# CLI COMMANDS

**Scope:** `internal/cli/`

## OVERVIEW
All Cobra subcommands for gitusr live here. Each command has its own file plus a matching `_test.go`. The root command (`root.go`) registers every subcommand and configures i18n-aware help templates.

## STRUCTURE
```
internal/cli/
├── root.go           # Root command registration and help template
├── root_test.go      # Root command tests
├── add.go            # Add a new user identity
├── add_test.go
├── current.go        # Show current git user config
├── current_test.go
├── init.go           # Initialize / import existing git config
├── init_test.go
├── list.go           # List saved identities
├── list_test.go
├── remove.go         # Remove a saved identity
├── remove_test.go
├── replace.go        # Replace identity in git history (filter-repo)
├── replace_test.go
├── use.go            # Switch active git user
└── use_test.go
```

## WHERE TO LOOK
| Task | Location | Notes |
|------|----------|-------|
| Add a new subcommand | `root.go` | Append to `cmd.AddCommand(...)` |
| Change command flags | `<cmd>.go` | Use `cmd.Flags().BoolP/StringP/IntP` |
| Test command output | `*_test.go` | Capture stdout/stderr, override `askNewUser` |
| Test interactive flow | `*_test.go` | Override `askNewUser` or `sel.SelectFunc` |

## CONVENTIONS
- Command constructors: `NewXxxCmd(store domain.UserStore) *cobra.Command`.
- Flags use short + long forms (`-g` / `--global`, `-n` / `--name`, etc.).
- `buildFilter(cmd)` helper extracts `UserFilter` from flags in `use.go`.
- Success output goes through `format.PrintUserInfo` or `format.PrintSuccess`.
- Errors are returned from `RunE`; never `fmt.Println` inside `RunE`.
- **Test mocking**: package-level vars (`askNewUser`, `SelectFunc`, `getConfigFn`, etc.) are overridden in tests. Use `t.Cleanup()` to restore original values.
- **Test helpers**: `executeCmd` (defined in `current_test.go`) runs a Cobra command and captures stdout/stderr. `mockStore` (defined in `list_test.go`) is the standard `domain.UserStore` fake.
- **i18n test naming**: use `_En` / `_ZhCN` suffixes for locale-specific test cases.

## ANTI-PATTERNS
- Do **not** add `os.Exit` inside `RunE`.
- Do **not** print errors directly to stdout — return them or use `format.PrintErr`.
- Do **not** create standalone CLI logic outside this package.

## COMMANDS
```bash
# Run only CLI package tests
go test ./internal/cli/... -v

# Run a specific command test
go test ./internal/cli/... -run TestAdd -v
```
