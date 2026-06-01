package hook

import (
	"fmt"
)

// shellGenerators maps shell types to their unified wrapper generation functions.
// Each unified wrapper contains both git() (clone+commit) and cd/chpwd hook
// behavior in a single script with disabled hook checks.
var shellGenerators = map[ShellType]func() string{
	ShellTypeBash: GenerateUnifiedBashWrapper,
	ShellTypeZsh:  GenerateUnifiedZshWrapper,
}

// Install installs hook wrappers for the specified hook type and shells.
// It is idempotent — if the hook type is already installed, it returns early.
// For each shell, it generates the wrapper code, writes it to disk via
// WriteWrapperFile, and appends a source line to the shell config via
// AppendSourceLine.
//
// Deprecated: Use InstallAll instead to install all hook types at once
// with a single unified wrapper file.
func Install(hookType HookType, shells []ShellType) ([]HookInstallResult, error) {
	// Check if already installed (idempotent)
	installed, err := IsInstalled(hookType)
	if err != nil {
		return nil, fmt.Errorf("check installed status: %w", err)
	}
	if installed {
		return nil, nil
	}

	var results []HookInstallResult

	for _, shell := range shells {
		generate, ok := shellGenerators[shell]
		if !ok {
			return results, fmt.Errorf("unsupported shell type: %s", shell)
		}

		// Generate wrapper code
		code := generate()

		// Write wrapper file to disk
		filePath, err := WriteWrapperFile(hookType, shell, code)
		if err != nil {
			return results, fmt.Errorf("write wrapper for %s: %w", shell, err)
		}

		// Append source line to shell config
		if err := AppendSourceLine(shell, filePath); err != nil {
			return results, fmt.Errorf("update shell config for %s: %w", shell, err)
		}

		results = append(results, HookInstallResult{
			Type:     hookType,
			Shell:    shell,
			FilePath: filePath,
		})
	}

	// Update state to mark the hook type as installed
	state, err := LoadState()
	if err != nil {
		return results, fmt.Errorf("load state: %w", err)
	}

	// Double-check not already in state (defensive)
	alreadyInState := false
	for _, t := range state.InstalledTypes {
		if t == hookType {
			alreadyInState = true
			break
		}
	}

	if !alreadyInState {
		state.InstalledTypes = append(state.InstalledTypes, hookType)
		if err := SaveState(state); err != nil {
			return results, fmt.Errorf("save state: %w", err)
		}
	}

	return results, nil
}

// InstallAll installs a single unified hook wrapper that covers all three hook
// types (clone, commit, cd) for the specified shells. The unified wrapper
// contains both git() and cd handler logic in one file, written as
// git-wrapper.{sh,zsh}.
//
// It is idempotent — if ALL three hook types are already installed, it returns
// nil results without making any changes.
//
// For each shell it:
//  1. Generates the unified wrapper code
//  2. Writes it to git-wrapper.{sh,zsh} via WriteWrapperFile
//  3. Appends a source line to the shell rc via AppendSourceLine
//  4. Saves state with all three types installed and none disabled
func InstallAll(shells []ShellType) ([]HookInstallResult, error) {
	// Check if ALL 3 hook types are already installed (idempotent)
	allInstalled := true
	for _, ht := range AllHookTypes {
		installed, err := IsInstalled(ht)
		if err != nil {
			return nil, fmt.Errorf("check installed status for %s: %w", ht, err)
		}
		if !installed {
			allInstalled = false
			break
		}
	}
	if allInstalled {
		return nil, nil
	}

	var results []HookInstallResult

	for _, shell := range shells {
		generate, ok := shellGenerators[shell]
		if !ok {
			return results, fmt.Errorf("unsupported shell type: %s", shell)
		}

		// Generate unified wrapper code (covers clone, commit, and cd)
		code := generate()

		// Write as git-wrapper.{sh,zsh} (unified wrapper, not cd-env)
		filePath, err := WriteWrapperFile(HookTypeClone, shell, code)
		if err != nil {
			return results, fmt.Errorf("write wrapper for %s: %w", shell, err)
		}

		// Append single source line — unified wrapper handles all hook types
		if err := AppendSourceLine(shell, filePath); err != nil {
			return results, fmt.Errorf("update shell config for %s: %w", shell, err)
		}

		results = append(results, HookInstallResult{
			Type:     HookTypeClone, // unified wrapper covers all types
			Shell:    shell,
			FilePath: filePath,
		})
	}

	// Update state: all three types installed, none disabled
	state, err := LoadState()
	if err != nil {
		return results, fmt.Errorf("load state: %w", err)
	}

	for _, ht := range AllHookTypes {
		if !containsHookType(state.InstalledTypes, ht) {
			state.InstalledTypes = append(state.InstalledTypes, ht)
		}
	}
	state.DisabledTypes = nil

	if err := SaveState(state); err != nil {
		return results, fmt.Errorf("save state: %w", err)
	}

	return results, nil
}
