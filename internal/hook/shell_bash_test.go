package hook

import (
	"strings"
	"testing"
)

// =============================================================================
// Tests for GenerateUnifiedBashWrapper
// =============================================================================

func TestUnifiedBashWrapper_ContainsGitFunction(t *testing.T) {
	script := GenerateUnifiedBashWrapper()
	if !strings.Contains(script, "git() {") {
		t.Error("unified wrapper should define a git() function")
	}
}

func TestUnifiedBashWrapper_ContainsCDHandler(t *testing.T) {
	script := GenerateUnifiedBashWrapper()
	if !strings.Contains(script, "__gitusrcd() {") {
		t.Error("unified wrapper should define __gitusrcd() function")
	}
	if !strings.Contains(script, "alias cd=__gitusrcd") {
		t.Error("unified wrapper should alias cd to __gitusrcd")
	}
}

func TestUnifiedBashWrapper_DefinesDataDir(t *testing.T) {
	script := GenerateUnifiedBashWrapper()
	if !strings.Contains(script, "__GITUSR_DATA_DIR") {
		t.Error("unified wrapper should define __GITUSR_DATA_DIR variable")
	}
	if !strings.Contains(script, "XDG_DATA_HOME") {
		t.Error("unified wrapper should reference XDG_DATA_HOME in data dir")
	}
}

func TestUnifiedBashWrapper_UsesCommandGit(t *testing.T) {
	script := GenerateUnifiedBashWrapper()
	count := strings.Count(script, "command git")
	if count < 2 {
		t.Errorf("unified wrapper should use 'command git' at least twice, got %d occurrence(s)", count)
	}
}

func TestUnifiedBashWrapper_HasSingleUserCheck(t *testing.T) {
	script := GenerateUnifiedBashWrapper()
	if !strings.Contains(script, "gitusr list") {
		t.Error("unified wrapper should check user count via gitusr list")
	}
	if !strings.Contains(script, "-le 1") {
		t.Error("unified wrapper should check if user count is <= 1")
	}
}

func TestUnifiedBashWrapper_HasDisabledChecks(t *testing.T) {
	script := GenerateUnifiedBashWrapper()

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

func TestUnifiedBashWrapper_DisabledCheckUsesRedirect(t *testing.T) {
	script := GenerateUnifiedBashWrapper()
	count := strings.Count(script, "2>/dev/null")
	// Should appear at least 3 times: clone disabled, commit disabled, cd disabled
	// + 2 for gitusr list user_count checks = 5 minimum
	if count < 3 {
		t.Errorf("unified wrapper should redirect stderr for disabled checks, got %d occurrences of 2>/dev/null", count)
	}
}

func TestUnifiedBashWrapper_HandlesClone(t *testing.T) {
	script := GenerateUnifiedBashWrapper()
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

func TestUnifiedBashWrapper_ReturnsToOriginalDir(t *testing.T) {
	script := GenerateUnifiedBashWrapper()
	if !strings.Contains(script, "original_dir=$(pwd)") {
		t.Error("unified wrapper should save the original directory with 'original_dir=$(pwd)'")
	}
	if !strings.Contains(script, `\cd "$original_dir"`) {
		t.Error("unified wrapper should return to the original directory with '\\cd \"$original_dir\"'")
	}
}

func TestUnifiedBashWrapper_AlwaysCallsGitusrUse(t *testing.T) {
	script := GenerateUnifiedBashWrapper()
	if !strings.Contains(script, "else") {
		t.Error("unified wrapper should have an else branch for unconditional gitusr use")
	}
	count := strings.Count(script, "gitusr use")
	if count < 4 {
		t.Errorf("unified wrapper should call 'gitusr use' at least 4 times (including fallback), got %d", count)
	}
}

func TestUnifiedBashWrapper_HandlesCommit(t *testing.T) {
	script := GenerateUnifiedBashWrapper()
	if !strings.Contains(script, "commit") {
		t.Error("unified wrapper should handle commit subcommand")
	}
	if !strings.Contains(script, ".gitusrrc") {
		t.Error("unified wrapper should check for .gitusrrc file")
	}
	if !strings.Contains(script, "apply-rc") {
		t.Error("unified wrapper should call gitusr hooks apply-rc")
	}
	if !strings.Contains(script, "command git commit") {
		t.Error("unified wrapper should use 'command git commit' for the actual commit")
	}
}

func TestUnifiedBashWrapper_CDHandlerUsesBackslashCd(t *testing.T) {
	script := GenerateUnifiedBashWrapper()
	// __gitusrcd should use \cd to avoid alias expansion
	if !strings.Contains(script, `\cd "$@"`) {
		t.Error("unified wrapper cd handler should use \\cd to avoid alias expansion")
	}
}

func TestUnifiedBashWrapper_CDHandlerChecksRC(t *testing.T) {
	script := GenerateUnifiedBashWrapper()
	// __gitusrcd should check for .gitusrrc
	if !strings.Contains(script, "[ -f .gitusrrc ]") {
		t.Error("unified wrapper cd handler should check for .gitusrrc file")
	}
}


func TestGenerateBashWrapper_ContainsGitFunction(t *testing.T) {
	script := GenerateBashWrapper()
	if !strings.Contains(script, "git() {") {
		t.Error("generated script should define a git() function")
	}
}

func TestGenerateBashWrapper_UsesCommandGit(t *testing.T) {
	script := GenerateBashWrapper()
	count := strings.Count(script, "command git")
	if count < 2 {
		t.Errorf("generated script should use 'command git' at least twice, got %d occurrence(s)", count)
	}
}

func TestGenerateBashWrapper_HasSingleUserCheck(t *testing.T) {
	script := GenerateBashWrapper()
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

func TestGenerateBashWrapper_HandlesClone(t *testing.T) {
	script := GenerateBashWrapper()
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
}

func TestGenerateBashWrapper_ReturnsToOriginalDir(t *testing.T) {
	script := GenerateBashWrapper()
	if !strings.Contains(script, "original_dir=$(pwd)") {
		t.Error("generated script should save the original directory with 'original_dir=$(pwd)'")
	}
	if !strings.Contains(script, `\cd "$original_dir"`) {
		t.Error("generated script should return to the original directory with '\\cd \"$original_dir\"'")
	}
}

func TestGenerateBashWrapper_AlwaysCallsGitusrUse(t *testing.T) {
	script := GenerateBashWrapper()
	if !strings.Contains(script, "else") {
		t.Error("generated script should have an else branch for unconditional gitusr use")
	}
	count := strings.Count(script, "gitusr use")
	if count < 4 {
		t.Errorf("generated script should call 'gitusr use' at least 4 times (including fallback), got %d", count)
	}
}

func TestGenerateBashWrapper_HandlesCommit(t *testing.T) {
	script := GenerateBashWrapper()
	if !strings.Contains(script, "commit") {
		t.Error("generated script should handle commit subcommand")
	}
	if !strings.Contains(script, ".gitusrrc") {
		t.Error("generated script should check for .gitusrrc file")
	}
	if !strings.Contains(script, "apply-rc") {
		t.Error("generated script should call gitusr hooks apply-rc")
	}
}
