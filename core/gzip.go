package core

import (
	"compress/gzip"
	"io"
	"os"
)

func GzipCompress(inputFile, outputFile string) (int64, error) {
	output, err := os.Create(outputFile)
	if err != nil {
		return int64(0), err
	}
	defer output.Close()

	input, err := os.Open(inputFile)
	if err != nil {
		return int64(0), err
	}
	defer input.Close()

	zipwriter := gzip.NewWriter(output)
	defer zipwriter.Close()

	return io.Copy(zipwriter, input)
}

func GzipInflate(inputFile, outputFile string) (int64, error) {
	input, err := os.Open(inputFile)
	if err != nil {
		return int64(0), err
	}
	defer input.Close()

	output, err := os.Create(outputFile)
	if err != nil {
		return int64(0), err
	}
	defer output.Close()

	zipreader, err := gzip.NewReader(input)
	if err != nil {
		return int64(0), err
	}
	defer zipreader.Close()

	return io.Copy(output, zipreader)
}
