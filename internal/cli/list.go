package cli

import (
	"fmt"

	"github.com/spf13/cobra"

	"gitusr/internal/domain"
	"gitusr/internal/format"
)

// NewListCmd creates the "list" command that shows all saved users.
func NewListCmd(store domain.UserStore) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "list",
		Aliases: []string{"ls"},
		Short:   "show all saved users",
		RunE: func(cmd *cobra.Command, args []string) error {
			if !store.IsInitialized() {
				format.PrintErr("no users saved yet, run 'gitusr add' first")
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
