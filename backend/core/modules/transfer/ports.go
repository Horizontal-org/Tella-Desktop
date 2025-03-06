package transfer

type Service interface {
	PrepareUpload(request *PrepareUploadRequest) (*PrepareUploadResponse, error)
	ValidateTransfer(sessionId, fileId, token string) bool
	CompleteTransfer(sessionId, fileId string) error
	GetTransferDetails(sessionId, fileId string) (*Transfer, error)
}
