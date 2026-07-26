// Package prompt provides a minimal, stdlib-only interactive yes/no
// confirmation and a TTY/CI detection helper.
package prompt

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

// Confirm asks a yes/no question on stderr and reads the answer from
// stdin. An empty answer resolves to defaultValue.
func Confirm(question string, defaultValue bool) bool {
	suffix := "y/N"
	if defaultValue {
		suffix = "Y/n"
	}
	fmt.Fprintf(os.Stderr, "%s (%s) ", question, suffix)

	reader := bufio.NewReader(os.Stdin)
	line, err := reader.ReadString('\n')
	if err != nil && line == "" {
		return defaultValue
	}
	answer := strings.ToLower(strings.TrimSpace(line))
	if answer == "" {
		return defaultValue
	}
	return answer == "y" || answer == "yes"
}

// IsInteractive reports whether we're attached to a real terminal and not
// running in CI, i.e. whether prompting the user makes sense.
func IsInteractive() bool {
	if os.Getenv("CI") != "" {
		return false
	}
	return isCharDevice(os.Stdin) && isCharDevice(os.Stderr)
}

func isCharDevice(f *os.File) bool {
	info, err := f.Stat()
	if err != nil {
		return false
	}
	return (info.Mode() & os.ModeCharDevice) != 0
}
