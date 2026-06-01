package cli

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/northwang-lucky/gitusr/internal/hook"
)

// NewHooksDisableCmd creates the "hooks disable" command.
// It disables a hook type so it won't run until re-enabled.
func NewHooksDisableCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "disable <clone|commit|cd>",
		Short: "Disable a hook type",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			typeStr := args[0]
			hookType := hook.HookType(typeStr)
			if !isValidHookType(hookType) {
				return fmt.Errorf("invalid hook type: %q, must be %q, %q, or %q",
					typeStr, hook.HookTypeClone, hook.HookTypeCommit, hook.HookTypeCD)
			}

			if err := hook.DisableHook(hookType); err != nil {
				return err
			}

			fmt.Printf("Hook %s disabled\n", hookType)
			return nil
		},
	}

	return cmd
}
