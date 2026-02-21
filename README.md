# AgeForge - CLI Idle Empire Builder

AgeForge is a text-based idle/clicker game where you forge an empire from nothing, progressing through 22 ages of history — all within your terminal.

## Overview

Start in the Primitive Age with bare hands and 15 food. Gather resources, build structures, recruit villagers, research technologies, launch military expeditions, trade with factions, and advance through ages that span months of real-time play.

## Features

- **Resource Management**: 21 resources across 22 ages with storage limits and production chains
- **Building System**: 80 buildings (58 standard + 22 Wonders) with scaling costs and construction queues
- **Villager System**: 8 types (Worker, Shaman, Scholar, Soldier, Merchant, Engineer, Hacker, Astronaut) with food economy
- **Tech Tree**: 52 technologies with prerequisites and permanent bonuses
- **Military**: 15 expeditions with risk/reward, soldier management, and defense ratings
- **Random Events**: 27 events (beneficial, harmful, mixed) with streak balancing
- **Milestones**: 33 achievements across 5 categories (Settlement, Scholar, Builder, Military, Ages) with milestone chains, progress tracking, civilization titles, and temporary speed boosts
- **Age Progression**: 22 ages from Primitive to Transcendent with exponential requirements
- **Trade System**: 15 trade routes and resource exchange with supply/demand pressure
- **Diplomacy**: 6 NPC factions with opinion tracking, gifts, and trade bonuses
- **Prestige**: Reset-and-grow system with 9 upgrades and passive production bonuses
- **Speed System**: Wonder-based speed multipliers (+0.5x per wonder built)
- **Full Wiki**: In-game wiki with live stats and complete documentation
- **Tab-based TUI**: 9 tabs (Economy, Research, Military, Trade, Stats, Wiki, Map, Wonders, Logs) with keyboard navigation
- **Save/Load**: JSON save system with auto-save every 60s and offline progress

## Build & Run

```bash
go build -o ageforge .
./ageforge

# or use the run script
./run.sh
```

## How to Play

### Getting Started
1. `gather wood` — collect wood (need 10 for first hut)
2. `build hut` — build shelter (+2 population cap)
3. `recruit worker` — recruit your first worker
4. `assign worker food` — put them to work gathering food
5. Keep ~1/3 of workers on food to sustain your population

### Commands
- `gather <resource> [n]` — manually gather resources
- `build <building> [n]` — construct buildings
- `recruit <type> [n]` — recruit villagers
- `assign <type> <resource> [n|all]` — assign villagers to gather
- `unassign <type> <resource> [n|all]` — remove assignment
- `research <tech_key>` — start researching a technology
- `expedition <key>` — launch a military expedition
- `trade <from> <to> <amount>` — exchange resources
- `route start|stop <key>` — manage trade routes
- `diplomacy <faction> <action>` — interact with factions
- `upgrade <building>` — upgrade buildings to next tier
- `prestige` — reset with bonuses (requires Medieval Age+)
- `speed <multiplier>` — set game speed (requires wonders)
- `status` — detailed overview
- `save/load [name]` — save or load game

### Navigation
- F1-F9 — switch between tabs
- ESC — auto-save and return to menu
- Arrow keys / PgUp/PgDn — navigate wiki (in Wiki tab)
- v — toggle verbose logs (in Logs tab)

## Contributing

### Requirements

- Go 1.23+

### Dev Scripts

```bash
# Quick compile check (build + vet)
./dev.sh check

# Build + vet + run the full test suite (formatted output)
./dev.sh test

# Build + run the game
./dev.sh run

# Build + vet + run (default)
./dev.sh

# Auto-rebuild on file changes (requires: brew install fswatch)
./dev.sh watch
```

Or use `make`:

```bash
make check      # build + vet
make test       # build + vet + tests (formatted output)
make test-raw   # build + vet + tests (raw go test -v, for CI/piping)
make run        # build + run
make clean      # remove binary
make release    # cross-compile for darwin/linux/windows
```

