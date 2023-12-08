package util_test

import (
	"os"
	"path"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/APTrust/dart-runner/util"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFileExists(t *testing.T) {
	f := util.PathToUnitTestBag("example.edu.sample_good.tar")
	assert.True(t, util.FileExists(f))
	assert.True(t, util.FileExists(util.ProjectRoot()))
	assert.False(t, util.FileExists("NonExistentFile.xyz"))
}

func TestIsDir(t *testing.T) {
	f := util.PathToUnitTestBag("example.edu.sample_good.tar")
	assert.False(t, util.IsDirectory(f))
	assert.False(t, util.IsDirectory("NonExistentFile.xyz"))
	assert.True(t, util.IsDirectory(util.ProjectRoot()))
}

func TestExpandTilde(t *testing.T) {
	expanded, err := util.ExpandTilde("~/tmp")
	assert.Nil(t, err)
	assert.True(t, len(expanded) > 6)
	assert.True(t, strings.HasSuffix(expanded, "tmp"))

	expanded, err = util.ExpandTilde("/nothing/to/expand")
	assert.Nil(t, err)
	assert.Equal(t, "/nothing/to/expand", expanded)
}

func TestLooksSafeToDelete(t *testing.T) {
	assert.True(t, util.LooksSafeToDelete("/mnt/apt/data/some_dir", 15, 3))
	assert.False(t, util.LooksSafeToDelete("/usr/local", 12, 3))
}

func TestCopyFile(t *testing.T) {
	src := util.PathToUnitTestBag("example.edu.sample_good.tar")
	dest := filepath.Join(util.ProjectRoot(), "example.edu.sample_good.tar")
	_, err := util.CopyFile(dest, src)
	defer os.Remove(dest)
	assert.Nil(t, err)
}

func TestHasValidExtensionForMimeType(t *testing.T) {
	okFiles := map[string]string{
		"file.7z":     "application/x-7z-compressed",
		"file.7Z":     "application/x-7z-compressed",
		"file.tar":    "application/tar",
		"file2.tar":   "application/x-tar",
		"file.zip":    "application/zip",
		"file.gzip":   "application/gzip",
		"file.gz":     "application/gzip",
		"file.rar":    "application/x-rar-compressed",
		"file.tgz":    "application/tar+gzip",
		"file.tar.gz": "application/tar+gzip",
	}
	badFiles := map[string]string{
		"file.7z":  "application/tar",
		"file.tar": "application/x-7z-compressed",
		"file.zip": "application/gzip",
	}
	errFiles := map[string]string{
		"file.7z":  "application/binary",
		"file.tar": "application/kompressed",
	}

	for filename, mimeType := range okFiles {
		ok, err := util.HasValidExtensionForMimeType(filename, mimeType)
		assert.True(t, ok)
		assert.Nil(t, err)
	}

	for filename, mimeType := range badFiles {
		ok, err := util.HasValidExtensionForMimeType(filename, mimeType)
		assert.False(t, ok)
		assert.Nil(t, err)
	}

	for filename, mimeType := range errFiles {
		ok, err := util.HasValidExtensionForMimeType(filename, mimeType)
		assert.False(t, ok)
		assert.NotNil(t, err)
	}
}

func TestRecursiveFileList(t *testing.T) {
	testdir := util.PathToTestData()
	files, err := util.RecursiveFileList(testdir, false)
	require.Nil(t, err)
	sample := []string{
		"test.edu.btr-wasabi-or.tar",
		"test.edu.btr_good_sha512.tar",
		"bag-info.txt",
		"manifest-sha256.txt",
	}

	for _, expected := range sample {
		found := false
		for _, xFileInfo := range files {
			if xFileInfo.Name() == expected {
				found = true
				break
			}
		}
		assert.True(t, found, "File %s was not in list", expected)
	}
}

