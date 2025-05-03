package main

import (
	"os"
	"path"
	"testing"

	"github.com/stretchr/testify/assert"
)

var testDirPath = "./test-assets"

func setupFileUtilsTest(t *testing.T) string {
	// create a temporary directory for testing
	tempDir := t.TempDir()

	return tempDir
}

func TestCreateDirIfIsNotExist_EmptyPath(t *testing.T) {
	// act: try to create a directory with an empty path
	createDirErr := createDirIfIsNotExist("")

	// assert: check if the error is as expected
	assert.Error(t, createDirErr, "Should return an error for empty path")
}

func TestCreateDirIfIsNotExist_Success(t *testing.T) {
	tempDir := setupFileUtilsTest(t)

	// arrange: define the directory path for the test
	dirPath := path.Join(tempDir, "create-dir-test")

	// act: create the directory for the test
	createDirErr := createDirIfIsNotExist(dirPath)
	if createDirErr != nil {
		t.Fatalf("Expected no error, got %v", createDirErr)
	}

	// assert: check if directory was created
	if _, statErr := os.Stat(dirPath); os.IsNotExist(statErr) {
		t.Fatalf("Expected directory to be created, but it does not exist")
	}
}
