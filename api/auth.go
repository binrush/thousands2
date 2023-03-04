package main

import (
	"crypto/rand"
	"encoding/base64"
	"log"
	"net/http"

	"github.com/alexedwards/scs/v2"
	"golang.org/x/oauth2"
)

const (
	UserIdKey      = "UserId"
	OauthStateKey  = "OauthState"
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
	DB        *Database
	SM        *scs.SessionManager
}

func (h *AuthServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var head string
	head, r.URL.Path = ShiftPath(r.URL.Path)
	switch head {
	case "oauth":
		h.RedirectToProvider(w, r)
	case "authorized":
		h.Authorized(w, r)
	default:
		http.NotFound(w, r)
	}
}

func (h *AuthServer) Authorized(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path == "/" {
		http.NotFound(w, r)
		return
	}
	var head string
	head, r.URL.Path = ShiftPath(r.URL.Path)

	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	provider, ok := h.Providers[head]
	if !ok {
		http.NotFound(w, r)
		return
	}
	if r.FormValue("error") != "" {
		log.Printf("Error returned by oauth provider: %s, %s",
			r.FormValue("error"), r.FormValue("error_description"))
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}
	expectedState := h.SM.Pop(r.Context(), OauthStateKey).(string)
	if expectedState != r.FormValue("state") {
		log.Printf("Error checking state parameter: value not match")
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}
	code := r.FormValue("code")
	token, err := provider.GetConfig().Exchange(r.Context(), code)
	if err != nil {
		log.Printf("Error obtaining oauth access token: %s", err)
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}
	oauthUserId, err := provider.GetUserId(token)
	if err != nil {
		log.Printf("Failed to obtain oauth user ID: %s", err)
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}
	user, err := GetUser(h.DB, oauthUserId, provider.GetSrcId())
	if err != nil {
		log.Printf("Failed to obtain user data from DB: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	var userId int64
	if user != nil {
		// user exists log him in and redirect to profile page
		userId = user.Id
	} else {
		// user does not exist yet, register
		userId, err = provider.Register(token, h.DB, r.Context())
		if err != nil {
			log.Printf("Failed to register user: %v", err)
			http.Error(w, "Invalid request", http.StatusBadRequest)
			return
		}
	}
	h.SM.Put(r.Context(), UserIdKey, userId)
	http.Redirect(w, r, "/profile", http.StatusTemporaryRedirect)
}

func (h *AuthServer) RedirectToProvider(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path == "/" {
		http.NotFound(w, r)
		return
	}
	var head string
	head, r.URL.Path = ShiftPath(r.URL.Path)

	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}

	provider, ok := h.Providers[head]
	if !ok {
		http.NotFound(w, r)
		return
	}
	if h.SM.GetInt64(r.Context(), UserIdKey) != 0 { // user already logged in
		http.Redirect(w, r, "/profile", http.StatusTemporaryRedirect)
		return
	}

	oauthState := GenerateRandomString(OauthStateSize)
	h.SM.Put(r.Context(), OauthStateKey, oauthState)
	http.Redirect(
		w, r, provider.GetConfig().AuthCodeURL(oauthState, oauth2.AccessTypeOffline),
		http.StatusTemporaryRedirect)
}
