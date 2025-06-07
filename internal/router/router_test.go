package router

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"deploy-to-vm/internal/config"
	"encoding/hex"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path"
	"testing"

	deploy_to_vm_github "deploy-to-vm/internal/github"

	"github.com/gin-gonic/gin"
	"github.com/google/go-github/v71/github"
	"github.com/stretchr/testify/assert"
)

type MockGithubClient struct {
	DownloadAssetFunc  func(url string, outputPath string) error
	DownloadAssetsFunc func(assets []*github.ReleaseAsset, releaseDir string) (error, deploy_to_vm_github.DownloadAssetStatusCode)
}

func (m *MockGithubClient) DownloadAsset(url string, outputPath string) error {
	if m.DownloadAssetFunc != nil {
		return m.DownloadAssetFunc(url, outputPath)
	}

	return nil
}

func (m *MockGithubClient) DownloadAssets(assets []*github.ReleaseAsset, releaseDir string) (error, deploy_to_vm_github.DownloadAssetStatusCode) {
	if m.DownloadAssetsFunc != nil {
		return m.DownloadAssetsFunc(assets, releaseDir)
	}

	return nil, deploy_to_vm_github.DownloadAsset_Success
}

type MockNginxClient struct {
	ReloadFunc func() error
}

func (m *MockNginxClient) Reload() error {
	if m.ReloadFunc != nil {
		return m.ReloadFunc()
	}

	return nil
}

type MockNotificationClient struct {
	NotifyFunc func(message string) error
}

func (m *MockNotificationClient) LoadWebhookUrl() error {
	return nil
}

func (m *MockNotificationClient) Notify(message string) error {
	if m.NotifyFunc != nil {
		return m.NotifyFunc(message)
	}

	return nil
}

func setupTestRouter() *gin.Engine {
	return SetupRouter(RouterOptions{
		ConfigClient: &config.ConfigClient{},
	})
}

func TestDeployWithGH_MissingContentType(t *testing.T) {
	router := setupTestRouter()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/deploy-with-gh", bytes.NewBuffer(nil))
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "Invalid payload")
}

func TestDeployWithGH_MissingEventType(t *testing.T) {
	router := setupTestRouter()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/deploy-with-gh", bytes.NewBuffer(nil))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "Unrecognized event type")
}

