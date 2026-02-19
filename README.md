# AgeForge - CLI Idle Empire Builder

AgeForge is a text-based idle/clicker game where you forge an empire from nothing, progressing through 8 ages of history — all within your terminal.

## Overview

Start in the Primitive Age with bare hands and 15 food. Gather resources, build structures, recruit villagers, research technologies, launch military expeditions, and advance through ages that span months of real-time play.

## Features

- **Resource Management**: 12 resources across 8 ages with storage limits and production chains
- **Building System**: 28 buildings with scaling costs, construction queues, and 4 Wonders
- **Villager System**: 4 types (Worker, Scholar, Soldier, Merchant) with food economy
- **Tech Tree**: 30 technologies with prerequisites and permanent bonuses
- **Military**: Expeditions with risk/reward, soldier management, and defense ratings
- **Random Events**: 15 events (beneficial, harmful, mixed) that trigger during play
- **Milestones**: 20 achievements with permanent bonus rewards
- **Age Progression**: 8 ages with exponential requirements (designed for months of play)
- **Full Wiki**: In-game wiki with live stats and complete documentation
- **Tab-based TUI**: 5 tabs (Economy, Research, Military, Stats, Wiki) with keyboard navigation
- **Save/Load**: JSON save system with auto-save on exit

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
- `build <building>` — construct a building
- `recruit <type> [n]` — recruit villagers
- `assign <type> <resource> [n]` — assign villagers to gather
- `unassign <type> <resource> [n]` — remove assignment
- `research <tech_key>` — start researching a technology
- `expedition <key>` — launch a military expedition
- `status` — detailed overview
- `save/load [name]` — save or load game

### Navigation
- F1-F5 / Tab — switch between dashboard tabs
- ESC — auto-save and return to menu
- Arrow keys / PgUp/PgDn — navigate wiki (in Wiki tab)

## Requirements

- Go 1.23+ to build from source
