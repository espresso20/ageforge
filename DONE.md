# AgeForge — Completed Work

Items are moved here from TODO.md when finished. Do not re-implement anything listed here.

---

## Phase 21: Balance Pass + Bug Fix (2026-03-08)

**Code bugs:**
- [x] H-1: Worker production floor restored — unassigned buildings now produce 20% base rate (was 0%)
- [x] H-2: RemovePct truncation fixed — reconciliation loop prevents idle count going negative
- [x] H-5: GenerateRuins now decrements `counts` — no more double production from ruined buildings
- [x] H-6: Wonder bank float epsilon fixed — `remaining <= 0.001` prevents float drift blocking full detection
- [x] Balance #11: Victorian circular dep removed — `oil_derrick: 5` req → `bessemer_plant: 3`
- [x] Balance #17: Global Network description corrected (+30.0 data, not +10.0)

**Balance:**
- [x] Renaissance steel gate: 20,500 → 2,000 (achievable via Steel Forging tech before Foundry unlocks)
- [x] Colonial→Industrial gold: 5,340,000 → 2,500,000; knowledge: 4,125,000 → 2,000,000
- [x] Metallurgy build times: all 18 buildings capped to era norms (max was 5,000,000 ticks → 30,000)
- [x] Quantum age building reqs: 300/200/150 → 120/100/80
- [x] Faith worker class names: 5 pre-medieval entries added (Devotee→Initiate progression)
- [x] Engineering worker class names: 7 pre-victorian entries added (Apprentice→Machinist)

**Map:**
- [x] Terrain noise coarsened: 6×6px cells instead of per-pixel (eliminates TV-static appearance)
- [x] Ground colour variation clamped to [0.3, 0.7] range (no more jarring adjacent-pixel colour jumps)
- [x] Tree density halved, noise coarsened to 3×3px cells (cleaner terrain backdrop)
- [x] F1 → zone map, F2 → living city map keybinds wired
- [x] `livingmap` / `live` / `lmap` commands added

---

## Phase 20: Living City Map System (2026-03-08)
- [x] New zone grid system (`ui/map_zones.go`) — divides canvas into typed zones (food/hearth/commerce/civic/research/military/industry/storage/wonder/wilderness) based on building counts
- [x] 11 scene files (`ui/map_scenes_*.go`) — pixel-art 10×20 ASCII-palette scenes per zone type × era (9 eras), plus wonder scenes (20×40)
- [x] Scene renderer (`ui/map_renderer.go`) — composites zones onto RGBA canvas, applies era palette + level brightness
- [x] `RenderCityMap(state, w, h)` — top-level function: builds zone grid → renders all zones → returns `image.Image`
- [x] `tab_livingmap.go` — animated overlay: goroutine-driven frame counter, particle shimmer, 12fps target
- [x] `citypulse.go` — sidebar panel replacing minimap: era label, city tier, domain activity bar, F1/F2 hint
- [x] `tab_map.go` rewired to use `RenderCityMap` (dropped `GenerateMapImage`/`MapGenConfig`)
- [x] `age_splash.go` updated to use `RenderCityMap`
- [x] `mapHashKey` function (local to tab_map.go, replacing package-level `hashKey`)
- [x] F1 keybind → zone map overlay; F2 keybind → living city map overlay
- [x] Bug fixes: renderer empty-row guard (out-of-bounds), storage scene palette case mismatch (`y`→`Y`)

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

## Economy Redesign — Phase 9: Catastrophe System (2026-03-04)
- [x] `config/epochs.go`: Added `LegacyBonusForEpoch(epochKey) map[string]float64` (per-resource rate bonuses per epoch succumb); `CatastropheInfo(epochKey) (name, flavor string)` for modal text
- [x] `game/types.go`: Added `RuinState` struct; added `RuinCount int` to `BuildingState`; added `LegacyBonuses map[string]bool`, `CatastropheHistory []string` to `GameState`
- [x] `game/buildings.go`: Added `ruins map[string]int` to `BuildingManager`; `GenerateRuins(n)` — picks n random non-wonder buildings for ruins; `GetAllRuins()`, `LoadRuins()`; ruins included in `WorkerScaledProduction` at 50% base rate; ruins annotate `BuildingState.RuinCount` in `Snapshot()`
- [x] `game/engine.go`:
  - Added `legacyBonuses map[string]bool` and `catastropheHistory []string` to struct (persist through Prestige + Succumb)
  - `reapplyLegacyBonuses()` — applies all active legacy bonuses to `permanentBonuses` after any reset
  - `Endure() error` — 20% buildings destroyed, resources →15%, workers -25%, timed production -10% for 216 ticks, Survived marker set, history recorded
  - `Succumb() error` — generates 8 ruins, awards legacy bonus + Ancient Knowledge (+25% research speed), full reset preserving ruins/legacyBonuses/catastropheHistory/prestige/epochHistory, reapplies legacy bonuses
  - `DeferCatastrophe()` — engine no-op; UI handles hiding modal
  - `DoPrestige()` updated: preserves ruins across prestige, resets epoch state, calls `reapplyLegacyBonuses()`, logs ruins/legacy count
  - `GetState()` exposes `LegacyBonuses`, `CatastropheHistory`
