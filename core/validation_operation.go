package core

import (
	"fmt"
	"strings"

	"github.com/APTrust/dart-runner/util"
)

type ValidationOperation struct {
	Errors    map[string]string `json:"errors"`
	PathToBag string            `json:"pathToBag"`
	Result    *OperationResult  `json:"result"`
}

func NewValidationOperation(pathToBag string) *ValidationOperation {
	return &ValidationOperation{
		Errors:    make(map[string]string),
		PathToBag: pathToBag,
		Result:    NewOperationResult("validation", "validator"),
	}
}

func (op *ValidationOperation) Validate() bool {
	op.Errors = make(map[string]string)
	if strings.TrimSpace(op.PathToBag) == "" {
		op.Errors["ValidationOperation.pathToBag"] = "You must specify the path to the bag you want to validate."
	} else if !util.FileExists(op.PathToBag) {
		op.Errors["ValidationOperation.pathToBag"] = fmt.Sprintf("The bag to be validated does not exist at %s", op.PathToBag)
	}
	return len(op.Errors) == 0
}
