package cli

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/northwang-lucky/gitusr/internal/domain"
	"github.com/northwang-lucky/gitusr/internal/hook"
	"github.com/northwang-lucky/gitusr/internal/i18n"
	"github.com/northwang-lucky/gitusr/internal/store"
)

// newTestHostStore creates an isolated hosts.json-backed rule store.
func newTestHostStore(t *testing.T) domain.HostRuleStore {
	t.Helper()
	return store.NewJSONHostRuleStore(filepath.Join(t.TempDir(), "hosts.json"))
}

// setupHookState creates a temporary hook state file with the given disabled types
// and sets XDG_DATA_HOME so that hook.LoadState reads from the temp directory.
func setupHookState(t *testing.T, disabledTypes []hook.HookType) {
	t.Helper()

	dir := t.TempDir()
	xdgData := filepath.Join(dir, "gitusr")
	if err := os.MkdirAll(xdgData, 0755); err != nil {
		t.Fatalf("failed to create xdg data dir: %v", err)
	}

	state := hook.HookState{
		InstalledTypes: []hook.HookType{},
		DisabledTypes:  disabledTypes,
	}

	data, err := json.MarshalIndent(state, "", "\t")
	if err != nil {
		t.Fatalf("failed to marshal state: %v", err)
	}

	if err := os.WriteFile(filepath.Join(xdgData, "hook-state.json"), data, 0644); err != nil {
		t.Fatalf("failed to write state: %v", err)
	}

	t.Setenv("XDG_DATA_HOME", dir)
}

// --- Subcommand structure tests ---

func TestNewHooksCmd_Subcommands(t *testing.T) {
	i18n.ResetForTesting()
	i18n.InitWithLocale("en")

	store := &mockStore{}
	cmd := NewHooksCmd(store, newTestHostStore(t))

	subs := cmd.Commands()
	if len(subs) != 7 {
		t.Fatalf("expected 7 subcommands, got %d", len(subs))
	}

	names := make(map[string]bool)
	hidden := make(map[string]bool)
	for _, s := range subs {
		names[s.Name()] = true
		if s.Hidden {
			hidden[s.Name()] = true
		}
	}

	for _, name := range []string{"install", "uninstall", "enable", "disable", "apply-rc", "apply-host", "is-disabled"} {
		if !names[name] {
			t.Errorf("expected subcommand %q", name)
		}
	}

	if !hidden["apply-rc"] {
		t.Error("expected 'apply-rc' subcommand to be hidden")
	}
	if !hidden["apply-host"] {
		t.Error("expected 'apply-host' subcommand to be hidden")
	}
	if !hidden["is-disabled"] {
		t.Error("expected 'is-disabled' subcommand to be hidden")
	}
}

// --- Help text tests ---

func TestNewHooksCmd_Help(t *testing.T) {
	i18n.ResetForTesting()
	i18n.InitWithLocale("en")

	store := &mockStore{}
	cmd := NewHooksCmd(store, newTestHostStore(t))

	help, _, err := executeCmd(cmd, "--help")
	if err != nil {
		t.Fatalf("executeCmd --help unexpected error: %v", err)
	}

	for _, name := range []string{"install", "uninstall", "enable", "disable"} {
		if !strings.Contains(help, name) {
			t.Errorf("help should mention visible subcommand %q", name)
		}
	}

	for _, name := range []string{"apply-rc", "is-disabled"} {
		if strings.Contains(help, name) {
			t.Errorf("help should NOT mention hidden subcommand %q", name)
		}
	}
}

// --- is-disabled command tests ---

func TestNewHooksCmd_IsDisabled_WhenDisabled(t *testing.T) {
	i18n.ResetForTesting()
	i18n.InitWithLocale("en")

	setupHookState(t, []hook.HookType{hook.HookTypeClone})

	store := &mockStore{}
	cmd := NewHooksCmd(store, newTestHostStore(t))
	_, _, err := executeCmd(cmd, "is-disabled", "clone")

	if err != nil {
		t.Fatalf("expected exit 0 (disabled), got error: %v", err)
	}
}

