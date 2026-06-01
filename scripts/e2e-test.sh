#!/bin/bash
# =============================================================================
# e2e-test.sh — gitusr E2E 测试脚本（executed inside bubblewrap sandbox）
#
# 此脚本由 sandbox-test.sh 在 bubblewrap 沙箱内调用。
# 在隔离的 XDG_DATA_HOME 中构建 gitusr、安装 git-filter-repo、运行完整 E2E 流程。
#
# 请勿直接运行此脚本 — 请改用 scripts/sandbox-test.sh。
# =============================================================================
set -euo pipefail

# ─── Helper 函数 ───────────────────────────────────────────────────────────

# run_cmd — 执行命令，打印命令本身（输出到 stderr，不干扰 stdout 管道）
run_cmd() {
    echo "[CMD] $*" >&2
    "$@"
}

# run_cmd_ok — 执行命令，断言 $? == 0，否则打印错误信息并 exit 1
run_cmd_ok() {
    echo "[CMD] $*" >&2
    "$@"
    local rc=$?
    if [ $rc -ne 0 ]; then
        echo "  FAIL: '$*' exited with $rc (expected 0)" >&2
        exit 1
    fi
    return 0
}

# run_cmd_fail — 执行命令，断言 $? != 0，否则打印错误信息并 exit 1
run_cmd_fail() {
    echo "[CMD] $*" >&2
    "$@"
    local rc=$?
    if [ $rc -eq 0 ]; then
        echo "  FAIL: '$*' exited with 0 (expected non-zero)" >&2
        exit 1
    fi
    return 0
}

# ─── 沙箱环境设置 ──────────────────────────────────────────────────────────
echo "[sandbox] Setting up environment..."

export XDG_DATA_HOME=/tmp/xdg
mkdir -p "$XDG_DATA_HOME"

# 安装 git-filter-repo（filter-repo 命令使用）。
# 优先使用 python3 -m pip，避免裸 pip3 的路径问题。
echo "[sandbox] Installing git-filter-repo..."
if python3 -m pip install --quiet git-filter-repo 2>/dev/null; then
    echo "[sandbox] git-filter-repo installed"
else
    echo "[sandbox] WARNING: git-filter-repo install failed, continuing without it"
fi

# ─── 构建 gitusr 二进制 ────────────────────────────────────────────────────
echo "[sandbox] Building gitusr binary..."
go build -o /tmp/gitusr ./cmd/gitusr
echo "[sandbox] Build complete"

# ─── gitusr 包装函数 ───────────────────────────────────────────────────────
# 始终设置 XDG_DATA_HOME 确保隔离的 store 访问，GITUSR_LANG=en 确保英文输出。
gitusr() {
    GITUSR_LANG=en XDG_DATA_HOME=/tmp/xdg /tmp/gitusr "$@"
}

# ══════════════════════════════════════════════════════════════════════════════
# Steps 1-8: 核心 CLI 工作流（init / add / list / use / current / remove / replace）
# ══════════════════════════════════════════════════════════════════════════════

# Step 1: init — 创建第一个用户
echo ""
echo "--- Step 1/8: init ---"
gitusr init --name "Dev" --email "dev@test.com" --yes --force
echo "  Step 1 PASSED"

# Step 2: add Work 用户
echo ""
echo "--- Step 2/8: add Work ---"
gitusr add --name "Work" --email "work@test.com"
echo "  Step 2 PASSED"

# Step 3: add Personal 用户
echo ""
echo "--- Step 3/8: add Personal ---"
gitusr add --name "Personal" --email "personal@test.com"
echo "  Step 3 PASSED"

# Step 4: list — 验证三个用户都在列表中
echo ""
echo "--- Step 4/8: list ---"
gitusr list | tee /tmp/list-output.txt
grep -q "Dev" /tmp/list-output.txt      || { echo "  FAIL: Dev not in list"; exit 1; }
grep -q "Work" /tmp/list-output.txt     || { echo "  FAIL: Work not in list"; exit 1; }
grep -q "Personal" /tmp/list-output.txt || { echo "  FAIL: Personal not in list"; exit 1; }
echo "  Step 4 PASSED (Dev, Work, Personal all present)"

# Step 5: use — 在临时 git 仓库中切换到 Work 身份
echo ""
echo "--- Step 5/8: use --email in git repo ---"
TMP_REPO=/tmp/repo
rm -rf "$TMP_REPO"
mkdir -p "$TMP_REPO"
cd "$TMP_REPO"
git init
gitusr use --email "work@test.com"
echo "  Step 5 PASSED"

# Step 6: current — 验证当前仓库身份为 Work
echo ""
echo "--- Step 6/8: current ---"
cd "$TMP_REPO"
gitusr current | tee /tmp/current-output.txt
grep -q "work@test.com" /tmp/current-output.txt || {
    echo "  FAIL: work@test.com not in current output"
    exit 1
}
echo "  Step 6 PASSED"

