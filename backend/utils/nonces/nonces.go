package nonces

import (
	"errors"
	"sync"
)

type NonceManager struct {
	// seen[nonce]bool
	seen sync.Map
}

func NewNonceManager() *NonceManager {
	return &NonceManager{seen: sync.Map{}}
}

var ErrNonceReuse = errors.New("nonce has already been seen before")

func (n *NonceManager) Add(nonce string) error {
	if _, ok := n.seen.Load(nonce); ok {
		return ErrNonceReuse
	}
	n.seen.Store(nonce, true)
	return nil
}
