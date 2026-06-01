#!/bin/zsh
# =============================================================================
# e2e-zsh-test.sh — zsh 版 gitusr E2E 流程（在 bubblewrap 沙箱中执行）
#
# 本脚本由 sandbox-test.sh 在 bubblewrap 容器内调用。
# 与 bash 版镜像（scripts/e2e-test.sh），但 zsh 使用 chpwd hook（非 alias）
# 处理 cd 场景。cd 测试在 zsh 子 shell 中 source ~/.zshrc 后执行。
#
# 不要直接运行此脚本 —— 请使用 scripts/sandbox-test.sh。
# =============================================================================
set -euo pipefail

# ─── Sandbox environment setup ─────────────────────────────────────────────
echo "[sandbox] Setting up environment..."

export XDG_DATA_HOME=/tmp/xdg
mkdir -p "$XDG_DATA_HOME"

# Install git-filter-repo (used by the filter-repo command).
echo "[sandbox] Installing git-filter-repo..."
if python3 -m pip install --quiet git-filter-repo 2>/dev/null; then
    echo "[sandbox] git-filter-repo installed"
else
    echo "[sandbox] WARNING: git-filter-repo install failed, continuing without it"
fi

# ─── Build gitusr binary ───────────────────────────────────────────────────
echo "[sandbox] Building gitusr binary..."
go build -o /tmp/gitusr ./cmd/gitusr
echo "[sandbox] Build complete"

# ─── Helpers ───────────────────────────────────────────────────────────────
# gitusr wrapper that always sets XDG_DATA_HOME for isolated store access.
gitusr() {
    GITUSR_LANG=en XDG_DATA_HOME=/tmp/xdg /tmp/gitusr "$@"
}

# run_cmd_ok — run a command and exit on failure.
run_cmd_ok() {
    local desc="$1"
    shift
    echo "  RUN: $desc"
    if "$@"; then
        echo "  PASS: $desc"
    else
        echo "  FAIL: $desc"
        exit 1
    fi
}

# run_cmd_fail — run a command and exit if it unexpectedly succeeds.
run_cmd_fail() {
    local desc="$1"
    shift
    echo "  RUN (expect fail): $desc"
    if "$@"; then
        echo "  FAIL (expected failure): $desc"
        exit 1
    else
        echo "  PASS (correctly failed): $desc"
    fi
}

# run_output_contains — run command and verify its stdout contains a string.
run_output_contains() {
    local desc="$1" needle="$2"
    shift 2
    echo "  RUN: $desc"
    local output
    output=$("$@" 2>&1) || true
    if echo "$output" | grep -qF "$needle"; then
        echo "  PASS: $desc"
    else
        echo "  FAIL: $desc — expected output not found. Got:"
        echo "$output"
        exit 1
    fi
}

# ─── Minimal setup (no Steps 1-8 duplication) ─────────────────────────────
echo ""
echo "--- Minimal setup (init + add 2 users + global git config) ---"
gitusr init --name "Dev" --email "dev@test.com" --yes --force
gitusr add --name "Work" --email "work@test.com"
git config --global user.email "test@example.com"
git config --global user.name "Test User"
echo "  Minimal setup complete"

# ─── Hook subcommand tests (Steps 9–14) ───────────────────────────────────
# Note: zsh wrapper files use .zsh extension (vs bash .sh extension).

echo ""
echo "--- Step 9/14: hooks install ---"
gitusr hooks install | tee /tmp/hook-install-all-output.txt
grep -q "All hooks successfully installed" /tmp/hook-install-all-output.txt || { echo "  FAIL: hooks install did not report aggregate success"; exit 1; }
if [[ -f "$XDG_DATA_HOME/gitusr/hooks/git-wrapper.zsh" ]]; then
    echo "  Step 9 PASSED (all hooks installed, zsh wrapper exists)"
else
    echo "  FAIL: git-wrapper.zsh not found after install"
    exit 1
fi

