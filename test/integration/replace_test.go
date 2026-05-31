package integration

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// ---------------------------------------------------------------------------
// TestReplaceAuthor — replace command rewrites commit history
// ---------------------------------------------------------------------------
func TestReplaceAuthor(t *testing.T) {
	// Ensure git-filter-repo is available
	gitFilterRepo := ensureGitFilterRepo(t)

	homeDir := t.TempDir()
	xdgDataHome := t.TempDir()

	env := map[string]string{
		"HOME":          homeDir,
		"XDG_DATA_HOME": xdgDataHome,
		"PATH":          filepath.Dir(gitFilterRepo) + ":" + os.Getenv("PATH"),
	}

	// Pre-configure store with two users
	storePath := gitusrStore(xdgDataHome)
	users := []map[string]string{
		{"name": "Old Author", "email": "old@example.com"},
		{"name": "New Author", "email": "new@example.com"},
	}
	writeJSONFile(t, storePath, users)

	// Create a temp git repo with a commit by old author
	repoDir := t.TempDir()
	runGit(t, repoDir, nil, "init")
	runGit(t, repoDir, nil, "config", "user.name", "Old Author")
	runGit(t, repoDir, nil, "config", "user.email", "old@example.com")

	testFile := filepath.Join(repoDir, "test.txt")
	if err := os.WriteFile(testFile, []byte("hello world"), 0644); err != nil {
		t.Fatalf("write test file: %v", err)
	}
	runGit(t, repoDir, nil, "add", ".")
	runGit(t, repoDir, nil, "commit", "-m", "initial commit")

	// Verify the commit has the old author
	logOut := runGitOutput(t, repoDir, nil, "log", "--format=%an <%ae>", "-1")
	if !strings.Contains(logOut, "Old Author <old@example.com>") {
		t.Fatalf("setup: expected old author, got: %s", logOut)
	}

	// === Execute replace with --yes ===
	output, err := runGitusrInDir(t, env, repoDir, "replace", "old@example.com", "--with-index", "1", "--yes")
	if err != nil {
		t.Fatalf("replace failed: %v\noutput: %s", err, output)
	}

	// Verify backup branch was created
	branchOut := runGitOutput(t, repoDir, nil, "branch", "-a")
	if !strings.Contains(branchOut, "backup/pre-replace-") {
		t.Errorf("backup branch not created, branches: %s", branchOut)
	}

	// Verify author was replaced in history
	logOut = runGitOutput(t, repoDir, nil, "log", "--format=%an <%ae>", "-1")
	if !strings.Contains(logOut, "New Author <new@example.com>") {
		t.Errorf("author not replaced, got: %s", logOut)
	}

	// Verify git config was updated (because --yes)
	if name := gitConfig(t, repoDir, "name"); name != "New Author" {
		t.Errorf("git config user.name = %q, want 'New Author'", name)
	}
	if email := gitConfig(t, repoDir, "email"); email != "new@example.com" {
		t.Errorf("git config user.email = %q, want 'new@example.com'", email)
	}
}

// ---------------------------------------------------------------------------
// TestReplaceWithFlags — replace using --with-name and --with-email
// ---------------------------------------------------------------------------
func TestReplaceWithFlags(t *testing.T) {
	gitFilterRepo := ensureGitFilterRepo(t)

	homeDir := t.TempDir()
	xdgDataHome := t.TempDir()

	env := map[string]string{
		"HOME":          homeDir,
		"XDG_DATA_HOME": xdgDataHome,
		"PATH":          filepath.Dir(gitFilterRepo) + ":" + os.Getenv("PATH"),
	}

	storePath := gitusrStore(xdgDataHome)
	users := []map[string]string{
		{"name": "Alice", "email": "alice@example.com"},
		{"name": "Bob", "email": "bob@example.com"},
	}
	writeJSONFile(t, storePath, users)

	repoDir := t.TempDir()
	runGit(t, repoDir, nil, "init")
	runGit(t, repoDir, nil, "config", "user.name", "Alice")
	runGit(t, repoDir, nil, "config", "user.email", "alice@example.com")

	testFile := filepath.Join(repoDir, "test.txt")
	os.WriteFile(testFile, []byte("test"), 0644)
	runGit(t, repoDir, nil, "add", ".")
	runGit(t, repoDir, nil, "commit", "-m", "test")

	// Use --with-name and --with-email instead of --with-index
	output, err := runGitusrInDir(t, env, repoDir, "replace", "alice@example.com",
		"--with-name", "Bob", "--with-email", "bob@example.com", "--yes")
	if err != nil {
		t.Fatalf("replace with flags failed: %v\noutput: %s", err, output)
	}

	logOut := runGitOutput(t, repoDir, nil, "log", "--format=%an <%ae>", "-1")
	if !strings.Contains(logOut, "Bob <bob@example.com>") {
		t.Errorf("author not replaced with flags, got: %s", logOut)
	}
}

