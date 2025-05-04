package main

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"os"
	"path"
	"testing"

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
