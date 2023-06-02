package main

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"strconv"
	"strings"
	"time"

	"gopkg.in/yaml.v2"
)

const (
	ImageLarge  = "L"
	ImageMedium = "M"
	ImageSmall  = "S"
)

func toSqlNullInt64(value int64) sql.NullInt64 {
	var result sql.NullInt64
	if value == 0 {
		result.Valid = false
		return result
	}
	result.Valid = true
	result.Int64 = value
	return result
}

type InexactDate struct {
	Year  int64
	Month int64
	Day   int64
}

func (id *InexactDate) parseString(date string) ([]int64, error) {
	if date == "" {
		return []int64{}, nil
	}
	parts := strings.Split(date, ".")
	if len(parts) > 3 {
		return nil, fmt.Errorf("Invalid date format: %s", date)
	}
	partsInt := make([]int64, len(parts))
	for i, p := range parts {
		pi, err := strconv.ParseInt(p, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("Invalid date format: %s", date)
		}
		partsInt[i] = pi
	}
	return partsInt, nil
}

func (id *InexactDate) Parse(date string) error {
	parts, err := id.parseString(date)
	if err != nil {
		return err
	}
	dateFormat := "2006-01-02"
	var validationValue string
	var year, month, day int64
	switch len(parts) {
	case 0:
		id.Year, id.Month, id.Day = 0, 0, 0
		return nil
	case 1:
		year, month, day = parts[0], 0, 0
		validationValue = fmt.Sprintf("%04d-01-01", year)
	case 2:
		year, month, day = parts[1], parts[0], 0
		validationValue = fmt.Sprintf("%04d-%02d-01", year, month)
	case 3:
		year, month, day = parts[2], parts[1], parts[0]
		validationValue = fmt.Sprintf("%04d-%02d-%02d", year, month, day)
	default:
		// should not happen
		return fmt.Errorf("Invalid data: %v", parts)
	}
	_, err = time.Parse(dateFormat, validationValue)
	if err != nil {
		return fmt.Errorf("Failed to parse inexact date %s: %v", date, err)
	}

	id.Year, id.Month, id.Day = year, month, day
	return nil
}

type Ridge struct {
	Id    string `json:"id"`
	Name  string `json:"name"`
	Color string `json:"color"`
}

type SummitImage struct {
	Filename string `json:"filename"`
	Comment  string `json:"comment"`
}

type Summit struct {
	Id             string        `json:"id"`
	Name           *string       `json:"name"`
	NameAlt        *string       `json:"name_alt" yaml:"name_alt"`
	Interpretation *string       `json:"interpretation"`
	Description    *string       `json:"description"`
	Height         int           `json:"height"`
	Coordinates    [2]float32    `json:"coordinates"`
	Ridge          *Ridge        `json:"ridge"`
	Images         []SummitImage `json:"images"`
}

func (s *Summit) JSON() ([]byte, error) {
	// FIXME: use markdown instead of HTML
	// for summits description to avoid
	// using this custom encoder
	buffer := &bytes.Buffer{}
	encoder := json.NewEncoder(buffer)
	encoder.SetEscapeHTML(false)
	err := encoder.Encode(s)
	return buffer.Bytes(), err
}

type SummitsTableItem struct {
	Id     string  `json:"id"`
	Name   *string `json:"name"`
	Height int     `json:"height"`
	// Latitude is needed for sorting
	Lat       float32 `json:"lat"`
	RidgeName string  `json:"ridge"`
	RidgeId   string  `json:"ridge_id"`
	Visitors  int     `json:"visitors"`
	Rank      int     `json:"rank"`
	IsMain    bool    `json:"is_main"`
	Climbed   bool    `json:"climbed"`
}

type SummitsTable struct {
	Summits []SummitsTableItem `json:"summits"`
}

type TopItem struct {
	UserId    int    `json:"user_id"`
	UserName  string `json:"user_name"`
	ClimbsNum int    `json:"climbs_num"`
	Place     int    `json:"place"`
}

type Top struct {
	Items      []TopItem `json:"items"`
	Page       int       `json:"page"`
	TotalPages int       `json:"total_pages"`
}

type User struct {
	Id      int64
	OauthId string
	Src     int
	Name    string
}

