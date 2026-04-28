package filestore

import (
	"Tella-Desktop/backend/utils/authutils"
	"Tella-Desktop/backend/utils/filestoreutils"
	util "Tella-Desktop/backend/utils/genericutil"
	"Tella-Desktop/backend/utils/devlog"
	"Tella-Desktop/backend/utils/transferutils"
	"context"
	"database/sql"
	"errors"
	"fmt"
	"io"
	"os"
	"time"
	"crypto/sha256"

	"github.com/gabriel-vasile/mimetype"
	"github.com/google/uuid"
)

var log = devlog.Logger("filestore")

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

var errStoreFile = errors.New("failed to store file")
// StoreFile encrypts and stores a file in TVault
func (s *service) StoreFile(folderID, claimedSize int64, claimedHash string, fileName string, claimedMimeType string, reader io.Reader) (*FileMetadata, error) {
	// Begin Transaction
	tx, err := s.db.Begin()
	if err != nil {
		log("failed to begin transaction: %w", err)
		return nil, errStoreFile
	}
	defer tx.Rollback()

	// Generate UUID for the file
	fileUUID := uuid.New().String()

	// Read the entire file into memory
	// TODO cblgh(2026-02-16): when sending a ~200MB (video/quicktime) file i get a 'i/o timeout' error.
	// this happens when i send from Tella iOS a quicktime video at the same time as a bunch of heic files.
	// error message:
	// Upload failed: failed to store file: failed to read file data: i/o timeout
	//
	// as a piece of debugging information, it happens after ~150MB is sent.
	fileData, err := io.ReadAll(reader)
	log("filestore err?", fileName, err)
	if err != nil {
		// need to return "%w" here so we can unwrap it in package transfer
		return nil, fmt.Errorf("failed to read file data: %w", err)
	}

	// TODO (2026-03-14): convert to use incremental hashing if incremental file storagage + reading is implemented
	sum := sha256.Sum256(fileData)
	if fmt.Sprintf("%x", sum) != claimedHash {
		return nil, transferutils.ErrTransferHashMismatch
	}

	inferredMIME := mimetype.Detect(fileData)
	// TODO cblgh(2026-03-13): decide how to handle mimetype mismatch
	if inferredMIME != nil && !inferredMIME.Is("application/octet-stream") && !inferredMIME.Is(claimedMimeType) {
		log("MISMATCH DETECTED: claimed mimetype does not match mimetype based on file data")
	}

	originalSize := int64(len(fileData))
	log("filestore", fileName, "read size", originalSize)
	if originalSize != claimedSize {
		log("file %q: downloaded size (%d) did not match claimed size (%d) from prepareUpload (difference: %d)", fileName, originalSize, claimedSize, originalSize-claimedSize)
		return nil, errStoreFile
	}
	fileKey := filestoreutils.GenerateFileKey(fileUUID, s.dbKey)

	// TODO cblgh(2026-02-12): to overwrite fileData with encryptedData, do fileData[:0] -- but will the capacity be sufficient?
	encryptedData, err := authutils.EncryptData(fileData, fileKey)
	if err != nil {
		log("failed to encrypt file: %w", err)
		return nil, errStoreFile
	}
	// at this point we have transformed fileData into encryptedData: erase fileData's contents.
	util.SecureZeroMemory(fileData)
	// while we're at it: erase encryptedData once we're done here
	defer util.SecureZeroMemory(encryptedData)

	encryptedSize := int64(len(encryptedData))

	// Find space in TVault to store the file
	offset, err := filestoreutils.FindSpace(tx, encryptedSize, s.tvaultPath)
	if err != nil {
		log("failed to find space in TVault: %w", err)
		return nil, errStoreFile
	}

	// Open TVault file
	tvault, err := os.OpenFile(s.tvaultPath, os.O_RDWR, util.USER_ONLY_FILE_PERMS)
	if err != nil {
		log("failed to open TVault: %w", err)
		return nil, errStoreFile
	}
	defer tvault.Close()

	// Write encrypted data to TVault
	_, err = tvault.WriteAt(encryptedData, offset)
	if err != nil {
		log("failed to write to TVault: %w", err)
		return nil, errStoreFile
	}

	// Insert file metadata into database
	fileID, err := filestoreutils.InsertFileMetadata(tx, fileUUID, fileName, originalSize, claimedMimeType, folderID, offset, encryptedSize)
	if err != nil {
		log("failed to insert file metadata: %w", err)
		return nil, errStoreFile
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		log("failed to commit transaction: %w", err)
		return nil, errStoreFile
	}

	// Return metadata
	metadata := &FileMetadata{
		ID:        fileID,
		UUID:      fileUUID,
		Name:      fileName,
		Size:      originalSize,
		MimeType:  claimedMimeType,
		FolderID:  folderID,
		Offset:    offset,
		Length:    encryptedSize,
		CreatedAt: time.Now(),
	}

	log("Stored file %s (%s) at offset %d with size (encrypted) %d\n", fileName, fileUUID, offset, encryptedSize)
	return metadata, nil
}

