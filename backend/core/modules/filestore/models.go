package filestore

import "time"

type FilesInFolderResponse struct {
	FolderName string     `json:"folderName"`
	Files      []FileInfo `json:"files"`
}

type FileInfo struct {
	ID        int64  `json:"id"`
	Name      string `json:"name"`
	MimeType  string `json:"mimeType"`
	Timestamp string `json:"timestamp"`
	Size      int64  `json:"size"`
}

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

type FolderInfo struct {
	ID        int64  `json:"id"`
	Name      string `json:"name"`
	Timestamp string `json:"timestamp"`
	FileCount int    `json:"fileCount"`
}
