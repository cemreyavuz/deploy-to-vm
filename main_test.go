package main

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPingRoute(t *testing.T) {
	router := setupRouter()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/ping", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "pong", w.Body.String())
}

func TestDeployWithGH_MissingContentType(t *testing.T) {
	router := setupRouter()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/deploy-with-gh", bytes.NewBuffer(nil))
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "Invalid payload")
}

func TestDeployWithGH_MissingEventType(t *testing.T) {
	router := setupRouter()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/deploy-with-gh", bytes.NewBuffer(nil))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "Unrecognized event type")
}

func TestDeployWithGH_UnrecognizedEventType(t *testing.T) {
	router := setupRouter()

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
	router := setupRouter()

	w := httptest.NewRecorder()
	payload := `{"action":"created"}`
	req, _ := http.NewRequest("POST", "/deploy-with-gh", bytes.NewBuffer(([]byte(payload))))
	req.Header.Set("X-GitHub-Event", "deployment")
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "Unsupported event type")
}

func TestDeployWithGH_Success(t *testing.T) {
	router := setupRouter()

	w := httptest.NewRecorder()
	payload := `{"action":"created","release":{"assets":[{"browser_download_url":"https://example.com/asset"}]},"repository":{"id":973821242,"name":"deploy-to-vm","owner":{"login":"cemreyavuz"}}}`
	req, _ := http.NewRequest("POST", "/deploy-with-gh", bytes.NewBuffer(([]byte(payload))))
	req.Header.Set("X-GitHub-Event", "release")
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), `{"action":"created"}`)
}
