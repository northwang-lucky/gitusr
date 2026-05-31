package hook

import (
	"strings"
	"testing"
)

func TestGenerateBashEnv_ContainsAlias(t *testing.T) {
	script := GenerateBashEnv()
	if !strings.Contains(script, "alias cd=") {
		t.Error("generated script should define an alias for cd")
	}
	if !strings.Contains(script, "__gitusrcd") {
		t.Error("generated script should reference __gitusrcd")
	}
}

func TestGenerateBashEnv_UsesBackslashCd(t *testing.T) {
	script := GenerateBashEnv()
	if !strings.Contains(script, `\cd`) {
		t.Error("generated script should use \\cd to avoid recursive alias expansion")
	}
}

func TestGenerateBashEnv_ChecksRC(t *testing.T) {
	script := GenerateBashEnv()
	if !strings.Contains(script, ".gitusrrc") {
		t.Error("generated script should check for .gitusrrc file")
	}
	if !strings.Contains(script, "apply-rc --silent-if-unchanged") {
		t.Error("generated script should call gitusr hook apply-rc --silent-if-unchanged")
	}
}

func TestGenerateBashEnv_SingleUserCheck(t *testing.T) {
	script := GenerateBashEnv()
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
