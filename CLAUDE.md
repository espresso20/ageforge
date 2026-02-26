# AgeForge - Project Conventions

## Overview
CLI idle/clicker empire builder game built with Go + tview/tcell. 22 ages, 80 buildings, 52 techs, 33 milestones with chains/titles, trade, diplomacy, prestige, and wonder-based speed system.

## Architecture
- **config/** - Data definitions (resources, buildings, techs, ages, milestones, events, trade, diplomacy, prestige). Pure data, no logic.
- **game/** - Game engine, managers, tick loop. No UI imports.
- **ui/** - tview-based TUI. Reads GameState snapshots, never touches engine internals.
- **main.go** - Entry point, wires engine + UI.

## Key Systems
- **ResourceManager** - 21 resources with rates, storage caps, breakdowns
- **BuildingManager** - 80 buildings (58 standard + 22 wonders) with scaling costs, build queues, upgrades
- **VillagerManager** - 8 types with food drain, assignments, idle tracking
- **ResearchManager** - 52 techs with prerequisites, age gating, permanent bonuses
- **MilitaryManager** - 15 expeditions with risk/reward, soldiers, loot
- **EventManager** - 28 random events with sentiment streaks, timed effects, InjectEvent() for chain boosts
- **MilestoneManager** - 33 milestones in 5 categories, 5 chains with speed boosts/titles, progress tracking, hidden milestone visibility
- **TradeManager** - 15 trade routes, resource exchange with supply/demand pressure
- **DiplomacyManager** - 6 NPC factions with opinion, status, trade bonuses
- **PrestigeManager** - Reset system with 9 upgrades and passive bonuses
- **ProgressManager** - Age advancement, unlock tracking across 22 ages

## Patterns
- **Event Bus**: Systems communicate via `game.EventBus` (pub/sub). Events: BuildingBuilt, ResearchDone, AgeAdvanced, MilestoneCompleted, ChainCompleted, etc.
- **Config-Driven**: All content defined as data in config/, referenced by string keys.
- **Manager Pattern**: Each system has a manager struct with clear API.
- **GameState Snapshot**: `engine.GetState()` returns a read-only snapshot for UI consumption.
- **Toast Notifications**: Bus subscribers in dashboard show temporary toast messages for milestones, chains, wonders, age advances.

## Conventions
- Package names: lowercase, single word (config, game, ui)
- Config keys: snake_case strings ("lumber_mill", "stone_age")
- Use `float64` for resource amounts, `int` for building counts
- Error handling: return errors up, log at boundaries
- No global state; pass dependencies explicitly
- Bus handlers run under engine write lock — never call GetState() or other lock-acquiring methods inside them

## TODO / DONE Workflow
- **TODO.md** tracks all pending and in-progress work, organized by phase.
- **DONE.md** is the permanent history of completed work — items are moved here (never deleted) when finished.
- When working on any multi-step plan or feature:
  - Mark items `[~]` (in progress) in TODO.md before starting them
  - Mark items `[x]` and move them to DONE.md (with a brief date/note) when finished
  - Add newly discovered sub-tasks to TODO.md immediately — do not hold them in context only
  - Never implement something already listed in DONE.md; check it first to avoid duplication
- The goal: full session-resumability. TODO.md + DONE.md should always reflect true current state.

## Build & Run
```bash
go build -o ageforge .
./ageforge
# or
go run main.go
```

## Testing
```bash
go test ./...
```
