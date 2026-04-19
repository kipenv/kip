package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"fmt"
)

// Decrypt decrypts ciphertext using AES-256-GCM with the provided key and nonce.
func Decrypt(ciphertext, nonce, key []byte) ([]byte, error) {
	if len(key) != KeySize {
		return nil, fmt.Errorf("decrypt: key must be %d bytes, got %d", KeySize, len(key))
	}
	if len(nonce) != NonceSize {
		return nil, fmt.Errorf("decrypt: nonce must be %d bytes, got %d", NonceSize, len(nonce))
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("decrypt: new cipher: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("decrypt: new gcm: %w", err)
	}

	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, fmt.Errorf("decrypt: %w", err)
	}

	return plaintext, nil
}

// DecryptWithPassword decrypts ciphertext using a password-derived key (Argon2id).
func DecryptWithPassword(ciphertext, nonce, salt, password []byte) ([]byte, error) {
	key := deriveKey(password, salt)
	plaintext, err := Decrypt(ciphertext, nonce, key)
	if err != nil {
		return nil, fmt.Errorf("decrypt with password: %w", err)
	}
	return plaintext, nil
}
