package cli

import (
	"bytes"
	"io"
	"os"
	"strings"
	"testing"

	"gitusr/internal/domain"
	"gitusr/internal/i18n"
)

// mockStore implements domain.UserStore for testing.
type mockStore struct {
	initialized bool
	users       []domain.User
	listErr     error
}

func (m *mockStore) List() ([]domain.User, error) {
	if m.listErr != nil {
		return nil, m.listErr
	}
	return m.users, nil
}

func (m *mockStore) Add(user domain.User) error  { return nil }
func (m *mockStore) Remove(index int) error       { return nil }
func (m *mockStore) SaveAll(users []domain.User) error { return nil }
func (m *mockStore) IsInitialized() bool          { return m.initialized }

// captureOutput captures both stdout and stderr during f execution.
func captureOutput(f func()) (stdout string, stderr string) {
	oldOut := os.Stdout
	oldErr := os.Stderr

	rOut, wOut, _ := os.Pipe()
	rErr, wErr, _ := os.Pipe()
	os.Stdout = wOut
	os.Stderr = wErr

	f()

	wOut.Close()
	wErr.Close()
	os.Stdout = oldOut
	os.Stderr = oldErr

	var bufOut, bufErr bytes.Buffer
	io.Copy(&bufOut, rOut)
	io.Copy(&bufErr, rErr)

	return bufOut.String(), bufErr.String()
}

func TestList_NotInitialized(t *testing.T) {
	i18n.ResetForTesting()
	i18n.InitWithLocale("en")

	store := &mockStore{initialized: false}
	cmd := NewListCmd(store)
	cmd.SetArgs([]string{})

	var runErr error
	stdout, stderr := captureOutput(func() {
		runErr = cmd.Execute()
	})

	if runErr == nil {
		t.Fatal("expected error when store is not initialized")
	}

	if !strings.Contains(stderr, "gitusr error:") {
		t.Errorf("stderr should contain 'gitusr error:', got %q", stderr)
	}

	if stdout != "" {
		t.Errorf("stdout should be empty, got %q", stdout)
	}
}

