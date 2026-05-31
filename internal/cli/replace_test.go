package cli

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"gitusr/internal/domain"
	"gitusr/internal/i18n"
	sel "gitusr/internal/select"
)

// initI18nForTest resets and initializes i18n with the given locale.
// It must be called from every test that checks user-facing messages.
func initI18nForTest(locale string) {
	i18n.ResetForTesting()
	i18n.InitWithLocale(locale)
}

// TestReplace_NotInRepo checks that running replace outside a git repo returns an error.
func TestReplace_NotInRepo(t *testing.T) {
	initI18nForTest("en")

	dir := t.TempDir()
	t.Chdir(dir)

	store := &mockStore{initialized: true, users: []domain.User{
		{Name: "Alice", Email: "alice@example.com"},
	}}

	cmd := NewReplaceCmd(store)
	_, stderr, err := executeCmd(cmd, "old@example.com", "--with-name", "Alice")

	if err == nil {
		t.Fatal("expected error when not in a git repo, got nil")
	}

	errMsg := err.Error()
	if !strings.Contains(errMsg, "not a git repository") {
		t.Errorf("error message should mention 'not a git repository', got: %q", errMsg)
	}
	_ = stderr
}

// TestReplace_UncommittedChanges checks that uncommitted changes block the command.
func TestReplace_UncommittedChanges(t *testing.T) {
	initI18nForTest("en")

	dir := t.TempDir()
	initGitRepo(t, dir)
	t.Chdir(dir)

	// Create an untracked file to trigger HasUncommittedChanges
	if err := os.WriteFile(filepath.Join(dir, "untracked.txt"), []byte("dirty"), 0644); err != nil {
		t.Fatal(err)
	}

	store := &mockStore{initialized: true, users: []domain.User{
		{Name: "Alice", Email: "alice@example.com"},
	}}

	cmd := NewReplaceCmd(store)
	_, _, err := executeCmd(cmd, "old@example.com", "--with-name", "Alice")

	if err == nil {
		t.Fatal("expected error for uncommitted changes, got nil")
	}

	if !strings.Contains(err.Error(), "uncommitted") {
		t.Errorf("error should mention uncommitted changes, got: %q", err.Error())
	}
}

// TestReplace_UserNotFound checks error when no user matches the filter.
func TestReplace_UserNotFound(t *testing.T) {
	dir := t.TempDir()
	initGitRepo(t, dir)
	t.Chdir(dir)

	store := &mockStore{initialized: true, users: []domain.User{
		{Name: "Alice", Email: "alice@example.com"},
	}}

	cmd := NewReplaceCmd(store)
	_, _, err := executeCmd(cmd, "old@example.com", "--with-name", "Bob")

	if err == nil {
		t.Fatal("expected error for non-existent user, got nil")
	}

	if !strings.Contains(err.Error(), "not found") {
		t.Errorf("error should mention 'not found', got: %q", err.Error())
	}
}

// TestReplace_FilterRepoNotInstalled checks that the error from git-filter-repo
// not being installed is propagated correctly.
func TestReplace_FilterRepoNotInstalled(t *testing.T) {
	dir := t.TempDir()
	initGitRepo(t, dir)
	t.Chdir(dir)

	store := &mockStore{initialized: true, users: []domain.User{
		{Name: "Alice", Email: "alice@example.com"},
	}}

	// Mock filterRepoFunc to simulate git-filter-repo not installed.
	origFilterRepo := filterRepoFunc
	defer func() { filterRepoFunc = origFilterRepo }()

	filterRepoFunc = func(oldEmail, newName, newEmail string) error {
		return fmt.Errorf("git-filter-repo is not installed. Install it with: pip install git-filter-repo")
	}

	cmd := NewReplaceCmd(store)
	_, _, err := executeCmd(cmd, "old@example.com", "--with-name", "Alice")

	if err == nil {
		t.Fatal("expected error when filter-repo is not installed, got nil")
	}

	errMsg := err.Error()
	if !strings.Contains(errMsg, "git-filter-repo is not installed") {
		t.Errorf("error should mention git-filter-repo not installed, got: %q", errMsg)
	}
	if !strings.Contains(errMsg, "pip install git-filter-repo") {
		t.Errorf("error should include install instructions, got: %q", errMsg)
	}
}

