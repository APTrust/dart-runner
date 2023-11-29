package core

import (
	"archive/tar"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/APTrust/dart-runner/constants"
	"github.com/APTrust/dart-runner/util"
)

// TarredBagReader reads a tarred BagIt file to collect metadata for
// bagit and ingest processing. See ProcessNextEntry() below.
type TarredBagReader struct {
	validator *Validator
	reader    io.ReadSeekCloser
	tarReader *tar.Reader
}

// NewTarredBagReader creates a new TarredBagReader.
func NewTarredBagReader(validator *Validator) (*TarredBagReader, error) {
	file, err := os.Open(validator.PathToBag)
	if err != nil {
		Dart.Log.Errorf("TarredBagReader can't open file %s: %v", validator.PathToBag, err)
		return nil, err
	}
	return &TarredBagReader{
		reader:    file,
		validator: validator,
	}, nil
}

// ScanMetadata does the following:
//
// * gets a list of all files and creates a FileRecord for each
// * parses all payload and tag manifests
// * parses all parsable tag files
func (r *TarredBagReader) ScanMetadata() error {
	r.reader.Seek(0, io.SeekStart)
	r.tarReader = tar.NewReader(r.reader)
	for {
		err := r.processMetaEntry()
		if err == io.EOF {
			Dart.Log.Debugf("TarredBagReader.ScanMetadata finished reading metadata in %s", r.validator.PathToBag)
			break
		}
		if err != nil {
			Dart.Log.Errorf("TarredBagReader.ScanMetadata error reading %s: %v", r.validator.PathToBag, err)
			return err
		}
	}
	return nil
}

// ScanPayload scans the entire bad, adding checksums for all files.
func (r *TarredBagReader) ScanPayload() error {
	r.reader.Seek(0, io.SeekStart)
	r.tarReader = tar.NewReader(r.reader)
	for {
		err := r.processPayloadEntry()
		if err == io.EOF {
			Dart.Log.Debugf("TarredBagReader.ScanPayload finished reading payload in %s", r.validator.PathToBag)
			break
		}
		if err != nil {
			Dart.Log.Errorf("TarredBagReader.ScanPayload error reading %s: %v", r.validator.PathToBag, err)
			return err
		}
	}
	r.mergePayloadManifestChecksums()
	return nil
}

// Because payload manifests may have entries in tag manifest
// files, we need to make sure their file records and checksums
// appear in TagFiles map as well as the PayloadManifests map.
func (r *TarredBagReader) mergePayloadManifestChecksums() {
	for name, fileRecord := range r.validator.PayloadManifests.Files {
		tagFileRecord := r.validator.TagFiles.Files[name]
		if tagFileRecord != nil {
			tagFileRecord.Size = fileRecord.Size
			for _, cs := range fileRecord.Checksums {
				tagFileRecord.Checksums = append(tagFileRecord.Checksums, cs)
			}
		}
	}
}

// processMetaEntry parses manifests and tag files.
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
			err = r.parseManifest(pathInBag, r.validator.PayloadFiles)
		case constants.FileTypeTagManifest:
			err = r.parseManifest(pathInBag, r.validator.TagFiles)
		case constants.FileTypeTag:
			r.parseTagFile(pathInBag)
		}
		fileMap := r.validator.MapForPath(pathInBag)
		r.addOrUpdateFileRecord(fileMap, pathInBag, header.Size)
	}
	return err
}

// processPayloadEntry adds files and checksums to our validator.
// See ensureFileRecord below.
func (r *TarredBagReader) processPayloadEntry() error {
	header, err := r.tarReader.Next()
	if err != nil {
		return err
	}
	if header.Typeflag == tar.TypeReg || header.Typeflag == tar.TypeRegA {
		err = r.ensureFileRecord(header)
	}
	return nil
}

