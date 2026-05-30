package integration

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// gitusrBin is the path to the built gitusr binary.
var gitusrBin string

func TestMain(m *testing.M) {
	moduleRoot := findModuleRoot()

	// Build the binary once for all tests.
	tmpBin := filepath.Join(os.TempDir(), "gitusr-integration-test")
	buildCmd := exec.Command("go", "build", "-o", tmpBin, "./cmd/gitusr")
	buildCmd.Dir = moduleRoot
	if out, err := buildCmd.CombinedOutput(); err != nil {
		fmt.Fprintf(os.Stderr, "failed to build gitusr: %v\n%s", err, string(out))
		os.Exit(1)
	}
	gitusrBin = tmpBin

	code := m.Run()

	os.Remove(tmpBin)
	os.Exit(code)
}

func findModuleRoot() string {
	dir, err := os.Getwd()
	if err != nil {
		return "."
	}
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "."
		}
		dir = parent
	}
}

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

// runGitusr executes the gitusr binary with the given args and env vars.
// Returns combined stdout+stderr.
func runGitusr(t *testing.T, env map[string]string, args ...string) (string, error) {
	t.Helper()

	cmd := exec.Command(gitusrBin, args...)
	cmd.Env = os.Environ()
	for k, v := range env {
		cmd.Env = append(cmd.Env, k+"="+v)
	}

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	output := stdout.String()
	if stderr.Len() > 0 {
		output += "\n" + stderr.String()
	}
	return output, err
}

// runGitusrStdin runs gitusr with piped stdin and returns combined output.
func runGitusrStdin(t *testing.T, env map[string]string, stdin string, args ...string) (string, error) {
	t.Helper()

	cmd := exec.Command(gitusrBin, args...)
	cmd.Env = os.Environ()
	for k, v := range env {
		cmd.Env = append(cmd.Env, k+"="+v)
	}
	cmd.Stdin = strings.NewReader(stdin)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	output := stdout.String()
	if stderr.Len() > 0 {
		output += "\n" + stderr.String()
	}
	return output, err
}

// runGitusrInDir runs gitusr with the given working directory and returns combined output.
func runGitusrInDir(t *testing.T, env map[string]string, dir string, args ...string) (string, error) {
	t.Helper()

	cmd := exec.Command(gitusrBin, args...)
	cmd.Dir = dir
	cmd.Env = os.Environ()
	for k, v := range env {
		cmd.Env = append(cmd.Env, k+"="+v)
	}

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	output := stdout.String()
	if stderr.Len() > 0 {
		output += "\n" + stderr.String()
	}
	return output, err
}

// runGit runs a git command in the specified directory with optional env vars.
func runGit(t *testing.T, dir string, env map[string]string, args ...string) {
	t.Helper()

	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	cmd.Env = os.Environ()
	for k, v := range env {
		cmd.Env = append(cmd.Env, k+"="+v)
	}

	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("git %s failed in %s: %v\n%s", strings.Join(args, " "), dir, err, string(out))
	}
}

// gitConfig reads a git config value from the given repo directory.
func gitConfig(t *testing.T, dir string, key string) string {
	t.Helper()

	cmd := exec.Command("git", "config", "user."+key)
	cmd.Dir = dir
	out, err := cmd.Output()
	if err != nil {
		t.Fatalf("git config user.%s failed in %s: %v", key, dir, err)
	}
	return strings.TrimSpace(string(out))
}

// writeJSONFile writes data as JSON to the given path, creating the parent dir if needed.
func writeJSONFile(t *testing.T, path string, data any) {
	t.Helper()

	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		t.Fatalf("create dir for %s: %v", path, err)
	}

	raw, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		t.Fatalf("marshal json: %v", err)
	}
	raw = append(raw, '\n')

	if err := os.WriteFile(path, raw, 0644); err != nil {
		t.Fatalf("write file %s: %v", path, err)
	}
}

// readJSONFile reads and unmarshals a JSON file from the given path.
func readJSONFile(t *testing.T, path string, v any) {
	t.Helper()

	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read file %s: %v", path, err)
	}

	if err := json.Unmarshal(raw, v); err != nil {
		t.Fatalf("unmarshal %s: %v", path, err)
	}
}

// gitusrStore returns the path to the gitusr store file under the given
// XDG_DATA_HOME directory.
func gitusrStore(xdgDataHome string) string {
	return filepath.Join(xdgDataHome, "gitusr", "user-list.json")
}

