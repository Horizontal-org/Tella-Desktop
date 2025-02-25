package authutils

import (
	"bytes"
	"crypto/rand"
	"fmt"
	"testing"

	"Tella-Desktop/backend/utils/constants"
)

func TestEncryptDecrypt(t *testing.T) {
	// Generate random data and key
	data := make([]byte, 1024)
	if _, err := rand.Read(data); err != nil {
		t.Fatalf("Failed to generate random data: %v", err)
	}

	key := make([]byte, constants.KeyLength)
	if _, err := rand.Read(key); err != nil {
		t.Fatalf("Failed to generate random key: %v", err)
	}

	// Encrypt the data
	encrypted, err := EncryptData(data, key)
	if err != nil {
		t.Fatalf("Failed to encrypt data: %v", err)
	}

	// Encrypted data should be different from original data
	if bytes.Equal(encrypted, data) {
		t.Errorf("Encrypted data matches original data")
	}

	// Decrypt the data
	decrypted, err := DecryptData(encrypted, key)
	if err != nil {
		t.Fatalf("Failed to decrypt data: %v", err)
	}

	// Decrypted data should match original data
	if !bytes.Equal(decrypted, data) {
		t.Errorf("Decrypted data does not match original data")
	}
}

func TestDecryptWithWrongKey(t *testing.T) {
	// Generate random data and keys
	data := make([]byte, 1024)
	if _, err := rand.Read(data); err != nil {
		t.Fatalf("Failed to generate random data: %v", err)
	}

	key := make([]byte, constants.KeyLength)
	if _, err := rand.Read(key); err != nil {
		t.Fatalf("Failed to generate random key: %v", err)
	}

	wrongKey := make([]byte, constants.KeyLength)
	if _, err := rand.Read(wrongKey); err != nil {
		t.Fatalf("Failed to generate random wrong key: %v", err)
	}

	// Encrypt with correct key
	encrypted, err := EncryptData(data, key)
	if err != nil {
		t.Fatalf("Failed to encrypt data: %v", err)
	}

	// Try to decrypt with wrong key
	_, err = DecryptData(encrypted, wrongKey)
	if err == nil {
		t.Errorf("Expected error when decrypting with wrong key, got none")
	}
}

func TestEncryptionWithVariousSizes(t *testing.T) {
	key := make([]byte, constants.KeyLength)
	if _, err := rand.Read(key); err != nil {
		t.Fatalf("Failed to generate random key: %v", err)
	}

	// Test various data sizes
	testSizes := []int{0, 1, 16, 100, 1024, 1024 * 10}

	for _, size := range testSizes {
		t.Run(fmt.Sprintf("Size-%d", size), func(t *testing.T) {
			data := make([]byte, size)
			if size > 0 {
				if _, err := rand.Read(data); err != nil {
					t.Fatalf("Failed to generate random data: %v", err)
				}
			}

			// Encrypt data
			encrypted, err := EncryptData(data, key)
			if err != nil {
				t.Fatalf("Failed to encrypt data of size %d: %v", size, err)
			}

			// Verify encrypted data is different and larger (contains nonce + tag)
			if size > 0 && bytes.Equal(encrypted, data) {
				t.Errorf("Encrypted data of size %d matches original data", size)
			}

			// AES-GCM adds 12 bytes for nonce and 16 bytes for tag
			expectedMinSize := size + 12 + 16
			if len(encrypted) < expectedMinSize {
				t.Errorf("Encrypted data size %d is smaller than expected minimum %d",
					len(encrypted), expectedMinSize)
			}

			// Decrypt and verify
			decrypted, err := DecryptData(encrypted, key)
			if err != nil {
				t.Fatalf("Failed to decrypt data of size %d: %v", size, err)
			}

			if !bytes.Equal(decrypted, data) {
				t.Errorf("Decrypted data of size %d does not match original", size)
			}
		})
	}
}

func TestEncryptWithInvalidKey(t *testing.T) {
	data := []byte("test data")

	// Try with key that's too short
	shortKey := make([]byte, constants.KeyLength-1)
	if _, err := rand.Read(shortKey); err != nil {
		t.Fatalf("Failed to generate random key: %v", err)
	}

	_, err := EncryptData(data, shortKey)
	if err == nil {
		t.Errorf("Expected error when encrypting with short key, got none")
	}

	// Try with key that's too long
	longKey := make([]byte, constants.KeyLength+1)
	if _, err := rand.Read(longKey); err != nil {
		t.Fatalf("Failed to generate random key: %v", err)
	}

	// This might actually work depending on the implementation, but worth testing
	_, err = EncryptData(data, longKey)
	if err != nil {
		t.Logf("Encrypting with long key failed as expected: %v", err)
	}
}

func TestDecryptCorruptData(t *testing.T) {
	data := []byte("test data")

	key := make([]byte, constants.KeyLength)
	if _, err := rand.Read(key); err != nil {
		t.Fatalf("Failed to generate random key: %v", err)
	}

	// Encrypt the data
	encrypted, err := EncryptData(data, key)
	if err != nil {
		t.Fatalf("Failed to encrypt data: %v", err)
	}

	// Corrupt the encrypted data
	corruptedData := make([]byte, len(encrypted))
	copy(corruptedData, encrypted)
	if len(corruptedData) > 0 {
		// Modify the first byte
		corruptedData[0] ^= 0xFF
	}

	// Try to decrypt corrupted data
	_, err = DecryptData(corruptedData, key)
	if err == nil {
		t.Errorf("Expected error when decrypting corrupted data, got none")
	}

	// Test with truncated data
	if len(encrypted) > 1 {
		truncatedData := encrypted[:len(encrypted)-1]
		_, err = DecryptData(truncatedData, key)
		if err == nil {
			t.Errorf("Expected error when decrypting truncated data, got none")
		}
	}
}

func BenchmarkEncryption(b *testing.B) {
	data := make([]byte, 1024*1024) // 1MB
	if _, err := rand.Read(data); err != nil {
		b.Fatalf("Failed to generate random data: %v", err)
	}

	key := make([]byte, constants.KeyLength)
	if _, err := rand.Read(key); err != nil {
		b.Fatalf("Failed to generate random key: %v", err)
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, err := EncryptData(data, key)
		if err != nil {
			b.Fatalf("Encryption failed: %v", err)
		}
	}
}

func BenchmarkDecryption(b *testing.B) {
	data := make([]byte, 1024*1024) // 1MB
	if _, err := rand.Read(data); err != nil {
		b.Fatalf("Failed to generate random data: %v", err)
	}

	key := make([]byte, constants.KeyLength)
	if _, err := rand.Read(key); err != nil {
		b.Fatalf("Failed to generate random key: %v", err)
	}

	encrypted, err := EncryptData(data, key)
	if err != nil {
		b.Fatalf("Failed to encrypt data: %v", err)
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, err := DecryptData(encrypted, key)
		if err != nil {
			b.Fatalf("Decryption failed: %v", err)
		}
	}
}
