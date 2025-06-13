package pm2

import (
	deploy_to_vm_exec "deploy-to-vm/internal/exec"
	"errors"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

type MockExecCommand struct {
	CombinedOutputFunc func() ([]byte, error)
}

func (m *MockExecCommand) CombinedOutput() ([]byte, error) {
	return m.CombinedOutputFunc()
}

type MockExecClient struct {
	CommandFunc        func(name string, arg ...string) deploy_to_vm_exec.ExecCommandInterface
	CombinedOutputFunc func() ([]byte, error)
}

func (m *MockExecClient) Command(name string, arg ...string) deploy_to_vm_exec.ExecCommandInterface {
	if m.CommandFunc != nil {
		return m.CommandFunc(name, arg...)
	}

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

func TestPm2Client_Reload_Success(t *testing.T) {
	// Arrange: create a mock ExecClient that simulates the command execution
	var command string
	mockExecClient := &MockExecClient{
		CommandFunc: func(name string, arg ...string) deploy_to_vm_exec.ExecCommandInterface {
			command = name + " " + strings.Join(arg, " ")
			return &MockExecCommand{
				CombinedOutputFunc: func() ([]byte, error) {
					return nil, nil
				},
			}
		},
	}

	// Arrange: create an instance of Pm2Client with the mock ExecClient
	pm2Client := NewPm2Client(mockExecClient)

	// Act: call the Reload method
	err := pm2Client.Reload("test-process")

	// Assert: check if there was no error
	assert.NoError(t, err, "Expected no error when reloading pm2 process")

	// Assert: check if correct command is called
	assert.Equal(t, "pm2 reload test-process", command)
}

func TestPm2Client_Reload_Error(t *testing.T) {
	// Arrange: create a mock ExecClient that simulates the command execution
	mockExecClient := &MockExecClient{
		CombinedOutputFunc: func() ([]byte, error) {
			return nil, errors.New("mock error")
		},
	}

	// Arrange: create an instance of Pm2Client with the mock ExecClient
	pm2Client := NewPm2Client(mockExecClient)

	// Act: call the Reload method
	err := pm2Client.Reload("test-process")

	// Assert: check if there was an error
	assert.EqualError(t, err, "mock error")
}

func TestNewPm2Client_EmptyExecClient(t *testing.T) {
	// Arrange: create a new Pm2Client with nil ExecClient
	pm2Client := NewPm2Client(nil)

	// Assert: check if the ExecClient is not nil and is of type ExecClient
	assert.NotNil(t, pm2Client.ExecClient)
	assert.IsType(t, &deploy_to_vm_exec.ExecClient{}, pm2Client.ExecClient)
}

func TestNewPm2Client_OverrideExecClient(t *testing.T) {
	// Arrange: create a mock ExecClient
	mockExecClient := &MockExecClient{
		CombinedOutputFunc: func() ([]byte, error) {
			return nil, nil
		},
	}

	// Act: create a new Pm2Client with the mock ExecClient
	pm2Client := NewPm2Client(mockExecClient)

	// Assert: check if the ExecClient is the mock ExecClient
	assert.NotNil(t, pm2Client.ExecClient)
	assert.Equal(t, mockExecClient, pm2Client.ExecClient)
	assert.IsType(t, &MockExecClient{}, pm2Client.ExecClient)
}
