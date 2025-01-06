package server

import (
    "context"
    "fmt"
    "net/http"
    "github.com/wailsapp/wails/v2/pkg/runtime"
    
    "Tella-Desktop/backend/core/modules/registration"
    "Tella-Desktop/backend/core/modules/transfer"
)

type service struct {
    server   *http.Server
    running  bool
    port     int
    ctx      context.Context
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

    s.server.Addr = fmt.Sprintf(":%d", port)
    s.port = port

    go func() {
        if err := s.server.ListenAndServe(); err != http.ErrServerClosed {
            runtime.LogError(s.ctx, fmt.Sprintf("HTTP server error: %v", err))
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