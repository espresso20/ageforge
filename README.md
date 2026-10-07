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
    <img src="https://img.shields.io/badge/Go-1.26-00ADD8?style=for-the-badge&logo=go&logoColor=white&labelColor=1a1a1a" alt="Go 1.26">
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

AgeForge is an idle civilization game for the terminal: you forge an empire from nothing and take it through 22 ages of history, from a campfire to the Transcendent Age, one typed command at a time.

> Version note: this README and the [wiki](https://ageforge.io/docs/) describe the game on master, version 4.0, which is coming soon. The latest release is 3.6, which still has some older systems (the `speed` command, the old city map).

## Overview

A new game starts in the Primitive Age with 25 food and 50 wood. Gather, build, let workers staff your buildings, research technologies, send expeditions and campaigns, trade and make deals with other civilizations, and advance through the ages. The first run is slow on purpose: about a week of real time to the Modern Age and a full run's prestige. The game keeps playing while you are away (up to 24 hours, at half rate), and the build plan spends your income for you.

## Features

- **Resource Management**: 26 resources across 22 ages with storage limits and production chains
- **Building System**: 301 buildings (250 lineage buildings + 21 storage + 22 Wonders + 4 cultural monuments + 4 standalone: the Nano Foundry and 3 diplomatic buildings) with scaling costs and construction queues
- **Worker System**: 12 domains (food, faith, knowledge, military, trade, engineering, hacker, astronaut, lumber, masonry, metallurgy, energy); workers arrive on their own and follow your worker shares (`workers share knowledge 40`), with auto-recruit as housing and food allow
- **Tech Tree**: 202 technologies with prerequisites and permanent bonuses
- **Military**: 13 campaigns that cost soldiers and 3 scouting expeditions that cost resources; your garrison's Defense Rating blunts raids, war raids and catastrophe losses
- **Epochs and Catastrophes**: 7 epochs; from the Iron Era on a doom may be fated to strike at any moment of an era, and a harbinger comes first (Appease, Brace or Invite). When it strikes you Endure or Succumb, and Succumb's legacy bonuses carry across runs. In the Cosmic Era prestige itself is the Last Passage
- **Build Plan**: queue builds, techs, trades and an advance (up to 60 items); the game starts each one as the resources come in, and production a full store would waste goes to the current wonder, then to the plan
- **The Map**: your town drawn from your real buildings, in two styles (a roguelike glyph world and a side-on skyline), changing with every age, with a mini map on the dashboard
- **Random Events**: 61 events (26 base + 35 epoch-exclusive) with streak balancing
- **Milestones**: 77 milestones across 6 chains with civilization titles and temporary speed boosts
- **Age Progression**: 22 ages from Primitive to Transcendent with exponential requirements; on each advance, lineage buildings with a next tier can be upgraded to it (`upgrade`)
- **Trade System**: 21 trade routes and resource exchange with supply/demand pressure
- **Diplomacy**: 11 civilizations with opinion tracking, gifts, alliances, rotating faction trade deals, boons and setbacks
- **Prestige**: Reset-and-grow system from the Medieval Age on. Points pay by depth (every completed age pays, 3 times as much for each era: 9 from the Medieval Age, 120 from the Modern Age) and buy the legacy kit, 3 kit items that carry your build plan, worker shares and the civilizations you met into later runs. Era Mastery: every age a run completes runs faster on later runs, up to 4.2x
- **Command-driven interface**: everything is typed at one prompt, with completion as you type; panels (research, plan, workers, army, trade, factions, stats, wonders, logs, epoch, harbinger, map and more) open by name, and `help` lists every command
- **Themes and accounts**: 11 themes, dark and light, including colorblind-safe and high-contrast palettes; local accounts with backups and a recovery code
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
1. `gather wood` and `gather food`: top up by hand while production is small
2. `build gathering_camp` and `build wood_camp`: your first food and wood producers
3. `build hut`: a hut costs 14 wood and gives +10 housing; workers arrive on their own to staff your buildings, as housing and food allow
4. `workers`: see them, and steer the split with `workers share knowledge 40`
5. Bank the Sacred Grove, the Primitive Age's wonder, then `advance`

The [first-age walkthrough](https://ageforge.io/docs/#/first-ten-minutes) goes step by step.

### Commands
Everything is typed at the prompt. The full reference is the in-game `help` panel and the [Commands](https://ageforge.io/docs/#/commands) wiki page; the ones you will use most:

- `build <building> [count|max]`, `research <tech>`, `advance`
- `plan`: the build plan (`plan build`, `plan research`, `plan advance`)
- `workers share <domain> [percent|auto]`: steer where your workers go
- `wonder collect <resource|all> [amount|max]`: bank resources into the current wonder
- `trade`, `factions`, `army`, `map`, `stats`: open a panel
- `prestige`: view prestige status (it opens in the Medieval Age, and a run to the Modern Age pays far more); `prestige shop` and `prestige buy <item>` for the legacy kit
- `save [name]` / `load [name]`, and `quit` to save and quit

### Keys
- Tab or →: take the completion shown in dim text after the cursor; Tab again cycles through the others
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

Produces: `feat: add iron smeltery building` with the bullets as body. These subjects feed the release notes.