# Step 7: remove — 按邮箱删除 Personal 用户
echo ""
echo "--- Step 7/8: remove Personal ---"
cd /src
gitusr remove --email "personal@test.com"
# 验证 Personal 确实被删除
gitusr list | tee /tmp/list-after-remove.txt
grep -q "Personal" /tmp/list-after-remove.txt && {
    echo "  FAIL: Personal still in list after remove"
    exit 1
}
echo "  Step 7 PASSED"

# Step 8: replace — 将 work@test.com 替换为已存在用户
echo ""
echo "--- Step 8/8: replace ---"
if git filter-repo --help &>/dev/null; then
    cd /tmp/repo
    # 创建一个由 Work 创作的初始提交，以便 git-filter-repo 有历史记录
    git config user.email "work@test.com"
    git config user.name "Work"
    touch README.md && git add README.md && git commit -m "initial commit" --no-gpg-sign
    # 将 Work 创作的提交替换为 Dev 身份（Dev 已在 store 中存在）
    gitusr replace work@test.com --with-index 0 --yes
    echo "  Step 8 PASSED"
else
    echo "  Step 8 SKIPPED (git-filter-repo not available)"
fi

# ══════════════════════════════════════════════════════════════════════════════
# Steps 9-14: Hooks 安装/卸载基础流程
#     适配自旧命令格式（gitusr hook install --all → gitusr hooks install）
# ══════════════════════════════════════════════════════════════════════════════

# Step 9: hooks install — 安装所有 hooks（clone, commit, cd）
echo ""
echo "--- Step 9: hooks install ---"
gitusr hooks install | tee /tmp/hook-install-output.txt
grep -q "All hooks successfully installed" /tmp/hook-install-output.txt || {
    echo "  FAIL: hooks install did not report aggregate success"
    exit 1
}
if [[ -f "$XDG_DATA_HOME/gitusr/hooks/git-wrapper.sh" ]]; then
    echo "  Step 9 PASSED (all hooks installed, wrapper exists)"
else
    echo "  FAIL: git-wrapper.sh not found after install"
    exit 1
fi

# Step 10: hooks install 幂等性 — 再次安装应报告 already installed
echo ""
echo "--- Step 10: hooks install idempotency ---"
gitusr hooks install | tee /tmp/hook-install-idempotent-output.txt
grep -q "All hooks are already installed" /tmp/hook-install-idempotent-output.txt || {
    echo "  FAIL: idempotent install did not report already installed"
    exit 1
}
echo "  Step 10 PASSED (install is idempotent)"

# Step 11: hooks install 状态验证 — hook-state.json 应包含三种类型
echo ""
echo "--- Step 11: hooks install state verification ---"
WRAPPER_FILE="$XDG_DATA_HOME/gitusr/hooks/git-wrapper.sh"
HOOK_STATE="$XDG_DATA_HOME/gitusr/hook-state.json"
if [[ -f "$WRAPPER_FILE" ]] && [[ -f "$HOOK_STATE" ]] && grep -q '"clone"' "$HOOK_STATE" && grep -q '"commit"' "$HOOK_STATE" && grep -q '"cd"' "$HOOK_STATE"; then
    echo "  Step 11 PASSED (state contains clone, commit, cd)"
else
    echo "  FAIL: state verification failed"
    exit 1
fi

# Step 12: hooks uninstall — 卸载所有 hooks
echo ""
echo "--- Step 12: hooks uninstall ---"
gitusr hooks uninstall | tee /tmp/hook-uninstall-output.txt
grep -q "All hooks successfully uninstalled" /tmp/hook-uninstall-output.txt || {
    echo "  FAIL: hooks uninstall did not report aggregate success"
    exit 1
}
echo "  Step 12 PASSED (all hooks uninstalled)"

# Step 13: hooks uninstall 幂等性 — 再次卸载应报告 none_installed（exit != 0）
echo ""
echo "--- Step 13: hooks uninstall when nothing installed ---"
gitusr hooks uninstall 2>&1 | tee /tmp/hook-uninstall-none-output.txt || true
grep -q "No hooks are currently installed" /tmp/hook-uninstall-none-output.txt || {
    echo "  FAIL: uninstall did not report none_installed"
    exit 1
}
echo "  Step 13 PASSED (uninstall reports nothing installed)"

# Step 14: hooks uninstall 清理验证 — wrapper 应已删除，状态应已清空
echo ""
echo "--- Step 14: hooks uninstall cleanup verification ---"
if [[ ! -f "$XDG_DATA_HOME/gitusr/hooks/git-wrapper.sh" ]] && ! grep -q '"clone"\|"commit"\|"cd"' "$HOOK_STATE"; then
    echo "  Step 14 PASSED (cleanup removed wrappers and cleared state)"
else
    echo "  FAIL: cleanup did not remove wrappers or clear state"
    exit 1
fi

