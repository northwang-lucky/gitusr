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

// =============================================================================
// UninstallAll tests
// =============================================================================

// TestUninstallAll_Success verifies the full uninstall flow:
// 1. Install all hooks via InstallAll
// 2. Uninstall via UninstallAll
// 3. Verify all RC blocks removed, wrapper files deleted, state cleared
func TestUninstallAll_Success(t *testing.T) {
	tmpHome := t.TempDir()
	t.Setenv("HOME", tmpHome)
	tmpData := t.TempDir()
	t.Setenv("XDG_DATA_HOME", tmpData)

	// Install all hooks via InstallAll (bash + zsh)
	results, err := InstallAll([]ShellType{ShellTypeBash, ShellTypeZsh})
	if err != nil {
		t.Fatalf("InstallAll() returned error: %v", err)
	}
	if len(results) != 2 {
		t.Fatalf("expected 2 results from InstallAll, got %d", len(results))
	}

	// Record wrapper paths for verification
	bashWrapper := results[0].FilePath
	zshWrapper := results[1].FilePath
	if results[0].Shell == ShellTypeZsh {
		bashWrapper = results[1].FilePath
		zshWrapper = results[0].FilePath
	}

	// Verify initial state has all 3 types
	state, err := LoadState()
	if err != nil {
		t.Fatalf("LoadState() before uninstall: %v", err)
	}
	if len(state.InstalledTypes) != 3 {
		t.Fatalf("expected 3 installed types before uninstall, got %d", len(state.InstalledTypes))
	}

	// Execute UninstallAll
	if err := UninstallAll([]ShellType{ShellTypeBash, ShellTypeZsh}); err != nil {
		t.Fatalf("UninstallAll() returned error: %v", err)
	}

	// Verify: hook blocks removed from .bashrc
	bashrcPath := filepath.Join(tmpHome, ".bashrc")
	bashrcData, err := os.ReadFile(bashrcPath)
	if err != nil {
		t.Fatal(err)
	}
	bashrcContent := string(bashrcData)
	if strings.Contains(bashrcContent, "# gitusr hook begin") {
		t.Error("expected '# gitusr hook begin' marker to be removed from .bashrc")
	}
	if strings.Contains(bashrcContent, "# gitusr cd begin") {
		t.Error("expected '# gitusr cd begin' marker to be removed from .bashrc")
	}

	// Verify: hook blocks removed from .zshrc
	zshrcPath := filepath.Join(tmpHome, ".zshrc")
	zshrcData, err := os.ReadFile(zshrcPath)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(zshrcData), "# gitusr hook begin") {
		t.Error("expected '# gitusr hook begin' marker to be removed from .zshrc")
	}

	// Verify: wrapper files deleted (both shells)
	for _, wp := range []string{bashWrapper, zshWrapper} {
		if _, err := os.Stat(wp); !os.IsNotExist(err) {
			t.Errorf("expected wrapper file %s to be deleted", wp)
		}
	}

	// Verify: cd-env wrapper files also cleaned
	hooksDir := filepath.Join(tmpData, "gitusr", "hooks")
	for _, ext := range []string{"sh", "zsh"} {
		cdPath := filepath.Join(hooksDir, "cd-env."+ext)
		if _, err := os.Stat(cdPath); !os.IsNotExist(err) {
			t.Errorf("expected cd-env.%s to be deleted", ext)
		}
	}

	// Verify: state cleared (no installed types, no disabled types)
	state, err = LoadState()
	if err != nil {
		t.Fatalf("LoadState() after uninstall: %v", err)
	}
	if len(state.InstalledTypes) != 0 {
		t.Errorf("expected empty InstalledTypes, got %v", state.InstalledTypes)
	}
	if len(state.DisabledTypes) != 0 {
		t.Errorf("expected empty DisabledTypes, got %v", state.DisabledTypes)
	}
}

