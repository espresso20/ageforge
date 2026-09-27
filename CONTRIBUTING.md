# Contributing to AgeForge

AgeForge is a terminal-based idle empire builder written in Go.

---

## Requirements

- Go 1.24+

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

CI (`.github/workflows/go.yml`) runs `gofmt -l`, `go vet`, `go build`, a cross-compile of the release targets and `go test -race` on every PR and every push to `master`. Run `gofmt -w .` before pushing.

### Smoke testing

The unit tests prove pieces work. The smoke suite proves the game holds together: it boots the real engine with no UI and plays it, saves and reloads it, closes it for hours, types garbage at it, and checks the docs against it. It is one runner (`cmd/smoke`, package `smoke`) with named scenarios that share a report. Ticks run synchronously through `GameEngine.StepTicks`, so days of play take seconds, and seeds are fully reproducible: the same seed gives the same report.

```bash
make smoke                                   # fast tier, what every PR runs (a few minutes)
make smoke-full                              # full tier, what the nightly runs
go run ./cmd/smoke -list                     # the scenarios
go run ./cmd/smoke -scenario saveload -v     # one scenario (comma-separate several)
go run ./cmd/smoke -scenario progression -seed-base 7 -seeds 1 -trace -v   # one seed, every bot action in smoke-report/trace-7.log
go run ./cmd/smoke -scenario styles -style idle -tier full   # one play style
go run ./cmd/smoke -h                        # every flag
```

#### Scenarios

