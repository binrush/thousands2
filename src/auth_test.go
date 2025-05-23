package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/cookiejar"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/alexedwards/scs/v2"
	"github.com/go-chi/chi/v5"
	"golang.org/x/oauth2"
)

const (
	MockClientId     = "mock_client_id"
	MockClientSecret = "mock_client_secret"
	MockAccessToken  = "mock_access_token"
	MockOauthUserId  = "2343"
	MockUserId       = int64(5)
)

type MockSessionStore struct {
	userId int64
}

func NewMockSessionStore(userId int64) *MockSessionStore {
	return &MockSessionStore{userId: userId}
}

func (mss *MockSessionStore) Delete(token string) error {
	return nil
}

func (mss *MockSessionStore) Find(token string) (b []byte, found bool, err error) {
	if token != "mock_session_token" {
		return nil, false, nil
	}
	mockData := make(map[string]interface{})
	mockData[UserIdKey] = mss.userId
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
func (mp *MockProviderSuccess) Register(token *oauth2.Token, storage *Storage, ctx context.Context) (int64, error) {
	userId, err := storage.CreateUser("Mock Mock", "2343", mp.GetSrcId())
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

func (*MockProviderRegisterError) Register(token *oauth2.Token, storage *Storage, ctx context.Context) (int64, error) {
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
	db := MockDatabase(t)
	storage := NewStorage(db)
	providers := make(AuthProviders)
	providers["mock"] = mockOauthProvider
	sm := scs.New()
	conf := &RuntimeConfig{Datadir: "testdata/summits"}
	api := NewApi(conf, storage, sm)
	as := NewAuthServer(providers, storage, sm)
	app := &App{
		Api:        api,
		AuthServer: as,
		SM:         sm,
		router:     chi.NewRouter(),
	}
	app.router.Use(sm.LoadAndSave)
	app.router.Mount("/api", api.router)
	app.router.Mount("/auth", as.router)
	return app
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
			"/user/me",
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
		appServer := httptest.NewServer(app.router)
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
		profileEndpoint := "/api/user/me"
		resp, err = client.Get(appServer.URL + profileEndpoint)
		if err != nil {
			t.Fatal(err)
		}
		defer resp.Body.Close()

		if tt.expectedUrlPath == "/user/me" {
			// registration successful
			if resp.StatusCode != http.StatusOK {
				t.Errorf("Unexpected status code %d from %s, expected %d",
					resp.StatusCode, profileEndpoint, http.StatusOK)
			}
			var user User
			err = json.NewDecoder(resp.Body).Decode(&user)
			if err != nil {
				t.Errorf("Failed to decode response: %v", err)
			}
			if user.Name != "Mock Mock" {
				t.Errorf("Unexpected user name for new user: %v", user.Name)
			}
			if (user.Src != 1) || (user.OauthId != "2343") {
				t.Errorf("Unexpected oauth data of new user: Src=%v, oauthId=%v", user.Src, user.OauthId)
			}
		} else {
			// registration error
			if resp.StatusCode != http.StatusUnauthorized {
				t.Errorf("Unexpected status code %d from %s, expected %d",
					resp.StatusCode, profileEndpoint, http.StatusUnauthorized)
			}
			var apiError ApiError
			err = json.NewDecoder(resp.Body).Decode(&apiError)
			if err != nil {
				t.Errorf("Failed to decode response: %v", err)
			}
			if apiError.Message != "Authentication required" {
				t.Errorf("Unexpected error message returned: %v", apiError.Message)
			}
		}
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
	appServer := httptest.NewServer(app.router)
	defer appServer.Close()

	oauthProvider.SetConfig(
		NewMockOauthConfig(mockOauthServer.URL, appServer.URL+"/auth/authorized/mock"))

	_, err := app.AuthServer.Storage.CreateUser("Mock Mock", MockOauthUserId, 1)
	if err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	client := NewTestClient(t)
	resp, err := client.Get(appServer.URL + "/auth/oauth/mock")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	// Check that we were finally redirected to expected endpoint
	expectedUrlPath := "/user/me"
	if resp.Request.URL.Path != expectedUrlPath {
		t.Fatalf("Unexpected url path: %s, expected %s", resp.Request.URL.Path, expectedUrlPath)
	}
	requestHistory := Redirections(resp)
	if len(requestHistory) != 4 {
		t.Fatalf("Expected full oauth flow (4 steps), got: %v", requestHistory)
	}

	// Check that logged user is redirected directly to /user/me
	resp, err = client.Get(appServer.URL + "/auth/oauth/mock")
	if err != nil {
		t.Fatal(err)
	}
	if resp.Request.URL.Path != expectedUrlPath {
		t.Fatalf("Unexpected url path: %s, expected %s", resp.Request.URL.Path, expectedUrlPath)
	}
	requestHistory = Redirections(resp)
	if len(requestHistory) != 2 {
		t.Fatalf("Expected redirect to /user/me (2 steps), got: %v", requestHistory)
	}

	// check that profile is returned for authrorized user
	profileUrl := "/api/user/me"
	resp, err = client.Get(appServer.URL + profileUrl)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("Unexpected status code %d from %s, expected %d",
			resp.StatusCode, profileUrl, http.StatusUnauthorized)
	}

	// Check logout
	resp, err = client.Get(appServer.URL + "/auth/logout")
	if err != nil {
		t.Fatal(err)
	}
	expectedUrlPath = "/"
	if resp.Request.URL.Path != expectedUrlPath {
		t.Fatalf("Unexpected url path: %s, expected %s", resp.Request.URL.Path, expectedUrlPath)
	}
	resp, err = client.Get(appServer.URL + profileUrl)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("Unexpected status code %d from %s, expected %d",
			resp.StatusCode, profileUrl, http.StatusUnauthorized)
	}
}