var errGetFolders = errors.New("failed to get folders")
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
		log("failed to query folders: %w", err)
		return nil, errGetFolders
	}
	defer rows.Close()

	var folders []FolderInfo
	for rows.Next() {
		var folder FolderInfo
		if err := rows.Scan(&folder.ID, &folder.Name, &folder.Timestamp, &folder.FileCount); err != nil {
			log("failed to scan folder: %w", err)
			return nil, errGetFolders
		}
		folders = append(folders, folder)
	}

	if err := rows.Err(); err != nil {
		log("error iterating folders: %w", err)
		return nil, errGetFolders 
	}

	return folders, nil
}

var errGetFilesFolder = errors.New("failed to get files in folder")
func (s *service) GetFilesInFolder(folderID int64) (*FilesInFolderResponse, error) {
	var folderName string
	err := s.db.QueryRow("SELECT name FROM folders WHERE id = ?", folderID).Scan(&folderName)
	if err != nil {
		if err == sql.ErrNoRows {
			log("folder not found with ID: %d", folderID)
			return nil, errGetFilesFolder
		}
		log("failed to get folder name: %w", err)
		return nil, errGetFilesFolder
	}

	rows, err := s.db.Query(`
		SELECT id, name, mime_type, created_at, size 
		FROM files 
		WHERE folder_id = ? AND is_deleted = 0 
		ORDER BY created_at DESC
	`, folderID)
	if err != nil {
		log("failed to query files in folder: %w", err)
		return nil, errGetFilesFolder
	}
	defer rows.Close()

	var files []FileInfo
	for rows.Next() {
		var file FileInfo
		if err := rows.Scan(&file.ID, &file.Name, &file.MimeType, &file.Timestamp, &file.Size); err != nil {
			log("failed to scan file: %w", err)
			return nil, errGetFilesFolder
		}
		files = append(files, file)
	}

	if err := rows.Err(); err != nil {
		log("error iterating files: %w", err)
		return nil, errGetFilesFolder
	}

	return &FilesInFolderResponse{
		FolderName: folderName,
		Files:      files,
	}, nil
}

var errExportFiles = errors.New("failed to export files")
func (s *service) ExportFiles(ids []int64) ([]string, error) {
	if len(ids) == 0 {
		log("no file IDs provided")
		return nil, errExportFiles
	}

	if len(ids) == 1 {
		log("Exporting single file with ID: %d", ids[0])
	} else {
		log("Exporting %d files in batch", len(ids))
	}

	var exportedPaths []string
	var failedFiles []string

	// Get export directory once
	exportDir := authutils.GetExportDir()
	if err := os.MkdirAll(exportDir, util.USER_ONLY_DIR_PERMS); err != nil {
		log("failed to create export dir: %w", err)
		return nil, errExportFiles
	}

	// Open TVault once for all operations
	tvault, err := os.Open(s.tvaultPath)
	if err != nil {
		log("failed to open TVault: %w", err)
		return nil, errExportFiles
	}
	defer tvault.Close()

	for _, id := range ids {
		// Export each file individually
		exportPath, err := filestoreutils.ExportSingleFile(s.db, s.dbKey, id, tvault, exportDir)
		if err != nil {
			log("Failed to export file ID %d: %v", id, err)
			failedFiles = append(failedFiles, fmt.Sprintf("ID %d", id))
			continue
		}

		exportedPaths = append(exportedPaths, exportPath)
		if len(ids) == 1 {
			log("File exported successfully to: %s", exportPath)
		} else {
			log("File ID %d exported successfully to: %s", id, exportPath)
		}
	}

	// Return results with error info if some files failed
	if len(failedFiles) > 0 {
		if len(exportedPaths) == 0 {
			log("all files failed to export: %v", failedFiles)
			return nil, errExportFiles
		}
		log("Warning: Some files failed to export: %v", failedFiles)
	}

	if len(ids) == 1 {
		log("Export completed successfully")
	} else {
		log("Batch export completed: %d/%d files exported successfully", len(exportedPaths), len(ids))
	}

	return exportedPaths, nil
}

