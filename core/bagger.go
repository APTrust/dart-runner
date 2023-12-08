package core

import (
	"embed"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/APTrust/dart-runner/constants"
	"github.com/APTrust/dart-runner/util"
)

//go:embed bagit.txt
var bagitTxt embed.FS

type Bagger struct {
	Profile           *BagItProfile
	OutputPath        string
	FilesToBag        []*util.ExtendedFileInfo
	Errors            map[string]string
	MessageChannel    chan *EventMessage
	PayloadFiles      *FileMap
	PayloadManifests  *FileMap
	TagFiles          *FileMap
	TagManifests      *FileMap
	ManifestArtifacts map[string]string
	TagFileArtifacts  map[string]string
	writer            BagWriter
	pathPrefix        string
	bagName           string
	currentFileNum    int64
	totalFileCount    int64
}

func NewBagger(outputPath string, profile *BagItProfile, filesToBag []*util.ExtendedFileInfo) *Bagger {
	return &Bagger{
		Profile:           profile,
		OutputPath:        outputPath,
		FilesToBag:        filesToBag,
		PayloadFiles:      NewFileMap(constants.FileTypePayload),
		PayloadManifests:  NewFileMap(constants.FileTypeManifest),
		TagFiles:          NewFileMap(constants.FileTypeTag),
		TagManifests:      NewFileMap(constants.FileTypeTagManifest),
		ManifestArtifacts: make(map[string]string),
		TagFileArtifacts:  make(map[string]string),
		Errors:            make(map[string]string),
	}
}

// Run builds the bag and returns the number of files bagged.
func (b *Bagger) Run() bool {
	b.reset()
	b.calculatePathPrefix()
	b.calculateBagName()
	Dart.Log.Infof("Starting to build bag %s", b.bagName)

	if !b.validateProfile() {
		return b.finish()
	}

	if !b.initWriter() {
		return b.finish()
	}

	if !b.addPayloadFiles() {
		return b.finish()
	}

	// Here we should have enough info to print
	// the Payload-Oxum in bag-info.txt.
	if !b.addTagFiles() {
		return b.finish()
	}

	// Payload manifests
	if !b.addManifests(constants.FileTypeManifest) {
		return b.finish()
	}

	// Tag manifests must be added last because they
	// need to run checksums on tag files and payload manifests.
	if !b.addManifests(constants.FileTypeTagManifest) {
		return b.finish()
	}

	return b.finish()
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

// GetTotalFilesBagged returns the total number of files bagged,
// including payload files, tag files, and manifests. This will
// be zero if the bagger has not run yet, so call this after
// calling Run() to get an accurate number.
func (b *Bagger) GetTotalFilesBagged() int64 {
	return b.currentFileNum
}

// ArtifactsDir returns the name of the directory in which the bagger
// will leave artifacts, including manifests and tag files. These
// items will remain after bagging is complete.
//
// Note that the bagger does not create or populate this directory.
// That happens in JobRunner.saveArtifactsToFileSystem, and it happens
// only when DART is running in command-line mode. Otherwise, we
// save artifacts to the DB.
func (b *Bagger) ArtifactsDir() string {
	// If bag is directory, artifacts will go into the directory,
	// and we don't want that. Put artifacts in their own directory,
	// outside the bag.
	var artifactsDir string
	if util.IsDirectory(b.OutputPath) {
		artifactsDir = b.OutputPath + "_artifacts"
	} else {
		artifactsDir = path.Dir(b.OutputPath)
		artifactsDir = filepath.Join(artifactsDir, fmt.Sprintf("%s_artifacts", b.bagName))
	}
	return artifactsDir
}

func (b *Bagger) reset() {
	b.totalFileCount = int64(len(b.FilesToBag) + len(b.Profile.TagFileNames()) + len(b.Profile.ManifestsRequired))
	b.currentFileNum = 0
	b.Errors = make(map[string]string)
}

func (b *Bagger) addPayloadFiles() bool {
	for _, xFileInfo := range b.FilesToBag {
		b.info(fmt.Sprintf("Adding %s", xFileInfo.FullPath))
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
		b.info(fmt.Sprintf("Adding %s", pathInBag))
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

	// New for DART3: keep a copy of manifest contents to save as
	// an Artifact when job completes.
	manifestContents, _ := os.ReadFile(tempFilePath)
	b.ManifestArtifacts[filename] = string(manifestContents)

	return tempFilePath, pathInBag, true
}

func (b *Bagger) addTagFiles() bool {
	b.setBagInfoAutoValues()
	for _, tagFileName := range b.Profile.TagFileNames() {
		b.info(fmt.Sprintf("Adding %s", tagFileName))
		contents, err := b.Profile.GetTagFileContents(tagFileName)
		if err != nil {
			b.Errors[tagFileName] = fmt.Sprintf("Error getting tag file contents: %s", err.Error())
			return false
		}

		// New for DART3: keep a copy of tag file contents to save as
		// an Artifact when job completes.
		b.TagFileArtifacts[tagFileName] = contents

		tempFilePath := filepath.Join(os.TempDir(), fmt.Sprintf("%s-%d", tagFileName, time.Now().UnixNano()))
		defer os.Remove(tempFilePath)
		err = os.WriteFile(tempFilePath, []byte(contents), 0644)
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
	if !b.Profile.Validate() {
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
// Returns true if bagging finished without errors,
// false otherwise.
func (b *Bagger) finish() bool {
	if b.writer != nil {
		err := b.writer.Close()
		if err != nil {
			b.Errors["BagWriter"] = fmt.Sprintf("Error closing bag writer: %s", err.Error())
		}
	}
	Dart.Log.Infof("Finished bag %s", b.bagName)
	if len(b.Errors) > 0 {
		Dart.Log.Errorf("Bagging %s failed with the following errors:", b.bagName)
	}
	for key, value := range b.Errors {
		Dart.Log.Errorf("%s: %s", key, value)
	}
	return len(b.Errors) == 0
}

func (b *Bagger) info(message string) {
	Dart.Log.Info(message)
	if b.MessageChannel == nil {
		b.currentFileNum += 1
		return
	}
	eventMessage := InfoEvent(constants.StagePackage, message)
	eventMessage.Current = b.currentFileNum
	eventMessage.Total = b.totalFileCount
	if b.currentFileNum > 0 {
		eventMessage.Percent = int(float64(b.currentFileNum) * 100 / float64(b.totalFileCount))
	}

	// How can percent surpass 100, and why do we fudge it?
	//
	// We do this because, while we know ahead of time the number of payload
	// files we have to bag, we don't know the number of tag files and manifests.
	//
	// When calculating totalFileCount, we guess the number of manifests using
	// the BagIt profile's required manifests, and we guess the number of tag
	// files using the Profile's RequiredTagFiles. However, the user can add
	// additional tag files and manifests on a per-job basis. If they do, the
	// actual total file cound will be slightly higher than we anticipated.
	//
	// Tag and manifest files are usually small and all of them together are
	// usually added to the bag in a fraction of a second. So the user may see
	// the progress bar stick at 100% for a fraction of a second.
	if eventMessage.Percent > 100 {
		eventMessage.Percent = 100
	}

	b.MessageChannel <- eventMessage
	b.currentFileNum += 1
}
