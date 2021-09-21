package core

import (
	"fmt"
	"strings"

	"github.com/APTrust/dart-runner/util"
)

type PackageOperation struct {
	BagItSerialization string            `json:"bagItSerialization"`
	Errors             map[string]string `json:"errors"`
	OutputPath         string            `json:"outputPath"`
	PackageFormat      string            `json:"packageFormat"`
	PackageName        string            `json:"packageName"`
	PayloadSize        int64             `json:"payloadSize"`
	PluginId           string            `json:"pluginId"`
	Result             *OperationResult  `json:"result"`
	SkipFiles          []string          `json:"skipFiles"`
	SourceFiles        []string          `json:"sourceFiles"`
	TrimLeadingPaths   bool              `json:"_trimLeadingPaths"`
}

func NewPackageOperation(packageName, outputPath string) *PackageOperation {
	return &PackageOperation{
		PackageName:      packageName,
		OutputPath:       outputPath,
		SourceFiles:      make([]string, 0),
		SkipFiles:        make([]string, 0),
		TrimLeadingPaths: true,
		Result:           NewOperationResult("package", "packager"),
		Errors:           make(map[string]string),
	}
}

func (p *PackageOperation) Validate() bool {
	p.Errors = make(map[string]string)
	if strings.TrimSpace(p.PackageName) == "" {
		p.Errors["PackageOperation.PackageName"] = "Package name is required."
	}
	if strings.TrimSpace(p.OutputPath) == "" {
		p.Errors["PackageOperation.OutputPath"] = "Output path is required."
	}
	if p.SourceFiles == nil || util.IsEmptyStringList(p.SourceFiles) {
		p.Errors["PackageOperation.sourceFiles"] = "Specify at least one file or directory to package."
	}
	missingFiles := make([]string, 0)
	for _, sourceFile := range p.SourceFiles {
		if !util.FileExists(sourceFile) {
			missingFiles = append(missingFiles, sourceFile)
		}
	}
	if len(missingFiles) > 0 {
		p.Errors["PackageOperation.sourceFiles"] = fmt.Sprintf("The following files are missing: %s", strings.Join(missingFiles, ""))
	}
	return len(p.Errors) == 0
}

// PruneSourceFiles removes any non-existent files from the list
// of source files to be packaged. This is useful on DART desktop
// because sometimes users create a job, then delete files, then run
// the job. This is questionable on a server.
func (p *PackageOperation) PruneSourceFiles() {
	existingFiles := make([]string, 0)
	for _, file := range p.SourceFiles {
		if util.FileExists(file) {
			existingFiles = append(existingFiles, file)
		}
	}
	p.SourceFiles = existingFiles
}

func (p *PackageOperation) GetWriter() {
	// To do: return the tar writer, or whatever kind of
	// writer we're using.
}
