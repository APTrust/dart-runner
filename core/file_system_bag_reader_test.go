package core_test

import (
	"archive/tar"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/APTrust/dart-runner/core"
	"github.com/APTrust/dart-runner/util"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFileSystemBagReader(t *testing.T) {
	expected := loadValidatorFromJson(t, "tagsample_good_unserialized_metadata.json")
	require.NotNil(t, expected)

	profile := loadProfile(t, "aptrust-v2.2.json")
	pathToBag := util.PathToUnitTestBag("example.edu.tagsample_good.tar")

	tempDir := filepath.Join(os.TempDir(), "untarred-bag")
	require.NoError(t, Untar(pathToBag, tempDir))
	defer os.RemoveAll(tempDir)

	pathToUntarredBag := filepath.Join(tempDir, "example.edu.tagsample_good")

	validator, err := core.NewValidator(pathToUntarredBag, profile)
	require.Nil(t, err)
	reader, err := core.NewFileSystemBagReader(validator)
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
	fsReaderTestFileMaps(t, expected.PayloadFiles, validator.PayloadFiles)
	fsReaderTestFileMaps(t, expected.PayloadManifests, validator.PayloadManifests)
	fsReaderTestFileMaps(t, expected.TagFiles, validator.TagFiles)
	fsReaderTestFileMaps(t, expected.TagManifests, validator.TagManifests)

	fsReaderTestTags(t, expected.Tags, validator.Tags)
}

func fsReaderTestFileMaps(t *testing.T, expected, actual *core.FileMap) {
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

func fsReaderTestTags(t *testing.T, expected, actual []*core.Tag) {
	require.Equal(t, len(expected), len(actual))
	for i, expectedTag := range expected {
		message := expectedTag.FullyQualifiedName()
		actualTag := actual[i]
		assert.Equal(t, expectedTag.TagFile, actualTag.TagFile, message)
		assert.Equal(t, expectedTag.TagName, actualTag.TagName, message)
		assert.Equal(t, expectedTag.Value, actualTag.Value, message)
	}
}

func Untar(pathToTarFile, outputDir string) error {
	tarFile, err := os.Open(pathToTarFile)
	if err != nil {
		return err
	}
	tarReader := tar.NewReader(tarFile)
	for {
		header, err := tarReader.Next()
		switch {
		case err == io.EOF:
			return nil
		case err != nil:
			return err
		case header == nil:
			continue
		}

		target := filepath.Join(outputDir, header.Name)
		switch header.Typeflag {
		case tar.TypeDir:
			if _, err := os.Stat(target); err != nil {
				if err := os.MkdirAll(target, 0755); err != nil {
					return err
				}
			}
		case tar.TypeReg:
			f, err := os.OpenFile(target, os.O_CREATE|os.O_RDWR, os.FileMode(header.Mode))
			if err != nil {
				return err
			}
			if _, err := io.Copy(f, tarReader); err != nil {
				return err
			}
			f.Close()
		}
	}
}
