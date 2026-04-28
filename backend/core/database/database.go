package database

import (
	"Tella-Desktop/backend/utils/authutils"
	util "Tella-Desktop/backend/utils/genericutil"
	"Tella-Desktop/backend/utils/devlog"
	"database/sql"
	"errors"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"

	_ "github.com/mutecomm/go-sqlcipher/v4"
)

var log = devlog.Logger("db")

type DB struct {
	*sql.DB
}

var initFailed = errors.New("initialization failed")
// Initialize creates a new database connection and runs migrations
func Initialize(dbPath string, key []byte) (*DB, error) {
	// Ensure directory exists
	dbDir := filepath.Dir(dbPath)
	if err := os.MkdirAll(dbDir, util.USER_ONLY_DIR_PERMS); err != nil {
		log("failed to create database directory: %v", err)
		return nil, initFailed
	}

	// Convert the key to hex string
	hexKey := hex.EncodeToString(key)
	// Use the DSN format recommended by go-sqlcipher
	connStr := fmt.Sprintf("%s?_pragma_key=x'%s'&_pragma_cipher_page_size=4096&_pragma_kdf_iter=64000&_pragma_cipher_hmac_algorithm=HMAC_SHA512&_pragma_cipher_compatibility=3", dbPath, hexKey)

	db, err := sql.Open("sqlite3", connStr)
	if err != nil {
		log("failed to open database: %v", err)
		return nil, initFailed
	}

	db.SetMaxOpenConns(1)
	db.SetMaxIdleConns(1)
	db.SetConnMaxLifetime(0)

	_, err = db.Exec("PRAGMA busy_timeout = 30000")
	if err != nil {
		db.Close()
		log("failed to set busy timeout: %v", err)
		return nil, initFailed
	}

	_, err = db.Exec("PRAGMA journal_mode = WAL")
	if err != nil {
		db.Close()
		log("failed to set WAL mode: %v", err)
		return nil, initFailed
	}

	// Verify we can read the database
	var count int
	err = db.QueryRow("SELECT count(*) FROM sqlite_master").Scan(&count)
	if err != nil {
		db.Close()
		log("failed to verify database decryption: %v", err)
		return nil, initFailed
	}

	// Run migrations
	if err := runMigrations(db); err != nil {
		db.Close()
		log("failed to run migrations: %v", err)
		return nil, initFailed
	}

	return &DB{db}, nil
}

var errFailedMigration = errors.New("failed to run migration")
func runMigrations(db *sql.DB) error {
	// Begin transaction
	tx, err := db.Begin()
	if err != nil {
		log("failed to begin transaction: %v", err)
		return errFailedMigration
	}
	defer tx.Rollback()
	for _, migration := range getMigrations() {
		if _, err := tx.Exec(string(migration.Content)); err != nil {
			log("failed to execute migration %s: %v", migration.Name, err)
			return errFailedMigration
		}
	}
	// Commit transaction
	if err := tx.Commit(); err != nil {
		log("failed to commit transaction: %v", err)
		return errFailedMigration
	}

	return nil
}

// GetDatabasePath returns the path where the database should be stored
func GetDatabasePath() string {
	return authutils.GetDatabasePath()
}
