package main

import (
	"database/sql"
	"fmt"
	"log"
	"sync"

	_ "github.com/mattn/go-sqlite3"
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
			`CREATE TABLE users (
                id INTEGER PRIMARY KEY, 
                oauth_id TEXT NOT NULL, 
                src INTEGER NOT NULL, 
                name TEXT NOT NULL
            )`,
			`CREATE TABLE user_images (
				user_id INTEGER NOT NULL,
				size TEXT NOT NULL,
				url TEXT NOT NULL,
				PRIMARY KEY (user_id, size),
				FOREIGN KEY (user_id) REFERENCES users(id)
			)`,
			`CREATE TABLE climbs (
                user_id INTEGER NOT NULL, 
                summit_id TEXT NOT NULL, 
                year INTEGER, month INTEGER, day INTEGER, 
                comment TEXT,
                PRIMARY KEY (user_id, summit_id),
                FOREIGN KEY(user_id) REFERENCES users(id)
            )`,
			`CREATE TABLE ridges (
				id TEXT NOT NULL PRIMARY KEY,
				name TEXT NOT NULL,
				color TEXT NOT NULL
			)`,
			`CREATE TABLE summits (
				id TEXT NOT NULL PRIMARY KEY,
				ridge_id TEXT NOT NULL,
				name TEXT,
				name_alt TEXT,
				interpretation TEXT,
				description TEXT,
				height INTEGER NOT NULL,
				lat REAL NOT NULL,
				lng REAL NOT NULL,
				FOREIGN KEY (ridge_id) REFERENCES ridges(id)
			)`,
			`CREATE TABLE summit_images (
				url TEXT PRIMARY KEY,
				summit_id TEXT NOT NULL,
				comment TEXT NOT NULL,
				FOREIGN KEY (summit_id) REFERENCES summits(id)
			)`,
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
		if cnt > 0 {
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
				return fmt.Errorf("statement %s failed with error: %v", stmt, err)
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
