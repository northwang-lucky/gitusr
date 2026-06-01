package hook

import (
	"encoding/json"
	"os"
	"path/filepath"

	"github.com/northwang-lucky/gitusr/internal/xdgpath"
)

// LoadState reads hook state from disk.
// Returns empty state if the file does not exist or contains invalid JSON.
func LoadState() (*HookState, error) {
	path, err := stateFilePath()
	if err != nil {
		return nil, err
	}

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return &HookState{InstalledTypes: []HookType{}}, nil
		}
		return nil, err
	}

	if len(data) == 0 {
		return &HookState{InstalledTypes: []HookType{}}, nil
	}

	var state HookState
	if err := json.Unmarshal(data, &state); err != nil {
		return &HookState{InstalledTypes: []HookType{}}, nil
	}

	if state.InstalledTypes == nil {
		state.InstalledTypes = []HookType{}
	}
	if state.DisabledTypes == nil {
		state.DisabledTypes = []HookType{}
	}

	return &state, nil
}

// SaveState writes hook state to disk as JSON.
func SaveState(state *HookState) error {
	path, err := stateFilePath()
	if err != nil {
		return err
	}

	data, err := json.MarshalIndent(state, "", "\t")
	if err != nil {
		return err
	}

	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	return os.WriteFile(path, data, 0644)
}

// IsInstalled checks whether a given hook type is currently installed.
func IsInstalled(t HookType) (bool, error) {
	state, err := LoadState()
	if err != nil {
		return false, err
	}

	for _, installed := range state.InstalledTypes {
		if installed == t {
			return true, nil
		}
	}

	return false, nil
}

// containsHookType checks whether a HookType exists in a slice.
func containsHookType(slice []HookType, t HookType) bool {
	for _, v := range slice {
		if v == t {
			return true
		}
	}
	return false
}

// EnableHook removes a hook type from the disabled list, re-enabling it.
func EnableHook(t HookType) error {
	state, err := LoadState()
	if err != nil {
		return err
	}

	filtered := make([]HookType, 0, len(state.DisabledTypes))
	for _, dt := range state.DisabledTypes {
		if dt != t {
			filtered = append(filtered, dt)
		}
	}
	state.DisabledTypes = filtered

	return SaveState(state)
}

// DisableHook adds a hook type to the disabled list (no duplicates).
func DisableHook(t HookType) error {
	state, err := LoadState()
	if err != nil {
		return err
	}

	if !containsHookType(state.DisabledTypes, t) {
		state.DisabledTypes = append(state.DisabledTypes, t)
	}

	return SaveState(state)
}

// IsEnabled checks whether a hook type is not disabled.
// Returns true when state is empty (no state file or clean state).
func IsEnabled(t HookType) (bool, error) {
	state, err := LoadState()
	if err != nil {
		return false, err
	}

	return !containsHookType(state.DisabledTypes, t), nil
}

// stateFilePath returns the path to hook-state.json based on XDG data directory.
func stateFilePath() (string, error) {
	userListPath, err := xdgpath.DataFilePath()
	if err != nil {
		return "", err
	}
	dir := filepath.Dir(userListPath)
	return filepath.Join(dir, "hook-state.json"), nil
}
