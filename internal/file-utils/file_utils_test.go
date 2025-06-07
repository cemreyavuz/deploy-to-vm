package file_utils

import (
	"archive/tar"
	"compress/gzip"
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

// Helper to create a tar.gz file with one file inside
func createTestTarGz(t *testing.T, tarGzPath, fileName string, content []byte) {
	f, err := os.Create(tarGzPath)
	assert.NoError(t, err)
	defer f.Close()

	gw := gzip.NewWriter(f)
	defer gw.Close()

	tw := tar.NewWriter(gw)
	defer tw.Close()

	hdr := &tar.Header{
		Name: fileName,
		Mode: 0600,
		Size: int64(len(content)),
	}
	assert.NoError(t, tw.WriteHeader(hdr))
	_, err = tw.Write(content)
	assert.NoError(t, err)
}

func TestCreateDirIfIsNotExist_EmptyPath(t *testing.T) {
	// act: try to create a directory with an empty path
	createDirErr := CreateDirIfIsNotExist("")

	// assert: check if the error is as expected
	assert.Error(t, createDirErr, "Should return an error for empty path")
}

func TestCreateDirIfIsNotExist_MkdirAll_Error(t *testing.T) {
	tempDir := setupFileUtilsTest(t)

	// Arrange: create a directory with no permissions
	noPermDir := path.Join(tempDir, "no-perm-dir")
	err := os.MkdirAll(noPermDir, 0555) // Create with read and execute permissions only - no write permissions
	assert.NoError(t, err, "Expected no error creating directory with no permissions")

	// Act: try to create a directory in the no-perm directory
	createDirErr := CreateDirIfIsNotExist(path.Join(noPermDir, "subdir"))

	// Assert: check if the error is as expected
	assert.Error(t, createDirErr, "Should return an error for mkdirall failure")
	assert.Contains(t, createDirErr.Error(), "Failed to create directory:", "Expected error message to contain 'Failed to create directory'")
}

func TestCreateDirIfIsNotExist_AlreadyExists(t *testing.T) {
	tempDir := setupFileUtilsTest(t)

	// Act: create a directory that already exists
	createDirErr := CreateDirIfIsNotExist(tempDir)

	// Assert: check if the error is nil (directory already exists)
	assert.NoError(t, createDirErr, "Expected no error when directory already exists")
}

func TestCreateDirIfIsNotExist_Success(t *testing.T) {
	tempDir := setupFileUtilsTest(t)

	// arrange: define the directory path for the test
	dirPath := path.Join(tempDir, "create-dir-test")

	// act: create the directory for the test
	createDirErr := CreateDirIfIsNotExist(dirPath)
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
	_, emptyTagErr := CreateReleaseDirIfIsNotExist(tempDir, owner, repo, "")
	assert.Error(t, emptyTagErr, "Should return an error for empty tag parameter")

	// act/assert: try to create a release directory with empty repo parameter
	_, emptyRepoErr := CreateReleaseDirIfIsNotExist(tempDir, owner, "", tag)
	assert.Error(t, emptyRepoErr, "Should return an error for empty repo parameter")

	// act/assert: try to create a release directory with empty owner parameter
	_, emptyOwnerErr := CreateReleaseDirIfIsNotExist(tempDir, "", repo, tag)
	assert.Error(t, emptyOwnerErr, "Should return an error for empty owner parameter")

	// act/assert: try to create a release directory with empty assetsDir parameter
	_, assetsDirErr := CreateReleaseDirIfIsNotExist("", owner, repo, tag)
	assert.Error(t, assetsDirErr, "Should return an error for empty assetsDir parameter")

	// act/assert: try to create a release directory with empty parameters
	_, emptyParamsErr := CreateReleaseDirIfIsNotExist("", "", "", "")
	assert.Error(t, emptyParamsErr, "Should return an error for empty parameters")
}

func TestCreateReleaseDirIfIsNotExist_ClearContent(t *testing.T) {
	tempDir := setupFileUtilsTest(t)

	// Arrange: define the directory path for the test
	owner := "test-owner"
	repo := "test-repo"
	tag := "v1.0.0"
	dirPath := path.Join(tempDir, owner, repo, tag)

	// Arrange: create the release dir
	err := os.MkdirAll(dirPath, 0755)
	assert.NoError(t, err, "Expected no error creating initial directory")

	// Arrange: create a file in the directory
	filePath := path.Join(dirPath, "testfile.txt")
	err = os.WriteFile(filePath, []byte("content"), 0644)
	assert.NoError(t, err, "Expected no error creating test file")

	// Act: call CreateReleaseDirIfIsNotExist to clear the content
	CreateReleaseDirIfIsNotExist(tempDir, owner, repo, tag)

	// Assert: check if the directory was cleared
	if _, statErr := os.Stat(filePath); !os.IsNotExist(statErr) {
		t.Fatalf("Expected content to be cleared but file still exists")
	}
}

func TestCreateReleaseDirIfIsNotExist_ClearContentError(t *testing.T) {
	tempDir := setupFileUtilsTest(t)

	// Arrange: define the directory path for the test
	owner := "test-owner"
	repo := "test-repo"
	tag := "v1.0.0"
	dirPath := path.Join(tempDir, owner, repo, tag)

	// Arrange: create the release dir that cannot be modified
	err := os.MkdirAll(dirPath, 0755) // no permissions
	assert.NoError(t, err, "Expected no error creating initial directory")

	// Arrange: create a file in the directory
	filePath := path.Join(dirPath, "testfile.txt")
	err = os.WriteFile(filePath, []byte("content"), 0644)
	assert.NoError(t, err, "Expected no error creating test file")

	os.Chmod(dirPath, 0444) // set directory to read-only

	// Act: call CreateReleaseDirIfIsNotExist to clear the content
	_, err = CreateReleaseDirIfIsNotExist(tempDir, owner, repo, tag)

	assert.Error(t, err, "Expected error when trying to clear content of read-only directory")
	assert.Contains(t, err.Error(), "error clearing release directory")

	t.Cleanup(func() {
		os.Chmod(dirPath, 0755) // restore permissions for cleanup
	})
}

func TestCreateReleaseDirIfIsNotExist_Success(t *testing.T) {
	tempDir := setupFileUtilsTest(t)

	// arrange: define the directory path for the test
	owner := "test-owner"
	repo := "test-repo"
	tag := "v1.0.0"
	dirPath := path.Join(tempDir, owner, repo, tag)

	// act: create the release directory
	releaseDir, createDirErr := CreateReleaseDirIfIsNotExist(tempDir, owner, repo, tag)

	// assert: check if the release directory path is correct
	assert.Equal(t, dirPath, releaseDir, "Expected release directory path to match")

	// assert: check if the error is nil
	assert.NoError(t, createDirErr, "Expected no error when creating release directory")

	// assert: check if directory was created
	if _, statErr := os.Stat(dirPath); os.IsNotExist(statErr) {
		t.Fatalf("Expected directory to be created, but it does not exist")
	}
}

func TestReadFilesInDir_EmptyDir(t *testing.T) {
	tempDir := setupFileUtilsTest(t)

	// Act: read files in an empty directory
	files, err := ReadFilesInDir(tempDir)

	// Assert: check if the error is nil and files slice is empty
	assert.NoError(t, err, "Expected no error for empty directory")
	assert.Empty(t, files, "Expected no files in empty directory")
}

func TestReadFilesInDir_SingleFile(t *testing.T) {
	tempDir := setupFileUtilsTest(t)

	// Arrange: create a single file
	filePath := path.Join(tempDir, "testfile.txt")
	err := os.WriteFile(filePath, []byte("content"), 0644)
	assert.NoError(t, err, "Expected no error creating file")

	// Act: read files in the directory
	files, readErr := ReadFilesInDir(tempDir)

	// Assert: check if the error is nil and files slice contains the file
	assert.NoError(t, readErr, "Expected no error reading directory")
	assert.Len(t, files, 1, "Expected one file in directory")
	assert.Equal(t, filePath, files[0], "Expected file path to match")
}

func TestReadFilesInDir_NestedFiles(t *testing.T) {
	tempDir := setupFileUtilsTest(t)

	// Arrange: create nested directories and files
	nestedDir := path.Join(tempDir, "nested")
	err := os.MkdirAll(nestedDir, 0755)
	assert.NoError(t, err, "Expected no error creating nested directory")

	// Arrange: create files in the nested directory and the main directory
	file1 := path.Join(tempDir, "file1.txt")
	file2 := path.Join(nestedDir, "file2.txt")
	os.WriteFile(file1, []byte("a"), 0644)
	os.WriteFile(file2, []byte("b"), 0644)

	// Act: read files in the directory
	files, readErr := ReadFilesInDir(tempDir)

	// Assert: check if the error is nil and files slice contains both files
	assert.NoError(t, readErr, "Expected no error reading directory")
	assert.ElementsMatch(t, []string{file1, file2}, files, "Expected all files to be listed")
}

func TestReadFilesInDir_NonExistentDir(t *testing.T) {
	// Arrange: create a non-existent directory path
	nonExistentDir := path.Join(t.TempDir(), "does-not-exist-12345")

	// Act: try to read files in the non-existent directory
	files, err := ReadFilesInDir(nonExistentDir)

	// Assert: check if the error is not nil and files slice is nil
	assert.Error(t, err, "Expected error for non-existent directory")
	assert.ElementsMatch(t, []string{}, files, "Expected files to be empty array")
}

func TestUntarGzFilesInDir_NonExistentDir(t *testing.T) {
	// Act: untar files in an empty directory
	_, err := UntarGzFilesInDir("")

	// Assert: check if the error is not nil
	assert.Error(t, err, "Expected error for non-existent directory")
	assert.Contains(t, err.Error(), "Error while reading the directory:", "Expected error for non-existent directory")
}

func TestUntarGzFilesInDir_OpenError(t *testing.T) {
	tempDir := setupFileUtilsTest(t)

	// Arrange: create a tar.gz file with no read permissions
	badTarGz := path.Join(tempDir, "missing.tar.gz")
	os.WriteFile(badTarGz, []byte{}, 0000)

	// Act
	_, err := UntarGzFilesInDir(tempDir)

	// Assert: expect an error due to permission issues
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "Error while opening the file:")
}

