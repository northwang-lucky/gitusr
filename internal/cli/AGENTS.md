# CLI COMMANDS

**Scope:** `internal/cli/`

## OVERVIEW
Cobra command layer for gitusr. It owns flag parsing, i18n help text, user-facing command errors, and package-level test seams around prompts/git/hook operations.

## STRUCTURE
```
internal/cli/
├── root.go              # Root command registration + custom i18n usage template
├── add.go              # Add saved identity; non-interactive flags or prompt
├── current.go          # Show repo/global git config
├── hook.go             # `hook install/uninstall`; delegates to internal/hook
├── hook_apply_rc.go    # Hidden `hook apply-rc`; called by shell wrappers
├── init.go             # Initialize user list and legacy XDG migration
├── list.go             # List saved identities
├── remove.go           # Remove identity by index/email/name
├── replace.go          # git-filter-repo author replacement flow
├── use.go              # Apply selected identity repo/global
└── *_test.go           # Command-local tests and shared mocks/helpers
```

## WHERE TO LOOK
| Task | Location | Notes |
|------|----------|-------|
| Add subcommand | `root.go` | Append to `cmd.AddCommand(...)`; update i18n keys |
| Add command flags | Command file | Prefer long + short when existing UX has one |
| Change hook CLI UX | `hook.go`, `hook_apply_rc.go` | `installFunc`/`uninstallFunc` are test seams |
| Change selection flags | `use.go`, `remove.go`, `replace.go` | `buildFilter(cmd)` exists in `use.go` only |
| Change init/migration | `init.go`, `init_test.go` | Many locale and legacy-path cases live here |
| Test command output | `*_test.go` | Use `executeCmd` from `current_test.go` |
| Mock persistence | `list_test.go` | `mockStore` is the standard fake `domain.UserStore` |

## CONVENTIONS
- Command constructors are `NewXxxCmd(store domain.UserStore) *cobra.Command`; commands without persistence omit the store.
- Root command accepts `name string` so the same binary logic supports `gitusr` and `gu` aliases.
- `RunE` returns errors; do not print and return the same error.
- Use `errors.New(i18n.T(...))` for translated validation failures.
- i18n-specific tests use `_En` / `_ZhCN` suffixes and reset locale with `i18n.ResetForTesting()`.
- Function vars used for mocking must be restored with `t.Cleanup()`.
- Hidden `hook apply-rc` must stay hidden; shell wrappers call it directly.

## ANTI-PATTERNS
- Do **not** add standalone CLI logic outside this package.
- Do **not** call `os.Exit` from command code.
- Do **not** touch real git config, HOME, or XDG paths in tests; mock functions or use temp dirs.
- Do **not** forget to update both `active.en.toml` and `active.zh-CN.toml` when adding message IDs.

## COMMANDS
```bash
# Run only CLI package tests
go test ./internal/cli/... -v

# Run a specific command test
go test ./internal/cli/... -run TestHook -v
```
