package core

import (
	"github.com/APTrust/dart-runner/bagit"
)

// WorkflowCSVEntry represents a single entry from a workflow
// CSV file. Bag up whatever's in RootDir and run it through
// the workflow.
type WorkflowCSVEntry struct {
	BagName string
	RootDir string
	Tags    []*bagit.Tag
}

// NewWorkflowCSVEntry creates a new WorkflowCSVEntry.
func NewWorkflowCSVEntry(bagName, rootDir string) *WorkflowCSVEntry {
	return &WorkflowCSVEntry{
		BagName: bagName,
		RootDir: rootDir,
		Tags:    make([]*bagit.Tag, 0),
	}
}

// AddTag adds a tag to this entry. Tag values from the CSV file will
// be written into the bag.
func (entry *WorkflowCSVEntry) AddTag(tagFile, tagName, value string) {
	entry.Tags = append(entry.Tags, bagit.NewTag(tagFile, tagName, value))
}

// FindTags returns all tags with the given name from the specified
// tag file. Returns an empty list if there are no matches. If multiple
// instances of a tag exist, they'll be returned in the order they
// were defined in the CSV file. This is in keeping with the BagIt spec
// that says order may be important.
func (entry *WorkflowCSVEntry) FindTags(tagFile, tagName string) []*bagit.Tag {
	tags := make([]*bagit.Tag, 0)
	for _, tag := range entry.Tags {
		if tag.TagFile == tagFile && tag.TagName == tagName {
			tags = append(tags, tag)
		}
	}
	return tags
}
