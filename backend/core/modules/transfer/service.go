package transfer

import (
	"Tella-Desktop/backend/utils/transferutils"
	"context"
	"fmt"
	"io"
	"sync"

	"Tella-Desktop/backend/core/modules/filestore"

	"github.com/google/uuid"
)

type service struct {
	ctx         context.Context
	transfers   sync.Map
	fileService filestore.Service
}

func NewService(ctx context.Context, fileSerservice filestore.Service) Service {
	return &service{
		ctx:         ctx,
		transfers:   sync.Map{},
		fileService: fileSerservice,
	}
}

func (s *service) PrepareUpload(request *PrepareUploadRequest) (*PrepareUploadResponse, error) {
	var responseFiles []FileTransmissionInfo

	for _, fileInfo := range request.Files {
		transmissionID := uuid.New().String()
		transfer := &Transfer{
			ID:        transmissionID,
			SessionID: request.SessionID,
			FileInfo:  fileInfo,
		}
		s.transfers.Store(fileInfo.ID, transfer)
		responseFiles = append(responseFiles, FileTransmissionInfo{
			ID:             fileInfo.ID,
			TransmissionID: transmissionID,
		})
	}

	return &PrepareUploadResponse{
		Files: responseFiles,
	}, nil
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

	metadata, err := s.fileService.StoreFile(folderID, fileName, mimeType, reader)
	if err != nil {
		transfer.Status = "failed"
		s.transfers.Store(fileID, transfer)
		return fmt.Errorf("failed to store file: %w", err)
	}

	transfer.Status = "completed"
	s.transfers.Store(fileID, transfer)

	fmt.Printf("File stored successfully. ID: %s, Name: %s", metadata.UUID, metadata.Name)
	return nil
}
