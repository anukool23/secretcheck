// Package ignore implements gitignore-lite glob matching used to exclude
// paths from scanning, plus the default exclude list and .secretcheckignore
// file loading. It intentionally has no third-party dependency: glob
// patterns are compiled to a small regexp instead of pulling in a glob
// library.
package ignore

import (
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strings"
)

// DefaultExcludes are always skipped, on top of anything in
// .secretcheckignore or the config file.
var DefaultExcludes = []string{
	"**/node_modules/**",
	"**/.git/**",
	"**/dist/**",
	"**/build/**",
	"**/coverage/**",
	"**/*.min.js",
	"**/*.map",
	"package-lock.json",
	"**/package-lock.json",
	"yarn.lock",
	"**/yarn.lock",
	"pnpm-lock.yaml",
	"**/pnpm-lock.yaml",
	"go.sum",
	"**/go.sum",
	"**/*.png",
	"**/*.jpg",
	"**/*.jpeg",
	"**/*.gif",
	"**/*.ico",
	"**/*.webp",
	"**/*.pdf",
	"**/*.woff",
	"**/*.woff2",
	"**/*.ttf",
	"**/*.eot",
	"**/*.zip",
	"**/*.gz",
	"**/*.tar",
	"**/*.jar",
	"**/*.class",
	"**/*.exe",
	"**/*.dll",
	"**/*.so",
	"**/*.dylib",
}

const ignoreFileName = ".secretcheckignore"

// LoadIgnoreFile reads .secretcheckignore from the repo root, if present.
// Format is gitignore-lite: one glob per line; blank lines and lines
// starting with "#" are skipped.
func LoadIgnoreFile(repoRoot string) ([]string, error) {
	p := filepath.Join(repoRoot, ignoreFileName)
	data, err := os.ReadFile(p)
	if os.IsNotExist(err) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	var patterns []string
	for _, line := range strings.Split(string(data), "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		patterns = append(patterns, line)
	}
	return patterns, nil
}

var globCache = map[string]*regexp.Regexp{}

// compileGlob converts a gitignore-lite pattern ("*", "**", "?", literal
// path segments separated by "/") into an anchored regexp.
func compileGlob(pattern string) *regexp.Regexp {
	if re, ok := globCache[pattern]; ok {
		return re
	}

	var sb strings.Builder
	sb.WriteString("^")

	runes := []rune(pattern)
	for i := 0; i < len(runes); {
		c := runes[i]
		switch {
		case c == '*' && i+1 < len(runes) && runes[i+1] == '*':
			j := i
			for j < len(runes) && runes[j] == '*' {
				j++
			}
			precededBySlashOrStart := i == 0 || runes[i-1] == '/'
			followedBySlash := j < len(runes) && runes[j] == '/'
			if precededBySlashOrStart && followedBySlash {
				sb.WriteString("(?:.*/)?")
				j++ // also consume the slash
			} else {
				sb.WriteString(".*")
			}
			i = j
		case c == '*':
			sb.WriteString("[^/]*")
			i++
		case c == '?':
			sb.WriteString("[^/]")
			i++
		default:
			sb.WriteString(regexp.QuoteMeta(string(c)))
			i++
		}
	}
	sb.WriteString("$")

	re := regexp.MustCompile(sb.String())
	globCache[pattern] = re
	return re
}

// IsExcluded reports whether file (a repo-relative path) matches any of
// the given glob patterns. Patterns with no "/" also match against the
// file's base name at any depth, mirroring gitignore semantics.
func IsExcluded(file string, patterns []string) bool {
	normalized := filepath.ToSlash(file)
	base := path.Base(normalized)
	for _, p := range patterns {
		re := compileGlob(p)
		if re.MatchString(normalized) {
			return true
		}
		if !strings.Contains(p, "/") && re.MatchString(base) {
			return true
		}
	}
	return false
}
