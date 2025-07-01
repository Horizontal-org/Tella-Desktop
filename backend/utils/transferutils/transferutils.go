package transferutils

import (
	"errors"
)

var (
	ErrTransferNotFound    = errors.New("transfer not found")
	ErrInvalidSession      = errors.New("invalid session")
	ErrInvalidTransmission = errors.New("invalid transmission")
	ErrTransferComplete    = errors.New("transfer already completed")
)

func ValidateUploadRequest(sessionID, transmissionID string, fileID string) error {
	if sessionID == "" {
		return errors.New("sessionId is required")
	}
	if transmissionID == "" {
		return errors.New("transmissionId is required")
	}
	if fileID == "" {
		return errors.New("fileId is required")
	}
	return nil
}
