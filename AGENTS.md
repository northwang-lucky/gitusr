# PROJECT KNOWLEDGE BASE

**Generated:** 2026-06-01
**Commit:** 75070aa
**Branch:** main

## OVERVIEW
`gitusr` is a Go 1.26.3 Cobra CLI for managing and switching Git user identities. It stores identities under XDG data, applies them through `git config`, matches clone URLs against host rules, rewrites mistaken author history with `git-filter-repo`, and installs bash/zsh shell hooks for clone/commit/cd workflows.

## STRUCTURE
```
.
├── cmd/gitusr/          # Thin main: i18n init, JSONStore wiring, Cobra execute
├── internal/
│   ├── cli/             # Cobra commands, i18n help templates, command tests
│   ├── domain/          # User, UserFilter, HostRule + store interfaces
│   ├── format/          # User-facing stdout/stderr formatting helpers
│   ├── gitcmd/          # `git config`, backup branch, git-filter-repo wrappers
│   ├── hook/            # Shell wrapper install/uninstall + .gitusrrc application
│   ├── hosts/           # Clone URL parsing and host rule matching logic
│   ├── i18n/            # Embedded active.*.toml translations and locale state; see child AGENTS
│   ├── prompt/          # survey/v2 prompts for user creation/selection
│   ├── select/          # User resolution: index > email > name > interactive
│   ├── store/           # JSON persistence for saved users and host rules
│   ├── version/         # Build-time Version ldflag target
│   └── xdgpath/         # XDG data path resolution and legacy path helpers
├── scripts/             # Manual sandbox/E2E shell scripts
├── test/integration/    # Real binary + real git repo workflow tests
├── .agents/skills/      # Project-local OpenCode/agent skills shared with this repo
├── dist/                # Release artifacts; do not treat as source
├── mise.toml            # Build/test/install tasks
├── .goreleaser.yaml     # Release packaging and Homebrew Formula config
└── release-please-config.json # Release Please changelog/version config
```

## WHERE TO LOOK
| Task | Location | Notes |
|------|----------|-------|
| Add or register a CLI command | `internal/cli/` | Register in `root.go`; command constructors are `NewXxxCmd(...)` |
| Change hook install/cd/commit behavior | `internal/hook/`, `internal/cli/hook*.go` | Shell snippets are generated Go raw strings; CLI has hidden `hooks apply-rc` and `hooks apply-host` |
| Change host routing rules | `internal/cli/hosts.go`, `internal/hosts/`, `internal/store/hosts.go` | CLI group `hosts`; matching logic is pure functions; rules live in `hosts.json` |
| Change user data model | `internal/domain/user.go` | Update `UserStore`, `store.JSONStore`, tests, docs |
| Change persistence path/legacy migration | `internal/xdgpath/`, `internal/store/`, `internal/cli/init.go` | User list and hook state share the XDG gitusr directory |
| Change git interaction | `internal/gitcmd/runner.go` | Wraps real `git`; replace flow depends on `git-filter-repo` |
| Change prompts / UX | `internal/prompt/prompt.go`, `internal/i18n/*.toml` | survey prompts + translated messages must stay aligned |
| Change formatting | `internal/format/format.go` | Keep stdout success messages separate from stderr errors |
| Change translations | `internal/i18n/`, `internal/i18n/AGENTS.md` | Keep `active.en.toml` and `active.zh-CN.toml` message IDs aligned |
| Add unit tests | Package-local `*_test.go` | Most tests override package-level function vars with `t.Cleanup()` |
| Add full workflow tests | `test/integration/` | Builds binary once, uses temp HOME/XDG and real git repos |
| Release packaging | `.goreleaser.yaml`, `release-please-config.json` | GoReleaser injects `internal/version.Version`; Homebrew tap token comes from env |
| Publish binary release | `.agents/skills/publish-binary/SKILL.md` | Project-local SOP for branch checks, PR merge, Release Please, GoReleaser, and GitHub Actions follow-up |

## CODE MAP
| Symbol | Type | Location | Role |
|--------|------|----------|------|
| `main` | func | `cmd/gitusr/main.go` | Initializes path/i18n/store, derives binary alias, executes root command |
| `NewRootCmd` | func | `internal/cli/root.go` | Registers add/current/hook/hosts/init/list/remove/replace/use and custom i18n usage template |
| `domain.UserStore` | interface | `internal/domain/user.go` | Persistence seam used by CLI, select, hooks apply-rc |
| `domain.HostRuleStore` | interface | `internal/domain/hostrule.go` | Persistence seam for the ordered host rule list |
| `store.JSONStore` | struct | `internal/store/store.go` | JSON file implementation; creates empty list on first `List()` |
| `store.JSONHostRuleStore` | struct | `internal/store/hosts.go` | Reads/writes `hosts.json`; missing file means empty rules |
| `hosts.MatchHost` | func | `internal/hosts/match.go` | Parses clone URL, matches rules: exact beats wildcard, first-rule-wins |
| `hosts.ValidateHost` | func | `internal/hosts/validate.go` | Accepts bare hostnames and `*.suffix` wildcards; rejects URLs/ports |
| `sel.ResolveUser` | func | `internal/select/resolver.go` | Resolution priority: index, email, name, then interactive selection |
| `gitcmd.SetConfig` | func | `internal/gitcmd/runner.go` | Calls `git config [--global] user.name/user.email` |
| `gitcmd.FilterRepo` | func | `internal/gitcmd/runner.go` | Runs `git-filter-repo` with inline Python author callback |
| `hook.InstallAll` | func | `internal/hook/installer.go` | Writes unified bash/zsh wrappers for clone/commit/cd and saves hook state |
| `hook.Uninstall` | func | `internal/hook/uninstaller.go` | Removes hook type from state; cleans wrappers only when no hook types remain |
| `hook.ParseRC` | func | `internal/hook/rc.go` | Reads `.gitusrrc`; nil means absent, error means invalid JSON/empty config |
| `hook.MatchAndApplyRC` | func | `internal/hook/rc.go` | Matches saved user by email then name, applies repo-local git config |
| `cli.NewHostsCmd` | func | `internal/cli/hosts.go` | `hosts set/list/remove/move`; rule order is the matching priority |
| `cli.NewHooksApplyHostCmd` | func | `internal/cli/hooks_apply_host.go` | Hidden bridge: matches host rule, applies user or warns when missing |
| `i18n.T` | func | `internal/i18n/bundle.go` | Fail-open message lookup; returns msgID if uninitialized/missing |
| `format.PrintErr` | func | `internal/format/format.go` | Central stderr path used by `main.go` |
| `format.PrintWarn` | func | `internal/format/format.go` | Non-fatal stderr warning path (e.g. stale host rule) |

