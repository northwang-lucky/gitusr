package prompt

import (
	"errors"
	"fmt"
	"regexp"

	"github.com/AlecAivazis/survey/v2"
	"gitusr/internal/domain"
)

const emailRegex = `\w+([-+.]\w+)*@\w+([-.]\w+)*\.\w+([-.]\w+)*`

var compiledEmailRegex = regexp.MustCompile("^" + emailRegex + "$")

// validateEmail checks whether the given string matches the email format
// expected by this package.
func validateEmail(email string) error {
	if !compiledEmailRegex.MatchString(email) {
		return errors.New("invalid email format")
	}
	return nil
}

// AskNewUser interactively prompts for user.name and user.email.
// The email input is validated against a standard email regex.
func AskNewUser() (domain.User, error) {
	var name string
	err := survey.AskOne(&survey.Input{
		Message: "Please enter the user.name:",
	}, &name)
	if err != nil {
		return domain.User{}, err
	}

	var email string
	err = survey.AskOne(&survey.Input{
		Message: "Please enter the user.email:",
	}, &email, survey.WithValidator(survey.ComposeValidators(
		survey.Required,
		func(val interface{}) error {
			s, ok := val.(string)
			if !ok {
				return errors.New("internal error: expected string value")
			}
			return validateEmail(s)
		},
	)))
	if err != nil {
		return domain.User{}, err
	}

	return domain.User{Name: name, Email: email}, nil
}

// formatSelectOptions builds a slice of display strings for use in a
// survey.Select prompt. Names are left-padded to the longest name in
// the list so that the vertical pipe separator aligns.
func formatSelectOptions(users []domain.User) []string {
	if len(users) == 0 {
		return nil
	}

	maxNameLen := 0
	for _, u := range users {
		if len(u.Name) > maxNameLen {
			maxNameLen = len(u.Name)
		}
	}

	options := make([]string, len(users))
	for i, u := range users {
		options[i] = fmt.Sprintf("Name: %-*s | Email: %s", maxNameLen, u.Name, u.Email)
	}
	return options
}

// SelectUser displays a numbered list of users and returns the index of
// the selected user. An error is returned when the list is empty or the
// user cancels the prompt.
func SelectUser(users []domain.User) (int, error) {
	if len(users) == 0 {
		return 0, errors.New("no users to select from")
	}

	options := formatSelectOptions(users)

	var selected int
	err := survey.AskOne(&survey.Select{
		Message: "Select a user:",
		Options: options,
	}, &selected)
	if err != nil {
		return 0, err
	}

	return selected, nil
}

// Confirm asks the user a yes/no question and returns the user's answer.
func Confirm(msg string, defaultVal bool) (bool, error) {
	var result bool
	err := survey.AskOne(&survey.Confirm{
		Message: msg,
		Default: defaultVal,
	}, &result)
	if err != nil {
		return false, err
	}
	return result, nil
}
