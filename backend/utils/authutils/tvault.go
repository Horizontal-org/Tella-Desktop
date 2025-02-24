package authutils

import (
	"Tella-Desktop/backend/utils/constants"
	"io"
	"os"
)

// TVault helper methods -> move this to utils
func WriteTVaultHeader(salt, encryptDBKey []byte) error {
	//Create tvault file if it doesn't exist
	file, err := os.Create(GetTVaultPath())
	if err != nil {
		return err
	}
	defer file.Close()

	// Write version
	if _, err := file.Write([]byte{1}); err != nil {
		return err
	}

	// write salt
	if _, err := file.Write(salt); err != nil {
		return err
	}

	// write encrypted db key
	if _, err := file.Write(encryptDBKey); err != nil {
		return err
	}

	headerSize := 1 + len(salt) + len(encryptDBKey)
	if headerSize < constants.TVaultHeaderSize {
		padding := make([]byte, constants.TVaultHeaderSize-headerSize)
		if _, err := file.Write(padding); err != nil {
			return err
		}
	}

	return nil
}

func ReadTVaultHeader() ([]byte, []byte, error) {
	//check if tvault file exists
	file, err := os.Open(GetTVaultPath())
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil, constants.ErrTVaultNotFound
		}
		return nil, nil, err
	}
	defer file.Close()

	// Read version byte
	versionByte := make([]byte, 1)
	if _, err := file.Read(versionByte); err != nil {
		return nil, nil, constants.ErrCorruptedTVault
	}

	// Read salt
	salt := make([]byte, constants.SaltLength)
	if _, err := file.Read(salt); err != nil {
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
