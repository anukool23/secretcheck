// Command secretcheck scans staged git files for secrets and blocks
// commits that contain them, with a clear option to override.
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"

	"github.com/anukool23/secretcheck/internal/colors"
	"github.com/anukool23/secretcheck/internal/config"
	"github.com/anukool23/secretcheck/internal/gitutil"
	"github.com/anukool23/secretcheck/internal/hook"
	"github.com/anukool23/secretcheck/internal/ignore"
	"github.com/anukool23/secretcheck/internal/prompt"
	"github.com/anukool23/secretcheck/internal/report"
	"github.com/anukool23/secretcheck/internal/scanner"
)

// version is overridden at build time via -ldflags "-X main.version=...".
var version = "dev"

const usage = `secretcheck - scan staged git files for secrets before they're committed

Usage:
  secretcheck init [--force]    Install the pre-commit hook
  secretcheck uninstall         Remove the pre-commit hook
  secretcheck scan [flags]      Scan files for secrets
  secretcheck version           Print the version
  secretcheck help              Show this message

scan flags:
  --all         Scan all git-tracked files instead of just staged ones
  --json        Print findings as JSON and skip the interactive prompt
  --no-prompt   Never prompt interactively; exit non-zero on findings
`

func main() {
	if len(os.Args) < 2 {
		fmt.Fprint(os.Stderr, usage)
		os.Exit(1)
	}

	cmd := os.Args[1]
	args := os.Args[2:]

	var err error
	switch cmd {
	case "init":
		err = runInit(args)
	case "uninstall":
		err = runUninstall(args)
	case "scan":
		err = runScan(args)
	case "version", "--version", "-v":
		fmt.Println(version)
		return
	case "help", "--help", "-h":
		fmt.Print(usage)
		return
	default:
		fmt.Fprintf(os.Stderr, "secretcheck: unknown command %q\n\n", cmd)
		fmt.Fprint(os.Stderr, usage)
		os.Exit(1)
	}

	if err != nil {
		fmt.Fprintln(os.Stderr, colors.Red(err.Error()))
		os.Exit(1)
	}
}

func requireGitRepo() error {
	if !gitutil.IsRepo("") {
		return fmt.Errorf("secretcheck must be run inside a git repository")
	}
	return nil
}

func runInit(args []string) error {
	fs := flag.NewFlagSet("init", flag.ExitOnError)
	force := fs.Bool("force", false, "Reinstall/refresh the hook block even if already present")
	if err := fs.Parse(args); err != nil {
		return err
	}

	if err := requireGitRepo(); err != nil {
		return err
	}

	status, err := hook.Install("", *force)
	if err != nil {
		return err
	}
	target, err := hook.ResolveTarget("")
	if err != nil {
		return err
	}

	switch status {
	case hook.Created:
		fmt.Println(colors.Green(fmt.Sprintf("✔ Installed pre-commit hook at %s", target.HookPath)))
	case hook.Updated:
		fmt.Println(colors.Green(fmt.Sprintf("✔ Updated pre-commit hook at %s", target.HookPath)))
	case hook.AlreadyInstalled:
		fmt.Println(colors.Yellow(fmt.Sprintf("secretcheck is already installed at %s. Use --force to reinstall.", target.HookPath)))
	}
	if target.IsHusky {
		fmt.Println("(Detected Husky — hook was written to .husky/pre-commit.)")
	}
	return nil
}

func runUninstall(args []string) error {
	fs := flag.NewFlagSet("uninstall", flag.ExitOnError)
	if err := fs.Parse(args); err != nil {
		return err
	}

	if err := requireGitRepo(); err != nil {
		return err
	}

	status, err := hook.Uninstall("")
	if err != nil {
		return err
	}
	target, err := hook.ResolveTarget("")
	if err != nil {
		return err
	}

	switch status {
	case hook.NotInstalled:
		fmt.Println(colors.Yellow(fmt.Sprintf("secretcheck was not installed at %s.", target.HookPath)))
	case hook.DeletedFile:
		fmt.Println(colors.Green(fmt.Sprintf("✔ Removed empty hook file %s", target.HookPath)))
	case hook.Removed:
		fmt.Println(colors.Green(fmt.Sprintf("✔ Removed secretcheck block from %s", target.HookPath)))
	}
	return nil
}

type jsonOutput struct {
	Findings     []scanner.Finding `json:"findings"`
	FilesScanned int               `json:"filesScanned"`
}

func runScan(args []string) error {
	fs := flag.NewFlagSet("scan", flag.ExitOnError)
	all := fs.Bool("all", false, "Scan all git-tracked files instead of just staged ones")
	jsonOut := fs.Bool("json", false, "Print findings as JSON and skip the interactive prompt")
	noPrompt := fs.Bool("no-prompt", false, "Never prompt interactively; exit non-zero on findings")
	if err := fs.Parse(args); err != nil {
		return err
	}

	if err := requireGitRepo(); err != nil {
		return err
	}

	repoRoot, err := gitutil.RepoRoot("")
	if err != nil {
		return err
	}

	cfg, err := config.Load(repoRoot)
	if err != nil {
		return err
	}
	ruleSet, err := config.ResolveRules(cfg)
	if err != nil {
		return err
	}

	ignoreFilePatterns, err := ignore.LoadIgnoreFile(repoRoot)
	if err != nil {
		return err
	}
	extraExcludes := append(append([]string{}, cfg.IgnorePaths...), ignoreFilePatterns...)

	var result scanner.Result
	if *all {
		result, err = scanner.ScanAll(repoRoot, ruleSet, extraExcludes)
	} else {
		result, err = scanner.ScanStaged(repoRoot, ruleSet, extraExcludes)
	}
	if err != nil {
		return err
	}

	if len(result.Findings) == 0 {
		if *jsonOut {
			out, _ := json.MarshalIndent(jsonOutput{Findings: []scanner.Finding{}, FilesScanned: result.FilesScanned}, "", "  ")
			fmt.Println(string(out))
		} else {
			fmt.Println(colors.Green(fmt.Sprintf("✔ No secrets detected (%d file(s) scanned).", result.FilesScanned)))
		}
		return nil
	}

	if *jsonOut {
		out, _ := json.MarshalIndent(jsonOutput{Findings: result.Findings, FilesScanned: result.FilesScanned}, "", "  ")
		fmt.Println(string(out))
		os.Exit(1)
	}

	report.PrintFindings(result.Findings)

	if v := os.Getenv("SECRETCHECK_ALLOW"); v == "1" || v == "true" {
		fmt.Println(colors.Yellow(colors.Bold("⚠ SECRETCHECK_ALLOW is set — committing despite detected secrets.")))
		return nil
	}

	if !*noPrompt && prompt.IsInteractive() {
		proceed := prompt.Confirm(colors.Yellow(colors.Bold("Secrets were detected above. Commit anyway?")), false)
		if proceed {
			fmt.Println(colors.Yellow("⚠ Proceeding with commit despite detected secrets."))
			return nil
		}
		fmt.Fprintln(os.Stderr, colors.Red("Commit blocked."))
		os.Exit(1)
	}

	report.PrintBypassInstructions()
	os.Exit(1)
	return nil
}
