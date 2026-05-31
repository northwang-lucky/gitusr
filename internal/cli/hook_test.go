package cli

import (
	"fmt"
	"strings"
	"testing"

	"github.com/northwang-lucky/gitusr/internal/hook"
	"github.com/northwang-lucky/gitusr/internal/i18n"
)

func TestHookInstall_Success(t *testing.T) {
	i18n.ResetForTesting()
	i18n.InitWithLocale("en")

	origInstall := installFunc
	t.Cleanup(func() { installFunc = origInstall })

	installFunc = func(hookType hook.HookType, shells []hook.ShellType) ([]hook.HookInstallResult, error) {
		return []hook.HookInstallResult{
			{Type: hookType, Shell: hook.ShellTypeBash, FilePath: "/tmp/git-wrapper.sh"},
		}, nil
	}

	store := &mockStore{initialized: true}
	cmd := NewHookInstallCmd(store)
	stdout, _, err := executeCmd(cmd, "--type", "clone")

	if err != nil {
		t.Fatalf("NewHookInstallCmd().Execute(--type clone) unexpected error: %v", err)
	}

	if !strings.Contains(stdout, "successfully installed") {
		t.Errorf("expected 'successfully installed', got %q", stdout)
	}
}

func TestHookInstall_AlreadyInstalled(t *testing.T) {
	i18n.ResetForTesting()
	i18n.InitWithLocale("en")

	origInstall := installFunc
	t.Cleanup(func() { installFunc = origInstall })

	installFunc = func(hookType hook.HookType, shells []hook.ShellType) ([]hook.HookInstallResult, error) {
		// nil results means already installed (idempotent)
		return nil, nil
	}

	store := &mockStore{initialized: true}
	cmd := NewHookInstallCmd(store)
	stdout, _, err := executeCmd(cmd, "--type", "commit")

	if err != nil {
		t.Fatalf("NewHookInstallCmd().Execute(--type commit) unexpected error: %v", err)
	}

	if !strings.Contains(stdout, "already installed") {
		t.Errorf("expected 'already installed', got %q", stdout)
	}
}

func TestHookInstall_InvalidType(t *testing.T) {
	i18n.ResetForTesting()
	i18n.InitWithLocale("en")

	store := &mockStore{initialized: true}
	cmd := NewHookInstallCmd(store)
	_, _, err := executeCmd(cmd, "--type", "invalid")

	if err == nil {
		t.Fatal("expected error for invalid hook type")
	}

	if !strings.Contains(err.Error(), "invalid hook type") {
		t.Errorf("error should mention 'invalid hook type', got: %q", err.Error())
	}
}

func TestHookInstall_MissingType(t *testing.T) {
	i18n.ResetForTesting()
	i18n.InitWithLocale("en")

	store := &mockStore{initialized: true}
	cmd := NewHookInstallCmd(store)
	_, _, err := executeCmd(cmd)

	if err == nil {
		t.Fatal("expected error for missing --type flag")
	}

	if !strings.Contains(err.Error(), "required") {
		t.Errorf("error should mention 'required', got: %q", err.Error())
	}
}

func TestHookInstall_Error(t *testing.T) {
	i18n.ResetForTesting()
	i18n.InitWithLocale("en")

	origInstall := installFunc
	t.Cleanup(func() { installFunc = origInstall })

	installFunc = func(hookType hook.HookType, shells []hook.ShellType) ([]hook.HookInstallResult, error) {
		return nil, fmt.Errorf("write error: permission denied")
	}

	store := &mockStore{initialized: true}
	cmd := NewHookInstallCmd(store)
	_, _, err := executeCmd(cmd, "--type", "clone")

	if err == nil {
		t.Fatal("expected error from install failure")
	}

	if !strings.Contains(err.Error(), "permission denied") {
		t.Errorf("error should contain 'permission denied', got: %q", err.Error())
	}
}

func TestHookUninstall_Success(t *testing.T) {
	i18n.ResetForTesting()
	i18n.InitWithLocale("en")

	origUninstall := uninstallFunc
	t.Cleanup(func() { uninstallFunc = origUninstall })

	uninstallFunc = func(hookType hook.HookType, shells []hook.ShellType) error {
		return nil
	}

	store := &mockStore{initialized: true}
	cmd := NewHookUninstallCmd(store)
	stdout, _, err := executeCmd(cmd, "--type", "clone")

	if err != nil {
		t.Fatalf("NewHookUninstallCmd().Execute(--type clone) unexpected error: %v", err)
	}

	if !strings.Contains(stdout, "successfully uninstalled") {
		t.Errorf("expected 'successfully uninstalled', got %q", stdout)
	}
}

func TestHookUninstall_NotInstalled(t *testing.T) {
	i18n.ResetForTesting()
	i18n.InitWithLocale("en")

	origUninstall := uninstallFunc
	t.Cleanup(func() { uninstallFunc = origUninstall })

	uninstallFunc = func(hookType hook.HookType, shells []hook.ShellType) error {
		return fmt.Errorf("hook %s is not installed", hookType)
	}

	store := &mockStore{initialized: true}
	cmd := NewHookUninstallCmd(store)
	_, _, err := executeCmd(cmd, "--type", "commit")

	if err == nil {
		t.Fatal("expected error for not installed hook")
	}

	if !strings.Contains(err.Error(), "not installed") {
		t.Errorf("error should contain 'not installed', got: %q", err.Error())
	}
}

