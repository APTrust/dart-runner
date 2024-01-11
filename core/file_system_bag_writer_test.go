package core_test

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/APTrust/dart-runner/core"
	"github.com/APTrust/dart-runner/util"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// var digestAlgs = []string{
// 	constants.AlgMd5,
// 	constants.AlgSha256,
// }

// func listTestFiles(t *testing.T) []*util.ExtendedFileInfo {
// 	files, err := util.RecursiveFileList(util.PathToTestData(), false)
// 	require.Nil(t, err)
// 	return files
// }

// func assertChecksums(t *testing.T, checksums map[string]string, filename string) {
// 	assert.NotNil(t, checksums, filename)
// 	for _, alg := range digestAlgs {
// 		assert.NotEmpty(t, checksums[alg], "Missing %s for %s", alg, filename)
// 	}
// }

// Note: digestAlgs, listTestFiles, and assertChecksums are in tarred_bag_writer_test.go

func getFSWriter(t *testing.T, filename string) (*core.FileSystemBagWriter, string) {
	tempFilePath := filepath.Join(os.TempDir(), filename)
	w := core.NewFileSystemBagWriter(tempFilePath, digestAlgs)
	assert.NotNil(t, w)
	assert.Equal(t, tempFilePath, w.OutputPath())
	return w, tempFilePath
}

func listFilesRecursive(dir string) ([]string, error) {
	// Include top-level dir in list
	files := []string{dir}
	err := filepath.Walk(dir, func(filePath string, f os.FileInfo, err error) error {
		if f.Mode().IsRegular() {
			files = append(files, f.Name())
		}
		return nil
	})
	return files, err
}

func TestFSBagWriterAndCloseOpen(t *testing.T) {
	w, tempFileName := getFSWriter(t, "fs-testbag-1")
	defer w.Close()
	defer os.Remove(tempFileName)
	err := w.Open()
	require.Nil(t, err)
	require.True(t, util.FileExists(w.OutputPath()), "Directory does not exist at %s", w.OutputPath())
	err = w.Close()
	assert.Nil(t, err)
}

func TestFSBagWriterAddFile(t *testing.T) {
	w, tempFileName := getFSWriter(t, "fs-testbag-2")
	defer w.Close()
	defer os.Remove(tempFileName)
	err := w.Open()
	assert.Nil(t, err)
	require.True(t, util.FileExists(w.OutputPath()), "Directory does not exist at %s", w.OutputPath())

	// Note that the first "file" added to the bag is the root directory,
	// which has the same name as the bag, minus the .tar extension
	filesAdded := []string{
		util.CleanBagName(filepath.Base(w.OutputPath())),
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
			filesAdded = append(filesAdded, pathInBag)
			if len(filesAdded) > 4 {
				break
			}
		}
	}

	w.Close()

	require.True(t, util.FileExists(w.OutputPath()))
	filesInBag, err := listFilesRecursive(w.OutputPath())
	require.NoError(t, err)

	// Make sure root directory and all files are present.
	require.Equal(t, len(filesAdded), len(filesInBag))
	for i, isInArchive := range filesInBag {
		shouldBeInArchive := filesAdded[i]
		assert.Equal(t, shouldBeInArchive, isInArchive)
	}
}

func TestFSBagWriterAddFileWithClosedWriter(t *testing.T) {
	w, tempFileName := getTarWriter(t, "fs-testbag-3")
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
	require.True(t, util.FileExists(w.OutputPath()))

	checksums, err = w.AddFile(files[0], files[0].Name())
	require.NotNil(t, err)
	assert.Empty(t, checksums)
	assert.True(t, strings.Contains(err.Error(), "tar: write after close"), err.Error())

}

func TestFSBagWriterAddFileWithBadFilePath(t *testing.T) {
	w, tempFileName := getTarWriter(t, "fs-testbag-4")
	defer w.Close()
	defer os.Remove(tempFileName)
	err := w.Open()
	assert.Nil(t, err)
	require.True(t, util.FileExists(w.OutputPath()))

	// We need a valid FileInfo object for our constructor, but...
	fInfo, err := os.Stat(util.PathToUnitTestBag("example.edu.sample_good.tar"))
	require.Nil(t, err)

	// ...the path we give here points to a file that does not exist.
	// Make sure we get the right error.
	xFileInfo := util.NewExtendedFileInfo("file-does-not-exist.pdf", fInfo)
	checksums, err := w.AddFile(xFileInfo, "file.pdf")
	require.NotNil(t, err)
	assert.Empty(t, checksums)
	expectedErr := "no such file or directory"
	if runtime.GOOS == "windows" {
		expectedErr = "The system cannot find the file specified."
	}
	assert.True(t, strings.Contains(err.Error(), expectedErr), err.Error())
}
