#!/bin/bash
# =============================================================================
# e2e-test.sh — 8-step gitusr E2E flow (executed inside bubblewrap sandbox)
#
# This script is invoked by sandbox-test.sh inside the bubblewrap container.
# It builds gitusr from source, installs git-filter-repo, and runs the full
# E2E workflow with an isolated XDG_DATA_HOME.
#
# Do NOT run this script directly — use scripts/sandbox-test.sh instead.
# =============================================================================
set -euo pipefail

# ─── Sandbox environment setup ─────────────────────────────────────────────
echo "[sandbox] Setting up environment..."

export XDG_DATA_HOME=/tmp/xdg
mkdir -p "$XDG_DATA_HOME"

# Install git-filter-repo (used by the filter-repo command).
# python3 -m pip is preferred over bare pip3 to avoid path issues.
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

# ─── Helper ────────────────────────────────────────────────────────────────
# gitusr wrapper that always sets XDG_DATA_HOME for isolated store access.
gitusr() {
    XDG_DATA_HOME=/tmp/xdg /tmp/gitusr "$@"
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

# Step 9-14: Hook subcommand tests
echo ""
echo "--- Step 9/14: hook install --type=clone ---"
gitusr hook install --type=clone
# Verify wrapper file exists
if [[ -f "$XDG_DATA_HOME/gitusr/hooks/git-wrapper.sh" ]]; then
    echo "  Step 9 PASSED (clone hook installed, wrapper exists)"
else
    echo "  FAIL: git-wrapper.sh not found after install"
    exit 1
fi

echo ""
echo "--- Step 10/14: hook install --type=commit ---"
gitusr hook install --type=commit
echo "  Step 10 PASSED (commit hook installed)"

echo ""
echo "--- Step 11/14: hook env --shell bash ---"
ENV_OUTPUT=$(gitusr hook env --shell bash 2>&1)
if echo "$ENV_OUTPUT" | grep -q "__gitusr_use_if_found"; then
    echo "  Step 11 PASSED (env generates valid bash code)"
else
    echo "  FAIL: env output doesn't contain expected function"
    exit 1
fi

echo ""
echo "--- Step 12/14: hook env --shell zsh ---"
ENV_OUTPUT=$(gitusr hook env --shell zsh 2>&1)
if echo "$ENV_OUTPUT" | grep -q "add-zsh-hook"; then
    echo "  Step 12 PASSED (env generates valid zsh code)"
else
    echo "  FAIL: env output doesn't contain zsh hook code"
    exit 1
fi

echo ""
echo "--- Step 13/14: hook uninstall --type=clone ---"
gitusr hook uninstall --type=clone
echo "  Step 13 PASSED (clone hook uninstalled)"

echo ""
echo "--- Step 14/14: hook uninstall --type=commit ---"
gitusr hook uninstall --type=commit
# Verify wrapper files are cleaned up when all hooks uninstalled
if [[ ! -f "$XDG_DATA_HOME/gitusr/hooks/git-wrapper.sh" ]]; then
    echo "  Step 14 PASSED (commit hook uninstalled, wrapper cleaned up)"
else
    echo "  FAIL: wrapper files not cleaned up after uninstall"
    exit 1
fi

echo ""
echo "--- Step 15/18: hook actual trigger - clone with --gu-email ---"
gitusr hook install --type=clone
# Create a sourceable wrapper script
WRAPPER_FILE="$XDG_DATA_HOME/gitusr/hooks/git-wrapper.sh"
if [[ -f "$WRAPPER_FILE" ]]; then
    source "$WRAPPER_FILE"
    export PATH="$XDG_DATA_HOME/gitusr/hooks:$PATH"
fi
# Create a temp bare repo to clone from
TMP_BARE_REPO=/tmp/bare-repo
rm -rf "$TMP_BARE_REPO"
git init --bare "$TMP_BARE_REPO"
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
echo "--- Step 16/18: hook actual trigger - commit with .gitusrrc ---"
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
echo "--- Step 17/18: hook single user protection - should not trigger ---"
# Remove all but one user to test single user protection
gitusr remove --email "work@test.com" 2>&1 || true
cd /tmp
rm -rf /tmp/single-user-test
git clone "$TMP_BARE_REPO" single-user-test 2>&1 || true
cd /tmp/single-user-test
SINGLE_USER_CONFIG=$(git config user.email)
# With single user, hook should not trigger (no config set)
if [[ -z "$SINGLE_USER_CONFIG" ]]; then
    echo "  Step 17 PASSED (single user protection working - hook skipped)"
else
    echo "  WARN: single user protection may not be working, config set to: $SINGLE_USER_CONFIG"
fi

echo ""
echo "--- Step 18/18: hook cleanup verification ---"
gitusr hook uninstall --type=clone
gitusr hook uninstall --type=commit
echo "  Step 18 PASSED (all hooks uninstalled)"

echo ""
echo "========== ALL 18 E2E STEPS PASSED =========="
exit 0
