package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"deploy-to-vm/internal/config"
	file_utils "deploy-to-vm/internal/file-utils"
	deploy_to_vm_github "deploy-to-vm/internal/github"
	"deploy-to-vm/internal/nginx"
	"deploy-to-vm/internal/notification"
	"deploy-to-vm/internal/pm2"
	"deploy-to-vm/internal/router"

	"github.com/cloudflare/tableflip"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func NewUpgrader(options tableflip.Options) (*tableflip.Upgrader, error) {
	upg, err := tableflip.New(options)
	if err != nil {
		return nil, fmt.Errorf("error creating tableflip upgrader: %v", err)
	}

	return upg, nil
}

func startServer(r *gin.Engine, upg *tableflip.Upgrader) {
	port := os.Getenv("DEPLOY_TO_VM_PORT")
	if port == "" {
		log.Fatalln("Environment variable DEPLOY_TO_VM_PORT is not set")
	}

	// Do an upgrade on SIGHUP
	go func() {
		sig := make(chan os.Signal, 1)
		signal.Notify(sig, syscall.SIGHUP)
		for range sig {
			err := upg.Upgrade()
			if err != nil {
				log.Printf("Error during upgrade: %v", err)
			}
		}
	}()

	listenAddr := ":" + port

	// Listen must be called before ready
	ln, err := upg.Listen("tcp", listenAddr)
	if err != nil {
		log.Fatalln("Error listening on", listenAddr, ":", err)
	}
	defer upg.Stop()

	server := &http.Server{
		Addr:    listenAddr,
		Handler: r,
	}

	go func() {
		if err := server.Serve(ln); err != nil && err != http.ErrServerClosed {
			log.Fatalln("HTTP server:", err)
		}
	}()

	log.Println("Server is ready!")
	if err := upg.Ready(); err != nil {
		panic(err)
	}
	<-upg.Exit()

	// Make sure to set a deadline on exiting the process after upg.Exit() is
	// called. No new upgrades can be performed if the parent doesn't exist.
	time.AfterFunc(30*time.Second, func() {
		log.Println("Graceful shutdown timed out, forcing exit")
		os.Exit(1)
	})

	// Wait for connections to drain
	server.Shutdown(context.Background())
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
		log.Println("No .env file found or error loading .env file")
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

	// Create pm2 client
	pm2Client := pm2.NewPm2Client(nil)

	// Read secret token from environment variable
	secretToken := os.Getenv("DEPLOY_TO_VM_SECRET_TOKEN")
	if secretToken == "" {
		log.Fatal("Environment variable DEPLOY_TO_VM_SECRET_TOKEN is not set")
	}

	// Create notification client
	notificationClient := notification.SetupNotificationClient()

	pidFile := os.Getenv("DEPLOY_TO_VM_PID_FILE")
	if pidFile == "" {
		// TODO(cemreyavuz): alternatively, we can just decide to not support graceful restart
		log.Fatalln("Environment variable DEPLOY_TO_VM_PID_FILE is not set")
	} else {
		log.Printf("pidFile: %s", pidFile)
	}

	// Create upgrader
	upg, err := NewUpgrader(tableflip.Options{
		PIDFile: pidFile,
	})
	if err != nil {
		log.Fatalf("Error creating upgrader: %v", err)
	}

	// Create router
	r := router.SetupRouter(router.RouterOptions{
		AssetsDir:          assetsDir,
		ConfigClient:       configClient,
		GithubClient:       githubClient,
		NginxClient:        nginxClient,
		NotificationClient: notificationClient,
		Pm2Client:          pm2Client,
		SecretToken:        secretToken,
		Upgrader:           upg,
	})

	// Start the server
	startServer(r, upg)
}
