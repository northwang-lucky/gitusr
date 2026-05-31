package cli

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/northwang-lucky/gitusr/internal/domain"
	"github.com/northwang-lucky/gitusr/internal/format"
	"github.com/northwang-lucky/gitusr/internal/i18n"
	"github.com/northwang-lucky/gitusr/internal/store"
)

// captureStdout runs fn and returns captured stdout as a string.
// os.Stdout is restored via t.Cleanup.
func captureStdout(t *testing.T, fn func()) string {
	t.Helper()
	orig := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w
	t.Cleanup(func() { os.Stdout = orig })

	fn()

	w.Close()
	var buf bytes.Buffer
	_, _ = io.Copy(&buf, r)
	return buf.String()
}

// testStore creates a JSONStore backed by a temp file.
func testStore(t *testing.T) *store.JSONStore {
	t.Helper()
	return store.NewJSONStore(filepath.Join(t.TempDir(), "test-users.json"))
}

// preloadedStore creates a store with the given users already saved.
func preloadedStore(t *testing.T, users []domain.User) *store.JSONStore {
	t.Helper()
	s := testStore(t)
	if err := s.SaveAll(users); err != nil {
		t.Fatalf("preloadedStore: SaveAll failed: %v", err)
	}
	return s
}

// useLocale initializes i18n with the given locale for a test.
// It must be called before any T()-translated strings are used.
func useLocale(t *testing.T, locale string) {
	t.Helper()
	i18n.ResetForTesting()
	i18n.InitWithLocale(locale)
	t.Cleanup(func() { i18n.ResetForTesting() })
}

// ---------------------------------------------------------------------------
// TestInit_WithGlobalGitUser
// ---------------------------------------------------------------------------
func TestInit_WithGlobalGitUser(t *testing.T) {
	// Override getConfig to return a valid user
	origGet := getConfigFn
	t.Cleanup(func() { getConfigFn = origGet })
	getConfigFn = func(key string, global bool) (string, error) {
		switch key {
		case "name":
			return "Test User", nil
		case "email":
			return "test@example.com", nil
		}
		return "", fmt.Errorf("unknown key: %s", key)
	}

	// confirm should not be called because store is empty
	origConfirm := confirmFn
	t.Cleanup(func() { confirmFn = origConfirm })
	confirmFnCalled := false
	confirmFn = func(msg string, defaultVal bool) (bool, error) {
		confirmFnCalled = true
		return false, nil
	}

	// Capture printUserInfo
	var printedUser domain.User
	var printedOpts format.PrintOptions
	origPrint := printUserInfoFn
	t.Cleanup(func() { printUserInfoFn = origPrint })
	printUserInfoFn = func(user domain.User, opts format.PrintOptions) {
		printedUser = user
		printedOpts = opts
	}

	s := testStore(t)
	cmd := NewInitCmd(s)
	cmd.SetArgs([]string{})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("init failed: %v", err)
	}

	// Verify store
	users, err := s.List()
	if err != nil {
		t.Fatalf("List() failed: %v", err)
	}
	if len(users) != 1 {
		t.Fatalf("expected 1 user, got %d", len(users))
	}
	if users[0].Name != "Test User" {
		t.Errorf("Name = %q, want %q", users[0].Name, "Test User")
	}
	if users[0].Email != "test@example.com" {
		t.Errorf("Email = %q, want %q", users[0].Email, "test@example.com")
	}

	// confirm should not have been called
	if confirmFnCalled {
		t.Error("confirm was called unexpectedly")
	}

	// printUserInfo should have been called
	if printedUser.Name != "Test User" {
		t.Errorf("printed Name = %q, want %q", printedUser.Name, "Test User")
	}
	if !printedOpts.Global {
		t.Error("expected Global=true in PrintOptions")
	}
	if !printedOpts.ShowSuccess {
		t.Error("expected ShowSuccess=true in PrintOptions")
	}
}

