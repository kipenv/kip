// Package envfile provides utilities for parsing, generating, and comparing
// .env files.
package envfile

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

// Entry represents a single line in a .env file.
type Entry struct {
	Key   string
	Value string
}

// Parse reads a .env file and returns its key-value pairs.
// Supports: KEY=VALUE, KEY="quoted", KEY='quoted', # comments,
// empty lines, and export KEY=VALUE.
func Parse(path string) ([]Entry, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("parse env file: %w", err)
	}
	defer f.Close()

	return ParseReader(bufio.NewReader(f))
}

// ParseReader parses .env content from a reader.
func ParseReader(r *bufio.Reader) ([]Entry, error) {
	var entries []Entry
	scanner := bufio.NewScanner(r)

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		// Skip empty lines and comments
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// Strip "export " prefix
		line = strings.TrimPrefix(line, "export ")

		// Split on first '='
		idx := strings.Index(line, "=")
		if idx < 0 {
			continue
		}

		key := strings.TrimSpace(line[:idx])
		value := strings.TrimSpace(line[idx+1:])

		// Remove surrounding quotes
		value = unquote(value)

		if key == "" {
			continue
		}

		entries = append(entries, Entry{Key: key, Value: value})
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("parse env file: %w", err)
	}

	return entries, nil
}

// ParseString parses .env content from a string.
func ParseString(content string) ([]Entry, error) {
	return ParseReader(bufio.NewReader(strings.NewReader(content)))
}

// ToMap converts entries to a map for easy lookup.
func ToMap(entries []Entry) map[string]string {
	m := make(map[string]string, len(entries))
	for _, e := range entries {
		m[e.Key] = e.Value
	}
	return m
}

// unquote removes surrounding single or double quotes from a value.
func unquote(s string) string {
	if len(s) < 2 {
		return s
	}
	if (s[0] == '"' && s[len(s)-1] == '"') || (s[0] == '\'' && s[len(s)-1] == '\'') {
		return s[1 : len(s)-1]
	}
	return s
}
