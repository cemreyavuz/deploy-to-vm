package main

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMakePostRequest_Error(t *testing.T) {
	// Act: make a POST request to an invalid URL
	_, err := makePostRequest("http://invalid-url-that-doesnt-exist.example", []byte(`{}`))

	// Assert: check if an error is returned
	assert.Error(t, err, "Expected an error for invalid URL, but got none")
}

func TestMakePostRequest_Success(t *testing.T) {
	// Arrange: setup test server to simulate receiving our POST request
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check if request method is POST
		assert.Equal(t, http.MethodPost, r.Method, "Expected POST request")

		// Check content type header
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"), "Expected Content-Type: application/json")

		// Read request body
		body, _ := io.ReadAll(r.Body)
		expectedBody := `{"test":"data"}`
		assert.Equal(t, expectedBody, string(body), "Expected request body to match")

		// Set content type to application/json
		w.Header().Set("Content-Type", "application/json")

		// Set response status code to 200 OK
		w.WriteHeader(http.StatusOK)

		// Write response body
		w.Write([]byte(`{"status":"success"}`))
	}))
	defer server.Close()

	// Act: call the function being tested
	resp, err := makePostRequest(server.URL, []byte(`{"test":"data"}`))

	// Assert: check if request was successful
	assert.NoError(t, err, "Expected no error, but got one")

	// Assert: check response status code
	assert.Equal(t, http.StatusOK, resp.StatusCode, "Expected status code 200 OK")

	// Assert: check response content type
	assert.Equal(t, "application/json", resp.Header.Get("Content-Type"), "Expected Content-Type: application/json")

	// Assert: read and verify response body
	responseBody, _ := io.ReadAll(resp.Body)
	expected := `{"status":"success"}`
	assert.Equal(t, expected, string(responseBody), "Expected response body to match")
}
