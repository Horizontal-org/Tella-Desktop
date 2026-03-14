package constants

import (
	"errors"
)

// Authentication constants
const (
	LengthFieldSize      = 4
	KeyLength            = 32
	SaltLength           = 32
	TVaultHeaderSize     = 256
	CurrentTVaultVersion = 1
	PasswordMinLength    = 6
	PasswordMaxLength    = 1000
)

// Authentication errors
var (
	ErrInvalidPassword    = errors.New("invalid password")
	ErrTVaultNotFound     = errors.New("tvault file not found")
	ErrDatabaseNotFound   = errors.New("database file not found")
	ErrCorruptedTVault    = errors.New("corrupted tvault header")
	ErrPasswordTooShort   = errors.New("password must be at least 6 characters")
	ErrPasswordTooLong    = errors.New("password must not exceed 1000 characters")
	ErrHeaderTooLarge     = errors.New("tvault header too large")
	ErrUnsupportedVersion = errors.New("unsupported tvault version")
)
