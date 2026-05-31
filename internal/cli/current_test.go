package cli

import (
	"bytes"
	"io"
	"os"
	"os/exec"
	"strings"
	"testing"

	"gitusr/internal/i18n"

	"github.com/spf13/cobra"
)

// --- helpers ---

// initGitRepo initializes a new git repository in dir, configures a minimal
// user identity, and creates an initial empty commit.
func initGitRepo(t *testing.T, dir string) {
	t.Helper()

	runGit(t, dir, "init")
	runGit(t, dir, "config", "user.name", "Repo User")
	runGit(t, dir, "config", "user.email", "repo@example.com")
	runGit(t, dir, "commit", "--allow-empty", "-m", "initial")
}

func runGit(t *testing.T, dir string, args ...string) {
	t.Helper()
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git %s failed in %s: %v\n%s", strings.Join(args, " "), dir, err, out)
	}
}

// executeCmd runs a cobra command and captures OS-level stdout/stderr.
// This is needed because format.PrintUserInfo uses fmt.Print (os.Stdout)
// rather than cobra's command output writer.
func executeCmd(cmd *cobra.Command, args ...string) (stdout, stderr string, err error) {
	oldOut := os.Stdout
	oldErr := os.Stderr

	rOut, wOut, _ := os.Pipe()
	rErr, wErr, _ := os.Pipe()
	os.Stdout = wOut
	os.Stderr = wErr

	cmd.SetArgs(args)
	err = cmd.Execute()

	wOut.Close()
	wErr.Close()
	os.Stdout = oldOut
	os.Stderr = oldErr

	var bufOut, bufErr bytes.Buffer
	io.Copy(&bufOut, rOut)
	io.Copy(&bufErr, rErr)

	return bufOut.String(), bufErr.String(), err
}

// --- TestCurrent_InRepo ---

func TestCurrent_InRepo(t *testing.T) {
	i18n.InitWithLocale("en")

	dir := t.TempDir()
	initGitRepo(t, dir)
	t.Chdir(dir)

	cmd := NewCurrentCmd()
	stdout, _, err := executeCmd(cmd)

	if err != nil {
		t.Fatalf("NewCurrentCmd().Execute() unexpected error: %v", err)
	}

	if !strings.Contains(stdout, "Your repo git user is:") {
		t.Errorf("expected 'Your repo git user is:', got %q", stdout)
	}
	if !strings.Contains(stdout, "Repo User") {
		t.Errorf("expected 'Repo User', got %q", stdout)
	}
	if !strings.Contains(stdout, "repo@example.com") {
		t.Errorf("expected 'repo@example.com', got %q", stdout)
	}
}

// --- TestCurrent_Global ---

func TestCurrent_Global(t *testing.T) {
	i18n.InitWithLocale("en")

	dir := t.TempDir()
	initGitRepo(t, dir)
	t.Chdir(dir)

	cmd := NewCurrentCmd()
	stdout, _, err := executeCmd(cmd, "--global")

	if err != nil {
		t.Fatalf("NewCurrentCmd().Execute(--global) unexpected error: %v", err)
	}

	if !strings.Contains(stdout, "Your global git user is:") {
		t.Errorf("expected 'Your global git user is:', got %q", stdout)
	}
}

// --- TestCurrent_GlobalShortFlag ---

func TestCurrent_GlobalShortFlag(t *testing.T) {
	i18n.InitWithLocale("en")

	dir := t.TempDir()
	initGitRepo(t, dir)
	t.Chdir(dir)

	cmd := NewCurrentCmd()
	stdout, _, err := executeCmd(cmd, "-g")

	if err != nil {
		t.Fatalf("NewCurrentCmd().Execute(-g) unexpected error: %v", err)
	}

	if !strings.Contains(stdout, "Your global git user is:") {
		t.Errorf("expected 'Your global git user is:', got %q", stdout)
	}
}

// --- TestCurrent_NotInRepo ---

func TestCurrent_NotInRepo(t *testing.T) {
	i18n.InitWithLocale("en")

	dir := t.TempDir()
	t.Chdir(dir)

	cmd := NewCurrentCmd()
	_, stderr, err := executeCmd(cmd)

	if err == nil {
		t.Fatal("NewCurrentCmd().Execute() outside repo expected error, got nil")
	}

	errMsg := err.Error()
	if !strings.Contains(errMsg, "not a git repository") || !strings.Contains(strings.ToLower(errMsg), "git") {
		t.Errorf("error message should mention git repository, got: %q", errMsg)
	}

	// Stderr should also contain the error
	if stderr == "" && errMsg == "" {
		t.Error("expected error output on stderr or via error return")
	}
}

// --- TestCurrent_GlobalOutsideRepo ---

func TestCurrent_GlobalOutsideRepo(t *testing.T) {
	i18n.InitWithLocale("en")

	dir := t.TempDir()
	t.Chdir(dir)

	cmd := NewCurrentCmd()
	stdout, _, err := executeCmd(cmd, "--global")

	if err != nil {
		t.Fatalf("NewCurrentCmd().Execute(--global) outside repo unexpected error: %v", err)
	}

	if !strings.Contains(stdout, "Your global git user is:") {
		t.Errorf("expected 'Your global git user is:', got %q", stdout)
	}
}

