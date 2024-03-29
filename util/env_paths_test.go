package util_test

import (
	"os"
	"regexp"
	"runtime"
	"testing"

	"github.com/APTrust/dart-runner/util"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPaths(t *testing.T) {
	p := util.NewPaths()
	require.NotNil(t, p)

	homeDir, _ := os.UserHomeDir()
	assert.Contains(t, p.DataDir, homeDir)
	assert.Contains(t, p.Config, homeDir)
	assert.Contains(t, p.Cache, homeDir)
	assert.Contains(t, p.Home, homeDir)
	assert.Contains(t, p.LogDir, homeDir)
	assert.Contains(t, p.TempDir, util.AppName)

	switch runtime.GOOS {
	case "darwin":
		testMacOsPaths(t, p)
	case "windows":
		testWindowsPaths(t, p)
	case "linux":
		testLinuxPaths(t, p)
	}
}

func TestLogFile(t *testing.T) {
	p := util.NewPaths()
	require.NotNil(t, p)
	logFile, err := p.LogFile()
	require.Nil(t, err)
	assert.Contains(t, logFile, p.LogDir)

	pattern := regexp.MustCompile(`dart(\d)*.log$`)
	assert.True(t, pattern.MatchString(logFile))
}

func testMacOsPaths(t *testing.T, p *util.Paths) {
	assert.Contains(t, p.DataDir, "Library/Application Support")
	assert.Contains(t, p.Config, "Library/Preferences")
	assert.Contains(t, p.Cache, "Library/Caches")
	assert.Contains(t, p.LogDir, "Library/Logs")
}

func testWindowsPaths(t *testing.T, p *util.Paths) {
	assert.Contains(t, p.DataDir, "Data")
	assert.Contains(t, p.Config, "Config")
	assert.Contains(t, p.Cache, "Cache")
	assert.Contains(t, p.LogDir, "Log")
}

func testLinuxPaths(t *testing.T, p *util.Paths) {
	assert.Contains(t, p.DataDir, util.AppName)
	assert.Contains(t, p.Config, util.AppName)
	assert.Contains(t, p.Cache, util.AppName)
	assert.Contains(t, p.LogDir, util.AppName)
}

func TestDefaultPaths(t *testing.T) {
	p := util.NewPaths()
	defaultPaths, err := p.DefaultPaths()
	require.Nil(t, err)
	require.NotEmpty(t, defaultPaths)
	assert.True(t, len(defaultPaths) > 4)
}
