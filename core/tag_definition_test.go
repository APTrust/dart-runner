package core_test

import (
	"strings"
	"testing"

	"github.com/APTrust/dart-runner/core"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTagDefIsLegalValue(t *testing.T) {
	tagDef := &core.TagDefinition{
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

func testTagDefinitionCopy(t *testing.T, tagDef *core.TagDefinition) {
	copyOfTagDef := tagDef.Copy()
	require.NotNil(t, copyOfTagDef)

	// Values inside original and copy should be the same
	assert.Equal(t, tagDef, copyOfTagDef)

	// But the pointers should poind to different addresses
	assert.NotSame(t, tagDef, copyOfTagDef)
}

func TestTagDefGetValue(t *testing.T) {
	tagDef := &core.TagDefinition{}
	assert.Empty(t, tagDef.GetValue())

	tagDef.DefaultValue = "Homer"
	assert.Equal(t, "Homer", tagDef.GetValue())

	tagDef.UserValue = "Marge"
	assert.Equal(t, "Marge", tagDef.GetValue())
}

func TestTagDefToFormattedString(t *testing.T) {
	tagDef := &core.TagDefinition{
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

func TestTagDefToForm(t *testing.T) {
	tagDef := &core.TagDefinition{
		ID:           uuid.NewString(),
		TagName:      "FavoriteColor",
		TagFile:      "custom-tags.txt",
		Help:         "This is the help text.",
		Values:       []string{"red", "green", "blue"},
		DefaultValue: "green",
	}
	form := tagDef.ToForm()
	assert.Equal(t, tagDef.ID, form.Fields["ID"].Value)
	assert.Equal(t, tagDef.TagName, form.Fields["TagName"].Value)
	assert.Equal(t, tagDef.TagFile, form.Fields["TagFile"].Value)
	assert.Empty(t, form.Fields["TagName"].Attrs["readonly"])
	assert.Empty(t, form.Fields["TagFile"].Attrs["readonly"])
	assert.Equal(t, tagDef.DefaultValue, form.Fields["DefaultValue"].Value)
	assert.Equal(t, tagDef.Help, form.Fields["Help"].Value)
	assert.Equal(t, "false", form.Fields["Required"].Value)
	assert.Equal(t, 3, len(form.Fields["Required"].Choices))
	assert.Equal(t, len(tagDef.Values), len(strings.Split(form.Fields["Values"].Value, "\n")))
	assert.Contains(t, form.Fields["Values"].Value, "red")
	assert.Contains(t, form.Fields["Values"].Value, "green")
	assert.Contains(t, form.Fields["Values"].Value, "blue")

	tagDef.IsBuiltIn = true
	form = tagDef.ToForm()
	assert.Equal(t, "readonly", form.Fields["TagName"].Attrs["readonly"])
	assert.Equal(t, "readonly", form.Fields["TagFile"].Attrs["readonly"])
}

func TestTagDefValidate(t *testing.T) {
	tagDef := &core.TagDefinition{}
	assert.False(t, tagDef.Validate())
	assert.Equal(t, 2, len(tagDef.Errors))
	assert.Equal(t, "You must specify a tag name.", tagDef.Errors["TagName"])
	assert.Equal(t, "You must specify a tag file.", tagDef.Errors["TagFile"])

	tagDef.TagName = "Test Tag"
	tagDef.TagFile = "Test File"
	assert.True(t, tagDef.Validate())
	assert.Empty(t, tagDef.Errors)

	tagDef.Values = []string{
		"Spongebob",
		"Patrick",
		"Mister Crabs",
	}
	assert.True(t, tagDef.Validate())
	assert.Empty(t, tagDef.Errors)

	tagDef.DefaultValue = "Alice Cooper"
	tagDef.UserValue = "Taylor Swift"

	assert.False(t, tagDef.Validate())
	assert.Equal(t, 2, len(tagDef.Errors))
	assert.Equal(t, "The default value must be one of the allowed values.", tagDef.Errors["DefaultValue"])
	assert.Equal(t, "The value must be one of the allowed values.", tagDef.Errors["UserValue"])
}
