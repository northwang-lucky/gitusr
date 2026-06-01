package integration

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// ---------------------------------------------------------------------------
// TestHookApplyRCByEmail — apply-rc matches user by email and sets git config
// ---------------------------------------------------------------------------
func TestHookApplyRCByEmail(t *testing.T) {
	homeDir := t.TempDir()
	xdgDataHome := t.TempDir()

	env := map[string]string{
		"HOME":          homeDir,
		"XDG_DATA_HOME": xdgDataHome,
	}

	// Initialize store with users
	storePath := gitusrStore(xdgDataHome)
	users := []map[string]string{
		{"name": "Alice", "email": "alice@example.com"},
		{"name": "Bob", "email": "bob@example.com"},
	}
	writeJSONFile(t, storePath, users)

	// Create a git repo
	repoDir := t.TempDir()
	runGit(t, repoDir, nil, "init")

	// Create .gitusrrc with email
	rcContent := `{"email": "bob@example.com"}`
	rcPath := filepath.Join(repoDir, ".gitusrrc")
	if err := os.WriteFile(rcPath, []byte(rcContent), 0644); err != nil {
		t.Fatalf("write .gitusrrc: %v", err)
	}

	// Run hooks apply-rc
	output, err := runGitusrInDir(t, env, repoDir, "hooks", "apply-rc")
	if err != nil {
		t.Fatalf("hooks apply-rc failed: %v\noutput: %s", err, output)
	}

	// Verify git config was updated
	if name := gitConfig(t, repoDir, "name"); name != "Bob" {
		t.Errorf("git config user.name = %q, want 'Bob'", name)
	}
	if email := gitConfig(t, repoDir, "email"); email != "bob@example.com" {
		t.Errorf("git config user.email = %q, want 'bob@example.com'", email)
	}
}

// ---------------------------------------------------------------------------
// TestHookApplyRCByName — apply-rc falls back to name match when email doesn't match
// ---------------------------------------------------------------------------
func TestHookApplyRCByName(t *testing.T) {
	homeDir := t.TempDir()
	xdgDataHome := t.TempDir()

	env := map[string]string{
		"HOME":          homeDir,
		"XDG_DATA_HOME": xdgDataHome,
	}

	storePath := gitusrStore(xdgDataHome)
	users := []map[string]string{
		{"name": "Alice", "email": "alice@example.com"},
		{"name": "Bob", "email": "bob@example.com"},
	}
	writeJSONFile(t, storePath, users)

	repoDir := t.TempDir()
	runGit(t, repoDir, nil, "init")

	// Create .gitusrrc with name only (email doesn't match any user)
	rcContent := `{"name": "Alice", "email": "nonexistent@example.com"}`
	rcPath := filepath.Join(repoDir, ".gitusrrc")
	if err := os.WriteFile(rcPath, []byte(rcContent), 0644); err != nil {
		t.Fatalf("write .gitusrrc: %v", err)
	}

	output, err := runGitusrInDir(t, env, repoDir, "hooks", "apply-rc")
	if err != nil {
		t.Fatalf("hooks apply-rc failed: %v\noutput: %s", err, output)
	}

	// Should match by name fallback
	if name := gitConfig(t, repoDir, "name"); name != "Alice" {
		t.Errorf("git config user.name = %q, want 'Alice'", name)
	}
	if email := gitConfig(t, repoDir, "email"); email != "alice@example.com" {
		t.Errorf("git config user.email = %q, want 'alice@example.com'", email)
	}
}

