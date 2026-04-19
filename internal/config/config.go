// Package config manages CLI configuration for kip.
package config

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
)

const (
	defaultServerURL = "http://localhost:8080"
	configDir        = "kip"
	configFile       = "config.json"
)

// Config holds the CLI configuration.
type Config struct {
	ServerURL string `json:"server_url"`
}

// DefaultConfig returns the default configuration.
func DefaultConfig() Config {
	return Config{
		ServerURL: defaultServerURL,
	}
}

// configPath returns the path to the config file.
func configPath() (string, error) {
	cfgDir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(cfgDir, configDir, configFile), nil
}

// Load reads the config from disk, or returns defaults if not found.
func Load() (Config, error) {
	path, err := configPath()
	if err != nil {
		return DefaultConfig(), nil
	}

	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return DefaultConfig(), nil
		}
		return Config{}, err
	}

	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return Config{}, err
	}

	if cfg.ServerURL == "" {
		cfg.ServerURL = defaultServerURL
	}

	return cfg, nil
}

// Save writes the config to disk.
func Save(cfg Config) error {
	path, err := configPath()
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
