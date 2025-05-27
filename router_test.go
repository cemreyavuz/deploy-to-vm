package main

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
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

type MockNginxClient struct{}

func (m *MockNginxClient) Reload() error {
	return nil
}

type MockNotificationClient struct{}

func (m *MockNotificationClient) LoadWebhookUrl() error {
	return nil
}

func (m *MockNotificationClient) Notify(message string) error {
	return nil
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

func TestDeployWithGH_ReleaseNotReleased(t *testing.T) {
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
	assert.Contains(t, w.Body.String(), `{"message":"Only \"released\" action is supported, ignoring..."}`)
}

func TestDeployWithGH_WithSignature_Success(t *testing.T) {
	tempDir := t.TempDir()
	mockGithubClient := &MockGithubClient{}
	mockNginxClient := &MockNginxClient{}
	mockNotificationClient := &MockNotificationClient{}

	configClient := &ConfigClient{}
	configClient.Config = &DeployToVmConfig{
		Repositories: []DeployToVmConfigRepository{
			{
				Name:       "deploy-to-vm",
				Owner:      "cemreyavuz",
				SourceType: "github",
				TargetDir:  t.TempDir(),
				TargetType: "nginx",
			},
		},
	}

	router := setupRouter(RouterOptions{
		AssetsDir:          tempDir,
		ConfigClient:       configClient,
		GithubClient:       mockGithubClient,
		NginxClient:        mockNginxClient,
		NotificationClient: mockNotificationClient,
		SecretToken:        "test",
	})

	w := httptest.NewRecorder()
	payload := `{"action":"released","release":{"assets":[{"url":"https://example.com/asset","name":"example-asset"}],"tag_name":"dev.0"},"repository":{"id":973821242,"name":"deploy-to-vm","owner":{"login":"cemreyavuz"}}}`

	mac := hmac.New(sha256.New, []byte("test"))
	mac.Write([]byte(payload))
	signature := "sha256=" + hex.EncodeToString(mac.Sum(nil))

	req, _ := http.NewRequest("POST", "/deploy-with-gh", bytes.NewBuffer(([]byte(payload))))
	req.Header.Set("X-GitHub-Event", "release")
	req.Header.Set("X-Hub-Signature-256", signature)
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), `{"action":"released"}`)
}

func TestDeployWithGH_WithoutSignature_Success(t *testing.T) {
	tempDir := t.TempDir()
	mockGithubClient := &MockGithubClient{}
	mockNginxClient := &MockNginxClient{}
	mockNotificationClient := &MockNotificationClient{}

	configClient := &ConfigClient{}
	configClient.Config = &DeployToVmConfig{
		Repositories: []DeployToVmConfigRepository{
			{
				Name:       "deploy-to-vm",
				Owner:      "cemreyavuz",
				SourceType: "github",
				TargetDir:  t.TempDir(),
				TargetType: "nginx",
			},
		},
	}

	router := setupRouter(RouterOptions{
		AssetsDir:          tempDir,
		ConfigClient:       configClient,
		GithubClient:       mockGithubClient,
		NginxClient:        mockNginxClient,
		NotificationClient: mockNotificationClient,
	})

	w := httptest.NewRecorder()
	payload := `{"action":"released","release":{"assets":[{"url":"https://example.com/asset","name":"example-asset"}],"tag_name":"dev.0"},"repository":{"id":973821242,"name":"deploy-to-vm","owner":{"login":"cemreyavuz"}}}`
	req, _ := http.NewRequest("POST", "/deploy-with-gh", bytes.NewBuffer(([]byte(payload))))
	req.Header.Set("X-GitHub-Event", "release")
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), `{"action":"released"}`)
}
