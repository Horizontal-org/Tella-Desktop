package authutils

import (
	"Tella-Desktop/backend/utils/constants"
	"encoding/binary"
	"io"
	"os"
)

// Initialize the TVault file with the salt and encrypted db key
func InitializeTVaultHeader(salt, encryptDBKey []byte) error {
	versionSize := 1
	saltSize := len(salt)
	keySize := len(encryptDBKey)

	// Calculate total header size
	headerSize := versionSize +
		constants.LengthFieldSize + saltSize +
		constants.LengthFieldSize + keySize

	if headerSize > constants.TVaultHeaderSize {
		return constants.ErrHeaderTooLarge
	}

	file, err := os.Create(GetTVaultPath())
	if err != nil {
		return err
	}
	defer file.Close()

	// write version
	if _, err := file.Write([]byte{1}); err != nil {
		return err
	}

	// write salt length and salt
	if err := writeLengthAndData(file, salt); err != nil {
		return err
	}

	// write encrypted db key length and key
	if err := writeLengthAndData(file, encryptDBKey); err != nil {
		return err
	}

	// add padding to reach tvault header size
	paddingNeeded := constants.TVaultHeaderSize - headerSize
	if paddingNeeded > 0 {
		padding := make([]byte, paddingNeeded)
		if _, err := file.Write(padding); err != nil {
			return err
		}
	}

	return nil

}

func writeLengthAndData(file *os.File, data []byte) error {
	lenBuf := make([]byte, constants.LengthFieldSize)
	binary.LittleEndian.PutUint32(lenBuf, uint32(len(data)))
	if _, err := file.Write(lenBuf); err != nil {
		return err
	}

	if _, err := file.Write(data); err != nil {
		return err
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
	salt, err := readLengthPrefixedData(file)
	if err != nil {
		return nil, nil, constants.ErrCorruptedTVault
	}

	// Read encrypted key
	encryptedKey, err := readLengthPrefixedData(file)
	if err != nil {
		return nil, nil, constants.ErrCorruptedTVault
	}

	return salt, encryptedKey, nil

}

func readLengthPrefixedData(file *os.File) ([]byte, error) {
	lenBuf := make([]byte, constants.LengthFieldSize)
	if _, err := io.ReadFull(file, lenBuf); err != nil {
		return nil, err
	}
	dataLen := binary.LittleEndian.Uint32(lenBuf)

	data := make([]byte, dataLen)
	if _, err := io.ReadFull(file, data); err != nil {
		return nil, err
	}

	return data, nil
}
