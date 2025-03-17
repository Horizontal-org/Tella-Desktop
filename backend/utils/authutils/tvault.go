package authutils

import (
	"Tella-Desktop/backend/utils/constants"
	"encoding/binary"
	"io"
	"os"
)

// Initialize the TVault file with the salt and encrypted db key
func InitializeTVaultHeader(salt, encryptDBKey []byte) error {
	file, err := os.Create(GetTVaultPath())
	if err != nil {
		return err
	}
	defer file.Close()

	actualBytesWritten := 0

	n, err := file.Write([]byte{constants.CurrentTVaultVersion})
	if err != nil {
		return err
	}
	actualBytesWritten += n

	// Write salt
	n, err = writeLengthAndData(file, salt)
	if err != nil {
		return err
	}
	actualBytesWritten += n

	// Write encrypted key
	n, err = writeLengthAndData(file, encryptDBKey)
	if err != nil {
		return err
	}
	actualBytesWritten += n

	// add padding to reach tvault header size
	paddingNeeded := constants.TVaultHeaderSize - actualBytesWritten
	if paddingNeeded > 0 {
		padding := make([]byte, paddingNeeded)
		if _, err := file.Write(padding); err != nil {
			return err
		}
	}

	return nil
}

func writeLengthAndData(file *os.File, data []byte) (int, error) {
	totalBytesWritten := 0

	lenBuf := make([]byte, constants.LengthFieldSize)
	binary.LittleEndian.PutUint32(lenBuf, uint32(len(data)))

	n, err := file.Write(lenBuf)
	if err != nil {
		return totalBytesWritten, err
	}
	totalBytesWritten += n
	n, err = file.Write(data)
	if err != nil {
		return totalBytesWritten, err
	}
	totalBytesWritten += n

	return totalBytesWritten, nil
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

	version := int(versionByte[0])
	if version <= 0 || version > constants.CurrentTVaultVersion {
		return nil, nil, constants.ErrUnsupportedVersion
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
