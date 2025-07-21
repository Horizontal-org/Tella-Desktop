package database

import (
	"Tella-Desktop/backend/utils/authutils"
	"database/sql"
	"embed"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"

	_ "github.com/mutecomm/go-sqlcipher/v4"
)

//go:embed migrations/*.sql
var migrationsFS embed.FS

type DB struct {
	*sql.DB
}

// Initialize creates a new database connection and runs migrations
func Initialize(dbPath string, key []byte) (*DB, error) {
	// Ensure directory exists
	dbDir := filepath.Dir(dbPath)
	if err := os.MkdirAll(dbDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create database directory: %v", err)
	}

	// Convert the key to hex string
	hexKey := hex.EncodeToString(key)

	// Use the DSN format recommended by go-sqlcipher
	connStr := fmt.Sprintf("%s?_pragma_key=x'%s'&_pragma_cipher_page_size=4096&_pragma_kdf_iter=64000&_pragma_cipher_hmac_algorithm=HMAC_SHA512&_pragma_cipher_compatibility=3", dbPath, hexKey)

	db, err := sql.Open("sqlite3", connStr)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %v", err)
	}

	db.SetMaxOpenConns(1)
	db.SetMaxIdleConns(1)
	db.SetConnMaxLifetime(0)

	_, err = db.Exec("PRAGMA busy_timeout = 30000")
	if err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to set busy timeout: %v", err)
	}

	_, err = db.Exec("PRAGMA journal_mode = WAL")
	if err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to set WAL mode: %v", err)
	}

	// Verify we can read the database
	var count int
	err = db.QueryRow("SELECT count(*) FROM sqlite_master").Scan(&count)
	if err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to verify database decryption: %v", err)
	}

	// Run migrations
	if err := runMigrations(db); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to run migrations: %v", err)
	}

	return &DB{db}, nil
}

func runMigrations(db *sql.DB) error {
	// Read migration files
	files, err := migrationsFS.ReadDir("migrations")
	if err != nil {
		return fmt.Errorf("failed to read migrations directory: %v", err)
	}

	// Begin transaction
	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %v", err)
	}
	defer tx.Rollback()

	// Execute each migration file
	for _, file := range files {
		if filepath.Ext(file.Name()) != ".sql" {
			continue
		}

		content, err := migrationsFS.ReadFile(filepath.Join("migrations", file.Name()))
		if err != nil {
			return fmt.Errorf("failed to read migration file %s: %v", file.Name(), err)
		}

		if _, err := tx.Exec(string(content)); err != nil {
			return fmt.Errorf("failed to execute migration %s: %v", file.Name(), err)
		}
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %v", err)
	}

	return nil
}

// GetDatabasePath returns the path where the database should be stored
func GetDatabasePath() string {
	return authutils.GetDatabasePath()
}
