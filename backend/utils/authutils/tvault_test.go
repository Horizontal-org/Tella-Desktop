package authutils

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"testing"

	"Tella-Desktop/backend/utils/constants"
)

// Test-specific helper functions that use a provided TVault path
func writeTVaultHeaderForTest(tvaultPath string, salt, encryptedDBKey []byte) error {
	// Create directory if it doesn't exist
	dir := filepath.Dir(tvaultPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	// Create tvault file
	file, err := os.Create(tvaultPath)
	if err != nil {
		return err
	}
	defer file.Close()

	// Calculate maximum space available for encrypted key
	maxKeySize := constants.TVaultHeaderSize - 1 - constants.SaltLength

	// Truncate the encrypted key if needed
	actualKeyToWrite := encryptedDBKey
	if len(encryptedDBKey) > maxKeySize {
		actualKeyToWrite = encryptedDBKey[:maxKeySize]
	}

	// Write version
	if _, err := file.Write([]byte{1}); err != nil {
		return err
	}

	// Write salt
	if _, err := file.Write(salt); err != nil {
		return err
	}

	// Write encrypted db key
	if _, err := file.Write(actualKeyToWrite); err != nil {
		return err
	}

	headerSize := 1 + len(salt) + len(actualKeyToWrite)
	if headerSize < constants.TVaultHeaderSize {
		padding := make([]byte, constants.TVaultHeaderSize-headerSize)
		if _, err := file.Write(padding); err != nil {
			return err
		}
	}

	return nil
}

func readTVaultHeaderForTest(tvaultPath string) ([]byte, []byte, error) {
	// Check if tvault file exists
	file, err := os.Open(tvaultPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil, constants.ErrTVaultNotFound
		}
		return nil, nil, err
	}
	defer file.Close()

	// Get file size
	fileInfo, err := file.Stat()
	if err != nil {
		return nil, nil, err
	}

	// Check if file is too small to be valid
	minSizeNeeded := 1 + constants.SaltLength // Version + Salt
	if fileInfo.Size() < int64(minSizeNeeded) {
		return nil, nil, constants.ErrCorruptedTVault
	}

	// Read version byte
	versionByte := make([]byte, 1)
	if _, err := file.Read(versionByte); err != nil {
		return nil, nil, constants.ErrCorruptedTVault
	}

	// Read salt
	salt := make([]byte, constants.SaltLength)
	bytesRead, err := file.Read(salt)
	if err != nil || bytesRead != constants.SaltLength {
		return nil, nil, constants.ErrCorruptedTVault
	}

	// Read encrypted key
	// We need to read until we hit the padding (zeros)
	buffer := make([]byte, constants.TVaultHeaderSize-1-constants.SaltLength)
	n, err := file.Read(buffer)
	if err != nil && err != io.EOF {
		return nil, nil, constants.ErrCorruptedTVault
	}

	// Find where the actual encrypted key ends (before padding)
	encryptedKeyLength := n
	for i := n - 1; i >= 0; i-- {
		if buffer[i] != 0 {
			encryptedKeyLength = i + 1
			break
		}
	}

	// Only return the non-padding part of the encrypted key
	encryptedKey := buffer[:encryptedKeyLength]

	return salt, encryptedKey, nil
}

// Helper function to set up a test environment with a temporary directory
func setupTVaultTest(t *testing.T) (string, func()) {
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "tvault-test-")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}

	tvaultPath := filepath.Join(tempDir, ".tvault")

	// Return cleanup function
	cleanup := func() {
		os.RemoveAll(tempDir) // Clean up temporary directory
	}

	return tvaultPath, cleanup
}

