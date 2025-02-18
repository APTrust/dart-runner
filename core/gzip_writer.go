package core

import (
	"compress/gzip"
	"io"
	"os"
)

type GzipWriter struct {
	InputFile string
}

func NewGzipWriter(inputFile string) *GzipWriter {
	return &GzipWriter{
		InputFile: inputFile,
	}
}

func (g *GzipWriter) ZipToFile(outfile string) (int64, error) {
	output, err := os.Create(outfile)
	if err != nil {
		return int64(0), err
	}
	defer output.Close()

	input, err := os.Open(g.InputFile)
	if err != nil {
		return int64(0), err
	}
	defer input.Close()

	zipwriter := gzip.NewWriter(output)
	defer zipwriter.Close()

	return io.Copy(zipwriter, input)
}
