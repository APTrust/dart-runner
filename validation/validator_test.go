package validation_test

import (
	"fmt"
	"testing"

	"github.com/APTrust/dart-runner/util"
	"github.com/APTrust/dart-runner/validation"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func getValidator(t *testing.T, bag, profile string) *validation.Validator {
	bagItProfile, err := loadProfile(profile)
	require.Nil(t, err)

	pathToBag := util.PathToUnitTestBag(bag)
	validator, err := validation.NewValidator(pathToBag, bagItProfile)
	require.Nil(t, err)
	return validator
}

func TestValidator_ScanBag(t *testing.T) {
	expected, err := loadValidatorFromJson("tagsample_good_metadata.json")
	require.Nil(t, err)
	require.NotNil(t, expected)

	v := getValidator(t, "example.edu.tagsample_good.tar", "aptrust-v2.2.json")

	err = v.ScanBag()
	require.Nil(t, err)

	// The scanner should have loaded the validator with
	// the same info as in our JSON file (except PathToBag,
	// which will differ on each machine).
	tarReaderTestFileMaps(t, expected.PayloadFiles, v.PayloadFiles)
	tarReaderTestFileMaps(t, expected.PayloadManifests, v.PayloadManifests)
	tarReaderTestFileMaps(t, expected.TagFiles, v.TagFiles)
	tarReaderTestFileMaps(t, expected.TagManifests, v.TagManifests)

	tarReaderTestTags(t, expected.Tags, v.Tags)
}

func TestValidator_ValidateBasic(t *testing.T) {
	goodBags := []string{
		"example.edu.sample_good.tar",
		"example.edu.tagsample_good.tar",
	}
	profiles := []string{
		"aptrust-v2.2.json",
		"empty_profile.json",
	}
	for _, bag := range goodBags {
		for _, profile := range profiles {
			v := getValidator(t, bag, profile)
			// APTrust profile actually requires an md5 manifest,
			// which is not part of this bag. Let's test without
			// that requirement.
			v.Profile.ManifestsRequired = []string{}

			message := fmt.Sprintf("Bag %s, Profile %s", bag, profile)
			err := v.ScanBag()
			require.Nil(t, err, message)

			isValid := v.Validate()
			fmt.Println(v.ErrorString())
			assert.True(t, isValid, message)
			assert.Empty(t, v.Errors, message)
		}
	}
}

func TestValidator_BadOxum(t *testing.T) {
	profiles := []string{
		"aptrust-v2.2.json",
		"empty_profile.json",
		"btr-v1.0.json",
	}
	errMsg := "Payload-Oxum does not match payload"
	// We should get bad oxum error regardless of the
	// profile. Also, this error occurs before we even
	// get to scan the payload.
	for _, profile := range profiles {
		v := getValidator(t, "example.edu.sample_bad_oxum.tar", profile)
		err := v.ScanBag()
		require.NotNil(t, err)
		assert.Equal(t, errMsg, err.Error())
		assert.Equal(t, 1, len(v.Errors))
		assert.Equal(t, errMsg, v.Errors["Payload-Oxum"])
	}
}

// ---------------------------------------------------------------------
// Uncomment serialization test when we have a working zip reader.
// Until then, we can't run this.
// ---------------------------------------------------------------------

// func TestValidator_BadSerialization(t *testing.T) {
// 	// APTrust profile doesn't permit zip, only tar.
// 	v := getValidator(t, "example.edu.sample_good.zip", "aptrust-v2.2.json")
// 	err := v.ScanBag()
// 	require.Nil(t, err)

// 	assert.False(t, v.Validate())
// 	assert.Equal(t, 1, len(v.Errors))
// 	assert.True(t, strings.HasPrefix(v.Errors["Serialization"], "Bag has extension"))
// }

func TestValidator_MissingPayloadFile(t *testing.T) {
	v := getValidator(t, "example.edu.sample_missing_data_file.tar", "empty_profile.json")
	err := v.ScanBag()
	require.Nil(t, err)
	isValid := v.Validate()
	assert.False(t, isValid)
	assert.Equal(t, 1, len(v.Errors))
	assert.Equal(t, "file is missing from bag", v.Errors["data/datastream-DC"])
}

func TestValidator_MissingBagInfoFile(t *testing.T) {
	v := getValidator(t, "example.edu.sample_no_bag_info.tar", "aptrust-v2.2.json")
	v.Profile.ManifestsRequired = []string{}
	err := v.ScanBag()
	require.Nil(t, err)
	isValid := v.Validate()
	assert.False(t, isValid)
	assert.Equal(t, 3, len(v.Errors))
	fmt.Println(v.ErrorString())

	assert.Equal(t, "Required tag is missing.", v.Errors["aptrust-info.txt/Storage-Option"])
	assert.Equal(t, "Required tag is missing.", v.Errors["bag-info.txt/Source-Organization"])
	assert.Equal(t, "Required tag is missing.", v.Errors["aptrust-info.txt/Access"])

	// This bag is valid with the empty profile, because it doesn't
	// require any tags from the bag-info.txt file.
	v = getValidator(t, "example.edu.sample_no_bag_info.tar", "empty_profile.json")
	v.Profile.ManifestsRequired = []string{}
	err = v.ScanBag()
	require.Nil(t, err)
	isValid = v.Validate()
	assert.True(t, isValid)
	assert.Equal(t, 0, len(v.Errors))
}

func TestValidator_MissingDataDir(t *testing.T) {
	v := getValidator(t, "example.edu.sample_no_data_dir.tar", "empty_profile.json")
	err := v.ScanBag()
	require.Nil(t, err)

	isValid := v.Validate()
	assert.False(t, isValid)
	assert.Equal(t, 4, len(v.Errors))
	for _, msg := range v.Errors {
		assert.Equal(t, "file is missing from bag", msg)
	}
}

func TestValidator_MissingManifest(t *testing.T) {
	v := getValidator(t, "example.edu.sample_no_md5_manifest.tar", "aptrust-v2.2.json")
	err := v.ScanBag()
	require.Nil(t, err)

	isValid := v.Validate()
	assert.False(t, isValid)

	expected := `md5 -> Required manifest is missing.
aptrust-info.txt/Storage-Option -> Required tag is missing.`
	assert.Equal(t, expected, v.ErrorString())
}
