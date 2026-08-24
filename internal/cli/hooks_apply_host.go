package cli

import (
	"github.com/spf13/cobra"

	"github.com/northwang-lucky/gitusr/internal/domain"
	"github.com/northwang-lucky/gitusr/internal/format"
	"github.com/northwang-lucky/gitusr/internal/hosts"
	"github.com/northwang-lucky/gitusr/internal/i18n"
)

// NewHooksApplyHostCmd creates the hidden "hooks apply-host" command.
// The clone hook wrapper calls it after a successful clone: it matches the
// repository URL's host against the configured host rules and applies the
// matched saved user through the same set-and-print path as "gitusr use".
//
// Behaviour per design decision:
//   - No matching rule → nil, no output (silent skip).
//   - Rule matches but the email no longer exists in the user list → one
//     warning line on stderr, nil exit (skipped, non-fatal).
//   - Rule matches and the user exists → repo-level git config is set and
//     the user info is printed, unless --silent-if-unchanged detects an
//     already-matching config.
func NewHooksApplyHostCmd(store domain.UserStore, hostStore domain.HostRuleStore) *cobra.Command {
	cmd := &cobra.Command{
		Use:    "apply-host <url>",
		Hidden: true,
		Args:   cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			rules, err := hostStore.List()
			if err != nil {
				return err
			}

			rule, ok := hosts.MatchHost(rules, args[0])
			if !ok {
				return nil
			}

			user, found, err := findSavedUser(store, rule.Email)
			if err != nil {
				return err
			}
			if !found {
				format.PrintWarn(i18n.T("cli.hooks.apply_host.missing_user", map[string]interface{}{
					"Host": rule.Host, "Email": rule.Email,
				}))
				return nil
			}

			silent, _ := cmd.Flags().GetBool("silent-if-unchanged")
			if silent {
				currentName, nameErr := getConfigFn("name", false)
				currentEmail, emailErr := getConfigFn("email", false)
				if nameErr == nil && emailErr == nil &&
					currentName == user.Name && currentEmail == user.Email {
					return nil
				}
			}

			if err := setConfigFn("name", user.Name, false); err != nil {
				return err
			}
			if err := setConfigFn("email", user.Email, false); err != nil {
				return err
			}

			format.PrintUserInfo(user, format.PrintOptions{ShowSuccess: true})
			return nil
		},
	}

	cmd.Flags().BoolP("silent-if-unchanged", "s", false,
		i18n.T("cli.use.flag_silent", nil))

	return cmd
}
