package validation

import (
	"archive/tar"
	"crypto/md5"
	"crypto/sha1"
	"crypto/sha256"
	"crypto/sha512"
	"fmt"
	"io"

	"github.com/APTrust/dart-runner/constants"
	"github.com/APTrust/dart-runner/util"
)

// TarredBagReader reads a tarred BagIt file to collect metadata for
// validation and ingest processing. See ProcessNextEntry() below.
type TarredBagReader struct {
	validator *Validator
	reader    io.ReadCloser
	TarReader *tar.Reader
}

// NewTarredBagReader creates a new TarredBagReader.
//
// Param reader is an io.ReadCloser from which to read the tarred
// BagIt file.
//
// Param ingestObject contains info about the bag in the tarred BagIt
// file.
//
// Param tempDir should be the path to a directory in which the scanner
// can temporarily store files it extracts from the tarred bag. These
// files include manifests, tag manifests, and 2-3 tag files. All of the
// extracted files are text files, and they are typically small.
//
// Note that the TarredBagReader does NOT delete temp files when it's
// done. It stores the paths to the temp files in the TempFiles attribute
// (a string slice). The caller should process the temp files as it pleases,
// and then delete them using this object's DeleteTempFiles method.
//
// For an example of how to use this object, see the Run method in
// ingest/metadata_gatherer.go
func NewTarredBagReader(reader io.ReadCloser, validator *Validator) *TarredBagReader {
	return &TarredBagReader{
		reader:    reader,
		TarReader: tar.NewReader(reader),
	}
}

func (r *TarredBagReader) ScanMetadata() error {
	return nil
}

func (r *TarredBagReader) ScanPayload() error {
	return nil
}

func (r *TarredBagReader) processNextEntry() error {
	header, err := r.TarReader.Next()
	if err != nil {
		return err
	}
	if header.Typeflag == tar.TypeReg || header.Typeflag == tar.TypeRegA {
		return r.processFileEntry(header)
	}
	return nil
}

// Process a single file in the tarball.
func (r *TarredBagReader) processFileEntry(header *tar.Header) error {
	fileRecord, pathInBag, err := r.createFileRecord(header)
	if err != nil {
		return err
	}
	r.addChecksums(pathInBag, fileRecord)
	return nil
}

func (r *TarredBagReader) createFileRecord(header *tar.Header) (*FileRecord, string, error) {
	pathInBag, err := util.TarPathToBagPath(header.Name)
	if err != nil {
		return nil, "", err
	}
	fileRecord := NewFileRecord()
	fileType := util.BagFileType(pathInBag)
	switch fileType {
	case constants.FileTypeManifest:
		r.validator.PayloadManifests.Files[pathInBag] = fileRecord
	case constants.FileTypePayload:
		r.validator.PayloadFiles.Files[pathInBag] = fileRecord
	case constants.FileTypeTagManifest:
		r.validator.TagManifests.Files[pathInBag] = fileRecord
	default:
		r.validator.TagFiles.Files[pathInBag] = fileRecord
	}
	return fileRecord, pathInBag, nil
}

func (r *TarredBagReader) addChecksums(pathInBag string, fileRecord *FileRecord) error {
	md5Hash := md5.New()
	sha1Hash := sha1.New()
	sha256Hash := sha256.New()
	sha512Hash := sha512.New()
	writers := []io.Writer{
		md5Hash,
		sha1Hash,
		sha256Hash,
		sha512Hash,
	}
	multiWriter := io.MultiWriter(writers...)
	_, err := io.Copy(multiWriter, r.TarReader)
	if err != nil {
		return err
	}

	fileRecord.AddChecksum(constants.FileTypePayload, constants.AlgMd5,
		fmt.Sprintf("%x", md5Hash.Sum(nil)))
	fileRecord.AddChecksum(constants.FileTypePayload, constants.AlgSha1,
		fmt.Sprintf("%x", sha1Hash.Sum(nil)))
	fileRecord.AddChecksum(constants.FileTypePayload, constants.AlgSha256,
		fmt.Sprintf("%x", sha256Hash.Sum(nil)))
	fileRecord.AddChecksum(constants.FileTypePayload, constants.AlgSha512,
		fmt.Sprintf("%x", sha256Hash.Sum(nil)))

	return nil
}

// CloseReader closes the io.ReadCloser() that was passed into
// NewTarredBagReader. If you neglect this call in a long-running
// worker process, you'll run the system out of filehandles.
func (r *TarredBagReader) CloseReader() {
	if r.reader != nil {
		r.reader.Close()
	}
}
