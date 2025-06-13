package pm2

import (
	deploy_to_vm_exec "deploy-to-vm/internal/exec"
	"fmt"
	"log"
)

type Pm2Client struct {
	ExecClient deploy_to_vm_exec.ExecClientInterface
}

type Pm2ClientInterface interface {
	Reload(targetProcessName string) error
}

func (c *Pm2Client) Reload(targetProcessName string) error {
	if targetProcessName == "" {
		return fmt.Errorf("targetProcessName cannot be empty")
	}

	out, err := c.ExecClient.Command("pm2", "reload", targetProcessName).CombinedOutput()
	if err != nil {
		log.Printf("Error reloading pm2 process \"%s\": %v", targetProcessName, string(out))
	} else {
		log.Printf("Reloaded pm2 process \"%s\"", targetProcessName)
	}

	return err
}

func NewPm2Client(execClient deploy_to_vm_exec.ExecClientInterface) *Pm2Client {
	// If execClient is nil, create a new ExecClient instance
	if execClient == nil {
		return &Pm2Client{
			ExecClient: &deploy_to_vm_exec.ExecClient{},
		}
	}

	// If execClient is provided, use it to create the Pm2Client
	return &Pm2Client{
		ExecClient: execClient,
	}
}
