package cli

import (
	"fmt"

	"gitusr/internal/domain"
	"gitusr/internal/format"
	"gitusr/internal/i18n"
	sel "gitusr/internal/select"

	"github.com/spf13/cobra"
)

// NewRemoveCmd creates the "remove" (alias "rm") command that deletes a saved user.
func NewRemoveCmd(store domain.UserStore) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "remove",
		Aliases: []string{"rm"},
		Short:   i18n.T("cli.remove.short", nil),
		RunE: func(cmd *cobra.Command, args []string) error {
			if !store.IsInitialized() {
				return fmt.Errorf("%s", i18n.T("cli.error.store_not_init", nil))
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

			fmt.Println(i18n.T("cli.remove.success", map[string]interface{}{"Name": user.Name, "Email": user.Email}))
			return nil
		},
	}

	cmd.Flags().StringP("name", "n", "", i18n.T("cli.remove.flag_name", nil))
	cmd.Flags().StringP("email", "e", "", i18n.T("cli.remove.flag_email", nil))
	cmd.Flags().IntP("index", "i", -1, i18n.T("cli.remove.flag_index", nil))

	return cmd
}
