package core_test

import (
	"path"
	"testing"

	"github.com/APTrust/dart-runner/constants"
	"github.com/APTrust/dart-runner/core"
	"github.com/APTrust/dart-runner/util"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func loadProfile(t *testing.T, name string) *core.BagItProfile {
	filename := path.Join(util.ProjectRoot(), "profiles", name)
	profile, err := core.BagItProfileLoad(filename)
	assert.Nil(t, err)
	require.NotNil(t, profile)
	return profile
}

func TestNewBagItProfile(t *testing.T) {
	p := core.NewBagItProfile()
	assert.NotNil(t, p)
	assert.NotNil(t, p.AcceptBagItVersion)
	assert.NotNil(t, p.AcceptSerialization)
	assert.False(t, p.AllowFetchTxt)
	assert.NotNil(t, p.BagItProfileInfo)
	assert.NotNil(t, p.Errors)
	assert.NotNil(t, p.ManifestsAllowed)
	assert.NotNil(t, p.ManifestsRequired)
	assert.Equal(t, constants.SerializationOptional, p.Serialization)
	assert.NotNil(t, p.TagManifestsAllowed)
	assert.NotNil(t, p.TagManifestsRequired)
	assert.NotNil(t, p.Tags)
}

// This also implicitly tests BagItProfileFromJson
func TestBagItProfileLoad(t *testing.T) {
	profile := loadProfile(t, "aptrust-v2.2.json")

	// Spot check
	assert.Equal(t, "support@aptrust.org", profile.BagItProfileInfo.ContactEmail)
	assert.Equal(t, 14, len(profile.Tags))
	assert.Equal(t, "BagIt-Version", profile.Tags[0].TagName)
	assert.Equal(t, "Storage-Option", profile.Tags[13].TagName)
	assert.Equal(t, 9, len(profile.Tags[13].Values))

	// Test with bad filename
	_, err := core.BagItProfileLoad("__file_does_not_exist__")
	assert.NotNil(t, err)

	// Test with non-JSON file. This is a tar file.
	filename := path.Join(util.PathToUnitTestBag("example.edu.tagsample_good.tar"))
	_, err = core.BagItProfileLoad(filename)
	assert.NotNil(t, err)

	// Test to/from JSON
	str, err := profile.ToJSON()
	require.Nil(t, err)

	copyOfProfile, err := core.BagItProfileFromJSON(str)
	assert.Nil(t, err)
	require.NotNil(t, copyOfProfile)
	assert.Equal(t, profile, copyOfProfile)
}

func TestGetTagDef(t *testing.T) {
	profile := loadProfile(t, "aptrust-v2.2.json")

	tagDef := profile.GetTagDef("aptrust-info.txt", "Access")
	require.NotNil(t, tagDef)
	assert.Equal(t, "aptrust-info.txt", tagDef.TagFile)
	assert.Equal(t, "Access", tagDef.TagName)

	tagDef = profile.GetTagDef("aptrust-info.txt", "Tag-Does-Not-Exist")
	assert.Nil(t, tagDef)
}

func TestTagFileNames(t *testing.T) {
	aptExpected := []string{
		"aptrust-info.txt",
		"bag-info.txt",
		"bagit.txt",
	}
	apt := loadProfile(t, "aptrust-v2.2.json")
	aptActual := apt.TagFileNames()
	assert.Equal(t, len(aptExpected), len(aptActual))
	for i, _ := range aptExpected {
		assert.Equal(t, aptExpected[i], aptActual[i])
	}

	btrExpected := []string{
		"bag-info.txt",
		"bagit.txt",
	}
	btr := loadProfile(t, "btr-v1.0.json")
	btrActual := btr.TagFileNames()
	assert.Equal(t, len(btrExpected), len(btrActual))
	for i, _ := range btrExpected {
		assert.Equal(t, btrExpected[i], btrActual[i])
	}
}

func TestGetTagFileContents(t *testing.T) {
	apt := loadProfile(t, "aptrust-v2.2.json")

	descriptionTag, err := apt.FirstMatchingTag("TagName", "Description")
	require.Nil(t, err)
	require.NotNil(t, descriptionTag)
	descriptionTag.UserValue = "This here bag belongs to Yosemite Sam!"

	sourceOrgTag, err := apt.FirstMatchingTag("TagName", "Source-Organization")
	require.Nil(t, err)
	require.NotNil(t, sourceOrgTag)
	sourceOrgTag.UserValue = "Warner Bros."

	aptInfoExpected := "Title: \nAccess: Institution\nDescription: This here bag belongs to Yosemite Sam!\nStorage-Option: Standard\n"
	bagInfoExpected := "Source-Organization: Warner Bros.\nBag-Count: \nBagging-Date: \nBagging-Software: \nBag-Group-Identifier: \nInternal-Sender-Description: \nInternal-Sender-Identifier: \nPayload-Oxum: \n"

	aptActual, err := apt.GetTagFileContents("aptrust-info.txt")
	require.Nil(t, err)
	assert.Equal(t, aptInfoExpected, aptActual)

	infoActual, err := apt.GetTagFileContents("bag-info.txt")
	require.Nil(t, err)
	assert.Equal(t, bagInfoExpected, infoActual)
}

func TestMultipleTagValues(t *testing.T) {
	apt := loadProfile(t, "aptrust-v2.2.json")

	rights1 := &core.TagDefinition{
		TagFile:   "bag-info.txt",
		TagName:   "Rights-ID",
		UserValue: "1",
	}
	rights2 := &core.TagDefinition{
		TagFile:   "bag-info.txt",
		TagName:   "Rights-ID",
		UserValue: "2",
	}
	apt.Tags = append(apt.Tags, rights1, rights2)

	sourceOrgTag, err := apt.FirstMatchingTag("TagName", "Source-Organization")
	require.Nil(t, err)
	require.NotNil(t, sourceOrgTag)
	sourceOrgTag.UserValue = "Warner Bros."

	bagInfoExpected := "Source-Organization: Warner Bros.\nBag-Count: \nBagging-Date: \nBagging-Software: \nBag-Group-Identifier: \nInternal-Sender-Description: \nInternal-Sender-Identifier: \nPayload-Oxum: \nRights-ID: 1\nRights-ID: 2\n"

	infoActual, err := apt.GetTagFileContents("bag-info.txt")
	require.Nil(t, err)
	assert.Equal(t, bagInfoExpected, infoActual)
}

func TestSetTagValue(t *testing.T) {
	profile := loadProfile(t, "aptrust-v2.2.json")

	profile.SetTagValue("bag-info.txt", "Payload-Oxum", "12345.2")
	tag := profile.GetTagDef("bag-info.txt", "Payload-Oxum")
	require.NotNil(t, tag)
	assert.Equal(t, "12345.2", tag.GetValue())

	profile.SetTagValue("bag-info.txt", "Flava-Flave", "911")
	tag = profile.GetTagDef("bag-info.txt", "Flava-Flave")
	require.NotNil(t, tag)
	assert.Equal(t, "911", tag.GetValue())
}

func TestCloneProfile(t *testing.T) {
	profile := loadProfile(t, "aptrust-v2.2.json")

	clone := core.BagItProfileClone(profile)
	assert.Equal(t, profile.AllowFetchTxt, clone.AllowFetchTxt)
	assert.Equal(t, profile.Serialization, clone.Serialization)
	assert.ElementsMatch(t, profile.ManifestsAllowed, clone.ManifestsAllowed)
	assert.ElementsMatch(t, profile.ManifestsRequired, clone.ManifestsRequired)
	assert.ElementsMatch(t, profile.TagFilesAllowed, clone.TagFilesAllowed)
	assert.ElementsMatch(t, profile.TagManifestsAllowed, clone.TagManifestsAllowed)
	assert.ElementsMatch(t, profile.TagManifestsRequired, clone.TagManifestsRequired)
	assert.Empty(t, clone.Errors)
	require.Equal(t, len(profile.Tags), len(clone.Tags))
	for i, tag := range profile.Tags {
		clonedTag := clone.Tags[i]
		assert.Equal(t, tag.DefaultValue, clonedTag.DefaultValue)
		assert.Equal(t, tag.EmptyOK, clonedTag.EmptyOK)
		assert.Equal(t, tag.Help, clonedTag.Help)
		assert.Equal(t, tag.ID, clonedTag.ID)
		assert.Equal(t, tag.Required, clonedTag.Required)
		assert.Equal(t, tag.TagFile, clonedTag.TagFile)
		assert.Equal(t, tag.TagName, clonedTag.TagName)
		assert.Equal(t, tag.UserValue, clonedTag.UserValue)
		assert.Equal(t, tag.Values, clonedTag.Values)

		// This is key. Make sure the copy of the tags does not
		// point back to the original tag. This was the source
		// of the race condition that caused bags to get the
		// wrong tag values. See https://trello.com/c/0yoY0FBS
		//
		// Note that we want to assert NotSame here to test that
		// pointer addresses are not the same.
		assert.NotSame(t, tag, clonedTag)
	}
	assert.Equal(t, profile.BagItProfileInfo.BagItProfileVersion, clone.BagItProfileInfo.BagItProfileVersion)
	assert.Equal(t, profile.BagItProfileInfo.BagItProfileIdentifier, clone.BagItProfileInfo.BagItProfileIdentifier)
	assert.Equal(t, profile.BagItProfileInfo.ContactEmail, clone.BagItProfileInfo.ContactEmail)
	assert.Equal(t, profile.BagItProfileInfo.ContactName, clone.BagItProfileInfo.ContactName)
	assert.Equal(t, profile.BagItProfileInfo.ExternalDescription, clone.BagItProfileInfo.ExternalDescription)
	assert.Equal(t, profile.BagItProfileInfo.SourceOrganization, clone.BagItProfileInfo.SourceOrganization)
}

func TestBagItProfilePersistence(t *testing.T) {
	defer core.ClearDartTable()
	// apt := loadProfile(t, "aptrust-v2.2.json")
	// btr := loadProfile(t, "btr-v1.0.json")

	// TODO: Write test
}

func TestBagItProfileValidation(t *testing.T) {
	defer core.ClearDartTable()

	// TODO: Write test
}

func TestBagItProfileToForm(t *testing.T) {

	// TODO: Write test
}

func TestBagItProfilePersistentObject(t *testing.T) {

	// TODO: Write test
}
