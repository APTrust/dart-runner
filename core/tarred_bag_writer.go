package core

import (
	"archive/tar"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/APTrust/dart-runner/util"
)

type TarredBagWriter struct {
	OutputPath     string
	rootDirName    string
	tarWriter      *tar.Writer
	digestAlgs     []string
	rootDirCreated bool
}

func NewTarredBagWriter(outputPath string, digestAlgs []string) *TarredBagWriter {
	return &TarredBagWriter{
		OutputPath:     outputPath,
		rootDirName:    util.CleanBagName(filepath.Base(outputPath)),
		digestAlgs:     digestAlgs,
		rootDirCreated: false,
	}
}

// DigestAlgs returns a list of digest algoritms that the
// writer calculates as it writes. E.g. ["md5", "sha256"].
// These are defined in the contants package.
func (writer *TarredBagWriter) DigestAlgs() []string {
	return writer.digestAlgs
}

func (writer *TarredBagWriter) Open() error {
	tarFile, err := os.Create(writer.OutputPath)
	if err != nil {
		message := fmt.Sprintf("Error creating tar file: %v", err)
		Dart.Log.Error(message)
		return fmt.Errorf(message)
	}
	writer.tarWriter = tar.NewWriter(tarFile)
	return nil
}

func (writer *TarredBagWriter) Close() error {
	if writer.tarWriter != nil {
		return writer.tarWriter.Close()
	}
	return nil
}

func (writer *TarredBagWriter) initRootDir(uid, gid int) error {
	header := &tar.Header{
		Name:     writer.rootDirName,
		Size:     0,
		Mode:     0755,
		ModTime:  time.Now(),
		Uid:      uid,
		Gid:      gid,
		Typeflag: tar.TypeDir,
	}
	err := writer.tarWriter.WriteHeader(header)
	if err == nil {
		writer.rootDirCreated = true
	}
	return err
}

// AddFile as a file to a tar archive. Returns a map of checksums
// where key is the algorithm and value is the digest. E.g.
// checksums["md5"] = "0987654321"
func (writer *TarredBagWriter) AddFile(xFileInfo *util.ExtendedFileInfo, pathWithinArchive string) (map[string]string, error) {

	checksums := make(map[string]string)
	hashes := util.GetHashes(writer.digestAlgs)

	if writer.tarWriter == nil {
		message := "Underlying TarWriter is nil. Has it been opened?"
		Dart.Log.Error(message)
		return checksums, fmt.Errorf(message)
	}

	// This returns actual owner and group id on posix systems,
	// 0,0 on Windows.
	uid, gid := xFileInfo.OwnerAndGroup()
	if !writer.rootDirCreated {
		err := writer.initRootDir(uid, gid)
		if err != nil {
			Dart.Log.Errorf("Tar writer can't create root directory header: %v", err)
			return checksums, err
		}
	}

	header := &tar.Header{
		Name:    pathWithinArchive,
		Size:    xFileInfo.Size(),
		Mode:    int64(xFileInfo.Mode().Perm()),
		ModTime: xFileInfo.ModTime(),
		Uid:     uid,
		Gid:     gid,
	}

	// Note that because we support only files and directories.
	// BagIt files probably shouldn't contain links or devices.
	if xFileInfo.IsDir() {
		header.Typeflag = tar.TypeDir
		header.Size = 0
	} else {
		header.Typeflag = tar.TypeReg
	}

	// Write the header entry
	if err := writer.tarWriter.WriteHeader(header); err != nil {
		// Most likely error is archive/tar: write after close
		Dart.Log.Errorf("Tar writer can't write header: %v", err)
		return checksums, err
	}

	// For directory entries, there's no content to write,
	// so just stop here.
	if header.Typeflag == tar.TypeDir {
		return checksums, nil
	}

	// Open the file whose data we're going to add.
	file, err := os.Open(xFileInfo.FullPath)
	if err != nil {
		Dart.Log.Errorf("TarWriter can't open file %s: %v", xFileInfo.FullPath, err)
		return checksums, err
	}
	defer file.Close()

	// Copy the contents of the file into the tarWriter,
	// passing it through the hashes along the way.
	writers := make([]io.Writer, len(writer.digestAlgs)+1)
	for i, alg := range writer.digestAlgs {
		writers[i] = hashes[alg]
	}
	writers[len(writers)-1] = writer.tarWriter
	multiWriter := io.MultiWriter(writers...)
	bytesWritten, err := io.Copy(multiWriter, file)
	if bytesWritten != header.Size {
		message := fmt.Sprintf("TarWriter.addToArchive() copied only %d of %d bytes for file %s", bytesWritten, header.Size, xFileInfo.FullPath)
		Dart.Log.Error(message)
		return checksums, fmt.Errorf(message)
	}
	if err != nil {
		message := fmt.Sprintf("Error copying %s into tar archive: %v", xFileInfo.FullPath, err)
		Dart.Log.Error(message)
		return checksums, fmt.Errorf(message)
	}

	// Gather the checksums.
	for _, alg := range writer.digestAlgs {
		hash := hashes[alg]
		checksums[alg] = fmt.Sprintf("%x", hash.Sum(nil))
	}

	return checksums, nil
}
