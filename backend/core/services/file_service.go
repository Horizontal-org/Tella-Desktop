package services

import (
    "Tella-Desktop/backend/core/models"
    "io"
    "os"
    "path/filepath"
    "github.com/google/uuid"
    "github.com/wailsapp/wails/v2/pkg/runtime"
    "context"
)

type FileService struct {
    ctx       context.Context
    uploadDir string
    transfers map[string]*models.FileTransfer
}

func NewFileService(ctx context.Context) *FileService {
    uploadDir := getUploadDirectory()
    return &FileService{
        ctx:       ctx,
        uploadDir: uploadDir,
        transfers: make(map[string]*models.FileTransfer),
    }
}

func getUploadDirectory() string {
    homeDir, err := os.UserHomeDir()
    if err != nil {
        return filepath.Join(".", "uploads")
    }
    uploadsDir := filepath.Join(homeDir, "Documents", "TellaUploads")
    os.MkdirAll(uploadsDir, 0755)
    return uploadsDir
}

func (s *FileService) PrepareUpload(request *models.PrepareUploadRequest) (*models.PrepareUploadResponse, error) {
    sessionId := uuid.New().String()
    fileTokens := make(map[string]string)

    for fileId, fileInfo := range request.Files {
        token := uuid.New().String()
        fileTokens[fileId] = token

        s.transfers[fileId] = &models.FileTransfer{
            ID:        fileId,
            SessionID: sessionId,
            Token:     token,
            FileName:  fileInfo.FileName,
            Size:      fileInfo.Size,
            FileType:  fileInfo.FileType,
            Status:    "preparing",
        }
    }

    return &models.PrepareUploadResponse{
        SessionID: sessionId,
        Files:    fileTokens,
    }, nil
}

func (s *FileService) SaveFile(sessionId string, fileId string, token string, fileName string, reader io.Reader) error {
    filePath := filepath.Join(s.uploadDir, fileName)
    
    // Create destination file
    dst, err := os.Create(filePath)
    if err != nil {
        return err
    }
    defer dst.Close()

    // Copy file data
    if _, err := io.Copy(dst, reader); err != nil {
        return err
    }

    // Notify UI about received file
    runtime.EventsEmit(s.ctx, "file-received", fileName)
    
    return nil
}

func (s *FileService) ValidateTransfer(sessionId string, fileId string, token string) bool {
    transfer, exists := s.transfers[fileId]
    if !exists {
        return false
    }
    return transfer.SessionID == sessionId && transfer.Token == token
}
