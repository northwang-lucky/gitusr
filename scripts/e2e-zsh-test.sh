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

# ─── 8-step E2E workflow ──────────────────────────────────────────────────

# Step 1: init — create the first user
echo ""
echo "--- Step 1/8: init ---"
gitusr init --name "Dev" --email "dev@test.com" --yes --force
echo "  Step 1 PASSED"

# Step 2: add Work user
echo ""
echo "--- Step 2/8: add Work ---"
gitusr add --name "Work" --email "work@test.com"
echo "  Step 2 PASSED"

# Step 3: add Personal user
echo ""
echo "--- Step 3/8: add Personal ---"
gitusr add --name "Personal" --email "personal@test.com"
echo "  Step 3 PASSED"

# Step 4: list — verify all three users are present
echo ""
echo "--- Step 4/8: list ---"
gitusr list | tee /tmp/list-output.txt
grep -q "Dev" /tmp/list-output.txt      || { echo "  FAIL: Dev not in list"; exit 1; }
grep -q "Work" /tmp/list-output.txt     || { echo "  FAIL: Work not in list"; exit 1; }
grep -q "Personal" /tmp/list-output.txt || { echo "  FAIL: Personal not in list"; exit 1; }
echo "  Step 4 PASSED (Dev, Work, Personal all present)"

# Step 5: use — switch to Work identity inside a fresh git repo
echo ""
echo "--- Step 5/8: use --email in git repo ---"
TMP_REPO=/tmp/repo
rm -rf "$TMP_REPO"
mkdir -p "$TMP_REPO"
cd "$TMP_REPO"
git init
gitusr use --email "work@test.com"
echo "  Step 5 PASSED"

# Step 6: current — verify the active repo identity is Work
echo ""
echo "--- Step 6/8: current ---"
cd "$TMP_REPO"
gitusr current | tee /tmp/current-output.txt
grep -q "work@test.com" /tmp/current-output.txt || {
    echo "  FAIL: work@test.com not in current output"
    exit 1
}
echo "  Step 6 PASSED"

# Step 7: remove — delete Personal user by email
echo ""
echo "--- Step 7/8: remove Personal ---"
cd /src
gitusr remove --email "personal@test.com"
# Verify Personal was actually removed
gitusr list | tee /tmp/list-after-remove.txt
grep -q "Personal" /tmp/list-after-remove.txt && {
    echo "  FAIL: Personal still in list after remove"
    exit 1
}
echo "  Step 7 PASSED"

# Step 8: replace — rename work@test.com to an existing user
echo ""
echo "--- Step 8/8: replace ---"
if git filter-repo --help &>/dev/null; then
    cd /tmp/repo
    # Create an initial commit authored by Work so that git-filter-repo has history
    git config user.email "work@test.com"
    git config user.name "Work"
    touch README.md && git add README.md && git commit -m "initial commit" --no-gpg-sign
    # Replace Work-authored commits with Dev identity (Dev already exists in store)
    gitusr replace work@test.com --with-index 0 --yes
    echo "  Step 8 PASSED"
else
    echo "  Step 8 SKIPPED (git-filter-repo not available)"
fi

# ─── Hook subcommand tests (Steps 9–14) ───────────────────────────────────
# Note: zsh wrapper files use .zsh extension (vs bash .sh extension).

echo ""
echo "--- Step 9/14: hook install --all ---"
gitusr hook install --all | tee /tmp/hook-install-all-output.txt
grep -q "Hook clone successfully installed" /tmp/hook-install-all-output.txt || { echo "  FAIL: clone hook was not installed by --all"; exit 1; }
grep -q "Hook commit successfully installed" /tmp/hook-install-all-output.txt || { echo "  FAIL: commit hook was not installed by --all"; exit 1; }
grep -q "Hook cd successfully installed" /tmp/hook-install-all-output.txt || { echo "  FAIL: cd hook was not installed by --all"; exit 1; }
grep -q "All hooks successfully installed" /tmp/hook-install-all-output.txt || { echo "  FAIL: --all install did not report aggregate success"; exit 1; }
if [[ -f "$XDG_DATA_HOME/gitusr/hooks/git-wrapper.zsh" ]]; then
    echo "  Step 9 PASSED (all hooks installed, zsh wrapper exists)"
else
    echo "  FAIL: git-wrapper.zsh not found after install"
    exit 1
fi

echo ""
echo "--- Step 10/14: hook install --all idempotency ---"
gitusr hook install --all | tee /tmp/hook-install-all-idempotent-output.txt
grep -q "Hook clone is already installed" /tmp/hook-install-all-idempotent-output.txt || { echo "  FAIL: clone hook was not reported already installed"; exit 1; }
grep -q "Hook commit is already installed" /tmp/hook-install-all-idempotent-output.txt || { echo "  FAIL: commit hook was not reported already installed"; exit 1; }
grep -q "Hook cd is already installed" /tmp/hook-install-all-idempotent-output.txt || { echo "  FAIL: cd hook was not reported already installed"; exit 1; }
echo "  Step 10 PASSED (--all install is idempotent)"

