// Package config loads .secretcheckrc.json and resolves it against the
// built-in rule set (disabling/adding rules).
package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"

	"github.com/anukool23/secretcheck/internal/rules"
)

const configFileName = ".secretcheckrc.json"

// CustomRule is a user-supplied detection pattern read from
// .secretcheckrc.json. Pattern is a Go regexp (RE2) source string; for
// case-insensitive matching either set CaseInsensitive or prefix the
// pattern yourself with "(?i)".
type CustomRule struct {
	ID              string `json:"id"`
	Description     string `json:"description"`
	Pattern         string `json:"pattern"`
	CaseInsensitive bool   `json:"caseInsensitive"`
}

// Config is the shape of .secretcheckrc.json.
type Config struct {
	IgnorePaths  []string     `json:"ignorePaths"`
	DisableRules []string     `json:"disableRules"`
	CustomRules  []CustomRule `json:"customRules"`
}

// Load reads .secretcheckrc.json from repoRoot. A missing file is not an
// error; it just means "no overrides".
func Load(repoRoot string) (Config, error) {
	p := filepath.Join(repoRoot, configFileName)
	data, err := os.ReadFile(p)
	if os.IsNotExist(err) {
		return Config{}, nil
	}
	if err != nil {
		return Config{}, err
	}
	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return Config{}, fmt.Errorf("failed to parse %s: %w", configFileName, err)
	}
	return cfg, nil
}

// ResolveRules combines the built-in rule set with the config: removes
// disabled rules and compiles/appends any custom rules.
func ResolveRules(cfg Config) ([]rules.Rule, error) {
	disabled := make(map[string]bool, len(cfg.DisableRules))
	for _, id := range cfg.DisableRules {
		disabled[id] = true
	}

	active := make([]rules.Rule, 0)
	for _, r := range rules.Default() {
		if !disabled[r.ID] {
			active = append(active, r)
		}
	}

	for _, cr := range cfg.CustomRules {
		pattern := cr.Pattern
		if cr.CaseInsensitive {
			pattern = "(?i)" + pattern
		}
		re, err := regexp.Compile(pattern)
		if err != nil {
			return nil, fmt.Errorf("custom rule %q: invalid pattern: %w", cr.ID, err)
		}
		active = append(active, rules.Rule{ID: cr.ID, Description: cr.Description, Regex: re})
	}

	return active, nil
}
