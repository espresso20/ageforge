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
| 2026-03-03 | 88 total events: 28 universal + 35 epoch-exclusive + 10 good epoch + 8 bad epoch + 7 catastrophe | Event pool shifts each epoch; transition events are separate from regular random pool |
| 2026-03-03 | Catastrophe is NOT guaranteed every epoch — it's ~15% per epoch transition | Players would rage if forced into Endure/Succumb 7 times; rarity makes it feel special |
| 2026-03-03 | Faith influences epoch event roll (40/50/60% good chance) | Meaningful new use for faith investment beyond morale/diplomacy |
| 2026-03-03 | Culture gates good event tier (minor/major/legendary) | Meaningful new use for culture investment |
| 2026-03-03 | Endure: 20% buildings destroyed (down from 30%) | 30% felt punishing; 20% is painful but survivable |
| 2026-03-03 | Voluntary catastrophe always available via Epoch tab | Players who want Legacy Bonuses don't have to wait for RNG |
| 2026-09-26 | Catastrophe overhaul: Defer removed; a pending catastrophe blocks advancing and prestige (Esc + `catastrophe` reopen the choice); no catastrophes before the Iron epoch; `catastrophe invoke` removed in favour of the Harbinger's upcoming "Invite it" (dev console `/catastrophe` remains for testing); Succumb's +25% research speed stacks per distinct epoch and is permanent (derived from legacy flags); ruins capped at 24, lowest value dropped first | Defer let the choice be skipped and a later roll overwrite it; Stone Era invoke→Succumb was a free loop, and choosing a catastrophe fits better as a Harbinger decision than a bare command; the research bonus was silently wiped by the next reset; the choice should be deliberate and its rewards should stick |
| 2026-09-26 | Harbinger replaces invoke: each epoch whose transition can bring a catastrophe (Stone to Neon) has a harbinger thread from its first age (or first tick of a new run) to the transition, and each age's roster figure takes up the warning in turn (18 of 22 figures appear; the Cosmic Era's four never do); answers carry across figures; false prophets rolled once per thread at its first figure (Primitive 8/64, Iron 5/64, Renaissance 2/64, else 0), the claim kept as a fixed multiple of the real chance; Appease ×0.6 per level (2 levels, 15%/30% of the passage storage in faith, plus culture from the Steel Era); Brace 12%/24% of the most the epoch still asks of each core resource, tiers 15%/30% and 10%/45% (buildings lost / stock kept on Endure); Invite arms the persisted `catastropheInvited`. See epochs.md → Harbinger | Choosing a catastrophe fits better as an answer to a warning than as a bare command; a thread gives the warning time to matter; false prophets make early warnings worth doubting until the odds are printed; pricing off the passage keeps the price the same in every age of the epoch, so paying early is not a discount; ×0.6 still lowers the odds after the worst faith-band drop the price can cause |
| 2026-09-26 | Prestige is the Cosmic Era's passage (the Last Passage); Endure keeps 50/70/85% of the run's points by Brace; Succumb grants a one-time Cosmic Legacy (+10% production). See epochs.md → The Last Passage | The Cosmic Era has no transition out, so its four harbingers had nothing to warn of; prestige is its only passage, and points are what a prestige can lose, so Endure and Brace act on them; the Cosmic Legacy is derived from a flag, like Ancient Knowledge, so no reset can drop or double it |
| 2026-09-26 | Gate Covenant: for every advance, against the most storage buildable by the end of the age you leave and with no build_cost discounts, (1) every required building is buildable in that age, (2) its last required copy costs at most half that storage, (3) each resource requirement is at most 80% of it, (4) every resource the gate needs has a source that does not need it first (a building that doesn't cost it, hand gathering, an exchange, or tech output covering the amount within 48h). No building may cost a resource with no source in its own age. Levers in order: retarget to the lineage's building in that age, raise storage where the age is out of band (median first copy above 3–8% of max storage), lower the count (multiple of 5), drop the unobtainable resource from the bootstrap producer. Storage never transforms. Enforced by `TestGateCovenant` (smoke/static_test.go). See economy.md | "Storage ≥ 2× the next affordable building" was unenforced and every gate from Stone onward had drifted out of it: the normalized 1.15 curves made copy #50–500 unaffordable at any storage, the age lock made requirements on older buildings traps, and several ages' core resources had no way in. A mechanical rule plus a unit test keeps balance passes from reopening the same soft-lock |
