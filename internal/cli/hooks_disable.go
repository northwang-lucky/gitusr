package cli

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/northwang-lucky/gitusr/internal/hook"
	"github.com/northwang-lucky/gitusr/internal/i18n"
)

// NewHooksDisableCmd creates the "hooks disable" command.
// It disables a hook type so it won't run until re-enabled.
func NewHooksDisableCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "disable <clone|commit|cd>",
		Short: i18n.T("cli.hooks.disable.short", nil),
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			typeStr := args[0]
			hookType := hook.HookType(typeStr)
			if !isValidHookType(hookType) {
				return fmt.Errorf(i18n.T("cli.hooks.invalid_type", nil),
					typeStr, hook.HookTypeClone, hook.HookTypeCommit, hook.HookTypeCD)
			}

			if err := hook.DisableHook(hookType); err != nil {
				return err
			}

			fmt.Printf(i18n.T("cli.hooks.disable.success", nil), hookType)
			return nil
		},
	}

	return cmd
}
