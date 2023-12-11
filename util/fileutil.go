package util

import (
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"os/user"
	"path"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
)

// Returns true if the file at path exists, false if not.
func FileExists(path string) bool {
	_, err := os.Stat(path)
	if err != nil && os.IsNotExist(err) {
		return false
	}
	return true
}

// Returns true if path is a directory.
func IsDirectory(path string) bool {
	stat, err := os.Stat(path)
	if err != nil && os.IsNotExist(err) {
		return false
	}
	return stat.IsDir()
}

// Expands the tilde in a directory path to the current
// user's home directory. For example, on Linux, ~/data
// would expand to something like /home/josie/data
func ExpandTilde(filePath string) (string, error) {
	if !strings.Contains(filePath, "~") {
		return filePath, nil
	}
	usr, err := user.Current()
	if err != nil {
		return "", err
	}
	homeDir := usr.HomeDir + "/"
	expandedDir := strings.Replace(filePath, "~/", homeDir, 1)
	return expandedDir, nil
}

// Returns true if the path specified by dir has at least minLength
// characters and at least minSeparators path separators. This is
// for testing paths you want pass into os.RemoveAll(), so you don't
// wind up deleting "/" or "/etc" or something catastrophic like that.
func LooksSafeToDelete(dir string, minLength, minSeparators int) bool {
	separator := string(os.PathSeparator)
	separatorCount := (len(dir) - len(strings.Replace(dir, separator, "", -1)))
	return len(dir) >= minLength && separatorCount >= minSeparators
}

// CopyFile copies a file from src to dest
func CopyFile(dest, src string) (int64, error) {
	finfo, err := os.Stat(src)
	if err != nil {
		return int64(0), err
	}
	from, err := os.Open(src)
	if err != nil {
		return int64(0), err
	}
	defer from.Close()

	to, err := os.OpenFile(dest, os.O_RDWR|os.O_CREATE, finfo.Mode())
	if err != nil {
		return int64(0), err
	}
	defer to.Close()
	return io.Copy(to, from)
}

// ReadFile reads an entire file into a byte array.
func ReadFile(filepath string) ([]byte, error) {
	file, err := os.Open(filepath)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	return io.ReadAll(file)
}

func HasValidExtensionForMimeType(filename, mimeType string) (bool, error) {
	var err error
	valid := false
	ext := strings.ToLower(path.Ext(filename))
	switch mimeType {
	case "application/x-7z-compressed":
		valid = (ext == ".7z")
	case "application/tar", "application/x-tar":
		valid = (ext == ".tar")
	case "application/zip":
		valid = (ext == ".zip")
	case "application/gzip":
		valid = (ext == ".gz" || ext == ".gzip")
	case "application/x-rar-compressed":
		valid = (ext == ".rar")
	case "application/tar+gzip":
		valid = (ext == ".tgz" || strings.HasSuffix(filename, ".tar.gz"))
	default:
		if !valid {
			err = fmt.Errorf("dart-runner doesn't know about serialization type %s", mimeType)
		}
	}
	return valid, err
}

// RecursiveFileList a list of all items inside of dir.
// If includeIrregulars is false, this will NOT return links, pipes,
// devices, or anything else besides regular files and directories.
//
// We generally do want to omit items like symlinks, pipes, etc.
// when bagging because we cannot bag them.
func RecursiveFileList(dir string, includeIrregulars bool) ([]*ExtendedFileInfo, error) {
	files := make([]*ExtendedFileInfo, 0)
	err := filepath.Walk(dir, func(filePath string, f os.FileInfo, err error) error {
		if f.Mode().IsRegular() || f.Mode().IsDir() || includeIrregulars {
			files = append(files, NewExtendedFileInfo(filePath, f))
		}
		return nil
	})
	return files, err
}

// GetDirectoryStats returns the file count, directory count and total
// number of bytes found recursively under directory dir.
func GetDirectoryStats(dir string) *DirectoryStats {
	dirStats := &DirectoryStats{
		FullPath: dir,
		BaseName: filepath.Base(dir),
	}
	rootStat, err := os.Stat(dir)
	if err != nil {
		dirStats.Error = err.Error()
		return dirStats
	}
	dirStats.RootIsFile = !rootStat.IsDir()
	items, err := RecursiveFileList(dir, false)
	if err != nil {
		dirStats.Error = err.Error()
		return dirStats
	}
	for _, item := range items {
		if item.FileInfo.IsDir() {
			dirStats.DirCount++
		} else {
			dirStats.FileCount++
			dirStats.TotalBytes += item.FileInfo.Size()
		}
	}
	return dirStats
}

