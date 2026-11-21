#!/bin/bash
# AgeForge dev script — build, vet, run
# Usage:
#   ./dev.sh          Build + vet + run
#   ./dev.sh check    Build + vet only (no run)
#   ./dev.sh run      Build + run (skip vet)
#   ./dev.sh watch    Rebuild + run on file changes (requires fswatch)

set -euo pipefail

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[0;33m'
CYAN='\033[0;36m'
GRAY='\033[0;90m'
BOLD='\033[1m'
RESET='\033[0m'

log()   { echo -e "${CYAN}[ageforge]${RESET} $1"; }
ok()    { echo -e "${GREEN}[  OK  ]${RESET} $1"; }
warn()  { echo -e "${YELLOW}[ WARN ]${RESET} $1"; }
fail()  { echo -e "${RED}[ FAIL ]${RESET} $1"; }

step_vet() {
    log "Running go vet..."
    if output=$(go vet ./... 2>&1); then
        ok "go vet passed"
    else
        fail "go vet found issues:"
        echo "$output"
        return 1
    fi
}

step_build() {
    log "Building ageforge..."
    if output=$(go build -o ageforge . 2>&1); then
        ok "Build succeeded"
    else
        fail "Build failed:"
        format_errors "$output"
        return 1
    fi
}

# Format Go compiler errors into readable output
format_errors() {
    local raw="$1"
    echo "$raw" | while IFS= read -r line; do
        # Go compile error: file.go:line:col: message
        if echo "$line" | grep -qE '^[a-zA-Z0-9_/]+\.go:[0-9]+:[0-9]+:'; then
            file=$(echo "$line" | cut -d: -f1)
            lineno=$(echo "$line" | cut -d: -f2)
            msg=$(echo "$line" | cut -d: -f4-)
            echo -e "  ${RED}error${RESET} ${GRAY}${file}:${lineno}${RESET} →${msg}"
        # Undefined/type errors
        elif echo "$line" | grep -qE '(undefined|cannot use|not enough arguments|too many arguments|declared and not used)'; then
            echo -e "  ${RED}✗${RESET} $line"
        else
            echo "  $line"
        fi
    done
}

step_run() {
    log "Starting AgeForge..."
    echo ""
    ./ageforge
}

cmd="${1:-default}"

case "$cmd" in
    check)
        step_build
        step_vet
        ok "All checks passed — ready to run"
        ;;
    run)
        step_build
        step_run
        ;;
    watch)
        if ! command -v fswatch &>/dev/null; then
            fail "fswatch not installed. Install with: brew install fswatch"
            exit 1
        fi
        log "Watching for changes... (Ctrl+C to stop)"
        step_build && step_vet && step_run &
        GAME_PID=$!
        fswatch -o --exclude '\.git' --include '\.go$' . | while read; do
            kill $GAME_PID 2>/dev/null || true
            wait $GAME_PID 2>/dev/null || true
            echo ""
            log "Files changed — rebuilding..."
            if step_build && step_vet; then
                step_run &
                GAME_PID=$!
            else
                warn "Fix errors above, will rebuild on next change"
            fi
        done
        ;;
    default)
        step_build
        step_vet
        step_run
        ;;
    *)
        echo "Usage: ./dev.sh [check|run|watch]"
        exit 1
        ;;
esac
