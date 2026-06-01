package cli

import (
	"errors"
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/northwang-lucky/gitusr/internal/domain"
	"github.com/northwang-lucky/gitusr/internal/i18n"
)

// NewHooksUninstallCmd creates the "hooks uninstall" command.
// It removes all shell hook wrappers and clears hook configuration.
func NewHooksUninstallCmd(store domain.UserStore) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "uninstall",
		Short: i18n.T("cli.hooks.uninstall.short", nil),
		RunE: func(cmd *cobra.Command, args []string) error {
			err := uninstallFunc(nil, defaultShells())
			if err != nil {
				if strings.Contains(err.Error(), "not installed") || strings.Contains(err.Error(), "no hooks") {
					return errors.New(i18n.T("cli.hooks.uninstall.none_installed", nil))
				}
				return err
			}

			fmt.Println(i18n.T("cli.hooks.uninstall.success", nil))
			return nil
		},
	}

	return cmd
}
