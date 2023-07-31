package util

type DirectoryStats struct {
	RootIsFile bool
	FullPath   string
	BaseName   string
	DirCount   int
	Error      string
	FileCount  int
	TotalBytes int64
}
