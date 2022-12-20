package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
)

type Api struct {
	Config *RuntimeConfig
	DB     *Database
}

func (h *Api) HandleSummits(w http.ResponseWriter, r *http.Request) {
	var head string
	head, r.URL.Path = ShiftPath(r.URL.Path)
	switch head {
	case "table":
		h.HandleSummitsTable(w, r)
	default:
		http.NotFound(w, r)
	}
}

func (h *Api) HandleSummit(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path == "/" {
		http.NotFound(w, r)
		return
	}

	var ridgeId, summitId string
	ridgeId, r.URL.Path = ShiftPath(r.URL.Path)
	if r.URL.Path == "/" {
		http.NotFound(w, r)
		return
	}

	summitId, r.URL.Path = ShiftPath(r.URL.Path)
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}

	summit, err := FetchSummit(h.DB, ridgeId, summitId)
	if err != nil {
		log.Printf("Error: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
    if summit == nil { // summit not found
		log.Printf("Summit %s/%s not found", ridgeId, summitId)
        http.NotFound(w, r)
        return
    }
	resp, err := summit.JSON()
	if err != nil {
		log.Printf("Error: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	fmt.Fprintf(w, string(resp))
}

func (h *Api) HandleSummitsTable(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "Only GET is allowed", http.StatusMethodNotAllowed)
		return
	}
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	summits, err := FetchSummitsTable(h.DB)
	if err != nil {
		log.Printf("Error: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	resp, err := json.Marshal(summits)
	if err != nil {
		log.Printf("Error: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	fmt.Fprintf(w, string(resp))
}

func (h *Api) HandleTop(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "Only GET is allowed", http.StatusMethodNotAllowed)
		return
	}
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	var page int
	var err error
	pageParam := r.URL.Query()["page"]
	if pageParam == nil {
		page = 1
	} else if len(pageParam) != 1 {
		http.Error(w, "Invalid page parameger provided", http.StatusBadRequest)
		return
	} else {
		page, err = strconv.Atoi(pageParam[0])
		if (err != nil) || (page <= 0) {
			http.Error(w, "Invalid page parameter provided", http.StatusBadRequest)
			return
		}
	}
	top, err := FetchTop(h.DB, page, h.Config.ItemsPerPage)
	if err != nil {
		log.Printf("Error fetching top: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	resp, err := json.Marshal(top)
	if err != nil {
		log.Printf("Error marshalling top: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	fmt.Fprintf(w, string(resp))
}

func (h *Api) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var head string
	head, r.URL.Path = ShiftPath(r.URL.Path)
	switch head {
	case "summit":
		h.HandleSummit(w, r)
	case "summits":
		h.HandleSummits(w, r)
	case "top":
		h.HandleTop(w, r)
	default:
		http.NotFound(w, r)
	}
}
