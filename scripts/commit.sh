#!/usr/bin/env bash
# scripts/commit.sh — interactive conventional commit helper
# Replaces:  git add . && git commit -v && git push
# Usage:     bash scripts/commit.sh   OR   make commit

set -uo pipefail

GOLD='\033[0;33m'
GREEN='\033[0;32m'
CYAN='\033[0;36m'
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
echo -e "  ${GOLD}1${RESET}  feat      — new feature or content          → ### Added in release notes"
echo -e "  ${GOLD}2${RESET}  fix       — bug fix                         → ### Fixed in release notes"
echo -e "  ${GOLD}3${RESET}  refactor  — cleanup, no behavior change     → ### Changed in release notes"
echo -e "  ${GOLD}4${RESET}  chore     — build/tooling/deps              (skipped in release notes)"
echo -e "  ${GOLD}5${RESET}  docs      — docs/comments only              (skipped in release notes)"
echo ""
read -rp "  Choice [1-5, default 1]: " TYPE_CHOICE || true
echo ""

case "${TYPE_CHOICE:-1}" in
  2) TYPE="fix" ;;
  3) TYPE="refactor" ;;
  4) TYPE="chore" ;;
  5) TYPE="docs" ;;
  *) TYPE="feat" ;;
esac

# ── Plain-English description — we handle formatting ─────────────────────────
read -rp "  What did you change? (plain English, no prefix needed): " DESC || true
echo ""

# Trim whitespace + trailing period, lowercase first char
DESC="$(echo "$DESC" | sed 's/^[[:space:]]*//' | sed 's/[[:space:]]*$//' | sed 's/\.$//')"
FIRST="$(echo "${DESC:0:1}" | tr '[:upper:]' '[:lower:]')"
DESC="${FIRST}${DESC:1}"

MSG="${TYPE}: ${DESC}"

# ── Preview ───────────────────────────────────────────────────────────────────
echo -e "${GOLD}┌── Commit Message ──────────────────────────────────────────────┐${RESET}"
echo -e "│  ${BOLD}${MSG}${RESET}"
echo -e "${GOLD}└────────────────────────────────────────────────────────────────┘${RESET}"
echo ""

read -rp "  Commit? [Y/n]: " CONFIRM_COMMIT || true
if [[ "${CONFIRM_COMMIT:-y}" =~ ^[Nn] ]]; then
  echo -e "${GRAY}Aborted — unstaging changes.${RESET}"
  git restore --staged .
  exit 0
fi

git commit -m "$MSG"
echo ""

# ── Push ──────────────────────────────────────────────────────────────────────
read -rp "  Push to origin/master now? [Y/n]: " CONFIRM_PUSH || true
if [[ "${CONFIRM_PUSH:-y}" =~ ^[Nn] ]]; then
  echo -e "${GREEN}✓ Committed locally.${RESET} Push later with: git push origin master"
  exit 0
fi

git push origin master
echo ""
echo -e "${GREEN}✓ ${BOLD}${MSG}${RESET}"
