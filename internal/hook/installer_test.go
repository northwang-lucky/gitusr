package hook

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// verifyBashrcHasCDBlock checks that the .bashrc file contains exactly one gitusr CD block.
func verifyBashrcHasCDBlock(t *testing.T, homeDir string, wrapperPath string) {
	t.Helper()

	bashrc := filepath.Join(homeDir, ".bashrc")
	data, err := os.ReadFile(bashrc)
	if err != nil {
		t.Fatalf("failed to read .bashrc: %v", err)
	}

	content := string(data)

	// Verify CD markers appear exactly once
	beginCount := strings.Count(content, "# gitusr cd begin")
	if beginCount != 1 {
		t.Errorf("expected 1 %q, got %d", "# gitusr cd begin", beginCount)
	}

	endCount := strings.Count(content, "# gitusr cd end")
	if endCount != 1 {
		t.Errorf("expected 1 %q, got %d", "# gitusr cd end", endCount)
	}

	// Verify source line contains the wrapper path
	if !strings.Contains(content, "source "+wrapperPath) {
		t.Errorf("expected source %s in .bashrc, got:\n%s", wrapperPath, content)
	}
}

// verifyBashrcHasBlock checks that the .bashrc file contains exactly one gitusr hook block.
func verifyBashrcHasBlock(t *testing.T, homeDir string, wrapperPath string) {
	t.Helper()

	bashrc := filepath.Join(homeDir, ".bashrc")
	data, err := os.ReadFile(bashrc)
	if err != nil {
		t.Fatalf("failed to read .bashrc: %v", err)
	}

	content := string(data)

	// Verify markers appear exactly once
	beginCount := strings.Count(content, "# gitusr hook begin")
	if beginCount != 1 {
		t.Errorf("expected 1 %q, got %d", "# gitusr hook begin", beginCount)
	}

	endCount := strings.Count(content, "# gitusr hook end")
	if endCount != 1 {
		t.Errorf("expected 1 %q, got %d", "# gitusr hook end", endCount)
	}

	// Verify source line contains the wrapper path
	if !strings.Contains(content, "source "+wrapperPath) {
		t.Errorf("expected source %s in .bashrc, got:\n%s", wrapperPath, content)
	}
}

// verifyWrapperFileExists checks that the wrapper file exists and is non-empty.
func verifyWrapperFileExists(t *testing.T, path string) {
	t.Helper()

	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("wrapper file not found at %s: %v", path, err)
	}

	if info.Size() == 0 {
		t.Errorf("wrapper file at %s is empty", path)
	}
}

func TestInstall_FirstTime(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("HOME", tmpDir)
	t.Setenv("XDG_DATA_HOME", tmpDir)

	results, err := Install(HookTypeClone, []ShellType{ShellTypeBash})
	if err != nil {
		t.Fatalf("Install() returned error: %v", err)
	}

	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}

	r := results[0]
	if r.Type != HookTypeClone {
		t.Errorf("result.Type = %q, want %q", r.Type, HookTypeClone)
	}
	if r.Shell != ShellTypeBash {
		t.Errorf("result.Shell = %q, want %q", r.Shell, ShellTypeBash)
	}

	// Verify .bashrc was updated
	verifyBashrcHasBlock(t, tmpDir, r.FilePath)

	// Verify wrapper file exists
	verifyWrapperFileExists(t, r.FilePath)

	// Verify state was saved
	installed, err := IsInstalled(HookTypeClone)
	if err != nil {
		t.Fatalf("IsInstalled() returned error: %v", err)
	}
	if !installed {
		t.Error("expected IsInstalled(HookTypeClone) to be true after install")
	}
}

