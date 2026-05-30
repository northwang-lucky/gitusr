package cli

import (
	"errors"

	"github.com/spf13/cobra"

	"gitusr/internal/domain"
	"gitusr/internal/format"
	"gitusr/internal/gitcmd"
	sel "gitusr/internal/select"
)

// NewUseCmd creates the "use" command which switches the active git user
// for the current repository or globally.
func NewUseCmd(store domain.UserStore) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "use",
		Short: "switch user in a git repo or globally",
		RunE: func(cmd *cobra.Command, args []string) error {
			global, err := cmd.Flags().GetBool("global")
			if err != nil {
				return err
			}

			if !global && !gitcmd.IsGitRepo() {
				return errors.New("not a git repository (or any of the parent directories): .git")
			}

			if !store.IsInitialized() {
				return errors.New("no users saved yet, run 'gitusr add' first")
			}

			filter, err := buildFilter(cmd)
			if err != nil {
				return err
			}

			user, _, err := sel.ResolveUser(store, filter)
			if err != nil {
				return err
			}

			if err := gitcmd.SetConfig("name", user.Name, global); err != nil {
				return err
			}

			if err := gitcmd.SetConfig("email", user.Email, global); err != nil {
				return err
			}

			format.PrintUserInfo(user, format.PrintOptions{
				Global:      global,
				ShowSuccess: true,
			})

			return nil
		},
	}

	cmd.Flags().BoolP("global", "g", false, "switch global user")
	cmd.Flags().StringP("name", "n", "", "switch by name")
	cmd.Flags().StringP("email", "e", "", "switch by email")
	cmd.Flags().IntP("index", "i", -1, "switch by index")

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
