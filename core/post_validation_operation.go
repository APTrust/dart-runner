package core

import (
	"fmt"
	"strings"

	"github.com/APTrust/dart-runner/constants"
)

type PostValidationOperation struct {
	Errors           map[string]string `json:"errors"`
	Command          string            `json:"goCommand"`
	CommandArgs      []string          `json:"commandArgs"`
	CommandType      string            `json:"commandType"`
	NamedCommandArgs map[string]string `json:"namedCommandArgs"`
	Result           *OperationResult  `json:"result"`
}

func NewPostValidationGoOp(goCommand string, namedCommandArgs map[string]string) *PostValidationOperation {
	return &PostValidationOperation{
		Command:          goCommand,
		CommandArgs:      make([]string, 0),
		CommandType:      constants.PostValidateCommandTypeGo,
		NamedCommandArgs: namedCommandArgs,
		Errors:           make(map[string]string),
		Result:           NewOperationResult("post validate", "DART - "+constants.AppVersion),
	}
}

func NewPostValidationSyspemOp(systemCommand string, args ...string) *PostValidationOperation {
	// TODO: Whitelist system commands.
	return &PostValidationOperation{
		Command:          systemCommand,
		CommandArgs:      args,
		CommandType:      constants.PostValidateCommandTypeSystem,
		NamedCommandArgs: make(map[string]string),
		Errors:           make(map[string]string),
		Result:           NewOperationResult("post validate", "DART - "+constants.AppVersion),
	}
}

func NewGzipAfterValidation(inputFile string, outputFile string) *PostValidationOperation {
	namedCommandArgs := make(map[string]string)
	namedCommandArgs["inputFile"] = inputFile
	namedCommandArgs["outputFile"] = outputFile
	return &PostValidationOperation{
		Command:          constants.PostValidateGzipCommand,
		CommandArgs:      []string{},
		NamedCommandArgs: namedCommandArgs,
		Errors:           make(map[string]string),
		Result:           NewOperationResult("post validate", "DART.gzip - "+constants.AppVersion),
	}
}

func (op *PostValidationOperation) Validate() bool {
	op.Errors = make(map[string]string)
	if strings.TrimSpace(op.Command) == "" {
		op.Errors["PostValidationOp.command"] = "You must specify either a Go command or a system command."
	}
	// TODO: Validate that go command has named args and sys command has args?
	for key, value := range op.Errors {
		Dart.Log.Infof("%s: %s", key, value)
	}
	return len(op.Errors) == 0
}

func (op *PostValidationOperation) Run(messageChannel chan *EventMessage) error {
	var err error
	// For now, the only operation we support is gzip, which is a golang operation.

	if messageChannel != nil {
		//progress = NewStreamProgress(u.PayloadSize, messageChannel)
		messageChannel <- StartEvent(constants.StagePostValidation, fmt.Sprintf("Running post-validation command %s", op.Command))
	}
	switch op.Command {
	case constants.PostValidateGzipCommand:
		_, err = GzipCompress(op.NamedCommandArgs["inputFile"], op.NamedCommandArgs["outputFile"])
	default:
		err = fmt.Errorf("Unsupported post validation operation: %s", op.Command)
	}
	return err
}
