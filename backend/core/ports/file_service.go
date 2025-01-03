package ports

import (
    "Tella-Desktop/backend/core/models"
    "io"
)

type FileService interface {
    PrepareUpload(request *models.PrepareUploadRequest) (*models.PrepareUploadResponse, error)
    SaveFile(sessionId string, fileId string, token string, fileName string, reader io.Reader) error
    ValidateTransfer(sessionId string, fileId string, token string) bool
}