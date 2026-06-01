package hook

import (
	"strings"
	"testing"
)

func TestGenerateZshWrapper_ContainsGitFunction(t *testing.T) {
	script := GenerateZshWrapper()
	if !strings.Contains(script, "git() {") {
		t.Error("generated script should define a git() function")
	}
}

func TestGenerateZshWrapper_UsesCommandGit(t *testing.T) {
	script := GenerateZshWrapper()
	count := strings.Count(script, "command git")
	if count < 2 {
		t.Errorf("generated script should use 'command git' at least twice, got %d occurrence(s)", count)
	}
}

func TestGenerateZshWrapper_HasSingleUserCheck(t *testing.T) {
	script := GenerateZshWrapper()
	if !strings.Contains(script, "gitusr list") {
		t.Error("generated script should check user count via gitusr list")
	}
	if !strings.Contains(script, "user_count") {
		t.Error("generated script should use a user_count variable")
	}
	if !strings.Contains(script, "-le 1") {
		t.Error("generated script should check if user count is <= 1")
	}
}

func TestGenerateZshWrapper_HandlesClone(t *testing.T) {
	script := GenerateZshWrapper()
	if !strings.Contains(script, "clone") {
		t.Error("generated script should handle clone subcommand")
	}
	if !strings.Contains(script, "--gu-name") {
		t.Error("generated script should handle --gu-name argument")
	}
	if !strings.Contains(script, "--gu-email") {
		t.Error("generated script should handle --gu-email argument")
	}
	if !strings.Contains(script, "command git clone") {
		t.Error("generated script should use 'command git clone' for the actual clone")
	}
	if !strings.Contains(script, "gitusr use") {
		t.Error("generated script should call gitusr use after clone")
	}
	if !strings.Contains(script, "local -a") {
		t.Error("generated script should declare arrays with 'local -a' (zsh syntax)")
	}
}

func TestGenerateZshWrapper_ReturnsToOriginalDir(t *testing.T) {
	script := GenerateZshWrapper()
	if !strings.Contains(script, "original_dir=$(pwd)") {
		t.Error("generated script should save the original directory with 'original_dir=$(pwd)'")
	}
	if !strings.Contains(script, `\cd "$original_dir"`) {
		t.Error("generated script should return to the original directory with '\\cd \"$original_dir\"'")
	}
}

func TestGenerateZshWrapper_AlwaysCallsGitusrUse(t *testing.T) {
	script := GenerateZshWrapper()
	if !strings.Contains(script, "else") {
		t.Error("generated script should have an else branch for unconditional gitusr use")
	}
	count := strings.Count(script, "gitusr use")
	if count < 4 {
		t.Errorf("generated script should call 'gitusr use' at least 4 times (including fallback), got %d", count)
	}
}

func TestGenerateZshWrapper_HandlesCommit(t *testing.T) {
	script := GenerateZshWrapper()
	if !strings.Contains(script, "commit") {
		t.Error("generated script should handle commit subcommand")
	}
	if !strings.Contains(script, ".gitusrrc") {
		t.Error("generated script should check for .gitusrrc file")
	}
	if !strings.Contains(script, "apply-rc") {
		t.Error("generated script should call gitusr hook apply-rc")
	}
}

// =============================================================================
// Tests for GenerateUnifiedZshWrapper
// =============================================================================

func TestUnifiedZshWrapper_ContainsGitFunction(t *testing.T) {
	script := GenerateUnifiedZshWrapper()
	if !strings.Contains(script, "git() {") {
		t.Error("unified wrapper should define a git() function")
	}
}

func TestUnifiedZshWrapper_ContainsChpwdHook(t *testing.T) {
	script := GenerateUnifiedZshWrapper()
	if !strings.Contains(script, "__gitusr_autoload_hook() {") {
		t.Error("unified wrapper should define __gitusr_autoload_hook()")
	}
	if !strings.Contains(script, "add-zsh-hook chpwd __gitusr_autoload_hook") {
		t.Error("unified wrapper should register chpwd hook with add-zsh-hook")
	}
}

func TestUnifiedZshWrapper_UsesAddZshHookRemoveFirst(t *testing.T) {
	script := GenerateUnifiedZshWrapper()
	if !strings.Contains(script, "add-zsh-hook -D chpwd __gitusr_autoload_hook") {
		t.Error("unified wrapper should remove old hook first with add-zsh-hook -D")
	}
}

func TestUnifiedZshWrapper_DefinesDataDir(t *testing.T) {
	script := GenerateUnifiedZshWrapper()
	if !strings.Contains(script, "__GITUSR_DATA_DIR") {
		t.Error("unified wrapper should define __GITUSR_DATA_DIR variable")
	}
	if !strings.Contains(script, "XDG_DATA_HOME") {
		t.Error("unified wrapper should reference XDG_DATA_HOME in data dir")
	}
}

