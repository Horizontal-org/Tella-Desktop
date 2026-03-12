package registration

import (
	"context"
	"errors"
	"time"

	"Tella-Desktop/backend/utils/constants"
	"github.com/google/uuid"
)

type service struct {
	ctx         context.Context
	sessions    map[string]*Session
	pinCode     string
	rateLimiter map[string]int
	done chan struct{}
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
		done:       make(chan struct{}),
	}
}

func (s *service) CreateSession(pin, nonce string) (string, error) {
	// TODO cblgh(2026-02-17): guard ratelimiter with mutex alt. use sync.Map to prevent crash from malicious behaviour?
	// TODO cblgh(2026-03-11) change this rate limiter to not be on the nonce any more?
	if s.rateLimiter[nonce] >= 3 { // check this with the team
		return "", errors.New("too many invalid attempts")
	}

	if pin != s.pinCode {
		s.rateLimiter[nonce]++
		return "", errors.New("Invalid pin")
	}

	sessionID := uuid.New().String()
	// cleaned up by ForgetSession, which is called during session management cleanup in core/module/transfer/service.go
	//
	// the sessionID is controlled by calling registration.SessionIsValid(incSessionID)
	s.sessions[sessionID] = &Session{
		ID:        sessionID,
		Nonce:     nonce,
		CreatedAt: time.Now(),
	}

	// cleanup fallback in case of lifecycle fuckup elsewhere / transfer service's session management
	// TODO cblgh(2026-02-17): add explicit lifecycle 'close' function which would also drain this goroutine (otherwise
	// risk for goroutine leak since it's only cleaned up 10h after starting). 
	// 
	// note: this is currently taken care of by s.ForgetSession, but a more orderly exit would be prefered :)
	go (func(sid string) {
		// 'done' channel fires when application has been locked -> 
		// exit goroutine and allow GC to cleanup reference to this service
		select {
		case <-s.done:
		case <-time.After(constants.CLEAN_UP_SESSION_TIMEOUT_MIN * time.Minute):
			if s == nil {
				return
			}
			s.ForgetSession(sid)
		}
	})(sessionID)

	delete(s.rateLimiter, nonce) // if pin is success we delete the rate limiter

	return sessionID, nil
}

func (s *service) SessionIsValid(sessionID string) bool {
	_, exists := s.sessions[sessionID]
	return exists
}

func (s *service) ForgetSession(sessionID string) {
	delete(s.sessions, sessionID)
	// drain the goroutine
	close(s.done)
	// setup new channel
	s.done = make(chan struct{})
}

func (s *service) SetPINCode(pinCode string) {
	s.pinCode = pinCode
}

func (s *service) Lock() {
	for k := range s.sessions {
		delete(s.sessions, k)
	}
	for k := range s.rateLimiter {
		delete(s.rateLimiter, k)
	}
	close(s.done)
}
