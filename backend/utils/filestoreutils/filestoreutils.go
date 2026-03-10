package filestoreutils

import (
	"Tella-Desktop/backend/utils/authutils"
	util "Tella-Desktop/backend/utils/genericutil"
	"archive/zip"
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gabriel-vasile/mimetype"
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
	lookup := mimetype.Lookup(mimeType)

	// our library knows about the given mimetype, return the associated extension
	if lookup != nil {
		return lookup.Extension()
	}

	// if our library doesn't have any record of this mimetype, try to extract something useful from the mimeType. 
	// worst case, if mime type is not anything intelligible,  default to returning just `.file` as extension
	prefixes := []string{"image/", "video/", "audio/", "text/"}
	for _, prefix := range prefixes {
		extractedType, success := strings.CutPrefix(mimeType, prefix)
		if success {
			return "." + extractedType
		}
	}
	return ".file"
}

// EnsureFileExtension ensures a filename has the correct extension based on its mimetype
func EnsureFileExtension(fileName, inferredMIMEType, metadataMimeType string) string {
	// Check if the filename already has an extension
	if filepath.Ext(fileName) != "" {
		return fileName // Already has an extension, keep it
	}

	mimeType := metadataMimeType
	if inferredMIMEType != "" {
		mimeType = inferredMIMEType
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

	// to eliminate the risk of SQLi due to dynamic query construction, we split up the query into one two steps: step 1
	// uses a static prepared sql statement to get all relevant files. step 2 does a Go-based filtering on the returned db results.

	// step 1: get all the files for the given folder marked as not deleted 
	filesInFolderQuery := `SELECT id, name, mime_type, created_at, size 
	FROM files 
	WHERE folder_id = ? AND is_deleted = 0 
	ORDER BY created_at DESC`

	// Query creates a prepared stmt under the hood
	rows, err := db.Query(filesInFolderQuery, folderID)
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
		// step 2: filter retrieved files to the subset defined by fileIDs. 
		// in this case, we only append to slice `files` if file.ID matches one of the ids in fileIDs.
		for _, fid := range fileIDs {
			if file.ID == fid {
				files = append(files, file)
				break
			}
		}
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating files: %w", err)
	}

	return files, nil
}

func decryptAndGetFilename(db *sql.DB, fid int64, dbKey []byte, tvault *os.File) ([]byte, string, error) {
	metadata, err := GetFileMetadataByID(db, fid)
	if err != nil {
		return nil, "", err
	}

	// Read encrypted data from TVault
	encryptedData := make([]byte, metadata.Length)
	_, err = tvault.ReadAt(encryptedData, metadata.Offset)
	if err != nil {
		return nil, "", fmt.Errorf("failed to read file from TVault: %w", err)
	}
	defer util.SecureZeroMemory(encryptedData)

	// Generate file key and decrypt
	fileKey := GenerateFileKey(metadata.UUID, dbKey)
	decryptedData, err := authutils.DecryptData(encryptedData, fileKey)
	if err != nil {
		return nil, "", fmt.Errorf("failed to decrypt file: %w", err)
	}

	inferredMIME := mimetype.Detect(decryptedData)
	var detectedMIME string
	if !inferredMIME.Is("application/octet-stream") {
		detectedMIME = inferredMIME.String()
	}
	// Ensure filename has proper extension based on mimetype
	fileName := EnsureFileExtension(metadata.Name, detectedMIME,  metadata.MimeType)
	return decryptedData, fileName, nil
}

// ExportSingleFile exports a single file to the specified directory
func ExportSingleFile(db *sql.DB, dbKey []byte, id int64, tvault *os.File, exportDir string) (string, error) {
	decryptedData, fileName, err := decryptAndGetFilename(db, id, dbKey, tvault)
	if err != nil {
		return "", err
	}
	defer util.SecureZeroMemory(decryptedData)

	// Create unique filename in export directory
	exportPath := CreateUniqueFilename(exportDir, fileName)

	// Create the exported file
	exportFile, err := util.NarrowCreate(exportPath)
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
	err = os.Chmod(exportPath, util.USER_ONLY_FILE_PERMS)
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
	zipFile, err := util.NarrowCreate(zipPath)
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
	if err := os.Chmod(zipPath, util.USER_ONLY_FILE_PERMS); err != nil {
		fmt.Printf("Failed to set ZIP file permissions: %v", err)
	}

	return zipPath, nil
}

// AddFileToZip adds a single file to an existing ZIP writer
func AddFileToZip(db *sql.DB, dbKey []byte, zipWriter *zip.Writer, file FileInfo, tvault *os.File) error {
	decryptedData, fileName, err := decryptAndGetFilename(db, file.ID, dbKey, tvault)
	if err != nil {
		return err
	}
	defer util.SecureZeroMemory(decryptedData)

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
	file, err := os.OpenFile(tvaultPath, os.O_WRONLY, util.USER_ONLY_FILE_PERMS)
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

	metadataQuery := `
		SELECT uuid, name, size, folder_id, offset, length, created_at 
		FROM files 
		WHERE id = ? AND is_deleted = 0
	`

	// NOTE: we iteratively execute the static sql query to eliminate SQLi risk from dynamic query construction
	// TODO cblgh(2026-02-09): gather up all of these queries and execute in a batch?
	var filesMetadata []FileMetadata

	for _, fileID := range ids {
		var metadata FileMetadata
		metadata.ID = fileID
		var createdAtStr string

		err := tx.QueryRow(metadataQuery, fileID).Scan(
			&metadata.UUID, &metadata.Name,
			&metadata.Size, &metadata.FolderID, &metadata.Offset,
			&metadata.Length, &createdAtStr,
		)

		// TODO cblgh(2026-02-09): decide how best to handle these errors now that we're iterating; terminating too early
		// would be bad and risk disabling the program if a file is malformed / returns an error
		switch {
		case err == sql.ErrNoRows:
			fmt.Printf("no file with id %d\n", fileID)
		case err != nil:
			fmt.Printf("failed to query file metadata: %v", err)
		}

		// Parse timestamp - try RFC3339 first, then fallback to SQLite format
		timeFormats := []string{time.RFC3339, "2006-01-02 15:04:05"}
		var createdAt time.Time
		for _, timeFmt := range timeFormats {
			createdAt, err = time.Parse(timeFmt, createdAtStr)
			if err == nil {
				break
			}
		}
		if createdAt.IsZero() { 
			createdAt = time.Now()
		}
		metadata.CreatedAt = createdAt

		filesMetadata = append(filesMetadata, metadata)
	}

	return filesMetadata, nil
}
