package core_test

import (
	"fmt"
	"path"
	"testing"

	"github.com/APTrust/dart-runner/constants"
	"github.com/APTrust/dart-runner/core"
	"github.com/APTrust/dart-runner/util"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBagItProfileToStandardFormat(t *testing.T) {
	btrDartProfile := loadProfile(t, "btr-v1.0-1.3.0.json")
	btrStandardProfile := loadStandardProfile(t, "btr_standard_profile.json")
	convertedToStandard := btrDartProfile.ToStandardFormat()

	assert.Equal(t, btrStandardProfile.AcceptBagItVersion, convertedToStandard.AcceptBagItVersion)
	assert.Equal(t, btrStandardProfile.AcceptSerialization, convertedToStandard.AcceptSerialization)
	assert.Equal(t, btrStandardProfile.AllowFetchTxt, convertedToStandard.AllowFetchTxt)
	assert.Equal(t, btrStandardProfile.ManifestsAllowed, convertedToStandard.ManifestsAllowed)
	assert.Equal(t, btrStandardProfile.ManifestsRequired, convertedToStandard.ManifestsRequired)
	assert.Equal(t, btrStandardProfile.Serialization, convertedToStandard.Serialization)
	assert.Equal(t, btrStandardProfile.TagFilesAllowed, convertedToStandard.TagFilesAllowed)
	assert.Equal(t, btrStandardProfile.TagFilesRequired, convertedToStandard.TagFilesRequired)
	assert.Equal(t, btrStandardProfile.TagManifestsAllowed, convertedToStandard.TagManifestsAllowed)
	assert.Equal(t, btrStandardProfile.TagManifestsRequired, convertedToStandard.TagManifestsRequired)

	assert.Equal(t, btrStandardProfile.BagItProfileInfo.BagItProfileIdentifier, convertedToStandard.BagItProfileInfo.BagItProfileIdentifier)
	assert.Equal(t, btrStandardProfile.BagItProfileInfo.BagItProfileVersion, convertedToStandard.BagItProfileInfo.BagItProfileVersion)
	assert.Equal(t, btrStandardProfile.BagItProfileInfo.ContactEmail, convertedToStandard.BagItProfileInfo.ContactEmail)
	assert.Equal(t, btrStandardProfile.BagItProfileInfo.ContactName, convertedToStandard.BagItProfileInfo.ContactName)
	assert.Equal(t, btrStandardProfile.BagItProfileInfo.ExternalDescription, convertedToStandard.BagItProfileInfo.ExternalDescription)
	assert.Equal(t, btrStandardProfile.BagItProfileInfo.SourceOrganization, convertedToStandard.BagItProfileInfo.SourceOrganization)
	assert.Equal(t, btrStandardProfile.BagItProfileInfo.Version, convertedToStandard.BagItProfileInfo.Version)

	assert.Equal(t, len(btrStandardProfile.BagInfo), len(convertedToStandard.BagInfo))

	for name, tag := range convertedToStandard.BagInfo {
		expectedTag := btrStandardProfile.BagInfo[name]
		require.NotNil(t, expectedTag)
		if expectedTag.Recommended {
			assert.Equal(t, fmt.Sprintf("(Recommended) %s", expectedTag.Description), tag.Description)
		} else {
			assert.Equal(t, expectedTag.Description, tag.Description)
		}
		assert.Equal(t, expectedTag.Recommended, tag.Recommended)
		assert.Equal(t, expectedTag.Required, tag.Required)
		assert.Equal(t, expectedTag.Values, tag.Values)
	}
}

func loadFileAsBytes(t *testing.T, pathToFile string) []byte {
	data, err := util.ReadFile(pathToFile)
	require.Nil(t, err)
	return data
}

func loadTestProfile(t *testing.T, subDir, filename string) []byte {
	pathToFile := path.Join(util.ProjectRoot(), "testdata", "profiles", subDir, filename)
	return loadFileAsBytes(t, pathToFile)
}

func loadDartProfile(t *testing.T, filename string) []byte {
	pathToFile := path.Join(util.ProjectRoot(), "profiles", filename)
	return loadFileAsBytes(t, pathToFile)
}

func loadTestFile(t *testing.T, subDir, filename string) []byte {
	pathToFile := path.Join(util.ProjectRoot(), "testdata", subDir, filename)
	return loadFileAsBytes(t, pathToFile)
}

func TestGuessProfileTypeFromJson(t *testing.T) {

	// Standard profile types
	jsonData := loadTestProfile(t, "standard", "bagProfileBar.json")
	testGuessProfileTypeFromJson(t, jsonData, constants.ProfileTypeStandard)

	jsonData = loadTestProfile(t, "standard", "bagProfileFoo.json")
	testGuessProfileTypeFromJson(t, jsonData, constants.ProfileTypeStandard)

	jsonData = loadTestProfile(t, "", "btr_standard_profile.json")
	testGuessProfileTypeFromJson(t, jsonData, constants.ProfileTypeStandard)

	// Library of Congress profile types
	jsonData = loadTestProfile(t, "loc", "SANC-state-profile.json")
	testGuessProfileTypeFromJson(t, jsonData, constants.ProfileTypeLOCOrdered)

	jsonData = loadTestProfile(t, "loc", "unordered-loc-profile.json")
	testGuessProfileTypeFromJson(t, jsonData, constants.ProfileTypeLOCUnordered)

	// Dart types
	dartProfiles := []string{
		"aptrust-v2.2.json",
		"btr-v1.0-1.3.0.json",
		"btr-v1.0.json",
		"empty_profile.json",
	}
	for _, profile := range dartProfiles {
		jsonData = loadDartProfile(t, profile)
		testGuessProfileTypeFromJson(t, jsonData, constants.ProfileTypeDart)
	}

	// DART profile that requires multiple manifests
	jsonData = loadTestProfile(t, "", "multi_manifest.json")
	testGuessProfileTypeFromJson(t, jsonData, constants.ProfileTypeDart)

	// This profile is invalid because it fails to specify manifest info,
	// but it IS in DART format, and that's all we're testing for here.
	jsonData = loadTestProfile(t, "", "invalid_profile.json")
	testGuessProfileTypeFromJson(t, jsonData, constants.ProfileTypeDart)

	// Invalid/unknown types
	// Not even a BagIt profile. We want to make sure this doesn't blow up.
	jsonData = loadTestFile(t, "files", "sample_job.json")
	testGuessProfileTypeFromJson(t, jsonData, constants.ProfileTypeUnknown)

}

func testGuessProfileTypeFromJson(t *testing.T, jsonData []byte, expectedType string) {
	profileType, err := core.GuessProfileTypeFromJson(jsonData)
	require.Nil(t, err)
	assert.Equal(t, expectedType, profileType)
}
