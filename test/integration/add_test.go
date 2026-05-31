package integration

import (
	"strings"
	"testing"
)

// ---------------------------------------------------------------------------
// TestAddNonInteractive — add with --name and --email flags
// ---------------------------------------------------------------------------
func TestAddNonInteractive(t *testing.T) {
	homeDir := t.TempDir()
	xdgDataHome := t.TempDir()

	env := map[string]string{
		"HOME":          homeDir,
		"XDG_DATA_HOME": xdgDataHome,
	}

	// Initialize store with one user
	storePath := gitusrStore(xdgDataHome)
	users := []map[string]string{
		{"name": "Alice", "email": "alice@example.com"},
	}
	writeJSONFile(t, storePath, users)

	// Add a new user non-interactively
	output, err := runGitusr(t, env, "add", "--name", "Bob", "--email", "bob@example.com")
	if err != nil {
		t.Fatalf("add failed: %v\noutput: %s", err, output)
	}
	if !strings.Contains(output, "Bob") {
		t.Errorf("add output should contain 'Bob', got: %s", output)
	}
	if !strings.Contains(output, "bob@example.com") {
		t.Errorf("add output should contain 'bob@example.com', got: %s", output)
	}

	// Verify store now has two users
	var stored []map[string]string
	readJSONFile(t, storePath, &stored)
	if len(stored) != 2 {
		t.Fatalf("expected 2 users in store, got %d", len(stored))
	}
	if stored[1]["name"] != "Bob" || stored[1]["email"] != "bob@example.com" {
		t.Errorf("stored user = %v, want Bob/bob@example.com", stored[1])
	}
}

// ---------------------------------------------------------------------------
// TestAddDuplicateName — add with duplicate name returns error
// ---------------------------------------------------------------------------
func TestAddDuplicateName(t *testing.T) {
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

	output, err := runGitusr(t, env, "add", "--name", "Alice", "--email", "different@example.com")
	if err == nil {
		t.Fatalf("add duplicate name should fail, got: %s", output)
	}
	if !strings.Contains(output, "already exists") && !strings.Contains(output, "already") {
		t.Errorf("output should mention duplicate, got: %s", output)
	}

	// Verify store was NOT modified
	var stored []map[string]string
	readJSONFile(t, storePath, &stored)
	if len(stored) != 1 {
		t.Errorf("store should still have 1 user, got %d", len(stored))
	}
}

// ---------------------------------------------------------------------------
// TestAddDuplicateEmail — add with duplicate email returns error
// ---------------------------------------------------------------------------
func TestAddDuplicateEmail(t *testing.T) {
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

	output, err := runGitusr(t, env, "add", "--name", "Different", "--email", "alice@example.com")
	if err == nil {
		t.Fatalf("add duplicate email should fail, got: %s", output)
	}
	if !strings.Contains(output, "already exists") && !strings.Contains(output, "already") {
		t.Errorf("output should mention duplicate, got: %s", output)
	}
}

// ---------------------------------------------------------------------------
// TestAddInvalidEmail — add with invalid email format returns error
// ---------------------------------------------------------------------------
func TestAddInvalidEmail(t *testing.T) {
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

	output, err := runGitusr(t, env, "add", "--name", "Bob", "--email", "not-an-email")
	if err == nil {
		t.Fatalf("add invalid email should fail, got: %s", output)
	}
	if !strings.Contains(strings.ToLower(output), "invalid") && !strings.Contains(strings.ToLower(output), "email") {
		t.Errorf("output should mention invalid email, got: %s", output)
	}
}

// ---------------------------------------------------------------------------
// TestAddMissingFlags — add with only --name or only --email returns error
// ---------------------------------------------------------------------------
func TestAddMissingFlags(t *testing.T) {
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

	// Only --name
	output, err := runGitusr(t, env, "add", "--name", "Bob")
	if err == nil {
		t.Fatalf("add with only --name should fail, got: %s", output)
	}

	// Only --email
	output, err = runGitusr(t, env, "add", "--email", "bob@example.com")
	if err == nil {
		t.Fatalf("add with only --email should fail, got: %s", output)
	}
}

// ---------------------------------------------------------------------------
// TestAddNotInitialized — add with empty store returns error
// ---------------------------------------------------------------------------
func TestAddNotInitialized(t *testing.T) {
	homeDir := t.TempDir()
	xdgDataHome := t.TempDir()

	env := map[string]string{
		"HOME":          homeDir,
		"XDG_DATA_HOME": xdgDataHome,
	}

	// Do NOT create store
	output, err := runGitusr(t, env, "add", "--name", "Bob", "--email", "bob@example.com")
	if err == nil {
		t.Fatalf("add with uninitialized store should fail, got: %s", output)
	}
	if !strings.Contains(output, "initialized") && !strings.Contains(output, "save") {
		t.Errorf("output should mention store not initialized, got: %s", output)
	}
}
