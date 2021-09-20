package core

type WorkflowBatch struct {
	Errors         map[string]string `json:"errors"`
	JobParamsArray []*JobParams      `json:"jobParamsArray"`
	PathToCSVFile  string            `json:"pathToCSVFile"`
	WorkflowID     string            `json:"workflowId"`
}
