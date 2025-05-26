package main

import (
	"os/exec"
)

type ExecClient struct{}

type ExecClientInterface interface {
	Command(command string, args ...string) *ExecCommand
}

func (execClient *ExecClient) Command(command string, args ...string) *ExecCommand {
	cmd := exec.Command(command, args...)

	return &ExecCommand{
		cmd: cmd,
	}
}
