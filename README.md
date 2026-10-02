<p align="center">
  <a href="https://ageforge.io">
    <img src="https://raw.githubusercontent.com/espresso20/ageforge/master/ageforge-3.png" alt="AgeForge" width="200">
  </a>
</p>

<h1 align="center">AgeForge</h1>

<p align="center">
  <em>Forge a civilization from nothing, all within your terminal.</em>
</p>

<p align="center">
  <a href="https://github.com/espresso20/ageforge/releases/latest">
    <img src="https://img.shields.io/github/v/release/espresso20/ageforge?style=for-the-badge&color=f0a500&labelColor=1a1a1a&logo=github&logoColor=white" alt="Latest Release">
  </a>
  &nbsp;
  <a href="https://golang.org">
    <img src="https://img.shields.io/badge/Go-1.24-00ADD8?style=for-the-badge&logo=go&logoColor=white&labelColor=1a1a1a" alt="Go 1.24">
  </a>
  &nbsp;
  <a href="https://github.com/espresso20/ageforge/actions/workflows/go.yml">
    <img src="https://img.shields.io/github/actions/workflow/status/espresso20/ageforge/go.yml?branch=master&style=for-the-badge&label=CI&logo=githubactions&logoColor=white&labelColor=1a1a1a" alt="CI">
  </a>
  &nbsp;
  <a href="https://github.com/espresso20/ageforge/blob/master/LICENSE">
    <img src="https://img.shields.io/badge/license-non--commercial-4a4a4a?style=for-the-badge&labelColor=1a1a1a" alt="License">
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

AgeForge is a text-based idle/clicker game where you forge an empire from nothing and take it through 22 ages of history, all within your terminal.

## Overview

A new game starts in the Primitive Age with 25 food and 50 wood. Gather resources, build structures, recruit workers, research technologies, send scouting expeditions and military campaigns, trade with other civilizations, and advance through ages that take days of real-time play.

## Features

- **Resource Management**: 26 resources across 22 ages with storage limits and production chains
- **Building System**: 301 buildings (250 lineage buildings + 21 storage + 22 Wonders + 4 cultural monuments + 4 standalone: the Nano Foundry and 3 diplomatic buildings) with scaling costs and construction queues
- **Worker System**: 12 domains (food, faith, knowledge, military, trade, engineering, hacker, astronaut, lumber, masonry, metallurgy, energy) with per-domain class progression and food economy
- **Tech Tree**: 73 technologies with prerequisites and permanent bonuses
- **Military**: 13 campaigns that cost soldiers and 3 scouting expeditions that cost resources, each with a chance of failure and set rewards
- **Epoch System**: 7 epochs with faith-gated event rolls, catastrophe choices from the Iron Era on (Endure/Succumb), and legacy bonuses that carry across runs
- **Random Events**: 61 events (26 base + 35 epoch-exclusive) with streak balancing
- **Milestones**: 77 milestones across 6 chains with civilization titles and temporary speed boosts
- **Age Progression**: 22 ages from Primitive to Transcendent with exponential requirements; on each advance, lineage buildings with a next tier can be upgraded to it (`upgrade`)
- **Trade System**: 21 trade routes and resource exchange with supply/demand pressure
- **Diplomacy**: 11 civilizations with opinion tracking, gifts, alliances, trade deals and trade bonuses
- **Prestige**: Reset-and-grow system with 9 upgrades and passive production bonuses (requires Modern Age)
- **Command-driven interface**: everything is typed at one prompt; panels (research, army, trade, stats, wonders, logs, epoch, map and more) open by name, and `help` lists every command
- **Wiki**: full player documentation at [ageforge.io/docs](https://ageforge.io/docs/)
- **Save/Load**: JSON save system with auto-save every 60s and offline progress; saves live in `data/` next to the binary

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
1. `build hut`: a hut costs 14 wood and gives +10 housing
2. `recruit 5`: fill the new housing with workers
3. `build gathering_camp`, then `assign gathering_camp 3`: put workers on food
4. `gather wood`: top up by hand while production is small
5. Workers eat food, so keep enough of them producing it

### Commands
The full list is in the in-game `help` panel and on the [Commands](https://ageforge.io/docs/#/commands) wiki page. The ones you will use most:

- `gather <food|wood|stone> [n]`: gather by hand (until the Medieval Age)
- `build <building> [count|max]`: construct buildings
- `upgrade`: list buildings you can upgrade; `upgrade <building> [count|all]` upgrades them to the next age tier
- `recruit [count|max]`: recruit workers into free housing
- `assign <building> [n|all]`: assign workers to a building
- `unassign <building> [n|all]`: return workers to the idle pool
- `research <tech>`: start researching a technology
- `advance`: advance to the next age once it is ready (`plan advance` does it for you as soon as it is)
- `plan`: queue builds, techs and trades that start when the resources come in
- `campaign <key>`: wage a military campaign (costs soldiers)
- `expedition <key>`: send a scouting expedition (costs resources)
- `trade <give> <get> <amount to give>`: exchange resources
- `trade route start|stop <route>`: manage trade routes
- `factions` (alias `diplomacy`): open the Factions panel; `diplomacy ally|rival|gift|embargo|neutral <civ>` acts on a civilization
- `wonder collect <resource|all> [amount|max]`: bank resources into the current wonder
- `prestige`: view prestige status; `prestige confirm yes` resets with bonuses (requires Modern Age)
- `status` / `rates`: detailed overview / resource rate breakdown
- `save [name]` / `load [name]` / `saves`: save, load or list saves
- `quit`: save and quit

### Keys
- Tab / Shift+Tab: cycle command completions
- ↑ / ↓: step through command history
- PgUp / PgDn: scroll the main view
- Esc: close the open panel; with no panel open, save, stop the game and return to the main menu

## Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md) for the full dev guide: commit workflow, release process, test patterns, project structure, adding content, and how the math works.

Active work is tracked on the [Project Board](https://trello.com/b/tf31C2cz/ageforge). Bugs, features, balance and refactor work each have their own lane.

### Commit style

Use `make commit`. It prompts you interactively and formats the message correctly:

```
What kind of change?
  1  feat      - new feature or content
  2  fix       - bug fix
  3  balance   - tuning costs, rates, numbers
  4  refactor  - cleanup, no behavior change
  5  chore     - build/tooling/deps
  6  docs      - docs/comments only

Short summary (≤72 chars): add iron smeltery building
Optional details (bullet then Enter, empty Enter when done):
  · costs 50 stone and 20 coal
  · produces iron at 0.5/s
  ·
```

Produces: `feat: add iron smeltery building` with the bullets as body. This drives the auto-generated changelog on every release.
