package core_test

import (
	"bytes"
	"fmt"
	"strings"
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

// TestWriteManifestDoesNotStripBagNameFromMiddleOfPath tests that
// WriteManifest only strips the trimFromPath prefix when it appears at
// the START of a path, not from the middle. This covers the case where
// an unserialized (directory) bag has a source directory whose name
// matches the bag name. For example, bagging a directory called "testbag"
// into an output bag also called "testbag" must produce manifest entries
// like "data/testbag/file1.txt", not the incorrect "data/file1.txt".
// See https://github.com/APTrust/dart-runner/issues/XX
func TestWriteManifestDoesNotStripBagNameFromMiddleOfPath(t *testing.T) {
	// Simulate a directory bag where source dir and bag name are both "mybag".
	// PathForPayloadFile produces paths like "data/mybag/file1.txt".
	// trimFromPath is "mybag/" (the bag name).
	// The old strings.Replace would turn "data/mybag/file1.txt" into
	// "data/file1.txt" — wrong. TrimPrefix leaves it unchanged.
	fm := core.NewFileMap(constants.FileTypePayload)
	files := []string{
		"data/mybag/file1.txt",
		"data/mybag/file2.txt",
		"data/mybag/subdir/file3.txt",
	}
	for _, name := range files {
		fr := core.NewFileRecord()
		fr.AddChecksum(constants.FileTypePayload, constants.AlgMd5, "abc123")
		fr.AddChecksum(constants.FileTypeManifest, constants.AlgMd5, "abc123")
		fm.Files[name] = fr
	}

	var buf bytes.Buffer
	// "mybag/" is the bag name — for a tar bag it would be a path prefix,
	// but for a directory bag these paths already have no such prefix.
	err := fm.WriteManifest(&buf, constants.FileTypePayload, constants.AlgMd5, "mybag/")
	require.Nil(t, err)

	output := buf.String()
	for _, name := range files {
		assert.True(t, strings.Contains(output, name),
			"manifest should contain %s but got:\n%s", name, output)
	}
	// The incorrect "stripped" paths must not appear.
	for _, bad := range []string{"data/file1.txt", "data/file2.txt", "data/subdir/file3.txt"} {
		assert.False(t, strings.Contains(output, bad),
			"manifest must not contain incorrectly stripped path %s", bad)
	}
}

// TestWriteManifestStripsLeadingBagName tests that for tarred bags the bag
// name IS correctly stripped from paths that begin with it.
func TestWriteManifestStripsLeadingBagName(t *testing.T) {
	fm := core.NewFileMap(constants.FileTypePayload)
	files := map[string]string{
		"mybag/data/file1.txt": "data/file1.txt",
		"mybag/data/file2.txt": "data/file2.txt",
	}
	for name := range files {
		fr := core.NewFileRecord()
		fr.AddChecksum(constants.FileTypePayload, constants.AlgMd5, "abc123")
		fr.AddChecksum(constants.FileTypeManifest, constants.AlgMd5, "abc123")
		fm.Files[name] = fr
	}

	var buf bytes.Buffer
	err := fm.WriteManifest(&buf, constants.FileTypePayload, constants.AlgMd5, "mybag/")
	require.Nil(t, err)

	output := buf.String()
	for _, want := range files {
		assert.True(t, strings.Contains(output, want),
			"manifest should contain %s but got:\n%s", want, output)
	}
}
