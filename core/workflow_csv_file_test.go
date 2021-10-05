package core_test

import (
	"encoding/json"
	"io"
	"path"
	"testing"

	"github.com/APTrust/dart-runner/core"
	"github.com/APTrust/dart-runner/util"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWorkflowCSVFile(t *testing.T) {
	pathToFile := path.Join(util.PathToTestData(), "files", "csv_workflow_batch.csv")
	csv, err := core.NewWorkflowCSVFile(pathToFile)
	require.Nil(t, err)
	require.NotNil(t, csv)
	require.Equal(t, 7, len(csv.Headers()))
	require.Equal(t, 7, len(csv.HeaderTags()))

	expected := make([]*core.WorkflowCSVEntry, 0)
	err = json.Unmarshal([]byte(expectedEntries), &expected)
	require.Nil(t, err)

	for i := 0; ; i++ {
		entry, err := csv.ReadNext()
		if err == io.EOF {
			break
		}
		require.Nil(t, err)

		expectedEntry := expected[i]

		assert.Equal(t, expectedEntry.BagName, entry.BagName)
		assert.Equal(t, expectedEntry.RootDir, entry.RootDir)

		// There are 7 columns/headers, but two are
		// BagName and RootDir. The other 5 are tags.
		assert.Equal(t, 5, len(entry.Tags))

		for j, tag := range entry.Tags {
			assert.Equal(t, expectedEntry.Tags[j].TagFile, tag.TagFile)
			assert.Equal(t, expectedEntry.Tags[j].TagName, tag.TagName)
			assert.Equal(t, expectedEntry.Tags[j].Value, tag.Value)
		}
	}
}

var expectedEntries = `
[{
	"BagName": "bag_one",
	"RootDir": "/users/joe/photos",
	"Tags": [{
		"tagFile": "aptrust-info.txt",
		"tagName": "Title",
		"value": "Bag of Photos"
	}, {
		"tagFile": "aptrust-info.txt",
		"tagName": "Description",
		"value": "A bag of joe's photos"
	}, {
		"tagFile": "aptrust-info.txt",
		"tagName": "Access",
		"value": "Consortia"
	}, {
		"tagFile": "bag-info.txt",
		"tagName": "Source-Organization",
		"value": "Test University"
	}, {
		"tagFile": "bag-info.txt",
		"tagName": "Custom-Tag",
		"value": "Custom value one"
	}]
}, {
	"BagName": "bag_two",
	"RootDir": "/users/amy/music",
	"Tags": [{
		"tagFile": "aptrust-info.txt",
		"tagName": "Title",
		"value": "Amy's Music"
	}, {
		"tagFile": "aptrust-info.txt",
		"tagName": "Description",
		"value": "A whole bunch of MP3's ripped from Amy's old CD collection."
	}, {
		"tagFile": "aptrust-info.txt",
		"tagName": "Access",
		"value": "Institution"
	}, {
		"tagFile": "bag-info.txt",
		"tagName": "Source-Organization",
		"value": "Staging University"
	}, {
		"tagFile": "bag-info.txt",
		"tagName": "Custom-Tag",
		"value": "Custom value two"
	}]
}, {
	"BagName": "bag_three",
	"RootDir": "/var/www/news",
	"Tags": [{
		"tagFile": "aptrust-info.txt",
		"tagName": "Title",
		"value": "News Site Archive"
	}, {
		"tagFile": "aptrust-info.txt",
		"tagName": "Description",
		"value": "Snapshot of news site from summer 2020"
	}, {
		"tagFile": "aptrust-info.txt",
		"tagName": "Access",
		"value": "Restricted"
	}, {
		"tagFile": "bag-info.txt",
		"tagName": "Source-Organization",
		"value": "University of Virginia"
	}, {
		"tagFile": "bag-info.txt",
		"tagName": "Custom-Tag",
		"value": "Custom value three"
	}]
}]
`
