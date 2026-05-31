package hook

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/northwang-lucky/gitusr/internal/domain"
	"github.com/northwang-lucky/gitusr/internal/gitcmd"
)

// ParseRC reads and parses .gitusrrc from the given directory
// Returns nil, nil if file doesn't exist
// Returns error if JSON is invalid or both name and email are empty
func ParseRC(dir string) (*GitusrRC, error) {
	rcPath := filepath.Join(dir, ".gitusrrc")
	data, err := os.ReadFile(rcPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}

	var rc GitusrRC
	if err := json.Unmarshal(data, &rc); err != nil {
		return nil, fmt.Errorf("invalid .gitusrrc JSON: %w", err)
	}

	if rc.Name == "" && rc.Email == "" {
		return nil, errors.New(".gitusrrc must have at least name or email")
	}

	return &rc, nil
}

// MatchAndApplyRC searches for a matching user in the store and applies git config
// Priority: email > name
// Returns error if no matching user found
func MatchAndApplyRC(store domain.UserStore, rc *GitusrRC) error {
	users, err := store.List()
	if err != nil {
		return err
	}

	var matched *domain.User

	// Try email match first (priority)
	if rc.Email != "" {
		for _, u := range users {
			if u.Email == rc.Email {
				matched = &u
				break
			}
		}
	}

	// If no email match, try name match
	if matched == nil && rc.Name != "" {
		for _, u := range users {
			if u.Name == rc.Name {
				matched = &u
				break
			}
		}
	}

	if matched == nil {
		return errors.New("user specified in .gitusrrc not found in saved users")
	}

	// Apply git config (repo-level, not global)
	if err := gitcmd.SetConfig("name", matched.Name, false); err != nil {
		return err
	}
	if err := gitcmd.SetConfig("email", matched.Email, false); err != nil {
		return err
	}

	return nil
}
