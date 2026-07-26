// Package report formats scan findings for terminal output.
package report

import (
	"fmt"
	"os"

	"github.com/anukool23/secretcheck/internal/colors"
	"github.com/anukool23/secretcheck/internal/scanner"
)

// PrintFindings writes a human-readable, grouped-by-file report of
// findings to stderr.
func PrintFindings(findings []scanner.Finding) {
	order := make([]string, 0)
	byFile := make(map[string][]scanner.Finding)
	for _, f := range findings {
		if _, ok := byFile[f.File]; !ok {
			order = append(order, f.File)
		}
		byFile[f.File] = append(byFile[f.File], f)
	}

	fmt.Fprintln(os.Stderr)
	fmt.Fprintln(os.Stderr, colors.Red(colors.Bold(fmt.Sprintf("✖ secretcheck found %d potential secret(s):", len(findings)))))
	fmt.Fprintln(os.Stderr)

	for _, file := range order {
		fmt.Fprintln(os.Stderr, colors.Bold(file))
		for _, item := range byFile[file] {
			fmt.Fprintf(os.Stderr, "  %s  %s  %s\n",
				colors.Dim(fmt.Sprintf("line %d", item.Line)),
				colors.Yellow(item.Description),
				colors.Cyan(item.Match),
			)
			fmt.Fprintf(os.Stderr, "    %s\n", colors.Dim(item.Context))
		}
		fmt.Fprintln(os.Stderr)
	}
}

// PrintBypassInstructions explains how to proceed when running
// non-interactively (or when the user declines the prompt).
func PrintBypassInstructions() {
	fmt.Fprintln(os.Stderr, `To commit anyway:
  - Re-run and answer "y" at the prompt, or
  - Run: SECRETCHECK_ALLOW=1 git commit ...  (skips the check for this commit), or
  - Run: git commit --no-verify  (skips all git hooks)

If this is a false positive, add an inline comment on the line:
  // secretcheck-disable-line
or exclude the path via .secretcheckignore / .secretcheckrc.json`)
}
