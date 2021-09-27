// +build windows

package util

// OwnerAndGrooup returns the file's owner id and group id
// on posix systems. Returns zero, zero on Windows.
func (fi *ExtendedFileInfo) OwnerAndGroup() (int, int) {
	return 0, 0
}
