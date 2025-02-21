package server

import "context"

type Service interface {
	Start(port int) error
	Stop(ctx context.Context) error
	IsRunning() bool
}
