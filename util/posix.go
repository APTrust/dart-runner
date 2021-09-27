// +build !windows

package util

import (
	"syscall"
)

// OwnerAndGrooup returns the file's owner id and group id
// on posix systems. Returns zero, zero on Windows.
func (fi *ExtendedFileInfo) OwnerAndGroup() (uid int, gid int) {
	systat := fi.FileInfo.Sys().(*syscall.Stat_t)
	if systat != nil {
		uid = int(systat.Uid)
		gid = int(systat.Gid)
	}
	return uid, gid
}
