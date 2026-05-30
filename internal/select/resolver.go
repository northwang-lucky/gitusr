package sel

import (
	"errors"
	"fmt"

	"gitusr/internal/domain"
	"gitusr/internal/prompt"
)

// SelectFunc is the function used for interactive user selection.
// It defaults to prompt.SelectUser but can be overridden in tests.
var SelectFunc func([]domain.User) (int, error) = prompt.SelectUser

// ResolveUser finds a user from the store based on the filter criteria.
// Priority: Index > Email > Name > Interactive (via SelectFunc).
// Returns the user, its index in the store list, and any error encountered.
func ResolveUser(store domain.UserStore, filter domain.UserFilter) (domain.User, int, error) {
	users, err := store.List()
	if err != nil {
		return domain.User{}, 0, fmt.Errorf("list users: %w", err)
	}

	if len(users) == 0 {
		return domain.User{}, 0, errors.New("no saved users")
	}

	// 1. Index — highest priority
	if filter.Index != nil {
		idx := *filter.Index
		if idx < 0 || idx >= len(users) {
			return domain.User{}, 0, fmt.Errorf("index %d out of range (0-%d)", idx, len(users)-1)
		}
		return users[idx], idx, nil
	}

	// 2. Email — medium priority
	if filter.Email != nil {
		for i, u := range users {
			if u.Email == *filter.Email {
				return u, i, nil
			}
		}
		return domain.User{}, 0, fmt.Errorf("user with email %q not found", *filter.Email)
	}

	// 3. Name — lowest non-interactive priority
	if filter.Name != nil {
		for i, u := range users {
			if u.Name == *filter.Name {
				return u, i, nil
			}
		}
		return domain.User{}, 0, fmt.Errorf("user with name %q not found", *filter.Name)
	}

	// 4. Interactive selection — all filter fields are nil
	idx, err := SelectFunc(users)
	if err != nil {
		return domain.User{}, 0, err
	}

	return users[idx], idx, nil
}
