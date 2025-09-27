package filestoreutils

import (
	"Tella-Desktop/backend/utils/authutils"
	"archive/zip"
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
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

// GetFileExtensionFromMimeType returns the appropriate file extension for a given mimetype
func GetFileExtensionFromMimeType(mimeType string) string {
	// Common image formats
	switch mimeType {
	case "image/jpeg", "image/jpg":
		return ".jpg"
	case "image/png":
		return ".png"
	case "image/gif":
		return ".gif"
	case "image/webp":
		return ".webp"
	case "image/tiff":
		return ".tiff"
	case "image/bmp":
		return ".bmp"
	case "image/heic":
		return ".heic"
	case "image/heif":
		return ".heif"

	// Video formats
	case "video/mp4":
		return ".mp4"
	case "video/avi":
		return ".avi"
	case "video/mov", "video/quicktime":
		return ".mov"
	case "video/wmv":
		return ".wmv"
	case "video/flv":
		return ".flv"
	case "video/webm":
		return ".webm"
	case "video/3gpp":
		return ".3gp"

	// Audio formats
	case "audio/mpeg", "audio/mp3":
		return ".mp3"
	case "audio/wav":
		return ".wav"
	case "audio/aac":
		return ".aac"
	case "audio/ogg":
		return ".ogg"
	case "audio/flac":
		return ".flac"
	case "audio/m4a":
		return ".m4a"

	// Document formats
	case "application/pdf":
		return ".pdf"
	case "application/msword":
		return ".doc"
	case "application/vnd.openxmlformats-officedocument.wordprocessingml.document":
		return ".docx"
	case "application/vnd.ms-excel":
		return ".xls"
	case "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet":
		return ".xlsx"
	case "application/vnd.ms-powerpoint":
		return ".ppt"
	case "application/vnd.openxmlformats-officedocument.presentationml.presentation":
		return ".pptx"
	case "text/plain":
		return ".txt"
	case "text/html":
		return ".html"
	case "text/css":
		return ".css"
	case "application/javascript", "text/javascript":
		return ".js"
	case "application/json":
		return ".json"
	case "application/xml", "text/xml":
		return ".xml"

	// Archive formats
	case "application/zip":
		return ".zip"
	case "application/x-rar-compressed":
		return ".rar"
	case "application/x-tar":
		return ".tar"
	case "application/gzip":
		return ".gz"

	// Default case: try to extract from mimetype
	default:
		if strings.HasPrefix(mimeType, "image/") {
			return ".img"
		}
		if strings.HasPrefix(mimeType, "video/") {
			return ".video"
		}
		if strings.HasPrefix(mimeType, "audio/") {
			return ".audio"
		}
		if strings.HasPrefix(mimeType, "text/") {
			return ".txt"
		}
		return ".file"
	}
}

// EnsureFileExtension ensures a filename has the correct extension based on its mimetype
func EnsureFileExtension(fileName, mimeType string) string {
	// Check if the filename already has an extension
	if filepath.Ext(fileName) != "" {
		return fileName // Already has an extension, keep it
	}

	// No extension found, add one based on mimetype
	extension := GetFileExtensionFromMimeType(mimeType)
	return fileName + extension
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
	CreatedAt time.Time
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

	// Ensure filename has proper extension based on mimetype
	fileName := EnsureFileExtension(metadata.Name, metadata.MimeType)

	// Create unique filename in export directory
	exportPath := CreateUniqueFilename(exportDir, fileName)

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

	// Ensure filename has proper extension for ZIP entry
	fileName := EnsureFileExtension(file.Name, file.MimeType)

	// Create file in ZIP
	fileWriter, err := zipWriter.Create(fileName)
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

// Delete files
func SecurelyOverwriteFileData(tvaultPath string, offset, length int64) error {
	file, err := os.OpenFile(tvaultPath, os.O_WRONLY, 0600)
	if err != nil {
		return fmt.Errorf("failed to open TVault for writing: %w", err)
	}
	defer file.Close()

	// Generate random data to overwrite the file content
	randomData := make([]byte, length)
	if _, err := rand.Read(randomData); err != nil {
		return fmt.Errorf("failed to generate random data: %w", err)
	}

	// Overwrite the file data at the specified offset
	_, err = file.WriteAt(randomData, offset)
	if err != nil {
		return fmt.Errorf("failed to overwrite file data: %w", err)
	}

	// Force write to disk
	if err := file.Sync(); err != nil {
		return fmt.Errorf("failed to sync file changes: %w", err)
	}

	return nil
}

// AddFreeSpace records a new free space area in the database
func AddFreeSpace(tx *sql.Tx, offset, length int64) error {
	_, err := tx.Exec(`
		INSERT INTO free_spaces (offset, length, created_at)
		VALUES (?, ?, datetime('now'))
	`, offset, length)

	if err != nil {
		return fmt.Errorf("failed to add free space record: %w", err)
	}

	return nil
}

// GetFileMetadataForDeletion retrieves file metadata needed for deletion
func GetFileMetadataForDeletion(tx *sql.Tx, ids []int64) ([]FileMetadata, error) {
	if len(ids) == 0 {
		return nil, fmt.Errorf("no file IDs provided")
	}

	// Create placeholders for SQL IN clause
	placeholders := make([]string, len(ids))
	args := make([]interface{}, len(ids))

	for i, id := range ids {
		placeholders[i] = "?"
		args[i] = id
	}

	query := fmt.Sprintf(`
		SELECT id, uuid, name, size, folder_id, offset, length, created_at 
		FROM files 
		WHERE id IN (%s) AND is_deleted = 0
	`, strings.Join(placeholders, ","))

	rows, err := tx.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query file metadata: %w", err)
	}
	defer rows.Close()

	var filesMetadata []FileMetadata
	for rows.Next() {
		var metadata FileMetadata
		var createdAtStr string

		err := rows.Scan(
			&metadata.ID, &metadata.UUID, &metadata.Name,
			&metadata.Size, &metadata.FolderID, &metadata.Offset,
			&metadata.Length, &createdAtStr,
		)

		if err != nil {
			return nil, fmt.Errorf("failed to scan file metadata: %w", err)
		}

		// Parse timestamp - try RFC3339 first, then fallback to SQLite format
		createdAt, err := time.Parse(time.RFC3339, createdAtStr)
		if err != nil {
			createdAt, err = time.Parse("2006-01-02 15:04:05", createdAtStr)
			if err != nil {
				createdAt = time.Now()
			}
		}
		metadata.CreatedAt = createdAt

		filesMetadata = append(filesMetadata, metadata)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating file metadata: %w", err)
	}

	return filesMetadata, nil
}
