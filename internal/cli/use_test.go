package cli

import (
	"testing"

	"gitusr/internal/domain"
	"gitusr/internal/gitcmd"
)

// TestUseByIndexInRepo switches a user by index inside a git repo and
// verifies the local git config is updated.
func TestUseByIndexInRepo(t *testing.T) {
	dir := t.TempDir()
	initGitRepo(t, dir)
	t.Chdir(dir)

	store := &mockStore{
		initialized: true,
		users: []domain.User{
			{Name: "Alice", Email: "alice@example.com"},
			{Name: "Bob", Email: "bob@example.com"},
		},
	}

	cmd := NewUseCmd(store)
	stdout, _, err := executeCmd(cmd, "--index", "0")

	if err != nil {
		t.Fatalf("NewUseCmd().Execute(--index 0) unexpected error: %v", err)
	}

	// Verify stdout contains success info.
	if !containsAny(stdout, "Success") {
		t.Errorf("expected 'Success' in stdout, got %q", stdout)
	}

	// Verify local git config was updated.
	name, err := gitcmd.GetConfig("name", false)
	if err != nil {
		t.Fatalf("git config user.name: %v", err)
	}
	if name != "Alice" {
		t.Errorf("user.name = %q, want %q", name, "Alice")
	}

	email, err := gitcmd.GetConfig("email", false)
	if err != nil {
		t.Fatalf("git config user.email: %v", err)
	}
	if email != "alice@example.com" {
		t.Errorf("user.email = %q, want %q", email, "alice@example.com")
	}
}

// TestUseGlobal switches the global user and verifies the global git config
// is updated.
func TestUseGlobal(t *testing.T) {
	dir := t.TempDir()
	initGitRepo(t, dir)
	t.Chdir(dir)

	store := &mockStore{
		initialized: true,
		users: []domain.User{
			{Name: "GlobalUser", Email: "global@example.com"},
		},
	}

	cmd := NewUseCmd(store)
	stdout, _, err := executeCmd(cmd, "--global", "--index", "0")

	if err != nil {
		t.Fatalf("NewUseCmd().Execute(--global --index 0) unexpected error: %v", err)
	}

	// Verify stdout contains success info.
	if !containsAny(stdout, "Success") {
		t.Errorf("expected 'Success' in stdout, got %q", stdout)
	}

	// Verify global git config was updated.
	name, err := gitcmd.GetConfig("name", true)
	if err != nil {
		t.Fatalf("git config --global user.name: %v", err)
	}
	if name != "GlobalUser" {
		t.Errorf("global user.name = %q, want %q", name, "GlobalUser")
	}

	email, err := gitcmd.GetConfig("email", true)
	if err != nil {
		t.Fatalf("git config --global user.email: %v", err)
	}
	if email != "global@example.com" {
		t.Errorf("global user.email = %q, want %q", email, "global@example.com")
	}
}

// TestUseNotInRepo ensures an error is returned when running use outside a
// git repo without --global.
func TestUseNotInRepo(t *testing.T) {
	dir := t.TempDir()
	t.Chdir(dir)

	store := &mockStore{
		initialized: true,
		users: []domain.User{
			{Name: "Alice", Email: "alice@example.com"},
		},
	}

	cmd := NewUseCmd(store)
	_, stderr, err := executeCmd(cmd, "--index", "0")

	if err == nil {
		t.Fatal("expected error outside git repo, got nil")
	}

	errMsg := err.Error()
	if !containsAny(errMsg, "git") {
		t.Errorf("error should mention git, got: %q", errMsg)
	}

	_ = stderr
}

// TestUseNotInitialized ensures an error is returned when the store is not
// initialized.
func TestUseNotInitialized(t *testing.T) {
	dir := t.TempDir()
	initGitRepo(t, dir)
	t.Chdir(dir)

	store := &mockStore{
		initialized: false,
		users:       []domain.User{},
	}

	cmd := NewUseCmd(store)
	_, _, err := executeCmd(cmd, "--index", "0")

	if err == nil {
		t.Fatal("expected error when store not initialized, got nil")
	}
}

// TestUseUserNotFound ensures an error is returned when the specified user
// cannot be found.
func TestUseUserNotFound(t *testing.T) {
	dir := t.TempDir()
	initGitRepo(t, dir)
	t.Chdir(dir)

	store := &mockStore{
		initialized: true,
		users: []domain.User{
			{Name: "Alice", Email: "alice@example.com"},
		},
	}

	cmd := NewUseCmd(store)
	_, _, err := executeCmd(cmd, "--email", "notfound@example.com")

	if err == nil {
		t.Fatal("expected error for non-existent user, got nil")
	}
}

// containsAny reports whether s contains any of the given substrings.
func containsAny(s string, subs ...string) bool {
	for _, sub := range subs {
		if len(sub) > 0 && len(s) >= len(sub) {
			for i := 0; i <= len(s)-len(sub); i++ {
				if s[i:i+len(sub)] == sub {
					return true
				}
			}
		}
	}
	return false
}
