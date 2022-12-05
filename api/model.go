package main

import (
	"database/sql"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"os"
	"path"
	"strings"
	"errors"
	"fmt"
)

type Ridge struct {
	Id    string `json:"id"`
	Name  string `json:"name"`
	Color string `json:"color"`
}

type Summit struct {
	Id             string     `json:"id"`
	Name           *string    `json:"name"`
	AltName        *string    `json:"alt_name"`
	Interpretation *string    `json:"interpretation"`
	Description    *string    `json:"description"`
	Height         int        `json:"height"`
	Coordinates    [2]float32 `json:"coordinates"`
	Ridge          *Ridge     `json:"ridge"`
}

type SummitsTableItem struct {
	Id        string `json:"id"`
	Name      string `json:"name"`
	Height    int    `json:"height"`
	RidgeName string `json:"ridge"`
	Visitors  int    `json:"visitors"`
}

type SummitsTable struct {
	Summits []SummitsTableItem `json:"summits"`
}

type TopItem struct {
	UserId      int    `json:"user_id"`
	UserName    string `json:"user_name"`
	UserPreview string `json:"user_preview"`
	ClimbsNum   int    `json:"climbs_num"`
	Place       int    `json:"place"`
}

type Top struct {
	Items      []TopItem `json:"items"`
	Page       int       `json:"page"`
	TotalPages int       `json:"total_pages"`
}

func LoadRidge(dir string, ridgeId string, tx *sql.Tx) error {
	summitDirs, err := os.ReadDir(dir)
	if err != nil {
		return err
	}
	summitsStmt, err := tx.Prepare(
		`INSERT INTO summits 
			(id, ridge_id, name, alt_name, interpretation, description, height, lat, lng)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`)
	defer summitsStmt.Close()
	if err != nil {
		return err
	}
	summitsNum := 0
	for _, summitDir := range summitDirs {
		if !summitDir.IsDir() {
			continue
		}
		summitId := summitDir.Name()
		if strings.HasPrefix(summitId, ".") {
			continue
		}
		summitPath := path.Join(dir, summitId)
		summitData, err := ioutil.ReadFile(path.Join(summitPath, "meta.yaml"))
		if err != nil {
			return err
		}
		var summit Summit
		err = yaml.Unmarshal(summitData, &summit)
		if err != nil {
			return err
		}
		summit.Id = summitId
		_, err = summitsStmt.Exec(
			summit.Id, ridgeId, summit.Name, summit.AltName,
			summit.Interpretation, summit.Description,
			summit.Height, summit.Coordinates[0], summit.Coordinates[1])
		if err != nil {
			return err
		}
		summitsNum += 1
	}
	if summitsNum <= 0 {
		return errors.New(fmt.Sprintf("Error: empty ridges are not allowed: %s", ridgeId))
	}
	return nil
}

func LoadSummits(dataDir string, db *Database) error {
	tx, err := db.Pool.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	cleanupQueries := []string{
		"DELETE FROM summits",
		"DELETE FROM ridges",
	}
	for _, sql := range cleanupQueries {
		_, err = tx.Exec(sql)
		if err != nil {
			return err
		}
	}
	ridgeStmt, err := tx.Prepare("INSERT INTO ridges VALUES (?, ?, ?)")
	if err != nil {
		return err
	}
	defer ridgeStmt.Close()

	ridgeDirs, err := os.ReadDir(dataDir)
	if err != nil {
		return err
	}
	for _, ridgeDir := range ridgeDirs {
		if !ridgeDir.IsDir() {
			continue
		}
		ridgeId := ridgeDir.Name()
		if strings.HasPrefix(ridgeId, ".") {
			continue
		}
		ridgePath := path.Join(dataDir, ridgeId)
		ridgeData, err := ioutil.ReadFile(path.Join(ridgePath, "meta.yaml"))
		if err != nil {
			return err
		}
		var ridge Ridge
		if err := yaml.Unmarshal(ridgeData, &ridge); err != nil {
			return err
		}
		ridge.Id = ridgeId
		_, err = ridgeStmt.Exec(ridge.Id, ridge.Name, ridge.Color)
		if err != nil {
			return err
		}

		if err = LoadRidge(ridgePath, ridge.Id, tx); err != nil {
			return err
		}
	}
	if err := tx.Commit(); err != nil {
		return err
	}
	return nil
}

func FetchSummitsTable(db *Database) (*SummitsTable, error) {
	summits := make([]SummitsTableItem, 0)
	sql := `SELECT s.id, COALESCE(s.name, s.height), s.height, r.name, COUNT(c.user_id)
		FROM ridges r 
			INNER JOIN summits s ON r.id = s.ridge_id
			LEFT JOIN climbs c ON c.summit_id = s.id
		GROUP BY s.id, s.name, s.height
	`
	rows, err := db.Pool.Query(sql)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var s SummitsTableItem
		if err := rows.Scan(&s.Id, &s.Name, &s.Height, &s.RidgeName, &s.Visitors); err != nil {
			return nil, err
		}
		summits = append(summits, s)
	}
	return &SummitsTable{summits}, nil
}

func FetchTop(db *Database, page, itemsPerPage int) (*Top, error) {
	var result Top
	result.Page = page
	tx, err := db.Pool.Begin()
	defer tx.Commit()
	sql := `SELECT COUNT(DISTINCT user_id) 
        FROM users INNER JOIN climbs ON users.id=climbs.user_id`
	if err != nil {
		return nil, err
	}
	var totalItems int
	err = tx.QueryRow(sql).Scan(&totalItems)
	if err != nil {
		return nil, err
	}
	result.TotalPages = totalItems/itemsPerPage + 1
	result.Items = make([]TopItem, 0)
	sql = `SELECT users.id, users.name, users.preview, 
            count(*) as climbs, 
            MAX(coalesce(day, 32) | (coalesce(month, 13) << 8) | (coalesce(year, 2100) << 16)) 
                AS last_climb 
        FROM users INNER JOIN climbs ON users.id=climbs.user_id 
        GROUP BY users.id, users.name, users.preview 
        ORDER BY climbs DESC, last_climb ASC 
        LIMIT ? OFFSET ?`
	offset := (page - 1) * itemsPerPage
	rows, err := tx.Query(sql, itemsPerPage, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	i := 0
	for rows.Next() {
		var ti TopItem
		var lastClimb int
		err := rows.Scan(&ti.UserId, &ti.UserName, &ti.UserPreview, &ti.ClimbsNum, &lastClimb)
		if err != nil {
			return nil, err
		}
		ti.Place = offset + i + 1
		result.Items = append(result.Items, ti)
		i++
	}
	return &result, nil
}
