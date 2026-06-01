package integration

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// ---------------------------------------------------------------------------
// TestHookInstallUninstall_Clone — install/uninstall clone hook flow
// ---------------------------------------------------------------------------
func TestHookInstallUninstall_Clone(t *testing.T) {
	homeDir := t.TempDir()
	xdgDataHome := t.TempDir()

	env := map[string]string{
		"HOME":          homeDir,
		"XDG_DATA_HOME": xdgDataHome,
	}

	// Initialize store with 2 users (required for wrapper to work)
	storePath := gitusrStore(xdgDataHome)
	users := []map[string]string{
		{"name": "Alice", "email": "alice@example.com"},
		{"name": "Bob", "email": "bob@example.com"},
	}
	writeJSONFile(t, storePath, users)

	// Install hooks (unified: clone, commit, cd)
	output, err := runGitusr(t, env, "hooks", "install")
	if err != nil {
		t.Fatalf("hooks install failed: %v\noutput: %s", err, output)
	}
	if !strings.Contains(output, "All hooks successfully installed") {
		t.Errorf("expected 'All hooks successfully installed' in output, got: %s", output)
	}

	// Verify .bashrc contains the hook begin marker
	bashrcPath := filepath.Join(homeDir, ".bashrc")
	bashrc, err := os.ReadFile(bashrcPath)
	if err != nil {
		t.Fatalf("read .bashrc: %v", err)
	}
	if !strings.Contains(string(bashrc), "# gitusr hook begin") {
		t.Errorf(".bashrc should contain '# gitusr hook begin', got:\n%s", string(bashrc))
	}

	// Verify wrapper file exists
	wrapperPath := filepath.Join(xdgDataHome, "gitusr", "hooks", "git-wrapper.sh")
	if _, err := os.Stat(wrapperPath); os.IsNotExist(err) {
		t.Fatalf("wrapper file does not exist at %s", wrapperPath)
	}

	// Uninstall all hooks
	output, err = runGitusr(t, env, "hooks", "uninstall")
	if err != nil {
		t.Fatalf("hooks uninstall failed: %v\noutput: %s", err, output)
	}
	if !strings.Contains(output, "All hooks successfully uninstalled") {
		t.Errorf("expected 'All hooks successfully uninstalled' in output, got: %s", output)
	}

	// Verify .bashrc no longer contains the hook block
	bashrc, err = os.ReadFile(bashrcPath)
	if err != nil {
		t.Fatalf("read .bashrc after uninstall: %v", err)
	}
	if strings.Contains(string(bashrc), "# gitusr hook begin") {
		t.Errorf(".bashrc should NOT contain '# gitusr hook begin' after uninstall, got:\n%s", string(bashrc))
	}
}

// ---------------------------------------------------------------------------
// TestHookInstallUninstall_CD — install/uninstall cd hook flow
// ---------------------------------------------------------------------------
func TestHookInstallUninstall_CD(t *testing.T) {
	homeDir := t.TempDir()
	xdgDataHome := t.TempDir()

	env := map[string]string{
		"HOME":          homeDir,
		"XDG_DATA_HOME": xdgDataHome,
	}

	// Initialize store with 2 users (required for wrapper to work)
	storePath := gitusrStore(xdgDataHome)
	users := []map[string]string{
		{"name": "Alice", "email": "alice@example.com"},
		{"name": "Bob", "email": "bob@example.com"},
	}
	writeJSONFile(t, storePath, users)

	// Install hooks (unified: all types including cd)
	output, err := runGitusr(t, env, "hooks", "install")
	if err != nil {
		t.Fatalf("hooks install failed: %v\noutput: %s", err, output)
	}
	if !strings.Contains(output, "All hooks successfully installed") {
		t.Errorf("expected 'All hooks successfully installed' in output, got: %s", output)
	}

	// Verify .bashrc contains the hook begin marker (CD now uses unified hook markers)
	bashrcPath := filepath.Join(homeDir, ".bashrc")
	bashrc, err := os.ReadFile(bashrcPath)
	if err != nil {
		t.Fatalf("read .bashrc: %v", err)
	}
	if !strings.Contains(string(bashrc), "# gitusr hook begin") {
		t.Errorf(".bashrc should contain '# gitusr hook begin', got:\n%s", string(bashrc))
	}

	// Verify wrapper file exists and contains cd-specific code (unified wrapper uses __gitusrcd for bash)
	wrapperPath := filepath.Join(xdgDataHome, "gitusr", "hooks", "git-wrapper.sh")
	wrapper, err := os.ReadFile(wrapperPath)
	if err != nil {
		t.Fatalf("read wrapper file: %v", err)
	}
	if !strings.Contains(string(wrapper), "__gitusrcd") {
		t.Errorf("wrapper should contain '__gitusrcd', got:\n%s", string(wrapper))
	}

	// Uninstall all hooks
	output, err = runGitusr(t, env, "hooks", "uninstall")
	if err != nil {
		t.Fatalf("hooks uninstall failed: %v\noutput: %s", err, output)
	}
	if !strings.Contains(output, "All hooks successfully uninstalled") {
		t.Errorf("expected 'All hooks successfully uninstalled' in output, got: %s", output)
	}

	// Verify .bashrc no longer contains the hook block
	bashrc, err = os.ReadFile(bashrcPath)
	if err != nil {
		t.Fatalf("read .bashrc after uninstall: %v", err)
	}
	if strings.Contains(string(bashrc), "# gitusr hook begin") {
		t.Errorf(".bashrc should NOT contain '# gitusr hook begin' after uninstall, got:\n%s", string(bashrc))
	}
}

