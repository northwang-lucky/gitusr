package integration

import (
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// gitLocalConfig 读取仓库级 git config；HOME/XDG_CONFIG_HOME 都指向仓库目录，
// 保证不会读到任何真实用户的全局配置。key 不存在时 git config 退出码为 1，
// 按空字符串处理。
func gitLocalConfig(t *testing.T, dir, key string) string {
	t.Helper()

	cmd := exec.Command("git", "config", "--local", key)
	cmd.Dir = dir
	cmd.Env = []string{
		"HOME=" + dir,
		"XDG_CONFIG_HOME=" + dir,
		"GIT_CONFIG_GLOBAL=" + filepath.Join(dir, "gitconfig-none"),
		"PATH=" + os.Getenv("PATH"),
	}
	out, err := cmd.Output()
	if err != nil {
		var ee *exec.ExitError
		if errors.As(err, &ee) && ee.ExitCode() == 1 {
			return ""
		}
		t.Fatalf("git config --local %s failed in %s: %v", key, dir, err)
	}
	return strings.TrimSpace(string(out))
}

// setupCloneWrapperEnv 搭建沙箱环境：两个已保存用户（wrapper 仅在用户数 > 1
// 时启用拦截）、一条 github.com host 规则，并安装真实 bash wrapper。
// 返回 env、wrapper 路径、以及暴露了 gitusr 可执行文件的 bin 目录。
func setupCloneWrapperEnv(t *testing.T) (env map[string]string, wrapperPath, binDir string) {
	t.Helper()

	homeDir := t.TempDir()
	xdgDataHome := t.TempDir()
	configHome := t.TempDir()

	env = map[string]string{
		"HOME":              homeDir,
		"XDG_DATA_HOME":     xdgDataHome,
		"XDG_CONFIG_HOME":   configHome,
		"GIT_CONFIG_GLOBAL": filepath.Join(configHome, "git", "config"),
	}
	// GIT_CONFIG_GLOBAL 指向的文件不会由 git 自动建父目录
	if err := os.MkdirAll(filepath.Join(configHome, "git"), 0o755); err != nil {
		t.Fatalf("create sandbox git config dir: %v", err)
	}

	writeJSONFile(t, gitusrStore(xdgDataHome), []map[string]string{
		{"name": "Alice", "email": "alice@example.com"},
		{"name": "Bob", "email": "bob@example.com"},
	})
	writeJSONFile(t, hostsFile(xdgDataHome), []map[string]string{
		{"host": "github.com", "email": "alice@example.com"},
	})

	if out, err := runGitusr(t, env, "hooks", "install"); err != nil {
		t.Fatalf("hooks install failed: %v\n%s", err, out)
	}
	wrapperPath = filepath.Join(xdgDataHome, "gitusr", "hooks", "git-wrapper.sh")

	// wrapper 按名字调用 gitusr，构建产物需要以该名出现在 PATH 中
	binDir = t.TempDir()
	if err := os.Symlink(gitusrBin, filepath.Join(binDir, "gitusr")); err != nil {
		t.Fatalf("expose gitusr on PATH: %v", err)
	}
	return env, wrapperPath, binDir
}

// runWrapperClone 在真实 bash 中 source 已安装的 wrapper 并执行 git clone。
// 通过全局 insteadOf 把 cloneURL 改写到本地 bare 仓库，测试全程不联网；
// git() 函数与 command git 都走沙箱 HOME/XDG。返回 clone 发生的目录。
func runWrapperClone(t *testing.T, env map[string]string, wrapperPath, binDir, cloneURL string) string {
	t.Helper()

	bash, err := exec.LookPath("bash")
	if err != nil {
		t.Skip("bash not available in PATH")
	}

	// 本地 bare 仓库，作为 cloneURL 的改写目标
	srcDir := t.TempDir()
	srcPath := filepath.Join(srcDir, "b.git")
	runGit(t, srcDir, env, "init", "-b", "main", "--bare", srcPath)
	runGit(t, srcDir, env, "config", "--global", "url."+srcPath+".insteadOf", cloneURL)

	workRoot := t.TempDir()
	script := "set -e\nsource \"$1\"\ncd \"$2\"\ngit clone \"$3\"\n"
	cmd := exec.Command(bash, "-c", script, "bash", wrapperPath, workRoot, cloneURL)
	cmd.Env = append(os.Environ(),
		"PATH="+binDir+string(os.PathListSeparator)+os.Getenv("PATH"),
		"HOME="+env["HOME"],
		"XDG_DATA_HOME="+env["XDG_DATA_HOME"],
		"XDG_CONFIG_HOME="+env["XDG_CONFIG_HOME"],
		"GIT_CONFIG_GLOBAL="+env["GIT_CONFIG_GLOBAL"],
	)
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("wrapper git clone failed: %v\n%s", err, out)
	}
	return workRoot
}

// ---------------------------------------------------------------------------
// clone wrapper host 规则 E2E（回归：wrapper 曾把 "remote.origin.url <url>"
// 整行传给 apply-host，https URL 解析失败导致 host 规则静默不生效）
// ---------------------------------------------------------------------------

func TestCloneWrapperHostRuleAppliesForHTTPS(t *testing.T) {
	env, wrapperPath, binDir := setupCloneWrapperEnv(t)

	workRoot := runWrapperClone(t, env, wrapperPath, binDir, "https://github.com/a/b.git")

	cloned := filepath.Join(workRoot, "b")
	if got := gitLocalConfig(t, cloned, "user.name"); got != "Alice" {
		t.Errorf("cloned repo user.name = %q, want Alice (host rule must apply for https URLs)", got)
	}
	if got := gitLocalConfig(t, cloned, "user.email"); got != "alice@example.com" {
		t.Errorf("cloned repo user.email = %q, want alice@example.com", got)
	}
}

func TestCloneWrapperHostRuleAppliesForScpSSH(t *testing.T) {
	env, wrapperPath, binDir := setupCloneWrapperEnv(t)

	workRoot := runWrapperClone(t, env, wrapperPath, binDir, "git@github.com:a/b.git")

	cloned := filepath.Join(workRoot, "b")
	if got := gitLocalConfig(t, cloned, "user.name"); got != "Alice" {
		t.Errorf("cloned repo user.name = %q, want Alice (host rule must apply for scp-like ssh URLs)", got)
	}
}
