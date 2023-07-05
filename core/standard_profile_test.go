package core_test

import (
	"os"
	"path"
	"testing"

	"github.com/APTrust/dart-runner/core"
	"github.com/APTrust/dart-runner/util"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func loadStandardProfile(t *testing.T, name string) *core.StandardProfile {
	filename := path.Join(util.ProjectRoot(), "testdata", "files", name)
	jsonBytes, err := os.ReadFile(filename)
	require.Nil(t, err)
	profile, err := core.StandardProfileFromJson(jsonBytes)
	require.Nil(t, err)
	require.NotNil(t, profile)
	return profile
}

func TestNewStandardProfile(t *testing.T) {
	profile := core.NewStandardProfile()
	assert.NotNil(t, profile.AcceptBagItVersion)
	assert.NotNil(t, profile.AcceptSerialization)
	assert.NotNil(t, profile.BagInfo)
	assert.NotNil(t, profile.ManifestsAllowed)
	assert.NotNil(t, profile.ManifestsRequired)
	assert.NotNil(t, profile.TagFilesAllowed)
	assert.NotNil(t, profile.TagFilesRequired)
	assert.NotNil(t, profile.TagManifestsAllowed)
	assert.NotNil(t, profile.TagFilesRequired)
}

func TestStandardProfileFromJson(t *testing.T) {
	profile := loadStandardProfile(t, "btr_standard_profile.json")

	versions := []string{
		"0.97",
		"1.0",
	}
	assert.Equal(t, versions, profile.AcceptBagItVersion)

	acceptSerialization := []string{
		"application/zip",
		"application/tar",
		"application/x-tar",
		"application/gzip",
		"application/x-gzip",
		"application/x-7z-compressed",
	}
	assert.Equal(t, acceptSerialization, profile.AcceptSerialization)
	assert.False(t, profile.AllowFetchTxt)
	assert.True(t, profile.BagInfo["Source-Organization"].Required)
	assert.True(t, profile.BagInfo["Bagging-Date"].Required)
	assert.True(t, profile.BagInfo["Payload-Oxum"].Required)
	assert.False(t, profile.BagInfo["Organization-Address"].Required)
	assert.False(t, profile.BagInfo["Contact-Name"].Required)
	assert.False(t, profile.BagInfo["Contact-Phone"].Required)
	assert.False(t, profile.BagInfo["Contact-Email"].Required)
	assert.False(t, profile.BagInfo["External-Description"].Required)
	assert.False(t, profile.BagInfo["External-Identifier"].Required)
	assert.False(t, profile.BagInfo["Bag-Group-Identifier"].Required)
	assert.False(t, profile.BagInfo["Bag-Count"].Required)
	assert.False(t, profile.BagInfo["Bag-Size"].Required)
	assert.False(t, profile.BagInfo["Internal-Sender-Identifier"].Required)
	assert.False(t, profile.BagInfo["Internal-Sender-Description"].Required)
	assert.False(t, profile.BagInfo["Payload-Identifier"].Required)
	assert.False(t, profile.BagInfo["Bag-Producing-Organization"].Required)

	assert.Equal(t, "https://github.com/dpscollaborative/btr_bagit_profile/releases/download/1.0/btr-bagit-profile.json", profile.BagItProfileInfo.BagItProfileIdentifier)
	assert.Equal(t, "1.3.0", profile.BagItProfileInfo.BagItProfileVersion)
	assert.Equal(t, "", profile.BagItProfileInfo.ContactEmail)
	assert.Equal(t, "", profile.BagItProfileInfo.ContactName)
	assert.Equal(t, "Bagit Profile for Consistent Deposit to Distributed Digital Preservation Services", profile.BagItProfileInfo.ExternalDescription)
	assert.Equal(t, "Beyond the Repository Bagit Profile Group", profile.BagItProfileInfo.SourceOrganization)
	assert.Equal(t, "1.0", profile.BagItProfileInfo.Version)

	manifestsAllowed := []string{
		"md5",
		"sha1",
		"sha256",
		"sha512",
	}
	assert.Equal(t, manifestsAllowed, profile.ManifestsAllowed)
	assert.Empty(t, profile.ManifestsRequired)
	assert.Equal(t, manifestsAllowed, profile.TagManifestsAllowed)
	assert.Empty(t, profile.TagManifestsRequired)

	assert.Equal(t, "optional", profile.Serialization)
	assert.Empty(t, profile.TagFilesRequired)
	assert.Equal(t, []string{"*"}, profile.TagFilesAllowed)
}

func TestStandardProfileToDartProfile(t *testing.T) {
	profile := loadStandardProfile(t, "btr_standard_profile.json")
	dartBTRProfile := loadProfile(t, "btr-v1.0-1.3.0.json")

	converted := profile.ToDartProfile()

	assert.Equal(t, dartBTRProfile.AcceptBagItVersion, converted.AcceptBagItVersion)
	assert.Equal(t, dartBTRProfile.AcceptSerialization, converted.AcceptSerialization)
	assert.Equal(t, dartBTRProfile.AllowFetchTxt, converted.AllowFetchTxt)

	assert.Equal(t, dartBTRProfile.BagItProfileInfo.BagItProfileIdentifier, converted.BagItProfileInfo.BagItProfileIdentifier)
	assert.Equal(t, dartBTRProfile.BagItProfileInfo.BagItProfileVersion, converted.BagItProfileInfo.BagItProfileVersion)
	assert.Equal(t, dartBTRProfile.BagItProfileInfo.ExternalDescription, converted.BagItProfileInfo.ExternalDescription)
	assert.Equal(t, dartBTRProfile.BagItProfileInfo.SourceOrganization, converted.BagItProfileInfo.SourceOrganization)
	assert.Equal(t, dartBTRProfile.BagItProfileInfo.Version, converted.BagItProfileInfo.Version)

	assert.Equal(t, "Bagit Profile for Consistent Deposit to Distributed Digital Preservation Services", converted.Description)
	assert.True(t, util.LooksLikeUUID(converted.ID))
	assert.False(t, converted.IsBuiltIn)

	assert.Equal(t, dartBTRProfile.ManifestsAllowed, converted.ManifestsAllowed)
	assert.Equal(t, dartBTRProfile.ManifestsRequired, converted.ManifestsRequired)
	assert.Equal(t, "Beyond the Repository Bagit Profile Group (version 1.0)", converted.Name)

	assert.Equal(t, dartBTRProfile.Serialization, converted.Serialization)
	assert.Equal(t, dartBTRProfile.TagFilesAllowed, converted.TagFilesAllowed)
	assert.Equal(t, dartBTRProfile.TagFilesRequired, converted.TagFilesRequired)
	assert.Equal(t, dartBTRProfile.TagManifestsAllowed, converted.TagManifestsAllowed)
	assert.Equal(t, dartBTRProfile.TagManifestsRequired, converted.TagManifestsRequired)
	assert.False(t, converted.TarDirMustMatchName)

	for name, tag := range profile.BagInfo {
		tagDef, err := converted.FirstMatchingTag("TagName", name)
		require.Nil(t, err, name)
		require.NotNil(t, tagDef, name)
		assert.Equal(t, tag.Required, tagDef.Required, name)
		assert.Equal(t, tag.Required, !tagDef.EmptyOK, name)
		assert.Equal(t, tag.Values, tagDef.Values)
		assert.Contains(t, tagDef.Help, tag.Description, name)
		if tag.Recommended {
			assert.Contains(t, tagDef.Help, "Recommended", name)
		}
	}
}

func TestStandardProfileToJson(t *testing.T) {
	profile := loadStandardProfile(t, "btr_standard_profile.json")
	jsonBytes, err := profile.ToJSON()
	require.Nil(t, err)

	deserializedProfile, err := core.StandardProfileFromJson([]byte(jsonBytes))
	require.Nil(t, err)
	require.NotNil(t, deserializedProfile)

	assert.Equal(t, profile, deserializedProfile)
}
