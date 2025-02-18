package core

import (
	"compress/gzip"
	"io"
	"os"
)

type GzipReader struct {
	InputFile string
	file      *os.File
	zipreader *gzip.Reader
}

func NewGzipReader(inputFile string) *GzipReader {
	return &GzipReader{
		InputFile: inputFile,
	}
}

func (g *GzipReader) open() error {
	var err error
	g.file, err = os.Open(g.InputFile)
	if err != nil {
		return err
	}

	g.zipreader, err = gzip.NewReader(g.file)
	if err != nil {
		g.file.Close()
		return err
	}
	return nil
}

// Read reads unzips data and puts it into buf.
// This allows you to stream the data.
func (g *GzipReader) Read(buf []byte) (int, error) {
	return g.zipreader.Read(buf)
}

func (g *GzipReader) UnzipToFile(outfile string) (int64, error) {
	err := g.open()
	if err != nil {
		return int64(0), err
	}
	defer g.file.Close()

	file, err := os.Create(outfile)
	if err != nil {
		return int64(0), err
	}
	defer file.Close()
	defer g.zipreader.Close()
	return io.Copy(file, g.zipreader)
}
