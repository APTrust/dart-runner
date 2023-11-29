package core

import (
	"encoding/csv"
	"fmt"
	"os"
	"strings"

	"github.com/dimchansky/utfbom"
)

// WorkflowCSVFile reads a CSV file that describes what items should
// be run through a workflow.
type WorkflowCSVFile struct {
	PathToFile string
	reader     *csv.Reader
	headers    []string
	headerTags []*Tag
	file       *os.File
}

// NewWorkflowCSVFile creates a new WorkflowCSVFile object. See the
// ReadNext() function for extracting data from the file.
func NewWorkflowCSVFile(pathToFile string) (*WorkflowCSVFile, error) {
	f, err := os.Open(pathToFile)
	if err != nil {
		Dart.Log.Error("Can't open CSV file %s: %v", pathToFile, err)
		return nil, err
	}
	// If the CSV file was exported from Excel, it probably
	// starts with a byte order marker (BOM) that will trip up a
	// pure CSV parser. The line below removed the BOM if it exists.
	sr, _ := utfbom.Skip(f)
	csvFile := &WorkflowCSVFile{
		PathToFile: pathToFile,
		reader:     csv.NewReader(sr),
		file:       f,
	}
	err = csvFile.parseHeaders()
	if err != nil {
		f.Close()
		Dart.Log.Error("Error parsing headers in CSV file %s: %v", pathToFile, err)
		return nil, err
	}
	return csvFile, nil
}

// parseHeaders parses the headers in the first line of the file.
func (csvFile *WorkflowCSVFile) parseHeaders() error {
	headers, err := csvFile.reader.Read()
	if err != nil {
		return err
	}
	headerTags := make([]*Tag, len(headers))
	hasBagName := false
	hasRootDir := false
	for i, h := range headers {
		if h == "Bag-Name" {
			hasBagName = true
			headerTags[i] = NewTag("", h, "")
			continue
		}
		if h == "Root-Directory" {
			hasRootDir = true
			headerTags[i] = NewTag("", h, "")
			continue
		}
		parts := strings.Split(h, "/")
		if len(parts) != 2 {
			return fmt.Errorf("Bag tag header '%s' in column %d. Header name should use tagFile/tagName pattern.", h, i)
		}
		headerTags[i] = NewTag(parts[0], parts[1], "")
	}
	if !hasBagName {
		return fmt.Errorf("CSV file is missing Bag-Name column")
	}
	if !hasRootDir {
		return fmt.Errorf("CSV file is missing Root-Directory column")
	}
	csvFile.headers = headers
	csvFile.headerTags = headerTags
	return nil
}

// Headers returns the headers (column names) from the first line of
// the file. Other than the required headers Bag-Name and Root-Directory,
// all headers should be in the format FileName/TagName. For example,
// "bag-info.txt/Source-Organization".
func (csvFile *WorkflowCSVFile) Headers() []string {
	return csvFile.headers
}

// HeaderTags returns the headers from the first line, converted to
// tag objects with empty values. This is primarily for internal use,
// but may be useful in debugging and error reporting.
//
// Tags Bag-Name and Root-Directory are essentially throw-aways. The
// bagger won't use them. All other tags will have TagFile and TagName
// attributes, with an empty Value. These are used internally to help
// construct the tags in each WorkflowCSVEntry.
func (csvFile *WorkflowCSVFile) HeaderTags() []*Tag {
	return csvFile.headerTags
}

// ReadNext reads the next entry in the CSV file, returning a
// WorkflowCSVEntry. If this returns error io.EOF, it's finished
// reading and there are no more entries. If it returns any other
// error, something went wrong.
func (csvFile *WorkflowCSVFile) ReadNext() (*WorkflowCSVEntry, error) {
	record, err := csvFile.reader.Read()
	if err != nil {
		return nil, err
	}
	entry := NewWorkflowCSVEntry("", "")
	for i, value := range record {
		if csvFile.headers[i] == "Bag-Name" {
			entry.BagName = value
		} else if csvFile.headers[i] == "Root-Directory" {
			entry.RootDir = value
		} else {
			tag := csvFile.headerTags[i]
			entry.AddTag(tag.TagFile, tag.TagName, value)
		}
	}
	return entry, nil
}

// Close closes the CSV's underlying file object.
func (csvFile *WorkflowCSVFile) Close() {
	if csvFile.file != nil {
		csvFile.file.Close()
	}
}
