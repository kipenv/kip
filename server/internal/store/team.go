package store

import (
	"context"
	"time"
)

// Team represents a team.
type Team struct {
	ID         string    `json:"id"`
	Name       string    `json:"name"`
	InviteCode string    `json:"invite_code"`
	CreatedAt  time.Time `json:"created_at"`
}

// Member represents a team member.
type Member struct {
	ID       string    `json:"id"`
	TeamID   string    `json:"team_id"`
	Username string    `json:"username"`
	Token    string    `json:"token,omitempty"`
	JoinedAt time.Time `json:"joined_at"`
}

// Pin represents a pinned env file for a team.
type Pin struct {
	ID         string    `json:"id"`
	TeamID     string    `json:"team_id"`
	Filename   string    `json:"filename"`
	Ciphertext []byte    `json:"ciphertext"`
	Nonce      []byte    `json:"nonce"`
	PinnedBy   string    `json:"pinned_by"`
	UpdatedAt  time.Time `json:"updated_at"`
}

// ShareLogEntry represents a share history entry.
type ShareLogEntry struct {
	ID        string    `json:"id"`
	TeamID    string    `json:"team_id"`
	FromUser  string    `json:"from_user"`
	ToUser    string    `json:"to_user,omitempty"`
	Filename  string    `json:"filename"`
	SecretID  string    `json:"secret_id"`
	ExpiresAt time.Time `json:"expires_at"`
	CreatedAt time.Time `json:"created_at"`
}

// CreateTeamInput contains the data needed to create a team.
type CreateTeamInput struct {
	Name     string
	Username string
}

// CreateTeamOutput contains the result of creating a team.
type CreateTeamOutput struct {
	Team        Team
	MemberToken string
}

// JoinTeamInput contains the data needed to join a team.
type JoinTeamInput struct {
	InviteCode string
	Username   string
}

// JoinTeamOutput contains the result of joining a team.
type JoinTeamOutput struct {
	Team        Team
	MemberToken string
}

// CreateShareInput contains the data needed to log a share.
type CreateShareInput struct {
	TeamID   string
	FromUser string
	ToUser   string
	Filename string
	SecretID string
	TTL      time.Duration
}

// UpsertPinInput contains the data needed to pin an env file.
type UpsertPinInput struct {
	TeamID     string
	Filename   string
	Ciphertext []byte
	Nonce      []byte
	PinnedBy   string
}

// TeamStore defines the interface for persistent team storage.
type TeamStore interface {
	// CreateTeam creates a new team and adds the creator as the first member.
	CreateTeam(ctx context.Context, input CreateTeamInput) (*CreateTeamOutput, error)

	// JoinTeam adds a member to a team by invite code.
	JoinTeam(ctx context.Context, input JoinTeamInput) (*JoinTeamOutput, error)

	// LeaveTeam removes a member from a team. Deletes the team if no members remain.
	LeaveTeam(ctx context.Context, teamID, memberID string) error

	// GetTeamMembers returns all members of a team.
	GetTeamMembers(ctx context.Context, teamID string) ([]Member, error)

	// GetMemberByToken returns a member by their auth token.
	GetMemberByToken(ctx context.Context, token string) (*Member, error)

	// GetTeam returns a team by ID.
	GetTeam(ctx context.Context, id string) (*Team, error)

	// GetTeamsByMember returns all teams a member belongs to.
	GetTeamsByMember(ctx context.Context, memberToken string) ([]Team, error)

	// CreateShareLog logs a share event.
	CreateShareLog(ctx context.Context, input CreateShareInput) (*ShareLogEntry, error)

	// GetInbox returns pending shares for a user in a team.
	GetInbox(ctx context.Context, teamID, username string) ([]ShareLogEntry, error)

	// GetShareLog returns share history for a team.
	GetShareLog(ctx context.Context, teamID string) ([]ShareLogEntry, error)

	// UpsertPin creates or updates a pinned env file.
	UpsertPin(ctx context.Context, input UpsertPinInput) (*Pin, error)

	// GetPins returns all pinned envs for a team.
	GetPins(ctx context.Context, teamID string) ([]Pin, error)

	// GetPin returns a specific pinned env by team and filename.
	GetPin(ctx context.Context, teamID, filename string) (*Pin, error)

	// Close closes the store connection.
	Close() error
}
