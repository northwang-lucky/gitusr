package hook

import (
	"fmt"
	"os"
	"path/filepath"
)

// Uninstall removes a specific hook type from the specified shell configurations.
// It checks if the hook type is installed, removes the hook block from each shell
// config file, updates the persistent state, and cleans up wrapper files.
//
// Per-hook cleanup:
//   - CD hook: removes cd source blocks and cd-env.* wrapper files immediately
//   - When no hook types remain: removes main source blocks and all remaining wrapper files
func Uninstall(hookType HookType, shells []ShellType) error {
	installed, err := IsInstalled(hookType)
	if err != nil {
		return err
	}
	if !installed {
		return fmt.Errorf("hook %s is not installed", hookType)
	}

	// Update state - remove hookType from InstalledTypes
	state, err := LoadState()
	if err != nil {
		return err
	}

	var updated []HookType
	for _, t := range state.InstalledTypes {
		if t != hookType {
			updated = append(updated, t)
		}
	}
	state.InstalledTypes = updated

	if err := SaveState(state); err != nil {
		return err
	}

	// Remove CD-specific source blocks and wrapper files immediately when uninstalling CD
	if hookType == HookTypeCD {
		for _, shell := range shells {
			if err := RemoveCDSourceBlock(shell); err != nil {
				return err
			}
		}
		if err := deleteCDWrapperFiles(); err != nil {
			return err
		}
	}

	// Clean up remaining artifacts when no hook types are left
	if len(updated) == 0 {
		for _, shell := range shells {
			if err := RemoveSourceBlock(shell); err != nil {
				return err
			}
		}
		if err := deleteWrapperFiles(); err != nil {
			return err
		}
	}

	return nil
}

// deleteCDWrapperFiles removes cd-env scripts from the hooks directory.
// Non-existent files are silently skipped.
func deleteCDWrapperFiles() error {
	dir, err := hooksDir()
	if err != nil {
		return err
	}

	for _, ext := range []string{"sh", "zsh"} {
		fp := filepath.Join(dir, wrapperFileName(HookTypeCD, ext))
		if err := os.Remove(fp); err != nil && !os.IsNotExist(err) {
			return err
		}
	}

	return nil
}

// UninstallAll removes all hook types, source blocks, wrapper files, and clears
// the hook state entirely. It is the counterpart to InstallAll.
//
// For each shell it:
//  1. Removes the main source block (hooks for clone/commit)
//  2. Removes the CD source block (if it exists from legacy installs)
//  3. Deletes all wrapper files
//  4. Clears hook state (InstalledTypes and DisabledTypes both set to [])
//
// If no hook types are currently installed, it returns an error.
//
// Deprecated: Uninstall is kept for backward compatibility with per-type uninstall.
func UninstallAll(shells []ShellType) error {
	state, err := LoadState()
	if err != nil {
		return err
	}

	// Check if ANY hook type is installed
	if len(state.InstalledTypes) == 0 {
		return fmt.Errorf("no hooks are currently installed")
	}

	// Remove source blocks from all shell configs
	for _, shell := range shells {
		_ = RemoveSourceBlock(shell)
		_ = RemoveCDSourceBlock(shell)
	}

	// Delete all wrapper files (both git-wrapper and cd-env)
	if err := deleteWrapperFiles(); err != nil {
		return err
	}

	// Clear state entirely
	state.InstalledTypes = []HookType{}
	state.DisabledTypes = []HookType{}
	return SaveState(state)
}

// deleteWrapperFiles removes all git-wrapper scripts from the hooks directory.
// Non-existent files are silently skipped.
func deleteWrapperFiles() error {
	dir, err := hooksDir()
	if err != nil {
		return err
	}

	for _, ext := range []string{"sh", "zsh"} {
		fp := filepath.Join(dir, fmt.Sprintf("git-wrapper.%s", ext))
		if err := os.Remove(fp); err != nil && !os.IsNotExist(err) {
			return err
		}

		fp = filepath.Join(dir, wrapperFileName(HookTypeCD, ext))
		if err := os.Remove(fp); err != nil && !os.IsNotExist(err) {
			return err
		}
	}

	return nil
}
