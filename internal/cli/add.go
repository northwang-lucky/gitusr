package cli

import (
	"fmt"

	"github.com/spf13/cobra"

	"gitusr/internal/domain"
	"gitusr/internal/prompt"
)

// askNewUser is the function used to prompt for a new user.
// It is a package-level variable so that tests can replace it with a mock.
var askNewUser = prompt.AskNewUser

// NewAddCmd creates the "add" command for interactively adding and saving a user.
func NewAddCmd(store domain.UserStore) *cobra.Command {
	return &cobra.Command{
		Use:   "add",
		Short: "add and save a user",
		RunE: func(cmd *cobra.Command, args []string) error {
			if !store.IsInitialized() {
				return fmt.Errorf("store not initialized")
			}

			user, err := askNewUser()
			if err != nil {
				return fmt.Errorf("prompt: %w", err)
			}

			users, err := store.List()
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

			if err := store.Add(user); err != nil {
				return fmt.Errorf("add user: %w", err)
			}

			fmt.Printf("Success! User (name: %s | email: %s) has been saved!\n", user.Name, user.Email)
			return nil
		},
	}
}
