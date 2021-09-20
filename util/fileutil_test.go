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
	fmt.Println(f)
	assert.True(t, util.FileExists(f))
	assert.True(t, util.FileExists(util.ProjectRoot()))
	assert.False(t, util.FileExists("NonExistentFile.xyz"))
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
