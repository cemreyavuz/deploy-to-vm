package main

import (
	"os"
	"path"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetConfig_AlreadyLoadedConfigSuccess(t *testing.T) {
	// Arrange: create a ConfigClient instance
	configClient := &ConfigClient{}
	configClient.Config = &DeployToVmConfig{
		Repositories: []DeployToVmConfigRepository{
			{
				Name:       "test-repo",
				Owner:      "test-owner",
				SourceType: "github",
				TargetDir:  "/var/www/test-repo",
				TargetType: "nginx",
			},
		},
	}

	// Act: call GetConfig to retrieve the already loaded config
	config := configClient.GetConfig()

	// Assert: check if the config is as expected
	assert.NotNil(t, config, "Expected config to be not nil")
	assert.Equal(t, 1, len(config.Repositories), "Expected one repository in config")
	assert.Equal(t, "test-repo", config.Repositories[0].Name, "Expected repository name to match")
	assert.Equal(t, "test-owner", config.Repositories[0].Owner, "Expected repository owner to match")
	assert.Equal(t, "github", config.Repositories[0].SourceType, "Expected source type to match")
	assert.Equal(t, "/var/www/test-repo", config.Repositories[0].TargetDir, "Expected target directory to match")
	assert.Equal(t, "nginx", config.Repositories[0].TargetType, "Expected target type to match")
}

func TestGetConfig_EmptyConfigSuccess(t *testing.T) {
	// Arrange: create a dummy config in a temporary directory
	tempDir := t.TempDir()
	configFilePath := path.Join(tempDir, "config.json")
	os.WriteFile(configFilePath, []byte("{}"), 0644)

	// Arrange: set the environment variable for config file path
	os.Setenv("DEPLOY_TO_VM_CONFIG_FILE_PATH", configFilePath)
	defer os.Unsetenv("DEPLOY_TO_VM_CONFIG_FILE_PATH")

	// Arrange: create a ConfigClient instance with an empty config
	configClient := &ConfigClient{}

	// Act: call GetConfig to retrieve the config
	config := configClient.GetConfig()

	// Assert: check if the config is initialized and empty
	assert.NotNil(t, config, "Expected config to be not nil")
	assert.Equal(t, 0, len(config.Repositories), "Expected no repositories in config")
}

func TestGetRepository_Success(t *testing.T) {
	// Arrange: create a ConfigClient instance with a loaded config
	configClient := &ConfigClient{}
	configClient.Config = &DeployToVmConfig{
		Repositories: []DeployToVmConfigRepository{
			{
				Name:       "test-repo",
				Owner:      "test-owner",
				SourceType: "github",
				TargetDir:  "/var/www/test-repo",
				TargetType: "nginx",
			},
		},
	}

	// Act: call GetRepository to retrieve the repository
	repo := configClient.GetRepository("test-repo", "test-owner")

	// Assert: check if the repository is as expected
	assert.NotNil(t, repo, "Expected repository to be not nil")
	assert.Equal(t, "test-repo", repo.Name, "Expected repository name to match")
	assert.Equal(t, "test-owner", repo.Owner, "Expected repository owner to match")
	assert.Equal(t, "github", repo.SourceType, "Expected source type to match")
	assert.Equal(t, "/var/www/test-repo", repo.TargetDir, "Expected target directory to match")
	assert.Equal(t, "nginx", repo.TargetType, "Expected target type to match")
}

func TestGetRepository_NotFound(t *testing.T) {
	// Arrange: create a ConfigClient instance with a loaded config
	configClient := &ConfigClient{}
	configClient.Config = &DeployToVmConfig{
		Repositories: []DeployToVmConfigRepository{
			{
				Name:       "test-repo",
				Owner:      "test-owner",
				SourceType: "github",
				TargetDir:  "/var/www/test-repo",
				TargetType: "nginx",
			},
		},
	}

	// Act: call GetRepository with a non-existing repository
	repo := configClient.GetRepository("non-existing-repo", "test-owner")

	// Assert: check if the repository is nil
	assert.Nil(t, repo, "Expected repository to be nil for non-existing repo")
}

func TestLoadConfig_Success(t *testing.T) {
	// Arrange: create a temporary config file for testing
	tempDir := t.TempDir()
	configFilePath := path.Join(tempDir, "config.json")
	expectedConfigContent := `{
	"repositories": [
		{
			"name": "test-repo",
			"owner": "test-owner",
			"sourceType": "github",
			"targetDir": "/var/www/test-repo",
			"targetType": "nginx"
		}
	]
}`
	os.WriteFile(configFilePath, []byte(expectedConfigContent), 0644)
	os.Setenv("DEPLOY_TO_VM_CONFIG_FILE_PATH", configFilePath)
	defer os.Unsetenv("DEPLOY_TO_VM_CONFIG_FILE_PATH")

	// Act: load the configuration from the file
	configClient := &ConfigClient{}
	loadErr := configClient.LoadConfig()

	// Assert: check if there is no error reading the config file
	assert.NoError(t, loadErr, "Expected no error reading config file, got %v", loadErr)

	config := configClient.Config

	// Assert: check if the file content is as expected
	assert.Equal(t, 1, len(config.Repositories), "Expected one repository in config")
	assert.Equal(t, "test-repo", config.Repositories[0].Name, "Expected repository name to match")
	assert.Equal(t, "test-owner", config.Repositories[0].Owner, "Expected repository owner to match")
	assert.Equal(t, "github", config.Repositories[0].SourceType, "Expected source type to match")
	assert.Equal(t, "/var/www/test-repo", config.Repositories[0].TargetDir, "Expected target directory to match")
	assert.Equal(t, "nginx", config.Repositories[0].TargetType, "Expected target type to match")
}

func TestLoadConfig_EnvVarNotSet(t *testing.T) {
	// Arrange: unset the environment variable for config file path
	os.Unsetenv("DEPLOY_TO_VM_CONFIG_FILE_PATH")

	// Act: create a ConfigClient and attempt to load the config
	configClient := &ConfigClient{}
	loadErr := configClient.LoadConfig()

	// Assert: check if the error is as expected
	assert.Error(t, loadErr, "Expected error when DEPLOY_TO_VM_CONFIG_FILE_PATH is not set")
	assert.Equal(t, os.ErrNotExist.Error(), loadErr.Error())
}

func TestLoadConfig_FileNotFound(t *testing.T) {
	// Arrange: set the environment variable to a non-existing file path
	os.Setenv("DEPLOY_TO_VM_CONFIG_FILE_PATH", "/non/existing/path/config.json")
	defer os.Unsetenv("DEPLOY_TO_VM_CONFIG_FILE_PATH")

	// Act: create a ConfigClient and attempt to load the config
	configClient := &ConfigClient{}
	loadErr := configClient.LoadConfig()

	// Assert: check if the error is as expected
	assert.Error(t, loadErr, "Expected error when config file does not exist")
	assert.Contains(t, loadErr.Error(), "no such file or directory", "Expected file not found error")
}

func TestLoadConfig_InvalidJSON(t *testing.T) {
	// Arrange: create a temporary config file with invalid JSON
	tempDir := t.TempDir()
	configFilePath := path.Join(tempDir, "config.json")
	os.WriteFile(configFilePath, []byte("{invalid json}"), 0644)
	os.Setenv("DEPLOY_TO_VM_CONFIG_FILE_PATH", configFilePath)
	defer os.Unsetenv("DEPLOY_TO_VM_CONFIG_FILE_PATH")

	// Act: create a ConfigClient and attempt to load the config
	configClient := &ConfigClient{}
	loadErr := configClient.LoadConfig()

	// Assert: check if the error is as expected
	assert.Error(t, loadErr, "Expected error when config file contains invalid JSON")
	assert.Contains(t, loadErr.Error(), "invalid character", "Expected JSON parsing error")
}
