package core

type WorkflowBatch struct {
	Errors map[string]string `json:"errors"`
	//JobParamsArray []*JobParams      `json:"jobParamsArray"`
	PathToCSVFile string    `json:"pathToCSVFile"`
	Workflow      *Workflow `json:"workflow"`
}

// TODO: Pass in path to output dir

func NewWorkflowBatch(workflow *Workflow, pathToCSVFile string) *WorkflowBatch {
	return &WorkflowBatch{
		Errors:        make(map[string]string),
		PathToCSVFile: pathToCSVFile,
		Workflow:      workflow,
	}
}

func (w *WorkflowBatch) Validate() bool {
	// Validate workflow
	// Validate CSV file - Bag-Name comes from CSV file

	return true
}

func (w *WorkflowBatch) Run() bool {
	// Parse the CSV file.
	// For each line in CSV file:
	//   Create new JobParams
	//   Convert to Job
	//   Validate the job
	//   Run the PackageOp if there is one
	//   Run the ValidationOp if there is one
	//   Run all of the UploadOps if there are any
	//   Collect errors, if any
	//   Continue on error to get through all we can
	//   Return true if all jobs succeeded, false if not

	return true
}
