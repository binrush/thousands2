package main

import (
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
		defer db.Pool.Close()
		err := LoadSummits(datadir, db)
		if err == nil {
			t.Fatalf("Error expected to be non-nil for %s", datadir)
		}
		for _, tbl := range tables {
			var numRows int
			err := db.Pool.QueryRow("SELECT COUNT(*) FROM " + tbl).Scan(&numRows)
			if err != nil {
				t.Fatalf("Error running sql query: %v", err)
			}
			if numRows > 0 {
				t.Fatalf("Unexpected rows in %s table: %v (expected 0)", tbl, numRows)
			}
		}
	}
}
