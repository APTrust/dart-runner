package core_test

import (
	"database/sql"
	"fmt"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/APTrust/dart-runner/constants"
	"github.com/APTrust/dart-runner/core"
	"github.com/APTrust/dart-runner/util"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func loadProfile(t *testing.T, name string) *core.BagItProfile {
	filename := filepath.Join(util.ProjectRoot(), "profiles", name)
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

	// Make sure basic tag definitions are present
	// for bagit.txt and bag-info.txt.
	bagitTags, err := p.FindMatchingTags("TagFile", "bagit.txt")
	require.Nil(t, err)
	assert.Equal(t, 2, len(bagitTags))
	for _, tag := range bagitTags {
		assert.True(t, tag.Required)
	}

	bagitTags, err = p.FindMatchingTags("TagFile", "bag-info.txt")
	require.Nil(t, err)
	assert.Equal(t, 15, len(bagitTags))
	for _, tag := range bagitTags {
		assert.False(t, tag.Required)
	}

}

// This also implicitly tests BagItProfileFromJson
func TestBagItProfileLoad(t *testing.T) {
	profile := loadProfile(t, "aptrust-v2.2.json")

	// Spot check
	assert.Equal(t, "support@aptrust.org", profile.BagItProfileInfo.ContactEmail)
	assert.Equal(t, 14, len(profile.Tags))
	assert.Equal(t, "BagIt-Version", profile.Tags[0].TagName)
	assert.Equal(t, "Storage-Option", profile.Tags[13].TagName)
	assert.Equal(t, 10, len(profile.Tags[13].Values))

	// Test with bad filename
	_, err := core.BagItProfileLoad("__file_does_not_exist__")
	assert.NotNil(t, err)

	// Test with non-JSON file. This is a tar file.
	filename := filepath.Join(util.PathToUnitTestBag("example.edu.tagsample_good.tar"))
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

func TestGetTagByFQName(t *testing.T) {
	profile := loadProfile(t, "aptrust-v2.2.json")

	tagDef := profile.GetTagByFullyQualifiedName("aptrust-info.txt/Access")
	require.NotNil(t, tagDef)
	assert.Equal(t, "aptrust-info.txt", tagDef.TagFile)
	assert.Equal(t, "Access", tagDef.TagName)

	tagDef = profile.GetTagByFullyQualifiedName("aptrust-info.txt/Tag-Does-Not-Exist")
	assert.Nil(t, tagDef)

	tagDef = profile.GetTagByFullyQualifiedName("malformed tag name")
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

func TestFlagUserAddedTagFiles(t *testing.T) {
	profile := loadProfile(t, "aptrust-v2.2.json")
	profile.FlagUserAddedTagFiles()
	for _, tagDef := range profile.Tags {
		assert.False(t, tagDef.IsUserAddedFile)
	}

	t1 := &core.TagDefinition{
		TagFile:        "custom1.txt",
		TagName:        "Tag-1",
		IsUserAddedTag: true,
	}
	t2 := &core.TagDefinition{
		TagFile:        "custom1.txt",
		TagName:        "Tag-2",
		IsUserAddedTag: true,
	}
	t3 := &core.TagDefinition{
		TagFile:        "custom2.txt",
		TagName:        "Tag-3",
		IsUserAddedTag: true,
	}
	t4 := &core.TagDefinition{
		TagFile:        "custom2.txt",
		TagName:        "Tag-4",
		IsUserAddedTag: false,
	}
	profile.Tags = append(profile.Tags, t1, t2, t3, t4)

	profile.FlagUserAddedTagFiles()
	for _, tagDef := range profile.Tags {
		if tagDef.TagFile == "custom1.txt" {
			assert.True(t, tagDef.IsUserAddedFile)
		} else {
			assert.False(t, tagDef.IsUserAddedFile)
		}
	}

}

func TestCloneProfile(t *testing.T) {
	profile := loadProfile(t, "aptrust-v2.2.json")

	clone := core.BagItProfileClone(profile)
	assert.Equal(t, profile.AcceptBagItVersion, clone.AcceptBagItVersion)
	assert.Equal(t, profile.AcceptSerialization, clone.AcceptSerialization)
	assert.Equal(t, profile.AllowFetchTxt, clone.AllowFetchTxt)
	assert.Equal(t, profile.BaseProfileID, clone.BaseProfileID)
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
	aptProfile := loadProfile(t, "aptrust-v2.2.json")
	btrProfile := loadProfile(t, "btr-v1.0.json")
	emptyProfile := loadProfile(t, "empty_profile.json")

	assert.NoError(t, core.ObjSave(aptProfile))
	assert.NoError(t, core.ObjSave(btrProfile))
	assert.NoError(t, core.ObjSave(emptyProfile))

	// Make sure profile was saved.
	result := core.ObjFind(aptProfile.ID)
	require.Nil(t, result.Error)
	profile := result.BagItProfile()
	require.NotNil(t, profile)
	assert.Equal(t, aptProfile.ID, profile.ID)
	assert.Equal(t, aptProfile.Name, profile.Name)
	assert.Equal(t, aptProfile.TagFilesAllowed, profile.TagFilesAllowed)

	// Make sure order, offset and limit work on list query.
	result = core.ObjList(constants.TypeBagItProfile, "obj_name", 1, 0)
	require.Nil(t, result.Error)
	require.Equal(t, 1, len(result.BagItProfiles))
	assert.Equal(t, profile.ID, result.BagItProfiles[0].ID)

	// Make sure we can get all results.
	result = core.ObjList(constants.TypeBagItProfile, "obj_name", 100, 0)
	require.Nil(t, result.Error)
	require.Equal(t, 3, len(result.BagItProfiles))
	assert.Equal(t, aptProfile.ID, result.BagItProfiles[0].ID)
	assert.Equal(t, btrProfile.ID, result.BagItProfiles[1].ID)
	assert.Equal(t, emptyProfile.ID, result.BagItProfiles[2].ID)

	clonedProfile := core.BagItProfileClone(emptyProfile)
	clonedProfile.Name = fmt.Sprintf("Copy of %s", emptyProfile.Name)
	clonedProfile.IsBuiltIn = false
	assert.NoError(t, core.ObjSave(clonedProfile))

	// Make sure delete works. Should return no error.
	assert.Nil(t, core.ObjDelete(clonedProfile))

	// Make sure the profile was truly deleted.
	result = core.ObjFind(clonedProfile.ID)
	assert.Equal(t, sql.ErrNoRows, result.Error)
	assert.Nil(t, result.BagItProfile())

	// User should not be able to delete APTrust
	// profile because it's a built-in.
	assert.Equal(t, constants.ErrNotDeletable, core.ObjDelete(aptProfile))

}

func TestBagItProfileValidation(t *testing.T) {
	defer core.ClearDartTable()
	p := core.NewBagItProfile()
	p.ID = ""
	p.AcceptBagItVersion = make([]string, 0)
	p.AcceptSerialization = make([]string, 0)
	p.Tags = make([]*core.TagDefinition, 0)
	p.Serialization = ""
	assert.False(t, p.Validate())

	assert.Equal(t, "Profile ID is missing.", p.Errors["ID"])
	assert.Equal(t, "Profile requires a name.", p.Errors["Name"])
	assert.Equal(t, "Profile must accept at least one BagIt version.", p.Errors["AcceptBagItVersion"])
	assert.Equal(t, "Profile must allow at least one manifest algorithm.", p.Errors["ManifestsAllowed"])
	assert.Equal(t, "Profile lacks requirements for bagit.txt tag file.", p.Errors["BagIt"])
	assert.Equal(t, "Profile lacks requirements for bag-info.txt tag file.", p.Errors["BagInfo"])

	expected := fmt.Sprintf("Serialization must be one of: %s.", strings.Join(constants.SerializationOptions, ","))
	assert.Equal(t, expected, p.Errors["Serialization"])

	// No error here unless serialization is required.
	assert.Empty(t, p.Errors["AcceptSerialization"])

	p.Serialization = constants.SerializationRequired
	p.Validate()
	assert.Equal(t, "When serialization is allowed, you must specify at least one serialization format.", p.Errors["AcceptSerialization"])

}

func TestTagsInFile(t *testing.T) {
	profile := loadProfile(t, "aptrust-v2.2.json")
	tags := profile.TagsInFile("bagit.txt")
	require.Equal(t, 2, len(tags))
	assert.Equal(t, "BagIt-Version", tags[0].TagName)
	assert.Equal(t, "Tag-File-Character-Encoding", tags[1].TagName)

	tags = profile.TagsInFile("bag-info.txt")
	require.Equal(t, 8, len(tags))
	assert.Equal(t, "Bag-Count", tags[0].TagName)
	assert.Equal(t, "Source-Organization", tags[7].TagName)

	tags = profile.TagsInFile("aptrust-info.txt")
	require.Equal(t, 4, len(tags))
	assert.Equal(t, "Access", tags[0].TagName)
	assert.Equal(t, "Title", tags[3].TagName)

	tags = profile.TagsInFile("no-such-file.txt")
	require.Equal(t, 0, len(tags))
}

func TestBagItProfileToForm(t *testing.T) {
	profile := loadProfile(t, "aptrust-v2.2.json")
	form := profile.ToForm()

	require.NotNil(t, form.Fields["ID"])
	assert.Equal(t, profile.ID, form.Fields["ID"].Value)

	require.NotNil(t, form.Fields["AcceptBagItVersion"])
	assert.Equal(t, constants.AcceptBagItVersion, form.Fields["AcceptBagItVersion"].Values)

	require.NotNil(t, form.Fields["AcceptSerialization"])
	assert.Equal(t, constants.AcceptSerialization, form.Fields["AcceptSerialization"].Values)

	require.NotNil(t, form.Fields["AllowFetchTxt"])
	assert.Equal(t, strconv.FormatBool(profile.AllowFetchTxt), form.Fields["AllowFetchTxt"].Value)
	assert.Equal(t, core.YesNoChoices(profile.AllowFetchTxt), form.Fields["AllowFetchTxt"].Choices)

	require.NotNil(t, form.Fields["BaseProfileID"])
	assert.Equal(t, profile.BaseProfileID, form.Fields["BaseProfileID"].Value)

	require.NotNil(t, form.Fields["Description"])
	assert.Equal(t, profile.Description, form.Fields["Description"].Value)

	require.NotNil(t, form.Fields["IsBuiltIn"])
	assert.Equal(t, "true", form.Fields["IsBuiltIn"].Value)

	aptrustAlgs := []string{"md5", "sha256"}
	require.NotNil(t, form.Fields["ManifestsAllowed"])
	assert.Equal(t, aptrustAlgs, form.Fields["ManifestsAllowed"].Values)

	require.NotNil(t, form.Fields["ManifestsRequired"])
	assert.Equal(t, 1, len(form.Fields["ManifestsRequired"].Values))
	assert.Equal(t, "md5", form.Fields["ManifestsRequired"].Values[0])

	require.NotNil(t, form.Fields["Name"])
	assert.Equal(t, profile.Name, form.Fields["Name"].Value)
	assert.Equal(t, "readonly", form.Fields["Name"].Attrs["readonly"])

	require.NotNil(t, form.Fields["Serialization"])
	assert.Equal(t, profile.Serialization, form.Fields["Serialization"].Value)
	assert.Equal(t, 4, len(form.Fields["Serialization"].Choices))

	require.NotNil(t, form.Fields["TagFilesAllowed"])
	assert.Equal(t, []string{"*\n"}, form.Fields["TagFilesAllowed"].Values)

	require.NotNil(t, form.Fields["TagFilesRequired"])
	assert.Empty(t, form.Fields["TagFilesRequired"].Values)

	require.NotNil(t, form.Fields["TagManifestsAllowed"])
	assert.Equal(t, aptrustAlgs, form.Fields["TagManifestsAllowed"].Values)

	require.NotNil(t, form.Fields["TagManifestsRequired"])
	assert.Empty(t, form.Fields["TagManifestsRequired"].Values)

	require.NotNil(t, form.Fields["TarDirMustMatchName"])
	assert.Equal(t, strconv.FormatBool(profile.TarDirMustMatchName), form.Fields["TarDirMustMatchName"].Value)
	assert.Equal(t, core.YesNoChoices(profile.TarDirMustMatchName), form.Fields["TarDirMustMatchName"].Choices)

	require.NotNil(t, form.Fields["InfoIdentifier"])
	assert.Equal(t, profile.BagItProfileInfo.BagItProfileIdentifier, form.Fields["InfoIdentifier"].Value)

	require.NotNil(t, form.Fields["InfoContactEmail"])
	assert.Equal(t, profile.BagItProfileInfo.ContactEmail, form.Fields["InfoContactEmail"].Value)

	require.NotNil(t, form.Fields["InfoContactName"])
	assert.Equal(t, profile.BagItProfileInfo.ContactName, form.Fields["InfoContactName"].Value)

	require.NotNil(t, form.Fields["InfoExternalDescription"])
	assert.Equal(t, profile.BagItProfileInfo.ExternalDescription, form.Fields["InfoExternalDescription"].Value)

	require.NotNil(t, form.Fields["InfoSourceOrganization"])
	assert.Equal(t, profile.BagItProfileInfo.SourceOrganization, form.Fields["InfoSourceOrganization"].Value)

	require.NotNil(t, form.Fields["InfoVersion"])
	assert.Equal(t, profile.BagItProfileInfo.Version, form.Fields["InfoVersion"].Value)
}

func TestBagItProfilePersistentObject(t *testing.T) {
	defer core.ClearDartTable()

	profile := loadProfile(t, "aptrust-v2.2.json")
	assert.Equal(t, constants.TypeBagItProfile, profile.ObjType())
	assert.Equal(t, profile.ID, profile.ObjID())
	assert.True(t, util.LooksLikeUUID(profile.ObjID()))
	assert.Equal(t, profile.Name, profile.ObjName())
	assert.Equal(t, "BagItProfile: APTrust", profile.String())
	assert.Empty(t, profile.GetErrors())

	assert.False(t, profile.IsDeletable())
	profile.IsBuiltIn = false
	assert.True(t, profile.IsDeletable())

	profile.Errors = map[string]string{
		"Error 1": "Message 1",
		"Error 2": "Message 2",
	}

	assert.Equal(t, 2, len(profile.GetErrors()))
	assert.Equal(t, "Message 1", profile.GetErrors()["Error 1"])
	assert.Equal(t, "Message 2", profile.GetErrors()["Error 2"])
}

func TestNewBagItProfileCreationForm(t *testing.T) {
	defer core.ClearDartTable()

	profilesToLoad := []string{
		"aptrust-v2.2.json",
		"btr-v1.0.json",
		"empty_profile.json",
	}
	for _, name := range profilesToLoad {
		profile := loadProfile(t, name)
		require.Nil(t, core.ObjSave(profile))
	}

	form, err := core.NewBagItProfileCreationForm()
	require.Nil(t, err)
	field := form.Fields["BaseProfileID"]
	require.NotNil(t, field)
	assert.Equal(t, len(profilesToLoad), len(field.Choices))
}
