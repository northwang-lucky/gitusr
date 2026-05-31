# INTEGRATION TESTS

**Scope:** `test/integration/`

## OVERVIEW
End-to-end tests that build the real `gitusr` binary and exercise it against temporary git repositories. These tests verify the full CLI workflow across commands.

## STRUCTURE
```
test/integration/
└── integration_test.go   # 559 lines, 9 test functions
```

## WHERE TO LOOK
| Task | Location | Notes |
|------|----------|-------|
| Add a new E2E scenario | `integration_test.go` | Append to `TestFullWorkflow` or add a new `Test*` function |
| Change binary build flags | `TestMain` (line 17) | `go build -o <tmp> ./cmd/gitusr` |
| Change env isolation | `runGitusr` / `runGitusrInDir` | Each test uses independent `t.TempDir()` for `HOME` |

## CONVENTIONS
- `TestMain` builds the binary once before all tests and cleans it up after.
- `runGitusr(t, env, args...)` executes the built binary with optional env overrides.
- `runGitusrInDir(t, env, dir, args...)` runs the binary in a specific working directory.
- `runGit(t, dir, env, args...)` executes raw `git` commands for setup/assertions.
- `gitConfig(t, dir, key)` reads git config values for assertions.
- Each test creates isolated `HOME` and `XDG_DATA_HOME` via `t.TempDir()` to prevent cross-test pollution.

## ANTI-PATTERNS
- Do **not** call `t.Parallel()` — the shared `gitusrBin` global and real git repos make parallel execution unsafe.
- Do **not** rely on the developer's real `~/.gitconfig` — always set `HOME` to a temp dir.

## COMMANDS
```bash
# Run integration tests only
go test ./test/integration/... -v

# Run the full test suite (unit + integration)
mise run test
```

## NOTES
- Requires `git` to be installed and available in `$PATH`.
- The binary is built to `os.TempDir()` and removed in `TestMain` cleanup.
- `findModuleRoot()` walks up the directory tree to locate `go.mod` for the build command.
