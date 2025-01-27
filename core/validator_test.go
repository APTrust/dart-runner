package core_test

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/APTrust/dart-runner/constants"
	"github.com/APTrust/dart-runner/core"
	"github.com/APTrust/dart-runner/util"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var aptrustProfile = "aptrust-v2.2.json"
var btrProfile = "btr-v1.0.json"
var emptyProfile = "empty_profile.json"

func getValidator(t *testing.T, bag, profile string) *core.Validator {
	bagItProfile := loadProfile(t, profile)

	pathToBag := util.PathToUnitTestBag(bag)
	validator, err := core.NewValidator(pathToBag, bagItProfile)
	require.Nil(t, err)
	return validator
}

func TestValidator_ScanBag(t *testing.T) {
	expected := loadValidatorFromJson(t, "tagsample_good_metadata.json")
	require.NotNil(t, expected)

	v := getValidator(t, "example.edu.tagsample_good.tar", aptrustProfile)

	err := v.ScanBag()
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
	assert.Equal(t, "file is missing from bag: data/datastream-DC", v.Errors["data/datastream-DC"])
}

func TestValidator_MissingBagInfoFile(t *testing.T) {
	v := getValidator(t, "example.edu.sample_no_bag_info.tar", aptrustProfile)
	v.Profile.ManifestsRequired = []string{}
	err := v.ScanBag()
	require.Nil(t, err)
	isValid := v.Validate()
	assert.False(t, isValid)
	assert.Equal(t, 3, len(v.Errors))

	assert.Equal(t, "Required tag is missing: aptrust-info.txt/Storage-Option", v.Errors["aptrust-info.txt/Storage-Option"])
	assert.Equal(t, "Required tag is missing: bag-info.txt/Source-Organization", v.Errors["bag-info.txt/Source-Organization"])
	assert.Equal(t, "Required tag is missing: aptrust-info.txt/Access", v.Errors["aptrust-info.txt/Access"])

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
		assert.True(t, strings.HasPrefix(msg, "file is missing from bag: "))
	}
}

func TestValidator_MissingManifest(t *testing.T) {
	v := getValidator(t, "example.edu.sample_no_md5_manifest.tar", aptrustProfile)
	err := v.ScanBag()
	require.Nil(t, err)
	isValid := v.Validate()
	assert.False(t, isValid)
	assert.Equal(t, 2, len(v.Errors))
	assert.Equal(t, "Required manifest 'manifest-md5.txt' is missing.", v.Errors["manifest-md5.txt"])
	assert.Equal(t, "Required tag is missing: aptrust-info.txt/Storage-Option", v.Errors["aptrust-info.txt/Storage-Option"])
}