### Running Tests

The test suite covers all game systems with **74 tests** across 10 test files:

| File | Tests | What it covers |
|------|-------|----------------|
| `resources_test.go` | 7 | Add, storage cap, remove, pay/afford, rates, unlock, save/load |
| `buildings_test.go` | 5 | Unlock, cost scaling, pop capacity, get all, load counts |
| `villagers_test.go` | 9 | Recruit, cap limits, unlock, assign/unassign, food drain, production, soldiers, save/load |
| `research_test.go` | 9 | Start, afford check, age gating, prereqs, tick completion, bonuses, cancel, duplicate, save/load |
| `milestones_test.go` | 8 | First shelter, population, age gating, chains, titles, snapshots, hidden visibility, save/load |
| `prestige_test.go` | 5 | Can prestige, point calc, diminishing returns, level grants, save/load |
| `progress_test.go` | 5 | Age order, next age, display names, advancement check, requirements |
| `bus_test.go` | 4 | Subscribe/publish, multiple subscribers, no subscribers, event isolation |
| `events_test.go` | 3 | Inject event, expiration, save/load |
| `engine_test.go` | 19 | Full integration: init, resources, gather, build, recruit, assign, research, cancel, state consistency, speed, reset, milestone events, chain events, build multiple, save/load |

**Run the suite:**

```bash
./dev.sh test
# or
make test
```

Output shows a per-test checklist with pass/fail indicators, then a summary:

```
  ✓ TestBuildingManager_UnlockAndCount (0.00s)
  ✓ TestBuildingManager_CostScaling (0.00s)
  ✗ TestSomething_Broken (0.00s)

━━━ Test Summary ━━━

Packages:
  ✓ game (0.18s)

Results:  73 passed  1 failed  0 skipped  (74 total)

Failures:
  ✗ TestSomething_Broken
    → some_test.go:42 → expected 10, got 5
```

**Run a single test:**

```bash
go test ./game/ -run TestEngine_BuildMultiple -v
```

**Run tests for one package with raw output:**

```bash
go test ./game/ -v -count=1
```

**Common test patterns used:**

- Tests create isolated managers (`NewResourceManager()`, `NewBuildingManager()`, etc.) — no shared state
- Resource tests must respect `BaseStorage` caps (food: 50, wood: 50, knowledge: 30) — use `AddStorage()` before `Add()` if you need large amounts
- Milestone tests use `fullAgeOrder()` (via `NewProgressManager().GetAgeOrder()`) to get the complete age map — incomplete maps cause milestones with missing `MinAge` entries to auto-complete
- Engine tests access internals via `ge.mu.Lock()` for setup, then use public API methods for the actual test
- Save/load round-trip tests create a file, defer cleanup with `defer os.Remove(...)`, and verify state survives serialization

### Project Structure

```
config/     Data definitions (resources, buildings, techs, ages, milestones).
            Pure data, no logic. All content is config-driven.
game/       Game engine, managers, tick loop. No UI imports.
ui/         tview-based TUI. Reads GameState snapshots only.
main.go     Entry point, wires engine + UI.
```

### Key Patterns

- **Config-Driven Content**: All game content (buildings, techs, ages, milestones, events, trade routes) is defined as data in `config/`. Add new content there, not in game logic.
- **Manager Pattern**: Each system (resources, buildings, villagers, research, military, milestones, trade, diplomacy, prestige) has its own manager struct in `game/` with a clear API.
- **GameState Snapshot**: `engine.GetState()` returns a read-only snapshot. UI reads snapshots, never touches engine internals.
- **Event Bus**: Systems communicate via `game.EventBus` (pub/sub). Subscribe in `ui/dashboard.go` for toasts, in managers for cross-system reactions.
- **No Global State**: Pass dependencies explicitly. No singletons.

### Adding Content

