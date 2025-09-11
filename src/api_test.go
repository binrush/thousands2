package main

import (
	"bufio"
	"database/sql"
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

func MockDatabase(t *testing.T) *sql.DB {
	db, err := NewDatabase(":memory:")
	require.NoError(t, err)

	err = Migrate(db)
	require.NoError(t, err)

	file, err := os.Open("testdata/mock-db.sql")
	require.NoError(t, err)
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		_, err := db.Exec(scanner.Text())
		require.NoError(t, err)
	}
	return db
}

func GetMockApp(t *testing.T, userId int64, config *RuntimeConfig) *App {
	db := MockDatabase(t)
	sm := scs.New()
	sm.Store = NewMockSessionStore(userId)
	storage := NewStorage(db)
	err := storage.LoadSummits(config.Datadir)
	require.NoError(t, err)
	return NewAppServer(config, storage, sm)
}

func TestSummitsTableHandler(t *testing.T) {
	app := GetMockApp(t, 5, &RuntimeConfig{Datadir: "testdata/summits"})

	cases := []struct {
		name                    string
		cookie                  *http.Cookie
		expectedResonseFilename string
	}{
		{
			name:                    "unauthenticated request",
			cookie:                  nil,
			expectedResonseFilename: "summits-table-expected.json",
		},
		{
			name:                    "authenticated request",
			cookie:                  &http.Cookie{Name: "session", Value: "mock_session_token"},
			expectedResonseFilename: "summits-table-expected-authenticated.json",
		},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			rr := httptest.NewRecorder()
			req, err := http.NewRequest("GET", "/api/summits", nil)
			require.NoError(t, err)

			if tt.cookie != nil {
				req.AddCookie(tt.cookie)
			}

			app.router.ServeHTTP(rr, req)

			assert.Equal(t, http.StatusOK, rr.Code, "handler returned wrong status code")

			expected, err := os.ReadFile("testdata/" + tt.expectedResonseFilename)
			require.NoError(t, err, "Failed to read expected json data")

			assert.JSONEq(t, string(expected), rr.Body.String(), "Response body mismatch")
		})
	}
}

func TestHandlersClientErrors(t *testing.T) {
	cases := []struct {
		name         string
		url          string
		expectedCode int
	}{
		{"invalid page 0", "/api/top?page=0", http.StatusBadRequest},
		{"negative page", "/api/top?page=-1", http.StatusBadRequest},
		{"multiple pages", "/api/top?page=1&page=2", http.StatusBadRequest},
		{"missing summit path", "/api/summit", http.StatusNotFound},
		{"incomplete summit path", "/api/summit/kyrel", http.StatusNotFound},
		{"invalid summit path", "/api/summit/malidak/kyrel/1", http.StatusNotFound},
		{"invalid oauth provider", "/auth/oauth/invalidprovider", http.StatusNotFound},
		{"invalid oauth path", "/auth/oauth/vk/123", http.StatusNotFound},
		{"incomplete oauth path", "/auth/oauth", http.StatusNotFound},
		{"invalid authorized provider", "/auth/authorized/invalidprovider", http.StatusNotFound},
		{"invalid authorized path", "/auth/authorized/vk/123", http.StatusNotFound},
		{"incomplete authorized path", "/auth/authorized", http.StatusNotFound},
		{"random auth endpoint", "/auth/randomendpoint", http.StatusNotFound},
		{"unauthorized me", "/api/user/me", http.StatusUnauthorized},
		{"invalid user path", "/api/user/", http.StatusNotFound},
		{"non-numeric user id", "/api/user/abcd", http.StatusNotFound},
		{"invalid user check", "/api/user/1/check", http.StatusNotFound},
		{"non-existent user", "/api/user/123", http.StatusNotFound},
	}
	app := GetMockApp(t, 5, &RuntimeConfig{Datadir: "testdata/summits"})

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			req, err := http.NewRequest("GET", tt.url, nil)
			require.NoError(t, err)

			rr := httptest.NewRecorder()
			app.router.ServeHTTP(rr, req)

			assert.Equal(t, tt.expectedCode, rr.Code, "handler returned wrong status code for %s", tt.url)
		})
	}
}