// ListDirectory returns a list of directory contents one level deep. It does
// not recurse. The list will contain directories first, followed by files.
func ListDirectory(dir string) ([]*ExtendedFileInfo, error) {
	dirEntries, err := os.ReadDir(dir)
	if err != nil {
		return nil, nil
	}
	files := make([]*ExtendedFileInfo, 0)
	directories := make([]*ExtendedFileInfo, 0)
	for _, entry := range dirEntries {
		fileInfo, err := entry.Info()
		if err != nil {
			return nil, nil
		}
		extFileInfo := NewExtendedFileInfo(filepath.Join(dir, fileInfo.Name()), fileInfo)
		if entry.IsDir() {
			directories = append(directories, extFileInfo)
		} else {
			files = append(files, extFileInfo)
		}
	}
	allEntries := append(directories, files...)
	return allEntries, nil
}

// ListDirectoriesWithSort returns a list of all items in the specified
// directory, showing folders first, then files. Folders and files are
// sorted in alpha order, case insensitive.
func ListDirectoryWithSort(dir string) ([]*ExtendedFileInfo, error) {
	entries, err := ListDirectory(dir)
	if err != nil {
		return entries, err
	}
	directories := make([]*ExtendedFileInfo, 0)
	files := make([]*ExtendedFileInfo, 0)
	for _, exFileInfo := range entries {
		if exFileInfo.IsDir() {
			directories = append(directories, exFileInfo)
		} else {
			files = append(files, exFileInfo)
		}
	}
	sort.Slice(directories, func(i, j int) bool {
		this := directories[i]
		that := directories[j]
		return strings.ToLower(this.Name()) < strings.ToLower(that.Name())
	})
	sort.Slice(files, func(i, j int) bool {
		this := files[i]
		that := files[j]
		return strings.ToLower(this.Name()) < strings.ToLower(that.Name())
	})
	return append(directories, files...), nil
}

// GetWindowsDrives returns a list of Windows drives.
func GetWindowsDrives() []string {
	drives := make([]string, 0)
	if runtime.GOOS == "windows" {
		for _, drive := range "ABCDEFGHIJKLMNOPQRSTUVWXYZ" {
			fullDriveName := string(drive) + ":\\"
			_, err := os.Stat(fullDriveName)
			if err == nil {
				drives = append(drives, fullDriveName)
			}
		}
	}
	return drives
}

// ParseCSV parses the CSV file at pathToCSVFile and returns the following:
//
// 1. A slice of strings containing the entries in the first line of the
// file. These are assumed to be headers / field names.
//
// 2. A slice of url.Values objects in which names are column headers
// and values are the values parsed from one line of the file. We use
// url.Values because, unlike a map, it preserves the order of the fields
// and allows us to have multiple values per key. This is essential as
// the BagIt spec allows a tag value to be specified multiple times.
//
// 3. An error, which will typically be one of: "file does not exist",
// "file can't be read (permissions)" or "csv parse error".
//
// This will not parse correctly if the first line of the file does not
// contain headers.
func ParseCSV(pathToCSVFile string) ([]string, []*NameValuePairList, error) {
	f, err := os.Open(pathToCSVFile)
	if err != nil {
		return nil, nil, err
	}
	reader := csv.NewReader(f)
	records, err := reader.ReadAll()
	if err != nil {
		return nil, nil, err
	}
	if len(records) < 2 {
		return nil, nil, fmt.Errorf("csv file contains no records")
	}

	// allRecords will hold one record for each line
	// in the CSV file (minus the first line, which contains headers)
	allRecords := make([]*NameValuePairList, 0)

	// First line contains headers. It may also contain a
	// byte order marker (BOM) if the CSV file was saved
	// from Excel. We need to remove the BOM and other
	// non-printables from the field names, so we call
	// StripNonPrintable().
	fieldNames := records[0]
	for i, _ := range fieldNames {
		fieldNames[i] = StripNonPrintable(fieldNames[i])
	}

	// Now make one []NameValuePair list for every record
	// in the file.
	for i := 1; i < len(records); i++ {
		currentRecord := records[i]
		record := NewNameValuePairList()
		lastField := Min(len(fieldNames), len(currentRecord))
		for j := 0; j < lastField; j++ {
			name := fieldNames[j]
			value := currentRecord[j]
			record.Add(name, value)
		}
		allRecords = append(allRecords, record)
	}
	return fieldNames, allRecords, nil
}
