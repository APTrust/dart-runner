package core

import (
	"path"
	"strings"

	"github.com/APTrust/dart-runner/util"
)

type CSVBatchParser struct {
	PathToCSVFile string
	Workflow      *Workflow
}

// NewCSVBatchParser creates a new parser to convert records
// in a CSV batch file to a slice of JobParams. Param pathToCSVFile
// is the path the CSV file you want to parse. Param workfow is the
// workflow that you want to apply to all items in that CSV file.
func NewCSVBatchParser(pathToCSVFile string, workflow *Workflow) *CSVBatchParser {
	return &CSVBatchParser{
		PathToCSVFile: pathToCSVFile,
		Workflow:      workflow,
	}
}

// ParseAll converts the records in a CSV batch file to a slice of
// JobParams objects. Param outputDir will be passed through to each
// JobParams object, describing the directory in which to create bags.
// Unless you have some special reason, outputDir should be set to
// the value of the built-in app setting called "Bagging Directory".
// You can get that with a call to core.GetAppSetting("Bagging Directory").
func (p *CSVBatchParser) ParseAll(outputDir string) ([]*JobParams, error) {
	jobParamsList := make([]*JobParams, 0)
	_, records, err := util.ParseCSV(p.PathToCSVFile)
	if err != nil {
		return nil, err
	}
	for _, nvpList := range records {
		packageNamePair, _ := nvpList.FirstMatching("Bag-Name")
		packageName := packageNamePair.Value

		filesToBagPair, _ := nvpList.FirstMatching("Root-Directory")
		filesToBag := []string{filesToBagPair.Value}

		outputFile := path.Join(outputDir, packageName)
		tags := p.parseTags(nvpList)
		jobParams := NewJobParams(p.Workflow, packageName, outputFile, filesToBag, tags)
		jobParamsList = append(jobParamsList, jobParams)
	}
	return jobParamsList, nil
}

// parseTags converts a single csv record to BagIt tag objects.
// The record is a map that typically looks like this:
//
//	{
//	   "BagIt-Version": "1.0",
//	   "custom-tag-file.txt/User-Name": "Spongebob",
//	   "custom-tag-file.txt/Copyright-Expires": "2067",
//	}
func (p *CSVBatchParser) parseTags(record *util.NameValuePairList) []*Tag {
	tags := make([]*Tag, 0)
	for _, nvp := range record.Items {
		var tagName string
		var tagFile string
		// Field name is in format file-name.txt/Tag-Name.
		// We need to split this into file name and tag name.
		parts := strings.SplitN(nvp.Name, "/", 2)
		if len(parts) == 0 {
			continue // bad entry?
		}
		if len(parts) == 1 {
			// No tag file specified. Assume bag-info.txt.
			tagName = parts[0]
			tagFile = "bag-info.txt"
		} else {
			tagFile = parts[0]
			tagName = parts[1]
		}
		tags = append(tags, NewTag(tagFile, tagName, nvp.Value))
	}
	return tags
}