// TestReplace_BackupBranchCreated checks that a backup branch is created
// before filter-repo runs.
func TestReplace_BackupBranchCreated(t *testing.T) {
	initI18nForTest("en")

	dir := t.TempDir()
	initGitRepo(t, dir)
	t.Chdir(dir)

	store := &mockStore{initialized: true, users: []domain.User{
		{Name: "Alice", Email: "alice@example.com"},
	}}

	// Mock filterRepoFunc and confirmFunc to succeed without side effects.
	origFilterRepo := filterRepoFunc
	origConfirm := confirmFunc
	defer func() {
		filterRepoFunc = origFilterRepo
		confirmFunc = origConfirm
	}()

	filterRepoFunc = func(oldEmail, newName, newEmail string) error {
		return nil
	}
	confirmFunc = func(msg string, defaultVal bool) (bool, error) {
		return false, nil
	}

	cmd := NewReplaceCmd(store)
	stdout, _, err := executeCmd(cmd, "old@example.com", "--with-name", "Alice")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(stdout, "Created backup branch: backup/pre-replace-") {
		t.Errorf("output should mention created backup branch, got: %q", stdout)
	}

	// Verify a backup branch actually exists.
	gitOut, err := exec.Command("git", "branch").Output()
	if err != nil {
		t.Fatalf("git branch failed: %v", err)
	}
	if !strings.Contains(string(gitOut), "backup/pre-replace-") {
		t.Errorf("no backup/pre-replace-* branch found in:\n%s", string(gitOut))
	}
}

// TestReplace_Success tests the full happy path: backup branch created,
// filter-repo mock succeeds, and (when confirmed) repo user config is updated.
func TestReplace_Success(t *testing.T) {
	initI18nForTest("en")

	dir := t.TempDir()
	initGitRepo(t, dir)
	t.Chdir(dir)

	store := &mockStore{initialized: true, users: []domain.User{
		{Name: "Alice", Email: "alice@example.com"},
	}}

	origFilterRepo := filterRepoFunc
	origConfirm := confirmFunc
	defer func() {
		filterRepoFunc = origFilterRepo
		confirmFunc = origConfirm
	}()

	filterRepoCalled := false
	filterRepoFunc = func(oldEmail, newName, newEmail string) error {
		filterRepoCalled = true
		if oldEmail != "old@example.com" {
			t.Errorf("filterRepoFunc oldEmail = %q, want %q", oldEmail, "old@example.com")
		}
		if newName != "Alice" {
			t.Errorf("filterRepoFunc newName = %q, want %q", newName, "Alice")
		}
		if newEmail != "alice@example.com" {
			t.Errorf("filterRepoFunc newEmail = %q, want %q", newEmail, "alice@example.com")
		}
		return nil
	}

	confirmCalled := false
	confirmFunc = func(msg string, defaultVal bool) (bool, error) {
		confirmCalled = true
		if !strings.Contains(msg, "Switch repo user to Alice") {
			t.Errorf("confirm message should mention Alice, got: %q", msg)
		}
		return true, nil
	}

	cmd := NewReplaceCmd(store)
	stdout, _, err := executeCmd(cmd, "old@example.com", "--with-name", "Alice")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify backup branch
	if !strings.Contains(stdout, "Created backup branch: backup/pre-replace-") {
		t.Errorf("output should mention created backup branch, got: %q", stdout)
	}

	// Verify filterRepoFunc was called
	if !filterRepoCalled {
		t.Error("filterRepoFunc was not called")
	}

	// Verify confirm was called
	if !confirmCalled {
		t.Error("confirmFunc was not called")
	}

	// Verify repo user config was updated
	gitCmd := exec.Command("git", "config", "--get", "user.name")
	out, err := gitCmd.Output()
	if err != nil {
		t.Fatalf("git config --get user.name failed: %v", err)
	}
	if strings.TrimSpace(string(out)) != "Alice" {
		t.Errorf("user.name = %q, want %q", strings.TrimSpace(string(out)), "Alice")
	}

	gitCmd = exec.Command("git", "config", "--get", "user.email")
	out, err = gitCmd.Output()
	if err != nil {
		t.Fatalf("git config --get user.email failed: %v", err)
	}
	if strings.TrimSpace(string(out)) != "alice@example.com" {
		t.Errorf("user.email = %q, want %q", strings.TrimSpace(string(out)), "alice@example.com")
	}
}

