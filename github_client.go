package main

import (
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
)

// HttpClient is an interface that defines the Do method for making HTTP
// requests. This allows for easier testing and mocking of HTTP requests in
// unit tests. The interface can be implemented by any struct that has a Do
// method with the same signature.
type HttpClient interface {
	Do(req *http.Request) (*http.Response, error)
}

// GithubClient is a struct that represents a client for interacting with the
// GitHub API. It contains an access token for authentication and an HTTP client
// for making requests.
type GithubClient struct {
	AccessToken string
	HttpClient  HttpClient
}

// GithubClientInterface is an interface that defines the methods for the
// GithubClient struct. This allows for easier testing and mocking of the
// GithubClient in unit tests. The interface can be implemented by any struct
// that has the same methods as the GithubClient struct.
type GithubClientInterface interface {
	DownloadAsset(url string, outputPath string) error
}

// DownloadAsset is a method of the GithubClient struct that downloads an asset
// from a given URL and saves it to a specified output path.
func (c *GithubClient) DownloadAsset(url string, outputPath string) error {
	// create a new HTTP request
	req, createRequestErr := http.NewRequest("GET", url, nil)
	if createRequestErr != nil {
		return errors.New("Error creating request:" + createRequestErr.Error())
	}

	// set the authorization header
	req.Header.Set("Authorization", "Bearer "+c.AccessToken)

	// set the accept header to application/octet-stream
	req.Header.Set("Accept", "application/octet-stream")

	log.Println("Downloading asset from URL:", url)

	// perform the HTTP request
	res, downloadErr := c.HttpClient.Do(req)
	if downloadErr != nil {
		return errors.New("Error downloading asset: " + downloadErr.Error())
	}
	defer res.Body.Close()

	// check if the response status is OK
	if res.StatusCode != http.StatusOK {
		return errors.New(fmt.Sprintf("Error downloading asset, status code: %v", res.StatusCode))
	}

	// create the output file
	outputFile, createFileErr := os.Create(outputPath)
	if createFileErr != nil {
		return errors.New("Error creating output file:" + createFileErr.Error())
	}
	defer outputFile.Close()

	// copy the response body to the output file
	_, writeToFileErr := io.Copy(outputFile, res.Body)
	if writeToFileErr != nil {
		return errors.New("Error writing to output file:" + writeToFileErr.Error())
	}

	log.Printf("Asset downloaded successfully to: \"%s\"", outputPath)
	return nil
}
