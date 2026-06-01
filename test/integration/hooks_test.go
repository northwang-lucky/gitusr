package integration

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/northwang-lucky/gitusr/internal/hook"
)

// ---------------------------------------------------------------------------
// TestHooksEnableDisable — install then disable/enable a hook, verify state
// ---------------------------------------------------------------------------
func TestHooksEnableDisable(t *testing.T) {
	homeDir := t.TempDir()
	xdgDataHome := t.TempDir()

	env := map[string]string{
		"HOME":          homeDir,
		"XDG_DATA_HOME": xdgDataHome,
	}

	// Initialize store with 2 users (required for hook install)
	storePath := gitusrStore(xdgDataHome)
	users := []map[string]string{
		{"name": "Alice", "email": "alice@example.com"},
		{"name": "Bob", "email": "bob@example.com"},
	}
	writeJSONFile(t, storePath, users)

	// Step 1: Install all hooks (unified)
	output, err := runGitusr(t, env, "hooks", "install")
	if err != nil {
		t.Fatalf("hooks install failed: %v\noutput: %s", err, output)
	}

	// Step 2: Disable cd hook
	output, err = runGitusr(t, env, "hooks", "disable", "cd")
	if err != nil {
		t.Fatalf("hooks disable cd failed: %v\noutput: %s", err, output)
	}
	if !strings.Contains(output, "disabled") {
		t.Errorf("disable output should contain 'disabled', got: %s", output)
	}
	if !strings.Contains(output, "Hook cd disabled") {
		t.Errorf("disable output should contain 'Hook cd disabled', got: %s", output)
	}

	// Step 3: Verify hook-state.json disabled_types contains "cd"
	statePath := filepath.Join(xdgDataHome, "gitusr", "hook-state.json")
	var state hook.HookState
	readJSONFile(t, statePath, &state)
	found := false
	for _, dt := range state.DisabledTypes {
		if dt == hook.HookTypeCD {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("hook-state.json disabled_types should contain 'cd', got: %v", state.DisabledTypes)
	}

	// Step 4: Enable cd hook
	output, err = runGitusr(t, env, "hooks", "enable", "cd")
	if err != nil {
		t.Fatalf("hooks enable cd failed: %v\noutput: %s", err, output)
	}
	if !strings.Contains(output, "enabled") {
		t.Errorf("enable output should contain 'enabled', got: %s", output)
	}
	if !strings.Contains(output, "Hook cd enabled") {
		t.Errorf("enable output should contain 'Hook cd enabled', got: %s", output)
	}

	// Step 5: Verify hook-state.json disabled_types is empty
	readJSONFile(t, statePath, &state)
	if len(state.DisabledTypes) != 0 {
		t.Errorf("hook-state.json disabled_types should be empty after enable, got: %v", state.DisabledTypes)
	}
}

// ---------------------------------------------------------------------------
// TestHooksDisableInvalidType — disable with invalid hook type returns error
// ---------------------------------------------------------------------------
func TestHooksDisableInvalidType(t *testing.T) {
	homeDir := t.TempDir()
	xdgDataHome := t.TempDir()

	env := map[string]string{
		"HOME":          homeDir,
		"XDG_DATA_HOME": xdgDataHome,
	}

	output, err := runGitusr(t, env, "hooks", "disable", "invalid_hook")
	if err == nil {
		t.Fatalf("hooks disable invalid_hook should return error, got output: %s", output)
	}
	if !strings.Contains(strings.ToLower(output), "invalid hook type") {
		t.Errorf("error output should mention 'invalid hook type', got: %s", output)
	}
}

// ---------------------------------------------------------------------------
// TestHooksEnableInvalidType — enable with invalid hook type returns error
// ---------------------------------------------------------------------------
func TestHooksEnableInvalidType(t *testing.T) {
	homeDir := t.TempDir()
	xdgDataHome := t.TempDir()

	env := map[string]string{
		"HOME":          homeDir,
		"XDG_DATA_HOME": xdgDataHome,
	}

	output, err := runGitusr(t, env, "hooks", "enable", "invalid_hook")
	if err == nil {
		t.Fatalf("hooks enable invalid_hook should return error, got output: %s", output)
	}
	if !strings.Contains(strings.ToLower(output), "invalid hook type") {
		t.Errorf("error output should mention 'invalid hook type', got: %s", output)
	}
}

// ---------------------------------------------------------------------------
// TestHooksIsDisabled — hidden is-disabled command exit code and output
// ---------------------------------------------------------------------------
func TestHooksIsDisabled(t *testing.T) {
	homeDir := t.TempDir()
	xdgDataHome := t.TempDir()

	env := map[string]string{
		"HOME":          homeDir,
		"XDG_DATA_HOME": xdgDataHome,
	}

	// Initialize store with 2 users (required for hook install)
	storePath := gitusrStore(xdgDataHome)
	users := []map[string]string{
		{"name": "Alice", "email": "alice@example.com"},
		{"name": "Bob", "email": "bob@example.com"},
	}
	writeJSONFile(t, storePath, users)

	// Install hooks (unified)
	output, err := runGitusr(t, env, "hooks", "install")
	if err != nil {
		t.Fatalf("hooks install failed: %v\noutput: %s", err, output)
	}

	// is-disabled should exit non-zero when the hook is NOT disabled (i.e. enabled)
	_, err = runGitusr(t, env, "hooks", "is-disabled", "cd")
	if err == nil {
		t.Fatal("is-disabled on enabled hook should exit non-zero")
	}

	// Disable the hook
	output, err = runGitusr(t, env, "hooks", "disable", "cd")
	if err != nil {
		t.Fatalf("hooks disable cd failed: %v\noutput: %s", err, output)
	}

	// is-disabled should exit 0 (success) when the hook IS disabled
	_, err = runGitusr(t, env, "hooks", "is-disabled", "cd")
	if err != nil {
		t.Fatalf("is-disabled on disabled hook should exit 0, got error: %v", err)
	}
}
