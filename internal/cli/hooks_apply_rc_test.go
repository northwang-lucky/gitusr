package cli

import (
	"strings"
	"testing"

	"github.com/northwang-lucky/gitusr/internal/domain"
	"github.com/northwang-lucky/gitusr/internal/i18n"
)

// TestHooksApplyRC_NoRCFile verifies that when no .gitusrrc file exists in the
// current directory, the command succeeds silently (no output, no error).
func TestHooksApplyRC_NoRCFile(t *testing.T) {
	i18n.ResetForTesting()
	i18n.InitWithLocale("en")

	dir := t.TempDir()
	t.Chdir(dir)

	store := &mockStore{
		initialized: true,
		users:       []domain.User{},
	}

	cmd := NewHooksApplyRCCmd(store)
	stdout, stderr, err := executeCmd(cmd)

	if err != nil {
		t.Fatalf("unexpected error when no .gitusrrc: %v", err)
	}
	if stdout != "" {
		t.Errorf("expected empty stdout, got %q", stdout)
	}
	if stderr != "" {
		t.Errorf("expected empty stderr, got %q", stderr)
	}
}

// TestHooksApplyRC_UserFound_Applied verifies that when a valid .gitusrrc
// references an existing user, the git config is applied locally.
func TestHooksApplyRC_UserFound_Applied(t *testing.T) {
	i18n.ResetForTesting()
	i18n.InitWithLocale("en")

	dir := t.TempDir()
	initGitRepo(t, dir)
	t.Chdir(dir)
	writeRCFile(t, dir, `{"name":"Alice","email":"alice@example.com"}`)

	store := &mockStore{
		initialized: true,
		users: []domain.User{
			{Name: "Alice", Email: "alice@example.com"},
		},
	}

	cmd := NewHooksApplyRCCmd(store)
	stdout, stderr, err := executeCmd(cmd)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// MatchAndApplyRC does not produce output on success
	if stdout != "" {
		t.Errorf("expected empty stdout, got %q", stdout)
	}
	if stderr != "" {
		t.Errorf("expected empty stderr, got %q", stderr)
	}

	// Verify git config was applied
	name := getGitConfig(t, dir, "user.name", false)
	if name != "Alice" {
		t.Errorf("user.name = %q, want %q", name, "Alice")
	}
	email := getGitConfig(t, dir, "user.email", false)
	if email != "alice@example.com" {
		t.Errorf("user.email = %q, want %q", email, "alice@example.com")
	}
}

// TestHooksApplyRC_UserNotFound verifies that when the .gitusrrc references a
// user not in the store, an error is returned with the appropriate i18n message.
func TestHooksApplyRC_UserNotFound(t *testing.T) {
	i18n.ResetForTesting()
	i18n.InitWithLocale("en")

	dir := t.TempDir()
	initGitRepo(t, dir)
	t.Chdir(dir)
	writeRCFile(t, dir, `{"email":"unknown@example.com"}`)

	store := &mockStore{
		initialized: true,
		users: []domain.User{
			{Name: "Alice", Email: "alice@example.com"},
		},
	}

	cmd := NewHooksApplyRCCmd(store)
	_, _, err := executeCmd(cmd)

	if err == nil {
		t.Fatal("expected error when user not found")
	}

	errMsg := err.Error()
	if !strings.Contains(errMsg, "User specified in .gitusrrc not found") {
		t.Errorf("error should contain i18n message, got: %q", errMsg)
	}
}

// TestHooksApplyRC_SilentIfUnchanged_SameUser verifies that with --silent-if-unchanged
// when the current git config already matches the target user, the command
// returns nil without producing any output.
func TestHooksApplyRC_SilentIfUnchanged_SameUser(t *testing.T) {
	i18n.ResetForTesting()
	i18n.InitWithLocale("en")

	dir := t.TempDir()
	initGitRepo(t, dir)
	t.Chdir(dir)
	writeRCFile(t, dir, `{"email":"alice@example.com"}`)

	// Pre-configure git to match the target user
	runGit(t, dir, "config", "user.name", "Alice")
	runGit(t, dir, "config", "user.email", "alice@example.com")

	store := &mockStore{
		initialized: true,
		users: []domain.User{
			{Name: "Alice", Email: "alice@example.com"},
		},
	}

	cmd := NewHooksApplyRCCmd(store)
	stdout, stderr, err := executeCmd(cmd, "-s")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if stdout != "" {
		t.Errorf("expected empty stdout when user unchanged, got %q", stdout)
	}
	if stderr != "" {
		t.Errorf("expected empty stderr when user unchanged, got %q", stderr)
	}
}

