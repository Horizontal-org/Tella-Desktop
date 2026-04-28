package transferutils

import (
	"errors"
)

var (
	ErrTransferNotFound         = errors.New("transfer not found")
	ErrInvalidSession           = errors.New("invalid session")
	ErrInvalidTransmission      = errors.New("invalid transmission")
	ErrTransferComplete         = errors.New("transfer already completed")
	ErrTransferRejected         = errors.New("transfer rejected")
	ErrTransferTooLarge         = errors.New("content too large")
	ErrTransferInsufficentSpace = errors.New("Insufficient storage space")
	ErrTransferHashMismatch			= errors.New("File hash mismatch")
)

// TODO cblgh(2026-02-12): actually implement validation
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
