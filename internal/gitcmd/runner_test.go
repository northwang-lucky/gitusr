package gitcmd

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// --- helpers ---

// initGitRepo initializes a new git repository in dir, configures a minimal
// user identity, and creates an initial empty commit so that the repo has a
// valid HEAD.  The caller is responsible for creating dir (e.g. via t.TempDir).
func initGitRepo(t *testing.T, dir string) {
	t.Helper()

	runGit(t, dir, "init")
	runGit(t, dir, "config", "user.name", "Test User")
	runGit(t, dir, "config", "user.email", "test@example.com")
	// Create an empty commit so the repo has a HEAD
	runGit(t, dir, "commit", "--allow-empty", "-m", "initial")
}

// runGit runs a git command in dir and fails the test if the command fails.
func runGit(t *testing.T, dir string, args ...string) {
	t.Helper()
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git %s failed in %s: %v\n%s", strings.Join(args, " "), dir, err, out)
	}
}

// gitConfigGet is a thin wrapper to read a config value via git for assertions.
func gitConfigGet(t *testing.T, dir, key string) string {
	t.Helper()
	cmd := exec.Command("git", "config", "--get", key)
	cmd.Dir = dir
	out, err := cmd.Output()
	if err != nil {
		t.Fatalf("git config --get %s failed: %v", key, err)
	}
	return strings.TrimSpace(string(out))
}

// isFilterRepoInstalled returns true when git-filter-repo is available.
func isFilterRepoInstalled() bool {
	return exec.Command("git", "filter-repo", "--help").Run() == nil
}

// --- IsGitRepo tests ---

func TestIsGitRepo_InRepo(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("HOME", t.TempDir())
	initGitRepo(t, dir)
	t.Chdir(dir)

	if !IsGitRepo() {
		t.Error("IsGitRepo() = false inside a git repo, want true")
	}
}

func TestIsGitRepo_NotInRepo(t *testing.T) {
	dir := t.TempDir()
	t.Chdir(dir)

	if IsGitRepo() {
		t.Error("IsGitRepo() = true outside a git repo, want false")
	}
}

func TestIsGitRepo_Worktree(t *testing.T) {
	// .git as a file (worktree / submodule pointer) should still return true
	dir := t.TempDir()
	t.Chdir(dir)

	if err := os.WriteFile(".git", []byte("gitdir: /some/path\n"), 0644); err != nil {
		t.Fatal(err)
	}

	if !IsGitRepo() {
		t.Error("IsGitRepo() = false when .git is a file, want true")
	}
}

// --- GetConfig tests ---

func TestGetConfig_Existing(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("HOME", t.TempDir())
	initGitRepo(t, dir)
	t.Chdir(dir)

	got, err := GetConfig("name", false)
	if err != nil {
		t.Fatalf("GetConfig(name) unexpected error: %v", err)
	}
	if got != "Test User" {
		t.Errorf("GetConfig(name) = %q, want %q", got, "Test User")
	}
}

func TestGetConfig_Missing(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("HOME", t.TempDir())
	initGitRepo(t, dir)
	t.Chdir(dir)

	_, err := GetConfig("nonexistent", false)
	if err == nil {
		t.Error("GetConfig(nonexistent) expected error, got nil")
	}
}

// --- SetConfig tests ---

func TestSetConfig(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("HOME", t.TempDir())
	initGitRepo(t, dir)
	t.Chdir(dir)

	if err := SetConfig("email", "new@example.com", false); err != nil {
		t.Fatalf("SetConfig(email) unexpected error: %v", err)
	}

	got := gitConfigGet(t, ".", "user.email")
	if got != "new@example.com" {
		t.Errorf("after SetConfig: user.email = %q, want %q", got, "new@example.com")
	}
}

func TestSetConfig_MultipleKeys(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("HOME", t.TempDir())
	initGitRepo(t, dir)
	t.Chdir(dir)

	if err := SetConfig("name", "Alice", false); err != nil {
		t.Fatalf("SetConfig(name) unexpected error: %v", err)
	}
	if err := SetConfig("email", "alice@example.com", false); err != nil {
		t.Fatalf("SetConfig(email) unexpected error: %v", err)
	}

	if got := gitConfigGet(t, ".", "user.name"); got != "Alice" {
		t.Errorf("user.name = %q, want %q", got, "Alice")
	}
	if got := gitConfigGet(t, ".", "user.email"); got != "alice@example.com" {
		t.Errorf("user.email = %q, want %q", got, "alice@example.com")
	}
}

