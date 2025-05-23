package main

import (
	"log"
	"net/http"
	"os"

	"github.com/joho/godotenv"
)

var db = make(map[string]string)

func main() {
	// set the log entry prefix
	log.SetPrefix("[deploy-to-vm] ")
	log.Println("Starting deploy-to-vm server...")

	// load .env file
	dotenvErr := godotenv.Load()
	if dotenvErr != nil {
		log.Fatalf("No .env file found or error loading .env file")
	}

	// create assets folder if not exists
	assetsDir := os.Getenv("DEPLOY_TO_VM_ASSETS_DIR")
	err := createDirIfIsNotExist(assetsDir)
	if err != nil {
		log.Fatalf("Error creating assets directory: \"%v\"", err)
	}

	// Create db client
	dbClient := &DBClient{}
	dbClient.Connect()

	// create github client
	githubAccessToken := os.Getenv("DEPLOY_TO_VM_GITHUB_ACCESS_TOKEN")
	githubClient := &GithubClient{
		AccessToken: githubAccessToken,
		HttpClient:  &http.Client{},
	}

	// create nginx client
	nginxClient := &NginxClient{}

	// create router
	r := setupRouter(RouterOptions{
		AssetsDir:    assetsDir,
		GithubClient: githubClient,
		NginxClient:  nginxClient,
	})

	// Listen and Server in 0.0.0.0:8080
	r.Run(":8080")
}
