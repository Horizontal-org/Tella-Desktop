package server

import (
	"context"
	"fmt"
	"math/rand"
	"net"
	"net/http"
	"sync"
	"time"

	"Tella-Desktop/backend/core/modules/filestore"
	"Tella-Desktop/backend/core/modules/registration"
	"Tella-Desktop/backend/core/modules/transfer"
	"Tella-Desktop/backend/utils/network"
	"Tella-Desktop/backend/utils/tls"
)

type service struct {
	server              *http.Server
	listener            net.Listener
	running             bool
	port                int
	pin                 string
	ctx                 context.Context
	registrationService registration.Service
	registrationHandler *registration.Handler
	transferService     transfer.Service
	fileService         filestore.Service
	defaultFolderID     int64
	mu                  sync.RWMutex
}

func NewService(
	ctx context.Context,
	registrationService registration.Service,
	registrationHandler *registration.Handler,
	transferService transfer.Service,
	fileService filestore.Service,
	defaultFolderID int64,
) Service {
	srv := &service{
		ctx:                 ctx,
		running:             false,
		registrationService: registrationService,
		registrationHandler: registrationHandler,
		transferService:     transferService,
		fileService:         fileService,
		defaultFolderID:     defaultFolderID,
	}

	return srv
}

func (s *service) Start(port int) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.running {
		return fmt.Errorf("server is already running")
	}

	// Generate new PIN for each start
	s.pin = generateRandomPIN()
	s.registrationService.SetPINCode(s.pin)

	ipStrings, err := network.GetLocalIPs()
	if err != nil {
		return fmt.Errorf("failed to get local IPs: %v", err)
	}

	// Parse strings ip into net.IP
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

	mux := http.NewServeMux()

	transferHandler := transfer.NewHandler(s.transferService, s.fileService, s.defaultFolderID)

	handler := NewHandler(mux, s.registrationHandler, transferHandler)
	handler.SetupRoutes()

	s.server = &http.Server{
		Addr:         fmt.Sprintf(":%d", port),
		Handler:      mux,
		TLSConfig:    tlsConfig,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	s.port = port

	serverErrors := make(chan error, 1)

	// Start server in goroutine
	go func() {
		if err := s.server.ListenAndServeTLS("", ""); err != nil && err != http.ErrServerClosed {
			fmt.Printf("HTTP server error: %v\n", err)
			serverErrors <- err
			s.mu.Lock()
			s.running = false
			s.mu.Unlock()
		}
	}()

	// Give the server time to start up properly
	time.Sleep(500 * time.Millisecond)

	// Check if there were any immediate startup errors
	select {
	case err := <-serverErrors:
		return fmt.Errorf("server failed to start: %v", err)
	default:
		// server started successfully
	}

	s.running = true
	fmt.Printf("HTTPS Server started on port %d with PIN %s\n", port, s.pin)
	return nil
}

func (s *service) Stop(ctx context.Context) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.running {
		return nil
	}

	fmt.Printf("Stopping HTTPS Server...\n")

	shutdownCtx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	if err := s.server.Shutdown(shutdownCtx); err != nil {
		fmt.Printf("Graceful shutdown failed: %v, forcing close\n", err)
	}

	s.running = false
	s.server = nil

	fmt.Printf("HTTPS Server stopped\n")

	// Add delay to ensure port is fully released
	time.Sleep(1 * time.Second)

	return nil
}

func (s *service) IsRunning() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.running
}

func (s *service) GetPIN() string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.pin
}

func generateRandomPIN() string {
	pinNumber := 100000 + rand.Intn(900000)
	return fmt.Sprintf("%06d", pinNumber)
}