// ---------------------------------------------------------------------------
// TestReplaceNotInRepo — replace outside git repo returns error
// ---------------------------------------------------------------------------
func TestReplaceNotInRepo(t *testing.T) {
	homeDir := t.TempDir()
	xdgDataHome := t.TempDir()

	env := map[string]string{
		"HOME":          homeDir,
		"XDG_DATA_HOME": xdgDataHome,
	}

	nonRepoDir := t.TempDir()
	output, err := runGitusrInDir(t, env, nonRepoDir, "replace", "old@example.com", "--with-index", "0", "--yes")
	if err == nil {
		t.Fatalf("replace outside repo should fail, got: %s", output)
	}
}

// ---------------------------------------------------------------------------
// TestReplaceUncommitted — replace with uncommitted changes returns error
// ---------------------------------------------------------------------------
func TestReplaceUncommitted(t *testing.T) {
	homeDir := t.TempDir()
	xdgDataHome := t.TempDir()

	env := map[string]string{
		"HOME":          homeDir,
		"XDG_DATA_HOME": xdgDataHome,
	}

	repoDir := t.TempDir()
	runGit(t, repoDir, nil, "init")
	runGit(t, repoDir, nil, "config", "user.name", "Test")
	runGit(t, repoDir, nil, "config", "user.email", "test@example.com")

	// Create an uncommitted file
	uncommittedFile := filepath.Join(repoDir, "dirty.txt")
	os.WriteFile(uncommittedFile, []byte("dirty"), 0644)

	output, err := runGitusrInDir(t, env, repoDir, "replace", "test@example.com", "--with-index", "0", "--yes")
	if err == nil {
		t.Fatalf("replace with uncommitted changes should fail, got: %s", output)
	}
	if !strings.Contains(strings.ToLower(output), "uncommitted") {
		t.Errorf("output should mention uncommitted changes, got: %s", output)
	}
}

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

// ensureGitFilterRepo checks for git-filter-repo in PATH or downloads it.
// Returns the path to the git-filter-repo executable.
func ensureGitFilterRepo(t *testing.T) string {
	t.Helper()

	// Check if already in PATH
	if path, err := exec.LookPath("git-filter-repo"); err == nil {
		return path
	}

	// Try the temp location from previous download
	tmpPath := "/tmp/git-filter-repo"
	if _, err := os.Stat(tmpPath); err == nil {
		return tmpPath
	}

	// Download it
	t.Log("git-filter-repo not found, downloading...")
	cmd := exec.Command("curl", "-sL",
		"https://raw.githubusercontent.com/newren/git-filter-repo/main/git-filter-repo",
		"-o", tmpPath)
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Skipf("git-filter-repo not available and download failed: %v\n%s", err, string(out))
	}
	if err := os.Chmod(tmpPath, 0755); err != nil {
		t.Skipf("failed to chmod git-filter-repo: %v", err)
	}

	// Verify it works
	verifyCmd := exec.Command(tmpPath, "--help")
	if err := verifyCmd.Run(); err != nil {
		t.Skipf("downloaded git-filter-repo does not work: %v", err)
	}

	return tmpPath
}

// runGitOutput runs a git command and returns the trimmed stdout.
func runGitOutput(t *testing.T, dir string, env map[string]string, args ...string) string {
	t.Helper()

	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	cmd.Env = os.Environ()
	for k, v := range env {
		cmd.Env = append(cmd.Env, k+"="+v)
	}

	out, err := cmd.Output()
	if err != nil {
		stderr := ""
		if exitErr, ok := err.(*exec.ExitError); ok {
			stderr = string(exitErr.Stderr)
		}
		t.Fatalf("git %s failed: %v\nstderr: %s", strings.Join(args, " "), err, stderr)
	}
	return strings.TrimSpace(string(out))
}
