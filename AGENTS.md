# PROJECT KNOWLEDGE BASE

**Generated:** 2026-05-31
**Commit:** 51662be
**Branch:** main

## OVERVIEW
`gitusr` is a Go CLI tool for managing and switching between git user identities (name/email). It persists users to a JSON file and applies them via `git config`.

## STRUCTURE
```
.
├── cmd/gitusr/          # Entry point (main.go)
├── internal/
│   ├── cli/             # Cobra commands (7 subcommands + tests)
│   ├── domain/          # Core types: User, UserStore interface, UserFilter
│   ├── format/          # Stderr/stdout formatting helpers
│   ├── gitcmd/          # Wrappers around `git config`, `git filter-repo`
│   ├── i18n/            # Internationalization (go-i18n/v2, embed TOML)
│   ├── prompt/          # Interactive survey prompts (AskNewUser, SelectUser)
│   ├── select/          # User resolution logic (ResolveUser)
│   ├── store/           # JSON persistence (JSONStore implements UserStore)
│   └── xdgpath/         # XDG data directory resolution
├── scripts/             # E2E and sandbox test scripts
└── test/integration/    # End-to-end integration tests
```

## WHERE TO LOOK
| Task | Location | Notes |
|------|----------|-------|
| Add a new CLI command | `internal/cli/` | Register in `root.go` |
| Change user data model | `internal/domain/user.go` | Update `UserStore` interface too |
| Change persistence format | `internal/store/store.go` | `JSONStore` implements `UserStore` |
| Change git interaction | `internal/gitcmd/runner.go` | Wraps `exec.Command("git", ...)` |
| Change prompts / UX | `internal/prompt/prompt.go` | Uses `survey/v2` |
| Change translations | `internal/i18n/*.toml` | Messages loaded via `//go:embed` |
| Add integration tests | `test/integration/` | Spins up real git repos |
| Mock interactive flow | `internal/cli/*_test.go` | Replace `askNewUser` or `SelectFunc` |

## CODE MAP

| Symbol | Type | Location | Role |
|--------|------|----------|------|
| `NewRootCmd` | func | `internal/cli/root.go` | Registers all subcommands |
| `domain.UserStore` | interface | `internal/domain/user.go` | Persistence abstraction |
| `store.JSONStore` | struct | `internal/store/store.go` | JSON file implementation |
| `ResolveUser` | func | `internal/select/resolver.go` | Priority: Index > Email > Name > Interactive |
| `gitcmd.SetConfig` | func | `internal/gitcmd/runner.go` | Calls `git config [--global]` |
| `prompt.AskNewUser` | func | `internal/prompt/prompt.go` | Interactive name/email prompt |
| `i18n.T` | func | `internal/i18n/bundle.go` | Message lookup with template data |
| `format.PrintErr` | func | `internal/format/format.go` | Prints errors to stderr |

## CONVENTIONS
- **Doc comments** on every exported symbol.
- **Error wrapping**: `fmt.Errorf("context: %w", err)` — always wrap, never swallow.
- **Constructor pattern**: command constructors are `NewXxxCmd(store domain.UserStore) *cobra.Command`.
- **Dependency injection**: commands depend on `domain.UserStore` interface, not concrete `JSONStore`.
- **Test mocking**: package-level vars (`askNewUser`, `SelectFunc`) are overridden in tests to avoid real interactivity.
- **Tests colocated**: `foo_test.go` lives next to `foo.go` (Go standard).
- **Global vs repo scope**: commands accept a `--global` / `-g` flag; without it they verify `gitcmd.IsGitRepo()`.
- **i18n message IDs**: use dot-separated paths like `cli.use.short` or `prompt.ask_name`.

## ANTI-PATTERNS (THIS PROJECT)
- Do **not** call `os.Exit` inside command `RunE` — return errors and let `main.go` handle exit codes.
- Do **not** duplicate validation that `store` already performs (e.g., duplicate user checks exist in both `cli/add.go` and `store.Add` — legacy overlap).
- Do **not** print directly to stdout inside `RunE` except for success messages — use `format.PrintErr` for errors.
- Do **not** call `format.PrintErr(err)` inside `RunE` and then `return err` — this causes double-printing because `main.go` also calls `format.PrintErr`. (`list.go` and `remove.go` have this bug.)
- Do **not** duplicate the `emailRegex` constant — it exists in both `prompt/prompt.go` and `cli/add.go`.
- Do **not** use `fmt.Errorf("%s", i18n.T(...))` — use `errors.New(i18n.T(...))` instead.

## UNIQUE STYLES
- `internal/select` is imported with alias `sel` throughout the codebase.
- `JSONStore` auto-creates an empty `[]domain.User{}` file on `List()` when missing.
- `FilterRepo` in `gitcmd` invokes `git-filter-repo` with an inline Python callback — requires `pip install git-filter-repo`.
- Locale detection priority: `GITUSR_LANG` > `LANGUAGE` > `LANG` > `"en"`. Chinese (`zh*`) is normalized to `zh-CN`, everything else to `en`.

## COMMANDS
```bash
# Build binary to bin/gitusr
mise run build

# Run all tests (unit + integration)
mise run test

# Install to ~/.local/bin/gitusr + symlink gu
mise run install

# Clean build artifacts
mise run clean
```

## NOTES
- Data file path: `$XDG_DATA_HOME/gitusr/user-list.json` (falls back to `~/.local/share/...`).
- `SilenceErrors: true` and `SilenceUsage: true` on root command means `main.go` must print errors explicitly via `format.PrintErr`.
- Integration tests create temporary git repositories on disk and exercise real `git` commands — require `git` to be installed.
- E2E scripts in `scripts/` spin up temporary environments and exercise the built binary.
- No CI / linting config present — no `.github/workflows/`, `.golangci.yml`, or pre-commit hooks.
