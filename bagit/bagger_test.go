package bagit_test

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"testing"

	"github.com/APTrust/dart-runner/bagit"
	"github.com/APTrust/dart-runner/util"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func getBagger(t *testing.T, bagName, profileName string, files []*util.ExtendedFileInfo) *bagit.Bagger {
	tempDir, err := ioutil.TempDir("", "bagger_test")
	require.Nil(t, err)
	outputPath := path.Join(tempDir, bagName)
	profilePath := path.Join(util.ProjectRoot(), "profiles", profileName)
	profile, err := bagit.ProfileLoad(profilePath)
	require.Nil(t, err)
	require.NotNil(t, profile)
	bagger := bagit.NewBagger(outputPath, profile, files)
	return bagger
}

func TestBaggerRun(t *testing.T) {
	files, err := util.RecursiveFileList(util.PathToTestData())
	require.Nil(t, err)
	bagger := getBagger(t, "bag01.tar", "aptrust-v2.2.json", files)
	defer os.Remove(bagger.OutputPath)
	ok := bagger.Run()
	assert.True(t, ok)
	assert.Empty(t, bagger.Errors)
	fmt.Println(bagger.OutputPath)
	assert.True(t, false)
}
