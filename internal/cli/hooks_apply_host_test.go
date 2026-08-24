package cli

import (
	"os/exec"
	"strings"
	"testing"

	"github.com/northwang-lucky/gitusr/internal/domain"
	"github.com/northwang-lucky/gitusr/internal/i18n"
)

// setGitConfig writes a repo-local git config value for test setup.
func setGitConfig(t *testing.T, dir, key, value string) {
	t.Helper()
	cmd := exec.Command("git", "config", key, value)
	cmd.Dir = dir
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git config %s %s failed: %v\n%s", key, value, err, string(out))
	}
}

// TestHooksApplyHost_MatchApplied verifies that a matching host rule selects
// the saved user and applies it to the repo-local git config with output.
func TestHooksApplyHost_MatchApplied(t *testing.T) {
	i18n.ResetForTesting()
	i18n.InitWithLocale("en")

	dir := t.TempDir()
	initGitRepo(t, dir)
	t.Chdir(dir)

	hostStore := newTestHostStore(t)
	if err := hostStore.SaveAll([]domain.HostRule{{Host: "github.com", Email: "alice@example.com"}}); err != nil {
		t.Fatalf("save hosts: %v", err)
	}
	store := &mockStore{
		initialized: true,
		users:       []domain.User{{Name: "Alice", Email: "alice@example.com"}},
	}

	cmd := NewHooksApplyHostCmd(store, hostStore)
	stdout, stderr, err := executeCmd(cmd, "https://github.com/a/b.git")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if stderr != "" {
		t.Errorf("expected empty stderr, got %q", stderr)
	}
	if !strings.Contains(stdout, "Alice") || !strings.Contains(stdout, "alice@example.com") {
		t.Errorf("expected user info in stdout, got %q", stdout)
	}

	if name := getGitConfig(t, dir, "user.name", false); name != "Alice" {
		t.Errorf("user.name = %q, want %q", name, "Alice")
	}
	if email := getGitConfig(t, dir, "user.email", false); email != "alice@example.com" {
		t.Errorf("user.email = %q, want %q", email, "alice@example.com")
	}
}

// TestHooksApplyHost_NoMatch verifies that unmatched URLs are silently skipped.
func TestHooksApplyHost_NoMatch(t *testing.T) {
	i18n.ResetForTesting()
	i18n.InitWithLocale("en")

	dir := t.TempDir()
	initGitRepo(t, dir)
	t.Chdir(dir)

	hostStore := newTestHostStore(t)
	store := &mockStore{initialized: true, users: []domain.User{{Name: "Alice", Email: "alice@example.com"}}}

	cmd := NewHooksApplyHostCmd(store, hostStore)
	stdout, stderr, err := executeCmd(cmd, "https://example.com/a/b.git")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if stdout != "" || stderr != "" {
		t.Errorf("expected no output, got stdout=%q stderr=%q", stdout, stderr)
	}
}

// TestHooksApplyHost_MissingUserWarns verifies that a rule referencing a
// deleted user emits one warning line and is skipped without an error.
func TestHooksApplyHost_MissingUserWarns(t *testing.T) {
	i18n.ResetForTesting()
	i18n.InitWithLocale("en")

	dir := t.TempDir()
	initGitRepo(t, dir)
	t.Chdir(dir)

	hostStore := newTestHostStore(t)
	if err := hostStore.SaveAll([]domain.HostRule{{Host: "github.com", Email: "deleted@example.com"}}); err != nil {
		t.Fatalf("save hosts: %v", err)
	}
	store := &mockStore{initialized: true, users: []domain.User{{Name: "Alice", Email: "alice@example.com"}}}

	cmd := NewHooksApplyHostCmd(store, hostStore)
	stdout, stderr, err := executeCmd(cmd, "https://github.com/a/b.git")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if stdout != "" {
		t.Errorf("expected empty stdout, got %q", stdout)
	}
	if !strings.Contains(stderr, "deleted@example.com") {
		t.Errorf("expected warning about missing user in stderr, got %q", stderr)
	}
}

// TestHooksApplyHost_SilentUnchanged verifies that --silent-if-unchanged
// suppresses output when the repo config already matches the rule.
func TestHooksApplyHost_SilentUnchanged(t *testing.T) {
	i18n.ResetForTesting()
	i18n.InitWithLocale("en")

	dir := t.TempDir()
	initGitRepo(t, dir)
	t.Chdir(dir)
	setGitConfig(t, dir, "user.name", "Alice")
	setGitConfig(t, dir, "user.email", "alice@example.com")

	hostStore := newTestHostStore(t)
	if err := hostStore.SaveAll([]domain.HostRule{{Host: "github.com", Email: "alice@example.com"}}); err != nil {
		t.Fatalf("save hosts: %v", err)
	}
	store := &mockStore{
		initialized: true,
		users:       []domain.User{{Name: "Alice", Email: "alice@example.com"}},
	}

	cmd := NewHooksApplyHostCmd(store, hostStore)
	stdout, stderr, err := executeCmd(cmd, "https://github.com/a/b.git", "--silent-if-unchanged")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if stdout != "" || stderr != "" {
		t.Errorf("expected no output when unchanged, got stdout=%q stderr=%q", stdout, stderr)
	}
}
