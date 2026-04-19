package handler

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"

	"github.com/kipenv/kip/server/internal/store"
)

func newTestShareHandler(t *testing.T) (*ShareHandler, *TeamHandler) {
	t.Helper()
	dbPath := filepath.Join(t.TempDir(), "test.db")
	teamStore, err := store.NewSQLiteStore(dbPath)
	if err != nil {
		t.Fatalf("NewSQLiteStore() error: %v", err)
	}
	t.Cleanup(func() { teamStore.Close() })
	secretStore := store.NewMemoryStore()
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	return NewShareHandler(teamStore, secretStore, logger), NewTeamHandler(teamStore, logger)
}

func setupTeamWithMembers(t *testing.T, th *TeamHandler) (teamID, aliceToken, bobToken string) {
	t.Helper()
	created := createTeamHelper(t, th, "test-team", "alice")

	joinBody, _ := json.Marshal(JoinTeamRequest{InviteCode: created.InviteCode, Username: "bob"})
	joinReq := httptest.NewRequest(http.MethodPost, "/api/v1/teams/join", bytes.NewReader(joinBody))
	joinRec := httptest.NewRecorder()
	th.Join(joinRec, joinReq)
	var joinResp JoinTeamResponse
	json.NewDecoder(joinRec.Body).Decode(&joinResp)

	return created.ID, created.MemberToken, joinResp.MemberToken
}

func TestShareToTeam(t *testing.T) {
	sh, th := newTestShareHandler(t)
	teamID, aliceToken, _ := setupTeamWithMembers(t, th)

	body, _ := json.Marshal(ShareRequest{
		Ciphertext: base64.StdEncoding.EncodeToString([]byte("encrypted")),
		Nonce:      base64.StdEncoding.EncodeToString([]byte("nonce12bytes")),
		Filename:   ".env",
		TTLSeconds: 3600,
	})

	req := httptest.NewRequest(http.MethodPost, "/api/v1/teams/"+teamID+"/share", bytes.NewReader(body))
	req.SetPathValue("id", teamID)
	req.Header.Set("Authorization", "Bearer "+aliceToken)
	rec := httptest.NewRecorder()
	sh.Share(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("status = %d, want %d. body: %s", rec.Code, http.StatusCreated, rec.Body.String())
	}

	var resp ShareResponse
	json.NewDecoder(rec.Body).Decode(&resp)
	if resp.ID == "" {
		t.Error("share ID is empty")
	}
}

func TestShareToSpecificUser(t *testing.T) {
	sh, th := newTestShareHandler(t)
	teamID, aliceToken, _ := setupTeamWithMembers(t, th)

	body, _ := json.Marshal(ShareRequest{
		Ciphertext: base64.StdEncoding.EncodeToString([]byte("data")),
		Nonce:      base64.StdEncoding.EncodeToString([]byte("nonce")),
		ToUser:     "bob",
		TTLSeconds: 3600,
	})

	req := httptest.NewRequest(http.MethodPost, "/api/v1/teams/"+teamID+"/share", bytes.NewReader(body))
	req.SetPathValue("id", teamID)
	req.Header.Set("Authorization", "Bearer "+aliceToken)
	rec := httptest.NewRecorder()
	sh.Share(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusCreated)
	}
}

func TestShareUnauthorized(t *testing.T) {
	sh, th := newTestShareHandler(t)
	teamID, _, _ := setupTeamWithMembers(t, th)

	body, _ := json.Marshal(ShareRequest{
		Ciphertext: base64.StdEncoding.EncodeToString([]byte("data")),
		Nonce:      base64.StdEncoding.EncodeToString([]byte("nonce")),
	})

	req := httptest.NewRequest(http.MethodPost, "/api/v1/teams/"+teamID+"/share", bytes.NewReader(body))
	req.SetPathValue("id", teamID)
	rec := httptest.NewRecorder()
	sh.Share(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusUnauthorized)
	}
}

func TestInbox(t *testing.T) {
	sh, th := newTestShareHandler(t)
	teamID, aliceToken, bobToken := setupTeamWithMembers(t, th)

	// Alice shares to all
	body, _ := json.Marshal(ShareRequest{
		Ciphertext: base64.StdEncoding.EncodeToString([]byte("data")),
		Nonce:      base64.StdEncoding.EncodeToString([]byte("nonce")),
		Filename:   ".env",
		TTLSeconds: 3600,
	})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/teams/"+teamID+"/share", bytes.NewReader(body))
	req.SetPathValue("id", teamID)
	req.Header.Set("Authorization", "Bearer "+aliceToken)
	rec := httptest.NewRecorder()
	sh.Share(rec, req)

	// Bob checks inbox
	inboxReq := httptest.NewRequest(http.MethodGet, "/api/v1/teams/"+teamID+"/inbox", nil)
	inboxReq.SetPathValue("id", teamID)
	inboxReq.Header.Set("Authorization", "Bearer "+bobToken)
	inboxRec := httptest.NewRecorder()
	sh.Inbox(inboxRec, inboxReq)

	if inboxRec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", inboxRec.Code, http.StatusOK)
	}

	var inbox []InboxEntry
	json.NewDecoder(inboxRec.Body).Decode(&inbox)
	if len(inbox) != 1 {
		t.Fatalf("got %d inbox items, want 1", len(inbox))
	}
	if inbox[0].FromUser != "alice" {
		t.Errorf("from_user = %q, want %q", inbox[0].FromUser, "alice")
	}
}

