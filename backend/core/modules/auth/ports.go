package auth

import "context"

type Service interface {
	// Initialize creates necessary directories
	Initialize(ctx context.Context) error

	// IsFirstTimeSetup checks if this is the first launch
	IsFirstTimeSetup() bool

	// CreatePassword sets up encryption with a new password
	CreatePassword(password string) error

	// DecryptDatabaseKey decrypts the database key with the given password
	DecryptDatabaseKey(password string) error

	// GetDBKey returns the current database key (only if unlocked)
	GetDBKey() ([]byte, error)

	// ClearSession clears the current authentication session
	ClearSession()
}
