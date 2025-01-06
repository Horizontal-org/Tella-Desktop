package server

import "context"

type Service interface {
    Start(ctx context.Context, port int) error
    Stop(ctx context.Context) error
    IsRunning() bool
}