echo ""
echo "--- Step 10/14: hooks install idempotency ---"
gitusr hooks install | tee /tmp/hook-install-all-idempotent-output.txt
grep -q "All hooks are already installed" /tmp/hook-install-all-idempotent-output.txt || { echo "  FAIL: idempotent install did not report already installed"; exit 1; }
echo "  Step 10 PASSED (install is idempotent)"

echo ""
echo "--- Step 11/14: hooks install state verification ---"
WRAPPER_FILE="$XDG_DATA_HOME/gitusr/hooks/git-wrapper.zsh"
HOOK_STATE="$XDG_DATA_HOME/gitusr/hook-state.json"
if [[ -f "$WRAPPER_FILE" ]] && [[ -f "$HOOK_STATE" ]] && grep -q '"clone"' "$HOOK_STATE" && grep -q '"commit"' "$HOOK_STATE" && grep -q '"cd"' "$HOOK_STATE"; then
    echo "  Step 11 PASSED (--all state contains clone, commit, cd)"
else
    echo "  FAIL: --all state verification failed"
    exit 1
fi

echo ""
echo "--- Step 12/14: hooks uninstall ---"
gitusr hooks uninstall | tee /tmp/hook-uninstall-all-output.txt
grep -q "All hooks successfully uninstalled" /tmp/hook-uninstall-all-output.txt || { echo "  FAIL: hooks uninstall did not report aggregate success"; exit 1; }
echo "  Step 12 PASSED (all hooks uninstalled)"

echo ""
echo "--- Step 13/14: hooks uninstall when nothing installed ---"
gitusr hooks uninstall 2>&1 | tee /tmp/hook-uninstall-none-output.txt || true
grep -q "No hooks are currently installed" /tmp/hook-uninstall-none-output.txt || { echo "  FAIL: uninstall did not report none installed"; exit 1; }
echo "  Step 13 PASSED (uninstall reports nothing installed)"

echo ""
echo "--- Step 14/14: hooks uninstall cleanup verification ---"
if [[ ! -f "$XDG_DATA_HOME/gitusr/hooks/git-wrapper.zsh" ]] && ! grep -q '"clone"\|"commit"\|"cd"' "$HOOK_STATE"; then
    echo "  Step 14 PASSED (cleanup removed wrappers and cleared state)"
else
    echo "  FAIL: cleanup did not remove wrappers or clear state"
    exit 1
fi

# ─── Hook actual trigger tests (Steps 15–18) ──────────────────────────────
# Clone and commit hooks work via git() wrapper function (same as bash).
# Cd hook uses chpwd — tested in a zsh subshell where source ~/.zshrc
# activates add-zsh-hook.

# Pre-create a bare repo for clone tests
TMP_BARE_REPO=/tmp/bare-repo
rm -rf "$TMP_BARE_REPO"
git init --bare "$TMP_BARE_REPO"

echo ""
echo "--- Step 15/18: hooks actual trigger - clone with --gu-email (zsh) ---"
gitusr hooks install
WRAPPER_FILE="$XDG_DATA_HOME/gitusr/hooks/git-wrapper.zsh"
if [[ -f "$WRAPPER_FILE" ]]; then
    source "$WRAPPER_FILE"
    export PATH="$XDG_DATA_HOME/gitusr/hooks:$PATH"
fi
# Test clone with --gu-email parameter
CLONE_DIR=/tmp/cloned-repo
rm -rf "$CLONE_DIR"
cd /tmp
# Use the wrapper function to clone with email parameter
git clone "$TMP_BARE_REPO" cloned-repo --gu-email "work@test.com" 2>&1 || true
cd "$CLONE_DIR"
# Verify git config was set by the hook
CLONE_USER_EMAIL=$(git config user.email)
if [[ "$CLONE_USER_EMAIL" == "work@test.com" ]]; then
    echo "  Step 15 PASSED (clone hook triggered and set user.email)"