// TestHooksApplyRC_SilentIfUnchanged_DifferentUser verifies that with
// --silent-if-unchanged when the current git config differs from the target,
// the command applies the new config normally.
func TestHooksApplyRC_SilentIfUnchanged_DifferentUser(t *testing.T) {
	i18n.ResetForTesting()
	i18n.InitWithLocale("en")

	dir := t.TempDir()
	initGitRepo(t, dir)
	t.Chdir(dir)
	writeRCFile(t, dir, `{"email":"alice@example.com"}`)

	store := &mockStore{
		initialized: true,
		users: []domain.User{
			{Name: "Alice", Email: "alice@example.com"},
		},
	}

	cmd := NewHooksApplyRCCmd(store)
	stdout, stderr, err := executeCmd(cmd, "-s")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if stdout != "" {
		t.Errorf("expected empty stdout, got %q", stdout)
	}
	if stderr != "" {
		t.Errorf("expected empty stderr, got %q", stderr)
	}

	// Git config should have been updated
	name := getGitConfig(t, dir, "user.name", false)
	if name != "Alice" {
		t.Errorf("user.name = %q, want %q", name, "Alice")
	}
	email := getGitConfig(t, dir, "user.email", false)
	if email != "alice@example.com" {
		t.Errorf("user.email = %q, want %q", email, "alice@example.com")
	}
}

// TestHooksApplyRC_NoSilent_SameUser verifies that without --silent-if-unchanged,
// the command applies git config even when it already matches.
func TestHooksApplyRC_NoSilent_SameUser(t *testing.T) {
	i18n.ResetForTesting()
	i18n.InitWithLocale("en")

	dir := t.TempDir()
	initGitRepo(t, dir)
	t.Chdir(dir)
	writeRCFile(t, dir, `{"email":"alice@example.com"}`)

	// Pre-configure to match
	runGit(t, dir, "config", "user.name", "Alice")
	runGit(t, dir, "config", "user.email", "alice@example.com")

	store := &mockStore{
		initialized: true,
		users: []domain.User{
			{Name: "Alice", Email: "alice@example.com"},
		},
	}

	cmd := NewHooksApplyRCCmd(store)
	_, _, err := executeCmd(cmd)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Config should remain as-is (re-applied)
	name := getGitConfig(t, dir, "user.name", false)
	if name != "Alice" {
		t.Errorf("user.name = %q, want %q", name, "Alice")
	}
}

// TestHooksApplyRC_InvalidJSON verifies that an invalid .gitusrrc file produces
// a parse error.
func TestHooksApplyRC_InvalidJSON(t *testing.T) {
	i18n.ResetForTesting()
	i18n.InitWithLocale("en")

	dir := t.TempDir()
	initGitRepo(t, dir)
	t.Chdir(dir)
	writeRCFile(t, dir, `not valid json`)

	store := &mockStore{initialized: true}

	cmd := NewHooksApplyRCCmd(store)
	_, _, err := executeCmd(cmd)

	if err == nil {
		t.Fatal("expected error for invalid JSON")
	}
	if !strings.Contains(err.Error(), "invalid .gitusrrc JSON") {
		t.Errorf("error should contain 'invalid .gitusrrc JSON', got: %q", err.Error())
	}
}

// TestHooksApplyRC_IsHidden verifies that the command is marked as hidden.
func TestHooksApplyRC_IsHidden(t *testing.T) {
	store := &mockStore{}
	cmd := NewHooksApplyRCCmd(store)

	if !cmd.Hidden {
		t.Error("NewHooksApplyRCCmd() should have Hidden = true")
	}
}