func LoadSummitImages(images []SummitImage, summitId string, tx *sql.Tx) error {
	imageStmt, err := tx.Prepare(
		`INSERT INTO summit_images 
			(filename, summit_id, comment) VALUES (?, ?, ?)`)
	if err != nil {
		return err
	}
	defer imageStmt.Close()
	for _, img := range images {
		_, err = imageStmt.Exec(img.Filename, summitId, img.Comment)
		if err != nil {
			return fmt.Errorf("Failed to load image %s: %v", img.Filename, err)
		}
	}
	return nil
}

func LoadRidge(dir string, ridgeId string, tx *sql.Tx) error {
	summitDirs, err := os.ReadDir(dir)
	if err != nil {
		return err
	}
	summitsStmt, err := tx.Prepare(
		`INSERT INTO summits 
			(id, ridge_id, name, name_alt, interpretation, description, height, lat, lng)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`)
	if err != nil {
		return err
	}
	defer summitsStmt.Close()
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
			summit.Id, ridgeId, summit.Name, summit.NameAlt,
			summit.Interpretation, summit.Description,
			summit.Height, summit.Coordinates[0], summit.Coordinates[1])
		if err != nil {
			return err
		}
		if len(summit.Images) > 0 {
			err = LoadSummitImages(summit.Images, summit.Id, tx)
			if err != nil {
				return err
			}
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
		"DELETE FROM summit_images",
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

func FetchSummits(db *Database, userId int64) (*SummitsTable, error) {
	summits := make([]SummitsTableItem, 0)
	query := `SELECT s.id, s.name, s.height, s.lat, r.name, r.id, COUNT(c.user_id), 
			ROW_NUMBER() OVER (ORDER BY s.height DESC) as rank,
			EXISTS(
				SELECT * FROM 
					(
						SELECT ridge_id, max(height) AS maxheight
						FROM summits
						WHERE ridge_id=s.ridge_id GROUP BY ridge_id
					) as smtsg
					INNER JOIN summits smts
						ON smtsg.ridge_id=smts.ridge_id
						AND smts.height=smtsg.maxheight 
					WHERE id=s.id
			) AS is_main,
			EXISTS(
				SELECT * FROM climbs
				WHERE summit_id=s.id AND user_id = ?
			) as climbed
		FROM ridges r 
			INNER JOIN summits s ON r.id = s.ridge_id
			LEFT JOIN climbs c ON c.summit_id = s.id
		GROUP BY s.id, s.name, s.height, s.lat, r.name
		ORDER BY s.id
	`
	rows, err := db.Pool.Query(query, userId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var s SummitsTableItem
		err := rows.Scan(
			&s.Id, &s.Name, &s.Height, &s.Lat,
			&s.RidgeName, &s.RidgeId, &s.Visitors,
			&s.Rank, &s.IsMain, &s.Climbed,
		)
		if err != nil {
			return nil, err
		}
		summits = append(summits, s)
	}
	return &SummitsTable{summits}, nil
}

func FetchSummitImages(summit *Summit) error {
	return nil
}

func FetchSummit(db *Database, ridgeId, summitId string) (*Summit, error) {
	var summit Summit
	var ridge Ridge
	summit.Name = new(string)
	summit.NameAlt = new(string)
	summit.Images = make([]SummitImage, 0)
	summit.Ridge = &ridge
	summit.Coordinates = [2]float32{}

	query := `SELECT
		s.id, s.name, s.name_alt, s.interpretation, s.description, s.height, s.lat, s.lng,
		r.id, r.name, r.color
	FROM summits s INNER JOIN ridges r ON s.ridge_id = r.id
	WHERE r.id = ? AND s.id = ?
	`
	err := db.Pool.QueryRow(query, ridgeId, summitId).Scan(&summit.Id,
		&summit.Name, &summit.NameAlt, &summit.Interpretation,
		&summit.Description, &summit.Height, &summit.Coordinates[0], &summit.Coordinates[1],
		&summit.Ridge.Id, &summit.Ridge.Name, &summit.Ridge.Color,
	)
	if err == sql.ErrNoRows {
		return nil, nil // summit not found
	}
	if err != nil {
		return nil, err
	}

	err = FetchSummitImages(&summit)
	if err != nil {
		return nil, err
	}
	return &summit, nil
}

func FetchTop(db *Database, page, itemsPerPage int) (*Top, error) {
	var result Top
	result.Page = page
	tx, err := db.Pool.Begin()
	defer tx.Commit()
	query := `SELECT COUNT(DISTINCT user_id) 
        FROM users INNER JOIN climbs ON users.id=climbs.user_id`
	if err != nil {
		return nil, err
	}
	var totalItems int
	err = tx.QueryRow(query).Scan(&totalItems)
	if err != nil {
		return nil, err
	}
	result.TotalPages = totalItems/itemsPerPage + 1
	result.Items = make([]TopItem, 0)
	query = `SELECT users.id, users.name, count(*) as climbs, 
            MAX(coalesce(day, 32) | (coalesce(month, 13) << 8) | (coalesce(year, 2100) << 16)) 
                AS last_climb 
        FROM users INNER JOIN climbs ON users.id=climbs.user_id 
        GROUP BY users.id, users.name
        ORDER BY climbs DESC, last_climb ASC 
        LIMIT ? OFFSET ?`
	offset := (page - 1) * itemsPerPage
	rows, err := tx.Query(query, itemsPerPage, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	i := 0
	for rows.Next() {
		var ti TopItem
		var lastClimb int
		err := rows.Scan(&ti.UserId, &ti.UserName, &ti.ClimbsNum, &lastClimb)
		if err != nil {
			return nil, err
		}
		ti.Place = offset + i + 1
		result.Items = append(result.Items, ti)
		i++
	}
	return &result, nil
}

func CreateUser(db *Database, Name, OauthId string, Src int /*, Image, Preview string*/) (int64, error) {
	db.WriteLock.Lock()
	defer db.WriteLock.Unlock()

	query := "INSERT INTO users (name, oauth_id, src) VALUES (?, ?, ?)"
	res, err := db.Pool.Exec(query, Name, OauthId, Src)
	if err != nil {
		return 0, err
	}
	var userId int64
	userId, err = res.LastInsertId()
	if err != nil {
		return 0, err
	}
	return userId, nil
}

func getUser(row *sql.Row) (*User, error) {
	var user User
	err := row.Scan(&user.Id, &user.Name, &user.OauthId, &user.Src)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func GetUserById(db *Database, id int64) (*User, error) {
	query := "SELECT id, name, oauth_id, src FROM users WHERE id=?"
	return getUser(db.Pool.QueryRow(query, id))
}

func GetUser(db *Database, oauthId string, src int) (*User, error) {
	query := "SELECT id, name, oauth_id, src FROM users WHERE oauth_id=? AND src=?"
	return getUser(db.Pool.QueryRow(query, oauthId, src))
}

func UpdateUserImage(db *Database, userId int64, size string, data []byte) error {
	db.WriteLock.Lock()
	defer db.WriteLock.Unlock()

	query := `INSERT INTO user_images (user_id, size, data) VALUES (?, ?, ?)
	ON CONFLICT (user_id, size) DO UPDATE SET data=excluded.data
	`
	_, err := db.Pool.Exec(query, userId, size, data)
	if err != nil {
		return err
	}
	return nil
}

func GetUserImage(db *Database, userId int64, size string) ([]byte, error) {
	query := "SELECT data FROM user_images WHERE user_id=? AND size=?"
	var img []byte
	row := db.Pool.QueryRow(query, userId, size)
	err := row.Scan(&img)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return img, nil
}

func UpdateClimb(db *Database, summitId string, userId int64, date InexactDate, comment string) error {
	db.WriteLock.Lock()
	defer db.WriteLock.Unlock()

	query := `INSERT INTO climbs (
		user_id, summit_id, year, month, day, comment
	) VALUES (?, ?, ?, ?, ?, ?)
	ON CONFLICT (user_id, summit_id) 
	DO UPDATE SET year=excluded.year, month=excluded.month, day=excluded.day, comment=excluded.comment
	`
	_, err := db.Pool.Exec(
		query, userId, summitId,
		toSqlNullInt64(date.Year), toSqlNullInt64(date.Month), toSqlNullInt64(date.Day), comment)
	if err != nil {
		return err
	}
	return nil
}
