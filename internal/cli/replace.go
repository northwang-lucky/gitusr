package cli

import (
	"errors"
	"fmt"
	"time"

	"github.com/spf13/cobra"

	"gitusr/internal/domain"
	"gitusr/internal/gitcmd"
	"gitusr/internal/prompt"
	sel "gitusr/internal/select"
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
		Short: "replace author in git history",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			targetEmail := args[0]

			// 1. Safety: must be inside a git repository
			if !gitcmd.IsGitRepo() {
				return errors.New("not a git repository (or any of the parent directories): .git")
			}

			// 2. Safety: no uncommitted changes allowed
			if gitcmd.HasUncommittedChanges() {
				return errors.New("uncommitted changes detected, please commit or stash them first")
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
			fmt.Printf("Created backup branch: %s\n", backupName)

			// 7. Rewrite history using git-filter-repo
			if err := filterRepoFunc(targetEmail, newUser.Name, newUser.Email); err != nil {
				return fmt.Errorf("filter-repo: %w", err)
			}

			// 8. Optionally update the repository-level git config
			yes, err := confirmFunc(
				fmt.Sprintf("Switch repo user to %s <%s>?", newUser.Name, newUser.Email),
				false,
			)
			if err != nil {
				return fmt.Errorf("confirm: %w", err)
			}

			if yes {
				if err := gitcmd.SetConfig("name", newUser.Name, false); err != nil {
					return fmt.Errorf("set name config: %w", err)
				}
				if err := gitcmd.SetConfig("email", newUser.Email, false); err != nil {
					return fmt.Errorf("set email config: %w", err)
				}
			}

			return nil
		},
	}

	cmd.Flags().String("with-name", "", "new user name")
	cmd.Flags().String("with-email", "", "new user email")
	cmd.Flags().Int("with-index", -1, "new user index")

	return cmd
}
