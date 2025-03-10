package authutils

import (
	"os"
	"path/filepath"

	"github.com/adrg/xdg"
)

// Directory constants
const (
	TellaAppName = "Tella"
	TVaultFile   = ".tvault"
	TellaDBFile  = ".tella.db"
	TempDir      = "temp"
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
	path, err := xdgCacheFile(filepath.Join(TellaAppName, TempDir, "placeholder"))
	if err != nil {
		// Fallback to local directory
		return filepath.Join(".", TempDir)
	}

	// Return just the directory part
	dir := filepath.Dir(path)

	// Ensure the directory exists
	if err := os.MkdirAll(dir, 0755); err != nil {
		return filepath.Join(".", TempDir)
	}

	return dir
}
