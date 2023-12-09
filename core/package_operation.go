package core

import (
	"path/filepath"
	"strings"

	"github.com/APTrust/dart-runner/constants"
	"github.com/APTrust/dart-runner/util"
)

type PackageOperation struct {
	BagItSerialization string            `json:"bagItSerialization"`
	Errors             map[string]string `json:"errors"`
	OutputPath         string            `json:"outputPath"`
	PackageName        string            `json:"packageName"`
	PackageFormat      string            `json:"packageFormat"`
	PayloadSize        int64             `json:"payloadSize"`
	Result             *OperationResult  `json:"result"`
	SourceFiles        []string          `json:"sourceFiles"`
}

func NewPackageOperation(packageName, outputPath string, sourceFiles []string) *PackageOperation {
	if filepath.Base(outputPath) != packageName {
		outputPath = filepath.Join(outputPath, packageName)
	}
	return &PackageOperation{
		PackageName: packageName,
		OutputPath:  outputPath,
		SourceFiles: sourceFiles,
		Result:      NewOperationResult("package", "Bagger - "+constants.AppVersion),
		Errors:      make(map[string]string),
	}
}

func (p *PackageOperation) Validate() bool {
	p.Errors = make(map[string]string)
	p.PruneSourceFiles()
	if strings.TrimSpace(p.PackageName) == "" {
		p.Errors["PackageOperation.PackageName"] = "Package name is required."
	}
	if strings.TrimSpace(p.OutputPath) == "" {
		p.Errors["PackageOperation.OutputPath"] = "Output path is required."
	}
	if p.SourceFiles == nil || util.IsEmptyStringList(p.SourceFiles) {
		p.Errors["PackageOperation.SourceFiles"] = "Specify at least one file or directory to package."
	}
	for key, value := range p.Errors {
		Dart.Log.Errorf("%s: %s", key, value)
	}
	return len(p.Errors) == 0
}

// PruneSourceFiles removes non-existent files and directories
// from SouceFiles and de-dupes duplicates.
func (p *PackageOperation) PruneSourceFiles() {
	alreadySeen := make(map[string]bool)
	cleanSourceFileList := make([]string, 0)
	for _, sourceFile := range p.SourceFiles {
		if util.FileExists(sourceFile) && !alreadySeen[sourceFile] {
			cleanSourceFileList = append(cleanSourceFileList, sourceFile)
		}
		alreadySeen[sourceFile] = true
	}
	p.SourceFiles = cleanSourceFileList
}

func (p *PackageOperation) ToForm() *Form {
	form := NewForm("PackageOperation", "", p.Errors)

	serialization := form.AddField("BagItSerialization", "Serialization", p.BagItSerialization, false)
	serialization.Choices = MakeChoiceList(constants.SerializationOptions, p.BagItSerialization)

	form.AddField("OutputPath", "Output Path", p.OutputPath, true)
	form.AddField("PackageName", "Package Name", p.PackageName, true)

	packageFormat := form.AddField("PackageFormat", "Package Format", p.PackageFormat, true)
	packageFormat.Choices = MakeChoiceList(constants.PackageFormats, p.PackageFormat)

	form.AddMultiValueField("SourceFiles", "Files to Package", p.SourceFiles, false)

	return form
}
