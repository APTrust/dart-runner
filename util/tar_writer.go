package util

import (
	"archive/tar"
	"fmt"
	"io"
	"os"
)

type TarWriter struct {
	PathToTarFile string
	tarWriter     *tar.Writer
}

func NewTarWriter(pathToTarFile string) *TarWriter {
	return &TarWriter{
		PathToTarFile: pathToTarFile,
	}
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

// AddFile as a file to a tar archive.
func (writer *TarWriter) AddFile(xFileInfo *ExtendedFileInfo, pathWithinArchive string) error {
	if writer.tarWriter == nil {
		return fmt.Errorf("Underlying TarWriter is nil. Has it been opened?")
	}
	// This returns actual owner and group id on posix systems,
	// 0,0 on Windows.
	uid, gid := xFileInfo.OwnerAndGroup()
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
		return err
	}

	// For directory entries, there's no content to write,
	// so just stop here.
	if header.Typeflag == tar.TypeDir {
		return nil
	}

	// Open the file whose data we're going to add.
	file, err := os.Open(xFileInfo.FullPath)
	defer file.Close()
	if err != nil {
		return err
	}

	// Copy the contents of the file into the tarWriter.
	bytesWritten, err := io.Copy(writer.tarWriter, file)
	if bytesWritten != header.Size {
		return fmt.Errorf("addToArchive() copied only %d of %d bytes for file %s",
			bytesWritten, header.Size, xFileInfo.FullPath)
	}
	if err != nil {
		return fmt.Errorf("Error copying %s into tar archive: %v",
			xFileInfo.FullPath, err)
	}

	return nil
}
