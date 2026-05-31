package cli

import (
	"errors"

	"github.com/spf13/cobra"

	"github.com/northwang-lucky/gitusr/internal/domain"
	"github.com/northwang-lucky/gitusr/internal/hook"
	"github.com/northwang-lucky/gitusr/internal/i18n"
)

// NewHookApplyRCCmd creates the hidden "hook apply-rc" command.
// It reads .gitusrrc from the current directory, matches a user from the store
// by email (priority) then name, and applies the git config locally.
//
// This command is called by the shell hook wrappers when processing
// git operations (clone, commit). It is hidden from help output.
func NewHookApplyRCCmd(store domain.UserStore) *cobra.Command {
	cmd := &cobra.Command{
		Use:    "apply-rc",
		Hidden: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			rc, err := hook.ParseRC("")
			if err != nil {
				return err
			}

			// No .gitusrrc file — nothing to do
			if rc == nil {
				return nil
			}

			// Find matching user from store (email priority, then name)
			users, err := store.List()
			if err != nil {
				return err
			}

			var matched *domain.User

			if rc.Email != "" {
				for i := range users {
					if users[i].Email == rc.Email {
						matched = &users[i]
						break
					}
				}
			}

			if matched == nil && rc.Name != "" {
				for i := range users {
					if users[i].Name == rc.Name {
						matched = &users[i]
						break
					}
				}
			}

			if matched == nil {
				return errors.New(i18n.T("cli.hook.rc.not_found", nil))
			}

			// --silent-if-unchanged: skip if git config already matches
			silent, _ := cmd.Flags().GetBool("silent-if-unchanged")
			if silent {
				currentName, nameErr := getConfigFn("name", false)
				currentEmail, emailErr := getConfigFn("email", false)
				if nameErr == nil && emailErr == nil &&
					currentName == matched.Name && currentEmail == matched.Email {
					return nil
				}
			}

			return hook.MatchAndApplyRC(store, rc)
		},
	}

	cmd.Flags().BoolP("silent-if-unchanged", "s", false,
		i18n.T("cli.use.flag_silent", nil))

	return cmd
}
