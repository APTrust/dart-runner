package bagit_test

import (
	"testing"

	"github.com/APTrust/dart-runner/bagit"
	"github.com/APTrust/dart-runner/constants"
	"github.com/APTrust/dart-runner/util"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func getTestFileRecord() *bagit.FileRecord {
	fr := bagit.NewFileRecord()
	fr.AddChecksum(constants.FileTypePayload, constants.AlgMd5, "1234")
	fr.AddChecksum(constants.FileTypePayload, constants.AlgSha256, "5678")
	fr.AddChecksum(constants.FileTypeManifest, constants.AlgMd5, "1234")
	fr.AddChecksum(constants.FileTypeManifest, constants.AlgSha256, "5678")
	return fr
}

func TestNewFileRecord(t *testing.T) {
	fr := bagit.NewFileRecord()
	assert.NotNil(t, fr)
	assert.NotNil(t, fr.Checksums)
}

func TestAddChecksum(t *testing.T) {
	fr := getTestFileRecord()
	assert.Equal(t, 4, len(fr.Checksums))
}

func TestFileRecordValidate(t *testing.T) {
	algs := []string{
		constants.AlgMd5,
		constants.AlgSha256,
	}
	fr := getTestFileRecord()
	err := fr.Validate(constants.FileTypePayload, algs)
	assert.Nil(t, err)

	// Test digest missing from manifest. We have a sha512 checksum
	// for the file, but there's no checksum in sha512 manifest.
	xtra := append(algs, constants.AlgSha512)
	fr.AddChecksum(constants.FileTypePayload, constants.AlgSha512, "9999")
	err = fr.Validate(constants.FileTypePayload, xtra)
	require.NotNil(t, err)
	assert.Equal(t, "file is missing from manifest-sha512.txt", err.Error())

	// Test mismatched checksums. File digest doesn't match manifest digest.
	fr.AddChecksum(constants.FileTypeManifest, constants.AlgSha512, "0000")
	err = fr.Validate(constants.FileTypePayload, xtra)
	require.NotNil(t, err)
	assert.Equal(t, "Digest 0000 in manifest-sha512.txt does not match digest 9999 in payload file", err.Error())

	// Test algorithm missing. In this case, the validator never
	// calculated the requested digest because the file wasn't there.
	sha1 := []string{constants.AlgSha1}
	err = fr.Validate(constants.FileTypePayload, sha1)
	require.NotNil(t, err)
	assert.Equal(t, "file is missing from bag", err.Error())

	// Test payload file missing. In this case, the sha256 manifest
	// has a digest for the file, but we have no digest calculated
	// from the file itself (i.e. where source = SourcePayloadFile)
	sha256 := []string{constants.AlgSha256}
	fr2 := bagit.NewFileRecord()
	fr2.AddChecksum(constants.FileTypeManifest, constants.AlgSha256, "5678")
	err = fr2.Validate(constants.FileTypePayload, sha256)
	require.NotNil(t, err)
	assert.Equal(t, "file is missing from bag", err.Error())

	// Test tag file. It's OK for a tag file to appear in the bag
	// and be omitted from the tag manifest.
	fr3 := bagit.NewFileRecord()
	fr3.AddChecksum(constants.FileTypeTag, constants.AlgSha256, "5678")
	err = fr3.Validate(constants.FileTypeTag, sha256)
	assert.Nil(t, err)

	// Not OK for a tag file to be listed in a manifest if the tag
	// file is missing from the bag.
	fr4 := bagit.NewFileRecord()
	fr4.AddChecksum(constants.FileTypeTagManifest, constants.AlgSha256, "5678")
	err = fr4.Validate(constants.FileTypeTag, sha256)
	require.NotNil(t, err)
	assert.Equal(t, "file is missing from bag", err.Error())
}

func TestFileRecordGetChecksum(t *testing.T) {
	fr := getTestFileRecord()

	cs := fr.GetChecksum(constants.AlgMd5, constants.FileTypeManifest)
	assert.NotNil(t, cs)
	assert.Equal(t, constants.AlgMd5, cs.Algorithm)
	assert.Equal(t, constants.FileTypeManifest, cs.Source)

	cs = fr.GetChecksum(constants.AlgSha256, constants.FileTypePayload)
	assert.NotNil(t, cs)
	assert.Equal(t, constants.AlgSha256, cs.Algorithm)
	assert.Equal(t, constants.FileTypePayload, cs.Source)

	cs = fr.GetChecksum(constants.AlgSha256, constants.FileTypeTag)
	assert.Nil(t, cs)
}

func TestDigestAlgorithms(t *testing.T) {
	fr := getTestFileRecord()
	algs := fr.DigestAlgorithms()
	assert.Equal(t, 2, len(algs))
	assert.True(t, util.StringListContains(algs, constants.AlgMd5))
	assert.True(t, util.StringListContains(algs, constants.AlgSha256))
}
