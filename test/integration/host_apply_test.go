package integration

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// hostsFile returns the hosts.json path under the given XDG data home.
func hostsFile(xdgDataHome string) string {
	return filepath.Join(xdgDataHome, "gitusr", "hosts.json")
}

// ---------------------------------------------------------------------------
// TestHostsSetThenApplyHost — hosts set 后 apply-host 命中并设置 git config
// ---------------------------------------------------------------------------
func TestHostsSetThenApplyHost(t *testing.T) {
	homeDir := t.TempDir()
	xdgDataHome := t.TempDir()

	env := map[string]string{
		"HOME":          homeDir,
		"XDG_DATA_HOME": xdgDataHome,
	}

	// 预置一个已保存用户
	storePath := gitusrStore(xdgDataHome)
	writeJSONFile(t, storePath, []map[string]string{
		{"name": "Alice", "email": "alice@example.com"},
	})

	// 配置 host 规则
	if _, err := runGitusr(t, env, "hosts", "set", "github.com", "alice@example.com"); err != nil {
		t.Fatalf("hosts set failed: %v", err)
	}

	repoDir := t.TempDir()
	runGit(t, repoDir, nil, "init")

	// clone 场景下由 wrapper 调用 apply-host(传仓库 remote URL)
	output, err := runGitusrInDir(t, env, repoDir, "hooks", "apply-host", "https://github.com/a/b.git")
	if err != nil {
		t.Fatalf("hooks apply-host failed: %v\noutput: %s", err, output)
	}
	if !strings.Contains(output, "Alice") {
		t.Errorf("expected user info output, got: %q", output)
	}

	if name := gitConfig(t, repoDir, "name"); name != "Alice" {
		t.Errorf("git config user.name = %q, want 'Alice'", name)
	}
	if email := gitConfig(t, repoDir, "email"); email != "alice@example.com" {
		t.Errorf("git config user.email = %q, want 'alice@example.com'", email)
	}
}

// ---------------------------------------------------------------------------
// TestHooksApplyHostNoMatch — host 无匹配时静默跳过
// ---------------------------------------------------------------------------
func TestHooksApplyHostNoMatch(t *testing.T) {
	homeDir := t.TempDir()
	xdgDataHome := t.TempDir()

	env := map[string]string{
		"HOME":          homeDir,
		"XDG_DATA_HOME": xdgDataHome,
	}

	writeJSONFile(t, gitusrStore(xdgDataHome), []map[string]string{
		{"name": "Alice", "email": "alice@example.com"},
	})
	writeJSONFile(t, hostsFile(xdgDataHome), []map[string]string{
		{"host": "github.com", "email": "alice@example.com"},
	})

	repoDir := t.TempDir()
	runGit(t, repoDir, nil, "init")

	output, err := runGitusrInDir(t, env, repoDir, "hooks", "apply-host", "https://example.com/a/b.git")
	if err != nil {
		t.Fatalf("hooks apply-host failed: %v\noutput: %s", err, output)
	}
	if output != "" {
		t.Errorf("expected silent skip, got output: %q", output)
	}
}

// ---------------------------------------------------------------------------
// TestHooksApplyHostMissingUserWarning — 规则引用的用户被删除时输出 warning
// ---------------------------------------------------------------------------
func TestHooksApplyHostMissingUserWarning(t *testing.T) {
	homeDir := t.TempDir()
	xdgDataHome := t.TempDir()

	env := map[string]string{
		"HOME":          homeDir,
		"XDG_DATA_HOME": xdgDataHome,
	}

	writeJSONFile(t, gitusrStore(xdgDataHome), []map[string]string{
		{"name": "Alice", "email": "alice@example.com"},
	})
	writeJSONFile(t, hostsFile(xdgDataHome), []map[string]string{
		{"host": "github.com", "email": "deleted@example.com"},
	})

	repoDir := t.TempDir()
	runGit(t, repoDir, nil, "init")

	output, err := runGitusrInDir(t, env, repoDir, "hooks", "apply-host", "https://github.com/a/b.git")
	if err != nil {
		t.Fatalf("hooks apply-host should not fail: %v\noutput: %s", err, output)
	}
	if !strings.Contains(output, "deleted@example.com") {
		t.Errorf("expected warning about missing user, got output: %q", output)
	}
}

// ---------------------------------------------------------------------------
// TestHostsListAndMove — hosts list/move 工作流闭环
// ---------------------------------------------------------------------------
func TestHostsListAndMove(t *testing.T) {
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

	for _, set := range [][]string{
		{"hosts", "set", "github.com", "alice@example.com"},
		{"hosts", "set", "code.byted.org", "bob@example.com"},
	} {
		if _, err := runGitusr(t, env, set...); err != nil {
			t.Fatalf("hosts set failed: %v", err)
		}
	}

	// 把 code.byted.org 移到首位,让它成为同优先级下的先匹配规则
	if _, err := runGitusr(t, env, "hosts", "move", "code.byted.org", "first"); err != nil {
		t.Fatalf("hosts move failed: %v", err)
	}

	output, err := runGitusr(t, env, "hosts", "list")
	if err != nil {
		t.Fatalf("hosts list failed: %v", err)
	}
	if !strings.Contains(output, "0: code.byted.org") {
		t.Errorf("expected code.byted.org first, got: %q", output)
	}

	// 文件内容校验顺序
	var rules []map[string]string
	readJSONFile(t, hostsFile(xdgDataHome), &rules)
	if len(rules) != 2 || rules[0]["host"] != "code.byted.org" {
		t.Errorf("unexpected hosts.json content: %+v", rules)
	}
}

// ---------------------------------------------------------------------------
// TestHostsSetValidatesEmail — host 规则必须引用已保存用户
// ---------------------------------------------------------------------------
func TestHostsSetValidatesEmail(t *testing.T) {
	homeDir := t.TempDir()
	xdgDataHome := t.TempDir()

	env := map[string]string{
		"HOME":          homeDir,
		"XDG_DATA_HOME": xdgDataHome,
	}

	writeJSONFile(t, gitusrStore(xdgDataHome), []map[string]string{
		{"name": "Alice", "email": "alice@example.com"},
	})

	output, err := runGitusr(t, env, "hosts", "set", "github.com", "ghost@example.com")
	if err == nil {
		t.Fatalf("expected error for unknown email, got output: %s", output)
	}

	if _, err := os.Stat(hostsFile(xdgDataHome)); err == nil {
		t.Error("hosts.json should not be created on validation failure")
	}
}
