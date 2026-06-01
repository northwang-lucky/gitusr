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

# 设置全局 git 身份，避免 commit 场景因缺少身份而失败
# （commit hook 会在 commit 前/后设置 local config 覆盖此全局配置）
git config --global user.email "test@example.com"
git config --global user.name "Test User"

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
# 重新安装 hooks（Step 14 已卸载），为运行时场景做准备
# =============================================================================

echo ""
echo "--- Re-installing hooks for runtime scenarios ---"
gitusr hooks install
WRAPPER="$XDG_DATA_HOME/gitusr/hooks/git-wrapper.sh"
source "$WRAPPER"
# 在非交互式 shell 中启用 alias 扩展（cd hook 依赖 alias）
shopt -s expand_aliases

# ══════════════════════════════════════════════════════════════════════════════
# Task 7: Clone Scenarios (CL-1 → CL-8)
# =============================================================================

# 创建共享 bare repo 供全部 clone 场景使用
CL_BARE="/tmp/bare-clone-test"
rm -rf "$CL_BARE"
git init --bare "$CL_BARE"

# --- CL-1: --gu-email val 格式 → 验证 user.email 被设置 ---
echo ""
echo "--- Step 15: CL-1 --gu-email val format ---"
\cd /tmp
rm -rf /tmp/test-cl1
run_cmd_ok git clone "$CL_BARE" test-cl1 --gu-email work@test.com
\cd /tmp/test-cl1
ACTUAL=$(git config user.email || true)
\cd /tmp
[ "$ACTUAL" = "work@test.com" ] || { echo "  FAIL: expected user.email=work@test.com, got '$ACTUAL'"; exit 1; }
echo "  CL-1 PASSED"

# --- CL-2: --gu-name val 格式 → 验证 user.name 被设置 ---
echo ""
echo "--- Step 16: CL-2 --gu-name val format ---"
\cd /tmp
rm -rf /tmp/test-cl2
run_cmd_ok git clone "$CL_BARE" test-cl2 --gu-name "Dev"
\cd /tmp/test-cl2
ACTUAL=$(git config user.name || true)
\cd /tmp
[ "$ACTUAL" = "Dev" ] || { echo "  FAIL: expected user.name=Dev, got '$ACTUAL'"; exit 1; }
echo "  CL-2 PASSED"

# --- CL-3: --gu-email=val 等号格式 → 验证 user.email 被设置 ---
echo ""
echo "--- Step 17: CL-3 --gu-email=val format ---"
\cd /tmp
rm -rf /tmp/test-cl3
run_cmd_ok git clone "$CL_BARE" test-cl3 --gu-email=work@test.com
\cd /tmp/test-cl3
ACTUAL=$(git config user.email || true)
\cd /tmp
[ "$ACTUAL" = "work@test.com" ] || { echo "  FAIL: expected user.email=work@test.com, got '$ACTUAL'"; exit 1; }
echo "  CL-3 PASSED"

# --- CL-4: --gu-name + --gu-email 同时指定 → 验证 name+email 均被设置 ---
echo ""
echo "--- Step 18: CL-4 --gu-name + --gu-email together ---"
\cd /tmp
rm -rf /tmp/test-cl4
run_cmd_ok git clone "$CL_BARE" test-cl4 --gu-name "Dev" --gu-email dev@test.com
\cd /tmp/test-cl4
ACTUAL_EMAIL=$(git config user.email || true)
ACTUAL_NAME=$(git config user.name || true)
\cd /tmp
[ "$ACTUAL_EMAIL" = "dev@test.com" ] || { echo "  FAIL: expected user.email=dev@test.com, got '$ACTUAL_EMAIL'"; exit 1; }
[ "$ACTUAL_NAME" = "Dev" ]         || { echo "  FAIL: expected user.name=Dev, got '$ACTUAL_NAME'"; exit 1; }
echo "  CL-4 PASSED"

