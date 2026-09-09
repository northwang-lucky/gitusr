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

// writeHookState creates a temp XDG_DATA_HOME holding the given hook state
// so that hook.LoadState reads it, mirroring setupHookState's sandboxing.
func writeHookState(t *testing.T, installed, disabled []hook.HookType) {
	t.Helper()

	dir := t.TempDir()
	xdgData := filepath.Join(dir, "gitusr")
	if err := os.MkdirAll(xdgData, 0755); err != nil {
		t.Fatalf("failed to create xdg data dir: %v", err)
	}

	data, err := json.MarshalIndent(hook.HookState{
		InstalledTypes: installed,
		DisabledTypes:  disabled,
	}, "", "\t")
	if err != nil {
		t.Fatalf("failed to marshal state: %v", err)
	}
	if err := os.WriteFile(filepath.Join(xdgData, "hook-state.json"), data, 0644); err != nil {
		t.Fatalf("failed to write state: %v", err)
	}

	t.Setenv("XDG_DATA_HOME", dir)
}

func TestHooksStatusCmd_MixedStates_En(t *testing.T) {
	i18n.ResetForTesting()
	i18n.InitWithLocale("en")
	t.Cleanup(i18n.ResetForTesting)

	writeHookState(t,
		[]hook.HookType{hook.HookTypeClone, hook.HookTypeCommit},
		[]hook.HookType{hook.HookTypeCommit})

	stdout, _, err := executeCmd(NewHooksStatusCmd())
	if err != nil {
		t.Fatalf("hooks status failed: %v", err)
	}

	for _, want := range []string{"clone: enabled", "commit: disabled", "cd: not installed"} {
		if !strings.Contains(stdout, want) {
			t.Errorf("expected line %q in output, got: %q", want, stdout)
		}
	}
}

// 未安装的 hook 即使出现在 disabled 列表里，也应展示为未安装而非已禁用。
func TestHooksStatusCmd_NotInstalledWins(t *testing.T) {
	i18n.ResetForTesting()
	i18n.InitWithLocale("en")
	t.Cleanup(i18n.ResetForTesting)

	writeHookState(t, nil, []hook.HookType{hook.HookTypeClone})

	stdout, _, err := executeCmd(NewHooksStatusCmd())
	if err != nil {
		t.Fatalf("hooks status failed: %v", err)
	}
	if !strings.Contains(stdout, "clone: not installed") {
		t.Errorf("expected clone as not installed, got: %q", stdout)
	}
}

func TestHooksStatusCmd_AllStates_WhenNothingInstalled(t *testing.T) {
	i18n.ResetForTesting()
	i18n.InitWithLocale("en")
	t.Cleanup(i18n.ResetForTesting)

	writeHookState(t, nil, nil)

	stdout, _, err := executeCmd(NewHooksStatusCmd())
	if err != nil {
		t.Fatalf("hooks status failed: %v", err)
	}
	if got := strings.Count(stdout, "not installed"); got != 3 {
		t.Errorf("expected 3 not-installed lines, got %d in: %q", got, stdout)
	}
}

func TestHooksStatusCmd_MixedStates_ZhCN(t *testing.T) {
	i18n.ResetForTesting()
	i18n.InitWithLocale("zh-CN")
	t.Cleanup(i18n.ResetForTesting)

	writeHookState(t,
		[]hook.HookType{hook.HookTypeClone, hook.HookTypeCommit, hook.HookTypeCD},
		[]hook.HookType{hook.HookTypeCD})

	stdout, _, err := executeCmd(NewHooksStatusCmd())
	if err != nil {
		t.Fatalf("hooks status failed: %v", err)
	}

	for _, want := range []string{"clone: 已启用", "commit: 已启用", "cd: 已禁用"} {
		if !strings.Contains(stdout, want) {
			t.Errorf("expected line %q in output, got: %q", want, stdout)
		}
	}
}
