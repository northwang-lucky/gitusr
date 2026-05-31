#!/bin/bash
# =============================================================================
# sandbox-test.sh — Run gitusr E2E tests inside a bubblewrap sandbox
#
# Usage: ./scripts/sandbox-test.sh
#
# This script:
#   1. Resolves tool paths dynamically (go, bwrap, git, python3)
#   2. Invokes bubblewrap with an isolated filesystem
#   3. Delegates the 8-step E2E flow to scripts/e2e-test.sh inside the sandbox
#
# Requirements: go, bwrap, git, python3
# =============================================================================
set -euo pipefail

# ─── Pre-flight checks ────────────────────────────────────────────────────
check_tool() {
    if ! command -v "$1" &>/dev/null; then
        echo "ERROR: '$1' is required but not found in PATH" >&2
        exit 1
    fi
}

check_tool go
check_tool bwrap
check_tool git
check_tool python3

# ─── Dynamic paths ─────────────────────────────────────────────────────────
GOROOT="${GOROOT:-$(go env GOROOT)}"
PROJECT_ROOT="$(cd "$(dirname "$0")/.." && pwd)"

if [[ ! -f "$PROJECT_ROOT/go.mod" ]]; then
    echo "ERROR: go.mod not found at $PROJECT_ROOT. Run from the repo root." >&2
    exit 1
fi

E2E_SCRIPT="$PROJECT_ROOT/scripts/e2e-test.sh"
if [[ ! -f "$E2E_SCRIPT" ]]; then
    echo "ERROR: e2e-test.sh not found at $E2E_SCRIPT" >&2
    exit 1
fi

echo "=== gitusr E2E Sandbox Test ==="
echo "GOROOT:      $GOROOT"
echo "PROJECT:     $PROJECT_ROOT"

# ─── Run E2E inside bubblewrap sandbox ─────────────────────────────────────
# The sandbox inherits the host filesystem and selectively overrides mounts.
# /tmp and /root are isolated tmpfs volumes.
# Go toolchain is injected via $GOROOT → /usr/local/go.
# The project source is bind-mounted at /src (read-write).
bwrap \
    --tmpfs /tmp \
    --tmpfs /root \
    --ro-bind "$GOROOT" /usr/local/go \
    --ro-bind /bin /bin \
    --ro-bind /usr/bin /usr/bin \
    --ro-bind /usr/lib /usr/lib \
    --ro-bind /lib /lib \
    --ro-bind /lib64 /lib64 \
    --ro-bind /etc/ssl/certs /etc/ssl/certs \
    --ro-bind /usr/include /usr/include \
    --bind "$PROJECT_ROOT" /src \
    --proc /proc \
    --dev /dev \
    --ro-bind /etc/resolv.conf /etc/resolv.conf \
    --chdir /src \
    --setenv HOME /root \
    --setenv GOROOT /usr/local/go \
    --setenv PATH "/usr/local/go/bin:/usr/bin:/bin" \
    bash /src/scripts/e2e-test.sh

EXIT_CODE=$?

if [[ $EXIT_CODE -eq 0 ]]; then
    echo "=== E2E Sandbox Test PASSED ==="
else
    echo "=== E2E Sandbox Test FAILED (exit: $EXIT_CODE) ===" >&2
fi
exit $EXIT_CODE
