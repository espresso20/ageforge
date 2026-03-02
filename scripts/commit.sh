#!/usr/bin/env bash
# scripts/commit.sh — interactive conventional commit helper
# Replaces:  git add . && git commit -v && git push
# Usage:     bash scripts/commit.sh   OR   make commit

set -uo pipefail

GOLD='\033[0;33m'
GREEN='\033[0;32m'
CYAN='\033[0;36m'
RED='\033[0;31m'
GRAY='\033[0;90m'
BOLD='\033[1m'
RESET='\033[0m'

# Stage everything
git add .

# Bail early if nothing changed
if git diff --cached --quiet; then
  echo -e "${GRAY}Nothing staged — nothing to commit.${RESET}"
  exit 0
fi

# ── Show what's staged ────────────────────────────────────────────────────────
echo ""
echo -e "${GOLD}┌── Staged Changes ──────────────────────────────────────────────┐${RESET}"
git diff --cached --stat | sed 's/^/│  /'
echo -e "${GOLD}└────────────────────────────────────────────────────────────────┘${RESET}"
echo ""

# ── Pick type ─────────────────────────────────────────────────────────────────
echo -e "${CYAN}${BOLD}What kind of change?${RESET}"
echo -e "  ${GOLD}1${RESET}  feat      — new feature or content          → ### Added"
echo -e "  ${GOLD}2${RESET}  fix       — bug fix                         → ### Fixed"
echo -e "  ${GOLD}3${RESET}  balance   — tuning costs, rates, numbers    → ### Balance"
echo -e "  ${GOLD}4${RESET}  refactor  — cleanup, no behavior change     → ### Changed"
echo -e "  ${GOLD}5${RESET}  chore     — build/tooling/deps              (skipped in notes)"
echo -e "  ${GOLD}6${RESET}  docs      — docs/comments only              (skipped in notes)"
echo ""
read -rp "  Choice [1-6, default 1]: " TYPE_CHOICE || true
echo ""

case "${TYPE_CHOICE:-1}" in
  2) TYPE="fix" ;;
  3) TYPE="balance" ;;
  4) TYPE="refactor" ;;
  5) TYPE="chore" ;;
  6) TYPE="docs" ;;
  *) TYPE="feat" ;;
esac

# ── Short subject line (≤72 chars enforced) ───────────────────────────────────
while true; do
  read -rp "  Short summary (≤72 chars): " DESC || true
  DESC="$(echo "$DESC" | sed 's/^[[:space:]]*//' | sed 's/[[:space:]]*$//' | sed 's/\.$//')"
  FIRST="$(echo "${DESC:0:1}" | tr '[:upper:]' '[:lower:]')"
  DESC="${FIRST}${DESC:1}"
  SUBJECT="${TYPE}: ${DESC}"
  LEN=${#SUBJECT}
  if [[ $LEN -le 72 ]]; then
    break
  fi
  echo -e "  ${RED}Too long (${LEN} chars — max 72). Keep it short here; add details below.${RESET}"
  echo ""
done
echo ""

# ── Optional bullet-point body ────────────────────────────────────────────────
echo -e "  ${GRAY}Details / bullet points? (one per line, blank line to finish, skip with Enter)${RESET}"
BODY_LINES=()
while true; do
  read -rp "  · " LINE || true
  [[ -z "$LINE" ]] && break
  BODY_LINES+=("- $LINE")
done
echo ""

# ── Preview ───────────────────────────────────────────────────────────────────
echo -e "${GOLD}┌── Commit Message ──────────────────────────────────────────────┐${RESET}"
echo -e "│  ${BOLD}${SUBJECT}${RESET}"
if [[ ${#BODY_LINES[@]} -gt 0 ]]; then
  echo -e "│"
  for line in "${BODY_LINES[@]}"; do
    echo -e "│  ${GRAY}${line}${RESET}"
  done
fi
echo -e "${GOLD}└────────────────────────────────────────────────────────────────┘${RESET}"
echo ""

read -rp "  Commit? [Y/n]: " CONFIRM_COMMIT || true
if [[ "${CONFIRM_COMMIT:-y}" =~ ^[Nn] ]]; then
  echo -e "${GRAY}Aborted — unstaging changes.${RESET}"
  git restore --staged .
  exit 0
fi

# Build full commit message
if [[ ${#BODY_LINES[@]} -gt 0 ]]; then
  FULL_MSG="${SUBJECT}"$'\n\n'
  for line in "${BODY_LINES[@]}"; do
    FULL_MSG+="${line}"$'\n'
  done
  git commit -m "$FULL_MSG"
else
  git commit -m "$SUBJECT"
fi
echo ""

# ── Push ──────────────────────────────────────────────────────────────────────
read -rp "  Push to origin/master now? [Y/n]: " CONFIRM_PUSH || true
if [[ "${CONFIRM_PUSH:-y}" =~ ^[Nn] ]]; then
  echo -e "${GREEN}✓ Committed locally.${RESET} Push later with: git push origin master"
  exit 0
fi

git push origin master
echo ""
echo -e "${GREEN}✓ ${BOLD}${SUBJECT}${RESET}"
