package util

import (
	"os"
)

// ExtendedFileInfo adds some data to os.FileInfo including
// FullPath, Owner() and Group(). The latter two are platform
// specific. They return zero on Windows and actual ids on
// posix systems.
type ExtendedFileInfo struct {
	os.FileInfo
	FullPath string
}

// NewExtendedFileInfo creates a new ExtendedFileInfo object.
// This takes two params because RecursiveFileList stats the
// files for us. Otherwise, we'd do it in the constructor.
func NewExtendedFileInfo(path string, fileInfo os.FileInfo) *ExtendedFileInfo {
	return &ExtendedFileInfo{
		FileInfo: fileInfo,
		FullPath: path,
	}
}
