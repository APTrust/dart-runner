package core

import (
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/APTrust/dart-runner/bagit"
)

type WorkflowCSVFile struct {
	PathToFile string
	reader     *csv.Reader
	headers    []string
	headerTags []*bagit.Tag
	file       *os.File
}

func NewWorkflowCSVFile(pathToFile string) (*WorkflowCSVFile, error) {
	f, err := os.Open(pathToFile)
	if err != nil {
		return nil, err
	}
	csv := &WorkflowCSVFile{
		PathToFile: pathToFile,
		reader:     csv.NewReader(f),
		file:       f,
	}
	err = csv.parseHeaders()
	if err != nil {
		f.Close()
		return nil, err
	}
	return csv, nil
}

func (csv *WorkflowCSVFile) parseHeaders() error {
	headers, err := csv.reader.Read()
	if err != nil {
		return err
	}
	headerTags := make([]*bagit.Tag, len(headers))
	hasBagName := false
	hasRootDir := false
	for i, h := range headers {
		if h == "Bag-Name" {
			hasBagName = true
			headerTags[i] = bagit.NewTag("", h, "")
			continue
		}
		if h == "Root-Directory" {
			hasRootDir = true
			headerTags[i] = bagit.NewTag("", h, "")
			continue
		}
		parts := strings.Split(h, "/")
		if len(parts) != 2 {
			return fmt.Errorf("Bag tag header '%s' in column %d. Header name should use tagFile/tagName pattern.", h, i)
		}
		headerTags[i] = bagit.NewTag(parts[0], parts[1], "")
	}
	if !hasBagName {
		return fmt.Errorf("CSV file is missing Bag-Name column")
	}
	if !hasRootDir {
		return fmt.Errorf("CSV file is missing Root-Directory column")
	}
	csv.headers = headers
	csv.headerTags = headerTags
	return nil
}

func (csv *WorkflowCSVFile) Headers() []string {
	return csv.headers
}

func (csv *WorkflowCSVFile) HeaderTags() []*bagit.Tag {
	return csv.headerTags
}

func (csv *WorkflowCSVFile) ReadNext() (*WorkflowCSVEntry, error) {
	record, err := csv.reader.Read()
	if err == io.EOF {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	entry := NewWorkflowCSVEntry("", "")
	for i, value := range record {
		if value == "Bag-Name" {
			entry.BagName = value
		} else if value == "Root-Directory" {
			entry.RootDir = value
		} else {
			tag := csv.headerTags[i]
			entry.AddTag(tag.TagFile, tag.TagName, value)
		}
	}
	return entry, nil
}

func (csv *WorkflowCSVFile) Close() {
	if csv.file != nil {
		csv.file.Close()
	}
}
