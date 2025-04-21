package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
	"golang.org/x/oauth2"
)

const (
	MockVKClientID     = "mock_vk_client_id"
	MockVKClientSecret = "mock_vk_client_secret"

	MockUserName = "Climbing User"
)

func TestVkGetUserId(t *testing.T) {
	vk := &VKProvider{}

	tokenDataRaw := `
	{
		"access_token": "533bacf01e11f55b536a565b57531ac114461ae8736d6506a3",
		"expires_in": 43200,
		"user_id": 2343
	}`
	tokenExtra := make(map[string]interface{})
	json.Unmarshal([]byte(tokenDataRaw), &tokenExtra)

	tokenWithUserId := (&oauth2.Token{}).WithExtra(tokenExtra)

	var cases = []struct {
		token          *oauth2.Token
		expectedUserId string
		expectedErr    error
	}{
		{
			&oauth2.Token{},
			"",
			errors.New("failed to get VK user Id"),
		},
		{
			tokenWithUserId,
			"2343",
			nil,
		},
	}
	for _, tt := range cases {
		userId, err := vk.GetUserId(tt.token)

		if userId != tt.expectedUserId {
			t.Fatalf("Incorrect userId returned: %v, expected %v", userId, tt.expectedUserId)
		}
		if tt.expectedErr == nil {
			if err != nil {
				t.Fatalf("Unexpected error returned: %v", tt.expectedErr)
			}
			return
		}
		if err.Error() != tt.expectedErr.Error() {
			t.Fatalf("Incorrect err returned: %v, expected %v", err, tt.expectedErr)
		}
	}
}

func HandleUserGet(w http.ResponseWriter, r *http.Request) {
	successfulResponse := `
	{
		"response": [
			{
				"id": %d,
				"photo_200_orig": "http://%s/img/ava_m.jpg",
				"has_photo": 1,
				"photo_50": "http://%s/img/ava_s.jpg",
				"first_name": "Climbing",
				"last_name": "User",
				"can_access_closed": true,
				"is_closed": false
			}
		]
	}
	`
	errorResponse := `
	{
		"error": {
			"error_code": 5,
			"error_msg": "User authorization failed: no access_token passed.",
			"request_params": [
				{
					"key": "method",
					"value": "users.get"
				},
				{
					"key": "oauth",
					"value": "1"
				},
				{
					"key": "v",
					"value": "5.131"
				}
			]
		}
	}
	`
	if r.Header.Get("Authorization") != "Bearer "+MockAccessToken {
		fmt.Fprint(w, errorResponse)
	} else {
		userId, _ := strconv.Atoi(MockOauthUserId)
		fmt.Fprintf(w, successfulResponse, userId, r.Host, r.Host)
	}

}

func HandleImage(w http.ResponseWriter, r *http.Request) {
	// Extract the image name from the URL path
	imgName := path.Base(r.URL.Path)
	f, err := os.Open(path.Join("testdata", imgName))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer f.Close()

	w.Header().Set("Content-Type", "image/jpeg")

	_, err = io.Copy(w, f)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func MockVKHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path == "/method/users.get" {
		HandleUserGet(w, r)
		return
	}
	if (r.URL.Path == "/img/ava_s.jpg") || (r.URL.Path == "/img/ava_m.jpg") {
		HandleImage(w, r)
		return
	}
	http.Error(w, "404", http.StatusNotFound)
}

func TestVKRegister(t *testing.T) {
	mockVKServer := httptest.NewServer(http.HandlerFunc(MockVKHandler))
	defer mockVKServer.Close()

	vk := &VKProvider{
		&oauth2.Config{},
		mockVKServer.URL,
	}
	token := &oauth2.Token{
		AccessToken: MockAccessToken,
	}
	db := MockDatabase(t)
	storage := NewStorage(db)
	userId, err := vk.Register(token, storage, context.Background())
	if err != nil {
		t.Fatalf("Registration error: %v", err)
	}
	user, err := storage.GetUser(MockOauthUserId, 1)
	if err != nil {
		t.Fatalf("Failed to get user %d: %v", userId, err)
	}

	expectedUser := User{
		Id:      userId,
		OauthId: MockOauthUserId,
		Src:     1,
		Name:    MockUserName,
		ImageS:  "users/13_S.jpg",
		ImageM:  "users/13_M.jpg",
	}
	if *user != expectedUser {
		t.Fatalf("Unexpected user created: %v, expected %v", *user, expectedUser)
	}
	cases := []struct {
		size     string
		expected string
	}{
		{ImageSmall, "testdata/ava_s.jpg"},
		{ImageMedium, "testdata/ava_m.jpg"},
	}
	for _, tt := range cases {
		img, err := storage.GetUserImage(userId, tt.size)
		if err != nil {
			t.Fatalf("Failed to get image for user %d: %v", userId, err)
		}
		assert.Equal(t, fmt.Sprintf("users/%d_%s.jpg", userId, tt.size), img)
		// TODO: check if file is uploaded to S3
		/*
			f, err := os.Open(tt.expected)
			if err != nil {
				panic(err)
			}
			defer f.Close()
			expectedImg, err := io.ReadAll(f)
			if err != nil {
				panic(err)
			}
			if !reflect.DeepEqual(img, expectedImg) {
				t.Fatalf("Unexpected image stored for user %d", userId)
			}
		*/
	}
}

func TestVKRegisterError(t *testing.T) {
	mockVKServer := httptest.NewServer(http.HandlerFunc(MockVKHandler))
	defer mockVKServer.Close()

	vk := &VKProvider{
		&oauth2.Config{},
		mockVKServer.URL,
	}
	token := &oauth2.Token{
		AccessToken: "incorrect token",
	}
	db := MockDatabase(t)
	storage := NewStorage(db)
	userId, err := vk.Register(token, storage, context.Background())
	if err == nil {
		t.Fatalf("Expected registration error, got userId=%d, err=%v", userId, err)
	}
	expectedError := errors.New("VK API error 5: User authorization failed: no access_token passed.")
	if err.Error() != expectedError.Error() {
		t.Fatalf("Unexpected error returned %v, expected %v", err, expectedError)
	}
}