// ---------------------------------------------------------------------------
// TestFullWorkflow — e2e: init → add → list → use → current → remove
// ---------------------------------------------------------------------------
func TestFullWorkflow(t *testing.T) {
	// === Setup isolated environment ===
	homeDir := t.TempDir()
	xdgDataHome := t.TempDir()

	env := map[string]string{
		"HOME":           homeDir,
		"XDG_DATA_HOME":  xdgDataHome,
	}

	// Pre-configure git global user so init can discover it without prompts.
	runGit(t, homeDir, env, "config", "--global", "user.name", "Alice")
	runGit(t, homeDir, env, "config", "--global", "user.email", "alice@example.com")

	// === Step 1: Init ===
	output, err := runGitusr(t, env, "init", "--force")
	if err != nil {
		t.Fatalf("init --force failed: %v\noutput: %s", err, output)
	}
	if !strings.Contains(output, "Success") {
		t.Errorf("init: expected 'Success' in output, got: %s", output)
	}
	if !strings.Contains(output, "Alice") {
		t.Errorf("init: expected 'Alice' in output, got: %s", output)
	}

	// Verify store file was created
	storePath := gitusrStore(xdgDataHome)
	var users []map[string]string
	readJSONFile(t, storePath, &users)
	if len(users) != 1 {
		t.Fatalf("init: expected 1 user in store, got %d", len(users))
	}
	if users[0]["name"] != "Alice" || users[0]["email"] != "alice@example.com" {
		t.Errorf("init: store user = %v, want Alice/alice@example.com", users[0])
	}

	// === Step 2: List ===
	output, err = runGitusr(t, env, "list")
	if err != nil {
		t.Fatalf("list failed: %v\noutput: %s", err, output)
	}
	if !strings.Contains(output, "Alice") {
		t.Errorf("list: expected 'Alice' in output, got: %s", output)
	}
	if !strings.Contains(output, "alice@example.com") {
		t.Errorf("list: expected 'alice@example.com' in output, got: %s", output)
	}
	// Should show index "0:"
	if !strings.Contains(output, "0:") {
		t.Errorf("list: expected index '0:' in output, got: %s", output)
	}

	// === Step 3: Add user (direct store manipulation — add is interactive) ===
	// We test add by directly inserting into the store JSON to cover the
	// full workflow end-to-end. The add command itself relies on interactive
	// prompts, which are tested at the unit level.
	users = append(users, map[string]string{"name": "Bob", "email": "bob@example.com"})
	writeJSONFile(t, storePath, users)

	// Verify list shows both users
	output, err = runGitusr(t, env, "list")
	if err != nil {
		t.Fatalf("list after add failed: %v\noutput: %s", err, output)
	}
	if !strings.Contains(output, "Alice") || !strings.Contains(output, "Bob") {
		t.Errorf("list after add: expected both Alice and Bob, got: %s", output)
	}
	if !strings.Contains(output, "0:") || !strings.Contains(output, "1:") {
		t.Errorf("list after add: expected indices 0 and 1, got: %s", output)
	}

	// === Step 4: Use — switch to Bob (index 1) in a temp git repo ===
	repoDir := t.TempDir()
	runGit(t, repoDir, nil, "init")
	runGit(t, repoDir, nil, "config", "user.name", "placeholder")
	runGit(t, repoDir, nil, "config", "user.email", "placeholder@example.com")

	output, err = runGitusrInDir(t, env, repoDir, "use", "--index", "1")
	if err != nil {
		t.Fatalf("use --index 1 failed: %v\noutput: %s", err, output)
	}
	if !strings.Contains(output, "Success") {
		t.Errorf("use: expected 'Success' in output, got: %s", output)
	}
	if !strings.Contains(output, "Bob") {
		t.Errorf("use: expected 'Bob' in output, got: %s", output)
	}

	// Verify local git config was updated
	if name := gitConfig(t, repoDir, "name"); name != "Bob" {
		t.Errorf("use: git config user.name = %q, want 'Bob'", name)
	}
	if email := gitConfig(t, repoDir, "email"); email != "bob@example.com" {
		t.Errorf("use: git config user.email = %q, want 'bob@example.com'", email)
	}

	// === Step 5: Current — verify shows correct user ===
	output, err = runGitusrInDir(t, env, repoDir, "current")
	if err != nil {
		t.Fatalf("current failed: %v\noutput: %s", err, output)
	}
	if !strings.Contains(output, "repo") {
		t.Errorf("current: expected 'repo' in output, got: %s", output)
	}
	if !strings.Contains(output, "Bob") {
		t.Errorf("current: expected 'Bob' in output, got: %s", output)
	}
	if !strings.Contains(output, "bob@example.com") {
		t.Errorf("current: expected 'bob@example.com' in output, got: %s", output)
	}

	// === Step 6: Current --global ===
	output, err = runGitusr(t, env, "current", "--global")
	if err != nil {
		t.Fatalf("current --global failed: %v\noutput: %s", err, output)
	}
	if !strings.Contains(output, "global") {
		t.Errorf("current --global: expected 'global' in output, got: %s", output)
	}

	// === Step 7: Remove Bob by email ===
	output, err = runGitusr(t, env, "remove", "--email", "bob@example.com")
	if err != nil {
		t.Fatalf("remove --email failed: %v\noutput: %s", err, output)
	}
	if !strings.Contains(output, "Success") {
		t.Errorf("remove: expected 'Success' in output, got: %s", output)
	}
	if !strings.Contains(output, "Bob") {
		t.Errorf("remove: expected 'Bob' in output, got: %s", output)
	}

	// === Step 8: List — verify Bob is gone ===
	output, err = runGitusr(t, env, "list")
	if err != nil {
		t.Fatalf("list after remove failed: %v\noutput: %s", err, output)
	}
	if strings.Contains(output, "Bob") || strings.Contains(output, "bob@example.com") {
		t.Errorf("list after remove: Bob should be gone, got: %s", output)
	}
	if !strings.Contains(output, "Alice") {
		t.Errorf("list after remove: Alice should still be present, got: %s", output)
	}
	// Only index 0 should exist (Alice)
	if strings.Contains(output, "1:") {
		t.Errorf("list after remove: index '1:' should not appear, got: %s", output)
	}
}

