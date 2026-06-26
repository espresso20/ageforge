# Contributing to AgeForge

AgeForge is a terminal-based idle empire builder written in Go.

---

## Requirements

- Go 1.23+

```bash
git clone https://github.com/espresso20/ageforge.git
cd ageforge
go build -o ageforge .
./ageforge
```

---

## Dev Scripts

```bash
make check        # build + vet + config validation
make test         # build + vet + full test suite (formatted output)
make test-raw     # build + vet + tests (raw go test -v, for CI/piping)
make run          # build + run the game
make clean        # remove binary
make commit       # interactive conventional commit (see below)
make release-patch   # bump patch version, update CHANGELOG, tag, push
make release-minor   # bump minor version, update CHANGELOG, tag, push
make release-major   # bump major version, update CHANGELOG, tag, push
```

---

## Writing Commits

Use `make commit` instead of `git add . && git commit && git push`. It stages everything, walks you through the message interactively, and enforces the format that drives automatic release notes.

```
make commit
```

Example session:

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

### Commit Types

| Type | Use for | Shows up in release notes as |
|---|---|---|
| `feat` | New game content, new commands, new UI features | `### Added` |
| `fix` | Bug fixes — wrong behavior, crashes, display errors | `### Fixed` |
| `balance` | Tuning numbers — costs, rates, durations, caps | `### Balance` |
| `refactor` | Code cleanup with no behavior change | `### Changed` |
| `chore` | Build scripts, CI, tooling, deps | *(skipped)* |
| `docs` | README, comments, wiki pages only | *(skipped)* |

### Rules

- **Subject line ≤ 72 chars.** The script enforces this and re-prompts if you go over. This is what shows in `git log` and on GitHub.
- **Write in plain English.** The script lowercases the first letter and strips trailing periods. Just describe what changed.
- **Use bullet points for multiple changes.** Enter them one per line at the details prompt. Blank line to finish. Details go into the commit body and are readable in `git log`.
- **One concern per commit.** If you changed a bug fix and a balance tweak, make two commits. Release notes are cleaner and reverting is safer.

### Examples

**Good — focused fix:**
```
fix: knowledge rate was displaying +0.0 for values below 0.1
```

**Good — balance change with details:**
```
balance: rebalance primitive age pacing

- hut build time raised from 10 to 20 ticks
- altar knowledge output raised from 0.004 to 0.008
- stash max count raised from 10 to 50
```

**Good — new feature:**
```
feat: manual age advancement — type 'advance' when ready
```

**Bad — too vague:**
```
fix: stuff
```

**Bad — too long for a subject line, no detail separation:**
```
balance: stash now has max count of 50, all buildings in primitive take longer to build, altar production raised from .004 to .008 knowledge
```

---

## Release Process

Releases are cut manually when ready. From `master` with a clean working tree:

```bash
make release-patch   # v2.4.5 → v2.4.6  (bug fixes, balance tweaks)
make release-minor   # v2.4.5 → v2.5.0  (new features, new content)
make release-major   # v2.4.5 → v3.0.0  (breaking changes, save format changes)
```

**When to use which:**
- `patch` — fixes, balance changes, small improvements. No new gameplay systems.
- `minor` — new commands, new ages/buildings/techs/mechanics. Backwards-compatible saves.
- `major` — save format changes, full system rewrites, anything that could break existing saves.

**What the script does** (`scripts/release.sh`):
1. Validates you're on `master` with a clean working tree
2. Scrapes `git log` since the last tag, groups commits by type into `### Added / Fixed / Balance / Changed / Other`
3. Stamps `CHANGELOG.md` with the new version block and generated notes
4. Commits, tags (`vX.Y.Z`), and pushes to `origin/master`
5. GitHub Actions picks up the tag and: builds 5 binaries (Linux/macOS/Windows × amd64/arm64), generates `SHA256SUMS.txt`, creates the GitHub Release with the changelog as the body, and posts a Discord embed

Version is baked in at build time: `go build -ldflags "-X main.version=vX.Y.Z"`. Dev builds report `dev`.

### How the release pipeline works

This is a **tag-based release** pattern. The git tag is the source of truth for what has shipped. Here's the full chain:

```
make release-patch
      │
      ├─ bumps version, writes CHANGELOG.md
      ├─ git commit "chore: release vX.Y.Z"
      ├─ git tag vX.Y.Z
      └─ git push origin master + vX.Y.Z
                              │
                              ▼
                     GitHub Actions (release.yml)
                              │
                              ├─ builds 5 binaries
                              ├─ generates SHA256SUMS.txt
                              ├─ creates GitHub Release (with changelog body)
                              └─ posts Discord embed
```

### What happens when a build fails

All commits since the **last successful tag** accumulate and get picked up by the next release — the script runs `git log <last-tag>..HEAD` to scrape commit messages. This is sometimes called a **release train**: commits queue up and ship together on the next run that succeeds.

