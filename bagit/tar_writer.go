package bagit

import (
	"archive/tar"
	"fmt"
	"io"
	"os"
	"path"
	"time"

	"github.com/APTrust/dart-runner/util"
)

type TarWriter struct {
	PathToTarFile  string
	rootDirName    string
	tarWriter      *tar.Writer
	digestAlgs     []string
	rootDirCreated bool
}

func NewTarWriter(pathToTarFile string, digestAlgs []string) *TarWriter {
	return &TarWriter{
		PathToTarFile:  pathToTarFile,
		rootDirName:    util.CleanBagName(path.Base(pathToTarFile)),
		digestAlgs:     digestAlgs,
		rootDirCreated: false,
	}
}

// DigestAlgs returns a list of digest algoritms that the
// writer calculates as it writes. E.g. ["md5", "sha256"].
// These are defined in the contants package.
func (writer *TarWriter) DigestAlgs() []string {
	return writer.digestAlgs
}

func (writer *TarWriter) Open() error {
	tarFile, err := os.Create(writer.PathToTarFile)
	if err != nil {
		return fmt.Errorf("Error creating tar file: %v", err)
	}
	writer.tarWriter = tar.NewWriter(tarFile)
	return nil
}

func (writer *TarWriter) Close() error {
	if writer.tarWriter != nil {
		return writer.tarWriter.Close()
	}
	return nil
}

func (writer *TarWriter) initRootDir(uid, gid int) error {
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
func (writer *TarWriter) AddFile(xFileInfo *util.ExtendedFileInfo, pathWithinArchive string) (map[string]string, error) {

	checksums := make(map[string]string)
	hashes := util.GetHashes(writer.digestAlgs)

	if writer.tarWriter == nil {
		return checksums, fmt.Errorf("Underlying TarWriter is nil. Has it been opened?")
	}

	// This returns actual owner and group id on posix systems,
	// 0,0 on Windows.
	uid, gid := xFileInfo.OwnerAndGroup()
	if !writer.rootDirCreated {
		err := writer.initRootDir(uid, gid)
		if err != nil {
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
		return checksums, err
	}

	// For directory entries, there's no content to write,
	// so just stop here.
	if header.Typeflag == tar.TypeDir {
		return checksums, nil
	}

	// Open the file whose data we're going to add.
	file, err := os.Open(xFileInfo.FullPath)
	defer file.Close()
	if err != nil {
		return checksums, err
	}

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
		return checksums, fmt.Errorf("addToArchive() copied only %d of %d bytes for file %s",
			bytesWritten, header.Size, xFileInfo.FullPath)
	}
	if err != nil {
		return checksums, fmt.Errorf("Error copying %s into tar archive: %v",
			xFileInfo.FullPath, err)
	}

	// Gather the checksums.
	for _, alg := range writer.digestAlgs {
		hash := hashes[alg]
		checksums[alg] = fmt.Sprintf("%x", hash.Sum(nil))
	}

	return checksums, nil
}
