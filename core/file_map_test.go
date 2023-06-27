package core_test

import (
	"fmt"
	"testing"

	"github.com/APTrust/dart-runner/constants"
	"github.com/APTrust/dart-runner/core"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var fmAlgs = []string{
	constants.AlgMd5,
	constants.AlgSha1,
}

var fmDigests = []string{
	"1234",
	"5678",
}

func TestNewFileMap(t *testing.T) {
	fm := core.NewFileMap(constants.FileTypePayload)
	require.NotNil(t, fm)
	assert.Equal(t, constants.FileTypePayload, fm.Type)
	require.NotNil(t, fm.Files)
	assert.Equal(t, 0, len(fm.Files))
}

func TestValidatePayloadChecksums(t *testing.T) {
	fm := makePayloadFileMap()
	require.NotNil(t, fm)
	assert.Equal(t, 5, len(fm.Files))

	errs := fm.ValidateChecksums(fmAlgs)
	assert.Equal(t, 0, len(errs))

	// Should get five errors because all files are missing
	// the sha512 checksum.
	xtra := append(fmAlgs, "sha512")
	errs = fm.ValidateChecksums(xtra)
	assert.Equal(t, 5, len(errs))
}

func TestValidateTagFileChecksums(t *testing.T) {
	fm := makeTagFileMap()
	require.NotNil(t, fm)
	assert.Equal(t, 5, len(fm.Files))

	errs := fm.ValidateChecksums(fmAlgs)
	assert.Equal(t, 0, len(errs))

	// Should get five errors because all files are missing
	// the sha512 checksum.
	xtra := append(fmAlgs, "sha512")
	errs = fm.ValidateChecksums(xtra)
	assert.Equal(t, 5, len(errs))

	// OK for tag files not to appear in manifests
	fm2 := makeTagFileMapWithoutManifestEntries()
	errs = fm2.ValidateChecksums(fmAlgs)
	assert.Equal(t, 0, len(errs))
}

func makePayloadFileMap() *core.FileMap {
	return makeFileMap(constants.FileTypePayload, constants.FileTypeManifest, "data/file")
}

func makeTagFileMap() *core.FileMap {
	return makeFileMap(constants.FileTypeTag, constants.FileTypeTagManifest, "tagfile")
}

func makeFileMap(fileType, manifestType, filename string) *core.FileMap {
	fileMap := core.NewFileMap(fileType)
	for i := 0; i < 5; i++ {
		filename := fmt.Sprintf("%s%d", filename, i)
		fr := core.NewFileRecord()
		for j, alg := range fmAlgs {
			fr.AddChecksum(fileType, alg, fmDigests[j])
			fr.AddChecksum(manifestType, alg, fmDigests[j])
		}
		fileMap.Files[filename] = fr
	}
	return fileMap
}

func makeTagFileMapWithoutManifestEntries() *core.FileMap {
	fileMap := core.NewFileMap(constants.FileTypeTag)
	for i := 0; i < 5; i++ {
		filename := fmt.Sprintf("tagfile%d", i)
		fr := core.NewFileRecord()
		for j, alg := range fmAlgs {
			fr.AddChecksum(constants.FileTypeTag, alg, fmDigests[j])
		}
		fileMap.Files[filename] = fr
	}
	return fileMap
}
