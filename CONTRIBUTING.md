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
  1  feat      - new feature or content          → ### Added
  2  fix       - bug fix                         → ### Fixed
  3  balance   - tuning costs, rates, numbers    → ### Balance
  4  refactor  - cleanup, no behavior change     → ### Changed
  5  chore     - build/tooling/deps              (skipped in notes)
  6  docs      - docs/comments only              (skipped in notes)

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
| `fix` | Bug fixes: wrong behavior, crashes, display errors | `### Fixed` |
| `balance` | Tuning numbers: costs, rates, durations, caps | `### Balance` |
| `refactor` | Code cleanup with no behavior change | `### Changed` |
| `chore` | Build scripts, CI, tooling, deps | *(skipped)* |
| `docs` | README, comments, wiki pages only | *(skipped)* |

### Rules

- **Subject line ≤ 72 chars.** The script enforces this and re-prompts if you go over. This is what shows in `git log` and on GitHub.
- **Write in plain English.** The script lowercases the first letter and strips trailing periods. Just describe what changed.
- **Use bullet points for multiple changes.** Enter them one per line at the details prompt. Blank line to finish. Details go into the commit body and are readable in `git log`.
- **One concern per commit.** If you changed a bug fix and a balance tweak, make two commits. Release notes are cleaner and reverting is safer.

### Examples

**Good: focused fix**
```
fix: knowledge rate was displaying +0.0 for values below 0.1
```

**Good: balance change with details**
```
balance: rebalance primitive age pacing

- hut build time raised from 10 to 20 ticks
- altar knowledge output raised from 0.004 to 0.008
- stash max count raised from 10 to 50
```

**Good: new feature**
```
feat: manual age advancement, type 'advance' when ready
```

**Bad: too vague**
```
fix: stuff
```

**Bad: too long for a subject line, no detail separation**
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
- `patch`: fixes, balance changes, small improvements. No new gameplay systems.
- `minor`: new commands, new ages/buildings/techs/mechanics. Backwards-compatible saves.
- `major`: save format changes, full system rewrites, anything that could break existing saves.

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

All commits since the **last successful tag** accumulate and get picked up by the next release: the script runs `git log <last-tag>..HEAD` to scrape commit messages. This is sometimes called a **release train**: commits queue up and ship together on the next run that succeeds.

**If the Actions build fails after the tag was already pushed**, you have two options:

**Option 1: Re-run the workflow (preferred for infra failures)**

