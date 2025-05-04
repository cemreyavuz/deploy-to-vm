package main

import (
	"errors"
	"fmt"
	"log"
	"os"
	"path"
)

// Checks if a directory exists and creates it if it doesn't
func createDirIfIsNotExist(path string) error {
	if path == "" {
		return fmt.Errorf("Path cannot be empty")
	}

	if _, err := os.Stat(path); os.IsNotExist(err) {
		log.Printf("Directory does not exist, creating: \"%s\"", path)
		err = os.MkdirAll(path, os.ModePerm)
		if err != nil {
			return fmt.Errorf("Failed to create directory: %w", err)
		}
	} else {
		log.Printf("Directory already exists: \"%s\"", path)
	}
	return nil
}

// Checks if a release directory exists and creates it if it doesn't
func createReleaseDirIfIsNotExist(assetsDir string, owner string, repo string, tag string) error {
	if assetsDir == "" || owner == "" || repo == "" || tag == "" {
		return errors.New("Assets directory, owner, repo, or tag cannot be empty")
	}

	releaseDirPath := path.Join(assetsDir, owner, repo, tag)

	return createDirIfIsNotExist(releaseDirPath)
}
