package crypto

import (
	"bytes"
	"testing"
)

func TestEncryptProducesOutput(t *testing.T) {
	plaintext := []byte("DB_PASSWORD=supersecret123")

	ciphertext, nonce, key, err := Encrypt(plaintext)
	if err != nil {
		t.Fatalf("Encrypt() error: %v", err)
	}

	if len(ciphertext) == 0 {
		t.Error("Encrypt() produced empty ciphertext")
	}
	if len(nonce) != NonceSize {
		t.Errorf("Encrypt() nonce length = %d, want %d", len(nonce), NonceSize)
	}
	if len(key) != KeySize {
		t.Errorf("Encrypt() key length = %d, want %d", len(key), KeySize)
	}
}

func TestEncryptCiphertextDiffersFromPlaintext(t *testing.T) {
	plaintext := []byte("STRIPE_KEY=sk_live_abc123")

	ciphertext, _, _, err := Encrypt(plaintext)
	if err != nil {
		t.Fatalf("Encrypt() error: %v", err)
	}

	if bytes.Equal(ciphertext, plaintext) {
		t.Error("Encrypt() ciphertext equals plaintext — encryption failed")
	}
}

func TestEncryptUniqueOutputs(t *testing.T) {
	plaintext := []byte("same-input-every-time")

	ct1, n1, k1, _ := Encrypt(plaintext)
	ct2, n2, k2, _ := Encrypt(plaintext)

	if bytes.Equal(k1, k2) {
		t.Error("Encrypt() produced same key twice")
	}
	if bytes.Equal(n1, n2) {
		t.Error("Encrypt() produced same nonce twice")
	}
	if bytes.Equal(ct1, ct2) {
		t.Error("Encrypt() produced same ciphertext twice")
	}
}

func TestEncryptRoundTrip(t *testing.T) {
	testCases := []struct {
		name      string
		plaintext string
	}{
		{"simple env", "DB_HOST=localhost\nDB_PORT=5432"},
		{"empty value", "EMPTY_VAR="},
		{"special chars", "PASS=p@$$w0rd!#%^&*()"},
		{"unicode", "NAME=こんにちは世界"},
		{"large payload", string(make([]byte, 1024*1024))}, // 1MB
		{"single byte", "x"},
		{"multiline", "KEY1=val1\nKEY2=val2\nKEY3=val3\n"},
		{"with quotes", `DB_URL="postgres://user:pass@host:5432/db"`},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			plaintext := []byte(tc.plaintext)

			ciphertext, nonce, key, err := Encrypt(plaintext)
			if err != nil {
				t.Fatalf("Encrypt() error: %v", err)
			}

			decrypted, err := Decrypt(ciphertext, nonce, key)
			if err != nil {
				t.Fatalf("Decrypt() error: %v", err)
			}

			if !bytes.Equal(decrypted, plaintext) {
				t.Errorf("round-trip failed: got %q, want %q", decrypted, plaintext)
			}
		})
	}
}

func TestEncryptEmptyPlaintext(t *testing.T) {
	ciphertext, nonce, key, err := Encrypt([]byte{})
	if err != nil {
		t.Fatalf("Encrypt() error: %v", err)
	}

	decrypted, err := Decrypt(ciphertext, nonce, key)
	if err != nil {
		t.Fatalf("Decrypt() error: %v", err)
	}

	if len(decrypted) != 0 {
		t.Errorf("expected empty plaintext, got %d bytes", len(decrypted))
	}
}

func TestEncryptWithKeyInvalidKeySize(t *testing.T) {
	_, _, err := EncryptWithKey([]byte("data"), []byte("short"))
	if err == nil {
		t.Error("EncryptWithKey() should fail with short key")
	}
}

func TestEncryptWithPassword(t *testing.T) {
	plaintext := []byte("AWS_SECRET=wJalrXUtnFEMI/K7MDENG/bPxRfiCY")
	password := []byte("my-strong-password-123")

	ciphertext, nonce, salt, err := EncryptWithPassword(plaintext, password)
	if err != nil {
		t.Fatalf("EncryptWithPassword() error: %v", err)
	}

	if len(salt) != SaltSize {
		t.Errorf("salt length = %d, want %d", len(salt), SaltSize)
	}

	decrypted, err := DecryptWithPassword(ciphertext, nonce, salt, password)
	if err != nil {
		t.Fatalf("DecryptWithPassword() error: %v", err)
	}

	if !bytes.Equal(decrypted, plaintext) {
		t.Errorf("round-trip failed: got %q, want %q", decrypted, plaintext)
	}
}

func TestEncryptWithPasswordWrongPassword(t *testing.T) {
	plaintext := []byte("SECRET=value")
	password := []byte("correct-password")

	ciphertext, nonce, salt, err := EncryptWithPassword(plaintext, password)
	if err != nil {
		t.Fatalf("EncryptWithPassword() error: %v", err)
	}

	_, err = DecryptWithPassword(ciphertext, nonce, salt, []byte("wrong-password"))
	if err == nil {
		t.Error("DecryptWithPassword() should fail with wrong password")
	}
}

func TestEncryptWithPasswordDifferentSalts(t *testing.T) {
	plaintext := []byte("SECRET=value")
	password := []byte("password")

	_, _, salt1, _ := EncryptWithPassword(plaintext, password)
	_, _, salt2, _ := EncryptWithPassword(plaintext, password)

	if bytes.Equal(salt1, salt2) {
		t.Error("EncryptWithPassword() produced same salt twice")
	}
}
