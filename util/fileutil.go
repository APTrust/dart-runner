package util

import (
	"fmt"
	"io"
	"os"
	"os/user"
	"path"
	"path/filepath"
	"runtime"
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
		BaseName: path.Base(dir),
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
		extFileInfo := NewExtendedFileInfo(path.Join(dir, fileInfo.Name()), fileInfo)
		if entry.IsDir() {
			directories = append(directories, extFileInfo)
		} else {
			files = append(files, extFileInfo)
		}
	}
	allEntries := append(directories, files...)
	return allEntries, nil
}

// GetWindowsDrives returns a list of Windows drives.
func GetWindowsDrives() []string {
	drives := make([]string, 0)
	if runtime.GOOS == "windows" {
		for _, drive := range "ABCDEFGHIJKLMNOPQRSTUVWXYZ" {
			_, err := os.Stat(string(drive) + ":\\")
			if err == nil {
				drives = append(drives, string(drive))
			}
		}
	}
	return drives
}
