package http_utils

import (
	"bytes"
	"net/http"
)

// Make post request to the given URL with the provided data
func MakePostRequest(url string, data []byte) (*http.Response, error) {
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(data))
	return resp, err
}