# --- CL-5: clone disabled → pass-through（config 不变）---
echo ""
echo "--- Step 19: CL-5 clone disabled → pass-through ---"
gitusr hooks disable clone
\cd /tmp
rm -rf /tmp/test-cl5
# disabled 时 --gu-* 参数会透传给 git，导致 git 报 unknown option，因此不加 --gu-*
run_cmd_ok git clone "$CL_BARE" test-cl5
\cd /tmp/test-cl5
LOCAL_EMAIL=$(git config --local user.email || true)
LOCAL_NAME=$(git config --local user.name || true)
\cd /tmp
[ -z "$LOCAL_EMAIL" ] || { echo "  FAIL: local user.email should be empty (got '$LOCAL_EMAIL')"; exit 1; }
[ -z "$LOCAL_NAME" ]  || { echo "  FAIL: local user.name should be empty (got '$LOCAL_NAME')"; exit 1; }
gitusr hooks enable clone  # 恢复，供后续场景使用
echo "  CL-5 PASSED"

# --- CL-6: 单用户 pass-through（config 不变）---
echo ""
echo "--- Step 20: CL-6 single user pass-through ---"
gitusr remove --email "work@test.com"
\cd /tmp
rm -rf /tmp/test-cl6
# 单用户时 --gu-* 会透传给 real git 导致 unknown option，不加 --gu-*
run_cmd_ok git clone "$CL_BARE" test-cl6
\cd /tmp/test-cl6
LOCAL_EMAIL=$(git config --local user.email || true)
\cd /tmp
[ -z "$LOCAL_EMAIL" ] || { echo "  FAIL: local user.email should be empty (got '$LOCAL_EMAIL')"; exit 1; }
# 重新添加 Work 用户供后续场景使用
gitusr add --name "Work" --email "work@test.com"
echo "  CL-6 PASSED"

# --- CL-7: clone 失败 URL → 验证非零 exit code ---
echo ""
echo "--- Step 21: CL-7 clone failure (invalid URL) ---"
\cd /tmp
set +e
git clone /nonexistent/path/NOPE /tmp/test-cl7 2>/dev/null
CL7_RC=$?
set -e
if [ "$CL7_RC" -eq 0 ]; then
    echo "  FAIL: CL-7 clone of invalid URL should have failed (non-zero exit)"
    exit 1
fi
echo "  CL-7 PASSED (clone failed as expected, exit=$CL7_RC)"

# --- CL-8: 无 --gu-* 参数 → 验证不崩溃 ---
echo ""
echo "--- Step 22: CL-8 no --gu-* parameters → no crash ---"
\cd /tmp
rm -rf /tmp/test-cl8
# wrapper 会尝试调用 gitusr use（无参数=交互式），在 non-TTY 会失败。
# 用 set +e 避免脚本退出，验证 git clone 本身成功（目录被创建）即算不崩溃。
set +e
git clone "$CL_BARE" test-cl8 > /tmp/cl8-output.txt 2>&1
set -e
[ -d "/tmp/test-cl8" ] || { echo "  FAIL: test-cl8 directory not created"; exit 1; }
echo "  CL-8 PASSED (no crash, git clone succeeded, gitusr use attempted)"

# ══════════════════════════════════════════════════════════════════════════════
# Task 8: Commit Scenarios (CM-1 → CM-4)
# =============================================================================

# 注意：wrapper 已 source，git 函数已激活。commit 场景需确保有至少 2 个用户。

# --- CM-1: .gitusrrc 存在 → 验证 commit 后 user.email 被设置 ---
echo ""
echo "--- Step 23: CM-1 .gitusrrc present → apply user.email ---"
\cd /tmp
rm -rf /tmp/test-cm1
mkdir -p /tmp/test-cm1 && \cd /tmp/test-cm1
git init
echo '{"email":"dev@test.com"}' > .gitusrrc
touch cm1-file && git add cm1-file
git commit -m "CM-1 test" --no-gpg-sign
CM1_EMAIL=$(git config --local user.email || true)
CM1_NAME=$(git config --local user.name || true)
[ "$CM1_EMAIL" = "dev@test.com" ] || { echo "  FAIL: expected local user.email=dev@test.com, got '$CM1_EMAIL'"; exit 1; }
[ "$CM1_NAME" = "Dev" ]          || { echo "  FAIL: expected local user.name=Dev, got '$CM1_NAME'"; exit 1; }
echo "  CM-1 PASSED"

