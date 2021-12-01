package main_test

import (
	// "encoding/json"
	"fmt"
	//"os"
	"path"
	//"strings"
	"testing"
	//"time"

	main "github.com/APTrust/dart-runner"
	"github.com/APTrust/dart-runner/constants"
	"github.com/APTrust/dart-runner/core"
	"github.com/APTrust/dart-runner/util"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// NOTE: This file tests that functions in main run without error.
// For more extensive tests, including output json and bags created,
// see post_build_test.go.
//
// As in the post build test, these tests assume that scripts/test.rb
// has set up the output directory under ~/tmp/bags

func TestRunJob(t *testing.T) {
	opts := optsForJobParams(t)
	assert.Equal(t, constants.ExitOK, main.RunJob(opts))
}

func TestRunWorkflow(t *testing.T) {
	filesDir, _, outputDir := dirs(t)
	opts := &core.Options{
		WorkflowFilePath:  fmt.Sprintf("%s/postbuild_test_workflow.json", filesDir),
		BatchFilePath:     fmt.Sprintf("%s/postbuild_test_batch.csv", filesDir),
		OutputDir:         outputDir,
		DeleteAfterUpload: true,
		Concurrency:       1,
	}
	assert.Equal(t, constants.ExitOK, main.RunWorkflow(opts))
}

func TestInitParams(t *testing.T) {
	opts := optsForJobParams(t)
	jobParams, err := main.InitParams(opts)
	require.Nil(t, err)
	require.NotNil(t, jobParams)
}

// Note: post_build_test tests output of this function
func TestShowHelp(t *testing.T) {
	assert.NotPanics(t, func() { main.ShowHelp() })
}

// Note: post_build_test tests output of this function
func TestShowVersion(t *testing.T) {
	assert.NotPanics(t, func() { main.ShowVersion() })
}

func optsForJobParams(t *testing.T) *core.Options {
	filesDir, _, outputDir := dirs(t)
	jobParamsJson, err := util.ReadFile(path.Join(filesDir, "postbuild_test_params.json"))
	require.Nil(t, err)
	return &core.Options{
		OutputDir:        outputDir,
		WorkflowFilePath: fmt.Sprintf("%s/postbuild_test_workflow.json", filesDir),
		StdinData:        jobParamsJson,
	}
}
