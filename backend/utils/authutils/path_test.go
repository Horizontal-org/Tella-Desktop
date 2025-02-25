package authutils

import (
	"os"
	"path/filepath"
	"testing"
)

func TestGetBasePath(t *testing.T) {
	// Save original functions
	originalGetOS := getOS
	originalGetUserHomeDir := getUserHomeDir
	defer func() {
		getOS = originalGetOS
		getUserHomeDir = originalGetUserHomeDir
	}()

	// Mock home directory
	const mockHomeDir = "/mock/home"
	getUserHomeDir = func() (string, error) {
		return mockHomeDir, nil
	}

	// Test cases
	tests := []struct {
		name     string
		mockOS   string
		expected string
	}{
		{
			name:     "Windows",
			mockOS:   "windows",
			expected: filepath.Join(mockHomeDir, "Documents"),
		},
		{
			name:     "macOS",
			mockOS:   "darwin",
			expected: filepath.Join(mockHomeDir, "Documents"),
		},
		{
			name:     "Linux",
			mockOS:   "linux",
			expected: mockHomeDir,
		},
		{
			name:     "Other OS",
			mockOS:   "freebsd",
			expected: mockHomeDir,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set mock OS
			getOS = func() string {
				return tt.mockOS
			}

			result, err := getBasePath()
			if err != nil {
				t.Fatalf("getBasePath() returned unexpected error: %v", err)
			}
			if result != tt.expected {
				t.Errorf("getBasePath() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestGetBasePathError(t *testing.T) {
	// Save original functions
	originalGetUserHomeDir := getUserHomeDir
	defer func() {
		getUserHomeDir = originalGetUserHomeDir
	}()

	// Mock home directory to fail
	getUserHomeDir = func() (string, error) {
		return "", os.ErrNotExist
	}

	// Test error case
	_, err := getBasePath()
	if err == nil {
		t.Error("getBasePath() expected to return error when getUserHomeDir fails")
	}
}

func TestBuildPath(t *testing.T) {
	// Save original functions
	originalGetOS := getOS
	originalGetUserHomeDir := getUserHomeDir
	defer func() {
		getOS = originalGetOS
		getUserHomeDir = originalGetUserHomeDir
	}()

	// Mock home directory
	const mockHomeDir = "/mock/home"
	getUserHomeDir = func() (string, error) {
		return mockHomeDir, nil
	}

	// Test cases for successful path building
	tests := []struct {
		name       string
		mockOS     string
		components []string
		expected   string
	}{
		{
			name:       "Windows - TellaVault Directory",
			mockOS:     "windows",
			components: []string{TellaVaultDir, TVaultFile},
			expected:   filepath.Join(mockHomeDir, "Documents", TellaVaultDir, TVaultFile),
		},
		{
			name:       "macOS - TellaVault Directory",
			mockOS:     "darwin",
			components: []string{TellaVaultDir, TVaultFile},
			expected:   filepath.Join(mockHomeDir, "Documents", TellaVaultDir, TVaultFile),
		},
		{
			name:       "Linux - TellaVault Directory",
			mockOS:     "linux",
			components: []string{TellaVaultDir, TVaultFile},
			expected:   filepath.Join(mockHomeDir, TellaVaultDir, TVaultFile),
		},
		{
			name:       "Windows - Public Directory",
			mockOS:     "windows",
			components: []string{TellaPublicDir, TempDir},
			expected:   filepath.Join(mockHomeDir, "Documents", TellaPublicDir, TempDir),
		},
		{
			name:       "macOS - Public Directory",
			mockOS:     "darwin",
			components: []string{TellaPublicDir, TempDir},
			expected:   filepath.Join(mockHomeDir, "Documents", TellaPublicDir, TempDir),
		},
		{
			name:       "Linux - Public Directory",
			mockOS:     "linux",
			components: []string{TellaPublicDir, TempDir},
			expected:   filepath.Join(mockHomeDir, TellaPublicDir, TempDir),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set mock OS
			getOS = func() string {
				return tt.mockOS
			}

			result := buildPath(tt.components...)
			if result != tt.expected {
				t.Errorf("buildPath(%v) = %v, want %v", tt.components, result, tt.expected)
			}
		})
	}
}

func TestBuildPathFallback(t *testing.T) {
	// Save original functions
	originalGetUserHomeDir := getUserHomeDir
	defer func() {
		getUserHomeDir = originalGetUserHomeDir
	}()

	// Mock home directory to fail
	getUserHomeDir = func() (string, error) {
		return "", os.ErrNotExist
	}

	// Test fallback cases
	tests := []struct {
		name       string
		components []string
		expected   string
	}{
		{
			name:       "TellaVault fallback",
			components: []string{TellaVaultDir, TVaultFile},
			expected:   filepath.Join(".", TellaVaultDir, TVaultFile),
		},
		{
			name:       "TellaPublic fallback",
			components: []string{TellaPublicDir, TempDir},
			expected:   filepath.Join(".", TellaPublicDir, TempDir),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := buildPath(tt.components...)
			if result != tt.expected {
				t.Errorf("buildPath(%v) with fallback = %v, want %v", tt.components, result, tt.expected)
			}
		})
	}
}

func TestGetTVaultPath(t *testing.T) {
	// Save original functions
	originalGetOS := getOS
	originalGetUserHomeDir := getUserHomeDir
	defer func() {
		getOS = originalGetOS
		getUserHomeDir = originalGetUserHomeDir
	}()

	// Mock home directory
	const mockHomeDir = "/mock/home"
	getUserHomeDir = func() (string, error) {
		return mockHomeDir, nil
	}

	// Test for different OS
	tests := []struct {
		name     string
		mockOS   string
		expected string
	}{
		{
			name:     "Windows",
			mockOS:   "windows",
			expected: filepath.Join(mockHomeDir, "Documents", TellaVaultDir, TVaultFile),
		},
		{
			name:     "macOS",
			mockOS:   "darwin",
			expected: filepath.Join(mockHomeDir, "Documents", TellaVaultDir, TVaultFile),
		},
		{
			name:     "Linux",
			mockOS:   "linux",
			expected: filepath.Join(mockHomeDir, TellaVaultDir, TVaultFile),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set mock OS
			getOS = func() string {
				return tt.mockOS
			}

			result := GetTVaultPath()
			if result != tt.expected {
				t.Errorf("GetTVaultPath() = %v, want %v", result, tt.expected)
			}
		})
	}

	// Test fallback path
	t.Run("Fallback path", func(t *testing.T) {
		getUserHomeDir = func() (string, error) {
			return "", os.ErrNotExist
		}
		expected := filepath.Join(".", TellaVaultDir, TVaultFile)
		result := GetTVaultPath()
		if result != expected {
			t.Errorf("GetTVaultPath() with fallback = %v, want %v", result, expected)
		}
	})
}

func TestGetDatabasePath(t *testing.T) {
	// Save original functions
	originalGetOS := getOS
	originalGetUserHomeDir := getUserHomeDir
	defer func() {
		getOS = originalGetOS
		getUserHomeDir = originalGetUserHomeDir
	}()

	// Mock home directory
	const mockHomeDir = "/mock/home"
	getUserHomeDir = func() (string, error) {
		return mockHomeDir, nil
	}

	// Test for different OS
	tests := []struct {
		name     string
		mockOS   string
		expected string
	}{
		{
			name:     "Windows",
			mockOS:   "windows",
			expected: filepath.Join(mockHomeDir, "Documents", TellaVaultDir, TellaDBFile),
		},
		{
			name:     "macOS",
			mockOS:   "darwin",
			expected: filepath.Join(mockHomeDir, "Documents", TellaVaultDir, TellaDBFile),
		},
		{
			name:     "Linux",
			mockOS:   "linux",
			expected: filepath.Join(mockHomeDir, TellaVaultDir, TellaDBFile),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set mock OS
			getOS = func() string {
				return tt.mockOS
			}

			result := GetDatabasePath()
			if result != tt.expected {
				t.Errorf("GetDatabasePath() = %v, want %v", result, tt.expected)
			}
		})
	}

	// Test fallback path
	t.Run("Fallback path", func(t *testing.T) {
		getUserHomeDir = func() (string, error) {
			return "", os.ErrNotExist
		}
		expected := filepath.Join(".", TellaVaultDir, TellaDBFile)
		result := GetDatabasePath()
		if result != expected {
			t.Errorf("GetDatabasePath() with fallback = %v, want %v", result, expected)
		}
	})
}

