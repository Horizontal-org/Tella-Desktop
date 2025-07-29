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
	metadata, err := filestoreutils.GetFileMetadataByID(s.db, id)
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

	tempFilePath := filestoreutils.CreateUniqueFilename(tempDir, metadata.Name)
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

	if err := filestoreutils.RecordTempFile(s.db, id, tempFilePath); err != nil {
		fmt.Printf("Failed to record temp file in database: %v", err)
	}

	err = filestoreutils.OpenFileWithDefaultApp(tempFilePath)
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
		exportPath, err := filestoreutils.ExportSingleFile(s.db, s.dbKey, id, tvault, exportDir)
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

func (s *service) ExportZipFolders(folderIDs []int64, selectedFileIDs []int64) ([]string, error) {
	if len(folderIDs) == 0 {
		return nil, fmt.Errorf("no folder IDs provided")
	}

	var exportedPaths []string
	exportDir := authutils.GetExportDir()

	// Open TVault once for all operations
	tvault, err := os.Open(s.tvaultPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open TVault: %w", err)
	}
	defer tvault.Close()

	for _, folderID := range folderIDs {
		// Get folder info using filestoreutils
		folderInfo, err := filestoreutils.GetFolderInfo(s.db, folderID)
		if err != nil {
			fmt.Printf("Failed to get folder info for ID %d: %v", folderID, err)
			continue
		}

		var filesToExport []filestoreutils.FileInfo

		if len(selectedFileIDs) > 0 && len(folderIDs) == 1 {
			// Scenario 1: Export selected files from within a folder
			fmt.Printf("Exporting %d selected files from folder '%s' as ZIP", len(selectedFileIDs), folderInfo.Name)
			filesToExport, err = filestoreutils.GetSelectedFilesInFolder(s.db, folderID, selectedFileIDs)
		} else {
			// Scenario 2: Export entire folder(s)
			fmt.Printf("Exporting entire folder '%s' as ZIP", folderInfo.Name)
			response, err := s.GetFilesInFolder(folderID)
			if err != nil {
				fmt.Printf("Failed to get files in folder %d: %v", folderID, err)
				continue
			}
			// Convert from service FileInfo to filestoreutils FileInfo
			for _, file := range response.Files {
				filesToExport = append(filesToExport, filestoreutils.FileInfo{
					ID:        file.ID,
					Name:      file.Name,
					MimeType:  file.MimeType,
					Timestamp: file.Timestamp,
					Size:      file.Size,
				})
			}
		}

		if err != nil {
			fmt.Printf("Failed to get files for folder %d: %v", folderID, err)
			continue
		}

		if len(filesToExport) == 0 {
			fmt.Printf("No files to export in folder '%s'", folderInfo.Name)
			continue
		}

		// Create ZIP file using filestoreutils
		zipPath, err := filestoreutils.CreateZipFile(s.db, s.dbKey, folderInfo.Name, filesToExport, tvault, exportDir)
		if err != nil {
			fmt.Printf("Failed to create ZIP for folder '%s': %v", folderInfo.Name, err)
			continue
		}

		exportedPaths = append(exportedPaths, zipPath)
		fmt.Printf("ZIP created successfully: %s", zipPath)
	}

	if len(exportedPaths) == 0 {
		return nil, fmt.Errorf("no ZIP files were created successfully")
	}

	fmt.Printf("ZIP export completed: %d ZIP files created", len(exportedPaths))
	return exportedPaths, nil
}
