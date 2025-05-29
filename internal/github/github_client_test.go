package github

import (
	"bytes"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path"
	"testing"

	"github.com/google/go-github/v71/github"
	"github.com/stretchr/testify/assert"
)

func setupGithubClientTest(t *testing.T) (string, string) {
	// create a temporary directory for testing
	tempDir := t.TempDir()

	var accessToken = "github-client-test-access-token"

	return accessToken, tempDir
}

// MockHttpClient is a mock implementation of the HttpClient interface for
// testing purposes. It allows you to define custom behavior for the Do method.
// This is useful for simulating different HTTP responses without making actual
// network calls.
type MockHttpClient struct {
	DoFunc func(req *http.Request) (*http.Response, error)
}

// Do is the mock implementation of the HttpClient's Do method. It calls the
// custom behavior defined in the DoFunc field. If DoFunc is nil, it returns
// nil for both response and error.
func (mock *MockHttpClient) Do(req *http.Request) (*http.Response, error) {
	if mock.DoFunc != nil {
		return mock.DoFunc(req)
	}
	return nil, nil
}

func TestDownloadAsset_Success(t *testing.T) {
	// arrange: get test helpers
	accessToken, tempDir := setupGithubClientTest(t)

	// arrange: create a mock HTTP client
	w := httptest.NewRecorder()
	var testFileContent = "download-asset-test-file-content"
	w.Body = bytes.NewBufferString(testFileContent)
	mockHttpClient := &MockHttpClient{
		DoFunc: func(req *http.Request) (*http.Response, error) {
			assert.Equal(t, "GET", req.Method)
			assert.Equal(t, "Bearer "+accessToken, req.Header.Get("Authorization"))
			assert.Equal(t, "application/octet-stream", req.Header.Get("Accept"))
			return &http.Response{
				StatusCode: http.StatusOK,
				Body:       w.Result().Body,
			}, nil
		},
	}

	// arrange: create a Github client with the mock HTTP client
	client := &GithubClient{
		HttpClient:  mockHttpClient,
		AccessToken: accessToken,
	}

	// arrange: define the test file path
	var testFilePath = path.Join(tempDir, "output.txt")

	// act: download the asset
	downloadErr := client.DownloadAsset("https://example.com/asset", testFilePath)

	// assert: check if the file was created
	assert.NoError(t, downloadErr, "Expected no error")

	// assert: check if the file content is as expected
	fileContent, readErr := os.ReadFile(testFilePath)
	if readErr != nil {
		t.Fatalf("Expected no error reading file, got %v", readErr)
	}
	assert.Equal(t, testFileContent, string(fileContent), "Expected file content to match")
}

func TestDownloadAssets_NoAssetsFound(t *testing.T) {
	// arrange: get test helpers
	accessToken, tempDir := setupGithubClientTest(t)

	// arrange: create a Github client
	client := &GithubClient{
		HttpClient:  &MockHttpClient{},
		AccessToken: accessToken,
	}

	// arrange: define an empty assets slice
	var testAssets []*github.ReleaseAsset

	// act: download the assets
	downloadErr, code := client.DownloadAssets(testAssets, tempDir)

	// assert: check if the error is as expected
	assert.Error(t, downloadErr, "Expected an error when no assets are found")
	assert.Equal(t, DownloadAsset_NoAssetsFound, code, "Expected no assets found error code")
}

func TestDownloadAssets_Single(t *testing.T) {
	// arrange: get test helpers
	accessToken, tempDir := setupGithubClientTest(t)

	// arrange: create a mock HTTP client
	w := httptest.NewRecorder()
	var testFileContent = "download-assets-test-file-content"
	w.Body = bytes.NewBufferString(testFileContent)
	mockHttpClient := &MockHttpClient{
		DoFunc: func(req *http.Request) (*http.Response, error) {
			assert.Equal(t, "GET", req.Method)
			assert.Equal(t, "Bearer "+accessToken, req.Header.Get("Authorization"))
			assert.Equal(t, "application/octet-stream", req.Header.Get("Accept"))
			return &http.Response{
				StatusCode: http.StatusOK,
				Body:       w.Result().Body,
			}, nil
		},
	}

	// arrange: create a Github client with the mock HTTP client
	client := &GithubClient{
		HttpClient:  mockHttpClient,
		AccessToken: accessToken,
	}

	// arrange: define the test file path
	var testAssets [1]*github.ReleaseAsset
	testAssets[0] = &github.ReleaseAsset{
		Name: github.Ptr("test-asset.txt"),
		URL:  github.Ptr("https://example.com/test-asset.txt"),
	}

	// act: download the asset
	downloadErr, code := client.DownloadAssets(testAssets[:], tempDir)

	// assert: check if the file was create
	assert.NoError(t, downloadErr, "Expected no error")
	assert.Equal(t, DownloadAsset_Success, code, "Expected no error code")

	// assert: check if the file content is as expected
	testAssetPath := path.Join(tempDir, "test-asset.txt")
	testAssetContent, testAssetReadErr := os.ReadFile(testAssetPath)
	if testAssetReadErr != nil {
		t.Fatalf("Expected no error reading file, got %v", testAssetReadErr)
	}
	assert.Equal(t, testFileContent, string(testAssetContent), "Expected file content to match")
}

