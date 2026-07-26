// Package colors provides a minimal ANSI helper (no dependency) that
// disables itself automatically on non-TTY output or when NO_COLOR is set.
package colors

import "os"

const esc = "\x1b"

func enabled() bool {
	if os.Getenv("NO_COLOR") != "" {
		return false
	}
	info, err := os.Stdout.Stat()
	if err != nil {
		return false
	}
	return (info.Mode() & os.ModeCharDevice) != 0
}

func wrap(code, text string) string {
	if !enabled() {
		return text
	}
	return esc + "[" + code + "m" + text + esc + "[0m"
}

func Red(s string) string    { return wrap("31", s) }
func Green(s string) string  { return wrap("32", s) }
func Yellow(s string) string { return wrap("33", s) }
func Cyan(s string) string   { return wrap("36", s) }
func Bold(s string) string   { return wrap("1", s) }
func Dim(s string) string    { return wrap("2", s) }
