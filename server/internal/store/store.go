// Package store defines storage interfaces and implementations for kip.
package store

import (
	"context"
	"errors"
	"time"
)

// ErrNotFound is returned when a secret does not exist or has expired.
var ErrNotFound = errors.New("secret not found")

// Secret represents an encrypted secret stored in Redis.
type Secret struct {
	ID                string    `json:"id"`
	Ciphertext        []byte    `json:"ciphertext"`
	Nonce             []byte    `json:"nonce"`
	Salt              []byte    `json:"salt,omitempty"`
	Filename          string    `json:"filename"`
	ReadsLeft         int       `json:"reads_left"`
	PasswordProtected bool      `json:"password_protected"`
	CreatedAt         time.Time `json:"created_at"`
}

// CreateSecretInput contains the data needed to create a new secret.
type CreateSecretInput struct {
	Ciphertext        []byte
	Nonce             []byte
	Salt              []byte
	Filename          string
	MaxReads          int
	PasswordProtected bool
	TTL               time.Duration
}

// SecretStore defines the interface for temporary secret storage.
type SecretStore interface {
	// Create stores a new encrypted secret and returns its ID.
	Create(ctx context.Context, input CreateSecretInput) (string, error)

	// Get retrieves a secret by ID, decrements reads_left, and deletes if exhausted.
	Get(ctx context.Context, id string) (*Secret, error)

	// Delete removes a secret by ID. Returns ErrNotFound if it does not exist.
	Delete(ctx context.Context, id string) error

	// Close closes the store connection.
	Close() error
}
