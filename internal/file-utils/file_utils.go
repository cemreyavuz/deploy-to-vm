package file_utils

import (
	"archive/tar"
	"compress/gzip"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"path"
	"path/filepath"
	"strings"
)

// Checks if a directory exists and creates it if it doesn't
func CreateDirIfIsNotExist(path string) error {
	if path == "" {
		return fmt.Errorf("Alper")
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
func CreateReleaseDirIfIsNotExist(assetsDir string, owner string, repo string, tag string) (string, error) {
	if assetsDir == "" || owner == "" || repo == "" || tag == "" {
		return "", errors.New("Assets directory, owner, repo, or tag cannot be empty")
	}

	releaseDirPath := path.Join(assetsDir, owner, repo, tag)

	createDirErr := CreateDirIfIsNotExist(releaseDirPath)
	return releaseDirPath, createDirErr
}

// Read files in a directory recursively
func ReadFilesInDir(dir string) ([]string, error) {
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

// Untar gz files in a directory. It reads all files in the directory, checks if
// they are tar files, and extracts them.
func UntarGzFilesInDir(dir string) ([]string, error) {
	// Read files in the directory recursively
	files, readErr := ReadFilesInDir(dir)
	if readErr != nil {
		return nil, errors.New(fmt.Sprintf("Error while reading the directory: %v", readErr))
	}

	log.Printf("Found files in the directory: \n- %v", strings.Join(files, "\n- "))

	processedFiles := make([]string, 0)

	// Iterate through each file
	for _, filePath := range files {
		if filepath.Ext(filePath) != ".gz" {
			processedFiles = append(processedFiles, filePath)
			log.Printf("Skipping non-tar file: %v", filePath)
			continue
		}

		// Get target folder for the extracted files
		targetDir := filepath.Dir(filePath)

		log.Println("Processing tar file:", filePath)

		file, openErr := os.Open(filePath)
		if openErr != nil {
			return nil, errors.New(fmt.Sprintf("Error while opening the file: %v", openErr))
		}

		gzr, newReaderErr := gzip.NewReader(file)
		if newReaderErr != nil {
			return nil, errors.New(fmt.Sprintf("Error while creating gzip reader: %v", newReaderErr))
		}

		tr := tar.NewReader(gzr)
		for {
			header, err := tr.Next()
			if err == io.EOF {
				break // End of tar archive
			}

			if err != nil {
				return nil, errors.New(fmt.Sprintf("Error while reading the tar file: %v", err))
			}

			switch header.Typeflag {
			case tar.TypeDir:
				// skip directories
				continue
			case tar.TypeReg:
				target := filepath.Join(targetDir, header.Name)
				if err := os.MkdirAll(filepath.Dir(target), 0755); err != nil {
					return nil, fmt.Errorf("mkdir for file: %w", err)
				}
				outFile, err := os.Create(target)
				if err != nil {
					return nil, fmt.Errorf("create file: %w", err)
				}
				if _, err := io.Copy(outFile, tr); err != nil {
					outFile.Close()
					return nil, fmt.Errorf("copy file: %w", err)
				}
				outFile.Close()
			}

			processedFiles = append(processedFiles, header.Name)
			// Log the extracted file
			log.Printf("Extracted file: %s", header.Name)
		}

		// Remove the original gz file after extraction
		removeErr := os.Remove(filePath)
		if removeErr != nil {
			return nil, errors.New(fmt.Sprintf("Error while removing the original gz file: %v", removeErr))
		} else {
			log.Printf("Removed original gz file: %s", filePath)
		}
	}

	return processedFiles, nil
}

// Link release assets to site directory
func LinkReleaseAssetsToSiteDir(releaseDir string, siteDir string) error {
	// Read files in release directory recursively
	filesInReleaseDir, readReleaseDirErr := ReadFilesInDir(releaseDir)
	if readReleaseDirErr != nil {
		return errors.New(fmt.Sprintf("Error while reading the release directory: %v", readReleaseDirErr))
	}
	log.Printf("Found files in the release directory: \n- %v", strings.Join(filesInReleaseDir, "\n- "))

	// Read files in site directory
	filesInSiteDir, readSiteDirErr := ReadFilesInDir(siteDir)
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