func TestUntarGzFilesInDir_ExtractsFilesNested(t *testing.T) {
	tempDir := setupFileUtilsTest(t)

	// Arrange: create a tar.gz file with a nested file structure
	tarGzPath := path.Join(tempDir, "nested.tar.gz")
	content := []byte("nested content")
	nestedFileName := "nested/hello.txt"
	createTestTarGz(t, tarGzPath, nestedFileName, content)

	// Act: extract files
	_, err := UntarGzFilesInDir(tempDir)

	// Assert: no error
	assert.NoError(t, err, "Expected no error extracting tar.gz")

	// Assert: check that the extracted file exists and has correct content
	extractedFilePath := path.Join(tempDir, nestedFileName)
	data, readErr := os.ReadFile(extractedFilePath)
	assert.NoError(t, readErr, "Expected to read extracted file")
	assert.Equal(t, content, data, "Extracted file content should match original")
}

func TestUntarGzFilesInDir_ExtractsFiles(t *testing.T) {
	tempDir := setupFileUtilsTest(t)

	// Arrange: create a tar.gz file with a single file inside
	tarGzPath := path.Join(tempDir, "test.tar.gz")
	content := []byte("hello world")
	fileName := "hello.txt"
	createTestTarGz(t, tarGzPath, fileName, content)

	// Act: extract files
	_, err := UntarGzFilesInDir(tempDir)

	// Assert: no error
	assert.NoError(t, err, "Expected no error extracting tar.gz")

	// Assert: check that the extracted file exists and has correct content
	extractedFilePath := path.Join(tempDir, fileName)
	data, readErr := os.ReadFile(extractedFilePath)
	assert.NoError(t, readErr, "Expected to read extracted file")
	assert.Equal(t, content, data, "Extracted file content should match original")

	// Assert: check that the tar.gz file is removed
	_, statErr := os.Stat(tarGzPath)
	assert.True(t, os.IsNotExist(statErr), "Expected tar.gz file to be removed after extraction")
}

