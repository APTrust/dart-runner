package core

import (
	"github.com/APTrust/dart-runner/bagit"
)

type WorkflowCSVEntry struct {
	BagName string
	RootDir string
	Tags    []*bagit.Tag
}

func NewWorkflowCSVEntry(bagName, rootDir string) *WorkflowCSVEntry {
	return &WorkflowCSVEntry{
		BagName: bagName,
		RootDir: rootDir,
		Tags:    make([]*bagit.Tag, 0),
	}
}

func (entry *WorkflowCSVEntry) AddTag(tagFile, tagName, value string) {
	entry.Tags = append(entry.Tags, bagit.NewTag(tagFile, tagName, value))
}

func (entry *WorkflowCSVEntry) FindTags(tagFile, tagName string) []*bagit.Tag {
	tags := make([]*bagit.Tag, 0)
	for _, tag := range entry.Tags {
		if tag.TagFile == tagFile && tag.TagName == tagName {
			tags = append(tags, tag)
		}
	}
	return tags
}
