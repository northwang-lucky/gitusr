package hook

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/northwang-lucky/gitusr/internal/domain"
	"github.com/northwang-lucky/gitusr/internal/gitcmd"
)

func TestApplyRC_WithRC(t *testing.T) {
	dir := t.TempDir()
	initGitRepo(t, dir)
	t.Chdir(dir)

	rcPath := filepath.Join(dir, ".gitusrrc")
	if err := os.WriteFile(rcPath, []byte(`{"name":"Alice","email":"alice@example.com"}`), 0644); err != nil {
		t.Fatal(err)
	}

	store := &hookTestStore{
		users: []domain.User{
			{Name: "Alice", Email: "alice@example.com"},
		},
	}

	if err := GenerateApplyRC(store); err != nil {
		t.Fatalf("GenerateApplyRC() returned error: %v", err)
	}

	name, err := gitcmd.GetConfig("name", false)
	if err != nil {
		t.Fatalf("git config user.name: %v", err)
	}
	if name != "Alice" {
		t.Errorf("user.name = %q, want %q", name, "Alice")
	}
}

func TestApplyRC_NoRC(t *testing.T) {
	dir := t.TempDir()
	t.Chdir(dir)

	store := &hookTestStore{
		users: []domain.User{
			{Name: "Bob", Email: "bob@example.com"},
		},
	}

	if err := GenerateApplyRC(store); err != nil {
		t.Fatalf("GenerateApplyRC() should return nil when no .gitusrrc, got: %v", err)
	}
}
