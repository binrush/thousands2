package main

import (
	"crypto/rand"
	"encoding/base64"
	"log/slog"
	"net/http"
	"net/url"

	"github.com/alexedwards/scs/v2"
	"github.com/go-chi/chi/v5"
	"golang.org/x/oauth2"
)

const (
	UserIdKey      = "UserId"
	OauthStateKey  = "OauthState"
	RedirectKey    = "Redirect"
	OauthStateSize = 16
)

func GenerateRandomString(size int) string {
	b := make([]byte, size)
	if _, err := rand.Read(b); err != nil {
		// this should never happened
		// or something wrong with OS's crypto pseudorandom generator
		panic(err)
	}
	return base64.RawURLEncoding.EncodeToString(b)
}

type AuthServer struct {
	Providers AuthProviders
	Storage   *Storage
	SM        *scs.SessionManager
	router    *chi.Mux
}

func NewAuthServer(providers AuthProviders, storage *Storage, sm *scs.SessionManager) *AuthServer {
	as := &AuthServer{
		Providers: providers,
		Storage:   storage,
		SM:        sm,
		router:    chi.NewRouter(),
	}

	// Set up routes
	as.router.Get("/oauth/{provider}", as.handleOAuthRedirect)
	as.router.Get("/authorized/{provider}", as.handleAuthorized)
	as.router.Get("/logout", as.handleLogout)

	return as
}

func (h *AuthServer) handleOAuthRedirect(w http.ResponseWriter, r *http.Request) {
	providerName := chi.URLParam(r, "provider")
	provider, ok := h.Providers[providerName]
	if !ok {
		http.NotFound(w, r)
		return
	}

	if h.SM.GetInt64(r.Context(), UserIdKey) != 0 { // user already logged in
		http.Redirect(w, r, "/user/me", http.StatusTemporaryRedirect)
		return
	}

	// Store the redirect URL from Referer header if available
	if referer := r.Header.Get("Referer"); referer != "" {
		// Parse the referer URL to get just the path and query
		if refererURL, err := url.Parse(referer); err == nil {
			redirectPath := refererURL.Path
			if refererURL.RawQuery != "" {
				redirectPath += "?" + refererURL.RawQuery
			}
			h.SM.Put(r.Context(), RedirectKey, redirectPath)
		}
	}

	oauthState := GenerateRandomString(OauthStateSize)
	h.SM.Put(r.Context(), OauthStateKey, oauthState)
	http.Redirect(
		w, r, provider.GetConfig().AuthCodeURL(oauthState, oauth2.AccessTypeOffline),
		http.StatusTemporaryRedirect)
}

func (h *AuthServer) handleAuthorized(w http.ResponseWriter, r *http.Request) {
	providerName := chi.URLParam(r, "provider")
	provider, ok := h.Providers[providerName]
	if !ok {
		http.NotFound(w, r)
		return
	}

	if r.FormValue("error") != "" {
		slog.Warn("Error returned by oauth provider",
			"error", r.FormValue("error"), "description", r.FormValue("error_description"))
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	expectedState, ok := h.SM.Pop(r.Context(), OauthStateKey).(string)
	if !ok {
		slog.Warn("OAuth state not found in session")
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}
	if expectedState != r.FormValue("state") {
		slog.Warn("Error checking state parameter: value not match")
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	code := r.FormValue("code")
	token, err := provider.GetConfig().Exchange(r.Context(), code)
	if err != nil {
		slog.Warn("Error obtaining oauth access token", "error", err)
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	oauthUserId, err := provider.GetUserId(token)
	if err != nil {
		slog.Warn("Failed to obtain oauth user ID", "error", err)
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	user, err := h.Storage.GetUser(oauthUserId, provider.GetSrcId())
	if err != nil {
		slog.Error("Failed to obtain user data from DB", "error", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	var userId int64
	if user != nil {
		// user exists log him in and redirect to profile page
		userId = user.Id
	} else {
		// user does not exist yet, register
		userId, err = provider.Register(token, h.Storage, r.Context())
		if err != nil {
			slog.Error("Failed to register user", "error", err)
			http.Error(w, "Invalid request", http.StatusBadRequest)
			return
		}
	}

	h.SM.Put(r.Context(), UserIdKey, userId)

	// Get the redirect URL from session, default to /user/me if not set
	redirectURL := "/user/me"
	if storedRedirect, ok := h.SM.Pop(r.Context(), RedirectKey).(string); ok && storedRedirect != "" {
		redirectURL = storedRedirect
	}

	http.Redirect(w, r, redirectURL, http.StatusTemporaryRedirect)
}

func (h *AuthServer) handleLogout(w http.ResponseWriter, r *http.Request) {
	err := h.SM.Destroy(r.Context())
	if err != nil {
		slog.Error("Failed to destroy session data", "error", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
}
