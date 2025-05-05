package main

import (
	"log"
	"os/exec"
)

// NginxClient is a struct that represents a client for interacting with the
// Nginx installation in the VM.
type NginxClient struct{}

// NginxClientInterface is an interface that defines the methods for the
// NginxClient struct. This allows for easier testing and mocking of the
// NginxClient in unit tests. The interface can be implemented by any struct
// that has the same methods as the NginxClient struct.
type NginxClientInterface interface {
	Reload() error
}

// Reloads the nginx unit in systemctl
func (nginxClient *NginxClient) Reload() error {
	out, err := exec.Command("systemctl", "reload", "nginx").CombinedOutput()
	if err != nil {
		log.Printf("Error reloading nginx unit: %v", string(out))
	} else {
		log.Printf("Reloaded nginx unit")
	}
	return err
}
