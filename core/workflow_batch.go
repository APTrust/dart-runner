package core

import "github.com/APTrust/dart-runner/util"

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

	return true
}

func (wb *WorkflowBatch) checkPaths(jobParams []*JobParams) bool {

	return true
}

func (wb *WorkflowBatch) checkRequiredTags(jobParams []*JobParams) bool {

	return true
}
