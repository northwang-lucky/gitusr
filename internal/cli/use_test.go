package cli

import (
	"strings"
	"testing"

	"gitusr/internal/domain"
	"gitusr/internal/gitcmd"
	"gitusr/internal/i18n"
)

// --- Existing tests (updated with i18n init) ---

// TestUseByIndexInRepo switches a user by index inside a git repo and
// verifies the local git config is updated.
func TestUseByIndexInRepo(t *testing.T) {
	i18n.ResetForTesting()
	i18n.InitWithLocale("en")

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
	if !strings.Contains(stdout, "Success") {
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
	i18n.ResetForTesting()
	i18n.InitWithLocale("en")

	// Isolate HOME to prevent polluting ~/.gitconfig
	t.Setenv("HOME", t.TempDir())

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
	if !strings.Contains(stdout, "Success") {
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
	i18n.ResetForTesting()
	i18n.InitWithLocale("en")

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
	if !strings.Contains(errMsg, "git") {
		t.Errorf("error should mention git, got: %q", errMsg)
	}

	_ = stderr
}

// TestUseNotInitialized ensures an error is returned when the store is not
// initialized.
func TestUseNotInitialized(t *testing.T) {
	i18n.ResetForTesting()
	i18n.InitWithLocale("en")

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
	i18n.ResetForTesting()
	i18n.InitWithLocale("en")

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

// --- New i18n tests ---

// TestUse_Success_En verifies the success flow with English locale.
func TestUse_Success_En(t *testing.T) {
	i18n.ResetForTesting()
	i18n.InitWithLocale("en")

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
	stdout, _, err := executeCmd(cmd, "--index", "0")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(stdout, "Success!") {
		t.Errorf("expected 'Success!' in stdout, got %q", stdout)
	}

	if !strings.Contains(stdout, "Your repo git user is:") {
		t.Errorf("expected 'Your repo git user is:' in stdout, got %q", stdout)
	}
}

// TestUse_Success_ZhCN verifies the success flow with Chinese locale.
func TestUse_Success_ZhCN(t *testing.T) {
	i18n.ResetForTesting()
	i18n.InitWithLocale("zh-CN")

	dir := t.TempDir()
	initGitRepo(t, dir)
	t.Chdir(dir)

	store := &mockStore{
		initialized: true,
		users: []domain.User{
			{Name: "张三", Email: "zhangsan@example.com"},
		},
	}

	cmd := NewUseCmd(store)
	stdout, _, err := executeCmd(cmd, "--index", "0")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(stdout, "成功！") {
		t.Errorf("expected '成功！' in stdout, got %q", stdout)
	}

	if !strings.Contains(stdout, "您的") {
		t.Errorf("expected '您的' in stdout, got %q", stdout)
	}

	if !strings.Contains(stdout, "张三") {
		t.Errorf("expected '张三' in stdout, got %q", stdout)
	}
}

// TestUse_NotRepo_ZhCN verifies the not-a-repo error message in Chinese.
func TestUse_NotRepo_ZhCN(t *testing.T) {
	i18n.ResetForTesting()
	i18n.InitWithLocale("zh-CN")

	dir := t.TempDir()
	t.Chdir(dir)

	store := &mockStore{
		initialized: true,
		users: []domain.User{
			{Name: "Alice", Email: "alice@example.com"},
		},
	}

	cmd := NewUseCmd(store)
	_, _, err := executeCmd(cmd, "--index", "0")

	if err == nil {
		t.Fatal("expected error outside git repo, got nil")
	}

	errMsg := err.Error()
	if !strings.Contains(errMsg, "不是") {
		t.Errorf("expected '不是' in error, got: %q", errMsg)
	}
	if !strings.Contains(errMsg, "git") {
		t.Errorf("expected 'git' in error, got: %q", errMsg)
	}
}

// TestUse_Help_ZhCN verifies the help output in Chinese locale.
func TestUse_Help_ZhCN(t *testing.T) {
	i18n.ResetForTesting()
	i18n.InitWithLocale("zh-CN")

	store := &mockStore{}
	cmd := NewUseCmd(store)

	stdout, stderr, err := executeCmd(cmd, "--help")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if stderr != "" {
		t.Errorf("expected empty stderr, got: %q", stderr)
	}

	if !strings.Contains(stdout, "在 git 仓库中或全局切换用户") {
		t.Errorf("expected Chinese Short description, got: %q", stdout)
	}

	if !strings.Contains(stdout, "切换全局用户") {
		t.Errorf("expected '切换全局用户' flag desc, got: %q", stdout)
	}

	if !strings.Contains(stdout, "按名称切换") {
		t.Errorf("expected '按名称切换' flag desc, got: %q", stdout)
	}

	if !strings.Contains(stdout, "按邮箱切换") {
		t.Errorf("expected '按邮箱切换' flag desc, got: %q", stdout)
	}

	if !strings.Contains(stdout, "按索引切换") {
		t.Errorf("expected '按索引切换' flag desc, got: %q", stdout)
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
