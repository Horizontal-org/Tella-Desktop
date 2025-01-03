package app

import (
	"context"
	"Tella-Desktop/backend/core/services"
)

// App struct
type App struct {
	ctx context.Context
	serverService *services.ServerService
}

// NewApp creates a new App application struct
func NewApp() *App {
	return &App{}
}

func (a *App) Startup(ctx context.Context) {
    a.ctx = ctx
	a.serverService = services.NewServerService(ctx)
}

func (a *App) StartServer(port int) error {
    return a.serverService.Start(a.ctx, port)
}

func (a *App) StopServer() error {
    return a.serverService.Stop(a.ctx)
}

func (a *App) IsServerRunning() bool {
    return a.serverService.IsRunning()
}
