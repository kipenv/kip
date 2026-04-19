package store

import (
	"context"
	"sync"
	"time"

	nanoid "github.com/matoous/go-nanoid/v2"
)

// memEntry is a secret with its expiration time.
type memEntry struct {
	secret    Secret
	expiresAt time.Time
}

// MemoryStore implements SecretStore using an in-memory map.
// Used for testing and development without Redis.
type MemoryStore struct {
	mu      sync.RWMutex
	secrets map[string]memEntry
}

// NewMemoryStore creates a new in-memory secret store.
func NewMemoryStore() *MemoryStore {
	return &MemoryStore{
		secrets: make(map[string]memEntry),
	}
}

// Create stores a secret in memory.
func (s *MemoryStore) Create(_ context.Context, input CreateSecretInput) (string, error) {
	id, err := nanoid.Generate(idAlphabet, idLength)
	if err != nil {
		return "", err
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	s.secrets[id] = memEntry{
		secret: Secret{
			ID:                id,
			Ciphertext:        input.Ciphertext,
			Nonce:             input.Nonce,
			Salt:              input.Salt,
			Filename:          input.Filename,
			ReadsLeft:         input.MaxReads,
			PasswordProtected: input.PasswordProtected,
			CreatedAt:         time.Now().UTC(),
		},
		expiresAt: time.Now().Add(input.TTL),
	}

	return id, nil
}

// Get retrieves a secret, decrements reads, and deletes if exhausted or expired.
func (s *MemoryStore) Get(_ context.Context, id string) (*Secret, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	entry, ok := s.secrets[id]
	if !ok {
		return nil, ErrNotFound
	}

	if time.Now().After(entry.expiresAt) {
		delete(s.secrets, id)
		return nil, ErrNotFound
	}

	entry.secret.ReadsLeft--
	if entry.secret.ReadsLeft <= 0 {
		delete(s.secrets, id)
	} else {
		s.secrets[id] = entry
	}

	result := entry.secret
	return &result, nil
}

// Delete removes a secret from memory by ID.
func (s *MemoryStore) Delete(_ context.Context, id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.secrets[id]; !ok {
		return ErrNotFound
	}
	delete(s.secrets, id)
	return nil
}

// Close is a no-op for the memory store.
func (s *MemoryStore) Close() error {
	return nil
}
