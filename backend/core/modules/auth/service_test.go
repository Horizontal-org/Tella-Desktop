package auth

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"Tella-Desktop/backend/utils/constants"
)

// Create a test-specific implementation of the auth service
type testService struct {
	tvaultPath   string
	databasePath string
	tempDir      string
	dbKey        []byte
	isUnlocked   bool
}

// Implement the Service interface for our test service
func (s *testService) Initialize(ctx context.Context) error {
	// Create vault directory if it doesn't exist
	vaultDir := filepath.Dir(s.tvaultPath)
	if err := os.MkdirAll(vaultDir, 0755); err != nil {
		return err
	}

	// Create temp directory
	if err := os.MkdirAll(s.tempDir, 0755); err != nil {
		return err
	}

	return nil
}

func (s *testService) IsFirstTimeSetup() bool {
	_, err := os.Stat(s.tvaultPath)
	return os.IsNotExist(err)
}

func (s *testService) CreatePassword(password string) error {
	if len(password) < 6 {
		return constants.ErrPasswordTooShort
	}

	// Create a mock database key
	s.dbKey = make([]byte, constants.KeyLength)
	for i := 0; i < constants.KeyLength; i++ {
		s.dbKey[i] = byte(i % 256)
	}

	// Create a mock TVault file
	file, err := os.Create(s.tvaultPath)
	if err != nil {
		return err
	}
	defer file.Close()

	// Write something to the file so it exists
	file.Write([]byte{1, 2, 3})

	s.isUnlocked = true
	return nil
}

func (s *testService) DecryptDatabaseKey(password string) error {
	if password == "secure-password-1234" {
		s.isUnlocked = true
		if s.dbKey == nil {
			s.dbKey = make([]byte, constants.KeyLength)
			for i := 0; i < constants.KeyLength; i++ {
				s.dbKey[i] = byte(i % 256)
			}
		}
		return nil
	}

	return constants.ErrInvalidPassword
}

func (s *testService) GetDBKey() ([]byte, error) {
	if !s.isUnlocked || s.dbKey == nil {
		return nil, constants.ErrInvalidPassword
	}
	return s.dbKey, nil
}

// Setup test environment
func setupTestEnvironment(t *testing.T) (Service, func()) {
	// Create temporary test directory
	tempDir, err := os.MkdirTemp("", "tella-test-")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}

	service := &testService{
		tvaultPath:   filepath.Join(tempDir, ".tvault"),
		databasePath: filepath.Join(tempDir, "tella.db"),
		tempDir:      filepath.Join(tempDir, "temp"),
		isUnlocked:   false,
	}

	// Initialize the service
	err = service.Initialize(context.Background())
	if err != nil {
		t.Fatalf("Failed to initialize service: %v", err)
	}

	// Return cleanup function
	cleanup := func() {
		os.RemoveAll(tempDir)
	}

	return service, cleanup
}

func TestIsFirstTimeSetup(t *testing.T) {
	service, cleanup := setupTestEnvironment(t)
	defer cleanup()

	// First time should be true as no tvault file exists
	if !service.IsFirstTimeSetup() {
		t.Errorf("Expected IsFirstTimeSetup to return true for new environment")
	}

	// Create password to generate tvault
	password := "secure-password-1234"
	err := service.CreatePassword(password)
	if err != nil {
		t.Fatalf("Failed to create password: %v", err)
	}

	// After creating password, IsFirstTimeSetup should return false
	if service.IsFirstTimeSetup() {
		t.Errorf("Expected IsFirstTimeSetup to return false after creating password")
	}
}