# --- CM-2: 无 .gitusrrc → 验证 config 不变（pass-through）---
echo ""
echo "--- Step 24: CM-2 no .gitusrrc → pass-through ---"
\cd /tmp
rm -rf /tmp/test-cm2
mkdir -p /tmp/test-cm2 && \cd /tmp/test-cm2
git init
touch cm2-file && git add cm2-file
git commit -m "CM-2 test" --no-gpg-sign
LOCAL_EMAIL=$(git config --local user.email || true)
[ -z "$LOCAL_EMAIL" ] || { echo "  FAIL: local user.email should be empty, got '$LOCAL_EMAIL'"; exit 1; }
echo "  CM-2 PASSED"

# --- CM-3: commit disabled → 验证 config 不变（pass-through）---
echo ""
echo "--- Step 25: CM-3 commit disabled → pass-through ---"
gitusr hooks disable commit
\cd /tmp
rm -rf /tmp/test-cm3
mkdir -p /tmp/test-cm3 && \cd /tmp/test-cm3
git init
echo '{"email":"work@test.com"}' > .gitusrrc
touch cm3-file && git add cm3-file
git commit -m "CM-3 test" --no-gpg-sign
LOCAL_EMAIL=$(git config --local user.email || true)
[ -z "$LOCAL_EMAIL" ] || { echo "  FAIL: local user.email should be empty (commit disabled), got '$LOCAL_EMAIL'"; exit 1; }
gitusr hooks enable commit  # 恢复
echo "  CM-3 PASSED"

# --- CM-4: 单用户 → 验证 config 不变（pass-through）---
echo ""
echo "--- Step 26: CM-4 single user → pass-through ---"
gitusr remove --email "work@test.com"
\cd /tmp
rm -rf /tmp/test-cm4
mkdir -p /tmp/test-cm4 && \cd /tmp/test-cm4
git init
echo '{"email":"dev@test.com"}' > .gitusrrc
touch cm4-file && git add cm4-file
git commit -m "CM-4 test" --no-gpg-sign
LOCAL_EMAIL=$(git config --local user.email || true)
[ -z "$LOCAL_EMAIL" ] || { echo "  FAIL: local user.email should be empty (single user), got '$LOCAL_EMAIL'"; exit 1; }
# 重新添加 Work 用户供后续场景使用
gitusr add --name "Work" --email "work@test.com"
echo "  CM-4 PASSED"

# ══════════════════════════════════════════════════════════════════════════════
# Task 9: CD + Other Scenarios (CD-1→4, OT-1→3)
# =============================================================================

# 注意：wrapper 已 source，alias cd=__gitusrcd 已生效。
# 使用 \cd 绕过 alias 设置目标目录，使用 cd（alias）触发 hook。

# --- CD-1: cd enabled + .gitusrrc 存在 → 验证 config 被应用 ---
echo ""
echo "--- Step 27: CD-1 cd with .gitusrrc → config applied ---"
\cd /tmp
rm -rf /tmp/test-cd1
mkdir -p /tmp/test-cd1 && \cd /tmp/test-cd1
git init
echo '{"email":"dev@test.com"}' > .gitusrrc
\cd /tmp  # bypass alias to leave the dir
cd /tmp/test-cd1  # use alias to trigger __gitusrcd()
CD1_EMAIL=$(git config --local user.email || true)
[ "$CD1_EMAIL" = "dev@test.com" ] || { echo "  FAIL: expected local user.email=dev@test.com, got '$CD1_EMAIL'"; exit 1; }
echo "  CD-1 PASSED"

