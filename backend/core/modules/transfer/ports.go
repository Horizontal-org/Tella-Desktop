package transfer

import "io"

type Service interface {
	PrepareUpload(request *PrepareUploadRequest) (*PrepareUploadResponse, error)
	AcceptTransfer(sessionID string) error
	RejectTransfer(sessionID string) error
	CloseConnection(sessionID string) error
	HandleUpload(sessionID, transmissionID, fileID string, reader io.Reader, fileName string, mimeType string, folderID int64) error
	PreUploadValidation(sessionID, transmissionID, fileID string) error
	GetTransfer(fileID string) (*Transfer, error)
	StopTransfer(sessionID string)
	GetMaxFileSizeLimit() int64
	Lock()
}
