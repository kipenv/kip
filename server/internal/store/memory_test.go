package store

import (
	"context"
	"testing"
	"time"
)

func TestMemoryStoreCreateAndGet(t *testing.T) {
	s := NewMemoryStore()
	ctx := context.Background()

	id, err := s.Create(ctx, CreateSecretInput{
		Ciphertext: []byte("encrypted-data"),
		Nonce:      []byte("nonce-12byte"),
		Filename:   ".env",
		MaxReads:   2,
		TTL:        time.Hour,
	})
	if err != nil {
		t.Fatalf("Create() error: %v", err)
	}
	if id == "" {
		t.Fatal("Create() returned empty ID")
	}

	// First read
	secret, err := s.Get(ctx, id)
	if err != nil {
		t.Fatalf("Get() first read error: %v", err)
	}
	if string(secret.Ciphertext) != "encrypted-data" {
		t.Errorf("Ciphertext = %q, want %q", secret.Ciphertext, "encrypted-data")
	}
	if secret.Filename != ".env" {
		t.Errorf("Filename = %q, want %q", secret.Filename, ".env")
	}
	if secret.ReadsLeft != 1 {
		t.Errorf("ReadsLeft = %d, want 1", secret.ReadsLeft)
	}

	// Second read — should succeed and then delete
	secret, err = s.Get(ctx, id)
	if err != nil {
		t.Fatalf("Get() second read error: %v", err)
	}
	if secret.ReadsLeft != 0 {
		t.Errorf("ReadsLeft = %d, want 0", secret.ReadsLeft)
	}

	// Third read — should fail
	_, err = s.Get(ctx, id)
	if err != ErrNotFound {
		t.Errorf("Get() third read error = %v, want ErrNotFound", err)
	}
}

func TestMemoryStoreNotFound(t *testing.T) {
	s := NewMemoryStore()
	_, err := s.Get(context.Background(), "nonexistent")
	if err != ErrNotFound {
		t.Errorf("Get() error = %v, want ErrNotFound", err)
	}
}

func TestMemoryStoreExpiration(t *testing.T) {
	s := NewMemoryStore()
	ctx := context.Background()

	id, _ := s.Create(ctx, CreateSecretInput{
		Ciphertext: []byte("data"),
		Nonce:      []byte("nonce"),
		Filename:   ".env",
		MaxReads:   10,
		TTL:        1 * time.Millisecond,
	})

	// Wait for expiration
	time.Sleep(5 * time.Millisecond)

	_, err := s.Get(ctx, id)
	if err != ErrNotFound {
		t.Errorf("Get() after expiration error = %v, want ErrNotFound", err)
	}
}

func TestMemoryStoreSingleRead(t *testing.T) {
	s := NewMemoryStore()
	ctx := context.Background()

	id, _ := s.Create(ctx, CreateSecretInput{
		Ciphertext: []byte("one-time"),
		Nonce:      []byte("nonce"),
		Filename:   ".env",
		MaxReads:   1,
		TTL:        time.Hour,
	})

	// First read succeeds
	_, err := s.Get(ctx, id)
	if err != nil {
		t.Fatalf("Get() error: %v", err)
	}

	// Second read fails
	_, err = s.Get(ctx, id)
	if err != ErrNotFound {
		t.Errorf("Get() error = %v, want ErrNotFound", err)
	}
}

func TestMemoryStoreUniqueIDs(t *testing.T) {
	s := NewMemoryStore()
	ctx := context.Background()
	ids := make(map[string]bool)

	for i := 0; i < 100; i++ {
		id, err := s.Create(ctx, CreateSecretInput{
			Ciphertext: []byte("data"),
			Nonce:      []byte("nonce"),
			MaxReads:   1,
			TTL:        time.Hour,
		})
		if err != nil {
			t.Fatalf("Create() error: %v", err)
		}
		if ids[id] {
			t.Fatalf("Create() produced duplicate ID: %s", id)
		}
		ids[id] = true
	}
}
