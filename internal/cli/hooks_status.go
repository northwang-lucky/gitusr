package cli

import (
	"fmt"
	"slices"

	"github.com/spf13/cobra"

	"github.com/northwang-lucky/gitusr/internal/hook"
	"github.com/northwang-lucky/gitusr/internal/i18n"
)

// NewHooksStatusCmd creates the "hooks status" command. It reports, for every
// known hook type in canonical order, whether the wrapper covers it and
// whether the interception is currently active:
//   - not installed: absent from hook state (never installed or uninstalled);
//     the disabled flag is irrelevant for such a type and takes no precedence
//   - disabled: installed but switched off via "hooks disable"
//   - enabled: installed and active
func NewHooksStatusCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "status",
		Short: i18n.T("cli.hooks.status.short", nil),
		RunE: func(cmd *cobra.Command, args []string) error {
			state, err := hook.LoadState()
			if err != nil {
				return err
			}

			for _, ht := range hook.AllHookTypes {
				fmt.Println(i18n.T("cli.hooks.status.row", map[string]interface{}{
					"Type":   ht,
					"Status": statusOfHook(state, ht),
				}))
			}
			return nil
		},
	}

	return cmd
}

// statusOfHook maps the raw hook state to the user-facing status label of a
// single hook type.
func statusOfHook(state *hook.HookState, ht hook.HookType) string {
	if !slices.Contains(state.InstalledTypes, ht) {
		return i18n.T("cli.hooks.status.not_installed", nil)
	}
	if slices.Contains(state.DisabledTypes, ht) {
		return i18n.T("cli.hooks.status.disabled", nil)
	}
	return i18n.T("cli.hooks.status.enabled", nil)
}
