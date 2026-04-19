package store

import (
	"context"
	"crypto/rand"
	"database/sql"
	_ "embed"
	"encoding/hex"
	"fmt"
	"time"

	_ "github.com/mattn/go-sqlite3"
	nanoid "github.com/matoous/go-nanoid/v2"
)

//go:embed migrations/001_init.sql
var initSQL string

// SQLiteStore implements TeamStore using SQLite.
type SQLiteStore struct {
	db *sql.DB
}

// NewSQLiteStore creates a new SQLite store and runs migrations.
func NewSQLiteStore(dbPath string) (*SQLiteStore, error) {
	db, err := sql.Open("sqlite3", dbPath+"?_journal_mode=WAL&_foreign_keys=on")
	if err != nil {
		return nil, fmt.Errorf("open sqlite: %w", err)
	}

	if _, err := db.Exec(initSQL); err != nil {
		db.Close()
		return nil, fmt.Errorf("run migrations: %w", err)
	}

	return &SQLiteStore{db: db}, nil
}

func generateToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return "envs_m_" + hex.EncodeToString(b), nil
}

func (s *SQLiteStore) CreateTeam(ctx context.Context, input CreateTeamInput) (*CreateTeamOutput, error) {
	teamID, err := nanoid.Generate(idAlphabet, idLength)
	if err != nil {
		return nil, err
	}

	inviteCode, err := nanoid.Generate(idAlphabet, 8)
	if err != nil {
		return nil, err
	}
	inviteCode = input.Name + "-" + inviteCode

	memberID, err := nanoid.Generate(idAlphabet, idLength)
	if err != nil {
		return nil, err
	}

	token, err := generateToken()
	if err != nil {
		return nil, err
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	now := time.Now().UTC()

	_, err = tx.ExecContext(ctx,
		"INSERT INTO teams (id, name, invite_code, created_at) VALUES (?, ?, ?, ?)",
		teamID, input.Name, inviteCode, now)
	if err != nil {
		return nil, fmt.Errorf("insert team: %w", err)
	}

	_, err = tx.ExecContext(ctx,
		"INSERT INTO members (id, team_id, username, token, joined_at) VALUES (?, ?, ?, ?, ?)",
		memberID, teamID, input.Username, token, now)
	if err != nil {
		return nil, fmt.Errorf("insert member: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	return &CreateTeamOutput{
		Team: Team{
			ID:         teamID,
			Name:       input.Name,
			InviteCode: inviteCode,
			CreatedAt:  now,
		},
		MemberToken: token,
	}, nil
}

func (s *SQLiteStore) JoinTeam(ctx context.Context, input JoinTeamInput) (*JoinTeamOutput, error) {
	var team Team
	err := s.db.QueryRowContext(ctx,
		"SELECT id, name, invite_code, created_at FROM teams WHERE invite_code = ?",
		input.InviteCode).Scan(&team.ID, &team.Name, &team.InviteCode, &team.CreatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrNotFound
		}
		return nil, err
	}

	memberID, err := nanoid.Generate(idAlphabet, idLength)
	if err != nil {
		return nil, err
	}

	token, err := generateToken()
	if err != nil {
		return nil, err
	}

	_, err = s.db.ExecContext(ctx,
		"INSERT INTO members (id, team_id, username, token, joined_at) VALUES (?, ?, ?, ?, ?)",
		memberID, team.ID, input.Username, token, time.Now().UTC())
	if err != nil {
		return nil, fmt.Errorf("insert member: %w", err)
	}

	return &JoinTeamOutput{
		Team:        team,
		MemberToken: token,
	}, nil
}

func (s *SQLiteStore) LeaveTeam(ctx context.Context, teamID, memberID string) error {
	result, err := s.db.ExecContext(ctx,
		"DELETE FROM members WHERE id = ? AND team_id = ?", memberID, teamID)
	if err != nil {
		return err
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return ErrNotFound
	}

	// Delete team if no members remain
	var count int
	s.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM members WHERE team_id = ?", teamID).Scan(&count)
	if count == 0 {
		s.db.ExecContext(ctx, "DELETE FROM pins WHERE team_id = ?", teamID)
		s.db.ExecContext(ctx, "DELETE FROM share_log WHERE team_id = ?", teamID)
		s.db.ExecContext(ctx, "DELETE FROM teams WHERE id = ?", teamID)
	}

	return nil
}

func (s *SQLiteStore) GetTeamMembers(ctx context.Context, teamID string) ([]Member, error) {
	rows, err := s.db.QueryContext(ctx,
		"SELECT id, team_id, username, joined_at FROM members WHERE team_id = ? ORDER BY joined_at",
		teamID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var members []Member
	for rows.Next() {
		var m Member
		if err := rows.Scan(&m.ID, &m.TeamID, &m.Username, &m.JoinedAt); err != nil {
			return nil, err
		}
		members = append(members, m)
	}
	return members, rows.Err()
}

func (s *SQLiteStore) GetMemberByToken(ctx context.Context, token string) (*Member, error) {
	var m Member
	err := s.db.QueryRowContext(ctx,
		"SELECT id, team_id, username, token, joined_at FROM members WHERE token = ?",
		token).Scan(&m.ID, &m.TeamID, &m.Username, &m.Token, &m.JoinedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &m, nil
}

func (s *SQLiteStore) GetTeam(ctx context.Context, id string) (*Team, error) {
	var t Team
	err := s.db.QueryRowContext(ctx,
		"SELECT id, name, invite_code, created_at FROM teams WHERE id = ?",
		id).Scan(&t.ID, &t.Name, &t.InviteCode, &t.CreatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &t, nil
}

func (s *SQLiteStore) GetTeamsByMember(ctx context.Context, memberToken string) ([]Team, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT t.id, t.name, t.invite_code, t.created_at
		 FROM teams t
		 JOIN members m ON m.team_id = t.id
		 WHERE m.token = ?
		 ORDER BY t.name`, memberToken)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var teams []Team
	for rows.Next() {
		var t Team
		if err := rows.Scan(&t.ID, &t.Name, &t.InviteCode, &t.CreatedAt); err != nil {
			return nil, err
		}
		teams = append(teams, t)
	}
	return teams, rows.Err()
}

func (s *SQLiteStore) CreateShareLog(ctx context.Context, input CreateShareInput) (*ShareLogEntry, error) {
	id, err := nanoid.Generate(idAlphabet, idLength)
	if err != nil {
		return nil, err
	}

	now := time.Now().UTC()
	expiresAt := now.Add(input.TTL)

	_, err = s.db.ExecContext(ctx,
		`INSERT INTO share_log (id, team_id, from_user, to_user, filename, secret_id, expires_at, created_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		id, input.TeamID, input.FromUser, sql.NullString{String: input.ToUser, Valid: input.ToUser != ""},
		input.Filename, input.SecretID, expiresAt, now)
	if err != nil {
		return nil, err
	}

	return &ShareLogEntry{
		ID:        id,
		TeamID:    input.TeamID,
		FromUser:  input.FromUser,
		ToUser:    input.ToUser,
		Filename:  input.Filename,
		SecretID:  input.SecretID,
		ExpiresAt: expiresAt,
		CreatedAt: now,
	}, nil
}

func (s *SQLiteStore) GetInbox(ctx context.Context, teamID, username string) ([]ShareLogEntry, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT id, team_id, from_user, to_user, filename, secret_id, expires_at, created_at
		 FROM share_log
		 WHERE team_id = ? AND (to_user = ? OR to_user IS NULL) AND expires_at > ?
		 ORDER BY created_at DESC`,
		teamID, username, time.Now().UTC())
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return scanShareLogs(rows)
}

func (s *SQLiteStore) GetShareLog(ctx context.Context, teamID string) ([]ShareLogEntry, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT id, team_id, from_user, to_user, filename, secret_id, expires_at, created_at
		 FROM share_log
		 WHERE team_id = ?
		 ORDER BY created_at DESC
		 LIMIT 50`,
		teamID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return scanShareLogs(rows)
}

func scanShareLogs(rows *sql.Rows) ([]ShareLogEntry, error) {
	var entries []ShareLogEntry
	for rows.Next() {
		var e ShareLogEntry
		var toUser sql.NullString
		if err := rows.Scan(&e.ID, &e.TeamID, &e.FromUser, &toUser, &e.Filename, &e.SecretID, &e.ExpiresAt, &e.CreatedAt); err != nil {
			return nil, err
		}
		e.ToUser = toUser.String
		entries = append(entries, e)
	}
	return entries, rows.Err()
}

func (s *SQLiteStore) UpsertPin(ctx context.Context, input UpsertPinInput) (*Pin, error) {
	id, err := nanoid.Generate(idAlphabet, idLength)
	if err != nil {
		return nil, err
	}

	now := time.Now().UTC()

	_, err = s.db.ExecContext(ctx,
		`INSERT INTO pins (id, team_id, filename, ciphertext, nonce, pinned_by, updated_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?)
		 ON CONFLICT(team_id, filename) DO UPDATE SET
		   ciphertext = excluded.ciphertext,
		   nonce = excluded.nonce,
		   pinned_by = excluded.pinned_by,
		   updated_at = excluded.updated_at`,
		id, input.TeamID, input.Filename, input.Ciphertext, input.Nonce, input.PinnedBy, now)
	if err != nil {
		return nil, err
	}

	return &Pin{
		ID:         id,
		TeamID:     input.TeamID,
		Filename:   input.Filename,
		Ciphertext: input.Ciphertext,
		Nonce:      input.Nonce,
		PinnedBy:   input.PinnedBy,
		UpdatedAt:  now,
	}, nil
}

func (s *SQLiteStore) GetPins(ctx context.Context, teamID string) ([]Pin, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT id, team_id, filename, ciphertext, nonce, pinned_by, updated_at
		 FROM pins WHERE team_id = ? ORDER BY filename`,
		teamID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var pins []Pin
	for rows.Next() {
		var p Pin
		if err := rows.Scan(&p.ID, &p.TeamID, &p.Filename, &p.Ciphertext, &p.Nonce, &p.PinnedBy, &p.UpdatedAt); err != nil {
			return nil, err
		}
		pins = append(pins, p)
	}
	return pins, rows.Err()
}

func (s *SQLiteStore) GetPin(ctx context.Context, teamID, filename string) (*Pin, error) {
	var p Pin
	err := s.db.QueryRowContext(ctx,
		`SELECT id, team_id, filename, ciphertext, nonce, pinned_by, updated_at
		 FROM pins WHERE team_id = ? AND filename = ?`,
		teamID, filename).Scan(&p.ID, &p.TeamID, &p.Filename, &p.Ciphertext, &p.Nonce, &p.PinnedBy, &p.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &p, nil
}

// Close closes the SQLite database connection.
func (s *SQLiteStore) Close() error {
	return s.db.Close()
}