func TestDownloadAssets_Multiple(t *testing.T) {
	// arrange: get test helpers
	accessToken, tempDir := setupGithubClientTest(t)

	// arrange: create a mock HTTP client
	w := httptest.NewRecorder()
	var testFileContent = "download-assets-test-file-content"
	w.Body = bytes.NewBufferString(testFileContent)
	mockHttpClient := &MockHttpClient{
		DoFunc: func(req *http.Request) (*http.Response, error) {
			assert.Equal(t, "GET", req.Method)
			assert.Equal(t, "Bearer "+accessToken, req.Header.Get("Authorization"))
			assert.Equal(t, "application/octet-stream", req.Header.Get("Accept"))
			return &http.Response{
				StatusCode: http.StatusOK,
				Body:       w.Result().Body,
			}, nil
		},
	}

	// arrange: create a Github client with the mock HTTP client
	client := &GithubClient{
		HttpClient:  mockHttpClient,
		AccessToken: accessToken,
	}

	// arrange: define the test file path
	var testAssets [2]*github.ReleaseAsset
	testAssets[0] = &github.ReleaseAsset{
		Name: github.Ptr("test-asset0.txt"),
		URL:  github.Ptr("https://example.com/test-asset0.txt"),
	}
	testAssets[1] = &github.ReleaseAsset{
		Name: github.Ptr("test-asset1.txt"),
		URL:  github.Ptr("https://example.com/test-asset1.txt"),
	}

	// act: download the asset
	downloadErr, code := client.DownloadAssets(testAssets[:], tempDir)

	// assert: check if the file was created
	assert.NoError(t, downloadErr, "Expected no error")
	assert.Equal(t, DownloadAsset_Success, code, "Expected no error code")

	// assert: check if the file content is as expected
	testAsset1Path := path.Join(tempDir, "test-asset0.txt")
	testAsset1Content, testAsset1ReadErr := os.ReadFile(testAsset1Path)
	if testAsset1ReadErr != nil {
		t.Fatalf("Expected no error reading file, got %v", testAsset1ReadErr)
	}
	assert.Equal(t, testFileContent, string(testAsset1Content), "Expected file content to match")

	testAsset0Path := path.Join(tempDir, "test-asset0.txt")
	testAsset0Content, testAsset0ReadErr := os.ReadFile(testAsset0Path)
	if testAsset1ReadErr != nil {
		t.Fatalf("Expected no error reading file, got %v", testAsset0ReadErr)
	}
	assert.Equal(t, testFileContent, string(testAsset0Content), "Expected file content to match")
}

func TestDownloadAssets_Error(t *testing.T) {
	// Arrange: get test helpers
	accessToken, _ := setupGithubClientTest(t)

	// Arrange: create a mock HTTP client that returns an error
	mockHttpClient := &MockHttpClient{
		DoFunc: func(req *http.Request) (*http.Response, error) {
			return nil, assert.AnError
		},
	}

	// Arrange: create a Github client with the mock HTTP client
	client := &GithubClient{
		HttpClient:  mockHttpClient,
		AccessToken: accessToken,
	}

	// Arrange: define the test assets
	testAssets := []*github.ReleaseAsset{
		{
			Name: github.Ptr("test-asset.txt"),
			URL:  github.Ptr("https://example.com/test-asset.txt"),
		},
	}

	// Act: attempt to download the assets
	err, code := client.DownloadAssets(testAssets, "/invalid/path")

	// Assert: check if the error is as expected
	assert.Error(t, err, "Expected an error when downloading assets")
	assert.Contains(t, err.Error(), "Error downloading asset:", "Expected error message to match")
	assert.Equal(t, DownloadAsset_UnknownError, code, "Expected error code to match")
}

