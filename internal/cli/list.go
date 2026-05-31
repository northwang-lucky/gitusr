package cli

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/northwang-lucky/gitusr/internal/domain"
	"github.com/northwang-lucky/gitusr/internal/format"
	"github.com/northwang-lucky/gitusr/internal/i18n"
)

// NewListCmd creates the "list" command that shows all saved users.
func NewListCmd(store domain.UserStore) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "list",
		Aliases: []string{"ls"},
		Short:   i18n.T("cli.list.short", nil),
		RunE: func(cmd *cobra.Command, args []string) error {
			if !store.IsInitialized() {
				format.PrintErr(i18n.T("cli.list.no_users", nil))
				return fmt.Errorf("store not initialized")
			}

			users, err := store.List()
			if err != nil {
				format.PrintErr(err.Error())
				return err
			}

			output := format.FormatUserList(users)
			fmt.Print(output)
			return nil
		},
	}
	return cmd
}