// --- TestCurrent_Alias ---

func TestCurrent_Alias(t *testing.T) {
	cmd := NewCurrentCmd()

	found := false
	for _, alias := range cmd.Aliases {
		if alias == "ct" {
			found = true
			break
		}
	}

	if !found {
		t.Error("NewCurrentCmd() should have alias 'ct'")
	}
}

// --- TestCurrent_FlagRegistration ---

func TestCurrent_FlagRegistration(t *testing.T) {
	cmd := NewCurrentCmd()

	globalFlag := cmd.Flags().Lookup("global")
	if globalFlag == nil {
		t.Fatal("NewCurrentCmd() should have --global flag")
	}
	if globalFlag.Shorthand != "g" {
		t.Errorf("--global flag shorthand = %q, want %q", globalFlag.Shorthand, "g")
	}
}

// --- TestCurrent_En ---

func TestCurrent_En(t *testing.T) {
	i18n.ResetForTesting()
	i18n.InitWithLocale("en")

	dir := t.TempDir()
	initGitRepo(t, dir)
	t.Chdir(dir)

	cmd := NewCurrentCmd()

	// Verify Short is translated
	if cmd.Short != "show current repo/global user" {
		t.Errorf("expected Short='show current repo/global user', got %q", cmd.Short)
	}

	// Verify flag usage is translated
	globalFlag := cmd.Flags().Lookup("global")
	if globalFlag == nil {
		t.Fatal("--global flag not found")
	}
	if globalFlag.Usage != "show global user" {
		t.Errorf("expected flag usage='show global user', got %q", globalFlag.Usage)
	}

	stdout, _, err := executeCmd(cmd)
	if err != nil {
		t.Fatalf("NewCurrentCmd().Execute() unexpected error: %v", err)
	}

	if !strings.Contains(stdout, "Your repo git user is:") {
		t.Errorf("expected 'Your repo git user is:' in stdout, got %q", stdout)
	}
	if !strings.Contains(stdout, "Repo User") {
		t.Errorf("expected 'Repo User' in stdout, got %q", stdout)
	}
	if !strings.Contains(stdout, "repo@example.com") {
		t.Errorf("expected 'repo@example.com' in stdout, got %q", stdout)
	}
}

// --- TestCurrent_ZhCN ---

func TestCurrent_ZhCN(t *testing.T) {
	i18n.ResetForTesting()
	i18n.InitWithLocale("zh-CN")

	dir := t.TempDir()
	initGitRepo(t, dir)
	t.Chdir(dir)

	cmd := NewCurrentCmd()

	// Verify Short is translated to Chinese
	if cmd.Short != "显示当前仓库/全局用户" {
		t.Errorf("expected Short='显示当前仓库/全局用户', got %q", cmd.Short)
	}

	// Verify flag usage is translated to Chinese
	globalFlag := cmd.Flags().Lookup("global")
	if globalFlag == nil {
		t.Fatal("--global flag not found")
	}
	if globalFlag.Usage != "显示全局用户" {
		t.Errorf("expected flag usage='显示全局用户', got %q", globalFlag.Usage)
	}

	stdout, _, err := executeCmd(cmd)
	if err != nil {
		t.Fatalf("NewCurrentCmd().Execute() unexpected error: %v", err)
	}

	if !strings.Contains(stdout, "您的 repo git 用户为：") {
		t.Errorf("expected '您的 repo git 用户为：' in stdout, got %q", stdout)
	}
	if !strings.Contains(stdout, "Repo User") {
		t.Errorf("expected 'Repo User' in stdout, got %q", stdout)
	}
	if !strings.Contains(stdout, "repo@example.com") {
		t.Errorf("expected 'repo@example.com' in stdout, got %q", stdout)
	}
}

// --- TestCurrent_NotRepo_En ---

func TestCurrent_NotRepo_En(t *testing.T) {
	i18n.ResetForTesting()
	i18n.InitWithLocale("en")

	dir := t.TempDir()
	t.Chdir(dir)

	cmd := NewCurrentCmd()
	_, _, err := executeCmd(cmd)

	if err == nil {
		t.Fatal("expected error outside repo, got nil")
	}

	errMsg := err.Error()
	expected := "not a git repository (or any of the parent directories): .git"
	if errMsg != expected {
		t.Errorf("expected error %q, got %q", expected, errMsg)
	}
}

// --- TestCurrent_NotRepo_ZhCN ---

func TestCurrent_NotRepo_ZhCN(t *testing.T) {
	i18n.ResetForTesting()
	i18n.InitWithLocale("zh-CN")

	dir := t.TempDir()
	t.Chdir(dir)

	cmd := NewCurrentCmd()
	_, _, err := executeCmd(cmd)

	if err == nil {
		t.Fatal("expected error outside repo, got nil")
	}

	errMsg := err.Error()
	expected := "不是 git 仓库（或其任何父目录）：.git"
	if errMsg != expected {
		t.Errorf("expected error %q, got %q", expected, errMsg)
	}
}