func TestInstall_Idempotent(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("HOME", tmpDir)
	t.Setenv("XDG_DATA_HOME", tmpDir)

	// First install
	results1, err := Install(HookTypeClone, []ShellType{ShellTypeBash})
	if err != nil {
		t.Fatalf("first Install() returned error: %v", err)
	}
	if len(results1) == 0 {
		t.Fatal("expected non-empty results from first install")
	}

	// Second install — should be idempotent (skip because already installed)
	results2, err := Install(HookTypeClone, []ShellType{ShellTypeBash})
	if err != nil {
		t.Fatalf("second Install() returned error: %v", err)
	}
	if results2 != nil {
		t.Errorf("expected nil results on second install (already installed), got %v", results2)
	}

	// Verify no duplicate markers in .bashrc
	bashrc := filepath.Join(tmpDir, ".bashrc")
	data, err := os.ReadFile(bashrc)
	if err != nil {
		t.Fatalf("failed to read .bashrc: %v", err)
	}

	content := string(data)
	beginCount := strings.Count(content, "# gitusr hook begin")
	if beginCount != 1 {
		t.Errorf("expected 1 %q after idempotent install, got %d\nContent:\n%s", "# gitusr hook begin", beginCount, content)
	}

	// Verify state still has only one HookTypeClone
	state, err := LoadState()
	if err != nil {
		t.Fatalf("LoadState() returned error: %v", err)
	}

	cloneCount := 0
	for _, t := range state.InstalledTypes {
		if t == HookTypeClone {
			cloneCount++
		}
	}
	if cloneCount != 1 {
		t.Errorf("expected 1 HookTypeClone in state, got %d", cloneCount)
	}
}

func TestInstall_InvalidShell(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("HOME", tmpDir)
	t.Setenv("XDG_DATA_HOME", tmpDir)

	_, err := Install(HookTypeClone, []ShellType{ShellType("invalid")})
	if err == nil {
		t.Fatal("expected error for invalid shell type")
	}
	if !strings.Contains(err.Error(), "unsupported shell type") {
		t.Errorf("error should mention 'unsupported shell type', got: %q", err.Error())
	}
}

func TestInstall_MultipleShells(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("HOME", tmpDir)
	t.Setenv("XDG_DATA_HOME", tmpDir)

	results, err := Install(HookTypeClone, []ShellType{ShellTypeBash, ShellTypeZsh})
	if err != nil {
		t.Fatalf("Install() returned error: %v", err)
	}

	if len(results) != 2 {
		t.Fatalf("expected 2 results, got %d", len(results))
	}

	// Verify both shell configs were updated
	verifyBashrcHasBlock(t, tmpDir, results[0].FilePath)

	zshrc := filepath.Join(tmpDir, ".zshrc")
	zshData, err := os.ReadFile(zshrc)
	if err != nil {
		t.Fatalf("failed to read .zshrc: %v", err)
	}
	if !strings.Contains(string(zshData), "# gitusr hook begin") {
		t.Error(".zshrc should contain the hook marker begin")
	}

	// Verify both wrapper files exist
	for _, r := range results {
		verifyWrapperFileExists(t, r.FilePath)
	}

	// Verify state
	state, err := LoadState()
	if err != nil {
		t.Fatalf("LoadState() returned error: %v", err)
	}
	if len(state.InstalledTypes) != 1 {
		t.Errorf("expected 1 installed type, got %d", len(state.InstalledTypes))
	}
}

func TestInstall_CD_FirstTime(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("HOME", tmpDir)
	t.Setenv("XDG_DATA_HOME", tmpDir)

	results, err := Install(HookTypeCD, []ShellType{ShellTypeBash})
	if err != nil {
		t.Fatalf("Install(HookTypeCD) returned error: %v", err)
	}

	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}

	r := results[0]
	if r.Type != HookTypeCD {
		t.Errorf("result.Type = %q, want %q", r.Type, HookTypeCD)
	}
	if r.Shell != ShellTypeBash {
		t.Errorf("result.Shell = %q, want %q", r.Shell, ShellTypeBash)
	}

	// Verify .bashrc was updated with CD markers
	verifyBashrcHasCDBlock(t, tmpDir, r.FilePath)

	// Verify wrapper file exists
	verifyWrapperFileExists(t, r.FilePath)

	// Verify state was saved
	installed, err := IsInstalled(HookTypeCD)
	if err != nil {
		t.Fatalf("IsInstalled() returned error: %v", err)
	}
	if !installed {
		t.Error("expected IsInstalled(HookTypeCD) to be true after install")
	}

	// Verify wrapper contains cd-specific code
	wrapper, err := os.ReadFile(r.FilePath)
	if err != nil {
		t.Fatalf("read wrapper file: %v", err)
	}
	if !strings.Contains(string(wrapper), "__gitusr_use_if_found") {
		t.Error("cd wrapper should contain __gitusr_use_if_found")
	}
}

