package main

import (
	"log"
	"os"
)

var db = make(map[string]string)

func main() {
	// set the log entry prefix
	log.SetPrefix("[deploy-to-vm] ")
	log.Println("Starting deploy-to-vm server...")

	// create assets folder if not exists
	assetsDir := os.Getenv("DEPLOY_TO_VM_ASSETS_DIR")
	err := createDirIfIsNotExist(assetsDir)
	if err != nil {
		log.Fatalf("Error creating assets directory: \"%v\"", err)
	}

	r := setupRouter()
	// Listen and Server in 0.0.0.0:8080
	r.Run(":8080")
}