- [x] `game/save.go`: Added `Ruins map[string]int`, `LegacyBonuses map[string]bool`, `CatastropheHistory []string` to `GameSave`; serialized in `buildSaveSnapshot()`; restored in `LoadGame()`
- [x] `ui/catastrophe_modal.go` (NEW): Full-screen overlay with epoch catastrophe name + flavor, Endure/Succumb/Defer buttons, keyboard shortcuts (E/S/D + Tab navigation), color-coded (dark red border, yellow Succumb), legacy bonus preview text
- [x] `ui/dashboard.go`: Added `catModalShown string` field; `refresh()` detects `PendingCatastrophe != ""` and shows modal once per new catastrophe; Defer closes modal without resetting `catModalShown` (prevents immediate re-show); clears `catModalShown` when catastrophe is resolved
- All tests passing, `go build ./...` clean

## Economy Redesign — Phase 10: Balance & Building Content (2026-03-04)
- [x] `config/buildings_new.go` (NEW): `newProductionBuildings()` — lineages 1–4 (Housing 21 tiers, Food 21 tiers, Organic Extraction 21 tiers w/ epoch output transitions wood→coal→oil→nanobots→quantum_flux, Geological Extraction 21 tiers w/ dual output at iron/classical tiers)
- [x] `config/buildings_new2.go` (NEW): `newProductionBuildings2()` — lineages 5–9 (Knowledge 21, Faith 21, Military 21, Trade 19 tiers bronze→quantum, Engineering 19 tiers bronze→quantum)
- [x] `config/buildings_new3.go` (NEW): `newProductionBuildings3()` — lineages 10–13 (Culture/Arts 17, Metallurgy 18, Energy 13, Hacker/Digital 8 tiers); culture buildings carry dual effects (production + culture storage cap)
- [x] `config/buildings_new_merge.go` (NEW): `NewProductionBuildings()` — merges all three helpers into a single slice
- [x] `config/buildings.go`: `BaseBuildings()` rewritten — prepends `NewProductionBuildings()`, then appends storage+wonder buildings filtered from legacy `baseBuildingsRaw()`; all metadata (`LineageKey`, `LineageTier`, `WorkerDomain`, `WorkerCapacity`, `EpochKey`, `OutputResource`) carried inline in new production defs
- [x] `config/ages.go`: Complete rewrite — all 22 ages updated with new `UnlockBuildings` arrays referencing new lineage building keys; `BuildingReqs` updated throughout; `coal` added to `UnlockResources` for `renaissance_age`
- [x] `config/trade.go`: Updated 7 stale `RequiredBld` keys to new building keys (`train_station`→`steam_works`, `oil_well`→`oil_derrick`, `power_grid`→`power_station`, `fiber_hub`→`server_farm`, `warp_gate`→`warp_drive_plant`, `galactic_hub`→`galactic_trade_hub`, `quantum_computer`→`reality_processor`)
- [x] `config/upgrades.go`: Updated stale building key references in upgrade chain defs (`apartment/skyscraper/neon_tower`→`tenement/tower_block/arcology_pod`; knowledge chain rewritten as `story_circle→scriptorium→library→university`; `mine`→`iron_mine`)
- Result: 284 total buildings (241 production across 13 lineages + 21 storage + 22 wonders), no duplicate keys, all tests passing, `go build ./...` clean
- Note: Storage Covenant and food-cost/multiplier tuning are initial-pass complete via lineages.md values; fine-tuning subject to playtesting