// ---------------------------------------------------------------------------
// TestUseWithStoreByIndex — use with different filter types
// ---------------------------------------------------------------------------
func TestUseWithStoreByIndex(t *testing.T) {
	homeDir := t.TempDir()
	xdgDataHome := t.TempDir()

	env := map[string]string{
		"HOME":          homeDir,
		"XDG_DATA_HOME": xdgDataHome,
	}

	// Pre-configure store with two users
	storePath := gitusrStore(xdgDataHome)
	users := []map[string]string{
		{"name": "UserA", "email": "a@test.com"},
		{"name": "UserB", "email": "b@test.com"},
	}
	writeJSONFile(t, storePath, users)

	// Create temp git repo
	repoDir := t.TempDir()
	runGit(t, repoDir, nil, "init")

	// Use by email
	output, err := runGitusrInDir(t, env, repoDir, "use", "--email", "b@test.com")
	if err != nil {
		t.Fatalf("use --email failed: %v\noutput: %s", err, output)
	}
	if !strings.Contains(output, "Success") || !strings.Contains(output, "UserB") {
		t.Errorf("use --email: expected 'Success' and 'UserB', got: %s", output)
	}
	if name := gitConfig(t, repoDir, "name"); name != "UserB" {
		t.Errorf("use --email: git config = %q, want 'UserB'", name)
	}
}

// ---------------------------------------------------------------------------
// TestUseGlobalSwitch — use --global flag
// ---------------------------------------------------------------------------
func TestUseGlobalSwitch(t *testing.T) {
	homeDir := t.TempDir()
	xdgDataHome := t.TempDir()

	env := map[string]string{
		"HOME":          homeDir,
		"XDG_DATA_HOME": xdgDataHome,
	}

	// Pre-configure store
	storePath := gitusrStore(xdgDataHome)
	users := []map[string]string{
		{"name": "GlobalUser", "email": "global@test.com"},
	}
	writeJSONFile(t, storePath, users)

	// Run use --global --index 0 (any dir works, but we use a temp dir)
	dir := t.TempDir()
	output, err := runGitusrInDir(t, env, dir, "use", "--global", "--index", "0")
	if err != nil {
		t.Fatalf("use --global --index 0 failed: %v\noutput: %s", err, output)
	}
	if !strings.Contains(output, "Success") || !strings.Contains(output, "GlobalUser") {
		t.Errorf("use --global: expected 'Success' and 'GlobalUser', got: %s", output)
	}

	// Verify global git config was updated
	name := gitConfig(t, homeDir, "name")
	if name != "GlobalUser" {
		t.Errorf("use --global: global git config user.name = %q, want 'GlobalUser'", name)
	}

	// Note: git config --global reads from HOME/.gitconfig
	cmd := exec.Command("git", "config", "--global", "user.email")
	cmd.Dir = dir
	cmd.Env = append(os.Environ(), "HOME="+homeDir)
	out, err := cmd.Output()
	if err != nil {
		t.Fatalf("git config --global user.email: %v", err)
	}
	email := strings.TrimSpace(string(out))
	if email != "global@test.com" {
		t.Errorf("use --global: global git config user.email = %q, want 'global@test.com'", email)
	}
}