func TestGetDirectoryStats(t *testing.T) {
	dir := filepath.Join(util.ProjectRoot(), "server", "views")
	dirStats := util.GetDirectoryStats(dir)
	require.Empty(t, dirStats.Error)
	assert.Equal(t, dir, dirStats.FullPath)
	assert.Equal(t, path.Base(dir), dirStats.BaseName)
	assert.True(t, dirStats.DirCount > 10)
	assert.True(t, dirStats.FileCount > dirStats.DirCount*2)
	assert.True(t, dirStats.TotalBytes > 40000)

	dirStats = util.GetDirectoryStats(filepath.Join(util.ProjectRoot(), "path-does", "not-exist"))
	assert.Contains(t, dirStats.Error, "no such file or directory")

	file := filepath.Join(util.ProjectRoot(), "server", "views", "partials", "nav.html")
	dirStats = util.GetDirectoryStats(file)
	assert.Empty(t, dirStats.Error)
	assert.Equal(t, file, dirStats.FullPath)
	assert.Equal(t, "nav.html", dirStats.BaseName)
	assert.Equal(t, 0, dirStats.DirCount)
	assert.Equal(t, 1, dirStats.FileCount)

	fileInfo, err := os.Stat(file)
	require.Nil(t, err)
	assert.Equal(t, fileInfo.Size(), dirStats.TotalBytes)
}

func TestListDirectory(t *testing.T) {
	dir := filepath.Join(util.ProjectRoot(), "util")
	files, err := util.ListDirectory(dir)
	require.Nil(t, err)
	require.NotEmpty(t, files)
	assert.True(t, len(files) > 10)
}

func TestListDirectoryWithSort(t *testing.T) {
	dir := filepath.Join(util.ProjectRoot())
	entries, err := util.ListDirectoryWithSort(dir)
	require.Nil(t, err)
	require.NotEmpty(t, entries)

	entryCount := len(entries)
	require.True(t, entryCount > 10)

	// Test that directories are first, in alpha order...
	assert.True(t, entries[0].IsDir())
	assert.True(t, strings.ToLower(entries[0].Name()) < strings.ToLower(entries[1].Name()))

	// ...followed by files in alpha order.
	last := entryCount - 1
	secondToLast := entryCount - 2
	assert.False(t, entries[last].IsDir())
	assert.True(t, strings.ToLower(entries[secondToLast].Name()) < strings.ToLower(entries[last].Name()))
}

func TestGetWindowsDrives(t *testing.T) {
	drives := util.GetWindowsDrives()
	if runtime.GOOS == "windows" {
		require.NotEmpty(t, drives)
		assert.Contains(t, drives, "C:\\")
	} else {
		assert.Empty(t, drives)
	}
}

func TestParseCSV(t *testing.T) {
	expectedHeaders := []string{
		"Bag-Name",
		"Root-Directory",
		"aptrust-info.txt/Title",
		"aptrust-info.txt/Description",
		"aptrust-info.txt/Access",
		"bag-info.txt/Source-Organization",
		"bag-info.txt/Custom-Tag",
		"custom-tag-file.txt/Tag-One",
		"custom-tag-file.txt/Tag-Two",
	}

	pathToCSVFile := filepath.Join(util.PathToTestData(), "files", "test_batch.csv")
	headers, records, err := util.ParseCSV(pathToCSVFile)
	require.Nil(t, err)
	assert.Equal(t, expectedHeaders, headers)
	assert.Equal(t, 3, len(records))

	// Make sure there are no empty values.
	for lineNumber, record := range records {
		for _, nameValuePair := range record.Items {
			assert.NotEmpty(t, nameValuePair.Value, lineNumber, nameValuePair.Name)
		}
	}

	// Spot check the first and last values of the last record
	lastRecord := records[2]

	bagName, found := lastRecord.FirstMatching("Bag-Name")
	assert.True(t, found, "Bag-Name")
	assert.Equal(t, "RunnerTestUtil", bagName.Value)

	tagTwo, found := lastRecord.FirstMatching("custom-tag-file.txt/Tag-Two")
	assert.True(t, found)
	assert.Equal(t, "Camembert", tagTwo.Value)
}
