package cli

import (
	"errors"
	"strings"
	"testing"

	"github.com/northwang-lucky/gitusr/internal/hook"
	"github.com/northwang-lucky/gitusr/internal/i18n"
)

func TestHookEnv_Bash(t *testing.T) {
	i18n.ResetForTesting()
	i18n.InitWithLocale("en")

	origFn := generateEnvFn
	t.Cleanup(func() { generateEnvFn = origFn })
	generateEnvFn = func(shell hook.ShellType) (string, error) {
		if shell != hook.ShellTypeBash {
			t.Fatalf("unexpected shell: %s", shell)
		}
		return "alias cd=__gitusrcd\n__gitusrcd() { ... }", nil
	}

	store := &mockStore{initialized: true}
	cmd := NewHookEnvCmd(store)
	stdout, _, err := executeCmd(cmd, "--shell", "bash")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(stdout, "alias cd=") {
		t.Errorf("expected bash env to contain cd alias, got %q", stdout)
	}
}

func TestHookEnv_Zsh(t *testing.T) {
	i18n.ResetForTesting()
	i18n.InitWithLocale("en")

	origFn := generateEnvFn
	t.Cleanup(func() { generateEnvFn = origFn })
	generateEnvFn = func(shell hook.ShellType) (string, error) {
		if shell != hook.ShellTypeZsh {
			t.Fatalf("unexpected shell: %s", shell)
		}
		return "chpwd() { __gitusr_autoload_hook }", nil
	}

	store := &mockStore{initialized: true}
	cmd := NewHookEnvCmd(store)
	stdout, _, err := executeCmd(cmd, "--shell", "zsh")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(stdout, "chpwd") {
		t.Errorf("expected zsh env to contain chpwd hook, got %q", stdout)
	}
}

func TestHookEnv_InvalidShell(t *testing.T) {
	i18n.ResetForTesting()
	i18n.InitWithLocale("en")

	origFn := generateEnvFn
	t.Cleanup(func() { generateEnvFn = origFn })
	generateEnvFn = func(shell hook.ShellType) (string, error) {
		return "", errors.New("unsupported shell type")
	}

	store := &mockStore{initialized: true}
	cmd := NewHookEnvCmd(store)
	_, _, err := executeCmd(cmd, "--shell", "fish")
	if err == nil {
		t.Fatal("expected error for invalid shell, got nil")
	}
	if !strings.Contains(err.Error(), "unsupported shell type") {
		t.Errorf("error should mention unsupported shell type, got %q", err.Error())
	}
}

func TestHookEnv_NoHooksInstalled(t *testing.T) {
	i18n.ResetForTesting()
	i18n.InitWithLocale("en")

	origFn := generateEnvFn
	t.Cleanup(func() { generateEnvFn = origFn })
	generateEnvFn = func(shell hook.ShellType) (string, error) {
		return "", errors.New("请先执行 hook install")
	}

	store := &mockStore{initialized: true}
	cmd := NewHookEnvCmd(store)
	_, _, err := executeCmd(cmd, "--shell", "bash")
	if err == nil {
		t.Fatal("expected error when no hooks installed, got nil")
	}
	if !strings.Contains(err.Error(), "请先执行 hook install") {
		t.Errorf("error should prompt for hook install first, got %q", err.Error())
	}
}
