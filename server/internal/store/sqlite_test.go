package store

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func newTestSQLiteStore(t *testing.T) *SQLiteStore {
	t.Helper()
	dbPath := filepath.Join(t.TempDir(), "test.db")
	s, err := NewSQLiteStore(dbPath)
	if err != nil {
		t.Fatalf("NewSQLiteStore() error: %v", err)
	}
	t.Cleanup(func() { s.Close() })
	return s
}

func TestSQLiteCreateTeam(t *testing.T) {
	s := newTestSQLiteStore(t)
	ctx := context.Background()

	out, err := s.CreateTeam(ctx, CreateTeamInput{
		Name:     "my-startup",
		Username: "juan",
	})
	if err != nil {
		t.Fatalf("CreateTeam() error: %v", err)
	}

	if out.Team.ID == "" {
		t.Error("team ID is empty")
	}
	if out.Team.Name != "my-startup" {
		t.Errorf("team name = %q, want %q", out.Team.Name, "my-startup")
	}
	if out.MemberToken == "" {
		t.Error("member token is empty")
	}
	if out.Team.InviteCode == "" {
		t.Error("invite code is empty")
	}
}

func TestSQLiteCreateTeamDuplicateName(t *testing.T) {
	s := newTestSQLiteStore(t)
	ctx := context.Background()

	_, err := s.CreateTeam(ctx, CreateTeamInput{Name: "dup", Username: "a"})
	if err != nil {
		t.Fatal(err)
	}

	_, err = s.CreateTeam(ctx, CreateTeamInput{Name: "dup", Username: "b"})
	if err == nil {
		t.Error("expected error for duplicate team name")
	}
}

func TestSQLiteJoinTeam(t *testing.T) {
	s := newTestSQLiteStore(t)
	ctx := context.Background()

	created, _ := s.CreateTeam(ctx, CreateTeamInput{Name: "team1", Username: "alice"})

	joined, err := s.JoinTeam(ctx, JoinTeamInput{
		InviteCode: created.Team.InviteCode,
		Username:   "bob",
	})
	if err != nil {
		t.Fatalf("JoinTeam() error: %v", err)
	}

	if joined.Team.ID != created.Team.ID {
		t.Errorf("joined team ID = %q, want %q", joined.Team.ID, created.Team.ID)
	}
	if joined.MemberToken == "" {
		t.Error("member token is empty")
	}
	if joined.MemberToken == created.MemberToken {
		t.Error("tokens should be different for different members")
	}
}

func TestSQLiteJoinTeamInvalidCode(t *testing.T) {
	s := newTestSQLiteStore(t)
	ctx := context.Background()

	_, err := s.JoinTeam(ctx, JoinTeamInput{InviteCode: "invalid", Username: "bob"})
	if err != ErrNotFound {
		t.Errorf("JoinTeam() error = %v, want ErrNotFound", err)
	}
}

func TestSQLiteJoinTeamDuplicateUsername(t *testing.T) {
	s := newTestSQLiteStore(t)
	ctx := context.Background()

	created, _ := s.CreateTeam(ctx, CreateTeamInput{Name: "team1", Username: "alice"})

	_, err := s.JoinTeam(ctx, JoinTeamInput{InviteCode: created.Team.InviteCode, Username: "alice"})
	if err == nil {
		t.Error("expected error for duplicate username in same team")
	}
}

func TestSQLiteGetTeamMembers(t *testing.T) {
	s := newTestSQLiteStore(t)
	ctx := context.Background()

	created, _ := s.CreateTeam(ctx, CreateTeamInput{Name: "team1", Username: "alice"})
	s.JoinTeam(ctx, JoinTeamInput{InviteCode: created.Team.InviteCode, Username: "bob"})

	members, err := s.GetTeamMembers(ctx, created.Team.ID)
	if err != nil {
		t.Fatalf("GetTeamMembers() error: %v", err)
	}
	if len(members) != 2 {
		t.Fatalf("got %d members, want 2", len(members))
	}
	if members[0].Username != "alice" {
		t.Errorf("first member = %q, want %q", members[0].Username, "alice")
	}
	if members[1].Username != "bob" {
		t.Errorf("second member = %q, want %q", members[1].Username, "bob")
	}
}

