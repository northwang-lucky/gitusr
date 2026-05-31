package cli

import (
	"strings"
	"testing"

	"gitusr/internal/i18n"
)

func TestRootHelp_En(t *testing.T) {
	i18n.ResetForTesting()
	i18n.InitWithLocale("en")

	store := &mockStore{}
	cmd := NewRootCmd(store, "gitusr")

	stdout, stderr, err := executeCmd(cmd, "--help")
	if err != nil {
		t.Fatalf("NewRootCmd().Execute(--help) unexpected error: %v", err)
	}

	if !strings.Contains(stdout, "A CLI that allows you to switch git users.") {
		t.Errorf("expected English Short description in help output, got: %q", stdout)
	}
	if stderr != "" {
		t.Errorf("expected empty stderr, got: %q", stderr)
	}
}

func TestRootHelp_ZhCN(t *testing.T) {
	i18n.ResetForTesting()
	i18n.InitWithLocale("zh-CN")

	store := &mockStore{}
	cmd := NewRootCmd(store, "gitusr")

	stdout, stderr, err := executeCmd(cmd, "--help")
	if err != nil {
		t.Fatalf("NewRootCmd().Execute(--help) unexpected error: %v", err)
	}

	if !strings.Contains(stdout, "用于切换 git 用户的 CLI 工具。") {
		t.Errorf("expected Chinese Short description in help output, got: %q", stdout)
	}
	if stderr != "" {
		t.Errorf("expected empty stderr, got: %q", stderr)
	}
}
