package auth

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"Tella-Desktop/backend/utils/authutils"
	"Tella-Desktop/backend/utils/constants"

	"github.com/matthewhartstonge/argon2"
	"github.com/wailsapp/wails/v2/pkg/runtime"
)

type service struct {
	ctx          context.Context
	tvaultPath   string
	databasePath string
	databaseKey  []byte
	isUnlocked   bool
}

func NewService(ctx context.Context) Service {
	return &service{
		ctx:          ctx,
		tvaultPath:   authutils.GetTVaultPath(),
		databasePath: authutils.GetDatabasePath(),
		isUnlocked:   false,
	}
}

func (s *service) Initialize(ctx context.Context) error {
	s.ctx = ctx

	// create directory if they don't exists
	vaultDir := filepath.Dir(s.tvaultPath)
	if err := os.MkdirAll(vaultDir, 0755); err != nil {
		return fmt.Errorf("failed to create vault directory: %w", err)
	}

	// create tmp directory for decrypted files
	tempDir := authutils.GetTempDir()
	if err := os.MkdirAll(tempDir, 0755); err != nil {
		return fmt.Errorf("failed to create temp directory: %w", err)
	}

	runtime.LogInfo(ctx, "Auth service initialized")
	return nil
}

func (s *service) IsFirstTimeSetup() bool {
	// Check if the tvault file exists
	_, err := os.Stat(s.tvaultPath)
	return os.IsNotExist(err)
}

func (s *service) CreatePassword(password string) error {
	if len(password) < 6 {
		return constants.ErrPasswordTooShort
	}

	//generate random database key | TODO: move this outside of this function
	dbKey := make([]byte, constants.KeyLength)
	if _, err := rand.Read(dbKey); err != nil {
		return fmt.Errorf("failed to generate database key: %w", err)
	}

	//generate random salt
	salt := make([]byte, constants.SaltLength)
	if _, err := rand.Read(salt); err != nil {
		return fmt.Errorf("failed to generate salt: %w", err)
	}

	config := argon2.MemoryConstrainedDefaults()

	raw, err := config.HashRaw([]byte(password))
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}

	encryptedDBKey, err := authutils.EncryptData(dbKey, raw.Hash)
	if err != nil {
		argon2.SecureZeroMemory(raw.Hash)
		return fmt.Errorf("failed to encrypt database key: %w", err)
	}

	if err := authutils.InitializeTVaultHeader(raw.Salt, encryptedDBKey); err != nil {
		argon2.SecureZeroMemory(raw.Hash)
		return fmt.Errorf("failed to initialize tvault header: %w", err)
	}

	argon2.SecureZeroMemory(raw.Hash)

	// Store database key in memory
	s.databaseKey = dbKey
	s.isUnlocked = true

	runtime.LogInfo(s.ctx, "Password created successfully")
	return nil
}

func (s *service) DecryptDatabaseKey(password string) error {
	runtime.LogInfo(s.ctx, "Verifying password")

	salt, encryptedDBKey, err := authutils.ReadTVaultHeader()
	if err != nil {
		return err
	}

	config := argon2.MemoryConstrainedDefaults()

	raw, err := config.Hash([]byte(password), salt)
	if err != nil {
		return fmt.Errorf("failed to derive key: %w", err)
	}

	dbKey, err := authutils.DecryptData(encryptedDBKey, raw.Hash)
	if err != nil {
		argon2.SecureZeroMemory(raw.Hash)
		runtime.LogInfo(s.ctx, "Invalid password")
		return constants.ErrInvalidPassword
	}

	argon2.SecureZeroMemory(raw.Hash)

	s.databaseKey = dbKey
	s.isUnlocked = true

	runtime.LogInfo(s.ctx, "Password verified successfully")
	return nil
}

func (s *service) GetDBKey() ([]byte, error) {
	if !s.isUnlocked || s.databaseKey == nil {
		return nil, errors.New("database is locked")
	}
	return s.databaseKey, nil
}

func (s *service) ClearSession() {
	// Clear the database key from memory
	if s.databaseKey != nil {
		// Zero out the key for security
		for i := range s.databaseKey {
			s.databaseKey[i] = 0
		}
		s.databaseKey = nil
	}
	s.isUnlocked = false
	runtime.LogInfo(s.ctx, "Session cleared")
}
