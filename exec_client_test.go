package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestExecClient_Command_ReturnsExecCommand(t *testing.T) {
	client := &ExecClient{}
	execCommand := client.Command("echo", "hello")

	assert.NotNil(t, execCommand, "Expected execCommand to be non-nil")
	assert.NotNil(t, execCommand.cmd, "Expected underlying execCommand.Cmd to be non-nil")
}

func TestExecClient_CombinedOutput_Success(t *testing.T) {
	client := &ExecClient{}
	execCommand := client.Command("echo", "test-output")

	output, err := execCommand.CombinedOutput()
	assert.NoError(t, err)
	assert.Contains(t, string(output), "test-output")
}

func TestExecClient_CombinedOutput_Error(t *testing.T) {
	client := &ExecClient{}
	execCommand := client.Command("cat", "non-existent-file.txt")

	output, err := execCommand.CombinedOutput()
	assert.Error(t, err)
	assert.Contains(t, string(output), "cat: non-existent-file.txt: No such file or directory")
}
