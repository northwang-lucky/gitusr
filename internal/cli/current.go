package cli

import (
	"errors"

	"gitusr/internal/domain"
	"gitusr/internal/format"
	"gitusr/internal/gitcmd"

	"github.com/spf13/cobra"
)

// NewCurrentCmd creates the "current" command which displays the current
// repository-level or global git user configuration.
func NewCurrentCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "current",
		Aliases: []string{"ct"},
		Short:   "show current repo/global user",
		RunE: func(cmd *cobra.Command, args []string) error {
			global, err := cmd.Flags().GetBool("global")
			if err != nil {
				return err
			}

			if !global && !gitcmd.IsGitRepo() {
				return errors.New("not a git repository (or any of the parent directories): .git")
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

	cmd.Flags().BoolP("global", "g", false, "show global user")

	return cmd
}
