package hook

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/northwang-lucky/gitusr/internal/domain"
	"github.com/northwang-lucky/gitusr/internal/gitcmd"
)

func TestGenerateEnv_Bash(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("XDG_DATA_HOME", tmpDir)

	state := &HookState{
		InstalledTypes: []HookType{HookTypeClone},
	}
	if err := SaveState(state); err != nil {
		t.Fatalf("SaveState() returned error: %v", err)
	}

	script, err := GenerateEnv(ShellTypeBash)
	if err != nil {
		t.Fatalf("GenerateEnv(Bash) returned error: %v", err)
	}
	if !strings.Contains(script, "alias cd=") {
		t.Error("expected bash env to contain cd alias")
	}
	if !strings.Contains(script, "__gitusrcd") {
		t.Error("expected bash env to contain __gitusrcd")
	}
}

func TestGenerateEnv_Zsh(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("XDG_DATA_HOME", tmpDir)

	state := &HookState{
		InstalledTypes: []HookType{HookTypeCommit},
	}
	if err := SaveState(state); err != nil {
		t.Fatalf("SaveState() returned error: %v", err)
	}

	script, err := GenerateEnv(ShellTypeZsh)
	if err != nil {
		t.Fatalf("GenerateEnv(Zsh) returned error: %v", err)
	}
	if !strings.Contains(script, "chpwd") {
		t.Error("expected zsh env to contain chpwd hook")
	}
	if !strings.Contains(script, "__gitusr_autoload_hook") {
		t.Error("expected zsh env to contain __gitusr_autoload_hook")
	}
}

func TestGenerateEnv_InvalidShell(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("XDG_DATA_HOME", tmpDir)

	state := &HookState{
		InstalledTypes: []HookType{HookTypeClone},
	}
	if err := SaveState(state); err != nil {
		t.Fatalf("SaveState() returned error: %v", err)
	}

	_, err := GenerateEnv(ShellType("fish"))
	if err == nil {
		t.Fatal("expected error for invalid shell type")
	}
	if !strings.Contains(err.Error(), "unsupported shell type") {
		t.Errorf("error should mention unsupported shell type, got: %q", err.Error())
	}
}

func TestGenerateEnv_NoHooksInstalled(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("XDG_DATA_HOME", tmpDir)

	// No state file means no hooks installed
	_, err := GenerateEnv(ShellTypeBash)
	if err == nil {
		t.Fatal("expected error when no hooks are installed")
	}
	if !strings.Contains(err.Error(), "请先执行 hook install") {
		t.Errorf("error should contain '请先执行 hook install', got: %q", err.Error())
	}
}

func TestApplyRC_WithRC(t *testing.T) {
	dir := t.TempDir()
	initGitRepo(t, dir)
	t.Chdir(dir)

	rcPath := filepath.Join(dir, ".gitusrrc")
	if err := os.WriteFile(rcPath, []byte(`{"name":"Alice","email":"alice@example.com"}`), 0644); err != nil {
		t.Fatal(err)
	}

	store := &hookTestStore{
		users: []domain.User{
			{Name: "Alice", Email: "alice@example.com"},
		},
	}

	if err := GenerateApplyRC(store); err != nil {
		t.Fatalf("GenerateApplyRC() returned error: %v", err)
	}

	name, err := gitcmd.GetConfig("name", false)
	if err != nil {
		t.Fatalf("git config user.name: %v", err)
	}
	if name != "Alice" {
		t.Errorf("user.name = %q, want %q", name, "Alice")
	}
}

func TestApplyRC_NoRC(t *testing.T) {
	dir := t.TempDir()
	t.Chdir(dir)

	store := &hookTestStore{
		users: []domain.User{
			{Name: "Bob", Email: "bob@example.com"},
		},
	}

	if err := GenerateApplyRC(store); err != nil {
		t.Fatalf("GenerateApplyRC() should return nil when no .gitusrrc, got: %v", err)
	}
}
