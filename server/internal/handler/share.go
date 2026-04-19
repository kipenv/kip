package handler

import (
	"encoding/base64"
	"encoding/json"
	"log/slog"
	"net/http"
	"time"

	"github.com/kipenv/kip/server/internal/store"
)

// ShareHandler handles team sharing endpoints.
type ShareHandler struct {
	teams   store.TeamStore
	secrets store.SecretStore
	logger  *slog.Logger
}

// NewShareHandler creates a new share handler.
func NewShareHandler(teams store.TeamStore, secrets store.SecretStore, logger *slog.Logger) *ShareHandler {
	return &ShareHandler{teams: teams, secrets: secrets, logger: logger}
}

// ShareRequest is the JSON body for sharing with a team.
type ShareRequest struct {
	Ciphertext string `json:"ciphertext"` // base64
	Nonce      string `json:"nonce"`      // base64
	Filename   string `json:"filename"`
	ToUser     string `json:"to_user,omitempty"`
	TTLSeconds int    `json:"ttl_seconds"`
}

// ShareResponse is the JSON response after sharing.
type ShareResponse struct {
	ID        string    `json:"id"`
	ExpiresAt time.Time `json:"expires_at"`
}

// PinRequest is the JSON body for pinning an env.
type PinRequest struct {
	Ciphertext string `json:"ciphertext"` // base64
	Nonce      string `json:"nonce"`      // base64
	Filename   string `json:"filename"`
}

// PinResponse is the JSON response for a pin.
type PinResponse struct {
	Filename   string `json:"filename"`
	Ciphertext string `json:"ciphertext"` // base64
	Nonce      string `json:"nonce"`      // base64
	PinnedBy   string `json:"pinned_by"`
	UpdatedAt  string `json:"updated_at"`
}

// InboxEntry is a share in the inbox.
type InboxEntry struct {
	ID        string `json:"id"`
	FromUser  string `json:"from_user"`
	Filename  string `json:"filename"`
	SecretID  string `json:"secret_id"`
	ExpiresAt string `json:"expires_at"`
	CreatedAt string `json:"created_at"`
}

// Share handles POST /api/v1/teams/{id}/share
func (h *ShareHandler) Share(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	member, ok := h.authenticate(w, r)
	if !ok {
		return
	}

	teamID := r.PathValue("id")
	if member.TeamID != teamID {
		writeError(w, http.StatusForbidden, "not a member of this team")
		return
	}

	var req ShareRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

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

	filename := req.Filename
	if filename == "" {
		filename = ".env"
	}

	ttl := time.Duration(req.TTLSeconds) * time.Second
	if ttl < minTTL {
		ttl = minTTL
	}
	if ttl > maxTTL {
		ttl = maxTTL
	}

	// Store the encrypted data in the secret store (Redis/memory)
	secretID, err := h.secrets.Create(r.Context(), store.CreateSecretInput{
		Ciphertext: ciphertext,
		Nonce:      nonce,
		Filename:   filename,
		MaxReads:   maxReads,
		TTL:        ttl,
	})
	if err != nil {
		h.logger.Error("failed to store shared secret", "error", err)
		writeError(w, http.StatusInternalServerError, "failed to store secret")
		return
	}

	// Log the share in SQLite
	entry, err := h.teams.CreateShareLog(r.Context(), store.CreateShareInput{
		TeamID:   teamID,
		FromUser: member.Username,
		ToUser:   req.ToUser,
		Filename: filename,
		SecretID: secretID,
		TTL:      ttl,
	})
	if err != nil {
		h.logger.Error("failed to log share", "error", err)
		writeError(w, http.StatusInternalServerError, "failed to log share")
		return
	}

	writeJSON(w, http.StatusCreated, ShareResponse{
		ID:        entry.ID,
		ExpiresAt: entry.ExpiresAt,
	})
}

// Inbox handles GET /api/v1/teams/{id}/inbox
func (h *ShareHandler) Inbox(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	member, ok := h.authenticate(w, r)
	if !ok {
		return
	}

	teamID := r.PathValue("id")
	if member.TeamID != teamID {
		writeError(w, http.StatusForbidden, "not a member of this team")
		return
	}

	entries, err := h.teams.GetInbox(r.Context(), teamID, member.Username)
	if err != nil {
		h.logger.Error("failed to get inbox", "error", err)
		writeError(w, http.StatusInternalServerError, "failed to get inbox")
		return
	}

	resp := make([]InboxEntry, len(entries))
	for i, e := range entries {
		resp[i] = InboxEntry{
			ID:        e.ID,
			FromUser:  e.FromUser,
			Filename:  e.Filename,
			SecretID:  e.SecretID,
			ExpiresAt: e.ExpiresAt.Format(time.RFC3339),
			CreatedAt: e.CreatedAt.Format(time.RFC3339),
		}
	}

	writeJSON(w, http.StatusOK, resp)
}

