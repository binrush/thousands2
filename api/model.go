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
