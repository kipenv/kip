package handler

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/kipenv/kip/internal/crypto"
)

func startTestServer(t *testing.T) *httptest.Server {
	t.Helper()
	h := newTestHandler()

	mux := http.NewServeMux()
	mux.HandleFunc("POST /api/v1/secret", h.Create)
	mux.HandleFunc("GET /api/v1/secret/{id}", h.Get)

	return httptest.NewServer(mux)
}

func TestPushPullIntegration(t *testing.T) {
	ts := startTestServer(t)
	defer ts.Close()

	dir := t.TempDir()
	envPath := filepath.Join(dir, ".env")
	envContent := "DB_HOST=localhost\nDB_PORT=5432\nSECRET_KEY=supersecret\n"
	if err := os.WriteFile(envPath, []byte(envContent), 0o600); err != nil {
		t.Fatal(err)
	}

	plaintext, err := os.ReadFile(envPath)
	if err != nil {
		t.Fatal(err)
	}

	ciphertext, nonce, key, err := crypto.Encrypt(plaintext)
	if err != nil {
		t.Fatal(err)
	}

	reqBody := fmt.Sprintf(`{"ciphertext":"%s","nonce":"%s","filename":".env","max_reads":3,"ttl_seconds":3600}`,
		base64.StdEncoding.EncodeToString(ciphertext),
		base64.StdEncoding.EncodeToString(nonce))

	resp, err := http.Post(ts.URL+"/api/v1/secret", "application/json", strings.NewReader(reqBody))
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		t.Fatalf("create status = %d, body: %s", resp.StatusCode, body)
	}

	var createResp struct {
		ID string `json:"id"`
	}
	json.NewDecoder(resp.Body).Decode(&createResp)

	getResp, err := http.Get(ts.URL + "/api/v1/secret/" + createResp.ID)
	if err != nil {
		t.Fatal(err)
	}
	defer getResp.Body.Close()

	var secretResp GetResponse
	json.NewDecoder(getResp.Body).Decode(&secretResp)

	ct, _ := base64.StdEncoding.DecodeString(secretResp.Ciphertext)
	n, _ := base64.StdEncoding.DecodeString(secretResp.Nonce)

	decrypted, err := crypto.Decrypt(ct, n, key)
	if err != nil {
		t.Fatalf("decrypt: %v", err)
	}

	if string(decrypted) != envContent {
		t.Errorf("round-trip failed: got %q, want %q", decrypted, envContent)
	}
	if secretResp.Filename != ".env" {
		t.Errorf("filename = %q, want %q", secretResp.Filename, ".env")
	}
	if secretResp.ReadsLeft != 2 {
		t.Errorf("reads_left = %d, want 2", secretResp.ReadsLeft)
	}
}

func TestSecretAutoDeleteAfterReads(t *testing.T) {
	ts := startTestServer(t)
	defer ts.Close()

	plaintext := []byte("API_KEY=abc123")
	ciphertext, nonce, _, _ := crypto.Encrypt(plaintext)

	reqBody := fmt.Sprintf(`{"ciphertext":"%s","nonce":"%s","filename":".env","max_reads":2,"ttl_seconds":3600}`,
		base64.StdEncoding.EncodeToString(ciphertext),
		base64.StdEncoding.EncodeToString(nonce))

	resp, _ := http.Post(ts.URL+"/api/v1/secret", "application/json", strings.NewReader(reqBody))
	var createResp CreateResponse
	json.NewDecoder(resp.Body).Decode(&createResp)
	resp.Body.Close()

	// First read
	r1, _ := http.Get(ts.URL + "/api/v1/secret/" + createResp.ID)
	var s1 GetResponse
	json.NewDecoder(r1.Body).Decode(&s1)
	r1.Body.Close()
	if s1.ReadsLeft != 1 {
		t.Errorf("first read: reads_left = %d, want 1", s1.ReadsLeft)
	}

	// Second read
	r2, _ := http.Get(ts.URL + "/api/v1/secret/" + createResp.ID)
	var s2 GetResponse
	json.NewDecoder(r2.Body).Decode(&s2)
	r2.Body.Close()
	if s2.ReadsLeft != 0 {
		t.Errorf("second read: reads_left = %d, want 0", s2.ReadsLeft)
	}

	// Third read — should be gone
	r3, _ := http.Get(ts.URL + "/api/v1/secret/" + createResp.ID)
	if r3.StatusCode != http.StatusNotFound {
		t.Errorf("third read: status = %d, want 404", r3.StatusCode)
	}
	r3.Body.Close()
}

func TestPasswordProtectedRoundTrip(t *testing.T) {
	ts := startTestServer(t)
	defer ts.Close()

	plaintext := []byte("SECRET=super-private-data")
	password := []byte("mypassword123")

	// Step 1: Encrypt with random key
	innerCt, innerNonce, key, err := crypto.Encrypt(plaintext)
	if err != nil {
		t.Fatal(err)
	}

	// Step 2: Pack nonce + ciphertext, then encrypt with password
	packed := make([]byte, len(innerNonce)+len(innerCt))
	copy(packed, innerNonce)
	copy(packed[len(innerNonce):], innerCt)

	outerCt, outerNonce, salt, err := crypto.EncryptWithPassword(packed, password)
	if err != nil {
		t.Fatal(err)
	}

	// Step 3: Upload
	reqBody := fmt.Sprintf(`{"ciphertext":"%s","nonce":"%s","salt":"%s","filename":".env","max_reads":1,"ttl_seconds":3600,"password_protected":true}`,
		base64.StdEncoding.EncodeToString(outerCt),
		base64.StdEncoding.EncodeToString(outerNonce),
		base64.StdEncoding.EncodeToString(salt))

	resp, err := http.Post(ts.URL+"/api/v1/secret", "application/json", strings.NewReader(reqBody))
	if err != nil {
		t.Fatal(err)
	}
	var createResp struct{ ID string `json:"id"` }
	json.NewDecoder(resp.Body).Decode(&createResp)
	resp.Body.Close()

	// Step 4: Download
	getResp, _ := http.Get(ts.URL + "/api/v1/secret/" + createResp.ID)
	var secretResp GetResponse
	json.NewDecoder(getResp.Body).Decode(&secretResp)
	getResp.Body.Close()

	if !secretResp.PasswordProtected {
		t.Fatal("expected password_protected = true")
	}
	if secretResp.Salt == "" {
		t.Fatal("expected salt to be present")
	}

	// Step 5: Decrypt outer (password) layer
	ct, _ := base64.StdEncoding.DecodeString(secretResp.Ciphertext)
	n, _ := base64.StdEncoding.DecodeString(secretResp.Nonce)
	s, _ := base64.StdEncoding.DecodeString(secretResp.Salt)

	unpacked, err := crypto.DecryptWithPassword(ct, n, s, password)
	if err != nil {
		t.Fatalf("password decrypt: %v", err)
	}

	// Step 6: Unpack and decrypt inner layer
	innerN := unpacked[:crypto.NonceSize]
	innerC := unpacked[crypto.NonceSize:]

	decrypted, err := crypto.Decrypt(innerC, innerN, key)
	if err != nil {
		t.Fatalf("inner decrypt: %v", err)
	}

	if string(decrypted) != string(plaintext) {
		t.Errorf("round-trip: got %q, want %q", decrypted, plaintext)
	}
}
