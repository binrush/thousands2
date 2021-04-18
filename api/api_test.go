package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
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

func TestSummitsHandler(t *testing.T) {
	req, err := http.NewRequest("GET", "/summits", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()

	conf := &RuntimeConfig{datadir: "testdata/summits"}
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
