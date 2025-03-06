package filestore

type FileInfo struct {
	Name      string `json:"name"`
	MimeType  string `json:"mimeType"`
	Timestamp string `json:"timestamp"`
}
