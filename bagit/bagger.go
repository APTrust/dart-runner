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
	writer           BagWriter
	pathPrefix       string
	bagName          string
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
	}
}

// Run builds the bag and returns the number of files bagged.
func (b *Bagger) Run() bool {
	b.reset()
	if !b.validateProfile() {
		return false
	}

	b.calculatePathPrefix()
	b.calculateBagName()

	if !b.initWriter() {
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
	return b.PayloadFiles.TotalBytes()
}

func (b *Bagger) PayloadFileCount() int64 {
	return b.PayloadFiles.FileCount()
}

func (b *Bagger) PayloadOxum() string {
	return fmt.Sprintf("%d.%d", b.PayloadFiles.TotalBytes(), b.PayloadFiles.FileCount())
}

func (b *Bagger) reset() {
	b.Errors = make(map[string]string)
}

func (b *Bagger) addPayloadFiles() bool {
	for _, xFileInfo := range b.FilesToBag {
		pathInBag := b.pathForPayloadFile(xFileInfo.FullPath)
		checksums, err := b.writer.AddFile(xFileInfo, pathInBag)
		if err != nil {
			b.Errors[xFileInfo.FullPath] = err.Error()
		}

		// Track the checksums, except for directory entries,
		// which won't have checksums because no actual data
		// is written.
		if !xFileInfo.IsDir() {
			fileRecord := NewFileRecord()
			fileRecord.Size = xFileInfo.Size()
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
	pathInBag := b.pathForTagFile(filename)
	tempFilePath := ""
	outputFile, err := os.CreateTemp("", fmt.Sprintf("%s-%d", filename, time.Now().UnixNano()))
	if outputFile != nil {
		tempFilePath = outputFile.Name()
		defer outputFile.Close()
	}
	if err != nil {
		b.Errors[filename] = fmt.Sprintf("Error opening temp file: %s", err.Error())
		return tempFilePath, pathInBag, false
	}
	trimFromPath := fmt.Sprintf("%s/", b.bagName)
	err = fileMap.WriteManifest(outputFile, subjectFileType, alg, trimFromPath)
	if err != nil {
		b.Errors[pathInBag] = fmt.Sprintf("Error writing manifest %s (type=%s, subjectType=%s)): %s", filename, whichKind, subjectFileType, err.Error())
		return tempFilePath, pathInBag, false
	}
	return tempFilePath, pathInBag, true
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
		pathInBag := b.pathForTagFile(tagFileName)
		checksums, err := b.writer.AddFile(xFileInfo, pathInBag)
		if err != nil {
			b.Errors[tagFileName] = fmt.Sprintf("Error writing tag file to bag: %s", err.Error())
			return false
		}

		// Track the checksums
		fileRecord := NewFileRecord()
		for alg, digest := range checksums {
			fileRecord.AddChecksum(constants.FileTypeTag, alg, digest)
		}
		b.TagFiles.Files[pathInBag] = fileRecord
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
	b.Profile.SetTagValue("bag-info.txt", "Bagging-Software", constants.AppVersion)
	b.Profile.SetTagValue("bag-info.txt", "Payload-Oxum", b.PayloadOxum())
	b.Profile.SetTagValue("bag-info.txt", "Bag-Size", util.ToHumanSize(b.PayloadBytes(), 1024))
	bpIdentifier := "http://example.com/unspecified_profile_identifier"
	if b.Profile.BagItProfileInfo.BagItProfileIdentifier != "" {
		bpIdentifier = b.Profile.BagItProfileInfo.BagItProfileIdentifier
	}
	b.Profile.SetTagValue("bag-info.txt", "BagIt-Profile-Identifier", bpIdentifier)
}

func (b *Bagger) calculateBagName() {
	b.bagName = path.Base(b.OutputPath)
	b.bagName = strings.TrimSuffix(b.bagName, path.Ext(b.bagName))
	// Handle common .tar.gz case
	b.bagName = strings.TrimSuffix(b.bagName, ".tar")
}

func (b *Bagger) pathForPayloadFile(fullPath string) string {
	shortPath := strings.Replace(fullPath, b.pathPrefix, "", 1)
	if !strings.HasPrefix(shortPath, "/") {
		shortPath = "/" + shortPath
	}
	return fmt.Sprintf("%s/data%s", b.bagName, shortPath)
}

func (b *Bagger) pathForTagFile(fullPath string) string {
	shortPath := strings.Replace(fullPath, b.pathPrefix, "", 1)
	return fmt.Sprintf("%s/%s", b.bagName, shortPath)
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
