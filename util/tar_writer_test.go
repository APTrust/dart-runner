package util_test

import (
	"archive/tar"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/APTrust/dart-runner/util"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func listTestFiles(t *testing.T) []*util.ExtendedFileInfo {
	files, err := util.RecursiveFileList(util.PathToTestData())
	require.Nil(t, err)
	return files
}

func getTarWriter(t *testing.T, filename string) (*util.TarWriter, string) {
	tempDir, err := ioutil.TempDir("", "tarwriter_test")
	if err != nil {
		assert.FailNow(t, "Cannot create temp dir", err.Error())
	}
	tempFilePath := filepath.Join(tempDir, filename)
	w := util.NewTarWriter(tempFilePath)
	assert.NotNil(t, w)
	assert.Equal(t, tempFilePath, w.PathToTarFile)
	return w, tempFilePath
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

func TestAddToArchive(t *testing.T) {
	w, tempFileName := getTarWriter(t, "test2.tar")
	defer w.Close()
	defer os.Remove(tempFileName)
	err := w.Open()
	assert.Nil(t, err)
	require.True(t, util.FileExists(w.PathToTarFile), "Tar file does not exist at %s", w.PathToTarFile)

	filesAdded := make([]string, 0)
	files := listTestFiles(t)
	for _, xFileInfo := range files {
		if xFileInfo.IsDir() {
			continue
		}
		// Use Sprintf with forward slash instead of path.Join()
		// because tar file paths should use / even on windows.
		pathInBag := fmt.Sprintf("data/%s", xFileInfo.Name())
		err = w.AddToArchive(xFileInfo, pathInBag)
		assert.Nil(t, err, xFileInfo.FullPath)
		filesAdded = append(filesAdded, pathInBag)
		if len(filesAdded) > 3 {
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
	require.Equal(t, len(filesAdded), len(filesInArchive))
	assert.Equal(t, filesAdded[0], filesInArchive[0])
	assert.Equal(t, filesAdded[1], filesInArchive[1])
	assert.Equal(t, filesAdded[2], filesInArchive[2])
	assert.Equal(t, filesAdded[3], filesInArchive[3])

}

func TestAddToArchiveWithClosedWriter(t *testing.T) {
	w, tempFileName := getTarWriter(t, "test3.tar")
	defer w.Close()
	defer os.Remove(tempFileName)

	// Note that we have not opened the writer
	files := listTestFiles(t)
	err := w.AddToArchive(files[0], files[0].Name())
	require.NotNil(t, err)
	assert.True(t, strings.HasPrefix(err.Error(), "Underlying TarWriter is nil"))

	// Open and close the writer, so the file exists.
	w.Open()
	w.Close()
	require.True(t, util.FileExists(w.PathToTarFile))

	err = w.AddToArchive(files[0], files[0].Name())
	require.NotNil(t, err)
	assert.True(t, strings.Contains(err.Error(), "tar: write after close"), err.Error())

}

func TestAddToArchiveWithBadFilePath(t *testing.T) {
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
	err = w.AddToArchive(xFileInfo, "file.pdf")
	require.NotNil(t, err)
	assert.True(t, strings.Contains(err.Error(), "no such file or directory"))
}
