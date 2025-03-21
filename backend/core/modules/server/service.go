package server

import (
	"context"
	"fmt"
	"math/rand"
	"net"
	"net/http"

	"github.com/wailsapp/wails/v2/pkg/runtime"

	"Tella-Desktop/backend/core/modules/filestore"
	"Tella-Desktop/backend/core/modules/registration"
	"Tella-Desktop/backend/core/modules/transfer"
	"Tella-Desktop/backend/utils/network"
	"Tella-Desktop/backend/utils/tls"
)

type service struct {
	server              *http.Server
	running             bool
	port                int
	pin                 string
	ctx                 context.Context
	registrationService registration.Service
	fileService         filestore.Service
	defaultFolderID     int64
}

func NewService(
	ctx context.Context,
	registrationService registration.Service,
	transferService transfer.Service,
	fileService filestore.Service,
	defaultFolderID int64,
) Service {
	srv := &service{
		ctx:                 ctx,
		running:             false,
		fileService:         fileService,
		defaultFolderID:     defaultFolderID,
		registrationService: registrationService,
	}

	// Initialize handlers
	registrationHandler := registration.NewHandler(registrationService)
	transferHandler := transfer.NewHandler(transferService, fileService, defaultFolderID)

	// Setup routes using handler
	mux := http.NewServeMux()
	handler := NewHandler(mux, registrationHandler, transferHandler)
	handler.SetupRoutes()

	srv.server = &http.Server{
		Handler: mux,
	}

	return srv
}

func (s *service) Start(port int) error {
	if s.running {
		return fmt.Errorf("server is already running")
	}

	s.pin = generateRandomPIN()

	s.registrationService.SetPINCode(s.pin)

	ipStrings, err := network.GetLocalIPs()

	if err != nil {
		return fmt.Errorf("failed to get local IPs: %v", err)
	}

	// parse strings ip into net.IP
	var ips []net.IP
	for _, ipStr := range ipStrings {
		if ip := net.ParseIP(ipStr); ip != nil {
			ips = append(ips, ip)
		}
	}
	tlsConfig, err := tls.GenerateTLSConfig(s.ctx, tls.Config{
		CommonName:   "Tella Desktop",
		Organization: []string{"Tella"},
		IPAddresses:  ips,
	})
	if err != nil {
		return fmt.Errorf("failed to generate TLS config: %v", err)
	}

	s.server.Addr = fmt.Sprintf(":%d", port)
	s.server.TLSConfig = tlsConfig
	s.port = port

	go func() {
		if err := s.server.ListenAndServeTLS("", ""); err != http.ErrServerClosed {
			runtime.LogError(s.ctx, fmt.Sprintf("HTTP server error: %v", err))
		}
	}()

	s.running = true
	runtime.LogInfo(s.ctx, fmt.Sprintf("HTTPS Server started on port %d", port))
	return nil
}

func (s *service) Stop(ctx context.Context) error {
	if !s.running {
		return nil
	}

	if err := s.server.Shutdown(ctx); err != nil {
		return fmt.Errorf("server shutdown error: %v", err)
	}

	s.running = false
	return nil
}

func (s *service) IsRunning() bool {
	return s.running
}

func (s *service) GetPIN() string {
	return s.pin
}

func generateRandomPIN() string {
	pinNumber := 100000 + rand.Intn(900000)
	return fmt.Sprintf("%06d", pinNumber)
}
