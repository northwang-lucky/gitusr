package cli

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/northwang-lucky/gitusr/internal/hook"
	"github.com/northwang-lucky/gitusr/internal/i18n"
)

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
	cmd := NewHooksCmd(store)

	subs := cmd.Commands()
	if len(subs) != 6 {
		t.Fatalf("expected 6 subcommands, got %d", len(subs))
	}

	names := make(map[string]bool)
	hidden := make(map[string]bool)
	for _, s := range subs {
		names[s.Name()] = true
		if s.Hidden {
			hidden[s.Name()] = true
		}
	}

	for _, name := range []string{"install", "uninstall", "enable", "disable", "apply-rc", "is-disabled"} {
		if !names[name] {
			t.Errorf("expected subcommand %q", name)
		}
	}

	if !hidden["apply-rc"] {
		t.Error("expected 'apply-rc' subcommand to be hidden")
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
	cmd := NewHooksCmd(store)

	help, _, err := executeCmd(cmd, "--help")
	if err != nil {
		t.Fatalf("executeCmd --help unexpected error: %v", err)
	}

	// Help should mention all visible subcommands
	for _, name := range []string{"install", "uninstall", "enable", "disable"} {
		if !strings.Contains(help, name) {
			t.Errorf("help should mention visible subcommand %q", name)
		}
	}

	// Help should NOT mention hidden subcommands
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

	// Setup: clone hook IS disabled in state
	setupHookState(t, []hook.HookType{hook.HookTypeClone})

	store := &mockStore{}
	cmd := NewHooksCmd(store)
	_, _, err := executeCmd(cmd, "is-disabled", "clone")

	// Exit 0: the hook IS disabled
	if err != nil {
		t.Fatalf("expected exit 0 (disabled), got error: %v", err)
	}
}

func TestNewHooksCmd_IsDisabled_WhenEnabled(t *testing.T) {
	i18n.ResetForTesting()
	i18n.InitWithLocale("en")

	// Setup: empty state (no disabled types) → all hooks enabled
	setupHookState(t, nil)

	store := &mockStore{}
	cmd := NewHooksCmd(store)
	_, _, err := executeCmd(cmd, "is-disabled", "clone")

	// Exit 1: the hook is NOT disabled
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
	cmd := NewHooksCmd(store)
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
	cmd := NewHooksCmd(store)
	_, _, err := executeCmd(cmd, "is-disabled")

	if err == nil {
		t.Fatal("expected error for missing arg")
	}
}

// --- enable command tests ---

func TestNewHooksCmd_Enable_FlagRegistration(t *testing.T) {
	cmd := NewHookEnableCmd()

	typeFlag := cmd.Flags().Lookup("type")
	if typeFlag == nil {
		t.Fatal("NewHookEnableCmd() should have --type flag")
	}
	if typeFlag.Usage == "" {
		t.Error("--type flag should have a usage description")
	}
}

func TestNewHooksCmd_Enable_InvalidType(t *testing.T) {
	i18n.ResetForTesting()
	i18n.InitWithLocale("en")

	store := &mockStore{}
	cmd := NewHooksCmd(store)
	_, _, err := executeCmd(cmd, "enable", "--type", "invalid")

	if err == nil {
		t.Fatal("expected error for invalid hook type")
	}
	if !strings.Contains(err.Error(), "invalid hook type") {
		t.Errorf("error should mention 'invalid hook type', got: %q", err.Error())
	}
}

func TestNewHooksCmd_Enable_MissingType(t *testing.T) {
	i18n.ResetForTesting()
	i18n.InitWithLocale("en")

	store := &mockStore{}
	cmd := NewHooksCmd(store)
	_, _, err := executeCmd(cmd, "enable")

	if err == nil {
		t.Fatal("expected error for missing --type flag")
	}
	if !strings.Contains(err.Error(), "--type flag") {
		t.Errorf("error should mention --type flag, got: %q", err.Error())
	}
}

// --- disable command tests ---

func TestNewHooksCmd_Disable_FlagRegistration(t *testing.T) {
	cmd := NewHookDisableCmd()

	typeFlag := cmd.Flags().Lookup("type")
	if typeFlag == nil {
		t.Fatal("NewHookDisableCmd() should have --type flag")
	}
	if typeFlag.Usage == "" {
		t.Error("--type flag should have a usage description")
	}
}

func TestNewHooksCmd_Disable_InvalidType(t *testing.T) {
	i18n.ResetForTesting()
	i18n.InitWithLocale("en")

	store := &mockStore{}
	cmd := NewHooksCmd(store)
	_, _, err := executeCmd(cmd, "disable", "--type", "invalid")

	if err == nil {
		t.Fatal("expected error for invalid hook type")
	}
	if !strings.Contains(err.Error(), "invalid hook type") {
		t.Errorf("error should mention 'invalid hook type', got: %q", err.Error())
	}
}

func TestNewHooksCmd_Disable_MissingType(t *testing.T) {
	i18n.ResetForTesting()
	i18n.InitWithLocale("en")

	store := &mockStore{}
	cmd := NewHooksCmd(store)
	_, _, err := executeCmd(cmd, "disable")

	if err == nil {
		t.Fatal("expected error for missing --type flag")
	}
	if !strings.Contains(err.Error(), "--type flag") {
		t.Errorf("error should mention --type flag, got: %q", err.Error())
	}
}

// --- Enable / disable integration tests ---

func TestNewHooksCmd_Enable_Success(t *testing.T) {
	i18n.ResetForTesting()
	i18n.InitWithLocale("en")

	// Start with clone disabled
	setupHookState(t, []hook.HookType{hook.HookTypeClone})

	store := &mockStore{}
	cmd := NewHooksCmd(store)
	stdout, _, err := executeCmd(cmd, "enable", "--type", "clone")

	if err != nil {
		t.Fatalf("enable --type clone unexpected error: %v", err)
	}
	if !strings.Contains(stdout, "Hook type \"clone\" enabled") {
		t.Errorf("expected success message mentioning clone, got: %s", stdout)
	}

	// Verify the hook is no longer disabled
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

	// Start with clean state (no disabled types)
	setupHookState(t, nil)

	store := &mockStore{}
	cmd := NewHooksCmd(store)
	stdout, _, err := executeCmd(cmd, "disable", "--type", "clone")

	if err != nil {
		t.Fatalf("disable --type clone unexpected error: %v", err)
	}
	if !strings.Contains(stdout, "Hook type \"clone\" disabled") {
		t.Errorf("expected success message mentioning clone, got: %s", stdout)
	}

	// Verify the hook is now disabled
	enabled, err := hook.IsEnabled(hook.HookTypeClone)
	if err != nil {
		t.Fatalf("IsEnabled error: %v", err)
	}
	if enabled {
		t.Error("expected clone hook to be disabled after disable command")
	}
}
