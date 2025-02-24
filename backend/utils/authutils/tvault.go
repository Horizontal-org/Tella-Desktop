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

	// Read salt
	salt := make([]byte, constants.SaltLength)
	if _, err := file.Read(salt); err != nil {
		return nil, nil, constants.ErrCorruptedTVault
	}

	// the rest of the header (minus version and salt) contains the encrypted database key
	// we'll read up to encryptedKeyMaxSize bytes to account for different encryption overhead
	encryptedKeyMaxSize := constants.TVaultHeaderSize - 1 - constants.SaltLength
	encryptedKey := make([]byte, encryptedKeyMaxSize)
	n, err := file.Read(encryptedKey)
	if err != nil && err != io.EOF {
		return nil, nil, constants.ErrCorruptedTVault
	}

	// Trim to actual size
	encryptedKey = encryptedKey[:n]

	return salt, encryptedKey, nil

}