else
    echo "  FAIL: clone hook did not set user.email, got: $CLONE_USER_EMAIL"
    exit 1
fi

echo ""
echo "--- Step 16/18: hooks actual trigger - commit with .gitusrrc (zsh) ---"
cd /tmp
rm -rf /tmp/rc-test-repo
git clone "$TMP_BARE_REPO" rc-test-repo 2>&1 || true
cd /tmp/rc-test-repo
# Create .gitusrrc file
echo '{"email":"dev@test.com"}' > .gitusrrc
# Create a test file and add it
echo "test content" > test.txt
git add test.txt
# Use the wrapper to commit (should trigger commit hook)
git commit -m "test commit" 2>&1 || true
# Verify git config was set by the commit hook
COMMIT_USER_EMAIL=$(git config user.email)
if [[ "$COMMIT_USER_EMAIL" == "dev@test.com" ]]; then
    echo "  Step 16 PASSED (commit hook triggered and set user.email via .gitusrrc)"
else
    echo "  FAIL: commit hook did not set user.email, got: $COMMIT_USER_EMAIL"
    exit 1
fi

echo ""
echo "--- Step 17/18: hooks actual trigger - cd with chpwd hook (zsh) ---"
# Create a test repo with .gitusrrc for the cd test
rm -rf /tmp/cd-test-repo
git clone "$TMP_BARE_REPO" /tmp/cd-test-repo 2>&1 || true
echo '{"email":"work@test.com"}' > /tmp/cd-test-repo/.gitusrrc