func TestAppendSourceLine_MarkersAlreadyExist(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("HOME", tmpDir)

	bashrc := filepath.Join(tmpDir, ".bashrc")

	// Write initial file with markers
	oldBlock := "# gitusr hook begin\nsource /old/path/wrapper.sh\n# gitusr hook end\n"
	if err := os.WriteFile(bashrc, []byte(oldBlock), 0644); err != nil {
		t.Fatal(err)
	}

	// Append a new source line — should replace old block
	newWrapperPath := "/new/path/wrapper.sh"
	if err := AppendSourceLine(ShellTypeBash, newWrapperPath); err != nil {
		t.Fatalf("AppendSourceLine() returned error: %v", err)
	}

	data, err := os.ReadFile(bashrc)
	if err != nil {
		t.Fatal(err)
	}
	content := string(data)

	// Should have exactly one block
	if strings.Count(content, "# gitusr hook begin") != 1 {
		t.Errorf("expected 1 marker begin, got %d\nContent:\n%s",
			strings.Count(content, "# gitusr hook begin"), content)
	}

	// Should contain the new source path
	if !strings.Contains(content, "source "+newWrapperPath) {
		t.Errorf("expected source %s in content:\n%s", newWrapperPath, content)
	}

	// Should not contain the old source path
	if strings.Contains(content, "/old/path") {
		t.Errorf("old source path should have been removed:\n%s", content)
	}
}

func TestInstall_AllDoesNotOverwrite(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("HOME", tmpDir)
	t.Setenv("XDG_DATA_HOME", tmpDir)

	// Install hooks sequentially: clone → commit → cd
	hookTypes := []HookType{HookTypeClone, HookTypeCommit, HookTypeCD}
	for _, ht := range hookTypes {
		results, err := Install(ht, []ShellType{ShellTypeBash})
		if err != nil {
			t.Fatalf("Install(%s) returned error: %v", ht, err)
		}
		if len(results) != 1 {
			t.Fatalf("Install(%s): expected 1 result, got %d", ht, len(results))
		}
	}

	// Assert git-wrapper.sh contains git() function definition
	wrapperPath := filepath.Join(tmpDir, "gitusr", "hooks", "git-wrapper.sh")
	wrapperData, err := os.ReadFile(wrapperPath)
	if err != nil {
		t.Fatalf("read git-wrapper.sh: %v", err)
	}
	if !strings.Contains(string(wrapperData), "git()") {
		t.Error("git-wrapper.sh should contain git() function definition")
	}

	// Assert cd-env.sh exists, contains __gitusr_use_if_found, and does NOT contain git()
	cdEnvPath := filepath.Join(tmpDir, "gitusr", "hooks", "cd-env.sh")
	cdEnvData, err := os.ReadFile(cdEnvPath)
	if err != nil {
		t.Fatalf("read cd-env.sh: %v", err)
	}
	if !strings.Contains(string(cdEnvData), "__gitusr_use_if_found") {
		t.Error("cd-env.sh should contain __gitusr_use_if_found")
	}
	if strings.Contains(string(cdEnvData), "git()") {
		t.Error("cd-env.sh should NOT contain git() function definition")
	}

	// Assert .bashrc contains both hook and cd markers
	bashrc := filepath.Join(tmpDir, ".bashrc")
	bashrcData, err := os.ReadFile(bashrc)
	if err != nil {
		t.Fatalf("read .bashrc: %v", err)
	}
	bashrcContent := string(bashrcData)
	if !strings.Contains(bashrcContent, "# gitusr hook begin") {
		t.Error(".bashrc should contain # gitusr hook begin marker")
	}
	if !strings.Contains(bashrcContent, "# gitusr cd begin") {
		t.Error(".bashrc should contain # gitusr cd begin marker")
	}

	// Assert state has all three types installed
	state, err := LoadState()
	if err != nil {
		t.Fatalf("LoadState(): %v", err)
	}
	if len(state.InstalledTypes) != 3 {
		t.Errorf("expected 3 installed types, got %d: %v", len(state.InstalledTypes), state.InstalledTypes)
	}

	hasClone := false
	hasCommit := false
	hasCD := false
	for _, typ := range state.InstalledTypes {
		switch typ {
		case HookTypeClone:
			hasClone = true
		case HookTypeCommit:
			hasCommit = true
		case HookTypeCD:
			hasCD = true
		}
	}
	if !hasClone {
		t.Error("state missing HookTypeClone")
	}
	if !hasCommit {
		t.Error("state missing HookTypeCommit")
	}
	if !hasCD {
		t.Error("state missing HookTypeCD")
	}
}
