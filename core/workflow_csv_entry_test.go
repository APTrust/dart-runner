package core_test

import (
	"testing"

	"github.com/APTrust/dart-runner/core"
	"github.com/stretchr/testify/assert"
)

func TestWorkflowCSVEntry(t *testing.T) {
	entry := core.NewWorkflowCSVEntry("photos.tar", "/usr/local/archive")
	assert.Equal(t, "photos.tar", entry.BagName)
	assert.Equal(t, "/usr/local/archive", entry.RootDir)
	assert.Empty(t, entry.Tags)

	entry.AddTag("file-1.txt", "tag-1", "one")
	entry.AddTag("file-2.txt", "tag-2", "two")
	entry.AddTag("file-3.txt", "tag-3", "three")
	entry.AddTag("file-3.txt", "tag-3", "three again")

	tags := entry.FindTags("file-1.txt", "tag-1")
	assert.Equal(t, 1, len(tags))
	assert.Equal(t, "file-1.txt", tags[0].TagFile)
	assert.Equal(t, "tag-1", tags[0].TagName)
	assert.Equal(t, "one", tags[0].Value)

	// If no match, we should get empty slice
	tags = entry.FindTags("file-100.txt", "tag-1")
	assert.Equal(t, 0, len(tags))

	// Multiple tags should come back in order
	tags = entry.FindTags("file-3.txt", "tag-3")
	assert.Equal(t, "three", tags[0].Value)
	assert.Equal(t, "three again", tags[1].Value)
}