func TestShareLog(t *testing.T) {
	sh, th := newTestShareHandler(t)
	teamID, aliceToken, _ := setupTeamWithMembers(t, th)

	// Create two shares
	for i := 0; i < 2; i++ {
		body, _ := json.Marshal(ShareRequest{
			Ciphertext: base64.StdEncoding.EncodeToString([]byte("data")),
			Nonce:      base64.StdEncoding.EncodeToString([]byte("nonce")),
			TTLSeconds: 3600,
		})
		req := httptest.NewRequest(http.MethodPost, "/api/v1/teams/"+teamID+"/share", bytes.NewReader(body))
		req.SetPathValue("id", teamID)
		req.Header.Set("Authorization", "Bearer "+aliceToken)
		rec := httptest.NewRecorder()
		sh.Share(rec, req)
	}

	// Check log
	logReq := httptest.NewRequest(http.MethodGet, "/api/v1/teams/"+teamID+"/log", nil)
	logReq.SetPathValue("id", teamID)
	logReq.Header.Set("Authorization", "Bearer "+aliceToken)
	logRec := httptest.NewRecorder()
	sh.Log(logRec, logReq)

	if logRec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", logRec.Code, http.StatusOK)
	}

	var log []InboxEntry
	json.NewDecoder(logRec.Body).Decode(&log)
	if len(log) != 2 {
		t.Fatalf("got %d log entries, want 2", len(log))
	}
}

func TestPinUpsertAndGet(t *testing.T) {
	sh, th := newTestShareHandler(t)
	teamID, aliceToken, bobToken := setupTeamWithMembers(t, th)

	// Alice pins .env
	body, _ := json.Marshal(PinRequest{
		Ciphertext: base64.StdEncoding.EncodeToString([]byte("pinned-data")),
		Nonce:      base64.StdEncoding.EncodeToString([]byte("nonce12bytes")),
		Filename:   ".env",
	})
	req := httptest.NewRequest(http.MethodPut, "/api/v1/teams/"+teamID+"/pin", bytes.NewReader(body))
	req.SetPathValue("id", teamID)
	req.Header.Set("Authorization", "Bearer "+aliceToken)
	rec := httptest.NewRecorder()
	sh.PinUpsert(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("pin status = %d, want %d. body: %s", rec.Code, http.StatusOK, rec.Body.String())
	}

	// Bob gets pins
	getReq := httptest.NewRequest(http.MethodGet, "/api/v1/teams/"+teamID+"/pin", nil)
	getReq.SetPathValue("id", teamID)
	getReq.Header.Set("Authorization", "Bearer "+bobToken)
	getRec := httptest.NewRecorder()
	sh.PinGet(getRec, getReq)

	if getRec.Code != http.StatusOK {
		t.Fatalf("get pins status = %d, want %d", getRec.Code, http.StatusOK)
	}

	var pins []PinResponse
	json.NewDecoder(getRec.Body).Decode(&pins)
	if len(pins) != 1 {
		t.Fatalf("got %d pins, want 1", len(pins))
	}
	if pins[0].Filename != ".env" {
		t.Errorf("filename = %q, want %q", pins[0].Filename, ".env")
	}
	if pins[0].PinnedBy != "alice" {
		t.Errorf("pinned_by = %q, want %q", pins[0].PinnedBy, "alice")
	}
}

func TestPinForbidden(t *testing.T) {
	sh, th := newTestShareHandler(t)
	teamID, _, _ := setupTeamWithMembers(t, th)

	// Create another team and use that token
	other := createTeamHelper(t, th, "other-team", "charlie")

	body, _ := json.Marshal(PinRequest{
		Ciphertext: base64.StdEncoding.EncodeToString([]byte("data")),
		Nonce:      base64.StdEncoding.EncodeToString([]byte("nonce")),
	})
	req := httptest.NewRequest(http.MethodPut, "/api/v1/teams/"+teamID+"/pin", bytes.NewReader(body))
	req.SetPathValue("id", teamID)
	req.Header.Set("Authorization", "Bearer "+other.MemberToken)
	rec := httptest.NewRecorder()
	sh.PinUpsert(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusForbidden)
	}
}
