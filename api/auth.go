package main

import (
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/vk"
	"net/http"
	"os"
	"crypto/rand"
	"encoding/base64"
	"fmt"
)

const (
	SessionCookieName = "TSID"
)

type AuthProviders map[string]*oauth2.Config

type Session struct {
	UserId      int
	OauthState  string
	RedirectUrl string
}

type SessionManager struct {
	Data map[string]*Session
}

func NewSessionManager() SessionManager {
	return SessionManager{
		Data: make(map[string]*Session),
	}
}

func generateSessionId() string {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		// this should never happened
		// or something wrong with OS's crypto pseudorandom generator
		panic(err)
	}
	return base64.RawURLEncoding.EncodeToString(b)
}

func (sm *SessionManager) StartSession(w http.ResponseWriter, r *http.Request) *Session {
	sessionIdCookie, err := r.Cookie(SessionCookieName)
	if err == nil {
		sessionData, ok := sm.Data[sessionIdCookie.Value]
		if ok {
			return sessionData
		}
	}
	sessionId := generateSessionId()
	sm.Data[sessionId] = &Session{}
	sessCookie := http.Cookie{
		Name:     SessionCookieName,
		Value:    sessionId,
		HttpOnly: true,
	}
	http.SetCookie(w, &sessCookie)
	return sm.Data[sessionId]
}

type AuthServer struct {
	Providers AuthProviders
	DB        *Database
}

func (h *AuthServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var head string
	head, r.URL.Path = ShiftPath(r.URL.Path)
	switch head {
	case "oauth":
		h.RedirectToProvider(w, r)
	case "authorized":
		fmt.Fprint(w, "Coming soon")
	default:
		http.NotFound(w, r)
	}
}

func (h *AuthServer) RedirectToProvider(w http.ResponseWriter, r *http.Request) {
	var head string
	head, r.URL.Path = ShiftPath(r.URL.Path)
	provider, ok := h.Providers[head]
	if !ok {
		http.NotFound(w, r)
	}
	http.Redirect(
		w, r, provider.AuthCodeURL("FIXME", oauth2.AccessTypeOffline),
		http.StatusTemporaryRedirect)
}

func GetAuthProviders(authorizedUrl string) AuthProviders {
	providers := make(map[string]*oauth2.Config)
	providers["vk"] = &oauth2.Config{
		RedirectURL:  authorizedUrl,
		ClientID:     os.Getenv("VK_CLIENT_ID"),
		ClientSecret: os.Getenv("VK_CLIENT_SECRET"),
		Endpoint:     vk.Endpoint,
	}
	return providers
}
