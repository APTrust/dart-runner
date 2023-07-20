package controllers_test

import (
	"fmt"
	"html"
	"net/http"
	"net/url"
	"path"
	"strconv"
	"strings"
	"testing"

	"github.com/APTrust/dart-runner/constants"
	"github.com/APTrust/dart-runner/core"
	"github.com/APTrust/dart-runner/util"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBagItProfileCreate(t *testing.T) {
	// POST /profiles/new
	defer core.ClearDartTable()
	saveTestProfiles(t)

	emptyProfile := loadProfile(t, constants.ProfileIDEmpty)

	expected := []string{
		"New profile based on Empty Profile",
		emptyProfile.Description,
		emptyProfile.BagItProfileInfo.BagItProfileIdentifier,
		"sha256",
		"sha512",
		"application/tar",
	}

	data := url.Values{}
	data.Set("BaseProfileID", constants.ProfileIDEmpty)
	settings := PostTestSettings{
		EndpointUrl:          "/profiles/new",
		Params:               data,
		ExpectedResponseCode: http.StatusFound,
		ExpectedContent:      expected,
	}
	DoPostTestWithRedirect(t, settings)

}

func TestBagItProfileDelete(t *testing.T) {
	// PUT /profiles/delete/:id
	// POST /profiles/delete/:id
	defer core.ClearDartTable()
	saveTestProfiles(t)

	// Note that this is an AJAX endpoint that returns
	// JSON on success. The AJAX handler will redirect
	// to the specified location.
	expected := []string{
		`{"location":"/profiles","status":"OK"}`,
	}

	emptyProfile := loadProfile(t, constants.ProfileIDEmpty)
	copyOfProfile := core.BagItProfileClone(emptyProfile)
	require.NoError(t, core.ObjSave(copyOfProfile))

	settings := PostTestSettings{
		EndpointUrl:          fmt.Sprintf("/profiles/delete/%s", copyOfProfile.ID),
		Params:               url.Values{},
		ExpectedResponseCode: http.StatusOK,
		ExpectedContent:      expected,
	}
	DoSimplePostTest(t, settings)
}

func TestBagItProfileEdit(t *testing.T) {
	defer core.ClearDartTable()
	saveTestProfiles(t)

	aptProfile := loadProfile(t, constants.ProfileIDAPTrust)

	// Make sure all elements of the profile appear in this form.
	expected := []string{
		aptProfile.Name,
		aptProfile.Description,
		aptProfile.BagItProfileInfo.BagItProfileIdentifier,
		aptProfile.BagItProfileInfo.BagItProfileVersion,
		aptProfile.BagItProfileInfo.ContactEmail,
		aptProfile.BagItProfileInfo.ContactName,
		aptProfile.BagItProfileInfo.ExternalDescription,
		aptProfile.BagItProfileInfo.SourceOrganization,
		aptProfile.BagItProfileInfo.Version,
		aptProfile.ID,
		aptProfile.Serialization,
	}
	stringLists := [][]string{
		aptProfile.AcceptBagItVersion,
		aptProfile.AcceptSerialization,
		aptProfile.ManifestsAllowed,
		aptProfile.ManifestsRequired,
		aptProfile.TagFilesAllowed,
		aptProfile.TagFilesRequired,
		aptProfile.TagManifestsAllowed,
		aptProfile.TagManifestsRequired,
		aptProfile.TagFileNames(),
	}
	for _, list := range stringLists {
		for _, item := range list {
			if item != "" {
				expected = append(expected, item)
			}
		}
	}
	for _, tag := range aptProfile.Tags {
		expected = append(expected, tag.TagName)
	}

	editURL := fmt.Sprintf("/profiles/edit/%s", aptProfile.ID)
	DoSimpleGetTest(t, editURL, expected)
}

func TestBagItProfileIndex(t *testing.T) {
	defer core.ClearDartTable()
	saveTestProfiles(t)

	expected := []string{
		"BagIt Profiles",
		"Import Profile",
		"New",
		"Name",
		"Description",
		"APTrust",
		"Beyond the Repository",
		"Empty Profile",
		constants.ProfileIDAPTrust,
		constants.ProfileIDBTR,
		constants.ProfileIDEmpty,
	}

	DoSimpleGetTest(t, "/profiles", expected)
}

func TestBagItProfileNew(t *testing.T) {
	defer core.ClearDartTable()
	saveTestProfiles(t)

	expected := []string{
		"New BagIt Profile",
		"Base this profile on...",
		"APTrust",
		"BTR SHA-512",
		"Empty Profile",
		constants.ProfileIDAPTrust,
		constants.ProfileIDBTR,
		constants.ProfileIDEmpty,
	}

	DoSimpleGetTest(t, "/profiles/new", expected)
}

func TestBagItProfileImportStart(t *testing.T) {
	expected := []string{
		"Import profile from",
		"A URL",
		"JSON Data",
		"BagItProfileImport_URL",
		"BagItProfileImport_JsonData",
	}
	DoSimpleGetTest(t, "/profiles/import", expected)
}

func TestBagItProfileImport(t *testing.T) {
	// POST /profiles/import
}

func TestBagItProfileExport(t *testing.T) {
	defer core.ClearDartTable()
	saveTestProfiles(t)

	aptProfile := loadProfile(t, constants.ProfileIDAPTrust)
	standardFormat := aptProfile.ToStandardFormat()

	expected := []string{
		standardFormat.Serialization,
		standardFormat.BagItProfileInfo.BagItProfileIdentifier,
		standardFormat.BagItProfileInfo.BagItProfileVersion,
		standardFormat.BagItProfileInfo.ContactEmail,
		standardFormat.BagItProfileInfo.ContactName,
		standardFormat.BagItProfileInfo.ExternalDescription,
		standardFormat.BagItProfileInfo.SourceOrganization,
		standardFormat.BagItProfileInfo.Version,
	}

	for tagName, tag := range standardFormat.BagInfo {
		expected = append(expected, tagName)
		if tag.Description != "" {
			expected = append(expected, html.EscapeString(tag.Description))
		}
	}

	exportURL := fmt.Sprintf("/profiles/export/%s", constants.ProfileIDAPTrust)
	DoSimpleGetTest(t, exportURL, expected)
}

func TestBagItProfileSave(t *testing.T) {
	// PUT /profiles/edit/:id
	// POST /profiles/edit/:id
	defer core.ClearDartTable()
	saveTestProfiles(t)

	profile := loadProfile(t, constants.ProfileIDEmpty)
	clonedProfile := core.BagItProfileClone(profile)
	require.NoError(t, core.ObjSave(clonedProfile))
	clonedProfile.BagItProfileInfo.BagItProfileIdentifier = "https://example.com/profiles/clone.json"
	clonedProfile.Name = "King Pikachu"
	clonedProfile.AcceptSerialization = []string{
		"application/tar",
		"application/zip",
		"application/gzip",
	}
	clonedProfile.Serialization = constants.SerializationRequired

	settings := PostTestSettings{
		EndpointUrl:              fmt.Sprintf("/profiles/edit/%s", clonedProfile.ID),
		Params:                   profileToFormData(clonedProfile),
		ExpectedResponseCode:     http.StatusFound,
		ExpectedRedirectLocation: "/profiles",
	}
	DoSimplePostTest(t, settings)

	result := core.ObjFind(clonedProfile.ID)
	require.Nil(t, result.Error)
	savedProfile := result.BagItProfile()
	require.NotNil(t, savedProfile)

	// Spot check a few properties, including the ones we changed.
	assert.Equal(t, clonedProfile.AcceptBagItVersion, savedProfile.AcceptBagItVersion)
	assert.Equal(t, clonedProfile.Name, savedProfile.Name)
	assert.Equal(t, clonedProfile.AcceptSerialization, savedProfile.AcceptSerialization)
	assert.Equal(t, clonedProfile.BagItProfileInfo.BagItProfileIdentifier, savedProfile.BagItProfileInfo.BagItProfileIdentifier)
}

func TestBagItProfileNewTag(t *testing.T) {
	defer core.ClearDartTable()
	saveTestProfiles(t)

	expected := []string{
		"In file aptrust-info.txt",
		"Tag Name",
		"Required",
		"Allowed Values",
		"Default Value",
		"Help Text",
	}

	tagURL := fmt.Sprintf("/profiles/new_tag/%s/aptrust-info.txt", constants.ProfileIDAPTrust)
	DoSimpleGetTest(t, tagURL, expected)
}

func TestBagItProfileEditTag(t *testing.T) {
	defer core.ClearDartTable()
	saveTestProfiles(t)

	profile := loadProfile(t, constants.ProfileIDAPTrust)
	tag, err := profile.FirstMatchingTag("TagName", "Storage-Option")
	require.Nil(t, err)

	expected := []string{
		"Storage-Option",
		tag.ID,
		tag.DefaultValue,
		"How do you want this bag to be stored in APTrust?",
	}
	expected = append(expected, tag.Values...)

	tagURL := fmt.Sprintf("/profiles/edit_tag/%s/%s", profile.ID, tag.ID)
	DoSimpleGetTest(t, tagURL, expected)
}

func TestBagItProfileSaveTag(t *testing.T) {
	// POST /profiles/edit_tag/:profile_id/:tag_id
	// PUT  /profiles/edit_tag/:profile_id/:tag_id
}

func TestBagItProfileDeleteTag(t *testing.T) {
	// POST /profiles/delete_tag/:profile_id/:tag_id
	// PUT  /profiles/delete_tag/:profile_id/:tag_id
}

func TestBagItProfileNewTagFile(t *testing.T) {
	defer core.ClearDartTable()
	saveTestProfiles(t)

	expected := []string{
		"Tag File Name",
		`name="Name"`,
		"Cancel",
		"Save",
	}

	tagFileURL := fmt.Sprintf("/profiles/new_tag_file/%s", constants.ProfileIDAPTrust)
	DoSimpleGetTest(t, tagFileURL, expected)
}

func TestBagItProfileCreateTagFile(t *testing.T) {
	// POST /profiles/new_tag_file/:profile_id
}

func TestBagItProfileDeleteTagFile(t *testing.T) {
	// POST /profiles/delete_tag_file/:profile_id
	// PUT  /profiles/delete_tag_file/:profile_id
}

// This loads our standard DART profiles from the profiles
// directory and saves them in the database.
func saveTestProfiles(t *testing.T) {
	profiles := []string{
		"aptrust-v2.2.json",
		"btr-v1.0.json",
		"empty_profile.json",
	}
	for _, filename := range profiles {
		pathToFile := path.Join(util.ProjectRoot(), "profiles", filename)
		data, err := util.ReadFile(pathToFile)
		require.Nil(t, err)
		profile, err := core.BagItProfileFromJSON(string(data))
		require.Nil(t, err)
		err = core.ObjSave(profile)
		require.Nil(t, err)
	}
}

func loadProfile(t *testing.T, profileID string) *core.BagItProfile {
	result := core.ObjFind(profileID)
	require.Nil(t, result.Error)
	profile := result.BagItProfile()
	require.NotNil(t, profile)
	return profile
}

// This converts a BagItProfile to a set of url.Values. Note
// that this does not include tag definition values, as those
// are edited on a separate form.
func profileToFormData(p *core.BagItProfile) url.Values {
	vals := url.Values{}
	vals.Set("AllowFetchTxt", strconv.FormatBool(p.AllowFetchTxt))
	vals.Set("BaseProfileID", p.BaseProfileID)
	vals.Set("Description", p.Description)
	vals.Set("ID", p.ID)
	vals.Set("InfoContactEmail", p.BagItProfileInfo.ContactEmail)
	vals.Set("InfoContactName", p.BagItProfileInfo.ContactName)
	vals.Set("InfoExternalDescription", p.BagItProfileInfo.ExternalDescription)
	vals.Set("InfoIdentifier", p.BagItProfileInfo.BagItProfileIdentifier)
	vals.Set("InfoSourceOrganization", p.BagItProfileInfo.SourceOrganization)
	vals.Set("InfoVersion", p.BagItProfileInfo.Version)
	vals.Set("IsBuiltIn", strconv.FormatBool(p.IsBuiltIn))
	vals.Set("Name", p.Name)
	vals.Set("Serialization", p.Serialization)
	vals.Set("TarDirMustMatchName", strconv.FormatBool(p.TarDirMustMatchName))

	for _, val := range p.AcceptBagItVersion {
		vals.Add("AcceptBagItVersion", val)
	}
	for _, val := range p.AcceptSerialization {
		vals.Add("AcceptSerialization", val)
	}
	for _, val := range p.ManifestsAllowed {
		vals.Add("ManifestsAllowed", val)
	}
	for _, val := range p.ManifestsRequired {
		vals.Add("ManifestsRequired", val)
	}
	for _, val := range p.TagFilesAllowed {
		vals.Add("TagFilesAllowed", val)
	}
	for _, val := range p.TagFilesRequired {
		vals.Add("TagFilesRequired", val)
	}
	for _, val := range p.TagManifestsAllowed {
		vals.Add("TagManifestsAllowed", val)
	}
	for _, val := range p.TagManifestsRequired {
		vals.Add("TagManifestsRequired", val)
	}
	return vals
}

// This converts a tag definition to a set of url.Values.
func tagDefToFormData(tag *core.TagDefinition) url.Values {
	vals := url.Values{}
	vals.Set("ID", tag.ID)
	vals.Set("TagFile", tag.TagFile)
	vals.Set("TagName", tag.TagName)
	vals.Set("DefaultValue", tag.DefaultValue)
	vals.Set("Help", tag.Help)
	vals.Set("Values", strings.Join(tag.Values, "\r\n"))
	vals.Set("Required", strconv.FormatBool(tag.Required))
	return vals
}
