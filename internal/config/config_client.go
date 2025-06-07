package config

import (
	"encoding/json"
	"log"
	"os"
)

type DeployToVmConfigRepository struct {
	Name       string `json:"name"`
	Owner      string `json:"owner"`
	SourceType string `json:"sourceType"`
	TargetDir  string `json:"targetDir"`
	TargetType string `json:"targetType"`
}

type DeployToVmConfig struct {
	Repositories []DeployToVmConfigRepository `json:"repositories"`
}

type ConfigClient struct {
	Config  *DeployToVmConfig
	DevFlag bool
}

type ConfigClientInterface interface {
	GetConfig() *DeployToVmConfig
	GetRepository(name string, owner string) *DeployToVmConfigRepository
	LoadConfig() error
}

func (c *ConfigClient) GetConfig() *DeployToVmConfig {
	// Load client config if not already loaded
	if c.Config == nil {
		loadErr := c.LoadConfig()
		if loadErr != nil {
			panic("Error loading config file: " + loadErr.Error())
		}
	}

	return c.Config
}

func (c *ConfigClient) GetRepository(name string, owner string) *DeployToVmConfigRepository {
	// Get the config
	config := c.GetConfig()

	// Search for the repository by name
	for _, repo := range config.Repositories {
		if repo.Name == name && repo.Owner == owner {
			return &repo
		}
	}

	// Return nil if not found
	return nil
}

func (c *ConfigClient) LoadConfig() error {
	// Read the config file path from environment variable
	configFilePath := os.Getenv("DEPLOY_TO_VM_CONFIG_FILE_PATH")
	if configFilePath == "" {
		log.Printf("Environment variable DEPLOY_TO_VM_CONFIG_FILE_PATH is not set")
		return os.ErrNotExist
	}

	// Read config file
	file, openErr := os.Open(configFilePath)
	if openErr != nil {
		log.Println("Error opening config file:", openErr)
		return openErr
	}

	// Ensure the file is closed after reading
	defer file.Close()

	// Decode the JSON config file into DeployToVmConfig struct
	var config DeployToVmConfig
	decoder := json.NewDecoder(file)
	if decodeErr := decoder.Decode(&config); decodeErr != nil {
		log.Println("Error decoding config file:", decodeErr)
		return decodeErr
	}

	c.Config = &config
	return nil
}
