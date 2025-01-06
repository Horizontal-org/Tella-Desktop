package transfer

import "io"

type Service interface {
    PrepareUpload(request *PrepareUploadRequest) (*PrepareUploadResponse, error)
    SaveFile(sessionId, fileId, token, fileName string, reader io.Reader) error
    ValidateTransfer(sessionId, fileId, token string) bool
}