package core_test

import (
	"strings"
	"testing"

	"github.com/APTrust/dart-runner/core"
	"github.com/stretchr/testify/assert"
	//"github.com/stretchr/testify/require"
)

var sampleJson = `
{
   "key1": "value one",
   "key2": "multiline
            value
            with
            embedded
            newlines"
}
`

func TestGetStdinData(t *testing.T) {
	reader := strings.NewReader(sampleJson)
	str := string(core.ReadInput(reader))
	assert.Equal(t, sampleJson, str)
}

func TestOptionsAreValid(t *testing.T) {
	opts := &core.Options{
		JobFilePath: "/path/to/job_file.json",
	}

	// Not valid: missing output dir
	assert.False(t, opts.AreValid())

	// Job file + output dir = valid
	opts.OutputDir = "/some/dir"
	assert.True(t, opts.AreValid())

	// Not valid without one of: job, workflow, stdin data
	opts.JobFilePath = ""
	assert.False(t, opts.AreValid())

	// Not valid because workflow requires batch file
	opts.WorkflowFilePath = "/path/to/workflow"
	assert.False(t, opts.AreValid())

	// output dir + workflow + batch = valid
	opts.BatchFilePath = "/path/to/batch.csv"
	assert.True(t, opts.AreValid())

	// clear for next test...
	opts.WorkflowFilePath = ""
	opts.BatchFilePath = ""
	assert.False(t, opts.AreValid())

	// output dir + stdin data = valid
	opts.StdinData = []byte(sampleJson)
	assert.True(t, opts.AreValid())

	// empty opts are invalid
	opts = &core.Options{}
	assert.False(t, opts.AreValid())

	// version is a valid option. User just wants to print version & exit.
	opts.Version = true
	assert.True(t, opts.AreValid())
}
