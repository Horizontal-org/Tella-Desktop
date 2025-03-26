package transfer

import "errors"

type Transfer struct {
	ID        string   `json:"id"`
	SessionID string   `json:"sessionId"`
	FileInfo  FileInfo `json:"fileInfo"`
	Status    string   `json:"status"`
}

type FileInfo struct {
	ID       string `json:"id"`
	FileName string `json:"fileName"`
	Size     int64  `json:"size"`
	FileType string `json:"fileType"`
	SHA256   string `json:"sha256"`
}

type PrepareUploadRequest struct {
	Title     string `json:"title"`
	SessionID string `json:"sessionId"`
	Metadata  struct {
		Files map[string]FileInfo `json:"files"`
	} `json:"metadata"`
}

type PrepareUploadResponse struct {
	TransmissionID string `json:"transmissionId"`
}

type UploadRequest struct {
	SessionID      string `json:"sessionId"`
	TransmissionID string `json:"transmissionId"`
	FileID         string `json:"fileId"`
	Data           []byte `json:"data"`
}

type UploadResponse struct {
	Success bool `json:"success"`
}

func (r *PrepareUploadRequest) Validate() error {
	if r.SessionID == "" {
		return errors.New("sessionId is required")
	}
	if len(r.Metadata.Files) == 0 {
		return errors.New("at least one file is required")
	}
	return nil
}
