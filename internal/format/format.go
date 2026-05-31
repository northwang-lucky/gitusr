package format

import (
	"fmt"
	"os"
	"strings"

	"gitusr/internal/domain"
	"gitusr/internal/i18n"
)

// PrintErr writes the given message prefixed with a localised error label to stderr.
func PrintErr(msg string) {
	prefix := i18n.T("format.error_prefix", map[string]interface{}{"Msg": msg})
	fmt.Fprintf(os.Stderr, "%s\n", prefix)
}

// PrintOptions controls the behaviour of PrintUserInfo.
type PrintOptions struct {
	Global      bool
	ShowSuccess bool
}

// PrintUserInfo prints the current git user configuration to stdout.
// If ShowSuccess is true, a localised success line is printed as a prefix.
func PrintUserInfo(user domain.User, opts PrintOptions) {
	var scope string
	if opts.Global {
		scope = "global"
	} else {
		scope = "repo"
	}

	var b strings.Builder
	if opts.ShowSuccess {
		b.WriteString(i18n.T("format.success_banner", nil) + "\n")
	}

	b.WriteString(i18n.T("format.userinfo_header", map[string]interface{}{"Scope": scope}) + "\n\n")
	b.WriteString(i18n.T("format.user_name_label", map[string]interface{}{"Name": user.Name}) + "\n")
	b.WriteString(i18n.T("format.user_email_label", map[string]interface{}{"Email": user.Email}) + "\n")

	fmt.Print(b.String())
}

// FormatUserList returns a numbered list of users with aligned columns.
// Each line uses a localised format like "N: Name: <name> | Email: <email>"
// where <name> is padded to the length of the longest name in the list.
func FormatUserList(users []domain.User) string {
	if len(users) == 0 {
		return ""
	}

	// Determine the longest name length for alignment
	maxNameLen := 0
	for _, u := range users {
		if len(u.Name) > maxNameLen {
			maxNameLen = len(u.Name)
		}
	}

	var b strings.Builder
	for i, u := range users {
		paddedName := fmt.Sprintf("%-*s", maxNameLen, u.Name)
		line := i18n.T("format.userlist_row", map[string]interface{}{
			"Index":   i,
			"NamePad": paddedName,
			"Email":   u.Email,
		}) + "\n"
		b.WriteString(line)
	}

	return b.String()
}
