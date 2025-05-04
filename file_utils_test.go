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

func TestCreateReleaseDirIfIsNotExist_EmptyParams(t *testing.T) {
	tempDir := setupFileUtilsTest(t)

	// arrange: define the directory path for the test
	owner := "test-owner"
	repo := "test-repo"
	tag := "v1.0.0"

	// act/assert: try to create a release directory with empty tag parameter
	_, emptyTagErr := createReleaseDirIfIsNotExist(tempDir, owner, repo, "")
	assert.Error(t, emptyTagErr, "Should return an error for empty tag parameter")

	// act/assert: try to create a release directory with empty repo parameter
	_, emptyRepoErr := createReleaseDirIfIsNotExist(tempDir, owner, "", tag)
	assert.Error(t, emptyRepoErr, "Should return an error for empty repo parameter")

	// act/assert: try to create a release directory with empty owner parameter
	_, emptyOwnerErr := createReleaseDirIfIsNotExist(tempDir, "", repo, tag)
	assert.Error(t, emptyOwnerErr, "Should return an error for empty owner parameter")

	// act/assert: try to create a release directory with empty assetsDir parameter
	_, assetsDirErr := createReleaseDirIfIsNotExist("", owner, repo, tag)
	assert.Error(t, assetsDirErr, "Should return an error for empty assetsDir parameter")

	// act/assert: try to create a release directory with empty parameters
	_, emptyParamsErr := createReleaseDirIfIsNotExist("", "", "", "")
	assert.Error(t, emptyParamsErr, "Should return an error for empty parameters")
}

func TestCreateReleaseDirIfIsNotExist_Success(t *testing.T) {
	tempDir := setupFileUtilsTest(t)

	// arrange: define the directory path for the test
	owner := "test-owner"
	repo := "test-repo"
	tag := "v1.0.0"
	dirPath := path.Join(tempDir, owner, repo, tag)

	// act: create the release directory
	releaseDir, createDirErr := createReleaseDirIfIsNotExist(tempDir, owner, repo, tag)

	// assert: check if the release directory path is correct
	assert.Equal(t, dirPath, releaseDir, "Expected release directory path to match")

	// assert: check if the error is nil
	assert.NoError(t, createDirErr, "Expected no error when creating release directory")

	// assert: check if directory was created
	if _, statErr := os.Stat(dirPath); os.IsNotExist(statErr) {
		t.Fatalf("Expected directory to be created, but it does not exist")
	}
}
