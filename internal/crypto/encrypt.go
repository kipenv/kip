package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"fmt"

	"golang.org/x/crypto/argon2"
)

// Argon2id parameters for password-based key derivation.
const (
	Argon2Time    = 1
	Argon2Memory  = 64 * 1024 // 64 MB
	Argon2Threads = 4
	SaltSize      = 16
)

// Encrypt encrypts plaintext using AES-256-GCM with a randomly generated key.
// Returns the ciphertext, nonce, and key. The key must be transmitted separately
// (e.g., in a URL fragment) and never stored alongside the ciphertext.
func Encrypt(plaintext []byte) (ciphertext, nonce, key []byte, err error) {
	key, err = GenerateKey()
	if err != nil {
		return nil, nil, nil, err
	}

	ciphertext, nonce, err = EncryptWithKey(plaintext, key)
	if err != nil {
		return nil, nil, nil, err
	}

	return ciphertext, nonce, key, nil
}

// EncryptWithKey encrypts plaintext using AES-256-GCM with the provided key.
func EncryptWithKey(plaintext, key []byte) (ciphertext, nonce []byte, err error) {
	if len(key) != KeySize {
		return nil, nil, fmt.Errorf("encrypt: key must be %d bytes, got %d", KeySize, len(key))
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, nil, fmt.Errorf("encrypt: new cipher: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, nil, fmt.Errorf("encrypt: new gcm: %w", err)
	}

	nonce, err = GenerateNonce()
	if err != nil {
		return nil, nil, err
	}

	ciphertext = gcm.Seal(nil, nonce, plaintext, nil)
	return ciphertext, nonce, nil
}

// EncryptWithPassword encrypts plaintext using a password-derived key (Argon2id).
// Returns the ciphertext, nonce, and salt needed for decryption.
// The caller must store/transmit the salt alongside the ciphertext.
func EncryptWithPassword(plaintext, password []byte) (ciphertext, nonce, salt []byte, err error) {
	salt = make([]byte, SaltSize)
	if _, err = readRand(salt); err != nil {
		return nil, nil, nil, fmt.Errorf("encrypt with password: generate salt: %w", err)
	}

	key := deriveKey(password, salt)
	ciphertext, nonce, err = EncryptWithKey(plaintext, key)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("encrypt with password: %w", err)
	}

	return ciphertext, nonce, salt, nil
}

// deriveKey uses Argon2id to derive a 32-byte key from a password and salt.
func deriveKey(password, salt []byte) []byte {
	return argon2.IDKey(password, salt, Argon2Time, Argon2Memory, Argon2Threads, KeySize)
}

// readRand is a variable to allow testing with deterministic random.
var readRand = randRead

func randRead(b []byte) (int, error) {
	return rand.Read(b) //nolint:staticcheck // rand.Read is fine for our use
}
