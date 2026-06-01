package cli

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/northwang-lucky/gitusr/internal/hook"
	"github.com/northwang-lucky/gitusr/internal/i18n"
)

// NewHooksEnableCmd creates the "hooks enable" command.
// It re-enables a previously disabled hook type by positional argument
// without reinstalling it.
func NewHooksEnableCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "enable <clone|commit|cd>",
		Short: i18n.T("cli.hooks.enable.short", nil),
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			typeStr := args[0]
			hookType := hook.HookType(typeStr)
			if !isValidHookType(hookType) {
				return fmt.Errorf(i18n.T("cli.hooks.invalid_type", nil),
					typeStr, hook.HookTypeClone, hook.HookTypeCommit, hook.HookTypeCD)
			}

			if err := hook.EnableHook(hookType); err != nil {
				return err
			}

			fmt.Printf(i18n.T("cli.hooks.enable.success", nil), hookType)
			return nil
		},
	}

	return cmd
}
