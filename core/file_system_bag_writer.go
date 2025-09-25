package core

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/APTrust/dart-runner/util"
)

type FileSystemBagWriter struct {
	outputPath     string
	rootDirName    string
	digestAlgs     []string
	rootDirCreated bool
}

func NewFileSystemBagWriter(outputPath string, digestAlgs []string) *FileSystemBagWriter {
	return &FileSystemBagWriter{
		outputPath:     outputPath,
		rootDirName:    util.CleanBagName(filepath.Base(outputPath)),
		digestAlgs:     digestAlgs,
		rootDirCreated: false,
	}
}

// DigestAlgs returns a list of digest algoritms that the
// writer calculates as it writes. E.g. ["md5", "sha256"].
// These are defined in the contants package.
func (writer *FileSystemBagWriter) DigestAlgs() []string {
	return writer.digestAlgs
}

func (writer *FileSystemBagWriter) OutputPath() string {
	return writer.outputPath
}

func (writer *FileSystemBagWriter) Open() error {
	err := os.MkdirAll(writer.outputPath, 0755)
	if err == nil {
		writer.rootDirCreated = true
	}
	return err
}

func (writer *FileSystemBagWriter) Close() error {
	// No-op. This is here to satisfy the BagWriter interface.
	return nil
}

// AddFile as a file to a tar archive. Returns a map of checksums
// where key is the algorithm and value is the digest. E.g.
// checksums["md5"] = "0987654321"
func (writer *FileSystemBagWriter) AddFile(xFileInfo *util.ExtendedFileInfo, pathWithinArchive string) (map[string]string, error) {
	absPath := filepath.Join(writer.outputPath, pathWithinArchive)
	checksums := make(map[string]string)
	hashes := util.GetHashes(writer.digestAlgs)

	// For directory entries, there's no content to write.
	// Make sure the directory exists, then return.
	if xFileInfo.IsDir() {
		return checksums, os.MkdirAll(absPath, 0755)
	}

	err := os.MkdirAll(filepath.Dir(absPath), 0755)
	if err != nil {
		return checksums, err
	}

	// Open the file whose data we're going to add.
	file, err := os.Open(xFileInfo.FullPath)
	if err != nil {
		Dart.Log.Errorf("FileSystemBagWriter can't open source file %s: %v", xFileInfo.FullPath, err)
		return checksums, err
	}
	defer file.Close()

	// Create a file inside the bag into which we'll copy the contents
	// of file.
	outfile, err := os.Create(absPath)
	if err != nil {
		Dart.Log.Errorf("FileSystemBagWriter can't open destination file %s: %v", xFileInfo.FullPath, err)
		return checksums, err
	}
	defer outfile.Close()

	// Copy the contents of the file into the tarWriter,
	// passing it through the hashes along the way.
	writers := make([]io.Writer, len(writer.digestAlgs)+1)
	for i, alg := range writer.digestAlgs {
		writers[i] = hashes[alg]
	}
	writers[len(writers)-1] = outfile
	multiWriter := io.MultiWriter(writers...)
	bytesWritten, err := io.Copy(multiWriter, file)
	if bytesWritten != xFileInfo.Size() {
		message := fmt.Sprintf("FileSystemBagWriter.addToArchive() copied only %d of %d bytes for file %s", bytesWritten, xFileInfo.Size(), xFileInfo.FullPath)
		Dart.Log.Error(message)
		return checksums, fmt.Errorf(message)
	}
	if err != nil {
		message := fmt.Sprintf("Error copying %s into tar archive: %v", xFileInfo.FullPath, err)
		Dart.Log.Error(message)
		return checksums, fmt.Errorf(message)
	}

	// This returns actual owner and group id on posix systems,
	// 0,0 on Windows.
	uid, gid := xFileInfo.OwnerAndGroup()
	if uid != 0 {
		os.Chown(absPath, uid, gid)
	}
	// We should add proper aTime to xFileInfo.
	os.Chtimes(absPath, xFileInfo.ModTime(), xFileInfo.ModTime())

	// Gather the checksums.
	for _, alg := range writer.digestAlgs {
		hash := hashes[alg]
		checksums[alg] = fmt.Sprintf("%x", hash.Sum(nil))
	}

	return checksums, nil
}
