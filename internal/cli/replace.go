package cli

import (
	"errors"
	"fmt"
	"time"

	"github.com/spf13/cobra"

	"github.com/northwang-lucky/gitusr/internal/domain"
	"github.com/northwang-lucky/gitusr/internal/gitcmd"
	"github.com/northwang-lucky/gitusr/internal/i18n"
	"github.com/northwang-lucky/gitusr/internal/prompt"
	sel "github.com/northwang-lucky/gitusr/internal/select"
)

// Package-level variables allow tests to inject mocks for git-filter-repo
// and interactive confirmation without touching the production packages.
var (
	filterRepoFunc = gitcmd.FilterRepo
	confirmFunc    = prompt.Confirm
)

// NewReplaceCmd creates the "replace" command that rewrites git history to
// replace the author of commits matching <target-email> with another saved user.
// A backup branch (backup/pre-replace-{unix_timestamp}) is created before any
// history rewriting takes place.
func NewReplaceCmd(store domain.UserStore) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "replace <target-email>",
		Short: i18n.T("cli.replace.short", nil),
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			targetEmail := args[0]

			// 1. Safety: must be inside a git repository
			if !gitcmd.IsGitRepo() {
				return errors.New(i18n.T("cli.error.not_repo", nil))
			}

			// 2. Safety: no uncommitted changes allowed
			if gitcmd.HasUncommittedChanges() {
				return errors.New(i18n.T("cli.replace.uncommitted", nil))
			}

			// 3. Read filter flags
			nameFilter, err := cmd.Flags().GetString("with-name")
			if err != nil {
				return err
			}
			emailFilter, err := cmd.Flags().GetString("with-email")
			if err != nil {
				return err
			}
			indexFilter, err := cmd.Flags().GetInt("with-index")
			if err != nil {
				return err
			}

			// 4. Build domain.UserFilter from flag values
			filter := domain.UserFilter{}
			if nameFilter != "" {
				filter.Name = &nameFilter
			}
			if emailFilter != "" {
				filter.Email = &emailFilter
			}
			if indexFilter >= 0 {
				filter.Index = &indexFilter
			}

			// 5. Resolve the replacement user from the store
			newUser, _, err := sel.ResolveUser(store, filter)
			if err != nil {
				return fmt.Errorf("resolve user: %w", err)
			}

			// 6. Create backup branch before any destructive operations
			backupName := fmt.Sprintf("backup/pre-replace-%d", time.Now().Unix())
			if err := gitcmd.CreateBackupBranch(backupName); err != nil {
				return fmt.Errorf("create backup branch: %w", err)
			}
			fmt.Print(i18n.T("cli.replace.backup_created", map[string]interface{}{"Branch": backupName}) + "\n")

			// 7. Rewrite history using git-filter-repo
			if err := filterRepoFunc(targetEmail, newUser.Name, newUser.Email); err != nil {
				return fmt.Errorf("filter-repo: %w", err)
			}

			// 8. Optionally update the repository-level git config
			yesFlag, err := cmd.Flags().GetBool("yes")
			if err != nil {
				return err
			}

			if !yesFlag {
				yes, err := confirmFunc(
					i18n.T("cli.replace.switch_confirm", map[string]interface{}{"Name": newUser.Name, "Email": newUser.Email}),
					false,
				)
				if err != nil {
					return fmt.Errorf("confirm: %w", err)
				}
				if !yes {
					return nil
				}
			}

			if err := gitcmd.SetConfig("name", newUser.Name, false); err != nil {
				return fmt.Errorf("set name config: %w", err)
			}
			if err := gitcmd.SetConfig("email", newUser.Email, false); err != nil {
				return fmt.Errorf("set email config: %w", err)
			}

			return nil
		},
	}

	cmd.Flags().String("with-name", "", i18n.T("cli.replace.flag_with_name", nil))
	cmd.Flags().String("with-email", "", i18n.T("cli.replace.flag_with_email", nil))
	cmd.Flags().Int("with-index", -1, i18n.T("cli.replace.flag_with_index", nil))
	cmd.Flags().BoolP("yes", "y", false, i18n.T("cli.replace.flag_yes", nil))

	return cmd
}
