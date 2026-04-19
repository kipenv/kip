// Package crypto provides AES-256-GCM encryption and decryption
// with zero-knowledge design. Keys are generated locally and never
// sent to the server.
package crypto

import (
	"crypto/rand"
	"fmt"

	"github.com/mr-tron/base58"
)

const (
	// KeySize is the size of the AES-256 key in bytes.
	KeySize = 32
	// NonceSize is the size of the GCM nonce in bytes.
	NonceSize = 12
)

// GenerateKey creates a cryptographically secure random 32-byte key.
func GenerateKey() ([]byte, error) {
	key := make([]byte, KeySize)
	if _, err := rand.Read(key); err != nil {
		return nil, fmt.Errorf("generate key: %w", err)
	}
	return key, nil
}

// GenerateNonce creates a cryptographically secure random 12-byte nonce.
func GenerateNonce() ([]byte, error) {
	nonce := make([]byte, NonceSize)
	if _, err := rand.Read(nonce); err != nil {
		return nil, fmt.Errorf("generate nonce: %w", err)
	}
	return nonce, nil
}

// EncodeKey encodes a key to base58 for use in URL fragments.
// Base58 avoids ambiguous characters (0, O, I, l) making it
// safe for URLs and easy to copy/paste.
func EncodeKey(key []byte) string {
	return base58.Encode(key)
}

// DecodeKey decodes a base58-encoded key.
func DecodeKey(encoded string) ([]byte, error) {
	key, err := base58.Decode(encoded)
	if err != nil {
		return nil, fmt.Errorf("decode key: %w", err)
	}
	if len(key) != KeySize {
		return nil, fmt.Errorf("decode key: expected %d bytes, got %d", KeySize, len(key))
	}
	return key, nil
}
