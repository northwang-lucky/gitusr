package cli

import (
	"errors"
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
				return errors.New(i18n.T("cli.list.no_users", nil))
			}

			users, err := store.List()
			if err != nil {
				return err
			}

			output := format.FormatUserList(users)
			fmt.Print(output)
			return nil
		},
	}
	return cmd
}
