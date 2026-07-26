package scanner

import (
	"strings"
	"testing"

	"github.com/anukool23/secretcheck/internal/rules"
)

func TestDetectsAWSAccessKey(t *testing.T) {
	findings := ScanContent("a.env", `AWS_KEY = "AKIAABCDEFGHIJKLWXYZ"`, rules.Default())
	if len(findings) != 1 {
		t.Fatalf("expected 1 finding, got %d", len(findings))
	}
	if findings[0].RuleID != "aws-access-key-id" {
		t.Errorf("expected aws-access-key-id, got %s", findings[0].RuleID)
	}
	if findings[0].Line != 1 {
		t.Errorf("expected line 1, got %d", findings[0].Line)
	}
}

func TestDetectsGitHubToken(t *testing.T) {
	content := "token = ghp_" + strings.Repeat("a", 36)
	findings := ScanContent("a.txt", content, rules.Default())
	found := false
	for _, f := range findings {
		if f.RuleID == "github-token" {
			found = true
		}
	}
	if !found {
		t.Errorf("expected a github-token finding, got %+v", findings)
	}
}

func TestDetectsPrivateKeyBlock(t *testing.T) {
	content := "-----BEGIN RSA PRIVATE KEY-----\nMIIB...\n-----END RSA PRIVATE KEY-----"
	findings := ScanContent("id_rsa", content, rules.Default())
	found := false
	for _, f := range findings {
		if f.RuleID == "private-key-block" {
			found = true
		}
	}
	if !found {
		t.Errorf("expected a private-key-block finding, got %+v", findings)
	}
}

func TestIgnoresPlaceholderValues(t *testing.T) {
	findings := ScanContent("config.js", `const password = "changeme";`, rules.Default())
	if len(findings) != 0 {
		t.Errorf("expected 0 findings for placeholder value, got %+v", findings)
	}
}

func TestCleanCodeNotFlagged(t *testing.T) {
	content := "console.log(\"hello world\");\nconst x = 1 + 2;"
	findings := ScanContent("index.js", content, rules.Default())
	if len(findings) != 0 {
		t.Errorf("expected 0 findings for clean code, got %+v", findings)
	}
}

func TestDisableLineMarker(t *testing.T) {
	content := `AWS_KEY = "AKIAABCDEFGHIJKLWXYZ" // secretcheck-disable-line`
	findings := ScanContent("a.env", content, rules.Default())
	if len(findings) != 0 {
		t.Errorf("expected 0 findings with disable-line marker, got %+v", findings)
	}
}

func TestDisableNextLineMarker(t *testing.T) {
	content := "// secretcheck-disable-next-line\nAWS_KEY = \"AKIAABCDEFGHIJKLWXYZ\""
	findings := ScanContent("a.env", content, rules.Default())
	if len(findings) != 0 {
		t.Errorf("expected 0 findings with disable-next-line marker, got %+v", findings)
	}
}

func TestRedactsMatchedValue(t *testing.T) {
	findings := ScanContent("a.env", `AWS_KEY = "AKIAABCDEFGHIJKLWXYZ"`, rules.Default())
	if len(findings) != 1 {
		t.Fatalf("expected 1 finding, got %d", len(findings))
	}
	if strings.Contains(findings[0].Match, "ABCDEFGHIJKL") {
		t.Errorf("expected redacted match, got %q", findings[0].Match)
	}
}
