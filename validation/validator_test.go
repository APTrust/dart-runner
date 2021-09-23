package validation_test

import (
	"testing"

	"github.com/APTrust/dart-runner/util"
	"github.com/APTrust/dart-runner/validation"
	"github.com/stretchr/testify/require"
)

func TestValidator_ScanBag(t *testing.T) {
	expected, err := loadValidatorFromJson("tagsample_good_metadata.json")
	require.Nil(t, err)
	require.NotNil(t, expected)

	profile, err := loadProfile("aptrust-v2.2.json")
	require.Nil(t, err)

	pathToBag := util.PathToUnitTestBag("example.edu.tagsample_good.tar")
	validator, err := validation.NewValidator(pathToBag, profile)

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
