package core

import (
	"github.com/APTrust/dart-runner/bagit"
)

type JobParams struct {
	Errors         map[string]string      `json:"errors"`
	Files          []string               `json:"files"`
	PackageName    string                 `json:"packageName"`
	Tags           []*bagit.TagDefinition `json:"tags"`
	WorkflowName   string                 `json:"workflowName"`
	bagItProfile   *bagit.Profile
	workflowObject *Workflow
}
