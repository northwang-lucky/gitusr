package cli

import (
	"errors"
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
		NewHookUninstallCmd(store),
		NewHookEnableCmd(),
		NewHookDisableCmd(),
		NewHooksApplyRCCmd(store),
		newHookIsDisabledCmd(),
	)

	return cmd
}

// NewHookEnableCmd creates the "hooks enable" command.
// It re-enables a previously disabled hook type without reinstalling it.
func NewHookEnableCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "enable",
		Short: "Enable a hook type",
		RunE: func(cmd *cobra.Command, args []string) error {
			typeStr, err := cmd.Flags().GetString("type")
			if err != nil {
				return err
			}

			if typeStr == "" {
				return errors.New("--type flag is required for enable")
			}

			hookType := hook.HookType(typeStr)
			if !isValidHookType(hookType) {
				return fmt.Errorf("invalid hook type: %q, must be %q, %q, or %q",
					typeStr, hook.HookTypeClone, hook.HookTypeCommit, hook.HookTypeCD)
			}

			if err := hook.EnableHook(hookType); err != nil {
				return err
			}

			fmt.Printf("Hook type %q enabled\n", hookType)
			return nil
		},
	}

	cmd.Flags().String("type", "", "Hook type to enable (clone, commit, cd)")

	return cmd
}

// NewHookDisableCmd creates the "hooks disable" command.
// It disables a hook type without uninstalling it, so it can be re-enabled later.
func NewHookDisableCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "disable",
		Short: "Disable a hook type",
		RunE: func(cmd *cobra.Command, args []string) error {
			typeStr, err := cmd.Flags().GetString("type")
			if err != nil {
				return err
			}

			if typeStr == "" {
				return errors.New("--type flag is required for disable")
			}

			hookType := hook.HookType(typeStr)
			if !isValidHookType(hookType) {
				return fmt.Errorf("invalid hook type: %q, must be %q, %q, or %q",
					typeStr, hook.HookTypeClone, hook.HookTypeCommit, hook.HookTypeCD)
			}

			if err := hook.DisableHook(hookType); err != nil {
				return err
			}

			fmt.Printf("Hook type %q disabled\n", hookType)
			return nil
		},
	}

	cmd.Flags().String("type", "", "Hook type to disable (clone, commit, cd)")

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

			// exit 0: hook IS disabled
			if !enabled {
				return nil
			}

			// exit 1: hook is NOT disabled
			return fmt.Errorf("hook type %q is not disabled", typeStr)
		},
	}

	return cmd
}
