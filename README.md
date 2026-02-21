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

## Requirements

- Go 1.23+ to build from source
