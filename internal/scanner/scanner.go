// Package scanner runs the rule set against staged or working-tree files.
package scanner

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"

	"github.com/anukool23/secretcheck/internal/gitutil"
	"github.com/anukool23/secretcheck/internal/ignore"
	"github.com/anukool23/secretcheck/internal/rules"
)

const (
	disableLineMarker     = "secretcheck-disable-line"
	disableNextLineMarker = "secretcheck-disable-next-line"
	maxContextLength      = 160
)

// Finding is a single detected secret.
type Finding struct {
	File        string `json:"file"`
	Line        int    `json:"line"`
	RuleID      string `json:"ruleId"`
	Description string `json:"description"`
	Match       string `json:"match"`   // redacted, safe to print
	Context     string `json:"context"` // trimmed source line, truncated
}

// Result summarizes a scan run.
type Result struct {
	Findings     []Finding
	FilesScanned int
	FilesSkipped int
}

func truncate(s string, max int) string {
	r := []rune(s)
	if len(r) > max {
		return string(r[:max]) + "…"
	}
	return s
}

// ScanContent scans a single file's content, line by line, against rules.
func ScanContent(file, content string, ruleSet []rules.Rule) []Finding {
	var findings []Finding
	lines := strings.Split(content, "\n")
	suppressNext := false

	for i, line := range lines {
		suppressThis := suppressNext || strings.Contains(line, disableLineMarker)
		suppressNext = strings.Contains(line, disableNextLineMarker)

		if suppressThis {
			continue
		}

		for _, rule := range ruleSet {
			matches := rule.Regex.FindAllStringSubmatch(line, -1)
			for _, m := range matches {
				valueToCheck := m[0]
				if rule.ID == "generic-api-key" && len(m) > 2 {
					valueToCheck = m[2]
				}
				if rules.IsLikelyPlaceholder(valueToCheck) {
					continue
				}
				findings = append(findings, Finding{
					File:        file,
					Line:        i + 1,
					RuleID:      rule.ID,
					Description: rule.Description,
					Match:       rules.Redact(m[0]),
					Context:     truncate(strings.TrimSpace(line), maxContextLength),
				})
			}
		}
	}

	return findings
}

func buildExcludes(extra []string) []string {
	out := make([]string, 0, len(ignore.DefaultExcludes)+len(extra))
	out = append(out, ignore.DefaultExcludes...)
	out = append(out, extra...)
	return out
}

// ScanStaged scans staged (indexed) files — what runs inside the
// pre-commit hook.
func ScanStaged(repoRoot string, ruleSet []rules.Rule, extraExcludes []string) (Result, error) {
	excludes := buildExcludes(extraExcludes)

	staged, err := gitutil.StagedFiles(repoRoot)
	if err != nil {
		return Result{}, err
	}
	binary, err := gitutil.BinaryStagedFiles(repoRoot)
	if err != nil {
		return Result{}, err
	}

	var result Result
	for _, file := range staged {
		if binary[file] || ignore.IsExcluded(file, excludes) {
			result.FilesSkipped++
			continue
		}
		content, err := gitutil.StagedFileContent(repoRoot, file)
		if err != nil {
			// File may have been deleted/renamed between listing and reading.
			result.FilesSkipped++
			continue
		}
		result.FilesScanned++
		result.Findings = append(result.Findings, ScanContent(file, content, ruleSet)...)
	}
	return result, nil
}

// ScanAll scans every git-tracked file on disk — useful for a one-off
// repo-wide audit.
func ScanAll(repoRoot string, ruleSet []rules.Rule, extraExcludes []string) (Result, error) {
	excludes := buildExcludes(extraExcludes)

	files, err := gitutil.AllTrackedFiles(repoRoot)
	if err != nil {
		return Result{}, err
	}

	var result Result
	for _, file := range files {
		if ignore.IsExcluded(file, excludes) {
			result.FilesSkipped++
			continue
		}
		absPath := filepath.Join(repoRoot, file)
		info, err := os.Stat(absPath)
		if err != nil || !info.Mode().IsRegular() {
			result.FilesSkipped++
			continue
		}
		data, err := os.ReadFile(absPath)
		if err != nil || isBinary(data) {
			result.FilesSkipped++
			continue
		}
		result.FilesScanned++
		result.Findings = append(result.Findings, ScanContent(file, string(data), ruleSet)...)
	}
	return result, nil
}

// isBinary is a cheap sniff: presence of a NUL byte in the first 8KB.
func isBinary(data []byte) bool {
	sample := data
	if len(sample) > 8000 {
		sample = sample[:8000]
	}
	return bytes.IndexByte(sample, 0) != -1
}
