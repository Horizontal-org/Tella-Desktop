package transfer

import "errors"

type Transfer struct {
	TransmissionID string   `json:"transmissionId"`
	SessionID      string   `json:"sessionId"`
	FileInfo       FileInfo `json:"fileInfo"`
	Status         string   `json:"status"`
}

type FileInfo struct {
	ID       string `json:"id"`
	FileName string `json:"fileName"`
	Size     int64  `json:"size"`
	FileType string `json:"fileType"`
	SHA256   string `json:"sha256"`
}

type PrepareUploadRequest struct {
	Title     string     `json:"title"`
	SessionID string     `json:"sessionId"`
	Nonce     string     `json:"nonce"`
	Files     []FileInfo `json:"files"`
}

type PrepareUploadResponse struct {
	Files []FileTransmissionInfo `json:"files"`
}

type FileTransmissionInfo struct {
	ID             string `json:"id"`
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
	if len(r.Files) == 0 {
		return errors.New("at least one file is required")
	}
	// TODO cblgh(2026-03-09): improve validation logic
	// go through each file and make sure it has fields
	//
	// "id": "file-uuid",
	// "fileName": "",
	// "size": number,
	// "sha256": "",
	// "fileType": "application/pdf",
	// "thumbnail": "thumbnail-data" (optional?)
	return nil
}