// TestReplace_ResolveByIndex checks that --with-index flag correctly resolves the user.
func TestReplace_ResolveByIndex(t *testing.T) {
	dir := t.TempDir()
	initGitRepo(t, dir)
	t.Chdir(dir)

	store := &mockStore{initialized: true, users: []domain.User{
		{Name: "Alice", Email: "alice@example.com"},
		{Name: "Bob", Email: "bob@example.com"},
	}}

	origFilterRepo := filterRepoFunc
	origConfirm := confirmFunc
	defer func() {
		filterRepoFunc = origFilterRepo
		confirmFunc = origConfirm
	}()

	filterRepoFunc = func(oldEmail, newName, newEmail string) error {
		if newName != "Bob" {
			t.Errorf("expected Bob (index 1), got %s", newName)
		}
		return nil
	}
	confirmFunc = func(msg string, defaultVal bool) (bool, error) {
		return false, nil
	}

	cmd := NewReplaceCmd(store)
	_, _, err := executeCmd(cmd, "old@example.com", "--with-index", "1")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

// TestReplace_ResolveInteractive tests fallback to interactive selection
// when no filter flags are provided.
func TestReplace_ResolveInteractive(t *testing.T) {
	dir := t.TempDir()
	initGitRepo(t, dir)
	t.Chdir(dir)

	store := &mockStore{initialized: true, users: []domain.User{
		{Name: "Alice", Email: "alice@example.com"},
		{Name: "Bob", Email: "bob@example.com"},
	}}

	origFilterRepo := filterRepoFunc
	origConfirm := confirmFunc
	origSelectFunc := sel.SelectFunc
	defer func() {
		filterRepoFunc = origFilterRepo
		confirmFunc = origConfirm
		sel.SelectFunc = origSelectFunc
	}()

	filterRepoFunc = func(oldEmail, newName, newEmail string) error {
		return nil
	}
	confirmFunc = func(msg string, defaultVal bool) (bool, error) {
		return false, nil
	}

	sel.SelectFunc = func(users []domain.User) (int, error) {
		return 1, nil // select Bob
	}

	cmd := NewReplaceCmd(store)
	_, _, err := executeCmd(cmd, "old@example.com")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

// TestReplace_FlagRegistration checks that all expected flags are registered.
func TestReplace_FlagRegistration(t *testing.T) {
	store := &mockStore{initialized: true}
	cmd := NewReplaceCmd(store)

	flags := []string{"with-name", "with-email", "with-index"}
	for _, name := range flags {
		if cmd.Flags().Lookup(name) == nil {
			t.Errorf("NewReplaceCmd() should have --%s flag", name)
		}
	}

	// Check with-index default value
	withIndexFlag := cmd.Flags().Lookup("with-index")
	if withIndexFlag.DefValue != "-1" {
		t.Errorf("--with-index default = %q, want %q", withIndexFlag.DefValue, "-1")
	}
}

// TestReplace_ArgsValidation checks that the command requires exactly one argument.
func TestReplace_ArgsValidation(t *testing.T) {
	store := &mockStore{initialized: true}
	cmd := NewReplaceCmd(store)

	// No args → error
	_, _, err := executeCmd(cmd)
	if err == nil {
		t.Error("expected error when no args provided")
	}

	// Two args → error
	cmd2 := NewReplaceCmd(store)
	_, _, err = executeCmd(cmd2, "a@b.com", "extra")
	if err == nil {
		t.Error("expected error when extra args provided")
	}
}

// --- i18n tests ---

// TestReplace_Success_En verifies all user-facing English strings on the success path.
func TestReplace_Success_En(t *testing.T) {
	initI18nForTest("en")

	dir := t.TempDir()
	initGitRepo(t, dir)
	t.Chdir(dir)

	store := &mockStore{initialized: true, users: []domain.User{
		{Name: "Alice", Email: "alice@example.com"},
	}}

	origFilterRepo := filterRepoFunc
	origConfirm := confirmFunc
	defer func() {
		filterRepoFunc = origFilterRepo
		confirmFunc = origConfirm
	}()

	filterRepoFunc = func(oldEmail, newName, newEmail string) error { return nil }

	confirmCalled := false
	confirmFunc = func(msg string, defaultVal bool) (bool, error) {
		confirmCalled = true
		if !strings.Contains(msg, "Switch repo user to Alice") {
			t.Errorf("confirm message should contain English text, got: %q", msg)
		}
		return false, nil
	}

	cmd := NewReplaceCmd(store)
	stdout, _, err := executeCmd(cmd, "old@example.com", "--with-name", "Alice")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(stdout, "Created backup branch: backup/pre-replace-") {
		t.Errorf("expected 'Created backup branch:', got: %q", stdout)
	}

	if !confirmCalled {
		t.Error("confirmFunc was not called")
	}
}

// TestReplace_Success_ZhCN verifies all user-facing Chinese strings on the success path.
func TestReplace_Success_ZhCN(t *testing.T) {
	initI18nForTest("zh-CN")

	dir := t.TempDir()
	initGitRepo(t, dir)
	t.Chdir(dir)

	store := &mockStore{initialized: true, users: []domain.User{
		{Name: "Alice", Email: "alice@example.com"},
	}}

	origFilterRepo := filterRepoFunc
	origConfirm := confirmFunc
	defer func() {
		filterRepoFunc = origFilterRepo
		confirmFunc = origConfirm
	}()

	filterRepoFunc = func(oldEmail, newName, newEmail string) error { return nil }

	confirmCalled := false
	confirmFunc = func(msg string, defaultVal bool) (bool, error) {
		confirmCalled = true
		if !strings.Contains(msg, "将仓库用户切换为") {
			t.Errorf("confirm message should contain Chinese text '将仓库用户切换为', got: %q", msg)
		}
		return false, nil
	}

	cmd := NewReplaceCmd(store)
	stdout, _, err := executeCmd(cmd, "old@example.com", "--with-name", "Alice")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(stdout, "已创建备份分支：backup/pre-replace-") {
		t.Errorf("expected '已创建备份分支：', got: %q", stdout)
	}

	if !confirmCalled {
		t.Error("confirmFunc was not called")
	}
}

// TestReplace_Uncommitted_ZhCN verifies the uncommitted-changes error message in Chinese.
func TestReplace_Uncommitted_ZhCN(t *testing.T) {
	initI18nForTest("zh-CN")

	dir := t.TempDir()
	initGitRepo(t, dir)
	t.Chdir(dir)

	if err := os.WriteFile(filepath.Join(dir, "untracked.txt"), []byte("dirty"), 0644); err != nil {
		t.Fatal(err)
	}

	store := &mockStore{initialized: true, users: []domain.User{
		{Name: "Alice", Email: "alice@example.com"},
	}}

	cmd := NewReplaceCmd(store)
	_, _, err := executeCmd(cmd, "old@example.com", "--with-name", "Alice")

	if err == nil {
		t.Fatal("expected error for uncommitted changes, got nil")
	}

	if !strings.Contains(err.Error(), "未提交的更改") {
		t.Errorf("error should mention '未提交的更改', got: %q", err.Error())
	}
}

// TestReplace_Confirm_ZhCN verifies that both the confirm prompt and backup branch
// message are in Chinese when the locale is zh-CN, and that accepting the confirm
// actually updates the git config.
func TestReplace_Confirm_ZhCN(t *testing.T) {
	initI18nForTest("zh-CN")

	dir := t.TempDir()
	initGitRepo(t, dir)
	t.Chdir(dir)

	store := &mockStore{initialized: true, users: []domain.User{
		{Name: "Alice", Email: "alice@example.com"},
	}}

	origFilterRepo := filterRepoFunc
	origConfirm := confirmFunc
	defer func() {
		filterRepoFunc = origFilterRepo
		confirmFunc = origConfirm
	}()

	filterRepoFunc = func(oldEmail, newName, newEmail string) error { return nil }

	confirmCalled := false
	confirmFunc = func(msg string, defaultVal bool) (bool, error) {
		confirmCalled = true
		if !strings.Contains(msg, "将仓库用户切换为") {
			t.Errorf("confirm prompt should be in Chinese, got: %q", msg)
		}
		if !strings.Contains(msg, "Alice") {
			t.Errorf("confirm prompt should contain user name Alice, got: %q", msg)
		}
		if !strings.Contains(msg, "alice@example.com") {
			t.Errorf("confirm prompt should contain email alice@example.com, got: %q", msg)
		}
		return true, nil
	}

	cmd := NewReplaceCmd(store)
	stdout, _, err := executeCmd(cmd, "old@example.com", "--with-name", "Alice")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(stdout, "已创建备份分支：backup/pre-replace-") {
		t.Errorf("expected Chinese backup branch message, got: %q", stdout)
	}

	if !confirmCalled {
		t.Error("confirmFunc was not called")
	}

	// Config was updated because confirm returned true
	gitCmd := exec.Command("git", "config", "--get", "user.name")
	out, err := gitCmd.Output()
	if err != nil {
		t.Fatalf("git config --get user.name failed: %v", err)
	}
	if strings.TrimSpace(string(out)) != "Alice" {
		t.Errorf("user.name = %q, want %q", strings.TrimSpace(string(out)), "Alice")
	}
}
