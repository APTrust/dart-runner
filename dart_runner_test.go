package main_test

import (
	"archive/tar"
	// "encoding/json"
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"

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

func init() {
	os.Setenv("DART_ENV", "test")
}

func TestRunJob(t *testing.T) {
	opts := optsForJobParams(t)
	assert.Equal(t, constants.ExitOK, main.RunJob(opts))
}

// Test for https://github.com/APTrust/dart-runner/issues/18
func TestSingleRelativePathBug(t *testing.T) {
	filesDir, _, outputDir := dirs(t)
	jobParamsJson, err := util.ReadFile(filepath.Join(filesDir, "job_params_relative_path.json"))
	require.Nil(t, err)
	opts := &core.Options{
		OutputDir:        outputDir,
		WorkflowFilePath: fmt.Sprintf("%s/postbuild_bag_only_workflow.json", filesDir),
		StdinData:        jobParamsJson,
	}
	require.Equal(t, constants.ExitOK, main.RunJob(opts))

	// Bag name RelativePaths.tar comes from job_params_relative_path.json
	tarredBag := path.Join(outputDir, "RelativePaths.tar")

	f, err := os.Open(tarredBag)
	require.Nil(t, err)
	defer f.Close()

	tr := tar.NewReader(f)
	foundExpected := false
	foundBadPath := false
	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break
		}
		require.Nil(t, err)
		// This is the file we SHOULD find if
		// https://github.com/APTrust/dart-runner/issues/18
		// is fixed.
		if hdr.Name == "RelativePaths/data/core/bagger.go" {
			foundExpected = true
		}
		// If this file appears in the bag, then
		// https://github.com/APTrust/dart-runner/issues/18
		// is not fixed, or we've had a regression.
		if hdr.Name == "RelativePaths/data/corebagger.go" {
			foundBadPath = true
		}
	}
	assert.True(t, foundExpected, "expected RelativePaths/data/core/bagger.go in tar")
	assert.False(t, foundBadPath, "RelativePaths/data/corebagger.go should not exist in tar")
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
	jobParamsJson, err := util.ReadFile(filepath.Join(filesDir, "postbuild_test_params.json"))
	require.Nil(t, err)
	return &core.Options{
		OutputDir:        outputDir,
		WorkflowFilePath: fmt.Sprintf("%s/postbuild_test_workflow.json", filesDir),
		StdinData:        jobParamsJson,
	}
}
