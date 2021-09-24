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

	validator := getValidator(t, "example.edu.tagsample_good.tar", "aptrust-v2.2.json")

	err = validator.ScanBag()
	require.Nil(t, err)

	// The scanner should have loaded the validator with
	// the same info as in our JSON file (except PathToBag,
	// which will differ on each machine).
	tarReaderTestFileMaps(t, expected.PayloadFiles, validator.PayloadFiles)
	tarReaderTestFileMaps(t, expected.PayloadManifests, validator.PayloadManifests)
	tarReaderTestFileMaps(t, expected.TagFiles, validator.TagFiles)
	tarReaderTestFileMaps(t, expected.TagManifests, validator.TagManifests)

	tarReaderTestTags(t, expected.Tags, validator.Tags)
}

func TestValidator_ValidateBasic(t *testing.T) {
	validator := getValidator(t, "example.edu.tagsample_good.tar", "aptrust-v2.2.json")
	// APTrust profile actually requires an md5 manifest,
	// which is not part of this bag. Let's test without
	// that requirement.
	validator.Profile.ManifestsRequired = []string{}

	err := validator.ScanBag()
	require.Nil(t, err)

	isValid := validator.Validate()
	fmt.Println(validator.ErrorString())
	assert.True(t, isValid)
}
