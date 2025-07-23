package transfer

import (
	"Tella-Desktop/backend/utils/transferutils"
	"context"
	"fmt"
	"io"
	"sync"
	"time"

	"Tella-Desktop/backend/core/modules/filestore"

	"github.com/google/uuid"
	"github.com/wailsapp/wails/v2/pkg/runtime"
)

type service struct {
	ctx              context.Context
	transfers        sync.Map
	pendingTransfers sync.Map
	fileService      filestore.Service
}

type PendingTransfer struct {
	SessionID    string     `json:"sessionId"`
	Title        string     `json:"title"`
	Files        []FileInfo `json:"files"`
	ResponseChan chan *PrepareUploadResponse
	ErrorChan    chan error
	CreatedAt    time.Time
}

func NewService(ctx context.Context, fileSerservice filestore.Service) Service {
	return &service{
		ctx:              ctx,
		transfers:        sync.Map{},
		pendingTransfers: sync.Map{},
		fileService:      fileSerservice,
	}
}

func (s *service) PrepareUpload(request *PrepareUploadRequest) (*PrepareUploadResponse, error) {
	pendingTransfer := &PendingTransfer{
		SessionID:    request.SessionID,
		Title:        request.Title,
		Files:        request.Files,
		ResponseChan: make(chan *PrepareUploadResponse, 1),
		ErrorChan:    make(chan error, 1),
		CreatedAt:    time.Now(),
	}

	s.pendingTransfers.Store(request.SessionID, pendingTransfer)

	runtime.EventsEmit(s.ctx, "prepare-upload-request", map[string]interface{}{
		"sessionId":  request.SessionID,
		"title":      request.Title,
		"files":      request.Files,
		"totalFiles": len(request.Files),
		"totalSize":  s.calculateTotalSize(request.Files),
	})

	select {
	case response := <-pendingTransfer.ResponseChan:
		s.pendingTransfers.Delete(request.SessionID)
		return response, nil
	case err := <-pendingTransfer.ErrorChan:
		s.pendingTransfers.Delete(request.SessionID)
		return nil, err
	case <-time.After(5 * time.Minute):
		s.pendingTransfers.Delete(request.SessionID)
		return nil, fmt.Errorf("request timeout - no response from recipient")
	}
}

func (s *service) AcceptTransfer(sessionID string) error {
	value, exists := s.pendingTransfers.Load(sessionID)
	if !exists {
		return fmt.Errorf("no pending transfer found for session: %s", sessionID)
	}

	pendingTransfer, ok := value.(*PendingTransfer)
	if !ok {
		return fmt.Errorf("invalid pending transfer data")
	}

	var responseFiles []FileTransmissionInfo
	for _, fileInfo := range pendingTransfer.Files {
		transmissionID := uuid.New().String()
		transfer := &Transfer{
			ID:        transmissionID,
			SessionID: sessionID,
			FileInfo:  fileInfo,
			Status:    "pending",
		}
		s.transfers.Store(fileInfo.ID, transfer)
		responseFiles = append(responseFiles, FileTransmissionInfo{
			ID:             fileInfo.ID,
			TransmissionID: transmissionID,
		})
	}

	response := &PrepareUploadResponse{
		Files: responseFiles,
	}

	select {
	case pendingTransfer.ResponseChan <- response:
		runtime.LogInfo(s.ctx, fmt.Sprintf("Transfer accepted for session: %s", sessionID))
		return nil
	default:
		return fmt.Errorf("failed to send acceptance response")
	}
}

func (s *service) RejectTransfer(sessionID string) error {
	value, exists := s.pendingTransfers.Load(sessionID)
	if !exists {
		return fmt.Errorf("no pending transfer found for session: %s", sessionID)
	}

	pendingTransfer, ok := value.(*PendingTransfer)
	if !ok {
		return fmt.Errorf("invalid pending transfer data")
	}

	select {
	case pendingTransfer.ErrorChan <- fmt.Errorf("transfer rejected by recipient"):
		runtime.LogInfo(s.ctx, fmt.Sprintf("Transfer rejected for session: %s", sessionID))
		return nil
	default:
		return fmt.Errorf("failed to send rejection response")
	}
}

func (s *service) GetTransfer(fileID string) (*Transfer, error) {
	if value, ok := s.transfers.Load(fileID); ok {
		if transfers, ok := value.(*Transfer); ok {
			return transfers, nil
		}
	}
	return nil, transferutils.ErrTransferNotFound
}

func (s *service) HandleUpload(sessionID, transmissionID, fileID string, reader io.Reader, fileName string, mimeType string, folderID int64) error {
	transfer, err := s.GetTransfer(fileID)
	if err != nil {
		return err
	}

	if transfer.SessionID != sessionID {
		return transferutils.ErrInvalidSession
	}

	if transfer.Status == "completed" {
		return transferutils.ErrTransferComplete
	}

	runtime.EventsEmit(s.ctx, "file-receiving", map[string]interface{}{
		"sessionId": sessionID,
		"fileId":    fileID,
		"fileName":  fileName,
		"fileSize":  transfer.FileInfo.Size,
	})

	metadata, err := s.fileService.StoreFile(folderID, fileName, mimeType, reader)
	if err != nil {
		transfer.Status = "failed"
		s.transfers.Store(fileID, transfer)
		return fmt.Errorf("failed to store file: %w", err)
	}

	transfer.Status = "completed"
	s.transfers.Store(fileID, transfer)

	runtime.EventsEmit(s.ctx, "file-received", map[string]interface{}{
		"sessionId": sessionID,
		"fileId":    fileID,
		"fileName":  fileName,
		"fileSize":  transfer.FileInfo.Size,
	})

	fmt.Printf("File stored successfully. ID: %s, Name: %s", metadata.UUID, metadata.Name)
	return nil
}

func (s *service) calculateTotalSize(files []FileInfo) int64 {
	var total int64
	for _, file := range files {
		total += file.Size
	}
	return total
}
