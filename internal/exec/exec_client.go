package exec

import (
	"os/exec"
)

type ExecClient struct{}

type ExecClientInterface interface {
	Command(command string, args ...string) ExecCommandInterface
}

func (execClient *ExecClient) Command(command string, args ...string) ExecCommandInterface {
	cmd := exec.Command(command, args...)

	return &ExecCommand{
		cmd: cmd,
	}
}
