package server

import (
	"context"
	"crypto/x509"
	"fmt"
	"net/http"

	"github.com/wailsapp/wails/v2/pkg/runtime"

	"Tella-Desktop/backend/core/modules/registration"
	"Tella-Desktop/backend/core/modules/transfer"
	"Tella-Desktop/backend/utils/tls"
)

type service struct {
	server  *http.Server
	running bool
	port    int
	ctx     context.Context
	cert    *x509.Certificate
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

	// Load or generate certificates
	certPath, keyPath, err := tls.LoadOrGenerateCertificateAndKey()
	if err != nil {
		return fmt.Errorf("failed to setup TLS certificates: %v", err)
	}

	// Load the certificate to get its information
	cert, err := tls.LoadCertificate(certPath)
	if err != nil {
		return fmt.Errorf("failed to load certificate: %v", err)
	}
	s.cert = cert

	// Display certificate hash in UI
	certInfo := tls.GetCertificateDisplayString(cert)
	runtime.EventsEmit(s.ctx, "certificate-info", certInfo)

	s.server.Addr = fmt.Sprintf(":%d", port)
	s.port = port

	go func() {
		runtime.LogInfo(s.ctx, fmt.Sprintf("Starting HTTPS server on port %d", port))
		if err := s.server.ListenAndServeTLS(certPath, keyPath); err != http.ErrServerClosed {
			runtime.LogError(s.ctx, fmt.Sprintf("HTTPS server error: %v", err))
		}
	}()

	s.running = true
	runtime.LogInfo(s.ctx, fmt.Sprintf("Server started on port %d", port))
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
