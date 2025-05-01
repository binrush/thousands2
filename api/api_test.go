package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/alexedwards/scs/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func AreEqualJSON(s1, s2 string) (bool, error) {
	var o1 interface{}
	var o2 interface{}

	var err error
	err = json.Unmarshal([]byte(s1), &o1)
	if err != nil {
		return false, fmt.Errorf("Error mashalling string 1 :: %s", err.Error())
	}
	err = json.Unmarshal([]byte(s2), &o2)
	if err != nil {
		return false, fmt.Errorf("Error mashalling string 2 :: %s", err.Error())
	}

	return reflect.DeepEqual(o1, o2), nil
}

func MockDatabase(t *testing.T) *Database {
	db, err := NewDatabase(":memory:")
	if err != nil {
		t.Fatal(err)
	}
	err = db.Migrate()
	if err != nil {
		t.Fatal(err)
	}
	file, err := os.Open("testdata/mock-db.sql")
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		_, err := db.Pool.Exec(scanner.Text())
		if err != nil {
			t.Fatal(err)
		}
	}
	return db
}

func GetMockApp(t *testing.T, userId int64, config *RuntimeConfig) *App {
	db := MockDatabase(t)
	sm := scs.New()
	sm.Store = NewMockSessionStore(userId)
	storage := NewStorage(db)
	if err := storage.LoadSummits(config.Datadir); err != nil {
		t.Fatal(err)
	}
	return NewAppServer(config, storage, sm, "")
}

func TestSummitsTableHandler(t *testing.T) {
	app := GetMockApp(t, 5, &RuntimeConfig{Datadir: "testdata/summits"})

	cases := []struct {
		cookie                  *http.Cookie
		expectedResonseFilename string
	}{
		{nil, "summits-table-expected.json"},
		{
			&http.Cookie{Name: "session", Value: "mock_session_token"},
			"summits-table-expected-authenticated.json",
		},
	}
	for _, tt := range cases {
		rr := httptest.NewRecorder()
		req, err := http.NewRequest("GET", "/api/summits", nil)
		if err != nil {
			t.Error(err)
			return
		}
		if tt.cookie != nil {
			req.AddCookie(tt.cookie)
		}

		app.router.ServeHTTP(rr, req)

		if status := rr.Code; status != http.StatusOK {
			t.Errorf("handler returned wrong status code: got %v want %v",
				status, http.StatusOK)
			return
		}

		expected, err := os.ReadFile("testdata/" + tt.expectedResonseFilename)
		if err != nil {
			t.Errorf("Failed to read expected json data: %v", err)
			return
		}
		expectedStr := string(expected)
		actualStr := rr.Body.String()
		require.JSONEq(t, expectedStr, actualStr, "Response body mismatch")
	}
}

func TestHandlersClientErrors(t *testing.T) {
	var cases = []struct {
		url          string
		expectedCode int
	}{
		{"/api/top?page=0", http.StatusBadRequest},
		{"/api/top?page=-1", http.StatusBadRequest},
		{"/api/top?page=1&page=2", http.StatusBadRequest},
		{"/api/summit", http.StatusNotFound},
		{"/api/summit/kyrel", http.StatusNotFound},
		{"/api/summit/malidak/kyrel/1", http.StatusNotFound},
		{"/auth/oauth/invalidprovider", http.StatusNotFound},
		{"/auth/oauth/vk/123", http.StatusNotFound},
		{"/auth/oauth", http.StatusNotFound},
		{"/auth/authorized/invalidprovider", http.StatusNotFound},
		{"/auth/authorized/vk/123", http.StatusNotFound},
		{"/auth/authorized", http.StatusNotFound},
		{"/auth/randomendpoint", http.StatusNotFound},
		{"/api/user/me", http.StatusUnauthorized},
		{"/api/user/", http.StatusNotFound},
		{"/api/user/abcd", http.StatusNotFound},
		{"/api/user/1/check", http.StatusNotFound},
		{"/api/user/123", http.StatusNotFound},
	}
	app := GetMockApp(t, 5, &RuntimeConfig{Datadir: "testdata/summits"})

	for _, tt := range cases {

		req, err := http.NewRequest("GET", tt.url, nil)
		if err != nil {
			t.Fatal(err)
		}
		rr := httptest.NewRecorder()

		app.router.ServeHTTP(rr, req)

		if status := rr.Code; status != tt.expectedCode {
			t.Errorf("handler returned wrong status code for %s: got %v want %v",
				tt.url, status, tt.expectedCode)
		}
	}
}

