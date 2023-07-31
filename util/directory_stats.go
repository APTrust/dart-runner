package util

type DirectoryStats struct {
	RootIsFile bool
	DirCount   int
	Error      string
	FileCount  int
	TotalBytes int64
}
