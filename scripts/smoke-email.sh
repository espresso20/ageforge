#!/usr/bin/env bash
# Builds the smoke-report email for the CI workflows (go.yml's smoke job and
# smoke.yml's nightly and weekly jobs): the body in $RUNNER_TEMP/smoke-email.md, and the
# subject and attachment list as step outputs.
#
# Usage: scripts/smoke-email.sh "<context>"   e.g. "PR #42", "nightly" or "weekly"
# Env:   RESULT (PASS|FAIL, from the smoke step's outcome), RUN_URL, SHA,
#        PR_URL (optional), RUNNER_TEMP, GITHUB_OUTPUT.
set -euo pipefail

context="$1"
body="${RUNNER_TEMP:-/tmp}/smoke-email.md"
out="${GITHUB_OUTPUT:-/dev/stdout}"

{
  if [ -f smoke-report/summary.md ]; then
    cat smoke-report/summary.md
  else
    echo "## AgeForge smoke: ${RESULT}"
    echo
    echo "The smoke run wrote no report; it failed before the suite finished. See the run log."
  fi
  echo
  echo "Run: ${RUN_URL}"
  if [ -n "${PR_URL:-}" ]; then
    echo
    echo "Pull request: ${PR_URL}"
  fi
} > "$body"

echo "subject=AgeForge smoke: ${RESULT} — ${context} — ${SHA:0:7}" >> "$out"

# Attach the full report while it stays small enough for a mail server.
attach=""
if [ -f smoke-report/report.md ] && [ "$(wc -c < smoke-report/report.md)" -lt 300000 ]; then
  attach="smoke-report/report.md"
fi
echo "attachments=${attach}" >> "$out"