**New building**: Add a `BuildingDef` to `config/buildings.go` with `BaseCost`, `CostScale`, `BuildTicks`, `Category`, and `Effects`. Unlock it in the appropriate age's `UnlockBuildings` list in `config/ages.go`. The cost formula is `floor(BaseCost * CostScale^count)` — typical CostScale values are 1.25-1.6.

**New milestone**: Add a `MilestoneDef` to `config/milestones.go` with a `Category` (settlement/builder/scholar/military/ages). Set `Hidden: true` if it should only appear when the player is close to completing it (>50% progress). If it belongs in a chain, add its key to the chain's `MilestoneKeys` in `MilestoneChains()`. The engine auto-detects chain completion and grants the speed boost.

**New age**: Add an `AgeDef` to `config/ages.go` with `ResourceReqs`, `BuildingReqs`, `UnlockResources`, `UnlockBuildings`, `UnlockVillagers`. Add a matching age milestone to `config/milestones.go` with `MinAge` set and `Category: "ages"`. On advancement, all resources are reduced to 25%.

**New tech**: Add a `TechDef` to `config/techs.go` with `Age` (gating), `Cost` (knowledge), `Prerequisites` (tech keys), and `Effects`. Effect types: `"production"` (per-tick output), `"bonus"` (multiplier on a rate like `"gold_rate"` or `"tick_speed"`).

**New random event**: Add an `EventDef` to `config/events.go` with `Sentiment` (good/bad/mixed), `Weight` (higher = more likely), `Cooldown` (min ticks between repeats), `Duration` (0 for instant), `MinAge`, and `Effects`. The streak system caps bad events at 2 consecutive and forces a bad event after 3 good ones.

**New expedition**: Add an expedition def to the `getExpeditions()` function in `game/military.go` with `SoldiersNeeded`, `Duration`, `DifficultyBase`, `Rewards`, and `MinAge`. Success chance = `random() > (DifficultyBase - military_bonus * 0.3)`.

**New trade route**: Add a `TradeRouteDef` to `config/trade.go` with `Export`/`Import` maps, `TicksPerRun`, `RequiredBuilding`, and `MinAge`. Routes auto-cycle: deduct exports, add imports scaled by diplomacy bonuses.

**New villager type**: Add a `VillagerTypeDef` to `game/villagers.go` with `FoodCost` (per tick) and `GatherRate` (per tick when assigned). Unlock it in the appropriate age in `config/ages.go`.

### How the Math Works

#### Tick Loop

The game runs on a tick loop. Each tick processes: build queue, research, random events, expeditions, trade routes, diplomacy, production rates, resource application, milestones, age advancement, and tick speed recalculation. The base tick interval is **2 seconds**, modified by bonuses.

#### Tick Speed

```
tick_interval = 2000ms / ((1.0 + tick_speed_bonus) * speed_multiplier)
minimum: 200ms (hard floor)
```

`tick_speed_bonus` comes from: research, milestones, prestige (+1% per level), and active chain boosts. `speed_multiplier` is player-set in 0.5x increments, capped at `1.0 + (wonders_built * 0.5)`.

#### Resource Rates

Rates are recalculated every tick in this order:

1. **Base rates**: building production + villager gathering + research effects + event effects
2. **`production_all` multiplier**: `rate *= (1.0 + bonus)` on all positive rates
3. **Per-resource multiplier**: e.g. `gold_rate` bonus: `rate *= (1.0 + bonus)`
4. **Gather rate bonus**: additive on villager rates
5. **Diplomacy trade bonuses**: multiplicative on positive rates
6. **Food drain subtracted**: `sum(villager_count * food_cost_per_type)`

Storage = `BaseStorage + all_storage_bonuses + per_resource_bonuses` (from buildings, research, milestones, prestige).

#### Building Costs

```
cost = floor(base_cost * cost_scale ^ current_count)
```

Example: Hut costs 30 wood with scale 1.3 — 1st: 30, 2nd: 39, 3rd: 50, 4th: 66...