func TestUnifiedZshWrapper_UsesZshSyntax(t *testing.T) {
	script := GenerateUnifiedZshWrapper()
	if !strings.Contains(script, "[[ ") {
		t.Error("unified wrapper should use zsh [[ ]] syntax for conditions")
	}
	if !strings.Contains(script, "local -a") {
		t.Error("unified wrapper should declare arrays with 'local -a' (zsh syntax)")
	}
	if !strings.Contains(script, "autoload -U add-zsh-hook") {
		t.Error("unified wrapper should autoload add-zsh-hook")
	}
}

func TestUnifiedZshWrapper_UsesCommandGit(t *testing.T) {
	script := GenerateUnifiedZshWrapper()
	count := strings.Count(script, "command git")
	if count < 2 {
		t.Errorf("unified wrapper should use 'command git' at least twice, got %d occurrence(s)", count)
	}
}

func TestUnifiedZshWrapper_HasSingleUserCheck(t *testing.T) {
	script := GenerateUnifiedZshWrapper()
	if !strings.Contains(script, "gitusr list") {
		t.Error("unified wrapper should check user count via gitusr list")
	}
	if !strings.Contains(script, "-le 1") {
		t.Error("unified wrapper should check if user count is <= 1")
	}
}

func TestUnifiedZshWrapper_HasDisabledChecks(t *testing.T) {
	script := GenerateUnifiedZshWrapper()

	checks := map[string]string{
		"clone":  "gitusr hooks is-disabled clone",
		"commit": "gitusr hooks is-disabled commit",
		"cd":     "gitusr hooks is-disabled cd",
	}

	for hookType, expected := range checks {
		if !strings.Contains(script, expected) {
			t.Errorf("unified wrapper should check disabled status for %s hook", hookType)
		}
	}
}

func TestUnifiedZshWrapper_DisabledCheckUsesRedirect(t *testing.T) {
	script := GenerateUnifiedZshWrapper()
	count := strings.Count(script, "2>/dev/null")
	// Should appear at least 5 times: gitusr list (git fn), gitusr list (chpwd),
	// clone disabled check, commit disabled check, cd disabled check
	if count < 5 {
		t.Errorf("unified wrapper should redirect stderr for disabled checks and list calls, got %d occurrences of 2>/dev/null", count)
	}
}

func TestUnifiedZshWrapper_HandlesClone(t *testing.T) {
	script := GenerateUnifiedZshWrapper()
	if !strings.Contains(script, "clone") {
		t.Error("unified wrapper should handle clone subcommand")
	}
	if !strings.Contains(script, "--gu-name") {
		t.Error("unified wrapper should handle --gu-name argument")
	}
	if !strings.Contains(script, "--gu-email") {
		t.Error("unified wrapper should handle --gu-email argument")
	}
	if !strings.Contains(script, "command git clone") {
		t.Error("unified wrapper should use 'command git clone' for the actual clone")
	}
	if !strings.Contains(script, "gitusr use") {
		t.Error("unified wrapper should call gitusr use after clone")
	}
}

func TestUnifiedZshWrapper_ReturnsToOriginalDir(t *testing.T) {
	script := GenerateUnifiedZshWrapper()
	if !strings.Contains(script, "original_dir=$(pwd)") {
		t.Error("unified wrapper should save the original directory with 'original_dir=$(pwd)'")
	}
	if !strings.Contains(script, `\cd "$original_dir"`) {
		t.Error("unified wrapper should return to the original directory with '\\cd \"$original_dir\"'")
	}
}

func TestUnifiedZshWrapper_AlwaysCallsGitusrUse(t *testing.T) {
	script := GenerateUnifiedZshWrapper()
	if !strings.Contains(script, "else") {
		t.Error("unified wrapper should have an else branch for unconditional gitusr use")
	}
	count := strings.Count(script, "gitusr use")
	if count < 4 {
		t.Errorf("unified wrapper should call 'gitusr use' at least 4 times (including fallback), got %d", count)
	}
}

func TestUnifiedZshWrapper_HandlesCommit(t *testing.T) {
	script := GenerateUnifiedZshWrapper()
	if !strings.Contains(script, "commit") {
		t.Error("unified wrapper should handle commit subcommand")
	}
	if !strings.Contains(script, ".gitusrrc") {
		t.Error("unified wrapper should check for .gitusrrc file")
	}
	if !strings.Contains(script, "apply-rc") {
		t.Error("unified wrapper should call gitusr hook apply-rc")
	}
	if !strings.Contains(script, "command git commit") {
		t.Error("unified wrapper should use 'command git commit' for the actual commit")
	}
}

func TestUnifiedZshWrapper_ChpwdHookChecksRC(t *testing.T) {
	script := GenerateUnifiedZshWrapper()
	// __gitusr_autoload_hook section should check for .gitusrrc with [[ -f ]]
	if !strings.Contains(script, "[[ -f .gitusrrc ]]") {
		t.Error("unified wrapper chpwd hook should check for .gitusrrc with [[ -f ]]")
	}
}

func TestUnifiedZshWrapper_ChpwdHookUsesApplyRC(t *testing.T) {
	script := GenerateUnifiedZshWrapper()
	if !strings.Contains(script, "gitusr hook apply-rc --silent-if-unchanged") {
		t.Error("unified wrapper chpwd hook should call gitusr hook apply-rc --silent-if-unchanged")
	}
}