var errExportZipFolders = errors.New("failed to export zip folders")
func (s *service) ExportZipFolders(folderIDs []int64, selectedFileIDs []int64) ([]string, error) {
	if len(folderIDs) == 0 {
		log("no folder IDs provided")
		return nil, errExportZipFolders
	}

	var exportedPaths []string
	exportDir := authutils.GetExportDir()
	if err := os.MkdirAll(exportDir, util.USER_ONLY_DIR_PERMS); err != nil {
		log("failed to create export dir: %w", err)
		return nil, errExportZipFolders
	}

	// Open TVault once for all operations
	tvault, err := os.Open(s.tvaultPath)
	if err != nil {
		log("failed to open TVault: %w", err)
		return nil, errExportZipFolders
	}
	defer tvault.Close()

	for _, folderID := range folderIDs {
		// Get folder info using filestoreutils
		folderInfo, err := filestoreutils.GetFolderInfo(s.db, folderID)
		if err != nil {
			log("Failed to get folder info for ID %d: %v", folderID, err)
			continue
		}

		var filesToExport []filestoreutils.FileInfo

		if len(selectedFileIDs) > 0 && len(folderIDs) == 1 {
			// Scenario 1: Export selected files from within a folder
			log("Exporting %d selected files from folder '%s' as ZIP", len(selectedFileIDs), folderInfo.Name)
			filesToExport, err = filestoreutils.GetSelectedFilesInFolder(s.db, folderID, selectedFileIDs)
		} else {
			// Scenario 2: Export entire folder(s)
			log("Exporting entire folder '%s' as ZIP", folderInfo.Name)
			response, err := s.GetFilesInFolder(folderID)
			if err != nil {
				log("Failed to get files in folder %d: %v", folderID, err)
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
			log("Failed to get files for folder %d: %v", folderID, err)
			continue
		}

		if len(filesToExport) == 0 {
			log("No files to export in folder '%s'", folderInfo.Name)
			continue
		}

		// Create ZIP file using filestoreutils
		zipPath, err := filestoreutils.CreateZipFile(s.db, s.dbKey, folderInfo.Name, filesToExport, tvault, exportDir)
		if err != nil {
			log("Failed to create ZIP for folder '%s': %v", folderInfo.Name, err)
			continue
		}

		exportedPaths = append(exportedPaths, zipPath)
		log("ZIP created successfully: %s", zipPath)
	}

	if len(exportedPaths) == 0 {
		log("no ZIP files were created successfully")
		return nil, errExportZipFolders
	}

	log("ZIP export completed: %d ZIP files created", len(exportedPaths))
	return exportedPaths, nil
}

var errDeleteFiles = errors.New("error when deleting files")
func (s *service) DeleteFiles(ids []int64) error {
	if len(ids) == 0 {
		log("no file IDs provided for deletion")
		return errDeleteFiles
	}

	// Start transaction
	tx, err := s.db.Begin()
	if err != nil {
		log("failed to begin transaction: %w", err)
		return errDeleteFiles
	}
	defer tx.Rollback()

	// Get file metadata for deletion
	filesMetadata, err := filestoreutils.GetFileMetadataForDeletion(tx, ids)
	if err != nil {
		log("failed to get file metadata for deletion: %w", err)
		return errDeleteFiles
	}

	if len(filesMetadata) == 0 {
		log("no files found for deletion")
		return errDeleteFiles
	}

	// Mark files as deleted in database and add to free spaces
	for _, metadata := range filesMetadata {
		_, err := tx.Exec(`
			UPDATE files 
			SET is_deleted = 1, updated_at = datetime('now')
			WHERE id = ?
		`, metadata.ID)

		if err != nil {
			log("failed to mark file %d as deleted: %w", metadata.ID, err)
			return errDeleteFiles
		}

		// Add the file's space to free_spaces table
		err = filestoreutils.AddFreeSpace(tx, metadata.Offset, metadata.Length)
		if err != nil {
			log("failed to add free space for file %d: %w", metadata.ID, err)
			return errDeleteFiles
		}
	}

	// Commit database transaction first
	if err := tx.Commit(); err != nil {
		log("failed to commit deletion transaction: %w", err)
		return errDeleteFiles
	}

	// Now securely overwrite the file data in TVault
	for _, metadata := range filesMetadata {
		err := filestoreutils.SecurelyOverwriteFileData(s.tvaultPath, metadata.Offset, metadata.Length)
		if err != nil {
			// Log error but don't fail the entire operation since DB is already updated
			log("Warning: Failed to securely overwrite data for file %s (ID: %d): %v\n",
				metadata.Name, metadata.ID, err)
		}
	}

	return nil
}

var errDeleteFolders = errors.New("error when deleting folders")
func (s *service) DeleteFolders(folderIDs []int64) error {
	if len(folderIDs) == 0 {
		log("no folder IDs provided for deletion")
		return errDeleteFolders
	}

	// First, get all file IDs in the selected folders
	fileIDs, err := s.getFileIDsInFolders(folderIDs)
	if err != nil {
		log("failed to get file IDs in folders: %w", err)
		return errDeleteFolders
	}

	// Delete all files using the existing DeleteFiles method
	if len(fileIDs) > 0 {
		err = s.DeleteFiles(fileIDs)
		if err != nil {
			log("failed to delete files in folders: %w", err)
			return errDeleteFolders
		}
	}

	// Now delete the empty folders
	tx, err := s.db.Begin()
	if err != nil {
		log("failed to begin transaction: %w", err)
		return errDeleteFolders
	}
	defer tx.Rollback()

	for _, folderID := range folderIDs {
		_, err := tx.Exec("DELETE FROM folders WHERE id = ?", folderID)
		if err != nil {
			log("failed to delete folder %d: %w", folderID, err)
			return errDeleteFolders
		}
	}

	if err := tx.Commit(); err != nil {
		log("failed to commit folder deletion: %w", err)
		return errDeleteFolders
	}

	return nil
}

// Helper method to get all file IDs in the specified folders
var errGetFileIDsFolders = errors.New("error when getting files")
func (s *service) getFileIDsInFolders(folderIDs []int64) ([]int64, error) {
	if len(folderIDs) == 0 {
		return nil, nil
	}

	filesInFolderQuery := `
	SELECT id FROM files 
	WHERE folder_id = ? AND is_deleted = 0
	`

	// NOTE: we iteratively execute the static sql query to eliminate SQLi risk from dynamic query construction
	// TODO (2026-02-09): gather up all of these queries and execute in a batch / transaction?
	var fileIDs []int64
	allRows := make([]*sql.Rows, len(folderIDs))
	for i, folderID := range folderIDs {
		// Query creates a prepared stmt under the hood
		rows, err := s.db.Query(filesInFolderQuery, folderID)
		allRows[i] = rows
		if err != nil {
			log("failed to query file IDs: %w", err)
			return nil, errGetFileIDsFolders
		}
		defer allRows[i].Close()

		for allRows[i].Next() {
			var fileID int64
			if err := allRows[i].Scan(&fileID); err != nil {
				log("failed to scan file ID: %w", err)
				return nil, errGetFileIDsFolders
			}
			fileIDs = append(fileIDs, fileID)
		}

		if err := allRows[i].Err(); err != nil {
			log("error iterating file IDs: %w", err)
			return nil, errGetFileIDsFolders
		}
	}

	return fileIDs, nil
}