// ---------------------------------------------------------------------------
// TestInit_WithoutGlobalUser
// ---------------------------------------------------------------------------
func TestInit_WithoutGlobalUser(t *testing.T) {
	// getConfig returns error (user not configured)
	origGet := getConfigFn
	t.Cleanup(func() { getConfigFn = origGet })
	getConfigFn = func(key string, global bool) (string, error) {
		return "", fmt.Errorf("config not set")
	}

	// askNewUser returns a test user
	origAsk := askNewUserFn
	t.Cleanup(func() { askNewUserFn = origAsk })
	askNewUserFn = func() (domain.User, error) {
		return domain.User{Name: "Prompted User", Email: "prompted@example.com"}, nil
	}

	// Capture setConfig calls
	origSet := setConfigFn
	t.Cleanup(func() { setConfigFn = origSet })
	setConfigCalls := make([]struct{ key, value string }, 0)
	setConfigFn = func(key, value string, global bool) error {
		setConfigCalls = append(setConfigCalls, struct{ key, value string }{key, value})
		return nil
	}

	// Capture printUserInfo
	var printedUser domain.User
	origPrint := printUserInfoFn
	t.Cleanup(func() { printUserInfoFn = origPrint })
	printUserInfoFn = func(user domain.User, opts format.PrintOptions) {
		printedUser = user
	}

	s := testStore(t)
	cmd := NewInitCmd(s)
	cmd.SetArgs([]string{})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("init failed: %v", err)
	}

	// Verify store
	users, err := s.List()
	if err != nil {
		t.Fatalf("List() failed: %v", err)
	}
	if len(users) != 1 {
		t.Fatalf("expected 1 user, got %d", len(users))
	}
	if users[0].Name != "Prompted User" {
		t.Errorf("Name = %q, want %q", users[0].Name, "Prompted User")
	}

	// Verify setConfig was called for both name and email
	if len(setConfigCalls) != 2 {
		t.Fatalf("expected 2 setConfig calls, got %d", len(setConfigCalls))
	}
	if setConfigCalls[0].key != "name" || setConfigCalls[0].value != "Prompted User" {
		t.Errorf("setConfig name = %q, want %q", setConfigCalls[0].value, "Prompted User")
	}
	if setConfigCalls[1].key != "email" || setConfigCalls[1].value != "prompted@example.com" {
		t.Errorf("setConfig email = %q, want %q", setConfigCalls[1].value, "prompted@example.com")
	}

	if printedUser.Name != "Prompted User" {
		t.Errorf("printed Name = %q, want %q", printedUser.Name, "Prompted User")
	}
}

// ---------------------------------------------------------------------------
// TestInit_WithExistingUsers_NoForce
// ---------------------------------------------------------------------------
func TestInit_WithExistingUsers_NoForce(t *testing.T) {
	useLocale(t, "en")

	// getConfig returns a valid user
	origGet := getConfigFn
	t.Cleanup(func() { getConfigFn = origGet })
	getConfigFn = func(key string, global bool) (string, error) {
		switch key {
		case "name":
			return "New User", nil
		case "email":
			return "new@example.com", nil
		}
		return "", nil
	}

	// confirm returns false (user declines override)
	origConfirm := confirmFn
	t.Cleanup(func() { confirmFn = origConfirm })
	confirmFn = func(msg string, defaultVal bool) (bool, error) {
		if !strings.Contains(msg, "Override existing") {
			t.Errorf("unexpected confirm message: %s", msg)
		}
		return false, nil
	}

	// Preload store with existing users
	existing := []domain.User{
		{Name: "Existing", Email: "existing@example.com"},
	}
	s := preloadedStore(t, existing)

	// Capture printUserInfo - should NOT be called
	origPrint := printUserInfoFn
	t.Cleanup(func() { printUserInfoFn = origPrint })
	printCalled := false
	printUserInfoFn = func(user domain.User, opts format.PrintOptions) {
		printCalled = true
	}

	cmd := NewInitCmd(s)
	cmd.SetArgs([]string{})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("init failed: %v", err)
	}

	// Store should still have only the existing user
	users, err := s.List()
	if err != nil {
		t.Fatalf("List() failed: %v", err)
	}
	if len(users) != 1 {
		t.Fatalf("expected 1 existing user, got %d", len(users))
	}
	if users[0].Name != "Existing" {
		t.Errorf("store should not be modified, got Name=%q", users[0].Name)
	}

	if printCalled {
		t.Error("printUserInfo should not have been called when user declines")
	}
}

