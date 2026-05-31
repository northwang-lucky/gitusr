package hook

import (
	"fmt"
	"os"
	"path/filepath"
)

// Uninstall removes a specific hook type from the specified shell configurations.
// It checks if the hook type is installed, removes the hook block from each shell
// config file, updates the persistent state, and cleans up wrapper files when no
// hook types remain installed.
func Uninstall(hookType HookType, shells []ShellType) error {
	installed, err := IsInstalled(hookType)
	if err != nil {
		return err
	}
	if !installed {
		return fmt.Errorf("hook %s is not installed", hookType)
	}

	for _, shell := range shells {
		if err := RemoveSourceBlock(shell); err != nil {
			return err
		}
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

	// If no hook types remain installed, clean up all wrapper files
	if len(state.InstalledTypes) == 0 {
		if err := deleteWrapperFiles(); err != nil {
			return err
		}
	}

	return nil
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
	}

	return nil
}
