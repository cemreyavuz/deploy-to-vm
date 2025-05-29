package main

import (
	"os"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestStartServer_NoEnvVariableSetForPort(t *testing.T) {
	// Arrange: create a Gin engine
	r := gin.Default()
	os.Unsetenv("DEPLOY_TO_VM_PORT")

	// Act: call startServer with the Gin engine
	err := startServer(r)

	// Assert: check if the error is returned
	assert.Error(t, err)
	assert.Equal(t, "Environment variable DEPLOY_TO_VM_PORT is not set", err.Error())
}