// ---------------------------------------------------------------------------
// TestInit_WithForce
// ---------------------------------------------------------------------------
func TestInit_WithForce(t *testing.T) {
	origGet := getConfigFn
	t.Cleanup(func() { getConfigFn = origGet })
	getConfigFn = func(key string, global bool) (string, error) {
		switch key {
		case "name":
			return "Forced User", nil
		case "email":
			return "forced@example.com", nil
		}
		return "", nil
	}

	// confirm should NOT be called with --force
	origConfirm := confirmFn
	t.Cleanup(func() { confirmFn = origConfirm })
	confirmFn = func(msg string, defaultVal bool) (bool, error) {
		t.Error("confirm should not have been called with --force")
		return false, nil
	}

	// Preload store
	existing := []domain.User{
		{Name: "Old User", Email: "old@example.com"},
	}
	s := preloadedStore(t, existing)

	cmd := NewInitCmd(s)
	cmd.SetArgs([]string{"--force"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("init --force failed: %v", err)
	}

	// Store should be overwritten
	users, err := s.List()
	if err != nil {
		t.Fatalf("List() failed: %v", err)
	}
	if len(users) != 1 {
		t.Fatalf("expected 1 user after force, got %d", len(users))
	}
	if users[0].Name != "Forced User" {
		t.Errorf("Name = %q, want %q", users[0].Name, "Forced User")
	}
}

// ---------------------------------------------------------------------------
// TestInit_LegacyMigration_Confirmed
// ---------------------------------------------------------------------------
func TestInit_LegacyMigration_Confirmed(t *testing.T) {
	useLocale(t, "en")

	// Setup legacy config
	legacyHome := t.TempDir()
	t.Setenv("HOME", legacyHome) // not used when userHomeDirFn is overridden

	origHome := userHomeDirFn
	t.Cleanup(func() { userHomeDirFn = origHome })
	userHomeDirFn = func() (string, error) { return legacyHome, nil }

	legacyDir := filepath.Join(legacyHome, ".gitusr")
	if err := os.MkdirAll(legacyDir, 0755); err != nil {
		t.Fatalf("create legacy dir: %v", err)
	}
	legacyUsers := []domain.User{
		{Name: "Legacy User", Email: "legacy@test.com"},
	}
	legacyData, _ := json.MarshalIndent(legacyUsers, "", "  ")
	legacyPath := filepath.Join(legacyDir, "user-list.json")
	if err := os.WriteFile(legacyPath, append(legacyData, '\n'), 0644); err != nil {
		t.Fatalf("write legacy config: %v", err)
	}

	// confirm returns true
	origConfirm := confirmFn
	t.Cleanup(func() { confirmFn = origConfirm })
	confirmFn = func(msg string, defaultVal bool) (bool, error) {
		if !strings.Contains(msg, "legacy") {
			t.Errorf("unexpected confirm message: %s", msg)
		}
		return true, nil
	}

	// getConfig should NOT be called (migration returns early)
	origGet := getConfigFn
	t.Cleanup(func() { getConfigFn = origGet })
	getConfigCalled := false
	getConfigFn = func(key string, global bool) (string, error) {
		getConfigCalled = true
		return "should-not-be-used", nil
	}

	s := testStore(t)
	cmd := NewInitCmd(s)

	var stdout string
	stdout = captureStdout(t, func() {
		cmd.SetArgs([]string{})
		if err := cmd.Execute(); err != nil {
			t.Fatalf("init failed: %v", err)
		}
	})

	// Verify store has legacy user
	users, err := s.List()
	if err != nil {
		t.Fatalf("List() failed: %v", err)
	}
	if len(users) != 1 {
		t.Fatalf("expected 1 migrated user, got %d", len(users))
	}
	if users[0].Name != "Legacy User" {
		t.Errorf("Name = %q, want %q", users[0].Name, "Legacy User")
	}
	if users[0].Email != "legacy@test.com" {
		t.Errorf("Email = %q, want %q", users[0].Email, "legacy@test.com")
	}

	// Verify stdout messages
	if !strings.Contains(stdout, "Migrated 1 users from legacy config") {
		t.Errorf("stdout missing migration message, got: %s", stdout)
	}

	if getConfigCalled {
		t.Error("getConfig should not have been called after successful migration")
	}
}

// ---------------------------------------------------------------------------
// TestInit_LegacyMigration_Skip
// ---------------------------------------------------------------------------
func TestInit_LegacyMigration_Skip(t *testing.T) {
	useLocale(t, "en")

	legacyHome := t.TempDir()

	origHome := userHomeDirFn
	t.Cleanup(func() { userHomeDirFn = origHome })
	userHomeDirFn = func() (string, error) { return legacyHome, nil }

	legacyDir := filepath.Join(legacyHome, ".gitusr")
	if err := os.MkdirAll(legacyDir, 0755); err != nil {
		t.Fatalf("create legacy dir: %v", err)
	}
	legacyUsers := []domain.User{
		{Name: "Skipped Legacy", Email: "skip@test.com"},
	}
	legacyData, _ := json.MarshalIndent(legacyUsers, "", "  ")
	legacyPath := filepath.Join(legacyDir, "user-list.json")
	if err := os.WriteFile(legacyPath, append(legacyData, '\n'), 0644); err != nil {
		t.Fatalf("write legacy config: %v", err)
	}

	// confirm returns false (user declines)
	origConfirm := confirmFn
	t.Cleanup(func() { confirmFn = origConfirm })
	confirmFn = func(msg string, defaultVal bool) (bool, error) {
		return false, nil
	}

	// getConfig returns a user (normal init proceeds)
	origGet := getConfigFn
	t.Cleanup(func() { getConfigFn = origGet })
	getConfigFn = func(key string, global bool) (string, error) {
		switch key {
		case "name":
			return "Normal Init User", nil
		case "email":
			return "normal@example.com", nil
		}
		return "", nil
	}

	// Capture printUserInfo
	origPrint := printUserInfoFn
	t.Cleanup(func() { printUserInfoFn = origPrint })
	var printedUser domain.User
	printUserInfoFn = func(user domain.User, opts format.PrintOptions) {
		printedUser = user
	}

	s := testStore(t)
	cmd := NewInitCmd(s)

	var stdout string
	stdout = captureStdout(t, func() {
		cmd.SetArgs([]string{})
		if err := cmd.Execute(); err != nil {
			t.Fatalf("init failed: %v", err)
		}
	})

	// Verify "Migration skipped." was printed
	if !strings.Contains(stdout, "Migration skipped") {
		t.Errorf("stdout missing 'Migration skipped', got: %s", stdout)
	}

	// Store should have the normal init user (NOT the legacy user)
	users, err := s.List()
	if err != nil {
		t.Fatalf("List() failed: %v", err)
	}
	if len(users) != 1 {
		t.Fatalf("expected 1 user from normal init, got %d", len(users))
	}
	if users[0].Name != "Normal Init User" {
		t.Errorf("Name = %q, want %q", users[0].Name, "Normal Init User")
	}
	if printedUser.Name != "Normal Init User" {
		t.Errorf("printed Name = %q, want %q", printedUser.Name, "Normal Init User")
	}
}

// ---------------------------------------------------------------------------
// TestInit_LegacyMigration_NoLegacyFile
// ---------------------------------------------------------------------------
func TestInit_LegacyMigration_NoLegacyFile(t *testing.T) {
	useLocale(t, "en")

	legacyHome := t.TempDir()

	origHome := userHomeDirFn
	t.Cleanup(func() { userHomeDirFn = origHome })
	userHomeDirFn = func() (string, error) { return legacyHome, nil }

	// No legacy file created

	origConfirm := confirmFn
	t.Cleanup(func() { confirmFn = origConfirm })
	confirmCalled := false
	confirmFn = func(msg string, defaultVal bool) (bool, error) {
		confirmCalled = true
		return false, nil
	}

	origGet := getConfigFn
	t.Cleanup(func() { getConfigFn = origGet })
	getConfigFn = func(key string, global bool) (string, error) {
		switch key {
		case "name":
			return "Normal User", nil
		case "email":
			return "normal@example.com", nil
		}
		return "", nil
	}

	origPrint := printUserInfoFn
	t.Cleanup(func() { printUserInfoFn = origPrint })
	var printedUser domain.User
	printUserInfoFn = func(user domain.User, opts format.PrintOptions) {
		printedUser = user
	}

	s := testStore(t)
	cmd := NewInitCmd(s)

	var stdout string
	stdout = captureStdout(t, func() {
		cmd.SetArgs([]string{})
		if err := cmd.Execute(); err != nil {
			t.Fatalf("init failed: %v", err)
		}
	})

	// confirm should NOT have been called for legacy migration
	if confirmCalled {
		t.Error("confirm was called unexpectedly (no legacy file)")
	}

	// Normal init should proceed
	users, err := s.List()
	if err != nil {
		t.Fatalf("List() failed: %v", err)
	}
	if len(users) != 1 {
		t.Fatalf("expected 1 user, got %d", len(users))
	}
	if users[0].Name != "Normal User" {
		t.Errorf("Name = %q, want %q", users[0].Name, "Normal User")
	}
	if printedUser.Name != "Normal User" {
		t.Errorf("printed Name = %q, want %q", printedUser.Name, "Normal User")
	}

	// No legacy messages
	if strings.Contains(stdout, "legacy") || strings.Contains(stdout, "Migration") {
		t.Errorf("unexpected legacy message in stdout: %s", stdout)
	}
}

// ---------------------------------------------------------------------------
// TestInit_GetConfigReturnsError_ThenPromptAndSet
// ---------------------------------------------------------------------------
func TestInit_GetConfigReturnsError_ThenPromptAndSet(t *testing.T) {
	// name is set, but email returns error → should still go to prompt
	origGet := getConfigFn
	t.Cleanup(func() { getConfigFn = origGet })
	getConfigFn = func(key string, global bool) (string, error) {
		if key == "name" {
			return "Partial", nil
		}
		return "", fmt.Errorf("not set")
	}

	origAsk := askNewUserFn
	t.Cleanup(func() { askNewUserFn = origAsk })
	askNewUserFn = func() (domain.User, error) {
		return domain.User{Name: "Full Prompt", Email: "full@example.com"}, nil
	}

	origSet := setConfigFn
	t.Cleanup(func() { setConfigFn = origSet })
	setConfigCalls := make([]string, 0)
	setConfigFn = func(key, value string, global bool) error {
		setConfigCalls = append(setConfigCalls, key+"="+value)
		return nil
	}

	s := testStore(t)
	cmd := NewInitCmd(s)
	cmd.SetArgs([]string{})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("init failed: %v", err)
	}

	// Verify setConfig was called for both
	if len(setConfigCalls) != 2 {
		t.Fatalf("expected 2 setConfig calls, got %d", len(setConfigCalls))
	}
	if setConfigCalls[0] != "name=Full Prompt" {
		t.Errorf("setConfig name: got %q", setConfigCalls[0])
	}
	if setConfigCalls[1] != "email=full@example.com" {
		t.Errorf("setConfig email: got %q", setConfigCalls[1])
	}

	users, _ := s.List()
	if len(users) != 1 || users[0].Name != "Full Prompt" {
		t.Errorf("stored user = %v", users)
	}
}

