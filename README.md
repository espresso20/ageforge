<p align="center">
  <a href="https://ageforge.io">
    <img src="https://raw.githubusercontent.com/espresso20/ageforge/master/ageforge-3.png" alt="AgeForge" width="200">
  </a>
</p>

<h1 align="center">AgeForge</h1>

<p align="center">
  <em>An almanac of twenty-two ages, begun at a campfire and kept in your terminal.</em>
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

AgeForge is an idle civilization game for the terminal. You are handed a few people, a fire, 25 food and 50 wood, and from there you raise a civilization through 22 ages, from ten thousand years before Christ to a year the calendar cannot print, one typed command at a time. No age lets you go until its Wonder stands. A harbinger comes before every doom, and whether you brace for it or let the whole civilization fall and carry its lessons into the next is yours to decide. Eleven neighbors wait past the edge of a map grown from its own seed, and the town keeps working while you are away.

> Version note: this README and the [wiki](https://ageforge.io/docs/) describe the game on master, version 4.0, which is coming soon. The latest release is 3.6, which still has some older systems (the `speed` command, the old city map).

## Overview

A new game starts in the Primitive Age. Gather, build, let workers find their own way to your buildings, research, send missions from the mission board, trade and treat with other civilizations, raise each age's Wonder and advance. Write a build plan and the game spends your income on it while you play and while you are away (up to 24 hours, at half rate); when you come back, the log tells you what happened, in order.

## Features

- **Resource Management**: 24 resources across 22 ages, one resource to each building line; stock has no cap
- **Building System**: 302 buildings (251 lineage buildings + 21 storage + 22 Wonders + 4 cultural monuments + 4 standalone: the Nano Foundry and 3 diplomatic buildings) with scaling costs and construction queues
- **Worker System**: 11 domains (food, faith, military, trade, engineering, hacker, astronaut, lumber, masonry, metallurgy, energy); workers arrive on their own and follow your worker shares (`workers share faith 40`), with auto-recruit as housing and food allow
- **Tech Tree**: 203 technologies with prerequisites and permanent bonuses
- **Missions and the Army**: a mission board of nine offers (short, long and grand), each rolled up front with its length, odds and haul, with road events on the way and a rare Legendary offer; grand military missions cost soldiers, and your garrison's Defense Rating blunts raids, war raids and catastrophe losses
- **Epochs and Catastrophes**: 7 epochs; from the Iron Era on a doom may be fated to strike at any moment of an era, and a harbinger comes first (Appease, Brace or Invite). When it strikes you Endure or Succumb, and Succumb's legacy bonuses carry across runs. In the Cosmic Era prestige itself is the Last Passage
- **Build Plan**: queue builds, techs, trades and an advance (up to 60 items); the game starts each one as the resources come in, and stock has no limit, so a price is the only thing that holds a build back
- **The Map**: your town drawn from your real buildings, in two styles (a roguelike glyph world and a side-on skyline that zooms from the whole town to every building's face), changing with every age; each age's Wonder stands on the skyline, and the Wonders panel keeps a gallery of every one you have built; `minimap on` adds a mini map above the Buildings list
- **The Advance panel**: the march of ages with a moving calendar year (10,000 BC to Beyond time), the four pillars to the next age (Knowledge, People, Industry, Monument) and what the next age brings
- **Random Events**: 61 events (26 base + 35 epoch-exclusive) with streak balancing
- **Milestones**: 77 milestones across 6 chains with civilization titles and temporary speed boosts
- **Age Progression**: 22 ages from Primitive to Transcendent with exponential requirements; on each advance, buildings with a next rung in their line (a line never switches resource) can be upgraded to it (`upgrade`)
- **Trade System**: 21 trade routes, harbors and a black market whose deals go through five times in six
- **Diplomacy**: 11 civilizations with opinion tracking, gifts, alliances, rotating faction trade deals, boons and setbacks
- **Prestige**: Reset-and-grow system from the Medieval Age on. Every point ever earned is a lasting production bonus in every age. Plan Template, Roster and Old Friends carry over automatically after the first prestige. Era Mastery: every age a run completes runs faster on later runs, up to 4.2x
- **Command-driven interface**: everything is typed at one prompt, with completion as you type; panels (research, plan, workers, army, trade, factions, stats, wonders, logs, epoch, harbinger, map and more) open by name over the full screen, and `help` lists every command. On a Mac, Option+arrows stand in for Page Up/Down and Home/End
- **VibeMode**: `vibemode on` swaps the game's words for a GenZ script; costs, rules, saves and badges stay exactly as they are
- **Themes and accounts**: 16 themes, dark and light, including colorblind-safe and high-contrast palettes; local accounts with backups and a recovery code
- **Badges**: 556 badges to earn on your account, kept in a badge case (`badges`): one for every age, wonder, lineage, civilization and harbinger, ladders that climb across all your runs, and one-off feats. Badges give points, titles and ten of the themes, and never change a run
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
4. `workers`: see them, and steer the split with `workers share faith 40`
5. Bank the Sacred Grove, the Primitive Age's wonder, then `advance`

The [first-age walkthrough](https://ageforge.io/docs/#/first-ten-minutes) goes step by step.

### Commands
Everything is typed at the prompt. The full reference is the in-game `help` panel and the [Commands](https://ageforge.io/docs/#/commands) wiki page; the ones you will use most:

- `build <building> [count|max]`, `research <tech>`, `advance`
- `plan`: the build plan (`plan build`, `plan research`, `plan advance`)
- `workers share <domain> [percent|auto]`: steer where your workers go
- `wonder collect <resource|all> [amount|max]`: bank resources into the current wonder
- `trade`, `factions`, `army`, `map`, `stats`: open a panel
- `prestige`: view prestige status, its lasting production bonus and carry-overs; it opens in the Medieval Age, and a run to the Modern Age pays far more
- `save [name]` / `load [name]`, and `quit` to save and quit

### Keys
- Tab or →: take the completion shown in dim text after the cursor; Tab again cycles through the others
- ↑ / ↓: step through command history
- PgUp / PgDn: scroll the main view (Option+↑ / Option+↓ on a Mac, which has no Page keys; Option+← / Option+→ are Home / End, and - and = zoom the Map)
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
