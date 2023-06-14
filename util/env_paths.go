package util

import (
	"os"
	"os/user"
	"path"
	"runtime"
)

const AppName = "DART"

// Paths contains paths to common directories used by DART.
// This was ported from https://github.com/sindresorhus/env-paths
// so that paths in DART3 will match paths used in prior versions
// of DART.
type Paths struct {
	appName   string
	DataDir   string
	ConfigDir string
	CacheDir  string
	HomeDir   string
	LogDir    string
	TempDir   string
}

// NewPaths returns a Paths struct appropriate to the current operating
// system. Currently supports only Windows, Mac and Linux.
func NewPaths() *Paths {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		panic(err)
	}
	envPaths := &Paths{
		appName: AppName,
		HomeDir: homeDir,
		TempDir: path.Join(os.TempDir(), AppName),
	}
	switch runtime.GOOS {
	case "darwin":
		envPaths.setMacOS()
	case "windows":
		envPaths.setWindows()
	case "linux":
		envPaths.setLinux()
	default:
		panic("OS " + runtime.GOOS + " not supported")
	}
	return envPaths
}

func (p *Paths) setMacOS() {
	library := path.Join(p.HomeDir, "Library")
	p.DataDir = path.Join(library, "Application Support", p.appName)
	p.ConfigDir = path.Join(library, "Preferences", p.appName)
	p.CacheDir = path.Join(library, "Caches", p.appName)
	p.LogDir = path.Join(library, "Logs", p.appName)
}

func (p *Paths) setWindows() {
	appData := os.Getenv("APPDATA")
	if appData == "" {
		appData = path.Join(p.HomeDir, "AppData", "Roaming")
	}
	localAppData := os.Getenv("LOCALAPPDATA")
	if localAppData == "" {
		localAppData = path.Join(p.HomeDir, "AppData", "Local")
	}
	p.DataDir = path.Join(localAppData, p.appName, "Data")
	p.ConfigDir = path.Join(appData, p.appName, "Config")
	p.CacheDir = path.Join(localAppData, p.appName, "Cache")
	p.LogDir = path.Join(localAppData, p.appName, "Log")
}

// setLinux returns Linix directory names based on
// https://specifications.freedesktop.org/basedir-spec/basedir-spec-latest.html and
// https://wiki.debian.org/XDGBaseDirectorySpecification#state
func (p *Paths) setLinux() {
	dataDir := os.Getenv("XDG_DATA_HOME")
	if dataDir == "" {
		dataDir = path.Join(p.HomeDir, ".local", "share", p.appName)
	}
	configDir := os.Getenv("XDG_CONFIG_HOME")
	if configDir == "" {
		configDir = path.Join(p.HomeDir, ".config", p.appName)
	}

	cacheDir := os.Getenv("XDG_CACHE_HOME")
	if cacheDir == "" {
		cacheDir = path.Join(p.HomeDir, ".cache", p.appName)
	}

	logDir := os.Getenv("XDG_STATE_HOME")
	if logDir == "" {
		logDir = path.Join(p.HomeDir, ".local", "state", p.appName)
	}

	user, err := user.Current()
	if err != nil {
		panic(err)
	}

	p.DataDir = dataDir
	p.ConfigDir = configDir
	p.CacheDir = cacheDir
	p.LogDir = logDir
	p.TempDir = path.Join(os.TempDir(), user.Name, AppName)
}