func TestNewHooksCmd_IsDisabled_WhenEnabled(t *testing.T) {
	i18n.ResetForTesting()
	i18n.InitWithLocale("en")

	setupHookState(t, nil)

	store := &mockStore{}
	cmd := NewHooksCmd(store, newTestHostStore(t))
	_, _, err := executeCmd(cmd, "is-disabled", "clone")

	if err == nil {
		t.Fatal("expected exit 1 (enabled), got nil error")
	}
	if !strings.Contains(err.Error(), "not disabled") {
		t.Errorf("error should mention 'not disabled', got: %q", err.Error())
	}
}

func TestNewHooksCmd_IsDisabled_InvalidType(t *testing.T) {
	i18n.ResetForTesting()
	i18n.InitWithLocale("en")

	store := &mockStore{}
	cmd := NewHooksCmd(store, newTestHostStore(t))
	_, _, err := executeCmd(cmd, "is-disabled", "invalid")

	if err == nil {
		t.Fatal("expected error for invalid hook type")
	}
	if !strings.Contains(err.Error(), "invalid hook type") {
		t.Errorf("error should mention 'invalid hook type', got: %q", err.Error())
	}
}

func TestNewHooksCmd_IsDisabled_MissingArg(t *testing.T) {
	i18n.ResetForTesting()
	i18n.InitWithLocale("en")

	store := &mockStore{}
	cmd := NewHooksCmd(store, newTestHostStore(t))
	_, _, err := executeCmd(cmd, "is-disabled")

	if err == nil {
		t.Fatal("expected error for missing arg")
	}
}

// --- enable command tests ---

func TestNewHooksCmd_Enable_InvalidType(t *testing.T) {
	i18n.ResetForTesting()
	i18n.InitWithLocale("en")

	store := &mockStore{}
	cmd := NewHooksCmd(store, newTestHostStore(t))
	_, _, err := executeCmd(cmd, "enable", "invalid")

	if err == nil {
		t.Fatal("expected error for invalid hook type")
	}
	if !strings.Contains(err.Error(), "invalid hook type") {
		t.Errorf("error should mention 'invalid hook type', got: %q", err.Error())
	}
}

func TestNewHooksCmd_Enable_MissingArg(t *testing.T) {
	i18n.ResetForTesting()
	i18n.InitWithLocale("en")

	store := &mockStore{}
	cmd := NewHooksCmd(store, newTestHostStore(t))
	_, _, err := executeCmd(cmd, "enable")

	if err == nil {
		t.Fatal("expected error for missing positional arg")
	}
}

// --- disable command tests ---

func TestNewHooksDisableCmd_NoTypeFlag(t *testing.T) {
	cmd := NewHooksDisableCmd()

	if typeFlag := cmd.Flags().Lookup("type"); typeFlag != nil {
		t.Fatal("NewHooksDisableCmd() should NOT have a --type flag")
	}

	if cmd.Args == nil {
		t.Fatal("NewHooksDisableCmd() should set Args")
	}
}

func TestNewHooksCmd_Disable_InvalidType(t *testing.T) {
	i18n.ResetForTesting()
	i18n.InitWithLocale("en")

	store := &mockStore{}
	cmd := NewHooksCmd(store, newTestHostStore(t))
	_, _, err := executeCmd(cmd, "disable", "invalid")

	if err == nil {
		t.Fatal("expected error for invalid hook type")
	}
	if !strings.Contains(err.Error(), "invalid hook type") {
		t.Errorf("error should mention 'invalid hook type', got: %q", err.Error())
	}
}

func TestNewHooksCmd_Disable_MissingArg(t *testing.T) {
	i18n.ResetForTesting()
	i18n.InitWithLocale("en")

	store := &mockStore{}
	cmd := NewHooksCmd(store, newTestHostStore(t))
	_, _, err := executeCmd(cmd, "disable")

	if err == nil {
		t.Fatal("expected error for missing positional arg")
	}
}

// --- Enable / disable integration tests ---

func TestNewHooksCmd_Enable_Success(t *testing.T) {
	i18n.ResetForTesting()
	i18n.InitWithLocale("en")

	setupHookState(t, []hook.HookType{hook.HookTypeClone})

	store := &mockStore{}
	cmd := NewHooksCmd(store, newTestHostStore(t))
	stdout, _, err := executeCmd(cmd, "enable", "clone")

	if err != nil {
		t.Fatalf("enable clone unexpected error: %v", err)
	}
	if !strings.Contains(stdout, "Hook clone enabled") {
		t.Errorf("expected success message mentioning clone, got: %s", stdout)
	}

	enabled, err := hook.IsEnabled(hook.HookTypeClone)
	if err != nil {
		t.Fatalf("IsEnabled error: %v", err)
	}
	if !enabled {
		t.Error("expected clone hook to be enabled after enable command")
	}
}

