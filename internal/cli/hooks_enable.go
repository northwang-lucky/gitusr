package cli

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/northwang-lucky/gitusr/internal/hook"
)

// NewHooksEnableCmd creates the "hooks enable" command.
// It re-enables a previously disabled hook type by positional argument
// without reinstalling it.
func NewHooksEnableCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "enable <clone|commit|cd>",
		Short: "Enable a hook type",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			typeStr := args[0]
			hookType := hook.HookType(typeStr)
			if !isValidHookType(hookType) {
				return fmt.Errorf("invalid hook type: %q, must be %q, %q, or %q",
					typeStr, hook.HookTypeClone, hook.HookTypeCommit, hook.HookTypeCD)
			}

			if err := hook.EnableHook(hookType); err != nil {
				return err
			}

			fmt.Printf("Hook %s enabled\n", hookType)
			return nil
		},
	}

	return cmd
}