// TestUninstallAll_NotInstalled verifies UninstallAll returns an error
// when no hooks are currently installed.
func TestUninstallAll_NotInstalled(t *testing.T) {
	tmpData := t.TempDir()
	t.Setenv("XDG_DATA_HOME", tmpData)

	// Save empty state
	state := &HookState{InstalledTypes: []HookType{}}
	if err := SaveState(state); err != nil {
		t.Fatal(err)
	}

	// Execute: should return error since nothing is installed
	err := UninstallAll([]ShellType{ShellTypeBash})

	// Verify error
	if err == nil {
		t.Fatal("expected error when no hooks are installed")
	}
	if !strings.Contains(err.Error(), "no hooks are currently installed") {
		t.Errorf("error should contain 'no hooks are currently installed', got: %v", err)
	}
}

// TestUninstallAll_PartialState verifies full cleanup when only some
// hooks were installed via the legacy Install (not InstallAll).
// This simulates migrating from old hook installs to the new unified approach.
func TestUninstallAll_PartialState(t *testing.T) {
	tmpHome := t.TempDir()
	t.Setenv("HOME", tmpHome)
	tmpData := t.TempDir()
	t.Setenv("XDG_DATA_HOME", tmpData)

	// Install only HookTypeClone via legacy Install (creates git-wrapper.sh)
	_, err := Install(HookTypeClone, []ShellType{ShellTypeBash})
	if err != nil {
		t.Fatalf("Install(HookTypeClone): %v", err)
	}

	// Install HookTypeCD via legacy Install (creates cd-env.sh + CD source block)
	_, err = Install(HookTypeCD, []ShellType{ShellTypeBash})
	if err != nil {
		t.Fatalf("Install(HookTypeCD): %v", err)
	}

	// Verify initial state has both types
	state, err := LoadState()
	if err != nil {
		t.Fatalf("LoadState() before uninstall: %v", err)
	}
	if len(state.InstalledTypes) != 2 {
		t.Fatalf("expected 2 installed types, got %d: %v", len(state.InstalledTypes), state.InstalledTypes)
	}

	// Execute UninstallAll — should clean up everything
	if err := UninstallAll([]ShellType{ShellTypeBash}); err != nil {
		t.Fatalf("UninstallAll() returned error: %v", err)
	}

	// Verify: both hook blocks removed from .bashrc
	bashrcPath := filepath.Join(tmpHome, ".bashrc")
	bashrcData, err := os.ReadFile(bashrcPath)
	if err != nil {
		t.Fatal(err)
	}
	bashrcContent := string(bashrcData)
	if strings.Contains(bashrcContent, "# gitusr hook begin") {
		t.Error("expected '# gitusr hook begin' marker to be removed from .bashrc")
	}
	if strings.Contains(bashrcContent, "# gitusr cd begin") {
		t.Error("expected '# gitusr cd begin' marker to be removed from .bashrc")
	}

	// Verify: all wrapper files deleted
	hooksDir := filepath.Join(tmpData, "gitusr", "hooks")
	for _, ext := range []string{"sh", "zsh"} {
		gwPath := filepath.Join(hooksDir, "git-wrapper."+ext)
		if _, err := os.Stat(gwPath); !os.IsNotExist(err) {
			t.Errorf("expected git-wrapper.%s to be deleted", ext)
		}
		cdPath := filepath.Join(hooksDir, "cd-env."+ext)
		if _, err := os.Stat(cdPath); !os.IsNotExist(err) {
			t.Errorf("expected cd-env.%s to be deleted", ext)
		}
	}

	// Verify: state fully cleared
	state, err = LoadState()
	if err != nil {
		t.Fatalf("LoadState() after uninstall: %v", err)
	}
	if len(state.InstalledTypes) != 0 {
		t.Errorf("expected empty InstalledTypes, got %v", state.InstalledTypes)
	}
	if len(state.DisabledTypes) != 0 {
		t.Errorf("expected empty DisabledTypes, got %v", state.DisabledTypes)
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

	// Prepare .bashrc with clone source line in hook block and CD source line in CD block
	bashrcPath := filepath.Join(tmpHome, ".bashrc")
	block := fmt.Sprintf("\n# gitusr hook begin\nsource %s\n# gitusr hook end\n\n# gitusr cd begin\nsource %s\n# gitusr cd end\n", gitWrapperPath, cdEnvPath)
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
