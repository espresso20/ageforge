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

# Check version
./ageforge --version

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
- F10 — Dev console tab (only visible after developer unlock)
- ESC — auto-save and return to menu
- Arrow keys / PgUp/PgDn — navigate wiki (in Wiki tab)
- v — toggle verbose logs (in Logs tab)

## Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md) for the full dev guide — commit workflow, release process, test patterns, project structure, adding content, and how the math works.

Quick reference:

```bash
make commit          # interactive commit helper (stages, prompts type + message, pushes)
make check           # build + vet + config validation
make test            # full test suite
make release-patch   # cut a patch release
make release-minor   # cut a minor release
make release-major   # cut a major release
```

### Writing Commits

Use `make commit` instead of `git add . && git commit && git push`. It stages everything, walks you through the message interactively, and enforces the format that drives automatic release notes.

```
make commit
```

```
┌── Staged Changes ──────────────────────────────────────────────┐
│   config/buildings.go | 4 ++--
│   game/engine.go      | 6 +++---
└────────────────────────────────────────────────────────────────┘

What kind of change?
  1  feat      — new feature or content          → ### Added
  2  fix       — bug fix                         → ### Fixed
  3  balance   — tuning costs, rates, numbers    → ### Balance
  4  refactor  — cleanup, no behavior change     → ### Changed
  5  chore     — build/tooling/deps              (skipped in notes)
  6  docs      — docs/comments only              (skipped in notes)

  Choice [1-6, default 1]: 2

  Short summary (≤72 chars): stash building count was capped at 1 instead of max

  Details / bullet points? (blank line to finish, skip with Enter)
  · BuildMultiple inQueue check was comparing bool instead of counting
  · fix applies to all buildings with MaxCount > 0
  ·

┌── Commit Message ──────────────────────────────────────────────┐
│  fix: stash building count was capped at 1 instead of max
│
│  - BuildMultiple inQueue check was comparing bool instead of counting
│  - fix applies to all buildings with MaxCount > 0
└────────────────────────────────────────────────────────────────┘

  Commit? [Y/n]: y
  Push to origin/master now? [Y/n]: y
```

#### Commit Types

| Type | Use for | Shows up in release notes as |
|---|---|---|
| `feat` | New game content, new commands, new UI features | `### Added` |
| `fix` | Bug fixes — wrong behavior, crashes, display errors | `### Fixed` |
| `balance` | Tuning numbers — costs, rates, durations, caps | `### Balance` |
| `refactor` | Code cleanup with no behavior change | `### Changed` |
| `chore` | Build scripts, CI, tooling, deps | *(skipped)* |
| `docs` | README, comments, wiki pages only | *(skipped)* |

#### Rules

- **Subject line ≤ 72 chars.** The script enforces this — it re-prompts if you go over. This is what shows in `git log` and GitHub.
- **Write in plain English.** The script lowercases the first letter and strips trailing periods. Just describe what changed.
- **Use bullet points for multiple changes.** Enter them one per line at the details prompt. Blank line to finish. Details go into the commit body and are readable in `git log --format=medium`.
- **One concern per commit.** If you changed both a bug fix and a balance tweak, make two commits. The release notes are cleaner and `git bisect` actually works.

#### Examples

**Good — a focused fix:**
```
fix: knowledge rate was displaying +0.0 for values below 0.1
```

**Good — a balance change with bullet details:**
```
balance: rebalance primitive age pacing

- hut build time raised from 10 to 20 ticks
- altar knowledge output raised from 0.004 to 0.008
- stash max count raised from 10 to 50
```

**Good — a new feature:**
```
feat: manual age advancement — type 'advance' when ready
```

**Bad — too vague:**
```
fix: stuff     ← doesn't say what broke or what changed
```

**Bad — too long for subject, no detail separation:**
```
balance: stash now has max count of 50, all buildings in primitive take longer to build, altar production raised from .004 to .008 knowledge
← this wraps in every tool and dumps everything into one line
```

---

### Release Process

Releases are cut manually when ready. From `master` with a clean working tree:

```bash
make release-patch   # v2.4.5 → v2.4.6  (bug fixes, balance tweaks)
make release-minor   # v2.4.5 → v2.5.0  (new features, new content)
make release-major   # v2.4.5 → v3.0.0  (breaking changes, major redesigns)
```

**When to use which:**
- `patch` — fixes, balance changes, small improvements. No new gameplay systems.
- `minor` — new commands, new ages/buildings/techs/mechanics. Backwards-compatible saves.
- `major` — save format changes, full system rewrites, anything that could break existing saves.

**What happens when you run it** (`scripts/release.sh`):
1. Validates you're on `master` with a clean working tree
2. Scrapes `git log` since the last tag, groups commits by type into `### Added / Fixed / Balance / Changed / Other`
3. Stamps `CHANGELOG.md` with the new version block and generated notes
4. Commits, tags (`vX.Y.Z`), and pushes to `origin/master`
5. GitHub Actions picks up the tag and: builds 5 binaries (Linux/macOS/Windows × amd64/arm64), generates `SHA256SUMS.txt`, creates the GitHub Release with the changelog as the body, and posts a rich Discord embed

Version is baked into the binary at build time. Dev builds report `dev`.
