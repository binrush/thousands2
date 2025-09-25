package main

import (
	"encoding/json"
	"fmt"
	"log/slog"
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
var authRequired = &ApiError{authRequiredMsg, http.StatusUnauthorized}

type Api struct {
	Config  *RuntimeConfig
	Storage *Storage
	SM      *scs.SessionManager
	router  *chi.Mux
}

func NewApi(config *RuntimeConfig, storage *Storage, sm *scs.SessionManager) *Api {
	api := &Api{
		Config:  config,
		Storage: storage,
		SM:      sm,
		router:  chi.NewRouter(),
	}

	// Set up routes
	api.router.Get("/summit/{ridgeId}/{summitId}", api.handleSummitGet)
	api.router.Put("/summit/{ridgeId}/{summitId}", api.handleSummitPut)
	api.router.Delete("/summit/{ridgeId}/{summitId}", api.handleSummitDelete)
	api.router.Get("/summit/{ridgeId}/{summitId}/climbs", api.handleSummitClimbs)
	api.router.Get("/summits", api.handleSummits)
	api.router.Get("/top", api.handleTop)
	api.router.Get("/user/me", api.handleUserMe)
	api.router.Get("/user/{userId}", api.handleUser)
	api.router.Get("/user/{userId}/climbs", api.handleUserClimbs)

	return api
}

func (h *Api) handleSummitGet(w http.ResponseWriter, r *http.Request) {
	ridgeId := chi.URLParam(r, "ridgeId")
	summitId := chi.URLParam(r, "summitId")

	userId := h.SM.GetInt64(r.Context(), UserIdKey)

	canonicalId, err := h.Storage.ResolveLegacyId(summitId)
	if err != nil {
		slog.Error("Failed to resolve legacy id", "summitId", summitId, "error", err)
		h.writeError(w, serverError)
		return
	}
	if canonicalId != "" {
		summitId = canonicalId
	}

	summit, err := h.Storage.FetchSummit(summitId, userId)
	if err != nil {
		slog.Error("Failed to fetch summit", "ridgeId", ridgeId, "summitId", summitId, "error", err)
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

	summit, err := h.Storage.FetchSummit(summitId, userId)
	if err != nil {
		slog.Error("Failed to fetch summit", "ridgeId", ridgeId, "summitId", summitId, "error", err)
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

	err = h.Storage.UpdateClimb(summit.Id, userId, ied, comment)
	if err != nil {
		slog.Error("Failed to update climb", "error", err)
		h.writeError(w, serverError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (h *Api) handleSummitDelete(w http.ResponseWriter, r *http.Request) {
	userId := h.SM.GetInt64(r.Context(), UserIdKey)
	if userId == 0 {
		h.writeError(w, authRequired)
		return
	}

	summitId := chi.URLParam(r, "summitId")
	err := h.Storage.DeleteClimb(summitId, userId)
	if err != nil {
		slog.Error("Failed to delete climb", "error", err)
		h.writeError(w, serverError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (h *Api) handleSummitClimbs(w http.ResponseWriter, r *http.Request) {
	ridgeId := chi.URLParam(r, "ridgeId")
	SummitId := chi.URLParam(r, "summitId")

	// Check if summit exists to correctly handle this case
	// (we do not want to return empty list for non-existing summit)
	summit, err := h.Storage.FetchSummit(SummitId, 0)
	if err != nil {
		slog.Error("Failed to find summit", "ridgeId", ridgeId, "summitId", SummitId, "error", err)
		h.writeError(w, serverError)
		return
	}
	if summit == nil {
		h.writeError(w, pathNotFoundError)
		return
	}

	page := 1
	pageParam := r.URL.Query()["page"]
	if len(pageParam) == 1 {
		if p, err := strconv.Atoi(pageParam[0]); err == nil && p > 0 {
			page = p
		}
	}

	climbs, totalClimbs, err := h.Storage.FetchSummitClimbs(summit.Id, page, h.Config.ItemsPerPage)
	if err != nil {
		slog.Error("Failed to fetch climbs for summit", "ridgeId", ridgeId, "summitId", SummitId, "realId", summit.Id, "error", err)
		h.writeError(w, serverError)
		return
	}

	response := struct {
		Climbs      []SummitClimb `json:"climbs"`
		TotalClimbs int           `json:"total_climbs"`
		Page        int           `json:"page"`
	}{
		Climbs:      climbs,
		TotalClimbs: totalClimbs,
		Page:        page,
	}

	h.writeJSON(w, response)
}

func (h *Api) handleSummits(w http.ResponseWriter, r *http.Request) {
	userId := h.SM.GetInt64(r.Context(), UserIdKey)
	summits, err := h.Storage.FetchSummits(userId)
	if err != nil {
		slog.Error("Failed to fetch summits from db", "error", err)
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

	top, err := h.Storage.FetchTop(page, h.Config.ItemsPerPage)
	if err != nil {
		slog.Error("Failed to fetch top", "error", err)
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

func (h *Api) handleUserById(w http.ResponseWriter, _ *http.Request, userId int64) {
	user, err := h.Storage.GetUserById(userId)
	if err != nil {
		slog.Error("Failed to get user by ID", "userId", userId, "error", err)
		h.writeError(w, serverError)
		return
	}
	if user == nil {
		slog.Warn("Unknown user id passed", "userId", userId)
		h.writeError(w, pathNotFoundError)
		return
	}
	h.writeJSON(w, user)
}

func (h *Api) handleUserClimbs(w http.ResponseWriter, r *http.Request) {
	userIdStr := chi.URLParam(r, "userId")
	userId, err := strconv.ParseInt(userIdStr, 10, 64)
	if err != nil {
		h.writeError(w, pathNotFoundError)
		return
	}

	climbs, err := h.Storage.FetchUserClimbs(userId)
	if err != nil {
		slog.Error("Failed to fetch climbs for user", "userId", userId, "error", err)
		h.writeError(w, serverError)
		return
	}
	h.writeJSON(w, climbs)
}

func (h *Api) writeJSON(w http.ResponseWriter, data interface{}) {
	jsonResp, err := json.Marshal(data)
	if err != nil {
		slog.Error("Error marshalling response", "error", err)
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