**If the Actions build fails after the tag was already pushed**, you have two options:

**Option 1 — Re-run the workflow (preferred for infra failures)**

Go to the [Actions tab](https://github.com/espresso20/ageforge/actions), find the failed run, and click **Re-run jobs**. The tag already exists so GitHub re-triggers the same job on the same tag. No new commit or tag needed. Use this when the failure was environmental — a flaky dependency download, a runner hiccup, a typo in a config file that you've since fixed and pushed.

**Option 2 — Delete the tag and re-release (for code bugs caught post-tag)**

Use this if you caught a real bug in the code after tagging but before anyone downloaded it.

```bash
git tag -d vX.Y.Z                      # delete local tag
git push origin :refs/tags/vX.Y.Z      # delete remote tag
# fix the bug, commit it
make release-patch                      # cuts a fresh vX.Y.Z+1
```

### What the tag represents

Once a tag exists in git, that version is considered "attempted." Even if the GitHub Release was never created (build failed), the tag marks that point in history. The next `make release-*` will always bump from the latest tag, so you will never accidentally re-release the same version number.

### Checking release status

```bash
git tag --sort=-version:refname | head -5    # recent tags
gh run list --workflow=release.yml            # recent workflow runs
gh release list                               # published GitHub releases
```

---

## Running Tests

The test suite covers all game systems with **86 tests** across 11 files:

| File | Tests | What it covers |
|------|-------|----------------|
| `config/validate_test.go` | 12 | Cross-validates all config keys — no bad references, no duplicates, all content reachable |
| `game/resources_test.go` | 7 | Add, storage cap, remove, pay/afford, rates, unlock, save/load |
| `game/buildings_test.go` | 5 | Unlock, cost scaling, pop capacity, get all, load counts |
| `game/villagers_test.go` | 9 | Recruit, cap limits, assign/unassign, food drain, production, soldiers, save/load |
| `game/research_test.go` | 9 | Start, afford check, age gating, prereqs, tick completion, bonuses, cancel, duplicate, save/load |
| `game/milestones_test.go` | 8 | First shelter, population, age gating, chains, titles, snapshots, hidden visibility, save/load |
| `game/prestige_test.go` | 5 | Can prestige, point calc, diminishing returns, level grants, save/load |
| `game/progress_test.go` | 5 | Age order, next age, display names, advancement check, requirements |
| `game/bus_test.go` | 4 | Subscribe/publish, multiple subscribers, no subscribers, event isolation |
| `game/events_test.go` | 3 | Inject event, expiration, save/load |
| `game/engine_test.go` | 19 | Full integration: init, gather, build, recruit, assign, research, speed, reset, milestones, save/load |

The **config validation tests** are the primary safety net. They cross-reference every string key in every config file against the canonical key lists — a typo like `"foods"` or `"woodcutter_camps"` anywhere will fail the test.

```bash
make test                                              # full suite, formatted
make test-raw                                          # raw go test -v
go test ./game/ -run TestEngine_BuildMultiple -v       # single test
go test ./game/ -v -count=1                            # one package
```

**Common test patterns:**
- Tests create isolated managers — no shared state between tests
- Resource tests must respect `BaseStorage` caps — use `AddStorage()` before `Add()` for large amounts
- Milestone tests use `NewProgressManager().GetAgeOrder()` for the full age map
- Engine tests access internals via `ge.mu.Lock()` for setup, then public API for assertions
- Save/load tests defer `os.Remove(...)` for cleanup and verify full round-trip

---

## Project Structure

```
config/         Data definitions — ages, buildings, techs, resources, milestones,
                events, trade, diplomacy, prestige. Pure data, no logic.
game/           Engine, managers, tick loop. No UI imports.
ui/             tview TUI. Reads GameState snapshots. Never writes to engine.
scripts/        release.sh, commit.sh — dev tooling
main.go         Entry point — wires engine + UI.
```

---

## Key Patterns

- **Config-Driven Content**: All game content is data in `config/`. Add buildings, techs, ages, events there — not in logic files.
- **Manager Pattern**: Each system has its own manager with a clean API. No cross-manager direct calls.
- **GameState Snapshot**: `engine.GetState()` returns a read-only snapshot. The UI refreshes from snapshots every 500ms and never touches engine internals.
- **Event Bus**: Systems communicate via `game.EventBus` (pub/sub, synchronous under write lock). Subscribe in `ui/dashboard.go` for toasts, in managers for cross-system reactions.
- **No Global State**: Pass dependencies explicitly. No singletons.

### Critical: Event Bus Deadlock

Bus handlers run synchronously under the engine's write lock. **Never call `engine.GetState()` or any lock-acquiring method inside a bus subscriber.** Use `config.*ByKey()` functions (pure data, no locks) for any lookups inside handlers.

---

## Developer Console

A hidden dev console is available for playtesting without grinding through all 22 ages.

**Unlock:** Press `Ctrl+K` anywhere in the dashboard → type the developer passphrase → press Enter. If correct, a **Dev** tab appears in the tab bar (`F10`).

**Access:** `F10` or `` ` `` (backtick). The normal `>` game input still works everywhere even while in the Dev tab.

| Command | Effect |
|---|---|
| `/ages` | List all 22 age keys |
| `/age <key>` | Jump to any age instantly |
| `/fill` | Fill all resources to storage cap |
| `/give <resource> <amount>` | Add a specific resource |
| `/techs` | Unlock all techs up to current age |
| `/build <key>` | Instantly place any building |
| `/prestige <n>` | Set prestige level 0–9 |
| `/speed <n>` | Set tick speed multiplier |
| `/god` | Toggle godmode — zero costs, instant builds |

The passphrase is stored as a SHA256 hash in `game/devmode.go` — never plain text. Dev mode never persists to disk; it resets on every restart.

---

## Adding Content

**New building** — add a `BuildingDef` to `config/buildings.go` with `BaseCost`, `CostScale`, `BuildTicks`, `Category`, and `Effects`. Unlock it in the matching age's `UnlockBuildings` in `config/ages.go`. Cost formula: `floor(BaseCost × CostScale^count)`. Typical `CostScale`: 1.25–1.6.

**New tech** — add a `TechDef` to `config/techs.go` with `Age` (gate), `Cost` (knowledge), `Prerequisites`, and `Effects`. Effect types: `"production"` (flat per-tick) or `"bonus"` (multiplier, e.g. `"gold_rate"`, `"tick_speed"`).

**New age** — add an `AgeDef` to `config/ages.go` with `ResourceReqs`, `BuildingReqs`, `UnlockResources`, `UnlockBuildings`, `UnlockVillagers`. Add a matching age milestone in `config/milestones.go` with `Category: "ages"` and `MinAge` set. On age transition, each resource is capped to ~`carryoverStarterBuildings` (8) of the cheapest new-age building that uses it (via `config.AgeEntryCosts`); resources no new-age building uses as a build cost keep `carryoverResidualPct` (10%, e.g. food — avoids a starvation spiral at the transition); amounts already below the cap are preserved; faith is exempt (cumulative). See `advanceAge` in `game/engine.go`.

Write the **raw** `ResourceReqs`/`BuildingReqs` you want as the *baseline* — they are not the final gate. `Ages()` runs every def through `normalizeAgeRequirements` (in `config/ages.go`) before returning, so all consumers (`AgeByKey`, `AgeOrder`, and `CheckAdvancement`) see the scaled values. The scaling (EPIC economy-rebalance sub-ticket 3, deliberately moderate because the cost-curve and carryover fixes already tightened pacing; tunable in that function): resource reqs are multiplied by a per-band factor — **2.0x** for stone/bronze/iron, **1.75x** for classical/medieval/renaissance, **1.5x** for colonial/industrial/victorian, **1.25x** for electric → transcendent — then rounded to 2 significant figures. Building reqs below **5** are raised to 5 for the early/mid ages (stone_age through information_age); digital_age onward and primitive_age are untouched.

**New milestone** — add a `MilestoneDef` to `config/milestones.go`. Set `Hidden: true` if it should only appear when progress > 50%. To include it in a chain, add its key to the chain's `MilestoneKeys` in `MilestoneChains()`. Chain completion auto-grants a title + speed boost.

**New random event** — add an `EventDef` to `config/events.go` with `Sentiment` (good/bad/mixed), `Weight`, `Cooldown`, `Duration` (0 = instant), `MinAge`, and `Effects`. Streak logic caps bad events at 2 consecutive and forces one after 3 good ones.

**New expedition** — add a def to `getExpeditions()` in `game/military.go` with `SoldiersNeeded`, `Duration`, `DifficultyBase`, `Rewards`, `MinAge`. Success: `random() > (DifficultyBase - military_bonus × 0.3)`.

**New trade route** — add a `TradeRouteDef` to `config/trade.go` with `Export`/`Import` maps, `TicksPerRun`, `RequiredBuilding`, `MinAge`. Routes auto-cycle, importing `amount × (1.0 + diplomacy_bonus)`.

**New villager type** — add a `VillagerTypeDef` to `game/villagers.go` with `FoodCost` and `GatherRate`. Unlock it in the matching age in `config/ages.go`.

---

## How the Math Works

### Tick Loop

Base interval: **2 seconds**. Each tick in order: build queue → research → events → expeditions → trade → diplomacy → production → resources → milestones → age check → tick speed recalc.

```
tick_interval = 2000ms / ((1.0 + tick_speed_bonus) × speed_multiplier)
minimum: 200ms
```

`tick_speed_bonus`: research + milestones + prestige (+1%/level) + active chain boosts.
`speed_multiplier`: player-set in 0.5× steps, capped at `1.0 + (wonders_built × 0.5)`.

### Resource Rates (per tick, in order)

1. Base: building production + villager gathering + research effects + event effects
2. `production_all` multiplier on all positive rates
3. Per-resource multiplier (e.g. `gold_rate` bonus)
4. Gather rate bonus: additive on villager rates
5. Diplomacy trade bonuses: multiplicative on positive rates
6. Food drain: `sum(villager_count × food_cost_per_type)` subtracted

### Building Costs

```
cost = floor(base_cost × cost_scale ^ current_count)
```
Example: Hut — 30 wood, scale 1.3: 1st=30, 2nd=39, 3rd=50, 4th=66...
Upgrades cost `target_base_cost × 0.25` (75% discount).

### Food Economy

Workers: 0.10/tick · Soldiers: 0.25/tick · Astronauts: 0.40/tick. Keep ~⅓ of workforce on food.

### Expeditions

```
adjusted_difficulty = max(0.05, base_difficulty - military_bonus × 0.3)
success  = random() > adjusted_difficulty
loot     = base_reward × (1.0 + expedition_bonus)   # success
loot     = base_reward × 0.3                         # failure
```

### Trade & Exchange

```
rate             = base_rate × (1.0 - pressure × 0.3)   # floor: 50% of base
pressure_gain    = 0.1 / (1.0 + market_count × 0.2)
pressure_decay   = pressure × 0.98 per tick
```

### Prestige

Requires Medieval Age+.

```
base   = age_order_index
bonus  = floor(milestones/10) + floor(techs/15) + floor(buildings/50)
points = floor((base + bonus) / sqrt(prestige_level + 1))
```
Each prestige: +2% production, +1% tick speed permanently. 9 upgrades purchasable with points.

### Offline Progress

```
offline_ticks = min(elapsed, 24h) / tick_interval
resource_gain = rate × offline_ticks × 0.5     # 50% efficiency, capped at storage
```

### Milestone Chains

5 chains: Settlement, Scholar, Builder, Military, Ancient Ages. Completing a chain grants a civilization title + a temporary tick speed boost via `InjectEvent()`. Hidden milestones become visible at >50% progress, or (for age milestones) when in the preceding age.

---

## How Systems Connect

```
config/*.go   →   game/*.go                  →   ui/*.go
──────────        ─────────                      ───────
BuildingDef   →   BuildingManager.Build()    →   EconomyTab.Refresh()
                  engine.processBuildQueue()
                        │
                        ▼
                  EventBus.Publish(BuildingBuilt)
                        │
                        ├──▶ dashboard.go → ToastManager.Show()
                        └──▶ MilestoneManager.Check()
                                    │
                                    ▼ (chain completes)
                             EventManager.InjectEvent(speed boost)
```

Flow: **Config** defines data → **Manager** owns state/logic → **Engine** orchestrates in `doTick()` → **EventBus** notifies other systems (synchronous, under write lock) → **UI** reads `GetState()` snapshots every 500ms, never writes.

---

## Conventions

- Package names: lowercase single word (`config`, `game`, `ui`)
- Config keys: `snake_case` strings (`"lumber_mill"`, `"stone_age"`)
- `float64` for resource amounts, `int` for building counts
- Return errors up, log at boundaries
- Keep changes minimal — don't refactor code you didn't need to touch

---

## Project Tracking

All active work is tracked on the [AgeForge Trello board](https://trello.com/b/tf31C2cz/ageforge).

| Lane | What lives there |
|---|---|
| **Bugs** | One ticket per bug — single fix per card |
| **Features** | New systems, commands, and content (`feat`) |
| **Balance** | Cost, rate, and number tuning (`balance`) |
| **Refactor** | Cleanup with no behavior change (`refactor`) |
| **Doing** | Actively in progress |
| **Done** | Shipped — full history of completed work |
| **Later Enhancements** | Shelved ideas for future consideration |

When picking up work: move the card to **Doing**. When done: move it to **Done** and close the PR.

Commit types map 1:1 to board lanes — if your change is a `fix`, it came from the Bugs lane. If it's `balance`, it came from Balance. This keeps the board and git history in sync.

---

## Reporting Bugs

Found a bug? [Open a GitHub Issue](https://github.com/espresso20/ageforge/issues) with what you were doing, what you expected, what happened, and your OS + terminal. It will be triaged into a Trello bug ticket.

## Questions

Join the [Discord](https://discord.gg/EPvyd5vjpj) or check the [project board](https://trello.com/b/tf31C2cz/ageforge) to see what's being worked on.
