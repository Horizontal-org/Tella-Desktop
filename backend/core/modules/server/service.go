package server

import (
	"context"
	"fmt"
	"net"
	"net/http"

	"github.com/wailsapp/wails/v2/pkg/runtime"

	"Tella-Desktop/backend/core/modules/registration"
	"Tella-Desktop/backend/core/modules/transfer"
	"Tella-Desktop/backend/utils/network"
	"Tella-Desktop/backend/utils/tls"
)

type service struct {
	server  *http.Server
	running bool
	port    int
	ctx     context.Context
}

func NewService(
	ctx context.Context,
	registrationService registration.Service,
	transferService transfer.Service,
) Service {
	srv := &service{
		ctx:     ctx,
		running: false,
	}

	// Initialize handlers
	registrationHandler := registration.NewHandler(registrationService)
	transferHandler := transfer.NewHandler(transferService)

	// Setup routes using handler
	mux := http.NewServeMux()
	handler := NewHandler(mux, registrationHandler, transferHandler)
	handler.SetupRoutes()

	srv.server = &http.Server{
		Handler: mux,
	}

	return srv
}

func (s *service) Start(ctx context.Context, port int) error {
	if s.running {
		return fmt.Errorf("server is already running")
	}

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

	tlsConfig, err := tls.GenerateTLSConfig(tls.Config{
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
