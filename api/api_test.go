package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"reflect"
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

func TestSummitsTableHandler(t *testing.T) {
	db := MockDatabase(t)
	conf := &RuntimeConfig{Datadir: "testdata/summits"}
	sm := scs.New()
	sm.Store = &MockSessionStore{}

	if err := LoadSummits(conf.Datadir, db); err != nil {
		t.Fatal(err)
		return
	}

	api := NewApi(conf, db, sm)
	handler := sm.LoadAndSave(api)

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
		req, err := http.NewRequest("GET", "/summits", nil)
		if err != nil {
			t.Error(err)
			return
		}
		if tt.cookie != nil {
			req.AddCookie(tt.cookie)
		}

		handler.ServeHTTP(rr, req)

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
	db := MockDatabase(t)
	sm := scs.New()
	conf := &RuntimeConfig{Datadir: "testdata/summits"}
	api := NewApi(conf, db, sm)
	providers := make(AuthProviders)
	as := NewAuthServer(providers, db, sm)
	app := sm.LoadAndSave(&App{
		Api:        api,
		AuthServer: as,
		SM:         sm,
	})

	for _, tt := range cases {

		req, err := http.NewRequest("GET", tt.url, nil)
		if err != nil {
			t.Fatal(err)
		}
		rr := httptest.NewRecorder()

		app.ServeHTTP(rr, req)

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
	}{
		{"/top", "top-1.json"},
		{"/top?page=1", "top-1.json"},
		{"/top?page=2", "top-2.json"},
		{"/top?page=3", "top-3.json"},
		{"/summit/malidak/kirel", "summit-1.json"},
		{"/summit/malidak/kirel?page=2", "summit-1-page-2.json"},
		{"/summit/stolby/1021", "summit-2.json"},
		{"/summit/malidak/malinovaja", "summit-3.json"},
		{"/user/5", "user-1.json"},
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

	api := NewApi(conf, db, nil)

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
