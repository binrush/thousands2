package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/vk"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"reflect"
	"testing"
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

func MockAuthProviders() AuthProviders {
	providers := make(AuthProviders)
	providers["vk"] = &oauth2.Config{
		RedirectURL:  "https://thousands.su/api/auth/authorized",
		ClientID:     "MOCK_VK_CLIENT_ID",
		ClientSecret: "MOCK_VK_CLIENT_SECRET",
		Scopes:       []string{},
		Endpoint:     vk.Endpoint,
	}
	return providers
}

func TestSummitsTableHandler(t *testing.T) {
	db := MockDatabase(t)
	conf := &RuntimeConfig{Datadir: "testdata/summits"}

	if err := LoadSummits(conf.Datadir, db); err != nil {
		t.Fatal(err)
		return
	}

	req, err := http.NewRequest("GET", "/summits/table", nil)
	if err != nil {
		t.Fatal(err)
		return
	}

	rr := httptest.NewRecorder()

	api := Api{Config: conf, DB: db}

	// Our handlers satisfy http.Handler, so we can call their ServeHTTP method
	// directly and pass in our Request and ResponseRecorder.
	api.ServeHTTP(rr, req)

	// Check the status code is what we expect.
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusOK)
		return
	}

	// Check the response body is what we expect.
	expected, err := ioutil.ReadFile("testdata/summits-table-expected.json")
	if err != nil {
		t.Errorf("Failed to read expected json data: %v", err)
		return
	}
	expected_str := string(expected)
	areEqual, err := AreEqualJSON(expected_str, rr.Body.String())
	if err != nil {
		fmt.Println("Error comparing response", err.Error())
		return
	}
	if !areEqual {
		t.Errorf("handler returned unexpected body: got %v want %v",
			rr.Body.String(), expected_str)
	}
}

/*func TestSummitHandler(t *testing.T) {
	db := MockDatabase(t)
	conf := &RuntimeConfig{Datadir: "testdata/summits"}

	if err := LoadSummits(conf.Datadir, db); err != nil {
		t.Fatal(err)
		return
	}

	req, err := http.NewRequest("GET", "/summit/malidak/kirel", nil)
	if err != nil {
		t.Fatal(err)
		return
	}

	rr := httptest.NewRecorder()

	api := Api{Config: conf, DB: db}

	// Our handlers satisfy http.Handler, so we can call their ServeHTTP method
	// directly and pass in our Request and ResponseRecorder.
	api.ServeHTTP(rr, req)

	// Check the status code is what we expect.
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusOK)
		return
	}
	expected, err := ioutil.ReadFile("testdata/summit-expected0.json")
	if err != nil {
		t.Errorf("Failed to read expected json data: %v", err)
		return
	}
	expected_str := string(expected)
	areEqual, err := AreEqualJSON(expected_str, rr.Body.String())
	if err != nil {
		fmt.Println("Error comparing response", err.Error())
		return
	}
	if !areEqual {
		t.Errorf("handler returned unexpected body: got %v want %v",
			rr.Body.String(), expected_str)
	}
}*/

func TestHandlersClientErrors(t *testing.T) {
	var cases = []struct {
		url          string
		expectedCode int
	}{
		{"/top?page=0", http.StatusBadRequest},
		{"/top?page=-1", http.StatusBadRequest},
		{"/top?page=1&page=2", http.StatusBadRequest},
		{"/summit", http.StatusNotFound},
		{"/summit/kyrel", http.StatusNotFound},
		{"/summit/malidak/kyrel/1", http.StatusNotFound},
	}

	for _, tt := range cases {

		req, err := http.NewRequest("GET", tt.url, nil)
		if err != nil {
			t.Fatal(err)
		}

		rr := httptest.NewRecorder()

		conf := &RuntimeConfig{Datadir: "testdata/summits"}
		api := Api{Config: conf}

		api.ServeHTTP(rr, req)

		if status := rr.Code; status != tt.expectedCode {
			t.Errorf("handler returned wrong status code: got %v want %v",
				status, http.StatusBadRequest)
		}
	}
}

func TestAuthHandler(t *testing.T) {
	req, err := http.NewRequest("GET", "/oauth/vk", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()

	as := AuthServer{Providers: MockAuthProviders()}

	as.ServeHTTP(rr, req)
	if status := rr.Code; status != http.StatusTemporaryRedirect {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusTemporaryRedirect)
	}
	resp := rr.Result()
	redirectUrl, err := url.Parse(resp.Header.Get("Location"))
	if err != nil {
		t.Errorf("Failed to parse url: %v", err)
	}
	if redirectUrl.Host != "oauth.vk.com" {
		t.Errorf("Invalid host in redirect url: got %v, expected oauth.vk.com",
			redirectUrl.Host)
	}
	//	query := redirectUrl.Query()
	//	expectedQuery := map[string][]string{
	//		"response_type": []string{"code"},
	//	}
	//	if ! reflect.DeepEqual(query, expectedQuery) {
	//		t.Fatalf("Unexpected query params: %v, expected %v", query, expectedQuery)
	//	}
}

func TestHandlersHappyPath(t *testing.T) {
	cases := []struct {
		url                string
		expectedResultFile string
	}{
		{"/top", "top-1.json"},
		{"/top?page=1", "top-1.json"},
		{"/top?page=2", "top-2.json"},
		{"/top?page=3", "top-3.json"},
		{"/summit/malidak/kirel", "summit-1.json"},
		{"/summit/stolby/1021", "summit-2.json"},
		{"/summit/malidak/malinovaja", "summit-3.json"},
	}
	db := MockDatabase(t)
	defer db.Pool.Close()
	conf := &RuntimeConfig{
		Datadir:      "testdata/summits",
		ItemsPerPage: 5,
	}

	if err := LoadSummits(conf.Datadir, db); err != nil {
		t.Fatal(err)
		return
	}

	api := Api{Config: conf, DB: db}

	for _, tt := range cases {

		req, err := http.NewRequest("GET", tt.url, nil)
		if err != nil {
			t.Fatal(err)
		}

		rr := httptest.NewRecorder()

		api.ServeHTTP(rr, req)
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
		expected, err := ioutil.ReadFile(
			filepath.Join("testdata/responses", tt.expectedResultFile))
		if err != nil {
			t.Errorf("Failed to read expected json data from %s: %v",
				tt.expectedResultFile, err)
			continue
		}
		expectedBody := string(expected)

		areEqual, err := AreEqualJSON(expectedBody, rr.Body.String())
		if err != nil {
			t.Errorf("Error comparing response: %v", err)
		}
		if !areEqual {
			t.Errorf("handler returned unexpected body for url %s: got %v want %v",
				tt.url, rr.Body.String(), expectedBody)
			continue
		}
	}

}
