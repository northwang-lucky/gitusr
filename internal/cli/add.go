package cli

import (
	"fmt"

	"github.com/spf13/cobra"

	"gitusr/internal/domain"
	"gitusr/internal/i18n"
	"gitusr/internal/prompt"
)

// askNewUser is the function used to prompt for a new user.
// It is a package-level variable so that tests can replace it with a mock.
var askNewUser = prompt.AskNewUser

// NewAddCmd creates the "add" command for interactively adding and saving a user.
func NewAddCmd(store domain.UserStore) *cobra.Command {
	return &cobra.Command{
		Use:   "add",
		Short: i18n.T("cli.add.short", nil),
		RunE: func(cmd *cobra.Command, args []string) error {
			if !store.IsInitialized() {
				return fmt.Errorf("%s", i18n.T("cli.error.store_not_init", nil))
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
					return fmt.Errorf("%s", i18n.T("cli.add.dup_name", map[string]interface{}{"Name": user.Name}))
				}
				if u.Email == user.Email {
					return fmt.Errorf("%s", i18n.T("cli.add.dup_email", map[string]interface{}{"Email": user.Email}))
				}
			}

			if err := store.Add(user); err != nil {
				return fmt.Errorf("add user: %w", err)
			}

			fmt.Print(i18n.T("cli.add.success", map[string]interface{}{"Name": user.Name, "Email": user.Email}) + "\n")
			return nil
		},
	}
}
