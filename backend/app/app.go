package app

import (
	"context"

	"Tella-Desktop/backend/core/services"
	"Tella-Desktop/backend/infrastructure/server"
)

// App struct
type App struct {
	ctx context.Context
	deviceService *services.DeviceService
	server *server.Server
}

// NewApp creates a new App application struct
func NewApp() *App {
	return &App{}
}

func (a *App) Startup(ctx context.Context) {
    a.ctx = ctx
    a.deviceService = services.NewDeviceService(ctx)
    a.server = server.NewServer(ctx, a.deviceService)
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