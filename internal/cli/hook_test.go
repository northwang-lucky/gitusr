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

func TestHookInstall_Success_CD(t *testing.T) {
	i18n.ResetForTesting()
	i18n.InitWithLocale("en")

	origInstall := installFunc
	t.Cleanup(func() { installFunc = origInstall })

	installFunc = func(hookType hook.HookType, shells []hook.ShellType) ([]hook.HookInstallResult, error) {
		return []hook.HookInstallResult{
			{Type: hookType, Shell: hook.ShellTypeBash, FilePath: "/tmp/git-wrapper-cd.sh"},
		}, nil
	}

	store := &mockStore{initialized: true}
	cmd := NewHookInstallCmd(store)
	stdout, _, err := executeCmd(cmd, "--type", "cd")

	if err != nil {
		t.Fatalf("NewHookInstallCmd().Execute(--type cd) unexpected error: %v", err)
	}

	if !strings.Contains(stdout, "successfully installed") {
		t.Errorf("expected 'successfully installed', got %q", stdout)
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

	uninstallFunc = func(_ any, shells []hook.ShellType) error {
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

	uninstallFunc = func(_ any, shells []hook.ShellType) error {
		return fmt.Errorf("hook commit is not installed")
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

func TestHookUninstall_Success_CD(t *testing.T) {
	i18n.ResetForTesting()
	i18n.InitWithLocale("en")

	origUninstall := uninstallFunc
	t.Cleanup(func() { uninstallFunc = origUninstall })

	uninstallFunc = func(_ any, shells []hook.ShellType) error {
		return nil
	}

	store := &mockStore{initialized: true}
	cmd := NewHookUninstallCmd(store)
	stdout, _, err := executeCmd(cmd, "--type", "cd")

	if err != nil {
		t.Fatalf("NewHookUninstallCmd().Execute(--type cd) unexpected error: %v", err)
	}

	if !strings.Contains(stdout, "successfully uninstalled") {
		t.Errorf("expected 'successfully uninstalled', got %q", stdout)
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

	uninstallFunc = func(_ any, shells []hook.ShellType) error {
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
// install, uninstall, and apply-rc as subcommands (apply-rc is hidden).
func TestHookCmd_Subcommands(t *testing.T) {
	store := &mockStore{initialized: true}
	cmd := NewHookCmd(store)

	subs := cmd.Commands()
	if len(subs) != 3 {
		t.Fatalf("expected 3 subcommands, got %d", len(subs))
	}

	names := make(map[string]bool)
	hidden := make(map[string]bool)
	for _, s := range subs {
		names[s.Name()] = true
		if s.Hidden {
			hidden[s.Name()] = true
		}
	}

	if !names["install"] {
		t.Error("expected 'install' subcommand")
	}
	if !names["uninstall"] {
		t.Error("expected 'uninstall' subcommand")
	}
	if !names["apply-rc"] {
		t.Error("expected 'apply-rc' subcommand")
	}
	if !hidden["apply-rc"] {
		t.Error("expected 'apply-rc' subcommand to be hidden")
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

	uninstallFunc = func(_ any, shells []hook.ShellType) error {
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

// TestHookInstall_All verifies --all installs all three hook types.
func TestHookInstall_All(t *testing.T) {
	i18n.ResetForTesting()
	i18n.InitWithLocale("en")

	origInstall := installFunc
	t.Cleanup(func() { installFunc = origInstall })

	callCount := 0
	installFunc = func(hookType hook.HookType, shells []hook.ShellType) ([]hook.HookInstallResult, error) {
		callCount++
		return []hook.HookInstallResult{
			{Type: hookType, Shell: hook.ShellTypeBash, FilePath: "/tmp/git-wrapper.sh"},
		}, nil
	}

	store := &mockStore{initialized: true}
	cmd := NewHookInstallCmd(store)
	stdout, _, err := executeCmd(cmd, "--all")

	if err != nil {
		t.Fatalf("--all unexpected error: %v", err)
	}
	if callCount != 3 {
		t.Errorf("expected 3 install calls, got %d", callCount)
	}
	for _, typ := range []string{"clone", "commit", "cd"} {
		if !strings.Contains(stdout, typ) {
			t.Errorf("expected output to mention %q, got: %s", typ, stdout)
		}
	}
	if !strings.Contains(stdout, "All hooks successfully installed") {
		t.Errorf("expected 'All hooks successfully installed', got: %s", stdout)
	}
}

// TestHookInstall_All_AlreadyInstalled verifies --all when all types are already installed.
func TestHookInstall_All_AlreadyInstalled(t *testing.T) {
	i18n.ResetForTesting()
	i18n.InitWithLocale("en")

	origInstall := installFunc
	t.Cleanup(func() { installFunc = origInstall })

	callCount := 0
	installFunc = func(hookType hook.HookType, shells []hook.ShellType) ([]hook.HookInstallResult, error) {
		callCount++
		return nil, nil
	}

	store := &mockStore{initialized: true}
	cmd := NewHookInstallCmd(store)
	stdout, _, err := executeCmd(cmd, "--all")

	if err != nil {
		t.Fatalf("--all unexpected error: %v", err)
	}
	if callCount != 3 {
		t.Errorf("expected 3 install calls, got %d", callCount)
	}
	for _, typ := range []string{"clone", "commit", "cd"} {
		if !strings.Contains(stdout, typ) {
			t.Errorf("expected output to mention %q, got: %s", typ, stdout)
		}
	}
	if !strings.Contains(stdout, "already installed") {
		t.Errorf("expected 'already installed' messages, got: %s", stdout)
	}
}

// TestHookInstall_All_PartialError verifies --all aborts on the first real error.
func TestHookInstall_All_PartialError(t *testing.T) {
	i18n.ResetForTesting()
	i18n.InitWithLocale("en")

	origInstall := installFunc
	t.Cleanup(func() { installFunc = origInstall })

	callCount := 0
	installFunc = func(hookType hook.HookType, shells []hook.ShellType) ([]hook.HookInstallResult, error) {
		callCount++
		if callCount == 2 {
			return nil, fmt.Errorf("write error: permission denied")
		}
		return []hook.HookInstallResult{
			{Type: hookType, Shell: hook.ShellTypeBash, FilePath: "/tmp/git-wrapper.sh"},
		}, nil
	}

	store := &mockStore{initialized: true}
	cmd := NewHookInstallCmd(store)
	_, _, err := executeCmd(cmd, "--all")

	if err == nil {
		t.Fatal("expected error from second install failure")
	}
	if !strings.Contains(err.Error(), "permission denied") {
		t.Errorf("error should contain 'permission denied', got: %q", err.Error())
	}
	if callCount != 2 {
		t.Errorf("expected abort after 2 calls, got %d", callCount)
	}
}

// TestHookInstall_All_And_Type_Exclusive verifies --all and --type are mutually exclusive.
func TestHookInstall_All_And_Type_Exclusive(t *testing.T) {
	i18n.ResetForTesting()
	i18n.InitWithLocale("en")

	store := &mockStore{initialized: true}
	cmd := NewHookInstallCmd(store)
	_, _, err := executeCmd(cmd, "--all", "--type", "clone")

	if err == nil {
		t.Fatal("expected error for mutually exclusive flags")
	}
	if !strings.Contains(err.Error(), "mutually exclusive") {
		t.Errorf("error should mention 'mutually exclusive', got: %q", err.Error())
	}
}

// TestHookUninstall_All verifies --all uninstalls all three hook types.
func TestHookUninstall_All(t *testing.T) {
	i18n.ResetForTesting()
	i18n.InitWithLocale("en")

	origUninstall := uninstallFunc
	t.Cleanup(func() { uninstallFunc = origUninstall })

	callCount := 0
	uninstallFunc = func(_ any, shells []hook.ShellType) error {
		callCount++
		return nil
	}

	store := &mockStore{initialized: true}
	cmd := NewHookUninstallCmd(store)
	stdout, _, err := executeCmd(cmd, "--all")

	if err != nil {
		t.Fatalf("--all unexpected error: %v", err)
	}
	if callCount != 3 {
		t.Errorf("expected 3 uninstall calls, got %d", callCount)
	}
	for _, typ := range []string{"clone", "commit", "cd"} {
		if !strings.Contains(stdout, typ) {
			t.Errorf("expected output to mention %q, got: %s", typ, stdout)
		}
	}
	if !strings.Contains(stdout, "All hooks successfully uninstalled") {
		t.Errorf("expected 'All hooks successfully uninstalled', got: %s", stdout)
	}
}

// TestHookUninstall_All_NotInstalled verifies --all skips not-installed types.
func TestHookUninstall_All_NotInstalled(t *testing.T) {
	i18n.ResetForTesting()
	i18n.InitWithLocale("en")

	origUninstall := uninstallFunc
	t.Cleanup(func() { uninstallFunc = origUninstall })

	callCount := 0
	uninstallFunc = func(_ any, shells []hook.ShellType) error {
		callCount++
		return fmt.Errorf("hook clone is not installed")
	}

	store := &mockStore{initialized: true}
	cmd := NewHookUninstallCmd(store)
	stdout, _, err := executeCmd(cmd, "--all")

	if err != nil {
		t.Fatalf("--all should not error when all not installed: %v", err)
	}
	if callCount != 3 {
		t.Errorf("expected 3 uninstall calls, got %d", callCount)
	}
	for _, typ := range []string{"clone", "commit", "cd"} {
		if !strings.Contains(stdout, typ) {
			t.Errorf("expected output to mention %q, got: %s", typ, stdout)
		}
	}
	if !strings.Contains(stdout, "not installed") {
		t.Errorf("expected 'not installed' messages, got: %s", stdout)
	}
}

// TestHookUninstall_All_And_Type_Exclusive verifies --all and --type are mutually exclusive for uninstall.
func TestHookUninstall_All_And_Type_Exclusive(t *testing.T) {
	i18n.ResetForTesting()
	i18n.InitWithLocale("en")

	store := &mockStore{initialized: true}
	cmd := NewHookUninstallCmd(store)
	_, _, err := executeCmd(cmd, "--all", "--type", "clone")

	if err == nil {
		t.Fatal("expected error for mutually exclusive flags")
	}
	if !strings.Contains(err.Error(), "mutually exclusive") {
		t.Errorf("error should mention 'mutually exclusive', got: %q", err.Error())
	}
}

// TestHookInstall_All_ZhCN verifies Chinese messages for --all install.
func TestHookInstall_All_ZhCN(t *testing.T) {
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
	stdout, _, err := executeCmd(cmd, "--all")

	if err != nil {
		t.Fatalf("--all unexpected error: %v", err)
	}
	if !strings.Contains(stdout, "所有钩子安装成功") {
		t.Errorf("expected Chinese all-success message, got: %s", stdout)
	}
}

// TestHookUninstall_All_ZhCN verifies Chinese messages for --all uninstall.
func TestHookUninstall_All_ZhCN(t *testing.T) {
	i18n.ResetForTesting()
	i18n.InitWithLocale("zh-CN")

	origUninstall := uninstallFunc
	t.Cleanup(func() { uninstallFunc = origUninstall })

	uninstallFunc = func(_ any, shells []hook.ShellType) error {
		return nil
	}

	store := &mockStore{initialized: true}
	cmd := NewHookUninstallCmd(store)
	stdout, _, err := executeCmd(cmd, "--all")

	if err != nil {
		t.Fatalf("--all unexpected error: %v", err)
	}
	if !strings.Contains(stdout, "所有钩子卸载成功") {
		t.Errorf("expected Chinese all-success message, got: %s", stdout)
	}
}


