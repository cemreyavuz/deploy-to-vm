package main

import (
	"os/exec"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestExecCommand_CombinedOutput_NilCmd(t *testing.T) {
	execCmd := &ExecCommand{cmd: nil}
	output, err := execCmd.CombinedOutput()
	assert.Nil(t, output)
	assert.NoError(t, err)
}

func TestExecCommand_CombinedOutput_Success(t *testing.T) {
	execCommand := exec.Command("echo", "test-output")

	output, err := execCommand.CombinedOutput()
	assert.NoError(t, err)
	assert.Contains(t, string(output), "test-output")
}

func TestExecCommand_CombinedOutput_Error(t *testing.T) {
	cmd := exec.Command("cat", "non-existent-file.txt")
	execCommand := &ExecCommand{cmd: cmd}

	output, err := execCommand.CombinedOutput()
	assert.Error(t, err)
	assert.Contains(t, string(output), "cat: non-existent-file.txt: No such file or directory")
}
