package util_test

import (
	"io"
	"io/ioutil"
	"os"
	"runtime"
	"testing"

	"github.com/APTrust/dart-runner/util"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func writeTestFile(t *testing.T) string {
	tempfile, err := ioutil.TempFile("", "test_file")
	require.Nil(t, err)
	defer tempfile.Close()
	_, err = io.WriteString(tempfile, "Le grill? What the hell is that?")
	require.Nil(t, err)
	return tempfile.Name()
}

// GetOwnerAndGroup should fill in the Uid and Gid fields of
// the tar header on Posix systems. On windows, it won't fill in
// anything, but it should not cause any errors.
func TestGetOwnerAndGroup(t *testing.T) {
	testFilePath := writeTestFile(t)
	defer os.Remove(testFilePath)
	fileInfo, err := os.Stat(testFilePath)
	require.Nil(t, err)
	require.NotNil(t, fileInfo)

	xInfo := util.NewExtendedFileInfo(testFilePath, fileInfo)
	uid, gid := xInfo.OwnerAndGroup()
	if runtime.GOOS == "linux" || runtime.GOOS == "darwin" ||
		runtime.GOOS == "unix" || runtime.GOOS == "bsd" {
		// We just wrote these files, so their uid and gid
		// should match ours.
		assert.EqualValues(t, os.Getuid(), uid)
		assert.EqualValues(t, os.Getgid(), gid)
	}
}
