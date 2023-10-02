package core_test

import (
	"fmt"
	"testing"

	"github.com/APTrust/dart-runner/core"
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
