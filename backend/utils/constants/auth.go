package constants

import "errors"

// Authentication constants
const (
	KeyLength        = 32
	SaltLength       = 32
	Iterations       = 10000
	TVaultHeaderSize = 256
)

// Authentication errors
var (
	ErrInvalidPassword  = errors.New("invalid password")
	ErrTVaultNotFound   = errors.New("tvault file not found")
	ErrDatabaseNotFound = errors.New("database file not found")
	ErrCorruptedTVault  = errors.New("corrupted tvault header")
	ErrPasswordTooShort = errors.New("password must be at least 6 characters")
)
