package authutils

import (
	"crypto/aes"
	"crypto/cipher"
	"errors"
)

func EncryptData(data, key []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCMWithRandomNonce(block)
	if err != nil {
		return nil, err
	}

	ciphertext := gcm.Seal(nil, nil, data, nil)
	return ciphertext, nil

}

func DecryptData(ciphertext, key []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCMWithRandomNonce(block)
	if err != nil {
		return nil, err
	}

	if len(ciphertext) < gcm.NonceSize() {
		return nil, errors.New("ciphertext too short")
	}

	//decrypt
	plaintext, err := gcm.Open(nil, nil, ciphertext, nil)
	if err != nil {
		return nil, err
	}

	return plaintext, nil
}
