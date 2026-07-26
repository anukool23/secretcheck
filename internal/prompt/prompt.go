// Package prompt provides a minimal, stdlib-only interactive yes/no
// confirmation and a TTY/CI detection helper.
package prompt

import (
	"bufio"
	"fmt"
	"os"
	"runtime"
	"strings"
)

// Confirm asks a yes/no question on stderr and reads the answer from the
// controlling terminal. An empty answer resolves to defaultValue.
func Confirm(question string, defaultValue bool) bool {
	suffix := "y/N"
	if defaultValue {
		suffix = "Y/n"
	}
	fmt.Fprintf(os.Stderr, "%s (%s) ", question, suffix)

	in, owned := openTTY()
	if owned {
		defer in.Close()
	}

	reader := bufio.NewReader(in)
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

// openTTY opens the controlling terminal directly rather than relying on
// os.Stdin. When this binary is invoked from a git hook, os.Stdin has
// passed through several layers (git -> shell hook script -> npm/pip
// shim -> this process), and in that chain it can end up reporting as a
// character device (so IsInteractive() says yes) without actually
// blocking on a read the way a direct terminal read does — the prompt
// would print but resolve to the default instantly instead of waiting
// for a keypress. Reading /dev/tty (or CONIN$ on Windows) sidesteps that
// entirely by talking to the terminal directly, the same technique
// ssh/sudo/git itself use for prompts that must work regardless of how
// stdin was piped down to the process.
//
// The returned bool reports whether the caller owns the file (and should
// close it) — false means it fell back to os.Stdin.
func openTTY() (*os.File, bool) {
	name := "/dev/tty"
	if runtime.GOOS == "windows" {
		name = "CONIN$"
	}
	f, err := os.OpenFile(name, os.O_RDONLY, 0)
	if err != nil {
		return os.Stdin, false
	}
	return f, true
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
