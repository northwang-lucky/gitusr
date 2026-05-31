package hook

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestUninstall_Success(t *testing.T) {
	tmpHome := t.TempDir()
	t.Setenv("HOME", tmpHome)
	tmpData := t.TempDir()
	t.Setenv("XDG_DATA_HOME", tmpData)

	// Prepare: create hooks dir and wrapper file
	hooksDir := filepath.Join(tmpData, "gitusr", "hooks")
	if err := os.MkdirAll(hooksDir, 0755); err != nil {
		t.Fatal(err)
	}
	wrapperPath := filepath.Join(hooksDir, "git-wrapper.sh")
	if err := os.WriteFile(wrapperPath, []byte("#!/bin/bash\necho 'wrapper'"), 0644); err != nil {
		t.Fatal(err)
	}

	// Prepare: create .bashrc with hook block
	bashrcPath := filepath.Join(tmpHome, ".bashrc")
	block := "\n# gitusr hook begin\nsource " + wrapperPath + "\n# gitusr hook end\n"
	if err := os.WriteFile(bashrcPath, []byte(block), 0644); err != nil {
		t.Fatal(err)
	}

	// Prepare: save state with HookTypeClone installed
	state := &HookState{InstalledTypes: []HookType{HookTypeClone}}
	if err := SaveState(state); err != nil {
		t.Fatal(err)
	}

	// Execute
	if err := Uninstall(HookTypeClone, []ShellType{ShellTypeBash}); err != nil {
		t.Fatalf("Uninstall() returned error: %v", err)
	}

	// Verify: block removed from .bashrc
	data, err := os.ReadFile(bashrcPath)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(data), "# gitusr hook begin") {
		t.Error("expected hook block to be removed from .bashrc")
	}

	// Verify: state updated (HookTypeClone removed)
	loaded, err := LoadState()
	if err != nil {
		t.Fatal(err)
	}
	if len(loaded.InstalledTypes) != 0 {
		t.Errorf("expected empty InstalledTypes, got %v", loaded.InstalledTypes)
	}
}

func TestUninstall_NotInstalled(t *testing.T) {
	tmpData := t.TempDir()
	t.Setenv("XDG_DATA_HOME", tmpData)

	// Prepare: save state with empty InstalledTypes
	state := &HookState{InstalledTypes: []HookType{}}
	if err := SaveState(state); err != nil {
		t.Fatal(err)
	}

	// Execute
	err := Uninstall(HookTypeClone, []ShellType{ShellTypeBash})

	// Verify: error returned
	if err == nil {
		t.Fatal("expected error for not installed hook")
	}
	if !strings.Contains(err.Error(), "not installed") {
		t.Errorf("error should contain 'not installed', got: %v", err)
	}
}

func TestUninstall_LastType(t *testing.T) {
	tmpHome := t.TempDir()
	t.Setenv("HOME", tmpHome)
	tmpData := t.TempDir()
	t.Setenv("XDG_DATA_HOME", tmpData)

	// Prepare: create hooks dir and wrapper files for both shells
	hooksDir := filepath.Join(tmpData, "gitusr", "hooks")
	if err := os.MkdirAll(hooksDir, 0755); err != nil {
		t.Fatal(err)
	}
	bashWrapper := filepath.Join(hooksDir, "git-wrapper.sh")
	zshWrapper := filepath.Join(hooksDir, "git-wrapper.zsh")
	if err := os.WriteFile(bashWrapper, []byte("#!/bin/bash"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(zshWrapper, []byte("#!/bin/zsh"), 0644); err != nil {
		t.Fatal(err)
	}

	// Prepare: create .bashrc with hook block
	bashrcPath := filepath.Join(tmpHome, ".bashrc")
	block := "\n# gitusr hook begin\nsource " + bashWrapper + "\n# gitusr hook end\n"
	if err := os.WriteFile(bashrcPath, []byte(block), 0644); err != nil {
		t.Fatal(err)
	}

	// Prepare: save state with HookTypeClone as the only installed type
	state := &HookState{InstalledTypes: []HookType{HookTypeClone}}
	if err := SaveState(state); err != nil {
		t.Fatal(err)
	}

	// Execute
	if err := Uninstall(HookTypeClone, []ShellType{ShellTypeBash}); err != nil {
		t.Fatalf("Uninstall() returned error: %v", err)
	}

	// Verify: block removed from .bashrc
	data, err := os.ReadFile(bashrcPath)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(data), "# gitusr hook begin") {
		t.Error("expected hook block to be removed from .bashrc")
	}

	// Verify: state empty
	loaded, err := LoadState()
	if err != nil {
		t.Fatal(err)
	}
	if len(loaded.InstalledTypes) != 0 {
		t.Errorf("expected empty InstalledTypes, got %v", loaded.InstalledTypes)
	}

	// Verify: wrapper files deleted (both shells)
	if _, err := os.Stat(bashWrapper); !os.IsNotExist(err) {
		t.Error("expected git-wrapper.sh to be deleted")
	}
	if _, err := os.Stat(zshWrapper); !os.IsNotExist(err) {
		t.Error("expected git-wrapper.zsh to be deleted")
	}
}
