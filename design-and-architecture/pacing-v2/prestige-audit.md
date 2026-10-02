# Prestige loop and storage caps audit

Date: 2026-09-29. Code: origin/master at 54c171d (#153). Worktree: `/Users/echo/code/ageforge-wt/prestige-audit` (throwaway instrumentation, nothing committed).

All times are simulated game time at 1x unless a row says otherwise. Seed 1 runs are deterministic: the same command gives the same numbers. Seed-to-seed spread in the nightly is small (time to Modern 2.19 to 2.28 days over 8 seeds), so single-seed comparisons are fair for relative effects. Treat differences under about 3% as noise: within one multi-run chain, each run draws different events.

## Terms

- **Run**: one playthrough from the Primitive Age until you prestige.
- **Prestige**: resetting the run from the Modern Age or later, for **points**.
- **Shop, perk, tier**: points buy perks in the prestige shop. Each perk has 5 tiers.
- **Passive bonus**: what each prestige level gives without buying anything: +2% all production and +1% tick speed.
- **The x3 cap**: all production bonuses are added together and the total multiplier is clamped at 3.0x. Each resource's own bonus pool (for example knowledge) has a separate 3.0x clamp.
- **Production multiplier (experiment)**: everything made per tick is multiplied by it, after every bonus and after the x3 cap.
- **Tick speed**: how fast the game clock runs. At 2x, everything happens twice as fast in real time, events and raids included.
- **`speed` command**: the player's own game-speed setting. Each built wonder raises its cap by 0.5x.
- **Storage cap**: the most of a resource you can hold. Production past the cap is lost, except what **overflow** banks into the age's wonder.
- **Storage Covenant**: the design rule that max storage holds 1.5 hours of an age's typical production.
- **Gate Covenant**: the design rule that every age requirement fits in max storage.
- **Wonder banking**: paying into the age's wonder a deposit at a time.
- **Active player**: the smoke greedy bot. It decides every 10 seconds of game time.
- **Check-in player**: the smoke idle bot. It looks in every 1, 3 or 8 hours and leaves a build plan.
- **Price unit**: one unit is the median price of a resource in that age. Used to add different resources together.
- **k-scaling**: production x k (after the x3 cap), construction and research times / k, and storage x k, all at once. It makes an age k times faster, like tick speed, while events and raids stay on the real clock.
- **Mastery**: in Option C, how many runs have completed a given age.
- **Depth points**: the proposed points formula, paying more for each age the deeper it is.

## Headline numbers

**Prestige today**
- Run 1 reaches Modern in **2.26 days** and earns **27 points**. Runs 2 to 10 earn 19, 15, 13, 12, 11, 10, 9, 9 and 8 (133 in all). Maxing the 277-point shop takes **about 38 runs** (2.5 months).
- Run 10 is **16% faster** than run 1 with the bot's own buying (1.91 days). Buying tick speed first gets runs 7 to 10 about **27% faster** (1.58 to 1.71 days). A maxed veteran (level 38, every perk at tier 5) reaches Modern in **1.18 days, 1.9x** faster.
- Pushing past Modern pays about 1 point per age: Information 30 points at 2.8 days, Digital 31 at 3.6, Space 40 at 6.1, Transcendent 49 at 13.3. **Prestiging at Modern is always the best points per day.**
- **Gather Boost is dead code.** Knowledge Production stops working in the Colonial Age, the passive's production part in the Victorian, and the four flat perks are worth about five minutes a run.

**Two things bigger than the whole shop**
- **The `speed` command reaches 7x by Modern.** A first run that uses it reaches Modern in **13 hours**, not 54. **Bug:** the setting survives prestige, so run 2 starts at 7x.
- **The x3 production cap is full from the Victorian Age on.** The prestige passive, the Cosmic Legacy and late production rewards all stack under it and do nothing from there.

**Why the percentages barely move pacing**
- Production x1.25, x1.5, x2 and x4 make the run 1.17x, 1.31x, 1.51x and 2.09x faster. **69% of the run scales with production; 31% is construction and research time**, which production can't touch.
- Tick speed is exact (2x speed, 2.00x faster). **Scaling production and construction and research together by k is nearly exact: 1.9 to 2.0x at k = 2** on every curve and seed tried ("k-scaling").

**Caps**
- **Storage caps do not grow with prestige percentages.** The storage perk is a flat +20 per tier and nothing multiplies storage.
- At 1.5x production the Storage Covenant drops from 1.5 hours to about 1 hour in most ages, at 2x to 45 minutes.
- An 8-hour check-in player loses **60%** of construction output at full storage, and **79 to 81%** at 1.5x to 2x production, for **no speedup at all** (6.2 to 6.3 days either way). On the one-week curve, tripling or sextupling storage barely helped either (9.1 days to 8.8 or 8.9): that player is limited by purchase rounds per visit, so what helps is time and automation, not production.

**Gates**
- All three tight rows (1.09x, 1.69x, 1.77x) are wonder parts, which are banked a deposit at a time. The largest payment a player must actually hold needs 9 to 15 of the age's 25 storage buildings, costing about one wonder. No player is at risk; only the static test is.

**Slowing the game down**
- **One week to the first prestige (primary):** today's targets x2.6 from the Bronze Age, a one-table change. Active **5.3 days**, 3-hour check-ins **7.4** (6.2 with tripled storage, which doesn't speed active play at all), 8-hour **9.1**.
- **Two weeks (your sketch, the alternative):** active 7.8 days, 1 to 3 hour check-ins about 9, 8-hour 12.8. It has the longest stretches with nothing new to decide (26 hours in the Atomic Age, 46 in Fusion).
- The idle gap narrows as ages lengthen: 8-hour check-ins go from 2.8x the active time today to 1.7x (one week) and 1.6x (two weeks). The 8-hour player is limited by purchase rounds per visit, not by storage, and prestige reaches them only partly (a 2x economy makes active play 1.94x faster but 8-hour play 1.38x), so the prestige layer must ship with automation that spends between visits.
- **An early reset at the Medieval Age** only works with steep depth points: with today's formula it would pay twice the points per day of a Modern run.
- **Succumb is not yet a taste of prestige:** it resets the whole run, and its bonus (+25% research speed) barely moves pacing.

**Recommendation: Option C, Era Mastery, on the one-week curve.** Each age you have completed runs faster next time (production up and construction and research down by the same factor, after the x3 clamp, with storage scaled to match): 2x on old ground in run 2, rising to 4x. The frontier always runs at 1x, so every run reaches further. Points pay by depth (x3 per epoch) and buy quality-of-life unlocks, starting with a legacy kit that keeps the plan and research queue across resets. Prestige opens at the Medieval Age as an optional early taste.

## 1. Points: what a run earns and what it buys

### The formula

Points = (age index + milestones/10 + techs/15 + buildings built/50), divided by the square root of (prestige level + 1), rounded down. Each fraction is rounded down too. The Modern Age is index 12.

A typical first run to Modern: 12 (age) + 3 (34 milestones) + 3 (45 techs) + 9 (491 buildings) = **27 points**. Most of it is the age index. The other three terms move slowly.

### Runs 1 to 10 at the Modern Age (seed 1, measured)

Two shop policies:
- **Cheapest first**: the smoke bot's policy today. It buys whatever next tier is cheapest.
- **Speed first** (my "sensible" policy): save for Temporal Mastery (tick speed) until it is maxed, then Knowledge Production, then cheapest first. Tick speed is the only perk that measurably speeds a run (see below).

| run | points earned | total earned | cheapest first: days to Modern | speed first: days to Modern | speed first: shop at the start of the run |
|---|---|---|---|---|---|
| 1 | 27 | 27 | 2.26 | 2.26 | nothing |
| 2 | 19 | 46 | 2.24 | 1.97 | tick speed 2 |
| 3 | 15 | 61 | 2.18 | 2.00 | tick speed 3 |
| 4 | 13 | 74 | 2.15 | 1.80 | tick speed 4 |
| 5 | 12 | 86 | 2.09 | 1.81 | tick speed 4 |
| 6 | 11 | 97 | 2.06 | 1.84 | tick speed 4 |
| 7 | 10 | 107 | 2.02 | 1.65 | tick speed 5, knowledge 2 |
| 8 | 9 | 116 | 1.94 | 1.62 | tick speed 5, knowledge 3 |
| 9 | 9 | 125 | 1.86 | 1.58 | tick speed 5, knowledge 4 |
| 10 | 8 | 133 | 1.91 | 1.71 | tick speed 5, knowledge 4 |

(Speed first earned 1 more point in runs 3 and 10; totals 134.)

- By run 10 the bot's own policy is **16% faster** than run 1. Speed first is **about 27% faster** from run 7 on.
- **What the bot buys** with cheapest first: every tier 1 and 2 first. By run 10 it holds Gather Boost 5, Starting Food 5, Starting Wood 5, Storage 4, Expedition Loot 4, Military 3, Housing 3, Knowledge 3 and **tick speed only at tier 1** (bought in run 8). Nearly all of that does nothing for pacing (next section).
- **Runs to max the shop: about 38.** Maxing costs 277 points. With about 27 raw points per run and the square-root divisor, run 38 is the one that crosses 277. That is roughly **2.5 months** of 2-day runs.

### What each perk actually does (probe: twin states, one with the perk, 600 ticks into each age)

| perk (max tier) | Stone | Medieval | Industrial | Electric | Atomic |
|---|---|---|---|---|---|
| Gather Boost 5 (+25% "worker output") | **no effect** | **no effect** | **no effect** | **no effect** | **no effect** |
| Knowledge Production 5 (+25% knowledge) | knowledge x1.25 | x1.14 | none (x3 knowledge cap reached) | none | none |
| Storage Bonus 5 (+100 storage) | +9.5% of storage | +0.008% | ~0 | ~0 | ~0 |
| Housing Bonus 5 (+10 housing) | +3.6% of housing | +0.6% | +0.23% | +0.016% | +0.008% |
| Passive at level 10 (+20% production) | x1.19 | x1.17 | x1.12 | **none (x3 cap)** | **none** |
| Passive at level 30 (+60% production) | x1.57 | x1.52 | x1.36 | **none** | **none** |

