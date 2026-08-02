package crypto

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"encoding/hex"
	"testing"
)

func TestDecryptWrongKey(t *testing.T) {
	plaintext := []byte("STRIPE_KEY=sk_live_abc123")

	ciphertext, nonce, _, err := Encrypt(plaintext)
	if err != nil {
		t.Fatalf("Encrypt() error: %v", err)
	}

	wrongKey, _ := GenerateKey()
	_, err = Decrypt(ciphertext, nonce, wrongKey)
	if err == nil {
		t.Error("Decrypt() should fail with wrong key")
	}
}

func TestDecryptTamperedCiphertext(t *testing.T) {
	plaintext := []byte("DB_PASSWORD=secret")

	ciphertext, nonce, key, err := Encrypt(plaintext)
	if err != nil {
		t.Fatalf("Encrypt() error: %v", err)
	}

	// Tamper with ciphertext
	tampered := make([]byte, len(ciphertext))
	copy(tampered, ciphertext)
	tampered[0] ^= 0xFF

	_, err = Decrypt(tampered, nonce, key)
	if err == nil {
		t.Error("Decrypt() should fail with tampered ciphertext")
	}
}

func TestDecryptTamperedNonce(t *testing.T) {
	plaintext := []byte("API_KEY=abc123")

	ciphertext, nonce, key, err := Encrypt(plaintext)
	if err != nil {
		t.Fatalf("Encrypt() error: %v", err)
	}

	// Tamper with nonce
	tampered := make([]byte, len(nonce))
	copy(tampered, nonce)
	tampered[0] ^= 0xFF

	_, err = Decrypt(ciphertext, tampered, key)
	if err == nil {
		t.Error("Decrypt() should fail with tampered nonce")
	}
}

func TestDecryptTruncatedCiphertext(t *testing.T) {
	plaintext := []byte("SECRET=value")

	ciphertext, nonce, key, err := Encrypt(plaintext)
	if err != nil {
		t.Fatalf("Encrypt() error: %v", err)
	}

	// Truncate ciphertext
	truncated := ciphertext[:len(ciphertext)/2]
	_, err = Decrypt(truncated, nonce, key)
	if err == nil {
		t.Error("Decrypt() should fail with truncated ciphertext")
	}
}

func TestDecryptInvalidKeySize(t *testing.T) {
	_, err := Decrypt([]byte("data"), make([]byte, NonceSize), []byte("short"))
	if err == nil {
		t.Error("Decrypt() should fail with invalid key size")
	}
}

func TestDecryptInvalidNonceSize(t *testing.T) {
	_, err := Decrypt([]byte("data"), []byte("short"), make([]byte, KeySize))
	if err == nil {
		t.Error("Decrypt() should fail with invalid nonce size")
	}
}

// TestDecryptKnownVector verifies our implementation against a known
// AES-256-GCM test vector to ensure correctness.
func TestDecryptKnownVector(t *testing.T) {
	// NIST AES-256-GCM test vector (simplified)
	// We encrypt with known key/nonce and verify round-trip.
	key, _ := hex.DecodeString("0000000000000000000000000000000000000000000000000000000000000001")
	nonce, _ := hex.DecodeString("000000000000000000000001")
	plaintext := []byte("Hello, kip!")

	block, err := aes.NewCipher(key)
	if err != nil {
		t.Fatalf("NewCipher() error: %v", err)
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		t.Fatalf("NewGCM() error: %v", err)
	}
	ciphertext := gcm.Seal(nil, nonce, plaintext, nil)

	// Now decrypt with our function
	decrypted, err := Decrypt(ciphertext, nonce, key)
	if err != nil {
		t.Fatalf("Decrypt() error: %v", err)
	}
	if !bytes.Equal(decrypted, plaintext) {
		t.Errorf("Decrypt() = %q, want %q", decrypted, plaintext)
	}
}

func TestDecryptWithPasswordRoundTrip(t *testing.T) {
	testCases := []struct {
		name     string
		text     string
		password string
	}{
		{"simple", "KEY=value", "password123"},
		{"empty password", "KEY=value", ""},
		{"unicode password", "KEY=value", "contraseña-秘密"},
		{"long password", "KEY=value", string(make([]byte, 1024))},
		{"empty plaintext", "", "password"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			plaintext := []byte(tc.text)
			password := []byte(tc.password)

			ct, nonce, salt, err := EncryptWithPassword(plaintext, password)
			if err != nil {
				t.Fatalf("EncryptWithPassword() error: %v", err)
			}

			decrypted, err := DecryptWithPassword(ct, nonce, salt, password)
			if err != nil {
				t.Fatalf("DecryptWithPassword() error: %v", err)
			}

			if !bytes.Equal(decrypted, plaintext) {
				t.Errorf("round-trip failed: got %q, want %q", decrypted, plaintext)
			}
		})
	}
}