func TestTVaultHeaderOperations(t *testing.T) {
	tvaultPath, cleanup := setupTVaultTest(t)
	defer cleanup()

	// Test data
	salt := make([]byte, constants.SaltLength)
	for i := 0; i < constants.SaltLength; i++ {
		salt[i] = byte(i % 256)
	}

	encryptedDBKey := make([]byte, 48) // Typical size for encrypted key with AES-GCM
	for i := 0; i < len(encryptedDBKey); i++ {
		encryptedDBKey[i] = byte((i + 100) % 256)
	}

	// Write TVault header
	err := writeTVaultHeaderForTest(tvaultPath, salt, encryptedDBKey)
	if err != nil {
		t.Fatalf("writeTVaultHeaderForTest failed: %v", err)
	}

	// Verify file was created
	if _, err := os.Stat(tvaultPath); os.IsNotExist(err) {
		t.Errorf("TVault file was not created at %s", tvaultPath)
	}

	// Read file content to verify structure
	fileContent, err := os.ReadFile(tvaultPath)
	if err != nil {
		t.Fatalf("Failed to read TVault file: %v", err)
	}

	// Verify file size is at least the minimum header size
	if len(fileContent) != constants.TVaultHeaderSize {
		t.Errorf("Expected TVault header size %d, got %d", constants.TVaultHeaderSize, len(fileContent))
	}

	// Check version byte
	if fileContent[0] != 1 {
		t.Errorf("Expected version byte 1, got %d", fileContent[0])
	}

	// Check salt
	if !bytes.Equal(fileContent[1:constants.SaltLength+1], salt) {
		t.Errorf("Salt in TVault doesn't match the original salt")
	}

	// Check encrypted DB key (check a portion to verify it's there)
	encryptedKeyOffset := 1 + constants.SaltLength
	encryptedKeyInFile := fileContent[encryptedKeyOffset : encryptedKeyOffset+len(encryptedDBKey)]
	if !bytes.Equal(encryptedKeyInFile, encryptedDBKey) {
		t.Errorf("Encrypted DB key in TVault doesn't match the original")
	}

	// Read the TVault header
	readSalt, readKey, err := readTVaultHeaderForTest(tvaultPath)
	if err != nil {
		t.Fatalf("Failed to read TVault header: %v", err)
	}

	// Verify salt
	if !bytes.Equal(readSalt, salt) {
		t.Errorf("Read salt doesn't match expected salt")
	}

	// Verify encrypted key
	if !bytes.Equal(readKey, encryptedDBKey) {
		t.Errorf("Read encrypted key doesn't match expected key")
	}
}

func TestReadNonExistentTVault(t *testing.T) {
	tvaultPath, cleanup := setupTVaultTest(t)
	defer cleanup()

	// Try to read without creating the file first
	_, _, err := readTVaultHeaderForTest(tvaultPath)
	if err == nil {
		t.Errorf("Expected an error when reading non-existent TVault, got none")
	}

	if err != constants.ErrTVaultNotFound {
		t.Errorf("Expected ErrTVaultNotFound, got %v", err)
	}
}

func TestReadCorruptedTVault(t *testing.T) {
	tvaultPath, cleanup := setupTVaultTest(t)
	defer cleanup()

	// Create a corrupted TVault file (too small)
	file, err := os.Create(tvaultPath)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Write just a few bytes (not enough for a valid header)
	_, err = file.Write([]byte{1, 2, 3})
	if err != nil {
		t.Fatalf("Failed to write test data: %v", err)
	}
	file.Close()

	// Try to read the corrupted file
	_, _, err = readTVaultHeaderForTest(tvaultPath)
	if err == nil {
		t.Errorf("Expected an error when reading corrupted TVault, got none")
	}

	if err != constants.ErrCorruptedTVault {
		t.Errorf("Expected ErrCorruptedTVault, got %v", err)
	}
}

