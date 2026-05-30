package store

import (
	"os"
	"path/filepath"
	"testing"

	"gitusr/internal/domain"
)

// helper to create a temp file path
func tempFilePath(t *testing.T) string {
	t.Helper()
	return filepath.Join(t.TempDir(), "users.json")
}

// helper to write raw content to a file
func writeFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}
}

func TestListEmpty(t *testing.T) {
	path := tempFilePath(t)

	store := NewJSONStore(path)
	users, err := store.List()

	if err != nil {
		t.Fatalf("List() on empty store should not error, got: %v", err)
	}

	if len(users) != 0 {
		t.Fatalf("expected empty list, got %d users", len(users))
	}
}

func TestListInvalidJSON(t *testing.T) {
	path := tempFilePath(t)
	writeFile(t, path, "not-json{{{")

	store := NewJSONStore(path)
	_, err := store.List()

	if err == nil {
		t.Fatal("List() on invalid JSON should return error")
	}
}

func TestAdd(t *testing.T) {
	path := tempFilePath(t)

	store := NewJSONStore(path)
	user := domain.User{Name: "Alice", Email: "alice@example.com"}

	err := store.Add(user)
	if err != nil {
		t.Fatalf("Add() failed: %v", err)
	}

	users, err := store.List()
	if err != nil {
		t.Fatalf("List() after Add() failed: %v", err)
	}

	if len(users) != 1 {
		t.Fatalf("expected 1 user, got %d", len(users))
	}

	if users[0].Name != "Alice" || users[0].Email != "alice@example.com" {
		t.Fatalf("unexpected user: %+v", users[0])
	}
}

func TestAddDuplicateName(t *testing.T) {
	path := tempFilePath(t)

	store := NewJSONStore(path)
	store.Add(domain.User{Name: "Bob", Email: "bob@example.com"})

	err := store.Add(domain.User{Name: "Bob", Email: "bob2@example.com"})
	if err == nil {
		t.Fatal("Add() with duplicate name should return error")
	}
}

func TestAddDuplicateEmail(t *testing.T) {
	path := tempFilePath(t)

	store := NewJSONStore(path)
	store.Add(domain.User{Name: "Bob", Email: "bob@example.com"})

	err := store.Add(domain.User{Name: "Bob2", Email: "bob@example.com"})
	if err == nil {
		t.Fatal("Add() with duplicate email should return error")
	}
}

func TestRemove(t *testing.T) {
	path := tempFilePath(t)

	store := NewJSONStore(path)
	store.Add(domain.User{Name: "A", Email: "a@example.com"})
	store.Add(domain.User{Name: "B", Email: "b@example.com"})
	store.Add(domain.User{Name: "C", Email: "c@example.com"})

	err := store.Remove(1) // remove "B"
	if err != nil {
		t.Fatalf("Remove(1) failed: %v", err)
	}

	users, err := store.List()
	if err != nil {
		t.Fatalf("List() after Remove() failed: %v", err)
	}

	if len(users) != 2 {
		t.Fatalf("expected 2 users after remove, got %d", len(users))
	}

	if users[0].Name != "A" || users[1].Name != "C" {
		t.Fatalf("unexpected users after remove: %+v", users)
	}
}

func TestRemoveInvalidIndex(t *testing.T) {
	path := tempFilePath(t)

	store := NewJSONStore(path)
	store.Add(domain.User{Name: "A", Email: "a@example.com"})

	err := store.Remove(-1)
	if err == nil {
		t.Fatal("Remove(-1) should return error")
	}

	err = store.Remove(5)
	if err == nil {
		t.Fatal("Remove(5) should return error")
	}

	// empty list - any index is invalid
	store2 := NewJSONStore(tempFilePath(t))
	err = store2.Remove(0)
	if err == nil {
		t.Fatal("Remove(0) on empty store should return error")
	}
}

func TestSaveAll(t *testing.T) {
	path := tempFilePath(t)

	store := NewJSONStore(path)
	store.Add(domain.User{Name: "X", Email: "x@example.com"})

	// Overwrite with new list
	newUsers := []domain.User{
		{Name: "P", Email: "p@example.com"},
		{Name: "Q", Email: "q@example.com"},
	}

	err := store.SaveAll(newUsers)
	if err != nil {
		t.Fatalf("SaveAll() failed: %v", err)
	}

	users, err := store.List()
	if err != nil {
		t.Fatalf("List() after SaveAll() failed: %v", err)
	}

	if len(users) != 2 {
		t.Fatalf("expected 2 users after SaveAll, got %d", len(users))
	}

	if users[0].Name != "P" || users[1].Name != "Q" {
		t.Fatalf("unexpected users after SaveAll: %+v", users)
	}
}

func TestSaveAllEmpty(t *testing.T) {
	path := tempFilePath(t)

	store := NewJSONStore(path)
	store.Add(domain.User{Name: "X", Email: "x@example.com"})

	err := store.SaveAll([]domain.User{})
	if err != nil {
		t.Fatalf("SaveAll([]) failed: %v", err)
	}

	users, err := store.List()
	if err != nil {
		t.Fatalf("List() after SaveAll([]) failed: %v", err)
	}

	if len(users) != 0 {
		t.Fatalf("expected empty list after SaveAll([]), got %d", len(users))
	}
}

func TestIsInitialized(t *testing.T) {
	path := tempFilePath(t)

	store := NewJSONStore(path)

	if store.IsInitialized() {
		t.Fatal("IsInitialized() on empty store should return false")
	}

	store.Add(domain.User{Name: "Z", Email: "z@example.com"})

	if !store.IsInitialized() {
		t.Fatal("IsInitialized() after Add() should return true")
	}
}

func TestListFileNotExist(t *testing.T) {
	path := tempFilePath(t)

	store := NewJSONStore(path)

	users, err := store.List()
	if err != nil {
		t.Fatalf("List() on non-existent file should not error, got: %v", err)
	}

	if len(users) != 0 {
		t.Fatalf("expected empty list for non-existent file, got %d", len(users))
	}
}

func TestListEmptyFile(t *testing.T) {
	path := tempFilePath(t)
	writeFile(t, path, "")

	store := NewJSONStore(path)
	_, err := store.List()
	if err == nil {
		t.Fatal("List() on empty file should return error")
	}
}

func TestRemoveLastUser(t *testing.T) {
	path := tempFilePath(t)

	store := NewJSONStore(path)
	store.Add(domain.User{Name: "Only", Email: "only@example.com"})

	err := store.Remove(0)
	if err != nil {
		t.Fatalf("Remove(0) failed: %v", err)
	}

	users, err := store.List()
	if err != nil {
		t.Fatalf("List() after removing last user failed: %v", err)
	}

	if len(users) != 0 {
		t.Fatalf("expected empty list after removing last user, got %d", len(users))
	}
}
