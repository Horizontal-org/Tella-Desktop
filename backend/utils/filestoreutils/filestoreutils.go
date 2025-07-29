package filestoreutils

import (
	"Tella-Desktop/backend/utils/authutils"
	"archive/zip"
	"crypto/sha256"
	"database/sql"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
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
			return 0, err
		}
		return offset, nil
	} else if err != sql.ErrNoRows {
		return 0, err
	}

	// No suitable free space found, append to end of file
	file, err := os.Stat(tvaultPath)
	if err != nil {
		return 0, err
	}

	return file.Size(), nil
}

// GenerateFileKey generates a file-specific encryption key
func GenerateFileKey(fileUUID string, dbKey []byte) []byte {
	hash := sha256.New()
	hash.Write(dbKey)
	hash.Write([]byte(fileUUID))
	return hash.Sum(nil)
}

// OpenFileWithDefaultApp opens a file with the system's default application
func OpenFileWithDefaultApp(filePath string) error {
	var cmd *exec.Cmd

	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("open", filePath)
	case "windows":
		cmd = exec.Command("cmd", "/c", "start", "", filePath)
	case "linux":
		cmd = exec.Command("xdg-open", filePath)
	default:
		return fmt.Errorf("unsupported operating system")
	}

	return cmd.Start()
}

// CreateUniqueFilename creates a unique filename by appending a counter if the file already exists
func CreateUniqueFilename(dir, fileName string) string {
	originalPath := filepath.Join(dir, fileName)
	if _, err := os.Stat(originalPath); os.IsNotExist(err) {
		return originalPath
	}

	ext := filepath.Ext(fileName)
	baseName := fileName[:len(fileName)-len(ext)]

	counter := 1
	for {
		newName := fmt.Sprintf("%s-%d%s", baseName, counter, ext)
		newPath := filepath.Join(dir, newName)

		if _, err := os.Stat(newPath); os.IsNotExist(err) {
			return newPath
		}

		counter++
	}
}

// FileInfo represents basic file information
type FileInfo struct {
	ID        int64  `json:"id"`
	Name      string `json:"name"`
	MimeType  string `json:"mimeType"`
	Timestamp string `json:"timestamp"`
	Size      int64  `json:"size"`
}

// FolderInfo represents basic folder information
type FolderInfo struct {
	ID        int64  `json:"id"`
	Name      string `json:"name"`
	Timestamp string `json:"timestamp"`
}

// FileMetadata represents complete file metadata including encryption details
type FileMetadata struct {
	ID        int64
	UUID      string
	Name      string
	Size      int64
	MimeType  string
	FolderID  int64
	Offset    int64
	Length    int64
	CreatedAt string
}

// GetFileMetadataByID retrieves file metadata from database by ID
func GetFileMetadataByID(db *sql.DB, id int64) (*FileMetadata, error) {
	var metadata FileMetadata

	err := db.QueryRow(`
		SELECT uuid, name, mime_type, offset, length
		FROM files
		WHERE id = ? AND is_deleted = 0
	`, id).Scan(&metadata.UUID, &metadata.Name, &metadata.MimeType, &metadata.Offset, &metadata.Length)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("file not found with ID: %d", id)
		}
		return nil, fmt.Errorf("failed to fetch file metadata: %w", err)
	}

	return &metadata, nil
}

// GetFolderInfo retrieves folder information from database by ID
func GetFolderInfo(db *sql.DB, folderID int64) (*FolderInfo, error) {
	var folder FolderInfo
	err := db.QueryRow(`
		SELECT id, name, created_at 
		FROM folders 
		WHERE id = ?
	`, folderID).Scan(&folder.ID, &folder.Name, &folder.Timestamp)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("folder not found with ID: %d", folderID)
		}
		return nil, fmt.Errorf("failed to get folder info: %w", err)
	}

	return &folder, nil
}

// GetSelectedFilesInFolder retrieves specific files within a folder from database
func GetSelectedFilesInFolder(db *sql.DB, folderID int64, fileIDs []int64) ([]FileInfo, error) {
	if len(fileIDs) == 0 {
		return nil, fmt.Errorf("no file IDs provided")
	}

	// Create placeholders for SQL IN clause
	placeholders := make([]string, len(fileIDs))
	args := make([]interface{}, len(fileIDs)+1)
	args[0] = folderID

	for i, id := range fileIDs {
		placeholders[i] = "?"
		args[i+1] = id
	}

	query := fmt.Sprintf(`
		SELECT id, name, mime_type, created_at, size 
		FROM files 
		WHERE folder_id = ? AND id IN (%s) AND is_deleted = 0 
		ORDER BY created_at DESC
	`, strings.Join(placeholders, ","))

	rows, err := db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query selected files: %w", err)
	}
	defer rows.Close()

	var files []FileInfo
	for rows.Next() {
		var file FileInfo
		if err := rows.Scan(&file.ID, &file.Name, &file.MimeType, &file.Timestamp, &file.Size); err != nil {
			return nil, fmt.Errorf("failed to scan file: %w", err)
		}
		files = append(files, file)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating files: %w", err)
	}

	return files, nil
}

