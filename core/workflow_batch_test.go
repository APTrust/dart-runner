package core_test

import (
	"fmt"
	"os"
	"path"
	"strings"
	"testing"

	"github.com/APTrust/dart-runner/core"
	"github.com/APTrust/dart-runner/util"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewWorkflowBatchForm(t *testing.T) {
	defer core.ClearDartTable()
	workflow := loadJsonWorkflow(t)

	// Make 5 workflows, so we'll have something to
	// show in our select list.
	workflowIds := make([]string, 5)
	workflowNames := make([]string, 5)
	for i := 0; i < 5; i++ {
		name := fmt.Sprintf("Workflow %d", i)
		id := uuid.NewString()
		workflow.Name = name
		workflow.ID = id
		require.Nil(t, core.ObjSave(workflow))
		workflowIds[i] = id
		workflowNames[i] = name
	}

	// Create the form with "Workflow 3" as the selected
	// workflow.
	selectedId := workflowIds[2]
	workflow.ID = selectedId
	wb := core.NewWorkflowBatch(workflow, "/path/to/file.csv")
	form := wb.ToForm()
	require.Equal(t, 2, len(form.Fields))

	// Make sure path field is present and set correctly
	assert.Equal(t, "/path/to/file.csv", form.Fields["PathToCSVFile"].Value)

	// Now test the workflow choices. This should appear in alpha
	// order, and Workflow 3 should be selected.
	workflowChoices := form.Fields["WorkflowID"].Choices
	require.Equal(t, 5, len(workflowChoices))

	for i := 0; i < 5; i++ {
		expectedName := workflowNames[i]
		expectedId := workflowIds[i]
		assert.Equal(t, expectedName, workflowChoices[i].Label, i)
		assert.Equal(t, expectedId, workflowChoices[i].Value, i)
		if expectedId == selectedId {
			assert.True(t, workflowChoices[i].Selected, i)
		} else {
			assert.False(t, workflowChoices[i].Selected, i)
		}
	}
}

func TestWorkflowBatchValidateBadParams(t *testing.T) {
	wb := core.NewWorkflowBatch(nil, "/path/does/not/exist.csv")
	assert.False(t, wb.Validate())
	assert.Equal(t, 2, len(wb.Errors))
	assert.Contains(t, wb.Errors["PathToCSVFile"], "file does not exist")
	assert.Contains(t, wb.Errors["WorkflowID"], "Please choose a workflow")

	invalidWorkflow := &core.Workflow{}
	wb = core.NewWorkflowBatch(invalidWorkflow, "/path/does/not/exist.csv")
	assert.False(t, wb.Validate())
	assert.Equal(t, "Workflow requires a name.", wb.Errors["Workflow_Name"])
	assert.Equal(t, "Workflow requires a package format.", wb.Errors["Workflow_PackageFormat"])

	// Make sure that when validating the workflow inside this batch,
	// we validate the workflow's underlying BagItProfile.
	// Just check that one error comes through here. The fuller test
	// for workflow and profile validation occurs in the workflow
	// profile tests.
	invalidWorkflow.BagItProfile = &core.BagItProfile{}
	assert.False(t, wb.Validate())
	assert.Equal(t, 10, len(wb.Errors))
	assert.Equal(t, "Profile must allow at least one manifest algorithm.", wb.Errors["Workflow_BagItProfile.ManifestsAllowed"])

	// This json file is unparsable as CSV
	pathToJSONFile := path.Join(util.PathToTestData(), "files", "sample_job.json")
	workflow := loadJsonWorkflow(t)
	wb = core.NewWorkflowBatch(workflow, pathToJSONFile)
	assert.False(t, wb.Validate())
	assert.Contains(t, wb.Errors["CSVFile"], "parse error")
	assert.Contains(t, wb.Errors["CSVFile"], "Be sure this is a valid CSV file")
	// fmt.Println(wb.Errors)
}

func TestWorkflowBatchValidateCSVContents(t *testing.T) {
	workflow := loadJsonWorkflow(t)
	pathToBatchFile := path.Join(util.PathToTestData(), "files", "postbuild_test_batch.csv")

	// This file is mostly valid, but it contains relative paths
	// that DART won't be able to find. The validator should note those.
	wb := core.NewWorkflowBatch(workflow, pathToBatchFile)
	assert.False(t, wb.Validate())
	assert.Equal(t, 3, len(wb.Errors))
	assert.Equal(t, "Line 1: file or directory does not exist: './core'.", wb.Errors["./core"])
	assert.Equal(t, "Line 2: file or directory does not exist: './server/controllers'.", wb.Errors["./server/controllers"])
	assert.Equal(t, "Line 3: file or directory does not exist: './util'.", wb.Errors["./util"])

	// Make the relative paths absolute, and the validator should be
	// happy because this file contains a complete and valid set of tags.
	tmpFile := makeTempCSVFileWithValidPaths(t, pathToBatchFile)
	defer func() { os.Remove(tmpFile) }()
	wb = core.NewWorkflowBatch(workflow, tmpFile)
	assert.True(t, wb.Validate())

	// Now let's test a file with missing and invalid tag values.
	// The tag values in this file do not satisfy the requirements
	// of the workflow's BagIt profile. This file contains quite a
	// few errors, but we're only going to spot check a handful of
	// basic cases.
	csvWithBadTags := path.Join(util.PathToTestData(), "files", "csv_batch_invalid_tags.csv")
	wb = core.NewWorkflowBatch(workflow, csvWithBadTags)
	assert.False(t, wb.Validate())
	assert.Equal(t, "Value Spongebob for tag aptrust-info.txt/Access on line 1 is not in the list of allowed values.", wb.Errors["1-aptrust-info.txt/Access"])
	assert.Equal(t, "Required tag aptrust-info.txt/Storage-Option on line 1 is missing or empty.", wb.Errors["1-aptrust-info.txt/Storage-Option"])
	assert.Equal(t, "Bag-Name is missing from line 3", wb.Errors["3-Bag-Name"])
	assert.Equal(t, "Required tag aptrust-info.txt/Title on line 3 is missing or empty.", wb.Errors["3-aptrust-info.txt/Title"])

	fmt.Println(wb.Errors)
}

// Paths in pastbuild_test_batch.csv file are relative.
// Create a temp file with absolute paths.
func makeTempCSVFileWithValidPaths(t *testing.T, pathToCSVFile string) string {
	tempFilePath := path.Join(os.TempDir(), "temp_batch.csv")
	csvContents, err := os.ReadFile(pathToCSVFile)
	require.Nil(t, err)
	absPrefix := util.ProjectRoot() + string(os.PathSeparator)
	csvWithAbsPaths := strings.ReplaceAll(string(csvContents), "./", absPrefix)
	require.NoError(t, os.WriteFile(tempFilePath, []byte(csvWithAbsPaths), 0666))
	return tempFilePath
}
