---
name: publish-binary
description: 本项目 gitusr 专用的二进制发布流程技能。凡是用户要求发布 gitusr、发布二进制、发版、创建 release/PR、触发 GitHub Actions/GoReleaser/Homebrew 发布，或从当前分支合入 main 并跟进发布结果时，必须使用本 skill。它只适用于当前仓库，不适用于通用项目发布。
compatibility: opencode
metadata:
  project: gitusr
  workflow: github-release-please-goreleaser
---

# publish-binary

本 skill 负责把当前 `gitusr` 仓库的分支变更安全地发布为 GitHub Release、GoReleaser 二进制归档和 Homebrew tap 更新。流程的核心原则是：不要在未确认状态下改写历史，不要直接推送 main，不要替用户跳过 code review，除非用户明确选择直接通过。

## 项目事实

- 发布仓库：`northwang-lucky/gitusr`。
- 默认分支：`main`。
- Release Please 工作流：`.github/workflows/release-please.yml`，在 `main` push 后运行并创建/更新 Release PR。
- Release 工作流：`.github/workflows/release.yml`，在 `Release Please` 完成或 `v*` tag push 后运行 GoReleaser。
- GoReleaser 配置：`.goreleaser.yaml`，构建 `darwin/linux` 的 `amd64/arm64` 二进制，并更新 `northwang-lucky/homebrew-tap`。
- Homebrew tap 需要仓库 secret：`HOMEBREW_TAP_GITHUB_TOKEN`。
- 本项目质量门禁：`mise run test`。

## 开始前

1. 如果可用，先加载 `git-master`，因为本流程包含分支、PR、合并和提交历史检查。
2. 使用 `git remote -v` 确认存在 GitHub 远端；如果没有远端或不是 GitHub，提示用户配置 GitHub 远端后结束。
3. 使用 `gh auth status` 确认 GitHub CLI 已登录；如果未登录，提示用户登录后结束。
4. 使用 `git status --short --branch` 查看当前分支和工作区。
5. 若工作区存在未提交变更：
   - 如果这些变更是当前任务产生的，先按项目规则提交。
   - 如果是用户或其他 agent 的变更，不要改动；提示用户处理或确认纳入发布后再继续。

## 1. 检查当前分支

先更新远端引用：

```bash
git fetch --prune origin
```

读取当前分支：

```bash
git branch --show-current
```

### 当前分支是 main

1. 比较本地 `main` 和 `origin/main`：

   ```bash
   git status --short --branch
   git log --oneline origin/main..HEAD
   git log --oneline HEAD..origin/main
   ```

2. 如果 `main` 相对 `origin/main` 没有领先提交，且工作区干净：报告“main 相对远端无待发布变更”，结束流程。
3. 如果 `main` 落后于 `origin/main`：先 `git pull --ff-only origin main` 更新，再重新检查。
4. 如果 `main` 相对 `origin/main` 有领先提交：创建发布分支，不要直接从 main 发 PR：

   ```bash
   git switch -c release/<short-summary>
   ```

   `<short-summary>` 用变更主题命名，例如 `release/hook-fixes`、`release/publish-binary-skill`。

### 当前分支不是 main

1. 确认分支相对远端 main 的状态：

   ```bash
   git rev-list --left-right --count origin/main...HEAD
   git merge-base --is-ancestor origin/main HEAD
   ```

2. 如果当前分支领先于 `origin/main` 且包含 `origin/main` 的最新提交，继续流程。
3. 如果当前分支落后于 `origin/main`，或 `origin/main` 不是当前分支祖先：提示用户先 rebase/merge main 并解决冲突，然后结束流程。
4. 如果有冲突状态或未完成 rebase/merge：提示用户解决冲突后再运行本流程，结束流程。

## 2. 分析分支变更并建议版本号类型

收集变更：

```bash
git diff --stat origin/main...HEAD
git log --oneline --decorate origin/main..HEAD
git diff --name-only origin/main...HEAD
```

按 Conventional Commits 和实际影响建议 bump 类型：

| 类型 | 选择条件 |
|------|----------|
| `major` | 有 breaking change、CLI 参数/输出/配置不兼容、数据格式不兼容、用户必须改迁移方式 |
| `minor` | 新增用户可见功能、命令、flag、hook 行为、发布能力，但保持兼容 |
| `patch` | bug fix、文案、测试、CI、文档、内部重构且不改变用户兼容行为 |

