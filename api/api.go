package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
)

const (
	internalServerErrorMsg = "Internal server error"
	notFoundMsg            = "Path not found"
)

type ApiError struct {
	Message    string `json:"error"`
	StatusCode int
}

func (e *ApiError) Error() string {
	return e.Message
}

var pathNotFoundError = &ApiError{notFoundMsg, http.StatusNotFound}
var serverError = &ApiError{internalServerErrorMsg, http.StatusInternalServerError}

type Api struct {
	Config *RuntimeConfig
	DB     *Database
}

func (h *Api) HandleSummit(r *http.Request) interface{} {
	if r.URL.Path == "/" {
		return pathNotFoundError
	}

	var ridgeId, summitId string
	ridgeId, r.URL.Path = ShiftPath(r.URL.Path)
	if r.URL.Path == "/" {
		return pathNotFoundError
	}

	summitId, r.URL.Path = ShiftPath(r.URL.Path)
	if r.URL.Path != "/" {
		return pathNotFoundError
	}

	summit, err := FetchSummit(h.DB, ridgeId, summitId)
	if err != nil {
		log.Printf("Failed to fetch summit %s/%s: %v", ridgeId, summitId, err)
		return serverError
	}
	if summit == nil { // summit not found
		return pathNotFoundError
	}
	return summit
}

func (h *Api) HandleSummitsTable(r *http.Request) interface{} {
	if r.URL.Path != "/" {
		return pathNotFoundError
	}
	summits, err := FetchSummitsTable(h.DB)
	if err != nil {
		log.Printf("Failed to fetch summits from db: %v", err)
		return serverError
	}
	return summits
}

func (h *Api) HandleTop(r *http.Request) interface{} {

	if r.URL.Path != "/" {
		return pathNotFoundError
	}
	var page int
	var err error
	pageParam := r.URL.Query()["page"]
	if pageParam == nil {
		page = 1
	} else if len(pageParam) != 1 {
		return &ApiError{"Invalid page parameter provided", http.StatusBadRequest}
	} else {
		page, err = strconv.Atoi(pageParam[0])
		if (err != nil) || (page <= 0) {
			return &ApiError{"Invalid page parameter provided", http.StatusBadRequest}
		}
	}
	top, err := FetchTop(h.DB, page, h.Config.ItemsPerPage)
	if err != nil {
		log.Printf("Failed to fetch top: %v", err)
		return serverError
	}
	return top
}

func (h *Api) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var head string
	head, r.URL.Path = ShiftPath(r.URL.Path)
	var resp interface{}

	switch head {
	case "summit":
		if r.Method != "GET" {
			http.Error(w, "Only GET is allowed", http.StatusMethodNotAllowed)
			return
		}
		resp = h.HandleSummit(r)
	case "summits":
		if r.Method != "GET" {
			http.Error(w, "Only GET is allowed", http.StatusMethodNotAllowed)
			return
		}
		resp = h.HandleSummitsTable(r)
	case "top":
		if r.Method != "GET" {
			http.Error(w, "Only GET is allowed", http.StatusMethodNotAllowed)
			return
		}
		resp = h.HandleTop(r)
	default:
		resp = pathNotFoundError
	}

	jsonResp, err := json.Marshal(resp)
	if err != nil {
		log.Printf("Error marshalling top: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")

	if apiError, ok := resp.(*ApiError); ok {
		http.Error(w, string(jsonResp), apiError.StatusCode)
		return
	}
	fmt.Fprintf(w, string(jsonResp))
}
