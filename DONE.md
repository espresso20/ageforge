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

## HTTP Wiki Server
- [x] Embedded localhost HTTP wiki server (port 7891) accessible from splash screen (2026-02-25)
  - `game/wiki_server.go` — `WikiServer` struct with `Start()`/`Stop()`/`IsRunning()`/`URL()`/`OpenBrowser()`
  - `OpenBrowser()` uses `os/exec` + `runtime.GOOS` switch (darwin=`open`, windows=`cmd/start`, linux=`xdg-open`)
  - 10 HTTP routes: `/`, `/ages`, `/buildings`, `/resources`, `/techs`, `/events`, `/milestones`, `/prestige`, `/villagers`, `/trade`
  - All data read from `config.*` — works without starting the game engine
  - Dark-themed HTML with gold nav bar, badge-colored effects, grouped sections
  - `ui/app.go` — `App.wikiServer` field, initialized in `setup()`, stopped in `Stop()`
  - `ui/splash.go` — Wiki button calls `wikiServer.Start()` + `wikiServer.OpenBrowser()` (no engine start from splash)
  - `showWipeConfirmation` and `CreateSplashPage` now take `*game.WikiServer` instead of `onWiki func()`

## Web Port — Phase 1 (Repo Restructure)
- [x] Split game/save.go into 3 files (2026-02-25):
  - `save_common.go` — GameSave structs + `buildSaveSnapshot` + `applySaveState` (no build tag, shared)
  - `save_native.go` — file I/O backend (`//go:build !js`)
  - `save_wasm.go` — localStorage backend (`//go:build js && wasm`)
- [x] `game/wiki_server.go` — added `//go:build !js` tag
- [x] `cmd/wasm/main.go` — WASM entry point; exports `init`, `start`, `stop`, `getState`, `command`, `loadSave`, `exportSave`, `importSave`, `saveExists`, `wipeSaves` to JS; full command dispatcher for all engine actions
- [x] `web/index.html` — full app shell: loading screen, top bar, 9-tab nav, modal overlay, toast container
- [x] `web/style.css` — complete dark theme: CSS vars, tab bar, tables, badges, cards, progress bars, toasts, modal
- [x] `web/game.js` — WASM loader, tick loop (1s), autosave (60s), all 9 tab renderers, command dispatcher, save/export/import UI, wonder bank panel
- [x] `web/map.js` — PixiJS v8 map renderer skeleton (placeholder until Phase 6 mapgen extraction)
- [x] `web/manifest.json` — PWA manifest (ageforge.io)
- [x] `web/sw.js` — service worker for offline + WASM caching
- [x] `Makefile` — added `make web` and `make web-clean` targets; `GOOS=js GOARCH=wasm go build ./cmd/wasm/`
- [x] `netlify.toml` — build command, publish dir, WASM MIME header, security headers, SPA redirect
- [x] `.go-version` — pins Go 1.23 for Netlify build image

## Economy Redesign — Phase 5: Config Foundation (2026-03-04)
- [x] Added 6 new fields to `BuildingDef` struct: `LineageKey`, `LineageTier`, `WorkerDomain`, `WorkerCapacity`, `EpochKey`, `OutputResource`
- [x] Added `EpochKey` field to `AgeDef`; all 22 ages populated with correct epoch key
- [x] `config/resources.go`: added 5 new resources (`marble`, `iron_ore`, `nanobots`, `titanium_ore`, `dark_matter_crystals`); fixed faith unlock age (medieval → primitive); fixed culture unlock age (renaissance → classical) — total now 25 resources
- [x] `config/epochs.go` (NEW): `EpochDef` struct + all 7 epoch definitions + `EpochByKey()` + `EpochForAge()` helper
- [x] `config/workers.go` (NEW): `WorkerClassDef` struct + 189 worker classes across 12 domains × up to 21 tiers; `WorkerClassByDomainAndAge()` + `WorkerDomains()` helpers; geometric scaling (×1.5 food, ×2.0 multiplier per tier)
- [x] `config/buildings.go`: renamed `BaseBuildings()` → `baseBuildingsRaw()`; added `buildingMeta()` map covering all 111 buildings; new `BaseBuildings()` merges metadata in — all buildings now carry lineage/domain/epoch/output data
- Note: 111 buildings total (89 standard + 22 wonders), not 80 as previously counted