echo ""
echo "--- Step 11/14: hook install --all state verification ---"
WRAPPER_FILE="$XDG_DATA_HOME/gitusr/hooks/git-wrapper.zsh"
HOOK_STATE="$XDG_DATA_HOME/gitusr/hook-state.json"
if [[ -f "$WRAPPER_FILE" ]] && [[ -f "$HOOK_STATE" ]] && grep -q '"clone"' "$HOOK_STATE" && grep -q '"commit"' "$HOOK_STATE" && grep -q '"cd"' "$HOOK_STATE"; then
    echo "  Step 11 PASSED (--all state contains clone, commit, cd)"
else
    echo "  FAIL: --all state verification failed"
    exit 1
fi

echo ""
echo "--- Step 12/14: hook uninstall --all ---"
gitusr hook uninstall --all | tee /tmp/hook-uninstall-all-output.txt
grep -q "Hook clone successfully uninstalled" /tmp/hook-uninstall-all-output.txt || { echo "  FAIL: clone hook was not uninstalled by --all"; exit 1; }
grep -q "Hook commit successfully uninstalled" /tmp/hook-uninstall-all-output.txt || { echo "  FAIL: commit hook was not uninstalled by --all"; exit 1; }
grep -q "Hook cd successfully uninstalled" /tmp/hook-uninstall-all-output.txt || { echo "  FAIL: cd hook was not uninstalled by --all"; exit 1; }
grep -q "All hooks successfully uninstalled" /tmp/hook-uninstall-all-output.txt || { echo "  FAIL: --all uninstall did not report aggregate success"; exit 1; }
echo "  Step 12 PASSED (all hooks uninstalled)"

echo ""
echo "--- Step 13/14: hook uninstall --all idempotency ---"
gitusr hook uninstall --all | tee /tmp/hook-uninstall-all-idempotent-output.txt
grep -q "Hook clone is not installed" /tmp/hook-uninstall-all-idempotent-output.txt || { echo "  FAIL: clone hook was not reported not installed"; exit 1; }
grep -q "Hook commit is not installed" /tmp/hook-uninstall-all-idempotent-output.txt || { echo "  FAIL: commit hook was not reported not installed"; exit 1; }
grep -q "Hook cd is not installed" /tmp/hook-uninstall-all-idempotent-output.txt || { echo "  FAIL: cd hook was not reported not installed"; exit 1; }
echo "  Step 13 PASSED (--all uninstall skips missing hooks)"

echo ""
echo "--- Step 14/14: hook uninstall --all cleanup verification ---"
if [[ ! -f "$XDG_DATA_HOME/gitusr/hooks/git-wrapper.zsh" ]] && ! grep -q '"clone"\|"commit"\|"cd"' "$HOOK_STATE"; then
    echo "  Step 14 PASSED (--all cleanup removed wrappers and state)"
else
    echo "  FAIL: --all cleanup did not remove wrappers or state"
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
echo "--- Step 15/18: hook actual trigger - clone with --gu-email (zsh) ---"
gitusr hook install --type=clone
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
echo "--- Step 16/18: hook actual trigger - commit with .gitusrrc (zsh) ---"
gitusr hook install --type=commit
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
echo "--- Step 17/18: hook actual trigger - cd with chpwd hook (zsh) ---"
gitusr hook install --type=cd
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

echo ""
echo "--- Step 18/18: hook cleanup verification ---"
gitusr hook uninstall --type=clone
gitusr hook uninstall --type=commit
gitusr hook uninstall --type=cd
echo "  Step 18 PASSED (all hooks uninstalled)"

# ─── TODO: Extended hook scenario coverage ─────────────────────────────────
# These scenarios should be implemented as individual test cases.
# CL = Clone, CM = Commit, CD = Chdir, OT = Other.
#
# TODO CL-1:  clone hook — repo URL with trailing .git
# TODO CL-2:  clone hook — --gu-name parameter (use by name)
# TODO CL-3:  clone hook — --gu-email parameter (covered in Step 15; add edge cases)
# TODO CL-4:  clone hook — --gu-name=value and --gu-email=value syntax
# TODO CL-5:  clone hook — single user in store (hook should skip, pass through)
# TODO CL-6:  clone hook — disabled hook (gitusr hooks disable clone)
# TODO CL-7:  clone hook — clone failure (hook should not call gitusr use)
# TODO CL-8:  clone hook — target directory explicitly specified
#
# TODO CM-1:  commit hook — .gitusrrc with email (covered in Step 16)
# TODO CM-2:  commit hook — .gitusrrc with name field (match by name)
# TODO CM-3:  commit hook — no .gitusrrc present (hook should pass through)
# TODO CM-4:  commit hook — disabled hook (gitusr hooks disable commit)
#
# TODO CD-1:  cd hook — .gitusrrc with email (covered in Step 17)
# TODO CD-2:  cd hook — .gitusrrc with name field (match by name)
# TODO CD-3:  cd hook — no .gitusrrc present (hook should skip)
# TODO CD-4:  cd hook — disabled hook (gitusr hooks disable cd)
#
# TODO OT-1:  other — add user while hooks are installed (list count update)
# TODO OT-2:  other — remove user while hooks are installed (single-user skip)
# TODO OT-3:  other — re-init with hooks installed (state reset)

echo ""
echo "========== ALL 18 E2E STEPS PASSED (ZSH) =========="
exit 0