Building upgrades cost `target_base_cost * 0.25` (75% discount).

#### Food Economy

Each villager type has a per-tick food cost. Workers cost 0.10/tick, soldiers 0.25/tick, astronauts 0.40/tick. Total drain = `sum(count * cost)`. When food hits 0, a starvation warning fires. Keep ~1/3 of your workforce on food.

#### Expeditions

```
adjusted_difficulty = max(0.05, base_difficulty - military_bonus * 0.3)
success = random() > adjusted_difficulty
loot = base_reward * (1.0 + expedition_bonus)    # on success
loot = base_reward * 0.3                          # on failure (partial)
soldier_loss: success = 0-1, failure = 1-2
```

#### Trade & Exchange

Resource exchange uses supply/demand pressure:

```
rate = base_rate * (1.0 - pressure * 0.3)     # min 50% of base
pressure_increase = 0.1 / (1.0 + market_count * 0.2)
pressure_decay = pressure * 0.98 per tick      # recovers over time
```

More markets = less pressure per trade. Pressure decays 2% per tick naturally.

Trade routes cycle every `TicksPerRun` ticks: deduct exports, add `imports * (1.0 + diplomacy_bonus)`.

#### Prestige

Requires Medieval Age+. Points formula:

```
base = age_order_index
bonus = floor(milestones/10) + floor(techs/15) + floor(buildings/50)
points = floor((base + bonus) / sqrt(prestige_level + 1))
```

Each prestige level grants +2% production, +1% tick speed permanently. 9 upgrades purchasable with prestige points.

#### Offline Progress

On load, the game simulates missed time:

```
offline_ticks = min(elapsed, 24h) / tick_interval
resource_gain = rate * offline_ticks * 0.5       # 50% efficiency
capped at remaining storage
```

#### Milestone Chains

5 chains (Settlement, Scholar, Builder, Military, Ancient Ages). Completing all milestones in a chain grants:
- A civilization title (displayed in status bar)
- A temporary tick_speed boost injected as an active event (e.g. Settlement: +3.0x for 180 ticks)

Hidden milestones become visible when progress exceeds 50%, or (for age milestones) when the player reaches the preceding age.

### Wiring: How Systems Connect

```
config/*.go          game/*.go                    ui/*.go
-----------          ---------                    -------
BuildingDef    -->   BuildingManager.Build()  --> EconomyTab.Refresh()
                     BuildingManager.GetCost()
                     engine.processBuildQueue()
                         |
                         v
                     EventBus.Publish(BuildingBuilt)
                         |
                         v
                     dashboard.go subscriber --> ToastManager.Show()
                     MilestoneManager.Check()
                         |
                         v (if chain completes)
                     EventManager.InjectEvent(speed boost)
                     EventBus.Publish(ChainCompleted)
```

The flow for any system follows this pattern:
1. **Config** defines the data (costs, effects, requirements)
2. **Manager** in `game/` owns the state and logic
3. **Engine** orchestrates managers in `doTick()` and public API methods
4. **EventBus** notifies other systems (pub/sub, synchronous under write lock)
5. **UI** reads `GetState()` snapshots every 500ms — never writes to engine

To add a new toast notification: subscribe to an event in `ui/dashboard.go`'s `build()` method. To add a new UI display: update the relevant tab's `Refresh()` method using fields from `GameState`.

### Important: Event Bus Deadlock

Bus handlers run synchronously under the engine's write lock. **Never call `engine.GetState()` or any lock-acquiring method inside a bus subscriber.** Use `config.*ByKey()` functions (pure data, no locks) for lookups in handlers.

### Conventions

- Package names: lowercase, single word (`config`, `game`, `ui`)
- Config keys: `snake_case` strings (`"lumber_mill"`, `"stone_age"`)
- `float64` for resource amounts, `int` for building counts
- Return errors up, log at boundaries
- Keep changes minimal — don't refactor code you didn't need to touch