## Economy Redesign — Phase 6: Worker-Building Coupling Engine (2026-03-05)
- [x] `game/villagers.go`: MAJOR REWRITE — domain-based system replacing 8-type resource-centric model
  - 12 worker domains (food/faith/knowledge/military/trade/engineering/hacker/astronaut/lumber/masonry/metallurgy/energy)
  - `domainRuntime` with `count` + `assignments map[string]int` (buildingKey → worker count)
  - `Assign(domain, buildingKey, count)` — validates building WorkerDomain matches
  - `FoodDrain()` uses WorkerClassDef food costs from config (geometric tier scaling)
  - `GetAssignedCount(domain, buildingKey)` — used by production formula
  - `GetDomainCount(domain)` — replaces direct `.types["soldier"].count` access
  - `SetAge(ageKey)` — updates WorkerClassDef lookups on age change
  - Legacy alias map: "worker"→"food", "shaman"→"faith", etc. for backward compat
  - `Snapshot()` keys by legacy type names for UI compat (domainToLegacy map)
  - `LoadVillagers()` handles both new (buildingKey) and old (resource key) save formats
  - `DefaultVillagerTypes()` + `VillagerTypeDef` kept for villager_panel.go backward compat
- [x] `game/types.go`: Added `WorkerDomain`, `WorkerCapacity`, `WorkersAssigned` to `BuildingState`
- [x] `game/buildings.go`: Added `WorkerScaledProduction(getAssigned func)` — computes per-resource
  rates with formula `base × count × (0.20 + 0.80 × assigned/totalCap)` (20% floor);
  updated `Snapshot(resources, getWorkerCount)` to populate worker fields
- [x] `game/engine.go`:
  - `recalculateRates()` now calls `Buildings.WorkerScaledProduction(Villagers.GetAssignedCount)`
  - `advanceAge()` calls `Villagers.SetAge(newAge)` for food cost / class name updates
  - `AssignVillager/AssignAll/UnassignAll/UnassignVillager` — all updated to (domain, buildingKey)
  - Removed all `ge.Villagers.types[...]` direct field accesses; replaced with `GetDomainCount`
  - `GetState()`: `Buildings.Snapshot` now passes `Villagers.GetAssignedCount`
  - Tutorial log updated to reference new assign syntax
- [x] `game/save.go`: `LoadGame()` calls `Villagers.SetAge(save.Age)` after restoring age
- [x] `game/villagers_test.go`: Rewrote all tests for domain+buildingKey semantics;
  added `TestVillagerManager_GetAssignedCount`, `TestVillagerManager_AssignWrongDomain`,
  `TestVillagerManager_LegacyLoadCompat`
- [x] `game/engine_test.go`: Updated `TestEngine_AssignVillager` to use new API
- [x] `ui/input.go`: Updated `cmdAssign`/`cmdUnassign` to use domain+building syntax; updated help text
- All tests passing, `go build ./...` clean

## Economy Redesign — Phase 7: Age Transition Transformation Pass (2026-03-04)
- [x] `config/buildings.go`: Added `BuildingNextTierForAge(lineageKey, currentTier, newAgeKey)` — returns next-tier BuildingDef or nil; all 16 known lineage→age transitions verified correct
- [x] `game/types.go`: Added `AgeAdvanceSummary`, `BuildingTransform` structs; added `IsLegacy bool` to `BuildingState`; added `LastAgeAdvanceSummary AgeAdvanceSummary` to `GameState`
- [x] `game/villagers.go`: Added `RenameAssignment(domain, oldKey, newKey)` — transfers building-keyed worker assignments to new key; called by TransformBuilding
- [x] `game/buildings.go`: Added `legacyBuildings map[string]bool` to `BuildingManager`; added `MarkLegacy`, `IsLegacy`, `GetLegacyBuildings`, `LoadLegacyBuildings`, `TransformBuilding`; updated `Snapshot` to set `IsLegacy=true` and `CanBuild=false` for legacy buildings
- [x] `game/engine.go`: Added `lastAgeAdvanceSummary` field; transformation pass in `advanceAge()`:
  - Collects all (lineage, tier, newAge) matches into a pending list (safe from map mutation)
  - Applies `TransformBuilding(old, new, RenameAssignment)` for each match
  - Legacy-marks buildings whose lineage has a higher-tier unlocked equivalent
  - Logs each transformation as a "success" entry
  - Exposes `LastAgeAdvanceSummary` in `GetState()`
