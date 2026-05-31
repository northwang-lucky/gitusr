package cli

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/northwang-lucky/gitusr/internal/domain"
	"github.com/northwang-lucky/gitusr/internal/i18n"
	"github.com/northwang-lucky/gitusr/internal/version"
)

// NewRootCmd creates the root cobra command for gitusr and registers all
// subcommands. The store parameter is injected into every subcommand that
// requires persistence. The name parameter controls the command's Use field,
// allowing "gitusr" or "gu" depending on the invocation alias.
func NewRootCmd(store domain.UserStore, name string) *cobra.Command {
	cmd := &cobra.Command{
		Use:           name,
		Short:         i18n.T("cli.root.short", nil),
		SilenceErrors: true,
		SilenceUsage:  true,
		Version:       version.Version,
	}

	cmd.AddCommand(
		NewAddCmd(store),
		NewCurrentCmd(),
		NewHookCmd(store),
		NewInitCmd(store),
		NewListCmd(store),
		NewRemoveCmd(store),
		NewReplaceCmd(store),
		NewUseCmd(store),
	)

	cmd.InitDefaultCompletionCmd()
	cmd.InitDefaultHelpCmd()
	cmd.InitDefaultHelpFlag()

	for _, sub := range cmd.Commands() {
		switch sub.Name() {
		case "completion":
			sub.Short = i18n.T("cli.completion.short", nil)
		case "help":
			sub.Short = i18n.T("cli.help.short", nil)
		}
	}

	if f := cmd.Flags().Lookup("help"); f != nil {
		f.Usage = i18n.T("cli.framework.help_flag", map[string]any{"Command": name})
	}

	usageTemplate := fmt.Sprintf(`Usage:
  {{.UseLine}}{{if .HasAvailableSubCommands}}

%s{{range .Commands}}{{if (or .IsAvailableCommand (eq .Name "help"))}}
  {{rpad .Name .NamePadding }} {{.Short}}{{end}}{{end}}{{end}}{{if .HasAvailableLocalFlags}}

%s
{{.LocalFlags.FlagUsages | trimTrailingWhitespaces}}{{end}}{{if .HasAvailableInheritedFlags}}

%s
{{.InheritedFlags.FlagUsages | trimTrailingWhitespaces}}{{end}}{{if .HasHelpSubCommands}}

%s{{range .Commands}}{{if .IsAdditionalHelpTopicCommand}}
  {{rpad .Name .NamePadding }} {{.Short}}{{end}}{{end}}{{end}}{{if .HasAvailableSubCommands}}

%s{{end}}`,
		i18n.T("cli.framework.available_commands", nil),
		i18n.T("cli.framework.flags", nil),
		i18n.T("cli.framework.global_flags", nil),
		i18n.T("cli.framework.additional_help", nil),
		i18n.T("cli.framework.use_help", nil),
	)
	cmd.SetUsageTemplate(usageTemplate)

	return cmd
}