// ExportSingleFile exports a single file to the specified directory
func ExportSingleFile(db *sql.DB, dbKey []byte, id int64, tvault *os.File, exportDir string) (string, error) {
	metadata, err := GetFileMetadataByID(db, id)
	if err != nil {
		return "", err
	}

	// Read encrypted data from TVault
	encryptedData := make([]byte, metadata.Length)
	_, err = tvault.ReadAt(encryptedData, metadata.Offset)
	if err != nil {
		return "", fmt.Errorf("failed to read file from TVault: %w", err)
	}

	// Generate file key and decrypt
	fileKey := GenerateFileKey(metadata.UUID, dbKey)
	decryptedData, err := authutils.DecryptData(encryptedData, fileKey)
	if err != nil {
		return "", fmt.Errorf("failed to decrypt file: %w", err)
	}

	// Create unique filename in export directory
	exportPath := CreateUniqueFilename(exportDir, metadata.Name)

	// Create the exported file
	exportFile, err := os.Create(exportPath)
	if err != nil {
		return "", fmt.Errorf("failed to create export file: %w", err)
	}
	defer exportFile.Close()

	// Write decrypted data to export file
	_, err = exportFile.Write(decryptedData)
	if err != nil {
		return "", fmt.Errorf("failed to write to export file: %w", err)
	}

	// Set appropriate file permissions
	err = os.Chmod(exportPath, 0644)
	if err != nil {
		fmt.Printf("Failed to set file permissions for %s: %v", exportPath, err)
	}

	return exportPath, nil
}

// CreateZipFile creates a ZIP file containing the specified files
func CreateZipFile(db *sql.DB, dbKey []byte, folderName string, files []FileInfo, tvault *os.File, exportDir string) (string, error) {
	// Create unique ZIP filename
	zipFileName := fmt.Sprintf("%s.zip", folderName)
	zipPath := CreateUniqueFilename(exportDir, zipFileName)

	// Create ZIP file
	zipFile, err := os.Create(zipPath)
	if err != nil {
		return "", fmt.Errorf("failed to create ZIP file: %w", err)
	}
	defer zipFile.Close()

	// Create ZIP writer
	zipWriter := zip.NewWriter(zipFile)
	defer zipWriter.Close()

	// Add each file to ZIP
	for _, file := range files {
		err := AddFileToZip(db, dbKey, zipWriter, file, tvault)
		if err != nil {
			fmt.Printf("Failed to add file '%s' to ZIP: %v", file.Name, err)
			continue // Continue with other files
		}
	}

	// Set appropriate file permissions
	if err := os.Chmod(zipPath, 0644); err != nil {
		fmt.Printf("Failed to set ZIP file permissions: %v", err)
	}

	return zipPath, nil
}

// AddFileToZip adds a single file to an existing ZIP writer
func AddFileToZip(db *sql.DB, dbKey []byte, zipWriter *zip.Writer, file FileInfo, tvault *os.File) error {
	// Get file metadata for decryption
	metadata, err := GetFileMetadataByID(db, file.ID)
	if err != nil {
		return fmt.Errorf("failed to get metadata for file %d: %w", file.ID, err)
	}

	// Read and decrypt file
	encryptedData := make([]byte, metadata.Length)
	_, err = tvault.ReadAt(encryptedData, metadata.Offset)
	if err != nil {
		return fmt.Errorf("failed to read encrypted data: %w", err)
	}

	fileKey := GenerateFileKey(metadata.UUID, dbKey)
	decryptedData, err := authutils.DecryptData(encryptedData, fileKey)
	if err != nil {
		return fmt.Errorf("failed to decrypt file: %w", err)
	}

	// Create file in ZIP
	fileWriter, err := zipWriter.Create(file.Name)
	if err != nil {
		return fmt.Errorf("failed to create file in ZIP: %w", err)
	}

	// Write decrypted data to ZIP entry
	_, err = fileWriter.Write(decryptedData)
	if err != nil {
		return fmt.Errorf("failed to write file data to ZIP: %w", err)
	}

	return nil
}

// RecordTempFile records a temporary file in the database for cleanup
func RecordTempFile(db *sql.DB, fileID int64, tempPath string) error {
	_, err := db.Exec(`
		INSERT INTO temp_files (file_id, temp_path, created_at)
		VALUES (?, ?, datetime('now'))
	`, fileID, tempPath)

	return err
}
