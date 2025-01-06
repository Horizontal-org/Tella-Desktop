package transfer

import (
    "io"
    "os"
    "path/filepath"
    "github.com/google/uuid"
    "github.com/wailsapp/wails/v2/pkg/runtime"
    "context"
)

type service struct {
    ctx       context.Context
    uploadDir string
    transfers map[string]*Transfer
}

func NewService(ctx context.Context) Service {
    return &service{
        ctx:       ctx,
        uploadDir: getUploadDirectory(),
        transfers: make(map[string]*Transfer),
    }
}

func (s *service) PrepareUpload(request *PrepareUploadRequest) (*PrepareUploadResponse, error) {
    sessionId := uuid.New().String()
    fileTokens := make(map[string]string)

    for fileId, fileInfo := range request.Files {
        token := uuid.New().String()
        fileTokens[fileId] = token

        s.transfers[fileId] = &Transfer{
            ID:        fileId,
            SessionID: sessionId,
            Token:     token,
            FileName:  fileInfo.FileName,
            Size:      fileInfo.Size,
            FileType:  fileInfo.FileType,
            Status:    "preparing",
        }
    }

    return &PrepareUploadResponse{
        SessionID: sessionId,
        Files:    fileTokens,
    }, nil
}

func (s *service) SaveFile(sessionId, fileId, token, fileName string, reader io.Reader) error {
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

func (s *service) ValidateTransfer(sessionId, fileId, token string) bool {
    transfer, exists := s.transfers[fileId]
    if !exists {
        return false
    }
    return transfer.SessionID == sessionId && transfer.Token == token
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
