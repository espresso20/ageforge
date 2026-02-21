#!/bin/bash
# AgeForge dev script — build, vet, test, run
# Usage:
#   ./dev.sh          Build + vet + run
#   ./dev.sh check    Build + vet only (no run)
#   ./dev.sh test     Run full test suite with formatted output
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

# Format Go compiler/test errors into readable output
format_errors() {
    local raw="$1"
    echo "$raw" | while IFS= read -r line; do
        # Go compile error: file.go:line:col: message
        if echo "$line" | grep -qE '^[a-zA-Z0-9_/]+\.go:[0-9]+:[0-9]+:'; then
            file=$(echo "$line" | cut -d: -f1)
            lineno=$(echo "$line" | cut -d: -f2)
            msg=$(echo "$line" | cut -d: -f4-)
            echo -e "  ${RED}error${RESET} ${GRAY}${file}:${lineno}${RESET} →${msg}"
        # Go test failure: filename_test.go:line: message
        elif echo "$line" | grep -qE '^\s+[a-zA-Z0-9_]+_test\.go:[0-9]+:'; then
            file=$(echo "$line" | sed 's/^\s*//' | cut -d: -f1)
            lineno=$(echo "$line" | sed 's/^\s*//' | cut -d: -f2)
            msg=$(echo "$line" | sed 's/^\s*//' | cut -d: -f3-)
            echo -e "  ${RED}✗${RESET} ${GRAY}${file}:${lineno}${RESET} →${msg}"
        # Undefined/type errors
        elif echo "$line" | grep -qE '(undefined|cannot use|not enough arguments|too many arguments|declared and not used)'; then
            echo -e "  ${RED}✗${RESET} $line"
        else
            echo "  $line"
        fi
    done
}

step_test() {
    log "Running test suite..."
    echo ""

    local raw_output
    local exit_code=0
    raw_output=$(go test ./... -v -count=1 2>&1) || exit_code=$?

    local total=0 passed=0 failed=0 skipped=0
    local -a failed_tests=()
    local -a failed_details=()
    local current_test=""
    local current_detail=""
    local in_failure=false
    local pkg_times=""

    while IFS= read -r line; do
        # Test result lines
        if echo "$line" | grep -qE '^--- PASS:'; then
            ((passed++)) || true
            ((total++)) || true
            test_name=$(echo "$line" | sed 's/--- PASS: //' | sed 's/ (.*//')
            duration=$(echo "$line" | grep -oE '\([0-9.]+s\)' || echo "")
            echo -e "  ${GREEN}✓${RESET} ${test_name} ${GRAY}${duration}${RESET}"
            in_failure=false
        elif echo "$line" | grep -qE '^--- FAIL:'; then
            ((failed++)) || true
            ((total++)) || true
            test_name=$(echo "$line" | sed 's/--- FAIL: //' | sed 's/ (.*//')
            duration=$(echo "$line" | grep -oE '\([0-9.]+s\)' || echo "")
            echo -e "  ${RED}✗${RESET} ${test_name} ${GRAY}${duration}${RESET}"
            failed_tests+=("$test_name")
            if [ -n "$current_detail" ]; then
                failed_details+=("$current_detail")
            fi
            current_detail=""
            in_failure=false
        elif echo "$line" | grep -qE '^--- SKIP:'; then
            ((skipped++)) || true
            ((total++)) || true
            test_name=$(echo "$line" | sed 's/--- SKIP: //' | sed 's/ (.*//')
            echo -e "  ${YELLOW}○${RESET} ${test_name} ${GRAY}(skipped)${RESET}"
            in_failure=false
        # Capture failure details (indented lines after a FAIL test header)
        elif echo "$line" | grep -qE '^\s+[a-zA-Z0-9_]+_test\.go:[0-9]+:'; then
            file=$(echo "$line" | sed 's/^\s*//' | cut -d: -f1)
            lineno=$(echo "$line" | sed 's/^\s*//' | cut -d: -f2)
            msg=$(echo "$line" | sed 's/^\s*//' | cut -d: -f3-)
            current_detail+="    ${RED}→${RESET} ${GRAY}${file}:${lineno}${RESET}${msg}\n"
            in_failure=true
        elif echo "$line" | grep -qE '^=== RUN'; then
            # Save any pending detail for previous test
            if [ -n "$current_detail" ] && [ ${#failed_tests[@]} -gt 0 ]; then
                failed_details+=("$current_detail")
                current_detail=""
            fi
        # Package summary
        elif echo "$line" | grep -qE '^(ok|FAIL)\s'; then
            pkg=$(echo "$line" | awk '{print $2}')
            dur=$(echo "$line" | grep -oE '[0-9.]+s' | head -1 || echo "")
            pkg_short=$(basename "$pkg")
            if echo "$line" | grep -qE '^ok'; then
                pkg_times+="  ${GREEN}✓${RESET} ${pkg_short} ${GRAY}(${dur})${RESET}\n"
            else
                pkg_times+="  ${RED}✗${RESET} ${pkg_short} ${GRAY}(${dur})${RESET}\n"
            fi
        fi
    done <<< "$raw_output"

    # Capture last failure detail
    if [ -n "$current_detail" ] && [ ${#failed_tests[@]} -gt 0 ] && [ ${#failed_details[@]} -lt ${#failed_tests[@]} ]; then
        failed_details+=("$current_detail")
    fi

    # Summary
    echo ""
    echo -e "${BOLD}━━━ Test Summary ━━━${RESET}"

    if [ -n "$pkg_times" ]; then
        echo -e "\n${BOLD}Packages:${RESET}"
        echo -e "$pkg_times"
    fi

    echo -e "${BOLD}Results:${RESET}  ${GREEN}${passed} passed${RESET}  ${RED}${failed} failed${RESET}  ${YELLOW}${skipped} skipped${RESET}  ${GRAY}(${total} total)${RESET}"

    # Show failure details
    if [ ${#failed_tests[@]} -gt 0 ]; then
        echo ""
        echo -e "${RED}${BOLD}Failures:${RESET}"
        for i in "${!failed_tests[@]}"; do
            echo -e "  ${RED}✗${RESET} ${BOLD}${failed_tests[$i]}${RESET}"
            if [ "$i" -lt ${#failed_details[@]} ] && [ -n "${failed_details[$i]}" ]; then
                echo -e "${failed_details[$i]}"
            fi
        done
    fi

    echo ""
    if [ "$exit_code" -eq 0 ]; then
        ok "All ${total} tests passed"
    else
        fail "${failed} test(s) failed"
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
