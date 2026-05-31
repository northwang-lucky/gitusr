package cli

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"gitusr/internal/domain"
	"gitusr/internal/format"
	"gitusr/internal/gitcmd"
	"gitusr/internal/i18n"
	"gitusr/internal/prompt"
)

// Package-level function vars allow tests to inject mock implementations.
var (
	getConfigFn     = gitcmd.GetConfig
	setConfigFn     = gitcmd.SetConfig
	askNewUserFn    = prompt.AskNewUser
	confirmFn       = prompt.Confirm
	printUserInfoFn = format.PrintUserInfo
	userHomeDirFn   = os.UserHomeDir
)

// NewInitCmd creates the "init" cobra command that initializes the user store
// from the git global config, with optional legacy-config migration.
func NewInitCmd(store domain.UserStore) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "init",
		Short: i18n.T("cli.init.short", nil),
		RunE: func(cmd *cobra.Command, args []string) error {
			force, _ := cmd.Flags().GetBool("force")
			name, _ := cmd.Flags().GetString("name")
			email, _ := cmd.Flags().GetString("email")
			yes, _ := cmd.Flags().GetBool("yes")
			return runInit(store, force, name, email, yes)
		},
	}
	cmd.Flags().BoolP("force", "f", false, i18n.T("cli.init.flag_force", nil))
	cmd.Flags().StringP("name", "n", "", i18n.T("cli.init.flag_name", nil))
	cmd.Flags().StringP("email", "e", "", i18n.T("cli.init.flag_email", nil))
	cmd.Flags().BoolP("yes", "y", false, i18n.T("cli.init.flag_yes", nil))
	return cmd
}

// runInit contains the core logic extracted for testability.
func runInit(store domain.UserStore, force bool, name string, email string, yes bool) error {
	// Step 0: Legacy config migration
	if migrated, err := migrateLegacy(store); err != nil {
		return err
	} else if migrated {
		return nil
	}

	// Step 1: Obtain global user from git config, prompt, or flags
	var globalUser domain.User
	if name != "" && email != "" {
		globalUser = domain.User{Name: name, Email: email}
	} else if name != "" || email != "" {
		return fmt.Errorf("%s", i18n.T("cli.init.err_name_and_email_required", nil))
	} else {
		var err error
		globalUser, err = getOrPromptGlobalUser()
		if err != nil {
			return err
		}
	}

	// Step 2: Check if store is already initialized
	if store.IsInitialized() && !force && !yes {
		confirmed, err := confirmFn(i18n.T("cli.init.override_confirm", nil), false)
		if err != nil {
			return err
		}
		if !confirmed {
			return nil
		}
	}

	// Step 3: Persist
	if err := store.SaveAll([]domain.User{globalUser}); err != nil {
		return err
	}

	// Step 4: Print result
	printUserInfoFn(globalUser, format.PrintOptions{Global: true, ShowSuccess: true})
	return nil
}

// migrateLegacy checks for a legacy config file at ~/.gitusr/user-list.json,
// prompts the user, and migrates the data if confirmed. Returns true when
// migration was performed (caller should then skip normal init).
func migrateLegacy(store domain.UserStore) (bool, error) {
	homeDir, err := userHomeDirFn()
	if err != nil {
		return false, err
	}
	legacyPath := filepath.Join(homeDir, ".gitusr", "user-list.json")

	data, err := os.ReadFile(legacyPath)
	if err != nil {
		// File does not exist or is not readable – no legacy config
		return false, nil
	}

	var oldUsers []domain.User
	if err := json.Unmarshal(data, &oldUsers); err != nil {
		// Malformed file – treat as if there is no legacy config
		return false, nil
	}

	if len(oldUsers) == 0 {
		// Empty list – nothing to migrate
		return false, nil
	}

	confirmed, err := confirmFn(i18n.T("cli.init.migrate_confirm", nil), true)
	if err != nil {
		return false, err
	}

	if !confirmed {
		fmt.Println(i18n.T("cli.init.migrate_skip", nil))
		return false, nil
	}

	if err := store.SaveAll(oldUsers); err != nil {
		return false, err
	}

	fmt.Print(i18n.T("cli.init.migrate_count", map[string]interface{}{"Count": len(oldUsers)}) + "\n")
	return true, nil
}

// getOrPromptGlobalUser reads the git global user.name and user.email config
// values. If either is missing it falls back to an interactive prompt and then
// persists the new values into the git global config.
func getOrPromptGlobalUser() (domain.User, error) {
	name, nameErr := getConfigFn("name", true)
	email, emailErr := getConfigFn("email", true)

	if nameErr != nil || emailErr != nil {
		user, err := askNewUserFn()
		if err != nil {
			return domain.User{}, err
		}
		if err := setConfigFn("name", user.Name, true); err != nil {
			return domain.User{}, err
		}
		if err := setConfigFn("email", user.Email, true); err != nil {
			return domain.User{}, err
		}
		return user, nil
	}

	return domain.User{Name: name, Email: email}, nil
}
