package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strconv"
)

var ErrPathNotFound = errors.New("Path not found")
var ErrInvalidParams = errors.New("Invalid parameters passed")

type InvalidParameterError struct {
	Message string
}

func (e *InvalidParameterError) Error() string {
	return e.Message
}

type ApiError struct {
	Error string `json:"error"`
}

type Api struct {
	Config *RuntimeConfig
	DB     *Database
}

func (h *Api) HandleSummit(r *http.Request) (interface{}, error) {
	if r.URL.Path == "/" {
		return nil, ErrPathNotFound
	}

	var ridgeId, summitId string
	ridgeId, r.URL.Path = ShiftPath(r.URL.Path)
	if r.URL.Path == "/" {
		return nil, ErrPathNotFound
	}

	summitId, r.URL.Path = ShiftPath(r.URL.Path)
	if r.URL.Path != "/" {
		return nil, ErrPathNotFound
	}

	summit, err := FetchSummit(h.DB, ridgeId, summitId)
	if err != nil {
		return nil, fmt.Errorf("Failed to fetch summit: %v", err)
	}
	if summit == nil { // summit not found
		return nil, ErrPathNotFound
	}
	return summit, nil
}

func (h *Api) HandleSummitsTable(r *http.Request) (interface{}, error) {
	if r.URL.Path != "/" {
		return nil, ErrPathNotFound
	}
	summits, err := FetchSummitsTable(h.DB)
	if err != nil {
		return nil, fmt.Errorf("Failed to fetch summits from db: %v", err)
	}
	return summits, nil
}

func (h *Api) HandleTop(r *http.Request) (interface{}, error) {

	if r.URL.Path != "/" {
		return nil, ErrPathNotFound
	}
	var page int
	var err error
	pageParam := r.URL.Query()["page"]
	if pageParam == nil {
		page = 1
	} else if len(pageParam) != 1 {
		return nil, &InvalidParameterError{"Invalid page parameter provided"}
	} else {
		page, err = strconv.Atoi(pageParam[0])
		if (err != nil) || (page <= 0) {
			return nil, &InvalidParameterError{"Invalid page parameter provided"}
		}
	}
	top, err := FetchTop(h.DB, page, h.Config.ItemsPerPage)
	if err != nil {
		return nil, fmt.Errorf("Failed to fetch top: %v", err)
	}
	return top, nil
}

func (h *Api) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var head string
	head, r.URL.Path = ShiftPath(r.URL.Path)
	var resp interface{}
	var handlerErr error

	switch head {
	case "summit":
		if r.Method != "GET" {
			http.Error(w, "Only GET is allowed", http.StatusMethodNotAllowed)
			return
		}
		resp, handlerErr = h.HandleSummit(r)
	case "summits":
		if r.Method != "GET" {
			http.Error(w, "Only GET is allowed", http.StatusMethodNotAllowed)
			return
		}
		resp, handlerErr = h.HandleSummitsTable(r)
	case "top":
		if r.Method != "GET" {
			http.Error(w, "Only GET is allowed", http.StatusMethodNotAllowed)
			return
		}
		resp, handlerErr = h.HandleTop(r)
	default:
		handlerErr = ErrPathNotFound
	}

	_, isParamError := handlerErr.(*InvalidParameterError)
	switch {
	case handlerErr == nil:
		jsonResp, err := json.Marshal(resp)
		if err != nil {
			log.Printf("Error marshalling top: %v", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, string(jsonResp))
	case handlerErr == ErrPathNotFound:
		http.Error(w, "Not found", http.StatusNotFound)
	case isParamError:
		jsonResp, err := json.Marshal(ApiError{handlerErr.Error()})
		if err != nil {
			log.Printf("Error marshalling error message: %v", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}
		http.Error(w, string(jsonResp), http.StatusBadRequest)
	default:
		fmt.Printf("Error handling request: %v", handlerErr)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}
}
