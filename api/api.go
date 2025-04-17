package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/alexedwards/scs/v2"
	"github.com/go-chi/chi/v5"
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
	router *chi.Mux
}

func NewApi(config *RuntimeConfig, db *Database, sm *scs.SessionManager) *Api {
	api := &Api{
		Config: config,
		DB:     db,
		SM:     sm,
		router: chi.NewRouter(),
	}

	// Set up routes
	api.router.Get("/summit/{ridgeId}/{summitId}", api.handleSummitGet)
	api.router.Put("/summit/{ridgeId}/{summitId}", api.handleSummitPut)
	api.router.Get("/summits", api.handleSummits)
	api.router.Get("/top", api.handleTop)
	api.router.Get("/user/me", api.handleUserMe)
	api.router.Get("/user/{userId}", api.handleUser)

	return api
}

func (h *Api) handleSummitGet(w http.ResponseWriter, r *http.Request) {
	ridgeId := chi.URLParam(r, "ridgeId")
	summitId := chi.URLParam(r, "summitId")

	page := 1
	pageParam := r.URL.Query()["page"]
	if len(pageParam) == 1 {
		if p, err := strconv.Atoi(pageParam[0]); err == nil && p > 0 {
			page = p
		}
	}

	summit, err := FetchSummit(h.DB, summitId, page, h.Config.ItemsPerPage)
	if err != nil {
		log.Printf("Failed to fetch summit %s/%s: %v", ridgeId, summitId, err)
		h.writeError(w, serverError)
		return
	}
	if summit == nil {
		h.writeError(w, pathNotFoundError)
		return
	}

	h.writeJSON(w, summit)
}

func (h *Api) handleSummitPut(w http.ResponseWriter, r *http.Request) {
	userId := h.SM.GetInt64(r.Context(), UserIdKey)
	if userId == 0 {
		h.writeError(w, authRequired)
		return
	}

	ridgeId := chi.URLParam(r, "ridgeId")
	summitId := chi.URLParam(r, "summitId")

	summit, err := FetchSummit(h.DB, summitId, 1, h.Config.ItemsPerPage)
	if err != nil {
		log.Printf("Failed to fetch summit %s/%s: %v", ridgeId, summitId, err)
		h.writeError(w, serverError)
		return
	}
	if summit == nil {
		h.writeError(w, pathNotFoundError)
		return
	}

	comment := r.PostFormValue("comment")
	date := r.PostFormValue("date")
	var ied InexactDate
	err = ied.Parse(date)
	if err != nil {
		h.writeError(w, &ApiError{"Invalid date format", http.StatusBadRequest})
		return
	}

	err = UpdateClimb(h.DB, summit.Id, userId, ied, comment)
	if err != nil {
		log.Printf("Failed to update climb: %v", err)
		h.writeError(w, serverError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (h *Api) handleSummits(w http.ResponseWriter, r *http.Request) {
	userId := h.SM.GetInt64(r.Context(), UserIdKey)
	summits, err := FetchSummits(h.DB, userId)
	if err != nil {
		log.Printf("Failed to fetch summits from db: %v", err)
		h.writeError(w, serverError)
		return
	}
	h.writeJSON(w, summits)
}

func (h *Api) handleTop(w http.ResponseWriter, r *http.Request) {
	page := 1
	pageParam := r.URL.Query()["page"]
	if pageParam != nil {
		if len(pageParam) != 1 {
			h.writeError(w, &ApiError{"Invalid page parameter provided", http.StatusBadRequest})
			return
		}
		var err error
		page, err = strconv.Atoi(pageParam[0])
		if err != nil || page <= 0 {
			h.writeError(w, &ApiError{"Invalid page parameter provided", http.StatusBadRequest})
			return
		}
	}

	top, err := FetchTop(h.DB, page, h.Config.ItemsPerPage)
	if err != nil {
		log.Printf("Failed to fetch top: %v", err)
		h.writeError(w, serverError)
		return
	}
	h.writeJSON(w, top)
}

func (h *Api) handleUserMe(w http.ResponseWriter, r *http.Request) {
	userId := h.SM.GetInt64(r.Context(), UserIdKey)
	if userId == 0 {
		h.writeError(w, authRequired)
		return
	}
	h.handleUserById(w, r, userId)
}

func (h *Api) handleUser(w http.ResponseWriter, r *http.Request) {
	userIdStr := chi.URLParam(r, "userId")
	userId, err := strconv.ParseInt(userIdStr, 10, 64)
	if err != nil {
		h.writeError(w, pathNotFoundError)
		return
	}
	h.handleUserById(w, r, userId)
}

func (h *Api) handleUserById(w http.ResponseWriter, r *http.Request, userId int64) {
	user, err := GetUserById(h.DB, userId)
	if err != nil {
		log.Printf("Failed to get user %d by ID: %v", userId, err)
		h.writeError(w, serverError)
		return
	}
	if user == nil {
		log.Printf("Unknown user id %d", userId)
		h.writeError(w, pathNotFoundError)
		return
	}
	h.writeJSON(w, user)
}

func (h *Api) writeJSON(w http.ResponseWriter, data interface{}) {
	jsonResp, err := json.Marshal(data)
	if err != nil {
		log.Printf("Error marshalling response: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	fmt.Fprint(w, string(jsonResp))
}

func (h *Api) writeError(w http.ResponseWriter, err *ApiError) {
	jsonResp, _ := json.Marshal(err)
	w.Header().Set("Content-Type", "application/json")
	http.Error(w, string(jsonResp), err.StatusCode)
}