func TestHookInstallUninstall_All(t *testing.T) {
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
	if !strings.Contains(output, "All hooks successfully installed") {
		t.Errorf("hooks install output should contain 'All hooks successfully installed', got: %s", output)
	}

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

	statePath := filepath.Join(xdgDataHome, "gitusr", "hook-state.json")
	state, err := os.ReadFile(statePath)
	if err != nil {
		t.Fatalf("read hook-state.json after install: %v", err)
	}
	for _, hookType := range []string{"clone", "commit", "cd"} {
		if !strings.Contains(string(state), hookType) {
			t.Errorf("hook-state.json should contain %q after install, got:\n%s", hookType, string(state))
		}
	}

	wrapperPath := filepath.Join(xdgDataHome, "gitusr", "hooks", "git-wrapper.sh")
	if _, err := os.Stat(wrapperPath); os.IsNotExist(err) {
		t.Fatalf("wrapper file does not exist at %s after install", wrapperPath)
	}

	output, err = runGitusr(t, env, "hooks", "uninstall")
	if err != nil {
		t.Fatalf("hooks uninstall failed: %v\noutput: %s", err, output)
	}
	if !strings.Contains(output, "All hooks successfully uninstalled") {
		t.Errorf("hooks uninstall output should contain 'All hooks successfully uninstalled', got: %s", output)
	}

	for _, rcPath := range []string{bashrcPath, zshrcPath} {
		rc, err := os.ReadFile(rcPath)
		if err != nil {
			t.Fatalf("read %s after uninstall: %v", rcPath, err)
		}
		if strings.Contains(string(rc), "# gitusr hook begin") {
			t.Errorf("%s should not contain hook marker after uninstall, got:\n%s", rcPath, string(rc))
		}
	}

	state, err = os.ReadFile(statePath)
	if err != nil {
		t.Fatalf("read hook-state.json after uninstall: %v", err)
	}
	for _, hookType := range []string{"clone", "commit", "cd"} {
		if strings.Contains(string(state), hookType) {
			t.Errorf("hook-state.json should not contain %q after uninstall, got:\n%s", hookType, string(state))
		}
	}
	if _, err := os.Stat(wrapperPath); !os.IsNotExist(err) {
		t.Fatalf("wrapper file should be removed after uninstall, stat error: %v", err)
	}
}

// ---------------------------------------------------------------------------
// TestHookSingleUserNoTrigger — wrapper code includes single-user pass-through
// ---------------------------------------------------------------------------
func TestHookSingleUserNoTrigger(t *testing.T) {
	homeDir := t.TempDir()
	xdgDataHome := t.TempDir()

	env := map[string]string{
		"HOME":          homeDir,
		"XDG_DATA_HOME": xdgDataHome,
	}

	// Initialize store with only 1 user
	storePath := gitusrStore(xdgDataHome)
	users := []map[string]string{
		{"name": "Alice", "email": "alice@example.com"},
	}
	writeJSONFile(t, storePath, users)

	// Install hooks (unified)
	_, err := runGitusr(t, env, "hooks", "install")
	if err != nil {
		t.Fatalf("hooks install failed: %v", err)
	}

	// Read the wrapper file and verify single-user pass-through logic
	wrapperPath := filepath.Join(xdgDataHome, "gitusr", "hooks", "git-wrapper.sh")
	wrapper, err := os.ReadFile(wrapperPath)
	if err != nil {
		t.Fatalf("read wrapper file: %v", err)
	}

	wrapperContent := string(wrapper)
	if !strings.Contains(wrapperContent, "user_count") {
		t.Error("wrapper should contain single-user check variable (user_count)")
	}
	if !strings.Contains(wrapperContent, "-le 1") {
		t.Error("wrapper should contain pass-through threshold (-le 1)")
	}
}