func TestNewHooksCmd_Disable_Success(t *testing.T) {
	i18n.ResetForTesting()
	i18n.InitWithLocale("en")

	setupHookState(t, nil)

	store := &mockStore{}
	cmd := NewHooksCmd(store, newTestHostStore(t))
	stdout, _, err := executeCmd(cmd, "disable", "clone")

	if err != nil {
		t.Fatalf("disable clone unexpected error: %v", err)
	}
	if !strings.Contains(stdout, "Hook clone disabled") {
		t.Errorf("expected success message mentioning clone, got: %s", stdout)
	}

	enabled, err := hook.IsEnabled(hook.HookTypeClone)
	if err != nil {
		t.Fatalf("IsEnabled error: %v", err)
	}
	if enabled {
		t.Error("expected clone hook to be disabled after disable command")
	}
}

// --- hooks uninstall command tests ---

func TestHooksUninstall_Success(t *testing.T) {
	i18n.ResetForTesting()
	i18n.InitWithLocale("en")

	origUninstall := uninstallFunc
	t.Cleanup(func() { uninstallFunc = origUninstall })

	uninstallFunc = func(_ any, shells []hook.ShellType) error {
		if len(shells) != 2 {
			t.Errorf("expected 2 shells, got %d", len(shells))
		}
		return nil
	}

	store := &mockStore{}
	cmd := NewHooksCmd(store, newTestHostStore(t))
	stdout, _, err := executeCmd(cmd, "uninstall")

	if err != nil {
		t.Fatalf("hooks uninstall unexpected error: %v", err)
	}
	if !strings.Contains(stdout, "All hooks successfully uninstalled") {
		t.Errorf("expected success message, got: %s", stdout)
	}
}

func TestHooksUninstall_NoneInstalled(t *testing.T) {
	i18n.ResetForTesting()
	i18n.InitWithLocale("en")

	origUninstall := uninstallFunc
	t.Cleanup(func() { uninstallFunc = origUninstall })

	uninstallFunc = func(_ any, shells []hook.ShellType) error {
		return fmt.Errorf("no hooks are currently installed")
	}

	store := &mockStore{}
	cmd := NewHooksCmd(store, newTestHostStore(t))
	_, _, err := executeCmd(cmd, "uninstall")

	if err == nil {
		t.Fatal("expected error for no hooks installed")
	}
	if !strings.Contains(err.Error(), "No hooks are currently installed") {
		t.Errorf("error should mention 'No hooks are currently installed', got: %q", err.Error())
	}
}

func TestHooksUninstall_NotInstalled(t *testing.T) {
	i18n.ResetForTesting()
	i18n.InitWithLocale("en")

	origUninstall := uninstallFunc
	t.Cleanup(func() { uninstallFunc = origUninstall })

	uninstallFunc = func(_ any, shells []hook.ShellType) error {
		return fmt.Errorf("hook clone is not installed")
	}

	store := &mockStore{}
	cmd := NewHooksCmd(store, newTestHostStore(t))
	_, _, err := executeCmd(cmd, "uninstall")

	if err == nil {
		t.Fatal("expected error for not installed hook")
	}
	if !strings.Contains(err.Error(), "No hooks are currently installed") {
		t.Errorf("error should mention 'No hooks are currently installed', got: %q", err.Error())
	}
}

func TestHooksUninstall_Error(t *testing.T) {
	i18n.ResetForTesting()
	i18n.InitWithLocale("en")

	origUninstall := uninstallFunc
	t.Cleanup(func() { uninstallFunc = origUninstall })

	uninstallFunc = func(_ any, shells []hook.ShellType) error {
		return fmt.Errorf("remove error: permission denied")
	}

	store := &mockStore{}
	cmd := NewHooksCmd(store, newTestHostStore(t))
	_, _, err := executeCmd(cmd, "uninstall")

	if err == nil {
		t.Fatal("expected error from uninstall failure")
	}
	if !strings.Contains(err.Error(), "permission denied") {
		t.Errorf("error should contain 'permission denied', got: %q", err.Error())
	}
}
