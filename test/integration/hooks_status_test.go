package integration

import (
	"strings"
	"testing"
)

// ---------------------------------------------------------------------------
// TestHooksStatusWorkflow — hooks status 随 install/disable/enable/uninstall 流转
// ---------------------------------------------------------------------------
func TestHooksStatusWorkflow(t *testing.T) {
	homeDir := t.TempDir()
	xdgDataHome := t.TempDir()

	env := map[string]string{
		"HOME":          homeDir,
		"XDG_DATA_HOME": xdgDataHome,
	}
	writeJSONFile(t, gitusrStore(xdgDataHome), []map[string]string{
		{"name": "Alice", "email": "alice@example.com"},
		{"name": "Bob", "email": "bob@example.com"},
	})

	// 安装前：三种 hook 都应为未安装
	out, err := runGitusr(t, env, "hooks", "status")
	if err != nil {
		t.Fatalf("hooks status failed: %v\n%s", err, out)
	}
	if got := strings.Count(out, "not installed"); got != 3 {
		t.Fatalf("expected 3 not-installed lines before install, got %d in %q", got, out)
	}

	// 安装后：三行全部 enabled
	if out, err := runGitusr(t, env, "hooks", "install"); err != nil {
		t.Fatalf("hooks install failed: %v\n%s", err, out)
	}
	out, err = runGitusr(t, env, "hooks", "status")
	if err != nil {
		t.Fatalf("hooks status failed: %v\n%s", err, out)
	}
	for _, want := range []string{"clone: enabled", "commit: enabled", "cd: enabled"} {
		if !strings.Contains(out, want) {
			t.Errorf("after install: missing %q in %q", want, out)
		}
	}

	// 禁用 commit 后：仅该行变化
	if out, err := runGitusr(t, env, "hooks", "disable", "commit"); err != nil {
		t.Fatalf("hooks disable commit failed: %v\n%s", err, out)
	}
	out, err = runGitusr(t, env, "hooks", "status")
	if err != nil {
		t.Fatalf("hooks status failed: %v\n%s", err, out)
	}
	if !strings.Contains(out, "commit: disabled") || !strings.Contains(out, "clone: enabled") {
		t.Errorf("after disable commit: unexpected output %q", out)
	}

	// 重新启用后恢复 enabled
	if out, err := runGitusr(t, env, "hooks", "enable", "commit"); err != nil {
		t.Fatalf("hooks enable commit failed: %v\n%s", err, out)
	}
	out, err = runGitusr(t, env, "hooks", "status")
	if err != nil {
		t.Fatalf("hooks status failed: %v\n%s", err, out)
	}
	if !strings.Contains(out, "commit: enabled") {
		t.Errorf("after enable commit: unexpected output %q", out)
	}

	// 卸载后回到未安装
	if out, err := runGitusr(t, env, "hooks", "uninstall"); err != nil {
		t.Fatalf("hooks uninstall failed: %v\n%s", err, out)
	}
	out, err = runGitusr(t, env, "hooks", "status")
	if err != nil {
		t.Fatalf("hooks status failed: %v\n%s", err, out)
	}
	if got := strings.Count(out, "not installed"); got != 3 {
		t.Errorf("after uninstall: expected 3 not-installed lines, got %d in %q", got, out)
	}
}
