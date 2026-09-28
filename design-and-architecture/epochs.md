# AgeForge — Epoch System

## Overview

**Epochs** are the meta-progression layer above ages. The 21 ages group into **7 epochs of 3 ages each**.
Epoch transitions are bigger civilizational milestones than age advances — they signal a fundamental
shift in what resources matter, what events can occur, and what dangers your civilization faces.

```
Ages 1–3   → Stone Era
Ages 4–6   → Iron Era
Ages 7–9   → Steel Era
Ages 10–12 → Electric Era
Ages 13–15 → Digital Era
Ages 16–18 → Neon Era
Ages 19–21 → Cosmic Era
```

Every 3rd age advance = 1 epoch transition. 7 epochs × 3 ages = 21 ages exactly.

---

## The 7 Epochs

| # | Epoch | Ages | Dominant Resources | Era Feel |
|---|-------|------|--------------------|---------|
| 1 | **Stone Era** | Primitive, Stone, Bronze | wood, stone | Dawn of civilization |
| 2 | **Iron Era** | Iron, Classical, Medieval | iron, marble | Classical empires, plague, philosophy |
| 3 | **Steel Era** | Renaissance, Colonial, Industrial | steel, coal | Industry, empire, war |
| 4 | **Electric Era** | Victorian, Electric, Atomic | electricity, oil, uranium | Power age, nuclear dawn |
| 5 | **Digital Era** | Modern, Information, Digital | titanium, data | Information revolution |
| 6 | **Neon Era** | Cyberpunk, Fusion, Space | plasma, dark_matter | Post-human, megacorp wars, space |
| 7 | **Cosmic Era** | Interstellar, Galactic, Quantum | antimatter, quantum_flux | Galactic civilization, reality manipulation |

### What Changes at an Epoch Transition

1. **Primary building costs shift** — buildings in the new epoch are denominated in new resources
2. **Extraction lineages switch output** — Organic and Geological lineages produce new resources
3. **Metallurgy's processing chain advances** — new ore → new refined metal
4. **Event pool changes** — epoch-exclusive events replace previous epoch's exclusive events
5. **Catastrophe may strike** — from the Iron Era on, a bad transition roll can escalate into the Civilizational Catastrophe modal (see below)
6. **UI epoch badge updates** — the epoch badge near the age indicator changes color and icon
7. **New worker domains may unlock** — Hacker (Digital Era), Astronaut (Neon Era)

---

## Epoch Event System

At each epoch transition, the game rolls a **major epoch event** — a single significant event
that could be a boon or a disaster. Catastrophe is **not guaranteed** — it is one possible bad
outcome among several, 12–18% per transition depending on faith, and never before the Iron Era.

### The Roll

```
Roll at each epoch transition:

  Faith < 25% cap:   40% Good / 60% Bad
  Faith 25–75% cap:  50% Good / 50% Bad   ← baseline
  Faith > 75% cap:   60% Good / 40% Bad

  If GOOD → roll from Good Event pool (weighted by Culture level):
      Low culture:    Minor Good only
      Medium culture: Minor or Major Good
      High culture:   Minor, Major, or Legendary Good (rare)

  If BAD → roll from Bad pool (weighted):
      70% → Challenging Event (severe but recoverable — 8 events)
      30% → Civilizational Catastrophe (the Endure / Succumb modal)
```

**Catastrophe probability per transition:** (1 − good chance) × 30% = **18% / 15% / 12%** at
low / mid / high faith. A run has 6 transitions (Iron through Cosmic; every run starts in the Stone
Era), so expected catastrophes are about 0.9 per run at mid faith, with at most 6 (one per epoch).

**Faith matters:** Keeping faith above 75% of its cap improves your odds of a good epoch event
at every transition. This is a new strategic reason to invest in Faith buildings and workers
beyond its existing morale/cohesion/diplomacy uses.

**Culture matters:** Culture level gates the tier of good events you can receive. High-culture
civilizations occasionally receive Legendary Good events — powerful outcomes unavailable at
lower culture levels.

### Good Epoch Events

**Minor Good — available at any culture level:**

1. **Age of Plenty** — all resource production ×2 for 72 real hours
2. **Population Surge** — all worker classes +15% count, immediately recruited
3. **Ancient Cache** — fills 40% of every resource's storage cap with current resources
4. **Trade Winds** — gold ×3 for 48h; all trade routes open regardless of requirements
5. **Cultural Festival** — culture +30%, faith +20% instantly; morale bonus for 48h

**Major Good — requires medium culture:**

6. **The Grand Discovery** — 3 free techs from current epoch's research tree
7. **Worker Innovation** — all worker output multipliers permanently +10% (stacks across events)
8. **The Architect's Gift** — 10 free buildings of any current-age type, instant, no resource cost
9. **Peaceful Century** — all negative random events suspended for 96h; all production +20%

**Legendary Good — requires high culture, rare:**

10. **Epoch Blessing** — permanent +15% production for current epoch's primary resource;
    one unique epoch wonder unlocked (exclusive to this event, not in the normal wonder list);
    recorded in civilization history as a golden age entry

### Challenging Bad Epoch Events (70% of bad outcomes)

Severe but recoverable — no reset, no modal. Applied immediately on epoch transition.

