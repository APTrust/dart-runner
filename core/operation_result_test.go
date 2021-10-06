package core_test

import (
	"testing"

	"github.com/APTrust/dart-runner/core"
	// "github.com/APTrust/dart-runner/util"
	"github.com/stretchr/testify/assert"
)

func TestOperationResult(t *testing.T) {
	result := core.NewOperationResult("op1", "provider1")
	assert.Equal(t, "op1", result.Operation)
	assert.Equal(t, "provider1", result.Provider)
	assert.NotNil(t, result.Errors)
	assert.Empty(t, result.Errors)

	assert.Empty(t, result.Started)
	assert.Empty(t, result.Completed)
	assert.EqualValues(t, 0, result.FileSize)
	assert.Empty(t, result.FileMTime)
	assert.Empty(t, result.RemoteChecksum)
	assert.Empty(t, result.RemoteURL)
	assert.Empty(t, result.Info)
	assert.Empty(t, result.Warning)

	result.Start()
	assert.NotEmpty(t, result.Started)
	assert.Equal(t, 1, result.Attempt)

	assert.True(t, result.WasAttempted())
	assert.False(t, result.WasCompleted())
	assert.False(t, result.Succeeded())

	// Completed without errors
	errors := make(map[string]string)
	result.Finish(errors)
	assert.NotEmpty(t, result.Completed)
	assert.True(t, result.WasCompleted())
	assert.True(t, result.Succeeded())
	assert.False(t, result.HasErrors())

	errors["oops"] = "Something went wrong."
	result.Finish(errors)
	assert.NotEmpty(t, result.Completed)
	assert.True(t, result.WasCompleted())
	assert.True(t, result.HasErrors())
	assert.False(t, result.Succeeded())
}
