# AgeForge - Project Conventions

## Overview
CLI idle/clicker empire builder game built with Go + tview/tcell. 22 ages, 284 buildings (13 lineages), 52 techs, 33 milestones, 7 epochs, catastrophe system, 12 worker domains, 25 resources, trade, diplomacy, prestige.

## Architecture
- **config/** - Data definitions (resources, buildings, techs, ages, milestones, events, trade, diplomacy, prestige, epochs, workers). Pure data, no logic.
- **game/** - Game engine, managers, tick loop. No UI imports.
- **ui/** - tview-based TUI. Reads GameState snapshots, never touches engine internals.
- **main.go** - Entry point, wires engine + UI.

## Key Systems
- **ResourceManager** - 25 resources with rates, storage caps, breakdowns
- **BuildingManager** - 284 buildings (241 production across 13 lineages + 21 storage + 22 wonders); lineage tier transforms on age advance; legacy building tracking
- **VillagerManager** - 12 worker domains (food/faith/knowledge/military/trade/engineering/hacker/astronaut/lumber/masonry/metallurgy/energy); domain+building keyed assignments; age-tiered class names; WorkerScaledProduction (20% floor)
- **ResearchManager** - 52 techs with prerequisites, age gating, permanent bonuses
- **MilitaryManager** - 15 expeditions with risk/reward, soldiers, loot
- **EventManager** - 28 random events with sentiment streaks, timed effects, InjectEvent() for chain boosts
- **MilestoneManager** - 33 milestones in 5 categories, 5 chains with speed boosts/titles
- **TradeManager** - 15 trade routes, resource exchange with supply/demand pressure
- **DiplomacyManager** - 6 NPC factions with opinion, status, trade bonuses
- **PrestigeManager** - Reset system with 9 upgrades and passive bonuses
- **ProgressManager** - Age advancement, unlock tracking across 22 ages
- **EpochSystem** - 7 epochs (3 ages each); epoch event rolls (faith→odds, culture→tier); Catastrophe modal (Endure/Succumb/Defer); legacy bonuses + ruins surviving prestige

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

## Website & Wiki Sync Rule
**Any change to the game MUST also update the appropriate pages in `site/`.**
- Balance change (building cost, resource rate, tech effect, age requirement) → update `site/docs/` wiki page for that system
- New building/tech/age/resource/wonder/expedition → add it to the relevant wiki page AND the landing page if it affects headline stats
- Mechanic change (new command, removed command, renamed key) → update `site/docs/commands.md` and any affected wiki page
- The landing page `site/index.html` hero stats must stay accurate
- Use a subagent for the wiki update if it spans more than 2 files

## Subagent Usage — Token Efficiency (CRITICAL)
This project is large. Burning main context on file reads is wasteful. Follow these rules strictly:

**What goes in subagents (NOT main context):**
- All implementation work for any sub-phase: reading files + writing code + running tests
- Any task touching more than 2 files
- All wiki/docs writes
- Any bulk read of config or source files

**What stays in main context:**
- Architecture decisions and reasoning
- Writing self-contained subagent prompts
- Small targeted edits (1-2 lines in a file already read this session)
- Reviewing subagent summaries and deciding next steps

**Subagent rules:**
- Use `subagent_type=general-purpose` for read+write tasks (implementation, docs)
- Use `subagent_type=Explore` for read-only research across multiple files
- Every subagent prompt must be **fully self-contained**: list exact files to read, exact changes needed, verify with `go build ./...` and `go test ./...`, return a summary of changes made
- Spawn parallel subagents for independent work (e.g. Phase 11a + 11b simultaneously; Phase 12 docs in 3-4 parallel batches)

**Never read these files in main context** (too large — always delegate):
- `config/buildings_new*.go`, `config/ages.go`, `config/buildings.go`
- `game/engine.go`, `game/villagers.go`, `game/buildings.go`
- `ui/tab_economy.go`, `ui/dashboard.go`
- Any file over ~200 lines unless doing a single targeted Grep lookup

**Per-phase workflow:**
1. Write a self-contained subagent prompt with exact spec
2. Spawn subagent — it reads, implements, tests, reports back
3. Review summary in main context
4. Commit if clean; update TODO.md + DONE.md