1. **The Famine** — food production -60% for 120 ticks; workers begin leaving if not corrected
2. **Merchant Betrayal** — gold -50%; all trade routes suspended for 72 ticks
3. **The Great Fire** — 8 random buildings destroyed; no targeted penalty on surrounding buildings
4. **Epidemic** — worker count -20%; food drain +15% for 96h; faith influences severity
5. **Resource Drought** — current epoch's primary output resource -70% for 90 ticks
6. **Political Instability** — faith -60%; military output -40%; knowledge production paused 60 ticks
7. **Economic Crash** — all gold halved; building costs +50% for 72h
8. **The Dark Age** — knowledge production and all research paused for 48h; one random tech gains a knowledge debt that must be cleared before it can be used

### Civilizational Catastrophe (30% of bad outcomes, 12–18% per transition)

This section describes what is built (updated 2026-09-26). Ideas from the original design that
were never built are kept under **Future ideas (not implemented)** at the end of it.

**Rules**

- **Iron-epoch gate.** No catastrophe before the epoch containing the Iron Age
  (`config.CatastropheGateEpoch = "iron_era"`). The Stone Era never has one.
  Good and challenging epoch events are unaffected.
- **One per epoch per run.** Each epoch's transition rolls once per run (`epochEventFired`), reset
  by Succumb, prestige and a new game.
- **No direct player trigger.** `catastrophe invoke` was removed (2026-09-26). Choosing to face
  a catastrophe is the Harbinger's **Invite** (see [Harbinger](#harbinger)), which arms
  `catastropheInvited`: the next transition into an allowed epoch produces a catastrophe instead
  of rolling (consumed when honoured, kept while gated or something is pending, reset by
  Succumb/prestige, saved as `catastrophe_invited`), and `CatastropheOutlook` reports probability
  1 while it is armed. The dev console's `/catastrophe` forces one for testing and respects the
  gate.
- **Pending blocks progress.** A catastrophe only sets `pendingCatastrophe`; nothing is destroyed
  until the player chooses. The game keeps running, but `AdvanceAge` and `DoPrestige` refuse while
  one is pending, and the roll never overwrites a pending catastrophe.
- **No Defer.** The modal has two choices. Esc closes it without choosing; a status-bar badge
  shows the pending catastrophe and the bare `catastrophe` command reopens the modal. A save with a
  pending catastrophe shows the modal again on load.
- **Seeded.** Every roll (epoch event, destroyed buildings, ruins) comes from the run's seeded
  `GameEngine.rng`, drawing from pools built in sorted-key order.
