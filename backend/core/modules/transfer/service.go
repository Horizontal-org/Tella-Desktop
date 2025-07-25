package transfer

import (
	"Tella-Desktop/backend/utils/transferutils"
	"context"
	"database/sql"
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
	db               *sql.DB
}

type PendingTransfer struct {
	SessionID    string     `json:"sessionId"`
	Title        string     `json:"title"`
	Files        []FileInfo `json:"files"`
	ResponseChan chan *PrepareUploadResponse
	ErrorChan    chan error
	CreatedAt    time.Time
}

type TransferSession struct {
	SessionID string
	FolderID  int64
	Title     string
	Files     map[string]*Transfer
}

func NewService(ctx context.Context, fileSerservice filestore.Service, db *sql.DB) Service {
	return &service{
		ctx:              ctx,
		transfers:        sync.Map{},
		pendingTransfers: sync.Map{},
		fileService:      fileSerservice,
		db:               db,
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

	folderID, err := s.createTransferFolder(pendingTransfer.Title)
	if err != nil {
		return fmt.Errorf("failed to create transfer folder: %w", err)
	}

	transferSession := &TransferSession{
		SessionID: sessionID,
		FolderID:  folderID,
		Title:     pendingTransfer.Title,
		Files:     make(map[string]*Transfer),
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

	s.transfers.Store(sessionID+"_session", transferSession)

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

	actualFolderID := folderID
	if sessionValue, exists := s.transfers.Load(sessionID + "_session"); exists {
		if session, ok := sessionValue.(*TransferSession); ok {
			actualFolderID = session.FolderID
		}
	}

	runtime.EventsEmit(s.ctx, "file-receiving", map[string]interface{}{
		"sessionId": sessionID,
		"fileId":    fileID,
		"fileName":  fileName,
		"fileSize":  transfer.FileInfo.Size,
	})

	metadata, err := s.fileService.StoreFile(actualFolderID, fileName, mimeType, reader)
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

	runtime.LogInfo(s.ctx, fmt.Sprintf("File stored successfully in folder %d. ID: %s, Name: %s", actualFolderID, metadata.UUID, metadata.Name))
	return nil
}

func (s *service) calculateTotalSize(files []FileInfo) int64 {
	var total int64
	for _, file := range files {
		total += file.Size
	}
	return total
}

func (s *service) createTransferFolder(title string) (int64, error) {
	tx, err := s.db.Begin()
	if err != nil {
		return 0, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Create folder with the transfer title
	result, err := tx.Exec(`
		INSERT INTO folders (name, parent_id, created_at, updated_at) 
		VALUES (?, NULL, datetime('now'), datetime('now'))
	`, title)
	if err != nil {
		return 0, fmt.Errorf("failed to create folder: %w", err)
	}

	folderID, err := result.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("failed to get folder ID: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return 0, fmt.Errorf("failed to commit transaction: %w", err)
	}

	runtime.LogInfo(s.ctx, fmt.Sprintf("Created transfer folder '%s' with ID: %d", title, folderID))
	return folderID, nil
}