| scenario | tier | what it checks |
|---|---|---|
| `static` | fast | The Gate Covenant from config alone: every age gate fits storage and every resource it needs has a source (also `go test ./smoke`). |
| `docsync` | fast | Every "N ages / buildings / technologies / ..." claim in `site/index.html` (hero stats included), `README.md`, `site/docs/README.md` and the opening lines of each wiki page matches the config count; `buildings.md`'s lineage heading and table match the lineage count; `commands.md` documents every command in `ui/input.go`'s `HandleCommand` and every subcommand the autocompleter offers, and documents nothing the handler doesn't accept. Fix the docs, not the check. |
| `progression` | fast: 3 seeds to the Bronze Age; full: 8 seeds, 2 prestige cycles, on to the Digital Age | The greedy bot (public engine methods only: no dev commands, god mode or speed changes) plays each seed. Fails on panics, soft-locks (30 simulated minutes with no new building, research or rising resources) and invariant violations (checked every 25 ticks: negative or non-finite values, resources over their cap, worker bookkeeping, a stale pending catastrophe, a requirement no storage can hold, a prestige that pays other than the documented formula or loses a legacy, ruins, upgrades or the Cosmic Legacy). Grades every age against the pacing table. |
| `saveload` | fast: 2 checkpoints; full: 4 on 2 seeds | At one checkpoint per age, saves through `SaveGame`, loads into a fresh engine and plays the same N ticks with a fresh bot. The reload must equal the saved state, re-save to the same data (in the same order: a set written in map order fails) and pass the signature check (a tampered copy must fail it), and the continued game must equal the uninterrupted run. The first divergent field is reported, with a replay that tells an RNG restart from lost state. |
| `offline` | fast: 1 base state; full: 3 | Closes the game for 1h, 8h, 24h and 30h through the offline catch-up (`GameEngine.SimulateOffline`, the function `LoadGame` runs), plus one end-to-end load of a save written 8h ago. Gains must be positive, within rate x ticks x 50%, under the caps, capped at 24h (30h pays exactly what 24h does), and the invariants must hold while the game ticks on. |
| `fuzz` | fast: 300 commands; full: 25,000 on 2 seeds | Random commands through the real handler (`ui.HandleCommand`): valid commands walked through the autocompleter's suggestions, valid commands with hostile arguments (huge, negative, NaN, unicode, path-like), mangled lines and garbage, with bot play between batches so later commands meet a real game. No panics, no hangs, invariants hold, ticks keep counting. A failure is replayed and shrunk to a short command list. |
| `accounts` | fast | In a temp data dir: create two accounts with a save each, switch, export and import (the export must round-trip; tampered or garbage exports are refused), back up, recover from a recovery code (garbage and typo'd codes refused), wipe one account (the other's files stay byte-identical, a backup is written), wipe the active one and restore it from its export. |
| `perf` | fast | The late-game `BenchmarkTick` and `BenchmarkGetState` (in `game/tick_perf_test.go`) against budgets (250µs and 1ms, generous for CI runners; actuals are reported), and a long run that samples the heap after GC and the size of every collection in `GameState`, failing on unbounded growth. |
| `ui` | fast: the default and one light theme at 80x24; full: every theme at 80x24 and 100x30 | Runs `TestSmokeUISweep` (every theme at 180x56: splash pages, a new game, every overlay and read-only command) and `TestSmokeUISmallTerminals` (the dashboard and every overlay at small sizes, then live resizes) from `ui/smoke_sweep_test.go` (build tag `smoke`). A panic, a frozen event loop or a blank screen fails it. |
| `styles` | full | The bot under five styles, 2 seeds each: `greedy`, `idle` (decides only at check-ins every 3 game-hours), `harbinger` (buys Appease and Brace), `succumber` (Succumbs every catastrophe), `cosmic` (Invites the Cosmic Era's harbinger, then Succumbs the Last Passage). Reports pacing and outcomes per style. |
| `prestige` | full | Plays for two prestige cycles (reporting how far it got if the budget runs out first), then drives the same mechanics through the engine's test hooks: a Succumb for a legacy bonus and ruins, prestiges from the Modern Age checked against the formula, upgrades measured against a twin engine that bought none, a succumbed Last Passage for the Cosmic Legacy, and the prestiges it must survive. |

#### Tiers and CI

- **Fast tier** (`make smoke`, `-tier fast`): the scenarios above marked fast. It runs on every pull request as the `smoke (fast tier)` job in `.github/workflows/go.yml` (skipped for doc-only changes, like the rest of that workflow), writes a summary to the job page and uploads `smoke-report/` as an artifact.
- **Full tier** (`make smoke-full`, `-tier full`): every scenario, deeper. It runs nightly in `.github/workflows/smoke.yml` (and on demand from the Actions tab) with the same summary and artifact.
- Both fail on panics, soft-locks, invariant violations, save/load divergence, fuzz failures, account failures, docs mismatches and blown perf budgets, and both run with `-pacing enforce`, so a first-cycle age of the progression scenario whose median across seeds leaves its pacing band fails them too (see Pacing targets below).
- **Known bugs** can be reported as warnings instead of failures so the job stays useful while they wait for a fix (`-strict` fails on them too); today there are none. When a known bug is fixed, delete its special case in the scenario.

#### Reading the report

Everything lands in `smoke-report/` (`-out` to change it):

- `summary.md`: the verdict, one row per scenario, the pacing table and the failures. It is what the CI job page and the email show.
- `report.md`: the full report, one section per scenario with every failure (message, repro command, and a state dump or stack under "detail"), the warnings, and the scenario's tables.
- `report.json`: the same for scripts.
- `progression.md`, `style-<name>.md`, `prestige-played.md`: the full bot report for each set of runs: outcome per seed, pacing, events and catastrophes, harbinger threads, bot actions, and each anomaly with a dump that names what was blocking the next age.

Before calling a soft-lock a bug, check whether a sensible player would be stuck too or whether the bot is just not clever enough (`-trace` helps). Improve the bot for the second case.

#### Pacing targets

`smoke/targets.go` holds the pacing contract: the time a player should spend in each age at 1x game time (15 minutes in the Primitive Age up to 12 hours in the Atomic Age, about 3 days to the Modern Age and the first prestige, then 12 to 24 hours per age). An age passes when it takes 0.5x to 2x its target; the report marks each age ✓, slow or fast. The per-age timeout is derived from the same table (4x the target, at least an hour): in report mode an age past it is flagged and play goes on, so a slow age reads as slow instead of killing the run, and only a true no-progress soft-lock fails.

The game derives its economy from the same table (payback times, build and research caps; see `design-and-architecture/economy.md`, Law 2), so the targets live in the game: to change them, edit `config.AgeTargets` in `config/pacing.go`, then make `PacingTargets` in `smoke/targets.go` match it. `TestPacingTargetsMatchConfig` fails while the two differ, and `TestTargetsCoverEveryAge` checks every age but the last has one. `PacingLow`, `PacingHigh`, `TimeoutFactor` and `TimeoutFloor` sit beside `PacingTargets`. Changing a target changes the game, not just the grade: every producer's rate follows it.

**Pacing is enforced in CI.** Both tiers run with `-pacing enforce`: in the progression scenario a first-cycle age whose median across the seeds is outside the band fails the scenario, and any age past its timeout fails its run, so a PR that makes the game meaningfully slower or faster fails its `smoke (fast tier)` check (the fast tier grades the Primitive and Stone Ages) and the nightly grades every age to the prestige. Later cycles run with prestige upgrades and are graded but never fail (the targets describe a first run), an age left by prestige is not a completed age and isn't graded, and the other scenarios always run in report mode (the idle style checks in every three hours; saveload and perf cut their runs short). If a balance change moves an age out of the band on purpose, change the target in the same PR.

#### Email

Both workflows can email the summary (with `report.md` attached when it is under 300 KB) after every run, pass or fail. It is optional: with no secrets the email steps skip, and they always skip on pull requests from forks, which get no secrets. To turn it on, add three repository secrets (Settings → Secrets and variables → Actions):

- `MAIL_USERNAME`: the Gmail address that sends the mail.
- `MAIL_PASSWORD`: a Gmail [app password](https://support.google.com/accounts/answer/185833) for it, not the account password.
- `MAIL_TO`: where the mail goes (comma-separate several).

To keep the nightly emails but stop the per-PR ones, add a repository variable `SMOKE_EMAIL_PR` set to `false`.

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
2. `production_all` multiplier — `× (1 + Σ production_all)`, applied via the resolver. The sum may be **negative** (e.g. the catastrophe Reconstruction Effort −10% debuff); a former `>0` gate that silently dropped negative pools is gone, so debuffs now apply. Result is floored at 10% of base so production can't be driven below it.
3. Per-resource multiplier (e.g. `gold_rate` bonus)
4. Gather rate bonus: additive on villager rates
5. Diplomacy trade bonuses: multiplicative on positive rates
6. Food drain: `sum(villager_count × food_cost_per_type)` subtracted

### Building Costs

```
cost = floor(base_cost × cost_scale ^ current_count) × (1 + Σ build_cost)
```
Example: Hut — 30 wood, scale 1.3: 1st=30, 2nd=39, 3rd=50, 4th=66...

`build_cost` modifiers are now consumed by `GetCost()`: build-cost-reducing milestone rewards and the Civil Engineering tech (each −3% to −5%) sum into `Σ build_cost` and multiply the scaled cost, floored at 10% of base. Previously these were defined but applied nowhere — `GetCost()` is the consumer the resolver migration wired in (see `design-and-architecture/multiplier-system.md` bug #2). Both the displayed and charged cost reflect the reduction.

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
