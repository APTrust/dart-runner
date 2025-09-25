package core

import (
	"fmt"
	"io"
	"os"
	"runtime"
	"strings"

	"github.com/APTrust/dart-runner/constants"
	"github.com/APTrust/dart-runner/util"
)

// FileSystemBagReader reads loose (unserialized) bags.
type FileSystemBagReader struct {
	validator *Validator
	fileList  []*util.ExtendedFileInfo
}

// NewFileSystemBagReader returns a reader that can parse loose
// (unserialized) bags.
func NewFileSystemBagReader(validator *Validator) (*FileSystemBagReader, error) {
	fileList, err := util.RecursiveFileList(validator.PathToBag, false)
	return &FileSystemBagReader{
		validator: validator,
		fileList:  fileList,
	}, err
}

// ScanMetadata does the following:
//
// * gets a list of all files and creates a FileRecord for each
// * parses all payload and tag manifests
// * parses all parsable tag files
func (r *FileSystemBagReader) ScanMetadata() error {
	for _, xFileInfo := range r.fileList {
		if xFileInfo.IsDir() {
			continue
		}
		err := r.processMetaEntry(xFileInfo)
		if err != nil {
			Dart.Log.Errorf("FileSystemBagReader.ScanMetadata error reading %s: %v", xFileInfo.FullPath, err)
			return err
		}
	}
	return nil
}

// ScanPayload scans the entire bag, adding checksums for all files.
func (r *FileSystemBagReader) ScanPayload() error {
	for _, xFileInfo := range r.fileList {
		err := r.processPayloadEntry(xFileInfo)
		if err == io.EOF {
			Dart.Log.Debugf("FileSystemBagReader.ScanPayload finished reading payload in %s", xFileInfo.FullPath)
			break
		}
		if err != nil {
			Dart.Log.Errorf("FileSystemBagReader.ScanPayload error reading %s: %v", xFileInfo.FullPath, err)
			return err
		}
	}
	r.mergePayloadManifestChecksums()
	return nil
}

// Close closes the FileSystemBagReader, which is a no-op.
func (r *FileSystemBagReader) Close() {
	// Unlike the TarFileBagReader, there is no underlying
	// reader to close here, so this is a no-op.
	// This method exists for compatibility with the BagReader
	// interface.
}

// processMetaEntry parses manifests and tag files.
func (r *FileSystemBagReader) processMetaEntry(xFileInfo *util.ExtendedFileInfo) error {
	pathInBag := r.getPathInBag(xFileInfo.FullPath)
	var err error
	fileType := util.BagFileType(pathInBag)
	switch fileType {
	case constants.FileTypeManifest:
		err = r.parseManifest(pathInBag, xFileInfo.FullPath, r.validator.PayloadFiles)
	case constants.FileTypeTagManifest:
		err = r.parseManifest(pathInBag, xFileInfo.FullPath, r.validator.TagFiles)
	case constants.FileTypeTag:
		r.parseTagFile(pathInBag, xFileInfo.FullPath)
	}
	fileMap := r.validator.MapForPath(pathInBag)
	r.addOrUpdateFileRecord(fileMap, pathInBag, xFileInfo.Size())
	return err
}

