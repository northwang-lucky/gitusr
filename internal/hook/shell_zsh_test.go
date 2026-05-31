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
