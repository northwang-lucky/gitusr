# INTEGRATION TESTS

**Scope:** `test/integration/`

## OVERVIEW
End-to-end tests build the real `gitusr` binary once and exercise CLI workflows against temporary HOME/XDG directories and real git repositories.

## STRUCTURE
```
test/integration/
├── integration_test.go       # Binary build, helpers, core workflows
└── hook_test.go              # Hook/.gitusrrc workflows
```

## WHERE TO LOOK
| Task | Location | Notes |
|------|----------|-------|
| Add a full CLI scenario | `integration_test.go` | Reuse binary helpers; isolate HOME/XDG |
| Change binary build flags | `TestMain` | Builds `./cmd/gitusr` to `os.TempDir()` |
| Change env isolation | `runGitusr`, `runGitusrInDir` | Pass env map; never rely on real HOME |
| Assert git config | `gitConfig` helper | Reads raw `git config` in temp repo |
| Add hook E2E coverage | `hook_test.go` | Must sandbox shell rc writes with temp HOME/XDG |

## CONVENTIONS
- `TestMain` owns binary build and cleanup; individual tests call the built binary.
- Helpers run real `git` for repo setup and assertions.
- Each test uses `t.TempDir()` for HOME and XDG paths.
- Keep tests sequential; shared binary path and external git operations make parallelism risky.
- Prefer non-interactive CLI flags in E2E tests.

## ANTI-PATTERNS
- Do **not** call `t.Parallel()` here.
- Do **not** read or write the developer's real `~/.gitconfig`, shell rc files, or XDG data.
- Do **not** assert translated prose unless the test explicitly fixes locale.

## COMMANDS
```bash
# Run integration tests only
go test ./test/integration/... -v

# Run full project test suite
mise run test
```

## NOTES
- Requires `git` in `PATH`; replace/history scenarios also require `git-filter-repo` when covered.
- `findModuleRoot()` walks upward to locate `go.mod` for the binary build.
