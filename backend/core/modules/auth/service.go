package auth

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"Tella-Desktop/backend/utils/authutils"
	"Tella-Desktop/backend/utils/constants"

	"github.com/wailsapp/wails/v2/pkg/runtime"
	"golang.org/x/crypto/pbkdf2"
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

	//derive key from password and salt
	derivedKey := pbkdf2.Key([]byte(password), salt, constants.Iterations, constants.KeyLength, sha256.New)

	//encrypt database key using derived key
	encryptedDBKey, err := authutils.EncryptData(dbKey, derivedKey)
	if err != nil {
		return fmt.Errorf("failed to encrypt database key: %w", err)
	}

	//create and write tvault header
	if err := authutils.InitializeTVaultHeader(salt, encryptedDBKey); err != nil {
		return fmt.Errorf("failed to write tvault header: %w", err)
	}

	// Store database key in memory
	s.databaseKey = dbKey
	s.isUnlocked = true

	runtime.LogInfo(s.ctx, "Password created successfully")
	return nil
}

func (s *service) DecryptDatabaseKey(password string) error {
	runtime.LogInfo(s.ctx, "Verifying password")

	// Read the salt and encrypted database key from tvault
	salt, encryptedDBKey, err := authutils.ReadTVaultHeader()
	if err != nil {
		return err
	}

	// Derive key from password and stored salt
	derivedKey := pbkdf2.Key([]byte(password), salt, constants.Iterations, constants.KeyLength, sha256.New)

	// Decrypt database key
	dbKey, err := authutils.DecryptData(encryptedDBKey, derivedKey)
	if err != nil {
		runtime.LogInfo(s.ctx, "Invalid password")
		return constants.ErrInvalidPassword
	}

	// Store database key in memory
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
