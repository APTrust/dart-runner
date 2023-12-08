package util

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

// Paths in pastbuild_test_batch.csv file are relative.
// Create a temp file with absolute paths.
func MakeTempCSVFileWithValidPaths(t *testing.T, pathToCSVFile string) string {
	tempFilePath := filepath.Join(os.TempDir(), "temp_batch.csv")
	csvContents, err := os.ReadFile(pathToCSVFile)
	require.Nil(t, err)
	absPrefix := ProjectRoot() + string(os.PathSeparator)
	csvWithAbsPaths := strings.ReplaceAll(string(csvContents), "./", absPrefix)
	require.NoError(t, os.WriteFile(tempFilePath, []byte(csvWithAbsPaths), 0666))
	return tempFilePath
}
