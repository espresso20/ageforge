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

# ── Scrape commits since last tag ─────────────────────────────────────────────
# Group conventional commits into sections for the release notes.

COMMITS="$(git log "${CURRENT}..HEAD" --pretty=format:"%s" --no-merges 2>/dev/null || true)"

python3 - "$COMMITS" <<'PYEOF'
import sys, re

raw = sys.argv[1]
lines = [l.strip() for l in raw.splitlines() if l.strip()]

sections = {"feat": [], "fix": [], "refactor": [], "perf": [], "other": []}
pattern = re.compile(r'^(feat|fix|refactor|perf|chore|docs|style|test)(\(.+?\))?!?:\s*(.+)$', re.IGNORECASE)

for line in lines:
    m = pattern.match(line)
    if m:
        kind = m.group(1).lower()
        msg  = m.group(3)
        if kind in ("feat",):
            sections["feat"].append(msg)
        elif kind in ("fix",):
            sections["fix"].append(msg)
        elif kind in ("refactor", "perf"):
            sections["refactor"].append(msg)
        elif kind in ("chore", "docs", "style", "test"):
            pass  # skip meta commits from release notes
        else:
            sections["other"].append(line)
    else:
        # non-conventional — keep if not a chore/version bump
        if not re.match(r'^chore:', line, re.IGNORECASE):
            sections["other"].append(line)

out = []
labels = [("feat", "### Added"), ("fix", "### Fixed"), ("refactor", "### Changed"), ("other", "### Other")]
for key, header in labels:
    if sections[key]:
        out.append(header)
        for item in sections[key]:
            out.append(f"- {item}")
        out.append("")

print("\n".join(out).strip())
PYEOF

# Capture the generated notes
NOTES="$(python3 - "$COMMITS" <<'PYEOF'
import sys, re

raw = sys.argv[1]
lines = [l.strip() for l in raw.splitlines() if l.strip()]

sections = {"feat": [], "fix": [], "refactor": [], "other": []}
pattern = re.compile(r'^(feat|fix|refactor|perf|chore|docs|style|test)(\(.+?\))?!?:\s*(.+)$', re.IGNORECASE)

for line in lines:
    m = pattern.match(line)
    if m:
        kind = m.group(1).lower()
        msg  = m.group(3)
        if kind == "feat":
            sections["feat"].append(msg)
        elif kind == "fix":
            sections["fix"].append(msg)
        elif kind in ("refactor", "perf"):
            sections["refactor"].append(msg)
    else:
        if not re.match(r'^chore:', line, re.IGNORECASE):
            sections["other"].append(line)

out = []
labels = [("feat", "### Added"), ("fix", "### Fixed"), ("refactor", "### Changed"), ("other", "### Other")]
for key, header in labels:
    if sections[key]:
        out.append(header)
        for item in sections[key]:
            out.append(f"- {item}")
        out.append("")

result = "\n".join(out).strip()
print(result if result else "- Maintenance and improvements.")
PYEOF
)"

# ── Update CHANGELOG ──────────────────────────────────────────────────────────

CHANGELOG="CHANGELOG.md"

if ! grep -q "^## \[Unreleased\]" "$CHANGELOG"; then
  echo "error: CHANGELOG.md has no '## [Unreleased]' section" >&2
  exit 1
fi

python3 - "$CHANGELOG" "$NEW_VERSION" "$TODAY" "$NOTES" <<'PYEOF'
import sys

path, ver, today, notes = sys.argv[1], sys.argv[2], sys.argv[3], sys.argv[4]
with open(path) as f:
    content = f.read()

versioned_block = f"## [Unreleased]\n\n---\n\n## [{ver}] — {today}\n\n{notes}"
content = content.replace("## [Unreleased]", versioned_block, 1)

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
