package core_test

import (
	"archive/tar"
	"fmt"
	"os"
	"path"
	"strings"
	"testing"

	"github.com/APTrust/dart-runner/constants"
	"github.com/APTrust/dart-runner/core"
	"github.com/APTrust/dart-runner/util"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var digestAlgs = []string{
	constants.AlgMd5,
	constants.AlgSha256,
}

func listTestFiles(t *testing.T) []*util.ExtendedFileInfo {
	files, err := util.RecursiveFileList(util.PathToTestData(), false)
	require.Nil(t, err)
	return files
}

func getTarWriter(t *testing.T, filename string) (*core.TarWriter, string) {
	tempFilePath := path.Join(os.TempDir(), filename)
	w := core.NewTarWriter(tempFilePath, digestAlgs)
	assert.NotNil(t, w)
	assert.Equal(t, tempFilePath, w.PathToTarFile)
	return w, tempFilePath
}

func assertChecksums(t *testing.T, checksums map[string]string, filename string) {
	assert.NotNil(t, checksums, filename)
	for _, alg := range digestAlgs {
		assert.NotEmpty(t, checksums[alg], "Missing %s for %s", alg, filename)
	}
}

func TestAndCloseOpen(t *testing.T) {
	w, tempFileName := getTarWriter(t, "test1.tar")
	defer w.Close()
	defer os.Remove(tempFileName)
	err := w.Open()
	require.Nil(t, err)
	require.True(t, util.FileExists(w.PathToTarFile), "Tar file does not exist at %s", w.PathToTarFile)
	err = w.Close()
	assert.Nil(t, err)
}

func TestAddFile(t *testing.T) {
	w, tempFileName := getTarWriter(t, "test2.tar")
	defer w.Close()
	defer os.Remove(tempFileName)
	err := w.Open()
	assert.Nil(t, err)
	require.True(t, util.FileExists(w.PathToTarFile), "Tar file does not exist at %s", w.PathToTarFile)

	// Note that the first "file" added to the bag is the root directory,
	// which has the same name as the bag, minus the .tar extension
	filesAdded := []string{
		util.CleanBagName(path.Base(w.PathToTarFile)),
	}
	files := listTestFiles(t)
	for _, xFileInfo := range files {
		// Use Sprintf with forward slash instead of path.Join()
		// because tar file paths should use / even on windows.
		pathInBag := fmt.Sprintf("data/%s", xFileInfo.Name())
		checksums, err := w.AddFile(xFileInfo, pathInBag)
		assert.Nil(t, err, xFileInfo.FullPath)
		if !xFileInfo.IsDir() {
			assertChecksums(t, checksums, xFileInfo.FullPath)
		}
		filesAdded = append(filesAdded, pathInBag)
		if len(filesAdded) > 4 {
			break
		}
	}

	w.Close()

	file, err := os.Open(w.PathToTarFile)
	if file != nil {
		defer file.Close()
	}
	require.Nil(t, err)
	filesInArchive := make([]string, 0)
	reader := tar.NewReader(file)
	for {
		header, err := reader.Next()
		if err != nil {
			break
		}
		filesInArchive = append(filesInArchive, header.Name)
	}

	// Make sure root directory and all files are present.
	require.Equal(t, len(filesAdded), len(filesInArchive))
	for i, isInArchive := range filesInArchive {
		shouldBeInArchive := filesAdded[i]
		assert.Equal(t, shouldBeInArchive, isInArchive)
	}
}

func TestAddFileWithClosedWriter(t *testing.T) {
	w, tempFileName := getTarWriter(t, "test3.tar")
	defer w.Close()
	defer os.Remove(tempFileName)

	// Note that we have not opened the writer
	files := listTestFiles(t)
	checksums, err := w.AddFile(files[0], files[0].Name())
	require.NotNil(t, err)
	assert.Empty(t, checksums)
	assert.True(t, strings.HasPrefix(err.Error(), "Underlying TarWriter is nil"))

	// Open and close the writer, so the file exists.
	w.Open()
	w.Close()
	require.True(t, util.FileExists(w.PathToTarFile))

	checksums, err = w.AddFile(files[0], files[0].Name())
	require.NotNil(t, err)
	assert.Empty(t, checksums)
	assert.True(t, strings.Contains(err.Error(), "tar: write after close"), err.Error())

}

func TestAddFileWithBadFilePath(t *testing.T) {
	w, tempFileName := getTarWriter(t, "test4.tar")
	defer w.Close()
	defer os.Remove(tempFileName)
	err := w.Open()
	assert.Nil(t, err)
	require.True(t, util.FileExists(w.PathToTarFile))

	// We need a valid FileInfo object for our constructor, but...
	fInfo, err := os.Stat(util.PathToUnitTestBag("example.edu.sample_good.tar"))
	require.Nil(t, err)

	// ...the path we give here points to a file that does not exist.
	// Make sure we get the right error.
	xFileInfo := util.NewExtendedFileInfo("file-does-not-exist.pdf", fInfo)
	checksums, err := w.AddFile(xFileInfo, "file.pdf")
	require.NotNil(t, err)
	assert.Empty(t, checksums)
	assert.True(t, strings.Contains(err.Error(), "no such file or directory"))
}
