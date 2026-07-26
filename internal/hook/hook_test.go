package hook

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func initTestRepo(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	run := func(args ...string) {
		cmd := exec.Command("git", args...)
		cmd.Dir = dir
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("git %v failed: %v\n%s", args, err, out)
		}
	}
	run("init", "-q")
	run("config", "user.email", "test@example.com")
	run("config", "user.name", "Test")
	return dir
}

func TestInstallCreatesHook(t *testing.T) {
	dir := initTestRepo(t)

	status, err := Install(dir, false)
	if err != nil {
		t.Fatalf("Install failed: %v", err)
	}
	if status != Created {
		t.Errorf("expected Created, got %s", status)
	}

	target, err := ResolveTarget(dir)
	if err != nil {
		t.Fatalf("ResolveTarget failed: %v", err)
	}
	data, err := os.ReadFile(target.HookPath)
	if err != nil {
		t.Fatalf("expected hook file to exist: %v", err)
	}
	if !strings.Contains(string(data), startMarker) {
		t.Error("expected hook file to contain the secretcheck marker")
	}
}

func TestInstallIsIdempotent(t *testing.T) {
	dir := initTestRepo(t)

	if _, err := Install(dir, false); err != nil {
		t.Fatalf("first install failed: %v", err)
	}
	status, err := Install(dir, false)
	if err != nil {
		t.Fatalf("second install failed: %v", err)
	}
	if status != AlreadyInstalled {
		t.Errorf("expected AlreadyInstalled, got %s", status)
	}
}

func TestUninstallRemovesEmptyHookFile(t *testing.T) {
	dir := initTestRepo(t)

	if _, err := Install(dir, false); err != nil {
		t.Fatalf("install failed: %v", err)
	}
	status, err := Uninstall(dir)
	if err != nil {
		t.Fatalf("uninstall failed: %v", err)
	}
	if status != DeletedFile {
		t.Errorf("expected DeletedFile, got %s", status)
	}

	target, err := ResolveTarget(dir)
	if err != nil {
		t.Fatalf("ResolveTarget failed: %v", err)
	}
	if _, err := os.Stat(target.HookPath); !os.IsNotExist(err) {
		t.Error("expected hook file to be removed")
	}
}

func TestInstallPreservesExistingHookContent(t *testing.T) {
	dir := initTestRepo(t)

	target, err := ResolveTarget(dir)
	if err != nil {
		t.Fatalf("ResolveTarget failed: %v", err)
	}
	if err := os.MkdirAll(filepath.Dir(target.HookPath), 0o755); err != nil {
		t.Fatalf("mkdir failed: %v", err)
	}
	existing := "#!/bin/sh\necho custom-hook\n"
	if err := os.WriteFile(target.HookPath, []byte(existing), 0o755); err != nil {
		t.Fatalf("writing existing hook failed: %v", err)
	}

	status, err := Install(dir, false)
	if err != nil {
		t.Fatalf("install failed: %v", err)
	}
	if status != Updated {
		t.Errorf("expected Updated, got %s", status)
	}

	data, err := os.ReadFile(target.HookPath)
	if err != nil {
		t.Fatalf("reading hook failed: %v", err)
	}
	content := string(data)
	if !strings.Contains(content, "echo custom-hook") {
		t.Error("expected existing hook content to be preserved")
	}
	if !strings.Contains(content, startMarker) {
		t.Error("expected secretcheck block to be added")
	}
}

func TestHuskyDetection(t *testing.T) {
	dir := initTestRepo(t)
	if err := os.MkdirAll(filepath.Join(dir, ".husky"), 0o755); err != nil {
		t.Fatalf("mkdir .husky failed: %v", err)
	}

	target, err := ResolveTarget(dir)
	if err != nil {
		t.Fatalf("ResolveTarget failed: %v", err)
	}
	if !target.IsHusky {
		t.Error("expected Husky to be detected")
	}
	if target.HookPath != filepath.Join(dir, ".husky", "pre-commit") {
		t.Errorf("unexpected hook path: %s", target.HookPath)
	}
}