- **Gather Boost is dead code.** Its bonus key, `gather_rate`, is applied to `WorkerManager.GetProductionRates()`, which has returned an empty map since worker output was folded into building output (game/villagers.go:211, game/engine.go:1576). Four techs and one milestone that grant `gather_rate` are dead the same way.
- **Knowledge Production** does nothing from the Colonial Age on: the knowledge pool passes its own x3 clamp there (section 2).
- **Military Power and Expedition Loot** don't show in these numbers: the pacing bot leaves the army and expeditions alone, as the pacing targets assume.
- **The starting resources** (+125 food, +125 wood) are 12% of the first age gate (1,000 food) and nothing after.
- **Temporal Mastery (tick speed)** is the only perk that works in every age: +25% at tier 5 for 89 points.

### Pushing past Modern (seed 1, first run, level 0)

| prestige at | points | days | points per day | extra points for the extra days |
|---|---|---|---|---|
| Modern | 27 | 2.26 | 11.9 | - |
| Information | 30 | 2.77 | 10.8 | +3 for +0.5 days |
| Digital | 31 | 3.61 | 8.6 | +4 for +1.3 days |
| Space | 40 | 6.08 | 6.6 | +13 for +3.8 days |
| Transcendent | 49 | 13.34 | 3.7 | +22 for +11.1 days |

- **Pushing further is not worth it today.** Each age adds 1 point. Prestiging at Modern every 2.26 days earns about 155 points in 30 days. Prestiging at Digital every 3.6 days earns about 130.
- Prestige from the Cosmic Era (Interstellar on) can also bring the Last Passage. Enduring it keeps only 50% of the points (70% or 85% braced). Succumbing pays 0 and grants the Cosmic Legacy once.

## 2. Two things that dwarf the whole shop

### The `speed` command reaches 7x by Modern

Each built wonder raises the speed cap by 0.5x, and every advance requires the age's wonder. So a player who types `speed` after each wonder runs at 2.5x in the Iron Age, 4.5x in the Colonial and 7x in the Atomic. The pacing bot never touches `speed`, so none of the pacing numbers include it.

Measured (seed 1, bot keeping speed at the cap):

| age | 1x hours | at the speed cap | speed at the end of the age |
|---|---|---|---|
| Primitive | 0.29 | 0.29 | 1x |
| Stone | 0.95 | 0.63 | 1.5x |
| Bronze | 1.73 | 0.85 | 2.5x |
| Iron | 3.02 | 1.21 | 2.5x |
| Classical | 3.82 | 1.27 | 3x |
| Medieval | 3.26 | 0.93 | 3.5x |
| Renaissance | 5.47 | 1.32 | 4.5x |
| Colonial | 4.98 | 1.11 | 4.5x |
| Industrial | 7.10 | 1.42 | 5x |
| Victorian | 6.22 | 1.13 | 5.5x |
| Electric | 8.63 | 1.42 | 6.5x |
| Atomic | 8.87 | 1.36 | 7x |
| **to Modern** | **54.3 h (2.26 d)** | **13.0 h (0.54 d)** | |

- **A first run that uses `speed` reaches Modern in 13 hours, 4.2x faster than the pacing target assumes.** That is more than every prestige perk combined, ever.
- **Bug: the speed setting survives prestige.** `completePrestige` (game/engine.go) resets every manager but not `speedMultiplier`. Succumb and Reset do reset it. Measured: set 7x, prestige, and the new run is still at 7x while its cap is 1.0x (100 ticks took 28 s instead of 200 s). The speed is also saved and reloaded. So run 2 of a player who used `speed` starts at 7x.
- Offline catch-up counts ticks at the current speed too, so a 7x player also gets 7x the offline ticks (still capped at 24 hours of real time, at 50% efficiency).

If the calendar is meant to be the clock, `speed` has to go, or be capped low, or become a prestige unlock. Every option below assumes it does.

### The x3 production cap swallows every production bonus from the Victorian Age on

**Where:** `productionCap = 3.0` in game/engine.go, applied in `recalculateRates` as `rate *= clamp(1 + sum of production_all bonuses, 0.10, 3.0)`. Each resource's own pool (`knowledge_rate`, `gold_rate` and so on) gets the same clamp separately.

**What counts toward it** (all added into one `production_all` pool before the clamp):
- techs: 17 of them grant production_all (Steam Power +30%, Industrialization +50%, Electrification +20%, Mass Production +40%, Power Distribution +30%, Chemical Engineering +20%, and more later)
- wonders: Crystal Palace +15%, Hoover Dam +20%, and larger ones later
- permanent bonuses: milestone rewards, epoch events (for example +10% and +15% golden ages)
- **the prestige passive (+2% per level)**
- **the Cosmic Legacy (+10%)**
- timed events: festivals (+20%), faction boons, "double production" events, catastrophe debuffs (-10%)

**Succumb's legacy** adds per-resource bonuses (`<resource>_rate`), which sit under the per-resource clamp, plus +25% research speed per epoch, which has no cap at all (see the table in section 10).

**How full the pool is** (seed 1, measured at the end of each age; the multiplier is 1 + the pool):

| age | production pool | applied | room left under x3 | knowledge pool | room left |
|---|---|---|---|---|---|
| Primitive to Bronze | +5% | x1.05 | 195 points | 0 to +5% | 195+ |
| Iron | +10% | x1.10 | 190 | +15% | 185 |
| Classical to Renaissance | +15% | x1.15 | 185 | +55% to +160% | 145 to 40 |
| Colonial | +25% | x1.25 | 175 | +205% | **0** |
| Industrial | +165% | x2.65 | 35 | +205% | 0 |
| Victorian | +235% | **x3.00 (clamped)** | **0** | +245% | 0 |
| Electric | +320% | x3.00 | 0 | +285% | 0 |
| Atomic | +335% | x3.00 | 0 | +285% | 0 |
| Modern | +405% | x3.00 | 0 | +425% | 0 |
| Information | +410% | x3.00 | 0 | +595% | 0 |

By Modern, 205 points of production bonus sit above the clamp doing nothing. That is where the achievements audit's dead late production milestone rewards go, along with late techs (Quantum Mechanics +100%, Transcendence +200%) and late wonders (Warp Nexus +80%, Singularity Core +200%).

**What that does to a prestige bonus** (seed 1, same bonus inside the pool, the way the passive works today, versus applied after the clamp):

| bonus | Primitive to Industrial | Victorian to Atomic | to Modern | faster by |
|---|---|---|---|---|
| none | 30.6 h | 23.7 h | 54.3 h | - |
| +50%, inside the pool | 23.6 h | 24.9 h | 48.5 h | 1.12x |
| x1.5, after the clamp | 23.7 h | 17.7 h | 41.4 h | 1.31x |
| +100%, inside the pool | 20.2 h | 23.2 h | 43.4 h | 1.25x |
| x2, after the clamp | 19.3 h | 16.6 h | 35.9 h | 1.51x |

- **Inside the pool, a prestige production bonus does nothing from the Victorian Age on**: 44% of the run to Modern, and every age after it.
- Before Victorian it works the same either way.
- So every option below applies prestige as its own multiplier, **after** the clamp.

**Why bypass rather than raise the cap.** The clamp was added because stacked timed buffs (faction boons, events, festivals) ran a soak to x20 on knowledge. That job is still needed, and the income model the pacing rules use (`TypicalIncome`) assumes the x3 ceiling, so raising it retunes every late age. Prestige is different: it is permanent, bounded by design, and must be felt in every age. A separate layer keeps the clamp doing its job.

**Worth doing alongside (any option):** split the pool in two: in-run permanent sources (techs, wonders, milestones, epoch events) and timed buffs (events, festivals, boons, debuffs). Clamp only the timed pool. That revives the dead late rewards, but it is a pacing change of its own (late ages speed up), so it belongs with the slowdown retune, not before it. Move the Cosmic Legacy's +10% into the prestige layer either way; today it vanishes from the Victorian Age on.

## 3. Why the percentages barely move pacing

### Production multiplier and tick speed, measured (seed 1, time to Modern)

The production multiplier here is applied after the x3 cap, so it is a best case for any "+X% production" perk.

| multiplier | production x | time to Modern | faster by | tick speed x | time to Modern | faster by |
|---|---|---|---|---|---|---|
| 1 | - | 54.3 h | - | - | 54.3 h | - |
| 1.25 | 1.25 | 46.3 h | 1.17x | 1.25 | 43.4 h (computed) | 1.25x |
| 1.5 | 1.5 | 41.4 h | 1.31x | 1.5 | 36.2 h (computed) | 1.5x |
| 2 | 2 | 35.9 h | 1.51x | 2 | 27.2 h (measured) | 2.00x |
| 4 | 4 | 25.9 h | 2.09x | 4 | 13.6 h (computed) | 4x |

- **Tick speed is exact.** Every age at 2x took exactly half as long. The bot decides every 5 ticks and every system counts ticks, so tick speed scales everything, including events, raids and catastrophes. The other tick-speed rows are computed from that.
- **Production saturates.** Doubling production saves only a third of the time; quadrupling saves half.

### Shortening construction and research too

I added a hook that divides every construction time, and research times, by a factor (the game has no such bonus today). Seed 1, time to Modern:

