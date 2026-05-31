package cli

import (
	"github.com/spf13/cobra"

	"gitusr/internal/domain"
	"gitusr/internal/i18n"
)

// NewRootCmd creates the root cobra command for gitusr and registers all
// subcommands. The store parameter is injected into every subcommand that
// requires persistence. The name parameter controls the command's Use field,
// allowing "gitusr" or "gu" depending on the invocation alias.
func NewRootCmd(store domain.UserStore, name string) *cobra.Command {
	cmd := &cobra.Command{
		Use:           name,
		Short:         i18n.T("cli.root.short", nil),
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
