package cli

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/northwang-lucky/gitusr/internal/domain"
	"github.com/northwang-lucky/gitusr/internal/hook"
	"github.com/northwang-lucky/gitusr/internal/i18n"
)

// generateEnvFn is a package-level variable that can be overridden in tests.
var generateEnvFn = hook.GenerateEnv

// NewHookEnvCmd creates the "hook env" command which generates shell code
// for cd-based auto-switching. The output is designed to be eval'd:
//
//	eval "$(gitusr hook env --shell bash)"
func NewHookEnvCmd(store domain.UserStore) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "env",
		Short: i18n.T("cli.hook.env.short", nil),
		RunE: func(cmd *cobra.Command, args []string) error {
			shell, err := cmd.Flags().GetString("shell")
			if err != nil {
				return err
			}

			output, err := generateEnvFn(hook.ShellType(shell))
			if err != nil {
				return err
			}

			fmt.Print(output)
			return nil
		},
	}

	cmd.Flags().String("shell", "", i18n.T("cli.hook.env.flag_shell", nil))
	cmd.MarkFlagRequired("shell")

	return cmd
}
