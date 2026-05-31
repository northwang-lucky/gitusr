package cli

import (
	"errors"
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/northwang-lucky/gitusr/internal/domain"
	"github.com/northwang-lucky/gitusr/internal/hook"
	"github.com/northwang-lucky/gitusr/internal/i18n"
)

// installFunc and uninstallFunc are package-level variables to allow
// test code to inject mock implementations.
var installFunc = hook.Install
var uninstallFunc = hook.Uninstall

// defaultShells returns the list of shell types that hook commands operate on by default.
func defaultShells() []hook.ShellType {
	return []hook.ShellType{hook.ShellTypeBash, hook.ShellTypeZsh}
}

// NewHookCmd creates the parent "hook" command that groups install and uninstall subcommands.
func NewHookCmd(store domain.UserStore) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "hook",
		Short: i18n.T("cli.hook.short", nil),
	}

	cmd.AddCommand(
		NewHookInstallCmd(store),
		NewHookUninstallCmd(store),
		NewHookApplyRCCmd(store),
	)

	return cmd
}

// NewHookInstallCmd creates the "hook install" command.
// It installs shell hook wrappers that enable automatic git user detection
// on clone, commit, and cd operations.
func NewHookInstallCmd(store domain.UserStore) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "install",
		Short: i18n.T("cli.hook.install.short", nil),
		RunE: func(cmd *cobra.Command, args []string) error {
			typeStr, err := cmd.Flags().GetString("type")
			if err != nil {
				return err
			}

			if typeStr == "" {
				return errors.New("required flag \"type\" not set")
			}

			hookType := hook.HookType(typeStr)
			if hookType != hook.HookTypeClone && hookType != hook.HookTypeCommit && hookType != hook.HookTypeCD {
				return fmt.Errorf("invalid hook type: %q, must be %q, %q, or %q",
					typeStr, hook.HookTypeClone, hook.HookTypeCommit, hook.HookTypeCD)
			}

			results, err := installFunc(hookType, defaultShells())
			if err != nil {
				return err
			}

			// nil results means the hook is already installed (idempotent).
			if results == nil {
				fmt.Println(i18n.T("cli.hook.install.already_installed",
					map[string]interface{}{"Type": string(hookType)}))
				return nil
			}

			fmt.Println(i18n.T("cli.hook.install.success",
				map[string]interface{}{"Type": string(hookType)}))
			return nil
		},
	}

	cmd.Flags().String("type", "", i18n.T("cli.hook.install.flag_type", nil))
	cmd.MarkFlagRequired("type")

	return cmd
}

// NewHookUninstallCmd creates the "hook uninstall" command.
// It removes shell hook wrappers and cleans up hook configuration.
func NewHookUninstallCmd(store domain.UserStore) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "uninstall",
		Short: i18n.T("cli.hook.uninstall.short", nil),
		RunE: func(cmd *cobra.Command, args []string) error {
			typeStr, err := cmd.Flags().GetString("type")
			if err != nil {
				return err
			}

			if typeStr == "" {
				return errors.New("required flag \"type\" not set")
			}

			hookType := hook.HookType(typeStr)
			if hookType != hook.HookTypeClone && hookType != hook.HookTypeCommit && hookType != hook.HookTypeCD {
				return fmt.Errorf("invalid hook type: %q, must be %q, %q, or %q",
					typeStr, hook.HookTypeClone, hook.HookTypeCommit, hook.HookTypeCD)
			}

			err = uninstallFunc(hookType, defaultShells())
			if err != nil {
				// Translate the "not installed" error for consistent user-facing output.
				if strings.Contains(err.Error(), "not installed") {
					return errors.New(i18n.T("cli.hook.uninstall.not_installed",
						map[string]interface{}{"Type": string(hookType)}))
				}
				return err
			}

			fmt.Println(i18n.T("cli.hook.uninstall.success",
				map[string]interface{}{"Type": string(hookType)}))
			return nil
		},
	}

	cmd.Flags().String("type", "", i18n.T("cli.hook.install.flag_type", nil))
	cmd.MarkFlagRequired("type")

	return cmd
}
