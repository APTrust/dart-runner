package core

import (
	"path"

	"github.com/APTrust/dart-runner/util"
)

type CSVBatchParser struct {
	PathToCSVFile string
	Workflow      *Workflow
}

func NewCSVBatchParser(pathToCSVFile string, workflow *Workflow) *CSVBatchParser {
	return &CSVBatchParser{
		PathToCSVFile: pathToCSVFile,
		Workflow:      workflow,
	}
}

func (p *CSVBatchParser) ParseAll() ([]*JobParams, error) {

	// TODO: Any way to avoid this coupling?
	outputDir, err := GetAppSetting("Bagging Directory")

	if err != nil {
		return nil, err
	}
	jobParamsList := make([]*JobParams, 0)
	_, records, err := util.ParseCSV(p.PathToCSVFile)
	if err != nil {
		return nil, err
	}
	for _, record := range records {
		packageName := record["Bag-Name"]
		outputFile := path.Join(outputDir, packageName)
		filesToBag := []string{record["Root-Directory"]}
		tags := p.parseTags()

		// TODO: Figure out if bag must be tarred, zipped, etc.
		// if p.Workflow.PackageFormat

		jobParams := NewJobParams(p.Workflow, packageName, outputFile, filesToBag, tags)
		jobParamsList = append(jobParamsList, jobParams)
	}
	return jobParamsList, nil
}

func (p *CSVBatchParser) parseTags() []*Tag {
	return nil
}
