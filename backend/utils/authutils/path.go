package authutils

import (
	"os"
	"path/filepath"
	"runtime"
)

var getOS = func() string {
	return runtime.GOOS
}

var getUserHomeDir = os.UserHomeDir

// Directory constants
const (
	TellaVaultDir  = ".TellaVault"
	TellaPublicDir = "TellaPublic"
	TVaultFile     = ".tvault"
	TellaDBFile    = ".tella.db"
	TempDir        = "temp"
)

// getBasePath returns the appropriate base directory based on OS
func getBasePath() (string, error) {
	homeDir, err := getUserHomeDir()
	if err != nil {
		return "", err
	}

	// For Windows and macOS, use Documents folder
	// For Linux and others, use home directory directly
	if getOS() == "windows" || getOS() == "darwin" {
		return filepath.Join(homeDir, "Documents"), nil
	}

	return homeDir, nil
}

// buildPath constructs a path with fallback option if there's an error
func buildPath(components ...string) string {
	basePath, err := getBasePath()
	if err != nil {
		// Determine the fallback path based on the first component
		if components[0] == TellaVaultDir {
			return filepath.Join(".", TellaVaultDir, components[1])
		}
		// Create a new slice with "." as the first element
		fallbackPath := make([]string, len(components)+1)
		fallbackPath[0] = "."
		copy(fallbackPath[1:], components)
		return filepath.Join(fallbackPath...)
	}

	// Create a new slice with basePath as the first element
	fullPath := make([]string, len(components)+1)
	fullPath[0] = basePath
	copy(fullPath[1:], components)
	return filepath.Join(fullPath...)
}

func GetTVaultPath() string {
	return buildPath(TellaVaultDir, TVaultFile)
}

func GetDatabasePath() string {
	return buildPath(TellaVaultDir, TellaDBFile)
}

func GetTempDir() string {
	return buildPath(TellaPublicDir, TempDir)
}
