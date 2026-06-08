# I18N PACKAGE

**Scope:** `internal/i18n/`

## OVERVIEW
Embedded translation bundle and locale state for CLI output. It loads `active.*.toml`, normalizes locale, and exposes fail-open `T()` lookups.

## WHERE TO LOOK
| Task | Location | Notes |
|------|----------|-------|
| Add or rename message IDs | `active.en.toml`, `active.zh-CN.toml` | Keep both files aligned by key; command tests assert both locales |
| Change locale detection | `bundle.go` | Priority is `GITUSR_LANG` > `LANGUAGE` > `LANG` > `en` |
| Test language behavior | `bundle_test.go`, command package tests | Use `ResetForTesting()` before switching locale |
| Change help text wording | `active.*.toml`, `internal/cli/root.go` | Root usage template pulls section labels from i18n |

## CODE MAP
| Symbol | Type | Location | Role |
|--------|------|----------|------|
| `Init` | func | `bundle.go` | Detects locale and initializes embedded bundle once |
| `InitWithLocale` | func | `bundle.go` | Test-oriented explicit locale initialization |
| `ResetForTesting` | func | `bundle.go` | Clears singleton bundle/localizer for locale tests |
| `T` | func | `bundle.go` | Returns translated message or the msgID when unavailable |
| `normalizeLocale` | func | `bundle.go` | Maps `zh*` to `zh-CN`; everything else to `en` |

## CONVENTIONS
- Translation files must be named `active.<locale>.toml`; `bundle.go` embeds only `active.*.toml`.
- Message lookup is fail-open: missing initialization or missing key returns the key itself.
- Tests that change locale must call `ResetForTesting()` and avoid leaking singleton state.
- Command tests with localized expectations use `_En` / `_ZhCN` suffixes.
- Root help section labels are translated through `internal/cli/root.go`, not Cobra defaults.

## ANTI-PATTERNS
- Do **not** add a message ID to only one locale file.
- Do **not** call `ResetForTesting()` from production code.
- Do **not** add unsupported locale files without updating normalization and tests.

## COMMANDS
```bash
# Run i18n unit tests
go test ./internal/i18n/... -v

# Run command tests affected by translations
go test ./internal/cli/... -run 'I18n|Locale|Help|Usage' -v
```
