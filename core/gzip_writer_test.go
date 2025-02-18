package core_test

import (
	"crypto/sha256"
	"fmt"
	"io"
	"os"
	"path"
	"testing"

	"github.com/APTrust/dart-runner/core"
	"github.com/APTrust/dart-runner/util"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGzipWriter_Write(t *testing.T) {

	// Use PathToTestData()
	inputFile := path.Join(util.PathToTestData(), "bags", "example.edu.sample_good.tar")
	outputFile, err := os.CreateTemp("", "gzip-test*.tar.gz")
	require.Nil(t, err)

	defer os.Remove(outputFile.Name())

	// Close the temp file, so our writer can open it.
	require.Nil(t, outputFile.Close())

	gzipWriter := core.NewGzipWriter(inputFile)
	bytesWritten, err := gzipWriter.ZipToFile(outputFile.Name())
	require.Nil(t, err)

	// Original file was about 25k, so this many bytes were
	// written into the gzip writer...
	assert.Equal(t, int64(23552), bytesWritten)

	// ... but fewer bytes end up in the resulting gzip file
	// because the data has been compressed.
	fileInfo, err := os.Stat(outputFile.Name())
	require.Nil(t, err)
	assert.Equal(t, int64(3608), fileInfo.Size())

	tmpFile, err := os.CreateTemp("", "gzip-test*.tar")
	require.Nil(t, err)
	defer os.Remove(tmpFile.Name())
	require.Nil(t, tmpFile.Close())

	// Unzip the file and make compare its checksum to the
	// original so we can be sure they're the same.
	zipReader := core.NewGzipReader(outputFile.Name())
	bytesCopied, err := zipReader.UnzipToFile(tmpFile.Name())
	require.Nil(t, err)
	assert.Equal(t, int64(23552), bytesCopied)

	originalChecksum := GetSha256(t, inputFile)
	unzippedChecksum := GetSha256(t, tmpFile.Name())
	assert.Equal(t, originalChecksum, unzippedChecksum)
}

func GetSha256(t *testing.T, filename string) string {
	file, err := os.Open(filename)
	require.Nil(t, err)
	defer file.Close()

	hash := sha256.New()
	_, err = io.Copy(hash, file)
	require.Nil(t, err)
	return fmt.Sprintf("%x", hash.Sum(nil))
}