func TestList_Empty(t *testing.T) {
	i18n.ResetForTesting()
	i18n.InitWithLocale("en")

	store := &mockStore{initialized: true, users: []domain.User{}}
	cmd := NewListCmd(store)
	cmd.SetArgs([]string{})

	stdout, stderr := captureOutput(func() {
		err := cmd.Execute()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	if stderr != "" {
		t.Errorf("stderr should be empty, got %q", stderr)
	}

	if stdout != "" {
		t.Errorf("stdout should be empty for empty list, got %q", stdout)
	}
}

func TestList_WithUsers(t *testing.T) {
	i18n.ResetForTesting()
	i18n.InitWithLocale("en")

	store := &mockStore{
		initialized: true,
		users: []domain.User{
			{Name: "Alice", Email: "alice@example.com"},
			{Name: "Bob", Email: "bob@example.com"},
		},
	}
	cmd := NewListCmd(store)
	cmd.SetArgs([]string{})

	stdout, stderr := captureOutput(func() {
		err := cmd.Execute()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	if stderr != "" {
		t.Errorf("stderr should be empty, got %q", stderr)
	}

	// Verify formatted output contains user data
	if !strings.Contains(stdout, "Alice") {
		t.Errorf("stdout should contain 'Alice', got %q", stdout)
	}
	if !strings.Contains(stdout, "alice@example.com") {
		t.Errorf("stdout should contain 'alice@example.com', got %q", stdout)
	}
	if !strings.Contains(stdout, "Bob") {
		t.Errorf("stdout should contain 'Bob', got %q", stdout)
	}
	if !strings.Contains(stdout, "bob@example.com") {
		t.Errorf("stdout should contain 'bob@example.com', got %q", stdout)
	}

	// Verify the numbering
	if !strings.Contains(stdout, "0:") {
		t.Errorf("stdout should contain line numbering starting at 0, got %q", stdout)
	}
	if !strings.Contains(stdout, "1:") {
		t.Errorf("stdout should contain line numbering, got %q", stdout)
	}
}

func TestList_AliasLS(t *testing.T) {
	i18n.ResetForTesting()
	i18n.InitWithLocale("en")

	store := &mockStore{initialized: true, users: []domain.User{}}
	cmd := NewListCmd(store)

	found := false
	for _, a := range cmd.Aliases {
		if a == "ls" {
			found = true
			break
		}
	}

	if !found {
		t.Errorf("expected 'ls' alias, got aliases: %v", cmd.Aliases)
	}
}

func TestList_En(t *testing.T) {
	i18n.ResetForTesting()
	i18n.InitWithLocale("en")

	store := &mockStore{
		initialized: true,
		users: []domain.User{
			{Name: "Alice", Email: "alice@example.com"},
			{Name: "Bob", Email: "bob@example.com"},
		},
	}
	cmd := NewListCmd(store)
	cmd.SetArgs([]string{})

	stdout, stderr := captureOutput(func() {
		err := cmd.Execute()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	if stderr != "" {
		t.Errorf("stderr should be empty, got %q", stderr)
	}

	// Verify English Short description
	if cmd.Short != "Show all saved users" {
		t.Errorf("Short = %q, want %q", cmd.Short, "Show all saved users")
	}

	// Verify English formatted output
	lines := strings.Split(strings.TrimSuffix(stdout, "\n"), "\n")
	if len(lines) != 2 {
		t.Fatalf("expected 2 lines, got %d", len(lines))
	}

	expected0 := "0: Name: Alice | Email: alice@example.com"
	expected1 := "1: Name: Bob   | Email: bob@example.com"

	if lines[0] != expected0 {
		t.Errorf("line 0:\nexpected %q\ngot      %q", expected0, lines[0])
	}
	if lines[1] != expected1 {
		t.Errorf("line 1:\nexpected %q\ngot      %q", expected1, lines[1])
	}
}

func TestList_ZhCN(t *testing.T) {
	i18n.ResetForTesting()
	i18n.InitWithLocale("zh-CN")

	store := &mockStore{
		initialized: true,
		users: []domain.User{
			{Name: "Alice", Email: "alice@example.com"},
			{Name: "Bob", Email: "bob@example.com"},
		},
	}
	cmd := NewListCmd(store)
	cmd.SetArgs([]string{})

	stdout, stderr := captureOutput(func() {
		err := cmd.Execute()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	if stderr != "" {
		t.Errorf("stderr should be empty, got %q", stderr)
	}

	// Verify Chinese Short description
	if cmd.Short != "显示所有已保存用户" {
		t.Errorf("Short = %q, want %q", cmd.Short, "显示所有已保存用户")
	}

	// Verify Chinese formatted output
	lines := strings.Split(strings.TrimSuffix(stdout, "\n"), "\n")
	if len(lines) != 2 {
		t.Fatalf("expected 2 lines, got %d", len(lines))
	}

	expected0 := "0：姓名：Alice | 邮箱：alice@example.com"
	expected1 := "1：姓名：Bob   | 邮箱：bob@example.com"

	if lines[0] != expected0 {
		t.Errorf("line 0:\nexpected %q\ngot      %q", expected0, lines[0])
	}
	if lines[1] != expected1 {
		t.Errorf("line 1:\nexpected %q\ngot      %q", expected1, lines[1])
	}
}

func TestList_Empty_En(t *testing.T) {
	i18n.ResetForTesting()
	i18n.InitWithLocale("en")

	store := &mockStore{initialized: true, users: []domain.User{}}
	cmd := NewListCmd(store)
	cmd.SetArgs([]string{})

	stdout, stderr := captureOutput(func() {
		err := cmd.Execute()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	if stderr != "" {
		t.Errorf("stderr should be empty, got %q", stderr)
	}

	if stdout != "" {
		t.Errorf("stdout should be empty for empty list, got %q", stdout)
	}
}

func TestList_Empty_ZhCN(t *testing.T) {
	i18n.ResetForTesting()
	i18n.InitWithLocale("zh-CN")

	store := &mockStore{initialized: true, users: []domain.User{}}
	cmd := NewListCmd(store)
	cmd.SetArgs([]string{})

	stdout, stderr := captureOutput(func() {
		err := cmd.Execute()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	if stderr != "" {
		t.Errorf("stderr should be empty, got %q", stderr)
	}

	if stdout != "" {
		t.Errorf("stdout should be empty for empty list, got %q", stdout)
	}
}
