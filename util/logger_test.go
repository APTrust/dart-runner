package util_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/APTrust/dart-runner/util"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRotateCurrentLog(t *testing.T) {
	tempDir := filepath.Join(os.TempDir(), "dart_log_test")
	require.NoError(t, os.Mkdir(tempDir, 0744))
	defer os.RemoveAll(tempDir)

	tempLog, err := os.Create(filepath.Join(tempDir, "dart.log"))
	require.Nil(t, err)
	tempLog.Write([]byte("Lorem ipsum pipsum squeakem... 12345678"))
	require.NoError(t, tempLog.Close())

	// Should return nil because file is < 500 bytes.
	newFileName, err := util.RotateCurrentLog(tempDir, tempLog.Name(), "dart", int64(500))
	require.Nil(t, err)
	assert.Empty(t, newFileName)

	// Should return new file name because file is > 5 bytes.
	newFileName, err = util.RotateCurrentLog(tempDir, tempLog.Name(), "dart", int64(5))
	require.Nil(t, err)
	assert.Equal(t, filepath.Join(tempDir, "dart_0001.log"), newFileName)

	// At this point, our original log file is gone.
	// Copy it back, so now the dir has dart.log and dart_0001.log
	_, err = util.CopyFile(tempLog.Name(), newFileName)
	require.NoError(t, err)

	// Should return nil because file is < 500 bytes.
	newFileName, err = util.RotateCurrentLog(tempDir, tempLog.Name(), "dart", int64(500))
	require.Nil(t, err)
	assert.Empty(t, newFileName)

	// Should return new file name because file is > 5 bytes.
	// And the new file should have number 0002.
	newFileName, err = util.RotateCurrentLog(tempDir, tempLog.Name(), "dart", int64(5))
	require.Nil(t, err)
	assert.Equal(t, filepath.Join(tempDir, "dart_0002.log"), newFileName)

}
