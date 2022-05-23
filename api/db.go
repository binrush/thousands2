package main

import (
	"database/sql"
	_ "github.com/mattn/go-sqlite3"
	"log"
	"sync"
)

type Database struct {
	Pool      *sql.DB
	WriteLock sync.Mutex
}

type Migration struct {
	Name    string
	Queries []string
}

var migrations []Migration = []Migration{
	{
		"Initial",
		[]string{
			"CREATE TABLE ridges (id TEXT, name TEXT, color TEXT, PRIMARY KEY (id))",
			`CREATE TABLE summits (
                id TEXT,
                ridge_id TEXT,
                name TEXT,
                name_alt TEXT,
                height INTEGER,
                description TEXT,
                interpretation TEXT,
                lat REAL,
                lon REAL,
                PRIMARY_KEY (id),
                FOREIGN KEY(ridge_id) REFERENCES ridges(id))
            `,
			"CREATE TABLE users (oauth_id TEXT, src TEXT, name TEXT, PRIMARY KEY (oauth_id, src))",
		},
	},
}

func NewDatabase(path string) (*Database, error) {
	pool, err := sql.Open("sqlite3", path)
	if err != nil {
		return nil, err
	}
	pool.SetMaxIdleConns(3)
	pool.SetMaxOpenConns(3)
	return &Database{Pool: pool}, nil
}

func (db *Database) Migrate() error {

	var err, rollbackErr error
	var stmt string

	stmt = "CREATE TABLE IF NOT EXISTS _migrations (name text, PRIMARY KEY (name))"
	_, err = db.Pool.Exec(stmt)
	if err != nil {
		return err
	}
	for _, m := range migrations {
		var cnt int
		stmt = "SELECT count(*) FROM _migrations WHERE name=?"
		err = db.Pool.QueryRow(stmt, m.Name).Scan(&cnt)
		if err != nil {
			return err
		}
		// migration already applied
		if cnt <= 0 {
			continue
		}
		// Running migration
		tx, err := db.Pool.Begin()
		if err != nil {
			return err
		}
		log.Printf("Executing migration %s\n", m.Name)
		for _, stmt := range m.Queries {
			_, err = tx.Exec(stmt)
			if err != nil {
				rollbackErr = tx.Rollback()
				if rollbackErr != nil {
					// log rollback error
					log.Printf("Rollback failed: %v\n", rollbackErr)
				}
				return err
			}
		}
		stmt = "INSERT INTO _migrations VALUES (?)"
		_, err = tx.Exec(stmt, m.Name)
		if err != nil {
			rollbackErr = tx.Rollback()
			if rollbackErr != nil {
				log.Printf("Rollback failed: %v\n", rollbackErr)
			}
		}
		err = tx.Commit()
		if err != nil {
			return err
		}
	}
	return nil
}