// ---------------------------------------------------------------------------
// TestInit_Success_En — init with English locale
// ---------------------------------------------------------------------------
func TestInit_Success_En(t *testing.T) {
	useLocale(t, "en")

	origGet := getConfigFn
	t.Cleanup(func() { getConfigFn = origGet })
	getConfigFn = func(key string, global bool) (string, error) {
		switch key {
		case "name":
			return "Alice", nil
		case "email":
			return "alice@example.com", nil
		}
		return "", fmt.Errorf("unknown key: %s", key)
	}

	origConfirm := confirmFn
	t.Cleanup(func() { confirmFn = origConfirm })
	confirmCalled := false
	confirmFn = func(msg string, defaultVal bool) (bool, error) {
		confirmCalled = true
		return false, nil
	}

	s := testStore(t)
	cmd := NewInitCmd(s)
	cmd.SetArgs([]string{})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("init failed: %v", err)
	}

	users, err := s.List()
	if err != nil {
		t.Fatalf("List() failed: %v", err)
	}
	if len(users) != 1 {
		t.Fatalf("expected 1 user, got %d", len(users))
	}
	if users[0].Name != "Alice" {
		t.Errorf("Name = %q, want %q", users[0].Name, "Alice")
	}
	if users[0].Email != "alice@example.com" {
		t.Errorf("Email = %q, want %q", users[0].Email, "alice@example.com")
	}
	if confirmCalled {
		t.Error("confirm should not have been called for empty store")
	}
}

