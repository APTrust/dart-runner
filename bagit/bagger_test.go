package bagit_test

import (
	"fmt"
	"os"
	"path"
	"strings"
	"testing"

	"github.com/APTrust/dart-runner/bagit"
	"github.com/APTrust/dart-runner/util"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func getBagger(t *testing.T, bagName, profileName string, files []*util.ExtendedFileInfo) *bagit.Bagger {
	outputPath := path.Join(os.TempDir(), bagName)
	profile, err := loadProfile(profileName)
	require.Nil(t, err)
	require.NotNil(t, profile)
	bagger := bagit.NewBagger(outputPath, profile, files)
	return bagger
}

func TestBaggerRun(t *testing.T) {
	files, err := util.RecursiveFileList(util.PathToTestData())
	require.Nil(t, err)
	bagger := getBagger(t, "bag01.tar", "aptrust-v2.2.json", files)
	// defer os.Remove(bagger.OutputPath)

	bagger.Profile.SetTagValue("bag-info.txt", "Source-Organization", "University of Virginia")
	bagger.Profile.SetTagValue("bag-info.txt", "Bag-Count", "1 of 1")
	bagger.Profile.SetTagValue("bag-info.txt", "Internal-Sender-Description", "My stuff")
	bagger.Profile.SetTagValue("bag-info.txt", "Internal-Sender-Identifier", "my-identifier")

	bagger.Profile.SetTagValue("aptrust-info.txt", "Title", "Test Bag #0001")
	bagger.Profile.SetTagValue("aptrust-info.txt", "Description", "Eloquence and elocution")
	bagger.Profile.SetTagValue("aptrust-info.txt", "Access", "Consortia")
	bagger.Profile.SetTagValue("aptrust-info.txt", "Storage-Option", "Glacier-Deep-OH")

	ok := bagger.Run()
	assert.True(t, ok)
	assert.Empty(t, bagger.Errors)
	fmt.Println(bagger.OutputPath)

	profile, err := loadProfile("aptrust-v2.2.json")
	require.Nil(t, err)
	require.NotNil(t, profile)
	validator, err := bagit.NewValidator(bagger.OutputPath, profile)
	require.Nil(t, err)

	err = validator.ScanBag()
	require.Nil(t, err)

	assert.True(t, validator.Validate())
	assert.Empty(t, validator.Errors)

	xFileInfoList, err := util.RecursiveFileList(util.PathToTestData())
	require.Nil(t, err)
	assertAllPayloadFilesPresent(t, xFileInfoList, validator.PayloadFiles.Files)
}

func assertAllPayloadFilesPresent(t *testing.T, expected []*util.ExtendedFileInfo, actual map[string]*bagit.FileRecord) {
	for _, xFileInfo := range expected {
		if xFileInfo.IsDir() {
			continue
		}
		shortPath := "data" + strings.Replace(xFileInfo.FullPath, util.PathToTestData(), "", 1)
		fileRecord := actual[shortPath]
		assert.NotNil(t, fileRecord, shortPath)
	}
}
