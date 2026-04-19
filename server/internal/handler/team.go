package handler

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"strings"

	"github.com/kipenv/kip/server/internal/store"
)

// TeamHandler handles team endpoints.
type TeamHandler struct {
	teams  store.TeamStore
	logger *slog.Logger
}

// NewTeamHandler creates a new team handler.
func NewTeamHandler(teams store.TeamStore, logger *slog.Logger) *TeamHandler {
	return &TeamHandler{teams: teams, logger: logger}
}

// CreateTeamRequest is the JSON body for creating a team.
type CreateTeamRequest struct {
	Name     string `json:"name"`
	Username string `json:"username"`
}

// CreateTeamResponse is the JSON response after creating a team.
type CreateTeamResponse struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	InviteCode  string `json:"invite_code"`
	MemberToken string `json:"member_token"`
}

// JoinTeamRequest is the JSON body for joining a team.
type JoinTeamRequest struct {
	InviteCode string `json:"invite_code"`
	Username   string `json:"username"`
}

// JoinTeamResponse is the JSON response after joining a team.
type JoinTeamResponse struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	MemberToken string `json:"member_token"`
}

// MemberResponse is a member in the members list.
type MemberResponse struct {
	Username string `json:"username"`
	JoinedAt string `json:"joined_at"`
}

// TeamListResponse is a team in the list.
type TeamListResponse struct {
	ID         string `json:"id"`
	Name       string `json:"name"`
	InviteCode string `json:"invite_code"`
}

// Create handles POST /api/v1/teams
func (h *TeamHandler) Create(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	var req CreateTeamRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.Name == "" || req.Username == "" {
		writeError(w, http.StatusBadRequest, "name and username are required")
		return
	}

	out, err := h.teams.CreateTeam(r.Context(), store.CreateTeamInput{
		Name:     req.Name,
		Username: req.Username,
	})
	if err != nil {
		if strings.Contains(err.Error(), "UNIQUE constraint") {
			writeError(w, http.StatusConflict, "team name already taken")
			return
		}
		h.logger.Error("failed to create team", "error", err)
		writeError(w, http.StatusInternalServerError, "failed to create team")
		return
	}

	writeJSON(w, http.StatusCreated, CreateTeamResponse{
		ID:          out.Team.ID,
		Name:        out.Team.Name,
		InviteCode:  out.Team.InviteCode,
		MemberToken: out.MemberToken,
	})
}

// Join handles POST /api/v1/teams/join
func (h *TeamHandler) Join(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	var req JoinTeamRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.InviteCode == "" || req.Username == "" {
		writeError(w, http.StatusBadRequest, "invite_code and username are required")
		return
	}

	out, err := h.teams.JoinTeam(r.Context(), store.JoinTeamInput{
		InviteCode: req.InviteCode,
		Username:   req.Username,
	})
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			writeError(w, http.StatusNotFound, "invalid invite code")
			return
		}
		if strings.Contains(err.Error(), "UNIQUE constraint") {
			writeError(w, http.StatusConflict, "username already taken in this team")
			return
		}
		h.logger.Error("failed to join team", "error", err)
		writeError(w, http.StatusInternalServerError, "failed to join team")
		return
	}

	writeJSON(w, http.StatusOK, JoinTeamResponse{
		ID:          out.Team.ID,
		Name:        out.Team.Name,
		MemberToken: out.MemberToken,
	})
}

// Leave handles DELETE /api/v1/teams/{id}/leave
func (h *TeamHandler) Leave(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	member, ok := h.authenticate(w, r)
	if !ok {
		return
	}

	teamID := r.PathValue("id")
	if teamID == "" {
		writeError(w, http.StatusBadRequest, "missing team id")
		return
	}

	if member.TeamID != teamID {
		writeError(w, http.StatusForbidden, "not a member of this team")
		return
	}

	if err := h.teams.LeaveTeam(r.Context(), teamID, member.ID); err != nil {
		h.logger.Error("failed to leave team", "error", err)
		writeError(w, http.StatusInternalServerError, "failed to leave team")
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"status": "left"})
}

// Members handles GET /api/v1/teams/{id}/members
func (h *TeamHandler) Members(w http.ResponseWriter, r *http.Request) {
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

	members, err := h.teams.GetTeamMembers(r.Context(), teamID)
	if err != nil {
		h.logger.Error("failed to get members", "error", err)
		writeError(w, http.StatusInternalServerError, "failed to get members")
		return
	}

	resp := make([]MemberResponse, len(members))
	for i, m := range members {
		resp[i] = MemberResponse{
			Username: m.Username,
			JoinedAt: m.JoinedAt.Format("2006-01-02"),
		}
	}

	writeJSON(w, http.StatusOK, resp)
}

// ListTeams handles GET /api/v1/teams
func (h *TeamHandler) ListTeams(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	token := extractToken(r)
	if token == "" {
		writeError(w, http.StatusUnauthorized, "missing authorization token")
		return
	}

	member, err := h.teams.GetMemberByToken(r.Context(), token)
	if err != nil {
		writeError(w, http.StatusUnauthorized, "invalid token")
		return
	}

	// Get all teams for this member — need to get teams where this user is a member
	// The token only maps to one team, so we return that team
	team, err := h.teams.GetTeam(r.Context(), member.TeamID)
	if err != nil {
		h.logger.Error("failed to get team", "error", err)
		writeError(w, http.StatusInternalServerError, "failed to get teams")
		return
	}

	writeJSON(w, http.StatusOK, []TeamListResponse{{
		ID:         team.ID,
		Name:       team.Name,
		InviteCode: team.InviteCode,
	}})
}

// authenticate extracts and validates the bearer token, returning the member.
func (h *TeamHandler) authenticate(w http.ResponseWriter, r *http.Request) (*store.Member, bool) {
	token := extractToken(r)
	if token == "" {
		writeError(w, http.StatusUnauthorized, "missing authorization token")
		return nil, false
	}

	member, err := h.teams.GetMemberByToken(r.Context(), token)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			writeError(w, http.StatusUnauthorized, "invalid token")
			return nil, false
		}
		h.logger.Error("auth error", "error", err)
		writeError(w, http.StatusInternalServerError, "auth error")
		return nil, false
	}

	return member, true
}

// extractToken gets the bearer token from the Authorization header.
func extractToken(r *http.Request) string {
	auth := r.Header.Get("Authorization")
	if !strings.HasPrefix(auth, "Bearer ") {
		return ""
	}
	return strings.TrimPrefix(auth, "Bearer ")
}
