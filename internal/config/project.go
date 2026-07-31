package config

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
)

const projectFile = ".kip"

// ProjectConfig represents the .kip file in a project directory.
type ProjectConfig struct {
	Team   string `json:"team"`
	Server string `json:"server,omitempty"`
}

// LoadProject reads the .kip file from the current or parent directories.
func LoadProject() (*ProjectConfig, string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return nil, "", err
	}

	for {
		path := filepath.Join(dir, projectFile)
		data, err := os.ReadFile(path)
		if err == nil {
			var cfg ProjectConfig
			if err = json.Unmarshal(data, &cfg); err != nil {
				return nil, "", err
			}
			return &cfg, path, nil
		}
		if !errors.Is(err, os.ErrNotExist) {
			return nil, "", err
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}

	return nil, "", nil
}

// SaveProject writes the .kip file to the given directory.
func SaveProject(dir string, cfg ProjectConfig) (string, error) {
	path := filepath.Join(dir, projectFile)
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return "", err
	}
	// 0600: .kip identifies the team this directory belongs to. Same treatment
	// as the other kip-owned files (config.json, teams.json).
	return path, os.WriteFile(path, data, 0o600)
}

// AddToGitignore adds .kip to the .gitignore file in the given directory.
func AddToGitignore(dir string) error {
	gitignorePath := filepath.Join(dir, ".gitignore")
	data, err := os.ReadFile(gitignorePath)
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return err
	}

	content := string(data)
	// Check if already present
	for _, line := range splitLines(content) {
		if line == projectFile {
			return nil
		}
	}

	if len(content) > 0 && content[len(content)-1] != '\n' {
		content += "\n"
	}
	content += projectFile + "\n"

	// 0644 by design: .gitignore is a shared, committed file that git and every
	// other tool on the machine must be able to read.
	return os.WriteFile(gitignorePath, []byte(content), 0o644) //nolint:gosec // see above
}

// AddToGitExclude adds .kip to .git/info/exclude.
func AddToGitExclude(dir string) error {
	excludePath := filepath.Join(dir, ".git", "info", "exclude")

	// Ensure directory exists
	if err := os.MkdirAll(filepath.Dir(excludePath), 0o755); err != nil {
		return err
	}

	data, err := os.ReadFile(excludePath)
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return err
	}

	content := string(data)
	for _, line := range splitLines(content) {
		if line == projectFile {
			return nil
		}
	}

	if len(content) > 0 && content[len(content)-1] != '\n' {
		content += "\n"
	}
	content += projectFile + "\n"

	// 0644 by design: git reads .git/info/exclude directly.
	return os.WriteFile(excludePath, []byte(content), 0o644) //nolint:gosec // see above
}

func splitLines(s string) []string {
	var lines []string
	start := 0
	for i := 0; i < len(s); i++ {
		if s[i] == '\n' {
			lines = append(lines, s[start:i])
			start = i + 1
		}
	}
	if start < len(s) {
		lines = append(lines, s[start:])
	}
	return lines
}
