package util

import (
	"io/fs"
	"io/ioutil"
	"os"
	"os/user"
	"path/filepath"
	"runtime"
	"sort"
	"time"
)

const AppName = "DART"

// Paths contains paths to common directories used by DART.
// This was ported from https://github.com/sindresorhus/env-paths
// so that paths in DART3 will match paths used in prior versions
// of DART.
type Paths struct {
	appName   string
	Cache     string
	Config    string
	DataDir   string
	Desktop   string
	Documents string
	Downloads string
	Music     string
	Photos    string
	Videos    string
	Home      string
	Public    string
	LogDir    string
	TempDir   string
	Root      string
	UserMount string
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
		Home:    homeDir,
		TempDir: filepath.Join(os.TempDir(), AppName),
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
	library := filepath.Join(p.Home, "Library")
	p.DataDir = filepath.Join(library, "Application Support", p.appName)
	p.Config = filepath.Join(library, "Preferences", p.appName)
	p.Cache = filepath.Join(library, "Caches", p.appName)
	p.LogDir = filepath.Join(library, "Logs", p.appName)

	p.Desktop = filepath.Join(p.Home, "Desktop")
	p.Documents = filepath.Join(p.Home, "Documents")
	p.Downloads = filepath.Join(p.Home, "Downloads")
	p.Music = filepath.Join(p.Home, "Music")
	p.Photos = filepath.Join(p.Home, "Pictures")
	p.Public = filepath.Join(p.Home, "Public")
	p.Videos = filepath.Join(p.Home, "Movies")
	p.Root = "/"
	p.UserMount = "/Volumes"
}

func (p *Paths) setWindows() {
	appData := os.Getenv("APPDATA")
	if appData == "" {
		appData = filepath.Join(p.Home, "AppData", "Roaming")
	}
	localAppData := os.Getenv("LOCALAPPDATA")
	if localAppData == "" {
		localAppData = filepath.Join(p.Home, "AppData", "Local")
	}
	p.DataDir = filepath.Join(localAppData, p.appName, "Data")
	p.Config = filepath.Join(appData, p.appName, "Config")
	p.Cache = filepath.Join(localAppData, p.appName, "Cache")
	p.LogDir = filepath.Join(localAppData, p.appName, "Log")

	p.Desktop = filepath.Join(p.Home, "Desktop")
	p.Documents = filepath.Join(p.Home, "Documents")
	p.Downloads = filepath.Join(p.Home, "Downloads")
	p.Music = filepath.Join(p.Home, "Music")
	p.Photos = filepath.Join(p.Home, "Pictures")
	p.Public = filepath.Join(p.Home, "c:\\Users\\Public")
	p.Videos = filepath.Join(p.Home, "Videos")

	p.Root = "c:\\"
	p.UserMount = ""
}

// setLinux returns Linix directory names based on
// https://specifications.freedesktop.org/basedir-spec/basedir-spec-latest.html and
// https://wiki.debian.org/XDGBaseDirectorySpecification#state
func (p *Paths) setLinux() {
	dataDir := os.Getenv("XDG_DATA_HOME")
	if dataDir == "" {
		dataDir = filepath.Join(p.Home, ".local", "share", p.appName)
	}
	configDir := os.Getenv("XDG_CONFIG_HOME")
	if configDir == "" {
		configDir = filepath.Join(p.Home, ".config", p.appName)
	}

	cacheDir := os.Getenv("XDG_CACHE_HOME")
	if cacheDir == "" {
		cacheDir = filepath.Join(p.Home, ".cache", p.appName)
	}

	logDir := os.Getenv("XDG_STATE_HOME")
	if logDir == "" {
		logDir = filepath.Join(p.Home, ".local", "state", p.appName)
	}

	user, err := user.Current()
	if err != nil {
		panic(err)
	}

	p.DataDir = dataDir
	p.Config = configDir
	p.Cache = cacheDir
	p.LogDir = logDir
	p.TempDir = filepath.Join(os.TempDir(), user.Name, AppName)

	p.Desktop = filepath.Join(p.Home, "Desktop")
	p.Documents = filepath.Join(p.Home, "Documents")
	p.Downloads = filepath.Join(p.Home, "Downloads")
	p.Music = filepath.Join(p.Home, "Music")
	p.Photos = filepath.Join(p.Home, "Pictures")
	p.Public = filepath.Join(p.Home, "Public")
	p.Videos = filepath.Join(p.Home, "Videos")
	p.Root = "/"
	p.UserMount = filepath.Join("/", "media", user.Name)
}

func (p *Paths) LogFile() (string, error) {
	files, err := ioutil.ReadDir(p.LogDir)
	if err != nil {
		return "", err
	}
	lastMod := time.Time{}
	var lastLog fs.FileInfo
	for _, f := range files {
		if f.ModTime().After(lastMod) {
			lastMod = f.ModTime()
			lastLog = f
		}
	}
	return filepath.Join(p.LogDir, lastLog.Name()), nil
}

// DefaultPaths returns a list of ExtendedFileInfo objects describing
// which directories we should should by default in our file browser.
func (p *Paths) DefaultPaths() ([]*ExtendedFileInfo, error) {
	exFileInfo := make([]*ExtendedFileInfo, 0)
	paths := []string{
		p.Root,
		p.Desktop,
		p.Documents,
		p.Downloads,
		p.Home,
		p.Music,
		p.Photos,
		p.Public,
	}
	if runtime.GOOS == "windows" {
		paths = append(paths, GetWindowsDrives()...)
	} else {
		mountedDirs, err := ListDirectory(p.UserMount)
		if err != nil {
			return nil, err
		}
		for _, dir := range mountedDirs {
			if dir.IsDir() {
				exFileInfo = append(exFileInfo, dir)
			}
		}
	}
	for _, _path := range paths {
		fstat, err := os.Stat(_path)
		if err != nil {
			// Need to log error, but we'll have a circular referece to core. :(
		} else {
			exFileInfo = append(exFileInfo, NewExtendedFileInfo(_path, fstat))
		}
	}
	sort.Slice(exFileInfo, func(i, j int) bool {
		return exFileInfo[i].FullPath < exFileInfo[j].FullPath
	})
	return exFileInfo, nil
}
