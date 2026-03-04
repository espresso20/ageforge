# AgeForge — Design & Architecture

This directory is the authoritative source for design decisions, game laws, and architecture
principles for AgeForge. When in doubt about why something works the way it does, or whether
a proposed change would break the game's balance, start here.

## Documents

| File | What it covers |
|------|---------------|
| [economy.md](economy.md) | Economy laws, cost scaling, production model, worker-building coupling |
| [workers.md](workers.md) | All 12 worker domains, age-tiered class names, food costs, output multipliers |
| [age-transitions.md](age-transitions.md) | Age advance transformation pass — building lineages, worker renames, legacy rules, UI summary |
| [lineages.md](lineages.md) | All 13 building lineages — full 21-tier tables, storage buildings, wonders policy |
| [resources.md](resources.md) | All 25 resources — faith mechanics (draining), culture mechanics (accumulating), epoch resource chain, 2-stage processing chain |
| [epochs.md](epochs.md) | 7 epochs × 3 ages, resource transitions per epoch, Civilizational Catastrophe system (Endure vs Succumb), 63 total events across 7 epoch pools, UI epoch badge |

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
| 2026-03-03 | Age-tiered worker classes (Gatherer→Serf→Drone→Harvester etc.) | Flavor + mechanical progression; higher tiers cost more food but produce significantly more |
| 2026-03-03 | Age transition transformation pass | Buildings upgrade in-place (count preserved), workers rename on age advance — civilisation feels like it advances rather than just unlocking more |
| 2026-03-03 | Wonders never transform, storage buildings don't transform (cumulative) | Wonders are landmarks; storage is additive infrastructure |
| 2026-03-03 | Legacy buildings: no next-tier = stays functional, grayed, unbuildable | Player is never punished by losing production they invested in |
| 2026-03-03 | Remove MaxCount from production/housing buildings | Geometric cost scaling is the natural cap |
| 2026-03-03 | Keep MaxCount on storage buildings and wonders | Unlimited storage breaks resource pressure; wonders are unique by design |
| 2026-03-03 | Storage cap ≥ 2× next affordable building cost at all times | Prevents the impossible-to-build problem |
| 2026-03-03 | Prestige loop target: weeks of real-time play | Long-running idle game; prestige unlocks faster paths and higher age tiers |
| 2026-03-03 | Save directory resolves relative to binary, not CWD | Prevents save files appearing in unexpected locations |
| 2026-03-03 | Split Raw Materials into three separate domains: Food, Lumber, Masonry | Each domain fills its own lineage; clear coupling between worker and building type |
| 2026-03-03 | Faith and Knowledge are separate domains with separate worker chains | Shaman is a faith leader, not a scholar; each domain drives distinct gameplay mechanics |
| 2026-03-03 | Culture/Arts lineage has no worker domain — auto-produces passively | Culture is a civilization ambient stat, not an assigned-labor product |
| 2026-03-03 | Culture accumulates permanently (20% persists through prestige), thresholds unlock permanent bonuses | Creates a long-term civilization identity investment separate from prestige resets |
| 2026-03-03 | Faith is a draining resource (must maintain); gates morale, cohesion, diplomacy, prestige multiplier | Meaningful idle management loop; neglecting faith has real costs |
| 2026-03-03 | 13 building lineages, 12 worker domains | Final counts — adding new content requires explicit justification |
| 2026-03-03 | 7 epochs (3 ages each) as meta-progression layer above ages | Cleaner than per-age transitions; 7 epochs × 3 ages = 21 exactly |
| 2026-03-03 | Organic Extraction and Geological Extraction change output resource per epoch | "Lumber" is not always wood; the role evolves. Building transforms name AND output at epoch boundary. |
| 2026-03-03 | 2-stage processing chain everywhere (Geological ore → Metallurgy → refined metal) | Adds satisfying supply chain optimization; every epoch has a bottleneck to balance |
| 2026-03-03 | 25 resources total (added marble, iron_ore, titanium_ore, dark_matter_crystals) | Intermediate ores enable 2-stage chains without exposing them as build costs |
| 2026-03-03 | Civilizational Catastrophe system: Endure vs Succumb at each epoch boundary | Narrative-driven alternative/complement to prestige; 7 catastrophes, epoch-specific Legacy Bonuses, 8 Ruins carry forward on Succumb |
| 2026-03-03 | 63 total events: 28 universal + 5 epoch-exclusive per epoch | Event pool shifts each epoch; events feel contextually appropriate to the era |
