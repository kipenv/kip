// Package scanner checks .env entries for leaked credentials, weak secrets,
// and other security issues using regex-based pattern matching.
package scanner

import (
	"fmt"
	"strings"

	"github.com/kipenv/kip/internal/envfile"
)

// Severity indicates the risk level of a finding.
type Severity string

const (
	SeverityWarn Severity = "WARN"
	SeverityOK   Severity = "OK"
)

// Finding represents a single scan result for an entry.
type Finding struct {
	Key      string
	Severity Severity
	Message  string
}

// Scan checks every entry against known patterns and heuristics, returning
// one Finding per entry. Entries that trigger no patterns get SeverityOK.
func Scan(entries []envfile.Entry) []Finding {
	findings := make([]Finding, 0, len(entries))

	for _, e := range entries {
		warns := check(e)
		if len(warns) > 0 {
			for _, msg := range warns {
				findings = append(findings, Finding{
					Key:      e.Key,
					Severity: SeverityWarn,
					Message:  msg,
				})
			}
		} else {
			findings = append(findings, Finding{
				Key:      e.Key,
				Severity: SeverityOK,
				Message:  "no issues detected",
			})
		}
	}

	return findings
}

// check runs all rules against a single entry and returns warning messages.
func check(e envfile.Entry) []string {
	var warns []string

	// Regex pattern matching against the value.
	for _, p := range patterns {
		if p.regex.MatchString(e.Value) {
			warns = append(warns, fmt.Sprintf("%s — %s", p.name, p.message))
		}
	}

	// Common weak/test values (case-insensitive).
	lower := strings.ToLower(e.Value)
	for _, weak := range weakValues {
		if lower == strings.ToLower(weak) {
			warns = append(warns, fmt.Sprintf("weak value — \"%s\" is a common test/default value", e.Value))
			break
		}
	}

	// Short secret detection: if the key name suggests a secret and the
	// value is under 8 characters, flag it.
	upperKey := strings.ToUpper(e.Key)
	if isSensitiveKey(upperKey) && len(e.Value) > 0 && len(e.Value) < 8 {
		warns = append(warns, fmt.Sprintf("short secret — value is only %d chars for a key that looks sensitive", len(e.Value)))
	}

	return warns
}

// isSensitiveKey returns true if the key name contains words that suggest
// the value should be a strong secret.
func isSensitiveKey(upperKey string) bool {
	for _, part := range sensitiveKeyParts {
		if strings.Contains(upperKey, part) {
			return true
		}
	}
	return false
}
