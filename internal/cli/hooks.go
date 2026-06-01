package cli

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/northwang-lucky/gitusr/internal/domain"
	"github.com/northwang-lucky/gitusr/internal/hook"
	"github.com/northwang-lucky/gitusr/internal/i18n"
)

// installFunc, installAllFunc and uninstallFunc are package-level variables
// to allow test code to inject mock implementations.
var installFunc = hook.Install
var installAllFunc = hook.InstallAll
var uninstallFunc func(any, []hook.ShellType) error = func(_ any, shells []hook.ShellType) error {
	return hook.UninstallAll(shells)
}

// defaultShells returns the list of shell types that hook commands operate on by default.
func defaultShells() []hook.ShellType {
	return []hook.ShellType{hook.ShellTypeBash, hook.ShellTypeZsh}
}

// isValidHookType returns true if ht is one of the known hook types.
func isValidHookType(ht hook.HookType) bool {
	return ht == hook.HookTypeClone || ht == hook.HookTypeCommit || ht == hook.HookTypeCD
}

// NewHooksCmd creates the parent "hooks" command that groups all hook-related
// subcommands including install, uninstall, enable, disable, apply-rc, and is-disabled.
func NewHooksCmd(store domain.UserStore) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "hooks",
		Short: i18n.T("cli.hooks.short", nil),
	}

	cmd.AddCommand(
		NewHookInstallCmd(store),
		NewHooksUninstallCmd(store),
		NewHooksEnableCmd(),
		NewHooksDisableCmd(),
		NewHooksApplyRCCmd(store),
		newHookIsDisabledCmd(),
	)

	return cmd
}

// newHookIsDisabledCmd creates the hidden "hooks is-disabled" command.
func newHookIsDisabledCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:    "is-disabled <hook-type>",
		Hidden: true,
		Args:   cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			typeStr := args[0]
			hookType := hook.HookType(typeStr)
			if !isValidHookType(hookType) {
				return fmt.Errorf(i18n.T("cli.hooks.invalid_type", nil),
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

// NewHookInstallCmd creates the "hooks install" command.
// It installs hook wrappers for all hook types (clone, commit, cd)
// on the default shells (bash, zsh). The operation is idempotent —
// if all hooks are already installed, it prints a message and exits.
func NewHookInstallCmd(store domain.UserStore) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "install",
		Short: i18n.T("cli.hooks.install.short", nil),
		RunE: func(cmd *cobra.Command, args []string) error {
			results, err := installAllFunc(defaultShells())
			if err != nil {
				return err
			}

			if results == nil {
				fmt.Println(i18n.T("cli.hooks.install.already_installed", nil))
				return nil
			}

			fmt.Println(i18n.T("cli.hooks.install.success", nil))
			return nil
		},
	}

	return cmd
}