Go to the [Actions tab](https://github.com/espresso20/ageforge/actions), find the failed run, and click **Re-run jobs**. The tag already exists so GitHub re-triggers the same job on the same tag. No new commit or tag needed. Use this when the failure was environmental: a flaky dependency download, a runner hiccup, a typo in a config file that you've since fixed and pushed.

**Option 2: Delete the tag and re-release (for code bugs caught post-tag)**

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
| `config/validate_test.go` | 12 | Cross-validates all config keys: no bad references, no duplicates, all content reachable |
| `game/resources_test.go` | 7 | Add, storage cap, remove, pay/afford, rates, unlock, save/load |
| `game/buildings_test.go` | 5 | Unlock, cost scaling, pop capacity, get all, load counts |
| `game/villagers_test.go` | 9 | Recruit, cap limits, assign/unassign, food drain, production, soldiers, save/load |
| `game/research_test.go` | 9 | Start, afford check, age gating, prereqs, tick completion, bonuses, cancel, duplicate, save/load |
| `game/milestones_test.go` | 8 | First shelter, population, age gating, chains, titles, snapshots, hidden visibility, save/load |
| `game/prestige_test.go` | 5 | Can prestige, point calc, diminishing returns, level grants, save/load |
| `game/progress_test.go` | 5 | Age order, next age, display names, advancement check, requirements |
| `game/bus_test.go` | 4 | Subscribe/publish, multiple subscribers, no subscribers, event isolation |
| `game/events_test.go` | 3 | Inject event, expiration, save/load |
| `game/engine_test.go` | 19 | Full integration: init, gather, build, recruit, assign, research, reset, milestones, save/load |

The **config validation tests** are the primary safety net. They cross-reference every string key in every config file against the canonical key lists. A typo like `"foods"` or `"woodcutter_camps"` anywhere will fail the test.

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
make smoke-deep                              # deep tier, what the weekly runs (five seeds to a Quantum Age prestige; about 13 minutes of work per seed on a CI runner)
go run ./cmd/smoke -list                     # the scenarios
go run ./cmd/smoke -scenario saveload -v     # one scenario (comma-separate several)
go run ./cmd/smoke -scenario progression -seed-base 7 -seeds 1 -trace -v   # one seed, every bot action in smoke-report/trace-7.log
go run ./cmd/smoke -scenario styles -style idle -tier full   # one play style
go run ./cmd/smoke -scenario styles -style idle -check-in 8h -max-sim 2000h   # idle check-ins every 8 game-hours, to the first prestige
go run ./cmd/smoke -scenario styles -style idle -no-shares   # the idle style recruiting and assigning by hand (auto-recruit off), to measure what worker shares are worth
go run ./cmd/smoke -scenario progression -army on -v   # from the Iron Age the bot keeps 4 of the age's newest military building, bought when one costs at most a quarter of its stock (off by default: the pacing targets assume no army)
go run ./cmd/smoke -scenario idle -tier full -pacing enforce   # the idle targets: 1h, 3h and 8h check-ins, 3 seeds each, plus the greedy bot on the same seeds
go run ./cmd/smoke -tier deep -merge a/report.json,b/report.json -pacing enforce   # pool per-seed reports and grade the median
go run ./cmd/smoke -tier full -only progression -seed-base 1 -seeds 4 -out shard-1   # one nightly shard: progression seeds 1-4 (-only/-skip narrow what -scenario picks)
go run ./cmd/smoke -tier full -skip progression,styles,idle -pacing enforce -out shard-rest   # the nightly's "rest" shard
go run ./cmd/smoke -tier full -merge shard-1,shard-2,shard-idle,shard-rest -pacing enforce   # merge shards like the nightly: progression pooled and graded, static run again, the rest carried over
go run ./cmd/smoke -h                        # every flag
```

#### Scenarios

| scenario | tier | what it checks |
|---|---|---|
| `static` | fast | The Gate Covenant from config alone: every age gate fits storage and every resource it needs has a source from a cold start, for a player who skipped every building no gate required, and the storage ladder (each age's first storage copy fits in the storage its gate forces); and the Storage Covenant: every age's most storage holds `config.StorageHoldHours` of its typical production; and the Milestone Covenant: every milestone, chain and title can be completed by the end of the last age before prestige opens, or of the milestone's own age when that is later, with each milestone's earliest age and tightest requirement in the report; and every harbinger Appease and Brace price, both levels, fits the most storage buildable in the first age of its era (all also `go test ./smoke`). |
| `docsync` | fast | Every "N ages / buildings / technologies / ..." claim in `site/index.html` (hero stats included), `README.md`, `site/docs/README.md` and the opening lines of each wiki page matches the config count; `buildings.md`'s lineage heading and table match the lineage count; `commands.md` documents every command and subcommand in the command registry (`ui/commands.go`), and documents nothing the registry doesn't list. Fix the docs, not the check. |
| `progression` | fast: 3 seeds to the Bronze Age; full: 8 seeds, 2 prestige cycles, on to the Digital Age; deep: 5 seeds (one per CI job), one cycle to a Quantum Age prestige through the Last Passage (the bot Invites it and Endures) | The greedy bot (public engine methods only: no dev commands, god mode or speed changes) plays each seed. Fails on panics, soft-locks (30 simulated minutes with no new building, research or rising resources) and invariant violations (checked every 25 ticks: negative or non-finite values, resources over their cap, worker bookkeeping, a stale pending catastrophe, storage that can never grow again, a requirement no storage can hold (storage is built up a copy at a time: each copy counts only once the caps before it can pay for it), a prestige that pays other than the documented formula or loses a legacy, ruins, upgrades or the Cosmic Legacy). Grades every age against the pacing table, reports each age's longest quiet stretch, and grades the first run to the Modern Age against the first-run band (see Pacing targets). |
| `saveload` | fast: 2 checkpoints; full: 4 on 2 seeds | At one checkpoint per age, saves through `SaveGame`, loads into a fresh engine and plays the same N ticks with a fresh bot. The reload must equal the saved state, re-save to the same data (in the same order: a set written in map order fails) and pass the signature check (a tampered copy must fail it), and the continued game must equal the uninterrupted run. The first divergent field is reported, with a replay that tells an RNG restart from lost state. |
| `offline` | fast: 1 base state; full: 3 | Closes the game for 1h, 8h, 24h and 30h through the offline catch-up (`GameEngine.SimulateOffline`, the function `LoadGame` runs), plus one end-to-end load of a save written 8h ago. Gains must be positive, within rate x ticks x 50% (the higher of the rates before and after: construction finishes while away), under the caps, capped at 24h (30h pays exactly what 24h does), and the invariants must hold while the game ticks on. Then a day away with a build plan of the age's producers: it must start some, log a summary, keep the invariants, come back identical twice, and take under 2 seconds. |
| `fuzz` | fast: 300 commands; full: 25,000 on 2 seeds | Random commands through the real handler (`ui.HandleCommand`): every registry command, walked through the prompt's completions, valid commands with hostile arguments (huge, negative, NaN, unicode, path-like), mangled lines and garbage, with bot play between batches so later commands meet a real game. No panics, no hangs, invariants hold, ticks keep counting. A failure is replayed and shrunk to a short command list. |
| `accounts` | fast | In a temp data dir: create two accounts with a save each, switch, export and import (the export must round-trip; tampered or garbage exports are refused), back up, recover from a recovery code (garbage and typo'd codes refused), wipe one account (the other's files stay byte-identical, a backup is written), wipe the active one and restore it from its export. |
| `perf` | fast | The late-game `BenchmarkTick` and `BenchmarkGetState` (in `game/tick_perf_test.go`) against budgets (250µs and 1ms, generous for CI runners; actuals are reported), and a long run that samples the heap after GC and the size of every collection in `GameState`, failing on unbounded growth. |
| `ui` | fast: the default and one light theme at 80x24; full: every theme at 80x24 and 100x30 | Runs `TestSmokeUISweep` (every theme at 180x56: splash pages, a new game, every overlay and read-only command) and `TestSmokeUISmallTerminals` (the dashboard and every overlay at small sizes, then live resizes) from `ui/smoke_sweep_test.go` (build tag `smoke`). A panic, a frozen event loop or a blank screen fails it. |
| `styles` | full | The bot under six styles, 2 seeds each: `greedy`, `idle` (a player who checks in every 3 game-hours, `-check-in` to change it: each visit builds storage, plays rounds until nothing more is worth doing, then leaves a build plan for the hours until the next visit, with wonder overflow on; it leaves its workers to the game's worker shares on auto, so the game recruits and staffs between visits, and never recruits or assigns by hand), `harbinger` (buys Appease and Brace), `succumber` (Succumbs every catastrophe), `cosmic` (Invites the Cosmic Era's harbinger, then Succumbs the Last Passage), `army` (keeps a modest garrison, as `-army on`; report-only: graded, never fails). Reports pacing (with the first run to the Modern Age, graded but never enforced) and outcomes per style. |
| `idle` | full | Check-in players at 1h, 3h and 8h (3 seeds each) who leave a build plan at every visit, graded on the median time to the first prestige against `IdleTargets` in `smoke/idle_targets.go` (7, 8 and 11 days); under `-pacing enforce` a miss fails it. Reports the time in each age per interval, and each interval's median as a ratio to the greedy bot on the same seeds (reported, not graded). `-no-plan`, `-no-overflow` and `-no-shares` switch off the plan, wonder overflow and worker shares (with `-no-shares` the bot recruits and assigns by hand at each check-in, with auto-recruit off), to measure what each is worth. |
| `prestige` | full | Plays for two prestige cycles (reporting how far it got if the budget runs out first), then drives the same mechanics through the engine's test hooks: a Succumb for a legacy bonus and ruins, prestiges from the Modern Age checked against the formula, upgrades measured against a twin engine that bought none, a succumbed Last Passage for the Cosmic Legacy, and the prestiges it must survive. |

#### Tiers and CI

- **Fast tier** (`make smoke`, `-tier fast`): the scenarios above marked fast. It runs on every pull request as the `smoke (fast tier)` job in `.github/workflows/go.yml` (skipped for doc-only changes, like the rest of that workflow), writes a summary to the job page and uploads `smoke-report/` as an artifact.
- **Full tier** (`make smoke-full`, `-tier full`): every scenario, deeper. It runs nightly in `.github/workflows/smoke.yml` (and on demand from the Actions tab, suite `nightly`, with a `scenarios` input to run only some, e.g. `idle`) with the same summary and artifact. It enforces the idle targets too (the `idle` scenario). The full tier is about 4.5 hours of work, past one job's 180 minutes, so the nightly runs it as four shards side by side (`nightly (...)` jobs: progression seeds 1-4 and seeds 5-8 in report mode, styles and idle, and the rest; each picks its share with `-only` or `-skip`, and a shard left with nothing writes an empty report). Then the `smoke (full tier)` job downloads their reports and merges them with `-merge` under `-pacing enforce`: the progression runs are pooled and graded on the median across all eight seeds, the static check runs again, and every other scenario's result carries over (its tables stay in its shard's artifact). The summary and the email come from the merged report. It fails if a shard breaks, a shard wrote no report, or the merged report fails.
- **Deep tier** (`make smoke-deep`, `-tier deep`): the static check and the progression scenario only, played to a Quantum Age prestige, so pacing is enforced on every first-cycle age from the Primitive to the Galactic (the nightly stops at the Digital Age). It runs weekly (Sunday 03:41 UTC) in the same workflow, and on demand with suite `weekly`: one job per seed (`deep (seed N)`, in report mode, about 13 minutes each on a runner; one job per seed leaves room under the 6-hour job limit for more seeds or cycles), then `smoke (deep tier)` downloads their reports, pools the runs with `-merge` and grades the median under `-pacing enforce`. It fails if a seed breaks, a seed job wrote no report, or a first-cycle age's median, or the first run's, leaves its band. It emails like the nightly.
- Both fail on panics, soft-locks, invariant violations, save/load divergence, fuzz failures, account failures, docs mismatches and blown perf budgets, and both run with `-pacing enforce` (the nightly in its merge job), so a first-cycle age of the progression scenario whose median across seeds leaves its pacing band fails them too, and so does a median first run to the Modern Age outside the first-run band once the runs get that far (see Pacing targets below).
- **Cross-machine determinism** (`.github/workflows/determinism.yml`, every pull request and push to master that runs the Go workflow): the same three progression seeds, played through a prestige with a state digest every 2000 ticks (`-digest-every`; every run's final `state_digest`, and the map model's fingerprint of that state as `map_digest`, are in `report.json` regardless), on Linux amd64, Linux arm64 and macOS arm64. Each runner writes a fingerprint (`scripts/determinism.sh fingerprint`: the report without wall-clock times, plus a hash of every exported config file) and runs the `detmath` golden-bits test; the `determinism (compare)` job fails unless all fingerprints match, and prints each seed's first differing digest, which brackets the ticks where the runs parted. On an Apple Silicon Mac, `scripts/determinism.sh local` runs the same check natively and as amd64 under Rosetta (about 3 minutes). To find the first differing tick and field, rerun both machines with a smaller `-digest-every` around that bracket, then compare `GameEngine.StateJSON` at that tick. See Float rules for simulation code.
- **Known bugs** can be reported as warnings instead of failures so the job stays useful while they wait for a fix (`-strict` fails on them too); today there are none. When a known bug is fixed, delete its special case in the scenario.

#### Reading the report

Everything lands in `smoke-report/` (`-out` to change it):

- `summary.md`: the verdict, one row per scenario, the pacing table and the failures. It is what the CI job page and the email show.
- `report.md`: the full report, one section per scenario with every failure (message, repro command, and a state dump or stack under "detail"), the warnings, and the scenario's tables.
- `report.json`: the same for scripts.
- `progression.md`, `style-<name>.md`, `prestige-played.md`: the full bot report for each set of runs: outcome per seed, pacing, events and catastrophes, harbinger threads, bot actions, the Army table (when a run endured a catastrophe or blunted a raid: buildings lost to Endure, mean stock kept, mean garrison share, buildings saved, raids blunted, workers saved), and each anomaly with a dump that names what was blocking the next age.

Before calling a soft-lock a bug, check whether a sensible player would be stuck too or whether the bot is just not clever enough (`-trace` helps). Improve the bot for the second case.

#### Pacing targets

`smoke/targets.go` holds the pacing contract: the time a player should spend in each age at 1x game time (15 minutes in the Primitive Age up to 31h 12m in the Atomic and Modern Ages, about a week to the Modern Age and the first prestige, then 36 to 62 hours per age). An age passes when it takes 0.5x to 2x its target; the report marks each age ✓, slow or fast. The per-age timeout is derived from the same table (4x the target, at least an hour; 124h 48m for the Atomic Age): in report mode an age past it is flagged and play goes on, so a slow age reads as slow instead of killing the run, and only a true no-progress soft-lock fails.

The first run as a whole has a band too: the progression scenario's median time to the Modern Age must fall within `FirstRunLow` to `FirstRunHigh` (4.8 to 6.2 days). The targets sum to about 7 days; the greedy bot, a near-perfect player, lands near 5.3, because build and research times are capped copies of hand-typed values and grow less than the targets. A seed that never gets there counts as slower than any that did. The pacing table also has a **longest quiet** column: per age, the median of each seed's longest stretch with no new building type built and no tech finished (reported only).

The game derives its economy from the same table (payback times, build and research caps; see `design-and-architecture/economy.md`, Law 2), so the targets live in the game: to change them, edit `baseAgeTargets` in `config/pacing.go` (`config.AgeTargets` is that table times `config.PacingStretch`, 2.6, from the Bronze Age on, and every clock counted in ticks stretches by the same factor through `config.StretchTicks`), then make `PacingTargets` in `smoke/targets.go` match `config.AgeTargets`. `TestPacingTargetsMatchConfig` fails while the two differ, and `TestTargetsCoverEveryAge` checks every age but the last has one. `PacingLow`, `PacingHigh`, `TimeoutFactor`, `TimeoutFloor`, `FirstRunLow` and `FirstRunHigh` sit beside `PacingTargets`. Changing a target changes the game, not just the grade: every producer's rate follows it.

**Pacing is enforced in CI.** Every tier runs with `-pacing enforce` (the nightly's and the deep tier's merge jobs, across their seeds): in the progression scenario a first-cycle age whose median across the seeds is outside the band fails the scenario, as does a median first run to the Modern Age outside the first-run band, and any age past its timeout fails its run, so a PR that makes the game meaningfully slower or faster fails its `smoke (fast tier)` check (the fast tier grades the Primitive and Stone Ages), the nightly grades every age to the prestige, and the weekly every age to the Galactic. Later cycles run with prestige upgrades and are graded but never fail (the targets describe a first run), an age left by prestige is not a completed age and isn't graded, and the other scenarios always run in report mode (the idle style checks in every three hours; saveload and perf cut their runs short). If a balance change moves an age out of the band on purpose, change the target in the same PR.

#### Email

The PR fast tier, the nightly and the weekly deep tier can each email the summary (with `report.md` attached when it is under 300 KB) after every run, pass or fail. It is optional: with no secrets the email steps skip, and they always skip on pull requests from forks, which get no secrets. To turn it on, add three repository secrets (Settings → Secrets and variables → Actions):

- `MAIL_USERNAME`: the Gmail address that sends the mail.
- `MAIL_PASSWORD`: a Gmail [app password](https://support.google.com/accounts/answer/185833) for it, not the account password.
- `MAIL_TO`: where the mail goes (comma-separate several).

To keep the nightly emails but stop the per-PR ones, add a repository variable `SMOKE_EMAIL_PR` set to `false`.

**Common test patterns:**
- Tests create isolated managers, with no shared state between tests
- Resource tests must respect `BaseStorage` caps: use `AddStorage()` before `Add()` for large amounts
- Milestone tests use `NewProgressManager().GetAgeOrder()` for the full age map
- Engine tests access internals via `ge.mu.Lock()` for setup, then public API for assertions
- Save/load tests defer `os.Remove(...)` for cleanup and verify full round-trip

---

## Project Structure

```
config/         Data definitions: ages, buildings, techs, resources, milestones,
                events, trade, diplomacy, prestige. Pure data, no logic.
game/           Engine, managers, tick loop. No UI imports.
detmath/        Log, Exp, Pow with the same bits on every architecture (see Float rules).
mapmodel/       The one map model: a GameState laid out for every map style (placement,
                civs, flows, recap, glyph tiers). Pure and deterministic.
ui/             tview TUI. Reads GameState snapshots. Never writes to engine.
ui/mapstyle/    Map styles drawn from mapmodel (roguelike, skyline) and their registry.
scripts/        release.sh, commit.sh, determinism.sh (dev tooling)
main.go         Entry point; wires engine + UI.
```

---

## Key Patterns

- **Config-Driven Content**: All game content is data in `config/`. Add buildings, techs, ages, events there, not in logic files.
- **Manager Pattern**: Each system has its own manager with a clean API. No cross-manager direct calls.
- **GameState Snapshot**: `engine.GetState()` returns a read-only snapshot. The UI refreshes from snapshots every 500ms and never touches engine internals.
- **Event Bus**: Systems communicate via `game.EventBus` (pub/sub, synchronous under write lock). Subscribe in `ui/dashboard.go` for toasts, in managers for cross-system reactions.
- **No Global State**: Pass dependencies explicitly. No singletons.
- **Two logs**: the main window's log and the Logs panel show the same lines (everything but `debug`, which only dumps carry), so the main window always says what a command did. The main log leaves out the tick number and prints each line in its category's color; the Logs panel keeps the tick on every line, tags each category and marks a routine confirmation (`game.LogRoutine`: a build started, workers assigned, a plan item added) with a `·`. `ui/log_routing.go` decides which log shows an entry, and `TestSuccessRepliesAreLogged` checks that every command that succeeds writes a line the main log shows.

### Critical: Event Bus Deadlock

Bus handlers run synchronously under the engine's write lock. **Never call `engine.GetState()` or any lock-acquiring method inside a bus subscriber.** Use `config.*ByKey()` functions (pure data, no locks) for any lookups inside handlers.

### Float rules for simulation code

A seed must play the same run on every machine: the same smoke numbers locally and in CI, and a save that carries on identically after moving from a Mac to a Linux box. Floating point is deterministic only if every machine does the same operations with the same rounding, and two things break that. Both rules apply to `game/`, `config/`, `boon/`, `flavor/`, `detmath/`, `mapmodel/` and `smoke/` (tests excluded), and the map styles under `ui/mapstyle/` keep the second one (`mapmodel.Sin`, `Cos`, `Log2` and `Noise` are there for them):

1. **Round every product that feeds an addition or subtraction: `float64(a*b) + c`, not `a*b + c`.** Go lets the compiler fuse `a*b + c` into one fused multiply-add (FMA) instruction, which skips the rounding of `a*b`. arm64 (Apple Silicon, Graviton) always fuses and amd64 at `GOAMD64=v3` does; default amd64 never does, so the last bit differs, and a few thousand ticks later so does the run. Fusion can reach across statements and through inlined calls (`t := a*b; x += t` fuses too, and so does `x += f()` when `f` returns a product), so a plain assignment is not enough: only an explicit conversion rounds. A conversion changes nothing on amd64 v1, so these fixes never move CI numbers.
2. **No transcendental functions from package `math`.** `math.Log`, `Exp`, `Pow` and the rest are assembly on some architectures, Go compiled with FMA on others, and `math.Exp` on amd64 even picks a path by CPU features. Use `detmath.Log`, `Log10`, `Exp` and `Pow`: the same algorithms with every product rounded. `detmath.Log` and `Log10` equal `math.Log` and `math.Log10` on amd64 bit for bit, and `detmath.Pow` with an integer exponent equals `math.Pow` everywhere. `math.Sqrt`, `Floor`, `Ceil`, `Round`, `Trunc`, `Abs`, `Mod`, `Min` and `Max` are exact and fine.

Map iteration order is the third trap (Go randomizes it per run); float sums and anything that rolls dice walk sorted keys (`sortedKeys`, `BuildingManager.eachBuilt`).

Tests in `detmath/rules_test.go` enforce both rules: `TestNoFusedMultiplyAdd` compiles these packages for arm64 and amd64 v3 and lists every FMA instruction with its file, line and source, and `TestNoStdlibTranscendentals` lists every banned `math` call. The cross-machine check (Smoke testing, below) catches whatever slips past them.

---

## Developer Console

A hidden dev console is available for playtesting without grinding through all 22 ages.

**Unlock:** Press `Ctrl+K` anywhere in the dashboard → type the developer passphrase → press Enter.

**Access:** once unlocked, type the `/` commands below at the normal `>` prompt. Their replies go to the game log.

| Command | Effect |
|---|---|
| `/ages` | List all 22 age keys |
| `/age <key>` | Jump to any age instantly |
| `/fill` | Fill all resources to storage cap |
| `/give <resource> <amount>` | Add a specific resource |
| `/techs` | Unlock all techs up to current age |
| `/build <key>` | Instantly place any building |
| `/prestige <n>` | Set prestige level 0-9 |
| `/speed <n>` | Set tick speed multiplier |
| `/god` | Toggle godmode: zero costs, instant builds |
| `/mastery <age\|all> <0-10>` | Set Era Mastery for one age or every age |
| `/record <key>` | Set the record (the deepest age ever entered), which catch-up reads |

The passphrase is stored as a SHA256 hash in `game/devmode.go`, never plain text. Dev mode never persists to disk; it resets on every restart.

Any dev command that succeeds (all of the above except the read-only `/ages`, plus `/catastrophe`, `/harbinger` and `/lastpassage`) marks the run **dev-touched**: from then on it records nothing to the account (no achievements, lifetime stats or theme unlocks), and the log says so once. The mark is saved with the run (`dev_touched` in the save), kept through prestige and Succumb, and cleared only by a new game; a new game or a load with `/god` still on starts marked. Unlocking the console without using it marks nothing, so local builds with dev mode on still earn records.

---

## Adding Content

**New building**: add a `BuildingDef` to `config/buildings.go` with `BaseCost`, `CostScale`, `BuildTicks`, `Category`, and `Effects`. Unlock it in the matching age's `UnlockBuildings` in `config/ages.go`. Cost formula: `floor(BaseCost × CostScale^count)`. Typical `CostScale`: 1.25-1.6.

**New tech**: add a `TechDef` to `config/techs.go` with `Age` (gate), `Cost` (knowledge), `Prerequisites`, and `Effects`. Effect types: `"production"` (flat per-tick) or `"bonus"` (multiplier, e.g. `"gold_rate"`, `"tick_speed"`).

**New age**: add an `AgeDef` to `config/ages.go` with `ResourceReqs`, `BuildingReqs`, `UnlockResources`, `UnlockBuildings`, `UnlockVillagers`. Add a matching age milestone in `config/milestones.go` with `Category: "ages"` and `MinAge` set. On age transition, each resource is capped to ~`carryoverStarterBuildings` (8) of the cheapest new-age building that uses it (via `config.AgeEntryCosts`); resources no new-age building uses as a build cost keep `carryoverResidualPct` (10%, e.g. food, which avoids a starvation spiral at the transition); amounts already below the cap are preserved; faith is exempt (cumulative). See `advanceAge` in `game/engine.go`.

Write the **raw** `ResourceReqs`/`BuildingReqs` you want as the *baseline*; they are not the final gate. `Ages()` runs every def through `normalizeAgeRequirements` (in `config/ages.go`) before returning, so all consumers (`AgeByKey`, `AgeOrder`, and `CheckAdvancement`) see the scaled values. The scaling (EPIC economy-rebalance sub-ticket 3, deliberately moderate because the cost-curve and carryover fixes already tightened pacing; tunable in that function): resource reqs are multiplied by a per-band factor (**2.0x** for stone/bronze/iron, **1.75x** for classical/medieval/renaissance, **1.5x** for colonial/industrial/victorian, **1.25x** for electric → transcendent), then rounded to 2 significant figures. Building reqs below **5** are raised to 5 for the early/mid ages (stone_age through information_age); digital_age onward and primitive_age are untouched.

**New milestone**: add a `MilestoneDef` to `config/milestones.go`. Set `Hidden: true` if it should only appear when progress > 50%. To include it in a chain, add its key to the chain's `MilestoneKeys` in `MilestoneChains()`. Chain completion auto-grants a title + speed boost. `TestMilestonesAreFeasible` (`smoke/static_milestones_test.go`) proves every milestone can be completed from config, so a count the game can't reach fails CI with the math; see the Milestone Covenant in `design-and-architecture/economy.md`. Use `MinBuildingSum` for consecutive tiers of one lineage so upgrading doesn't lose progress, and the run counters (`MinTotalBuilt`, `MinSoldiersTrained`, `MinWonders`, `MinKnowledgeWorkers`) rather than special-casing a key in the engine.

**New random event**: add an `EventDef` to `config/events.go` with `Sentiment` (good/bad/mixed), `Weight`, `Cooldown`, `Duration` (0 = instant), `MinAge`, and `Effects`. Streak logic caps bad events at 2 consecutive and forces one after 3 good ones.

**New expedition**: add a def to `getExpeditions()` in `game/military.go` with `SoldiersNeeded`, `Duration`, `DifficultyBase`, `Rewards`, `MinAge`. Success: `random() > (DifficultyBase - military_bonus × 0.3)`.

**New trade route**: add a `TradeRouteDef` to `config/trade.go` with `Export`/`Import` maps, `TicksPerRun`, `RequiredBuilding`, `MinAge`. Routes auto-cycle, importing `amount × (1.0 + diplomacy_bonus)`.

**New command**: register it in the command registry, `registry()` in `ui/commands.go`: its name and aliases, subcommands, argument slots (an `ArgKind` says what each takes, so completion offers the right game things), help rows, a Help section, and `Dangerous: true` if it can't be undone (Enter then never runs it from a completion). Then handle it in `HandleCommand` in `ui/input.go` and document it in `site/docs/commands.md`. Completion, the Help panel and the smoke fuzz corpus follow from the registry. `TestRegistryMatchesDispatcher` (`go test ./ui`) fails until the registry and the handler agree both ways, and the smoke `docsync` scenario until `commands.md` does.

**New villager type**: add a `VillagerTypeDef` to `game/villagers.go` with `FoodCost` and `GatherRate`. Unlock it in the matching age in `config/ages.go`.

---

## How the Math Works

### Tick Loop

Base interval: **2 seconds**. Each tick in order: build queue → research → events → expeditions → trade → diplomacy → production → resources → milestones → age check → tick speed recalc.

```
tick_interval = 2000ms / ((1.0 + tick_speed_bonus) × speed_multiplier)
minimum: 200ms
```

`tick_speed_bonus`: research + milestones + prestige (+1%/level) + active chain boosts.
`speed_multiplier`: always 1.0 for players. Game speed is fixed so the calendar paces the game: there is no player speed setting and wonders raise no cap. Only the dev console's `/speed` changes it, for the session; a load clamps it back to 1.0 (`clampPlayerSpeed`), and prestige, Succumb and a wipe reset it.

### Resource Rates (per tick, in order)

1. Base: building production + villager gathering + research effects + event effects
2. `production_all` multiplier: `× (1 + Σ production_all)`, applied via the resolver. The sum may be **negative** (e.g. the catastrophe Reconstruction Effort −10% debuff), and negative pools apply as debuffs; there is no `>0` gate. Result is floored at 10% of base so production can't be driven below it.
3. Per-resource multiplier (e.g. `gold_rate` bonus)
4. Gather rate bonus: additive on villager rates
5. Diplomacy trade bonuses: multiplicative on positive rates
6. Food drain: `sum(villager_count × food_cost_per_type)` subtracted

### Building Costs

```
cost = floor(base_cost × cost_scale ^ current_count) × (1 + Σ build_cost)
```
Example with a 30-wood base cost and scale 1.3: 1st=30, 2nd=39, 3rd=50, 4th=66...

`build_cost` modifiers are consumed by `GetCost()`: build-cost-reducing milestone rewards and the Civil Engineering tech (each −3% to −5%) sum into `Σ build_cost` and multiply the scaled cost, floored at 10% of base. `GetCost()` is the consumer the resolver migration wired in (see `design-and-architecture/multiplier-system.md` bug #2). Both the displayed and charged cost reflect the reduction.

### Food Economy

Workers: 0.10/tick · Soldiers: 0.25/tick · Astronauts: 0.40/tick. Keep ~⅓ of workforce on food. The worker shares routine recruits only while the net food rate stays above a small margin: one worker's food, or `recruitFoodMarginShare` (5%) of the food production if that is more (`game/shares.go`).

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
# in steps of 30 ticks (game.OfflineStepTicks), each:
resource_gain = rate × 30 × 0.5     # 50% efficiency, capped at storage; overflow to the wonder bank
construction and research advance 30 ticks; the build plan starts what the step paid for
the worker shares routine staffs idle workers and recruits into empty slots (shares.go)
```

With an empty plan, nothing under construction and nothing for the shares routine to do (no idle workers, no empty slot it may recruit into) this is the old lump sum. See `site/docs/plan.md` and `site/docs/workers-and-domains.md` (Worker Shares).

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
- Keep changes minimal: don't refactor code you didn't need to touch

---

## Project Tracking

All active work is tracked on the [AgeForge Trello board](https://trello.com/b/tf31C2cz/ageforge).

| Lane | What lives there |
|---|---|
| **Bugs** | One ticket per bug, single fix per card |
| **Features** | New systems, commands, and content (`feat`) |
| **Balance** | Cost, rate, and number tuning (`balance`) |
| **Refactor** | Cleanup with no behavior change (`refactor`) |
| **Doing** | Actively in progress |
| **Done** | Shipped: full history of completed work |
| **Later Enhancements** | Shelved ideas for future consideration |

When picking up work: move the card to **Doing**. When done: move it to **Done** and close the PR.

Commit types map 1:1 to board lanes: if your change is a `fix`, it came from the Bugs lane. If it's `balance`, it came from Balance. This keeps the board and git history in sync.

---

## Claude agent (GitHub Actions)

`.github/workflows/claude-agent.yml` runs a Claude Code agent on a GitHub runner. It works a task end to end (code, `go test -race`, `make smoke`), pushes a branch and opens a PR. It never merges.

- **Start one:** label an issue `agent` (the issue author and the labeller both need write access), or run the workflow manually with a brief (Actions → Claude agent → Run workflow). Manual runs can pick the model and effort.
- **Defaults:** `claude-opus-5-5` at `high` effort (pick `xhigh` on a manual run for balance work or big features), 400 turns max, 5-hour job timeout. One agent runs at a time; further runs queue.
- **Setup (owner only):**
  1. Run `claude setup-token` locally.
  2. Add the result as the repository secret `CLAUDE_CODE_OAUTH_TOKEN`. That bills to the Claude subscription. Do not also add `ANTHROPIC_API_KEY`: if both exist the API key wins and usage is billed per token.
  3. Install the Claude GitHub App on the repo. The workflow is skipped until the secret exists.

## Reporting Bugs

Found a bug? [Open a GitHub Issue](https://github.com/espresso20/ageforge/issues) with what you were doing, what you expected, what happened, and your OS + terminal. It will be triaged into a Trello bug ticket.

## Questions

Join the [Discord](https://discord.gg/EPvyd5vjpj) or check the [project board](https://trello.com/b/tf31C2cz/ageforge) to see what's being worked on.