func TestDownloadAsset_NewRequest_Error(t *testing.T) {
	// Arrange: get test helpers
	accessToken, _ := setupGithubClientTest(t)

	// Arrange: create a mock HTTP client that returns an error
	mockHttpClient := &MockHttpClient{
		DoFunc: func(req *http.Request) (*http.Response, error) {
			return nil, assert.AnError
		},
	}

	// Arrange: create a Github client with the mock HTTP client
	client := &GithubClient{
		HttpClient:  mockHttpClient,
		AccessToken: accessToken,
	}

	// Act: attempt to download an asset with an invalid URL
	err := client.DownloadAsset(":://", "output.txt")

	// Assert: check if the error is as expected
	assert.Error(t, err, "Expected an error when creating request")
	assert.Contains(t, err.Error(), "Error creating request", "Expected error message to match")
}

func TestDownloadAsset_HttpClientDo_Error(t *testing.T) {
	// Arrange: get test helpers
	accessToken, _ := setupGithubClientTest(t)

	// Arrange: create a mock HTTP client that returns an error
	mockHttpClient := &MockHttpClient{
		DoFunc: func(req *http.Request) (*http.Response, error) {
			return nil, assert.AnError
		},
	}

	// Arrange: create a Github client with the mock HTTP client
	client := &GithubClient{
		HttpClient:  mockHttpClient,
		AccessToken: accessToken,
	}

	// Act: attempt to download an asset
	err := client.DownloadAsset("https://example.com/asset", "output.txt")

	// Assert: check if the error is as expected
	assert.Error(t, err, "Expected an error when downloading asset")
	assert.Equal(t, "Error downloading asset: "+assert.AnError.Error(), err.Error(), "Expected error message to match")
}

func TestDownloadAsset_ResponseStatusNotOK(t *testing.T) {
	// Arrange: get test helpers
	accessToken, _ := setupGithubClientTest(t)

	// Arrange: create a mock HTTP client that returns a non-OK status code
	mockHttpClient := &MockHttpClient{
		DoFunc: func(req *http.Request) (*http.Response, error) {
			return &http.Response{
				StatusCode: http.StatusNotFound,
				Body:       http.NoBody,
			}, nil
		},
	}

	// Arrange: create a Github client with the mock HTTP client
	client := &GithubClient{
		HttpClient:  mockHttpClient,
		AccessToken: accessToken,
	}

	// Act: attempt to download an asset
	err := client.DownloadAsset("https://example.com/asset", "output.txt")

	// Assert: check if the error is as expected
	assert.Error(t, err, "Expected an error when response status is not OK")
	assert.Equal(t, "Error downloading asset, status code: 404", err.Error(), "Expected error message to match")
}

func TestDownloadAsset_CreateFile_Error(t *testing.T) {
	// Arrange: get test helpers
	accessToken, _ := setupGithubClientTest(t)

	// Arrange: create a mock HTTP client that returns a successful response
	mockHttpClient := &MockHttpClient{
		DoFunc: func(req *http.Request) (*http.Response, error) {
			return &http.Response{
				StatusCode: http.StatusOK,
				Body:       http.NoBody,
			}, nil
		},
	}

	// Arrange: create a Github client with the mock HTTP client
	client := &GithubClient{
		HttpClient:  mockHttpClient,
		AccessToken: accessToken,
	}

	// Act: attempt to download an asset to a directory that cannot be created
	err := client.DownloadAsset("https://example.com/asset", "/invalid/path/output.txt")

	// Assert: check if the error is as expected
	assert.Error(t, err, "Expected an error when creating file")
	assert.Contains(t, err.Error(), "Error creating output file:", "Expected error message to match")
}

type errorReader struct{}

func (e *errorReader) Read(p []byte) (int, error) {
	return 0, errors.New("simulated read error")
}

func TestDownloadAssest_WriteFile_Error(t *testing.T) {
	// Arrange: get test helpers
	accessToken, tempDir := setupGithubClientTest(t)

	// Arrange: create a mock HTTP client that returns a successful response
	mockHttpClient := &MockHttpClient{
		DoFunc: func(req *http.Request) (*http.Response, error) {
			return &http.Response{
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(&errorReader{}),
			}, nil
		},
	}

	// Arrange: create a Github client with the mock HTTP client
	client := &GithubClient{
		HttpClient:  mockHttpClient,
		AccessToken: accessToken,
	}

	// Act: attempt to download an asset
	err := client.DownloadAsset("https://example.com/asset", path.Join(tempDir, "output.txt"))

	// Assert: check if the error is as expected
	assert.Error(t, err, "Expected an error when writing to file")
	assert.Contains(t, err.Error(), "Error writing to output file:", "Expected error message to match")
}
