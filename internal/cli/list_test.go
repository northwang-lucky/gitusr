package cli

import (
	"bytes"
	"io"
	"os"
	"strings"
	"testing"

	"gitusr/internal/domain"
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