func TestUntarGzFilesInDir_SkipsNonGzFiles(t *testing.T) {
	tempDir := setupFileUtilsTest(t)

	// Arrange: create a regular file (not .gz)
	nonGzFile := path.Join(tempDir, "not-a-tar.txt")
	os.WriteFile(nonGzFile, []byte("data"), 0644)

	// Act: untar files in the directory
	_, err := UntarGzFilesInDir(tempDir)

	// Assert: no error
	assert.NoError(t, err, "Expected no error when only non-gz files are present")

	// Assert: check that the non-gz file still exists
	_, statErr := os.Stat(nonGzFile)
	assert.NoError(t, statErr, "Expected non-gz file to remain")
}

func TestUntarGzFilesInDir_InvalidGzFile(t *testing.T) {
	tempDir := setupFileUtilsTest(t)

	// Arrange: create an invalid .gz file
	invalidGz := path.Join(tempDir, "invalid.tar.gz")
	os.WriteFile(invalidGz, []byte("not a valid gzip"), 0644)

	// Act: untar files in the directory
	_, err := UntarGzFilesInDir(tempDir)

	// Assert: expect an error due to invalid gzip file
	assert.Error(t, err, "Expected error for invalid gzip file")
}

func TestUntarGzFilesInDir_RemoveTarGzAfterExtraction_Success(t *testing.T) {
	tempDir := setupFileUtilsTest(t)

	// Arrange: create a tar.gz file with a single file inside
	tarGzPath := path.Join(tempDir, "test.tar.gz")
	content := []byte("hello world")
	fileName := "hello.txt"
	createTestTarGz(t, tarGzPath, fileName, content)

	// Act: extract files
	_, err := UntarGzFilesInDir(tempDir)

	// Assert: no error
	assert.NoError(t, err, "Expected no error extracting tar.gz")

	// Assert: check that the extracted file exists and has correct content
	extractedFilePath := path.Join(tempDir, fileName)
	data, readErr := os.ReadFile(extractedFilePath)
	assert.NoError(t, readErr, "Expected to read extracted file")
	assert.Equal(t, content, data, "Extracted file content should match original")

	// Assert: check that the tar.gz file is removed
	_, statErr := os.Stat(tarGzPath)
	assert.True(t, os.IsNotExist(statErr), "Expected tar.gz file to be removed after extraction")
}

