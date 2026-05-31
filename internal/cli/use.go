package cli

import (
	"errors"

	"github.com/spf13/cobra"

	"github.com/northwang-lucky/gitusr/internal/domain"
	"github.com/northwang-lucky/gitusr/internal/format"
	"github.com/northwang-lucky/gitusr/internal/gitcmd"
	"github.com/northwang-lucky/gitusr/internal/i18n"
	sel "github.com/northwang-lucky/gitusr/internal/select"
)

// NewUseCmd creates the "use" command which switches the active git user
// for the current repository or globally.
func NewUseCmd(store domain.UserStore) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "use",
		Short: i18n.T("cli.use.short", nil),
		RunE: func(cmd *cobra.Command, args []string) error {
			global, err := cmd.Flags().GetBool("global")
			if err != nil {
				return err
			}

			if !global && !gitcmd.IsGitRepo() {
				return errors.New(i18n.T("cli.error.not_repo", nil))
			}

			if !store.IsInitialized() {
				return errors.New(i18n.T("cli.use.no_users", nil))
			}

			filter, err := buildFilter(cmd)
			if err != nil {
				return err
			}

			user, _, err := sel.ResolveUser(store, filter)
			if err != nil {
				return err
			}

			silent, _ := cmd.Flags().GetBool("silent-if-unchanged")
			if silent {
				currentName, nameErr := getConfigFn("name", global)
				currentEmail, emailErr := getConfigFn("email", global)
				if nameErr == nil && emailErr == nil && currentName == user.Name && currentEmail == user.Email {
					return nil
				}
			}

			if err := setConfigFn("name", user.Name, global); err != nil {
				return err
			}

			if err := setConfigFn("email", user.Email, global); err != nil {
				return err
			}

			format.PrintUserInfo(user, format.PrintOptions{
				Global:      global,
				ShowSuccess: true,
			})

			return nil
		},
	}

	cmd.Flags().BoolP("global", "g", false, i18n.T("cli.use.flag_global", nil))
	cmd.Flags().StringP("name", "n", "", i18n.T("cli.use.flag_name", nil))
	cmd.Flags().StringP("email", "e", "", i18n.T("cli.use.flag_email", nil))
	cmd.Flags().IntP("index", "i", -1, i18n.T("cli.use.flag_index", nil))
	cmd.Flags().BoolP("silent-if-unchanged", "s", false, i18n.T("cli.use.flag_silent", nil))

	return cmd
}

// buildFilter constructs a UserFilter from the command's flag values.
func buildFilter(cmd *cobra.Command) (domain.UserFilter, error) {
	filter := domain.UserFilter{}

	name, err := cmd.Flags().GetString("name")
	if err != nil {
		return filter, err
	}
	if name != "" {
		filter.Name = &name
	}

	email, err := cmd.Flags().GetString("email")
	if err != nil {
		return filter, err
	}
	if email != "" {
		filter.Email = &email
	}

	if cmd.Flags().Changed("index") {
		index, err := cmd.Flags().GetInt("index")
		if err != nil {
			return filter, err
		}
		filter.Index = &index
	}

	return filter, nil
}
