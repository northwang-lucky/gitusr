package hook

import (
	"strings"
	"testing"
)

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
		t.Error("generated script should call gitusr hook apply-rc")
	}
}