func TestHookUninstall_InvalidType(t *testing.T) {
	i18n.ResetForTesting()
	i18n.InitWithLocale("en")

	store := &mockStore{initialized: true}
	cmd := NewHookUninstallCmd(store)
	_, _, err := executeCmd(cmd, "--type", "nope")

	if err == nil {
		t.Fatal("expected error for invalid hook type")
	}

	if !strings.Contains(err.Error(), "invalid hook type") {
		t.Errorf("error should mention 'invalid hook type', got: %q", err.Error())
	}
}

func TestHookUninstall_Error(t *testing.T) {
	i18n.ResetForTesting()
	i18n.InitWithLocale("en")

	origUninstall := uninstallFunc
	t.Cleanup(func() { uninstallFunc = origUninstall })

	uninstallFunc = func(hookType hook.HookType, shells []hook.ShellType) error {
		return fmt.Errorf("remove error: permission denied")
	}

	store := &mockStore{initialized: true}
	cmd := NewHookUninstallCmd(store)
	_, _, err := executeCmd(cmd, "--type", "clone")

	if err == nil {
		t.Fatal("expected error from uninstall failure")
	}

	if !strings.Contains(err.Error(), "permission denied") {
		t.Errorf("error should contain 'permission denied', got: %q", err.Error())
	}
}

// TestHookCmd_Subcommands verifies the parent hook command correctly groups
// install and uninstall as subcommands.
func TestHookCmd_Subcommands(t *testing.T) {
	store := &mockStore{initialized: true}
	cmd := NewHookCmd(store)

	subs := cmd.Commands()
	if len(subs) != 2 {
		t.Fatalf("expected 2 subcommands, got %d", len(subs))
	}

	names := make(map[string]bool)
	for _, s := range subs {
		names[s.Name()] = true
	}

	if !names["install"] {
		t.Error("expected 'install' subcommand")
	}
	if !names["uninstall"] {
		t.Error("expected 'uninstall' subcommand")
	}
}

// TestHookCmd_Short verifies the parent hook command has the correct short description.
func TestHookCmd_Short(t *testing.T) {
	i18n.ResetForTesting()
	i18n.InitWithLocale("en")

	store := &mockStore{initialized: true}
	cmd := NewHookCmd(store)

	if cmd.Short != "Manage shell hooks" {
		t.Errorf("cmd.Short = %q, want %q", cmd.Short, "Manage shell hooks")
	}
}

// TestHookInstall_FlagRegistration verifies the --type flag is properly registered.
func TestHookInstall_FlagRegistration(t *testing.T) {
	store := &mockStore{initialized: true}
	cmd := NewHookInstallCmd(store)

	typeFlag := cmd.Flags().Lookup("type")
	if typeFlag == nil {
		t.Fatal("NewHookInstallCmd() should have --type flag")
	}

	if typeFlag.Usage == "" {
		t.Error("--type flag should have a usage description")
	}
}

// TestHookUninstall_FlagRegistration verifies the --type flag is properly registered.
func TestHookUninstall_FlagRegistration(t *testing.T) {
	store := &mockStore{initialized: true}
	cmd := NewHookUninstallCmd(store)

	typeFlag := cmd.Flags().Lookup("type")
	if typeFlag == nil {
		t.Fatal("NewHookUninstallCmd() should have --type flag")
	}

	if typeFlag.Usage == "" {
		t.Error("--type flag should have a usage description")
	}
}

// TestHookInstall_Success_ZhCN verifies Chinese success message.
func TestHookInstall_Success_ZhCN(t *testing.T) {
	i18n.ResetForTesting()
	i18n.InitWithLocale("zh-CN")

	origInstall := installFunc
	t.Cleanup(func() { installFunc = origInstall })

	installFunc = func(hookType hook.HookType, shells []hook.ShellType) ([]hook.HookInstallResult, error) {
		return []hook.HookInstallResult{
			{Type: hookType, Shell: hook.ShellTypeBash, FilePath: "/tmp/git-wrapper.sh"},
		}, nil
	}

	store := &mockStore{initialized: true}
	cmd := NewHookInstallCmd(store)
	stdout, _, err := executeCmd(cmd, "--type", "clone")

	if err != nil {
		t.Fatalf("NewHookInstallCmd().Execute(--type clone) unexpected error: %v", err)
	}

	if !strings.Contains(stdout, "钩子安装成功") {
		t.Errorf("expected Chinese success message, got %q", stdout)
	}
}

// TestHookUninstall_Success_ZhCN verifies Chinese success message.
func TestHookUninstall_Success_ZhCN(t *testing.T) {
	i18n.ResetForTesting()
	i18n.InitWithLocale("zh-CN")

	origUninstall := uninstallFunc
	t.Cleanup(func() { uninstallFunc = origUninstall })

	uninstallFunc = func(hookType hook.HookType, shells []hook.ShellType) error {
		return nil
	}

	store := &mockStore{initialized: true}
	cmd := NewHookUninstallCmd(store)
	stdout, _, err := executeCmd(cmd, "--type", "clone")

	if err != nil {
		t.Fatalf("NewHookUninstallCmd().Execute(--type clone) unexpected error: %v", err)
	}

	if !strings.Contains(stdout, "钩子卸载成功") {
		t.Errorf("expected Chinese success message, got %q", stdout)
	}
}


