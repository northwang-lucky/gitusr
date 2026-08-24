# HOOK PACKAGE

**Scope:** `internal/hook/`

## OVERVIEW
Shell-hook engine for automatic git user switching. It writes unified bash/zsh wrapper files, mutates shell rc files with marked blocks, tracks installed/disabled hook types, applies `.gitusrrc` to repo-local git config, and routes clone URLs to saved users via host rules (`hooks apply-host`).

## WHERE TO LOOK
| Task | Location | Notes |
|------|----------|-------|
| Add/remove hook type | `types.go`, `installer.go`, `uninstaller.go` | Update `AllHookTypes`, CLI validation, bash/zsh generation, tests |
| Change wrapper behavior | `shell_bash.go`, `shell_zsh.go` | Raw shell strings for git clone/commit wrappers |
| Change clone identity chain | `shell_bash.go`, `shell_zsh.go` | Priority: `--gu-*` args > `.gitusrrc` > `hosts.json` (`apply-host`) > interactive `use` |
| Change cd auto-apply behavior | `shell_bash.go`, `shell_zsh.go` | Bash aliases `cd`; zsh uses `add-zsh-hook chpwd` |
| Change shell rc mutation | `config_writer.go` | Markers are `# gitusr hook begin/end` |
| Change persistent hook state | `state.go` | State file is `hook-state.json` next to user list; tracks installed and disabled types |
| Change `.gitusrrc` parsing/matching | `rc.go` | Email match wins over name match |
| Add tests for filesystem writes | `*_test.go` | Use `t.TempDir()` + `t.Setenv("HOME"/"XDG_DATA_HOME")` |

## CODE MAP
| Symbol | Type | Location | Role |
|--------|------|----------|------|
| `HookTypeClone/Commit/CD` | const | `types.go` | Supported hook types and CLI `--type` values |
| `InstallAll` | func | `installer.go` | Idempotent install for clone/commit/cd; writes unified wrappers and hook state |
| `UninstallAll` | func | `uninstaller.go` | Removes unified wrappers, rc blocks, and hook state |
| `WriteWrapperFile` | func | `config_writer.go` | Writes `{XDG_DATA_HOME}/gitusr/hooks/git-wrapper.{sh,zsh}` |
| `AppendSourceLine` | func | `config_writer.go` | Replaces existing marked block before appending source line |
| `RemoveSourceBlock` | func | `config_writer.go` | No-op if shell config or marked block is absent |
| `LoadState` | func | `state.go` | Missing/empty/invalid JSON becomes empty installed list |
| `ParseRC` | func | `rc.go` | Absent `.gitusrrc` returns nil; invalid/empty config returns error |
| `MatchAndApplyRC` | func | `rc.go` | Finds saved user then calls repo-local `gitcmd.SetConfig` |

## CONVENTIONS
- Install idempotency contract: `InstallAll(...)` returns `nil, nil` when all hook types are already installed.
- `Install` is legacy/deprecated; new CLI flow should use unified `InstallAll` and `UninstallAll`.
- Clone snippets call `gitusr list` first and pass through when saved user count is `<= 1`.
- Shell snippets check `gitusr hooks is-disabled <type>` before clone/commit/cd behavior.
- Clone identity chain: `--gu-*` args > `.gitusrrc` > host rules (`apply-host`, only when `hosts.json` exists) > interactive `gitusr use`.
- The host rule branch reads the remote URL via `git config --local --get-regexp '^remote\..*\.url$'`; the clone URL is deliberately NOT parsed from argv (branch names look like URLs).
- Bash wrappers use `command git` and `\cd` to avoid function/alias recursion.
- Zsh cd hooks remove the existing hook with `add-zsh-hook -D` before re-adding it.
- Shell rc files are mutated only inside the marked block; existing config must be preserved.
- Tests assert marker counts and wrapper file contents rather than sourcing shell files.
- `HookState` stores both installed and disabled hook type lists; enable/disable must not remove installation state.

## ANTI-PATTERNS
- Do **not** write tests that touch the developer's real `.bashrc`, `.zshrc`, HOME, or XDG data.
- Do **not** change marker strings without updating every install/uninstall test and migration expectation.
- Do **not** delete wrapper files while any hook type remains installed.
- Do **not** make `.gitusrrc` name take priority over email.
- Do **not** add unsupported shells by only changing generators; update CLI defaults and rc-path handling too.
- Do **not** break hidden CLI bridge calls: wrappers depend on `hooks apply-rc --silent-if-unchanged`, `hooks apply-host --silent-if-unchanged`, and `hooks is-disabled`.
- Do **not** reorder the clone identity chain: `--gu-*` args and `.gitusrrc` both override host rules by design.

## COMMANDS
```bash
# Run hook package tests
go test ./internal/hook/... -v

# Run hook CLI tests too
go test ./internal/hook/... ./internal/cli/... -run Hook -v
```