// ---------------------------------------------------------------------------
// TestInit_Success_ZhCN — init with Chinese locale
// ---------------------------------------------------------------------------
func TestInit_Success_ZhCN(t *testing.T) {
	useLocale(t, "zh-CN")

	origGet := getConfigFn
	t.Cleanup(func() { getConfigFn = origGet })
	getConfigFn = func(key string, global bool) (string, error) {
		switch key {
		case "name":
			return "小明", nil
		case "email":
			return "xiaoming@example.com", nil
		}
		return "", fmt.Errorf("unknown key: %s", key)
	}

	origConfirm := confirmFn
	t.Cleanup(func() { confirmFn = origConfirm })
	confirmCalled := false
	confirmFn = func(msg string, defaultVal bool) (bool, error) {
		confirmCalled = true
		return false, nil
	}

	s := testStore(t)
	cmd := NewInitCmd(s)
	cmd.SetArgs([]string{})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("init failed: %v", err)
	}

	users, err := s.List()
	if err != nil {
		t.Fatalf("List() failed: %v", err)
	}
	if len(users) != 1 {
		t.Fatalf("expected 1 user, got %d", len(users))
	}
	if users[0].Name != "小明" {
		t.Errorf("Name = %q, want %q", users[0].Name, "小明")
	}
	if confirmCalled {
		t.Error("confirm should not have been called for empty store")
	}
}