// TestHooksApplyRC_FlagRegistration verifies that the --silent-if-unchanged
// flag is registered with the correct shorthand (-s).
func TestHooksApplyRC_FlagRegistration(t *testing.T) {
	store := &mockStore{}
	cmd := NewHooksApplyRCCmd(store)

	flag := cmd.Flags().Lookup("silent-if-unchanged")
	if flag == nil {
		t.Fatal("NewHooksApplyRCCmd() should have --silent-if-unchanged flag")
	}
	if flag.Shorthand != "s" {
		t.Errorf("--silent-if-unchanged shorthand = %q, want %q", flag.Shorthand, "s")
	}
	if flag.DefValue != "false" {
		t.Errorf("--silent-if-unchanged default = %q, want %q", flag.DefValue, "false")
	}
}

// TestHooksApplyRC_EmailPriority verifies that when both name and email are
// specified, email is used for matching (matching MatchAndApplyRC behavior).
func TestHooksApplyRC_EmailPriority(t *testing.T) {
	i18n.ResetForTesting()
	i18n.InitWithLocale("en")

	dir := t.TempDir()
	initGitRepo(t, dir)
	t.Chdir(dir)
	// Name matches user A, email matches user B — email should win
	writeRCFile(t, dir, `{"name":"Alice","email":"bob@example.com"}`)

	store := &mockStore{
		initialized: true,
		users: []domain.User{
			{Name: "Alice", Email: "alice@example.com"},
			{Name: "Bob", Email: "bob@example.com"},
		},
	}

	cmd := NewHooksApplyRCCmd(store)
	_, _, err := executeCmd(cmd)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should have applied Bob's config (email priority)
	name := getGitConfig(t, dir, "user.name", false)
	if name != "Bob" {
		t.Errorf("user.name = %q, want %q (email should take priority)", name, "Bob")
	}
	email := getGitConfig(t, dir, "user.email", false)
	if email != "bob@example.com" {
		t.Errorf("user.email = %q, want %q (email should take priority)", email, "bob@example.com")
	}
}

// --- Localized error message tests ---

// TestHooksApplyRC_UserNotFound_En verifies the English error message.
func TestHooksApplyRC_UserNotFound_En(t *testing.T) {
	i18n.ResetForTesting()
	i18n.InitWithLocale("en")

	dir := t.TempDir()
	initGitRepo(t, dir)
	t.Chdir(dir)
	writeRCFile(t, dir, `{"email":"nobody@example.com"}`)

	store := &mockStore{
		initialized: true,
		users: []domain.User{
			{Name: "Alice", Email: "alice@example.com"},
		},
	}

	cmd := NewHooksApplyRCCmd(store)
	_, _, err := executeCmd(cmd)

	if err == nil {
		t.Fatal("expected error when user not found")
	}

	want := "User specified in .gitusrrc not found in saved users"
	if err.Error() != want {
		t.Errorf("error = %q, want %q", err.Error(), want)
	}
}

// TestHooksApplyRC_UserNotFound_ZhCN verifies the Chinese error message.
func TestHooksApplyRC_UserNotFound_ZhCN(t *testing.T) {
	i18n.ResetForTesting()
	i18n.InitWithLocale("zh-CN")

	dir := t.TempDir()
	initGitRepo(t, dir)
	t.Chdir(dir)
	writeRCFile(t, dir, `{"email":"nobody@example.com"}`)

	store := &mockStore{
		initialized: true,
		users: []domain.User{
			{Name: "Alice", Email: "alice@example.com"},
		},
	}

	cmd := NewHooksApplyRCCmd(store)
	_, _, err := executeCmd(cmd)

	if err == nil {
		t.Fatal("expected error when user not found")
	}

	want := ".gitusrrc 中指定的用户在已保存用户中未找到"
	if err.Error() != want {
		t.Errorf("error = %q, want %q", err.Error(), want)
	}
}

// --- Silent-if-unchanged with getConfigFn override (unit-level) ---