// parseManifest parses manifest entries in manifest and adds them
// to the right file map. Payload manifest entries are added to the
// map of payload files. Tag manifest entries are added to the map
// of tag files.
func (r *FileSystemBagReader) parseManifest(pathInBag, fullPathToFile string, fileMap *FileMap) error {
	alg, err := util.AlgorithmFromManifestName(fullPathToFile)
	if err != nil {
		Dart.Log.Errorf("FileSystemBagReader.parseManifest error getting algs for %s: %v", fullPathToFile, err)
		return err
	}
	fileToParse, err := os.Open(fullPathToFile)
	if err != nil {
		Dart.Log.Errorf("FileSystemBagReader.parseManifest error opening file %s: %v", fullPathToFile, err)
		return err
	}
	entries, err := ParseManifest(fileToParse)
	if err != nil {
		Dart.Log.Errorf("FileSystemBagReader.parseManifest error parsing entries for %s: %v", fullPathToFile, err)
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
func (r *FileSystemBagReader) parseTagFile(pathInBag, fullPathToFile string) {
	if !strings.HasSuffix(pathInBag, ".txt") {
		return
	}
	fileToParse, err := os.Open(fullPathToFile)
	if err != nil {
		Dart.Log.Errorf("FileSystemBagReader.parseTagFile error opening file %s: %v", fullPathToFile, err)
		return
	}
	tags, err := ParseTagFile(fileToParse, pathInBag)
	if err != nil {
		r.validator.UnparsableTagFiles = append(r.validator.UnparsableTagFiles, pathInBag)
	} else {
		r.validator.Tags = append(r.validator.Tags, tags...)
	}
}

// addOrUpdateFileRecord adds or updates a FileRecord in fileMap.
// We call this when scanning manifests and when scanning the payload
// because some files may exist in one but not the other. If that's
// the case, we want to know because mismatched file lists can mean
// the bag is invalid.
func (r *FileSystemBagReader) addOrUpdateFileRecord(fileMap *FileMap, pathInBag string, size int64) *FileRecord {
	// avoid multiple map lookups
	fileRecord := fileMap.Files[pathInBag]
	if fileRecord == nil {
		fileRecord = NewFileRecord()
		fileMap.Files[pathInBag] = fileRecord
	}
	// If we encounter this file first in a manifest entry,
	// we don't know its size, so we pass -1, which we don't
	// want to record because it's invalid. Record only valid
	// sizes.
	if size >= 0 {
		fileRecord.Size = size
	}
	return fileRecord
}

// processPayloadEntry adds files and checksums to our validator.
// See ensureFileRecord below.
func (r *FileSystemBagReader) processPayloadEntry(xFileInfo *util.ExtendedFileInfo) error {
	if !xFileInfo.IsDir() {
		return r.ensureFileRecord(xFileInfo)
	}
	return nil
}

// ensureFileRecord makes sure we have a FileRecord in the right
// FileMap. It also calculates and stores the required checksums
// for the file.
func (r *FileSystemBagReader) ensureFileRecord(xFileInfo *util.ExtendedFileInfo) error {
	pathInBag := r.getPathInBag(xFileInfo.FullPath)

	fileMap := r.validator.MapForPath(pathInBag)
	fileRecord := r.addOrUpdateFileRecord(fileMap, pathInBag, xFileInfo.Size())

	var err error
	fileType := util.BagFileType(pathInBag)
	var algs []string
	if fileType == constants.FileTypePayload {
		algs, err = r.validator.PayloadManifestAlgs()
	} else {
		algs, err = r.validator.TagManifestAlgs()
	}
	if err != nil {
		Dart.Log.Errorf("FileSystemBagReader.ensureFileRecord for %s: %v", xFileInfo.FullPath, err)
		return err
	}
	return r.addChecksums(pathInBag, xFileInfo.FullPath, fileRecord, algs)
}

func (r *FileSystemBagReader) getPathInBag(fullPath string) string {
	pathInBag := strings.Replace(fullPath, r.validator.PathToBag, "", 1)
	if runtime.GOOS == "windows" {
		pathInBag = strings.ReplaceAll(pathInBag, "\\", "/")
	}
	if strings.HasPrefix(pathInBag, "/") {
		pathInBag = strings.Replace(pathInBag, "/", "", 1)
	}
	return pathInBag
}

// addChecksums calculates checksums on a file stream and adds those
// checksums to the FileRecord. We calculate one checksum for each
// known manifest algorithm. For example, if the  bag has md5, sha1,
// sha256, and sha512 manifests, we'll calculate all those checksums.
// If it has only md5 and sha256, we'll calculate just those two.
//
// We use a MultiWriter to calculate all of a file's checksums in a
// single read.
func (r *FileSystemBagReader) addChecksums(pathInBag, fullPathToFile string, fileRecord *FileRecord, algs []string) error {

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

	sourceFile, err := os.Open(fullPathToFile)
	if err != nil {
		Dart.Log.Errorf("FileSystemBagReader error opening file for read %s: %v", fullPathToFile, err)
		return err
	}

	_, err = io.Copy(multiWriter, sourceFile)
	if err != nil {
		Dart.Log.Errorf("FileSystemBagReader error adding checksums for file %s: %v", fullPathToFile, err)
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

// Because payload manifests may have entries in tag manifest
// files, we need to make sure their file records and checksums
// appear in TagFiles map as well as the PayloadManifests map.
func (r *FileSystemBagReader) mergePayloadManifestChecksums() {
	for name, fileRecord := range r.validator.PayloadManifests.Files {
		tagFileRecord := r.validator.TagFiles.Files[name]
		if tagFileRecord != nil {
			tagFileRecord.Size = fileRecord.Size
			tagFileRecord.Checksums = append(tagFileRecord.Checksums, fileRecord.Checksums...)
		}
	}
}
