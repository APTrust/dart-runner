package core_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/APTrust/dart-runner/constants"
	"github.com/APTrust/dart-runner/core"
	"github.com/APTrust/dart-runner/util"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	APTProfile   = "aptrust-v2.2.json"
	BTRProfile   = "btr-v1.0.json"
	EmptyProfile = "empty_profile.json"
)

func getBagger(t *testing.T, bagName, profileName string, files []*util.ExtendedFileInfo) *core.Bagger {
	outputPath := filepath.Join(os.TempDir(), bagName)
	profile := loadProfile(t, profileName)
	bagger := core.NewBagger(outputPath, profile, files)
	return bagger
}

func TestBaggerRun_APTrust(t *testing.T) {
	testBaggerRun(t, "apt_bag.tar", APTProfile)
}

func TestBaggerRun_BTR(t *testing.T) {
	testBaggerRun(t, "btr_bag.tar", BTRProfile)
}

func testBaggerRun(t *testing.T, bagName, profileName string) {
	files, err := util.RecursiveFileList(util.PathToTestData(), false)
	require.Nil(t, err)
	bagger := getBagger(t, bagName, profileName, files)
	defer os.Remove(bagger.OutputPath)

	setBagInfoTags(bagger.Profile)

	if profileName == APTProfile {
		setAptInfoTags(bagger.Profile)
	}

	// Create the bag and ensure no errors
	ok := bagger.Run()
	assert.True(t, ok)
	assert.Empty(t, bagger.Errors)

	// Validate the bag
	profile := loadProfile(t, profileName)
	validator, err := core.NewValidator(bagger.OutputPath, profile)
	require.Nil(t, err)

	err = validator.ScanBag()
	require.Nil(t, err)

	assert.True(t, validator.Validate())
	assert.Empty(t, validator.Errors)

	// In addition to being valid, make sure the payload
	// has everything we expect.
	xFileInfoList, err := util.RecursiveFileList(util.PathToTestData(), false)
	require.Nil(t, err)
	assertAllPayloadFilesPresent(t, xFileInfoList, validator.PayloadFiles.Files)

	// GetTotalFilesBagged should return the number of payload
	// and non-payload files bagged.
	assert.True(t, bagger.GetTotalFilesBagged() > bagger.PayloadFileCount())

	// Make sure the bagger kept artifacts for payload manifests
	// and tag files.
	if util.StringListContains(bagger.Profile.ManifestsRequired, constants.AlgSha256) {
		assert.True(t, len(bagger.ManifestArtifacts["manifest-sha256.txt"]) > 200)
	} else if util.StringListContains(bagger.Profile.ManifestsRequired, constants.AlgSha512) {
		assert.True(t, len(bagger.ManifestArtifacts["manifest-sha512.txt"]) > 200)
	}
	assert.True(t, len(bagger.TagFileArtifacts["bag-info.txt"]) > 100)

	if strings.Contains(bagger.OutputPath, "apt") {
		assert.Equal(t, filepath.Join(os.TempDir(), "apt_bag_artifacts"), bagger.ArtifactsDir())
	} else if strings.Contains(bagger.OutputPath, "btr") {
		assert.Equal(t, filepath.Join(os.TempDir(), "btr_bag_artifacts"), bagger.ArtifactsDir())
	}
}

// Set tags for bag-info.txt in the profile before we create the bag.
func setBagInfoTags(profile *core.BagItProfile) {
	profile.SetTagValue("bag-info.txt", "Source-Organization", "University of Virginia")
	profile.SetTagValue("bag-info.txt", "Bag-Count", "1 of 1")
	profile.SetTagValue("bag-info.txt", "Internal-Sender-Description", "My stuff")
	profile.SetTagValue("bag-info.txt", "Internal-Sender-Identifier", "my-identifier")
}

// Set tags for aptrust-info.txt in the profile before we create the bag.
func setAptInfoTags(profile *core.BagItProfile) {
	profile.SetTagValue("aptrust-info.txt", "Title", "Test Bag #0001")
	profile.SetTagValue("aptrust-info.txt", "Description", "Eloquence and elocution")
	profile.SetTagValue("aptrust-info.txt", "Access", "Consortia")
	profile.SetTagValue("aptrust-info.txt", "Storage-Option", "Glacier-Deep-OH")
}

// Make sure that all items expected to be in the payload are actually there.
// Just because manifest and payload match doesn't mean that payload contains
// everything we intended to bag.
func assertAllPayloadFilesPresent(t *testing.T, expected []*util.ExtendedFileInfo, actual map[string]*core.FileRecord) {
	for _, xFileInfo := range expected {
		if xFileInfo.IsDir() {
			continue
		}
		shortPath := "data" + strings.Replace(xFileInfo.FullPath, util.ProjectRoot(), "", 1)
		// Fix Windows paths to match bag paths, which only use forward slashes
		shortPath = strings.ReplaceAll(shortPath, "\\", "/")
		fileRecord := actual[shortPath]
		assert.NotNil(t, fileRecord, shortPath)
	}
}
