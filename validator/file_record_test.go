package validator_test

import (
	"testing"

	"github.com/APTrust/dart-runner/validator"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func getTestFileRecord() *validator.FileRecord {
	fr := validator.NewFileRecord()
	fr.AddChecksum(validator.SourcePayloadFile, validator.AlgMd5, "1234")
	fr.AddChecksum(validator.SourcePayloadFile, validator.AlgSha256, "5678")
	fr.AddChecksum(validator.SourceManifest, validator.AlgMd5, "1234")
	fr.AddChecksum(validator.SourceManifest, validator.AlgSha256, "5678")
	return fr
}

func TestNewFileRecord(t *testing.T) {
	fr := validator.NewFileRecord()
	assert.NotNil(t, fr)
	assert.NotNil(t, fr.Checksums)
}

func TestAddChecksum(t *testing.T) {
	fr := getTestFileRecord()
	assert.Equal(t, 4, len(fr.Checksums))
}

func TestFileRecordValidate(t *testing.T) {
	algs := []validator.DigestAlgorithm{
		validator.AlgMd5,
		validator.AlgSha256,
	}
	fr := getTestFileRecord()
	err := fr.Validate(validator.FileTypePayload, algs)
	assert.Nil(t, err)

	// Test digest missing from manifest. We have a sha512 checksum
	// for the file, but there's no checksum in sha512 manifest.
	xtra := append(algs, validator.AlgSha512)
	fr.AddChecksum(validator.SourcePayloadFile, validator.AlgSha512, "9999")
	err = fr.Validate(validator.FileTypePayload, xtra)
	require.NotNil(t, err)
	assert.Equal(t, "file is missing from manifest manifest-sha512.txt", err.Error())

	// Test mismatched checksums. File digest doesn't match manifest digest.
	fr.AddChecksum(validator.SourceManifest, validator.AlgSha512, "0000")
	err = fr.Validate(validator.FileTypePayload, xtra)
	require.NotNil(t, err)
	assert.Equal(t, "Digest 0000 in manifest-sha512.txt does not match digest 9999 in payload file", err.Error())

	// Test algorithm missing. In this case, the validator never
	// even calculated the requested digest.
	sha1 := []validator.DigestAlgorithm{validator.AlgSha1}
	err = fr.Validate(validator.FileTypePayload, sha1)
	require.NotNil(t, err)
	assert.Equal(t, "Digest sha1 was not calculated", err.Error())

	// Test payload file missing. In this case, the sha256 manifest
	// has a digest for the file, but we have no digest calculated
	// from the file itself (i.e. where source = SourcePayloadFile)
	sha256 := []validator.DigestAlgorithm{validator.AlgSha256}
	fr2 := validator.NewFileRecord()
	fr2.AddChecksum(validator.SourceManifest, validator.AlgSha256, "5678")
	err = fr2.Validate(validator.FileTypePayload, sha256)
	require.NotNil(t, err)
	assert.Equal(t, "file is missing from bag", err.Error())
}

func TestFileRecordGetChecksum(t *testing.T) {
	fr := getTestFileRecord()

	cs := fr.GetChecksum(validator.AlgMd5, validator.SourceManifest)
	assert.NotNil(t, cs)
	assert.Equal(t, validator.AlgMd5, cs.Algorithm)
	assert.Equal(t, validator.SourceManifest, cs.Source)

	cs = fr.GetChecksum(validator.AlgSha256, validator.SourcePayloadFile)
	assert.NotNil(t, cs)
	assert.Equal(t, validator.AlgSha256, cs.Algorithm)
	assert.Equal(t, validator.SourcePayloadFile, cs.Source)

	cs = fr.GetChecksum(validator.AlgSha256, validator.SourceTagFile)
	assert.Nil(t, cs)
}
