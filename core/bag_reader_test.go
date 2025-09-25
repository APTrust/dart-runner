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

func TestGetBagReader(t *testing.T) {
	pathToBag := filepath.Join(os.TempDir(), "testbag.tar")
	_, err := os.Create(pathToBag)
	defer os.Remove(pathToBag)
	require.NoError(t, err)

	profile := loadProfile(t, "aptrust-v2.2.json")
	validator, err := core.NewValidator(pathToBag, profile)
	require.Nil(t, err)

	// This should get us a file system bag reader
	fsReader, err := core.GetBagReader(constants.BagWriterTypeFileSystem, validator)
	require.NoError(t, err)
	require.NotNil(t, fsReader)
	_, ok := fsReader.(*core.FileSystemBagReader)
	assert.True(t, ok)

	// And this should get us a tarred bag writer
	tarReader, err := core.GetBagReader(constants.BagReaderTypeTar, validator)
	require.NoError(t, err)
	require.NotNil(t, fsReader)
	_, ok = tarReader.(*core.TarredBagReader)
	assert.True(t, ok)

	// And this should get us an error
	noWriter, err := core.GetBagWriter("bad type", filepath.Join(os.TempDir(), "test.tar"), []string{})
	require.Error(t, err)
	require.Nil(t, noWriter)
}