| lever | today's curve | faster by | two-week curve (section 7) | faster by |
|---|---|---|---|---|
| none | 54.3 h | - | 7.81 d | - |
| construction and research x2 only | 47.0 h | 1.16x | - | - |
| production x1.5 | 41.4 h | 1.31x | 5.62 d | 1.39x |
| production x2 | 35.9 h | 1.51x | 4.73 d | 1.65x |
| production x4 | 25.9 h | 2.09x | 3.32 d | 2.35x |
| **production x1.4 + construction and research x1.4** | - | - | 5.44 d | **1.44x** |
| **production x2 + construction and research x2** | 26.8 h | **2.03x** | 3.84 d | **2.03x** |
| production x4 + construction and research x2 | 18.1 h | 3.00x | 2.39 d | 3.27x |

- **Scaling production and construction and research together by k makes the run about k times faster**: 1.44x at k = 1.4; at k = 2, 2.03x and 2.02x on today's curve (seeds 1 and 2) and 2.03x and 1.87x on the two-week curve. It behaves like tick speed for the economy, while events, raids, catastrophes and harbingers stay on the real clock.
- Either lever alone saturates. Construction alone barely helps (1.16x); production alone gives about 1.5x per doubling.
- On the two-week curve production goes further (x2 gives 1.65x instead of 1.51x), because the hand-typed build times don't stretch with the targets.

I call "production x k, construction and research / k, storage x k" **k-scaling** below. It is the lever every option uses.

### What binds in each age

An age ends when three things are true: its resource and building requirements are met, and its wonder is built. The wonder has to be banked first, then built, and its build time is fixed (the pacing rule caps it at 1/6 of the age's target; every wonder sits at that cap). Times below are hours from entering the age, seed 1.

| age | 1x time | bank full | wonder built | wonder build share | last other requirement met | 2x prod | 4x prod | speedup at 4x |
|---|---|---|---|---|---|---|---|---|
| Primitive | 0.29 | 0.25 | 0.29 | 14% | knowledge at 0.22 | 0.23 | 0.19 | 1.55x |
| Stone | 0.95 | 0.82 | 0.95 | 13% | food at 0.82 | 0.74 | 0.53 | 1.78x |
| Bronze | 1.73 | 1.36 | 1.61 | 14% | knowledge at 1.73 | 1.00 | 0.69 | 2.50x |
| Iron | 3.02 | 2.60 | 3.02 | 14% | stone at 2.81 | 2.15 | 1.06 | 2.85x |
| Classical | 3.82 | 3.24 | 3.82 | 15% | Military Academy at 3.56 | 2.54 | 1.84 | 2.07x |
| Medieval | 3.26 | 2.55 | 3.26 | 22% | Guildhall at 3.22 | 2.13 | 1.59 | 2.05x |
| Renaissance | 5.47 | 2.71 | 3.66 | 17% | steel at 5.47 | 2.61 | 1.68 | 3.26x |
| Colonial | 4.98 | 3.87 | 4.98 | 22% | gold at 3.87 | 3.72 | 2.97 | 1.67x |
| Industrial | 7.10 | 5.94 | 7.10 | 16% | Tenement at 6.98 | 4.20 | 2.93 | 2.43x |
| Victorian | 6.22 | 4.92 | 6.22 | 21% | Steam Turbine at 5.99 | 4.15 | 3.14 | 1.98x |
| Electric | 8.63 | 6.01 | 7.46 | 17% | Power Station at 8.63 | 5.27 | 4.07 | 2.12x |
| Atomic | 8.87 | 7.06 | 8.80 | 20% | Bunker Complex at 8.87 | 7.15 | 5.24 | 1.69x |

What this shows:
1. **Fixed build times are the floor.** Wonder construction alone is 9.8 of the 54.3 hours (18%) at 1x. At 2x production it is 27% of the run, at 4x 38%. No production bonus touches it. Only tick speed does.
2. **Required buildings finish last in 6 of 12 ages** (Military Academy, Guildhall, Tenement, Steam Turbine, Power Station, Bunker Complex). Their copies get pricier (x1.13 each), are bought one batch at a time as they become affordable, and each batch has a build time. More production buys them sooner, but the last batch still takes its build time.
3. **Production-bound stretches are short.** The bank fills, then the age waits on construction. The ages that respond best to production (Bronze, Iron, Renaissance) are the ones whose wait is on a resource, not a build.
4. **The x3 cap removes the passive's production part from the Victorian Age on** (section 2), which is 23.7 of the 54.3 hours. Only the +1% tick speed per level survives there.
5. **Research rarely binds on its own.** Halving construction and research times together gained only 16%.

**One number for all of it.** A two-part model fits every lever run above to within about 2%: time = T x (share / production + (1 - share) / construction speed). The share that scales with production is **69% on today's curve and 78% on the two-week curve**; the rest scales only with construction and research times. That is why x4 production tops out near 2x, and why prestige has to shorten construction and research as well (or use tick speed) to make a run 3x faster.

## 4. Caps under multipliers

### Do storage caps grow with prestige? No.

Storage is computed in `recalculateRates` (game/engine.go) as base storage + flat bonuses from storage buildings, techs, milestones and the prestige Storage Bonus. Nothing multiplies it. A prestige production bonus, today's or any new one, raises what flows in but never the cap. The Storage Bonus perk adds a flat +20 per tier (+100 at tier 5) to every resource.

### Storage Covenant hold time under production multipliers

Max storage per age holds this many hours of the age's typical production (tightest resource of each age). The covenant asks for 1.5 hours.

| age | 1x | 1.25x | 1.5x | 2x | 4x |
|---|---|---|---|---|---|
| Primitive | 4.89 h | 3.91 | 3.26 | 2.45 | 1.22 |
| Stone | 1.54 | 1.23 | 1.03 | 0.77 | 0.38 |
| Bronze | 1.60 | 1.28 | 1.07 | 0.80 | 0.40 |
| Iron | 1.51 | 1.21 | 1.01 | 0.75 | 0.38 |
| Classical | 1.50 | 1.20 | 1.00 | 0.75 | 0.38 |
| Medieval | 1.52 | 1.21 | 1.01 | 0.76 | 0.38 |
| Renaissance | 1.89 | 1.51 | 1.26 | 0.95 | 0.47 |
| Colonial | 1.56 | 1.25 | 1.04 | 0.78 | 0.39 |
| Industrial | 1.54 | 1.24 | 1.03 | 0.77 | 0.39 |
| Victorian | 1.53 | 1.22 | 1.02 | 0.76 | 0.38 |
| Electric | 1.53 | 1.23 | 1.02 | 0.77 | 0.38 |
| Atomic | 2.24 | 1.79 | 1.49 | 1.12 | 0.56 |
| Modern | 2.61 | 2.09 | 1.74 | 1.31 | 0.65 |
| Information | 1.51 | 1.21 | 1.01 | 0.76 | 0.38 |
| Digital | 1.85 | 1.48 | 1.24 | 0.93 | 0.46 |
| Cyberpunk | 1.62 | 1.30 | 1.08 | 0.81 | 0.40 |
| Fusion | 2.52 | 2.02 | 1.68 | 1.26 | 0.63 |
| Space | 4.39 | 3.51 | 2.93 | 2.20 | 1.10 |
| Interstellar | 14.3 | 11.5 | 9.6 | 7.2 | 3.6 |
| Galactic | 24.0 | 19.2 | 16.0 | 12.0 | 6.0 |
| Quantum | 17.1 | 13.7 | 11.4 | 8.6 | 4.3 |

Most ages sit right on the 1.5-hour line today. **Any production bonus of more than about 1.0x breaks the covenant in most ages from Stone to Information.** At 1.5x they hold 1 hour; at 2x, 45 minutes.

### What a check-in player loses at full storage

Share of construction-resource output (in price units) that hit a full store and was lost, over the whole first run. Overflow is what the wonder bank caught.

| player | production | days to Modern | lost at the cap | overflow caught |
|---|---|---|---|---|
| active (greedy) | 1x | 2.26 | 17% | 0.1% |
| check-in 1 h | 1x | 2.99 | 16% | 1.3% |
| check-in 3 h | 1x | 3.87 | 30% | 2.0% |
| check-in 8 h | 1x | 6.27 | 60% | 1.5% |
| check-in 8 h | 1.5x | 6.22 | 81% | 0.8% |
| check-in 8 h | 2x | 6.33 | 79% | 0.8% |
| check-in 8 h | 4x | 4.35 | 87% | 0.7% |

- The active player's 17% is resources it doesn't need piling up. That is the floor; part of every check-in player's loss is the same surplus.
- **At 8-hour check-ins, 1.5x and 2x production make the run no faster at all.** 4x helps only 1.44x (active: 2.09x).
- **Bigger storage doesn't rescue it either.** On the one-week curve (section 7), tripling every storage cap took 3-hour check-ins from 7.4 to 6.2 days, but 8-hour check-ins only from 9.1 to 8.8 days, and six times the storage gave 8.9. The 8-hour player is limited by how many purchase rounds it gets (one per visit, with each copy 13% dearer than the last), not by what it can hold.
- Overflow catches almost nothing. It only banks the age's wonder, and only the resources the wonder still needs.

So a production perk is worth little to the 8-hour player however storage is sized; what helps that player is time (shorter construction and research) and automation that spends between visits. Storage x k still matters for active and 1 to 3 hour players.

### The flat perks after the first few minutes

| perk at tier 5 | what it adds | share of that age's cap or need | measured effect on a run |
|---|---|---|---|
| Storage Bonus | +100 storage per resource | Stone: 9.5% of storage; Bronze: 0.03% of max; Medieval on: under 0.01% | together with the next three: 5 minutes saved in the first three ages (172 to 167 min); no difference after that (53.8 h vs 54.2 h to Modern) |
| Housing Bonus | +10 housing | Stone 3.6%; Medieval 0.6%; Industrial 0.2%; Electric 0.02% | |
| Starting Food | +125 food | 12% of the first gate (1,000 food) | |
| Starting Wood | +125 wood | about 6 huts (14 wood, rising 13% a copy) | |

The flat perks cost 81 of the shop's 277 points (29%) and are worth about five minutes per run.

## 5. Gate headroom watch list

**Short answer: all three tight rows are wonder parts, and wonders are banked a deposit at a time.** A player never has to hold a wonder's cost at once (`BankWonderResource`, game/engine.go), and overflow banks what a full store would lose. So the 1.09x, 1.69x and 1.77x rows are tight only against the Gate Covenant's own rule (each wonder part must fit one full store), not against any player. What a player must be able to hold is the largest single non-wonder payment.

Below, "own-age copies" counts only that age's storage building (older ages' storage can no longer be built once you move on) plus base storage.

