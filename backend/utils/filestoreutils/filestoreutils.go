package filestoreutils

import (
	"Tella-Desktop/backend/utils/constants"
	"crypto/sha256"
	"database/sql"
	"fmt"
	"os"
)

// insertFileMetadata adds file metadata to the database
func InsertFileMetadata(
	tx *sql.Tx,
	fileUUID string,
	fileName string,
	size int64,
	mimeType string,
	folderID int64,
	offset int64,
	length int64,
) (int64, error) {
	result, err := tx.Exec(`
		INSERT INTO files (
			uuid, name, size, folder_id, mime_type, offset, length, 
			is_deleted, created_at, updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, 0, datetime('now'), datetime('now'))
	`,
		fileUUID, fileName, size, folderID, mimeType, offset, length,
	)

	if err != nil {
		return 0, err
	}

	return result.LastInsertId()
}

// findSpace looks for a suitable free space or returns the end of the file
func FindSpace(tx *sql.Tx, size int64, tvaultPath string) (int64, error) {
	// First try to find a free space that fits
	var freeSpaceID, offset int64
	err := tx.QueryRow(`
		SELECT id, offset FROM free_spaces 
		WHERE length >= ? 
		ORDER BY length ASC LIMIT 1
	`, size).Scan(&freeSpaceID, &offset)

	if err == nil {
		// Found a free space, remove or resize it
		_, err = tx.Exec("DELETE FROM free_spaces WHERE id = ?", freeSpaceID)
		if err != nil {
			return 0, fmt.Errorf("failed to remove free space: %w", err)
		}
		return offset, nil
	}

	if err != sql.ErrNoRows {
		return 0, fmt.Errorf("failed to query free spaces: %w", err)
	}

	// No suitable free space, append to end of file

	// First check if TVault header exists but no files yet
	tvaultInfo, err := os.Stat(tvaultPath)
	if err != nil {
		// TVault might not exist yet, try creating it
		if os.IsNotExist(err) {
			// This should not happen as TVault header should be created during setup
			return 0, fmt.Errorf("TVault file not found")
		}
		return 0, fmt.Errorf("failed to stat TVault: %w", err)
	}

	if tvaultInfo.Size() <= int64(constants.TVaultHeaderSize) {
		// Only header exists, start at header size
		return int64(constants.TVaultHeaderSize), nil
	}

	// Find the highest offset + length to determine the end of the file
	var endOfFile int64
	err = tx.QueryRow(`
		SELECT COALESCE(MAX(offset + length), ?) FROM files
	`, constants.TVaultHeaderSize).Scan(&endOfFile)

	if err != nil {
		return 0, fmt.Errorf("failed to determine end of file: %w", err)
	}

	return endOfFile, nil
}

// helper functions
// generateFileKey creates a unique key for a file based on the database key and file UUID
func GenerateFileKey(fileUUID string, key []byte) []byte {
	// Create a key by hashing the database key with the file UUID
	h := sha256.New()
	h.Write(key)
	h.Write([]byte(fileUUID))
	return h.Sum(nil)
}
