package main

import (
	"context"
	"fmt"
	"net/http"
	"net/http/cookiejar"
	"net/http/httptest"
	"net/url"
	"testing"

	"golang.org/x/oauth2"
)

const (
	MockClientId     = "mock_client_id"
	MockClientSecret = "mock_client_secret"
	MockAccessToken  = "mock_access_token"
	MockOauthUserId  = "2343"
)

type MockOauthHandler interface {
	HandleAccessToken(w http.ResponseWriter, r *http.Request)
	HandleAuthorize(w http.ResponseWriter, r *http.Request)
}

type MockOauthSuccessfulHandler struct {
	Code        string
	AccessToken string
	UserId      string
}

func (mos *MockOauthSuccessfulHandler) HandleAccessToken(w http.ResponseWriter, r *http.Request) {
	errorResponseTpl := "{\"error\":\"%s\",\"error_description\":\"Oauth Server Error\"}"
	accessTokenResponse := `{
		"token_type": "Bearer",
		"access_token": "%s",
		"expires_in": 43200,
		"user_id": "%s"
	}`
	w.Header().Set("Content-Type", "application/json")
	errorResponse := ""
	if !((r.FormValue("client_id") == MockClientId) && (r.FormValue("client_secret") == MockClientSecret)) {
		errorResponse = fmt.Sprintf(errorResponseTpl, "invalid_client")
	} else if r.FormValue("code") != mos.Code {
		errorResponse = fmt.Sprintf(errorResponseTpl, "invalid_grant")
	}

	if errorResponse != "" {
		http.Error(w, errorResponse, http.StatusUnauthorized)
	} else {
		fmt.Fprintf(w, accessTokenResponse, mos.AccessToken, mos.UserId)
	}
}

func (mos *MockOauthSuccessfulHandler) HandleAuthorize(w http.ResponseWriter, r *http.Request) {
	// TODO: check response_type, client_id, access_type params
	//redirectUrl := r.FormValue("redirect_uri")
	redirectUrl, err := url.Parse(r.FormValue("redirect_uri"))
	if err != nil {
		http.Error(w, "Failed to parse redirect url", http.StatusBadRequest)
		return
	}
	if r.FormValue("client_id") != MockClientId {
		http.Error(w, "Client ID does not match", http.StatusBadRequest)
		return
	}
	query := redirectUrl.Query()
	state := r.FormValue("state")
	query.Add("state", state)
	query.Add("code", mos.Code)
	redirectUrl.RawQuery = query.Encode()
	http.Redirect(w, r, redirectUrl.String(), http.StatusTemporaryRedirect)
}

type MockOauthIncorrectStateHandler struct {
	MockOauthSuccessfulHandler
}

func (mos *MockOauthIncorrectStateHandler) HandleAuthorize(w http.ResponseWriter, r *http.Request) {
	redirectUrl, err := url.Parse(r.FormValue("redirect_uri"))
	if err != nil {
		http.Error(w, "Failed to parse redirect url", http.StatusBadRequest)
		return
	}
	query := redirectUrl.Query()
	query.Add("state", "incorrect")
	query.Add("code", mos.Code)
	redirectUrl.RawQuery = query.Encode()
	http.Redirect(w, r, redirectUrl.String(), http.StatusTemporaryRedirect)
}

// Handler that returns authorize error
type MockOauthAuthorizeErrorHandler struct {
	MockOauthSuccessfulHandler
}

func (mos *MockOauthAuthorizeErrorHandler) HandleAuthorize(w http.ResponseWriter, r *http.Request) {
	redirectUrl, err := url.Parse(r.FormValue("redirect_uri"))
	if err != nil {
		http.Error(w, "Failed to parse redirect url", http.StatusBadRequest)
		return
	}
	query := redirectUrl.Query()
	query.Add("error", "invalid_request")
	query.Add("error_description", "Invalid Client Id")
	redirectUrl.RawQuery = query.Encode()
	http.Redirect(w, r, redirectUrl.String(), http.StatusTemporaryRedirect)
}

// Handler that returns error to access token request
type MockOauthTokenErrorHandler struct {
	MockOauthSuccessfulHandler
}

func (mos *MockOauthTokenErrorHandler) HandleAccessToken(w http.ResponseWriter, r *http.Request) {
	errorResponseTpl := "{\"error\":\"%s\",\"error_description\":\"Oauth Server Error\"}"
	w.Header().Set("Content-Type", "application/json")
	errorResponse := fmt.Sprintf(errorResponseTpl, "invalid_client")
	http.Error(w, errorResponse, http.StatusUnauthorized)
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
	if token.AccessToken != MockAccessToken {
		return "", fmt.Errorf("Unable to get user id: incorrect access token")
	}
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

func NewMockOauthServer(mockOauthHandler MockOauthHandler) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/access_token":
			mockOauthHandler.HandleAccessToken(w, r)
		case "/authorize":
			mockOauthHandler.HandleAuthorize(w, r)
		default:
			http.NotFound(w, r)
		}
	}))
}

func NewApp(mockOauthProvider provider, t *testing.T) *App {
	providers := make(AuthProviders)
	providers["mock"] = mockOauthProvider
	as := AuthServer{Providers: providers, DB: MockDatabase(t)}
	sm := NewSessionManager()
	return &App{
		AuthServer: &as,
		SM:         &sm,
	}
}

func NewTestClient(t *testing.T) *http.Client {
	jar, err := cookiejar.New(nil)
	if err != nil {
		t.Fatalf("Failed to create cookie jar: %v", err)
	}
	return &http.Client{Jar: jar}
}

func TestAuthFlow(t *testing.T) {
	var cases = []struct {
		oauthHandler       MockOauthHandler
		expectedUrlPath    string
		expectedStatusCode int
	}{
		{
			&MockOauthSuccessfulHandler{
				GenerateRandomString(32),
				MockAccessToken,
				MockOauthUserId,
			},
			"/profile",
			404,
		},
		{
			&MockOauthAuthorizeErrorHandler{},
			"/auth/authorized/mock",
			400,
		},
		{
			&MockOauthIncorrectStateHandler{},
			"/auth/authorized/mock",
			400,
		},
		{
			&MockOauthTokenErrorHandler{},
			"/auth/authorized/mock",
			400,
		},
	}

	for _, tt := range cases {
		mockOauthServer := NewMockOauthServer(tt.oauthHandler)
		defer mockOauthServer.Close()

		mockProvider := NewMockProvider(mockOauthServer.URL, MockClientId, MockClientSecret)
		app := NewApp(mockProvider, t)
		appServer := httptest.NewServer(app)
		defer appServer.Close()

		// update RedirectUrl in auth provider
		mockProvider.Config.RedirectURL = appServer.URL + "/auth/authorized/mock"

		client := NewTestClient(t)
		resp, err := client.Get(appServer.URL + "/auth/oauth/mock")
		if err != nil {
			t.Fatal(err)
		}
		defer resp.Body.Close()

		// Check that we were finally redirected to expected endpoint
		if resp.Request.URL.Path != tt.expectedUrlPath {
			t.Errorf("Unexpected url path: %s, expected %s", resp.Request.URL.Path, tt.expectedUrlPath)
		}
		if resp.StatusCode != tt.expectedStatusCode {
			t.Errorf("Unexpected status code: %d, expected %d", resp.StatusCode, tt.expectedStatusCode)
		}
		// TODO: call corresponding endpoint to check if account is created
	}
}
