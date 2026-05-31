package sel_test

import (
	"errors"
	"testing"

	"github.com/northwang-lucky/gitusr/internal/domain"

	sel "github.com/northwang-lucky/gitusr/internal/select"
)

// mockStore implements domain.UserStore for testing.
type mockStore struct {
	users   []domain.User
	listErr error
}

func (m *mockStore) List() ([]domain.User, error) {
	if m.listErr != nil {
		return nil, m.listErr
	}
	return m.users, nil
}

func (m *mockStore) Add(domain.User) error   { return nil }
func (m *mockStore) Remove(int) error        { return nil }
func (m *mockStore) SaveAll([]domain.User) error { return nil }
func (m *mockStore) IsInitialized() bool     { return len(m.users) > 0 }

// ptr returns a pointer to v.
func ptr[T any](v T) *T { return &v }

func TestResolveByIndex(t *testing.T) {
	store := &mockStore{users: []domain.User{
		{Name: "Alice", Email: "alice@example.com"},
		{Name: "Bob", Email: "bob@example.com"},
	}}

	filter := domain.UserFilter{Index: ptr(1)}
	user, idx, err := sel.ResolveUser(store, filter)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if idx != 1 {
		t.Errorf("expected index 1, got %d", idx)
	}
	if user.Name != "Bob" {
		t.Errorf("expected Bob, got %s", user.Name)
	}
}

func TestResolveByEmail(t *testing.T) {
	store := &mockStore{users: []domain.User{
		{Name: "Alice", Email: "alice@example.com"},
		{Name: "Bob", Email: "bob@example.com"},
	}}

	filter := domain.UserFilter{Email: ptr("bob@example.com")}
	user, idx, err := sel.ResolveUser(store, filter)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if idx != 1 {
		t.Errorf("expected index 1, got %d", idx)
	}
	if user.Name != "Bob" {
		t.Errorf("expected Bob, got %s", user.Name)
	}
}

func TestResolveByName(t *testing.T) {
	store := &mockStore{users: []domain.User{
		{Name: "Alice", Email: "alice@example.com"},
		{Name: "Bob", Email: "bob@example.com"},
	}}

	filter := domain.UserFilter{Name: ptr("Bob")}
	user, idx, err := sel.ResolveUser(store, filter)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if idx != 1 {
		t.Errorf("expected index 1, got %d", idx)
	}
	if user.Name != "Bob" {
		t.Errorf("expected Bob, got %s", user.Name)
	}
}

func TestResolvePriority(t *testing.T) {
	store := &mockStore{users: []domain.User{
		{Name: "Alice", Email: "alice@example.com"},
		{Name: "Bob", Email: "bob@example.com"},
	}}

	// Index should take precedence over email.
	filter := domain.UserFilter{
		Index: ptr(0),
		Email: ptr("bob@example.com"),
	}
	user, idx, err := sel.ResolveUser(store, filter)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if idx != 0 {
		t.Errorf("expected index 0 (index priority), got %d", idx)
	}
	if user.Name != "Alice" {
		t.Errorf("expected Alice, got %s", user.Name)
	}
}

func TestResolveNotFound(t *testing.T) {
	store := &mockStore{users: []domain.User{
		{Name: "Alice", Email: "alice@example.com"},
	}}

	filter := domain.UserFilter{Email: ptr("notfound@example.com")}
	_, _, err := sel.ResolveUser(store, filter)

	if err == nil {
		t.Fatal("expected error for not-found user, got nil")
	}
}

func TestResolveEmptyList(t *testing.T) {
	store := &mockStore{users: []domain.User{}}

	filter := domain.UserFilter{Index: ptr(0)}
	_, _, err := sel.ResolveUser(store, filter)

	if err == nil {
		t.Fatal("expected error for empty user list, got nil")
	}
}

func TestResolveIndexOutOfRange(t *testing.T) {
	store := &mockStore{users: []domain.User{
		{Name: "Alice", Email: "alice@example.com"},
	}}

	filter := domain.UserFilter{Index: ptr(5)}
	_, _, err := sel.ResolveUser(store, filter)

	if err == nil {
		t.Fatal("expected error for out-of-range index, got nil")
	}
}

func TestResolveInteractive(t *testing.T) {
	store := &mockStore{users: []domain.User{
		{Name: "Alice", Email: "alice@example.com"},
		{Name: "Bob", Email: "bob@example.com"},
	}}

	// Override SelectFunc for testing.
	origSelectFunc := sel.SelectFunc
	defer func() { sel.SelectFunc = origSelectFunc }()

	sel.SelectFunc = func(users []domain.User) (int, error) {
		if len(users) != 2 {
			t.Errorf("expected 2 users passed to select func, got %d", len(users))
		}
		return 1, nil // select Bob
	}

	filter := domain.UserFilter{} // all nil → interactive
	user, idx, err := sel.ResolveUser(store, filter)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if idx != 1 {
		t.Errorf("expected index 1, got %d", idx)
	}
	if user.Name != "Bob" {
		t.Errorf("expected Bob, got %s", user.Name)
	}
}

func TestResolveInteractiveCancelled(t *testing.T) {
	store := &mockStore{users: []domain.User{
		{Name: "Alice", Email: "alice@example.com"},
	}}

	origSelectFunc := sel.SelectFunc
	defer func() { sel.SelectFunc = origSelectFunc }()

	sel.SelectFunc = func(users []domain.User) (int, error) {
		return 0, errors.New("cancelled")
	}

	filter := domain.UserFilter{}
	_, _, err := sel.ResolveUser(store, filter)

	if err == nil {
		t.Fatal("expected error for cancelled interactive selection, got nil")
	}
}
