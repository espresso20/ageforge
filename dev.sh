#!/bin/bash
# AgeForge dev script — build, vet, test, run
# Usage:
#   ./dev.sh          Build + vet + run
#   ./dev.sh check    Build + vet only (no run)
#   ./dev.sh test     Run tests
#   ./dev.sh run      Build + run (skip vet)
#   ./dev.sh watch    Rebuild + run on file changes (requires fswatch)

set -euo pipefail

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[0;33m'
CYAN='\033[0;36m'
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
        echo "$output"
        return 1
    fi
}

step_test() {
    log "Running tests..."
    if output=$(go test ./... -v 2>&1); then
        ok "All tests passed"
        # Show test output if there were actual test files
        if echo "$output" | grep -q "^---"; then
            echo "$output" | grep -E "^(---|PASS|FAIL|ok)"
        else
            warn "No test files found in any package"
        fi
    else
        fail "Tests failed:"
        echo "$output"
        return 1
    fi
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
    test)
        step_build
        step_vet
        step_test
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
        echo "Usage: ./dev.sh [check|test|run|watch]"
        exit 1
        ;;
esac
