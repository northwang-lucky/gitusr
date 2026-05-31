package cli

import (
	"fmt"
	"regexp"

	"github.com/spf13/cobra"

	"gitusr/internal/domain"
	"gitusr/internal/i18n"
	"gitusr/internal/prompt"
)

// emailRegex is the regex used to validate email addresses from flags.
// It mirrors the pattern used in internal/prompt/prompt.go.
const emailRegex = `\w+([-+.]\w+)*@\w+([-.]\w+)*\.\w+([-.]\w+)*`

var compiledEmailRegex = regexp.MustCompile("^" + emailRegex + "$")

// askNewUser is the function used to prompt for a new user.
// It is a package-level variable so that tests can replace it with a mock.
var askNewUser = prompt.AskNewUser

// NewAddCmd creates the "add" command for interactively adding and saving a user.
func NewAddCmd(store domain.UserStore) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "add",
		Short: i18n.T("cli.add.short", nil),
		RunE: func(cmd *cobra.Command, args []string) error {
			if !store.IsInitialized() {
				return fmt.Errorf("%s", i18n.T("cli.error.store_not_init", nil))
			}

			nameFlag := cmd.Flags().Changed("name")
			emailFlag := cmd.Flags().Changed("email")

			var user domain.User
			var err error

			if nameFlag || emailFlag {
				if !nameFlag || !emailFlag {
					return fmt.Errorf("%s", i18n.T("cli.add.required_flags", nil))
				}

				name, _ := cmd.Flags().GetString("name")
				email, _ := cmd.Flags().GetString("email")

				if !compiledEmailRegex.MatchString(email) {
					return fmt.Errorf("%s", i18n.T("prompt.invalid_email", nil))
				}

				user = domain.User{Name: name, Email: email}
			} else {
				user, err = askNewUser()
				if err != nil {
					return fmt.Errorf("prompt: %w", err)
				}
			}

			users, err := store.List()
			if err != nil {
				return fmt.Errorf("list users: %w", err)
			}

			for _, u := range users {
				if u.Name == user.Name {
					return fmt.Errorf("%s", i18n.T("cli.add.dup_name", map[string]interface{}{"Name": user.Name}))
				}
				if u.Email == user.Email {
					return fmt.Errorf("%s", i18n.T("cli.add.dup_email", map[string]interface{}{"Email": user.Email}))
				}
			}

			if err := store.Add(user); err != nil {
				return fmt.Errorf("add user: %w", err)
			}

			fmt.Print(i18n.T("cli.add.success", map[string]interface{}{"Name": user.Name, "Email": user.Email}) + "\n")
			return nil
		},
	}

	cmd.Flags().StringP("name", "n", "", i18n.T("cli.add.flag_name", nil))
	cmd.Flags().StringP("email", "e", "", i18n.T("cli.add.flag_email", nil))

	return cmd
}
