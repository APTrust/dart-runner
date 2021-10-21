package core

import (
	"flag"
)

type Options struct {
	JobFilePath       string
	WorkflowFilePath  string
	BatchFilePath     string
	OutputDir         string
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

	return &Options{
		JobFilePath:       *jobFilePath,
		WorkflowFilePath:  *workflowFilePath,
		BatchFilePath:     *batchFilePath,
		OutputDir:         *outputDir,
		Concurrency:       *concurrency,
		DeleteAfterUpload: *deleteAfterUpload,
		ShowHelp:          *showHelp,
		Version:           *version,
	}
}

func (opts Options) AreValid() bool {
	if opts.Version {
		return true
	}
	if opts.JobFilePath != "" && opts.OutputDir != "" {
		return true
	}
	if opts.WorkflowFilePath != "" && opts.BatchFilePath != "" && opts.OutputDir != "" {
		return true
	}
	return false
}