// ---------------------------------------------------------------------------
// TestHookApplyRCNotFound — apply-rc with no matching user returns error
// ---------------------------------------------------------------------------
func TestHookApplyRCNotFound(t *testing.T) {
	homeDir := t.TempDir()
	xdgDataHome := t.TempDir()

	env := map[string]string{
		"HOME":          homeDir,
		"XDG_DATA_HOME": xdgDataHome,
	}

	storePath := gitusrStore(xdgDataHome)
	users := []map[string]string{
		{"name": "Alice", "email": "alice@example.com"},
	}
	writeJSONFile(t, storePath, users)

	repoDir := t.TempDir()
	runGit(t, repoDir, nil, "init")

	// Create .gitusrrc with non-matching user
	rcContent := `{"name": "Unknown", "email": "unknown@example.com"}`
	rcPath := filepath.Join(repoDir, ".gitusrrc")
	if err := os.WriteFile(rcPath, []byte(rcContent), 0644); err != nil {
		t.Fatalf("write .gitusrrc: %v", err)
	}

	output, err := runGitusrInDir(t, env, repoDir, "hooks", "apply-rc")
	if err == nil {
		t.Fatalf("hooks apply-rc with no match should fail, got: %s", output)
	}
	if !strings.Contains(strings.ToLower(output), "not found") {
		t.Errorf("output should mention not found, got: %s", output)
	}
}

// ---------------------------------------------------------------------------
// TestHookApplyRCNoFile — apply-rc with no .gitusrrc exits silently
// ---------------------------------------------------------------------------
func TestHookApplyRCNoFile(t *testing.T) {
	homeDir := t.TempDir()
	xdgDataHome := t.TempDir()

	env := map[string]string{
		"HOME":          homeDir,
		"XDG_DATA_HOME": xdgDataHome,
	}

	storePath := gitusrStore(xdgDataHome)
	users := []map[string]string{
		{"name": "Alice", "email": "alice@example.com"},
	}
	writeJSONFile(t, storePath, users)

	repoDir := t.TempDir()
	runGit(t, repoDir, nil, "init")

	// Do NOT create .gitusrrc

	output, err := runGitusrInDir(t, env, repoDir, "hooks", "apply-rc")
	if err != nil {
		t.Fatalf("hooks apply-rc without .gitusrrc should succeed silently, got error: %v\noutput: %s", err, output)
	}
}

// ---------------------------------------------------------------------------
// TestHookApplyRCSilentIfUnchanged — --silent-if-unchanged skips when config already matches
// ---------------------------------------------------------------------------
func TestHookApplyRCSilentIfUnchanged(t *testing.T) {
	homeDir := t.TempDir()
	xdgDataHome := t.TempDir()

	env := map[string]string{
		"HOME":          homeDir,
		"XDG_DATA_HOME": xdgDataHome,
	}

	storePath := gitusrStore(xdgDataHome)
	users := []map[string]string{
		{"name": "Alice", "email": "alice@example.com"},
	}
	writeJSONFile(t, storePath, users)

	repoDir := t.TempDir()
	runGit(t, repoDir, nil, "init")

	// Pre-set git config to match Alice
	runGit(t, repoDir, nil, "config", "user.name", "Alice")
	runGit(t, repoDir, nil, "config", "user.email", "alice@example.com")

	// Create .gitusrrc for Alice
	rcContent := `{"email": "alice@example.com"}`
	rcPath := filepath.Join(repoDir, ".gitusrrc")
	if err := os.WriteFile(rcPath, []byte(rcContent), 0644); err != nil {
		t.Fatalf("write .gitusrrc: %v", err)
	}

	// Run with --silent-if-unchanged — should exit without error and without changing config
	output, err := runGitusrInDir(t, env, repoDir, "hooks", "apply-rc", "--silent-if-unchanged")
	if err != nil {
		t.Fatalf("hooks apply-rc --silent-if-unchanged failed: %v\noutput: %s", err, output)
	}

	// Config should still be Alice
	if name := gitConfig(t, repoDir, "name"); name != "Alice" {
		t.Errorf("git config user.name = %q, want 'Alice'", name)
	}
	if email := gitConfig(t, repoDir, "email"); email != "alice@example.com" {
		t.Errorf("git config user.email = %q, want 'alice@example.com'", email)
	}
}
