package app

import (
	"context"

	"Tella-Desktop/backend/core/modules/client"
	"Tella-Desktop/backend/core/modules/registration"
	"Tella-Desktop/backend/core/modules/server"
	"Tella-Desktop/backend/core/modules/transfer"
	"Tella-Desktop/backend/utils/network"
)

// App struct
type App struct {
	ctx                 context.Context
	registrationService registration.Service
	transferService     transfer.Service
	serverService       server.Service
	clientService       client.Service
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

	a.registrationService = registration.NewService(a.ctx)
	a.transferService = transfer.NewService(a.ctx)

	a.serverService = server.NewService(
		a.ctx,
		a.registrationService,
		a.transferService,
	)
	a.clientService = client.NewService(a.ctx)
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