// ---------------------------------------------------------------------------
// TestErrors — error scenarios
// ---------------------------------------------------------------------------

// TestCurrentNotInRepo verifies that current in a non-git directory returns an error.
func TestCurrentNotInRepo(t *testing.T) {
	homeDir := t.TempDir()
	xdgDataHome := t.TempDir()

	env := map[string]string{
		"HOME":          homeDir,
		"XDG_DATA_HOME": xdgDataHome,
	}

	// Use a temp dir that is NOT a git repo
	nonRepoDir := t.TempDir()
	output, err := runGitusrInDir(t, env, nonRepoDir, "current")
	if err == nil {
		t.Fatalf("current in non-git repo should have returned an error\ngot output: %s", output)
	}
	if !strings.Contains(err.Error(), "exit status 1") {
		t.Errorf("expected exit status 1 error, got: %v", err)
	}
	if !strings.Contains(strings.ToLower(output), "git") &&
		!strings.Contains(strings.ToLower(output), "repository") {
		t.Errorf("output should mention git/repository, got: %s", output)
	}
}

// TestUseNotInitialized verifies that use with an empty store returns an error.
func TestUseNotInitialized(t *testing.T) {
	homeDir := t.TempDir()
	xdgDataHome := t.TempDir()

	env := map[string]string{
		"HOME":          homeDir,
		"XDG_DATA_HOME": xdgDataHome,
	}

	// Do NOT create a store — it should be empty/uninitialized

	// Create git repo so the "not a git repo" check passes
	repoDir := t.TempDir()
	runGit(t, repoDir, nil, "init")

	output, err := runGitusrInDir(t, env, repoDir, "use", "--index", "0")
	if err == nil {
		t.Fatalf("use with uninitialized store should have returned an error\ngot output: %s", output)
	}
	if !strings.Contains(output, "initialized") && !strings.Contains(output, "save") {
		t.Errorf("output should mention store not initialized, got: %s", output)
	}
}

// TestRemoveNonExistent verifies that removing a non-existent user returns an error.
func TestRemoveNonExistent(t *testing.T) {
	homeDir := t.TempDir()
	xdgDataHome := t.TempDir()

	env := map[string]string{
		"HOME":          homeDir,
		"XDG_DATA_HOME": xdgDataHome,
	}

	// Pre-configure store with one user
	storePath := gitusrStore(xdgDataHome)
	users := []map[string]string{
		{"name": "Charlie", "email": "charlie@example.com"},
	}
	writeJSONFile(t, storePath, users)

	// Try to remove a non-existent user
	output, err := runGitusr(t, env, "remove", "--email", "nonexistent@example.com")
	if err == nil {
		t.Fatalf("remove non-existent user should have returned an error\ngot output: %s", output)
	}
	if !strings.Contains(output, "not found") {
		t.Errorf("output should mention 'not found', got: %s", output)
	}

	// Verify the existing user was NOT removed
	var stored []map[string]string
	readJSONFile(t, storePath, &stored)
	if len(stored) != 1 || stored[0]["name"] != "Charlie" {
		t.Errorf("remove non-existent: store should still have Charlie, got %v", stored)
	}
}

// TestInitWithoutGitConfig verifies that init fails cleanly when there
// is no git global config AND no TTY for prompting.
func TestInitWithoutGitConfig(t *testing.T) {
	homeDir := t.TempDir()
	xdgDataHome := t.TempDir()

	env := map[string]string{
		"HOME":          homeDir,
		"XDG_DATA_HOME": xdgDataHome,
	}

	// Do NOT set up git global config — the init should fail since
	// there's no TTY in test environment for interactive prompts.
	output, err := runGitusr(t, env, "init", "--force")
	if err == nil {
		t.Fatalf("init without git config should have returned an error\ngot output: %s", output)
	}
}

// TestListNotInitialized verifies that list with an empty store returns an error.
func TestListNotInitialized(t *testing.T) {
	homeDir := t.TempDir()
	xdgDataHome := t.TempDir()

	env := map[string]string{
		"HOME":          homeDir,
		"XDG_DATA_HOME": xdgDataHome,
	}

	// Do NOT create a store
	output, err := runGitusr(t, env, "list")
	if err == nil {
		t.Fatalf("list with uninitialized store should have returned an error\ngot output: %s", output)
	}
	if !strings.Contains(output, "initialized") && !strings.Contains(output, "save") {
		t.Errorf("output should mention store not initialized, got: %s", output)
	}
}
