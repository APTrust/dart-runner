package core_test

import (
	"testing"

	"github.com/APTrust/dart-runner/constants"
	"github.com/APTrust/dart-runner/core"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewBagItProfileImport(t *testing.T) {
	bpi := core.NewBagItProfileImport(constants.ImportSourceUrl, "https://example.com/profile.json", nil)
	assert.Equal(t, "https://example.com/profile.json", bpi.URL)
	assert.Equal(t, constants.ImportSourceUrl, bpi.ImportSource)
	assert.Nil(t, bpi.JsonData)

	jsonData := []byte(`{"data":"foo"}`)
	bpi = core.NewBagItProfileImport(constants.ImportSourceJson, "", jsonData)
	assert.Equal(t, "", bpi.URL)
	assert.Equal(t, constants.ImportSourceJson, bpi.ImportSource)
	assert.Equal(t, jsonData, bpi.JsonData)

}

func TestBagItProfileImportValidate(t *testing.T) {
	bpi := core.NewBagItProfileImport(constants.ImportSourceUrl, "", nil)
	assert.False(t, bpi.Validate())
	assert.Equal(t, 1, len(bpi.Errors))
	assert.Equal(t, "Please specify a valid URL.", bpi.Errors["URL"])

	bpi = core.NewBagItProfileImport("", "", nil)
	assert.False(t, bpi.Validate())
	assert.Equal(t, 1, len(bpi.Errors))
	assert.Equal(t, "Please specify either URL or JSON as the import source.", bpi.Errors["ImportSource"])

	bpi = core.NewBagItProfileImport(constants.ImportSourceJson, "", nil)
	assert.False(t, bpi.Validate())
	assert.Equal(t, 1, len(bpi.Errors))
	assert.Equal(t, "Please enter JSON to be imported.", bpi.Errors["JsonData"])

	bpi = core.NewBagItProfileImport(constants.ImportSourceUrl, "https://example.com/profile.json", nil)
	assert.True(t, bpi.Validate())

	bpi = core.NewBagItProfileImport(constants.ImportSourceJson, "", []byte(`{"data":"foo"}`))
	assert.True(t, bpi.Validate())
}

func TestBagItProfileImportConvert(t *testing.T) {
	testBPIConvertFromURLs(t)
	testBPIConvertFromJson(t)
}

func testBPIConvertFromURLs(t *testing.T) {
	//
	// Note that this test requires an internet connection, as
	// we're converting profiles hosted on GitHub.
	//
	btrUrl := "https://raw.githubusercontent.com/dpscollaborative/btr_bagit_profile/master/btr-bagit-profile.json"
	bpi := core.NewBagItProfileImport(constants.ImportSourceUrl, btrUrl, nil)
	profile, err := bpi.Convert()
	require.Nil(t, err)
	require.NotNil(t, profile)
	assert.True(t, profile.Validate(), profile.Errors)

	// Do a quick spot test here. We test the conversion process
	// more thoroughly in bagit_profile_conversions_test.go
	manifestsAllowed := []string{
		"md5",
		"sha1",
		"sha256",
		"sha512",
	}
	assert.Equal(t, "Bagit Profile for Consistent Deposit to Distributed Digital Preservation Services", profile.BagItProfileInfo.ExternalDescription)
	assert.Equal(t, manifestsAllowed, profile.ManifestsAllowed)
	tag, err := profile.FirstMatchingTag("TagName", "Bag-Producing-Organization")
	require.Nil(t, err)
	require.NotNil(t, tag)
	assert.Equal(t, "(Recommended) Can be the same as source_organization recommended when not the same as source", tag.Help)

	locUrl := "https://raw.githubusercontent.com/LibraryOfCongress/bagger/master/bagger-business/src/main/resources/gov/loc/repository/bagger/profiles/SANC-local-profile.json"
	bpi = core.NewBagItProfileImport(constants.ImportSourceUrl, locUrl, nil)
	profile, err = bpi.Convert()
	require.Nil(t, err)
	require.NotNil(t, profile)
	tag, err = profile.FirstMatchingTag("TagName", "receivingInstitutionAddress")
	require.Nil(t, err)
	require.NotNil(t, tag)
	assert.Equal(t, "109 E. Jones St. Raleigh, NC 27601", tag.DefaultValue)
	assert.True(t, profile.Validate(), profile.Errors)
}

func testBPIConvertFromJson(t *testing.T) {
	locUnorderedJson := loadTestProfile(t, "loc", "unordered-loc-profile.json")
	standardJson := loadTestProfile(t, "standard", "bagProfileFoo.json")
	btrJson := loadTestProfile(t, "", "btr_standard_profile.json")

	jsonBlobs := [][]byte{
		locUnorderedJson,
		standardJson,
		btrJson,
	}

	for _, jsonBlob := range jsonBlobs {
		bpi := core.NewBagItProfileImport(constants.ImportSourceJson, "", jsonBlob)
		profile, err := bpi.Convert()
		require.Nil(t, err)
		require.NotNil(t, profile)
		assert.True(t, profile.Validate(), profile.Errors)
	}
}

func TestBagItProfileImportToForm(t *testing.T) {
	sampleUrl := "https://example.com/profile.json"
	sampleJson := []byte(`{"id": 1234, "name": "sample"}`)
	bpi := core.NewBagItProfileImport(constants.ImportSourceUrl, sampleUrl, sampleJson)
	form := bpi.ToForm()
	assert.Equal(t, 3, len(form.Fields))
	assert.Equal(t, 3, len(form.Fields["ImportSource"].Choices))
	assert.Equal(t, bpi.ImportSource, form.Fields["ImportSource"].Value)
	assert.Equal(t, bpi.URL, form.Fields["URL"].Value)
	assert.Equal(t, string(bpi.JsonData), form.Fields["JsonData"].Value)

}