// --- CreateBackupBranch tests ---

func TestCreateBackupBranch(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("HOME", t.TempDir())
	initGitRepo(t, dir)
	t.Chdir(dir)

	if err := CreateBackupBranch("backup/original"); err != nil {
		t.Fatalf("CreateBackupBranch(backup/original) unexpected error: %v", err)
	}

	// Verify the branch exists
	cmd := exec.Command("git", "branch", "--list", "backup/original")
	out, err := cmd.Output()
	if err != nil {
		t.Fatalf("git branch --list failed: %v", err)
	}
	if !strings.Contains(string(out), "backup/original") {
		t.Error("branch 'backup/original' not found after CreateBackupBranch")
	}
}

func TestCreateBackupBranch_Duplicate(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("HOME", t.TempDir())
	initGitRepo(t, dir)
	t.Chdir(dir)

	if err := CreateBackupBranch("dup"); err != nil {
		t.Fatalf("first CreateBackupBranch(dup) unexpected error: %v", err)
	}

	err := CreateBackupBranch("dup")
	if err == nil {
		t.Error("CreateBackupBranch(dup) second call expected error, got nil")
	}
}

// --- FilterRepo tests ---

func TestFilterRepo_NotInstalled(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("HOME", t.TempDir())
	initGitRepo(t, dir)
	t.Chdir(dir)

	if isFilterRepoInstalled() {
		t.Skip("git-filter-repo is installed; skipping 'not installed' test")
	}

	err := FilterRepo("old@email.com", "New Name", "new@email.com")
	if err == nil {
		t.Fatal("FilterRepo() expected error when git-filter-repo is not installed")
	}

	errMsg := err.Error()
	if !strings.Contains(errMsg, "git-filter-repo is not installed") {
		t.Errorf("error message does not mention installation: %s", errMsg)
	}
	if !strings.Contains(errMsg, "pip install git-filter-repo") {
		t.Errorf("error message does not include install instructions: %s", errMsg)
	}
}

// --- HasUncommittedChanges tests ---

func TestHasUncommittedChanges_Clean(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("HOME", t.TempDir())
	initGitRepo(t, dir)
	t.Chdir(dir)

	if HasUncommittedChanges() {
		t.Error("HasUncommittedChanges() = true on clean repo, want false")
	}
}

func TestHasUncommittedChanges_Modified(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("HOME", t.TempDir())
	initGitRepo(t, dir)
	t.Chdir(dir)

	// Create and commit a tracked file first, then modify it
	if err := os.WriteFile(filepath.Join(dir, "README.md"), []byte("hello"), 0644); err != nil {
		t.Fatal(err)
	}
	runGit(t, dir, "add", "README.md")
	runGit(t, dir, "commit", "-m", "add README")

	// Modify the tracked file
	if err := os.WriteFile(filepath.Join(dir, "README.md"), []byte("modified"), 0644); err != nil {
		t.Fatal(err)
	}

	if !HasUncommittedChanges() {
		t.Error("HasUncommittedChanges() = false with modified tracked file, want true")
	}
}

func TestHasUncommittedChanges_Untracked(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("HOME", t.TempDir())
	initGitRepo(t, dir)
	t.Chdir(dir)

	// Create an untracked file (not staged)
	if err := os.WriteFile(filepath.Join(dir, "newfile.txt"), []byte("untracked"), 0644); err != nil {
		t.Fatal(err)
	}

	if !HasUncommittedChanges() {
		t.Error("HasUncommittedChanges() = false with untracked file, want true")
	}
}

func TestHasUncommittedChanges_NotInRepo(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("HOME", t.TempDir())
	t.Chdir(dir)

	// Outside a git repo, git status --porcelain will fail, so the function
	// should return false (safe default).
	if HasUncommittedChanges() {
		t.Error("HasUncommittedChanges() = true outside a git repo, want false")
	}
}

// --- filterRepoCallback tests (internal helper) ---

func TestFilterRepoCallback(t *testing.T) {
	cb := filterRepoCallback("old@example.com", "Alice", "alice@new.com")

	checks := []string{
		"commit.author_email == b'old@example.com'",
		"commit.author_name = b'Alice'",
		"commit.author_email = b'alice@new.com'",
		"commit.committer_name = b'Alice'",
		"commit.committer_email = b'alice@new.com'",
	}

	for _, want := range checks {
		if !strings.Contains(cb, want) {
			t.Errorf("callback missing %q\ncallback:\n%s", want, cb)
		}
	}
}