// TestHooksApplyRC_SilentIfUnchanged_SameUser_Mock verifies the silent path
// using getConfigFn override without a real git repo.
func TestHooksApplyRC_SilentIfUnchanged_SameUser_Mock(t *testing.T) {
	i18n.ResetForTesting()
	i18n.InitWithLocale("en")

	dir := t.TempDir()
	t.Chdir(dir)
	writeRCFile(t, dir, `{"email":"alice@example.com"}`)

	// Override getConfigFn to simulate already-matching config
	origGet := getConfigFn
	t.Cleanup(func() { getConfigFn = origGet })
	getConfigFn = func(key string, global bool) (string, error) {
		switch key {
		case "name":
			return "Alice", nil
		case "email":
			return "alice@example.com", nil
		}
		return "", nil
	}

	store := &mockStore{
		initialized: true,
		users: []domain.User{
			{Name: "Alice", Email: "alice@example.com"},
		},
	}

	cmd := NewHooksApplyRCCmd(store)
	stdout, stderr, err := executeCmd(cmd, "-s")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if stdout != "" {
		t.Errorf("expected empty stdout when unchanged, got %q", stdout)
	}
	if stderr != "" {
		t.Errorf("expected empty stderr when unchanged, got %q", stderr)
	}
}

// TestHooksApplyRC_SilentIfUnchanged_DifferentUser_Mock verifies the silent path
// when current config differs, using getConfigFn override.
func TestHooksApplyRC_SilentIfUnchanged_DifferentUser_Mock(t *testing.T) {
	i18n.ResetForTesting()
	i18n.InitWithLocale("en")

	dir := t.TempDir()
	initGitRepo(t, dir)
	t.Chdir(dir)
	writeRCFile(t, dir, `{"email":"alice@example.com"}`)

	// Override getConfigFn to simulate different current config
	origGet := getConfigFn
	t.Cleanup(func() { getConfigFn = origGet })
	getConfigFn = func(key string, global bool) (string, error) {
		switch key {
		case "name":
			return "Bob", nil
		case "email":
			return "bob@example.com", nil
		}
		return "", nil
	}

	store := &mockStore{
		initialized: true,
		users: []domain.User{
			{Name: "Alice", Email: "alice@example.com"},
		},
	}

	cmd := NewHooksApplyRCCmd(store)
	_, _, err := executeCmd(cmd, "-s")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Git config should have been updated to Alice
	name := getGitConfig(t, dir, "user.name", false)
	if name != "Alice" {
		t.Errorf("user.name = %q, want %q", name, "Alice")
	}
	email := getGitConfig(t, dir, "user.email", false)
	if email != "alice@example.com" {
		t.Errorf("user.email = %q, want %q", email, "alice@example.com")
	}
}

// TestHooksApplyRC_SilentIfUnchanged_NotSet verifies that without -s, the
// command still applies config even when getConfigFn indicates the current
// config already matches. MatchAndApplyRC calls gitcmd.SetConfig directly,
// so we need a real git repo to verify the operation succeeds.
func TestHooksApplyRC_SilentIfUnchanged_NotSet(t *testing.T) {
	i18n.ResetForTesting()
	i18n.InitWithLocale("en")

	dir := t.TempDir()
	initGitRepo(t, dir)
	t.Chdir(dir)
	writeRCFile(t, dir, `{"email":"alice@example.com"}`)

	// Override getConfigFn to simulate already-matching config so that
	// the --silent-if-unchanged check would short-circuit if -s were given.
	// Without -s, the command should still proceed to MatchAndApplyRC.
	origGet := getConfigFn
	t.Cleanup(func() { getConfigFn = origGet })
	getConfigFn = func(key string, global bool) (string, error) {
		switch key {
		case "name":
			return "Alice", nil
		case "email":
			return "alice@example.com", nil
		}
		return "", nil
	}

	store := &mockStore{
		initialized: true,
		users: []domain.User{
			{Name: "Alice", Email: "alice@example.com"},
		},
	}

	cmd := NewHooksApplyRCCmd(store)
	_, _, err := executeCmd(cmd) // no -s flag

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify git config was actually applied by MatchAndApplyRC
	name := getGitConfig(t, dir, "user.name", false)
	if name != "Alice" {
		t.Errorf("user.name = %q, want %q", name, "Alice")
	}
}
