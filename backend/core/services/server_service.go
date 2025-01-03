package services

import (
    "context"
    "Tella-Desktop/backend/infrastructure/server"
)

type ServerService struct {
    server *server.Server
	ctx    context.Context
}

func NewServerService(ctx context.Context) *ServerService {
    return &ServerService{
        server: server.NewServer(ctx),
		ctx:    ctx,
    }
}

func (s *ServerService) Start(ctx context.Context, port int) error {
    return s.server.Start(ctx, port)
}

func (s *ServerService) Stop(ctx context.Context) error {
    return s.server.Stop(ctx)
}

func (s *ServerService) IsRunning() bool {
    return s.server.IsRunning()
}