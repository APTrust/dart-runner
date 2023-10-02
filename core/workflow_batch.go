package core

import (
	"github.com/APTrust/dart-runner/constants"
	"github.com/APTrust/dart-runner/util"
)

type WorkflowBatch struct {
	Workflow      *Workflow
	PathToCSVFile string
	Errors        map[string]string
}

func NewWorkflowBatch(workflow *Workflow, pathToCSVFile string) *WorkflowBatch {
	return &WorkflowBatch{
		Workflow:      workflow,
		PathToCSVFile: pathToCSVFile,
		Errors:        make(map[string]string),
	}
}

func (wb *WorkflowBatch) Validate() bool {
	wb.Errors = make(map[string]string)

	// Validate Workflow
	if wb.Workflow == nil {
		wb.Errors["WorkflowID"] = "Please choose a workflow."
	} else if !wb.Workflow.Validate() {
		for key, value := range wb.Workflow.Errors {
			wb.Errors["Workflow_"+key] = value
		}
	}

	// Validate CSV file
	if wb.PathToCSVFile == "" {
		wb.Errors["PathToCSVFile"] = "Please chose a CSV file."
	} else if !util.FileExists(wb.PathToCSVFile) {
		wb.Errors["PathToCSVFile"] = "CSV file does not exist."
	}

	return len(wb.Errors) == 0
}

func (wb *WorkflowBatch) validateCSVFile() bool {
	// headers, records, err := util.ParseCSV(wb.PathToCSVFile)
	// if err != nil {
	// 	wb.Errors["CSVFile"] = err.Error()
	// 	return false
	// }
	// for i, record := range records {
	// 	lineNumber := i + 1
	// 	if (!util.FileExists(record[""]))
	// }
	return true
}

func (wb *WorkflowBatch) checkPaths(jobParams []*JobParams) bool {

	return true
}

func (wb *WorkflowBatch) checkRequiredTags(jobParams []*JobParams) bool {

	return true
}

func (wb *WorkflowBatch) ToForm() *Form {
	form := NewForm("WorkflowBatch", "ID not applicable to this type", wb.Errors)
	form.AddField("PathToCSVFile", "CSV Batch File", wb.PathToCSVFile, true)
	workflowID := ""
	if wb.Workflow != nil {
		workflowID = wb.Workflow.ID
	}
	workflowField := form.AddField("WorkflowID", "Choose a Workflow", workflowID, true)
	workflowField.Choices = ObjChoiceList(constants.TypeWorkflow, []string{workflowID})
	return form
}
