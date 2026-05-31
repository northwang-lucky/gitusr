package integration

import (
	"strings"
	"testing"
)

// ---------------------------------------------------------------------------
// TestChineseLocale — verify Chinese (zh-CN) translations are used
// ---------------------------------------------------------------------------
func TestChineseLocale(t *testing.T) {
	homeDir := t.TempDir()
	xdgDataHome := t.TempDir()

	// Use Chinese locale
	env := map[string]string{
		"HOME":          homeDir,
		"XDG_DATA_HOME": xdgDataHome,
		"GITUSR_LANG":   "zh-CN",
	}

	// Initialize store with one user
	storePath := gitusrStore(xdgDataHome)
	users := []map[string]string{
		{"name": "测试用户", "email": "test@example.com"},
	}
	writeJSONFile(t, storePath, users)

	// Test list command output in Chinese
	output, err := runGitusr(t, env, "list")
	if err != nil {
		t.Fatalf("list failed: %v\noutput: %s", err, output)
	}

	// Verify Chinese characters are present in output
	// "姓名" means "Name" in Chinese
	// "邮箱" means "Email" in Chinese
	if !strings.Contains(output, "姓名") {
		t.Errorf("Chinese output should contain '姓名' (Name), got: %s", output)
	}
	if !strings.Contains(output, "邮箱") {
		t.Errorf("Chinese output should contain '邮箱' (Email), got: %s", output)
	}

	// Verify the msgID is NOT in output (would indicate translation missing)
	if strings.Contains(output, "format.userlist_row") {
		t.Errorf("Output should not contain msgID 'format.userlist_row', got: %s", output)
	}
}

// ---------------------------------------------------------------------------
// TestChineseLocaleInit — verify init command Chinese output
// ---------------------------------------------------------------------------
func TestChineseLocaleInit(t *testing.T) {
	homeDir := t.TempDir()
	xdgDataHome := t.TempDir()

	env := map[string]string{
		"HOME":          homeDir,
		"XDG_DATA_HOME": xdgDataHome,
		"GITUSR_LANG":   "zh-CN",
	}

	// Pre-configure git global user
	runGit(t, homeDir, env, "config", "--global", "user.name", "测试用户")
	runGit(t, homeDir, env, "config", "--global", "user.email", "test@example.com")

	// Run init --force
	output, err := runGitusr(t, env, "init", "--force")
	if err != nil {
		t.Fatalf("init --force failed: %v\noutput: %s", err, output)
	}

	// Verify Chinese output
	// "成功" means "Success" in Chinese
	if !strings.Contains(output, "成功") {
		t.Errorf("Chinese output should contain '成功' (Success), got: %s", output)
	}

	// Verify the user name is present
	if !strings.Contains(output, "测试用户") {
		t.Errorf("Output should contain user name '测试用户', got: %s", output)
	}
}

// ---------------------------------------------------------------------------
// TestChineseLocaleErrors — verify error messages in Chinese
// ---------------------------------------------------------------------------
func TestChineseLocaleErrors(t *testing.T) {
	homeDir := t.TempDir()
	xdgDataHome := t.TempDir()

	env := map[string]string{
		"HOME":          homeDir,
		"XDG_DATA_HOME": xdgDataHome,
		"GITUSR_LANG":   "zh-CN",
	}

	// Try to use without initialized store
	repoDir := t.TempDir()
	runGit(t, repoDir, nil, "init")

	output, err := runGitusrInDir(t, env, repoDir, "use", "--index", "0")
	if err == nil {
		t.Fatalf("use without init should fail")
	}

	// Verify Chinese error message (should contain "存储未初始化" or similar)
	// If translation is missing, we'll see the msgID instead
	chineseIndicators := []string{"存储", "初始化", "暂存", "请先"}
	foundChinese := false
	for _, indicator := range chineseIndicators {
		if strings.Contains(output, indicator) {
			foundChinese = true
			break
		}
	}
	if !foundChinese && strings.Contains(output, "cli.error.store_not_init") {
		t.Errorf("Error message not translated: %s", output)
	}
}

// ---------------------------------------------------------------------------
// TestChineseLocaleAddDuplicate — verify duplicate error in Chinese
// ---------------------------------------------------------------------------
func TestChineseLocaleAddDuplicate(t *testing.T) {
	homeDir := t.TempDir()
	xdgDataHome := t.TempDir()

	env := map[string]string{
		"HOME":          homeDir,
		"XDG_DATA_HOME": xdgDataHome,
		"GITUSR_LANG":   "zh-CN",
	}

	storePath := gitusrStore(xdgDataHome)
	users := []map[string]string{
		{"name": "用户A", "email": "a@example.com"},
	}
	writeJSONFile(t, storePath, users)

	// Try to add duplicate
	output, err := runGitusr(t, env, "add", "--name", "用户A", "--email", "b@example.com")
	if err == nil {
		t.Fatalf("add duplicate should fail")
	}

	// Verify Chinese error message contains "已存在" (already exists)
	if !strings.Contains(output, "已存在") {
		t.Logf("Output: %s", output)
		t.Errorf("Duplicate error should contain '已存在' (already exists)")
	}
}
