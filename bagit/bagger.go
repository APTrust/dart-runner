package bagit

import (
	"embed"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/APTrust/dart-runner/constants"
	"github.com/APTrust/dart-runner/util"
	"github.com/google/uuid"
)

//go:embed bagit.txt
var bagitTxt embed.FS

type Bagger struct {
	Profile      *Profile
	OutputPath   string
	Files        []*util.ExtendedFileInfo
	Errors       map[string]string
	payloadFiles int64
	payloadBytes int64
	writer       util.BagWriter
	pathPrefix   string
}

func NewBagger(outputPath string, profile *Profile, files []*util.ExtendedFileInfo) *Bagger {
	return &Bagger{
		Profile:      profile,
		OutputPath:   outputPath,
		Files:        files,
		Errors:       make(map[string]string),
		payloadFiles: 0,
		payloadBytes: 0,
		pathPrefix:   "",
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

func (b *Bagger) PayloadFiles() int64 {
	return b.payloadFiles
}

func (b *Bagger) PayloadOxum() string {
	return fmt.Sprintf("%d.%d", b.payloadBytes, b.payloadFiles)
}

func (b *Bagger) reset() {
	b.Errors = make(map[string]string)
	b.payloadFiles = 0
	b.payloadBytes = 0
}

func (b *Bagger) addBagItFile() bool {
	fInfo, err := os.Stat("bagit.txt")
	if err != nil {
		b.Errors["bagit.txt"] = err.Error()
		return false
	}
	xFileInfo := util.NewExtendedFileInfo("bagit.txt", fInfo)
	err = b.writer.AddFile(xFileInfo, "bagit.txt")
	if err != nil {
		b.Errors["bagit.txt"] = err.Error()
		return false
	}
	return true
}

func (b *Bagger) addPayloadFiles() bool {
	var err error
	for _, xFileInfo := range b.Files {
		// Always use forward slash for bag paths, even on Windows.
		pathInBag := "data" + strings.Replace(xFileInfo.FullPath, b.pathPrefix, "", 1)
		err = b.writer.AddFile(xFileInfo, pathInBag)
		if err != nil {
			b.Errors[xFileInfo.FullPath] = err.Error()
		}
		b.payloadFiles++
		b.payloadBytes += xFileInfo.Size()
	}
	return true
}

func (b *Bagger) addManifests(whichKind string) bool {
	return true
}

func (b *Bagger) addTagFiles() bool {
	// We've already added BagIt. Read the profile and build from
	// there. The profile should have tag defs merged from the job
	// or workflow.
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
	b.writer = util.NewTarWriter(b.OutputPath)
	b.writer.Open()
	return true
}

func (b *Bagger) calculatePathPrefix() {
	paths := make([]string, len(b.Files))
	for i, xFileInfo := range b.Files {
		paths[i] = xFileInfo.FullPath
	}
	b.pathPrefix = util.FindCommonPrefix(paths)
}

func (b *Bagger) setBagInfoAutoValues() {
	baggingDate, err := b.Profile.FirstMatchingTag("tagName", "Bagging-Date")
	if err == nil && baggingDate != nil {
		baggingDate.UserValue = time.Now().UTC().Format(time.RFC3339)
	}

	baggingSoftware, err := b.Profile.FirstMatchingTag("tagName", "Bagging-Software")
	if err == nil && baggingSoftware != nil {
		baggingSoftware.UserValue = constants.AppVersion()
	}

	// This supposedly required tag is often omitted from profiles in the wild.
	profileIdentifier, err := b.Profile.FirstMatchingTag("tagName", "BagIt-Profile-Identifier")
	if err == nil && profileIdentifier == nil {
		profileIdentifier = &TagDefinition{
			ID:       uuid.New().String(),
			TagFile:  "bag-info.txt",
			TagName:  "BagIt-Profile-Identifier",
			Required: true,
		}
		b.Profile.Tags = append(b.Profile.Tags, profileIdentifier)
	}
	if b.Profile.BagItProfileInfo.BagItProfileIdentifier != "" {
		profileIdentifier.UserValue = b.Profile.BagItProfileInfo.BagItProfileIdentifier
	} else {
		profileIdentifier.UserValue = "http://example.com/unspecified_profile_identifier"
	}

	payloadOxum, err := b.Profile.FirstMatchingTag("tagName", "Payload-Oxum")
	if err == nil && payloadOxum != nil {
		payloadOxum.UserValue = b.PayloadOxum()
	}

	bagSize, err := b.Profile.FirstMatchingTag("tagName", "Bag-Size")
	if err == nil && bagSize != nil {
		bagSize.UserValue = util.ToHumanSize(b.payloadBytes, 1024)
	}
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
