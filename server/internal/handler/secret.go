// Package handler provides HTTP handlers for the kip server.
package handler

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"time"

	"github.com/kipenv/kip/server/internal/store"
)

const (
	maxPayloadSize = 5 * 1024 * 1024 // 5 MB
	minTTL         = time.Minute
	maxTTL         = 30 * 24 * time.Hour // 30 days
	maxReads       = 100
)

// SecretHandler handles secret push/pull endpoints.
type SecretHandler struct {
	store  store.SecretStore
	logger *slog.Logger
}

// NewSecretHandler creates a new handler with the given store.
func NewSecretHandler(s store.SecretStore, logger *slog.Logger) *SecretHandler {
	return &SecretHandler{store: s, logger: logger}
}

// CreateRequest is the JSON body for creating a secret.
type CreateRequest struct {
	Ciphertext        string `json:"ciphertext"`     // base64 encoded
	Nonce             string `json:"nonce"`          // base64 encoded
	Salt              string `json:"salt,omitempty"` // base64 encoded, for password-protected
	Filename          string `json:"filename"`
	MaxReads          int    `json:"max_reads"`
	TTLSeconds        int    `json:"ttl_seconds"`
	PasswordProtected bool   `json:"password_protected,omitempty"`
}

// CreateResponse is the JSON response after creating a secret.
type CreateResponse struct {
	ID        string    `json:"id"`
	ExpiresAt time.Time `json:"expires_at"`
}

// GetResponse is the JSON response when retrieving a secret.
type GetResponse struct {
	Ciphertext        string `json:"ciphertext"`     // base64 encoded
	Nonce             string `json:"nonce"`          // base64 encoded
	Salt              string `json:"salt,omitempty"` // base64 encoded
	Filename          string `json:"filename"`
	ReadsLeft         int    `json:"reads_left"`
	PasswordProtected bool   `json:"password_protected,omitempty"`
}

// Create handles POST /api/v1/secret
func (h *SecretHandler) Create(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, maxPayloadSize)

	var req CreateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	// Validate
	if req.Ciphertext == "" || req.Nonce == "" {
		writeError(w, http.StatusBadRequest, "ciphertext and nonce are required")
		return
	}

	ciphertext, err := base64.StdEncoding.DecodeString(req.Ciphertext)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid ciphertext encoding")
		return
	}

	nonce, err := base64.StdEncoding.DecodeString(req.Nonce)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid nonce encoding")
		return
	}

	if req.MaxReads < 1 {
		req.MaxReads = 1
	}
	if req.MaxReads > maxReads {
		req.MaxReads = maxReads
	}

	ttl := time.Duration(req.TTLSeconds) * time.Second
	if ttl < minTTL {
		ttl = minTTL
	}
	if ttl > maxTTL {
		ttl = maxTTL
	}

	filename := req.Filename
	if filename == "" {
		filename = ".env"
	}

	var salt []byte
	if req.Salt != "" {
		salt, err = base64.StdEncoding.DecodeString(req.Salt)
		if err != nil {
			writeError(w, http.StatusBadRequest, "invalid salt encoding")
			return
		}
	}

	id, err := h.store.Create(r.Context(), store.CreateSecretInput{
		Ciphertext:        ciphertext,
		Nonce:             nonce,
		Salt:              salt,
		Filename:          filename,
		MaxReads:          req.MaxReads,
		PasswordProtected: req.PasswordProtected,
		TTL:               ttl,
	})
	if err != nil {
		h.logger.Error("failed to create secret", "error", err)
		writeError(w, http.StatusInternalServerError, "failed to store secret")
		return
	}

	writeJSON(w, http.StatusCreated, CreateResponse{
		ID:        id,
		ExpiresAt: time.Now().UTC().Add(ttl),
	})
}

// Get handles GET /api/v1/secret/{id}
func (h *SecretHandler) Get(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	id := r.PathValue("id")
	if id == "" {
		writeError(w, http.StatusBadRequest, "missing secret id")
		return
	}

	secret, err := h.store.Get(r.Context(), id)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			writeError(w, http.StatusNotFound, "secret not found or already consumed")
			return
		}
		h.logger.Error("failed to get secret", "error", err, "id", id)
		writeError(w, http.StatusInternalServerError, "failed to retrieve secret")
		return
	}

	resp := GetResponse{
		Ciphertext:        base64.StdEncoding.EncodeToString(secret.Ciphertext),
		Nonce:             base64.StdEncoding.EncodeToString(secret.Nonce),
		Filename:          secret.Filename,
		ReadsLeft:         secret.ReadsLeft,
		PasswordProtected: secret.PasswordProtected,
	}
	if len(secret.Salt) > 0 {
		resp.Salt = base64.StdEncoding.EncodeToString(secret.Salt)
	}

	writeJSON(w, http.StatusOK, resp)
}

// Delete handles DELETE /api/v1/secret/{id}
func (h *SecretHandler) Delete(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	id := r.PathValue("id")
	if id == "" {
		writeError(w, http.StatusBadRequest, "missing secret id")
		return
	}

	err := h.store.Delete(r.Context(), id)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			writeError(w, http.StatusNotFound, "secret not found or already consumed")
			return
		}
		h.logger.Error("failed to delete secret", "error", err, "id", id)
		writeError(w, http.StatusInternalServerError, "failed to delete secret")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v) //nolint:errcheck
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{"error": message})
}
