package transfer

import "io"

type Service interface {
	PrepareUpload(request *PrepareUploadRequest) (*PrepareUploadResponse, error)
	AcceptTransfer(sessionID string) error
	RejectTransfer(sessionID string) error
	HandleUpload(sessionID, transmissionID, fileID string, reader io.Reader, fileName string, mimeType string, folderID int64) error
	GetTransfer(fileID string) (*Transfer, error)
}
