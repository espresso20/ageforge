# AgeForge — Completed Work

Items are moved here from TODO.md when finished. Do not re-implement anything listed here.

---

## Prestige
- [x] Recalibrated prestige shop costs to ~277 total points (fills in ~70-80 perfect runs, ~100 normal runs) (2026-02-25)
  - Starting resources (food/wood): [1, 2, 3, 4, 5] = 15 each
  - Base upgrades (gather/storage): [2, 3, 4, 6, 8] = 23 each
  - Mid upgrades (research/military/population/expedition): [2, 3, 5, 8, 10] = 28 each
  - Temporal Mastery: [6, 10, 17, 23, 33] = 89 (32% of total shop cost)
- [x] Locked prestige behind Modern Age (order 12, was Medieval order 5) — updated CanPrestige() and tests (2026-02-25)

## Wonder System
- [x] Replaced broken cost display with per-wonder resource bank system (2026-02-25)
  - `BuildingManager.wonderBanks` — per-wonder resource ledger, separate from player storage
  - `wonder collect <res> <amt|all>` — deducts from storage into bank immediately
  - `build <wonder_key>` — gated on `IsWonderBankFull()`, no second resource deduction
  - Bank fill shown in Wonder panel (progress bar per resource) and Wonders tab (% banked)
  - Banks saved/loaded in GameSave; auto-clear on prestige/reset via NewBuildingManager()
  - Fixed prestige gate error message: "Medieval" → "Modern Age"
  - Fixed 4 stale engine tests from prior gameplay rebalancing (starting resources, gather, research costs)

## Map
- [x] Wonders placed in isolated far outer zone (60-78% of maxDist) away from city; infrastructure auto-draws a road out to them (2026-02-25)
- [x] City spread factor: rings expand when totalCity > 40 buildings (up to 1.8x at 240+), capped at 55% to preserve wonder gap (2026-02-25)
- [x] Era-specific layout patterns in placeBuildingsRadial (2026-02-25):
  - Era 0 (Primitive): tight organic scatter, compressed rings, vc=0.65
  - Era 1-2 (Ancient/Medieval): classic loose radial baseline
  - Era 3 (Industrial): 8-direction grid snap, buildings line up in rows
  - Era 4-5 (Modern/Digital): 12-sector clock-face radial
  - Era 6 (Cyberpunk): dense compressed rings, vc=0.90 (tall)
  - Era 7-8 (Space/Cosmic): perfect orbital circles evenly spaced by count, vc=0.85

## Villager Dashboard
- [x] Added VillagerPanel to bottom area of dashboard (2026-02-25)
  - `ui/villager_panel.go` — new `VillagerPanel` struct alongside WonderPanel + MiniMap
  - Shows all unlocked villager types in canonical definition order
  - Per-type: count, idle count, current assignments (resource:count), and gather rate
  - Effective rate shown with arrow (base → effective) when bonuses are active
  - Bonuses section: research gather_rate, prestige gather_boost, and prestige passive (prod.all)
  - Hash-based dirty check — only redraws on actual state change
  - Wired into `dashboard.go` as 4th column in `bottomArea` (log | villagers | wonder | minimap)

## Splash Screen Redesign
- [x] Rebuilt main menu splash screen (2026-02-25)
  - Left panel: procedural Civ-1 style Greek temple pixel art (400×250 image.RGBA, tview.TrueColor)
    - Sky gradient, sun with glow, 3 fluffy elliptical clouds, textured green/dirt ground
    - Temple: 3-step staircase, 6 Doric columns (capital + shaft + base), entablature with triglyphs/metopes, triangular pediment with raking cornice, acroterion at apex
  - Right panel: `tview.List`-based vertical menu with two groups:
    - Primary: **Load Game** (first/default), New Game, Wiki
    - Danger: Wipe Save, Quit (dark-red highlight)
    - Tab switches between primary and danger lists; arrow keys navigate each
  - Load Game button grayed label `(no save)` if no autosave exists
  - Wiki entry calls `dashboard.GoToWiki()` (switches to Wiki tab, starts engine)
  - Prestige level shown in title area when > 0
  - `showWipeConfirmation` now threaded through `onWiki` closure so rebuilt splash preserves wiki callback
  - New helpers: `splashFill`, `splashCloud`, `renderSplashTemple`
  - `app.go`: passes `onWiki` callback to CreateSplashPage; creates dashboard before splash
  - `dashboard.go`: added exported `GoToWiki()` method

## General / Misc
- [x] Established TODO.md + DONE.md workflow for session-resumable planning (2026-02-25)
