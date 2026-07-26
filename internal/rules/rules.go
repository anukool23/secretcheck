// Package rules holds the built-in secret detection patterns and the
// helpers used to redact/filter matches.
package rules

import "regexp"

// Rule is a single detector: an id (for disabling via config), a
// human-readable description shown in findings output, and the regexp
// used to test each line of staged content.
type Rule struct {
	ID          string
	Description string
	Regex       *regexp.Regexp
}

// Default returns a fresh copy of the built-in rule set. A fresh copy is
// returned (rather than a shared slice) so callers can safely filter it
// without affecting other callers.
func Default() []Rule {
	return []Rule{
		{
			ID:          "aws-access-key-id",
			Description: "AWS Access Key ID",
			Regex:       regexp.MustCompile(`\b(AKIA|ABIA|ACCA|ASIA)[0-9A-Z]{16}\b`),
		},
		{
			ID:          "aws-secret-access-key",
			Description: "AWS Secret Access Key",
			Regex:       regexp.MustCompile(`(?i)aws_secret(_access)?_key\s*[:=]\s*["']?[A-Za-z0-9/+=]{40}["']?`),
		},
		{
			ID:          "github-token",
			Description: "GitHub Token",
			Regex:       regexp.MustCompile(`gh[pousr]_[A-Za-z0-9]{36,255}`),
		},
		{
			ID:          "github-fine-grained-pat",
			Description: "GitHub Fine-Grained Personal Access Token",
			Regex:       regexp.MustCompile(`github_pat_[A-Za-z0-9_]{80,}`),
		},
		{
			ID:          "slack-token",
			Description: "Slack Token",
			Regex:       regexp.MustCompile(`xox[baprs]-[0-9A-Za-z-]{10,72}`),
		},
		{
			ID:          "slack-webhook",
			Description: "Slack Webhook URL",
			Regex:       regexp.MustCompile(`hooks\.slack\.com/services/T[0-9A-Za-z]{6,}/B[0-9A-Za-z]{6,}/[0-9A-Za-z]{16,}`),
		},
		{
			ID:          "google-api-key",
			Description: "Google API Key",
			Regex:       regexp.MustCompile(`AIza[0-9A-Za-z\-_]{35}`),
		},
		{
			ID:          "google-oauth-client-id",
			Description: "Google OAuth Client ID",
			Regex:       regexp.MustCompile(`[0-9]+-[0-9A-Za-z_]{32}\.apps\.googleusercontent\.com`),
		},
		{
			ID:          "stripe-live-key",
			Description: "Stripe Live Secret/Restricted Key",
			Regex:       regexp.MustCompile(`\b(sk|rk)_live_[0-9a-zA-Z]{20,247}\b`),
		},
		{
			ID:          "twilio-api-key",
			Description: "Twilio API Key",
			Regex:       regexp.MustCompile(`\bSK[0-9a-fA-F]{32}\b`),
		},
		{
			ID:          "mailgun-api-key",
			Description: "Mailgun API Key",
			Regex:       regexp.MustCompile(`\bkey-[0-9a-zA-Z]{32}\b`),
		},
		{
			ID:          "sendgrid-api-key",
			Description: "SendGrid API Key",
			Regex:       regexp.MustCompile(`\bSG\.[0-9A-Za-z\-_]{22}\.[0-9A-Za-z\-_]{43}\b`),
		},
		{
			ID:          "npm-token",
			Description: "npm Access Token",
			Regex:       regexp.MustCompile(`\bnpm_[A-Za-z0-9]{36}\b`),
		},
		{
			ID:          "private-key-block",
			Description: "Private Key Block",
			Regex:       regexp.MustCompile(`-----BEGIN\s?(RSA|DSA|EC|OPENSSH|PGP)?\s?PRIVATE KEY-----`),
		},
		{
			ID:          "jwt",
			Description: "JSON Web Token",
			Regex:       regexp.MustCompile(`\beyJ[A-Za-z0-9_-]{5,}\.[A-Za-z0-9_-]{5,}\.[A-Za-z0-9_-]{5,}\b`),
		},
		{
			ID:          "db-connection-string",
			Description: "Database Connection String with Embedded Credentials",
			Regex:       regexp.MustCompile(`(?i)(mongodb(\+srv)?|postgres(ql)?|mysql|redis)://[^:/\s'"]+:[^@/\s'"]+@[^/\s'"]+`),
		},
		{
			ID:          "heroku-api-key",
			Description: "Heroku API Key",
			Regex:       regexp.MustCompile(`(?i)heroku[a-z0-9_-]*\s*[:=]\s*["']?[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}["']?`),
		},
		{
			ID:          "generic-bearer-token",
			Description: "Bearer Token",
			Regex:       regexp.MustCompile(`(?i)bearer\s+[A-Za-z0-9\-_.=]{16,}`),
		},
		{
			ID:          "generic-api-key",
			Description: "Generic API Key / Secret / Password Assignment",
			Regex:       regexp.MustCompile(`(?i)(api[_-]?key|apikey|secret[_-]?key|client[_-]?secret|access[_-]?token|auth[_-]?token|private[_-]?key|password|passwd|pwd)\s*[:=]\s*["']([A-Za-z0-9/+_.-]{8,})["']`),
		},
	}
}