func TestHandlersHappyPath(t *testing.T) {
	cases := []struct {
		name               string
		url                string
		expectedResultFile string
		cookie             *http.Cookie
	}{
		{"top page 1", "/api/top?page=1", "top-1.json", nil},
		{"top page 2", "/api/top?page=2", "top-2.json", nil},
		{"top page 3", "/api/top?page=3", "top-3.json", nil},
		{
			"authenticated summit",
			"/api/summit/malidak/kirel",
			"summit-1.json",
			&http.Cookie{Name: "session", Value: "mock_session_token"},
		},
		{"summit page 2", "/api/summit/malidak/kirel?page=2", "summit-1-page-2.json", nil},
		{"summit climbs", "/api/summit/malidak/kirel/climbs", "summit-climbs-1.json", nil},
		{"summit climbs page 2", "/api/summit/malidak/kirel/climbs?page=2", "summit-climbs-1-page-2.json", nil},
		{"stolby climbs", "/api/summit/stolby/1021/climbs", "summit-climbs-2.json", nil},
		{"stolby summit", "/api/summit/stolby/1021", "summit-2.json", nil},
		{"malinovaja summit", "/api/summit/malidak/malinovaja", "summit-3.json", nil},
		{"user profile", "/api/user/5", "user-1.json", nil},
		{"user climbs", "/api/user/5/climbs", "user-climbs-1.json", nil},
	}
	conf := &RuntimeConfig{
		Datadir:      "testdata/summits",
		ItemsPerPage: 5,
	}
	app := GetMockApp(t, 7, conf)

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			req, err := http.NewRequest("GET", tt.url, nil)
			require.NoError(t, err)

			if tt.cookie != nil {
				req.AddCookie(tt.cookie)
			}
			rr := httptest.NewRecorder()
			app.router.ServeHTTP(rr, req)

			require.Equal(t, http.StatusOK, rr.Code, "handler returned wrong status code")
			require.Equal(t, "application/json", rr.Header().Get("Content-Type"), "Wrong Content-Type header")

			expected, err := os.ReadFile(filepath.Join("testdata/responses", tt.expectedResultFile))
			require.NoError(t, err, "Failed to read expected json data")

			assert.JSONEq(t, string(expected), rr.Body.String(), "Response body mismatch")
		})
	}
}

func TestSummitPutHandler(t *testing.T) {
	cases := []struct {
		name         string
		url          string
		date         string
		comment      string
		expectedDate InexactDate
	}{
		{
			name:         "new climb",
			url:          "/api/summit/malidak/malinovaja",
			date:         "12.2002",
			comment:      "This is new climb",
			expectedDate: InexactDate{2002, 12, 0},
		},
		{
			name:         "update existing climb",
			url:          "/api/summit/malidak/kirel",
			date:         "12.06.2023",
			comment:      "This is a new comment",
			expectedDate: InexactDate{2023, 6, 12},
		},
	}
	userId := int64(7)

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			conf := &RuntimeConfig{Datadir: "testdata/summits"}
			app := GetMockApp(t, userId, conf)

			// Test data
			formData := url.Values{}
			formData.Set("date", tt.date)
			formData.Set("comment", tt.comment)

			// Make the PUT request
			rr := httptest.NewRecorder()
			req, err := http.NewRequest("PUT", tt.url, strings.NewReader(formData.Encode()))
			require.NoError(t, err)

			req.AddCookie(&http.Cookie{Name: "session", Value: "mock_session_token"})
			req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

			app.router.ServeHTTP(rr, req)
			assert.Equal(t, http.StatusOK, rr.Code, "handler returned wrong status code")

			// Verify the climb was recorded
			rr = httptest.NewRecorder()
			req, err = http.NewRequest("GET", tt.url, nil)
			require.NoError(t, err)

			req.AddCookie(&http.Cookie{Name: "session", Value: "mock_session_token"})
			app.router.ServeHTTP(rr, req)

			assert.Equal(t, http.StatusOK, rr.Code, "climbs endpoint returned wrong status code")

			var response Summit
			err = json.NewDecoder(rr.Body).Decode(&response)
			require.NoError(t, err, "Failed to decode response")

			require.NotNil(t, response.ClimbData, "New climb not found in the climbs list")
			assert.Equal(t, tt.comment, response.ClimbData.Comment, "Wrong comment")
			assert.Equal(t, tt.expectedDate, response.ClimbData.Date, "Wrong date")
		})
	}
}