| advance | tight row (bankable) | largest single payment a player must hold | headroom | own-age storage copies needed | their cost (price units; the age's wonder costs 40) | build time |
|---|---|---|---|---|---|---|
| Electric to Atomic | Hoover Dam, 71B steel, 1.69x | copy 15 of Power Station or Electric Arc Furnace: 36.1B steel | 3.33x | 11 of 25 Electric Warehouses | 35 (maxing all 25: 250) | 12.5 min per copy, copies build in parallel |
| Digital to Cyberpunk | World Simulation, 34T steel, 1.77x | copy 15 of Neural Grid: 21.2T steel | 2.83x | 15 of 25 Digital Archives | 44 (maxing: 170) | 20 min per copy |
| Space to Interstellar | Dyson Scaffold, 5,500T steel, 1.09x | copy 10 of Solar Collector Array: 1,650T electricity | 3.64x | 9 of 25 Orbital Depots | 24 (maxing: 240) | 27.5 min per copy |

- **Can a normal player make it? Yes.** Building about half the age's storage covers every must-hold payment. Nobody needs every storage building maxed, or any older-age storage at all.
- **How long does it take? Build time is minutes; the cost is about one wonder's worth.** The copies build in parallel (storage queues up to its max count). Covering the must-hold payments costs 24 to 44 price units, roughly one wonder. Maxing the age's storage would cost 170 to 250, four to six wonders.
- **The one real risk is a player who tries to hold a whole wonder part.** With only Orbital Depots (25 x 200T = 5,000T) they can't hold Dyson Scaffold's 5,500T of steel. Banking in parts works; the UI should say so where it shows the wonder.
- **The static test is the fragile part.** A 10% price bump on Dyson Scaffold's steel would fail `go test ./smoke`, even though no player is affected. Option: exempt wonder parts from the margin rule, or keep it as a design tripwire and say so in its comment.

## 6. A maxed shop against a fresh account

A veteran who maxed the shop has prestiged about 38 times. I put the engine at prestige level 38 with every perk at tier 5 (through a real prestige, so starting resources and caps are the game's own) and played seed 1 to Modern. Three controls separate the parts.

| age | fresh | level 1, nothing bought | level 1, flat perks only | level 1, maxed shop | level 38, passive only | **level 38, maxed shop** | veteran speedup |
|---|---|---|---|---|---|---|---|
| Primitive | 0.29 h | 0.27 | 0.25 | 0.20 | 0.16 | 0.12 | 2.5x |
| Stone | 0.95 | 0.96 | 0.96 | 0.77 | 0.54 | 0.46 | 2.1x |
| Bronze | 1.73 | 1.64 | 1.58 | 1.20 | 0.78 | 0.70 | 2.5x |
| Iron | 3.02 | 2.84 | 3.16 | 2.59 | 1.42 | 1.22 | 2.5x |
| Classical | 3.82 | 4.18 | 3.82 | 2.93 | 1.87 | 1.68 | 2.3x |
| Medieval | 3.26 | 2.80 | 3.43 | 2.34 | 1.76 | 1.53 | 2.1x |
| Renaissance | 5.47 | 5.46 | 5.36 | 4.70 | 2.14 | 1.65 | 3.3x |
| Colonial | 4.98 | 4.75 | 4.69 | 4.09 | 2.88 | 2.50 | 2.0x |
| Industrial | 7.10 | 6.64 | 6.43 | 5.47 | 3.69 | 3.32 | 2.1x |
| Victorian | 6.22 | 6.52 | 6.14 | 4.49 | 4.47 | 3.61 | 1.7x |
| Electric | 8.63 | 8.62 | 8.54 | 6.62 | 6.31 | 5.77 | 1.5x |
| Atomic | 8.87 | 9.52 | 9.44 | 8.02 | 6.95 | 5.69 | 1.6x |
| **to Modern** | **54.3 h** | 54.2 h | 53.8 h | 43.4 h | 33.0 h | **28.2 h (1.18 d)** | **1.9x** |

- **A maxed veteran reaches Modern in 1.18 days, 1.9x faster than a fresh account.** It takes about 38 runs (2.5 months) to get there.
- **The passive does most of it** (54.3 to 33.0 hours): at level 38 it is +38% tick speed and +76% production. The shop adds the rest (33.0 to 28.2), almost all from its +25% tick speed.
- **The veteran's speedup halves after the Industrial Age** (2 to 2.5x early, 1.5 to 1.7x in Electric and Atomic). That is the x3 cap eating the passive's production.
- **The flat perks are worth about 5 minutes per run** (the first three ages take 172 minutes without them and 167 with them; the rest is noise).

## 7. Slowing the game down (scope change)

### First-run targets side by side

I swapped `AgeTargets` and changed nothing else. The whole economy derives from that table (producer payback, build-time and research-time caps, deal values, Appease prices), so each curve is a one-table change in the engine.
- **Today**: primitive 15m, stone 45m, bronze 1.5h ... atomic 12h.
- **One week (primary, per the research)**: today's targets x2.6 from the Bronze Age on: bronze 3.9h, iron 6.5h, classical 9.1h, medieval 11.7h, renaissance 15.6h, colonial 18.2h, industrial 20.8h, victorian 23.4h, electric 26h, atomic 31.2h; after Modern 31h rising to 62h.
- **Two weeks (your sketch, the alternative)**: bronze 4h, iron 12h, classical 16h, medieval 20h, renaissance 24h, colonial 28h, industrial 32h, victorian 36h, electric 40h, atomic 44h; after Modern 48h rising to 72h.
- (I also ran today x2: active 4.2 days, 1-hour 4.9, 3-hour 6.0, 8-hour 8.8.)

Days to the first prestige (Modern), seed 1:

| curve | sum of targets | active | 1-hour check-ins | 3-hour | 8-hour |
|---|---|---|---|---|---|
| today | 2.7 | 2.26 | 2.99 | 3.87 | 6.27 |
| **one week** | 7.0 | **5.30** | 6.59 | **7.38** | 9.08 |
| one week, storage x3 | | 5.40 | | 6.21 | 8.80 |
| two weeks | 10.7 | 7.81 | 8.86 | 8.95 | 12.83 |

(Seed 2 lands within 3 to 9%: two weeks active 7.60, 3-hour 8.62, 8-hour 11.73.)

Hours per age:

| age | today active | one week active | one week 3h | one week 8h | two weeks active | two weeks 3h | two weeks 8h |
|---|---|---|---|---|---|---|---|
| Primitive | 0.3 | 0.3 | 1.8 | 1.8 | 0.3 | 1.8 | 1.8 |
| Stone | 0.9 | 0.9 | 5.8 | 9.0 | 0.9 | 5.8 | 9.0 |
| Bronze | 1.7 | 3.4 | 5.8 | 13.3 | 3.5 | 5.8 | 13.3 |
| Iron | 3.0 | 6.3 | 10.6 | 23.9 | 8.0 | 10.2 | 33.5 |
| Classical | 3.8 | 8.4 | 13.5 | 19.2 | 14.6 | 14.0 | 20.2 |
| Medieval | 3.3 | 7.8 | 7.8 | 9.8 | 11.1 | 10.0 | 17.5 |
| Renaissance | 5.5 | 13.2 | 10.4 | 13.9 | 20.9 | 25.6 | 31.0 |
| Colonial | 5.0 | 11.8 | 13.4 | 15.8 | 16.1 | 19.6 | 24.2 |
| Industrial | 7.1 | 15.0 | 23.0 | 33.8 | 22.4 | 29.8 | 40.1 |
| Victorian | 6.2 | 16.1 | 21.9 | 16.0 | 22.7 | 23.1 | 30.3 |
| Electric | 8.6 | 17.4 | 26.6 | 34.2 | 29.1 | 26.6 | 41.7 |
| Atomic | 8.9 | 26.5 | 36.4 | 26.9 | 37.9 | 42.4 | 45.5 |

- **Both curves work as pure config changes.** The active player slows 2.3x (one week) and 3.5x (two weeks).
- The active bot lands at 0.76x (one week) and 0.73x (two weeks) of the targets (today 0.83x). Build and research times are capped copies of hand-typed literals, and the caps only ever shorten (the timer survey below), so they grow less than the targets.
- **The one-week curve hits the research target** for active play (5.3 days) and 3-hour check-ins (7.4 days). 8-hour check-ins still take 9.1 days.
- **Tripled storage narrows the gap without speeding the active player**: active 5.30 to 5.40 days (no change), 3-hour 7.38 to 6.21. That is exactly the "cut active play's lead" lever the research asks for. It does little at 8 hours (next part).
- **The two-week sketch against your intent** (attentive about 10 to 11 days, a few check-ins a day about 2 weeks): 7.8 days active, about 9 days at 1 to 3 hour check-ins, 12.8 at 8 hours. The greedy bot is a near-perfect player, so a real attentive human is nearer the 1-hour bot. If you want the bot itself at 10.5 days, scale Bronze through Atomic by about 1.35 and re-measure.

### The gap between active and idle play

Time to Modern as a multiple of the active player's time:

| check-in | today | one week | two weeks |
|---|---|---|---|
| 1 hour | 1.32x (2.99 d) | 1.24x (6.59 d) | 1.13x (8.9 d) |
| 3 hours | 1.71x (3.87 d) | 1.39x (7.38 d) | 1.15x (9.0 d) |
| 8 hours | 2.77x (6.27 d) | 1.71x (9.08 d) | 1.64x (12.8 d) |

Longer ages narrow the gap by themselves, because fewer ages are shorter than one visit.

Where the 8-hour player's time goes (two-week sketch; the one-week curve looks the same):
- **The first four ages cost 58 hours against the active player's 13** (a third of the gap). Primitive and Stone are shorter than one visit.
- **Purchase rounds are the binding item.** In the Iron Age the 8-hour player met every resource requirement at its first visit (7.9 h) but finished the required Agora and Trading Post copies only at its fourth (31.9 h), while 74% of its construction output hit full stores. The build plan already advances when ready and can plan the next age ahead, and the idle bot uses both; still, most copies are bought at visits. Bigger storage alone didn't fix it (section 4).
- Over the whole run 36% of construction output is lost at the cap (1-hour check-ins: 16%). Classical to Atomic run 1.2 to 1.5x the active times.

How to narrow it further, so pacing depends mostly on the calendar:
1. **Adopt a slower curve.** One week takes 3-hour check-ins from 1.71x to 1.39x of the active time; two weeks to 1.15x.
2. **Let overflow pay the plan, and let the plan hold more.** Overflow today only banks the age's wonder (it caught 1 to 2% of output). Letting it fund queued plan items, with room for more than 30 items and funding beyond 24 hours, gives the 8-hour player purchase rounds between visits.
3. **A research queue.** Research runs one tech at a time, so a check-in player starts one tech per visit.
4. **Size storage to a few hours, and scale it with k.** Tripling storage took 3-hour check-ins from 7.4 to 6.2 days on the one-week curve. It did little at 8 hours.

**More levers, measured on the one-week curve** (seed 1, days to Modern):

| lever | active | 3-hour | 8-hour | 3-hour / active | 8-hour / active |
|---|---|---|---|---|---|
| none | 5.30 | 7.38 | 9.08 | 1.39x | 1.71x |
| storage x3 | 5.40 | 6.21 | 8.80 | 1.15x | 1.63x |
| storage x6 | - | - | 8.93 | - | - |
| construction times x2 (longer) | 6.69 | 7.81 | 11.16 | 1.17x | 1.67x |
| k = 2, no storage scaling | - | - | 6.89 | - | - |
| **k = 2 with storage x2 (full k-scaling)** | **2.73** | **3.89** | **6.60** | 1.42x | **2.42x** |

- **Bigger storage is the cleanest gap-closer for 3-hour players**: it doesn't touch active play.
- **Longer construction** slows active play 26%, 3-hour players 6% and 8-hour players 23%: it narrows the 3-hour gap only.
- **Prestige widens the 8-hour gap unless automation comes with it.** Full k-scaling makes active play 1.94x faster and 3-hour play 1.90x, but 8-hour play only 1.38x. The 8-hour player needs purchase rounds between visits (overflow paying the plan, a research queue), or veteran runs leave them further behind.

Targets I would enforce on the one-week curve: 3-hour first prestige at most 1.3x the active time, 8-hour at most 1.5x.

### Every time-based system, and what a 4x slowdown does to it

The survey assumed the two-week sketch (about 4x). The one-week curve is a 2.6x slowdown: the same items break, a bit over half as badly.

One tick is 2 seconds at 1x. "Derived" means it already follows `AgeTargets`. "Cap only" means it is `min(hand-typed value, share of the target)`: a longer target raises the ceiling but never lengthens the hand-typed value. "Hard-coded" means a fixed number of ticks or hours. (Survey of origin/master; file:line references are in the appendix.)

| system | today | scales with AgeTargets? | under a 4x slowdown |
|---|---|---|---|
| Producer rates (Payback Rule) | payback 56 s in Primitive, 22 min in Iron, 1.9 h in Renaissance | derived | construction output drops to 1/4 (the slowdown itself). Food, faith, culture and soldiers keep hand-set rates |
| Building times | cap = age target / 6 (wonders), / 48 (storage). 173 of 301 buildings sit at the cap today, including all 22 wonders and all 21 storage buildings | cap only | only 67 buildings stretch the full 4x; the 50 Primitive to Medieval buildings below the cap don't lengthen at all; Colonial to Electric staples 1.2 to 1.7x; Atomic-and-later staples stay at 2 h |
| Research times | cap = age target / 8; all 73 techs sit at it | cap only | Primitive to Victorian techs lengthen 1.65 to 3.65x (Iron to Renaissance about 2x); Electric on scale 4x |
| Random events (including raids: bandits, pirates, data breach, tribal raid) | one every 5 to 20 min (150 to 600 ticks); buffs last up to 40 s (epoch ones up to 7 min) | hard-coded | 4x as many events and raids per age; each raid steals fixed amounts against 1/4 the income; timed buffs feel 4x shorter |
| War raids, diplomacy | war raid every 80 s; war ends after 10 min quiet; opinion drift every 50 s; deal re-roll every hour | hard-coded | 4x more per age; opinions settle 4x faster per age |
| Trade routes | 8 to 20 ticks per run, fixed lots | hard-coded | routes worth about 4x more against income |
| Catastrophes | one roll per epoch transition, 12 to 18% | event-driven | same count per run. Endure's debuff (7.2 min, -10%) becomes trivial |
| Harbingers | arrive on epoch entry, hand off each advance, resolve at the transition; no tick windows | event-driven | lead time grows with the ages automatically |
| Catastrophe redesign (planned) | strike tick drawn across the epoch, random harbinger lead time | not built | draw both as a share of the epoch's expected length, never as fixed ticks (see section 10) |
| Expeditions | 2 to 3.3 min each; auto-expedition every 30 min | hard-coded | 4x more runs and encounters per age, fixed loot |
| Festivals, black market, boons, milestone boosts | festival buff 5 min, cooldown 10 min; black market 8 min; boons 25 to 100 min; chain boosts 5 min | hard-coded | 4x more uses per age; each boost 4x less important |
| Survivor milestones | 5 h 33 m and 27 h 47 m of play (descriptions quote the times) | hard-coded | earned ages earlier |
| Offline catch-up | 24 h cap, 50% efficiency; plan funding looks 24 h ahead | hard-coded | a third to half of one slowed post-Modern age |
| Storage Covenant | 1.5 h of typical income | hours x derived income | literal storage holds about 4x longer (6 h); caps lose their bite unless storage is retuned on purpose |
| Idle targets (smoke) | 1 h: 3.5 d, 3 h: 5 d, 8 h: 8 d | hard-coded | fail under enforce; need new targets |
| Smoke `PacingTargets` | a literal copy of AgeTargets | copy (a test keeps them equal) | edit both together |
| Smoke age timeouts | 4 x target, at least 1 h | derived from the copy | automatic |
| Smoke budgets | fast: 3 seeds to Bronze, 300 h; full: 8 seeds, 2 cycles to Digital, 600 h; styles 300 to 720 h; prestige 400 h; soft-lock 30 min | hard-coded | full needs about 620 h per seed; styles, prestige and idle budgets run out |

**Smoke runtime.** Last night's full tier took 1 h 53 min (progression 33 min, styles 23 min, fuzz 44 min, idle 3 min, prestige 4 min). The bot-driven scenarios scale with simulated days, so at 3.5x they add about 2.5 hours: roughly **4.5 hours per nightly**, inside GitHub's 6-hour job limit but close. Halve the progression seeds (8 to 4) or shard them the way the weekly job already does (`-merge`). The PR fast tier stops at Bronze and grows only about 1.6x (Bronze is 4 h instead of 1.5 h).

**What breaks first:**
1. **Build and research times barely slow down**, because the caps only shorten. Either re-type the literals or make the normalizers scale (a building's time = a share of its age's target) instead of cap.
2. **Random events keep a 5 to 20 minute clock**, so raids per age quadruple against a quarter of the income. Tie the event delay to the age target.
3. **Flow gates get 4x easier** (food, faith, culture and soldiers keep their rates) and **Appease gets 4x dearer** against fixed faith and culture storage.
4. **Smoke budgets and the idle targets** fail the nightly under enforce until they're updated.
5. **The 24-hour offline cap** covers much less of an age. Consider 48 to 72 hours with the slowed curve.

## 8. How strong prestige has to be

Your starting take was run 2 at about 1.6 days, run 5 about 1 day and a veteran about 15 hours (on today's curve). In k-scaling terms that is k = 1.4, 2.3 and 3.6.
- **Production percentages alone can't get there.** x4 production (after the clamp, the best case) reaches Modern in 25.9 hours. Fifteen hours is out of reach without shortening construction and research.
- **k-scaling can.** Measured: 1.44x at k = 1.4, and 1.9 to 2.0x at k = 2 on every curve and seed tried.
- **The research sets the bar higher than your take for run 2**: old ground 2 to 4x faster (k = 2 or more on ages already done), not 1.4x. I size the options below to k = 2 on old ground in run 2, rising to about 4.
- **Today's prestige is far short of either.** Run 2 is 1% faster with the bot's own buying and 13% with speed first. The best case, a maxed veteran after about 38 runs, is 1.9x, and most of it comes from the passive's tick speed.

## 9. What the player and designer research changes

### The targets

| research finding | what it means here | measured today |
|---|---|---|
| Later runs must be clearly faster (Trimps and CivIdle reviews) | run 2 must feel different, not 5% | runs 1 to 3: 2.26, 2.24, 2.18 days |
| Run 2 should cover old ground 2 to 4x faster and go further (Pecorella, Kongregate) | k of about 2 on ages you have done, 1x on new ones | 1.0 to 1.15x |
| Players reset at +50 to 200% more prestige currency; the natural reset point should move deeper each run | points must climb steeply with depth | points per day fall with depth (11.9 at Modern, 3.7 at Transcendent) |
| About a week to the first prestige for everyone; most of the fix is cutting active play's lead | active about 5 to 6 days, a few check-ins a day about 7 | active 2.26, 8-hour 6.27 (2.8x apart) |
| Catching up after a reset should take hours | only possible if old ground is much faster or skipped | 2 days |
| Automation should survive the reset | keep the setup | the plan, research and trade routes are wiped |
| No multi-day walls; something new every day or two | watch the longest quiet stretch | see below |
| Absence must not be punished | no progress lost beyond what the player chose | see below |

### What a prestige wipes today (automation included)

| system | after prestige |
|---|---|
| Build plan (queued builds, the advance item) | wiped |
| Research (every tech); there is no research queue | wiped |
| Buildings, workers and their assignments | wiped |
| Trade routes | wiped |
| Factions met, opinions, deals | wiped |
| Milestones and their permanent bonuses | wiped (earned again) |
| Harbinger history, awakenings | wiped |
| Wonder overflow setting | kept |
| `speed` setting | kept (the bug) |
| Ruins, Succumb legacy, Cosmic Legacy, prestige | kept |
| Ancient Memory | 40% chance of one free tech |

So every run starts by rebuilding the plan and re-researching every tech by hand. The legacy kit (section 10) keeps the plan as a template, remembers the research order as a queue, keeps factions met and re-applies worker shares.

### The longest stretch with nothing new to decide

Longest gap between new decisions (a building type built for the first time, or a tech finished) inside each age, active player, seed 1:

| age | today | one week | two weeks |
|---|---|---|---|
| Renaissance | 1.6 h | 4.4 h | 6.7 h |
| Colonial | 1.1 h | 1.9 h | 3.3 h |
| Industrial | 2.5 h | 3.9 h | 7.5 h |
| Victorian | 1.7 h | 4.0 h | 8.4 h |
| Electric | 1.5 h | 9.7 h | 12.7 h |
| Atomic | 5.7 h | 18.4 h | 26.4 h |
| Modern | 3.1 h | 3.5 h | 7.1 h |
| Information | 9.4 h | 26.4 h | 15.6 h |
| Digital | 5.1 h | 10.2 h | 12.8 h |
| Cyberpunk | 6.7 h | 23.2 h | 23.3 h |
| Fusion | 13.8 h | 28.5 h | 46.4 h |

**Flag:** on both slowed curves the Atomic Age and the post-Modern ages (Information, Cyberpunk, Fusion) go 18 to 28 hours with no new building type or tech, and Fusion on the two-week curve goes 46 hours. That is inside "something new every day or two" for an active player, but a check-in player sees one or two visits in a row with nothing to decide. Each needs a mid-age unlock: a tech that opens a building halfway through, a milestone, or the harbinger (the "randomize the when" redesign can put its visits there). Information is also the one age that runs 1.4 to 1.5x its target on every curve (20 hours today, 54 on the one-week curve).

### An early, optional reset at the Medieval Age

`CanPrestige` requires the Modern Age today; the formula already guarantees 1 point from the Medieval Age. If prestige opened at Medieval with **today's formula**, resetting there would pay about twice the points per day of pushing to Modern:

| reset at | today's formula (level 0) | days (today's curve) | points per day |
|---|---|---|---|
| Medieval | 9 | 0.41 | 22 |
| Modern | 27 | 2.26 | 12 |

The square-root divisor makes spamming self-limiting only slowly (Medieval resets would pay 9, 6, 5, 4, 4 ...), and it punishes a veteran's deep runs just as much. So an early reset needs the depth formula:

| reset at | depth points, x3 per epoch | days (one-week curve, active) | points per day |
|---|---|---|---|
| Medieval | 9 | 0.81 | 11 |
| Modern | 120 | 5.30 | 23 |
| through Digital | 363 | 9.91 | 37 |

In the literature's terms, a square-root formula needs 4x the earnings to double the currency and a cube root 8x. Here the lever is depth: with weights tripling each epoch, doubling a run's points takes about 1.5 to 2 more ages at the frontier (Modern 120, through Information 282, through Digital 363, through Fusion 849), that is 3 to 5 days of frontier play on the one-week curve.

With weights tripling each epoch (1 per Stone-era age, 3 Iron, 9 Steel, 27 Electric, 81 Digital, 243 Neon, 729 Cosmic), a Medieval reset is a taste (it opens the shop and buys one quality-of-life unlock), and pushing on is clearly better per day. No divisor is needed: the +50-200% rule does the work. After a first Modern run (120 points), a second Medieval reset adds 9 (+8%), which nobody takes; a second Modern run adds 120 (+100%); by the time a player holds 600 points only a run through Digital (+60%) or deeper is worth it. The natural reset point moves about one epoch deeper every two runs.

### Absence must not be punished: where it is today

1. **A closed game** catches up at 50% efficiency, capped at 24 hours, and skips events, expeditions, trade routes, diplomacy and deals.
2. **Full storage** throws away output while the player is away (60% of construction output at 8-hour check-ins).
3. **Research runs one tech at a time**; it sits idle between visits.
4. **A pending catastrophe blocks advancing**, and the plan's advance item waits behind it, as does prestige. Today the roll comes at an epoch transition, which the plan can reach while the player sleeps; after the "randomize the when" redesign, mid-age strikes will land while the player is away more often. The economy keeps running, but the civilization is stuck at the gate until the player chooses. Fix: let the plan carry a standing choice ("Endure unless I answer"), or let the advance through while the choice waits.
5. **A pending Last Passage blocks prestige** the same way.
6. **The plan holds 30 items and funds 24 hours ahead**, so a two-day absence on the slowed curves outruns it.
7. Harbinger choices never expire (good), and deals only re-roll in live play (fine).

### Succumb as an early taste of prestige

| condition | today |
|---|---|
| Reachable on day 1 or 2 | partly. The first chance is the passage into the Iron epoch (the Bronze-to-Iron advance), about 5 hours into a run on the one-week curve. But it needs the 12 to 18% roll or an Invite, and after the "randomize the when" redesign an Invite needs a fated doom |
| Telegraphed and chosen, never applied offline | yes. The harbinger warns, and the choice waits for the player |
| Lost buildings rebuilt within hours | only early. Succumb resets the whole run to the Primitive Age. At the first passage that costs about 5 hours on the one-week curve; at the Iron-to-Steel passage (Medieval to Renaissance) about 1.1 days. Ruins (50% output, at most 24) soften it a little |
| Its bonus visibly works like prestige | no. +25% research speed per epoch barely moves pacing (research rarely binds), and the per-resource legacy (+20% iron for the Iron epoch) sits under the per-resource x3 clamp and touches one or two resources |

To make it a real taste: guarantee one Invite in the Iron epoch of every first run; make the reward a visible economy boost after the clamp (in Option C, +1 mastery for the epoch's three ages; in B, a Momentum bump; in A, a free tier); and let the player keep the current epoch's first age on Succumb so the rebuild is hours, not a day.


## 10. Redesign options

### Foundations every option needs

1. **Fix the `speed` carry-over bug.** Reset `speedMultiplier` in `completePrestige` and clamp it to the cap on load.
2. **Take `speed` out of normal play.** Stop wonders raising the cap (it stays at 1x), or make a higher cap a late prestige unlock. Otherwise no pacing target means anything.
3. **Apply prestige after the x3 clamp**, as its own multiplier. Move the Cosmic Legacy's +10% there too.
4. **Scale storage with the prestige factor** (storage x k), so the Storage Covenant holds at every prestige level.
5. **Points by depth, x3 per epoch, no divisor.** Each completed age pays 1 (Stone era), 3 (Iron), 9 (Steel), 27 (Electric), 81 (Digital), 243 (Neon) or 729 (Cosmic). Medieval pays 9, Modern 120, through Digital 363, through Space 1,092, the whole game 4,008. Keep the milestone, tech and building terms out (or as a small bonus).
6. **Open prestige at the Medieval Age** as the early, optional reset (section 9). The depth weights make it a taste, not a farm.
7. **A legacy kit so automation survives the reset**: the plan as a template, a research queue that remembers the last run's order, factions stay met, worker shares re-apply. The first two are the first shop items.
8. **Stop punishing absence**: a standing catastrophe choice for the plan, a research queue, storage sized to a check-in, overflow that pays the plan (section 7).
9. **Rewire the dead perks behind their existing keys** (upgrade keys are saved; never rename them).
10. **Re-run a real milestone feasibility check** after any retune. The achievements audit found 7 milestones that can never be completed under today's cost curves and caps, and `TestMilestonesAreFeasible` passes anyway.

### Option A: Repair the shop

Keep points, tiers and the passive; make the perks real.
- **Perks** (keys unchanged, effects rewired, 10 tiers, tier t costs 15 x 1.6^t depth points):
  - `gather_boost` becomes **Production**: +30% production per tier, after the clamp.
  - `research_speed` becomes **Construction and Research**: build and research times x0.9 per tier.
  - `storage_bonus` becomes +10% storage per tier (on top of storage x k).
  - `population_cap` becomes +5% housing per tier.
  - `starting_food` and `starting_wood` become **Head Start**: a starter stockpile that grows with the tier, and at tier 5 the run begins in the Bronze Age.
  - `tick_speed`, `military_power`, `expedition_loot` stay as they are.
- **Passive**: +2% production per level, moved after the clamp; keep +1% tick speed.
- **Maxed** (production x4, construction and research 2.9x faster): the two-part model gives 3.6x, so about 1.5 days to Modern on the one-week curve and 2.1 on the two-week curve.
- **For:** smallest change; players already know the shop. **Against:** the speedup depends on what each player buys; nothing makes the frontier feel different from known ages; with depth points a good run 2 buys a lot at once, so it front-loads (on the two-week curve the projection clears the whole game by run 3).

### Option B: Momentum

One number, earned by lifetime points, speeds the whole economy.
- **Momentum** k = 1 + 0.357 x P^0.215, where P is lifetime points (`TotalEarned`, already saved): 2.0 after a first run to Modern (120 points), about 4 at 20,000. Applied as k-scaling: production x k after the clamp, construction and research / k, storage x k.
- **The shop becomes quality-of-life unlocks** bought with points (spending never lowers Momentum): the legacy kit, Head Start I to III, overflow that pays the plan, extra storage hours for check-in players, a clearer harbinger forecast.
- **The passive retires** into Momentum.
- **Veteran** (20,000 lifetime points, k = 4): about 1.3 days to Modern on the one-week curve, 2.0 on the two-week curve.
- **For:** simplest to build and explain; one curve to tune and test; the +50-200% reset rule falls straight out of it. **Against:** it speeds up ages you have never seen, so new content goes by faster before you earn it (on the two-week curve the projection clears the whole game by run 3; on the one-week curve by run 9); no per-age goals.

### Option C: Era Mastery (recommended)

Each age remembers how often you have completed it.
- **Mastery** m of an age = the number of runs in which you completed it, up to 10. In that age the economy runs at k = 1 + m^0.5 (k-scaling): 2.0 at m = 1, 2.4 at 2, 3.0 at 4, 4.0 at 9.
- **The frontier always runs at 1x.** First visits stay long; known ages go 2 to 4x faster; each run reaches further.
- **Catch-up**: ages six or more behind your record run at k = 4 whatever their mastery, so a veteran's early epochs take hours.
- **Points** by depth (x3 per epoch). Points buy the same quality-of-life shop as Option B.
- **The passive retires** into mastery.
- **Maxed** (mastery 10, k = 4.2): about 1.3 days to Modern on the one-week curve, 1.9 on the two-week curve. Frontier ages still take their full time on a first visit.
- **For:** it is exactly the loop you described, and exactly the research's "old ground 2 to 4x faster, go further"; 22 ages x 10 mastery levels give the achievements catalog 220 natural ladder steps; Succumb can grant mastery; smoke can grade each age against its target / k(m). **Against:** a new saved map and a mastery column in the ages view.

### Projected run times (from the measured levers)

Model: an attentive player; each run lasts a fixed week (primary) or two weeks (alternative) and pushes as far as it can; an age at speed k takes its measured 1x time / k (k-scaling is measured exact within about 5%); Option A uses the fitted two-part model. Points by depth, x3 per epoch.

**One-week curve (primary): 7-day runs** (post-Modern ages measured: 5; the rest at the measured ratio 0.99 of target)

| run | A: days to Modern | B: days to Modern | C: days to Modern | C + catch-up: days to Modern | C: the run ends in | C: days to get back to the last run's furthest age |
|---|---|---|---|---|---|---|
| 1 | 5.3 | 5.3 | 5.3 | 5.3 | Information | - |
| 2 | 3.1 | 2.5 | 2.6 | 2.1 | Cyberpunk | 3.1 |
| 3 | 2.1 | 2.0 | 2.2 | 1.6 | Fusion | 4.4 |
| 4 | 1.8 | 1.9 | 1.9 | 1.5 | Space | 4.9 |
| 5 | 1.6 | 1.7 | 1.8 | 1.3 | Space | 5.2 |
| 6 | 1.5 | 1.6 | 1.6 | 1.3 | Space | 4.7 |
| 7 | 1.5 | 1.5 | 1.5 | 1.3 | Interstellar | 4.3 |
| 8 | 1.5 | 1.4 | 1.5 | 1.3 | Interstellar | 5.2 |
| 9 | 1.5 | 1.4 | 1.4 | 1.3 | Interstellar | 4.8 |
| 10 | 1.5 | 1.3 | 1.3 | 1.3 | Interstellar | 4.5 |

Runs until a run clears the last age: A >10, B 9, C >10, C with catch-up >10.

**Two-week curve (alternative): 14-day runs** (post-Modern ages measured: 5; the rest at the measured ratio 0.95 of target)

| run | A: days to Modern | B: days to Modern | C: days to Modern | C + catch-up: days to Modern | C: the run ends in | C: days to get back to the last run's furthest age |
|---|---|---|---|---|---|---|
| 1 | 7.8 | 7.8 | 7.8 | 7.8 | Cyberpunk | - |
| 2 | 3.9 | 3.4 | 3.9 | 2.7 | Space | 6.8 |
| 3 | 2.4 | 2.5 | 3.2 | 2.0 | Interstellar | 8.4 |
| 4 | 2.1 | 2.3 | 2.9 | 2.0 | Galactic | 8.7 |
| 5 | 2.1 | 2.1 | 2.6 | 2.0 | Quantum | 9.2 |
| 6 | 2.1 | 2.0 | 2.4 | 2.0 | Transcendent | 9.7 |
| 7 | 2.1 | 2.0 | 2.3 | 2.0 | End of game | 10.3 |
| 8 | 2.1 | 1.9 | 2.1 | 2.0 | End of game | 10.8 |
| 9 | 2.1 | 1.9 | 2.0 | 2.0 | End of game | 10.0 |
| 10 | 2.1 | 1.8 | 2.0 | 2.0 | End of game | 9.3 |

Runs until a run clears the last age: A 3, B 3, C 7, C with catch-up 6.


### Storage rule under multipliers (all options)

- **Storage x k.** Multiply every resource's storage by the same prestige factor that multiplies its production. The Storage Covenant's hours then hold at every prestige level.
- **Pick the hours on purpose.** Today's 1.5 hours was sized for a 3-day game. On the slowed curves literal storage already holds longer (income per hour drops). Tripling storage cut 3-hour check-ins from 7.4 to 6.2 days on the one-week curve without speeding active play, so aim for about 4 to 5 hours of typical income, with storage x k on top.
- **Keep flat bonuses flat.** Building and tech storage stays additive; only the prestige factor multiplies the total.

### Smoke enforcement for later runs

Today only cycle 1 is graded. Proposed checks (full tier unless noted):
1. **First prestige for everyone**: active 5 to 6 days, 3-hour check-ins at most 1.3x active, 8-hour at most 1.5x (primary curve). Check the same ratios on the veteran preset, where they drift apart today (item 3).
2. **Cycle 2 curve**: play 2 cycles on 3 seeds. Cycle 2 covers cycle 1's ages at least 1.8x faster and ends at least one age deeper.
3. **Veteran preset**: a test hook sets the veteran state directly (mastery 10 in C, lifetime points in B, maxed tiers in A; the audit's `AUDIT_LEVEL` and `AUDIT_UPGRADES` hooks show the pattern). One cycle on 3 seeds; each age within 0.5 to 2x of its target / k.
4. **Depth pays**: points per day rise with depth (a Digital run beats a Modern run beats a Medieval run), and a second Medieval reset gains under 10% of a Modern run's points.
5. **No long walls**: the longest stretch with no new building type or tech stays under 24 hours for the active player in every age.
6. **Storage under k**: a static check that storage x k holds the covenant at the largest k the design allows.
7. **Milestone feasibility**: replace `TestMilestonesAreFeasible` with a check that computes or plays real affordability under the new curves.
8. **PR tier**: the veteran preset to the Bronze Age (seconds of wall time), so a PR that breaks the prestige layer fails fast.

### Save compatibility

- **Add, never rename.** New fields go on `PrestigeSave` with `omitempty`: `Mastery map[string]int` (C), `ShopVersion int`, `Furthest string` (the deepest age ever entered). Signed saves keep their bytes because old saves simply lack the fields.
- **Keep every upgrade key.** Rewire effects in `config/prestige.go` (effects are not saved; only key and tier are).
- **Refund on version change.** On load, if `ShopVersion` is below the new version, add back the points spent under a frozen copy of today's cost table, zero the rewired tiers, set the version, and log one line ("The prestige shop changed. Your N points were refunded."). B and C refund everything and keep the old keys hidden.
- **Carry progress over, generously.** Old saves have `Level` and `TotalEarned` in old points. B: convert `TotalEarned` x 4.4 (a first Modern run pays 120 new points against 27 old). C: seed the mastery of every age up to Modern at min(10, Level).
- **Clamp `speedMultiplier`** to the new cap on load.

### How each option meets the planned work

| planned work | Option A | Option B | Option C |
|---|---|---|---|
| **Achievements catalog** (lifetime ladders counted in runs) | runs get long (about a week), so count depth, tiers bought and lifetime points; keep run-count ladders short (1, 3, 5, 10, 20) | ladders on Momentum levels and depth | 22 ages x 10 mastery levels is a ready-made ladder family; plus depth and first completions |
| **Catastrophe redesign** (a doom fated per epoch at a random tick; Succumb pays a bonus) | faster later runs shorten epochs in ticks, so draw the strike tick as a share of the epoch's expected length, not fixed ticks | same | same, and mastered epochs are shorter than frontier ones, so the share rule matters most here. Succumb grants +1 mastery to the epoch's three ages |
| Succumb's research bonus | today +25% per epoch is subtracted from research time and hits the 1-tick floor after about four epochs with milestones; make it multiplicative (x0.8 per epoch) | fold it into Momentum | replace it with mastery |
| **Worker Apprenticeship** (veteran workers, one more retune) | apprentices are another production multiplier: put them after the clamp too, or they die in the Victorian Age like the passive; include them in storage x k | count apprentice output inside k so one curve gets retuned | tie it to mastery (workers in a mastered age arrive as veterans); express promotion timers as shares of the age target, not ticks |

All three keep the expected catastrophes per run the same, because the roll is per epoch, not per hour.

## 11. Recommendation

**Option C, Era Mastery, on the one-week curve.**

Why C:
- It is the loop you described and the loop the research describes: old ground 2 to 4x faster, the frontier at 1x, each run reaching further.
- Per-age mastery gives the achievements catalog, Succumb and the Worker Apprenticeship a natural hook, and gives smoke something concrete to grade (each age against its target / k).
- It needs one new saved map; everything else is shared with the other options.
- B is the fallback if you want the smallest build (one number derived from the lifetime points already saved). A keeps the shop players know, but its speedup depends on what each player buys and it front-loads badly with depth points.

Why the one-week curve over the two-week sketch:
- It lands where the research says most players want a full cycle (active 5.3 days, 3-hour check-ins 7.4, or 6.2 with tripled storage).
- The two-week curve has the longest quiet stretches (26 hours in the Atomic Age and 46 in Fusion with nothing new to decide), and at the research's pace its 22 ages run out within 3 to 7 two-week runs (A and B by run 3, C by run 7). On the one-week curve C is still pushing into Interstellar at run 10.
- Keep the sketch as the tuning ceiling: if playtests say a week is still too fast, the same machinery scales to it.

Build order:
1. **Now, small and independent:** fix the `speed` carry-over bug and stop wonders raising the speed cap; fix `TestMilestonesAreFeasible`; remove or rewire the dead `gather_rate` path.
2. **With the catastrophe "randomize the when" work, or right after it:** the one-week curve (`AgeTargets` x2.6 from Bronze, and smoke's copy); build and research normalizers that scale instead of cap; the random-event delay tied to the age target; new idle targets, smoke budgets and a 48 to 72 hour offline cap; the doom strike drawn as a share of the epoch; a standing catastrophe choice for the plan.
3. **The prestige layer:** k-scaling after the clamp, storage x k, depth points, prestige from the Medieval Age, the mastery map, the legacy kit and quality-of-life shop, the save refund, the later-run smoke checks. Ship it with overflow paying the plan and a research queue, or veteran 8-hour players fall to 2.4x the active time.
4. **Achievements:** the mastery ladders.
5. **Worker Apprenticeship**, counted inside k, then the final retune.

Your starting take, tested:
- Run 2 at about 1.6 days and run 5 at about 1 day (today's curve): reachable only with k-scaling (k = 1.4 and 2.3), not with production percentages.
- A veteran at about 15 hours: needs k = 3.6. Production alone can't (x4 gives 25.9 hours); k-scaling can.
- The post-Modern ages as veteran content: yes under C, because the frontier stays at 1x.
- On the one-week curve the same shape gives run 2 about 2.6 days to Modern, and a veteran about 1.3 days.

Two decisions only you can make:
1. **One week or two weeks** to the first prestige.
2. **How fast run 2 covers old ground**: your 25 to 30% (k = 1.4) or the research's 2x (k = 2). The projections use 2x; the machinery is the same either way.


## Appendix A: how to reproduce

**Worktree:** `/Users/echo/code/ageforge-wt/prestige-audit` (detached at origin/master 54c171d). It carries throwaway instrumentation, never committed:
- `game/zz_audit_hooks.go`, plus one-line calls in `game/engine.go` (struct field, `auditScaleRates` before food drain, `auditBuildTicks` at build start, `auditModifiers` in `buildResolver`, `auditStorageMult` on every storage cap), `game/overflow.go` (cap-loss counters) and `game/catastrophe.go` (research speed).
- `smoke/zz_audit.go` (environment-variable driven), with one-line calls in `smoke/run.go`.
- `config/zz_audit_targets.go` (the alternative target tables behind `AUDIT_TARGETS=week|mid|slow`).
- Probes: `smoke/zz_audit_probe_test.go`, `smoke/zz_audit_probe2_test.go`, `cmd/auditprobe/main.go`.

**Knobs** (environment variables read by the smoke runner):

| variable | effect |
|---|---|
| `AUDIT_OUT`, `AUDIT_TAG` | write one JSON line per age to `<AUDIT_OUT>/<tag>-seed<N>.jsonl` |
| `AUDIT_PRODMULT=x` | production x, after the x3 clamp |
| `AUDIT_PRODALL=x` | +x to the production pool, inside the clamp (like today's passive) |
| `AUDIT_GATEMULT=x` | construction times / x (x below 1 lengthens them); research via research speed + (1 - 1/x), shortening only |
| `AUDIT_STORAGEMULT=x` | every storage cap x |
| `AUDIT_SPEED=x` | game speed set directly (tick speed stand-in) |
| `AUDIT_SPEEDMAX=1` | the bot keeps `speed` at the wonder cap |
| `AUDIT_BUY=cheapest\|speedfirst\|none` | prestige shop policy |
| `AUDIT_LEVEL=L`, `AUDIT_UPGRADES=all5` or `key:tier,...` | start at prestige level L with those tiers, through a real prestige |
| `AUDIT_COUNT=1` | count production lost at full storage |
| `AUDIT_TARGETS=week\|mid\|slow` | swap AgeTargets (and smoke's copy): today x2.6 from Bronze, today x2 from Bronze, or the two-week sketch |
| (always on) | new-decision tracking: the longest stretch per age with no new building type or tech (`quiet_secs` in the JSON lines) |

**Commands** (run from the worktree; `B=go run ./cmd/smoke`, or a built binary):

```
# runs 1-10, bot policy and speed-first policy
AUDIT_OUT=out AUDIT_TAG=p1 AUDIT_BUY=cheapest  $B -tier full -scenario progression -seeds 1 -seed-base 1 -cycles 10 -final-age "" -max-sim 1200h -parallel 1 -out out/p1
AUDIT_OUT=out AUDIT_TAG=p2 AUDIT_BUY=speedfirst $B ...same flags... -out out/p2
# one first run to Modern with a lever (repeat for 1.25, 1.5, 2, 4)
AUDIT_OUT=out AUDIT_TAG=mult2 AUDIT_PRODMULT=2 $B -tier full -scenario progression -seeds 1 -seed-base 1 -cycles 1 -final-age "" -max-sim 400h -parallel 1 -out out/mult2
AUDIT_SPEED=2 ... ; AUDIT_GATEMULT=2 ... ; AUDIT_PRODMULT=2 AUDIT_GATEMULT=2 ... ; AUDIT_PRODALL=0.5 ... ; AUDIT_SPEEDMAX=1 ...
# veteran and controls
AUDIT_LEVEL=38 AUDIT_UPGRADES=all5 ... ; AUDIT_LEVEL=38 AUDIT_UPGRADES=none:0 ... ; AUDIT_LEVEL=1 AUDIT_UPGRADES=all5 ...
AUDIT_LEVEL=1 AUDIT_UPGRADES=storage_bonus:5,population_cap:5,starting_food:5,starting_wood:5 ... ; AUDIT_LEVEL=1 AUDIT_UPGRADES=none:0 ...
# pushing past Modern
$B -tier full -scenario progression -seeds 1 -seed-base 1 -cycles 1 -final-age "" -prestige-age information_age -max-sim 800h ...   (also digital_age, space_age, transcendent_age with -max-sim 2000h)
# check-in players with cap-loss counting (repeat for 1h, 3h; and 8h with AUDIT_PRODMULT=1.5, 2, 4)
AUDIT_OUT=out AUDIT_TAG=idle8h AUDIT_COUNT=1 $B -tier full -scenario styles -style idle -check-in 8h -seeds 1 -seed-base 1 -max-sim 900h -parallel 1 -out out/idle8h
# one-week curve: AUDIT_TARGETS=week (add AUDIT_STORAGEMULT=3, AUDIT_GATEMULT=0.5, or AUDIT_PRODMULT=2 AUDIT_GATEMULT=2 for the gap tests)
# slowed curve (add AUDIT_PRODMULT / AUDIT_GATEMULT for the lever runs; -scenario styles -style idle for check-ins)
AUDIT_TARGETS=slow AUDIT_OUT=out AUDIT_TAG=slow $B -tier full -scenario progression -seeds 1 -seed-base 1 -cycles 1 -final-age "" -max-sim 2000h -parallel 1 -out out/slow
AUDIT_TARGETS=slow ... -prestige-age space_age -max-sim 4000h   (post-Modern times)
# probes: perks per age, speed carry-over, gate rows, storage covenant, storage cost
AUDIT_PROBE=1 go test ./smoke -run 'TestAuditProbe' -v -count=1 -timeout 60m
```

Scripts used to read the JSON lines (in the session scratchpad, `sims/`): `analyze.py` (per-cycle totals), `tables.py`, `loss.py` (cap losses), `project.py` (multi-run projections from the measured levers). Each seed-1 run to Modern takes about 1 minute of wall time on this Mac; 10 cycles about 10 minutes; a slowed run about 4 minutes.

## Appendix B: timer survey file references

Producer payback `config/pacing.go:113`; build and research caps `pacing.go:75, 78, 272, 325`; random event delay `game/events.go:47-48`; event data `config/events.go:19-21`; war raids `game/diplomacy.go:697` and `:33`; drift `:20`; deal refresh `game/deals.go:66`; trade routes `config/trade.go:132-296`; epoch roll `game/engine.go:1880`, chances `game/catastrophe.go:34-41`; Endure debuff `catastrophe.go:48`; harbinger thread `game/harbinger.go:221-300`; expeditions `game/military.go:82-220`, auto-expedition `game/auto_expedition.go:73`; boons `boon/catalog.go:80-162`; festival and black market `game/engine.go:51-64`; chain boosts `config/milestones.go:70-135`; survivor milestones `milestones.go:818-841`; offline cap `game/engine.go:3983-3984`; plan funding `game/plan.go:544`; smoke targets `smoke/targets.go:16, 51-53`; idle targets `smoke/idle_targets.go:19-31`; budgets `smoke/run.go:103-111`, `smoke/scenario_bot.go:117-125`, `smoke/scenario_styles.go:92-99`, `smoke/scenario_prestige.go:31`.
