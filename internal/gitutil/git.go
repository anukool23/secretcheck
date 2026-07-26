// Package gitutil wraps the small set of git plumbing commands secretcheck
// needs: listing staged/tracked files, reading staged file content, and
// locating the repo root and hooks directory.
package gitutil

import (
	"fmt"
	"os/exec"
	"strings"
)

// Error wraps a failed git invocation with its stderr output.
type Error struct {
	Args   []string
	Stderr string
}

func (e *Error) Error() string {
	return fmt.Sprintf("git %s failed: %s", strings.Join(e.Args, " "), strings.TrimSpace(e.Stderr))
}

func run(dir string, args ...string) (string, error) {
	cmd := exec.Command("git", args...)
	if dir != "" {
		cmd.Dir = dir
	}
	var stderr strings.Builder
	cmd.Stderr = &stderr
	out, err := cmd.Output()
	if err != nil {
		return "", &Error{Args: args, Stderr: stderr.String()}
	}
	return string(out), nil
}

func splitLines(s string) []string {
	lines := strings.Split(strings.TrimRight(s, "\n"), "\n")
	out := make([]string, 0, len(lines))
	for _, l := range lines {
		l = strings.TrimSpace(l)
		if l != "" {
			out = append(out, l)
		}
	}
	return out
}

// IsRepo reports whether dir (or the current directory, if empty) is
// inside a git working tree.
func IsRepo(dir string) bool {
	_, err := run(dir, "rev-parse", "--is-inside-work-tree")
	return err == nil
}

// RepoRoot returns the absolute path to the top level of the working tree.
func RepoRoot(dir string) (string, error) {
	out, err := run(dir, "rev-parse", "--show-toplevel")
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(out), nil
}

// HooksDir returns the effective hooks directory (respects core.hooksPath).
func HooksDir(dir string) (string, error) {
	out, err := run(dir, "rev-parse", "--git-path", "hooks")
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(out), nil
}

// StagedFiles returns staged files that are added, copied, modified, or
// renamed (i.e. have index content worth checking).
func StagedFiles(dir string) ([]string, error) {
	out, err := run(dir, "diff", "--cached", "--name-only", "--diff-filter=ACMR")
	if err != nil {
		return nil, err
	}
	return splitLines(out), nil
}

// AllTrackedFiles returns every file tracked by git.
func AllTrackedFiles(dir string) ([]string, error) {
	out, err := run(dir, "ls-files")
	if err != nil {
		return nil, err
	}
	return splitLines(out), nil
}

// BinaryStagedFiles returns the set of staged files git considers binary,
// derived from `diff --numstat` reporting "-" for both add/remove counts.
func BinaryStagedFiles(dir string) (map[string]bool, error) {
	out, err := run(dir, "diff", "--cached", "--numstat")
	if err != nil {
		return nil, err
	}
	binary := map[string]bool{}
	for _, line := range strings.Split(out, "\n") {
		if strings.TrimSpace(line) == "" {
			continue
		}
		parts := strings.SplitN(line, "\t", 3)
		if len(parts) < 3 {
			continue
		}
		if parts[0] == "-" && parts[1] == "-" {
			binary[strings.TrimSpace(parts[2])] = true
		}
	}
	return binary, nil
}

// StagedFileContent returns the content of file as it exists in the index
// (the staged version), not the working tree.
func StagedFileContent(dir, file string) (string, error) {
	return run(dir, "show", ":"+file)
}
