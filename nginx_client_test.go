package main

import (
	"errors"
	"testing"

	deploy_to_vm_exec "deploy-to-vm/internal/exec"

	"github.com/stretchr/testify/assert"
)

type MockExecCommand struct {
	CombinedOutputFunc func() ([]byte, error)
}

func (m *MockExecCommand) CombinedOutput() ([]byte, error) {
	return m.CombinedOutputFunc()
}

type MockExecClient struct {
	CombinedOutputFunc func() ([]byte, error)
}

func (m *MockExecClient) Command(name string, arg ...string) deploy_to_vm_exec.ExecCommandInterface {
	return &MockExecCommand{
		CombinedOutputFunc: m.CombinedOutputFunc,
	}
}

type TestableExecCommand struct {
	CombinedOutputFunc func() ([]byte, error)
}

func (t *TestableExecCommand) CombinedOutput() ([]byte, error) {
	return t.CombinedOutputFunc()
}

func TestNginxClient_Reload_Success(t *testing.T) {
	// Arrange: create a mock ExecClient that simulates the command execution
	mockExecClient := &MockExecClient{
		CombinedOutputFunc: func() ([]byte, error) {
			return nil, nil
		},
	}

	// Arrange: create an instance of NginxClient with the mock ExecClient
	nginxClient := NewNginxClient(mockExecClient)

	// Act: call the Reload method
	err := nginxClient.Reload()

	// Assert: check if there was no error
	assert.NoError(t, err, "Expected no error when reloading nginx")
}

func TestNginxClient_Reload_Error(t *testing.T) {
	// Arrange: create a mock ExecClient that simulates the command execution
	mockExecClient := &MockExecClient{
		CombinedOutputFunc: func() ([]byte, error) {
			return nil, errors.New("mock error")
		},
	}

	// Arrange: create an instance of NginxClient with the mock ExecClient
	nginxClient := NewNginxClient(mockExecClient)

	// Act: call the Reload method
	err := nginxClient.Reload()

	// Assert: check if there was an error
	assert.EqualError(t, err, "mock error", "Expected error message to match")
}

func TestNewNginxClient_EmptyExecClient(t *testing.T) {
	// Arrange: create a new NginxClient with nil ExecClient
	nginxClient := NewNginxClient(nil)

	// Assert: check if the ExecClient is not nil and is of type ExecClient
	assert.NotNil(t, nginxClient.ExecClient, "Expected ExecClient to be initialized")
	assert.IsType(t, &deploy_to_vm_exec.ExecClient{}, nginxClient.ExecClient, "Expected ExecClient to be of type ExecClient")
}

func TestNewNginxClient_OverrideExecClient(t *testing.T) {
	// Arrange: create a mock ExecClient
	mockExecClient := &MockExecClient{
		CombinedOutputFunc: func() ([]byte, error) {
			return nil, nil
		},
	}

	// Act: create a new NginxClient with the mock ExecClient
	nginxClient := NewNginxClient(mockExecClient)

	// Assert: check if the ExecClient is the mock ExecClient
	assert.NotNil(t, nginxClient.ExecClient, "Expected ExecClient to be initialized")
	assert.Equal(t, mockExecClient, nginxClient.ExecClient, "Expected ExecClient to be the mock ExecClient")
	assert.IsType(t, &MockExecClient{}, nginxClient.ExecClient, "Expected ExecClient to be of type MockExecClient")
}
