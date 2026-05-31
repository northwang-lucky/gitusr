package i18n

import (
	"testing"
)

// reset clears the package-level i18n state so that Init() and InitWithLocale()
// actually re-initialize instead of being a no-op. Must be called at the start
// of every test that depends on a specific locale configuration.
func reset() {
	mu.Lock()
	defer mu.Unlock()
	bundle = nil
	localizer = nil
}

// ---------------------------------------------------------------------------
// Init() – locale detection from environment variables
// ---------------------------------------------------------------------------

func TestInit_En(t *testing.T) {
	reset()
	t.Setenv("GITUSR_LANG", "en")

	Init()

	got := T("format.success_banner", nil)
	want := "Success!"
	if got != want {
		t.Errorf("T(\"format.success_banner\", nil) = %q, want %q", got, want)
	}
}

func TestInit_ZhCN(t *testing.T) {
	reset()
	t.Setenv("GITUSR_LANG", "zh-CN")

	Init()

	got := T("format.success_banner", nil)
	want := "成功！"
	if got != want {
		t.Errorf("T(\"format.success_banner\", nil) = %q, want %q", got, want)
	}
}

func TestInit_Priority(t *testing.T) {
	reset()
	// GITUSR_LANG should win over LANG
	t.Setenv("GITUSR_LANG", "en")
	t.Setenv("LANG", "zh_CN.UTF-8")

	Init()

	got := T("format.success_banner", nil)
	want := "Success!"
	if got != want {
		t.Errorf("T(\"format.success_banner\", nil) = %q, want %q", got, want)
	}
}

func TestInit_Fallback(t *testing.T) {
	reset()
	// Unsupported locale "fr" should fall back to English
	t.Setenv("GITUSR_LANG", "fr")

	Init()

	got := T("format.success_banner", nil)
	want := "Success!"
	if got != want {
		t.Errorf("T(\"format.success_banner\", nil) = %q, want %q", got, want)
	}
}

func TestInit_LanguageVar(t *testing.T) {
	reset()
	// No GITUSR_LANG, only LANGUAGE — should detect zh-CN
	t.Setenv("LANGUAGE", "zh_CN")

	Init()

	got := T("format.success_banner", nil)
	want := "成功！"
	if got != want {
		t.Errorf("T(\"format.success_banner\", nil) = %q, want %q", got, want)
	}
}

func TestInit_LangVar(t *testing.T) {
	reset()
	// No GITUSR_LANG or LANGUAGE, only LANG — should detect zh-CN
	t.Setenv("LANG", "zh_CN.UTF-8")

	Init()

	got := T("format.success_banner", nil)
	want := "成功！"
	if got != want {
		t.Errorf("T(\"format.success_banner\", nil) = %q, want %q", got, want)
	}
}

// ---------------------------------------------------------------------------
// T() – message lookup
// ---------------------------------------------------------------------------

func TestT_Existing(t *testing.T) {
	reset()
	InitWithLocale("en")

	got := T("format.success_banner", nil)
	want := "Success!"
	if got != want {
		t.Errorf("T(\"format.success_banner\", nil) = %q, want %q", got, want)
	}
}

func TestT_Missing(t *testing.T) {
	reset()
	InitWithLocale("en")

	got := T("nonexistent.key", nil)
	want := "nonexistent.key"
	if got != want {
		t.Errorf("T(\"nonexistent.key\", nil) = %q, want %q", got, want)
	}
}

func TestT_WithData(t *testing.T) {
	reset()
	InitWithLocale("en")

	got := T("cli.add.success", map[string]any{
		"Name":  "Alice",
		"Email": "a@b.com",
	})
	want := "Success! User (name: Alice | email: a@b.com) has been saved!"
	if got != want {
		t.Errorf("T(\"cli.add.success\", ...) = %q, want %q", got, want)
	}
}

// ---------------------------------------------------------------------------
// InitWithLocale() – direct locale bypassing environment variables
// ---------------------------------------------------------------------------

func TestInitWithLocale(t *testing.T) {
	reset()
	// No env vars set — InitWithLocale should work independently
	InitWithLocale("zh-CN")

	got := T("format.success_banner", nil)
	want := "成功！"
	if got != want {
		t.Errorf("T(\"format.success_banner\", nil) = %q, want %q", got, want)
	}
}
