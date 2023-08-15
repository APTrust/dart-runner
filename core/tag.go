package core

import "fmt"

// Tag describes a tag parsed from a BagIt file such as bag-info.txt.
type Tag struct {
	TagFile string `json:"tagFile"`
	TagName string `json:"tagName"`
	Value   string `json:"value"`
}

// NewTag returns a new Tag object. Params are self-explanatory.
func NewTag(sourceFile, label, value string) *Tag {
	return &Tag{
		TagFile: sourceFile,
		TagName: label,
		Value:   value,
	}
}

// FullyQualifiedName returns the tag's fully qualified name,
// which is the tag file followed by a slash followed by the
// tag name. E.g. bag-info.txt/Source-Organization.
func (t *Tag) FullyQualifiedName() string {
	return fmt.Sprintf("%s/%s", t.TagFile, t.TagName)
}