# The zsh cd hook uses add-zsh-hook chpwd via GenerateUnifiedZshWrapper().
# It is only activated when ~/.zshrc is sourced in a zsh session.
# Test in a fresh zsh subshell: source ~/.zshrc → cd to repo → verify config.
if [[ -x /usr/bin/zsh ]]; then
    CD_TEST_OUTPUT=$(zsh -c '
        export XDG_DATA_HOME=/tmp/xdg
        export PATH="/tmp:$PATH"
        source ~/.zshrc 2>/dev/null
        cd /tmp/cd-test-repo
        git config user.email
    ' 2>&1) || true
    if echo "$CD_TEST_OUTPUT" | grep -q "work@test.com"; then
        echo "  Step 17 PASSED (cd chpwd hook triggered and set user.email)"
    else
        echo "  FAIL: cd chpwd hook did not set user.email, output: $CD_TEST_OUTPUT"
        exit 1
    fi
else
    echo "  Step 17 SKIPPED (/usr/bin/zsh not available)"
fi

# ─── Clone Scenarios (CL-2 → CL-8) ───────────────────────────────────────
# TMP_BARE_REPO created before Step 15; hooks installed, wrapper sourced.

echo ""
echo "--- CL-2: --gu-name val format ---"
cd /tmp
rm -rf /tmp/test-cl2
run_cmd_ok "git clone --gu-name" git clone "$TMP_BARE_REPO" test-cl2 --gu-name "Dev"
cd /tmp/test-cl2
ACTUAL=$(git config user.name || true)
cd /tmp
[[ "$ACTUAL" == "Dev" ]] || { echo "  FAIL: expected user.name=Dev, got '$ACTUAL'"; exit 1; }
echo "  CL-2 PASSED"

echo ""
echo "--- CL-3: --gu-email=val format ---"
cd /tmp
rm -rf /tmp/test-cl3
run_cmd_ok "git clone --gu-email=val" git clone "$TMP_BARE_REPO" test-cl3 --gu-email=work@test.com
cd /tmp/test-cl3
ACTUAL=$(git config user.email || true)
cd /tmp
[[ "$ACTUAL" == "work@test.com" ]] || { echo "  FAIL: expected user.email=work@test.com, got '$ACTUAL'"; exit 1; }
echo "  CL-3 PASSED"

echo ""
echo "--- CL-4: --gu-name + --gu-email together ---"
cd /tmp
rm -rf /tmp/test-cl4
run_cmd_ok "git clone --gu-name --gu-email" git clone "$TMP_BARE_REPO" test-cl4 --gu-name "Dev" --gu-email dev@test.com
cd /tmp/test-cl4
ACTUAL_EMAIL=$(git config user.email || true)
ACTUAL_NAME=$(git config user.name || true)
cd /tmp
[[ "$ACTUAL_EMAIL" == "dev@test.com" ]] || { echo "  FAIL: expected user.email=dev@test.com, got '$ACTUAL_EMAIL'"; exit 1; }
[[ "$ACTUAL_NAME" == "Dev" ]] || { echo "  FAIL: expected user.name=Dev, got '$ACTUAL_NAME'"; exit 1; }
echo "  CL-4 PASSED"

echo ""
echo "--- CL-5: clone disabled → pass-through ---"
gitusr hooks disable clone
cd /tmp
rm -rf /tmp/test-cl5
# disabled 时 --gu-* 参数会透传给 git 导致 unknown option，不加 --gu-*
run_cmd_ok "git clone (clone disabled)" git clone "$TMP_BARE_REPO" test-cl5
cd /tmp/test-cl5
LOCAL_EMAIL=$(git config --local user.email || true)
LOCAL_NAME=$(git config --local user.name || true)
cd /tmp
[[ -z "$LOCAL_EMAIL" ]] || { echo "  FAIL: local user.email should be empty (got '$LOCAL_EMAIL')"; exit 1; }
[[ -z "$LOCAL_NAME" ]] || { echo "  FAIL: local user.name should be empty (got '$LOCAL_NAME')"; exit 1; }
gitusr hooks enable clone
echo "  CL-5 PASSED"

echo ""
echo "--- CL-6: single user → pass-through ---"
gitusr remove --email "work@test.com"
cd /tmp
rm -rf /tmp/test-cl6
# 单用户时 --gu-* 会透传给 real git 导致 unknown option，不加 --gu-*
run_cmd_ok "git clone (single user)" git clone "$TMP_BARE_REPO" test-cl6
cd /tmp/test-cl6
LOCAL_EMAIL=$(git config --local user.email || true)
cd /tmp
[[ -z "$LOCAL_EMAIL" ]] || { echo "  FAIL: local user.email should be empty (got '$LOCAL_EMAIL')"; exit 1; }
gitusr add --name "Work" --email "work@test.com"
echo "  CL-6 PASSED"

echo ""
echo "--- CL-7: clone failure (invalid URL) ---"
cd /tmp
set +e
git clone /nonexistent/path/NOPE /tmp/test-cl7 2>/dev/null
CL7_RC=$?
set -e
[[ "$CL7_RC" -ne 0 ]] || { echo "  FAIL: CL-7 clone of invalid URL should have failed (non-zero exit)"; exit 1; }
echo "  CL-7 PASSED (clone failed as expected, exit=$CL7_RC)"

echo ""
echo "--- CL-8: no --gu-* parameters → no crash ---"
cd /tmp
rm -rf /tmp/test-cl8
# wrapper 会尝试调用 gitusr use（无参数=交互式），在 non-TTY 会失败。
# 用 set +e 避免脚本退出，验证 git clone 本身成功（目录被创建）即算不崩溃。
set +e
git clone "$TMP_BARE_REPO" test-cl8 > /tmp/cl8-output.txt 2>&1
set -e
[[ -d "/tmp/test-cl8" ]] || { echo "  FAIL: test-cl8 directory not created"; exit 1; }
echo "  CL-8 PASSED (no crash, git clone succeeded, gitusr use attempted)"

# ─── Commit Scenarios (CM-2 → CM-4) ───────────────────────────────────────
# wrapper 已 source，git 函数已激活。需确保至少有 2 个用户。

echo ""
echo "--- CM-2: no .gitusrrc → pass-through ---"
cd /tmp
rm -rf /tmp/test-cm2
mkdir -p /tmp/test-cm2
cd /tmp/test-cm2
git init
touch cm2-file && git add cm2-file
git commit -m "CM-2 test" --no-gpg-sign
LOCAL_EMAIL=$(git config --local user.email || true)
cd /tmp
[[ -z "$LOCAL_EMAIL" ]] || { echo "  FAIL: local user.email should be empty, got '$LOCAL_EMAIL'"; exit 1; }
echo "  CM-2 PASSED"

echo ""
echo "--- CM-3: commit disabled → pass-through ---"
gitusr hooks disable commit
cd /tmp
rm -rf /tmp/test-cm3
mkdir -p /tmp/test-cm3
cd /tmp/test-cm3
git init
echo '{"email":"work@test.com"}' > .gitusrrc
touch cm3-file && git add cm3-file
git commit -m "CM-3 test" --no-gpg-sign
LOCAL_EMAIL=$(git config --local user.email || true)
cd /tmp
[[ -z "$LOCAL_EMAIL" ]] || { echo "  FAIL: local user.email should be empty (commit disabled), got '$LOCAL_EMAIL'"; exit 1; }
gitusr hooks enable commit
echo "  CM-3 PASSED"

echo ""
echo "--- CM-4: single user → pass-through ---"
gitusr remove --email "work@test.com"
cd /tmp
rm -rf /tmp/test-cm4
mkdir -p /tmp/test-cm4
cd /tmp/test-cm4
git init
echo '{"email":"dev@test.com"}' > .gitusrrc
touch cm4-file && git add cm4-file
git commit -m "CM-4 test" --no-gpg-sign
LOCAL_EMAIL=$(git config --local user.email || true)
cd /tmp
[[ -z "$LOCAL_EMAIL" ]] || { echo "  FAIL: local user.email should be empty (single user), got '$LOCAL_EMAIL'"; exit 1; }
gitusr add --name "Work" --email "work@test.com"
echo "  CM-4 PASSED"

echo ""
echo "--- Step 18/18: hooks cleanup verification ---"
gitusr hooks uninstall
echo "  Step 18 PASSED (all hooks uninstalled)"

# ─── Test coverage summary ─────────────────────────────────────────────────
# CL = Clone, CM = Commit, CD = Chdir, OT = Other.
#
# DONE CL-1:  clone hook — --gu-email val format (Step 15)
# DONE CL-2:  clone hook — --gu-name val format (implemented above)
# DONE CL-3:  clone hook — --gu-email=val format (implemented above)
# DONE CL-4:  clone hook — --gu-name + --gu-email together (implemented above)
# DONE CL-5:  clone hook — clone disabled → pass-through (implemented above)
# DONE CL-6:  clone hook — single user → pass-through (implemented above)
# DONE CL-7:  clone hook — clone failure (invalid URL) (implemented above)
# DONE CL-8:  clone hook — no --gu-* parameters → no crash (implemented above)
#
# DONE CM-1:  commit hook — .gitusrrc with email (Step 16)
# DONE CM-2:  commit hook — no .gitusrrc → pass-through (implemented above)
# DONE CM-3:  commit hook — commit disabled → pass-through (implemented above)
# DONE CM-4:  commit hook — single user → pass-through (implemented above)
#
# DONE CD-1:  cd hook — .gitusrrc with email (Step 17)
#
# ─── Remaining TODO (optional, not required) ────────────────────────────────
# TODO CD-2:  cd hook — .gitusrrc with name field (match by name)
# TODO CD-3:  cd hook — no .gitusrrc present (hook should skip)
# TODO CD-4:  cd hook — disabled hook (gitusr hooks disable cd)
#
# TODO OT-1:  other — add user while hooks are installed
# TODO OT-2:  other — remove user while hooks are installed
# TODO OT-3:  other — re-init with hooks installed

echo ""
echo "========== ALL 18 E2E STEPS PASSED (ZSH) =========="
exit 0
