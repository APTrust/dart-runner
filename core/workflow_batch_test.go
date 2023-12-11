package core_test

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/APTrust/dart-runner/constants"
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

	// Make sure the uplaod field is present and has correct filter
	assert.NotEmpty(t, form.Fields["CsvUpload"])
	assert.Equal(t, ".csv", form.Fields["CsvUpload"].Attrs["accept"])

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
	pathToJSONFile := filepath.Join(util.PathToTestData(), "files", "sample_job.json")
	workflow := loadJsonWorkflow(t)
	wb = core.NewWorkflowBatch(workflow, pathToJSONFile)
	assert.False(t, wb.Validate())
	assert.Contains(t, wb.Errors["CSVFile"], "parse error")
	assert.Contains(t, wb.Errors["CSVFile"], "Be sure this is a valid CSV file")
	// fmt.Println(wb.Errors)
}

func TestWorkflowBatchValidateCSVContents(t *testing.T) {
	workflow := loadJsonWorkflow(t)
	pathToBatchFile := filepath.Join(util.PathToTestData(), "files", "postbuild_test_batch.csv")

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
	tmpFile := util.MakeTempCSVFileWithValidPaths(t, pathToBatchFile)
	defer func() { os.Remove(tmpFile) }()
	wb = core.NewWorkflowBatch(workflow, tmpFile)
	assert.True(t, wb.Validate())

	// Now let's test a file with missing and invalid tag values.
	// The tag values in this file do not satisfy the requirements
	// of the workflow's BagIt profile. This file contains quite a
	// few errors, but we're only going to spot check a handful of
	// basic cases.
	csvWithBadTags := filepath.Join(util.PathToTestData(), "files", "csv_batch_invalid_tags.csv")
	wb = core.NewWorkflowBatch(workflow, csvWithBadTags)
	assert.False(t, wb.Validate())
	assert.Equal(t, "Value Spongebob for tag aptrust-info.txt/Access on line 1 is not in the list of allowed values.", wb.Errors["1-aptrust-info.txt/Access"])
	assert.Equal(t, "Required tag aptrust-info.txt/Storage-Option on line 1 is missing or empty.", wb.Errors["1-aptrust-info.txt/Storage-Option"])
	assert.Equal(t, "Bag-Name is missing from line 3", wb.Errors["3-Bag-Name"])
	assert.Equal(t, "Required tag aptrust-info.txt/Title on line 3 is missing or empty.", wb.Errors["3-aptrust-info.txt/Title"])

	//fmt.Println(wb.Errors)
}

func TestWBPersistentObjectInterface(t *testing.T) {
	defer core.ClearDartTable()
	workflow := loadJsonWorkflow(t)
	pathToBatchFile := filepath.Join(util.PathToTestData(), "files", "postbuild_test_batch.csv")

	// We need valid paths in our batch file, or we won't be
	// able to save this due to validation errors.
	tempFile := util.MakeTempCSVFileWithValidPaths(t, pathToBatchFile)
	defer func() { os.Remove(tempFile) }()

	wb := core.NewWorkflowBatch(workflow, tempFile)
	assert.True(t, util.LooksLikeUUID(wb.ID))
	assert.Equal(t, wb.ID, wb.ObjID())
	assert.Equal(t, fmt.Sprintf("WorkflowBatch: Runner Test Workflow => %s", tempFile), wb.String())
	assert.Equal(t, constants.TypeWorkflowBatch, wb.ObjType())
	assert.Equal(t, fmt.Sprintf("Runner Test Workflow => %s", tempFile), wb.ObjName())
	assert.True(t, wb.IsDeletable())

	// We have to reload the workflow for these new batches
	// because it's a pointer, so changing the name on one
	// changes the name on all.
	wb2 := core.NewWorkflowBatch(loadJsonWorkflow(t), tempFile)
	wb2.Workflow.Name = "Second test workflow"
	wb3 := core.NewWorkflowBatch(loadJsonWorkflow(t), tempFile)
	wb3.Workflow.Name = "Third test workflow"

	// Make sure we can save these objects
	assert.NoError(t, core.ObjSave(wb))
	assert.NoError(t, core.ObjSave(wb2))
	assert.NoError(t, core.ObjSave(wb3))

	// Make sure we can retrieve them
	result := core.ObjFind(wb.ID)
	assert.NoError(t, result.Error)
	assert.NotEmpty(t, result.WorkflowBatch())

	result = core.ObjFind(wb3.ID)
	assert.NoError(t, result.Error)
	assert.NotEmpty(t, result.WorkflowBatch())

	// Make sure we can list them in order
	result = core.ObjList(constants.TypeWorkflowBatch, "obj_name", 10, 0)
	assert.NoError(t, result.Error)
	batches := result.WorkflowBatches
	require.Equal(t, 3, len(batches))
	assert.True(t, strings.HasPrefix(batches[0].ObjName(), "Runner Test Workflow"))
	assert.True(t, strings.HasPrefix(batches[1].ObjName(), "Second test workflow"))
	assert.True(t, strings.HasPrefix(batches[2].ObjName(), "Third test workflow"))

	// Make sure we can get these by type and name
	result = core.ObjByNameAndType(wb.ObjName(), wb.ObjType())
	require.NoError(t, result.Error)
	assert.Equal(t, wb.ID, result.WorkflowBatch().ID)

	result = core.ObjByNameAndType(wb2.ObjName(), wb2.ObjType())
	require.NoError(t, result.Error)
	assert.Equal(t, wb2.ID, result.WorkflowBatch().ID)

}
