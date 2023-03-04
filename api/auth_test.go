package main

import (
	"context"
	"fmt"
	"net/http"
	"net/http/cookiejar"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/alexedwards/scs/v2"
	"golang.org/x/oauth2"
)

const (
	MockClientId     = "mock_client_id"
	MockClientSecret = "mock_client_secret"
	MockAccessToken  = "mock_access_token"
	MockOauthUserId  = "2343"
)

type MockSessionStore struct {
}

func (mss *MockSessionStore) Delete(token string) error {
	return nil
}

func (mss *MockSessionStore) Find(token string) (b []byte, found bool, err error) {
	if token != "mock_session_token" {
		return nil, false, nil
	}
	mockData := make(map[string]interface{})
	mockData[UserIdKey] = int64(5)
	codec := scs.GobCodec{}
	res, err := codec.Encode(time.Now().Add(24*time.Hour), mockData)
	if err != nil {
		panic("Failed to encode mock data")
	}
	return res, true, nil
}

func (mss *MockSessionStore) Commit(token string, b []byte, expiry time.Time) error {
	return nil
}

func Redirections(resp *http.Response) []string {
	history := []string{}
	for resp != nil {
		req := resp.Request
		history = append(history, req.URL.String())
		resp = req.Response
	}
	return history
}

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

type MockProvider interface {
	Provider
	SetConfig(*oauth2.Config)
}

type MockProviderSuccess struct {
	Config *oauth2.Config
}

// GetSrcId implements provider
func (*MockProviderSuccess) GetSrcId() int {
	return 1
}

// GetUserId implements provider
func (*MockProviderSuccess) GetUserId(token *oauth2.Token) (string, error) {
	if token.AccessToken != MockAccessToken {
		return "", fmt.Errorf("Unable to get user id: incorrect access token")
	}
	return "2343", nil
}

// Register implements provider
func (mp *MockProviderSuccess) Register(token *oauth2.Token, db *Database, ctx context.Context) (int64, error) {
	userId, err := CreateUser(db, "Mock Mock", "2343", mp.GetSrcId())
	if err != nil {
		return 0, err
	}
	return userId, nil
}

func (mp *MockProviderSuccess) GetConfig() *oauth2.Config {
	return mp.Config
}

func (mp *MockProviderSuccess) SetConfig(config *oauth2.Config) {
	mp.Config = config
}

type MockProviderUserIdError struct {
	MockProviderSuccess
}

func (*MockProviderUserIdError) GetUserId(token *oauth2.Token) (string, error) {
	return "", fmt.Errorf("Failed to get user ID")
}

type MockProviderRegisterError struct {
	MockProviderSuccess
}

func (*MockProviderRegisterError) Register(token *oauth2.Token, db *Database, ctx context.Context) (int64, error) {
	return 0, fmt.Errorf("Failed to register user")
}

func NewMockOauthConfig(serverUrl, redirectURL string) *oauth2.Config {
	return &oauth2.Config{
		RedirectURL:  redirectURL,
		ClientID:     MockClientId,
		ClientSecret: MockClientSecret,
		Endpoint: oauth2.Endpoint{
			AuthURL:  serverUrl + "/authorize",
			TokenURL: serverUrl + "/access_token",
		},
	}
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

func NewApp(mockOauthProvider Provider, t *testing.T) *App {
	providers := make(AuthProviders)
	providers["mock"] = mockOauthProvider
	sm := scs.New()
	as := AuthServer{Providers: providers, DB: MockDatabase(t), SM: sm}
	return &App{
		AuthServer: &as,
		SM:         sm,
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
		oauthProvider      MockProvider
		expectedUrlPath    string
		expectedStatusCode int
	}{
		{
			&MockOauthSuccessfulHandler{
				GenerateRandomString(32),
				MockAccessToken,
				MockOauthUserId,
			},
			&MockProviderSuccess{},
			"/profile",
			404,
		},
		{
			&MockOauthAuthorizeErrorHandler{},
			&MockProviderSuccess{},
			"/auth/authorized/mock",
			400,
		},
		{
			&MockOauthIncorrectStateHandler{},
			&MockProviderSuccess{},
			"/auth/authorized/mock",
			400,
		},
		{
			&MockOauthTokenErrorHandler{},
			&MockProviderSuccess{},
			"/auth/authorized/mock",
			400,
		},
		{
			&MockOauthSuccessfulHandler{
				GenerateRandomString(32),
				MockAccessToken,
				MockOauthUserId,
			},
			&MockProviderUserIdError{},
			"/auth/authorized/mock",
			400,
		},
		{
			&MockOauthSuccessfulHandler{
				GenerateRandomString(32),
				MockAccessToken,
				MockOauthUserId,
			},
			&MockProviderRegisterError{},
			"/auth/authorized/mock",
			400,
		},
	}

	for _, tt := range cases {
		mockOauthServer := NewMockOauthServer(tt.oauthHandler)
		defer mockOauthServer.Close()

		app := NewApp(tt.oauthProvider, t)
		appServer := httptest.NewServer(app.SM.LoadAndSave(app))
		defer appServer.Close()

		tt.oauthProvider.SetConfig(
			NewMockOauthConfig(mockOauthServer.URL, appServer.URL+"/auth/authorized/mock"))

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

func TestAuthFlowUserExists(t *testing.T) {
	oauthHandler := &MockOauthSuccessfulHandler{
		GenerateRandomString(32),
		MockAccessToken,
		MockOauthUserId,
	}
	mockOauthServer := NewMockOauthServer(oauthHandler)
	defer mockOauthServer.Close()

	// to ensure register is not called
	oauthProvider := &MockProviderRegisterError{}
	app := NewApp(oauthProvider, t)
	appServer := httptest.NewServer(app.SM.LoadAndSave(app))
	defer appServer.Close()

	oauthProvider.SetConfig(
		NewMockOauthConfig(mockOauthServer.URL, appServer.URL+"/auth/authorized/mock"))

	_, err := CreateUser(app.AuthServer.DB, "Mock Mock", MockOauthUserId, 1)

	client := NewTestClient(t)
	resp, err := client.Get(appServer.URL + "/auth/oauth/mock")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	// Check that we were finally redirected to expected endpoint
	expectedUrlPath := "/profile"
	if resp.Request.URL.Path != expectedUrlPath {
		t.Fatalf("Unexpected url path: %s, expected %s", resp.Request.URL.Path, expectedUrlPath)
	}
	expectedStatusCode := http.StatusNotFound
	if resp.StatusCode != expectedStatusCode {
		t.Fatalf("Unexpected status code: %d, expected %d", resp.StatusCode, expectedStatusCode)
	}
	requestHistory := Redirections(resp)
	if len(requestHistory) != 4 {
		t.Fatalf("Expected full oauth flow (4 steps), got: %v", requestHistory)
	}

	// Check that logged user is redirected directly to /profile
	client.Get(appServer.URL + "/auth/oauth/mock")
	if resp.Request.URL.Path != expectedUrlPath {
		t.Fatalf("Unexpected url path: %s, expected %s", resp.Request.URL.Path, expectedUrlPath)
	}
	if resp.StatusCode != expectedStatusCode {
		t.Fatalf("Unexpected status code: %d, expected %d", resp.StatusCode, expectedStatusCode)
	}
	requestHistory = Redirections(resp)
	if len(requestHistory) != 4 {
		t.Fatalf("Expected redirect to /profile (2 steps), got: %v", requestHistory)
	}
}
