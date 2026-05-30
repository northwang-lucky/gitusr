package format

import (
	"fmt"
	"os"
	"strings"

	"gitusr/internal/domain"
)

// PrintErr writes the given message prefixed with "gitusr error: " to stderr.
func PrintErr(msg string) {
	fmt.Fprintf(os.Stderr, "gitusr error: %s\n", msg)
}

// PrintOptions controls the behaviour of PrintUserInfo.
type PrintOptions struct {
	Global      bool
	ShowSuccess bool
}

// PrintUserInfo prints the current git user configuration to stdout.
// If ShowSuccess is true, a "Success!" line is printed as a prefix.
func PrintUserInfo(user domain.User, opts PrintOptions) {
	var scope string
	if opts.Global {
		scope = "global"
	} else {
		scope = "repo"
	}

	var b strings.Builder
	if opts.ShowSuccess {
		b.WriteString("Success!\n")
	}

	b.WriteString(fmt.Sprintf("Your %s git user is:\n\n", scope))
	b.WriteString(fmt.Sprintf("user.name  = %s\n", user.Name))
	b.WriteString(fmt.Sprintf("user.email = %s\n", user.Email))

	fmt.Print(b.String())
}

// FormatUserList returns a numbered list of users with aligned columns.
// Each line has the format: "N: Name: <name> | Email: <email>"
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
		line := fmt.Sprintf("%d: Name: %-*s | Email: %s\n", i, maxNameLen, u.Name, u.Email)
		b.WriteString(line)
	}

	return b.String()
}
