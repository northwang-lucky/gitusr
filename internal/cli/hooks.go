package cli

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/northwang-lucky/gitusr/internal/domain"
	"github.com/northwang-lucky/gitusr/internal/hook"
)

// NewHooksCmd creates the parent "hooks" command that groups all hook-related
// subcommands including install, uninstall, enable, disable, apply-rc, and is-disabled.
func NewHooksCmd(store domain.UserStore) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "hooks",
		Short: "Manage shell hooks",
	}

	cmd.AddCommand(
		NewHookInstallCmd(store),
		NewHooksUninstallCmd(store),
		NewHooksEnableCmd(),
		NewHooksDisableCmd(),
		NewHookApplyRCCmd(store),
		newHookIsDisabledCmd(),
	)

	return cmd
}

// newHookIsDisabledCmd creates the hidden "hooks is-disabled" command.
// It checks whether a hook type is disabled and returns exit code 0 if disabled
// (useful for scripting). Exit code 1 indicates the hook type is enabled.
func newHookIsDisabledCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:    "is-disabled <hook-type>",
		Hidden: true,
		Args:   cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			typeStr := args[0]
			hookType := hook.HookType(typeStr)
			if !isValidHookType(hookType) {
				return fmt.Errorf("invalid hook type: %q, must be %q, %q, or %q",
					typeStr, hook.HookTypeClone, hook.HookTypeCommit, hook.HookTypeCD)
			}

			enabled, err := hook.IsEnabled(hookType)
			if err != nil {
				return err
			}

			if !enabled {
				return nil
			}

			return fmt.Errorf("hook type %q is not disabled", typeStr)
		},
	}

	return cmd
}
