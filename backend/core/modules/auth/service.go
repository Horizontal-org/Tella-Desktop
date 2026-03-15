package auth

import (
	"context"
	"crypto/rand"
	"errors"
	"os"
	"path/filepath"

	"Tella-Desktop/backend/utils/authutils"
	"Tella-Desktop/backend/utils/constants"
	util "Tella-Desktop/backend/utils/genericutil"
	"Tella-Desktop/backend/utils/devlog"

	"github.com/matthewhartstonge/argon2"
)

var log = devlog.Logger("auth")

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

var initFailed = errors.New("initialization failed")
func (s *service) Initialize(ctx context.Context) error {
	s.ctx = ctx

	// create directory if they don't exists
	vaultDir := filepath.Dir(s.tvaultPath)
	if err := os.MkdirAll(vaultDir, util.USER_ONLY_DIR_PERMS); err != nil {
		log("failed to create vault directory: %w", err)
		return initFailed
	}

	// create tmp directory for decrypted files
	tempDir := authutils.GetTempDir()
	if err := os.MkdirAll(tempDir, util.USER_ONLY_DIR_PERMS); err != nil {
		log("failed to create temp directory: %w", err)
		return initFailed
	}

	log("Auth service initialized")
	return nil
}

func (s *service) IsFirstTimeSetup() bool {
	// Check if the tvault file exists
	_, err := os.Stat(s.tvaultPath)
	return os.IsNotExist(err)
}

var errCreatePassword = errors.New("create password failed")
func (s *service) CreatePassword(password string) error {
	if len(password) < constants.PasswordMinLength {
		return constants.ErrPasswordTooShort
	}

	// basic input invalidation to prevent attacks that overflow memory somehow
	if len(password) > constants.PasswordMaxLength {
		return constants.ErrPasswordTooLong
	}

	//generate random database key | TODO: move this outside of this function
	dbKey := make([]byte, constants.KeyLength)
	if _, err := rand.Read(dbKey); err != nil {
		log("failed to generate database key: %w", err)
		return errCreatePassword
	}

	//generate random salt
	salt := make([]byte, constants.SaltLength)
	if _, err := rand.Read(salt); err != nil {
		log("failed to generate salt: %w", err)
		return errCreatePassword
	}

	config := argon2.MemoryConstrainedDefaults()

	raw, err := config.HashRaw([]byte(password))
	defer argon2.SecureZeroMemory(raw.Hash)
	if err != nil {
		log("failed to hash password: %w", err)
		return errCreatePassword
	}

	encryptedDBKey, err := authutils.EncryptData(dbKey, raw.Hash)
	if err != nil {
		log("failed to encrypt database key: %w", err)
		return errCreatePassword
	}

	if err := authutils.InitializeTVaultHeader(raw.Salt, encryptedDBKey); err != nil {
		log("failed to initialize tvault header: %w", err)
		return errCreatePassword
	}

	// Store database key in memory
	s.databaseKey = dbKey
	s.isUnlocked = true

	log("Password created successfully")
	return nil
}

var errDecryptDatabase = errors.New("failed to decrypt database")
func (s *service) DecryptDatabaseKey(password string) error {
	log("Verifying password")

	// basic input invalidation to prevent attacks that overflow memory somehow
	if len(password) > constants.PasswordMaxLength {
		return constants.ErrPasswordTooLong
	}

	salt, encryptedDBKey, err := authutils.ReadTVaultHeader()
	if err != nil {
		log("error reading tvault header %v", err)
		return errDecryptDatabase
	}

	config := argon2.MemoryConstrainedDefaults()

	raw, err := config.Hash([]byte(password), salt)
	defer argon2.SecureZeroMemory(raw.Hash)
	if err != nil {
		log("failed to derive key: %w", err)
		return errDecryptDatabase
	}

	dbKey, err := authutils.DecryptData(encryptedDBKey, raw.Hash)
	if err != nil {
		log("Invalid password")
		return constants.ErrInvalidPassword
	}

	s.databaseKey = dbKey
	s.isUnlocked = true

	log("Password verified successfully")
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
	log("Session cleared")
}
