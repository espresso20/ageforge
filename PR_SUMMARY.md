# PR Summary — feat/espresso/game-balances--ages-and

## Overview

This PR covers a full balance pass across all 22 ages, a major worker system refactor, new UI overlays, milestone fixes, event system improvements, and a cleanup of experimental visual systems that were removed.

---

## Game Balance

### Age-by-Age Balance Pass (All 22 Ages)
- **Primitive / Stone Age** — gathering_camp rate 0.05→0.50/tick, build time 40→12 ticks; forager_post rate 0.10→1.00/tick; hut build time 80→8 ticks; worker food drain reduced to 0.08/tick
- **Bronze / Iron Age** — full rate, build time, and food cost pass across all domain lineages
- **Classical / Medieval** — rate and build time tuning to ensure meaningful progression gates
- **Renaissance → Quantum** — full rate pass, ensuring late-game pacing feels rewarded not stalled
- stone_camp moved to stone_age (stone not available in primitive); hunting_lodge moved to iron_age

### Building Production Floor
- Buildings with zero workers now produce 0 (was producing at base rate regardless of staffing)
- Worker scaling formula: `base × count × (0.20 + 0.80 × assigned / totalCap)`

---

## Worker System (Full Refactor)

- Renamed VillagerManager → WorkerManager; `villagers` → `workers` throughout (save JSON tag preserved as `"villagers"` for backward compat)
- Collapsed 12 separate domain worker pools into a **single generic worker pool**
- `recruit [count|max]` — no domain arg; workers become a domain class when assigned to a building
- `assign <building> [count|all]` / `unassign <building> [count|all]` — domain auto-resolved from building definition
- `AssignWorker` now blocks assignment to unbuilt buildings (Count == 0) with a clear error
- Removed all backward-compat shims and `resolveDomain()` aliases
- `config.WorkerClassByDomainAndAge(domain, ageKey)` provides class name, food cost, and multiplier per age

---

## UI Changes

### Overlay Panel System
- Replaced F-key tab navigation with command-summoned floating overlay panels
- New overlays: `milestones` / `ms`, `research` / `res`, `army`, `trade`, `stats`, `wonders`, `logs`, `epoch`
- ESC closes any open overlay panel

### Worker Panel
- Per-domain assignment groups with capacity indicators
- Class name display resolves from current age + domain (e.g. "Farmer" in stone age, "Agrarian" in bronze)
- Assignment UI shows assigned/capacity per building

### Resource Bars
- Fill-level colour coding (green → yellow → red)
- Status glyphs for near-cap and depleted states

### Age Splash Screen
- Era-aware city label on age advancement splash

### Research Panel
- Research available as both a command and a summoned overlay panel

---

## Milestone & Event Fixes

- `master_builder` and `scholars_haven` milestones wired to correct game state fields
- `scholars_haven` now requires 5 knowledge workers assigned (was firing incorrectly)
- `war_machine` milestone no longer fires in primitive age
- Worker loss event (`worker_loss` effect) implemented with damage summary in ended message

---

## Visual Maps — Removed

Experimental visual map systems (Ebiten-rendered skyline, civ map, city map) were built and iterated across multiple cycles but did not meet quality standards. All visual rendering code has been removed:

- Removed: `visual/` package (Ebiten subprocess renderer)
- Removed: `ui/civmap.go`, `ui/citymap.go`, and all associated terrain/building render files
- Removed: `ui/tab_civ.go`, `ui/tab_map.go` visual render logic, `ui/tab_livingmap.go`
- Removed: `ui/visual_launcher.go`, `ui/citypulse.go`, `ui/draw_primitives.go`
- Removed: `cmd/mapdemo/` directory
- Removed: `civ`, `city`, `capital` commands from input handler and autocomplete

> Visual city representations may be revisited in a future PR with a different approach.

---

## Documentation & Code Quality

- Inline comments added across `game/`, `ui/`, and `config/` packages
- GoDoc-style doc comments on all exported types and functions
- `// NOTE:` / `// IMPORTANT:` annotations on locking patterns and bus handler constraints
- Wiki (`site/docs/`) audited and synced to match current game mechanics, commands, and balance values
- First-ten-minutes tutorial rewritten to match current balance and worker command syntax

---

## Bug Fixes

- Ticker freeze (3 bugs): stopCh not reset on restart; getTickInterval divide-by-zero guard; Bus replaced on reset losing dashboard subscriptions
- Wipe/prestige dashboard flash: atomic page replace instead of remove-then-add
- Worker assignment: blocked assignment to buildings with Count == 0
- Gather cap typo: `amount = 10000` corrected to `amount = 10`
- Worker panel: class name always resolves correctly for any domain + age combination
- Map tab: SetDrawFunc moved from Refresh() to constructor (was re-registering on every tick)

---

## Stats (current)

| Metric | Count |
|---|---|
| Ages | 22 |
| Buildings | 284 (241 production across 13 lineages + 21 storage + 22 wonders) |
| Techs | 52 |
| Milestones | 33 |
| Resources | 25 |
| Worker Domains | 12 |
| Epochs | 7 |
| Expeditions | 15 |
| Trade Routes | 15 |
| Diplomacy Factions | 6 |
| Prestige Upgrades | 9 |
