package filestore

import "io"

type Service interface {
	// StoreFile encrypts and stores a file in TVault, returning its metadata
	StoreFile(folderID int64, fileName string, mimeType string, reader io.Reader) (*FileMetadata, error)

	// GetStoredFiles returns a list of stored files
	GetStoredFiles() ([]FileInfo, error)
}