func TestSummitPutHandlerErrors(t *testing.T) {
	cases := []struct {
		name           string
		url            string
		date           string
		comment        string
		userId         int64
		expectedStatus int
	}{
		{
			name:           "unauthenticated request",
			url:            "/api/summit/malidak/kirel",
			date:           "12.06.2023",
			comment:        "Test climb",
			userId:         0,
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "invalid date format",
			url:            "/api/summit/malidak/kirel",
			date:           "invalid-date",
			comment:        "Test climb",
			userId:         5,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "non-existing summit",
			url:            "/api/summit/malidak/nonexistent",
			date:           "12.06.2023",
			comment:        "Test climb",
			userId:         5,
			expectedStatus: http.StatusNotFound,
		},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			conf := &RuntimeConfig{Datadir: "testdata/summits"}
			app := GetMockApp(t, tt.userId, conf)

			formData := url.Values{}
			formData.Set("date", tt.date)
			formData.Set("comment", tt.comment)

			rr := httptest.NewRecorder()
			req, err := http.NewRequest("PUT", tt.url, strings.NewReader(formData.Encode()))
			require.NoError(t, err)

			if tt.userId > 0 {
				req.AddCookie(&http.Cookie{Name: "session", Value: "mock_session_token"})
			}
			req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

			app.router.ServeHTTP(rr, req)
			assert.Equal(t, tt.expectedStatus, rr.Code, "handler returned wrong status code")
		})
	}
}

func TestSummitDeleteHandler(t *testing.T) {
	conf := &RuntimeConfig{Datadir: "testdata/summits"}
	app := GetMockApp(t, 5, conf)

	// Make the DELETE request
	rr := httptest.NewRecorder()
	req, err := http.NewRequest("DELETE", "/api/summit/kurkak/kurkak", nil)
	require.NoError(t, err)

	req.AddCookie(&http.Cookie{Name: "session", Value: "mock_session_token"})
	app.router.ServeHTTP(rr, req)
	assert.Equal(t, http.StatusOK, rr.Code, "handler returned wrong status code")

	// Verify the climb was deleted
	rr = httptest.NewRecorder()
	req, err = http.NewRequest("GET", "/api/summit/kurkak/kurkak", nil)
	require.NoError(t, err)

	req.AddCookie(&http.Cookie{Name: "session", Value: "mock_session_token"})
	app.router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code, "climbs endpoint returned wrong status code")

	var response Summit
	err = json.NewDecoder(rr.Body).Decode(&response)
	require.NoError(t, err, "Failed to decode response")
	assert.Nil(t, response.ClimbData, "Climb data should be nil after deletion")
}

func TestSummitDeleteHandlerUnauthenticated(t *testing.T) {
	app := GetMockApp(t, 0, &RuntimeConfig{Datadir: "testdata/summits"})

	rr := httptest.NewRecorder()
	req, err := http.NewRequest("DELETE", "/api/summit/kurkak/kurkak", nil)
	require.NoError(t, err)

	app.router.ServeHTTP(rr, req)
	assert.Equal(t, http.StatusUnauthorized, rr.Code, "handler returned wrong status code")
}
