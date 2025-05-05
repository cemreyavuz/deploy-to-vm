package main

import (
	"errors"
	"fmt"
	"log"
	"os"
	"path"
	"path/filepath"
	"strings"
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
func createReleaseDirIfIsNotExist(assetsDir string, owner string, repo string, tag string) (string, error) {
	if assetsDir == "" || owner == "" || repo == "" || tag == "" {
		return "", errors.New("Assets directory, owner, repo, or tag cannot be empty")
	}

	releaseDirPath := path.Join(assetsDir, owner, repo, tag)

	createDirErr := createDirIfIsNotExist(releaseDirPath)
	return releaseDirPath, createDirErr
}

// Read files in a directory recursively
func readFilesInDir(dir string) ([]string, error) {
	files := make([]string, 0)
	readErr := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		files = append(files, path)
		return nil
	})

	return files, readErr
}

// Link release assets to site directory
func linkReleaseAssetsToSiteDir(releaseDir string, siteDir string) error {
	// Read files in release directory recursively
	filesInReleaseDir, readReleaseDirErr := readFilesInDir(releaseDir)
	if readReleaseDirErr != nil {
		return errors.New(fmt.Sprintf("Error while reading the release directory: %v", readReleaseDirErr))
	}
	log.Printf("Found files in the release directory: \n- %v", strings.Join(filesInReleaseDir, "\n- "))

	// Read files in site directory
	filesInSiteDir, readSiteDirErr := readFilesInDir(siteDir)
	if readSiteDirErr != nil {
		return errors.New(fmt.Sprintf("Error while reading the site directory: %v", readSiteDirErr))
	}
	log.Printf("Found files in the site directory: \n- %v", strings.Join(filesInSiteDir, "\n- "))

	// Remove files in site directory
	for _, file := range filesInSiteDir {
		removeErr := os.Remove(file)
		if removeErr != nil {
			return errors.New(fmt.Sprintf("Error while removing the file from the site directory: %v", removeErr))
		}
	}
	log.Printf("Removed files in the site directory")

	// Link files in release directory to site directory
	for _, file := range filesInReleaseDir {
		filePathRelativeToReleaseDir, relErr := filepath.Rel(releaseDir, file)
		if relErr != nil {
			return errors.New(fmt.Sprintf("Error while calculating the relative path for the asset: %v", relErr))
		}

		// Generate new file path in site directory
		filePathInSiteDir := path.Join(siteDir, filePathRelativeToReleaseDir)

		// Make sure the parent dir for file is created in site directory
		parentDir, _ := path.Split(filePathInSiteDir)
		mkdirErr := os.MkdirAll(parentDir, os.ModePerm)
		if mkdirErr != nil {
			return errors.New(fmt.Sprintf("Error while creating the parent directory for asset: %v", mkdirErr))
		}

		// Link release asset to site directory
		linkErr := os.Link(file, filePathInSiteDir)
		if linkErr != nil {
			return errors.New(fmt.Sprintf("Error while linking the asset: %v", linkErr))
		}

		log.Printf("File is linked: %v", file)
	}

	return nil
}
