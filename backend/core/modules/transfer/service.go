package transfer

import (
	"context"

	"github.com/google/uuid"
	"github.com/wailsapp/wails/v2/pkg/runtime"
)

type service struct {
	ctx       context.Context
	transfers map[string]*Transfer
	sessions  map[string]string
}

func NewService(ctx context.Context) Service {
	return &service{
		ctx:       ctx,
		transfers: make(map[string]*Transfer),
		sessions:  make(map[string]string),
	}
}

func (s *service) PrepareUpload(request *PrepareUploadRequest) (*PrepareUploadResponse, error) {
	sessionId := uuid.New().String()
	fileTokens := make(map[string]string)

	// Store sender information
	s.sessions[sessionId] = request.Info.Fingerprint

	// Process each file in the request
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

	runtime.LogInfo(s.ctx, "Prepared upload session: "+sessionId)

	return &PrepareUploadResponse{
		SessionID: sessionId,
		Files:     fileTokens,
	}, nil
}

func (s *service) ValidateTransfer(sessionId, fileId, token string) bool {
	transfer, exists := s.transfers[fileId]
	if !exists {
		return false
	}

	if transfer.Status != "preparing" && transfer.Status != "in_progress" {
		// Transfer already completed or failed
		return false
	}

	// Validate session and token
	valid := transfer.SessionID == sessionId && transfer.Token == token

	if valid {
		// Update status to in_progress
		transfer.Status = "in_progress"
	}

	return valid
}

func (s *service) CompleteTransfer(sessionId, fileId string) error {
	transfer, exists := s.transfers[fileId]
	if !exists {
		return nil
	}

	if transfer.SessionID != sessionId {
		return nil
	}

	// Update status to completed
	transfer.Status = "completed"

	runtime.LogInfo(s.ctx, "Completed transfer for file: "+fileId)

	return nil
}

func (s *service) GetTransferDetails(sessionId, fileId string) (*Transfer, error) {
	transfer, exists := s.transfers[fileId]
	if !exists {
		return nil, nil
	}

	if transfer.SessionID != sessionId {
		return nil, nil
	}

	return transfer, nil
}