// Log handles GET /api/v1/teams/{id}/log
func (h *ShareHandler) Log(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	member, ok := h.authenticate(w, r)
	if !ok {
		return
	}

	teamID := r.PathValue("id")
	if member.TeamID != teamID {
		writeError(w, http.StatusForbidden, "not a member of this team")
		return
	}

	entries, err := h.teams.GetShareLog(r.Context(), teamID)
	if err != nil {
		h.logger.Error("failed to get share log", "error", err)
		writeError(w, http.StatusInternalServerError, "failed to get share log")
		return
	}

	resp := make([]InboxEntry, len(entries))
	for i, e := range entries {
		resp[i] = InboxEntry{
			ID:        e.ID,
			FromUser:  e.FromUser,
			Filename:  e.Filename,
			SecretID:  e.SecretID,
			ExpiresAt: e.ExpiresAt.Format(time.RFC3339),
			CreatedAt: e.CreatedAt.Format(time.RFC3339),
		}
	}

	writeJSON(w, http.StatusOK, resp)
}

// PinUpsert handles PUT /api/v1/teams/{id}/pin
func (h *ShareHandler) PinUpsert(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	member, ok := h.authenticate(w, r)
	if !ok {
		return
	}

	teamID := r.PathValue("id")
	if member.TeamID != teamID {
		writeError(w, http.StatusForbidden, "not a member of this team")
		return
	}

	var req PinRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

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

	filename := req.Filename
	if filename == "" {
		filename = ".env"
	}

	pin, err := h.teams.UpsertPin(r.Context(), store.UpsertPinInput{
		TeamID:     teamID,
		Filename:   filename,
		Ciphertext: ciphertext,
		Nonce:      nonce,
		PinnedBy:   member.Username,
	})
	if err != nil {
		h.logger.Error("failed to upsert pin", "error", err)
		writeError(w, http.StatusInternalServerError, "failed to pin env")
		return
	}

	writeJSON(w, http.StatusOK, PinResponse{
		Filename:   pin.Filename,
		Ciphertext: base64.StdEncoding.EncodeToString(pin.Ciphertext),
		Nonce:      base64.StdEncoding.EncodeToString(pin.Nonce),
		PinnedBy:   pin.PinnedBy,
		UpdatedAt:  pin.UpdatedAt.Format(time.RFC3339),
	})
}

// PinGet handles GET /api/v1/teams/{id}/pin
func (h *ShareHandler) PinGet(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	member, ok := h.authenticate(w, r)
	if !ok {
		return
	}

	teamID := r.PathValue("id")
	if member.TeamID != teamID {
		writeError(w, http.StatusForbidden, "not a member of this team")
		return
	}

	pins, err := h.teams.GetPins(r.Context(), teamID)
	if err != nil {
		h.logger.Error("failed to get pins", "error", err)
		writeError(w, http.StatusInternalServerError, "failed to get pins")
		return
	}

	resp := make([]PinResponse, len(pins))
	for i, p := range pins {
		resp[i] = PinResponse{
			Filename:   p.Filename,
			Ciphertext: base64.StdEncoding.EncodeToString(p.Ciphertext),
			Nonce:      base64.StdEncoding.EncodeToString(p.Nonce),
			PinnedBy:   p.PinnedBy,
			UpdatedAt:  p.UpdatedAt.Format(time.RFC3339),
		}
	}

	writeJSON(w, http.StatusOK, resp)
}

// authenticate extracts and validates the bearer token.
func (h *ShareHandler) authenticate(w http.ResponseWriter, r *http.Request) (*store.Member, bool) {
	token := extractToken(r)
	if token == "" {
		writeError(w, http.StatusUnauthorized, "missing authorization token")
		return nil, false
	}

	member, err := h.teams.GetMemberByToken(r.Context(), token)
	if err != nil {
		writeError(w, http.StatusUnauthorized, "invalid token")
		return nil, false
	}

	return member, true
}
