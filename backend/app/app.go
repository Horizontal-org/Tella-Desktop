package app

import (
	"context"
	"database/sql"

	"Tella-Desktop/backend/core/database"
	"Tella-Desktop/backend/core/modules/auth"
	"Tella-Desktop/backend/core/modules/client"
	"Tella-Desktop/backend/core/modules/filestore"
	"Tella-Desktop/backend/core/modules/registration"
	"Tella-Desktop/backend/core/modules/server"
	"Tella-Desktop/backend/core/modules/transfer"
	"Tella-Desktop/backend/utils/authutils"
	"Tella-Desktop/backend/utils/network"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// App struct
type App struct {
	ctx                 context.Context
	db                  *database.DB
	authService         auth.Service
	registrationService registration.Service
	transferService     transfer.Service
	serverService       server.Service
	clientService       client.Service
	fileService         filestore.Service
	defaultFolderID     int64
}

func (a *App) RegisterWithDevice(ip string, port int) error {
	return a.clientService.RegisterWithDevice(ip, port)
}

func (a *App) SendTestFile(ip string, port int, pin string) error {
	return a.clientService.SendTestFile(ip, port, pin)
}

// Auth related methods to expose to frontend
func (a *App) IsFirstTimeSetup() bool {
	return a.authService.IsFirstTimeSetup()
}

func (a *App) CreatePassword(password string) error {
	err := a.authService.CreatePassword((password))
	if err != nil {
		return err
	}

	if err := a.initializeDatabase(); err != nil {
		runtime.LogError(a.ctx, "Failed to initialize database during setup: "+err.Error())
		return err
	}

	runtime.LogInfo(a.ctx, "Database created and encrypted successfully")
	return nil
}

func (a *App) VerifyPassword(password string) error {
	err := a.authService.DecryptDatabaseKey(password)

	if err != nil {
		return err
	}

	// Initialize database after successful password verification
	if a.db == nil {
		if err := a.initializeDatabase(); err != nil {
			runtime.LogError(a.ctx, "Failed to initialize database after login: "+err.Error())
			return err
		}
	}

	return nil
}

// NewApp creates a new App application struct
func NewApp() *App {
	return &App{}
}

func (a *App) Startup(ctx context.Context) {
	a.ctx = ctx

	// Initialize auth service first
	a.authService = auth.NewService(ctx)
	if err := a.authService.Initialize(ctx); err != nil {
		runtime.LogFatalf(ctx, "Failed to initialize auth service: %v", err)
		return
	}

	a.registrationService = registration.NewService(a.ctx)
	a.transferService = transfer.NewService(a.ctx)
	a.clientService = client.NewService(a.ctx)
}

// Helper method to initialize the database with encryption
func (a *App) initializeDatabase() error {
	// Get database key from auth service
	dbKey, err := a.authService.GetDBKey()
	if err != nil {
		return err
	}

	// Initialize database with encryption key
	dbPath := authutils.GetDatabasePath()
	db, err := database.Initialize(dbPath, dbKey)
	if err != nil {
		return err
	}

	a.db = db
	runtime.LogInfo(a.ctx, "Database initialized successfully with encryption")

	// Create default folder for uploads if it doesn't exist
	defaultFolder, err := a.ensureDefaultFolder(db.DB)
	if err != nil {
		runtime.LogError(a.ctx, "Failed to create default folder: "+err.Error())
		return err
	}
	a.defaultFolderID = defaultFolder

	// Initialize filestore service with database and encryption key
	a.fileService = filestore.NewService(a.ctx, db.DB, dbKey)
	runtime.LogInfo(a.ctx, "File storage service initialized")

	// Re-initialize transfer and server services with filestore service
	a.serverService = server.NewService(
		a.ctx,
		a.registrationService,
		a.transferService,
		a.fileService,
		a.defaultFolderID,
	)
	return nil
}

// Create the default folder for storing files if it doesn't exist
func (a *App) ensureDefaultFolder(db *sql.DB) (int64, error) {
	// Check if default folder exists
	var folderId int64
	err := db.QueryRow("SELECT id FROM folders WHERE name = 'Received Files' AND parent_id IS NULL").Scan(&folderId)
	if err != nil {
		if err == sql.ErrNoRows {
			// Create default folder
			result, err := db.Exec(
				"INSERT INTO folders (name, parent_id, created_at, updated_at) VALUES (?, NULL, datetime('now'), datetime('now'))",
				"Received Files",
			)
			if err != nil {
				return 0, err
			}
			return result.LastInsertId()
		}
		return 0, err
	}
	return folderId, nil
}

func (a *App) Shutdown(ctx context.Context) {
	if a.db != nil {
		a.db.Close()
	}
}

func (a *App) StartServer(port int) error {
	return a.serverService.Start(port)
}

func (a *App) StopServer() error {
	return a.serverService.Stop(a.ctx)
}

func (a *App) IsServerRunning() bool {
	return a.serverService.IsRunning()
}

// network functions
func (a *App) GetLocalIPs() ([]string, error) {
	return network.GetLocalIPs()
}

// Filestore functions
func (a *App) GetStoredFiles() ([]filestore.FileInfo, error) {
	return a.fileService.GetStoredFiles()
}

func (a *App) OpenFileByID(id int64) error {
	return a.fileService.OpenFileByID(id)
}