// placeholderPatterns match captured values that commonly trigger the
// generic assignment rule but are not real secrets.
var placeholderPatterns = []*regexp.Regexp{
	regexp.MustCompile(`(?i)^x{4,}$`),
	regexp.MustCompile(`^\*{4,}$`),
	regexp.MustCompile(`(?i)^changeme$`),
	regexp.MustCompile(`(?i)^change[-_]?me[-_]?please$`),
	regexp.MustCompile(`(?i)^example$`),
	regexp.MustCompile(`(?i)^placeholder$`),
	regexp.MustCompile(`(?i)^dummy$`),
	regexp.MustCompile(`(?i)^sample$`),
	regexp.MustCompile(`(?i)^test(ing)?$`),
	regexp.MustCompile(`(?i)^fake$`),
	regexp.MustCompile(`(?i)^redacted$`),
	regexp.MustCompile(`(?i)^your[-_].*$`),
	regexp.MustCompile(`(?i)^insert[-_].*$`),
	regexp.MustCompile(`^<.*>$`),
	regexp.MustCompile(`^\$\{.*\}$`),
	regexp.MustCompile(`^\{\{.*\}\}$`),
	regexp.MustCompile(`(?i)^todo$`),
	regexp.MustCompile(`(?i)^n/a$`),
	regexp.MustCompile(`(?i)^none$`),
	regexp.MustCompile(`(?i)^null$`),
	regexp.MustCompile(`(?i)^undefined$`),
	regexp.MustCompile(`^0+$`),
	regexp.MustCompile(`^1234(5678)?9?0?$`),
}

// IsLikelyPlaceholder reports whether value looks like a placeholder/dummy
// value rather than a real secret (used to cut false positives on the
// broad generic-api-key rule).
func IsLikelyPlaceholder(value string) bool {
	for _, p := range placeholderPatterns {
		if p.MatchString(value) {
			return true
		}
	}
	return false
}

// Redact masks the middle of a matched secret for safe display, keeping a
// few characters on each end so a finding can still be located/verified.
func Redact(value string) string {
	if len(value) <= 8 {
		return repeat('*', len(value))
	}
	head := value[:4]
	tail := value[len(value)-4:]
	maskLen := len(value) - 8
	if maskLen < 4 {
		maskLen = 4
	}
	return head + repeat('*', maskLen) + tail
}

func repeat(r rune, n int) string {
	out := make([]rune, n)
	for i := range out {
		out[i] = r
	}
	return string(out)
}
