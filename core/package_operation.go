package core

import (
	"fmt"
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
	if strings.TrimSpace(p.PackageName) == "" {
		p.Errors["PackageOperation.PackageName"] = "Package name is required."
	}
	if strings.TrimSpace(p.OutputPath) == "" {
		p.Errors["PackageOperation.OutputPath"] = "Output path is required."
	}
	if p.SourceFiles == nil || util.IsEmptyStringList(p.SourceFiles) {
		p.Errors["PackageOperation.SourceFiles"] = "Specify at least one file or directory to package."
	}
	missingFiles := make([]string, 0)
	duplicateFiles := make([]string, 0)
	alreadySeen := make(map[string]bool)
	for _, sourceFile := range p.SourceFiles {
		if !util.FileExists(sourceFile) {
			missingFiles = append(missingFiles, sourceFile)
		}
		if alreadySeen[sourceFile] {
			duplicateFiles = append(duplicateFiles, sourceFile)
		}
		alreadySeen[sourceFile] = true
	}
	if len(missingFiles) > 0 {
		p.Errors["PackageOperation.sourceFiles"] = fmt.Sprintf("The following files are missing: %s", strings.Join(missingFiles, ""))
	}
	if len(duplicateFiles) > 0 {
		p.Errors["PackageOperation.sourceFiles"] += fmt.Sprintf("The following files are included more than once. Please remove duplicates: %s", strings.Join(duplicateFiles, ""))
	}
	return len(p.Errors) == 0
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
