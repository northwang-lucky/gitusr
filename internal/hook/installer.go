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

var cdShellGenerators = map[ShellType]func() string{
	ShellTypeBash: GenerateUnifiedBashWrapper,
	ShellTypeZsh:  GenerateUnifiedZshWrapper,
}

// Install installs hook wrappers for the specified hook type and shells.
// It is idempotent — if the hook type is already installed, it returns early.
// For each shell, it generates the wrapper code, writes it to disk via
// WriteWrapperFile, and appends a source line to the shell config via
// AppendSourceLine.
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

	generators := shellGenerators
	if hookType == HookTypeCD {
		generators = cdShellGenerators
	}

	for _, shell := range shells {
		generate, ok := generators[shell]
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
		if hookType == HookTypeCD {
			err = AppendCDSourceLine(shell, filePath)
		} else {
			err = AppendSourceLine(shell, filePath)
		}
		if err != nil {
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
