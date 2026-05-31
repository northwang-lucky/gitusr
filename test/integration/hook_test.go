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

	// Install clone hook
	output, err := runGitusr(t, env, "hook", "install", "--type", "clone")
	if err != nil {
		t.Fatalf("hook install --type clone failed: %v\noutput: %s", err, output)
	}
	if !strings.Contains(output, "successfully installed") {
		t.Errorf("expected 'successfully installed' in output, got: %s", output)
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

	// Uninstall clone hook
	output, err = runGitusr(t, env, "hook", "uninstall", "--type", "clone")
	if err != nil {
		t.Fatalf("hook uninstall --type clone failed: %v\noutput: %s", err, output)
	}
	if !strings.Contains(output, "successfully uninstalled") {
		t.Errorf("expected 'successfully uninstalled' in output, got: %s", output)
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

	// Install cd hook
	output, err := runGitusr(t, env, "hook", "install", "--type", "cd")
	if err != nil {
		t.Fatalf("hook install --type cd failed: %v\noutput: %s", err, output)
	}
	if !strings.Contains(output, "successfully installed") {
		t.Errorf("expected 'successfully installed' in output, got: %s", output)
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

	// Verify wrapper file exists and contains cd-specific code
	wrapperPath := filepath.Join(xdgDataHome, "gitusr", "hooks", "git-wrapper.sh")
	wrapper, err := os.ReadFile(wrapperPath)
	if err != nil {
		t.Fatalf("read wrapper file: %v", err)
	}
	if !strings.Contains(string(wrapper), "__gitusr_use_if_found") {
		t.Errorf("wrapper should contain '__gitusr_use_if_found', got:\n%s", string(wrapper))
	}

	// Uninstall cd hook
	output, err = runGitusr(t, env, "hook", "uninstall", "--type", "cd")
	if err != nil {
		t.Fatalf("hook uninstall --type cd failed: %v\noutput: %s", err, output)
	}
	if !strings.Contains(output, "successfully uninstalled") {
		t.Errorf("expected 'successfully uninstalled' in output, got: %s", output)
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

	// Install clone hook
	_, err := runGitusr(t, env, "hook", "install", "--type", "clone")
	if err != nil {
		t.Fatalf("hook install failed: %v", err)
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