func TestValidator_BadTags(t *testing.T) {
	v := getValidator(t, "example.edu.tagsample_bad.tar", aptrustProfile)
	err := v.ScanBag()
	require.Nil(t, err)
	isValid := v.Validate()
	assert.False(t, isValid)
	assert.Equal(t, 6, len(v.Errors))
	assert.Equal(t, "file is missing from bag: custom_tags/tag_file_xyz.pdf", v.Errors["custom_tags/tag_file_xyz.pdf"])
	assert.Equal(t, "Required tag 'aptrust-info.txt/Title' is present but has no value.", v.Errors["aptrust-info.txt/Title"])
	assert.Equal(t, "Tag 'aptrust-info.txt/Access' has illegal value 'acksess'. Allowed values are: Consortia,Institution,Restricted", v.Errors["aptrust-info.txt/Access"])
	assert.Equal(t, "Tag 'aptrust-info.txt/Storage-Option' has illegal value 'Cardboard-Box'. Allowed values are: Standard,Glacier-OH,Glacier-OR,Glacier-VA,Glacier-Deep-OH,Glacier-Deep-OR,Glacier-Deep-VA,Wasabi-OR,Wasabi-VA", v.Errors["aptrust-info.txt/Storage-Option"])
	assert.Equal(t, "Digest This-checksum-is-bad-on-purpose.-The-validator-should-catch-it!! in manifest-sha256.txt does not match digest cf9cbce80062932e10ee9cd70ec05ebc24019deddfea4e54b8788decd28b4bc7 in payload file: data/datastream-descMetadata", v.Errors["data/datastream-descMetadata"])
	assert.Equal(t, "file is missing from bag: data/file-not-in-bag", v.Errors["data/file-not-in-bag"])

	// This bag has the required tag files but is missing
	// some required tags.
	v = getValidator(t, "virginia.edu.uva-lib_2278801.tar", aptrustProfile)
	err = v.ScanBag()
	require.Nil(t, err)
	isValid = v.Validate()
	assert.False(t, isValid)
	assert.Equal(t, 2, len(v.Errors))
	assert.Equal(t, "Required tag is missing: aptrust-info.txt/Access", v.Errors["aptrust-info.txt/Access"])
	assert.Equal(t, "Required tag is missing: aptrust-info.txt/Storage-Option", v.Errors["aptrust-info.txt/Storage-Option"])
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
	assert.Equal(t, "Digest 00000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000 in manifest-sha512.txt does not match digest b598afc318e6d06f54a162e8e43bbc9cb071fcf0ffb3766b719011d8403d01290d6f2d7a9decc504395501f28f6c452c5a4317ee7bd309d4cd597984227d176d in payload file: data/netutil/listen.go", v.Errors["data/netutil/listen.go"])
}

func TestValidator_BTRExtraFile(t *testing.T) {
	v := getValidator(t, "test.edu.btr_bad_extraneous_file.tar", btrProfile)
	err := v.ScanBag()
	require.NotNil(t, err)
	assert.Equal(t, 1, len(v.Errors))
	assert.Equal(t, "Payload-Oxum does not match payload", v.Errors["Payload-Oxum"])

	// Validator should flag the bad file explicitly if
	// we set IgnoreOxumMismatch to true.
	v = getValidator(t, "test.edu.btr_bad_extraneous_file.tar", btrProfile)
	v.IgnoreOxumMismatch = true
	err = v.ScanBag()
	require.Nil(t, err)
	isValid := v.Validate()
	assert.False(t, isValid)
	assert.Equal(t, 2, len(v.Errors))
	assert.Equal(t, "file is missing from manifest-sha512.txt: data/nsqd.dat", v.Errors["data/nsqd.dat"])
	assert.Equal(t, "Payload-Oxum does not match payload", v.Errors["Payload-Oxum"])
}

func TestValidator_BTRMissingPayloadFile(t *testing.T) {
	v := getValidator(t, "test.edu.btr_bad_missing_payload_file.tar", btrProfile)
	err := v.ScanBag()
	require.NotNil(t, err)
	assert.Equal(t, 1, len(v.Errors))
	assert.Equal(t, "Payload-Oxum does not match payload", v.Errors["Payload-Oxum"])

	// Get a more specific error with IgnoreOxumMismatch to true.
	v = getValidator(t, "test.edu.btr_bad_missing_payload_file.tar", btrProfile)
	v.IgnoreOxumMismatch = true
	err = v.ScanBag()
	require.Nil(t, err)
	isValid := v.Validate()
	assert.False(t, isValid)
	assert.Equal(t, 2, len(v.Errors))
	assert.Equal(t, "file is missing from bag: data/netutil/listen.go", v.Errors["data/netutil/listen.go"])
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
	assert.Equal(t, "Required tag is missing: bag-info.txt/Bagging-Date", v.Errors["bag-info.txt/Bagging-Date"])
	assert.Equal(t, "Required tag is missing: bag-info.txt/Payload-Oxum", v.Errors["bag-info.txt/Payload-Oxum"])
	assert.Equal(t, "Required tag is missing: bag-info.txt/Source-Organization", v.Errors["bag-info.txt/Source-Organization"])
}

