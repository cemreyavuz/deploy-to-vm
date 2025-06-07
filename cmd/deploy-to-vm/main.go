package main

import (
	"errors"
	"flag"
	"log"
	"os"

	"deploy-to-vm/internal/config"
	file_utils "deploy-to-vm/internal/file-utils"
	deploy_to_vm_github "deploy-to-vm/internal/github"
	"deploy-to-vm/internal/nginx"
	"deploy-to-vm/internal/notification"
	"deploy-to-vm/internal/router"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func startServer(r *gin.Engine) error {
	port := os.Getenv("DEPLOY_TO_VM_PORT")
	if port == "" {
		return errors.New("Environment variable DEPLOY_TO_VM_PORT is not set")
	}

	return r.Run(":" + port)
}

func main() {
	// set the log entry prefix
	log.SetPrefix("[deploy-to-vm] ")
	log.Println("Starting deploy-to-vm server...")

	// Define command line flags
	devFlag := flag.Bool("dev", false, "Runs the server in development mode. In this mode, the server will not validate payloads for webhooks, allowing for easier testing and development.")
	flag.Parse()
	if *devFlag {
		log.Println("\"dev\" flag is set to true. Running in development mode.")
	}

	// load .env file
	dotenvErr := godotenv.Load()
	if dotenvErr != nil {
		log.Fatalf("No .env file found or error loading .env file")
	}

	// Create config client and load config
	configClient := &config.ConfigClient{
		DevFlag: *devFlag,
	}
	loadConfigErr := configClient.LoadConfig()
	if loadConfigErr != nil {
		log.Fatalf("Error loading config: \"%v\"", loadConfigErr)
	} else {
		log.Println("Config loaded successfully")
	}

	// create assets folder if not exists
	assetsDir := os.Getenv("DEPLOY_TO_VM_ASSETS_DIR")
	if assetsDir == "" {
		log.Fatal("Environment variable DEPLOY_TO_VM_ASSETS_DIR is not set")
	}
	err := file_utils.CreateDirIfIsNotExist(assetsDir)
	if err != nil {
		log.Fatalf("Error creating assets directory: \"%v\"", err)
	}

	// Create github client
	githubClient, err := deploy_to_vm_github.SetupGithubClient()
	if err != nil {
		log.Fatalf("Error setting up GitHub client: \"%v\"", err)
	}

	// create nginx client
	nginxClient := nginx.NewNginxClient(nil)

	// Read secret token from environment variable
	secretToken := os.Getenv("DEPLOY_TO_VM_SECRET_TOKEN")
	if secretToken == "" {
		log.Fatal("Environment variable DEPLOY_TO_VM_SECRET_TOKEN is not set")
	}

	// Create notification client
	notificationClient := notification.SetupNotificationClient()

	// Create router
	r := router.SetupRouter(router.RouterOptions{
		AssetsDir:          assetsDir,
		ConfigClient:       configClient,
		GithubClient:       githubClient,
		NginxClient:        nginxClient,
		NotificationClient: notificationClient,
		SecretToken:        secretToken,
	})

	// Start the server
	if err := startServer(r); err != nil {
		log.Fatalf("Error starting server: \"%v\"", err)
	}
}
