package hook

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestWriteWrapperFile(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("XDG_DATA_HOME", tmpDir)

	content := "test wrapper content"
	path, err := WriteWrapperFile(ShellTypeBash, content)
	if err != nil {
		t.Fatalf("WriteWrapperFile failed: %v", err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read written file: %v", err)
	}
	if string(data) != content {
		t.Errorf("content = %q, want %q", string(data), content)
	}

	// Verify path ends with gitusr/hooks/git-wrapper.sh
	wantSuffix := filepath.Join("gitusr", "hooks", "git-wrapper.sh")
	if !strings.HasSuffix(path, wantSuffix) {
		t.Errorf("path = %q, want suffix %q", path, wantSuffix)
	}
}

func TestWriteWrapperFile_CreateDir(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("XDG_DATA_HOME", tmpDir)

	hooksDir := filepath.Join(tmpDir, "gitusr", "hooks")
	if _, err := os.Stat(hooksDir); err == nil {
		t.Fatal("hooks dir should not exist yet")
	}

	path, err := WriteWrapperFile(ShellTypeZsh, "content")
	if err != nil {
		t.Fatalf("WriteWrapperFile failed: %v", err)
	}

	if _, err := os.Stat(hooksDir); os.IsNotExist(err) {
		t.Fatal("hooks dir should exist after WriteWrapperFile")
	}
	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Fatal("file should exist after WriteWrapperFile")
	}

	// Verify zsh extension
	if !strings.HasSuffix(path, "git-wrapper.zsh") {
		t.Errorf("path = %q, want suffix .zsh", path)
	}
}

func TestAppendSourceLine(t *testing.T) {
	tmpHome := t.TempDir()
	t.Setenv("HOME", tmpHome)

	bashrc := filepath.Join(tmpHome, ".bashrc")
	if err := os.WriteFile(bashrc, []byte("existing config"), 0644); err != nil {
		t.Fatal(err)
	}

	if err := AppendSourceLine(ShellTypeBash, "/path/to/wrapper.sh"); err != nil {
		t.Fatalf("AppendSourceLine failed: %v", err)
	}

	data, err := os.ReadFile(bashrc)
	if err != nil {
		t.Fatalf("read .bashrc: %v", err)
	}

	content := string(data)
	if !strings.Contains(content, "# gitusr hook begin") {
		t.Error("expected start marker")
	}
	if !strings.Contains(content, "# gitusr hook end") {
		t.Error("expected end marker")
	}
	if !strings.Contains(content, "source /path/to/wrapper.sh") {
		t.Error("expected source line")
	}
	if !strings.HasPrefix(content, "existing config") {
		t.Error("should preserve existing config")
	}
}

func TestAppendSourceLine_Idempotent(t *testing.T) {
	tmpHome := t.TempDir()
	t.Setenv("HOME", tmpHome)

	bashrc := filepath.Join(tmpHome, ".bashrc")
	if err := os.WriteFile(bashrc, []byte("config"), 0644); err != nil {
		t.Fatal(err)
	}

	// First append
	if err := AppendSourceLine(ShellTypeBash, "/path/to/wrapper.sh"); err != nil {
		t.Fatal(err)
	}

	// Second append with same path
	if err := AppendSourceLine(ShellTypeBash, "/path/to/wrapper.sh"); err != nil {
		t.Fatal(err)
	}

	data, err := os.ReadFile(bashrc)
	if err != nil {
		t.Fatalf("read .bashrc: %v", err)
	}

	content := string(data)
	beginCount := strings.Count(content, "# gitusr hook begin")
	if beginCount != 1 {
		t.Errorf("expected 1 start marker, got %d\nContent:\n%s", beginCount, content)
	}
	if !strings.HasPrefix(content, "config") {
		t.Error("should preserve existing config")
	}
}