## Phase 11d: Culture Progress Bar + Faith Threshold Indicator (2026-03-04)
- [x] `ui/tab_economy.go`: culture and faith resource rows special-cased in `refreshResources()`
  - `cultureThresholds []float64` — 11 threshold values (500 → 1B)
  - `cultureThresholdLabels []string` — effect label per threshold
  - `cultureProgressBar(current, max)` — 10-char `▓`/`░` bar (distinct from existing `ProgressBar`)
  - `formatCultureRow` — progress bar toward next threshold + grey label + rate; "✦ Culture Mastered" at max
  - `faithBand` struct + `faithBandFor(amount, storage)` — 6 bands: Dead/Dim/Low/Mid/High/Full
  - `formatFaithRow` — band label (color-coded) + percentage + epoch odds hint in grey + rate
  - tview-safe: bar wrapped in `\u005b`/`\u005d` to prevent color-tag misparse
- All tests passing, `go build ./...` clean

## Phase 11a: Economy Tab Worker Assignment Display (2026-03-04)
- [x] `ui/tab_economy.go`: `refreshVillagers()` rewritten — per-domain headers with class name from `WorkerClassByDomainAndAge`, total/idle count, per-building assignment bars (▓/░ 10-char)
- [x] `refreshBuildings()` updated — inline worker line per building when `WorkerCapacity > 0` showing domain label, assigned/totalCap, bar
- [x] New helpers: `legacyKeyToDomain`, `domainToLabel`, `workerAssignBar`
- [x] All tests passing, `go build ./...` clean

## Phase 11b: Villager Panel Domain-Grouped Rewrite (2026-03-04)
- [x] `ui/villager_panel.go`: replaced `DefaultVillagerTypes()` legacy loop with live 12-domain data from `GameState.Villagers`
- [x] Four domain groups: Materials (food/lumber/masonry) / Knowledge (knowledge/faith) / Civil (trade/engineering) / Late-game (military/metallurgy/energy/hacker/astronaut)
- [x] Per-domain: class name via `config.WorkerClassByDomainAndAge`, count, idle, per-building assignments sorted alphabetically
- [x] Hash updated: covers TotalPop, FoodDrain, all Assignments map contents
- [x] All tests passing, `go build ./...` clean

## Phase 11c: Age Advance Modal — Transformation Summary + Epoch Reveal (2026-03-04)
- [x] `config/epochs.go`: added `EpochEventByKey()` — lock-free event lookup by key for UI use
- [x] `ui/age_splash.go`: `ShowAgeSplashFull()` — extends splash with `AgeAdvanceSummary` + `epochChanged bool` + `EpochEventRecord`; `ShowAgeSplash()` kept as zero-value wrapper
  - Transformation section: old→new building names with counts, legacy buildings in grey
  - Epoch section: color-coded (green/red) event name + flavor text + epoch icon
- [x] `ui/dashboard.go`: `pendingEpochChanged` field; epoch detection in bus handler via `config.EpochForAge()` (pure data, lock-safe); `refresh()` passes summary + epoch event to `ShowAgeSplashFull`
- [x] All tests passing, `go build ./...` clean

## Phase 11e: Epoch Tab — F10 (2026-03-04)
- [x] `ui/tab_epoch.go` (NEW): EpochTab struct with 4 sections — current epoch info, epoch history table, legacy bonuses, civilization log
  - Catastrophe status: PENDING / ✓ Survived / Succumbed / not yet triggered
  - Helper fns: findEpochEvent, epochHasCatastrophe, epochEventColor, formatAgeKey
- [x] `ui/dashboard.go`: Epoch tab at index 9 (F10); Dev tab stays on backtick; tab bar updated to F1-F10=Tabs
- [x] `ui/input.go`: `catastrophe invoke` command wired, calls `engine.InvokeCatastrophe()`, returns error toast on failure
- [x] `ui/autocomplete.go`: `catastrophe` + `invoke` subcommand added
- [x] All tests passing, `go build ./...` clean

## Phase 11f: Stats Tab — Epoch/Legacy Fields (2026-03-04)
- [x] `ui/tab_stats.go`: new Epoch & Legacy section appended to `refreshStats()`
  - Current epoch (icon + name in cyan), catastrophe count (Endured/Succumbed breakdown), legacy bonuses per epoch with resource+% formatting
  - config imported for LegacyBonusForEpoch + EpochByKey
- [x] All tests passing, `go build ./...` clean

