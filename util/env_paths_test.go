package util_test

import (
	"os"
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
	assert.Contains(t, p.ConfigDir, homeDir)
	assert.Contains(t, p.CacheDir, homeDir)
	assert.Contains(t, p.HomeDir, homeDir)
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

func testMacOsPaths(t *testing.T, p *util.Paths) {
	assert.Contains(t, p.DataDir, "Library/Application Support")
	assert.Contains(t, p.ConfigDir, "Library/Preferences")
	assert.Contains(t, p.CacheDir, "Library/Caches")
	assert.Contains(t, p.LogDir, "Library/Logs")
}

func testWindowsPaths(t *testing.T, p *util.Paths) {
	assert.Contains(t, p.DataDir, "Data")
	assert.Contains(t, p.ConfigDir, "Config")
	assert.Contains(t, p.CacheDir, "Cache")
	assert.Contains(t, p.LogDir, "Log")
}

func testLinuxPaths(t *testing.T, p *util.Paths) {
	assert.Contains(t, p.DataDir, util.AppName)
	assert.Contains(t, p.ConfigDir, util.AppName)
	assert.Contains(t, p.CacheDir, util.AppName)
	assert.Contains(t, p.LogDir, util.AppName)
}
