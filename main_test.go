package main

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/go-github/v71/github"
	"github.com/stretchr/testify/assert"
)

type MockGithubClient struct{}

func (m *MockGithubClient) DownloadAsset(url string, outputPath string) error {
	return nil
}

func (m *MockGithubClient) DownloadAssets(assets []*github.ReleaseAsset, releaseDir string) error {
	return nil
}

func TestPingRoute(t *testing.T) {
	router := setupRouter(RouterOptions{})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/ping", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "pong", w.Body.String())
}

func TestDeployWithGH_MissingContentType(t *testing.T) {
	router := setupRouter(RouterOptions{})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/deploy-with-gh", bytes.NewBuffer(nil))
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "Invalid payload")
}

func TestDeployWithGH_MissingEventType(t *testing.T) {
	router := setupRouter(RouterOptions{})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/deploy-with-gh", bytes.NewBuffer(nil))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "Unrecognized event type")
}

func TestDeployWithGH_UnrecognizedEventType(t *testing.T) {
	router := setupRouter(RouterOptions{})

	w := httptest.NewRecorder()
	payload := `{"action": "created"}`
	req, _ := http.NewRequest("POST", "/deploy-with-gh", bytes.NewBuffer(([]byte(payload))))
	req.Header.Set("X-GitHub-Event", "unknown_event")
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "Unrecognized event type")
}

func TestDeployWithGH_UnsupportedEventType(t *testing.T) {
	router := setupRouter(RouterOptions{})

	w := httptest.NewRecorder()
	payload := `{"action":"created"}`
	req, _ := http.NewRequest("POST", "/deploy-with-gh", bytes.NewBuffer(([]byte(payload))))
	req.Header.Set("X-GitHub-Event", "deployment")
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "Unsupported event type")
}

func TestDeployWithGH_ReleaseNotPublished(t *testing.T) {
	// Arrange: create a new router
	router := setupRouter(RouterOptions{})

	// Arrange: create a new HTTP request with the release event
	w := httptest.NewRecorder()
	payload := `{"action":"created"}`
	req, _ := http.NewRequest("POST", "/deploy-with-gh", bytes.NewBuffer(([]byte(payload))))
	req.Header.Set("X-GitHub-Event", "release")
	req.Header.Set("Content-Type", "application/json")

	// Act: send the request to the router
	router.ServeHTTP(w, req)

	// Assert: check if the response status code is 200 OK
	assert.Equal(t, http.StatusOK, w.Code)

	// Assert: check if the response body contains the expected message
	assert.Contains(t, w.Body.String(), `{"message":"Release is not published yet, ignoring."}`)
}

func TestDeployWithGH_Success(t *testing.T) {
	tempDir := t.TempDir()
	mockGithubClient := &MockGithubClient{}

	router := setupRouter(RouterOptions{
		AssetsDir:    tempDir,
		GithubClient: mockGithubClient,
	})

	w := httptest.NewRecorder()
	payload := `{"action":"published","release":{"assets":[{"url":"https://example.com/asset","name":"example-asset"}],"tag_name":"dev.0"},"repository":{"id":973821242,"name":"deploy-to-vm","owner":{"login":"cemreyavuz"}}}`
	req, _ := http.NewRequest("POST", "/deploy-with-gh", bytes.NewBuffer(([]byte(payload))))
	req.Header.Set("X-GitHub-Event", "release")
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), `{"action":"published"}`)
}
