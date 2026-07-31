package scanner

import (
	"testing"

	"github.com/kipenv/kip/internal/envfile"
)

func TestPatterns(t *testing.T) {
	tests := []struct {
		name    string
		value   string
		pattern string
		match   bool
	}{
		// AWS keys
		{"aws key positive", "AKIAIOSFODNN7EXAMPLE", "AWS Access Key", true},
		{"aws key negative", "not-an-aws-key", "AWS Access Key", false},

		// Stripe live keys
		{"stripe live positive", "sk_live_4eC39HqLyjWDarjtT1zdp7dc", "Stripe Live Key", true},
		{"stripe test negative", "sk_test_4eC39HqLyjWDarjtT1zdp7dc", "Stripe Live Key", false},

		// GitHub tokens
		{"github token positive", "ghp_ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghij", "GitHub Token", true},
		{"github token short", "ghp_short", "GitHub Token", false},

		// Private keys
		{"rsa private key", "-----BEGIN RSA PRIVATE KEY-----", "Private Key", true},
		{"ec private key", "-----BEGIN EC PRIVATE KEY-----", "Private Key", true},
		{"generic private key", "-----BEGIN PRIVATE KEY-----", "Private Key", true},
		{"public key negative", "-----BEGIN PUBLIC KEY-----", "Private Key", false},

		// Production URLs
		{"prod url", "https://api.prod.example.com/v1", "Production URL", true},
		{"production url", "https://production.example.com", "Production URL", true},
		{"live url", "https://live.example.com/api", "Production URL", true},
		{"staging url negative", "https://staging.example.com", "Production URL", false},
		{"localhost negative", "http://localhost:3000", "Production URL", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var matched bool
			for _, p := range patterns {
				if p.name == tt.pattern {
					matched = p.regex.MatchString(tt.value)
					break
				}
			}
			if matched != tt.match {
				t.Errorf("pattern %q against %q: got %v, want %v",
					tt.pattern, tt.value, matched, tt.match)
			}
		})
	}
}

func TestWeakValues(t *testing.T) {
	tests := []struct {
		key   string
		value string
		warn  bool
	}{
		{"DB_PASSWORD", "password", true},
		{"SECRET", "123456", true},
		{"API_KEY", "secret", true},
		{"TOKEN", "changeme", true},
		{"DEBUG", "test", true},
		{"URL", "example", true},
		{"NOTE", "TODO", true},
		{"DB_PASSWORD", "correct-horse-battery-staple", false},
	}

	for _, tt := range tests {
		t.Run(tt.key+"="+tt.value, func(t *testing.T) {
			findings := Scan([]envfile.Entry{{Key: tt.key, Value: tt.value}})
			hasWarn := false
			for _, f := range findings {
				if f.Severity == SeverityWarn {
					hasWarn = true
					break
				}
			}
			if hasWarn != tt.warn {
				t.Errorf("weak value check for %s=%s: got warn=%v, want %v",
					tt.key, tt.value, hasWarn, tt.warn)
			}
		})
	}
}

func TestShortSecrets(t *testing.T) {
	tests := []struct {
		key       string
		value     string
		warnShort bool
	}{
		{"API_SECRET", "abc", true},
		{"AUTH_TOKEN", "short", true},
		{"DB_PASSWORD", "1234567", true},
		{"APP_KEY", "x", true},
		{"API_SECRET", "long-enough-value", false},
		{"DEBUG", "true", false}, // not a sensitive key name
		{"APP_NAME", "myapp", false},
	}

	for _, tt := range tests {
		t.Run(tt.key+"="+tt.value, func(t *testing.T) {
			findings := Scan([]envfile.Entry{{Key: tt.key, Value: tt.value}})
			hasShortWarn := false
			for _, f := range findings {
				if f.Severity == SeverityWarn && contains(f.Message, "short secret") {
					hasShortWarn = true
					break
				}
			}
			if hasShortWarn != tt.warnShort {
				t.Errorf("short secret check for %s=%s: got warn=%v, want %v",
					tt.key, tt.value, hasShortWarn, tt.warnShort)
			}
		})
	}
}

func TestScanRealisticEnv(t *testing.T) {
	entries := []envfile.Entry{
		{Key: "DB_HOST", Value: "localhost"},
		{Key: "DB_PORT", Value: "5432"},
		{Key: "AWS_ACCESS_KEY_ID", Value: "AKIAIOSFODNN7EXAMPLE"},
		{Key: "STRIPE_SECRET_KEY", Value: "sk_live_4eC39HqLyjWDarjtT1zdp7dc"},
		{Key: "GITHUB_TOKEN", Value: "ghp_ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghij"},
		{Key: "DB_PASSWORD", Value: "password"},
		{Key: "API_SECRET", Value: "abc"},
		{Key: "APP_URL", Value: "https://api.prod.example.com"},
		{Key: "DEBUG", Value: "true"},
		{Key: "TLS_KEY", Value: "-----BEGIN RSA PRIVATE KEY-----"},
	}

	findings := Scan(entries)

	warnCount := 0
	okCount := 0
	for _, f := range findings {
		switch f.Severity {
		case SeverityWarn:
			warnCount++
		case SeverityOK:
			okCount++
		}
	}

	// We expect warnings for: AWS key, Stripe key, GitHub token, DB_PASSWORD
	// (weak + short), API_SECRET (short), APP_URL (prod URL), TLS_KEY (private key + short).
	if warnCount == 0 {
		t.Error("expected warnings for realistic .env, got 0")
	}

	// DB_HOST, DB_PORT, DEBUG should be OK.
	if okCount < 3 {
		t.Errorf("expected at least 3 OK findings, got %d", okCount)
	}

	// Verify specific entries got flagged.
	warnKeys := make(map[string]bool)
	for _, f := range findings {
		if f.Severity == SeverityWarn {
			warnKeys[f.Key] = true
		}
	}

	mustWarn := []string{
		"AWS_ACCESS_KEY_ID", "STRIPE_SECRET_KEY", "GITHUB_TOKEN",
		"DB_PASSWORD", "API_SECRET", "APP_URL", "TLS_KEY",
	}
	for _, key := range mustWarn {
		if !warnKeys[key] {
			t.Errorf("expected warning for %s, but got none", key)
		}
	}
}

func TestScanCleanEnv(t *testing.T) {
	entries := []envfile.Entry{
		{Key: "DB_HOST", Value: "localhost"},
		{Key: "DB_PORT", Value: "5432"},
		{Key: "DEBUG", Value: "false"},
	}

	findings := Scan(entries)
	for _, f := range findings {
		if f.Severity == SeverityWarn {
			t.Errorf("unexpected warning for clean env: %s — %s", f.Key, f.Message)
		}
	}
}

func contains(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
