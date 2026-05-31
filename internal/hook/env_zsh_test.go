package hook

import (
	"strings"
	"testing"
)

func TestGenerateZshEnv_ContainsChpwdHook(t *testing.T) {
	script := GenerateZshEnv()
	if !strings.Contains(script, "chpwd") {
		t.Error("generated script should reference chpwd hook")
	}
	if !strings.Contains(script, "__gitusr_autoload_hook") {
		t.Error("generated script should define __gitusr_autoload_hook")
	}
}

func TestGenerateZshEnv_UsesRemoveFirst(t *testing.T) {
	script := GenerateZshEnv()
	if !strings.Contains(script, "add-zsh-hook -D chpwd") {
		t.Error("generated script should remove old hook first with -D flag")
	}
}

func TestGenerateZshEnv_ChecksRC(t *testing.T) {
	script := GenerateZshEnv()
	if !strings.Contains(script, ".gitusrrc") {
		t.Error("generated script should check for .gitusrrc file")
	}
	if !strings.Contains(script, "apply-rc --silent-if-unchanged") {
		t.Error("generated script should call gitusr hook apply-rc --silent-if-unchanged")
	}
}

func TestGenerateZshEnv_SingleUserCheck(t *testing.T) {
	script := GenerateZshEnv()
	if !strings.Contains(script, "gitusr hook count-users") {
		t.Error("generated script should check user count via gitusr hook count-users")
	}
	if !strings.Contains(script, "gu_count") {
		t.Error("generated script should use a gu_count variable")
	}
	if !strings.Contains(script, "-le 1") {
		t.Error("generated script should check if user count is <= 1")
	}
}
