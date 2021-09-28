package bagit_test

import (
	"path"
	"testing"

	"github.com/APTrust/dart-runner/bagit"
	"github.com/APTrust/dart-runner/util"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// This also implicitly tests BagItProfileFromJson
func TestBagItProfileLoad(t *testing.T) {
	filename := path.Join(util.ProjectRoot(), "profiles", "aptrust-v2.2.json")
	profile, err := bagit.ProfileLoad(filename)
	assert.Nil(t, err)
	require.NotNil(t, profile)

	// Spot check
	assert.Equal(t, "support@aptrust.org", profile.BagItProfileInfo.ContactEmail)
	assert.Equal(t, 14, len(profile.Tags))
	assert.Equal(t, "BagIt-Version", profile.Tags[0].TagName)
	assert.Equal(t, "Storage-Option", profile.Tags[13].TagName)
	assert.Equal(t, 9, len(profile.Tags[13].Values))

	// Test with bad filename
	_, err = bagit.ProfileLoad("__file_does_not_exist__")
	assert.NotNil(t, err)

	// Test with non-JSON file. This is a tar file.
	filename = path.Join(util.PathToUnitTestBag("example.edu.tagsample_good.tar"))
	_, err = bagit.ProfileLoad(filename)
	assert.NotNil(t, err)
}

func TestGetTagDef(t *testing.T) {
	filename := path.Join(util.ProjectRoot(), "profiles", "aptrust-v2.2.json")
	profile, err := bagit.ProfileLoad(filename)
	assert.Nil(t, err)
	require.NotNil(t, profile)

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
	aptPath := path.Join(util.ProjectRoot(), "profiles", "aptrust-v2.2.json")
	apt, err := bagit.ProfileLoad(aptPath)
	require.Nil(t, err)
	aptActual := apt.TagFileNames()
	assert.Equal(t, len(aptExpected), len(aptActual))
	for i, _ := range aptExpected {
		assert.Equal(t, aptExpected[i], aptActual[i])
	}

	btrExpected := []string{
		"bag-info.txt",
		"bagit.txt",
	}
	btrPath := path.Join(util.ProjectRoot(), "profiles", "btr-v1.0.json")
	btr, err := bagit.ProfileLoad(btrPath)
	require.Nil(t, err)
	btrActual := btr.TagFileNames()
	assert.Equal(t, len(btrExpected), len(btrActual))
	for i, _ := range btrExpected {
		assert.Equal(t, btrExpected[i], btrActual[i])
	}
}

func TestGetTagFileContents(t *testing.T) {
	aptPath := path.Join(util.ProjectRoot(), "profiles", "aptrust-v2.2.json")
	apt, err := bagit.ProfileLoad(aptPath)
	require.Nil(t, err)

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

func TestSetTagValue(t *testing.T) {
	aptPath := path.Join(util.ProjectRoot(), "profiles", "aptrust-v2.2.json")
	profile, err := bagit.ProfileLoad(aptPath)
	require.Nil(t, err)

	profile.SetTagValue("bag-info.txt", "Payload-Oxum", "12345.2")
	tag := profile.GetTagDef("bag-info.txt", "Payload-Oxum")
	require.NotNil(t, tag)
	assert.Equal(t, "12345.2", tag.GetValue())

	profile.SetTagValue("bag-info.txt", "Flava-Flave", "911")
	tag = profile.GetTagDef("bag-info.txt", "Flava-Flave")
	require.NotNil(t, tag)
	assert.Equal(t, "911", tag.GetValue())
}
