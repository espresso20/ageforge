#!/usr/bin/env bash
set -euo pipefail

# Usage: bash scripts/release.sh [patch|minor|major]
BUMP="${1:-patch}"

# ── Validation ────────────────────────────────────────────────────────────────

BRANCH="$(git rev-parse --abbrev-ref HEAD)"
if [[ "$BRANCH" != "master" ]]; then
  echo "error: must be on master (currently on '$BRANCH')" >&2
  exit 1
fi

if ! git diff --quiet || ! git diff --cached --quiet; then
  echo "error: working tree is dirty — commit or stash changes first" >&2
  exit 1
fi

# ── Compute new version ───────────────────────────────────────────────────────

CURRENT="$(git describe --tags --abbrev=0 2>/dev/null || echo "v0.0.0")"
# Strip leading 'v'
RAW="${CURRENT#v}"
MAJOR="$(echo "$RAW" | cut -d. -f1)"
MINOR="$(echo "$RAW" | cut -d. -f2)"
PATCH="$(echo "$RAW" | cut -d. -f3)"

case "$BUMP" in
  major) MAJOR=$((MAJOR + 1)); MINOR=0; PATCH=0 ;;
  minor) MINOR=$((MINOR + 1)); PATCH=0 ;;
  patch) PATCH=$((PATCH + 1)) ;;
  *)
    echo "error: bump must be patch, minor, or major (got '$BUMP')" >&2
    exit 1
    ;;
esac

NEW_VERSION="v${MAJOR}.${MINOR}.${PATCH}"
TODAY="$(date +%Y-%m-%d)"

echo "Releasing ${NEW_VERSION} (${TODAY})..."

# ── Update CHANGELOG ──────────────────────────────────────────────────────────

CHANGELOG="CHANGELOG.md"

if ! grep -q "^## \[Unreleased\]" "$CHANGELOG"; then
  echo "error: CHANGELOG.md has no '## [Unreleased]' section" >&2
  exit 1
fi

# Replace "## [Unreleased]" with the versioned header, then prepend a fresh
# [Unreleased] block at the top of the entries section.
python3 - "$CHANGELOG" "$NEW_VERSION" "$TODAY" <<'PYEOF'
import sys, re

path, ver, today = sys.argv[1], sys.argv[2], sys.argv[3]
with open(path) as f:
    content = f.read()

# Replace first [Unreleased] header with versioned header
content = content.replace(
    "## [Unreleased]",
    f"## [Unreleased]\n\n---\n\n## [{ver}] — {today}",
    1
)

with open(path, "w") as f:
    f.write(content)
PYEOF

echo "  Updated $CHANGELOG"

# ── Commit, tag, push ─────────────────────────────────────────────────────────

git add "$CHANGELOG"
git commit -m "chore: release ${NEW_VERSION}"
git tag -a "${NEW_VERSION}" -m "Release ${NEW_VERSION}"

git push origin master
git push origin "${NEW_VERSION}"

echo ""
echo "  Released ${NEW_VERSION}"
echo "  GitHub Actions: https://github.com/espresso20/ageforge/actions"
echo "  Releases:       https://github.com/espresso20/ageforge/releases/tag/${NEW_VERSION}"
