package cli

import (
	"errors"

	"github.com/northwang-lucky/gitusr/internal/domain"
	"github.com/northwang-lucky/gitusr/internal/format"
	"github.com/northwang-lucky/gitusr/internal/gitcmd"
	"github.com/northwang-lucky/gitusr/internal/i18n"

	"github.com/spf13/cobra"
)

// NewCurrentCmd creates the "current" command which displays the current
// repository-level or global git user configuration.
func NewCurrentCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "current",
		Aliases: []string{"ct"},
		Short:   i18n.T("cli.current.short", nil),
		RunE: func(cmd *cobra.Command, args []string) error {
			global, err := cmd.Flags().GetBool("global")
			if err != nil {
				return err
			}

			if !global && !gitcmd.IsGitRepo() {
				return errors.New(i18n.T("cli.error.not_repo", nil))
			}

			name, err := gitcmd.GetConfig("name", global)
			if err != nil {
				return err
			}

			email, err := gitcmd.GetConfig("email", global)
			if err != nil {
				return err
			}

			user := domain.User{Name: name, Email: email}
			opts := format.PrintOptions{Global: global, ShowSuccess: false}
			format.PrintUserInfo(user, opts)

			return nil
		},
	}

	cmd.Flags().BoolP("global", "g", false, i18n.T("cli.current.flag_global", nil))

	return cmd
}
