package store

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"

	nanoid "github.com/matoous/go-nanoid/v2"
)

const (
	secretPrefix = "secret:"
	idLength     = 12
	idAlphabet   = "0123456789abcdefghijklmnopqrstuvwxyz"
)

// RedisStore implements SecretStore using Redis.
type RedisStore struct {
	client *redis.Client
}

// NewRedisStore creates a new Redis-backed secret store.
func NewRedisStore(redisURL string) (*RedisStore, error) {
	opts, err := redis.ParseURL(redisURL)
	if err != nil {
		return nil, fmt.Errorf("parse redis url: %w", err)
	}

	client := redis.NewClient(opts)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("redis ping: %w", err)
	}

	return &RedisStore{client: client}, nil
}

// Create stores a new secret in Redis with TTL.
func (s *RedisStore) Create(ctx context.Context, input CreateSecretInput) (string, error) {
	id, err := nanoid.Generate(idAlphabet, idLength)
	if err != nil {
		return "", fmt.Errorf("generate id: %w", err)
	}

	secret := Secret{
		ID:         id,
		Ciphertext: input.Ciphertext,
		Nonce:      input.Nonce,
		Filename:   input.Filename,
		ReadsLeft:  input.MaxReads,
		CreatedAt:  time.Now().UTC(),
	}

	data, err := json.Marshal(secret)
	if err != nil {
		return "", fmt.Errorf("marshal secret: %w", err)
	}

	key := secretPrefix + id
	if err := s.client.Set(ctx, key, data, input.TTL).Err(); err != nil {
		return "", fmt.Errorf("redis set: %w", err)
	}

	return id, nil
}

// Get retrieves a secret, decrements reads_left, and deletes if exhausted.
func (s *RedisStore) Get(ctx context.Context, id string) (*Secret, error) {
	key := secretPrefix + id

	// Use a transaction to atomically read, decrement, and conditionally delete.
	var secret Secret

	err := s.client.Watch(ctx, func(tx *redis.Tx) error {
		data, getErr := tx.Get(ctx, key).Bytes()
		if getErr != nil {
			if errors.Is(getErr, redis.Nil) {
				return ErrNotFound
			}
			return fmt.Errorf("redis get: %w", getErr)
		}

		if err := json.Unmarshal(data, &secret); err != nil {
			return fmt.Errorf("unmarshal secret: %w", err)
		}

		secret.ReadsLeft--

		_, pipeErr := tx.TxPipelined(ctx, func(pipe redis.Pipeliner) error {
			if secret.ReadsLeft <= 0 {
				pipe.Del(ctx, key)
			} else {
				updated, err := json.Marshal(secret)
				if err != nil {
					return err
				}
				// Keep the existing TTL
				pipe.Set(ctx, key, updated, redis.KeepTTL)
			}
			return nil
		})

		return pipeErr
	}, key)
	if err != nil {
		return nil, err
	}

	return &secret, nil
}

// Delete removes a secret from Redis by ID.
func (s *RedisStore) Delete(ctx context.Context, id string) error {
	key := secretPrefix + id
	n, err := s.client.Del(ctx, key).Result()
	if err != nil {
		return fmt.Errorf("redis del: %w", err)
	}
	if n == 0 {
		return ErrNotFound
	}
	return nil
}

// Close closes the Redis connection.
func (s *RedisStore) Close() error {
	return s.client.Close()
}
