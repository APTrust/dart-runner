package bagit

import (
	"embed"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"strings"
	"time"

	"github.com/APTrust/dart-runner/constants"
	"github.com/APTrust/dart-runner/util"
)

//go:embed bagit.txt
var bagitTxt embed.FS

type Bagger struct {
	Profile          *Profile
	OutputPath       string
	FilesToBag       []*util.ExtendedFileInfo
	Errors           map[string]string
	PayloadFiles     *FileMap
	PayloadManifests *FileMap
	TagFiles         *FileMap
	TagManifests     *FileMap
	payloadFileCount int64
	payloadBytes     int64
	writer           BagWriter
	pathPrefix       string
}

func NewBagger(outputPath string, profile *Profile, filesToBag []*util.ExtendedFileInfo) *Bagger {
	return &Bagger{
		Profile:          profile,
		OutputPath:       outputPath,
		FilesToBag:       filesToBag,
		PayloadFiles:     NewFileMap(constants.FileTypePayload),
		PayloadManifests: NewFileMap(constants.FileTypeManifest),
		TagFiles:         NewFileMap(constants.FileTypeTag),
		TagManifests:     NewFileMap(constants.FileTypeTagManifest),
		Errors:           make(map[string]string),
		payloadFileCount: 0,
		payloadBytes:     0,
		pathPrefix:       "",
	}
}

// Run builds the bag and returns the number of files bagged.
func (b *Bagger) Run() bool {
	b.reset()
	if !b.validateProfile() {
		return false
	}

	b.calculatePathPrefix()

	if !b.initWriter() {
		return false
	}

	if !b.addBagItFile() {
		// Write bagit.txt into the bag.
		return false
	}

	if !b.addPayloadFiles() {
		return false
	}

	// Here we should have enough info to print
	// the Payload-Oxum in bag-info.txt.
	if !b.addTagFiles() {
		return false
	}

	// Payload manifests
	if !b.addManifests(constants.FileTypeManifest) {
		return false
	}

	// Tag manifests must be added last because they
	// need to run checksums on tag files and payload manifests.
	if !b.addManifests(constants.FileTypeTagManifest) {
		return false
	}

	b.finish()

	return len(b.Errors) == 0
}

func (b *Bagger) PayloadBytes() int64 {
	return b.payloadBytes
}

func (b *Bagger) PayloadFileCount() int64 {
	return b.payloadFileCount
}

func (b *Bagger) PayloadOxum() string {
	return fmt.Sprintf("%d.%d", b.payloadBytes, b.payloadFileCount)
}

func (b *Bagger) reset() {
	b.Errors = make(map[string]string)
	b.payloadFileCount = 0
	b.payloadBytes = 0
}

func (b *Bagger) addBagItFile() bool {
	fInfo, err := os.Stat("bagit.txt")
	if err != nil {
		b.Errors["bagit.txt"] = err.Error()
		return false
	}
	xFileInfo := util.NewExtendedFileInfo("bagit.txt", fInfo)
	checksums, err := b.writer.AddFile(xFileInfo, "bagit.txt")
	if err != nil {
		b.Errors["bagit.txt"] = err.Error()
		return false
	}

	// Track the checksum
	fileRecord := NewFileRecord()
	for alg, digest := range checksums {
		fileRecord.AddChecksum(constants.FileTypePayload, alg, digest)
	}
	b.TagFiles.Files["bagit.txt"] = fileRecord
	return true
}

func (b *Bagger) addPayloadFiles() bool {
	for _, xFileInfo := range b.FilesToBag {
		// Always use forward slash for bag paths, even on Windows.
		pathInBag := "data" + strings.Replace(xFileInfo.FullPath, b.pathPrefix, "", 1)
		checksums, err := b.writer.AddFile(xFileInfo, pathInBag)
		if err != nil {
			b.Errors[xFileInfo.FullPath] = err.Error()
		}
		b.payloadFileCount++
		b.payloadBytes += xFileInfo.Size()

		// Track the checksums, except for directory entries,
		// which won't have checksums because no actual data
		// is written.
		if !xFileInfo.IsDir() {
			fileRecord := NewFileRecord()
			for alg, digest := range checksums {
				fileRecord.AddChecksum(constants.FileTypePayload, alg, digest)
			}
			b.PayloadFiles.Files[pathInBag] = fileRecord
		}
	}
	return true
}

func (b *Bagger) addManifests(whichKind string) bool {
	for _, alg := range b.writer.DigestAlgs() {
		tempFilePath, pathInBag, ok := b.writeManifest(whichKind, alg)
		defer os.Remove(tempFilePath)
		if !ok {
			return false
		}
		fileInfo, err := os.Stat(tempFilePath)
		if err != nil {
			b.Errors[pathInBag] = err.Error()
			return false
		}
		xFileInfo := util.NewExtendedFileInfo(tempFilePath, fileInfo)
		checksums, err := b.writer.AddFile(xFileInfo, pathInBag)

		// Tag manifests should contain digests of payload manifests.
		// In this context, a payload manifest is a type of tag file.
		// We want to mark it as such because when we write tagmanifests,
		// we're going to ask the FileMap for all tag file checksums.
		if whichKind == constants.FileTypeManifest {
			fileRecord := NewFileRecord()
			for alg, digest := range checksums {
				fileRecord.AddChecksum(constants.FileTypeTag, alg, digest)
			}
			b.TagFiles.Files[pathInBag] = fileRecord
		}
	}
	return true
}

