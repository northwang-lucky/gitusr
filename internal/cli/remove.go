package cli

import (
	"fmt"

	"gitusr/internal/domain"
	"gitusr/internal/format"
	sel "gitusr/internal/select"

	"github.com/spf13/cobra"
)

// NewRemoveCmd creates the "remove" (alias "rm") command that deletes a saved user.
func NewRemoveCmd(store domain.UserStore) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "remove",
		Aliases: []string{"rm"},
		Short:   "delete a user",
		RunE: func(cmd *cobra.Command, args []string) error {
			if !store.IsInitialized() {
				return fmt.Errorf("store not initialized")
			}

			// Build UserFilter from flags
			filter := domain.UserFilter{}

			name, _ := cmd.Flags().GetString("name")
			if name != "" {
				filter.Name = &name
			}

			email, _ := cmd.Flags().GetString("email")
			if email != "" {
				filter.Email = &email
			}

			index, _ := cmd.Flags().GetInt("index")
			if index >= 0 {
				filter.Index = &index
			}

			user, idx, err := sel.ResolveUser(store, filter)
			if err != nil {
				format.PrintErr(err.Error())
				return err
			}

			if err := store.Remove(idx); err != nil {
				format.PrintErr(err.Error())
				return err
			}

			fmt.Printf("Success! User (name: %s | email: %s) has been removed!\n", user.Name, user.Email)
			return nil
		},
	}

	cmd.Flags().StringP("name", "n", "", "delete user by name")
	cmd.Flags().StringP("email", "e", "", "delete user by email")
	cmd.Flags().IntP("index", "i", -1, "delete user by index")

	return cmd
}