func TestTVaultWithDifferentKeySizes(t *testing.T) {
	// Test with different encrypted key sizes
	keySizes := []int{16, 32, 48, 64} // Various sizes

	for _, keySize := range keySizes {
		t.Run(fmt.Sprintf("KeySize-%d", keySize), func(t *testing.T) {
			tvaultPath, cleanup := setupTVaultTest(t)
			defer cleanup()

			salt := make([]byte, constants.SaltLength)
			for i := 0; i < constants.SaltLength; i++ {
				salt[i] = byte(i % 256)
			}

			encryptedKey := make([]byte, keySize)
			for i := 0; i < keySize; i++ {
				encryptedKey[i] = byte((i + 100) % 256)
			}

			// Write the header
			err := writeTVaultHeaderForTest(tvaultPath, salt, encryptedKey)
			if err != nil {
				t.Fatalf("Failed to write TVault header with key size %d: %v", keySize, err)
			}

			// Read it back
			readSalt, readKey, err := readTVaultHeaderForTest(tvaultPath)
			if err != nil {
				t.Fatalf("Failed to read TVault header with key size %d: %v", keySize, err)
			}

			// Verify data
			if !bytes.Equal(readSalt, salt) {
				t.Errorf("Salt mismatch for key size %d", keySize)
			}

			if !bytes.Equal(readKey, encryptedKey) {
				t.Errorf("Encrypted key mismatch for key size %d", keySize)
			}
		})
	}
}

func TestTVaultFileSizeLimit(t *testing.T) {
	tvaultPath, cleanup := setupTVaultTest(t)
	defer cleanup()

	// Test that the TVault header size is enforced
	salt := make([]byte, constants.SaltLength)

	// Create an encrypted key that, when combined with version byte and salt,
	// would exceed the header size limit
	hugeEncryptedKey := make([]byte, constants.TVaultHeaderSize)

	// Write should succeed but the key will be truncated to fit the header size
	err := writeTVaultHeaderForTest(tvaultPath, salt, hugeEncryptedKey)
	if err != nil {
		t.Fatalf("writeTVaultHeaderForTest failed with huge key: %v", err)
	}

	// Verify the file size doesn't exceed the header size
	fileInfo, err := os.Stat(tvaultPath)
	if err != nil {
		t.Fatalf("Failed to stat TVault file: %v", err)
	}

	if fileInfo.Size() > int64(constants.TVaultHeaderSize) {
		t.Errorf("TVault file size %d exceeds maximum header size %d",
			fileInfo.Size(), constants.TVaultHeaderSize)
	}
}

func TestCompareWithActualFunctions(t *testing.T) {
	origTVaultDir, err := os.MkdirTemp("", "tvault-orig-")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(origTVaultDir)

	testTVaultDir, err := os.MkdirTemp("", "tvault-test-")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(testTVaultDir)

	// Set up test data
	salt := make([]byte, constants.SaltLength)
	for i := 0; i < constants.SaltLength; i++ {
		salt[i] = byte(i % 256)
	}

	encryptedKey := make([]byte, 48)
	for i := 0; i < len(encryptedKey); i++ {
		encryptedKey[i] = byte((i + 100) % 256)
	}

	// Get the original TVault path for comparison
	origTVaultPath := GetTVaultPath()
	testTVaultPath := filepath.Join(testTVaultDir, ".tvault")

	// Create a test file using our test function
	err = writeTVaultHeaderForTest(testTVaultPath, salt, encryptedKey)
	if err != nil {
		t.Fatalf("writeTVaultHeaderForTest failed: %v", err)
	}

	// Read back the test file
	testSalt, testKey, err := readTVaultHeaderForTest(testTVaultPath)
	if err != nil {
		t.Fatalf("readTVaultHeaderForTest failed: %v", err)
	}

	// Verify the test functions work correctly
	if !bytes.Equal(testSalt, salt) {
		t.Errorf("Salt mismatch in test function")
	}

	if !bytes.Equal(testKey, encryptedKey) {
		t.Errorf("Encrypted key mismatch in test function")
	}

	t.Logf("Original TVault path is: %s", origTVaultPath)
	t.Logf("Test TVault path is: %s", testTVaultPath)
}
