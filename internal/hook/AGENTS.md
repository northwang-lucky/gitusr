# HOOK PACKAGE

**Scope:** `internal/hook/`

## OVERVIEW
Shell-hook engine for automatic git user switching. It writes bash/zsh wrapper files, mutates shell rc files with marked blocks, tracks installed hook types, and applies `.gitusrrc` to repo-local git config.

## WHERE TO LOOK
| Task | Location | Notes |
|------|----------|-------|
| Add/remove hook type | `types.go`, `installer.go`, `uninstaller.go` | Update `AllHookTypes`, CLI validation, tests |
| Change wrapper behavior | `shell_bash.go`, `shell_zsh.go` | Raw shell strings for git clone/commit wrappers |
| Change cd auto-apply behavior | `env_bash.go`, `env_zsh.go` | Bash aliases `cd`; zsh uses `add-zsh-hook chpwd` |
| Change shell rc mutation | `config_writer.go` | Markers are `# gitusr hook begin/end` |
| Change persistent hook state | `state.go` | State file is `hook-state.json` next to user list |
| Change `.gitusrrc` parsing/matching | `rc.go`, `env.go` | Email match wins over name match |
| Add tests for filesystem writes | `*_test.go` | Use `t.TempDir()` + `t.Setenv("HOME"/"XDG_DATA_HOME")` |

## CODE MAP
| Symbol | Type | Location | Role |
|--------|------|----------|------|
| `HookTypeClone/Commit/CD` | const | `types.go` | Supported hook types and CLI `--type` values |
| `Install` | func | `installer.go` | Idempotent install; writes wrapper, rc block, hook state |
| `Uninstall` | func | `uninstaller.go` | Removes hook from state; deletes rc block/wrappers only when no types remain |
| `WriteWrapperFile` | func | `config_writer.go` | Writes `{XDG_DATA_HOME}/gitusr/hooks/git-wrapper.{sh,zsh}` |
| `AppendSourceLine` | func | `config_writer.go` | Replaces existing marked block before appending source line |
| `RemoveSourceBlock` | func | `config_writer.go` | No-op if shell config or marked block is absent |
| `LoadState` | func | `state.go` | Missing/empty/invalid JSON becomes empty installed list |
| `ParseRC` | func | `rc.go` | Absent `.gitusrrc` returns nil; invalid/empty config returns error |
| `MatchAndApplyRC` | func | `rc.go` | Finds saved user then calls repo-local `gitcmd.SetConfig` |

## CONVENTIONS
- Install idempotency contract: `Install(...)` returns `nil, nil` when the hook type is already installed.
- `HookTypeCD` uses `cdShellGenerators`; clone/commit use `shellGenerators`.
- Shell snippets call `gitusr list` first and pass through when saved user count is `<= 1`.
- Bash wrappers use `command git` and `\cd` to avoid function/alias recursion.
- Zsh cd hooks remove the existing hook with `add-zsh-hook -D` before re-adding it.
- Shell rc files are mutated only inside the marked block; existing config must be preserved.
- Tests assert marker counts and wrapper file contents rather than sourcing shell files.

## ANTI-PATTERNS
- Do **not** write tests that touch the developer's real `.bashrc`, `.zshrc`, HOME, or XDG data.
- Do **not** change marker strings without updating every install/uninstall test and migration expectation.
- Do **not** delete wrapper files while any hook type remains installed.
- Do **not** make `.gitusrrc` name take priority over email.
- Do **not** add unsupported shells by only changing generators; update CLI defaults and rc-path handling too.

## COMMANDS
```bash
# Run hook package tests
go test ./internal/hook/... -v

# Run hook CLI tests too
go test ./internal/hook/... ./internal/cli/... -run Hook -v
```
