package cli

import (
	"github.com/spf13/cobra"

	"gitusr/internal/domain"
)

// NewRootCmd creates the root cobra command for gitusr and registers all
// subcommands. The store parameter is injected into every subcommand that
// requires persistence.
func NewRootCmd(store domain.UserStore) *cobra.Command {
	cmd := &cobra.Command{
		Use:           "gitusr",
		Short:         "A CLI that allows you to switch git users.",
		SilenceErrors: true,
		SilenceUsage:  true,
	}

	cmd.AddCommand(
		NewAddCmd(store),
		NewCurrentCmd(),
		NewInitCmd(store),
		NewListCmd(store),
		NewRemoveCmd(store),
		NewReplaceCmd(store),
		NewUseCmd(store),
	)

	return cmd
}
