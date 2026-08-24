package store

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/northwang-lucky/gitusr/internal/domain"
)

// JSONHostRuleStore implements domain.HostRuleStore using a JSON file
// (hosts.json) next to the user list. The array order is the matching order.
type JSONHostRuleStore struct {
	filePath string
}

// NewJSONHostRuleStore creates a JSONHostRuleStore for the given file path.
func NewJSONHostRuleStore(filePath string) *JSONHostRuleStore {
	return &JSONHostRuleStore{filePath: filePath}
}

// List returns all host rules in configuration order.
// A missing file is treated as an empty rule list.
func (s *JSONHostRuleStore) List() ([]domain.HostRule, error) {
	if _, err := os.Stat(s.filePath); os.IsNotExist(err) {
		return []domain.HostRule{}, nil
	}

	data, err := os.ReadFile(s.filePath)
	if err != nil {
		return nil, fmt.Errorf("read hosts file: %w", err)
	}

	var rules []domain.HostRule
	if err := json.Unmarshal(data, &rules); err != nil {
		return nil, fmt.Errorf("unmarshal hosts: %w", err)
	}

	return rules, nil
}

// SaveAll overwrites the entire rule list, preserving order.
func (s *JSONHostRuleStore) SaveAll(rules []domain.HostRule) error {
	data, err := json.MarshalIndent(rules, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal hosts: %w", err)
	}

	// Append newline for POSIX compatibility
	data = append(data, '\n')

	if err := os.WriteFile(s.filePath, data, 0o644); err != nil {
		return fmt.Errorf("write hosts file: %w", err)
	}

	return nil
}