## Phase 12: Documentation Rebuild (2026-03-04)
- [x] **New pages created**: `epochs.md` (170 lines), `faith.md` (81 lines), `knowledge.md` (96 lines), `workers-and-domains.md` (full 12-domain reference with all 189 class tiers from config/workers.go)
- [x] **Full rewrites**: `villagers.md` (12-domain system, class progression, assign syntax), `buildings.md` (13-lineage system, output transitions, legacy buildings), `resources.md` (25 resources, corrected unlock ages, faith/culture mechanics, ore chain)
- [x] **Substantial updates**: `prestige.md` (gate fixed: Transcendent→Modern Age; Prestige & Epochs section added), `ages.md` (epoch column, epoch boundary note, stale building keys removed)
- [x] **Minor updates**: `commands.md` (new assign syntax, catastrophe invoke, prestige gate fix), `_sidebar.md` (Deep Dive section), `first-ten-minutes.md` (all assign/recruit syntax), `how-to-play.md` (12-domain worker section), `milestones.md` (Culture Thresholds section), `technologies.md` (Ancient Knowledge callout)
- [x] Notable corrections found: faith resource unlocks Primitive Age (not Medieval), culture unlocks Classical Age (not Renaissance), iron_ore/marble both unlock Classical Age — all corrected
- [x] All tests passing, `go build ./...` clean, 15 files changed (+1276/-306 lines)

## General / Misc
- [x] Established TODO.md + DONE.md workflow for session-resumable planning (2026-02-25)

## Phase 13: Epoch-Exclusive Regular Events (2026-03-04)
- [x] Added `EpochKey string` field to `EventDef` struct in `config/events.go`
- [x] Added `EpochExclusiveEvents() []EventDef` with 35 events (5 per epoch × 7 epochs): stone_era, bronze_era, iron_era, medieval_era, industrial_era, digital_era, cosmic_era
- [x] Updated `EventByKey()` to include epoch-exclusive events in its lookup map
- [x] Modified `getEligible()` in `game/events.go` to accept `currentEpoch string` and append matching exclusive events to the candidate pool
- [x] Threaded `currentEpoch` through `Events.Tick()` signature and engine call site
- [x] Updated `game/events_test.go` — two Tick() calls pass `"stone_era"` as epoch argument
- [x] All tests passing, `go build ./...` clean — commit bc9ed5b

## Phase 14: Critical Bug Fix Pass (2026-03-04)
- [x] gather cap typo fixed: `amount = 10000` → `amount = 10` (ui/input.go)
- [x] Over-assignment blocked: AssignVillager/AssignAll check Buildings.GetCount > 0 before assigning
- [x] Snapshot domain key alignment: Snapshot() now keys Types by domain ("food") not legacy ("worker"); removed legacyKeyToDomain boilerplate from tab_economy.go and villager_panel.go

## Phase 14b: Purge Dead Legacy Building Code (2026-03-04)
- [x] Removed 57 dead production/research/housing/military entries from baseBuildingsRaw() in config/buildings.go (695 lines deleted)
- [x] Also committed user's refactor: buildings_new{1,2,3}.go → 13 buildings_lineage_*.go files

## Phase 17a: Stone + Military Age Gating (2026-03-04)
- [x] stone_camp moved to stone_age (stone resource not available in primitive_age)
- [x] hunting_lodge moved to iron_age (military worker domain starts at iron_age per workers.go)

## Phase 17b: Primitive Age Balance Pass (2026-03-04)
- [x] Worker food drain: 1.0 → 0.08/tick (food domain, primitive+stone age only)
- [x] gathering_camp: rate 0.05 → 0.50/tick, build time 40 → 12 ticks
- [x] forager_post: rate 0.10 → 1.00/tick, build time 100 → 30 ticks
- [x] hut: build time 80 → 8 ticks
- [x] Break-even verified: 5 food workers in 2 camps sustain 10 total workers

## Phase 19: Legacy Code & Villager→Worker Cleanup (2026-03-05)
- [x] 19a: Removed legacyAlias, domainToLegacy, resolveDomain() from game/villagers.go; updated config/ages.go UnlockVillagers to domain keys
- [x] 19b: Renamed VillagerManager→WorkerManager, VillagerState→WorkerState, VillagerTypeState→WorkerDomainState throughout; renamed engine methods RecruitVillager→RecruitWorker, AssignVillager→AssignWorker, UnassignVillager→UnassignWorker; save JSON tags preserved for backward compat
- [x] 19c: Dropped domain param from assign/unassign commands; assign/unassign now building-centric (domain auto-resolved); updated autocomplete chains; updated recruit usage to domain keys
- [x] 19d: Updated site docs — commands.md, workers-and-domains, any villager references
- [x] 19e: Final verification pass — zero legacy symbol matches across all Go source; all four command signatures confirmed building-centric; go build ./... clean; go test ./... passing
