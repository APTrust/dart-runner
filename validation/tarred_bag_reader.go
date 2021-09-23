package validation

import (
	"archive/tar"
	"crypto/md5"
	"crypto/sha1"
	"crypto/sha256"
	"crypto/sha512"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/APTrust/dart-runner/constants"
	"github.com/APTrust/dart-runner/util"
)

// TarredBagReader reads a tarred BagIt file to collect metadata for
// validation and ingest processing. See ProcessNextEntry() below.
type TarredBagReader struct {
	validator *Validator
	reader    io.ReadSeekCloser
	tarReader *tar.Reader
}

// NewTarredBagReader creates a new TarredBagReader.
func NewTarredBagReader(validator *Validator) (*TarredBagReader, error) {
	file, err := os.Open(validator.PathToBag)
	if err != nil {
		return nil, err
	}
	return &TarredBagReader{
		reader:    file,
		validator: validator,
	}, nil
}

func (r *TarredBagReader) ScanMetadata() error {
	// Get a list of all files and create a FileRecord for each.
	// Parse all payload and tag manifests.
	// Parse all tag files.
	r.reader.Seek(0, io.SeekStart)
	r.tarReader = tar.NewReader(r.reader)
	for {
		err := r.processMetaEntry()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
	}
	return nil
}

func (r *TarredBagReader) ScanPayload() error {
	r.reader.Seek(0, io.SeekStart)
	r.tarReader = tar.NewReader(r.reader)

	return nil
}

func (r *TarredBagReader) processMetaEntry() error {
	header, err := r.tarReader.Next()
	if err != nil {
		return err
	}
	if header.Typeflag == tar.TypeReg || header.Typeflag == tar.TypeRegA {
		pathInBag, err := util.TarPathToBagPath(header.Name)
		if err != nil {
			return err
		}
		fileType := util.BagFileType(pathInBag)
		switch fileType {
		case constants.FileTypeManifest:
			r.addOrUpdateFileRecord(r.validator.PayloadManifests, pathInBag, header.Size)

			err = r.parseManifest(pathInBag, r.validator.PayloadFiles)
		case constants.FileTypeTagManifest:
			r.addOrUpdateFileRecord(r.validator.TagManifests, pathInBag, header.Size)
			err = r.parseManifest(pathInBag, r.validator.TagFiles)
		case constants.FileTypeTag:
			r.addOrUpdateFileRecord(r.validator.TagFiles, pathInBag, header.Size)
			r.parseTagFile(pathInBag)
		default:
			// skip payload files for now
		}
	}
	return err
}

func (r *TarredBagReader) processNextEntry() error {
	header, err := r.tarReader.Next()
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
	err := r.ensureFileRecord(header)
	if err != nil {
		return err
	}
	return nil
}

func (r *TarredBagReader) ensureFileRecord(header *tar.Header) error {
	pathInBag, err := util.TarPathToBagPath(header.Name)
	if err != nil {
		return err
	}
	fileType := util.BagFileType(pathInBag)
	switch fileType {
	case constants.FileTypeManifest:
		r.addOrUpdateFileRecord(r.validator.PayloadManifests, pathInBag, header.Size)
		err = r.parseManifest(pathInBag, r.validator.PayloadFiles)
		if err != nil {
			return err
		}
	case constants.FileTypePayload:
		r.addOrUpdateFileRecord(r.validator.PayloadFiles, pathInBag, header.Size)
	case constants.FileTypeTagManifest:
		r.addOrUpdateFileRecord(r.validator.TagManifests, pathInBag, header.Size)
		err = r.parseManifest(pathInBag, r.validator.TagFiles)
		if err != nil {
			return err
		}
	default:
		r.addOrUpdateFileRecord(r.validator.TagFiles, pathInBag, header.Size)
		r.parseTagFile(pathInBag)
	}
	return nil
}

func (r *TarredBagReader) addOrUpdateFileRecord(fileMap *FileMap, pathInBag string, size int64) *FileRecord {
	// avoid multiple hash lookups
	fileRecord := fileMap.Files[pathInBag]
	if fileRecord == nil {
		fileRecord = NewFileRecord()
		fileMap.Files[pathInBag] = fileRecord
	}
	// if we encounter this file first in a manifest entry,
	// we don't know its size, so we pass -1, which we don't
	// want to record because it's invalid. Record only valid
	// sizes.
	if size >= 0 {
		fileRecord.Size = size
	}
	return fileRecord
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
	_, err := io.Copy(multiWriter, r.tarReader)
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

// Parse entries in manifest and add them to the right file map.
// Payload manifest entries are added to the map of payload files.
// Tag manifest entries are added to the map of tag files.
func (r *TarredBagReader) parseManifest(pathInBag string, fileMap *FileMap) error {
	alg, err := util.AlgorithmFromManifestName(pathInBag)
	if err != nil {
		return err
	}
	entries, err := ParseManifest(r.tarReader)
	if err != nil {
		return err
	}
	for filePath, digest := range entries {
		fileRecord := r.addOrUpdateFileRecord(fileMap, filePath, -1)
		fileRecord.AddChecksum(constants.FileTypeManifest, alg, digest)
	}
	return nil
}

func (r *TarredBagReader) parseTagFile(pathInBag string) {
	if !strings.HasSuffix(pathInBag, ".txt") {
		return
	}
	tags, err := ParseTagFile(r.tarReader, pathInBag)
	if err != nil {
		r.validator.UnparsableTagFiles = append(r.validator.UnparsableTagFiles, pathInBag)
	} else {
		r.validator.Tags = append(r.validator.Tags, tags...)
	}
}

// CloseReader closes the io.ReadCloser() that was passed into
// NewTarredBagReader. If you neglect this call in a long-running
// worker process, you'll run the system out of filehandles.
func (r *TarredBagReader) CloseReader() {
	if r.reader != nil {
		r.reader.Close()
	}
}
