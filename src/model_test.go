package main

import (
	"reflect"
	"testing"
)

/*
  Testing error handling in LoadSummits function
  Happy path is exercised in api_test.go
*/

func TestLoadSummits(t *testing.T) {
	cases := []string{
		"testdata/summits_broken0",
		"testdata/summits_broken1",
		"testdata/summits_broken2",
	}
	tables := []string{"ridges", "summits"}

	for _, datadir := range cases {
		db := MockDatabase(t)
		defer db.Close()
		storage := NewStorage(db)
		err := storage.LoadSummits(datadir)
		if err == nil {
			t.Fatalf("Error expected to be non-nil for %s", datadir)
		}
		for _, tbl := range tables {
			var numRows int
			err := db.QueryRow("SELECT COUNT(*) FROM " + tbl).Scan(&numRows)
			if err != nil {
				t.Fatalf("Error running sql query: %v", err)
			}
			if numRows > 0 {
				t.Fatalf("Unexpected rows in %s table: %v (expected 0)", tbl, numRows)
			}
		}
	}
}

func TestInexactDateParseValid(t *testing.T) {
	cases := []struct {
		input    string
		expected InexactDate
	}{
		{"", InexactDate{0, 0, 0}},
		{"2010", InexactDate{2010, 0, 0}},
		{"2.2010", InexactDate{2010, 2, 0}},
		{"12.06.2014", InexactDate{2014, 6, 12}},
	}
	for _, tt := range cases {
		var id InexactDate
		err := id.Parse(tt.input)
		if err != nil {
			t.Errorf("Unexpected error while parsing date string: %v", err)
		}
		if !reflect.DeepEqual(id, tt.expected) {
			t.Errorf("Unexpected parsing result: %v, expected %v", id, tt.expected)
		}
	}
}

func TestInexactDateParseInvalid(t *testing.T) {
	cases := []string{
		"abc", "1..2022", "1.1.02.2022", "13.2022", "29.2.2015",
	}
	for _, tt := range cases {
		var id InexactDate
		err := id.Parse(tt)
		if err == nil {
			t.Errorf("Error expected while parsing %v", tt)
		}
		if !reflect.DeepEqual(id, InexactDate{}) {
			t.Errorf("InexactData should not be changed in case of parsing error, got: %v", id)
		}
	}
}
