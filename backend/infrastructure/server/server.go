package server

import (
    "context"
    "fmt"
    "net/http"
	"github.com/wailsapp/wails/v2/pkg/runtime"

	"Tella-Desktop/backend/core/ports"
    "Tella-Desktop/backend/handlers"
)

type Server struct {
    server   *http.Server
    running  bool
    port     int
	ctx      context.Context
	deviceService ports.DeviceService
    fileService ports.FileService
}

func NewServer(ctx context.Context, deviceService ports.DeviceService, fileService ports.FileService) *Server {
    return &Server{
        running: false,
		ctx:     ctx,
		deviceService: deviceService,
        fileService:  fileService,
    }
}

func (s *Server) Start(ctx context.Context, port int) error {
    if s.running {
        return fmt.Errorf("server is already running")
    }

    mux := http.NewServeMux()
	
	// Initialize handlers
    registerHandler := handlers.NewRegisterHandler(s.deviceService)
    uploadHandler := handlers.NewUploadHandler(s.fileService)
    // Register routes
    mux.HandleFunc("/api/localsend/v2/register", registerHandler.Handle)
    mux.HandleFunc("/api/localsend/v2/prepare-upload", uploadHandler.HandlePrepare)
    mux.HandleFunc("/api/localsend/v2/upload", uploadHandler.HandleUpload)

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