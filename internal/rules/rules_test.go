package rules

import "testing"

func TestRedactShortValue(t *testing.T) {
	got := Redact("abc")
	if got != "***" {
		t.Errorf("Redact(abc) = %q, want ***", got)
	}
}

func TestRedactLongValueKeepsEnds(t *testing.T) {
	got := Redact("AKIAABCDEFGHIJKLWXYZ")
	if got[:4] != "AKIA" {
		t.Errorf("expected prefix AKIA, got %q", got)
	}
	if got[len(got)-4:] != "WXYZ" {
		t.Errorf("expected suffix WXYZ, got %q", got)
	}
	if got == "AKIAABCDEFGHIJKLWXYZ" {
		t.Error("expected value to be masked, got original")
	}
}

func TestIsLikelyPlaceholder(t *testing.T) {
	placeholders := []string{"changeme", "CHANGEME", "xxxx", "example", "<your-key>", "${SECRET}", "0000"}
	for _, v := range placeholders {
		if !IsLikelyPlaceholder(v) {
			t.Errorf("expected %q to be treated as a placeholder", v)
		}
	}

	real := []string{"AKIAABCDEFGHIJKLWXYZ", "gh" + "p_abcdefghijklmnopqrstuvwxyz1234567890"}
	for _, v := range real {
		if IsLikelyPlaceholder(v) {
			t.Errorf("expected %q to NOT be treated as a placeholder", v)
		}
	}
}
