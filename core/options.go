package core

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"os"

	"github.com/APTrust/dart-runner/constants"
)

type Options struct {
	JobFilePath       string
	WorkflowFilePath  string
	BatchFilePath     string
	OutputDir         string
	StdinData         []byte
	Concurrency       int
	DeleteAfterUpload bool
	ShowHelp          bool
	Version           bool
}

func ParseOptions() *Options {
	jobFilePath := flag.String("job", "", "Path to job json file")
	workflowFilePath := flag.String("workflow", "", "Path to workflow json file")
	batchFilePath := flag.String("batch", "", "Path to csv batch file")
	outputDir := flag.String("output-dir", "", "Path to output directory")
	concurrency := flag.Int("concurrency", 1, "Number of jobs to run simultaneously")
	deleteAfterUpload := flag.Bool("delete", true, "Delete bags after upload? true|false - Default = true.")
	showHelp := flag.Bool("help", false, "Show help.")
	version := flag.Bool("version", false, "Show version and exit.")

	flag.Parse()

	var jsonData []byte
	if StdinHasData() {
		jsonData = ReadInput(os.Stdin)
	}

	return &Options{
		JobFilePath:       *jobFilePath,
		WorkflowFilePath:  *workflowFilePath,
		BatchFilePath:     *batchFilePath,
		OutputDir:         *outputDir,
		Concurrency:       *concurrency,
		DeleteAfterUpload: *deleteAfterUpload,
		ShowHelp:          *showHelp,
		Version:           *version,
		StdinData:         jsonData,
	}
}

// AreValid returns true if options are valid. That is, they contain enough
// info for DART Runner to proceed.
func (opts Options) AreValid() bool {
	if opts.Version {
		return true
	}
	if opts.JobFilePath != "" && opts.OutputDir != "" {
		return true
	}
	if len(opts.StdinData) > 0 && opts.OutputDir != "" {
		// We'll validate stdin json later
		return true
	}
	if opts.WorkflowFilePath != "" && opts.BatchFilePath != "" && opts.OutputDir != "" {
		return true
	}
	return false
}

// ReadInput reads input from the specified reader (which, in practice,
// will usually be STDIN) and returns it with newlines preserved.
func ReadInput(reader io.Reader) []byte {
	var stdinData []byte
	scanner := bufio.NewScanner(reader)
	for scanner.Scan() {
		stdinData = append(stdinData, scanner.Bytes()...)
		stdinData = append(stdinData, []byte("\n")...)
	}
	return stdinData
}

// StdinHasData returns true if STDIN has data waiting to be read.
// This exits immediately if it can't access or stat STDIN.
func StdinHasData() bool {
	fi, err := os.Stdin.Stat()
	if err != nil {
		fmt.Fprintln(os.Stdout, "Error checking STDIN:", err)
		os.Exit(constants.ExitRuntimeErr)
	}
	return fi.Size() > 0
}