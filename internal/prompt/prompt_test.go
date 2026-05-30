package prompt

import (
	"testing"

	"gitusr/internal/domain"
)

func TestValidateEmail_Valid(t *testing.T) {
	validEmails := []string{
		"a@b.c",
		"user@example.com",
		"first.last@company.co.uk",
		"user+tag@domain.org",
		"u@d.co",
		"hello@world123.net",
		"x@y.z",
	}

	for _, email := range validEmails {
		err := validateEmail(email)
		if err != nil {
			t.Errorf("validateEmail(%q) expected nil error, got: %v", email, err)
		}
	}
}

func TestValidateEmail_Invalid(t *testing.T) {
	invalidEmails := []string{
		"",
		"notanemail",
		"@missinguser.com",
		"missingdomain@",
		"missing@dot",
		"spaces in@email.com",
		"no@tld.",
		"just@.",
	}

	for _, email := range invalidEmails {
		err := validateEmail(email)
		if err == nil {
			t.Errorf("validateEmail(%q) expected error, got nil", email)
		}
	}
}

func TestFormatSelectOptions_Empty(t *testing.T) {
	result := formatSelectOptions(nil)
	if len(result) != 0 {
		t.Errorf("expected empty slice, got %v", result)
	}

	result = formatSelectOptions([]domain.User{})
	if len(result) != 0 {
		t.Errorf("expected empty slice, got %v", result)
	}
}

func TestFormatSelectOptions_Single(t *testing.T) {
	users := []domain.User{
		{Name: "Alice", Email: "alice@example.com"},
	}

	result := formatSelectOptions(users)
	if len(result) != 1 {
		t.Fatalf("expected 1 option, got %d", len(result))
	}

	expected := "Name: Alice | Email: alice@example.com"
	if result[0] != expected {
		t.Errorf("expected %q, got %q", expected, result[0])
	}
}

func TestFormatSelectOptions_MultipleAligned(t *testing.T) {
	users := []domain.User{
		{Name: "Alice", Email: "a@t.com"},
		{Name: "Bob", Email: "b@t.com"},
		{Name: "Charlie", Email: "c@t.com"},
	}

	result := formatSelectOptions(users)
	if len(result) != 3 {
		t.Fatalf("expected 3 options, got %d", len(result))
	}

	// Names are padded to "Charlie" length (7 chars)
	expected := []string{
		"Name: Alice   | Email: a@t.com",
		"Name: Bob     | Email: b@t.com",
		"Name: Charlie | Email: c@t.com",
	}

	for i, exp := range expected {
		if result[i] != exp {
			t.Errorf("option %d:\nexpected %q\ngot      %q", i, exp, result[i])
		}
	}
}

// Test that exported functions are correctly typed and callable.
// We cannot easily test the full interactive survey flow without mocking,
// but we verify the function signatures compile and handle empty input.
func TestSelectUser_EmptyList(t *testing.T) {
	// survey.Select with empty options will error — verify behaviour
	_, err := SelectUser(nil)
	if err == nil {
		t.Error("SelectUser with nil users should return an error")
	}
}

func TestAskNewUser_Exported(t *testing.T) {
	// Verify AskNewUser returns the correct types.
	// Full interactive test requires terminal input mocking which is not
	// practical in a unit test. The function signature is validated by
	// compilation and we test the email validation separately.
	var _ func() (domain.User, error) = AskNewUser
}

func TestConfirm_Exported(t *testing.T) {
	// Verify Confirm returns the correct types.
	var _ func(string, bool) (bool, error) = Confirm
}
