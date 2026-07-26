// Package hook installs and removes the secretcheck pre-commit hook,
// either as a native git hook or into an existing Husky setup.
package hook

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/anukool23/secretcheck/internal/gitutil"
)

const (
	startMarker = "# >>> secretcheck >>>"
	endMarker   = "# <<< secretcheck <<<"
)

var hookBlock = strings.Join([]string{
	startMarker,
	`SECRETCHECK_BIN="$(git rev-parse --show-toplevel 2>/dev/null)/node_modules/.bin/secretcheck"`,
	`if [ -x "$SECRETCHECK_BIN" ]; then`,
	`  "$SECRETCHECK_BIN" scan`,
	`elif command -v secretcheck >/dev/null 2>&1; then`,
	`  secretcheck scan`,
	`else`,
	`  npx --yes secretcheck scan`,
	`fi`,
	`SECRETCHECK_EXIT_CODE=$?`,
	`if [ $SECRETCHECK_EXIT_CODE -ne 0 ]; then`,
	`  exit $SECRETCHECK_EXIT_CODE`,
	`fi`,
	endMarker,
}, "\n")

// Target describes where the pre-commit hook lives for this repo.
type Target struct {
	HookPath string
	IsHusky  bool
}

// ResolveTarget determines whether this repo uses Husky (.husky/pre-commit)
// or a native git hook, and returns the absolute path to write/read.
func ResolveTarget(dir string) (Target, error) {
	repoRoot, err := gitutil.RepoRoot(dir)
	if err != nil {
		return Target{}, err
	}

	huskyDir := filepath.Join(repoRoot, ".husky")
	if info, err := os.Stat(huskyDir); err == nil && info.IsDir() {
		return Target{HookPath: filepath.Join(huskyDir, "pre-commit"), IsHusky: true}, nil
	}

	hooksDir, err := gitutil.HooksDir(dir)
	if err != nil {
		return Target{}, err
	}
	if !filepath.IsAbs(hooksDir) {
		base := dir
		if base == "" {
			base, err = os.Getwd()
			if err != nil {
				return Target{}, err
			}
		}
		hooksDir = filepath.Join(base, hooksDir)
	}
	return Target{HookPath: filepath.Join(hooksDir, "pre-commit"), IsHusky: false}, nil
}

// InstallStatus reports what Install actually did.
type InstallStatus string

const Created InstallStatus = "created"
const Updated InstallStatus = "updated"
const AlreadyInstalled InstallStatus = "already-installed"

// Install writes (or merges) the secretcheck block into the pre-commit
// hook. If force is false and the block is already present, it is left
// untouched and AlreadyInstalled is returned.
func Install(dir string, force bool) (InstallStatus, error) {
	target, err := ResolveTarget(dir)
	if err != nil {
		return "", err
	}

	if err := os.MkdirAll(filepath.Dir(target.HookPath), 0o755); err != nil {
		return "", err
	}

	existing, err := os.ReadFile(target.HookPath)
	if os.IsNotExist(err) {
		content := "#!/bin/sh\n" + hookBlock + "\n"
		if err := os.WriteFile(target.HookPath, []byte(content), 0o755); err != nil {
			return "", err
		}
		return Created, nil
	}
	if err != nil {
		return "", err
	}

	existingStr := string(existing)
	if strings.Contains(existingStr, startMarker) {
		if !force {
			return AlreadyInstalled, nil
		}
		replaced := blockRegex().ReplaceAllString(existingStr, hookBlock+"\n")
		if err := os.WriteFile(target.HookPath, []byte(replaced), 0o755); err != nil {
			return "", err
		}
		return Updated, nil
	}

	// Preserve any existing hook logic; run our check first (fail fast),
	// right after the shebang line if there is one.
	lines := strings.Split(existingStr, "\n")
	insertAt := 0
	if len(lines) > 0 && strings.HasPrefix(lines[0], "#!") {
		insertAt = 1
	}
	newLines := make([]string, 0, len(lines)+3)
	newLines = append(newLines, lines[:insertAt]...)
	newLines = append(newLines, "", hookBlock, "")
	newLines = append(newLines, lines[insertAt:]...)

	if err := os.WriteFile(target.HookPath, []byte(strings.Join(newLines, "\n")), 0o755); err != nil {
		return "", err
	}
	return Updated, nil
}

// UninstallStatus reports what Uninstall actually did.
type UninstallStatus string

const Removed UninstallStatus = "removed"
const NotInstalled UninstallStatus = "not-installed"
const DeletedFile UninstallStatus = "deleted-file"

// Uninstall removes the secretcheck block from the pre-commit hook,
// deleting the file entirely if nothing else remains in it.
func Uninstall(dir string) (UninstallStatus, error) {
	target, err := ResolveTarget(dir)
	if err != nil {
		return "", err
	}

	existing, err := os.ReadFile(target.HookPath)
	if os.IsNotExist(err) {
		return NotInstalled, nil
	}
	if err != nil {
		return "", err
	}

	existingStr := string(existing)
	if !strings.Contains(existingStr, startMarker) {
		return NotInstalled, nil
	}

	withoutBlock := blockRegex().ReplaceAllString(existingStr, "")
	withoutBlock = collapseBlankLines(withoutBlock)

	remainder := strings.TrimSpace(shebangRegex().ReplaceAllString(withoutBlock, ""))
	if remainder == "" {
		if err := os.Remove(target.HookPath); err != nil {
			return "", err
		}
		return DeletedFile, nil
	}

	if err := os.WriteFile(target.HookPath, []byte(withoutBlock), 0o755); err != nil {
		return "", err
	}
	return Removed, nil
}

func blockRegex() *regexp.Regexp {
	return regexp.MustCompile(`(?s)` + regexp.QuoteMeta(startMarker) + `.*?` + regexp.QuoteMeta(endMarker) + `\n?`)
}

func shebangRegex() *regexp.Regexp {
	return regexp.MustCompile(`^#!.*\n`)
}

func collapseBlankLines(s string) string {
	return regexp.MustCompile(`\n{3,}`).ReplaceAllString(s, "\n\n")
}
