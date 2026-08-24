# CLI COMMANDS

**Scope:** `internal/cli/`

## OVERVIEW
Cobra command layer for gitusr. It owns flag parsing, i18n help text, user-facing command errors, hidden hook bridge commands, and package-level test seams around prompts/git/hook operations.

## STRUCTURE
```
internal/cli/
├── root.go              # Root command registration + custom i18n usage template
├── add.go              # Add saved identity; non-interactive flags or prompt
├── current.go          # Show repo/global git config
├── hosts.go            # hosts set/list/remove/move — ordered host rule routing
├── hooks_apply_host.go # Hidden hooks apply-host bridge for the clone wrapper
├── init.go             # Initialize user list and legacy XDG migration
├── list.go             # List saved identities
├── remove.go           # Remove identity by index/email/name
├── replace.go          # git-filter-repo author replacement flow
├── use.go              # Apply selected identity repo/global
├── hooks*.go           # Hook install/uninstall/enable/disable/apply-rc bridge
└── *_test.go           # Command-local tests and shared mocks/helpers
```

## WHERE TO LOOK
| Task | Location | Notes |
|------|----------|-------|
| Add subcommand | `root.go` | Append to `cmd.AddCommand(...)`; update i18n keys |
| Add command flags | Command file | Prefer long + short when existing UX has one |
| Change hook CLI UX | `hooks.go`, `hooks_apply_rc.go` | `installAllFunc`/`uninstallFunc` are test seams; hidden commands are shell-facing API |
| Change host rule CLI | `hosts.go` | Rule order is priority; `hosts set` updates in place, only `move` reorders |
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
- Hidden `hooks apply-rc`, `hooks apply-host`, and `hooks is-disabled` must stay hidden; shell wrappers call them directly.

## ANTI-PATTERNS
- Do **not** add standalone CLI logic outside this package.
- Do **not** call `os.Exit` from command code.
- Do **not** touch real git config, HOME, or XDG paths in tests; mock functions or use temp dirs.
- Do **not** forget to update both `active.en.toml` and `active.zh-CN.toml` when adding message IDs.
- Do **not** change hook command names casually; generated bash/zsh wrappers invoke exact strings.
- Do **not** duplicate host-rule validation in commands; `hosts.ValidateHost` is the single source.
- Do **not** make `hosts set` reorder rules; only `hosts move` changes priority. `resolveMoveTarget` owns all move-target semantics.

## COMMANDS
```bash
# Run only CLI package tests
go test ./internal/cli/... -v

# Run a specific command test
go test ./internal/cli/... -run TestHook -v
```
