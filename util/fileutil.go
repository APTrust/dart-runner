package util

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/user"
	"path"
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
	if strings.Index(filePath, "~") < 0 {
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
	return ioutil.ReadAll(file)
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
