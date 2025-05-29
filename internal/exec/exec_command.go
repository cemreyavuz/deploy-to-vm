package exec

import (
	"log"
	"os/exec"
)

type ExecCommand struct {
	cmd *exec.Cmd
}

type ExecCommandInterface interface {
	CombinedOutput() ([]byte, error)
}

func (execCmd *ExecCommand) CombinedOutput() ([]byte, error) {
	if execCmd.cmd == nil {
		log.Println("ExecCommand is nil, cannot execute command")
		return nil, nil
	}
	return execCmd.cmd.CombinedOutput()
}
