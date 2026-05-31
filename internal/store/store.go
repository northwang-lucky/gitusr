package store

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/northwang-lucky/gitusr/internal/domain"
)

// JSONStore implements domain.UserStore using a JSON file as persistence.
type JSONStore struct {
	filePath string
}

// NewJSONStore creates a new JSONStore for the given file path.
func NewJSONStore(filePath string) *JSONStore {
	return &JSONStore{filePath: filePath}
}

// List returns all users from the JSON file.
// If the file does not exist, it creates an empty array and returns an empty list.
func (s *JSONStore) List() ([]domain.User, error) {
	if _, err := os.Stat(s.filePath); os.IsNotExist(err) {
		if err := s.writeUsers([]domain.User{}); err != nil {
			return nil, fmt.Errorf("create store file: %w", err)
		}
		return []domain.User{}, nil
	}

	return s.readUsers()
}

// Add appends a user to the store.
// Returns an error if a user with the same name or email already exists.
func (s *JSONStore) Add(user domain.User) error {
	users, err := s.List()
	if err != nil {
		return fmt.Errorf("list users: %w", err)
	}

	for _, u := range users {
		if u.Name == user.Name {
			return fmt.Errorf("user with name %q already exists", user.Name)
		}
		if u.Email == user.Email {
			return fmt.Errorf("user with email %q already exists", user.Email)
		}
	}

	users = append(users, user)
	return s.writeUsers(users)
}

// Remove deletes a user at the given index.
// Returns an error if the index is out of bounds.
func (s *JSONStore) Remove(index int) error {
	users, err := s.List()
	if err != nil {
		return fmt.Errorf("list users: %w", err)
	}

	if index < 0 || index >= len(users) {
		return fmt.Errorf("index %d out of bounds (0-%d)", index, len(users)-1)
	}

	users = append(users[:index], users[index+1:]...)
	return s.writeUsers(users)
}

// SaveAll overwrites the entire user list with the given slice.
func (s *JSONStore) SaveAll(users []domain.User) error {
	return s.writeUsers(users)
}

// IsInitialized returns true if the store contains at least one user.
func (s *JSONStore) IsInitialized() bool {
	users, err := s.List()
	if err != nil {
		return false
	}
	return len(users) > 0
}

// readUsers reads and unmarshals users from the JSON file.
func (s *JSONStore) readUsers() ([]domain.User, error) {
	data, err := os.ReadFile(s.filePath)
	if err != nil {
		return nil, fmt.Errorf("read store file: %w", err)
	}

	var users []domain.User
	if err := json.Unmarshal(data, &users); err != nil {
		return nil, fmt.Errorf("unmarshal users: %w", err)
	}

	return users, nil
}

// writeUsers marshals and writes users to the JSON file.
func (s *JSONStore) writeUsers(users []domain.User) error {
	data, err := json.MarshalIndent(users, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal users: %w", err)
	}

	// Append newline for POSIX compatibility
	data = append(data, '\n')

	if err := os.WriteFile(s.filePath, data, 0o644); err != nil {
		return fmt.Errorf("write store file: %w", err)
	}

	return nil
}
