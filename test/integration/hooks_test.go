package integration

import (
	"os"
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
// TestHooksInstallUninstall — full install→verify→uninstall→verify cycle
// ---------------------------------------------------------------------------
func TestHooksInstallUninstall(t *testing.T) {
	homeDir := t.TempDir()
	xdgDataHome := t.TempDir()

	env := map[string]string{
		"HOME":          homeDir,
		"XDG_DATA_HOME": xdgDataHome,
	}

	// Initialize store with users (required for hook install)
	storePath := gitusrStore(xdgDataHome)
	users := []map[string]string{
		{"name": "Alice", "email": "alice@example.com"},
		{"name": "Bob", "email": "bob@example.com"},
	}
	writeJSONFile(t, storePath, users)

	// Step 1: Install all hooks
	output, err := runGitusr(t, env, "hooks", "install")
	if err != nil {
		t.Fatalf("hooks install failed: %v\noutput: %s", err, output)
	}
	if !strings.Contains(output, "All hooks successfully installed") {
		t.Errorf("install output should contain 'All hooks successfully installed', got: %s", output)
	}

	// Step 2: Verify .bashrc and .zshrc contain hook markers
	bashrcPath := filepath.Join(homeDir, ".bashrc")
	zshrcPath := filepath.Join(homeDir, ".zshrc")
	for _, rcPath := range []string{bashrcPath, zshrcPath} {
		rc, err := os.ReadFile(rcPath)
		if err != nil {
			t.Fatalf("read %s after install: %v", rcPath, err)
		}
		if !strings.Contains(string(rc), "# gitusr hook begin") {
			t.Errorf("%s should contain hook marker after install, got:\n%s", rcPath, string(rc))
		}
	}

	// Step 3: Verify wrapper file exists
	wrapperPath := filepath.Join(xdgDataHome, "gitusr", "hooks", "git-wrapper.sh")
	if _, err := os.Stat(wrapperPath); os.IsNotExist(err) {
		t.Fatalf("wrapper file does not exist at %s", wrapperPath)
	}

	// Step 4: Verify hook-state.json contains all three installed types
	statePath := filepath.Join(xdgDataHome, "gitusr", "hook-state.json")
	var state hook.HookState
	readJSONFile(t, statePath, &state)
	for _, ht := range []hook.HookType{hook.HookTypeClone, hook.HookTypeCommit, hook.HookTypeCD} {
		found := false
		for _, it := range state.InstalledTypes {
			if it == ht {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("hook-state.json InstalledTypes should contain %q, got: %v", ht, state.InstalledTypes)
		}
	}
	if len(state.DisabledTypes) != 0 {
		t.Errorf("hook-state.json DisabledTypes should be empty after install, got: %v", state.DisabledTypes)
	}

	// Step 5: Uninstall all hooks
	output, err = runGitusr(t, env, "hooks", "uninstall")
	if err != nil {
		t.Fatalf("hooks uninstall failed: %v\noutput: %s", err, output)
	}
	if !strings.Contains(output, "All hooks successfully uninstalled") {
		t.Errorf("uninstall output should contain 'All hooks successfully uninstalled', got: %s", output)
	}

	// Step 6: Verify .bashrc and .zshrc no longer contain hook markers
	for _, rcPath := range []string{bashrcPath, zshrcPath} {
		rc, err := os.ReadFile(rcPath)
		if err != nil {
			t.Fatalf("read %s after uninstall: %v", rcPath, err)
		}
		if strings.Contains(string(rc), "# gitusr hook begin") {
			t.Errorf("%s should NOT contain hook marker after uninstall, got:\n%s", rcPath, string(rc))
		}
	}

	// Step 7: Verify wrapper file is removed
	if _, err := os.Stat(wrapperPath); !os.IsNotExist(err) {
		t.Errorf("wrapper file should be removed after uninstall, but exists at %s", wrapperPath)
	}

	// Step 8: Verify hook-state.json has empty InstalledTypes
	readJSONFile(t, statePath, &state)
	if len(state.InstalledTypes) != 0 {
		t.Errorf("hook-state.json InstalledTypes should be empty after uninstall, got: %v", state.InstalledTypes)
	}
	if len(state.DisabledTypes) != 0 {
		t.Errorf("hook-state.json DisabledTypes should be empty after uninstall, got: %v", state.DisabledTypes)
	}
}

// ---------------------------------------------------------------------------
// TestHooksUninstallNotInstalled — uninstall when no hooks exist returns error
// ---------------------------------------------------------------------------
func TestHooksUninstallNotInstalled(t *testing.T) {
	homeDir := t.TempDir()
	xdgDataHome := t.TempDir()

	env := map[string]string{
		"HOME":          homeDir,
		"XDG_DATA_HOME": xdgDataHome,
	}

	// Do NOT install hooks, do NOT create hook state — clean environment
	output, err := runGitusr(t, env, "hooks", "uninstall")
	if err == nil {
		t.Fatalf("hooks uninstall with no hooks installed should return error, got output: %s", output)
	}
	if !strings.Contains(output, "No hooks are currently installed") {
		t.Errorf("error output should contain 'No hooks are currently installed', got: %s", output)
	}
}

// ---------------------------------------------------------------------------
// ---------------------------------------------------------------------------
// TestHooksInstallIdempotent — install twice: second call says "already installed"
// ---------------------------------------------------------------------------
func TestHooksInstallIdempotent(t *testing.T) {
	homeDir := t.TempDir()
	xdgDataHome := t.TempDir()

	env := map[string]string{
		"HOME":          homeDir,
		"XDG_DATA_HOME": xdgDataHome,
	}

	storePath := gitusrStore(xdgDataHome)
	users := []map[string]string{
		{"name": "Alice", "email": "alice@example.com"},
		{"name": "Bob", "email": "bob@example.com"},
	}
	writeJSONFile(t, storePath, users)

	output, err := runGitusr(t, env, "hooks", "install")
	if err != nil {
		t.Fatalf("first hooks install failed: %v\noutput: %s", err, output)
	}

	output, err = runGitusr(t, env, "hooks", "install")
	if err != nil {
		t.Fatalf("second hooks install should exit 0: %v\noutput: %s", err, output)
	}
	if !strings.Contains(strings.ToLower(output), "already installed") {
		t.Errorf("idempotent install should say 'already installed', got: %s", output)
	}
}

// ---------------------------------------------------------------------------
// TestHooksInstallResetsDisabled — install then disable clone, then delete
// installed_types from state, re-install: verify disabled_types is cleared.
// ---------------------------------------------------------------------------
func TestHooksInstallResetsDisabled(t *testing.T) {
	homeDir := t.TempDir()
	xdgDataHome := t.TempDir()

	env := map[string]string{
		"HOME":          homeDir,
		"XDG_DATA_HOME": xdgDataHome,
	}

	storePath := gitusrStore(xdgDataHome)
	users := []map[string]string{
		{"name": "Alice", "email": "alice@example.com"},
		{"name": "Bob", "email": "bob@example.com"},
	}
	writeJSONFile(t, storePath, users)

	output, err := runGitusr(t, env, "hooks", "install")
	if err != nil {
		t.Fatalf("hooks install failed: %v\noutput: %s", err, output)
	}

	output, err = runGitusr(t, env, "hooks", "disable", "clone")
	if err != nil {
		t.Fatalf("hooks disable clone failed: %v\noutput: %s", err, output)
	}

	// Manually remove installed_types from hook state to simulate corruption
	statePath := filepath.Join(xdgDataHome, "gitusr", "hook-state.json")
	var state hook.HookState
	readJSONFile(t, statePath, &state)
	if len(state.DisabledTypes) == 0 {
		t.Fatal("expected disabled_types to contain 'clone' after disable")
	}
	state.InstalledTypes = []hook.HookType{}
	writeJSONFile(t, statePath, state)

	// Re-install — should proceed (not idempotent) because installed_types is empty
	output, err = runGitusr(t, env, "hooks", "install")
	if err != nil {
		t.Fatalf("re-install after clearing installed_types failed: %v\noutput: %s", err, output)
	}

	// Verify disabled_types was cleared by InstallAll
	readJSONFile(t, statePath, &state)
	if len(state.DisabledTypes) != 0 {
		t.Errorf("disabled_types should be empty after re-install, got: %v", state.DisabledTypes)
	}
}

// ---------------------------------------------------------------------------
// TestHooksDisableThenInstallIdempotent — install, disable clone, install
// again: idempotent path must NOT reset disabled_types.
// ---------------------------------------------------------------------------
func TestHooksDisableThenInstallIdempotent(t *testing.T) {
	homeDir := t.TempDir()
	xdgDataHome := t.TempDir()

	env := map[string]string{
		"HOME":          homeDir,
		"XDG_DATA_HOME": xdgDataHome,
	}

	storePath := gitusrStore(xdgDataHome)
	users := []map[string]string{
		{"name": "Alice", "email": "alice@example.com"},
		{"name": "Bob", "email": "bob@example.com"},
	}
	writeJSONFile(t, storePath, users)

	output, err := runGitusr(t, env, "hooks", "install")
	if err != nil {
		t.Fatalf("hooks install failed: %v\noutput: %s", err, output)
	}

	output, err = runGitusr(t, env, "hooks", "disable", "clone")
	if err != nil {
		t.Fatalf("hooks disable clone failed: %v\noutput: %s", err, output)
	}

	// Install again — should be idempotent, NOT reset disabled_types
	output, err = runGitusr(t, env, "hooks", "install")
	if err != nil {
		t.Fatalf("second hooks install should exit 0: %v\noutput: %s", err, output)
	}
	if !strings.Contains(strings.ToLower(output), "already installed") {
		t.Errorf("idempotent install should say 'already installed', got: %s", output)
	}

	// Verify clone is still disabled
	statePath := filepath.Join(xdgDataHome, "gitusr", "hook-state.json")
	var state hook.HookState
	readJSONFile(t, statePath, &state)
	found := false
	for _, dt := range state.DisabledTypes {
		if dt == hook.HookTypeClone {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("disabled_types should still contain 'clone' after idempotent install, got: %v", state.DisabledTypes)
	}
}

// ---------------------------------------------------------------------------
// TestHooksLoadCorruptedState — write malformed hook-state.json, then install:
// verify the process does not panic and succeeds.
// ---------------------------------------------------------------------------
func TestHooksLoadCorruptedState(t *testing.T) {
	homeDir := t.TempDir()
	xdgDataHome := t.TempDir()

	env := map[string]string{
		"HOME":          homeDir,
		"XDG_DATA_HOME": xdgDataHome,
	}

	storePath := gitusrStore(xdgDataHome)
	users := []map[string]string{
		{"name": "Alice", "email": "alice@example.com"},
		{"name": "Bob", "email": "bob@example.com"},
	}
	writeJSONFile(t, storePath, users)

	// Write a corrupted hook-state.json to simulate disk corruption
	statePath := filepath.Join(xdgDataHome, "gitusr", "hook-state.json")
	if err := os.MkdirAll(filepath.Dir(statePath), 0755); err != nil {
		t.Fatalf("create dir for %s: %v", statePath, err)
	}
	if err := os.WriteFile(statePath, []byte("this is not valid json {{{"), 0644); err != nil {
		t.Fatalf("write corrupted state: %v", err)
	}

	// hooks install should not panic — LoadState treats invalid JSON as empty
	output, err := runGitusr(t, env, "hooks", "install")
	if err != nil {
		t.Fatalf("hooks install with corrupted state failed: %v\noutput: %s", err, output)
	}
	if !strings.Contains(strings.ToLower(output), "success") {
		t.Errorf("install with corrupted state should succeed, got: %s", output)
	}
}

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

// ---------------------------------------------------------------------------
// TestHooksWrapperBashContent — verify bash wrapper contains expected code
// ---------------------------------------------------------------------------
func TestHooksWrapperBashContent(t *testing.T) {
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

	// Read the bash wrapper file
	wrapperPath := filepath.Join(xdgDataHome, "gitusr", "hooks", "git-wrapper.sh")
	wrapper, err := os.ReadFile(wrapperPath)
	if err != nil {
		t.Fatalf("read wrapper file: %v", err)
	}

	wrapperContent := string(wrapper)

	// Verify __gitusrcd() function exists (cd hook)
	if !strings.Contains(wrapperContent, "__gitusrcd") {
		t.Error("wrapper should contain __gitusrcd() function")
	}

	// Verify three is-disabled checks (one per hook type: clone, commit, cd)
	disabledChecks := []string{
		"is-disabled clone",
		"is-disabled commit",
		"is-disabled cd",
	}
	for _, check := range disabledChecks {
		if !strings.Contains(wrapperContent, check) {
			t.Errorf("wrapper should contain is-disabled check for %s", check)
		}
	}

	// Verify --gu-name and --gu-email argument extraction
	if !strings.Contains(wrapperContent, "--gu-name") {
		t.Error("wrapper should contain --gu-name argument extraction")
	}
	if !strings.Contains(wrapperContent, "--gu-email") {
		t.Error("wrapper should contain --gu-email argument extraction")
	}

	// Verify single-user pass-through logic (user_count check and -le 1 threshold)
	if !strings.Contains(wrapperContent, "user_count") {
		t.Error("wrapper should contain single-user check variable (user_count)")
	}
	if !strings.Contains(wrapperContent, "-le 1") {
		t.Error("wrapper should contain pass-through threshold (-le 1)")
	}
}

// ---------------------------------------------------------------------------
// TestHooksWrapperZshContent — verify zsh wrapper contains expected code
// ---------------------------------------------------------------------------
func TestHooksWrapperZshContent(t *testing.T) {
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

	// Read the zsh wrapper file
	wrapperPath := filepath.Join(xdgDataHome, "gitusr", "hooks", "git-wrapper.zsh")
	wrapper, err := os.ReadFile(wrapperPath)
	if err != nil {
		t.Fatalf("read wrapper file: %v", err)
	}

	wrapperContent := string(wrapper)

	// Verify add-zsh-hook chpwd (zsh cd hook mechanism)
	if !strings.Contains(wrapperContent, "add-zsh-hook chpwd") {
		t.Error("wrapper should contain add-zsh-hook chpwd")
	}

	// Verify [[ syntax (zsh-specific conditional)
	if !strings.Contains(wrapperContent, "[[") {
		t.Error("wrapper should contain zsh [[ conditional syntax")
	}

	// Verify is-disabled checks (at least one present)
	if !strings.Contains(wrapperContent, "is-disabled") {
		t.Error("wrapper should contain is-disabled checks")
	}
}
