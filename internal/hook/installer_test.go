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

	// Verify wrapper contains cd-specific code (unified wrapper uses __gitusrcd for bash)
	wrapper, err := os.ReadFile(r.FilePath)
	if err != nil {
		t.Fatalf("read wrapper file: %v", err)
	}
	if !strings.Contains(string(wrapper), "__gitusrcd") {
		t.Error("cd wrapper should contain __gitusrcd (unified bash wrapper)")
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

// =============================================================================
// InstallAll tests
// =============================================================================

func TestInstallAll_FirstTime(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("HOME", tmpDir)
	t.Setenv("XDG_DATA_HOME", tmpDir)

	results, err := InstallAll([]ShellType{ShellTypeBash, ShellTypeZsh})
	if err != nil {
		t.Fatalf("InstallAll() returned error: %v", err)
	}

	if len(results) != 2 {
		t.Fatalf("expected 2 results, got %d", len(results))
	}

	// Verify each result has correct shell and file path
	for _, r := range results {
		if r.Shell != ShellTypeBash && r.Shell != ShellTypeZsh {
			t.Errorf("unexpected shell %q in results", r.Shell)
		}
		verifyWrapperFileExists(t, r.FilePath)
	}

	// Verify wrapper files are git-wrapper.{sh,zsh} (not cd-env)
	bashWrapper := results[0].FilePath
	if results[0].Shell == ShellTypeZsh {
		bashWrapper = results[1].FilePath
	}
	if !strings.Contains(bashWrapper, "git-wrapper.sh") {
		t.Errorf("expected git-wrapper.sh, got %s", bashWrapper)
	}

	// Verify .bashrc has hook markers (single source block for unified wrapper)
	bashrc := filepath.Join(tmpDir, ".bashrc")
	bashrcData, err := os.ReadFile(bashrc)
	if err != nil {
		t.Fatalf("failed to read .bashrc: %v", err)
	}
	bashrcContent := string(bashrcData)
	if strings.Count(bashrcContent, "# gitusr hook begin") != 1 {
		t.Errorf("expected 1 '# gitusr hook begin', got %d\n%s",
			strings.Count(bashrcContent, "# gitusr hook begin"), bashrcContent)
	}

	// Verify .zshrc has hook markers
	zshrc := filepath.Join(tmpDir, ".zshrc")
	zshrcData, err := os.ReadFile(zshrc)
	if err != nil {
		t.Fatalf("failed to read .zshrc: %v", err)
	}
	if !strings.Contains(string(zshrcData), "# gitusr hook begin") {
		t.Error(".zshrc should contain # gitusr hook begin marker")
	}

	// Verify state: all 3 types installed, none disabled
	state, err := LoadState()
	if err != nil {
		t.Fatalf("LoadState(): %v", err)
	}
	if len(state.InstalledTypes) != 3 {
		t.Errorf("expected 3 installed types, got %d: %v", len(state.InstalledTypes), state.InstalledTypes)
	}
	if len(state.DisabledTypes) != 0 {
		t.Errorf("expected 0 disabled types, got %d: %v", len(state.DisabledTypes), state.DisabledTypes)
	}

	// Verify each type individually
	for _, ht := range AllHookTypes {
		installed, err := IsInstalled(ht)
		if err != nil {
			t.Fatalf("IsInstalled(%s): %v", ht, err)
		}
		if !installed {
			t.Errorf("expected IsInstalled(%s) to be true", ht)
		}
	}
}

func TestInstallAll_Idempotent(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("HOME", tmpDir)
	t.Setenv("XDG_DATA_HOME", tmpDir)

	// First install
	results1, err := InstallAll([]ShellType{ShellTypeBash, ShellTypeZsh})
	if err != nil {
		t.Fatalf("first InstallAll() returned error: %v", err)
	}
	if len(results1) == 0 {
		t.Fatal("expected non-empty results from first install")
	}

	// Second install — should be idempotent (all types already installed)
	results2, err := InstallAll([]ShellType{ShellTypeBash, ShellTypeZsh})
	if err != nil {
		t.Fatalf("second InstallAll() returned error: %v", err)
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
		t.Errorf("expected 1 '# gitusr hook begin' after idempotent install, got %d\n%s", beginCount, content)
	}
	endCount := strings.Count(content, "# gitusr hook end")
	if endCount != 1 {
		t.Errorf("expected 1 '# gitusr hook end' after idempotent install, got %d\n%s", endCount, content)
	}

	// Verify state still has exactly 3 types without duplicates
	state, err := LoadState()
	if err != nil {
		t.Fatalf("LoadState(): %v", err)
	}
	if len(state.InstalledTypes) != 3 {
		t.Errorf("expected 3 installed types after idempotent call, got %d: %v",
			len(state.InstalledTypes), state.InstalledTypes)
	}
}

func TestInstallAll_PartiallyInstalled(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("HOME", tmpDir)
	t.Setenv("XDG_DATA_HOME", tmpDir)

	// Pre-install only HookTypeClone using old per-type Install (simulate partial state)
	_, err := Install(HookTypeClone, []ShellType{ShellTypeBash})
	if err != nil {
		t.Fatalf("pre-install Install(HookTypeClone): %v", err)
	}

	// Verify only clone is installed
	state, err := LoadState()
	if err != nil {
		t.Fatalf("LoadState(): %v", err)
	}
	if len(state.InstalledTypes) != 1 || state.InstalledTypes[0] != HookTypeClone {
		t.Fatalf("expected only HookTypeClone installed, got %v", state.InstalledTypes)
	}

	// Now call InstallAll — should proceed since not all 3 are installed
	results, err := InstallAll([]ShellType{ShellTypeBash})
	if err != nil {
		t.Fatalf("InstallAll() on partially installed state: %v", err)
	}
	if results == nil {
		t.Fatal("expected non-nil results when not all types are installed")
	}

	// Verify state now has all 3 types
	state, err = LoadState()
	if err != nil {
		t.Fatalf("LoadState() after InstallAll: %v", err)
	}
	if len(state.InstalledTypes) != 3 {
		t.Errorf("expected 3 installed types after InstallAll, got %d: %v",
			len(state.InstalledTypes), state.InstalledTypes)
	}
	if len(state.DisabledTypes) != 0 {
		t.Errorf("expected 0 disabled types, got %d", len(state.DisabledTypes))
	}
}

func TestInstallAll_WrapperContent(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("HOME", tmpDir)
	t.Setenv("XDG_DATA_HOME", tmpDir)

	results, err := InstallAll([]ShellType{ShellTypeBash, ShellTypeZsh})
	if err != nil {
		t.Fatalf("InstallAll() returned error: %v", err)
	}

	if len(results) != 2 {
		t.Fatalf("expected 2 results, got %d", len(results))
	}

	// Find the bash wrapper result
	var bashResult *HookInstallResult
	var zshResult *HookInstallResult
	for i := range results {
		switch results[i].Shell {
		case ShellTypeBash:
			bashResult = &results[i]
		case ShellTypeZsh:
			zshResult = &results[i]
		}
	}
	if bashResult == nil || zshResult == nil {
		t.Fatal("expected both bash and zsh results")
	}

	// Verify bash wrapper contains both git() and cd handler
	bashWrapper, err := os.ReadFile(bashResult.FilePath)
	if err != nil {
		t.Fatalf("read bash wrapper: %v", err)
	}
	bashContent := string(bashWrapper)
	if !strings.Contains(bashContent, "git()") {
		t.Error("bash wrapper should contain git() function definition")
	}
	if !strings.Contains(bashContent, "__gitusrcd") {
		t.Error("bash wrapper should contain __gitusrcd (cd handler)")
	}

	// Verify zsh wrapper contains both git() and chpwd hook
	zshWrapper, err := os.ReadFile(zshResult.FilePath)
	if err != nil {
		t.Fatalf("read zsh wrapper: %v", err)
	}
	zshContent := string(zshWrapper)
	if !strings.Contains(zshContent, "git()") {
		t.Error("zsh wrapper should contain git() function definition")
	}
	if !strings.Contains(zshContent, "chpwd") {
		t.Error("zsh wrapper should contain chpwd (cd handler)")
	}
}

func TestInstallAll_InvalidShell(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("HOME", tmpDir)
	t.Setenv("XDG_DATA_HOME", tmpDir)

	_, err := InstallAll([]ShellType{ShellType("invalid")})
	if err == nil {
		t.Fatal("expected error for invalid shell type")
	}
	if !strings.Contains(err.Error(), "unsupported shell type") {
		t.Errorf("error should mention 'unsupported shell type', got: %q", err.Error())
	}
}

func TestInstallAll_OnlyBash(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("HOME", tmpDir)
	t.Setenv("XDG_DATA_HOME", tmpDir)

	results, err := InstallAll([]ShellType{ShellTypeBash})
	if err != nil {
		t.Fatalf("InstallAll(bash): %v", err)
	}

	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	if results[0].Shell != ShellTypeBash {
		t.Errorf("expected bash result, got %s", results[0].Shell)
	}
	verifyBashrcHasBlock(t, tmpDir, results[0].FilePath)
	verifyWrapperFileExists(t, results[0].FilePath)

	// Verify .zshrc was NOT touched
	zshrc := filepath.Join(tmpDir, ".zshrc")
	if _, err := os.Stat(zshrc); !os.IsNotExist(err) {
		t.Error("expected .zshrc to not exist when only bash is installed")
	}

	// Verify all 3 types are in state even with only bash shell
	state, err := LoadState()
	if err != nil {
		t.Fatalf("LoadState(): %v", err)
	}
	if len(state.InstalledTypes) != 3 {
		t.Errorf("expected 3 installed types, got %d", len(state.InstalledTypes))
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

	// Assert cd-env.sh exists and contains __gitusrcd (unified wrapper includes cd handler)
	cdEnvPath := filepath.Join(tmpDir, "gitusr", "hooks", "cd-env.sh")
	cdEnvData, err := os.ReadFile(cdEnvPath)
	if err != nil {
		t.Fatalf("read cd-env.sh: %v", err)
	}
	if !strings.Contains(string(cdEnvData), "__gitusrcd") {
		t.Error("cd-env.sh should contain __gitusrcd (unified bash wrapper)")
	}
	// With unified wrappers, cd-env.sh also contains git() since both wrapper
	// types generate the same unified script with all hook behaviors.
	if !strings.Contains(string(cdEnvData), "git()") {
		t.Error("cd-env.sh should contain git() (unified wrapper includes all hooks)")
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