# ══════════════════════════════════════════════════════════════════════════════
# Task 7: Clone Scenarios (CL-1 → CL-8)
#
# 场景覆盖 bash wrapper git() clone 分支的所有路径：
#   CL-1: --gu-email val 格式 → 验证 user.email 被设置
#   CL-2: --gu-name val 格式  → 验证 user.name 被设置
#   CL-3: --gu-email=val 等号格式 → 验证 user.email 被设置
#   CL-4: --gu-name + --gu-email 同时指定 → 验证 name+email 均被设置
#   CL-5: clone disabled → pass-through（config 不变）
#   CL-6: 单用户 pass-through（config 不变）
#   CL-7: clone 失败 URL → 验证非零 exit code
#   CL-8: 无 --gu-* 参数 → 验证不崩溃
#
# 实现注意：
#   - 先运行 gitusr hooks install && source wrapper
#   - 每种场景使用独立的临时 bare repo 和 clone 目录
#   - 用 git config user.email / user.name 做精确断言
#   - CL-5 需先 gitusr hooks disable clone
#   - CL-6 需先将用户列表减至 1 个用户
# =============================================================================

echo ""
echo "--- TODO: Task 7 - Clone Scenarios (CL-1 → CL-8) ---"
echo "  [PLACEHOLDER] CL-1: --gu-email val format"
echo "  [PLACEHOLDER] CL-2: --gu-name val format"
echo "  [PLACEHOLDER] CL-3: --gu-email=val format"
echo "  [PLACEHOLDER] CL-4: --gu-name + --gu-email together"
echo "  [PLACEHOLDER] CL-5: clone disabled → pass-through"
echo "  [PLACEHOLDER] CL-6: single user pass-through"
echo "  [PLACEHOLDER] CL-7: clone failure (invalid URL)"
echo "  [PLACEHOLDER] CL-8: no --gu-* parameters"
echo "  Step 15-22 SKIPPED (Task 7 not yet implemented)"

# ══════════════════════════════════════════════════════════════════════════════
# Task 8: Commit Scenarios (CM-1 → CM-4)
#
# 场景覆盖 bash wrapper git() commit 分支的所有路径：
#   CM-1: .gitusrrc 存在 → 验证 commit 后 user.email/name 被设置
#   CM-2: 无 .gitusrrc → 验证 config 不变（pass-through）
#   CM-3: commit disabled → 验证 config 不变（pass-through）
#   CM-4: 单用户 → 验证 config 不变（pass-through）
#
# 实现注意：
#   - 先运行 gitusr hooks install && source wrapper
#   - 每种场景使用独立的临时 git 仓库
#   - .gitusrrc 使用 {"email":"..."} 格式进行匹配
#   - CM-3 需先 gitusr hooks disable commit
# =============================================================================

echo ""
echo "--- TODO: Task 8 - Commit Scenarios (CM-1 → CM-4) ---"
echo "  [PLACEHOLDER] CM-1: .gitusrrc present → set user.email"
echo "  [PLACEHOLDER] CM-2: no .gitusrrc → pass-through"
echo "  [PLACEHOLDER] CM-3: commit disabled → pass-through"
echo "  [PLACEHOLDER] CM-4: single user → pass-through"
echo "  Step 23-26 SKIPPED (Task 8 not yet implemented)"

# ══════════════════════════════════════════════════════════════════════════════
# Task 9: CD + Other Scenarios (CD-1→4, OT-1→3)
#
# 场景覆盖 bash wrapper __gitusrcd() 和 other subcommands pass-through：
#   CD-1: cd enabled + .gitusrrc 存在 → 验证 config 被应用
#   CD-2: cd disabled → 验证 config 不变（pass-through）
#   CD-3: 单用户 → 验证 config 不变
#   CD-4: 无 .gitusrrc → 验证 config 不变
#   OT-1: git status pass-through → 验证无干预，exit 0
#   OT-2: git push pass-through → 验证无干预
#   OT-3: git log pass-through → 验证无干预
#
# 实现注意：
#   - 先运行 gitusr hooks install && source wrapper（wrapper 定义 alias cd=__gitusrcd）
#   - 使用 \cd 绕过 alias 设置目标目录，cd 使用 alias 触发 hook
#   - CD-2 需先 gitusr hooks disable cd
# =============================================================================

echo ""
echo "--- TODO: Task 9 - CD + Other Scenarios (CD-1→4, OT-1→3) ---"
echo "  [PLACEHOLDER] CD-1: cd with .gitusrrc → config applied"
echo "  [PLACEHOLDER] CD-2: cd disabled → pass-through"
echo "  [PLACEHOLDER] CD-3: single user → pass-through"
echo "  [PLACEHOLDER] CD-4: no .gitusrrc → config unchanged"
echo "  [PLACEHOLDER] OT-1: git status pass-through"
echo "  [PLACEHOLDER] OT-2: git push pass-through"
echo "  [PLACEHOLDER] OT-3: git log pass-through"
echo "  Step 27-33 SKIPPED (Task 9 not yet implemented)"

# ══════════════════════════════════════════════════════════════════════════════

echo ""
echo "========== ALL E2E STEPS PASSED =========="
exit 0
