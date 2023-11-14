package restic

import (
	"os/exec"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRunIntegrityCheck(t *testing.T) {
	execCommandContext = mockExecOutputString("", 0)
	defer func() { execCommandContext = exec.CommandContext }()

	health, err := RunIntegrityCheck(testResticRepository, testResticPassword, &testEnvMap)
	assert.NoError(t, err)

	assert.True(t, health)
}

func TestRunIntegrityCheck_InvalidIndex(t *testing.T) {
	execCommandContext = mockExecOutputFileStderr("testdata/check_invalid_data.txt", 1)
	defer func() { execCommandContext = exec.CommandContext }()

	health, err := RunIntegrityCheck(testResticRepository, testResticPassword, &testEnvMap)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "error: error loading index 66c1a8dc: load <index/66c1a8dcc0>: invalid data returned")
	assert.False(t, health)
}

func TestRunIntegrityCheck_ExitError(t *testing.T) {
	execCommandContext = mockExecOutputString("", 1)
	defer func() { execCommandContext = exec.CommandContext }()

	health, err := RunIntegrityCheck(testResticRepository, testResticPassword, &testEnvMap)
	assert.Error(t, err)
	assert.False(t, health)
}
