package ports

import "context"

type ServerService interface {
    Start(ctx context.Context, port int) error
    Stop(ctx context.Context) error
    IsRunning() bool
}