package authutils

import (
	"os"
	"path/filepath"
)

// todo: we'll need to modify this to handle paths on linux and windows as well
func GetTVaultPath() string {
	homedir, err := os.UserHomeDir()
	if err != nil {
		return "./.TellaVault/.tvault"
	}

	return filepath.Join(homedir, "Documents", ".TellaVault", ".tvault")
}

func GetDatabasePath() string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "./.TellaVault/.tella.db"
	}
	return filepath.Join(homeDir, "Documents", ".TellaVault", ".tella.db")
}

func GetTempDir() string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "./TellaPublic/temp"
	}
	return filepath.Join(homeDir, "Documents", "TellaPublic", "temp")
}