- **Outlook.** `GameEngine.CatastropheOutlook()` (also `GameState.CatastropheOutlook`) reports the
  next transition's epoch, whether a catastrophe is possible there, its probability, a coarse tier
  (none/low/medium/high: under 14%, under 17%, 17% and up) and the faith fill driving it. The
  probability includes the Harbinger's Appease multiplier. It is read-only and lock-safe; the
  Harbinger reads it when a thread starts, at each handoff and for its panel. `Passage` says what
  the next passage is: `"epoch"` (`PassageEpoch`) or, in the final epoch, `"prestige"`
  (`PassagePrestige`, with `NextEpochKey` empty). See [The Last Passage](#the-last-passage).

**The modal** is a box floating over the dashboard, sized to its content:

```
╔════════════════════ ☄ Iron Era Catastrophe ════════════════════╗
║                    ☄ The Great Plague                            ║
║     A devastating plague sweeps your cities. The streets…        ║
║                                                                  ║
║ ── ENDURE — weather the catastrophe ──                           ║
║   • 20% of buildings destroyed (wonders are spared)              ║
║   • All resources reduced to 15%                                 ║
║   • 25% of workers lost; workers of destroyed buildings go idle  ║
║   • Production -10% for 216 ticks, morale -10                    ║
║   ✓ Age, research, wonders and prestige preserved                ║
║   ✓ Survived marker on the epoch badge                           ║
║                                                                  ║
║ ── SUCCUMB — let civilization fall ──                            ║
║   • Full reset to the Primitive Age: buildings, resources, …     ║
║   • No prestige points earned (level and upgrades are kept)      ║
║   ✓ Up to 8 buildings become ruins (50% output, max 24 ruins)    ║
║   ✓ Ancient Knowledge: research speed +25% (total +25%, …)       ║
║   ✓ Iron Era legacy: iron +20% (permanent)                       ║
║                                                                  ║
║              [E] ENDURE        [S] SUCCUMB                       ║
║   Esc: decide later · advancing waits · type 'catastrophe' …     ║
╚══════════════════════════════════════════════════════════════════╝
```

### ENDURE — Consequences

- `floor(non-wonder buildings / 5)` destroyed, at least 1 if any. Wonders are neither destroyed
  nor counted. With a Harbinger Brace (`pendingBraceLevel` 1 / 2) it is 15% / 10%, same floor.
- Workers assigned to destroyed buildings return to the idle pool (same rule as selling).
- All unlocked resources drop to 15% of their stored amounts (30% / 45% braced).
- 25% of the single worker pool is lost; every building's assignment shrinks by the same share,
  whatever its worker domain.
- Reconstruction Effort: `production_all` −10% for 216 ticks. Morale −0.10.
- "Survived" marker on the epoch badge and a civilization-log entry.
- Research, wonders, age and prestige are untouched.

### SUCCUMB — Consequences

- Up to 8 non-wonder buildings become ruins (50% base output, no workers). Ruins persist across
  Succumb and prestige, capped at **24**: past the cap the lowest-value ruins (earliest
  `RequiredAge`, then lowest base output, then key) are dropped first. The cap also trims old saves
  on load.
- The epoch's legacy flag is set. Its per-resource bonus goes into `permanentBonuses` via
  `reapplyLegacyBonuses` after every reset.
- **Ancient Knowledge:** +25% `research_speed` per distinct legacy epoch. Derived from the flags,
  never stored, emitted into the resolver as source `legacy` (shows in Active Multipliers). It
  survives save/load, Succumb and prestige by construction. Saves from before this change stored
  +25% in `permanentBonuses`; it is stripped on load (`succumb_research_derived` marks new saves).
- Full reset to the Primitive Age. Prestige level, points and upgrades are kept; no points earned.

### The 7 Catastrophes

| Epoch | Catastrophe Name | Flavor Text |
|-------|----------------|-------------|
| Stone Era | **The Great Meteor** | Defined in config but unreachable (Iron-epoch gate). |
| Iron Era | **The Great Plague** | A devastating plague sweeps your cities. The streets are silent. |
| Steel Era | **The World War** | Industrial warfare tears civilization apart. The factories are ash. |
| Electric Era | **The Nuclear Exchange** | Nations unleash the atom. Cities become glass. |
| Digital Era | **The Great Hack** | Every system falls silent. The AIs turn on their creators. |
| Neon Era | **Corporate Armageddon** | The megacorps end the world with a fusion bomb. |
| Cosmic Era | **The Reality Tear** | Exotic matter destabilizes spacetime. Reality cracks open. |

### SUCCUMB Legacy Bonuses by Epoch (built)

| Epoch | Legacy bonus |
|-------|--------------|
| Stone Era | wood +20%, stone +20% (kept by saves that earned it; no longer reachable) |
| Iron Era | iron +20% |
| Steel Era | steel +25%, coal +25% |
| Electric Era | electricity +25%, uranium +25% |
| Digital Era | data +30%, titanium_ore +30% |
| Neon Era | plasma +30%, dark_matter_crystals +30% |
| Cosmic Era | dark_matter +35% |

### Catastrophe vs Regular Prestige

| | Regular Prestige | Catastrophe Succumb |
|--|-----------------|---------------------|
| Trigger | Player-initiated from the Modern Age | Random (12–18% per transition, Iron Era on; lowered by Appease), or guaranteed by the Harbinger's Invite |
| Reset scope | Full | Full; up to 8 new ruins (24 max) carry forward |
| Bonus pool | Prestige upgrade tree + points | Epoch legacy bonus + Ancient Knowledge |
| Repeatable | Yes | Once per epoch per run; bonuses once per epoch ever |
| Blocked by a pending catastrophe | Yes | n/a |
| Can bring a catastrophe | From the Cosmic Era: the Last Passage | n/a |

### Future ideas (not implemented)

Kept from the original design for reference. None of this exists in the game.

- **Endure lasting consequences:** building costs +20% and worker food drain +10% for 72 real
  hours; random events 20% harder during recovery.
- **Endure permanent rewards:** a Reconstruction tech branch (5 epoch-specific recovery techs); a
  **Monument to the Fallen** wonder (free, +2,000 culture, +500 faith/tick, morale bonus); titles
  such as "The Undying Iron Lords" shown in the Stats tab.
- **Succumb extras:** Ancient Knowledge scoped to the fallen epoch's tech tree; a Catastrophe Title;
  **Faster Return** (techs from 2 epochs below the catastrophe auto-complete next run); an
  **Exclusive Starting Event** in the first 10 ticks of the next run; ruins marked ☒ in the
  Economy tab.
- **Exclusive uniques per catastrophe:** Meteor Fragment wonder; Ancient Immunity passive (events
  15% less severe) and a Plague Doctor worker class; War Doctrine tech and Armistice Monument;
  Fallout Shelter tech and Nuclear Vault; Ghost Protocol tech and Dead Drop Network; Phoenix
  Protocol tech and Corporate Ruins wonder; Reality Anchor tech and Scar in Reality wonder
  (the Reality Tear legacy was also meant to boost antimatter + quantum_flux instead of dark_matter).

---

## Harbinger

Built 2026-09-26; reworked the same day from a single last-age visit into epoch-long threads.
Code: `game/harbinger.go` (mechanic), `config/harbingers.go` (the 22-entry roster: `Name`,
`Description`, `AppeaseLabel`, `BraceLabel`, `InviteLabel`, plus the derived `ForecastPrecision`
and `FalseProphetChance`), `flavor/` (arrival, warning, action and outcome lines). Player docs:
`site/docs/harbinger.md`.

### Threads, start and handoff

A **thread** belongs to an epoch whose passage can roll a catastrophe: its outgoing transition
(Stone through Neon), or for the Cosmic Era, prestige (see [The Last Passage](#the-last-passage);
the thread's `TargetEpoch` is `""`). It lasts from the epoch's first age until that passage. The
speaker is always the current age's roster figure, so all 22 figures appear.

Three hooks start or advance a thread, all under the write lock:

- `harbingerOnAgeAdvance()`, at the end of `advanceAge` (after `detectEpochTransition` and
  `fireAwakening`). If a thread is live for the current epoch and its `Age` differs from the new
  age, `harbingerHandoff()` passes it to the new figure. Otherwise `maybeHarbingerArrive()` tries
  to start the new epoch's thread (the old one was already resolved by the transition).
- `harbingerTickCheck()`, near the top of `doTick`. Starts a thread on the first tick of an epoch
  that has none: a new game, Succumb, prestige or reset, none of which advance an age. The
  unpersisted `harbingerCheckedEpoch` limits it to one outlook check per epoch. This is how the
  Wild Man greets a new game.
- `restoreHarbingerState()`, after a load, ends with `harbingerOnAgeAdvance()`: a save in any age
  of a qualifying epoch without a thread gets one there (with that age's figure), and a saved
  thread whose `Age` lags the loaded age is handed off.

`maybeHarbingerArrive` requires no live thread and `harbingerArrived[currentEpoch]` unset;
`harbingerArrive` then requires `catastropheOutlook().Possible` (next epoch past the Iron gate and
not yet rolled this run). Starting a thread sets `harbingerArrived`, so it is once per epoch per
run.

A **handoff** appends the new age to `HarbingerSave.Chain`, sets `Age`, re-derives
`AnnouncedTier` from `harbingerDisplay()` and draws a fresh arrival and warning line in the new
voice. Levels, the invite, `FalseProphet` and `ClaimFactor` stay with the thread. Each start or
handoff logs, publishes `EventHarbingerArrived` (payload `handoff` true on a handoff, which the
toast renders as "takes up the warning") and keeps the status-bar badge up while
`GameState.Harbinger` is non-nil. `HarbingerView.Earlier` lists the earlier figures for the
panel's "Took up the warning from ..." line. Nothing expires.

The dev console's `/harbinger` (`summonHarbinger`) starts a thread with the current figure,
skipping the once-per-epoch rule but still needing `Possible`.

### False prophets and ClaimFactor

- The thread rolls once, at its first figure, against that age's `FalseProphetChance`
  (`(8 - ageIndex) / 64` before the Industrial Age, 0 from it). Among first ages that is
  Primitive 8/64, Iron 5/64, Renaissance 2/64, and 0 for Victorian, Modern and Cyberpunk. A thread
  started mid-epoch by a load uses that age's chance.
- A false thread claims medium or high (rolled). The claim is stored as
  `ClaimFactor = claimBase[tier] / realChance` at the start, with `claimBase` 0.15 for medium and
  0.18 for high (the real mid- and low-faith chances). `harbingerDisplay()` then reports
  `real × ClaimFactor` (capped at 1, tier never below low), so Appease, a faith-band change or an
  Invite move the false claim exactly as they move a true one, and every figure repeats it.
- `ForecastPrecision` is per current figure: numeric from the Industrial Age, where the panel
  prints `HarbingerView.Probability`. For a false Steel Era thread that reaches the Newsboy this
  is the claimed figure. `HarbingerView` carries no false-prophet flag, and the UI's
  `catastrophe` outlook and Epoch tab show the thread's tier and figure while one is live.
- Old saves with `FalseProphet` but no `claim_factor` load with `ClaimFactor = 1`.

### Actions and passage-based costs

Answers belong to the passage and carry across handoffs. Costs are pure functions of the thread's
epoch (`harbingerAppeaseCost(epochKey, level)`, `harbingerBraceCost(epochKey, level)`), rounded
up, level 2 at double, so the price is the same in every age of the epoch. Pricing off current
caps would make the epoch's first age (smallest caps) a discount.

- `harbingerAppeaseAges(epochKey)`: the epoch's ages, minus the game's last age.
- `harbingerHeldSinceStart(epochKey)`: resources unlocked (cumulative `UnlockResources`) by the
  epoch's first age, so every price is payable in every age of the epoch.
- `harbingerAdvanceAges(epochKey)`: the epoch's later ages plus the next epoch's first age.

| Action | Cost (level 1) | Effect | Cap |
|--------|----------------|--------|-----|
| Appease | `harbingerAppeaseIncomeShare` (1/4) of `config.FlowIncome` summed over `harbingerAppeaseAges` at their `AgeTargetTicks`, in faith, and in culture if culture is held since the start (Steel Era on); rounded up to 2 significant figures | Multiplies the real catastrophe chance by `harbingerAppeaseFactor` = 0.6 per level (0.36 at 2) | 2 |
| Brace | 12% of `harbingerBraceBasis`: per resource held since the start, minus faith and culture, the largest `ResourceReqs` across `harbingerAdvanceAges` | Endure destroys 15% / 10% of non-wonder buildings and keeps 30% / 45% of resources (unbraced 20% / 15%) | 2 |
| Invite | free | Sets `HarbingerSave.Invited` and arms `catastropheInvited`; Appease refuses afterwards, Brace does not | once |

Level-1 prices from the current config:

| Thread | Appease | Brace |
|--------|---------|-------|
| Stone | 59 faith | 9,600 food, 4,800 wood, 2,400 knowledge |
| Iron | 5,400 faith | 26,400 knowledge, 26,400 stone, 6,360 iron, 21,600 gold |
| Steel | 74K faith, 770K culture | 3.6M knowledge, 1.8M gold, 288K steel |
| Electric | 1.2M faith, 16M culture | 56.4M steel, 924K oil, 3.96M electricity |
| Digital | 12M faith, 180M culture | 156M gold, 117.6B electricity, 19.2B data |
| Neon | 130M faith, 2B culture | 288B electricity, 46.8B data, 3B crypto |

Appease follows income, not storage (2026-09-27): faith is a flow resource at hand-set rates
with no market, so the old 15%-of-passage-storage price was out of reach in most threads once
the pacing rebalance shortened the ages. `TestAppeasePayableWithinThread` checks that the
modelled income reaches level 1 before the thread ends and levels 1 and 2 by the passage. When storage cannot yet hold a price, `shortfall` adds "(your X storage must
reach N first)" to the refusal.

- **Appease** is applied through `harbingerAppeaseMultiplier()`, which scales
  `catastropheChanceOnBadRoll` in both `rollEpochEvent` (the real roll) and `catastropheOutlook`
  (the displayed odds), so the two cannot disagree. It lowers the real odds even for a false
  thread. Worst case the faith spend drops the fill from the top band to the bottom (12% to 18%,
  x1.5), and x0.6 still leaves 0.9x, so each level always lowers the odds.
- **Brace** lives on `HarbingerSave.BraceLevel` until resolution, then moves to
  `ge.pendingBraceLevel` if the catastrophe came. `Endure` reads and clears it
  (`braceDestroyPct`, `braceKeepFrac`). Succumb ignores it. Keeping it on the pending catastrophe
  means Esc-then-Endure and save/load both keep the discount.

### Resolution

`detectEpochTransition` calls `resolveHarbinger(newEpoch, came)` right after `rollEpochEvent`,
where `came` is "no catastrophe was pending before the roll and one is pending for this epoch
now". The verdict is spoken in the last figure's voice. Outcomes (`HarbingerRecord.Outcome`):

| Outcome | Condition |
|---------|-----------|
| `fulfilled` | came and invited |
| `vindicated` | came, not invited (a false thread here gets a note that the warning was invented) |
| `spared` | did not come, true thread |
| `discredited` | did not come, false thread |

It logs the verdict (and the braced Endure numbers if relevant), draws one flavor line, appends a
`HarbingerRecord` (with `Age`/`Name` of the last figure and the full `Chain`) to
`harbingerHistory`, and clears the live thread. The Epoch overlay lists each record's chain of
figures. The same `advanceAge` call then runs `harbingerOnAgeAdvance()`, which may start the new
epoch's thread.

### Determinism

All draws come from the seeded `ge.rng` under the write lock. A thread start draws, in order: the
false-prophet `Float64()` (always, whatever the age's chance, so the stream's shape does not
depend on it), then `Intn(2)` for the claimed tier only if the thread is false, then the arrival
and warning lines from the engine's flavor `Stream`. A handoff draws two lines; each action and
the resolution draw one.

### Persistence and resets

`GameSave` fields, all `omitempty` (old saves load clean):

```go
Harbinger          *HarbingerSave    `json:"harbinger,omitempty"`
HarbingerArrived   map[string]bool   `json:"harbinger_arrived,omitempty"`
CatastropheInvited bool              `json:"catastrophe_invited,omitempty"`
PendingBraceLevel  int               `json:"pending_brace_level,omitempty"`
HarbingerHistory   []HarbingerRecord `json:"harbinger_history,omitempty"`
```

`HarbingerSave` gained `chain` (ages that have spoken, first to current) and `claim_factor`
(false threads only); `HarbingerRecord` gained `chain`. All `omitempty`.

On load, levels are clamped to 2, an empty `Chain` becomes `[Age]`, an unknown roster age drops
the thread, and `pendingBraceLevel` is zeroed unless a catastrophe is pending.
`clearHarbingerRun` (live thread, arrivals, `harbingerCheckedEpoch`, invite, pending Brace) runs
on Succumb, prestige and reset, and the next tick starts the Stone Era thread.
`harbingerHistory` follows `epochEventHistory`: kept by Succumb, cleared by prestige.

---

## The Last Passage

Built 2026-09-26. Code: `game/last_passage.go`, with the Cosmic thread's pricing in
`game/harbinger.go`. Player docs: `site/docs/prestige.md#the-last-passage`.

### Prestige is the Cosmic Era's passage

Every other epoch's passage is its transition into the next epoch. The Cosmic Era has no next
epoch, so its passage is prestige. `CatastropheOutlook.Passage` is `"prestige"` there and
`NextEpochKey` is empty; `Possible` is false once the Last Passage is pending. The Cosmic
Era gets a harbinger thread like any other, starting at the Interstellar Age, with the four
cosmic figures handing off per age (Distress Beacon, Elder Relay, your future self, your unmade
self). Its figures are all past the Industrial Age, so no false prophets and numeric odds.

### The roll

`DoPrestige`, once confirmed, calls `rollLastPassage()` when `lastPassageApplies()` (final epoch,
past the Iron gate). One `ge.rng` `Float64()` is always drawn, so the stream's shape does not
depend on the odds or on an invite. The chance is `catastropheOutlook().Probability`, the same
formula as `rollEpochEvent`: `(1 - good chance) × 0.30 × harbingerAppeaseMultiplier()`, i.e.
18% / 15% / 12% by faith band, certain when `catastropheInvited` is armed (consumed on a hit).
Prestige from any earlier epoch never rolls.

- **Miss:** verdict `spared`, the RunEnding line, and prestige completes.
- **Hit:** `triggerLastPassage` sets `pendingLastPassage` and publishes a bus event for the toast.
  Prestige does not complete.

### Pending state

`pendingLastPassage` blocks only `DoPrestige` (`lastPassageBlockErr`); `AdvanceAge`, building
and everything else carry on. The dashboard shows the choice in the catastrophe modal
(`✦ The Last Passage`), Esc hides it, the status bar carries a `☄ LAST PASSAGE` badge and the bare
`catastrophe` command reopens it. The save list reports it as the pending choice. The dev
console's `/lastpassage` (`forceLastPassage`) sets it for testing, final epoch only.

### Endure and Succumb

Both go through `resolveLastPassage` and complete the prestige; the level rises either way.

| Choice | Points from the run | Other effect |
|--------|---------------------|--------------|
| Endure | `floor(full × keep)`, minimum 0, where `keep` = `lastPassageKeepFrac[brace]` = 0.50 / 0.70 / 0.85 at Brace 0 / 1 / 2 | Verdict `vindicated` (`fulfilled` if invited) |
| Succumb | 0 | Sets `cosmicLegacy` |

Brace in the Cosmic Era changes only the points share; nothing survives prestige to destroy.
`recordLastPassageOutcome` appends a `catastropheHistory` entry carrying the Endured / Succumbed
markers, so the Stats tallies count it.

**Cosmic Legacy.** A one-time permanent flag. `cosmicLegacyModifiers()` emits
`production_all +CosmicLegacyProductionBonus` (0.10) into the resolver as source `cosmic_legacy`,
derived from the flag and never stored as a bonus value, like the derived Succumb research bonus
(`legacy`). It survives every prestige and Succumb; only a full wipe clears it. With the flag
held, the modal disables Succumb ("You already carry the Cosmic Legacy. Succumb is closed to
you."), so Endure is the only choice.

**Invite** in the Cosmic Era arms `catastropheInvited` for the next prestige. It is the deliberate
path to the Cosmic Legacy; Appease is refused afterwards, Brace still raises the Endure share.

### The run's last lines

The prestige reset clears the log. `runEndingLines` writes the verdict, the Endure / Succumb
lines and one RunEnding flavor line aside (drawn in that order from `ge.rng`) and the new run's
log starts with them. `logRunEnding` runs on every prestige, from any age: the line is in the
voice of the age the run ended in, with the age's harbinger as the subject in the Cosmic Era.

### Costs

Fixed across the four Cosmic ages, level 2 at double:

| Action | Level 1 | Basis |
|--------|---------|-------|
| Appease | 1.2B faith + 19B culture | a quarter of `config.FlowIncome` over the Interstellar, Galactic and Quantum Ages at their targets |
| Brace | 1.56T dark matter + 75.6B titanium | 12% of the most the Cosmic Era's own advances ask of each resource held since the Interstellar Age (dark matter 13T for Quantum, titanium 630B for Galactic) |

Appease counts every Cosmic age but the Transcendent, which has no advance to pace
(`harbingerAppeaseAges`). A player who prestiges on arriving leaves before Appease is in reach;
one who stays for the epoch gets the same timing as every other thread. Antimatter and quantum flux unlock after the Interstellar Age, so the held-since-start rule
excludes them from Brace.

### Persistence

`GameSave` fields, both `omitempty`:

```go
PendingLastPassage bool `json:"pending_last_passage,omitempty"`
CosmicLegacy       bool `json:"cosmic_legacy,omitempty"`
```

### Tuning constants

- `game/last_passage.go`: `LastPassageKeep` (0.50, unbraced Endure share),
  `lastPassageKeepFrac` (0.50 / 0.70 / 0.85 by Brace), `CosmicLegacyProductionBonus` (0.10).
- `game/harbinger.go`: `harbingerAppeaseAges` (the Cosmic ages Appease is priced over).

---

## Epoch Event Pools

The existing universal events (drought, good harvest, plague, festival, trade windfall, etc.) remain
active in all epochs and scale their magnitude to current epoch resource production rates. Each epoch
adds 5 exclusive events that only appear during that epoch.

### Stone Era — Exclusive Events

1. **Sacred Grove Discovered** — an ancient forest is found; knowledge production +50% for 48 ticks
2. **Wandering Tribe** — nomad group joins; +80 pop, +6 workers across primitive classes
3. **Stone Idol** — workers uncover a carved idol; faith +300, morale +25% for 60 ticks
4. **Cave Paintings** — ancient art discovered; culture +500, +1 culture/tick permanently from next
   culture building built
5. **Bone Tools** — innovation event; wood production +100% for 30 ticks; one free early tech

### Iron Era — Exclusive Events

1. **The Spreading Plague** — population -15%, workers -10%; faith above 60% cap halves the losses
2. **Barbarian Horde** — military buildings take 40% damage unless military production > threshold
3. **Silk Road Opens** — all trade route gold income +80% for 90 ticks
4. **Philosopher's Academy** — knowledge production doubled for 60 ticks; one free knowledge tech
5. **Bronze Uprising** — production halved for 36 ticks unless faith > 40% cap

### Steel Era — Exclusive Events

1. **Industrial Accident** — 5 random Engineering buildings destroyed; surrounding output -20% for 48 ticks
2. **Worker Strike** — factory output halved until faith restored to >50% or 72 ticks elapse
3. **Colonial Gold Rush** — gold production ×3 for 60 ticks; 10 free Settler workers added
4. **Railroad Connection** — all trade route income +50% permanently until next epoch transition
5. **Colonial Revolt** — 20% of Colonial-era buildings damaged; gold income -30% for 60 ticks

### Electric Era — Exclusive Events

1. **Nuclear Test Fallout** — food production -30% for 120 ticks; faith -20% (public fear)
2. **Oil Crisis** — all electricity-dependent buildings offline for 36 ticks; oil building costs ×2 for 60 ticks
3. **Space Race Ignition** — knowledge +150% for 90 ticks; one free tech in knowledge tree
4. **Cold War Tension** — military production +60%, worker food drain +20% for 120 ticks (war footing)
5. **Power Grid Failure** — all Electric Era buildings offline for 18 ticks, then +50% electricity output
   on restoration (systems surge)

### Digital Era — Exclusive Events

1. **The Great Data Breach** — data -60%, crypto -30%; Hacker worker output -50% for 48 ticks
2. **AI Anomaly** — 3 random buildings swap their output resources for 60 ticks (chaos event)
3. **Biotech Breakthrough** — food production +100% for 90 ticks; biotech research branch available
4. **Silicon Drought** — titanium building costs +50% for 60 ticks; titanium production +20%
5. **Viral Memetic Storm** — culture and faith both halved for 48 ticks, then doubled for 48 ticks
   (net neutral but timing matters for players near faith thresholds)

### Neon Era — Exclusive Events

1. **Corporate War** — plasma production -40%, dark_matter +40%; military output +60% for 90 ticks
2. **Augmentation Rebellion** — 15% of workers revolt; food drain -10% permanently (fewer augmented workers)
3. **Fusion Breakthrough** — plasma production ×3 for 120 ticks; one free Neon Era tech
4. **Black Market Surge** — gold income +150% for 60 ticks; faith -20% (moral cost of dealings)
5. **Consciousness Upload** — 12% of population digitized; housing freed, knowledge +60% permanently

### Cosmic Era — Exclusive Events

1. **Alien Signal Received** — knowledge + data ×4 for 120 ticks; Xenology research branch unlocked
2. **Stellar Phenomena** — dark_matter ×2 for 60 ticks; antimatter disrupted -50% for 30 ticks
3. **Reality Distortion** — 5 random buildings swap output resources for 60 ticks (terrifying at scale)
4. **Quantum Resonance** — quantum_flux ×5 for 30 ticks (spike fills storage; plan for it)
5. **Dimensional Rift** — all production halted for 12 ticks, then ×3 for 60 ticks (terrifying/rewarding)

### Event Pool Summary

| Source | Count | Availability |
|--------|-------|-------------|
| Universal events (existing) | 28 | All epochs, magnitude scaled |
| Stone Era exclusive | 5 | Stone Era only |
| Iron Era exclusive | 5 | Iron Era only |
| Steel Era exclusive | 5 | Steel Era only |
| Electric Era exclusive | 5 | Electric Era only |
| Digital Era exclusive | 5 | Digital Era only |
| Neon Era exclusive | 5 | Neon Era only |
| Cosmic Era exclusive | 5 | Cosmic Era only |
| Good epoch events (major epoch roll) | 10 | One fires per epoch transition (if good) |
| Challenging bad epoch events (major epoch roll) | 8 | One fires per epoch transition (if bad, non-catastrophe) |
| Catastrophe events | 7 (6 reachable) | Iron Era on: at most one per epoch per run (transition roll) |
| **Total** | **88** | |

Note: The 10 good + 8 bad + 7 catastrophe events are **epoch transition events**, separate from
the regular random event pool. They fire exactly once per epoch transition, replacing the normal
"age advance" announcement.

---

## UI — Epoch Badge

The epoch badge appears in the status bar / header area, near the age indicator. It updates on
epoch transition with a brief color flash and a one-line status message.

```
 Age: Medieval  ·  ⚒ Iron Era
 Age: Industrial  ·  ⚙ Steel Era
 Age: Digital  ·  ⬡ Digital Era
 Age: Galactic  ·  ✦ Cosmic Era
```

### Badge Styling per Epoch

| Epoch | Icon | Color |
|-------|------|-------|
| Stone Era | ◈ | Gray/brown |
| Iron Era | ⚒ | Rust/orange |
| Steel Era | ⚙ | Silver/steel blue |
| Electric Era | ⚡ | Yellow/gold |
| Digital Era | ⬡ | Cyan |
| Neon Era | ✦ | Magenta/purple |
| Cosmic Era | ✧ | Deep blue / white |

If the player has **Survived** (Endured) a catastrophe in this epoch, the badge gains a subtle
scar marker: `⚒̶` or `[⚒ Iron Era · Survived]`.

### Epoch Transition Announcement

On epoch transition, the dashboard status bar briefly shows:
```
[yellow]✦ The Steel Era Dawns — The Age of Iron gives way to industry and empire.[-]
```
(tview-styled, 5-second timeout, then normal display resumes)

---

## Implementation Notes

### Data Model

- `AgeDef` needs an `EpochKey string` field mapping each age to its epoch
- `EpochDef` struct: `Key`, `Name`, `Icon`, `Color`, `Ages []string`, `CatastropheKey string`
- `CatastropheDef` struct: `EpochKey`, `Name`, `FlavorText`, `EndureConsequences`, `SuccumbLegacyBonus`
- `EpochEventDef` struct: `Key`, `Type` (good_minor/good_major/good_legendary/bad_challenging/catastrophe),
  `Name`, `FlavorText`, `Effects []EventEffect`
- `GameEngine` tracks: `currentEpoch string`, `epochEventFired map[string]bool`, `legacyBonuses map[string]bool`,
  `catastropheHistory []string` (for civilization log), `survivedEpochs map[string]bool`

### Epoch Transition Trigger

```
advanceAge() {
    // ... existing age advance logic ...
    newEpoch := epochForAge(ge.currentAge)
    if newEpoch != ge.currentEpoch {
        ge.currentEpoch = newEpoch
        ge.bus.Publish(EpochAdvanced, newEpoch)
        ge.rollEpochEvent(newEpoch)   // fires immediately on epoch transition
    }
}

rollEpochEvent(epoch string) {
    faithPct := ge.faith / ge.faithCap
    goodChance := 0.50
    if faithPct < 0.25 { goodChance = 0.40 }
    if faithPct > 0.75 { goodChance = 0.60 }

    if rand.Float64() < goodChance {
        ge.rollGoodEpochEvent()
    } else {
        if rand.Float64() < 0.30 {
            ge.pendingCatastrophe = epoch   // trigger UI modal
        } else {
            ge.rollChallengingEpochEvent()
        }
    }
}

rollGoodEpochEvent() {
    culturePct := ge.culture / ge.cultureCap
    var tier string
    switch {
    case culturePct > 0.75 && rand.Float64() < 0.15: tier = "legendary"
    case culturePct > 0.40: tier = "major"
    default: tier = "minor"
    }
    event := pickRandomFromPool(goodEpochEvents, tier)
    ge.applyEpochEvent(event)
    ge.bus.Publish(EpochEventFired, event)
}
```

### Event Pool Filtering (Regular Events)

EventManager's regular random event trigger filters the candidate pool by current epoch:
```
candidateEvents = universalEvents + epochExclusiveEvents[currentEpoch]
```
Previous epoch exclusive events are permanently removed from the pool on epoch transition.
Epoch transition events (good/bad/catastrophe) are a separate pool, fired once per epoch
transition — NOT drawn from the regular random event pool.

### Triggering (implementation)

In game/catastrophe.go:
- `triggerCatastrophe(epochKey, source)` sets `pendingCatastrophe`, appends an `EpochEventRecord`
  with `Outcome: "pending"`, and publishes `EventEpochEventFired` (`event_type: "catastrophe"`) so
  the dashboard toast fires. Sources: the transition roll, an honoured invite, or the dev
  console's `forceCatastrophe` (`/catastrophe`, via `game.DevConsoleCommand`).
- The dashboard's refresh loop shows the modal once per pending catastrophe; Esc hides it until
  the `catastrophe` command (or a save load, via `EventGameLoaded`) brings it back.

### Catastrophe Save State

In `GameSave`:
```go
LegacyBonuses        map[string]bool   `json:"legacy_bonuses,omitempty"`
CatastropheHistory   []string          `json:"catastrophe_history,omitempty"`
SurvivedEpochs       map[string]bool   `json:"survived_epochs,omitempty"`
EpochEventFired      map[string]bool   `json:"epoch_event_fired,omitempty"`
PendingCatastrophe   string            `json:"pending_catastrophe,omitempty"`
CatastropheFired     map[string]bool   `json:"catastrophe_fired,omitempty"` // deprecated, ignored, never written
SuccumbResearchDerived bool            `json:"succumb_research_derived,omitempty"`
Ruins                map[string]int    `json:"ruins,omitempty"`
```

`EpochEventRecord.Outcome` (`pending` / `endured` / `succumbed`) is stored on catastrophe records;
older saves get it reconstructed on load where the engine can tell, and are shown as "outcome not
recorded" where it can't.

Ruins persist across runs (they're part of your civilization's identity). Legacy bonuses are
permanent and never removed. `EpochEventFired` prevents a second transition roll in the same
epoch, which is also what limits catastrophes to one per epoch per run. `CatastropheFired` was
written by early builds of the overhaul; it stays in the struct only so those saves still verify.

---

## Decision Log

Epoch-system decisions. The project-wide log is in `README.md`.

| Date | Decision | Rationale |
|------|----------|-----------|
| 2026-09-26 | Prestige is the Cosmic Era's passage (the Last Passage), rolled once at confirmed prestige with the epoch odds; pending blocks only prestige; Endure keeps 50/70/85% of the run's points by Brace; Succumb grants a one-time Cosmic Legacy (+10% `production_all`, derived from a flag) | The Cosmic Era's four harbingers had nothing to warn of because the epoch has no transition out; prestige is the only passage it has, and points are what a prestige can lose, so Endure and Brace act on them |
| 2026-09-26 | Harbinger replaces invoke; epoch-long threads with the speaker changing each age; false prophets rolled once per thread (Stone, Iron, Steel Era); Appease x0.6 per level; Brace tiers (15%/30%, 10%/45%); costs priced off the passage | Choosing a catastrophe fits better as an answer to a warning than as a bare command; a thread gives the warning time to matter and uses 18 figures instead of 6; false prophets make early warnings worth doubting until the odds are printed; passage pricing keeps the price the same in every age so paying early is not a discount; x0.6 still lowers the odds after the worst faith-band drop the price can cause; Brace gives Endure-minded players something to buy without touching the odds |