func TestSQLiteLeaveTeam(t *testing.T) {
	s := newTestSQLiteStore(t)
	ctx := context.Background()

	created, _ := s.CreateTeam(ctx, CreateTeamInput{Name: "team1", Username: "alice"})
	joined, _ := s.JoinTeam(ctx, JoinTeamInput{InviteCode: created.Team.InviteCode, Username: "bob"})

	// Get bob's member ID
	bob, _ := s.GetMemberByToken(ctx, joined.MemberToken)

	err := s.LeaveTeam(ctx, created.Team.ID, bob.ID)
	if err != nil {
		t.Fatalf("LeaveTeam() error: %v", err)
	}

	members, _ := s.GetTeamMembers(ctx, created.Team.ID)
	if len(members) != 1 {
		t.Fatalf("got %d members, want 1", len(members))
	}
}

func TestSQLiteLeaveTeamDeletesEmptyTeam(t *testing.T) {
	s := newTestSQLiteStore(t)
	ctx := context.Background()

	created, _ := s.CreateTeam(ctx, CreateTeamInput{Name: "solo", Username: "alice"})
	alice, _ := s.GetMemberByToken(ctx, created.MemberToken)

	err := s.LeaveTeam(ctx, created.Team.ID, alice.ID)
	if err != nil {
		t.Fatalf("LeaveTeam() error: %v", err)
	}

	_, err = s.GetTeam(ctx, created.Team.ID)
	if err != ErrNotFound {
		t.Errorf("GetTeam() after last member left: error = %v, want ErrNotFound", err)
	}
}

func TestSQLiteGetMemberByToken(t *testing.T) {
	s := newTestSQLiteStore(t)
	ctx := context.Background()

	created, _ := s.CreateTeam(ctx, CreateTeamInput{Name: "team1", Username: "alice"})

	m, err := s.GetMemberByToken(ctx, created.MemberToken)
	if err != nil {
		t.Fatalf("GetMemberByToken() error: %v", err)
	}
	if m.Username != "alice" {
		t.Errorf("username = %q, want %q", m.Username, "alice")
	}
}

func TestSQLiteGetMemberByTokenNotFound(t *testing.T) {
	s := newTestSQLiteStore(t)
	ctx := context.Background()

	_, err := s.GetMemberByToken(ctx, "invalid-token")
	if err != ErrNotFound {
		t.Errorf("error = %v, want ErrNotFound", err)
	}
}

func TestSQLiteGetTeamsByMember(t *testing.T) {
	s := newTestSQLiteStore(t)
	ctx := context.Background()

	t1, _ := s.CreateTeam(ctx, CreateTeamInput{Name: "alpha", Username: "alice"})
	t2, _ := s.CreateTeam(ctx, CreateTeamInput{Name: "beta", Username: "bob"})

	// Alice joins beta too
	s.JoinTeam(ctx, JoinTeamInput{InviteCode: t2.Team.InviteCode, Username: "alice"})

	teams, err := s.GetTeamsByMember(ctx, t1.MemberToken)
	if err != nil {
		t.Fatalf("GetTeamsByMember() error: %v", err)
	}

	// Alice's original token is only for team alpha
	// She has a different token for beta
	if len(teams) != 1 {
		t.Fatalf("got %d teams, want 1", len(teams))
	}
	if teams[0].Name != "alpha" {
		t.Errorf("team name = %q, want %q", teams[0].Name, "alpha")
	}
}

func TestSQLiteUpsertPin(t *testing.T) {
	s := newTestSQLiteStore(t)
	ctx := context.Background()

	created, _ := s.CreateTeam(ctx, CreateTeamInput{Name: "team1", Username: "alice"})

	pin, err := s.UpsertPin(ctx, UpsertPinInput{
		TeamID:     created.Team.ID,
		Filename:   ".env",
		Ciphertext: []byte("encrypted"),
		Nonce:      []byte("nonce"),
		PinnedBy:   "alice",
	})
	if err != nil {
		t.Fatalf("UpsertPin() error: %v", err)
	}
	if pin.Filename != ".env" {
		t.Errorf("filename = %q, want %q", pin.Filename, ".env")
	}

	// Upsert same filename — should update
	pin2, err := s.UpsertPin(ctx, UpsertPinInput{
		TeamID:     created.Team.ID,
		Filename:   ".env",
		Ciphertext: []byte("updated"),
		Nonce:      []byte("nonce2"),
		PinnedBy:   "bob",
	})
	if err != nil {
		t.Fatalf("UpsertPin() update error: %v", err)
	}

	got, _ := s.GetPin(ctx, created.Team.ID, ".env")
	if string(got.Ciphertext) != "updated" {
		t.Errorf("ciphertext = %q, want %q", got.Ciphertext, "updated")
	}
	_ = pin2
}

