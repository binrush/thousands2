package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/alexedwards/scs/v2"
)

const (
	internalServerErrorMsg = "Internal server error"
	notFoundMsg            = "Path not found"
	methodNotAllowedMsg    = "Method not allowed"
	authRequiredMsg        = "Authentication required"
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
var methodNotAllowedError = &ApiError{methodNotAllowedMsg, http.StatusInternalServerError}
var authRequired = &ApiError{authRequiredMsg, http.StatusUnauthorized}

type Api struct {
	Config *RuntimeConfig
	DB     *Database
	SM     *scs.SessionManager
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
	switch r.Method {
	case http.MethodGet:
		return summit
	case http.MethodPut:
		return h.HandleUpdateClimb(r, summit)
	default:
		return methodNotAllowedError
	}
}

func (h *Api) HandleUpdateClimb(r *http.Request, summit *Summit) interface{} {
	userId := h.SM.GetInt64(r.Context(), UserIdKey)
	if userId == 0 { // not authenticated
		return authRequired
	}
	comment := r.PostFormValue("comment")
	date := r.PostFormValue("date")
	var ied InexactDate
	err := ied.Parse(date)
	if err != nil {
		// return validation error
	}
	err = UpdateClimb(h.DB, summit.Id, userId, ied, comment)
	return nil
}

func (h *Api) HandleSummits(r *http.Request) interface{} {
	if r.URL.Path != "/" {
		return pathNotFoundError
	}
	userId := h.SM.GetInt64(r.Context(), UserIdKey)
	summits, err := FetchSummits(h.DB, userId)
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

func (h *Api) HandleUser(r *http.Request) interface{} {
	if r.URL.Path == "/" {
		return pathNotFoundError
	}
	var userIdStr string
	userIdStr, r.URL.Path = ShiftPath(r.URL.Path)
	if r.URL.Path != "/" {
		return pathNotFoundError
	}

	var userId int64
	var err error
	if userIdStr == "me" {
		// return data for logged in user
		userId = h.SM.GetInt64(r.Context(), UserIdKey)
		if userId == 0 {
			// not authenticated
			return authRequired
		}
	} else {
		userId, err = strconv.ParseInt(userIdStr, 10, 64)
		if err != nil {
			return pathNotFoundError
		}
	}
	user, err := GetUserById(h.DB, userId)
	if err != nil {
		log.Printf("Failed to get user %d by ID: %v", userId, err)
		return serverError
	}
	if user == nil {
		log.Printf("Unknown user id %d", userId)
		return pathNotFoundError
	}
	return user
}

func (h *Api) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var head string
	head, r.URL.Path = ShiftPath(r.URL.Path)
	var resp interface{}

	switch head {
	case "summit":
		resp = h.HandleSummit(r)
	case "summits":
		if r.Method != http.MethodGet {
			resp = methodNotAllowedError
			break
		}
		resp = h.HandleSummits(r)
	case "top":
		if r.Method != http.MethodGet {
			resp = methodNotAllowedError
			break
		}
		resp = h.HandleTop(r)
	case "user":
		resp = h.HandleUser(r)
	default:
		resp = pathNotFoundError
	}

	jsonResp, err := json.Marshal(resp)
	if err != nil {
		log.Printf("Error marshalling response: %v", err)
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
