// Package i18n provides internationalization support for gitusr.
//
// It detects the user's locale from environment variables, embeds translation
// files via //go:embed, and exposes a simple T() function for message lookup
// with optional template data.
package i18n

import (
	"embed"
	"os"
	"strings"
	"sync"

	"github.com/BurntSushi/toml"
	"github.com/nicksnyder/go-i18n/v2/i18n"
	"golang.org/x/text/language"
)

//go:embed active.*.toml
var embedFS embed.FS

var (
	bundle    *i18n.Bundle
	localizer *i18n.Localizer
	mu        sync.Mutex
)

// Init initializes the i18n package with locale detection from environment
// variables. It checks GITUSR_LANG first, then LANGUAGE, then LANG, and
// defaults to "en" if none are set.
//
// Language normalization: if the detected language starts with "zh", it is
// mapped to "zh-CN"; all other values are mapped to "en".
//
// Init is safe to call multiple times; subsequent calls are no-ops.
// It panics if embedded message files cannot be loaded (acceptable for a CLI).
func Init() {
	mu.Lock()
	defer mu.Unlock()

	if localizer != nil {
		return
	}

	locale := detectLocale()
	initBundle(locale)
}

// InitWithLocale initializes the i18n package with a specific locale string.
// It does not read any environment variables, making it suitable for testing.
//
// The locale is normalized the same way as in Init: "zh*" → "zh-CN",
// everything else → "en".
func InitWithLocale(locale string) {
	mu.Lock()
	defer mu.Unlock()

	if localizer != nil {
		return
	}

	locale = normalizeLocale(locale)
	initBundle(locale)
}

// T translates the message identified by msgID using the configured localizer.
// Optional template data can be passed to interpolate placeholders in the
// translation messages.
//
// If the i18n package has not been initialized, or if the message ID is not
// found, T returns msgID unchanged (fail-open behaviour).
func T(msgID string, data map[string]interface{}) string {
	if localizer == nil {
		return msgID
	}

	msg, err := localizer.Localize(&i18n.LocalizeConfig{
		MessageID:   msgID,
		TemplateData: data,
	})
	if err != nil {
		return msgID
	}
	return msg
}

// detectLocale reads the preferred locale from environment variables.
// Priority: GITUSR_LANG > LANGUAGE > LANG > "en".
func detectLocale() string {
	for _, key := range []string{"GITUSR_LANG", "LANGUAGE", "LANG"} {
		if val := os.Getenv(key); val != "" {
			return normalizeLocale(val)
		}
	}
	return "en"
}

// normalizeLocale maps a raw locale string to a supported locale code.
// "zh" or any variant starting with "zh" → "zh-CN".
// Everything else → "en".
func normalizeLocale(lang string) string {
	// Trim surrounding whitespace and language tags (e.g. "zh_CN.UTF-8").
	lang = strings.TrimSpace(lang)
	if strings.HasPrefix(lang, "zh") {
		return "zh-CN"
	}
	return "en"
}

// initBundle creates the message Bundle, registers the TOML unmarshaler,
// loads all embedded translation files, and creates the Localizer.
//
// It panics on any error because a CLI cannot meaningfully recover from a
// broken i18n setup.
func initBundle(locale string) {
	b := i18n.NewBundle(language.English)
	b.RegisterUnmarshalFunc("toml", toml.Unmarshal)

	// Load all embedded translation files matching active.*.toml
	entries, err := embedFS.ReadDir(".")
	if err != nil {
		panic("i18n: failed to read embedded directory: " + err.Error())
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		if !strings.HasPrefix(name, "active.") || !strings.HasSuffix(name, ".toml") {
			continue
		}
		if _, err := b.LoadMessageFileFS(embedFS, name); err != nil {
			panic("i18n: failed to load message file " + name + ": " + err.Error())
		}
	}

	// Build the locale priority list: detected locale first, then "en", then
	// any other known locale that isn't already covered.
	locales := []string{locale}
	if locale != "en" {
		locales = append(locales, "en")
	}
	if locale != "zh-CN" {
		locales = append(locales, "zh-CN")
	}

	l := i18n.NewLocalizer(b, locales...)

	bundle = b
	localizer = l
}
