package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
    "net/url"
	"reflect"
	"testing"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/vk"
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


func TestSummitsHandler(t *testing.T) {
	db := MockDatabase(t)
	defer db.Pool.Close()

	req, err := http.NewRequest("GET", "/summits", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()

	conf := &RuntimeConfig{Datadir: "testdata/summits"}
	api := Api{Config: conf}

	// Our handlers satisfy http.Handler, so we can call their ServeHTTP method
	// directly and pass in our Request and ResponseRecorder.
	api.ServeHTTP(rr, req)

	// Check the status code is what we expect.
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}

	// Check the response body is what we expect.
	expected, err := ioutil.ReadFile("testdata/summits-expected.json")
	if err != nil {
		t.Errorf("Failed to read expected json data: %v", err)
	}
	expected_str := string(expected)
	areEqual, err := AreEqualJSON(expected_str, rr.Body.String())
	if err != nil {
		fmt.Println("Error comparing response", err.Error())
	}
	if !areEqual {
		t.Errorf("handler returned unexpected body: got %v want %v",
			rr.Body.String(), expected_str)
	}
}

func TestAuthHandler(t *testing.T) {
	req, err := http.NewRequest("GET", "/auth/vk", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()

	conf := &RuntimeConfig{AuthConfig: MockAuthProviders()}
	api := Api{Config: conf}

	api.ServeHTTP(rr, req)
	if status := rr.Code; status != http.StatusTemporaryRedirect {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusTemporaryRedirect)
	}
    resp := rr.Result()
    redirect_url, err := url.Parse(resp.Header.Get("Location"))
    if err != nil {
        t.Errorf("Failed to parse url: %v", err)
    }
    if redirect_url.Host != "oauth.vk.com" {
        t.Errorf("Invalid host in redirect url: got %v, expected oauth.vk.com",
                 redirect_url.Host)
    }

}
