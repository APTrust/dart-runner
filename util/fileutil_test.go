package util_test

import (
	"fmt"
	"os"
	"path"
	"strings"
	"testing"

	"github.com/APTrust/dart-runner/util"
	"github.com/stretchr/testify/assert"
)

func TestFileExists(t *testing.T) {
	f := util.PathToUnitTestBag("example.edu.sample_good.tar")
	assert.True(t, util.FileExists(f))
	assert.True(t, util.FileExists(util.ProjectRoot()))
	assert.False(t, util.FileExists("NonExistentFile.xyz"))
}

func TestIsDir(t *testing.T) {
	f := util.PathToUnitTestBag("example.edu.sample_good.tar")
	assert.False(t, util.IsDirectory(f))
	assert.False(t, util.IsDirectory("NonExistentFile.xyz"))
	assert.True(t, util.IsDirectory(util.ProjectRoot()))
}

func TestExpandTilde(t *testing.T) {
	expanded, err := util.ExpandTilde("~/tmp")
	assert.Nil(t, err)
	assert.True(t, len(expanded) > 6)
	assert.True(t, strings.HasSuffix(expanded, "tmp"))

	expanded, err = util.ExpandTilde("/nothing/to/expand")
	assert.Nil(t, err)
	assert.Equal(t, "/nothing/to/expand", expanded)
}

func TestLooksSafeToDelete(t *testing.T) {
	assert.True(t, util.LooksSafeToDelete("/mnt/apt/data/some_dir", 15, 3))
	assert.False(t, util.LooksSafeToDelete("/usr/local", 12, 3))
}

func TestCopyFile(t *testing.T) {
	src := util.PathToUnitTestBag("example.edu.sample_good.tar")
	dest := path.Join(util.ProjectRoot(), "example.edu.sample_good.tar")
	fmt.Println(src, dest)
	_, err := util.CopyFile(dest, src)
	defer os.Remove(dest)
	assert.Nil(t, err)
}

func TestHasValidExtensionForMimeType(t *testing.T) {
	okFiles := map[string]string{
		"file.7z":     "application/x-7z-compressed",
		"file.7Z":     "application/x-7z-compressed",
		"file.tar":    "application/tar",
		"file2.tar":   "application/x-tar",
		"file.zip":    "application/zip",
		"file.gzip":   "application/gzip",
		"file.gz":     "application/gzip",
		"file.rar":    "application/x-rar-compressed",
		"file.tgz":    "application/tar+gzip",
		"file.tar.gz": "application/tar+gzip",
	}
	badFiles := map[string]string{
		"file.7z":  "application/tar",
		"file.tar": "application/x-7z-compressed",
		"file.zip": "application/gzip",
	}
	errFiles := map[string]string{
		"file.7z":  "application/binary",
		"file.tar": "application/kompressed",
	}

	for filename, mimeType := range okFiles {
		ok, err := util.HasValidExtensionForMimeType(filename, mimeType)
		assert.True(t, ok)
		assert.Nil(t, err)
	}

	for filename, mimeType := range badFiles {
		ok, err := util.HasValidExtensionForMimeType(filename, mimeType)
		assert.False(t, ok)
		assert.Nil(t, err)
	}

	for filename, mimeType := range errFiles {
		ok, err := util.HasValidExtensionForMimeType(filename, mimeType)
		assert.False(t, ok)
		assert.NotNil(t, err)
	}
}
