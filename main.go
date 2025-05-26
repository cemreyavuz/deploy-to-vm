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

	// Create config client and load config
	configClient := &ConfigClient{}
	loadConfigErr := configClient.LoadConfig()
	if loadConfigErr != nil {
		log.Fatalf("Error loading config: \"%v\"", loadConfigErr)
	} else {
		log.Println("Config loaded successfully")
	}

	// create assets folder if not exists
	assetsDir := os.Getenv("DEPLOY_TO_VM_ASSETS_DIR")
	err := createDirIfIsNotExist(assetsDir)
	if err != nil {
		log.Fatalf("Error creating assets directory: \"%v\"", err)
	}

	// create github client
	githubAccessToken := os.Getenv("DEPLOY_TO_VM_GITHUB_ACCESS_TOKEN")
	githubClient := &GithubClient{
		AccessToken: githubAccessToken,
		HttpClient:  &http.Client{},
	}

	// create nginx client
	nginxClient := NewNginxClient(nil)

	// Read secret token from environment variable
	secretToken := os.Getenv("DEPLOY_TO_VM_SECRET_TOKEN")
	if secretToken == "" {
		log.Fatal("Environment variable DEPLOY_TO_VM_SECRET_TOKEN is not set")
	}

	// Create notification client
	notificationClient := &NotificationClient{}
	notificationClient.LoadWebhookUrl()
	if notificationClient.webhookURL == "" {
		log.Println("Notification webhook URL is not set, notifications will not be sent")
	} else {
		log.Println("Notification webhook URL is set, notifications will be sent")
	}

	// create router
	r := setupRouter(RouterOptions{
		AssetsDir:          assetsDir,
		ConfigClient:       configClient,
		GithubClient:       githubClient,
		NginxClient:        nginxClient,
		NotificationClient: notificationClient,
		SecretToken:        secretToken,
	})

	// Run the server on the specified port
	port := os.Getenv("DEPLOY_TO_VM_PORT")
	if port == "" {
		log.Fatalf("Environment variable DEPLOY_TO_VM_PORT is not set")
	}
	r.Run(":" + port)
}
