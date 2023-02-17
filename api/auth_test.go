package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"net/http/cookiejar"
	"net/http/httptest"
	"net/url"
	"testing"

	"golang.org/x/oauth2"
)

type MockOauthHandler struct {
	Code        string
	AccessToken string
	UserId      string
}

func (mos *MockOauthHandler) HandleAccessToken(w http.ResponseWriter, r *http.Request) {
	errorResponse := "{\"error\":\"invalid_request\",\"error_description\":\"Oauth Server Error\"}"
	accessTokenResponse := `{
		"access_token": "%s",
		"expires_in": 43200,
		"user_id": "%s"
	}`
	w.Header().Set("Content-Type", "application/json")
	if r.FormValue("code") != mos.Code {
		http.Error(
			w,
			errorResponse,
			http.StatusUnauthorized)
		return
	}
	fmt.Fprintf(w, accessTokenResponse, mos.AccessToken, mos.UserId)
}

func (mos *MockOauthHandler) HandleAuthorize(w http.ResponseWriter, r *http.Request) {
	// TODO: check response_type, client_id, access_type params
	//redirectUrl := r.FormValue("redirect_uri")
	redirectUrl, err := url.Parse(r.FormValue("redirect_uri"))
	if err != nil {
		http.Error(w, "Failed to parse redirect url", http.StatusInternalServerError)
		return
	}
	state := r.FormValue("state")
	query := redirectUrl.Query()
	query.Add("state", state)
	query.Add("code", mos.Code)
	redirectUrl.RawQuery = query.Encode()
	http.Redirect(w, r, redirectUrl.String(), http.StatusTemporaryRedirect)
}

func (mos *MockOauthHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.URL.Path {
	case "/access_token":
		mos.HandleAccessToken(w, r)
	case "/authorize":
		mos.HandleAuthorize(w, r)
	default:
		http.NotFound(w, r)
	}
}

type MockProvider struct {
	Config *oauth2.Config
}

// GetSrcId implements provider
func (*MockProvider) GetSrcId() int {
	return 1
}

// GetUserId implements provider
func (*MockProvider) GetUserId(token *oauth2.Token) (string, error) {
	return "2343", nil
}

// Register implements provider
func (mp *MockProvider) Register(token *oauth2.Token, db *Database, ctx context.Context) (int64, error) {
	userId, err := CreateUser(db, "Mock Mock", "2343", mp.GetSrcId())
	if err != nil {
		return 0, err
	}
	return userId, nil
}

func (mp *MockProvider) GetConfig() *oauth2.Config {
	return mp.Config
}

func NewMockProvider(serverUrl, client_id, client_secret string) *MockProvider {
	var mockProvider MockProvider
	mockProvider.Config = &oauth2.Config{
		// RedirectURL is updated when test server is started
		RedirectURL:  "",
		ClientID:     client_id,
		ClientSecret: client_secret,
		Endpoint: oauth2.Endpoint{
			AuthURL:  serverUrl + "/authorize",
			TokenURL: serverUrl + "/access_token",
		},
	}
	return &mockProvider

}

/*
func GetMockAuthProviders(baseUrl, serverUrl string) AuthProviders {
	providers := make(AuthProviders)
	providers["mock"] = NewMockProvider(baseUrl, serverUrl, "mock_client_id", "mock_client_secret")
	return providers
}
*/

func TestAuthFlow(t *testing.T) {
	var moh MockOauthHandler
	moh.AccessToken = GenerateRandomString(32)
	moh.Code = GenerateRandomString(32)
	mos := httptest.NewServer(&moh)
	defer mos.Close()
	log.Printf("Oauth server url: %v", mos.URL)

	mockProvider := NewMockProvider(mos.URL, "mock_client_id", "mock_client_secret")
	providers := make(AuthProviders)
	providers["mock"] = mockProvider
	as := AuthServer{Providers: providers, DB: MockDatabase(t)}
	sm := NewSessionManager()
	app := &App{
		AuthServer: &as,
		SM:         &sm,
	}
	appServer := httptest.NewServer(app)
	defer appServer.Close()
	// update RedirectUrl in auth provider
	mockProvider.Config.RedirectURL = appServer.URL + "/auth/authorized/mock"
	log.Printf("App server url: %v", appServer.URL)

	jar, err := cookiejar.New(nil)
	if err != nil {
		t.Fatalf("Failed to create cookie jar: %v", err)
	}
	client := &http.Client{Jar: jar}

	resp, err := client.Get(appServer.URL + "/auth/oauth/mock")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	// Check that we were finally redirected to /profile endpoint
	if resp.Request.URL.Path != "/profile" {
		t.Fatalf("Registration not finished, not redirected to /profile."+
			" Last url: %s, status code: %d", resp.Request.URL.String(), resp.StatusCode)
		// TODO: call corresponding endpoint to check if account is created
	}
}

/*
package main

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func mockOauthServer(code string) *httptest.Server {
	errorResponse := "{\"error\":\"invalid_request\",\"error_description\":\"Oauth Server Error\"}"
	accessToken := GenerateRandomString(32)
	mockUserId := 2343
	accessTokenResponse := `{
		"access_token": "%s",
		"expires_in": 43200,
		"user_id": %d
	}`
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.URL.Path == "/access_token" {
			if r.FormValue("code") != code {
				http.Error(
					w,
					errorResponse,
					http.StatusUnauthorized)
				return
			}
			fmt.Fprintf(w, accessTokenResponse, accessToken, mockUserId)
			return
		}
	}))

}

func MockAuthProviders(url string) AuthProviders {
	providers := make(AuthProviders)
	providers["vk"] = &oauth2.Config{
		RedirectURL:  "https://thousands.su/auth/authorized/vk",
		ClientID:     "MOCK_VK_CLIENT_ID",
		ClientSecret: "MOCK_VK_CLIENT_SECRET",
		Scopes:       []string{},
		Endpoint: oauth2.Endpoint{
			AuthURL:  url + "/authorize",
			TokenURL: url + "/access_token",
		},
	}
	return providers
}
*/