- [x] `game/save.go`: Added `LegacyBuildings []string` to `GameSave`; serialized in `buildSaveSnapshot`; restored in `LoadGame`
- Note: UI age advance summary modal deferred to Phase 11 (UI Completion)
- Note: Worker class rename/restat already handled by `SetAge()` from Phase 6 — no additional villagers.go changes needed
- All tests passing, `go build ./...` clean

## Economy Redesign — Phase 8: Epoch System (2026-03-04)
- [x] `config/epochs.go`: Added `EpochEventDef` struct; `GoodEpochEvents()` (10 events: 5 minor, 4 major, 1 legendary); `ChallengingEpochEvents()` (8 events). Event definitions are pure data (Key, Name, FlavorText, Type, Duration only); all application logic in engine.go.
- [x] `game/bus.go`: Added `EventEpochAdvanced` and `EventEpochEventFired` event constants
- [x] `game/types.go`: Added `EpochEventRecord` struct; added EpochKey, EpochName, EpochIcon, EpochColor, EpochSurvived, PendingCatastrophe, EpochEventHistory to `GameState`
- [x] `game/villagers.go`: Added `AddPctAll(pct)` — instant worker count boost across all domains; `RemovePct(pct)` — proportional removal (epidemic)
- [x] `game/buildings.go`: Added `DestroyRandom(count)` — destroys N random non-wonder buildings; added `math/rand` import
- [x] `game/research.go`: Added `ForceCompleteN(n, age, ageOrder)` — completes up to N techs from current age (Grand Discovery event)
- [x] `game/engine.go`:
  - Added epoch fields: `currentEpoch`, `epochEventFired`, `survivedEpochs`, `pendingCatastrophe`, `epochEventHistory`
  - Added `math/rand` import
  - `NewGameEngine()` and `Reset()` initialize epoch fields
  - `recalculateRates()` now sums `production_all` effects from active events (epoch boosts)
  - `advanceAge()` calls `detectEpochTransition()` after age change
  - `detectEpochTransition(newAge)` — detects epoch change, logs announcement, fires bus event, calls `rollEpochEvent()`
  - `rollEpochEvent(epochKey)` — faith-gated good/bad roll (40/50/60% good); bad: 70% challenging, 30% catastrophe
  - `rollGoodEpochEvent()` — culture-gated tier selection (minor/major/legendary), picks from pool, calls `applyGoodEpochEvent()`
  - `rollChallengingEpochEvent(epochKey)` — random pick from challenging pool, calls `applyChallengingEpochEvent()`
  - `applyGoodEpochEvent(ev)` — 10 cases: age_of_plenty (+production_all timed), population_surge (+15% workers), ancient_cache (fill 40% storage), trade_winds (gold boost), cultural_festival (culture+faith instant+timed), grand_discovery (3 free techs), worker_innovation (+10% permanent), architects_gift (10 free buildings), peaceful_century (+20% timed), epoch_blessing (+15% permanent)
  - `applyChallengingEpochEvent(ev, epochKey)` — 8 cases: famine (food debuff), merchant_betrayal (steal gold+debuff), great_fire (destroy 8 buildings), epidemic (remove 20% workers+debuff), resource_drought (primary resource debuff), political_instability (steal faith+knowledge debuff), economic_crash (steal gold+debuff), dark_age (cancel research+steal knowledge+debuff)
  - `InvokeCatastrophe() error` — voluntary trigger; errors if already fired this epoch; sets pendingCatastrophe
  - `GetState()` exposes all epoch fields
- [x] `game/save.go`: Added CurrentEpoch, EpochEventFired, SurvivedEpochs, PendingCatastrophe, EpochEventHistory to `GameSave`; serialize in `buildSaveSnapshot()`; restore in `LoadGame()`; added `copyBoolMap()` helper; added config import
- [x] `ui/dashboard.go`: Epoch badge in statusTV (icon + color + "·Survived" marker); subscribed to EventEpochAdvanced (toast 6s); subscribed to EventEpochEventFired (toast colored by event type 6s)
- Note: Catastrophe Endure/Succumb modal deferred to Phase 9
- Note: `survivedEpochs` tracking deferred to Phase 9 (Endure action)
- All tests passing, `go build ./...` clean

## General / Misc
- [x] Established TODO.md + DONE.md workflow for session-resumable planning (2026-02-25)
