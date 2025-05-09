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
	"github.com/wailsapp/wails/v2/pkg/runtime"
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

	runtime.LogInfo(s.ctx, fmt.Sprintf("Stored file %s (%s) at offset %d with size %d", fileName, fileUUID, offset, encryptedSize))
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

	runtime.LogInfo(s.ctx, fmt.Sprintf("Opening file: %s (ID: %d)", metadata.Name, id))

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
		runtime.LogWarning(s.ctx, fmt.Sprintf("Failed to set file permissions: %v", err))
	}

	runtime.LogInfo(s.ctx, fmt.Sprintf("File decrypted successfully to: %s", tempFilePath))

	if err := s.recordTempFile(id, tempFilePath); err != nil {
		runtime.LogWarning(s.ctx, fmt.Sprintf("Failed to record temp file in database: %v", err))
	}

	err = openFileWithDefaultApp(tempFilePath)
	if err != nil {
		runtime.LogWarning(s.ctx, fmt.Sprintf("Failed to open file automatically: %v", err))
		runtime.LogInfo(s.ctx, fmt.Sprintf("File decrypted and saved to: %s", tempFilePath))
		return nil
	}

	runtime.LogInfo(s.ctx, fmt.Sprintf("File decrypted and opened: %s", metadata.Name))
	return nil
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