报告格式：

```text
建议版本变更：<major|minor|patch>
依据：
- <依据 1>
- <依据 2>

请选择：
1. 接受建议
2. 改为 major
3. 改为 minor
4. 改为 patch
```

等待用户确认。若用户明确授权“你自行决定”，使用上述规则自行选择并继续。

> 注意：本仓库使用 Release Please，通常不需要手工修改版本号或 tag。版本 bump 主要通过提交信息和 Release Please 配置体现。若现有提交信息无法表达目标 bump，应先和用户确认是否补充一个符合 Conventional Commits 的提交，而不是篡改历史。

## 3. 创建合入 main 的 PR

推送当前分支：

```bash
git push -u origin HEAD
```

创建 PR：

```bash
gh pr create --base main --head "$(git branch --show-current)" --fill
```

如果 `--fill` 生成的标题/正文不能体现版本建议，改用显式标题和正文：

```bash
gh pr create --base main --head "$(git branch --show-current)" --title "<type(scope): summary>" --body "<summary and release impact>"
```

创建后：

1. 把 PR URL 告诉用户。
2. 询问用户是否要等待 code review，还是直接通过并合并。
3. 若用户选择等待 code review，停止自动合并，只保留 PR URL 和下一步说明。
4. 若用户选择直接通过，先检查 PR 状态和检查项：

   ```bash
   gh pr view --json url,mergeable,reviewDecision,statusCheckRollup
   gh pr checks --watch
   ```

5. 检查通过后合并。优先使用 squash，保持 main 历史清晰：

   ```bash
   gh pr merge --squash --delete-branch
   ```

   如果仓库规则不允许 squash，根据 `gh pr view` 和错误信息选择仓库允许的合并方式；不要使用强制合并或绕过保护规则。

## 4. PR 合并后更新 main

合并完成后：

```bash
git switch main
git pull --ff-only origin main
```

确认本地 main 与远端一致：

```bash
git status --short --branch
git log --oneline -5
```

## 5. 跟进 GitHub Actions 发布流程

先跟进 Release Please：

```bash
gh run list --workflow "Release Please" --branch main --limit 5
gh run watch <run-id> --exit-status
```

Release Please 成功后有两种结果：

1. 创建或更新 Release PR：提示用户需要合并 Release Please PR 才会创建 tag/release。
2. 已创建 release/tag：继续跟进 Release 工作流。

查找 Release Please PR：

```bash
gh pr list --search "Release Please" --state open --base main
```

如果出现 Release Please PR：

1. 打开 PR 摘要给用户确认 changelog 和版本号。
2. 用户确认后，按第 3 步的检查与合并规则合并该 PR。
3. 合并后再次 `git switch main && git pull --ff-only origin main`。

跟进 Release 工作流：

```bash
gh run list --workflow "Release" --limit 10
gh run watch <run-id> --exit-status
```

如果失败，收集失败原因：

```bash
gh run view <run-id> --json conclusion,status,event,headBranch,headSha,url
gh run view <run-id> --log-failed
```

分析时优先检查：

- `HOMEBREW_TAP_GITHUB_TOKEN` 是否缺失或权限不足。
- Go 版本是否与 `go.mod` / `mise.toml` 不一致。
- GoReleaser 配置或 Homebrew tap 路径是否错误。
- Release Please 是否只创建了 PR、尚未创建 tag。
- GitHub token 权限是否不足以写 release contents。

失败时不要盲目重跑。先向用户报告：

```text
发布失败：<workflow/job>
原因判断：<root cause>
证据：<关键日志行或 gh 输出摘要>
建议修复：<具体步骤>
```

发布成功时报告：

```text
发布成功：<version/tag>
GitHub Release：<url>
Homebrew tap：northwang-lucky/homebrew-tap
```

## 禁止事项

- 不要主动 `git push --force`。
- 不要直接推送 main。
- 不要在未得到用户确认时合并业务 PR 或 Release Please PR。
- 不要跳过失败的检查项。
- 不要手工创建 tag，除非用户明确要求并确认 Release Please 不适用。
- 不要修改 `.goreleaser.yaml`、Release Please 配置或 workflow 来“绕过”发布失败；先报告原因和修复方案。
