package app

import (
	"context"

	"Tella-Desktop/backend/core/services"
	"Tella-Desktop/backend/infrastructure/server"
	"Tella-Desktop/backend/utils/network"
)

// App struct
type App struct {
	ctx context.Context
	deviceService *services.DeviceService
	fileService *services.FileService
	clientService *services.ClientService
	server *server.Server
}

func (a *App) RegisterWithDevice(ip string, port int) error {
    return a.clientService.RegisterWithDevice(ip, port)
}

func (a *App) SendTestFile(ip string, port int, pin string) error {
    return a.clientService.SendTestFile(ip, port, pin)
}

// NewApp creates a new App application struct
func NewApp() *App {
	return &App{}
}

func (a *App) Startup(ctx context.Context) {
    a.ctx = ctx
    a.deviceService = services.NewDeviceService(ctx)
	a.fileService = services.NewFileService(ctx)
	a.clientService = services.NewClientService(ctx)
    a.server = server.NewServer(ctx, a.deviceService, a.fileService)
}

func (a *App) StartServer(port int) error {
    return a.server.Start(a.ctx, port)
}

func (a *App) StopServer() error {
    return a.server.Stop(a.ctx)
}

func (a *App) IsServerRunning() bool {
    return a.server.IsRunning()
}

// network functions
func (a *App) GetLocalIPs() ([]string, error) {
    return network.GetLocalIPs()
}