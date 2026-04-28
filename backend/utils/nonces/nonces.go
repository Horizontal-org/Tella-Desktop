package nonces

import (
	"errors"

	// pastedown.lru is safe for concurrent use
	"github.com/cespare/pastedown/lru"
)

type NonceManager struct {
	// seen[nonce] -> []byte{1}
	seen *lru.Cache
}

const CAPACITY_BYTES = 100000000 // max bytes of memory to use before evicting entries
// corresponds to the number of nonces that will be tracked.

func NewNonceManager() *NonceManager {
	return &NonceManager{seen: lru.New(CAPACITY_BYTES)}
}

var ErrNonceReuse = errors.New("nonce has already been seen before")
var ErrNonceZeroLength = errors.New("nonce is of length zero")

func (n *NonceManager) Add(nonce string) error {
	if nonce == "" {
		return ErrNonceZeroLength
	}
	if _, exists := n.seen.Get(nonce); exists {
		return ErrNonceReuse
	}
	n.seen.Insert(nonce, []byte{1})
	return nil
}
