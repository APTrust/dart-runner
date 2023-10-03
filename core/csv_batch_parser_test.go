package core_test

import (
	"path"
	"testing"

	"github.com/APTrust/dart-runner/core"
	"github.com/APTrust/dart-runner/util"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCSVBatchParserBadFile(t *testing.T) {
	workflow := loadJsonWorkflow(t)

	// File does not exist
	parser := core.NewCSVBatchParser("/path/-to-/nowhere/file.csv", workflow)
	jobParamsList, err := parser.ParseAll("/tmp")
	assert.Nil(t, jobParamsList)
	require.Error(t, err)
	assert.Equal(t, "open /path/-to-/nowhere/file.csv: no such file or directory", err.Error())

	// Attempt to parse a non-csv file should give us an error.
	jsonFile := path.Join(util.PathToTestData(), "files", "sample_job.json")
	parser = core.NewCSVBatchParser(jsonFile, workflow)
	jobParamsList, err = parser.ParseAll("/tmp")
	assert.Nil(t, jobParamsList)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "parse error")
}

func TestCSVBatchParserGoodFile(t *testing.T) {
	workflow := loadJsonWorkflow(t)
	csvFile := path.Join(util.PathToTestData(), "files", "postbuild_test_batch.csv")
	parser := core.NewCSVBatchParser(csvFile, workflow)
	jobParamsList, err := parser.ParseAll("/tmp/csvtest")
	require.Nil(t, err)
	require.NotEmpty(t, jobParamsList)

	assert.Equal(t, 3, len(jobParamsList))
	expectedTagLists := getExpectedTagLists()
	for i, jobParams := range jobParamsList {

		// Make sure the workflow and BagIt profile in the
		// jobParams look good.
		assert.Equal(t, workflow.ID, jobParams.Workflow.ID)
		assert.Equal(t, workflow.BagItProfile.ID, jobParams.Workflow.BagItProfile.ID)
		assert.Equal(t, len(workflow.BagItProfile.Tags), len(jobParams.Workflow.BagItProfile.Tags))

		// Make sure jobParams has the correct bag name,
		// output path and source file list.
		bagName := "unknown"
		fileList := make([]string, 0)
		for _, tag := range jobParams.Tags {
			if tag.TagName == "Bag-Name" {
				bagName = tag.Value
			} else if tag.TagName == "Root-Directory" {
				fileList = append(fileList, tag.Value)
			}
		}
		assert.Equal(t, bagName, jobParams.PackageName)
		assert.Equal(t, path.Join("/tmp/csvtest", bagName), jobParams.OutputPath)
		assert.NotEmpty(t, fileList)
		assert.Equal(t, fileList, jobParams.Files)

		// Make sure we got the tags we expected.
		expectedTags := expectedTagLists[i]
		assert.Equal(t, len(expectedTags), len(jobParams.Tags))
		for j, tag := range jobParams.Tags {
			assert.Equal(t, expectedTags[j].TagFile, tag.TagFile)
			assert.Equal(t, expectedTags[j].TagName, tag.TagName)
			assert.Equal(t, expectedTags[j].Value, tag.Value)
		}
	}
}

