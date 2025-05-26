package main

import (
	"log"
)

// NginxClient is a struct that represents a client for interacting with the
// Nginx installation in the VM.
type NginxClient struct {
	ExecClient ExecClientInterface
}

// NginxClientInterface is an interface that defines the methods for the
// NginxClient struct. This allows for easier testing and mocking of the
// NginxClient in unit tests. The interface can be implemented by any struct
// that has the same methods as the NginxClient struct.
type NginxClientInterface interface {
	Reload() error
}

// Reloads the nginx unit in systemctl
func (c *NginxClient) Reload() error {
	out, err := c.ExecClient.Command("systemctl", "reload", "nginx").CombinedOutput()
	if err != nil {
		log.Printf("Error reloading nginx unit: %v", string(out))
	} else {
		log.Printf("Reloaded nginx unit")
	}
	return err
}

func NewNginxClient(execClient ExecClientInterface) *NginxClient {
	// If execClient is nil, create a new ExecClient instance
	if execClient == nil {
		return &NginxClient{
			ExecClient: &ExecClient{},
		}
	}

	// If execClient is provided, use it to create the NginxClient
	return &NginxClient{
		ExecClient: execClient,
	}
}