## CONVENTIONS
- Exported symbols have doc comments; package comments exist where package purpose is non-obvious.
- Errors are wrapped with `fmt.Errorf("context: %w", err)` unless constructing a translated/user-facing sentinel with `errors.New(i18n.T(...))`.
- Commands depend on `domain.UserStore`, not `store.JSONStore`.
- `internal/select` is imported as `sel` in CLI packages.
- CLI `RunE` returns errors; `main.go` is responsible for final `format.PrintErr` and `os.Exit(1)`.
- Success output may use `fmt.Println` in command code or `format.PrintSuccess/PrintUserInfo`; errors must not be printed and returned from the same `RunE`.
- Locale detection priority: `GITUSR_LANG` > `LANGUAGE` > `LANG` > `en`; `zh*` normalizes to `zh-CN`, everything else to `en`.
- Test seams are package-level vars (`askNewUser`, `SelectFunc`, `getConfigFn`, `installFunc`, `uninstallFunc`, etc.) restored with `t.Cleanup()`.
- Tests that write HOME/XDG/shell rc state use `t.TempDir()` and `t.Setenv()`; do not touch real user files.
- Release config is split: Release Please controls changelog sections, GoReleaser controls binary archives and Homebrew Formula output.

## ANTI-PATTERNS (THIS PROJECT)
- Do **not** call `os.Exit` inside command `RunE`.
- Do **not** print an error inside `RunE` and then return it; that double-prints via `main.go`.
- Do **not** use `fmt.Errorf("%s", i18n.T(...))`; use `errors.New(i18n.T(...))`.
- Do **not** add new validation duplicates unless preserving existing user-facing behavior; `store.Add` already checks duplicate email.
- Do **not** duplicate `emailRegex`; it already exists in both `prompt/prompt.go` and `cli/add.go` as legacy overlap.
- Do **not** edit generated/release artifacts under `dist/` as source changes.
- Do **not** run file-mutating manual tests against real HOME or XDG directories; sandbox or temp env only.

## UNIQUE STYLES
- Root command calls `InitDefaultCompletionCmd/HelpCmd/HelpFlag` and then rewrites built-in help/completion descriptions through i18n.
- The executable name is `filepath.Base(os.Args[0])`, so `gitusr` and `gu` share the same root command with different `Use`.
- Hook state lives next to `user-list.json` as `hook-state.json`; shell wrappers live in `{XDG_DATA_HOME}/gitusr/hooks/`.
- `hooks install` calls `InstallAll`; nil results mean all hook types are already installed and CLI prints idempotent success.
- `.gitusrrc` matching uses email priority over name; `hooks apply-rc --silent-if-unchanged` avoids repeated output from shell hooks.
- Clone identity resolution in the wrappers: explicit `--gu-*` args > repo `.gitusrrc` > host rules (`hooks apply-host`) > interactive `gitusr use`. Host rules apply only when `hosts.json` exists.
- Host rules live in `hosts.json` as an ordered array; `hosts set` updates in place, only `hosts move` changes priority.
- `LoadState()` treats missing, empty, and invalid hook-state JSON as an empty state.

## COMMANDS
```bash
# Build binary to bin/gitusr with version ldflag
mise run build

# Run all Go tests (unit + integration)
mise run test

# Install to ~/.local/bin/gitusr and symlink gu; backs up ~/.gitconfig first
mise run install

# Remove local binary/symlink/data files
mise run uninstall

# Clean build artifacts
mise run clean

# Release packaging config is GoReleaser + Release Please owned
.goreleaser.yaml
release-please-config.json
```

## NOTES
- Data file path: `$XDG_DATA_HOME/gitusr/user-list.json`, falling back to `~/.local/share/gitusr/user-list.json`.
- Host rules live at `$XDG_DATA_HOME/gitusr/hosts.json`; matching is exact-first then wildcard, both in configuration order.
- Legacy data path is `~/.gitusr/user-list.json`; migration behavior is in init flow and xdgpath helpers.
- `replace` requires `git-filter-repo` available in `PATH` and creates a backup branch before rewriting history.
- Integration tests require `git`; hook tests also exercise shell rc file writes inside temp HOME.
- No GitHub Actions or golangci config is present; `mise run test` is the project quality gate.
- Existing focused child docs: `internal/cli/AGENTS.md`, `internal/hook/AGENTS.md`, `internal/i18n/AGENTS.md`, `test/integration/AGENTS.md`.
