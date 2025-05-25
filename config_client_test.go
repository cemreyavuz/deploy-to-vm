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
