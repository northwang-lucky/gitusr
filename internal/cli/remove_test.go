package cli

import (
	"strings"
	"testing"

	"gitusr/internal/domain"
	"gitusr/internal/store"
)

func TestRemove_ByIndex(t *testing.T) {
	tmpDir := t.TempDir()
	jsonStore := store.NewJSONStore(tmpDir + "/users.json")

	jsonStore.Add(domain.User{Name: "Alice", Email: "alice@example.com"})
	jsonStore.Add(domain.User{Name: "Bob", Email: "bob@example.com"})

	cmd := NewRemoveCmd(jsonStore)
	stdout, _, err := executeCmd(cmd, "--index", "0")

	if err != nil {
		t.Fatalf("NewRemoveCmd().Execute(--index 0) unexpected error: %v", err)
	}

	if !strings.Contains(stdout, "Success!") {
		t.Errorf("expected 'Success!', got %q", stdout)
	}
	if !strings.Contains(stdout, "Alice") {
		t.Errorf("expected 'Alice' in output, got %q", stdout)
	}
	if !strings.Contains(stdout, "alice@example.com") {
		t.Errorf("expected 'alice@example.com' in output, got %q", stdout)
	}

	// Verify user was actually removed
	users, err := jsonStore.List()
	if err != nil {
		t.Fatalf("List() after remove failed: %v", err)
	}
	if len(users) != 1 {
		t.Fatalf("expected 1 user after remove, got %d", len(users))
	}
	if users[0].Name != "Bob" {
		t.Fatalf("expected Bob to remain, got %s", users[0].Name)
	}
}

func TestRemove_ByEmail(t *testing.T) {
	tmpDir := t.TempDir()
	jsonStore := store.NewJSONStore(tmpDir + "/users.json")

	jsonStore.Add(domain.User{Name: "Alice", Email: "alice@example.com"})
	jsonStore.Add(domain.User{Name: "Bob", Email: "bob@example.com"})

	cmd := NewRemoveCmd(jsonStore)
	stdout, _, err := executeCmd(cmd, "--email", "bob@example.com")

	if err != nil {
		t.Fatalf("NewRemoveCmd().Execute(--email bob@example.com) unexpected error: %v", err)
	}

	if !strings.Contains(stdout, "Success!") {
		t.Errorf("expected 'Success!', got %q", stdout)
	}
	if !strings.Contains(stdout, "Bob") {
		t.Errorf("expected 'Bob' in output, got %q", stdout)
	}
}

func TestRemove_NotInitialized(t *testing.T) {
	tmpDir := t.TempDir()
	jsonStore := store.NewJSONStore(tmpDir + "/users.json")

	cmd := NewRemoveCmd(jsonStore)
	_, _, err := executeCmd(cmd, "--index", "0")

	if err == nil {
		t.Fatal("NewRemoveCmd().Execute() on uninitialized store should return error")
	}

	if !strings.Contains(err.Error(), "initialized") {
		t.Errorf("error should mention 'initialized', got: %q", err.Error())
	}
}

func TestRemove_UserNotFound(t *testing.T) {
	tmpDir := t.TempDir()
	jsonStore := store.NewJSONStore(tmpDir + "/users.json")

	jsonStore.Add(domain.User{Name: "Alice", Email: "alice@example.com"})

	cmd := NewRemoveCmd(jsonStore)
	_, _, err := executeCmd(cmd, "--email", "nonexistent@example.com")

	if err == nil {
		t.Fatal("NewRemoveCmd().Execute(--email nonexistent) should return error")
	}
}

func TestRemove_Alias(t *testing.T) {
	tmpDir := t.TempDir()
	jsonStore := store.NewJSONStore(tmpDir + "/users.json")

	cmd := NewRemoveCmd(jsonStore)

	found := false
	for _, alias := range cmd.Aliases {
		if alias == "rm" {
			found = true
			break
		}
	}

	if !found {
		t.Error("NewRemoveCmd() should have alias 'rm'")
	}
}

func TestRemove_FlagRegistration(t *testing.T) {
	tmpDir := t.TempDir()
	jsonStore := store.NewJSONStore(tmpDir + "/users.json")

	cmd := NewRemoveCmd(jsonStore)

	nameFlag := cmd.Flags().Lookup("name")
	if nameFlag == nil {
		t.Fatal("NewRemoveCmd() should have --name flag")
	}
	if nameFlag.Shorthand != "n" {
		t.Errorf("--name flag shorthand = %q, want %q", nameFlag.Shorthand, "n")
	}

	emailFlag := cmd.Flags().Lookup("email")
	if emailFlag == nil {
		t.Fatal("NewRemoveCmd() should have --email flag")
	}
	if emailFlag.Shorthand != "e" {
		t.Errorf("--email flag shorthand = %q, want %q", emailFlag.Shorthand, "e")
	}

	indexFlag := cmd.Flags().Lookup("index")
	if indexFlag == nil {
		t.Fatal("NewRemoveCmd() should have --index flag")
	}
	if indexFlag.Shorthand != "i" {
		t.Errorf("--index flag shorthand = %q, want %q", indexFlag.Shorthand, "i")
	}
}
