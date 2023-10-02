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

func NewWorkflowBatch(workflowID, pathToCSVFile string) (*WorkflowBatch, error) {
	result := ObjFind(workflowID)
	if result.Error != nil {
		return nil, result.Error
	}
	return &WorkflowBatch{
		Workflow:      result.Workflow(),
		PathToCSVFile: pathToCSVFile,
		Errors:        make(map[string]string),
	}, nil
}

func (wb *WorkflowBatch) Validate() bool {
	wb.Errors = make(map[string]string)

	// Validate Workflow
	if wb.Workflow == nil {
		wb.Errors["Workflow"] = "Workflow is missing. You must choose a workflow."
	}
	if !wb.Workflow.Validate() {
		for key, value := range wb.Workflow.Errors {
			wb.Errors["Workflow_"+key] = value
		}
	}

	// Validate CSV file
	if !util.FileExists(wb.PathToCSVFile) {
		wb.Errors["PathToCSVFile"] = "CSV file does not exist."
	}

	return len(wb.Errors) > 0
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
	form := NewForm("WorkflowBatch", "ID not applicable to this type", nil)
	form.AddField("PathToCSVFile", "CSV Batch File", wb.PathToCSVFile, true)
	workflowID := ""
	if wb.Workflow != nil {
		workflowID = wb.Workflow.ID
	}
	workflowField := form.AddField("WorkflowID", "Choose a Workflow", workflowID, true)
	workflowChoices := ObjChoiceList(constants.TypeWorkflow, []string{workflowID})
	emptyChoice := []Choice{{Label: "", Value: ""}}
	workflowField.Choices = append(emptyChoice, workflowChoices...)
	return form
}
