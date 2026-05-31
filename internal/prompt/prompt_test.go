package prompt

import (
	"testing"

	"gitusr/internal/domain"
	"gitusr/internal/i18n"
)

// --- Existing tests (updated for i18n) ---

func TestValidateEmail_Valid(t *testing.T) {
	i18n.ResetForTesting()
	i18n.InitWithLocale("en")

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
	i18n.ResetForTesting()
	i18n.InitWithLocale("en")

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
	i18n.ResetForTesting()
	i18n.InitWithLocale("en")

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
	i18n.ResetForTesting()
	i18n.InitWithLocale("en")

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
	i18n.ResetForTesting()
	i18n.InitWithLocale("en")

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
	i18n.ResetForTesting()
	i18n.InitWithLocale("en")

	// survey.Select with empty options will error — verify behaviour
	_, err := SelectUser(nil)
	if err == nil {
		t.Error("SelectUser with nil users should return an error")
	}

	want := "no users to select from"
	if err.Error() != want {
		t.Errorf("expected error %q, got %q", want, err.Error())
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

// --- New locale-aware tests ---

// TestAskNewUser_En verifies that the English prompt messages are correct.
// Since AskNewUser calls survey.AskOne (blocking), we test i18n.T() output
// directly for the keys used by AskNewUser.
func TestAskNewUser_En(t *testing.T) {
	i18n.ResetForTesting()
	i18n.InitWithLocale("en")

	tests := []struct {
		key  string
		want string
	}{
		{"prompt.ask_name", "Please enter the user.name:"},
		{"prompt.ask_email", "Please enter the user.email:"},
		{"prompt.internal_error", "internal error: expected string value"},
	}

	for _, tt := range tests {
		got := i18n.T(tt.key, nil)
		if got != tt.want {
			t.Errorf("i18n.T(%q, nil) = %q, want %q", tt.key, got, tt.want)
		}
	}
}

// TestAskNewUser_ZhCN verifies that the Chinese prompt messages are correct.
func TestAskNewUser_ZhCN(t *testing.T) {
	i18n.ResetForTesting()
	i18n.InitWithLocale("zh-CN")

	tests := []struct {
		key  string
		want string
	}{
		{"prompt.ask_name", "请输入 user.name："},
		{"prompt.ask_email", "请输入 user.email："},
		{"prompt.internal_error", "内部错误：期望字符串值"},
	}

	for _, tt := range tests {
		got := i18n.T(tt.key, nil)
		if got != tt.want {
			t.Errorf("i18n.T(%q, nil) = %q, want %q", tt.key, got, tt.want)
		}
	}
}

// TestSelectUser_En verifies the English select message and no-users error.
func TestSelectUser_En(t *testing.T) {
	i18n.ResetForTesting()
	i18n.InitWithLocale("en")

	got := i18n.T("prompt.select_user", nil)
	want := "Select a user:"
	if got != want {
		t.Errorf("i18n.T(%q, nil) = %q, want %q", "prompt.select_user", got, want)
	}

	got = i18n.T("prompt.no_users", nil)
	want = "no users to select from"
	if got != want {
		t.Errorf("i18n.T(%q, nil) = %q, want %q", "prompt.no_users", got, want)
	}
}

// TestSelectUser_ZhCN verifies the Chinese select message and no-users error.
func TestSelectUser_ZhCN(t *testing.T) {
	i18n.ResetForTesting()
	i18n.InitWithLocale("zh-CN")

	got := i18n.T("prompt.select_user", nil)
	want := "选择一个用户："
	if got != want {
		t.Errorf("i18n.T(%q, nil) = %q, want %q", "prompt.select_user", got, want)
	}

	got = i18n.T("prompt.no_users", nil)
	want = "没有可供选择的用户"
	if got != want {
		t.Errorf("i18n.T(%q, nil) = %q, want %q", "prompt.no_users", got, want)
	}
}

// TestValidateEmail_En verifies validateEmail returns the correct English error message.
func TestValidateEmail_En(t *testing.T) {
	i18n.ResetForTesting()
	i18n.InitWithLocale("en")

	err := validateEmail("notanemail")
	if err == nil {
		t.Fatal("expected error for invalid email")
	}

	want := "invalid email format"
	if err.Error() != want {
		t.Errorf("expected error %q, got %q", want, err.Error())
	}
}

// TestValidateEmail_ZhCN verifies validateEmail returns the correct Chinese error message.
func TestValidateEmail_ZhCN(t *testing.T) {
	i18n.ResetForTesting()
	i18n.InitWithLocale("zh-CN")

	err := validateEmail("notanemail")
	if err == nil {
		t.Fatal("expected error for invalid email")
	}

	want := "邮箱格式无效"
	if err.Error() != want {
		t.Errorf("expected error %q, got %q", want, err.Error())
	}
}

// TestFormatSelectOptions_ZhCN verifies Chinese locale formatting.
func TestFormatSelectOptions_ZhCN(t *testing.T) {
	i18n.ResetForTesting()
	i18n.InitWithLocale("zh-CN")

	users := []domain.User{
		{Name: "Alice", Email: "a@t.com"},
		{Name: "Bob", Email: "b@t.com"},
	}

	result := formatSelectOptions(users)
	if len(result) != 2 {
		t.Fatalf("expected 2 options, got %d", len(result))
	}

	// Names are padded to "Alice" length (5 chars)
	expected := []string{
		"姓名：Alice | 邮箱：a@t.com",
		"姓名：Bob   | 邮箱：b@t.com",
	}

	for i, exp := range expected {
		if result[i] != exp {
			t.Errorf("option %d:\nexpected %q\ngot      %q", i, exp, result[i])
		}
	}
}