// ---------------------------------------------------------------------------
// TestInit_Override_ZhCN — override confirm prompt in Chinese
// ---------------------------------------------------------------------------
func TestInit_Override_ZhCN(t *testing.T) {
	useLocale(t, "zh-CN")

	origGet := getConfigFn
	t.Cleanup(func() { getConfigFn = origGet })
	getConfigFn = func(key string, global bool) (string, error) {
		switch key {
		case "name":
			return "New User", nil
		case "email":
			return "new@example.com", nil
		}
		return "", nil
	}

	origConfirm := confirmFn
	t.Cleanup(func() { confirmFn = origConfirm })
	confirmFn = func(msg string, defaultVal bool) (bool, error) {
		want := "覆盖现有用户？"
		if msg != want {
			t.Errorf("confirm message = %q, want %q", msg, want)
		}
		return true, nil
	}

	existing := []domain.User{
		{Name: "Existing", Email: "existing@example.com"},
	}
	s := preloadedStore(t, existing)

	cmd := NewInitCmd(s)
	cmd.SetArgs([]string{})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("init failed: %v", err)
	}

	users, err := s.List()
	if err != nil {
		t.Fatalf("List() failed: %v", err)
	}
	if len(users) != 1 {
		t.Fatalf("expected 1 user after override, got %d", len(users))
	}
	if users[0].Name != "New User" {
		t.Errorf("Name = %q, want %q", users[0].Name, "New User")
	}
}

