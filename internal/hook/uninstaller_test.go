package hook

import (
	"fmt"
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

// TestUninstall_CD_Cleanup verifies that uninstalling HookTypeCD removes
// cd-env.sh and cleans up the CD source line from .bashrc, while preserving
// git-wrapper.sh for the still-installed clone hook.
func TestUninstall_CD_Cleanup(t *testing.T) {
	tmpHome := t.TempDir()
	t.Setenv("HOME", tmpHome)
	tmpData := t.TempDir()
	t.Setenv("XDG_DATA_HOME", tmpData)

	// Prepare: create hooks dir
	hooksDir := filepath.Join(tmpData, "gitusr", "hooks")
	if err := os.MkdirAll(hooksDir, 0755); err != nil {
		t.Fatal(err)
	}

	// Create git-wrapper.sh (for clone hook)
	gitWrapperPath := filepath.Join(hooksDir, "git-wrapper.sh")
	if err := os.WriteFile(gitWrapperPath, []byte("#!/bin/bash\necho 'clone wrapper'"), 0644); err != nil {
		t.Fatal(err)
	}

	// Create cd-env.sh (for cd hook)
	cdEnvPath := filepath.Join(hooksDir, "cd-env.sh")
	if err := os.WriteFile(cdEnvPath, []byte("#!/bin/bash\necho 'cd env'"), 0644); err != nil {
		t.Fatal(err)
	}

	// Prepare .bashrc with both source lines inside the marked block
	bashrcPath := filepath.Join(tmpHome, ".bashrc")
	block := fmt.Sprintf("\n# gitusr hook begin\nsource %s\nsource %s\n# gitusr hook end\n", gitWrapperPath, cdEnvPath)
	if err := os.WriteFile(bashrcPath, []byte(block), 0644); err != nil {
		t.Fatal(err)
	}

	// Prepare state with both HookTypeClone and HookTypeCD installed
	state := &HookState{InstalledTypes: []HookType{HookTypeClone, HookTypeCD}}
	if err := SaveState(state); err != nil {
		t.Fatal(err)
	}

	// Execute: uninstall only CD
	if err := Uninstall(HookTypeCD, []ShellType{ShellTypeBash}); err != nil {
		t.Fatalf("Uninstall() returned error: %v", err)
	}

	// Verify: git-wrapper.sh still exists (clone is still installed)
	if _, err := os.Stat(gitWrapperPath); os.IsNotExist(err) {
		t.Error("expected git-wrapper.sh to still exist after cd uninstall")
	}

	// Verify: cd-env.sh is deleted after cd uninstall
	if _, err := os.Stat(cdEnvPath); !os.IsNotExist(err) {
		t.Error("expected cd-env.sh to be deleted after cd uninstall")
	}

	// Verify: cd source line removed from .bashrc
	data, err := os.ReadFile(bashrcPath)
	if err != nil {
		t.Fatal(err)
	}
	bashrcContent := string(data)
	if strings.Contains(bashrcContent, cdEnvPath) {
		t.Error("expected cd source line to be removed from .bashrc")
	}
	// Clone source line should remain intact
	if !strings.Contains(bashrcContent, gitWrapperPath) {
		t.Error("expected clone source line to remain in .bashrc")
	}

	// Verify: state only has [clone] after cd uninstall
	loaded, err := LoadState()
	if err != nil {
		t.Fatal(err)
	}
	if len(loaded.InstalledTypes) != 1 || loaded.InstalledTypes[0] != HookTypeClone {
		t.Errorf("expected state to have only [clone], got %v", loaded.InstalledTypes)
	}
}