func TestValidator_IllegalControlCharacter(t *testing.T) {
	// Review the BagIt spec on this. It's actually pretty
	// lax and puts the burden on the bagger to ensure
	// good practices. Sections 5 and 6 give some guidance:
	// https://datatracker.ietf.org/doc/html/rfc8493

	// Get the app setting for how to deal with control chars,
	// so we can manipulate it in our tests. This setting won't
	// exists in tests, because we start with a blank database.
	_, err := core.GetAppSetting(constants.ControlCharactersInFileNames)
	require.NotNil(t, err)
	setting := core.NewAppSetting(constants.ControlCharactersInFileNames, constants.ControlCharIgnore)
	require.Nil(t, core.ObjSave(setting))

	// Get rid of this when we're done.
	defer func() {
		assert.NoError(t, core.ObjDelete(setting))
	}()

	// Create a temp tempFile to bag. The tempFile name contains a
	// control character. This tempFile name contains the unicode
	// bell character.
	tempFile, err := os.CreateTemp("", "\u0007-bell*")
	require.Nil(t, err)
	tempFile.Write([]byte("DART test file"))
	tempFile.Close()

	sourceFiles, err := util.RecursiveFileList(tempFile.Name(), false)
	require.Nil(t, err)

	// Bagger should ignore the control char in this file name
	// because setting says Ignore.
	bagger := getBagger(t, "control-char-bag.tar", APTProfile, sourceFiles)
	bagger.Profile.SetTagValue("aptrust-info.txt", "Title", "Test Bag with Control Characters")
	ok := bagger.Run()
	assert.True(t, ok)
	assert.Empty(t, bagger.Errors)
	assert.Empty(t, bagger.Warnings)

	// Now let's test some validation scenarios.
	// Validation should succeed with no warning when
	// AppSetting says to ignore control chars.
	validator, err := core.NewValidator(bagger.OutputPath, bagger.Profile)
	require.NoError(t, err)
	err = validator.ScanBag()
	require.NoError(t, err)
	ok = validator.Validate()
	assert.True(t, ok)
	assert.Empty(t, validator.Errors)
	assert.Empty(t, validator.Warnings)

	// When set to warn, bag should be valid, and we
	// get a warning.
	setting.Value = constants.ControlCharWarn
	require.Nil(t, core.ObjSave(setting))
	validator, err = core.NewValidator(bagger.OutputPath, bagger.Profile)
	require.NoError(t, err)
	err = validator.ScanBag()
	require.NoError(t, err)
	ok = validator.Validate()
	assert.True(t, ok)
	assert.Empty(t, validator.Errors)
	assert.Equal(t, 1, len(validator.Warnings))
	assert.True(t, strings.Contains(validator.Warnings["File Names"], filepath.Base(tempFile.Name())))

	// This will also pass, because the setting "Refuse to Bag"
	// applies only to the bagger, not to the validator.
	// But it should include a warning, since the user has
	// indicated concern about control chars.
	setting.Value = constants.ControlCharRefuseToBag
	require.Nil(t, core.ObjSave(setting))
	validator, err = core.NewValidator(bagger.OutputPath, bagger.Profile)
	require.NoError(t, err)
	err = validator.ScanBag()
	require.NoError(t, err)
	ok = validator.Validate()
	assert.True(t, ok)
	assert.Empty(t, validator.Errors)
	assert.Equal(t, 1, len(validator.Warnings))
	assert.True(t, strings.Contains(validator.Warnings["File Names"], filepath.Base(tempFile.Name())))

	// This should fail validation with an error, not
	// a warning.
	setting.Value = constants.ControlCharFailValidation
	require.Nil(t, core.ObjSave(setting))
	validator, err = core.NewValidator(bagger.OutputPath, bagger.Profile)
	require.NoError(t, err)
	err = validator.ScanBag()
	require.NoError(t, err)
	ok = validator.Validate()
	assert.False(t, ok) // failed!
	assert.Empty(t, validator.Warnings)
	assert.Equal(t, 1, len(validator.Errors))
	assert.True(t, strings.Contains(validator.Errors["File Names"], filepath.Base(tempFile.Name())))

}
