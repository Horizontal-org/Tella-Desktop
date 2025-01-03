package server

import (
    "context"
    "fmt"
    "net/http"
	"github.com/wailsapp/wails/v2/pkg/runtime"
)

type Server struct {
    server   *http.Server
    running  bool
    port     int
	ctx      context.Context
}

func NewServer(ctx context.Context) *Server {
    return &Server{
        running: false,
		ctx:     ctx,
    }
}

func (s *Server) Start(ctx context.Context, port int) error {
    if s.running {
        return fmt.Errorf("server is already running")
    }

    mux := http.NewServeMux()
    s.server = &http.Server{
        Addr:    fmt.Sprintf(":%d", port),
        Handler: mux,
    }
    s.port = port

    // Start server in a goroutine
    go func() {
        if err := s.server.ListenAndServe(); err != http.ErrServerClosed {
            runtime.LogError(s.ctx, fmt.Sprintf("HTTP server error: %v", err))
        }
    }()

    s.running = true
	runtime.LogInfo(s.ctx, fmt.Sprintf("Server is running on port %d", port))
    return nil
}

func (s *Server) Stop(ctx context.Context) error {
    if !s.running {
        return nil
    }
    
    if err := s.server.Shutdown(ctx); err != nil {
        return fmt.Errorf("server shutdown error: %v", err)
    }
    
    s.running = false
    return nil
}

func (s *Server) IsRunning() bool {
    return s.running
}