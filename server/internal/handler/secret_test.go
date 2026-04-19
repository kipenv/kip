package handler

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/kipenv/kip/server/internal/store"
)

func newTestHandler() *SecretHandler {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	return NewSecretHandler(store.NewMemoryStore(), logger)
}

func TestCreateSecret(t *testing.T) {
	h := newTestHandler()

	body := CreateRequest{
		Ciphertext: base64.StdEncoding.EncodeToString([]byte("encrypted")),
		Nonce:      base64.StdEncoding.EncodeToString([]byte("nonce12bytes")),
		Filename:   ".env",
		MaxReads:   1,
		TTLSeconds: 600,
	}
	b, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/secret", bytes.NewReader(b))
	rec := httptest.NewRecorder()

	h.Create(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("status = %d, want %d. body: %s", rec.Code, http.StatusCreated, rec.Body.String())
	}

	var resp CreateResponse
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if resp.ID == "" {
		t.Error("response ID is empty")
	}
}

func TestCreateSecretMissingFields(t *testing.T) {
	h := newTestHandler()

	body := `{"ciphertext":"","nonce":""}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/secret", bytes.NewBufferString(body))
	rec := httptest.NewRecorder()

	h.Create(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusBadRequest)
	}
}

func TestCreateSecretInvalidJSON(t *testing.T) {
	h := newTestHandler()

	req := httptest.NewRequest(http.MethodPost, "/api/v1/secret", bytes.NewBufferString("not-json"))
	rec := httptest.NewRecorder()

	h.Create(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusBadRequest)
	}
}

func TestCreateSecretWrongMethod(t *testing.T) {
	h := newTestHandler()

	req := httptest.NewRequest(http.MethodGet, "/api/v1/secret", nil)
	rec := httptest.NewRecorder()

	h.Create(rec, req)

	if rec.Code != http.StatusMethodNotAllowed {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusMethodNotAllowed)
	}
}

func TestGetSecret(t *testing.T) {
	h := newTestHandler()
	ctx := httptest.NewRequest(http.MethodGet, "/", nil).Context()

	// Create a secret first
	id, err := h.store.Create(ctx, store.CreateSecretInput{
		Ciphertext: []byte("encrypted"),
		Nonce:      []byte("nonce12bytes"),
		Filename:   ".env",
		MaxReads:   1,
		TTL:        10 * time.Minute,
	})
	if err != nil {
		t.Fatalf("store.Create() error: %v", err)
	}

	// Create a request with path value
	req := httptest.NewRequest(http.MethodGet, "/api/v1/secret/"+id, nil)
	req.SetPathValue("id", id)
	rec := httptest.NewRecorder()

	h.Get(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d. body: %s", rec.Code, http.StatusOK, rec.Body.String())
	}

	var resp GetResponse
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}

	ct, _ := base64.StdEncoding.DecodeString(resp.Ciphertext)
	if string(ct) != "encrypted" {
		t.Errorf("Ciphertext = %q, want %q", ct, "encrypted")
	}
	if resp.Filename != ".env" {
		t.Errorf("Filename = %q, want %q", resp.Filename, ".env")
	}
}

func TestGetSecretConsumed(t *testing.T) {
	h := newTestHandler()
	ctx := httptest.NewRequest(http.MethodGet, "/", nil).Context()

	id, _ := h.store.Create(ctx, store.CreateSecretInput{
		Ciphertext: []byte("data"),
		Nonce:      []byte("nonce"),
		MaxReads:   1,
		TTL:        10 * time.Minute,
	})

	// First read
	req := httptest.NewRequest(http.MethodGet, "/api/v1/secret/"+id, nil)
	req.SetPathValue("id", id)
	rec := httptest.NewRecorder()
	h.Get(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("first read: status = %d, want %d", rec.Code, http.StatusOK)
	}

	// Second read — should be gone
	req = httptest.NewRequest(http.MethodGet, "/api/v1/secret/"+id, nil)
	req.SetPathValue("id", id)
	rec = httptest.NewRecorder()
	h.Get(rec, req)
	if rec.Code != http.StatusNotFound {
		t.Errorf("second read: status = %d, want %d", rec.Code, http.StatusNotFound)
	}
}

func TestGetSecretNotFound(t *testing.T) {
	h := newTestHandler()

	req := httptest.NewRequest(http.MethodGet, "/api/v1/secret/nonexistent", nil)
	req.SetPathValue("id", "nonexistent")
	rec := httptest.NewRecorder()

	h.Get(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusNotFound)
	}
}

func TestGetSecretWrongMethod(t *testing.T) {
	h := newTestHandler()

	req := httptest.NewRequest(http.MethodPost, "/api/v1/secret/abc", nil)
	req.SetPathValue("id", "abc")
	rec := httptest.NewRecorder()

	h.Get(rec, req)

	if rec.Code != http.StatusMethodNotAllowed {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusMethodNotAllowed)
	}
}

func TestCreateSecretDefaultValues(t *testing.T) {
	h := newTestHandler()

	body := CreateRequest{
		Ciphertext: base64.StdEncoding.EncodeToString([]byte("data")),
		Nonce:      base64.StdEncoding.EncodeToString([]byte("nonce12bytes")),
		// No filename, max_reads, or ttl
	}
	b, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/secret", bytes.NewReader(b))
	rec := httptest.NewRecorder()

	h.Create(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusCreated)
	}
}

func TestDeleteSecret(t *testing.T) {
	h := newTestHandler()
	ctx := httptest.NewRequest(http.MethodGet, "/", nil).Context()

	id, err := h.store.Create(ctx, store.CreateSecretInput{
		Ciphertext: []byte("encrypted"),
		Nonce:      []byte("nonce12bytes"),
		Filename:   ".env",
		MaxReads:   5,
		TTL:        10 * time.Minute,
	})
	if err != nil {
		t.Fatalf("store.Create() error: %v", err)
	}

	// Delete the secret
	req := httptest.NewRequest(http.MethodDelete, "/api/v1/secret/"+id, nil)
	req.SetPathValue("id", id)
	rec := httptest.NewRecorder()
	h.Delete(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Fatalf("delete: status = %d, want %d. body: %s", rec.Code, http.StatusNoContent, rec.Body.String())
	}

	// Verify it's gone
	req = httptest.NewRequest(http.MethodGet, "/api/v1/secret/"+id, nil)
	req.SetPathValue("id", id)
	rec = httptest.NewRecorder()
	h.Get(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Errorf("get after delete: status = %d, want %d", rec.Code, http.StatusNotFound)
	}
}

func TestDeleteSecretNotFound(t *testing.T) {
	h := newTestHandler()

	req := httptest.NewRequest(http.MethodDelete, "/api/v1/secret/nonexistent", nil)
	req.SetPathValue("id", "nonexistent")
	rec := httptest.NewRecorder()

	h.Delete(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusNotFound)
	}
}

func TestDeleteSecretWrongMethod(t *testing.T) {
	h := newTestHandler()

	req := httptest.NewRequest(http.MethodPost, "/api/v1/secret/abc", nil)
	req.SetPathValue("id", "abc")
	rec := httptest.NewRecorder()

	h.Delete(rec, req)

	if rec.Code != http.StatusMethodNotAllowed {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusMethodNotAllowed)
	}
}

func TestCreateAndGetRoundTrip(t *testing.T) {
	h := newTestHandler()

	originalData := []byte("DB_HOST=localhost\nDB_PORT=5432")

	// Create
	createBody := CreateRequest{
		Ciphertext: base64.StdEncoding.EncodeToString(originalData),
		Nonce:      base64.StdEncoding.EncodeToString([]byte("nonce12bytes")),
		Filename:   ".env.production",
		MaxReads:   3,
		TTLSeconds: 3600,
	}
	b, _ := json.Marshal(createBody)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/secret", bytes.NewReader(b))
	rec := httptest.NewRecorder()
	h.Create(rec, req)

	var createResp CreateResponse
	json.NewDecoder(rec.Body).Decode(&createResp)

	// Get
	req = httptest.NewRequest(http.MethodGet, "/api/v1/secret/"+createResp.ID, nil)
	req.SetPathValue("id", createResp.ID)
	rec = httptest.NewRecorder()
	h.Get(rec, req)

	var getResp GetResponse
	json.NewDecoder(rec.Body).Decode(&getResp)

	ct, _ := base64.StdEncoding.DecodeString(getResp.Ciphertext)
	if string(ct) != string(originalData) {
		t.Errorf("round-trip: got %q, want %q", ct, originalData)
	}
	if getResp.Filename != ".env.production" {
		t.Errorf("Filename = %q, want %q", getResp.Filename, ".env.production")
	}
	if getResp.ReadsLeft != 2 {
		t.Errorf("ReadsLeft = %d, want 2", getResp.ReadsLeft)
	}
}
