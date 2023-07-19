package controllers_test

import (
	"path"
	"testing"

	"github.com/APTrust/dart-runner/constants"
	"github.com/APTrust/dart-runner/core"
	"github.com/APTrust/dart-runner/util"
	"github.com/stretchr/testify/require"
)

func TestBagItProfileCreate(t *testing.T) {

}

func TestBagItProfileDelete(t *testing.T) {

}

func TestBagItProfileEdit(t *testing.T) {

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

}

func TestBagItProfileImport(t *testing.T) {

}

func TestBagItProfileExport(t *testing.T) {

}

func TestBagItProfileSave(t *testing.T) {

}

func TestBagItProfileNewTag(t *testing.T) {

}

func TestBagItProfileEditTag(t *testing.T) {

}

func TestBagItProfileSaveTag(t *testing.T) {

}

func TestBagItProfileDeleteTag(t *testing.T) {

}

func TestBagItProfileNewTagFile(t *testing.T) {

}

func TestBagItProfileCreateTagFile(t *testing.T) {

}

func TestBagItProfileDeleteTagFile(t *testing.T) {

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