func TestLinkReleaseAssetsToSiteDir_Success(t *testing.T) {
	tempDir := setupFileUtilsTest(t)

	// Arrange: define release and site directories
	releaseDir := path.Join(tempDir, "release")
	siteDir := path.Join(tempDir, "site")

	// Arrange: create release directory and a file inside it
	err := os.MkdirAll(releaseDir, 0755)
	assert.NoError(t, err)
	fileName := "asset.txt"
	releaseFile := path.Join(releaseDir, fileName)
	content := []byte("asset content")
	err = os.WriteFile(releaseFile, content, 0644)
	assert.NoError(t, err)

	// Arrange: create site directory and a file to be removed
	err = os.MkdirAll(siteDir, 0755)
	assert.NoError(t, err)
	oldSiteFile := path.Join(siteDir, "old.txt")
	os.WriteFile(oldSiteFile, []byte("old"), 0644)

	// Act: link release assets to site directory
	linkErr := LinkReleaseAssetsToSiteDir(releaseDir, siteDir)

	// Assert: no error
	assert.NoError(t, linkErr, "Expected no error linking assets")

	// Assert: old file is removed
	_, statErr := os.Stat(oldSiteFile)
	assert.True(t, os.IsNotExist(statErr), "Expected old file to be removed from site directory")

	// Assert: asset is linked in site directory
	linkedFile := path.Join(siteDir, fileName)
	info, statErr := os.Stat(linkedFile)
	assert.NoError(t, statErr, "Expected linked file to exist in site directory")
	assert.False(t, info.IsDir(), "Expected linked file to be a file")

	// Assert: linked file is a hard link (same inode)
	releaseInfo, _ := os.Stat(releaseFile)
	assert.Equal(t, releaseInfo.Size(), info.Size(), "Expected linked file to have same size as source")
}

func TestLinkReleaseAssetsToSiteDir_EmptyReleaseDir(t *testing.T) {
	tempDir := setupFileUtilsTest(t)

	// Arrange: create empty release and site directories
	releaseDir := path.Join(tempDir, "release")
	siteDir := path.Join(tempDir, "site")
	os.MkdirAll(releaseDir, 0755)
	os.MkdirAll(siteDir, 0755)

	// Act: link release assets to site directory
	linkErr := LinkReleaseAssetsToSiteDir(releaseDir, siteDir)

	// Assert: no error, nothing to link
	assert.NoError(t, linkErr, "Expected no error when release dir is empty")
}

func TestLinkReleaseAssetsToSiteDir_ErrorOnInvalidReleaseDir(t *testing.T) {
	tempDir := setupFileUtilsTest(t)

	// Arrange: create a site directory but no valid release directory
	siteDir := path.Join(tempDir, "site")
	os.MkdirAll(siteDir, 0755)

	// Act: try to link a non-existent release directory
	linkErr := LinkReleaseAssetsToSiteDir("/nonexistent/release", siteDir)

	// Assert: error due to invalid release dir
	assert.Error(t, linkErr, "Expected error for invalid release directory")
}

func TestLinkReleaseAssetsToSiteDir_ErrorOnInvalidSiteDir(t *testing.T) {
	tempDir := setupFileUtilsTest(t)

	// Arrange: create a release directory but no valid site directory
	releaseDir := path.Join(tempDir, "release")
	os.MkdirAll(releaseDir, 0755)

	// Act: try to link to a non-existent site directory
	linkErr := LinkReleaseAssetsToSiteDir(releaseDir, "/nonexistent/site")

	// Assert: error due to invalid site dir
	assert.Error(t, linkErr, "Expected error for invalid site directory")
}
