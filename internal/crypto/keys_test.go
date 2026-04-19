package crypto

import (
	"testing"
)

func TestGenerateKey(t *testing.T) {
	key, err := GenerateKey()
	if err != nil {
		t.Fatalf("GenerateKey() error: %v", err)
	}
	if len(key) != KeySize {
		t.Errorf("GenerateKey() length = %d, want %d", len(key), KeySize)
	}
}

func TestGenerateKeyUniqueness(t *testing.T) {
	key1, err := GenerateKey()
	if err != nil {
		t.Fatalf("GenerateKey() error: %v", err)
	}
	key2, err := GenerateKey()
	if err != nil {
		t.Fatalf("GenerateKey() error: %v", err)
	}
	if string(key1) == string(key2) {
		t.Error("GenerateKey() produced identical keys")
	}
}

func TestGenerateNonce(t *testing.T) {
	nonce, err := GenerateNonce()
	if err != nil {
		t.Fatalf("GenerateNonce() error: %v", err)
	}
	if len(nonce) != NonceSize {
		t.Errorf("GenerateNonce() length = %d, want %d", len(nonce), NonceSize)
	}
}

func TestGenerateNonceUniqueness(t *testing.T) {
	nonce1, err := GenerateNonce()
	if err != nil {
		t.Fatalf("GenerateNonce() error: %v", err)
	}
	nonce2, err := GenerateNonce()
	if err != nil {
		t.Fatalf("GenerateNonce() error: %v", err)
	}
	if string(nonce1) == string(nonce2) {
		t.Error("GenerateNonce() produced identical nonces")
	}
}

func TestEncodeDecodeKey(t *testing.T) {
	key, err := GenerateKey()
	if err != nil {
		t.Fatalf("GenerateKey() error: %v", err)
	}

	encoded := EncodeKey(key)
	if len(encoded) == 0 {
		t.Fatal("EncodeKey() returned empty string")
	}

	decoded, err := DecodeKey(encoded)
	if err != nil {
		t.Fatalf("DecodeKey() error: %v", err)
	}

	if string(decoded) != string(key) {
		t.Error("DecodeKey(EncodeKey(key)) != key")
	}
}

func TestDecodeKeyInvalidBase58(t *testing.T) {
	_, err := DecodeKey("invalid!!!base58")
	if err == nil {
		t.Error("DecodeKey() should fail on invalid base58")
	}
}

func TestDecodeKeyWrongLength(t *testing.T) {
	// Encode a short byte slice (not 32 bytes)
	short := EncodeKey([]byte("too-short"))
	_, err := DecodeKey(short)
	if err == nil {
		t.Error("DecodeKey() should fail on wrong length")
	}
}

func TestEncodeKeyDeterministic(t *testing.T) {
	key, _ := GenerateKey()
	enc1 := EncodeKey(key)
	enc2 := EncodeKey(key)
	if enc1 != enc2 {
		t.Error("EncodeKey() should be deterministic for same input")
	}
}

func TestEncodeKeyNoAmbiguousChars(t *testing.T) {
	// Base58 should not contain 0, O, I, l
	for i := 0; i < 100; i++ {
		key, _ := GenerateKey()
		encoded := EncodeKey(key)
		for _, c := range encoded {
			if c == '0' || c == 'O' || c == 'I' || c == 'l' {
				t.Errorf("EncodeKey() contains ambiguous character '%c' in %s", c, encoded)
			}
		}
	}
}
