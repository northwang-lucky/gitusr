package hook

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/northwang-lucky/gitusr/internal/domain"
	"github.com/northwang-lucky/gitusr/internal/gitcmd"
)

// hookTestStore implements domain.UserStore for testing.
type hookTestStore struct {
	users []domain.User
}

func (m *hookTestStore) List() ([]domain.User, error) {
	return m.users, nil
}

func (m *hookTestStore) Add(user domain.User) error               { return nil }
func (m *hookTestStore) Remove(index int) error                    { return nil }
func (m *hookTestStore) SaveAll(users []domain.User) error         { return nil }
func (m *hookTestStore) IsInitialized() bool                       { return true }

// initGitRepo initializes a test git repository with a minimal user identity.
func initGitRepo(t *testing.T, dir string) {
	t.Helper()
	runGit(t, dir, "init")
	runGit(t, dir, "config", "user.name", "Test User")
	runGit(t, dir, "config", "user.email", "test@example.com")
	runGit(t, dir, "commit", "--allow-empty", "-m", "initial")
}

func runGit(t *testing.T, dir string, args ...string) {
	t.Helper()
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git %s failed in %s: %v\n%s", strings.Join(args, " "), dir, err, out)
	}
}

func TestParseRC_ValidJSON(t *testing.T) {
	dir := t.TempDir()
	rcPath := filepath.Join(dir, ".gitusrrc")
	if err := os.WriteFile(rcPath, []byte(`{"name":"Alice","email":"alice@example.com"}`), 0644); err != nil {
		t.Fatal(err)
	}

	rc, err := ParseRC(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rc == nil {
		t.Fatal("expected non-nil rc")
	}
	if rc.Name != "Alice" {
		t.Errorf("rc.Name = %q, want %q", rc.Name, "Alice")
	}
	if rc.Email != "alice@example.com" {
		t.Errorf("rc.Email = %q, want %q", rc.Email, "alice@example.com")
	}
}

func TestParseRC_NameOnly(t *testing.T) {
	dir := t.TempDir()
	rcPath := filepath.Join(dir, ".gitusrrc")
	if err := os.WriteFile(rcPath, []byte(`{"name":"Bob"}`), 0644); err != nil {
		t.Fatal(err)
	}

	rc, err := ParseRC(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rc == nil {
		t.Fatal("expected non-nil rc")
	}
	if rc.Name != "Bob" {
		t.Errorf("rc.Name = %q, want %q", rc.Name, "Bob")
	}
	if rc.Email != "" {
		t.Errorf("rc.Email = %q, want %q", rc.Email, "")
	}
}

func TestParseRC_InvalidJSON(t *testing.T) {
	dir := t.TempDir()
	rcPath := filepath.Join(dir, ".gitusrrc")
	if err := os.WriteFile(rcPath, []byte(`invalid json`), 0644); err != nil {
		t.Fatal(err)
	}

	rc, err := ParseRC(dir)
	if err == nil {
		t.Fatal("expected error for invalid JSON")
	}
	if rc != nil {
		t.Errorf("expected nil rc, got %v", rc)
	}
	if !strings.Contains(err.Error(), "invalid .gitusrrc JSON") {
		t.Errorf("error should contain 'invalid .gitusrrc JSON', got: %q", err.Error())
	}
}

func TestParseRC_FileNotFound(t *testing.T) {
	dir := t.TempDir()

	rc, err := ParseRC(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rc != nil {
		t.Errorf("expected nil rc, got %v", rc)
	}
}

func TestParseRC_EmptyFields(t *testing.T) {
	dir := t.TempDir()
	rcPath := filepath.Join(dir, ".gitusrrc")
	if err := os.WriteFile(rcPath, []byte(`{"name":"","email":""}`), 0644); err != nil {
		t.Fatal(err)
	}

	rc, err := ParseRC(dir)
	if err == nil {
		t.Fatal("expected error for empty name and email")
	}
	if rc != nil {
		t.Errorf("expected nil rc, got %v", rc)
	}
	if !strings.Contains(err.Error(), "must have at least name or email") {
		t.Errorf("error should contain 'must have at least name or email', got: %q", err.Error())
	}
}

func TestMatchAndApplyRC_EmailMatch(t *testing.T) {
	dir := t.TempDir()
	initGitRepo(t, dir)
	t.Chdir(dir)

	store := &hookTestStore{
		users: []domain.User{
			{Name: "Alice", Email: "alice@example.com"},
		},
	}

	rc := &GitusrRC{Email: "alice@example.com"}

	if err := MatchAndApplyRC(store, rc); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify git config was applied
	name, err := gitcmd.GetConfig("name", false)
	if err != nil {
		t.Fatalf("git config user.name: %v", err)
	}
	if name != "Alice" {
		t.Errorf("user.name = %q, want %q", name, "Alice")
	}

	email, err := gitcmd.GetConfig("email", false)
	if err != nil {
		t.Fatalf("git config user.email: %v", err)
	}
	if email != "alice@example.com" {
		t.Errorf("user.email = %q, want %q", email, "alice@example.com")
	}
}

func TestMatchAndApplyRC_NameMatch(t *testing.T) {
	dir := t.TempDir()
	initGitRepo(t, dir)
	t.Chdir(dir)

	store := &hookTestStore{
		users: []domain.User{
			{Name: "Bob", Email: "bob@example.com"},
		},
	}

	rc := &GitusrRC{Name: "Bob"}

	if err := MatchAndApplyRC(store, rc); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	name, err := gitcmd.GetConfig("name", false)
	if err != nil {
		t.Fatalf("git config user.name: %v", err)
	}
	if name != "Bob" {
		t.Errorf("user.name = %q, want %q", name, "Bob")
	}
}

func TestMatchAndApplyRC_EmailPriority(t *testing.T) {
	dir := t.TempDir()
	initGitRepo(t, dir)
	t.Chdir(dir)

	store := &hookTestStore{
		users: []domain.User{
			{Name: "Alice", Email: "alice@example.com"},
			{Name: "Alice2", Email: "alice2@example.com"},
		},
	}

	// Both name and email provided; email should take priority
	rc := &GitusrRC{
		Name:  "Alice",
		Email: "alice2@example.com",
	}

	if err := MatchAndApplyRC(store, rc); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	name, err := gitcmd.GetConfig("name", false)
	if err != nil {
		t.Fatalf("git config user.name: %v", err)
	}
	if name != "Alice2" {
		t.Errorf("user.name = %q, want %q", name, "Alice2")
	}
}

func TestMatchAndApplyRC_NoMatch(t *testing.T) {
	dir := t.TempDir()
	initGitRepo(t, dir)
	t.Chdir(dir)

	store := &hookTestStore{
		users: []domain.User{
			{Name: "Alice", Email: "alice@example.com"},
		},
	}

	rc := &GitusrRC{Email: "nonexistent@example.com"}

	err := MatchAndApplyRC(store, rc)
	if err == nil {
		t.Fatal("expected error for no match")
	}

	want := "user specified in .gitusrrc not found in saved users"
	if err.Error() != want {
		t.Errorf("error = %q, want %q", err.Error(), want)
	}
}
