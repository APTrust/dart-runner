package core_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/APTrust/dart-runner/constants"
	"github.com/APTrust/dart-runner/core"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetBagWriter(t *testing.T) {
	// This should get us a file system bag writer
	fsWriter, err := core.GetBagWriter(constants.BagWriterTypeFileSystem, os.TempDir(), []string{})
	require.NoError(t, err)
	require.NotNil(t, fsWriter)
	_, ok := fsWriter.(*core.FileSystemBagWriter)
	assert.True(t, ok)

	// And this should get us a tarred bag writer
	tarWriter, err := core.GetBagWriter(constants.BagWriterTypeTar, filepath.Join(os.TempDir(), "test.tar"), []string{})
	require.NoError(t, err)
	require.NotNil(t, fsWriter)
	_, ok = tarWriter.(*core.TarredBagWriter)
	assert.True(t, ok)

	// And this should get us an error
	noWriter, err := core.GetBagWriter("bad type", filepath.Join(os.TempDir(), "test.tar"), []string{})
	require.Error(t, err)
	require.Nil(t, noWriter)
}
