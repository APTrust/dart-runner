package bagit_test

import (
	//"fmt"
	"io/ioutil"
	"os"
	"path"
	"testing"

	"github.com/APTrust/dart-runner/bagit"
	"github.com/APTrust/dart-runner/util"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBaggerRun(t *testing.T) {
	tempDir, err := ioutil.TempDir("", "bagger_test")
	require.Nil(t, err)
	outputPath := path.Join(tempDir, "test.tar")
	defer os.Remove(outputPath)
	profilePath := path.Join(util.ProjectRoot(), "profiles", "aptrust-v2.2.json")
	profile, err := bagit.ProfileLoad(profilePath)
	require.Nil(t, err)
	require.NotNil(t, profile)
	files := make([]*util.ExtendedFileInfo, 0)
	bagger := bagit.NewBagger(outputPath, profile, files)
	ok := bagger.Run()
	assert.True(t, ok)
	assert.Empty(t, bagger.Errors)
}