// These are the lists of tags we expect to parse from the CSV file
// in the test above.
func getExpectedTagLists() [][]*core.Tag {
	tagList1 := []*core.Tag{
		{
			TagFile: "bag-info.txt",
			TagName: "Bag-Name",
			Value:   "RunnerTestCore",
		},
		{
			TagFile: "bag-info.txt",
			TagName: "Root-Directory",
			Value:   "./core",
		},
		{
			TagFile: "aptrust-info.txt",
			TagName: "Title",
			Value:   "Runner Test - Core Files",
		},
		{
			TagFile: "aptrust-info.txt",
			TagName: "Description",
			Value:   "Go source files from DART runner: core directory. These files are bagged as part of the DART runner workflow tests.",
		},
		{
			TagFile: "aptrust-info.txt",
			TagName: "Access",
			Value:   "Consortia",
		},
		{
			TagFile: "bag-info.txt",
			TagName: "Source-Organization",
			Value:   "Test University",
		},
		{
			TagFile: "bag-info.txt",
			TagName: "Custom-Tag",
			Value:   "Custom Value (Core)",
		},
		{
			TagFile: "custom-tag-file.txt",
			TagName: "Tag-One",
			Value:   "Alpha",
		},
		{
			TagFile: "custom-tag-file.txt",
			TagName: "Tag-Two",
			Value:   "Abondance",
		},
		{
			TagFile: "bag-info.txt",
			TagName: "Repeater",
			Value:   "1",
		},
		{
			TagFile: "bag-info.txt",
			TagName: "Repeater",
			Value:   "2",
		},
	}

	tagList2 := []*core.Tag{
		{
			TagFile: "bag-info.txt",
			TagName: "Bag-Name",
			Value:   "RunnerTestServer",
		},
		{
			TagFile: "bag-info.txt",
			TagName: "Root-Directory",
			Value:   "./server/controllers",
		},
		{
			TagFile: "aptrust-info.txt",
			TagName: "Title",
			Value:   "Runner Test - Server Files",
		},
		{
			TagFile: "aptrust-info.txt",
			TagName: "Description",
			Value:   "Go source files from DART runner: bagit directory. These files are bagged as part of the DART runner workflow tests.",
		},
		{
			TagFile: "aptrust-info.txt",
			TagName: "Access",
			Value:   "Institution",
		},
		{
			TagFile: "bag-info.txt",
			TagName: "Source-Organization",
			Value:   "Staging University",
		},
		{
			TagFile: "bag-info.txt",
			TagName: "Custom-Tag",
			Value:   "Custom Value (BagIt)",
		},
		{
			TagFile: "custom-tag-file.txt",
			TagName: "Tag-One",
			Value:   "Bravo",
		},
		{
			TagFile: "custom-tag-file.txt",
			TagName: "Tag-Two",
			Value:   "Brie",
		},
		{
			TagFile: "bag-info.txt",
			TagName: "Repeater",
			Value:   "3",
		},
		{
			TagFile: "bag-info.txt",
			TagName: "Repeater",
			Value:   "4",
		},
	}

	tagList3 := []*core.Tag{
		{
			TagFile: "bag-info.txt",
			TagName: "Bag-Name",
			Value:   "RunnerTestUtil",
		},
		{
			TagFile: "bag-info.txt",
			TagName: "Root-Directory",
			Value:   "./util",
		},
		{
			TagFile: "aptrust-info.txt",
			TagName: "Title",
			Value:   "Runner Test - Util Files",
		},
		{
			TagFile: "aptrust-info.txt",
			TagName: "Description",
			Value:   "Go source files from DART runner: util directory. These files are bagged as part of the DART runner workflow tests.",
		},
		{
			TagFile: "aptrust-info.txt",
			TagName: "Access",
			Value:   "Restricted",
		},
		{
			TagFile: "bag-info.txt",
			TagName: "Source-Organization",
			Value:   "University of Virginia",
		},
		{
			TagFile: "bag-info.txt",
			TagName: "Custom-Tag",
			Value:   "Custom Value (Util)",
		},
		{
			TagFile: "custom-tag-file.txt",
			TagName: "Tag-One",
			Value:   "Charlie",
		},
		{
			TagFile: "custom-tag-file.txt",
			TagName: "Tag-Two",
			Value:   "Camembert",
		},
		{
			TagFile: "bag-info.txt",
			TagName: "Repeater",
			Value:   "5",
		},
		{
			TagFile: "bag-info.txt",
			TagName: "Repeater",
			Value:   "6",
		},
	}

	return [][]*core.Tag{tagList1, tagList2, tagList3}
}
