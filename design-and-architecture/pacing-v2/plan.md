# Pacing v2: implementation plan

> **Status (2026-10-02).** The plan below is kept as written on 2026-09-30; this note records what has happened since.
>
> - **Done:** PR 1 (#166), PR 2 trimmed (#170), PR 3 (#172), PR 4 (#165), PR 5 (#177).
> - **Remaining:** PRs 6 and 7.
> - **Owner decisions:** a one-week first run, Era Mastery (Option C of the prestige audit), and the player `speed` setting retired.
> - **PR 2 trim:** the research queue and the standing catastrophe order were cut. Offline progress is unchanged at 24 hours (not the 72 below).
> - **Open questions:** every recommended default is accepted, except the 100% offline rate (question 2), which is deferred.

Plan only: nothing here is built yet. Written 2026-09-30 against origin/master 47a3822.

Sources:
- the prestige audit (scratchpad prestige-audit.md)
- your decisions of 2026-09-30
- the player and designer research
- six read-only surveys of the code
- new measurements made for this plan (appendix)

Line numbers point at origin/master 47a3822, except game/fate.go, which is on the catastrophe branch. Expect them to shift once the prerequisite PRs merge.

## The short version

What you get:
- A first run of about a week. The active bot takes 5.3 days; 3-hour check-ins about 7.
- Ages you have completed run 2 to 4.2 times faster. New ground stays at 1x.
- Prestige from the Medieval Age, points paid by depth, and a legacy kit that keeps your plan, research order, worker shares and contacts.
- Being away costs nothing you didn't choose: a research queue, a standing catastrophe order, overflow that pays the plan, and 72 hours of offline progress.

Seven PRs: one XL, three L, three M.

| # | PR | size |
|---|---|---|
| 1 | One-week curve and re-timed clocks | L |
| 2 | Away-proofing | L |
| 3 | Mid-age unlocks | M |
| 4 | Worker shares | M |
| 5 | Era Mastery engine | XL |
| 6 | Depth points, Medieval prestige, legacy kit | L |
| 7 | Legacy bonuses, dead perks, final tune | M |

Order:
- Main lane: 1, then 2, 5, 6, 7.
- Side lane: 3 and 4 can run once 1 has merged. 4 must land before 6.

New numbers from this plan:
- Veteran strength (k = 4.16) reaches Modern in 1.25 and 1.29 days on two seeds. So k-scaling holds up to k = 4, not just k = 2.
- The same veteran checking in every 8 hours takes 5.5 days (4.4 times the active time). That's why the legacy kit has to chain ages between visits.

Docs follow the wiki sync rule in every PR. The landing page's hero stats change only in PR 3 (the tech count).

## Words used here

- **k**: how many times faster an age runs. Production is multiplied by k after the ×3 cap. Build and research times are divided by k. Storage is multiplied by k. Events and raids stay on the real clock.
- **Mastery (m)**: how many runs have completed an age, capped at 10. k = 1 + √m, so m = 0 to 10 gives 1, 2, 2.41, 2.73, 3, 3.24, 3.45, 3.65, 3.83, 4 and 4.16.
- **Frontier**: an age you have never completed. It has m = 0, so k = 1.
- **Record**: the deepest age you have ever entered.
- **Catch-up**: ages 6 or more behind your record run at k = 4, or at their own k if that is higher (4.16 at m = 10).
- **Depth points**: every age completed in a run pays 3^epoch. That's 1 per Stone-era age, 3 Iron, 9 Steel, 27 Electric, 81 Digital, 243 Neon and 729 Cosmic, with no divisor.
- **AgeStretch**: 1 for the Primitive and Stone Ages, 2.6 from the Bronze Age on. PR 1 multiplies tick-counted clocks by it.
- **Legacy kit**: four shop items that carry automation across a prestige.
- **Taste**: an early prestige, from the Medieval Age up to the Atomic Age.

## Assumed merged before PR 1

1. **Retire player speed** (fix/espresso/retire-speed).
   - Speed becomes 1x only. The saved `speed_multiplier` is clamped on load and reset by prestige. Wonders no longer raise a speed cap. The dev console keeps /speed.
   - Without it no pacing target means anything: speed reached 7x by the Atomic Age.
2. **Catastrophe "randomize the when"** (feat/espresso/catastrophe-fated-timing).
   - It adds `expectedAgeTicks(age)` (game/fate.go:136), whose comment names it the one place a mastery factor should divide.
   - A fate's window and strike tick are absolute ticks, saved in `FateSave` when the era is entered.
   - A pending catastrophe still refuses AdvanceAge and prestige, and the plan's advance item waits behind it.
3. **Account data fixes** (fix/espresso/account-data-safety). account.json is signed and carries a version. PR 6 adds one field to its stats.
4. **Milestone feasibility and retune** (fix/espresso/milestone-feasibility). PR 1 re-times the two survivor milestones; whichever PR lands second rebases.

Status at writing: no PR is open for any of the four. Retire-speed and catastrophe timing have work in progress; the other two haven't started.

## PR 1: One-week curve and re-timed clocks (L)

**Goal.** A first run to Modern of about a week, active about 5.3 days, without making each age busier. Anything that counted real minutes keeps the same count per age.

**Changes**
- config/pacing.go:
  - `AgeTargets` becomes today's table × 2.6 from the Bronze Age: Bronze 3.9 h, Iron 6.5 h, Classical 9.1 h, Medieval 11.7 h, Renaissance 15.6 h, Colonial 18.2 h, Industrial 20.8 h, Victorian 23.4 h, Electric 26 h, Atomic and Modern 31.2 h, Information 36.4 h, rising to 62.4 h from Interstellar.
  - New `PacingStretch = 2.6` and `AgeStretch(age)`. A future curve change is then one number.
  - Keep the build and research normalizers as caps; that's what the audit measured. Stretching the hand-typed times is the lever to pull only if active play lands under 5 days (measured: +26% active).
- Multiply these clocks by AgeStretch: the random-event delay, event durations and cooldowns, epoch events, awakenings, war raids, war-ends-after-quiet, opinion drift, worker lending, deal refresh, trade route run length, the Endure debuff, expeditions, auto-expedition, boons and maluses, the festival, the black market and milestone chain boosts. Files and lines are in the clock table below.
- Divide boon instant lumps by AgeStretch. They're tuned to output per tick, which drops 2.6x.
- Survivor milestones: MinTick × 2.6 (about 14 h 26 m and 3 days), with their descriptions updated.
- Harbinger Appease and Brace prices grow 2.6x with the target, because they're priced as 0.25 × flow income × target (harbinger.go:523), but faith and culture storage doesn't grow.
  - Add a static check that every price fits max storage.
  - Where a price doesn't fit, price it on target ÷ AgeStretch.
- Smoke: the targets copy, idle targets, budgets and nightly sharding (below).

**New saved fields.** None. Timers already running in a save keep their old length once: festival cooldown end ticks, and a fate window rolled before the update.

**Tests**
- config: golden AgeTargets and AgeStretch values.
- One test per clock family:
  - events_test
  - diplomacy tests
  - trade routes
  - festival and black market
  - expeditions and auto-expedition
  - boon_tuning_test
  - the Endure debuff
  - chain boosts
  - survivor milestones, plus TestMilestoneDescriptionNumbers
- Appease and Brace prices fit faith and culture storage.

**Smoke**
- smoke/targets.go copy updated; TestPacingTargetsMatchConfig holds.
- New check: active time to Modern is 4.8 to 6.2 days (median of 8 seeds).
- Idle targets become 1 h ≤ 7.5 days, 3 h ≤ 8.5 days and 8 h ≤ 10.5 days.
  - This is required, not optional: TestIdleTargetsAreSane needs every idle target at or above the 167-hour cumulative target to Modern, and today's 84 h and 120 h fail it.
  - Ratios to active play are reported, not enforced, until PR 2.
- Promote the audit's quiet-stretch measure into smoke/run.go and report it per age. PR 3 enforces it.
- Budgets (starting values; set them from the first CI run):
  - full progression 600 h → 1,000 h (needed: about 460 h for the bot, 570 h at target)
  - styles 300 h → 800 h
  - prestige 400 h → 900 h
  - deep stays at 1,000 h (557 h needed)
- PR tier: unchanged. It stops on entering the Bronze Age, and the Primitive and Stone Ages keep their targets.
- Runtime:
  - Scenarios bound by simulated time grow about 2.3 to 2.6x. The nightly goes from 1 h 53 m to about 4.5 h of work, past its 180-minute limit.
  - Split the nightly into four parallel jobs: progression seeds 1–4, seeds 5–8, styles plus idle, and the rest. Then merge progression with `-merge`, as the weekly already does. Expect about 2 h wall time.
  - Weekly deep jobs: 60 → 150 minutes each.
  - Determinism jobs are fine: 3 seeds to Modern, about 3–7 minutes.

**Docs**
- ages.md: the target table and "about three days" (:21-38)
- epochs.md: hours per era (:23)
- buildings.md: build-time caps (:148) and the target table (:228-253)
- technologies.md: the "(~N min/tech)" headings (:109-309)
- how-to-play.md (:54)
- prestige.md (:17)
- harbinger.md: thread lengths (:208-224)
- events.md, trade.md, military.md (expedition times)
- design-and-architecture/economy.md
- CONTRIBUTING.md (:273, :298)
- CHANGELOG
- CLAUDE.md: "about 3 days to first prestige" becomes "about one week to first prestige"
- site/index.html: no change; it makes no time claim.

**Depends on:** retire-speed and catastrophe timing merged. If milestone feasibility lands first, rebase the survivor edit onto it.

**Difficulty: L.** The curve is one table. The work is about 20 small clock edits, a smoke re-baseline, nightly sharding and a dozen wiki pages. Low design risk, wide surface.

## PR 2: Away-proofing (L)

**Goal.** Checking in every 3 to 8 hours, or leaving for a weekend, costs nothing the player didn't choose.

**Changes**
1. **Research queue for everyone (M).**
   - An ordered queue on ResearchManager. A locked helper starts the next tech when the slot frees and knowledge covers it (minus what the plan reserves).
   - Techs whose age or prerequisites aren't met yet wait in place.
   - It runs live and in the offline loop.
   - Commands: `research queue <tech>...`, `research queue`, `research unqueue <tech>`, `research queue clear`.
   - The plan's own research items keep priority.
   - Files: game/research.go, game/engine.go (the tick near :1026/:1057 and the offline loop near :4055), the research panel, ui/commands.go.
2. **Standing catastrophe order (S).**
   - `catastrophe standing endure|ask`, default ask.
   - With endure set, a pending catastrophe resolves as Endure, at whatever brace level is held, once it has waited an hour of game time. That works live and offline, and logs one line.
   - It never Succumbs.
   - It needs a lock-free Endure path, since Endure takes the engine lock itself (catastrophe.go:355).
   - The setting survives prestige, like the overflow setting.
3. **Overflow pays the plan (M).**
   - After the wonder bank, whatever a full store would throw away goes into the bank of the first waiting build item that needs that resource.
   - The item pays from its bank first when it starts. Removing it returns the bank to stock, up to the cap.
   - An item priced above the cap becomes buyable; today it reserves nothing and never starts.
   - Files: game/overflow.go, game/plan.go.
4. **A bigger plan (S).** It holds 60 items (MaxPlanItems, plan.go:63) and funds 72 hours ahead; planFundTicks already follows MaxOfflineTime.
5. **Offline (S).** The cap goes from 24 h to 72 h (engine.go:3995). Efficiency is open question 2.
6. **Storage (S).**
   - StorageHoldHours becomes a per-age rule: 1.5 h for the Primitive and Stone Ages, 4.5 h from the Bronze Age.
   - After PR 1, literal storage already holds about 3.9 h, so storage buildings in about ten ages need up to +15% to pass.
7. **Idle bot (S).**
   - At check-ins it leaves a fuller plan and sets the research queue and the standing order.
   - New `-away` option: the bot closes the game between check-ins and comes back through offline catch-up. Today's check-in bot keeps the game running, so smoke never sees the offline rate.

**New saved fields** (all omitempty)
- `ResearchSave.Queue []string` `json:"queue,omitempty"`
- `GameSave.CatastropheStanding string` `json:"catastrophe_standing,omitempty"` ("" means ask)
- `PlanItem.Banked map[string]float64` `json:"banked,omitempty"`

**Tests**
- Research queue:
  - order is kept; waiting techs hold; researched ones are skipped
  - the plan's knowledge reservations are respected
  - save and load round trip
  - offline catch-up runs the queue
  - a tech finished by Grand Discovery (engine.go:2099) leaves the queue
  - Ancient Memory research doesn't break it
- Standing order: resolves after the hour at the held brace level, works offline, never Succumbs, off by default.
- Plan bank:
  - overflow deposits only after the wonder is served
  - the first matching item gets the deposit, with a fixed tie order
  - starting an item pays from its bank first
  - removing an item returns the bank to stock
  - save and load round trip
- MaxPlanItems 60 and planFundTicks at 72 h.
- An old signed save still verifies; empty new fields write nothing.

**Smoke**
- Idle ratios enforced, as the median over 3 seeds against the same seed's active time: 1 h ≤ 1.2x, 3 h ≤ 1.3x, 8 h ≤ 1.5x.
- Absolute limits: 3 h ≤ 7.5 days, 8 h ≤ 9 days.
- Active play stays at 4.8 to 6.2 days. Storage and automation must not speed it up much; the audit saw no change from tripled storage.
- The offline scenario covers 72 hours. smoke/scenario_offline.go:33-35 and smoke/scenario.go:185 assume 24 h today.
- `-away` runs are reported, not enforced.
- Static: the Storage Covenant row at 4.5 h from the Bronze Age. TestStorageCovenantCatchesBrokenStorage (static_test.go:62-111) needs new thresholds.
- PR tier: unchanged.
- Runtime: about +10 minutes nightly (longer idle runs).

**Docs**
- technologies.md: "only one technology" and "There is no queue" (:3, :28-30)
- knowledge.md (:70-76)
- plan.md: overflow, 60 items, 72 hours, the standing order
- wonders.md: overflow (:25-35)
- catastrophe.md: the standing order
- commands.md: research queue and catastrophe standing
- buildings.md: "holds at least an hour and a half" (:358)
- saving-and-loading.md: offline
- first-ten-minutes.md (:240)
- economy.md, Law 1
- CONTRIBUTING.md (:478-487)
- CHANGELOG

**Depends on:** PR 1.

**Difficulty: L.** Four features, each M or smaller. The plan bank is the risky one, because it meets reservations, saves and offline steps. If review wants smaller pieces, it splits cleanly into 2a (queue and standing order) and 2b (overflow, plan size, offline, storage).

## PR 3: Mid-age unlocks (M)

**Goal.** No age goes more than 12 hours without a new building type or tech for an active player.

**Changes**
- One mid-age tech in each of the Atomic, Information, Cyberpunk and Fusion Ages.
  - Their measured quiet stretches on the one-week curve are 18.4, 26.4, 23.2 and 28.5 hours.
  - Price and time each tech to land near the middle of its age, and have it unlock one or two of that age's existing buildings (moved from unlocking at age entry).
  - Add new content only where no existing building fits.
- Information Age: bring it from 1.4–1.5x its target down to 1.2x or less, through its gate or a PaybackAdjust entry. Its overshoot feeds the 26-hour stretch.
- Check the Space to Quantum Ages on the weekly deep run, and fix any stretch over 12 hours the same way.
- Smoke: quiet-stretch tracking becomes a graded field on each age's record.

**New saved fields.** None. Buildings an old save already has stay built; new copies need the tech.

**Tests**
- config: the new techs are valid (age, cost, research time at the cap), and each gated building's tech is in the same age.
- The static gate check: no age gate needs a building locked behind an unreachable tech.
- A unit test of the quiet-stretch metric on a synthetic run.
- Milestone feasibility still passes (the tech-count milestones get easier).

**Smoke**
- Enforce ≤ 12 hours per age (active, cycle 1, median over seeds) to Modern nightly. Post-Modern ages are enforced through the weekly merge.
- Information ≤ 1.2x its target.
- PR tier: unchanged.
- Runtime: none.

**Docs**
- technologies.md (the new techs), buildings.md (gated buildings), ages.md (the four age sections)
- site/index.html: "73 Technologies" at :9, :17, :125 and in the hero stats at :137-151
- README.md
- the CLAUDE.md overview's tech count
- CHANGELOG

**Depends on:** PR 1. It runs alongside PR 2 (different files).

**Difficulty: M.** Content in four ages plus one smoke metric. The risk is a gate soft-lock, and the static checks catch that.

## PR 4: Worker shares (M)

**Goal.** Set once how your workers split across domains, and the game staffs new buildings and recruits to match, live and offline. This is the fourth kit item, and it doesn't exist yet: workers are assigned by hand, a new run starts with zero workers, and the plan never recruits (plan.go:729-731).

**Changes**
- `workers share <domain> <percent>` and `workers auto-recruit on|off` (off by default).
- A routine each tick and each offline step:
  - fill buildings by domain toward the shares, up to capacity
  - if auto-recruit is on, recruit while food income stays positive and housing allows
- One owner for staffing: copies the plan starts (staffPlanCopy, plan.go:732-766) follow the shares when shares are set.
- Shares are per domain (the 12 in config/workers.go), not per building, because building keys change on upgrade (villagers.go:186).
- Files: game/villagers.go, game/plan.go, game/engine.go (tick and offline loop), ui/input.go, ui/commands.go, the workers panel.

**New saved fields**
- `GameSave.WorkerShares map[string]float64` `json:"worker_shares,omitempty"`
- `GameSave.AutoRecruit bool` `json:"auto_recruit,omitempty"`

**Tests**
- Shares fill in a fixed order.
- Auto-recruit never causes starvation and stops at housing.
- Offline steps apply shares.
- Plan staffing and shares agree.
- Save and load round trip.

**Smoke**
- The idle bot sets shares once per age instead of assigning by hand. Idle ratios must hold or improve.
- PR tier: unchanged.
- Runtime: none.

**Docs:** workers-and-domains.md, villagers.md, commands.md, first-ten-minutes.md, CHANGELOG.

**Depends on:** PR 1. Must land before PR 6. It overlaps the Worker Apprenticeship (open question 4).

**Difficulty: M.** New but contained. The care goes into food safety and into not fighting the plan's staffing.

## PR 5: Era Mastery engine (XL)

**Goal.** Ages you've completed run k times faster, the frontier runs at 1x, and the passive retires into mastery.

**Changes**
- **config/mastery.go (new).**
  - MasteryCap 10, CatchUpGap 6 and CatchUpK 4.
  - The k table uses `1 + math.Sqrt(m)`. Sqrt is exact on every CPU, and the float rules allow it.
  - The age's k is max(k(m), 4) when the age is 6 or more behind the record.
- **game/prestige.go.** PrestigeManager holds the mastery map, the record and this run's furthest age. It already survives prestige and Succumb, and is cleared only by Reset.
  - At prestige, every age below this run's furthest gets +1, capped at 10.
  - Mastery stays fixed during a run, so k never changes mid-run and fate windows stay consistent.
- **game/engine.go**
  - Cache the current age's k; refresh it on advance, load and prestige.
  - **Production.** In recalculateRates (engine.go:1468-1640), multiply every resource's net rate by k at the very end: after the ×3 clamps, flat tech output, event output, trade bonuses and the food drain. The whole economy then runs k times faster. Add a `MasteryRate` breakdown line so the rates panel can say why.
  - **Storage** (engine.go:1654): multiply by k.
    - **Grace rule:** when k drops (entering the frontier or leaving catch-up), stock above the new cap stays until spent; production simply stops adding to it.
    - Today `Add` (resources.go:84-100) and the storage line both clamp stock down, which would destroy up to three quarters of a veteran's stockpile on the first tick of new ground.
    - Graced resources are saved.
  - **Build times:** divide by k (round up, minimum 1 tick) at the two queue sites, startBuildLocked (~:2815) and BuildMultiple (~:2891), through one helper. The "Started building" log and ui/wonder_gallery.go:197 show the real time.
  - **Research times:** divide by k after the existing (1 − speed) step, at both sites (research.go:81-90 and :124-135). ui/overlay_research.go:165 and :257 show the real time.
  - **Float rules:** every new product that feeds a sum is written `float64(a*b)` (CONTRIBUTING.md:357). detmath's TestNoFusedMultiplyAdd enforces it.
- **The passive retires.** GetBonuses stops adding +2% production and +1% tick speed per level (prestige.go:123-127).
- **Catastrophes.** expectedAgeTicks(age) divides by that age's k, so a fate's window and the harbinger's lead fit a mastered era. The "strike at least 0.20 of an age away" floor divides by k too.
- **Existing saves:** one-time mastery seeding (see the migration section).
- **UI**
  - The ages view shows mastery and speed per age, only up to your record (the spoiler rule, game/spoilers.go).
  - The status line says "Known ground ×2.4" or "New ground".
  - A log line on entering an age.
  - The prestige panel shows what the next prestige adds to mastery.
  - Files: the ages view, ui/input.go (prestige panel :1605-1665), ui/overlay_stats.go (:162-202), ui/age_splash.go, ui/wonder_gallery.go, ui/overlay_research.go.
- **Dev console:** `/mastery <age|all> <0-10>` and `/record <age>` go in game/devcmd.go, which runs before devmode.go. Your local devmode.go is never touched.
- **Test hook:** SetMasteryForTest in game/sim.go, next to the other *ForTest hooks.

**New saved fields** (all omitempty)
- `PrestigeSave.Mastery map[string]int` `json:"mastery,omitempty"`
- `PrestigeSave.Furthest string` `json:"furthest,omitempty"` (the record)
- `PrestigeSave.RunFurthest string` `json:"run_furthest,omitempty"`
- `PrestigeSave.MasterySeeded bool` `json:"mastery_seeded,omitempty"`
- `GameSave.OverCapGrace map[string]bool` `json:"over_cap_grace,omitempty"`

**Tests**
- The k table's exact values; catch-up takes the larger k; the frontier is 1.
- Twin engines, m = 1 against m = 0 in the same age: rates exactly ×2, storage ×2, build ticks halved (rounded up), research ticks halved after the speed step.
- Commit at prestige: +1 below this run's furthest age, capped at 10. No commit at Succumb (unless PR 7 changes that); cleared at Reset.
- Seeding: a level-5 save gets mastery 5 from the Primitive to the Atomic Age, the record and run fields set, exactly once, with one log line. A level-0 save is untouched.
- Grace: stock above the cap after a k drop survives, stops growing, can be spent, loses its grace once under the cap, and survives save and load.
- Fate: the window divides by k, and a strike lands inside a mastered era rather than at its last transition.
- A spoiler guard: no mastery is shown past the record.
- The perf fixture and the determinism fingerprint both include mastery.

**Smoke**
- **Veteran preset hook**, through SetMasteryForTest:
  - `-preset veteran`: mastery 10 through the Space Age, record Interstellar
  - `-preset returning`: mastery 2 through Information and 1 on Digital, record Cyberpunk
- **Later-run grading** (full tier):
  - Cycle 1 plays to Modern, graded as today.
  - Cycle 2 plays for as long as cycle 1 took, then prestiges. This needs a new cycle policy in smoke/run.go, as a flag such as `-push-cycles`.
  - Checks: cycle 2 covers cycle 1's ages at least 1.9x faster, and ends at least one age deeper.
  - Drop the third run to the Digital Age. The weekly deep tier keeps covering post-Modern ages.
- **Veteran bands**, 3 seeds, 1 cycle:
  - each age from the Bronze Age within 0.5 to 2.0x of its target ÷ k
  - the Primitive and Stone Ages together ≤ 1 hour
  - Modern reached in 1.1 to 1.5 days
- **PR tier:** the veteran preset to the Iron Age on one seed. It takes seconds.
- smoke/prestige_rules.go: replace the "PassiveBonus = level × 0.02" check (:85) with mastery carry-over checks.
- smoke/scenario_prestige.go: add a mastery twin check.
- Perf: game/tick_perf_test.go gets a full mastery map and k ≠ 1. Budgets stay at 250 µs per tick and 1 ms per GetState.
- Determinism: scripts/determinism.sh adds one veteran-preset seed, replayed on linux-amd64, linux-arm64 and macOS arm64.
- Runtime, net shorter:
  - about +15 minutes (3 seeds of cycle-2 push, plus the veteran bands)
  - about −35 minutes (8 seeds lose the third run)

**Docs**
- prestige.md: the passive section (:170-179) becomes Era Mastery
- ages.md, epochs.md, catastrophe.md, harbinger.md
- buildings.md and technologies.md: times ÷ k on known ground
- how-to-play.md
- README.md (:60)
- site/index.html: the prestige copy (:231-236, :433-440)
- design-and-architecture/multiplier-system.md (:71-75), economy.md
- CONTRIBUTING.md (:425-426)
- CHANGELOG

**Depends on:** PR 1, PR 2 and catastrophe timing. Without PR 2's automation, 8-hour veterans take 4.4 times the active time (measured).

**Difficulty: XL.** It changes the hottest code (rates, storage, build and research times). It adds a save migration, reaches into fate timing and five UI panels, and brings new smoke machinery. Determinism and perf must hold throughout.

## PR 6: Depth points, Medieval prestige and the legacy kit (L)

**Goal.** Points pay by depth, prestige opens at the Medieval Age, the old shop is refunded, and points buy automation that survives the reset.

CONTRIBUTING.md has said prestige opens at the Medieval Age for a while. This PR makes it true.

**Changes**
- **Points (S)**, game/prestige.go:52-73.
  - The sum of 3^epoch over the ages completed in the run.
  - Drop the milestone, tech and building terms and the √(level + 1) divisor.
  - Medieval pays 9, Modern 120, Information 201, a run through Digital 363 and a run through Space 1,092.
  - The Last Passage still keeps 50%, 70% or 85%.
- **Opening (S).** PrestigeMinAge becomes "medieval_age" (prestige.go:76). The prestige panel previews the points and what one more age would add.
- **Shop v2 (M)**, config/prestige.go.
  - The nine old perks stay in the file, marked retired: hidden, no effect, can't be bought. Their costs stay frozen for the refund.
  - New one-tier items; their keys are permanent:
    - `legacy_plan`, Plan Template, 9 points
    - `legacy_research`, Research Memory, 18
    - `legacy_workers`, Worker Shares, 36
    - `legacy_factions`, Old Friends, 54
  - 117 points in all: a Medieval taste buys the template, and a first Modern run buys the rest.
- **Refund on shop version change (M).** See the migration section.
- **Plan Template (L).**
  - Record the plan as the player writes it, each item tagged with the age it was added in. A build's count is Count + Started; deals are skipped, because their offer IDs die with the run.
  - At run start and at each advance, add that age's slice, plus the advance item if the template had one. The plan then chains ages while you're away.
  - The live plan can't serve as the template: by prestige it only holds what's left.
  - A Succumb re-applies the template too.
- **Research Memory (S).** Record the run's completion order, both in finishResearch (engine.go:1186) and on the Grand Discovery path (engine.go:2099). At run start, fill the research queue with it.
- **Worker Shares (S).** Keep PR 4's shares and the auto-recruit setting across the prestige.
- **Old Friends (S).** Remember the factions met. They're met again as soon as your age reaches theirs, with no expedition and no two-age wait, at neutral opinion.
- **Account (S).** RecordPrestige records the age prestiged from, so badges can tell a taste from a full run.
- **Bot.** It buys the kit in order. A new "taste" style prestiges at the Medieval Age.

**New saved fields** (all omitempty)
- `PrestigeSave.ShopVersion int` `json:"shop_version,omitempty"`
- `PrestigeSave.LegacyPlan []PlanTemplateItem` `json:"legacy_plan,omitempty"`
- `PrestigeSave.LegacyResearch []string` `json:"legacy_research,omitempty"`
- `PrestigeSave.LegacyFactions []string` `json:"legacy_factions,omitempty"`
- `PrestigeSave.LegacyShares map[string]float64` `json:"legacy_shares,omitempty"`
- `GameSave.PlanLog []PlanTemplateItem` `json:"plan_log,omitempty"` (this run's plan as written)
- `ResearchSave.Order []string` `json:"order,omitempty"` (this run's completion order)
- In account.json stats: `PrestigesByAge map[string]int` `json:"prestiges_by_age,omitempty"`

**Tests**
- Points: the depth table, no divisor, the Last Passage shares.
- CanPrestige: from the Medieval Age, and not before.
- Refund, on a synthetic signed level-5 save (the worked example):
  - Available becomes 600 under the default rule; tiers go to 0; keys stay.
  - ShopVersion becomes 2, with one log line.
  - Loading twice doesn't refund twice.
  - Retired keys can't be bought.
- Kit:
  - template capture (age tags, counts, deals skipped) and re-apply at run start and each advance, within the 60-item cap
  - the research queue fills in order and holds techs that aren't available yet
  - remembered factions are met at their age
  - shares carry over
- Account: prestiges are recorded by age.

**Smoke**
- **Depth pays**, with no extra runs: points per day at each age come from first runs. Medieval and Modern come from the nightly cycle 1; Cyberpunk and deeper from the weekly deep run, which is a first run to Quantum.
  - Check: a run through Digital beats a Modern run, which beats a Medieval run, in points per day, each by at least 1.25x. Projected: 37, 23 and 11 points per day.
- **Static:** the weights table, and a second Medieval reset right after a Modern run gains under 10% (9 of 120).
- **Kit carry-over** (prestige_rules.go):
  - template items present at run start and after each advance
  - the queue filled in order
  - remembered factions met at their age
  - shares applied
- **scenario_prestige.go:** its hooked half (:144-165) tests the old perks and gets rewritten: old keys inert, the refund from the level-5 fixture, each kit item bought and checked.
- **Veteran idle ratios** (veteran preset, full kit, a canned template in smoke/testdata):
  - reported from PR 5, enforced here
  - 3 h ≤ 1.5x and 8 h ≤ 2.0x of veteran active play
  - measured without the kit or PR 2: 1.97x and 4.4x
- **docsync:** counts the active shop items (4) instead of len(config.PrestigeUpgrades()) (scenario_docsync.go:67, :263). TestPrestigeFormulaMatchesDocs (scenario_test.go:92) gets new worked examples.
- **PR tier:** the veteran check from PR 5 also buys the kit after a scripted prestige, so a broken refund or kit fails fast.
- **Runtime:** about +15 minutes nightly (veteran idle runs).

**Docs**
- prestige.md: rewrite
- commands.md: prestige section (:290-308) says the Medieval Age; add the shop items, plan template and research memory
- plan.md, technologies.md, knowledge.md, workers-and-domains.md
- trade.md: Old Friends
- account.md
- ages.md: "Prestige becomes available" moves from Modern (:298) to Medieval (:165)
- epochs.md (:402-410), catastrophe.md (:147)
- site/index.html: :231-236, :433-440, :543-548
- README.md: :60-64, where "9 upgrades" becomes the kit
- site/docs/README.md (:33)
- CONTRIBUTING.md: the formula (:467-476)
- CLAUDE.md: "about a week to the Modern Age on a first run; an optional early prestige opens at the Medieval Age"

**Depends on:** PR 5, PR 4, PR 2 (the queue).

**Difficulty: L.** The formula and the Medieval switch are small, and the refund is careful but contained. The plan template is the real work: capture, age tags, and re-applying at each advance under the plan cap.

## PR 7: Legacy bonuses, dead perks and the final tune (M)

**Goal.** Nothing permanent is swallowed by the ×3 cap, dead perk paths are gone, and every number is re-measured once with everything in.

**Changes**
- The Cosmic Legacy's +10% moves after the ×3 cap, as ×1.1. Today it does nothing from the Victorian Age on.
- Succumb's research bonus: today its +25% per epoch is subtracted from research time and floors research at one tick after about four epochs. Make it multiplicative (×0.8 per epoch).
- If you agree (open question 5), a Succumb also commits the run's completed ages as mastery.
- `gather_rate` is a dead key: four techs and a milestone grant it, and nothing reads it. Rewire it to production after the cap, or drop it from their texts.
- Re-measure every target and update the smoke tables and the docs.
- Hand the badge PR its calibration: measured cycle lengths, and the definition that a run is a prestige from Modern or deeper.

**New saved fields.** None.

**Tests**
- The Cosmic Legacy works from the Victorian Age on.
- Research never floors at one tick.
- The gather_rate techs do what they say.
- The Succumb commit, if approved.

**Smoke.** Re-baseline only; no new checks. PR tier: unchanged. Runtime: none.

**Docs:** catastrophe.md (legacy), technologies.md (the gather_rate techs), prestige.md (Cosmic Legacy), CHANGELOG.

**Depends on:** PR 6.

**Difficulty: M.** Small code, but it's the last full re-measure.

## Targets the smoke suite enforces

"From" is the PR that starts enforcing a check; before that it is reported. Measured numbers are seed 1 on the one-week curve (audit harness) unless noted.

**First run, active bot** (8 seeds, median)

| check | from | target | measured |
|---|---|---|---|
| each age | PR 1 | 0.5–2.0x of target | 0.65–1.26x |
| time to Modern | PR 1 | 4.8–6.2 days | 5.30 days |
| Information Age | PR 3 | ≤ 1.2x of target | 1.48x |

**Check-in players** (3 seeds, median; ratios are against the same seed's active time)

| check | from | target | measured |
|---|---|---|---|
| 1 h / 3 h / 8 h, days | PR 1 | ≤ 7.5 / 8.5 / 10.5 | 6.59 / 7.38 / 9.08 |
| 1 h / 3 h / 8 h, ratio | PR 2 | ≤ 1.2 / 1.3 / 1.5x | 1.24 / 1.39 / 1.71x |
| 3 h / 8 h, days | PR 2 | ≤ 7.5 / 9 | 7.38 / 9.08 |
| game closed between check-ins | PR 2 | reported | new |
| veteran 3 h / 8 h, ratio | PR 6 | ≤ 1.5 / 2.0x | 1.97 / 4.4x (no kit) |

**Later runs**

| check | from | target | measured or projected |
|---|---|---|---|
| cycle 2 on old ground | PR 5 | ≥ 1.9x faster | 2.4x projected; k = 2 measured 1.94x |
| cycle 2 depth | PR 5 | ≥ 1 age deeper | into Information, projected |
| veteran, each age from Bronze | PR 5 | 0.5–2.0x of target ÷ k | 0.58–1.12x |
| veteran, Primitive + Stone | PR 5 | ≤ 1 hour | 0.68 hours |
| veteran, to Modern | PR 5 | 1.1–1.5 days | 1.25 and 1.29 days |
| PR tier: veteran to Iron | PR 5 | same bands | seconds of wall time |

**Something new every half day**

| check | from | target | measured |
|---|---|---|---|
| longest gap with no new building type or tech, per age | PR 3 | ≤ 12 hours | 28.5 h Fusion, 26.4 h Information, 23.2 h Cyberpunk, 18.4 h Atomic; all others ≤ 10.2 h |

**Depth pays**

| check | from | target | projected |
|---|---|---|---|
| points per day: run through Digital > Modern > Medieval | PR 6 | each ≥ 1.25x | 37 / 23 / 11 |
| second Medieval reset after a Modern run | PR 6 | < 10% of 120 | 7.5% |

**Engine guards**

| check | from | target |
|---|---|---|
| Storage Covenant | PR 2 | ≥ 4.5 h from Bronze, 1.5 h before |
| storage hours under k | PR 5 | equal at k = 1 and k = 4.16 |
| perf | PR 5 | tick ≤ 250 µs, GetState ≤ 1 ms, with mastery |
| determinism | PR 5 | veteran seed identical on 3 machines |

**At a glance: what is enforced after each PR** (checks add up; nothing is dropped)
- **After PR 1:**
  - first-run bands per age, and active 4.8–6.2 days to Modern
  - idle check-ins ≤ 7.5, 8.5 and 10.5 days (1 h, 3 h, 8 h)
  - idle ratios and quiet stretches reported only
- **After PR 2:**
  - idle ratios ≤ 1.2, 1.3 and 1.5x; 3 h ≤ 7.5 days and 8 h ≤ 9 days
  - a 72-hour absence works
  - Storage Covenant at 4.5 h
- **After PR 3:** quiet stretch ≤ 12 h per age; Information ≤ 1.2x its target.
- **After PR 4:** nothing new; the idle ratios must hold.
- **After PR 5:**
  - cycle 2 ≥ 1.9x faster on old ground and at least one age deeper
  - veteran bands, and the veteran check on every PR
  - storage hours unchanged under k
  - perf and determinism with mastery in play
- **After PR 6:**
  - depth pays; a second taste gains under 10%
  - the kit carries over
  - veteran idle ≤ 1.5x (3 h) and 2.0x (8 h)
  - docsync counts the active shop items
- **After PR 7:** everything re-measured; no new checks.

## Save migration and refund

**Rules**
- Add, never rename or remove. Every new field is omitempty, so an old save re-marshals to the same bytes and its signature still checks.
- Migrations run after signature verification. The next save re-signs as usual. A migration that edited the struct before verifying would flag honest saves as tampered.
- The nine old upgrade keys stay in config, and stay in the save's upgrades map at tier 0.
- Two one-time steps, each with its own marker: mastery seeding (PR 5, `mastery_seeded`) and the shop refund (PR 6, `shop_version` 2).
- The curve itself needs no migration, but a first-run save in progress slows down about 2.3x from the Bronze Age on, from its next load. Say so in the CHANGELOG and in a one-time log line.

**Worked example: a level-5 save**

Before (the audit's bot: runs 1–5 prestiged at Modern, cheapest-first buying):
- Level 5, TotalEarned 86 (27 + 19 + 15 + 13 + 12), Available 3.
- Tiers held, 83 points spent in all:
  - Gather Boost 3, Storage Bonus 3, Knowledge Production 3, Military Power 3
  - Starting Food 5, Starting Wood 4, Housing Bonus 3, Expedition Loot 3
  - Temporal Mastery 0
- It's mid-run in the Industrial Age.

At PR 5 (mastery seeding):
- Every age a Modern prestige completes, Primitive to Atomic, gets mastery min(10, level) = 5. That's k = 3.24.
- Record = the deepest of:
  - the current age
  - Modern, since level ≥ 1
  - Interstellar, if the Cosmic Legacy is held
  - the first age of the deepest epoch in the saved catastrophe history
  
  Here that's Modern.
- This run's furthest age = Industrial.
- The passive goes: −10% production (inside the ×3 cap, and dead from the Victorian Age on anyway) and −5% tick speed.
- Right away the Industrial Age runs at ×3.24. Combined with the slower curve's ÷2.6, it plays about 1.25x faster than before the update.
- One log line: "Era Mastery: ages you have completed now run faster. Primitive to Atomic: mastery 5 (3.2x)."
- Its next prestige at Modern takes the Primitive to Atomic Ages to mastery 6 (3.45x).

At PR 6 (shop refund):
- Refund: 83 spent (at the frozen old costs) + 3 unspent = 86 old points.
- Convert (the recommended default, open question 1): each past prestige counts as one Modern run under the new formula, so 5 × 120 = 600. If 86 × 4.44 (120 new points per 27 old) came out higher, you'd take that; here it's 382, so 600 it is. Face value would be 86.
- Available becomes 600, and so does TotalEarned; lifetime points are in new units from here on.
- All nine old tiers go to 0. Keys are kept. ShopVersion becomes 2.
- One log line: "The prestige shop changed. Your old perks were refunded as 600 points (5 prestiges at 120 each)."
- 600 buys the whole kit (117) and leaves 483 for later shop items.
- Its next prestige at Modern pays 120 points. At level 5 under the √ divisor it paid 12.

Other cases:
- **Level 0, mid-run:** no refund and no mastery; only the slower curve.
- **Level 0, sitting in the Modern Age:** prestige pays 120 instead of 27.
- **Level 38, maxed shop** (277 spent): mastery 10 from Primitive to Atomic (4.16x), and a refund of max(38 × 120, 4.44 × points earned) = 4,560 points.

## Every clock that needs re-timing

**Already derived from the targets** (they move with the curve, no edit)
- Producer payback, config/pacing.go:113. PR 1.
- Build-time caps (target ÷ 6, storage ÷ 48), pacing.go:272. PR 1 keeps them as caps.
- Research-time cap (target ÷ 8), pacing.go:328. PR 1 keeps it as a cap.
- Appease and Brace prices, harbinger.go:523. PR 1 checks they fit storage.
- Fate windows and harbinger lead times (expectedAgeTicks). Merged first; PR 5 divides them by k.
- Smoke age timeouts (4 × target). Automatic.

**Hard-coded ticks, multiplied by AgeStretch in PR 1**
- Random event delay, 150–600 ticks: game/events.go:47-48
- Event durations 0–216 and cooldowns 50–300: config/events.go:19-21
- Epoch events, 60–288: config/epochs.go:161-254
- Awakenings, 200–500: config/awakenings.go:47-110
- War raids, every 40 ticks: game/diplomacy.go:701
- War ends after 300 quiet ticks: diplomacy.go:33
- Opinion drift every 25 ticks, plus its 50- and 100-tick steps: diplomacy.go:20, :589, :594
- Worker lending, 200 ticks: diplomacy.go:38, :654
- Deal refresh, 1,800 ticks: game/deals.go:66
- Trade routes, 8–20 ticks per run: config/trade.go:132-296
- Endure debuff, 216 ticks: game/catastrophe.go:48
- Expeditions, 60–160 ticks: game/military.go:82-220
- Auto-expedition, 900-tick base: game/auto_expedition.go:73
- Boons 750–3,000 and maluses 300–1,500: boon/catalog.go:80-245. Instant lumps are divided instead (catalog.go:23).
- Festival buff 150 and cooldown 300: game/engine.go:50-54
- Black market cooldown, 240: engine.go:64
- Chain boosts, 180 and 150: config/milestones.go:69-135
- Survivor milestones, 10,000 and 50,000 ticks, plus their texts: milestones.go:818-841

**Changed in PR 2**
- Offline cap 24 h → 72 h: engine.go:3995
- Plan funding horizon: follows the cap (plan.go:544)
- Plan size 30 → 60: plan.go:63
- Storage Covenant 1.5 h → 4.5 h from Bronze: pacing.go:516

**Divided by k in PR 5 (known ground only)**
- Production, storage, build ticks (2 sites), research ticks (2 sites), fate windows and harbinger lead.

**Left on the real clock on purpose**
- Market pressure decay, 0.98 per tick (game/trade.go:363); morale drift (engine.go:445); starvation (engine.go:782).
- The PR 1 clocks are stretched but never divided by k, so known ground sees fewer events, raids and routes per age. At k = 4 they matter less anyway, since their amounts are fixed.

**Retired**
- Prestige tick speed: the +1%/level passive goes in PR 5, and Temporal Mastery is refunded in PR 6. With player speed already gone, the only real-time speed-ups left are the short milestone chain boosts.

**Smoke copies (PR 1)**
- smoke/targets.go and smoke/idle_targets.go
- the budgets
- GateTrickleHours 48 (smoke/static.go:41): reconsider it at ×2.6

## Risks and mitigations

1. **Determinism across arm64 and amd64.** New float products in hot paths.
   - Wrap every product that feeds a sum in float64(). k comes from math.Sqrt, which is exact.
   - TestNoFusedMultiplyAdd and TestNoStdlibTranscendentals already guard this.
   - scripts/determinism.sh gains a veteran-preset seed, so all three machines replay k-scaled play.
2. **Perf (GetState's 1 ms budget).** Mastery is a 22-entry map and one cached k; each tick adds one multiply per resource.
   - Add mastery to the perf fixture.
   - The UI reads k from the snapshot instead of recomputing per age per frame.
3. **Idle players.** At veteran strength, without the kit or PR 2, 3-hour check-ins took 1.97 times the active time and 8-hour check-ins 4.4 times (5.5 days against 1.25), with 76% of output lost at full stores. The early ages are shorter than one visit, and the plan only looks one age ahead.
   - PR 2 lands before PR 5.
   - The plan template re-adds each age's slice at every advance, so the plan chains ages between visits.
   - Veteran ratios are enforced in PR 6.
4. **Offline.** The check-in bot keeps the game running, so smoke never sees the 50% offline rate.
   - PR 2's `-away` option measures it.
   - Open question 2.
5. **Early ages at high k.** At k = 4.16, the Primitive and Stone Ages run 2.2x and 3.1x their target ÷ k: the bot's 10-second decisions, one-tick build floors and hand gathering bind. These are minutes, not hours.
   - The veteran band starts at the Bronze Age.
   - Primitive plus Stone must stay under an hour.
6. **Storage shrinking at the frontier.** k drops from up to 4.16 to 1, and storage drops with it.
   - The grace rule in PR 5.
   - Smoke's "over the cap" invariant allows graced resources.
7. **Catastrophes on known ground.** The fate chance stays per era. A veteran crossing three mastered eras in about a day meets more dooms per hour than a first-run player, though not more per run. Harbinger warnings shrink to 12–36 minutes in a one-hour mastered age.
   - The standing order (PR 2) covers absent players.
   - Dividing expectedAgeTicks by k makes strikes land inside the era instead of piling up at its last transition.
8. **The achievements plan.** Rungs "counted in runs" assume 2.3-day runs; v2 runs take about a week, then get faster and deeper. A taste every few hours could farm the 10- and 25-prestige rungs, and tastes fill the early-age mastery ladders fast.
   - PR 6 records the prestige age on the account.
   - The badge PR counts only Modern-or-deeper prestiges as runs, and puts mastery ladders at m = 1, 3, 5 and 10 per age (88 badges; all ten steps would give 220).
   - It recalibrates from PR 7's measured cycles.
   - Resource ladders survive as they are: a run's production is unchanged by the stretch and by k.
9. **Save signing.** Migrations run after verification, and are tested against a signed level-5 fixture.
10. **Balance at high k.** The audit measured k-scaling only up to k = 2. This plan measured k = 3 at 1.71 days (3.1x) and k = 4.16 at 1.25 and 1.29 days (4.1 to 4.2x). The audit's research stand-in slightly overstates research speed-ups, so real play may run a few percent slower.
11. **Nightly runtime.** About 2.5 times longer, past the 180-minute limit. Shard it in PR 1.
12. **First-run saves in progress.** They slow down about 2.3x on update. A CHANGELOG line and a one-time log line.
13. **Flow gates.** Food, faith, culture and soldiers keep their hand-set rates, so their requirements are met 2.6 times sooner relative to the longer ages and stop binding. The measured runs already include this. Leave it for playtest.
14. **Your local devmode.go.** New dev commands go in devcmd.go, never devmode.go.
15. **Tastes dodge dooms.** A prestige clears an unstruck fate, so a Medieval taste can escape an Iron-era doom. It costs the run, so it's a choice, not an exploit. Worth one line in the docs.

## Open questions (each with a recommended default)

1. **Refund rate.** Face value (86 old points in the example) or the new rate (600)?
   - Default: the new rate. New prices are in depth points, and face value would leave a level-5 player worse off than a fresh Modern run.
2. **Offline rate.** Keep 50% with the 72-hour cap, or raise it to 100%?
   - Default: 100%. Otherwise a player who closes the game overnight has a much longer week than the check-in bot measures. Events, raids and routes still skip offline.
3. **Catch-up strength.** k = 4 covers the first seven ages in about 10 hours, but getting back to your record still takes 1 to 4 days. Is that the "hours" you meant?
   - Default: keep k = 4 as a config constant and revisit after a playtest. k = 8 would bring a veteran's Modern to about 16 hours.
4. **Worker shares.** Build them now (PR 4), or fold them into the Worker Apprenticeship?
   - Default: now. Without them a new run starts with zero workers and nothing recruits.
5. **Succumb and mastery.** Should a Succumb commit mastery for the ages completed so far, so the rebuild runs at 2x? A first-run Succumb early in the Steel era costs about 27 hours of replay on the new curve; with this, about 13.
   - Default: yes, in PR 7.
6. **Kit prices** 9 / 18 / 36 / 54.
   - Default: as listed. A Medieval taste buys the template; a first Modern run buys the rest.
7. **What counts as a prestige for badges.** Every prestige, or only Modern or deeper?
   - Default: first_prestige for any prestige; the 10 and 25 rungs for Modern or deeper.
8. **Split the ×3 cap** into permanent and timed pools? That revives the dead late rewards but speeds up late ages.
   - Default: not in v2. A follow-up card after the first playtest.

## Worker Apprenticeship inside k

The Worker Apprenticeship (card sAlxKj5k) comes after Pacing v2 and the badges. Veteran workers should be the visible face of mastery, not a second speed-up stacked on top of it: an age's total speed stays k(m), 1 on new ground and up to 4.16 on old ground. Apprentice output goes after the ×3 cap, like k; inside the cap it would die in the Victorian Age, as the passive did. In a mastered age, workers arrive already trained, and k's production factor shrinks by the same amount, so the product still equals k. Storage scales with the combined factor, so the covenant holds. On new ground, apprentices train on timers written as shares of the age's target (divided by k on known ground, never fixed ticks), and the first-run curve is re-measured with training in, so "the frontier runs at 1x" still means the measured week. Smoke keeps grading each age against target ÷ k, and the Apprenticeship PR does the one final retune.

## Follow-ups found while planning (need cards)

- A faction boon's TempWorkers duration is never enforced (game/faction_boon.go:210-215).
- Prestige sets morale to 0.70; Succumb sets it to 0.50.
- Prestige doesn't reset autoExpeditionTicksLeft, ageReady or starvationTicks.
- The prestige key `research_speed` actually boosts knowledge_rate (moot after the refund).
- The plan's 30-item cap isn't documented anywhere (PR 2 documents the new 60).
- Pacing v2 has no Trello card yet. Please add one, plus one per PR if you like, so the PR titles can carry its shortLink.

## Appendix: measurements made for this plan

**Method**
- The audit's AUDIT_* hooks, unchanged, in /Users/echo/code/ageforge-wt/prestige-audit.
- The smoke binary was built into the scratchpad, and every output went to scratchpad/pv2. Nothing was written to the repo.
- `AUDIT_TARGETS=week` is today's targets × 2.6 from the Bronze Age.
- For k: `AUDIT_PRODMULT`, `AUDIT_GATEMULT` and `AUDIT_STORAGEMULT` all set to k. That gives production after the ×3 cap, build ticks ÷ k, research through +(1 − 1/k) research speed, and storage × k.
- The k = 1 control reproduced the audit's per-age hours to its rounding: 0.29, 0.95, 3.39, 6.34, 8.44, 7.83, 13.16 and 11.76.

**Results** (seed 1 unless noted)

| run | days to Modern | note |
|---|---|---|
| k = 1, active | 5.30 | baseline (re-measured; matches the audit) |
| k = 2, active | 2.73 | 1.94x faster (re-measured; matches) |
| k = 3, active | 1.71 | 3.10x |
| k = 4.16, active | 1.25 (seed 2: 1.29) | 4.24x |
| k = 4.16, 3-hour check-ins | 2.46 | 1.97x the veteran active time |
| k = 4.16, 8-hour check-ins | 5.51 | 4.4x; 76% lost at full stores |
| k = 1, 8-hour check-ins | 9.08 | matches the audit |

- Per age at k = 4.16 (seed 1): Primitive 0.13 h (2.2x its target ÷ k), Stone 0.55 h (3.1x), Bronze to Atomic 0.66 to 1.03x.
- Longest quiet stretch at k = 4.16: 3 hours.

**Commands**

```
B=scratchpad/pv2/auditsmoke   # go build -o $B ./cmd/smoke, run in the audit worktree
AUDIT_TARGETS=week AUDIT_OUT=out AUDIT_TAG=k416 \
  AUDIT_PRODMULT=4.162 AUDIT_GATEMULT=4.162 AUDIT_STORAGEMULT=4.162 \
  $B -tier full -scenario progression -seeds 1 -seed-base 1 -cycles 1 \
  -final-age "" -max-sim 400h -parallel 1 -out rep/k416
# check-ins: -scenario styles -style idle -check-in 3h|8h -max-sim 900h (add AUDIT_COUNT=1 for cap losses)
```

Analysis script: scratchpad/pv2/an.py.
