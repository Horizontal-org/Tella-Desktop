package registration

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
)

type service struct {
	ctx         context.Context
	sessions    map[string]*Session
	pinCode     string
	rateLimiter map[string]int
}

type Session struct {
	ID        string
	Nonce     string
	CreatedAt time.Time
}

func NewService(ctx context.Context) Service {
	return &service{
		ctx:         ctx,
		sessions:    make(map[string]*Session),
		rateLimiter: make(map[string]int),
	}
}

func (s *service) CreateSession(pin string, nonce string) (string, error) {
	if s.rateLimiter[nonce] >= 3 { // check this with the team
		return "", errors.New("too many invalid attempts")
	}

	if pin != s.pinCode {
		s.rateLimiter[nonce]++
		return "", errors.New("Invalid pin")
	}

	sessionID := uuid.New().String()
	s.sessions[sessionID] = &Session{
		ID:        sessionID,
		Nonce:     nonce,
		CreatedAt: time.Now(),
	}

	delete(s.rateLimiter, nonce) // if pin is success we delete the rate limiter

	return sessionID, nil
}

func (s *service) SetPINCode(pinCode string) {
	s.pinCode = pinCode
}