func TestDeployWithGH_UnrecognizedEventType(t *testing.T) {
	router := setupTestRouter()

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
	router := setupTestRouter()

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
	router := setupTestRouter()

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

	configClient := &config.ConfigClient{}
	configClient.Config = &config.DeployToVmConfig{
		Repositories: []config.DeployToVmConfigRepository{
			{
				Name:       "deploy-to-vm",
				Owner:      "cemreyavuz",
				SourceType: "github",
				TargetDir:  t.TempDir(),
				TargetType: "nginx",
			},
		},
	}

	router := SetupRouter(RouterOptions{
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

	configClient := &config.ConfigClient{}
	configClient.Config = &config.DeployToVmConfig{
		Repositories: []config.DeployToVmConfigRepository{
			{
				Name:       "deploy-to-vm",
				Owner:      "cemreyavuz",
				SourceType: "github",
				TargetDir:  t.TempDir(),
				TargetType: "nginx",
			},
		},
	}

	router := SetupRouter(RouterOptions{
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

func TestDeployWithGH_NoAssetsFound(t *testing.T) {
	tempDir := t.TempDir()
	mockGithubClient := &MockGithubClient{
		DownloadAssetsFunc: func(assets []*github.ReleaseAsset, releaseDir string) (error, deploy_to_vm_github.DownloadAssetStatusCode) {
			return errors.New("mock error"), deploy_to_vm_github.DownloadAsset_NoAssetsFound
		},
	}

	configClient := &config.ConfigClient{}
	configClient.Config = &config.DeployToVmConfig{
		Repositories: []config.DeployToVmConfigRepository{
			{
				Name:       "deploy-to-vm",
				Owner:      "cemreyavuz",
				SourceType: "github",
				TargetDir:  t.TempDir(),
				TargetType: "nginx",
			},
		},
	}

	router := SetupRouter(RouterOptions{
		AssetsDir:    tempDir,
		ConfigClient: configClient,
		GithubClient: mockGithubClient,
	})

	w := httptest.NewRecorder()
	payload := `{"action":"released","release":{"assets":[],"tag_name":"dev.0"},"repository":{"id":973821242,"name":"deploy-to-vm","owner":{"login":"cemreyavuz"}}}`
	req, _ := http.NewRequest("POST", "/deploy-with-gh", bytes.NewBuffer(([]byte(payload))))
	req.Header.Set("X-GitHub-Event", "release")
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), `{"message":"No assets found for release \"dev.0\", will skip the request."}`)
}

func TestDeployWithGH_DownloadAssets_Error(t *testing.T) {
	tempDir := t.TempDir()
	mockGithubClient := &MockGithubClient{
		DownloadAssetsFunc: func(assets []*github.ReleaseAsset, releaseDir string) (error, deploy_to_vm_github.DownloadAssetStatusCode) {
			return fmt.Errorf("Failed to download assets"), deploy_to_vm_github.DownloadAsset_UnknownError
		},
	}

	router := SetupRouter(RouterOptions{
		AssetsDir:    tempDir,
		ConfigClient: &config.ConfigClient{},
		GithubClient: mockGithubClient,
	})

	w := httptest.NewRecorder()
	payload := `{"action":"released","release":{"assets":[{"url":"https://example.com/asset","name":"example-asset"}],"tag_name":"dev.0"},"repository":{"id":973821242,"name":"deploy-to-vm","owner":{"login":"cemreyavuz"}}}`
	req, _ := http.NewRequest("POST", "/deploy-with-gh", bytes.NewBuffer(([]byte(payload))))
	req.Header.Set("X-GitHub-Event", "release")
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.Contains(t, w.Body.String(), "Failed to download assets")
}

func TestDeployWithGH_Untar_Error(t *testing.T) {
	tempDir := t.TempDir()
	mockGithubClient := &MockGithubClient{
		DownloadAssetsFunc: func(assets []*github.ReleaseAsset, releaseDir string) (error, deploy_to_vm_github.DownloadAssetStatusCode) {
			corruptedTarFilePath := path.Join(releaseDir, "corrupted.tar.gz")
			os.WriteFile(corruptedTarFilePath, []byte("dummy content"), 0644)
			return nil, deploy_to_vm_github.DownloadAsset_Success
		},
	}

	configClient := &config.ConfigClient{}
	configClient.Config = &config.DeployToVmConfig{
		Repositories: []config.DeployToVmConfigRepository{
			{
				Name:       "deploy-to-vm",
				Owner:      "cemreyavuz",
				SourceType: "github",
				TargetDir:  t.TempDir(),
				TargetType: "nginx",
			},
		},
	}

	router := SetupRouter(RouterOptions{
		AssetsDir:    tempDir,
		ConfigClient: configClient,
		GithubClient: mockGithubClient,
	})

	w := httptest.NewRecorder()
	payload := `{"action":"released","release":{"assets":[{"url":"https://example.com/asset","name":"example-asset"}],"tag_name":"dev.0"},"repository":{"id":973821242,"name":"deploy-to-vm","owner":{"login":"cemreyavuz"}}}`
	req, _ := http.NewRequest("POST", "/deploy-with-gh", bytes.NewBuffer(([]byte(payload))))
	req.Header.Set("X-GitHub-Event", "release")
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.Contains(t, w.Body.String(), "Failed to untar files in release directory")
}

func TestDeployWithGH_GetRepository_Error(t *testing.T) {
	tempDir := t.TempDir()
	mockGithubClient := &MockGithubClient{}

	configClient := &config.ConfigClient{}
	configClient.Config = &config.DeployToVmConfig{
		Repositories: []config.DeployToVmConfigRepository{
			{
				Name:       "non-existent-repo",
				Owner:      "non-existent-owner",
				SourceType: "github",
				TargetDir:  t.TempDir(),
				TargetType: "nginx",
			},
		},
	}

	router := SetupRouter(RouterOptions{
		AssetsDir:    tempDir,
		ConfigClient: configClient,
		GithubClient: mockGithubClient,
	})

	w := httptest.NewRecorder()
	payload := `{"action":"released","release":{"assets":[{"url":"https://example.com/asset","name":"example-asset"}],"tag_name":"dev.0"},"repository":{"id":973821242,"name":"deploy-to-vm","owner":{"login":"cemreyavuz"}}}`
	req, _ := http.NewRequest("POST", "/deploy-with-gh", bytes.NewBuffer(([]byte(payload))))
	req.Header.Set("X-GitHub-Event", "release")
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.Contains(t, w.Body.String(), "Repository not found in config")
}

func TestDeployWithGH_MissingTargetDir(t *testing.T) {
	tempDir := t.TempDir()
	mockGithubClient := &MockGithubClient{}

	configClient := &config.ConfigClient{}
	configClient.Config = &config.DeployToVmConfig{
		Repositories: []config.DeployToVmConfigRepository{
			{
				Name:       "deploy-to-vm",
				Owner:      "cemreyavuz",
				SourceType: "github",
				TargetDir:  "", // Missing target directory
				TargetType: "nginx",
			},
		},
	}

	router := SetupRouter(RouterOptions{
		AssetsDir:    tempDir,
		ConfigClient: configClient,
		GithubClient: mockGithubClient,
	})

	w := httptest.NewRecorder()
	payload := `{"action":"released","release":{"assets":[{"url":"https://example.com/asset","name":"example-asset"}],"tag_name":"dev.0"},"repository":{"id":973821242,"name":"deploy-to-vm","owner":{"login":"cemreyavuz"}}}`
	req, _ := http.NewRequest("POST", "/deploy-with-gh", bytes.NewBuffer(([]byte(payload))))
	req.Header.Set("X-GitHub-Event", "release")
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.Contains(t, w.Body.String(), "Site directory not found for repository")
}

func TestDeployWithGH_NonExistentSiteDir(t *testing.T) {
	tempDir := t.TempDir()
	mockGithubClient := &MockGithubClient{}

	configClient := &config.ConfigClient{}
	configClient.Config = &config.DeployToVmConfig{
		Repositories: []config.DeployToVmConfigRepository{
			{
				Name:       "deploy-to-vm",
				Owner:      "cemreyavuz",
				SourceType: "github",
				TargetDir:  "/non/existent/dir",
				TargetType: "nginx",
			},
		},
	}

	router := SetupRouter(RouterOptions{
		AssetsDir:    tempDir,
		ConfigClient: configClient,
		GithubClient: mockGithubClient,
	})

	w := httptest.NewRecorder()
	payload := `{"action":"released","release":{"assets":[{"url":"https://example.com/asset","name":"example-asset"}],"tag_name":"dev.0"},"repository":{"id":973821242,"name":"deploy-to-vm","owner":{"login":"cemreyavuz"}}}`
	req, _ := http.NewRequest("POST", "/deploy-with-gh", bytes.NewBuffer(([]byte(payload))))
	req.Header.Set("X-GitHub-Event", "release")
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.Contains(t, w.Body.String(), "Failed to move release assets to site directory")
}

func TestDeployWithGH_Reload_Error(t *testing.T) {
	tempDir := t.TempDir()
	mockGithubClient := &MockGithubClient{}
	mockNginxClient := &MockNginxClient{
		ReloadFunc: func() error {
			return fmt.Errorf("Failed to reload nginx")
		},
	}

	configClient := &config.ConfigClient{}
	configClient.Config = &config.DeployToVmConfig{
		Repositories: []config.DeployToVmConfigRepository{
			{
				Name:       "deploy-to-vm",
				Owner:      "cemreyavuz",
				SourceType: "github",
				TargetDir:  t.TempDir(),
				TargetType: "nginx",
			},
		},
	}

	router := SetupRouter(RouterOptions{
		AssetsDir:    tempDir,
		ConfigClient: configClient,
		GithubClient: mockGithubClient,
		NginxClient:  mockNginxClient,
	})

	w := httptest.NewRecorder()
	payload := `{"action":"released","release":{"assets":[{"url":"https://example.com/asset","name":"example-asset"}],"tag_name":"dev.0"},"repository":{"id":973821242,"name":"deploy-to-vm","owner":{"login":"cemreyavuz"}}}`
	req, _ := http.NewRequest("POST", "/deploy-with-gh", bytes.NewBuffer(([]byte(payload))))
	req.Header.Set("X-GitHub-Event", "release")
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.Contains(t, w.Body.String(), "Failed to reload nginx unit")
}

func TestDeployWithGH_Notify_Error(t *testing.T) {
	tempDir := t.TempDir()
	mockGithubClient := &MockGithubClient{}
	mockNginxClient := &MockNginxClient{}
	mockNotificationClient := &MockNotificationClient{
		NotifyFunc: func(message string) error {
			return fmt.Errorf("Failed to send notification")
		},
	}

	configClient := &config.ConfigClient{}
	configClient.Config = &config.DeployToVmConfig{
		Repositories: []config.DeployToVmConfigRepository{
			{
				Name:       "deploy-to-vm",
				Owner:      "cemreyavuz",
				SourceType: "github",
				TargetDir:  t.TempDir(),
				TargetType: "nginx",
			},
		},
	}

	router := SetupRouter(RouterOptions{
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
}

func TestPingRoute(t *testing.T) {
	router := setupTestRouter()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/ping", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "pong", w.Body.String())
}
