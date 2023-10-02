package core_test

import (
	"fmt"
	"path"
	"strings"
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
	pathToFile := path.Join(util.PathToTestData(), "profiles", subDir, filename)
	return loadFileAsBytes(t, pathToFile)
}

func loadDartProfile(t *testing.T, filename string) []byte {
	pathToFile := path.Join(util.ProjectRoot(), "profiles", filename)
	return loadFileAsBytes(t, pathToFile)
}

func loadTestFile(t *testing.T, subDir, filename string) []byte {
	pathToFile := path.Join(util.PathToTestData(), subDir, filename)
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

func TestConvertFromLOCOrdered(t *testing.T) {
	jsonData := loadTestProfile(t, "loc", "SANC-state-profile.json")
	profile, err := core.ConvertProfile(jsonData, "https://example.com/ordered-profile.json")
	require.Nil(t, err)
	require.NotNil(t, profile)

	assert.True(t, strings.HasPrefix(profile.Name, "Profile imported from https://example.com/ordered-profile.json"))
	assert.True(t, strings.HasPrefix(profile.Description, "Profile imported from https://example.com/ordered-profile.json"))

	expectedTagNames := []string{
		"itemNumber",
		"rcNumber",
		"transferringAgencyName",
		"creatingAgencyName",
		"creatingAgencySubdivision",
		"transferringEmployee",
		"receivingInstitution",
		"receivingInstitutionAddress",
		"datesOfRecords (YYYY-MM-DD) - (YYYY-MM-DD)",
		"digitalOriginality",
		"Classification (for Access)",
		"digitalContentStructure",
		"Notes",
	}
	for _, tagName := range expectedTagNames {
		tag, _ := profile.FirstMatchingTag("TagName", tagName)
		assert.NotNil(t, tag, tagName)
	}

	tag, _ := profile.FirstMatchingTag("TagName", "receivingInstitutionAddress")
	assert.Equal(t, "109 E. Jones St. Raleigh, NC 27601", tag.DefaultValue)
	assert.True(t, tag.Required)

	tag, _ = profile.FirstMatchingTag("TagName", "Classification (for Access)")
	assert.Equal(t, "???", tag.DefaultValue)
	assert.Equal(t, 6, len(tag.Values))
	assert.True(t, tag.Required)
}

func TestConvertFromLOCUnordered(t *testing.T) {
	jsonData := loadTestProfile(t, "loc", "unordered-loc-profile.json")
	profile, err := core.ConvertProfile(jsonData, "https://example.com/unordered-profile.json")
	require.Nil(t, err)
	require.NotNil(t, profile)

	assert.True(t, strings.HasPrefix(profile.Name, "Profile imported from https://example.com/unordered-profile.json"))
	assert.True(t, strings.HasPrefix(profile.Description, "Profile imported from https://example.com/unordered-profile.json"))

	expectedTagNames := []string{
		"Send-To-Name",
		"Send-To-Phone",
		"Send-To-Email",
		"External-Identifier",
		"Media-Identifiers",
		"Number-Of-Media-Shipped",
		"Additional-Equipment",
		"Ship-Date",
		"Ship-Method",
		"Ship-Tracking-Number",
		"Ship-Media",
		"Ship-To-Address",
	}

	for _, tagName := range expectedTagNames {
		tag, _ := profile.FirstMatchingTag("TagName", tagName)
		assert.NotNil(t, tag, tagName)
	}

	// Make sure this was overwritten correctly
	tag, _ := profile.FirstMatchingTag("TagName", "External-Identifier")
	assert.True(t, tag.Required)

	tag, _ = profile.FirstMatchingTag("TagName", "Ship-To-Address")
	assert.Equal(t, "World Digital Library, Library of Congress, 101 Independence Ave, SE, Washington, DC 20540 USA", tag.DefaultValue)
	assert.True(t, tag.Required)
}

func TestConvertFromStandardProfile(t *testing.T) {
	stdProfileJson := loadTestProfile(t, "standard", "bagProfileBar.json")
	standardProfile, err := core.StandardProfileFromJson(stdProfileJson)
	require.Nil(t, err)
	require.NotNil(t, standardProfile)

	jsonData := loadTestProfile(t, "standard", "bagProfileBar.json")
	profile, err := core.ConvertProfile(jsonData, "")
	require.Nil(t, err)
	require.NotNil(t, profile)

	assert.Equal(t, standardProfile.AcceptBagItVersion, profile.AcceptBagItVersion)
	assert.Equal(t, standardProfile.AcceptSerialization, profile.AcceptSerialization)

	assert.Equal(t, standardProfile.AllowFetchTxt, profile.AllowFetchTxt)

	assert.Equal(t, standardProfile.BagItProfileInfo.BagItProfileIdentifier, profile.BagItProfileInfo.BagItProfileIdentifier)
	assert.Equal(t, standardProfile.BagItProfileInfo.BagItProfileVersion, profile.BagItProfileInfo.BagItProfileVersion)
	assert.Equal(t, standardProfile.BagItProfileInfo.ContactEmail, profile.BagItProfileInfo.ContactEmail)
	assert.Equal(t, standardProfile.BagItProfileInfo.ContactName, profile.BagItProfileInfo.ContactName)
	assert.Equal(t, standardProfile.BagItProfileInfo.ExternalDescription, profile.BagItProfileInfo.ExternalDescription)
	assert.Equal(t, standardProfile.BagItProfileInfo.SourceOrganization, profile.BagItProfileInfo.SourceOrganization)
	assert.Equal(t, standardProfile.BagItProfileInfo.Version, profile.BagItProfileInfo.Version)

	assert.Equal(t, constants.PreferredAlgsInOrder, profile.ManifestsAllowed)
	assert.Equal(t, standardProfile.ManifestsRequired, profile.ManifestsRequired)
	assert.Equal(t, standardProfile.Serialization, profile.Serialization)
	assert.Equal(t, standardProfile.TagFilesAllowed, profile.TagFilesAllowed)
	assert.Equal(t, standardProfile.TagFilesRequired, profile.TagFilesRequired)
	assert.Equal(t, constants.PreferredAlgsInOrder, profile.TagManifestsAllowed)
	assert.Equal(t, standardProfile.TagManifestsRequired, profile.TagManifestsRequired)

	for tagName, stdTagDef := range standardProfile.BagInfo {
		tag, _ := profile.FirstMatchingTag("TagName", tagName)
		assert.NotNil(t, tag)
		assert.Equal(t, "bag-info.txt", tag.TagFile)
		assert.Equal(t, tagName, tag.TagName)
		assert.Equal(t, stdTagDef.Required, tag.Required)
		assert.Equal(t, stdTagDef.Values, tag.Values)
		help := stdTagDef.Description
		if stdTagDef.Recommended {
			help = fmt.Sprintf("(Recommended) %s", stdTagDef.Description)
		}
		assert.Equal(t, help, tag.Help)
	}
}

func TestConvertFromDartProfile(t *testing.T) {
	pathToFile := path.Join(util.PathToTestData(), "profiles", "multi_manifest.json")
	multiManifestProfile, err := core.BagItProfileLoad(pathToFile)
	require.Nil(t, err)
	require.NotNil(t, multiManifestProfile)

	jsonData := loadTestProfile(t, "", "multi_manifest.json")
	profile, err := core.ConvertProfile(jsonData, "")
	require.Nil(t, err)
	require.NotNil(t, profile)

	assert.Equal(t, multiManifestProfile.ManifestsAllowed, profile.ManifestsAllowed)
	assert.Equal(t, multiManifestProfile.ManifestsRequired, profile.ManifestsRequired)
	assert.Equal(t, multiManifestProfile.TagManifestsAllowed, profile.TagManifestsAllowed)
	assert.Equal(t, multiManifestProfile.TagManifestsRequired, profile.TagManifestsRequired)
	assert.Equal(t, multiManifestProfile.Serialization, profile.Serialization)
	assert.Equal(t, multiManifestProfile.AcceptSerialization, profile.AcceptSerialization)
	assert.Equal(t, multiManifestProfile.AcceptBagItVersion, profile.AcceptBagItVersion)

	for _, tag := range multiManifestProfile.Tags {
		convertedTag, err := profile.FirstMatchingTag("TagName", tag.TagName)
		require.Nil(t, err)
		require.NotNil(t, convertedTag)
	}

	assert.True(t, profile.Validate(), profile.Errors)
}
