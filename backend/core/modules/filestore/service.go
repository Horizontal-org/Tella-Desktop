package filestore

import (
	"Tella-Desktop/backend/utils/authutils"
	"Tella-Desktop/backend/utils/filestoreutils"
	"context"
	"database/sql"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	goruntime "runtime"
	"time"

	"github.com/google/uuid"
)

type service struct {
	ctx        context.Context
	db         *sql.DB
	tvaultPath string
	dbKey      []byte
}

func NewService(ctx context.Context, db *sql.DB, dbKey []byte) Service {
	return &service{
		ctx:        ctx,
		db:         db,
		tvaultPath: authutils.GetTVaultPath(),
		dbKey:      dbKey,
	}
}

// StoreFile encrypts and stores a file in TVault
func (s *service) StoreFile(folderID int64, fileName string, mimeType string, reader io.Reader) (*FileMetadata, error) {
	// Begin Transaction
	tx, err := s.db.Begin()
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Generate UUID for the file
	fileUUID := uuid.New().String()

	// Read the entire file into memory
	fileData, err := io.ReadAll(reader)
	if err != nil {
		return nil, fmt.Errorf("failed to read file data: %w", err)
	}

	originalSize := int64(len(fileData))
	fileKey := filestoreutils.GenerateFileKey(fileUUID, s.dbKey)

	encryptedData, err := authutils.EncryptData(fileData, fileKey)
	if err != nil {
		return nil, fmt.Errorf("failed to encrypt file: %w", err)
	}

	encryptedSize := int64(len(encryptedData))

	// Find space in TVault to store the file
	offset, err := filestoreutils.FindSpace(tx, encryptedSize, s.tvaultPath)
	if err != nil {
		return nil, fmt.Errorf("failed to find space in TVault: %w", err)
	}

	// Open TVault file
	tvault, err := os.OpenFile(s.tvaultPath, os.O_RDWR, 0600)
	if err != nil {
		return nil, fmt.Errorf("failed to open TVault: %w", err)
	}
	defer tvault.Close()

	// Write encrypted data to TVault
	_, err = tvault.WriteAt(encryptedData, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to write to TVault: %w", err)
	}

	// Insert file metadata into database
	fileID, err := filestoreutils.InsertFileMetadata(tx, fileUUID, fileName, originalSize, mimeType, folderID, offset, encryptedSize)
	if err != nil {
		return nil, fmt.Errorf("failed to insert file metadata: %w", err)
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	// Return metadata
	metadata := &FileMetadata{
		ID:        fileID,
		UUID:      fileUUID,
		Name:      fileName,
		Size:      originalSize,
		MimeType:  mimeType,
		FolderID:  folderID,
		Offset:    offset,
		Length:    encryptedSize,
		CreatedAt: time.Now(),
	}

	fmt.Printf("Stored file %s (%s) at offset %d with size %d", fileName, fileUUID, offset, encryptedSize)
	return metadata, nil

}

func (s *service) GetStoredFiles() ([]FileInfo, error) {
	rows, err := s.db.Query(`
		SELECT id, name, mime_type, created_at 
		FROM files 
		WHERE is_deleted = 0 
		ORDER BY created_at DESC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var files []FileInfo
	for rows.Next() {
		var file FileInfo
		if err := rows.Scan(&file.ID, &file.Name, &file.MimeType, &file.Timestamp); err != nil {
			return nil, err
		}
		files = append(files, file)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return files, nil
}

func (s *service) OpenFileByID(id int64) error {
	metadata, err := s.getFileMetadataByID(id)
	if err != nil {
		return err
	}

	fmt.Printf("Opening file: %s (ID: %d)", metadata.Name, id)

	tvault, err := os.Open(s.tvaultPath)
	if err != nil {
		return fmt.Errorf("failed to open TVault: %w", err)
	}
	defer tvault.Close()

	encryptedData := make([]byte, metadata.Length)
	_, err = tvault.ReadAt(encryptedData, metadata.Offset)
	if err != nil {
		return fmt.Errorf("failed to read file from TVault: %w", err)
	}

	fileKey := filestoreutils.GenerateFileKey(metadata.UUID, s.dbKey)
	decryptedData, err := authutils.DecryptData(encryptedData, fileKey)
	if err != nil {
		return fmt.Errorf("failed to decrypt file: %w", err)
	}

	tempDir := authutils.GetTempDir()
	if err := os.MkdirAll(tempDir, 0755); err != nil {
		return fmt.Errorf("failed to create temp directory: %w", err)
	}

	tempFilePath := createUniqueFilename(tempDir, metadata.Name)
	tempFile, err := os.Create(tempFilePath)
	if err != nil {
		return fmt.Errorf("failed to create temp file: %w", err)
	}

	_, err = tempFile.Write(decryptedData)
	if err != nil {
		tempFile.Close()
		return fmt.Errorf("failed to write to temp file: %w", err)
	}
	tempFile.Close()

	err = os.Chmod(tempFilePath, 0644)
	if err != nil {
		fmt.Printf("Failed to set file permissions: %v", err)
	}

	fmt.Printf("File decrypted successfully to: %s", tempFilePath)

	if err := s.recordTempFile(id, tempFilePath); err != nil {
		fmt.Printf("Failed to record temp file in database: %v", err)
	}

	err = openFileWithDefaultApp(tempFilePath)
	if err != nil {
		fmt.Printf("Failed to open file automatically: %v", err)
		fmt.Printf("File decrypted and saved to: %s", tempFilePath)
		return nil
	}

	fmt.Printf("File decrypted and opened: %s", metadata.Name)
	return nil
}

func (s *service) GetStoredFolders() ([]FolderInfo, error) {
	rows, err := s.db.Query(`
		SELECT 
			f.id, 
			f.name, 
			f.created_at,
			COUNT(files.id) as file_count
		FROM folders f
		LEFT JOIN files ON f.id = files.folder_id AND files.is_deleted = 0
		GROUP BY f.id, f.name, f.created_at
		HAVING COUNT(files.id) > 0
		ORDER BY f.created_at DESC
	`)
	if err != nil {
		return nil, fmt.Errorf("failed to query folders: %w", err)
	}
	defer rows.Close()

	var folders []FolderInfo
	for rows.Next() {
		var folder FolderInfo
		if err := rows.Scan(&folder.ID, &folder.Name, &folder.Timestamp, &folder.FileCount); err != nil {
			return nil, fmt.Errorf("failed to scan folder: %w", err)
		}
		folders = append(folders, folder)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating folders: %w", err)
	}

	return folders, nil
}

func (s *service) GetFilesInFolder(folderID int64) (*FilesInFolderResponse, error) {
	var folderName string
	err := s.db.QueryRow("SELECT name FROM folders WHERE id = ?", folderID).Scan(&folderName)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("folder not found with ID: %d", folderID)
		}
		return nil, fmt.Errorf("failed to get folder name: %w", err)
	}

	rows, err := s.db.Query(`
		SELECT id, name, mime_type, created_at, size 
		FROM files 
		WHERE folder_id = ? AND is_deleted = 0 
		ORDER BY created_at DESC
	`, folderID)
	if err != nil {
		return nil, fmt.Errorf("failed to query files in folder: %w", err)
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

	return &FilesInFolderResponse{
		FolderName: folderName,
		Files:      files,
	}, nil
}

func (s *service) ExportFiles(ids []int64) ([]string, error) {
	if len(ids) == 0 {
		return nil, fmt.Errorf("no file IDs provided")
	}

	if len(ids) == 1 {
		fmt.Printf("Exporting single file with ID: %d", ids[0])
	} else {
		fmt.Printf("Exporting %d files in batch", len(ids))
	}

	var exportedPaths []string
	var failedFiles []string

	// Get export directory once
	exportDir := authutils.GetExportDir()

	// Open TVault once for all operations
	tvault, err := os.Open(s.tvaultPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open TVault: %w", err)
	}
	defer tvault.Close()

	for _, id := range ids {
		// Export each file individually
		exportPath, err := s.exportSingleFile(id, tvault, exportDir)
		if err != nil {
			fmt.Printf("Failed to export file ID %d: %v", id, err)
			failedFiles = append(failedFiles, fmt.Sprintf("ID %d", id))
			continue
		}

		exportedPaths = append(exportedPaths, exportPath)
		if len(ids) == 1 {
			fmt.Printf("File exported successfully to: %s", exportPath)
		} else {
			fmt.Printf("File ID %d exported successfully to: %s", id, exportPath)
		}
	}

	// Return results with error info if some files failed
	if len(failedFiles) > 0 {
		if len(exportedPaths) == 0 {
			return nil, fmt.Errorf("all files failed to export: %v", failedFiles)
		}
		fmt.Printf("Warning: Some files failed to export: %v", failedFiles)
	}

	if len(ids) == 1 {
		fmt.Printf("Export completed successfully")
	} else {
		fmt.Printf("Batch export completed: %d/%d files exported successfully", len(exportedPaths), len(ids))
	}

	return exportedPaths, nil
}

func openFileWithDefaultApp(filePath string) error {
	var cmd *exec.Cmd

	switch goruntime.GOOS {
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

func (s *service) getFileMetadataByID(id int64) (*FileMetadata, error) {
	var metadata FileMetadata

	err := s.db.QueryRow(`
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

func (s *service) recordTempFile(fileID int64, tempPath string) error {
	_, err := s.db.Exec(`
		INSERT INTO temp_files (file_id, temp_path, created_at)
		VALUES (?, ?, datetime('now'))
	`, fileID, tempPath)

	return err
}

func createUniqueFilename(dir, fileName string) string {
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

func (s *service) exportSingleFile(id int64, tvault *os.File, exportDir string) (string, error) {
	metadata, err := s.getFileMetadataByID(id)
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
	fileKey := filestoreutils.GenerateFileKey(metadata.UUID, s.dbKey)
	decryptedData, err := authutils.DecryptData(encryptedData, fileKey)
	if err != nil {
		return "", fmt.Errorf("failed to decrypt file: %w", err)
	}

	// Create unique filename in export directory
	exportPath := createUniqueFilename(exportDir, metadata.Name)

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
