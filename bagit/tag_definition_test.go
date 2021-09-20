package bagit_test

import (
	"github.com/APTrust/dart-runner/bagit"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestTagDefinitionIsLegalValue(t *testing.T) {
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

	tagDef.Values = nil
	assert.True(t, tagDef.IsLegalValue("homer"))
	assert.True(t, tagDef.IsLegalValue("marge"))
}