func (b *Bagger) writeManifest(whichKind, alg string) (string, string, bool) {
	fileMap := b.PayloadFiles
	prefix := "manifest"
	subjectFileType := constants.FileTypePayload
	if whichKind == constants.FileTypeTagManifest {
		fileMap = b.TagFiles
		prefix = "tagmanifest"
		subjectFileType = constants.FileTypeTag
	}
	filename := fmt.Sprintf("%s-%s.txt", prefix, alg)
	tempFilePath := ""
	outputFile, err := os.CreateTemp("", fmt.Sprintf("%s-%d", filename, time.Now().UnixNano()))
	if outputFile != nil {
		tempFilePath = outputFile.Name()
		defer outputFile.Close()
	}
	if err != nil {
		b.Errors[filename] = fmt.Sprintf("Error opening temp file: %s", err.Error())
		return tempFilePath, filename, false
	}
	err = fileMap.WriteManifest(outputFile, subjectFileType, alg)
	if err != nil {
		b.Errors[filename] = fmt.Sprintf("Error writing manifest %s (type=%s, subhectType=%s)): %s", filename, whichKind, subjectFileType, err.Error())
		return tempFilePath, filename, false
	}
	return tempFilePath, filename, true
}

func (b *Bagger) addTagFiles() bool {
	b.setBagInfoAutoValues()
	for _, tagFileName := range b.Profile.TagFileNames() {
		contents, err := b.Profile.GetTagFileContents(tagFileName)
		if err != nil {
			b.Errors[tagFileName] = fmt.Sprintf("Error getting tag file contents: %s", err.Error())
			return false
		}
		tempFilePath := path.Join(os.TempDir(), fmt.Sprintf("%s-%d", tagFileName, time.Now().UnixNano()))
		defer os.Remove(tempFilePath)
		err = ioutil.WriteFile(tempFilePath, []byte(contents), 0644)
		if err != nil {
			b.Errors[tagFileName] = fmt.Sprintf("Error writing tag file contents to temp file: %s", err.Error())
			return false
		}
		fileInfo, err := os.Stat(tempFilePath)
		if err != nil {
			b.Errors[tagFileName] = fmt.Sprintf("Error getting temp file stat: %s", err.Error())
			return false
		}
		xFileInfo := util.NewExtendedFileInfo(tempFilePath, fileInfo)
		checksums, err := b.writer.AddFile(xFileInfo, tagFileName)
		if err != nil {
			b.Errors[tagFileName] = fmt.Sprintf("Error writing tag file to bag: %s", err.Error())
			return false
		}

		// Track the checksums
		fileRecord := NewFileRecord()
		for alg, digest := range checksums {
			fileRecord.AddChecksum(constants.FileTypeTag, alg, digest)
		}
		b.TagFiles.Files[tagFileName] = fileRecord
	}
	return true
}

func (b *Bagger) validateProfile() bool {
	if b.Profile == nil {
		b.Errors["Profile"] = "BagIt profile cannot be nil"
	}
	if !b.Profile.IsValid() {
		b.Errors = b.Profile.Errors
	}
	return len(b.Errors) == 0
}

// In future, this will initialize the proper type of writer
// (zip, gzip, file system, etc.) For now, it supports tar only.
func (b *Bagger) initWriter() bool {
	digestAlgs := b.Profile.ManifestsRequired
	for _, alg := range b.Profile.TagManifestsRequired {
		if !util.StringListContains(digestAlgs, alg) {
			digestAlgs = append(digestAlgs, alg)
		}
	}
	// If no digest algs are required, pick one that's allowed.
	if len(digestAlgs) == 0 {
		digestAlgs = []string{
			b.getPreferredDigestAlg(),
		}
	}
	b.writer = NewTarWriter(b.OutputPath, digestAlgs)
	b.writer.Open()
	return true
}

func (b *Bagger) getPreferredDigestAlg() string {
	// What if TagManifestsAllowed differs from ManifestsAllowed?
	for _, alg := range constants.PreferredAlgsInOrder {
		if util.StringListContains(b.Profile.ManifestsAllowed, alg) {
			return alg
		}
	}
	// Nothing?? Try the tag manifest algs.
	for _, alg := range constants.PreferredAlgsInOrder {
		if util.StringListContains(b.Profile.TagManifestsAllowed, alg) {
			return alg
		}
	}
	// Still nothing? LOC recommends sha512, so that's what you get.
	return constants.AlgSha512
}

func (b *Bagger) calculatePathPrefix() {
	paths := make([]string, len(b.FilesToBag))
	for i, xFileInfo := range b.FilesToBag {
		paths[i] = xFileInfo.FullPath
	}
	b.pathPrefix = util.FindCommonPrefix(paths)
}

func (b *Bagger) setBagInfoAutoValues() {
	b.Profile.SetTagValue("bag-info.txt", "Bagging-Date", time.Now().UTC().Format(time.RFC3339))
	b.Profile.SetTagValue("bag-info.txt", "Bagging-Software", constants.AppVersion())
	b.Profile.SetTagValue("bag-info.txt", "Payload-Oxum", b.PayloadOxum())
	b.Profile.SetTagValue("bag-info.txt", "Bag-Size", util.ToHumanSize(b.payloadBytes, 1024))
	bpIdentifier := "http://example.com/unspecified_profile_identifier"
	if b.Profile.BagItProfileInfo.BagItProfileIdentifier != "" {
		bpIdentifier = b.Profile.BagItProfileInfo.BagItProfileIdentifier
	}
	b.Profile.SetTagValue("bag-info.txt", "BagIt-Profile-Identifier", bpIdentifier)
}

// Close the writer and do any other required cleanup.
func (b *Bagger) finish() bool {
	if b.writer != nil {
		err := b.writer.Close()
		if err != nil {
			b.Errors["BagWriter"] = fmt.Sprintf("Error closing bag writer: %s", err.Error())
		}
	}
	return true
}