func TestHandlersHappyPath(t *testing.T) {
	cases := []struct {
		url                string
		expectedResultFile string
		cookie             *http.Cookie
	}{
		{"/api/top?page=1", "top-1.json", nil},
		{"/api/top?page=2", "top-2.json", nil},
		{"/api/top?page=3", "top-3.json", nil},
		{"/api/summit/malidak/kirel", "summit-1.json", &http.Cookie{Name: "session", Value: "mock_session_token"}},
		{"/api/summit/malidak/kirel?page=2", "summit-1-page-2.json", nil},
		{"/api/summit/malidak/kirel/climbs", "summit-climbs-1.json", nil},
		{"/api/summit/malidak/kirel/climbs?page=2", "summit-climbs-1-page-2.json", nil},
		{"/api/summit/stolby/1021/climbs", "summit-climbs-2.json", nil},
		{"/api/summit/stolby/1021", "summit-2.json", nil},
		{"/api/summit/malidak/malinovaja", "summit-3.json", nil},
		{"/api/user/5", "user-1.json", nil},
	}
	conf := &RuntimeConfig{
		Datadir:      "testdata/summits",
		ItemsPerPage: 5,
	}
	app := GetMockApp(t, 7, conf)

	for _, tt := range cases {
		req, err := http.NewRequest("GET", tt.url, nil)
		if err != nil {
			t.Fatal(err)
		}
		if tt.cookie != nil {
			req.AddCookie(tt.cookie)
		}
		rr := httptest.NewRecorder()

		app.router.ServeHTTP(rr, req)
		res := rr.Result()
		if status := res.StatusCode; status != http.StatusOK {
			t.Errorf("handler returned wrong status code for url %s: got %v want %v.",
				tt.url, status, http.StatusOK)
			continue
		}
		contentType := res.Header.Get("Content-Type")
		if contentType != "application/json" {
			t.Errorf("Wrong Content-Type header returned for url %s: got %v, expected application/json",
				tt.url, contentType)
			continue
		}
		expected, err := os.ReadFile(
			filepath.Join("testdata/responses", tt.expectedResultFile))
		if err != nil {
			t.Errorf("Failed to read expected json data from %s: %v",
				tt.expectedResultFile, err)
			continue
		}
		expectedBody := string(expected)

		assert.JSONEq(t, expectedBody, rr.Body.String(), "Response body mismatch for %s", tt.url)
	}
}

func TestSummitPutHandler(t *testing.T) {
	cases := []struct {
		url          string
		date         string
		comment      string
		expectedDate InexactDate
	}{
		// non-existing climb
		{"/api/summit/malidak/malinovaja", "12.2002", "This is new climb", InexactDate{2002, 12, 0}},
		// existing climb
		{"/api/summit/malidak/kirel", "12.06.2023", "This is a new comment", InexactDate{2023, 6, 12}},
	}
	userId := int64(7)
	for _, tt := range cases {
		conf := &RuntimeConfig{Datadir: "testdata/summits"}
		app := GetMockApp(t, userId, conf)

		// Test data
		formData := url.Values{}
		formData.Set("date", tt.date)
		formData.Set("comment", tt.comment)

		// First, make the PUT request to record the climb
		rr := httptest.NewRecorder()
		req, err := http.NewRequest("PUT", tt.url, strings.NewReader(formData.Encode()))
		if err != nil {
			t.Error(err)
			continue
		}
		req.AddCookie(&http.Cookie{Name: "session", Value: "mock_session_token"})
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

		app.router.ServeHTTP(rr, req)

		if status := rr.Code; status != http.StatusOK {
			t.Errorf("handler returned wrong status code: got %v want %v",
				status, http.StatusOK)
			continue
		}

		// Now, verify the climb was recorded by checking the climbs endpoint
		rr = httptest.NewRecorder()
		req, err = http.NewRequest("GET", tt.url, nil)
		if err != nil {
			t.Error(err)
			continue
		}
		req.AddCookie(&http.Cookie{Name: "session", Value: "mock_session_token"})

		app.router.ServeHTTP(rr, req)

		if status := rr.Code; status != http.StatusOK {
			t.Errorf("climbs endpoint returned wrong status code: got %v want %v",
				status, http.StatusOK)
			continue
		}

		// Parse the response
		var response Summit
		if err := json.NewDecoder(rr.Body).Decode(&response); err != nil {
			t.Errorf("Failed to decode response: %v", err)
			continue
		}

		// Find our newly added climb
		if response.ClimbData == nil {
			t.Error("New climb not found in the climbs list")
			continue
		}

		// Verify the climb details
		if response.ClimbData.Comment != tt.comment {
			t.Errorf("Expected comment '%s', got '%s'", tt.comment, response.ClimbData.Comment)
		}
		if response.ClimbData.Date != tt.expectedDate {
			t.Errorf("Expected date %v, got %v", tt.expectedDate, response.ClimbData.Date)
		}
	}
}
