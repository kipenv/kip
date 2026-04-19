package envfile

import (
	"os"
	"path/filepath"
	"testing"
)

func TestParseString(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []Entry
	}{
		{
			name:  "simple key=value",
			input: "DB_HOST=localhost",
			expected: []Entry{
				{Key: "DB_HOST", Value: "localhost"},
			},
		},
		{
			name:  "multiple entries",
			input: "DB_HOST=localhost\nDB_PORT=5432\nDB_NAME=mydb",
			expected: []Entry{
				{Key: "DB_HOST", Value: "localhost"},
				{Key: "DB_PORT", Value: "5432"},
				{Key: "DB_NAME", Value: "mydb"},
			},
		},
		{
			name:  "double quoted value",
			input: `DB_URL="postgres://user:pass@host/db"`,
			expected: []Entry{
				{Key: "DB_URL", Value: "postgres://user:pass@host/db"},
			},
		},
		{
			name:  "single quoted value",
			input: `SECRET='my secret value'`,
			expected: []Entry{
				{Key: "SECRET", Value: "my secret value"},
			},
		},
		{
			name:  "export prefix",
			input: "export API_KEY=abc123",
			expected: []Entry{
				{Key: "API_KEY", Value: "abc123"},
			},
		},
		{
			name:     "comment lines",
			input:    "# This is a comment\n# Another comment\nKEY=value",
			expected: []Entry{{Key: "KEY", Value: "value"}},
		},
		{
			name:     "empty lines",
			input:    "\n\nKEY=value\n\n",
			expected: []Entry{{Key: "KEY", Value: "value"}},
		},
		{
			name:  "empty value",
			input: "EMPTY=",
			expected: []Entry{
				{Key: "EMPTY", Value: ""},
			},
		},
		{
			name:  "value with equals sign",
			input: "URL=postgres://host?sslmode=require",
			expected: []Entry{
				{Key: "URL", Value: "postgres://host?sslmode=require"},
			},
		},
		{
			name:  "whitespace around key and value",
			input: "  KEY  =  value  ",
			expected: []Entry{
				{Key: "KEY", Value: "value"},
			},
		},
		{
			name:     "line without equals (skipped)",
			input:    "invalid-line\nKEY=value",
			expected: []Entry{{Key: "KEY", Value: "value"}},
		},
		{
			name:     "empty key (skipped)",
			input:    "=value\nKEY=good",
			expected: []Entry{{Key: "KEY", Value: "good"}},
		},
		{
			name:     "empty input",
			input:    "",
			expected: nil,
		},
		{
			name:     "only comments",
			input:    "# comment 1\n# comment 2",
			expected: nil,
		},
		{
			name:  "mixed content",
			input: "# Database\nDB_HOST=localhost\n\n# Redis\nexport REDIS_URL=\"redis://localhost:6379\"\n\nDEBUG=true",
			expected: []Entry{
				{Key: "DB_HOST", Value: "localhost"},
				{Key: "REDIS_URL", Value: "redis://localhost:6379"},
				{Key: "DEBUG", Value: "true"},
			},
		},
		{
			name:  "value with special characters",
			input: `PASS=p@$$w0rd!#%^&*()`,
			expected: []Entry{
				{Key: "PASS", Value: "p@$$w0rd!#%^&*()"},
			},
		},
		{
			name:  "quoted value with spaces",
			input: `MSG="hello world"`,
			expected: []Entry{
				{Key: "MSG", Value: "hello world"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			entries, err := ParseString(tt.input)
			if err != nil {
				t.Fatalf("ParseString() error: %v", err)
			}

			if len(entries) != len(tt.expected) {
				t.Fatalf("ParseString() got %d entries, want %d", len(entries), len(tt.expected))
			}

			for i, e := range entries {
				if e.Key != tt.expected[i].Key || e.Value != tt.expected[i].Value {
					t.Errorf("entry[%d] = {%q, %q}, want {%q, %q}",
						i, e.Key, e.Value, tt.expected[i].Key, tt.expected[i].Value)
				}
			}
		})
	}
}

func TestParseFile(t *testing.T) {
	content := "DB_HOST=localhost\nDB_PORT=5432\n"
	dir := t.TempDir()
	path := filepath.Join(dir, ".env")

	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatalf("WriteFile() error: %v", err)
	}

	entries, err := Parse(path)
	if err != nil {
		t.Fatalf("Parse() error: %v", err)
	}

	if len(entries) != 2 {
		t.Fatalf("Parse() got %d entries, want 2", len(entries))
	}
	if entries[0].Key != "DB_HOST" || entries[0].Value != "localhost" {
		t.Errorf("entry[0] = {%q, %q}, want {DB_HOST, localhost}", entries[0].Key, entries[0].Value)
	}
}