func TestGetTempDir(t *testing.T) {
	// Save original functions
	originalGetOS := getOS
	originalGetUserHomeDir := getUserHomeDir
	defer func() {
		getOS = originalGetOS
		getUserHomeDir = originalGetUserHomeDir
	}()

	// Mock home directory
	const mockHomeDir = "/mock/home"
	getUserHomeDir = func() (string, error) {
		return mockHomeDir, nil
	}

	// Test for different OS
	tests := []struct {
		name     string
		mockOS   string
		expected string
	}{
		{
			name:     "Windows",
			mockOS:   "windows",
			expected: filepath.Join(mockHomeDir, "Documents", TellaPublicDir, TempDir),
		},
		{
			name:     "macOS",
			mockOS:   "darwin",
			expected: filepath.Join(mockHomeDir, "Documents", TellaPublicDir, TempDir),
		},
		{
			name:     "Linux",
			mockOS:   "linux",
			expected: filepath.Join(mockHomeDir, TellaPublicDir, TempDir),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set mock OS
			getOS = func() string {
				return tt.mockOS
			}

			result := GetTempDir()
			if result != tt.expected {
				t.Errorf("GetTempDir() = %v, want %v", result, tt.expected)
			}
		})
	}

	// Test fallback path
	t.Run("Fallback path", func(t *testing.T) {
		getUserHomeDir = func() (string, error) {
			return "", os.ErrNotExist
		}
		expected := filepath.Join(".", TellaPublicDir, TempDir)
		result := GetTempDir()
		if result != expected {
			t.Errorf("GetTempDir() with fallback = %v, want %v", result, expected)
		}
	})
}
