package controllers_test

import (
	"fmt"
	"net/http"
	"net/url"
	"testing"

	"github.com/APTrust/dart-runner/core"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestJobShowMetadata(t *testing.T) {
	defer core.ClearDartTable()
	job := loadTestJob(t)
	assert.NoError(t, core.ObjSave(job))

	expected := []string{}
	for _, tag := range job.BagItProfile.Tags {
		expected = append(expected, tag.TagName, tag.GetValue())
	}

	DoSimpleGetTest(t, fmt.Sprintf("/jobs/metadata/%s", job.ID), expected)
}

func TestJobSaveMetadata(t *testing.T) {
	defer core.ClearDartTable()
	job := loadTestJob(t)
	assert.NoError(t, core.ObjSave(job))

	tagsToSet := map[string]string{
		"aptrust-info.txt/Title":           "This is the new title",
		"aptrust-info.txt/Description":     "This is the new description",
		"bag-info.txt/Source-Organization": "The Krusty Krab",
	}

	// Emulate "Next" button click.
	params := url.Values{}
	params.Add("direction", "next")

	for key, value := range tagsToSet {
		params.Add(key, value)
	}

	postTestSettings := PostTestSettings{
		EndpointUrl:              fmt.Sprintf("/jobs/metadata/%s", job.ID),
		Params:                   params,
		ExpectedResponseCode:     http.StatusFound,
		ExpectedRedirectLocation: fmt.Sprintf("/jobs/upload/%s", job.ID),
	}
	DoSimplePostTest(t, postTestSettings)

	// Emulate "Previous" button click.
	params.Set("direction", "previous")
	postTestSettings.ExpectedRedirectLocation = fmt.Sprintf("/jobs/packaging/%s", job.ID)
	DoSimplePostTest(t, postTestSettings)

	// Make sure settings were saved.
	result := core.ObjFind(job.ID)
	require.Nil(t, result.Error)
	job = result.Job()

	for key, value := range tagsToSet {
		tag := job.BagItProfile.GetTagByFullyQualifiedName(key)
		require.NotNil(t, tag, key)
		assert.Equal(t, value, tag.GetValue(), tag.FullyQualifiedName())
	}
}

func TestJobSaveTag(t *testing.T) {

}

func TestJobDeleteTag(t *testing.T) {

}
