package main

import (
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"log"
	"os"
	"path"
	"strings"
)

type Ridge struct {
	Id    string `json:"id"`
	Name  string `json:"name"`
	Color string `json:"color"`
}

type Summit struct {
	Id             string     `json:"id"`
	Name           string     `json:"name"`
	AltName        string     `json:"alt_name"`
	Interpretation string     `json:"interpretation"`
	Description    string     `json:"description"`
	Height         int        `json:"height"`
	Coordinates    [2]float32 `json:"coordinates"`
	Ridge          *Ridge     `json:"ridge"`
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

func LoadRidge(ridge *Ridge, dir string, result []Summit) ([]Summit, error) {
	summitDirs, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}
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
			log.Printf("Failed to load summit metadata: %v", err)
			continue
		}
		var summit Summit
		err = yaml.Unmarshal(summitData, &summit)
		if err != nil {
			log.Printf("Failed to parse summit metadata: %v", err)
			continue
		}
		summit.Id = summitId
		summit.Ridge = ridge
		result = append(result, summit)
	}
	return result, nil
}

func LoadSummits(dataDir string) ([]Summit, error) {
	result := make([]Summit, 0, 300)
	ridgeDirs, err := os.ReadDir(dataDir)
	if err != nil {
		return nil, err
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
			log.Printf("Failed to load ridge metadata: %v", err)
			continue
		}
		var ridge Ridge
		err = yaml.Unmarshal(ridgeData, &ridge)
		if err != nil {
			log.Printf("Failed to parse ridge metadata: %v", err)
			continue
		}
		ridge.Id = ridgeId
		newResult, err := LoadRidge(&ridge, ridgePath, result)
		if err != nil {
			log.Printf("Failed to load from ridge dir : %v", err)
			continue
		}
		result = newResult
	}
	return result, nil
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
