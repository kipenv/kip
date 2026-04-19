package config

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
)

const teamsFile = "teams.json"

// TeamEntry stores the token and metadata for a team membership.
type TeamEntry struct {
	ID         string `json:"id"`
	Name       string `json:"name"`
	Token      string `json:"token"`
	InviteCode string `json:"invite_code,omitempty"`
	Username   string `json:"username"`
}

// TeamsConfig holds all team memberships.
type TeamsConfig struct {
	Teams []TeamEntry `json:"teams"`
}

// teamsPath returns the path to the teams config file.
func teamsPath() (string, error) {
	cfgDir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(cfgDir, configDir, teamsFile), nil
}

// LoadTeams reads team memberships from disk.
func LoadTeams() (*TeamsConfig, error) {
	path, err := teamsPath()
	if err != nil {
		return &TeamsConfig{}, nil
	}

	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return &TeamsConfig{}, nil
		}
		return nil, err
	}

	var cfg TeamsConfig
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}

// SaveTeams writes team memberships to disk.
func SaveTeams(cfg *TeamsConfig) error {
	path, err := teamsPath()
	if err != nil {
		return err
	}

	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return err
	}

	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0o600)
}

// AddTeam adds or updates a team entry.
func (c *TeamsConfig) AddTeam(entry TeamEntry) {
	for i, t := range c.Teams {
		if t.ID == entry.ID {
			c.Teams[i] = entry
			return
		}
	}
	c.Teams = append(c.Teams, entry)
}

// RemoveTeam removes a team entry by ID.
func (c *TeamsConfig) RemoveTeam(id string) bool {
	for i, t := range c.Teams {
		if t.ID == id {
			c.Teams = append(c.Teams[:i], c.Teams[i+1:]...)
			return true
		}
	}
	return false
}

// FindByName returns a team entry by name.
func (c *TeamsConfig) FindByName(name string) *TeamEntry {
	for i := range c.Teams {
		if c.Teams[i].Name == name {
			return &c.Teams[i]
		}
	}
	return nil
}

// FindByID returns a team entry by ID.
func (c *TeamsConfig) FindByID(id string) *TeamEntry {
	for i := range c.Teams {
		if c.Teams[i].ID == id {
			return &c.Teams[i]
		}
	}
	return nil
}
