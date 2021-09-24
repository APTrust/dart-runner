package validation_test

import (
	"fmt"
	"testing"

	"github.com/APTrust/dart-runner/util"
	"github.com/APTrust/dart-runner/validation"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var aptrustProfile = "aptrust-v2.2.json"
var btrProfile = "btr-v1.0.json"
var emptyProfile = "empty_profile.json"

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

	v := getValidator(t, "example.edu.tagsample_good.tar", aptrustProfile)

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
		aptrustProfile,
		emptyProfile,
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
			assert.True(t, isValid, message)
			assert.Empty(t, v.Errors, message)
		}
	}
}

func TestValidator_BadOxum(t *testing.T) {
	profiles := []string{
		aptrustProfile,
		emptyProfile,
		btrProfile,
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
// 	v := getValidator(t, "example.edu.sample_good.zip", aptrustProfile)
// 	err := v.ScanBag()
// 	require.Nil(t, err)

// 	assert.False(t, v.Validate())
// 	assert.Equal(t, 1, len(v.Errors))
// 	assert.True(t, strings.HasPrefix(v.Errors["Serialization"], "Bag has extension"))
// }

func TestValidator_MissingPayloadFile(t *testing.T) {
	v := getValidator(t, "example.edu.sample_missing_data_file.tar", emptyProfile)
	err := v.ScanBag()
	require.Nil(t, err)
	isValid := v.Validate()
	assert.False(t, isValid)
	assert.Equal(t, 1, len(v.Errors))
	assert.Equal(t, "file is missing from bag", v.Errors["data/datastream-DC"])
}

func TestValidator_MissingBagInfoFile(t *testing.T) {
	v := getValidator(t, "example.edu.sample_no_bag_info.tar", aptrustProfile)
	v.Profile.ManifestsRequired = []string{}
	err := v.ScanBag()
	require.Nil(t, err)
	isValid := v.Validate()
	assert.False(t, isValid)
	assert.Equal(t, 3, len(v.Errors))

	assert.Equal(t, "Required tag is missing.", v.Errors["aptrust-info.txt/Storage-Option"])
	assert.Equal(t, "Required tag is missing.", v.Errors["bag-info.txt/Source-Organization"])
	assert.Equal(t, "Required tag is missing.", v.Errors["aptrust-info.txt/Access"])

	// This bag is valid with the empty profile, because it doesn't
	// require any tags from the bag-info.txt file.
	v = getValidator(t, "example.edu.sample_no_bag_info.tar", emptyProfile)
	v.Profile.ManifestsRequired = []string{}
	err = v.ScanBag()
	require.Nil(t, err)
	isValid = v.Validate()
	assert.True(t, isValid)
	assert.Equal(t, 0, len(v.Errors))
}

func TestValidator_MissingDataDir(t *testing.T) {
	v := getValidator(t, "example.edu.sample_no_data_dir.tar", emptyProfile)
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
	v := getValidator(t, "example.edu.sample_no_md5_manifest.tar", aptrustProfile)
	err := v.ScanBag()
	require.Nil(t, err)
	isValid := v.Validate()
	assert.False(t, isValid)
	assert.Equal(t, 2, len(v.Errors))
	assert.Equal(t, "Required manifest is missing.", v.Errors["md5"])
	assert.Equal(t, "Required tag is missing.", v.Errors["aptrust-info.txt/Storage-Option"])
}

func TestValidator_BadTags(t *testing.T) {
	v := getValidator(t, "example.edu.tagsample_bad.tar", aptrustProfile)
	err := v.ScanBag()
	require.Nil(t, err)
	isValid := v.Validate()
	assert.False(t, isValid)
	assert.Equal(t, 1, len(v.Errors))
	assert.Equal(t, "file is missing from bag", v.Errors["custom_tags/tag_file_xyz.pdf"])

	// This bag has the required tag files but is missing
	// some required tags.
	v = getValidator(t, "virginia.edu.uva-lib_2278801.tar", aptrustProfile)
	err = v.ScanBag()
	require.Nil(t, err)
	isValid = v.Validate()
	assert.False(t, isValid)
	assert.Equal(t, 3, len(v.Errors))
	assert.Equal(t, "Required manifest is missing.", v.Errors["md5"])
	assert.Equal(t, "Required tag is missing.", v.Errors["aptrust-info.txt/Access"])
	assert.Equal(t, "Required tag is missing.", v.Errors["aptrust-info.txt/Storage-Option"])
}

func TestValidator_GoodBTRBags(t *testing.T) {
	bags := []string{
		"test.edu.btr-glacier-deep-oh.tar",
		"test.edu.btr-wasabi-or.tar",
		"test.edu.btr_good_sha256.tar",
		"test.edu.btr_good_sha512.tar",
	}
	profiles := []string{
		btrProfile,
		emptyProfile,
	}
	for _, bag := range bags {
		for _, profile := range profiles {
			v := getValidator(t, bag, profile)
			message := fmt.Sprintf("Bag %s, Profile %s", bag, profile)
			err := v.ScanBag()
			require.Nil(t, err, message)
			isValid := v.Validate()
			assert.True(t, isValid, message)
			assert.Empty(t, v.Errors, message)
		}
	}
}

func TestValidator_BTRBadChecksums(t *testing.T) {
	v := getValidator(t, "test.edu.btr_bad_checksums.tar", btrProfile)
	err := v.ScanBag()
	require.Nil(t, err)
	isValid := v.Validate()
	assert.False(t, isValid)
	assert.Equal(t, 1, len(v.Errors))
	assert.Equal(t, "Digest 00000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000 in manifest-sha512.txt does not match digest b598afc318e6d06f54a162e8e43bbc9cb071fcf0ffb3766b719011d8403d01290d6f2d7a9decc504395501f28f6c452c5a4317ee7bd309d4cd597984227d176d in payload file", v.Errors["data/netutil/listen.go"])
}

func TestValidator_BTRExtraFile(t *testing.T) {
	v := getValidator(t, "test.edu.btr_bad_extraneous_file.tar", btrProfile)
	err := v.ScanBag()
	require.NotNil(t, err)
	assert.Equal(t, 1, len(v.Errors))
	assert.Equal(t, "Payload-Oxum does not match payload", v.Errors["Payload-Oxum"])
}

func TestValidator_BTRMissingPayloadFile(t *testing.T) {
	v := getValidator(t, "test.edu.btr_bad_missing_payload_file.tar", btrProfile)
	err := v.ScanBag()
	require.NotNil(t, err)
	assert.Equal(t, 1, len(v.Errors))
	assert.Equal(t, "Payload-Oxum does not match payload", v.Errors["Payload-Oxum"])
}

func TestValidator_BTRMissingTags(t *testing.T) {
	v := getValidator(t, "test.edu.btr_bad_missing_required_tags.tar", btrProfile)
	err := v.ScanBag()
	require.Nil(t, err)
	isValid := v.Validate()
	assert.False(t, isValid)
	assert.Equal(t, 3, len(v.Errors))

	// Note that BTR requires Payload-Oxum while APTrust does not.
	assert.Equal(t, "Required tag is missing.", v.Errors["bag-info.txt/Bagging-Date"])
	assert.Equal(t, "Required tag is missing.", v.Errors["bag-info.txt/Payload-Oxum"])
	assert.Equal(t, "Required tag is missing.", v.Errors["bag-info.txt/Source-Organization"])
}

func TestValidator_IllegalControlCharacter(t *testing.T) {
	// Review the BagIt spec on this. It's actually pretty
	// lax and puts the burden on the bagger to ensure
	// good practices. Sections 5 and 6 give some guidance:
	// https://datatracker.ietf.org/doc/html/rfc8493
}

// TODO:
// BTR bags
// UVA bag
// bag with illegal control characters
