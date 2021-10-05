package core_test

import (
	"testing"

	"github.com/APTrust/dart-runner/core"
	"github.com/APTrust/dart-runner/util"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestValidationOperation(t *testing.T) {
	op := core.NewValidationOperation("")
	require.NotNil(t, op)
	assert.Empty(t, op.PathToBag)
	assert.False(t, op.Validate())
	assert.Equal(t, 1, len(op.Errors))
	assert.Equal(t, "You must specify the path to the bag you want to validate.", op.Errors["ValidationOperation.pathToBag"])

	op = core.NewValidationOperation("file-does-not-exist")
	require.NotNil(t, op)
	assert.Equal(t, "file-does-not-exist", op.PathToBag)
	assert.False(t, op.Validate())
	assert.Equal(t, 1, len(op.Errors))
	assert.Equal(t, "The bag to be validated does not exist at file-does-not-exist", op.Errors["ValidationOperation.pathToBag"])

	validPath := util.PathToUnitTestBag("test.edu.btr_good_sha256.tar")
	op = core.NewValidationOperation(validPath)
	require.NotNil(t, op)
	assert.Equal(t, validPath, op.PathToBag)
	assert.True(t, op.Validate())
	assert.Equal(t, 0, len(op.Errors))
}
