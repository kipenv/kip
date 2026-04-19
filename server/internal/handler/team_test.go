package handler

import (
	"bytes"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"

	"github.com/kipenv/kip/server/internal/store"
)

func newTestTeamHandler(t *testing.T) *TeamHandler {
	t.Helper()
	dbPath := filepath.Join(t.TempDir(), "test.db")
	s, err := store.NewSQLiteStore(dbPath)
	if err != nil {
		t.Fatalf("NewSQLiteStore() error: %v", err)
	}
	t.Cleanup(func() { s.Close() })
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	return NewTeamHandler(s, logger)
}

func createTeamHelper(t *testing.T, h *TeamHandler, name, username string) CreateTeamResponse {
	t.Helper()
	body, _ := json.Marshal(CreateTeamRequest{Name: name, Username: username})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/teams", bytes.NewReader(body))
	rec := httptest.NewRecorder()
	h.Create(rec, req)
	if rec.Code != http.StatusCreated {
		t.Fatalf("create team: status = %d, body: %s", rec.Code, rec.Body.String())
	}
	var resp CreateTeamResponse
	json.NewDecoder(rec.Body).Decode(&resp)
	return resp
}

func TestTeamCreate(t *testing.T) {
	h := newTestTeamHandler(t)
	resp := createTeamHelper(t, h, "my-team", "alice")

	if resp.ID == "" {
		t.Error("team ID is empty")
	}
	if resp.Name != "my-team" {
		t.Errorf("name = %q, want %q", resp.Name, "my-team")
	}
	if resp.InviteCode == "" {
		t.Error("invite code is empty")
	}
	if resp.MemberToken == "" {
		t.Error("member token is empty")
	}
}

func TestTeamCreateDuplicate(t *testing.T) {
	h := newTestTeamHandler(t)
	createTeamHelper(t, h, "dup-team", "alice")

	body, _ := json.Marshal(CreateTeamRequest{Name: "dup-team", Username: "bob"})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/teams", bytes.NewReader(body))
	rec := httptest.NewRecorder()
	h.Create(rec, req)

	if rec.Code != http.StatusConflict {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusConflict)
	}
}

func TestTeamCreateMissingFields(t *testing.T) {
	h := newTestTeamHandler(t)

	body, _ := json.Marshal(CreateTeamRequest{Name: "", Username: ""})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/teams", bytes.NewReader(body))
	rec := httptest.NewRecorder()
	h.Create(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusBadRequest)
	}
}

func TestTeamJoin(t *testing.T) {
	h := newTestTeamHandler(t)
	created := createTeamHelper(t, h, "joinable", "alice")

	body, _ := json.Marshal(JoinTeamRequest{InviteCode: created.InviteCode, Username: "bob"})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/teams/join", bytes.NewReader(body))
	rec := httptest.NewRecorder()
	h.Join(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d. body: %s", rec.Code, http.StatusOK, rec.Body.String())
	}

	var resp JoinTeamResponse
	json.NewDecoder(rec.Body).Decode(&resp)
	if resp.ID != created.ID {
		t.Errorf("team ID = %q, want %q", resp.ID, created.ID)
	}
	if resp.MemberToken == "" {
		t.Error("member token is empty")
	}
}

func TestTeamJoinInvalidCode(t *testing.T) {
	h := newTestTeamHandler(t)

	body, _ := json.Marshal(JoinTeamRequest{InviteCode: "bad-code", Username: "bob"})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/teams/join", bytes.NewReader(body))
	rec := httptest.NewRecorder()
	h.Join(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusNotFound)
	}
}

func TestTeamJoinDuplicateUsername(t *testing.T) {
	h := newTestTeamHandler(t)
	created := createTeamHelper(t, h, "team1", "alice")

	body, _ := json.Marshal(JoinTeamRequest{InviteCode: created.InviteCode, Username: "alice"})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/teams/join", bytes.NewReader(body))
	rec := httptest.NewRecorder()
	h.Join(rec, req)

	if rec.Code != http.StatusConflict {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusConflict)
	}
}

func TestTeamMembers(t *testing.T) {
	h := newTestTeamHandler(t)
	created := createTeamHelper(t, h, "team1", "alice")

	// Add bob
	joinBody, _ := json.Marshal(JoinTeamRequest{InviteCode: created.InviteCode, Username: "bob"})
	joinReq := httptest.NewRequest(http.MethodPost, "/api/v1/teams/join", bytes.NewReader(joinBody))
	joinRec := httptest.NewRecorder()
	h.Join(joinRec, joinReq)

	// List members
	req := httptest.NewRequest(http.MethodGet, "/api/v1/teams/"+created.ID+"/members", nil)
	req.SetPathValue("id", created.ID)
	req.Header.Set("Authorization", "Bearer "+created.MemberToken)
	rec := httptest.NewRecorder()
	h.Members(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d. body: %s", rec.Code, http.StatusOK, rec.Body.String())
	}

	var members []MemberResponse
	json.NewDecoder(rec.Body).Decode(&members)
	if len(members) != 2 {
		t.Fatalf("got %d members, want 2", len(members))
	}
}

func TestTeamMembersUnauthorized(t *testing.T) {
	h := newTestTeamHandler(t)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/teams/abc/members", nil)
	req.SetPathValue("id", "abc")
	rec := httptest.NewRecorder()
	h.Members(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusUnauthorized)
	}
}

func TestTeamLeave(t *testing.T) {
	h := newTestTeamHandler(t)
	created := createTeamHelper(t, h, "team1", "alice")

	// Add bob
	joinBody, _ := json.Marshal(JoinTeamRequest{InviteCode: created.InviteCode, Username: "bob"})
	joinReq := httptest.NewRequest(http.MethodPost, "/api/v1/teams/join", bytes.NewReader(joinBody))
	joinRec := httptest.NewRecorder()
	h.Join(joinRec, joinReq)
	var joinResp JoinTeamResponse
	json.NewDecoder(joinRec.Body).Decode(&joinResp)

	// Bob leaves
	req := httptest.NewRequest(http.MethodDelete, "/api/v1/teams/"+created.ID+"/leave", nil)
	req.SetPathValue("id", created.ID)
	req.Header.Set("Authorization", "Bearer "+joinResp.MemberToken)
	rec := httptest.NewRecorder()
	h.Leave(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d. body: %s", rec.Code, http.StatusOK, rec.Body.String())
	}

	// Verify only alice remains
	memReq := httptest.NewRequest(http.MethodGet, "/api/v1/teams/"+created.ID+"/members", nil)
	memReq.SetPathValue("id", created.ID)
	memReq.Header.Set("Authorization", "Bearer "+created.MemberToken)
	memRec := httptest.NewRecorder()
	h.Members(memRec, memReq)

	var members []MemberResponse
	json.NewDecoder(memRec.Body).Decode(&members)
	if len(members) != 1 {
		t.Fatalf("got %d members, want 1", len(members))
	}
}

func TestTeamLeaveForbidden(t *testing.T) {
	h := newTestTeamHandler(t)
	created := createTeamHelper(t, h, "team1", "alice")
	created2 := createTeamHelper(t, h, "team2", "bob")

	// Alice tries to leave team2
	req := httptest.NewRequest(http.MethodDelete, "/api/v1/teams/"+created2.ID+"/leave", nil)
	req.SetPathValue("id", created2.ID)
	req.Header.Set("Authorization", "Bearer "+created.MemberToken)
	rec := httptest.NewRecorder()
	h.Leave(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusForbidden)
	}
}
