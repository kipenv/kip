package scanner

import "regexp"

// pattern defines a regex-based check for sensitive values.
type pattern struct {
	name    string
	regex   *regexp.Regexp
	message string
}

// patterns holds all regex checks run against entry values.
var patterns = []pattern{
	{
		name:    "AWS Access Key",
		regex:   regexp.MustCompile(`AKIA[0-9A-Z]{16}`),
		message: "looks like an AWS access key ID",
	},
	{
		name:    "Stripe Live Key",
		regex:   regexp.MustCompile(`sk_live_[a-zA-Z0-9]+`),
		message: "looks like a Stripe live secret key",
	},
	{
		name:    "GitHub Token",
		regex:   regexp.MustCompile(`ghp_[a-zA-Z0-9]{36}`),
		message: "looks like a GitHub personal access token",
	},
	{
		name:    "Private Key",
		regex:   regexp.MustCompile(`-----BEGIN (RSA |EC )?PRIVATE KEY-----`),
		message: "contains a private key block",
	},
	{
		name:    "Production URL",
		regex:   regexp.MustCompile(`(?i)https?://[^\s]*(prod|production|live)[^\s]*`),
		message: "URL references a production/live environment",
	},
}

// weakValues are common placeholder or test values that should never be shared.
var weakValues = []string{
	"password",
	"123456",
	"secret",
	"changeme",
	"test",
	"example",
	"TODO",
}

// sensitiveKeyParts are substrings in key names that indicate the value
// should be treated as a secret.
var sensitiveKeyParts = []string{
	"SECRET",
	"KEY",
	"TOKEN",
	"PASSWORD",
	"PASS",
}
