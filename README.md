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

# Build + vet + run tests
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
make check    # build + vet
make test     # build + vet + tests
make run      # build + run
make clean    # remove binary
```

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

**New building**: Add a `BuildingDef` to `config/buildings.go`, unlock it in the appropriate age in `config/ages.go`.

**New milestone**: Add a `MilestoneDef` to `config/milestones.go` with a `Category`. If it belongs in a chain, add its key to the chain's `MilestoneKeys` in `MilestoneChains()`.

**New age**: Add an `AgeDef` to `config/ages.go` with resource/building requirements, unlocks, and wonder. Add a matching age milestone to `config/milestones.go`.

**New tech**: Add a `TechDef` to `config/techs.go` with age gating and prerequisites.

**New random event**: Add an `EventDef` to `config/events.go` with sentiment, weight, cooldown, and effects.

### Important: Event Bus Deadlock

Bus handlers run synchronously under the engine's write lock. **Never call `engine.GetState()` or any lock-acquiring method inside a bus subscriber.** Use `config.*ByKey()` functions (pure data, no locks) for lookups in handlers.

### Conventions

- Package names: lowercase, single word (`config`, `game`, `ui`)
- Config keys: `snake_case` strings (`"lumber_mill"`, `"stone_age"`)
- `float64` for resource amounts, `int` for building counts
- Return errors up, log at boundaries
- Keep changes minimal — don't refactor code you didn't need to touch
