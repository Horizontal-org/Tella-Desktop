package authutils

import (
	"github.com/adrg/xdg"
	"path/filepath"
)

// Directory constants
const (
	// TODO cblgh(2026-03-06): obfuscate `TellaAppName`?
	TellaAppName   = "Tella"
	TVaultFile     = ".tvault"
	TellaDBFile    = ".tella.db"
	TempDir        = "temp"
	ConfigFilename = "desktop-settings.toml"
)

// Create wrappers around XDG functions that we can mock in tests
var xdgDataFile = func(relPath string) (string, error) {
	return xdg.DataFile(relPath)
}

var xdgCacheFile = func(relPath string) (string, error) {
	return xdg.CacheFile(relPath)
}

var xdgConfigFile = func(relPath string) (string, error) {
	return xdg.ConfigFile(relPath)
}

func GetConfigFilePath() string {
	p, err := xdgConfigFile(ConfigFilename)
	if err != nil {
		return ""
	}
	return p
}

// TODO cblgh(2026-02-12): remove all "local directory" fallbacks to limit spreading data exposure across multiple directories?
// return err instead and let caller handle what to do?
func GetTVaultPath() string {
	path, err := xdgDataFile(filepath.Join(TellaAppName, TVaultFile))
	if err != nil {
		// Fallback to local directory
		return filepath.Join(".", TVaultFile)
	}
	return path
}

func GetDatabasePath() string {
	path, err := xdgDataFile(filepath.Join(TellaAppName, TellaDBFile))
	if err != nil {
		// Fallback to local directory
		return filepath.Join(".", TellaDBFile)
	}
	return path
}

func GetTempDir() string {
	tdir, err := xdgCacheFile(filepath.Join(TellaAppName, TempDir))
	if err != nil {
		// Fallback to local directory
		return filepath.Join(".", TempDir)
	}
	return tdir
}

func GetExportDir() string {
	return filepath.Join(xdg.UserDirs.Documents, "Exports")
}