// ensureFileRecord makes sure we have a FileRecord in the right
// FileMap. It also calculates and stores the required checksums
// for the file.
func (r *TarredBagReader) ensureFileRecord(header *tar.Header) error {
	pathInBag, err := util.TarPathToBagPath(header.Name)
	if err != nil {
		Dart.Log.Errorf("TarredBagReader: Can't convert header path %s to bag path: %v", header.Name, err)
		return err
	}
	fileMap := r.validator.MapForPath(pathInBag)
	fileRecord := r.addOrUpdateFileRecord(fileMap, pathInBag, header.Size)

	fileType := util.BagFileType(pathInBag)
	var algs []string
	if fileType == constants.FileTypePayload {
		algs, err = r.validator.PayloadManifestAlgs()
	} else {
		algs, err = r.validator.TagManifestAlgs()
	}
	if err != nil {
		Dart.Log.Errorf("TarredBagReader.ensureFileRecord for %s: %v", pathInBag, err)
		return err
	}
	return r.addChecksums(pathInBag, fileRecord, algs)
}

// addOrUpdateFileRecord adds or updates a FileRecord in fileMap.
// We call this when scanning manifests and when scanning the payload
// because some files may exist in one but not the other. If that's
// the case, we want to know because mismatched file lists can mean
// the bag is invalid.
func (r *TarredBagReader) addOrUpdateFileRecord(fileMap *FileMap, pathInBag string, size int64) *FileRecord {
	// avoid multiple map lookups
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

// addChecksums calculates checksums on a file stream and adds those
// checksums to the FileRecord. We calculate one checksum for each
// known manifest algorithm. For example, if the  bag has md5, sha1,
// sha256, and sha512 manifests, we'll calculate all those checksums.
// If it has only md5 and sha256, we'll calculate just those two.
//
// We use a MultiWriter to calculate all of a file's checksums in a
// single read.
func (r *TarredBagReader) addChecksums(pathInBag string, fileRecord *FileRecord, algs []string) error {

	// Get a hash for each of the digest algorithms we need
	// to calculate (md5, sha256, etc)
	hashes := util.GetHashes(algs)

	// Hashes implement io.Write. We'll write our file stream
	// through all of them at once.
	writers := make([]io.Writer, len(hashes))
	for i, alg := range algs {
		writers[i] = hashes[alg]
	}

	multiWriter := io.MultiWriter(writers...)
	_, err := io.Copy(multiWriter, r.tarReader)
	if err != nil {
		Dart.Log.Errorf("TarredBagReader error adding checksums for file %s: %v", pathInBag, err)
		return err
	}

	// Record where the checksum came from: tag file
	// or payload file. In this context, manifests count
	// as tag files because their checksums may appear
	// in tag manifests.
	fileType := util.BagFileType(pathInBag)
	if strings.Contains(fileType, "manifest") {
		fileType = constants.FileTypeTag
	}

	// For each hash we calculated, add a checksum to the
	// file record.
	for _, alg := range algs {
		digest := fmt.Sprintf("%x", hashes[alg].Sum(nil))
		fileRecord.AddChecksum(fileType, alg, digest)
	}

	return nil
}

// parseManifest parses manifest entries in manifest and adds them
// to the right file map. Payload manifest entries are added to the
// map of payload files. Tag manifest entries are added to the map
// of tag files.
func (r *TarredBagReader) parseManifest(pathInBag string, fileMap *FileMap) error {
	alg, err := util.AlgorithmFromManifestName(pathInBag)
	if err != nil {
		Dart.Log.Errorf("TarredBagReader.parseManifest error getting algs for %s: %v", pathInBag, err)
		return err
	}
	entries, err := ParseManifest(r.tarReader)
	if err != nil {
		Dart.Log.Errorf("TarredBagReader.parseManifest error parsing entries for %s: %v", pathInBag, err)
		return err
	}
	for filePath, digest := range entries {
		fileRecord := r.addOrUpdateFileRecord(fileMap, filePath, -1)
		fileRecord.AddChecksum(constants.FileTypeManifest, alg, digest)
	}
	return nil
}

// parseTagFile tries to parse a tag file if it has a .txt extension.
// It skips other file formats. If it can't parse a .txt tag file, it
// adds that file to the list of unparsables. This may or may not be
// an error, depending on the BagIt profile. The validator will determine
// that later. If a required tag file is unparsable, that's an error.
// If the profile says no tags from that file are required, it's not
// an error.
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

// Close closes the underlying reader (the file).
// If you neglect this call in a long-running
// worker process, you'll run the system out of filehandles.
func (r *TarredBagReader) Close() {
	if r.reader != nil {
		r.reader.Close()
	}
}