func TestSQLiteGetPins(t *testing.T) {
	s := newTestSQLiteStore(t)
	ctx := context.Background()

	created, _ := s.CreateTeam(ctx, CreateTeamInput{Name: "team1", Username: "alice"})

	s.UpsertPin(ctx, UpsertPinInput{TeamID: created.Team.ID, Filename: ".env", Ciphertext: []byte("a"), Nonce: []byte("n"), PinnedBy: "alice"})
	s.UpsertPin(ctx, UpsertPinInput{TeamID: created.Team.ID, Filename: ".env.local", Ciphertext: []byte("b"), Nonce: []byte("n"), PinnedBy: "alice"})

	pins, err := s.GetPins(ctx, created.Team.ID)
	if err != nil {
		t.Fatalf("GetPins() error: %v", err)
	}
	if len(pins) != 2 {
		t.Fatalf("got %d pins, want 2", len(pins))
	}
}

func TestSQLiteGetPinNotFound(t *testing.T) {
	s := newTestSQLiteStore(t)
	ctx := context.Background()

	_, err := s.GetPin(ctx, "nonexistent", ".env")
	if err != ErrNotFound {
		t.Errorf("error = %v, want ErrNotFound", err)
	}
}

func TestSQLiteShareLog(t *testing.T) {
	s := newTestSQLiteStore(t)
	ctx := context.Background()

	created, _ := s.CreateTeam(ctx, CreateTeamInput{Name: "team1", Username: "alice"})

	entry, err := s.CreateShareLog(ctx, CreateShareInput{
		TeamID:   created.Team.ID,
		FromUser: "alice",
		ToUser:   "bob",
		Filename: ".env",
		SecretID: "secret123",
		TTL:      time.Hour,
	})
	if err != nil {
		t.Fatalf("CreateShareLog() error: %v", err)
	}
	if entry.FromUser != "alice" {
		t.Errorf("from_user = %q, want %q", entry.FromUser, "alice")
	}

	log, err := s.GetShareLog(ctx, created.Team.ID)
	if err != nil {
		t.Fatalf("GetShareLog() error: %v", err)
	}
	if len(log) != 1 {
		t.Fatalf("got %d entries, want 1", len(log))
	}
}

func TestSQLiteInbox(t *testing.T) {
	s := newTestSQLiteStore(t)
	ctx := context.Background()

	created, _ := s.CreateTeam(ctx, CreateTeamInput{Name: "team1", Username: "alice"})

	// Share to bob
	s.CreateShareLog(ctx, CreateShareInput{
		TeamID: created.Team.ID, FromUser: "alice", ToUser: "bob",
		Filename: ".env", SecretID: "s1", TTL: time.Hour,
	})

	// Share to all
	s.CreateShareLog(ctx, CreateShareInput{
		TeamID: created.Team.ID, FromUser: "alice",
		Filename: ".env.local", SecretID: "s2", TTL: time.Hour,
	})

	// Share to charlie (bob shouldn't see this)
	s.CreateShareLog(ctx, CreateShareInput{
		TeamID: created.Team.ID, FromUser: "alice", ToUser: "charlie",
		Filename: ".env.prod", SecretID: "s3", TTL: time.Hour,
	})

	inbox, err := s.GetInbox(ctx, created.Team.ID, "bob")
	if err != nil {
		t.Fatalf("GetInbox() error: %v", err)
	}
	// Bob sees: his direct share + the --all share
	if len(inbox) != 2 {
		t.Fatalf("got %d inbox items, want 2", len(inbox))
	}
}

func TestSQLiteNewStoreInvalidPath(t *testing.T) {
	_, err := NewSQLiteStore("/nonexistent/path/test.db")
	if err == nil {
		t.Error("expected error for invalid path")
	}
}

func TestSQLiteNewStoreCreatesFile(t *testing.T) {
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "new.db")

	s, err := NewSQLiteStore(dbPath)
	if err != nil {
		t.Fatalf("NewSQLiteStore() error: %v", err)
	}
	defer s.Close()

	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		t.Error("database file was not created")
	}
}
