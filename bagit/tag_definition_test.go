package bagit_test

import (
	"testing"

	"github.com/APTrust/dart-runner/bagit"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTagDefIsLegalValue(t *testing.T) {
	tagDef := &bagit.TagDefinition{
		Values: []string{"one", "two", "three"},
	}
	assert.True(t, tagDef.IsLegalValue("one"))
	assert.True(t, tagDef.IsLegalValue("two"))
	assert.False(t, tagDef.IsLegalValue("six"))

	// If Values is nil or empty, any value is legal
	tagDef.Values = make([]string, 0)
	assert.True(t, tagDef.IsLegalValue("homer"))
	assert.True(t, tagDef.IsLegalValue("marge"))

	tagDef.UserValue = "homer"
	testTagDefinitionCopy(t, tagDef)

	tagDef.Values = nil
	assert.True(t, tagDef.IsLegalValue("homer"))
	assert.True(t, tagDef.IsLegalValue("marge"))
}

func testTagDefinitionCopy(t *testing.T, tagDef *bagit.TagDefinition) {
	copyOfTagDef := tagDef.Copy()
	require.NotNil(t, copyOfTagDef)

	// Values inside original and copy should be the same
	assert.Equal(t, tagDef, copyOfTagDef)

	// But the pointers should poind to different addresses
	assert.NotSame(t, tagDef, copyOfTagDef)
}

func TestTagDefGetValue(t *testing.T) {
	tagDef := &bagit.TagDefinition{}
	assert.Empty(t, tagDef.GetValue())

	tagDef.DefaultValue = "Homer"
	assert.Equal(t, "Homer", tagDef.GetValue())

	tagDef.UserValue = "Marge"
	assert.Equal(t, "Marge", tagDef.GetValue())
}

func TestTagDefToFormattedString(t *testing.T) {
	tagDef := &bagit.TagDefinition{
		TagName:   "Description",
		UserValue: "A bag of documents",
	}
	assert.Equal(t, "Description: A bag of documents", tagDef.ToFormattedString())

	tagDef.UserValue = `A
                        bag
                        of
                        documents
    `
	assert.Equal(t, "Description: A bag of documents", tagDef.ToFormattedString())
}