# --- CD-2: cd disabled → 验证 config 不变（pass-through）---
echo ""
echo "--- Step 28: CD-2 cd disabled → pass-through ---"
gitusr hooks disable cd
\cd /tmp
rm -rf /tmp/test-cd2
mkdir -p /tmp/test-cd2 && \cd /tmp/test-cd2
git init
echo '{"email":"work@test.com"}' > .gitusrrc
\cd /tmp  # bypass alias to leave the dir
cd /tmp/test-cd2  # use alias (should pass-through because cd is disabled)
LOCAL_EMAIL=$(git config --local user.email || true)
[ -z "$LOCAL_EMAIL" ] || { echo "  FAIL: local user.email should be empty (cd disabled), got '$LOCAL_EMAIL'"; exit 1; }
gitusr hooks enable cd  # 恢复
echo "  CD-2 PASSED"

# --- CD-3: 单用户 → 验证 config 不变 ---
echo ""
echo "--- Step 29: CD-3 single user → pass-through ---"
gitusr remove --email "work@test.com"
\cd /tmp
rm -rf /tmp/test-cd3
mkdir -p /tmp/test-cd3 && \cd /tmp/test-cd3
git init
echo '{"email":"dev@test.com"}' > .gitusrrc
\cd /tmp  # bypass alias to leave the dir
cd /tmp/test-cd3  # use alias (should pass-through because single user)
LOCAL_EMAIL=$(git config --local user.email || true)
[ -z "$LOCAL_EMAIL" ] || { echo "  FAIL: local user.email should be empty (single user), got '$LOCAL_EMAIL'"; exit 1; }
# 重新添加 Work 用户供后续场景使用
gitusr add --name "Work" --email "work@test.com"
echo "  CD-3 PASSED"

# --- CD-4: 无 .gitusrrc → 验证 config 不变 ---
echo ""
echo "--- Step 30: CD-4 no .gitusrrc → config unchanged ---"
\cd /tmp
rm -rf /tmp/test-cd4
mkdir -p /tmp/test-cd4 && \cd /tmp/test-cd4
git init
# 不创建 .gitusrrc
\cd /tmp  # bypass alias to leave the dir
cd /tmp/test-cd4  # use alias (no .gitusrrc → pass-through)
LOCAL_EMAIL=$(git config --local user.email || true)
[ -z "$LOCAL_EMAIL" ] || { echo "  FAIL: local user.email should be empty (no .gitusrrc), got '$LOCAL_EMAIL'"; exit 1; }
echo "  CD-4 PASSED"

# --- OT-1: git status pass-through → 验证无干预，exit 0 ---
echo ""
echo "--- Step 31: OT-1 git status pass-through ---"
\cd /tmp
rm -rf /tmp/test-ot1
mkdir -p /tmp/test-ot1 && \cd /tmp/test-ot1
git init
run_cmd_ok git status
echo "  OT-1 PASSED"

# --- OT-2: git push pass-through → 验证无干预 ---
echo ""
echo "--- Step 32: OT-2 git push pass-through ---"
\cd /tmp
rm -rf /tmp/test-ot2
mkdir -p /tmp/test-ot2 && \cd /tmp/test-ot2
git init
touch ot2-file && git add ot2-file && git commit -m "OT-2 init" --no-gpg-sign
PUSH_TARGET="/tmp/push-target-ot2"
rm -rf "$PUSH_TARGET"
git init --bare "$PUSH_TARGET"
git remote add origin "$PUSH_TARGET"
run_cmd_ok git push origin HEAD:refs/heads/main
echo "  OT-2 PASSED"

# --- OT-3: git log pass-through → 验证无干预 ---
echo ""
echo "--- Step 33: OT-3 git log pass-through ---"
\cd /tmp
rm -rf /tmp/test-ot3
mkdir -p /tmp/test-ot3 && \cd /tmp/test-ot3
git init
touch ot3-file && git add ot3-file && git commit -m "OT-3 init" --no-gpg-sign
run_cmd_ok git log --oneline
echo "  OT-3 PASSED"

# ══════════════════════════════════════════════════════════════════════════════

echo ""
echo "========== ALL E2E STEPS PASSED =========="
exit 0
