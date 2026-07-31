// Package client provides an HTTP client for the kip server API.
package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// Client is an HTTP client for the kip server.
type Client struct {
	baseURL    string
	httpClient *http.Client
}

// New creates a new kip API client.
func New(baseURL string) *Client {
	return &Client{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// do issues a request against the server. A nil body sends no payload; an
// empty token sends no Authorization header. The context lets callers cancel
// in-flight requests (the CLI wires it to SIGINT).
func (c *Client) do(ctx context.Context, method, path string, body []byte, token string) (*http.Response, error) {
	var payload io.Reader
	if body != nil {
		payload = bytes.NewReader(body)
	}

	req, err := http.NewRequestWithContext(ctx, method, c.baseURL+path, payload)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}

	return resp, nil
}

// CreateSecretRequest is the request body for creating a secret.
type CreateSecretRequest struct {
	Ciphertext        string `json:"ciphertext"`
	Nonce             string `json:"nonce"`
	Salt              string `json:"salt,omitempty"`
	Filename          string `json:"filename"`
	MaxReads          int    `json:"max_reads"`
	TTLSeconds        int    `json:"ttl_seconds"`
	PasswordProtected bool   `json:"password_protected,omitempty"`
}

// CreateSecretResponse is the response from creating a secret.
type CreateSecretResponse struct {
	ID        string    `json:"id"`
	ExpiresAt time.Time `json:"expires_at"`
}

// GetSecretResponse is the response from retrieving a secret.
type GetSecretResponse struct {
	Ciphertext        string `json:"ciphertext"`
	Nonce             string `json:"nonce"`
	Salt              string `json:"salt,omitempty"`
	Filename          string `json:"filename"`
	ReadsLeft         int    `json:"reads_left"`
	PasswordProtected bool   `json:"password_protected,omitempty"`
}

// ErrorResponse is returned when the server responds with an error.
type ErrorResponse struct {
	Error string `json:"error"`
}

// CreateSecret uploads an encrypted secret to the server.
func (c *Client) CreateSecret(ctx context.Context, req CreateSecretRequest) (*CreateSecretResponse, error) {
	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}

	resp, err := c.do(ctx, http.MethodPost, "/api/v1/secret", body, "")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		return nil, parseError(resp)
	}

	var result CreateSecretResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	return &result, nil
}

// GetSecret retrieves an encrypted secret from the server.
func (c *Client) GetSecret(ctx context.Context, id string) (*GetSecretResponse, error) {
	resp, err := c.do(ctx, http.MethodGet, "/api/v1/secret/"+id, nil, "")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, parseError(resp)
	}

	var result GetSecretResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	return &result, nil
}

// DeleteSecret revokes a secret by ID.
func (c *Client) DeleteSecret(ctx context.Context, id string) error {
	resp, err := c.do(ctx, http.MethodDelete, "/api/v1/secret/"+id, nil, "")
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent {
		return parseError(resp)
	}

	return nil
}

// Team types

// CreateTeamRequest is the request body for creating a team.
type CreateTeamRequest struct {
	Name     string `json:"name"`
	Username string `json:"username"`
}

// CreateTeamResponse is the response from creating a team.
type CreateTeamResponse struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	InviteCode  string `json:"invite_code"`
	MemberToken string `json:"member_token"`
}

// JoinTeamRequest is the request body for joining a team.
type JoinTeamRequest struct {
	InviteCode string `json:"invite_code"`
	Username   string `json:"username"`
}

// JoinTeamResponse is the response from joining a team.
type JoinTeamResponse struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	MemberToken string `json:"member_token"`
}

// MemberInfo represents a team member.
type MemberInfo struct {
	Username string `json:"username"`
	JoinedAt string `json:"joined_at"`
}

// CreateTeam creates a new team.
func (c *Client) CreateTeam(ctx context.Context, req CreateTeamRequest) (*CreateTeamResponse, error) {
	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}

	resp, err := c.do(ctx, http.MethodPost, "/api/v1/teams", body, "")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		return nil, parseError(resp)
	}

	var result CreateTeamResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	return &result, nil
}

// JoinTeam joins a team by invite code.
func (c *Client) JoinTeam(ctx context.Context, req JoinTeamRequest) (*JoinTeamResponse, error) {
	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}

	resp, err := c.do(ctx, http.MethodPost, "/api/v1/teams/join", body, "")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, parseError(resp)
	}

	var result JoinTeamResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	return &result, nil
}

// LeaveTeam leaves a team.
func (c *Client) LeaveTeam(ctx context.Context, teamID, token string) error {
	resp, err := c.do(ctx, http.MethodDelete, "/api/v1/teams/"+teamID+"/leave", nil, token)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return parseError(resp)
	}

	return nil
}

// GetTeamMembers returns all members of a team.
func (c *Client) GetTeamMembers(ctx context.Context, teamID, token string) ([]MemberInfo, error) {
	resp, err := c.do(ctx, http.MethodGet, "/api/v1/teams/"+teamID+"/members", nil, token)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, parseError(resp)
	}

	var result []MemberInfo
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	return result, nil
}

func parseError(resp *http.Response) error {
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("server error (status %d)", resp.StatusCode)
	}

	var errResp ErrorResponse
	if json.Unmarshal(body, &errResp) == nil && errResp.Error != "" {
		return fmt.Errorf("server error: %s", errResp.Error)
	}

	return fmt.Errorf("server error (status %d): %s", resp.StatusCode, string(body))
}
