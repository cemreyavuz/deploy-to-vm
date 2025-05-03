package main

import (
	"fmt"
	"os"
)

// Checks if a directory exists and creates it if it doesn't
func createDirIfIsNotExist(path string) error {
	if path == "" {
		return fmt.Errorf("Path cannot be empty")
	}

	if _, err := os.Stat(path); os.IsNotExist(err) {
		err = os.MkdirAll(path, os.ModePerm)
		if err != nil {
			return fmt.Errorf("Failed to create directory: %w", err)
		}
	}
	return nil
}
