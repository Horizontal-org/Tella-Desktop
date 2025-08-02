package filestore

import "io"

type Service interface {
	// StoreFile encrypts and stores a file in TVault, returning its metadata
	StoreFile(folderID int64, fileName string, mimeType string, reader io.Reader) (*FileMetadata, error)

	// GetStoredFolders returns a list of folders with file counts
	GetStoredFolders() ([]FolderInfo, error)

	// GetFilesInFolder returns files in a specific folder
	GetFilesInFolder(folderID int64) (*FilesInFolderResponse, error)

	// ExportFile exports a file by its ID to the user's downloads directory
	ExportFiles(ids []int64) ([]string, error)

	// ExportZipFolders exports files as ZIP archives
	ExportZipFolders(folderIDs []int64, selectedFileIDs []int64) ([]string, error)

	// DeleteFiles securely deletes files by their IDs
	DeleteFiles(ids []int64) error

	// DeleteFolders deletes folders and all their files by reusing DeleteFiles
	DeleteFolders(folderIDs []int64) error
}