func TestAppendSourceLine_Zsh(t *testing.T) {
	tmpHome := t.TempDir()
	t.Setenv("HOME", tmpHome)

	zshrc := filepath.Join(tmpHome, ".zshrc")
	if err := os.WriteFile(zshrc, []byte("zsh config"), 0644); err != nil {
		t.Fatal(err)
	}

	if err := AppendSourceLine(ShellTypeZsh, "/path/to/wrapper.zsh"); err != nil {
		t.Fatalf("AppendSourceLine failed: %v", err)
	}

	data, err := os.ReadFile(zshrc)
	if err != nil {
		t.Fatalf("read .zshrc: %v", err)
	}

	content := string(data)
	if !strings.Contains(content, "# gitusr hook begin") {
		t.Error("expected start marker")
	}
	if !strings.Contains(content, "# gitusr hook end") {
		t.Error("expected end marker")
	}
	if !strings.Contains(content, "source /path/to/wrapper.zsh") {
		t.Error("expected source line")
	}
}

func TestRemoveSourceBlock(t *testing.T) {
	tmpHome := t.TempDir()
	t.Setenv("HOME", tmpHome)

	bashrc := filepath.Join(tmpHome, ".bashrc")
	initialContent := "config\n# gitusr hook begin\nsource /path\n# gitusr hook end\n"
	if err := os.WriteFile(bashrc, []byte(initialContent), 0644); err != nil {
		t.Fatal(err)
	}

	if err := RemoveSourceBlock(ShellTypeBash); err != nil {
		t.Fatalf("RemoveSourceBlock failed: %v", err)
	}

	data, err := os.ReadFile(bashrc)
	if err != nil {
		t.Fatalf("read .bashrc: %v", err)
	}

	content := string(data)
	if strings.Contains(content, "# gitusr hook begin") {
		t.Error("should not contain start marker")
	}
	if strings.Contains(content, "# gitusr hook end") {
		t.Error("should not contain end marker")
	}
}

func TestRemoveSourceBlock_NotFound(t *testing.T) {
	tmpHome := t.TempDir()
	t.Setenv("HOME", tmpHome)

	bashrc := filepath.Join(tmpHome, ".bashrc")
	if err := os.WriteFile(bashrc, []byte("plain config without markers"), 0644); err != nil {
		t.Fatal(err)
	}

	// Should return nil when no block exists
	if err := RemoveSourceBlock(ShellTypeBash); err != nil {
		t.Fatalf("RemoveSourceBlock should return nil when no block found: %v", err)
	}

	// Content should be unchanged
	data, err := os.ReadFile(bashrc)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != "plain config without markers" {
		t.Errorf("content changed: %q", string(data))
	}
}

func TestRemoveSourceBlock_NoConfigFile(t *testing.T) {
	tmpHome := t.TempDir()
	t.Setenv("HOME", tmpHome)

	// Should return nil when config file doesn't exist
	if err := RemoveSourceBlock(ShellTypeZsh); err != nil {
		t.Fatalf("RemoveSourceBlock should return nil when no config file: %v", err)
	}
}

func TestAppendSourceLine_WithCDMarker(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("HOME", tmpDir)

	// First call: standard marker
	err := AppendSourceLine(ShellTypeBash, "/path/git-wrapper.sh")
	if err != nil {
		t.Fatal(err)
	}

	// Second call: different path (simulating CD hook)
	err = AppendSourceLine(ShellTypeBash, "/path/cd-env.sh")
	if err != nil {
		t.Fatal(err)
	}

	// Read .bashrc
	data, err := os.ReadFile(filepath.Join(tmpDir, ".bashrc"))
	if err != nil {
		t.Fatal(err)
	}
	content := string(data)

	// The current implementation REPLACES the marker block, so:
	// - There should be only ONE marker block
	// - The source line should point to cd-env.sh (the last one)
	// - The git-wrapper.sh source line should be GONE
	//
	// For TDD RED, we want to assert what SHOULD happen after fix:
	// Both source lines should exist in SEPARATE marker blocks
	// So this test will FAIL because currently only one marker block exists

	if !strings.Contains(content, "source /path/git-wrapper.sh") {
		t.Error("Expected .bashrc to contain source line for git-wrapper.sh")
	}
	if !strings.Contains(content, "source /path/cd-env.sh") {
		t.Error("Expected .bashrc to contain source line for cd-env.sh")
	}

	// Count marker blocks - after fix there should be 1 of each
	hookCount := strings.Count(content, "# gitusr hook begin")
	cdCount := strings.Count(content, "# gitusr cd begin")
	if hookCount != 1 || cdCount != 1 {
		t.Errorf("Expected 1 hook marker and 1 cd marker, got %d hook and %d cd", hookCount, cdCount)
	}
}
