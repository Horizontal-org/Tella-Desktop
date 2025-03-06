package filestore

import (
	"Tella-Desktop/backend/utils/authutils"
	"Tella-Desktop/backend/utils/filestoreutils"
	"context"
	"database/sql"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/google/uuid"
	"github.com/wailsapp/wails/v2/pkg/runtime"
)

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
		SELECT name, mime_type, created_at 
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
		if err := rows.Scan(&file.Name, &file.MimeType, &file.Timestamp); err != nil {
			return nil, err
		}
		files = append(files, file)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return files, nil
}