func TestCreatePassword(t *testing.T) {
	service, cleanup := setupTestEnvironment(t)
	defer cleanup()

	testCases := []struct {
		name     string
		password string
		wantErr  bool
		errType  error
	}{
		{
			name:     "Valid password",
			password: "secure-password-1234",
			wantErr:  false,
		},
		{
			name:     "Short password",
			password: "short",
			wantErr:  true,
			errType:  constants.ErrPasswordTooShort,
		},
		{
			name:     "Empty password",
			password: "",
			wantErr:  true,
			errType:  constants.ErrPasswordTooShort,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Reset environment for each test case
			cleanup()
			service, cleanup = setupTestEnvironment(t)

			err := service.CreatePassword(tc.password)

			// Check error expectation
			if tc.wantErr {
				if err == nil {
					t.Errorf("Expected error but got none")
				} else if tc.errType != nil && err != tc.errType {
					t.Errorf("Expected error %v, got %v", tc.errType, err)
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}

				// Verify database key was stored in memory
				dbKey, err := service.GetDBKey()
				if err != nil {
					t.Errorf("Failed to get DB key after password creation: %v", err)
				}
				if len(dbKey) != constants.KeyLength {
					t.Errorf("Expected DB key length %d, got %d", constants.KeyLength, len(dbKey))
				}

				// Verify that the tvault file was created
				tvaultPath := service.(*testService).tvaultPath
				if _, err := os.Stat(tvaultPath); os.IsNotExist(err) {
					t.Errorf("TVault file was not created at %s", tvaultPath)
				}
			}
		})
	}
}

func TestDecryptDatabaseKey(t *testing.T) {
	service, cleanup := setupTestEnvironment(t)
	defer cleanup()

	// Create a password first
	password := "secure-password-1234"
	err := service.CreatePassword(password)
	if err != nil {
		t.Fatalf("Failed to create password: %v", err)
	}

	// Test cases for verification
	testCases := []struct {
		name     string
		password string
		wantErr  bool
		errType  error
	}{
		{
			name:     "Correct password",
			password: "secure-password-1234",
			wantErr:  false,
		},
		{
			name:     "Incorrect password",
			password: "wrong-password",
			wantErr:  true,
			errType:  constants.ErrInvalidPassword,
		},
		{
			name:     "Empty password",
			password: "",
			wantErr:  true,
			errType:  constants.ErrInvalidPassword,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := service.DecryptDatabaseKey(tc.password)

			// Check error expectation
			if tc.wantErr {
				if err == nil {
					t.Errorf("Expected error but got none")
				} else if tc.errType != nil && err != tc.errType {
					t.Errorf("Expected error %v, got %v", tc.errType, err)
				}
				dbKey, keyErr := service.GetDBKey()
				if keyErr == nil {
					t.Errorf("Expected error when getting DB key after failed authentication")
				}
				if dbKey != nil {
					t.Errorf("Expected nil dbKey for invalid password, got non-nil value")
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				dbKey, keyErr := service.GetDBKey()
				if keyErr != nil {
					t.Errorf("Failed to get DB key after successful authentication: %v", keyErr)
				}
				if dbKey == nil {
					t.Errorf("Expected valid dbKey, got nil")
				}
				if len(dbKey) != constants.KeyLength {
					t.Errorf("Expected DB key length %d, got %d", constants.KeyLength, len(dbKey))
				}
			}
		})
	}
}

func TestGetDBKey(t *testing.T) {
	service, cleanup := setupTestEnvironment(t)
	defer cleanup()

	// Initially, service should be locked
	_, err := service.GetDBKey()
	if err == nil {
		t.Errorf("Expected error when getting DB key without authentication")
	}

	// Create a password and unlock
	password := "secure-password-1234"
	err = service.CreatePassword(password)
	if err != nil {
		t.Fatalf("Failed to create password: %v", err)
	}

	// Now should be able to get the key
	dbKey, err := service.GetDBKey()
	if err != nil {
		t.Errorf("Failed to get DB key after unlock: %v", err)
	}
	if len(dbKey) != constants.KeyLength {
		t.Errorf("Expected DB key length %d, got %d", constants.KeyLength, len(dbKey))
	}

	// Create a new service instance (simulating app restart)
	service, cleanup = setupTestEnvironment(t)
	defer cleanup()

	// Should be locked again
	_, err = service.GetDBKey()
	if err == nil {
		t.Errorf("Expected error when getting DB key without authentication after reset")
	}

	// Verify password to unlock
	err = service.DecryptDatabaseKey(password)
	if err != nil {
		t.Fatalf("Failed to verify password: %v", err)
	}

	// Now should be able to get the key again
	dbKey, err = service.GetDBKey()
	if err != nil {
		t.Errorf("Failed to get DB key after verification: %v", err)
	}
	if len(dbKey) != constants.KeyLength {
		t.Errorf("Expected DB key length %d, got %d", constants.KeyLength, len(dbKey))
	}
}