// ---------------------------------------------------------------------------
// TestInit_WithFlags_EmptyStore — flags bypass git-config and prompts
// ---------------------------------------------------------------------------
func TestInit_WithFlags_EmptyStore(t *testing.T) {
	useLocale(t, "en")

	// getConfig should NOT be called
	origGet := getConfigFn
	t.Cleanup(func() { getConfigFn = origGet })
	getConfigCalled := false
	getConfigFn = func(key string, global bool) (string, error) {
		getConfigCalled = true
		return "", nil
	}

	// askNewUser should NOT be called
	origAsk := askNewUserFn
	t.Cleanup(func() { askNewUserFn = origAsk })
	askCalled := false
	askNewUserFn = func() (domain.User, error) {
		askCalled = true
		return domain.User{}, nil
	}

	// Capture printUserInfo
	var printedUser domain.User
	var printedOpts format.PrintOptions
	origPrint := printUserInfoFn
	t.Cleanup(func() { printUserInfoFn = origPrint })
	printUserInfoFn = func(user domain.User, opts format.PrintOptions) {
		printedUser = user
		printedOpts = opts
	}

	s := testStore(t)
	cmd := NewInitCmd(s)
	cmd.SetArgs([]string{"--name", "FlagInit", "--email", "flag@init.com"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("init --name/--email failed: %v", err)
	}

	users, err := s.List()
	if err != nil {
		t.Fatalf("List() failed: %v", err)
	}
	if len(users) != 1 {
		t.Fatalf("expected 1 user, got %d", len(users))
	}
	if users[0].Name != "FlagInit" {
		t.Errorf("Name = %q, want %q", users[0].Name, "FlagInit")
	}
	if users[0].Email != "flag@init.com" {
		t.Errorf("Email = %q, want %q", users[0].Email, "flag@init.com")
	}

	if getConfigCalled {
		t.Error("getConfig was called unexpectedly when flags are provided")
	}
	if askCalled {
		t.Error("askNewUser was called unexpectedly when flags are provided")
	}

	if printedUser.Name != "FlagInit" {
		t.Errorf("printed Name = %q, want %q", printedUser.Name, "FlagInit")
	}
	if !printedOpts.Global {
		t.Error("expected Global=true in PrintOptions")
	}
	if !printedOpts.ShowSuccess {
		t.Error("expected ShowSuccess=true in PrintOptions")
	}
}

// ---------------------------------------------------------------------------
// TestInit_WithFlags_Override — --yes skips override confirmation
// ---------------------------------------------------------------------------
func TestInit_WithFlags_Override(t *testing.T) {
	useLocale(t, "en")

	origConfirm := confirmFn
	t.Cleanup(func() { confirmFn = origConfirm })
	confirmFn = func(msg string, defaultVal bool) (bool, error) {
		t.Error("confirm was called unexpectedly with --yes flag")
		return false, nil
	}

	existing := []domain.User{
		{Name: "Old", Email: "old@example.com"},
	}
	s := preloadedStore(t, existing)

	cmd := NewInitCmd(s)
	cmd.SetArgs([]string{"--name", "NewOverride", "--email", "new@override.com", "--yes"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("init --name/--email --yes failed: %v", err)
	}

	users, err := s.List()
	if err != nil {
		t.Fatalf("List() failed: %v", err)
	}
	if len(users) != 1 {
		t.Fatalf("expected 1 user after override, got %d", len(users))
	}
	if users[0].Name != "NewOverride" {
		t.Errorf("Name = %q, want %q", users[0].Name, "NewOverride")
	}
	if users[0].Email != "new@override.com" {
		t.Errorf("Email = %q, want %q", users[0].Email, "new@override.com")
	}
}

// ---------------------------------------------------------------------------
// TestInit_WithFlags_Partial — only one flag returns error
// ---------------------------------------------------------------------------
func TestInit_WithFlags_Partial(t *testing.T) {
	useLocale(t, "en")

	origGet := getConfigFn
	t.Cleanup(func() { getConfigFn = origGet })
	getConfigFn = func(key string, global bool) (string, error) {
		return "", fmt.Errorf("should not be called")
	}

	s := testStore(t)
	cmd := NewInitCmd(s)
	cmd.SetArgs([]string{"--name", "OnlyName"})
	err := cmd.Execute()

	if err == nil {
		t.Fatal("expected error when only --name is provided")
	}
	if !strings.Contains(err.Error(), "err_name_and_email_required") {
		t.Errorf("error should mention name and email required, got: %q", err.Error())
	}

	cmd2 := NewInitCmd(testStore(t))
	cmd2.SetArgs([]string{"--email", "only@email.com"})
	err2 := cmd2.Execute()

	if err2 == nil {
		t.Fatal("expected error when only --email is provided")
	}
	if !strings.Contains(err2.Error(), "err_name_and_email_required") {
		t.Errorf("error should mention name and email required, got: %q", err2.Error())
	}
}

// ---------------------------------------------------------------------------
// TestInit_Migrate_ZhCN — migration success in Chinese
// ---------------------------------------------------------------------------
func TestInit_Migrate_ZhCN(t *testing.T) {
	useLocale(t, "zh-CN")

	legacyHome := t.TempDir()

	origHome := userHomeDirFn
	t.Cleanup(func() { userHomeDirFn = origHome })
	userHomeDirFn = func() (string, error) { return legacyHome, nil }

	legacyDir := filepath.Join(legacyHome, ".gitusr")
	if err := os.MkdirAll(legacyDir, 0755); err != nil {
		t.Fatalf("create legacy dir: %v", err)
	}
	legacyUsers := []domain.User{
		{Name: "旧用户", Email: "old@test.com"},
		{Name: "旧用户2", Email: "old2@test.com"},
	}
	legacyData, _ := json.MarshalIndent(legacyUsers, "", "  ")
	legacyPath := filepath.Join(legacyDir, "user-list.json")
	if err := os.WriteFile(legacyPath, append(legacyData, '\n'), 0644); err != nil {
		t.Fatalf("write legacy config: %v", err)
	}

	origConfirm := confirmFn
	t.Cleanup(func() { confirmFn = origConfirm })
	confirmFn = func(msg string, defaultVal bool) (bool, error) {
		want := "在 ~/.gitusr/user-list.json 找到旧配置。迁移到 XDG 目录？"
		if msg != want {
			t.Errorf("confirm message = %q, want %q", msg, want)
		}
		return true, nil
	}

	s := testStore(t)
	cmd := NewInitCmd(s)

	var stdout string
	stdout = captureStdout(t, func() {
		cmd.SetArgs([]string{})
		if err := cmd.Execute(); err != nil {
			t.Fatalf("init failed: %v", err)
		}
	})

	users, err := s.List()
	if err != nil {
		t.Fatalf("List() failed: %v", err)
	}
	if len(users) != 2 {
		t.Fatalf("expected 2 migrated users, got %d", len(users))
	}
	if users[0].Name != "旧用户" {
		t.Errorf("Name = %q, want %q", users[0].Name, "旧用户")
	}

	wantMsg := "已从旧配置迁移 2 个用户。"
	if !strings.Contains(stdout, wantMsg) {
		t.Errorf("stdout missing %q, got: %s", wantMsg, stdout)
	}
}
