package authutils

import (
	"os"
	"path/filepath"
	"testing"
)

func TestGetTVaultPath(t *testing.T) {
	// Save original wrapper function
	originalXdgDataFile := xdgDataFile
	defer func() {
		xdgDataFile = originalXdgDataFile
	}()

	// Mock XDG DataFile wrapper
	xdgDataFile = func(path string) (string, error) {
		return filepath.Join("/mock/xdg/data", path), nil
	}

	expected := filepath.Join("/mock/xdg/data", TellaAppName, TVaultFile)
	result := GetTVaultPath()
	if result != expected {
		t.Errorf("GetTVaultPath() = %v, want %v", result, expected)
	}
}

func TestGetDatabasePath(t *testing.T) {
	originalXdgDataFile := xdgDataFile
	defer func() {
		xdgDataFile = originalXdgDataFile
	}()

	// Mock XDG DataFile wrapper
	xdgDataFile = func(path string) (string, error) {
		return filepath.Join("/mock/xdg/data", path), nil
	}

	expected := filepath.Join("/mock/xdg/data", TellaAppName, TellaDBFile)
	result := GetDatabasePath()
	if result != expected {
		t.Errorf("GetDatabasePath() = %v, want %v", result, expected)
	}
}

func TestGetTempDir(t *testing.T) {
	originalXdgCacheFile := xdgCacheFile
	defer func() {
		xdgCacheFile = originalXdgCacheFile
	}()

	// Mock XDG CacheFile wrapper
	xdgCacheFile = func(path string) (string, error) {
		return filepath.Join("/mock/xdg/cache", path), nil
	}

	expected := filepath.Join("/mock/xdg/cache", TellaAppName, TempDir)
	result := GetTempDir()

	if result != filepath.Dir(expected) {
		t.Errorf("GetTempDir() = %v, want %v", result, filepath.Dir(expected))
	}
}

func TestXDGFallback(t *testing.T) {
	originalXdgDataFile := xdgDataFile
	originalXdgCacheFile := xdgCacheFile
	originalXdgConfigFile := xdgConfigFile
	defer func() {
		xdgDataFile = originalXdgDataFile
		xdgCacheFile = originalXdgCacheFile
		xdgConfigFile = originalXdgConfigFile
	}()

	// Mock wrapper functions to return errors
	xdgDataFile = func(path string) (string, error) {
		return "", os.ErrNotExist
	}
	xdgCacheFile = func(path string) (string, error) {
		return "", os.ErrNotExist
	}
	xdgConfigFile = func(path string) (string, error) {
		return "", os.ErrNotExist
	}

	// Test fallbacks
	tests := []struct {
		name     string
		function func() string
		expected string
	}{
		{
			name:     "TVault fallback",
			function: GetTVaultPath,
			expected: filepath.Join(".", TVaultFile),
		},
		{
			name:     "Database fallback",
			function: GetDatabasePath,
			expected: filepath.Join(".", TellaDBFile),
		},
		{
			name:     "TempDir fallback",
			function: GetTempDir,
			expected: filepath.Join(".", TempDir),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.function()
			if result != tt.expected {
				t.Errorf("%s() = %v, want %v", tt.name, result, tt.expected)
			}
		})
	}
}
