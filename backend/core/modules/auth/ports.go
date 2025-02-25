package auth

import "context"

type Service interface {
	// Initialize creates necessary directories
	Initialize(ctx context.Context) error

	// IsFirstTimeSetup checks if this is the first launch
	IsFirstTimeSetup() bool

	// CreatePassword sets up encryption with a new password
	CreatePassword(password string) error

	// VerifyPassword validates the password and returns database key if valid
	VerifyPassword(password string) ([]byte, error)

	// GetDBKey returns the current database key (only if unlocked)
	GetDBKey() ([]byte, error)
}
