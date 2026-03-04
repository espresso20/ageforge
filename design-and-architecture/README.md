# AgeForge — Design & Architecture

This directory is the authoritative source for design decisions, game laws, and architecture
principles for AgeForge. When in doubt about why something works the way it does, or whether
a proposed change would break the game's balance, start here.

## Documents

| File | What it covers |
|------|---------------|
| [economy.md](economy.md) | Economy laws, cost scaling, production model, worker-building coupling |

## How to Use These Documents

- **Before changing any number** (cost, rate, storage, scale factor) — read the relevant doc
  and verify the change doesn't violate a Law or Covenant.
- **When designing new content** (new age, building, resource) — use the formulas here to
  derive costs and rates rather than guessing.
- **When a design decision is made in a session** — add it here so future sessions don't
  relitigate settled questions.

## Decision Log

Decisions recorded here are settled. They can be revisited but require explicit reasoning.

| Date | Decision | Rationale |
|------|----------|-----------|
| 2026-03-03 | Workers couple to buildings (Philosophy B) | Preserves assignment mechanic, creates unified production chain |
| 2026-03-03 | Remove MaxCount from production/housing buildings | Geometric cost scaling is the natural cap |
| 2026-03-03 | Keep MaxCount on storage buildings and wonders | Unlimited storage breaks resource pressure; wonders are unique by design |
| 2026-03-03 | Storage cap ≥ 2× next affordable building cost at all times | Prevents the impossible-to-build problem |
| 2026-03-03 | Prestige loop target: weeks of real-time play | Long-running idle game; prestige unlocks faster paths and higher age tiers |
| 2026-03-03 | Save directory resolves relative to binary, not CWD | Prevents save files appearing in unexpected locations |
