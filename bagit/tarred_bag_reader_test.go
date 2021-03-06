package bagit_test

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"testing"

	"github.com/APTrust/dart-runner/bagit"
	"github.com/APTrust/dart-runner/util"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func loadProfile(jsonFile string) (*bagit.Profile, error) {
	filePath := path.Join(util.ProjectRoot(), "profiles", jsonFile)
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	data, err := ioutil.ReadAll(file)
	if err != nil {
		return nil, err
	}
	profile := &bagit.Profile{}
	err = json.Unmarshal(data, profile)
	return profile, err
}

// We have json files containing metadata that a read should
// find when scanning a bag. We test our reader results against
// this known good data.
func loadValidatorFromJson(jsonFile string) (*bagit.Validator, error) {
	filePath := path.Join(util.PathToTestData(), "files", jsonFile)
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	data, err := ioutil.ReadAll(file)
	if err != nil {
		return nil, err
	}
	validator := &bagit.Validator{}
	err = json.Unmarshal(data, validator)
	if err != nil {
		validator.Profile, err = loadProfile("aptrust-v2.2.json")
	}
	return validator, err
}

func TestTarredBagScanner(t *testing.T) {
	expected, err := loadValidatorFromJson("tagsample_good_metadata.json")
	require.Nil(t, err)
	require.NotNil(t, expected)

	profile, err := loadProfile("aptrust-v2.2.json")
	require.Nil(t, err)
	pathToBag := util.PathToUnitTestBag("example.edu.tagsample_good.tar")
	validator, err := bagit.NewValidator(pathToBag, profile)
	require.Nil(t, err)
	reader, err := bagit.NewTarredBagReader(validator)
	require.Nil(t, err)

	// Scan the metadata...
	err = reader.ScanMetadata()
	require.Nil(t, err)

	// And the payload...
	err = reader.ScanPayload()
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

func tarReaderTestFileMaps(t *testing.T, expected, actual *bagit.FileMap) {
	require.Equal(t, len(expected.Files), len(actual.Files))
	for expectedName, expectedRecord := range expected.Files {
		actualRecord := actual.Files[expectedName]
		require.NotNil(t, actualRecord, expectedName)
		assert.Equal(t, expectedRecord.Size, actualRecord.Size, expectedName)
		for _, expectedChecksum := range expectedRecord.Checksums {
			message := fmt.Sprintf("File: %s, Checksum: %s", expectedName, expectedChecksum.Algorithm)
			actualChecksum := actualRecord.GetChecksum(expectedChecksum.Algorithm, expectedChecksum.Source)
			require.NotNil(t, actualChecksum, message)
			assert.Equal(t, expectedChecksum.Algorithm, actualChecksum.Algorithm)
			assert.Equal(t, expectedChecksum.Source, actualChecksum.Source)
			assert.Equal(t, expectedChecksum.Digest, actualChecksum.Digest)
		}
	}
}

func tarReaderTestTags(t *testing.T, expected, actual []*bagit.Tag) {
	require.Equal(t, len(expected), len(actual))
	for i, expectedTag := range expected {
		message := fmt.Sprintf("%s/%s", expectedTag.TagFile, expectedTag.TagName)
		actualTag := actual[i]
		assert.Equal(t, expectedTag.TagFile, actualTag.TagFile, message)
		assert.Equal(t, expectedTag.TagName, actualTag.TagName, message)
		assert.Equal(t, expectedTag.Value, actualTag.Value, message)
	}
}
