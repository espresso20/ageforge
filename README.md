<p align="center">
  <a href="https://ageforge.io">
    <img src="https://raw.githubusercontent.com/espresso20/ageforge/master/ageforge-3.png" alt="AgeForge" width="200">
  </a>
</p>

<h1 align="center">AgeForge</h1>

<p align="center">
  <em>Forge a civilization from nothing — all within your terminal.</em>
</p>

<p align="center">
  <a href="https://github.com/espresso20/ageforge/releases/latest">
    <img src="https://img.shields.io/github/v/release/espresso20/ageforge?style=for-the-badge&color=f0a500&labelColor=1a1a1a&logo=github&logoColor=white" alt="Latest Release">
  </a>
  &nbsp;
  <a href="https://golang.org">
    <img src="https://img.shields.io/badge/Go-1.23-00ADD8?style=for-the-badge&logo=go&logoColor=white&labelColor=1a1a1a" alt="Go 1.23">
  </a>
  &nbsp;
  <a href="https://github.com/espresso20/ageforge/blob/master/LICENSE">
    <img src="https://img.shields.io/github/license/espresso20/ageforge?style=for-the-badge&color=4a4a4a&labelColor=1a1a1a" alt="License">
  </a>
  &nbsp;
  <a href="https://ageforge.io">
    <img src="https://img.shields.io/badge/Website-ageforge.io-f0a500?style=for-the-badge&labelColor=1a1a1a&logoColor=white" alt="Website">
  </a>
  &nbsp;
  <a href="https://trello.com/b/tf31C2cz/ageforge">
    <img src="https://img.shields.io/badge/Project_Board-Trello-0052cc?style=for-the-badge&logo=trello&logoColor=white&labelColor=1a1a1a" alt="Project Board">
  </a>
</p>

<br>

AgeForge is a text-based idle/clicker game where you forge an empire from nothing, progressing through 22 ages of history — all within your terminal.

## Overview

Start in the Primitive Age with bare hands and 15 food. Gather resources, build structures, recruit villagers, research technologies, launch military expeditions, trade with factions, and advance through ages that span months of real-time play.

## Features

- **Resource Management**: 25 resources across 22 ages with storage limits and production chains
- **Building System**: 301 buildings (251 lineage buildings + 21 storage + 22 Wonders + 4 cultural monuments + 3 administrative) with scaling costs and construction queues
- **Worker System**: 12 domains (food, faith, knowledge, military, trade, engineering, hacker, astronaut, lumber, masonry, metallurgy, energy) with per-domain class progression and food economy
- **Tech Tree**: 52 technologies with prerequisites and permanent bonuses
- **Military**: 15 expeditions with risk/reward and defense ratings
- **Epoch System**: 7 epochs with faith-gated event rolls, catastrophe choices (Endure/Succumb/Defer), and legacy bonuses that carry across runs
- **Random Events**: 62 events (27 base + 35 epoch-exclusive) with streak balancing
- **Milestones**: 33 achievements across 5 chains with civilization titles and temporary speed boosts
- **Age Progression**: 22 ages from Primitive to Transcendent with exponential requirements and building transformation on advance
- **Trade System**: 15 trade routes and resource exchange with supply/demand pressure
- **Diplomacy**: 6 NPC factions with opinion tracking, gifts, and trade bonuses
- **Prestige**: Reset-and-grow system with 9 upgrades and passive production bonuses (requires Modern Age)
- **Speed System**: Wonder-based speed multipliers (+0.5x per wonder built)
- **Full Wiki**: In-game wiki with live stats and complete documentation
- **Tab-based TUI**: 10 tabs (Economy, Research, Military, Trade, Stats, Wiki, Map, Wonders, Logs, Epoch) with keyboard navigation
- **Save/Load**: JSON save system with auto-save every 60s and offline progress

## Build & Run

```bash
go build -o ageforge .
./ageforge

# Check version
./ageforge --version
```

Or use `make`:

| Command | What it does |
|---------|-------------|
| `make build` | Compile the binary |
| `make run` | Build and launch |
| `make test` | Run all tests |
| `make commit` | Interactive commit helper (conventional commits) |

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
- `assign <building> [n|all]` — assign workers to a building
- `unassign <building> [n|all]` — remove worker assignment
- `research <tech_key>` — start researching a technology
- `expedition <key>` — launch a military expedition
- `trade <from> <to> <amount>` — exchange resources
- `route start|stop <key>` — manage trade routes
- `diplomacy <faction> <action>` — interact with factions
- `upgrade <building>` — upgrade buildings to next tier
- `prestige` — reset with bonuses (requires Modern Age)
- `speed <multiplier>` — set game speed (requires wonders)
- `status` — detailed overview
- `save/load [name]` — save or load game

### Navigation
- F1-F9 — switch between tabs
- F10 — Dev console tab (only visible after developer unlock)
- ESC — auto-save and return to menu
- Arrow keys / PgUp/PgDn — navigate wiki (in Wiki tab)
- v — toggle verbose logs (in Logs tab)

## Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md) for the full dev guide — commit workflow, release process, test patterns, project structure, adding content, and how the math works.

Active work is tracked on the [Project Board](https://trello.com/b/tf31C2cz/ageforge) — bugs, features, balance, and refactor work each have their own lane.

### Commit style

Use `make commit` — it prompts you interactively and formats the message correctly:

```
What kind of change?
  1  feat      — new feature or content
  2  fix       — bug fix
  3  balance   — tuning costs, rates, numbers
  4  refactor  — cleanup, no behavior change
  5  chore     — build/tooling/deps
  6  docs      — docs/comments only

Short summary (≤72 chars): add iron smeltery building
Optional details (bullet then Enter, empty Enter when done):
  · costs 50 stone and 20 coal
  · produces iron at 0.5/s
  ·
```

Produces: `feat: add iron smeltery building` with the bullets as body. This drives the auto-generated changelog on every release.