func TestParseFileNotFound(t *testing.T) {
	_, err := Parse("/nonexistent/.env")
	if err == nil {
		t.Error("Parse() should fail for nonexistent file")
	}
}

func TestToMap(t *testing.T) {
	entries := []Entry{
		{Key: "A", Value: "1"},
		{Key: "B", Value: "2"},
		{Key: "C", Value: "3"},
	}

	m := ToMap(entries)

	if len(m) != 3 {
		t.Fatalf("ToMap() got %d entries, want 3", len(m))
	}
	if m["A"] != "1" || m["B"] != "2" || m["C"] != "3" {
		t.Errorf("ToMap() = %v, unexpected values", m)
	}
}

func TestToMapDuplicateKeys(t *testing.T) {
	entries := []Entry{
		{Key: "KEY", Value: "first"},
		{Key: "KEY", Value: "last"},
	}

	m := ToMap(entries)
	if m["KEY"] != "last" {
		t.Errorf("ToMap() with duplicates: got %q, want %q", m["KEY"], "last")
	}
}

func TestGenerateExample(t *testing.T) {
	entries := []Entry{
		{Key: "DB_HOST", Value: "localhost"},
		{Key: "DB_PORT", Value: "5432"},
		{Key: "DB_PASSWORD", Value: "secret123"},
		{Key: "API_KEY", Value: "sk_live_abc"},
		{Key: "DEBUG", Value: "true"},
		{Key: "APP_URL", Value: "https://example.com"},
		{Key: "ADMIN_EMAIL", Value: "admin@test.com"},
	}

	result := GenerateExample(entries)

	// Should contain keys but not values
	if !contains(result, "DB_HOST=<hostname>") {
		t.Error("should contain DB_HOST with hostname placeholder")
	}
	if !contains(result, "DB_PORT=<port>") {
		t.Error("should contain DB_PORT with port placeholder")
	}
	if !contains(result, "DB_PASSWORD=<password>") {
		t.Error("should contain DB_PASSWORD with password placeholder")
	}
	if !contains(result, "API_KEY=<key>") {
		t.Error("should contain API_KEY with key placeholder")
	}
	if !contains(result, "DEBUG=<true|false>") {
		t.Error("should contain DEBUG with boolean placeholder")
	}
	if !contains(result, "APP_URL=<url>") {
		t.Error("should contain APP_URL with url placeholder")
	}
	if !contains(result, "ADMIN_EMAIL=<email>") {
		t.Error("should contain ADMIN_EMAIL with email placeholder")
	}

	// Should NOT contain actual values
	if contains(result, "localhost") {
		t.Error("should not contain actual value 'localhost'")
	}
	if contains(result, "secret123") {
		t.Error("should not contain actual value 'secret123'")
	}
}

func TestDiff(t *testing.T) {
	local := map[string]string{
		"DB_HOST":   "localhost",
		"DB_PORT":   "5432",
		"LOCAL_VAR": "only-local",
	}
	remote := map[string]string{
		"DB_HOST":    "localhost",
		"DB_PORT":    "3306",
		"REMOTE_VAR": "only-remote",
	}

	result := Diff(local, remote)

	if len(result.Missing) != 1 || result.Missing[0] != "REMOTE_VAR" {
		t.Errorf("Missing = %v, want [REMOTE_VAR]", result.Missing)
	}
	if len(result.Extra) != 1 || result.Extra[0] != "LOCAL_VAR" {
		t.Errorf("Extra = %v, want [LOCAL_VAR]", result.Extra)
	}
	if len(result.Changed) != 1 || result.Changed[0].Key != "DB_PORT" {
		t.Errorf("Changed = %v, want [{DB_PORT 5432 3306}]", result.Changed)
	}
	if len(result.Same) != 1 || result.Same[0] != "DB_HOST" {
		t.Errorf("Same = %v, want [DB_HOST]", result.Same)
	}
	if !result.HasDifferences() {
		t.Error("HasDifferences() should be true")
	}
}

func TestDiffNoDifferences(t *testing.T) {
	env := map[string]string{"A": "1", "B": "2"}
	result := Diff(env, env)

	if result.HasDifferences() {
		t.Error("HasDifferences() should be false for identical maps")
	}
}

func TestDiffEmptyMaps(t *testing.T) {
	result := Diff(map[string]string{}, map[string]string{})
	if result.HasDifferences() {
		t.Error("HasDifferences() should be false for empty maps")
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && searchString(s, substr)
}

func searchString(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
