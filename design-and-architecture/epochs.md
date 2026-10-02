# AgeForge: Epoch System

## Overview

**Epochs** are the meta-progression layer above ages. The 22 ages group into **7 epochs**: six of 3 ages and the Cosmic Era with 4.
Epoch transitions are bigger civilizational milestones than age advances. They signal a fundamental
shift in what resources matter, what events can occur, and what dangers your civilization faces.

```
Ages 1-3   → Stone Era
Ages 4-6   → Iron Era
Ages 7-9   → Steel Era
Ages 10-12 → Electric Era
Ages 13-15 → Digital Era
Ages 16-18 → Neon Era
Ages 19-22 → Cosmic Era
```

Every 3rd age advance through the Space Age = 1 epoch transition (6 × 3 = 18 ages), and the Cosmic Era holds the last 4. It has no transition out; its passage is prestige (see [The Last Passage](#the-last-passage)).

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
| 7 | **Cosmic Era** | Interstellar, Galactic, Quantum, Transcendent | antimatter, quantum_flux | Galactic civilization, reality manipulation |

### What Changes at an Epoch Transition

1. **Primary building costs shift**: buildings in the new epoch are denominated in new resources
2. **Extraction lineages switch output**: Organic and Geological lineages produce new resources
3. **Metallurgy's processing chain advances**: new ore → new refined metal
4. **Event pool changes**: epoch-exclusive events replace previous epoch's exclusive events
5. **The new era's fate is rolled**: from the Iron Era on, the Cosmic Era included, a hidden roll decides whether a doom is fated to strike somewhere inside the era (see [Fated Dooms](#fated-dooms)). The transition itself never brings a catastrophe
6. **UI epoch badge updates**: the epoch badge near the age indicator changes color and icon
7. **New worker domains may unlock**: Hacker (Digital Era), Astronaut (Neon Era)

---

## Epoch Event System

At each epoch transition, the game rolls a **major epoch event**, a single significant event
that could be a boon or a setback. A transition never brings a catastrophe: from the Iron Era on a
doom is fated in secret on entering an era and strikes inside it (see
[Fated Dooms](#fated-dooms)).

### The Roll

```
Roll at each epoch transition (rollEpochEvent):

  Faith < 25% cap:   40% Good / 60% Bad
  Faith 25-75% cap:  50% Good / 50% Bad   ← baseline
  Faith > 75% cap:   60% Good / 40% Bad

  If GOOD → roll from Good Event pool (weighted by Culture level):
      Low culture:    Minor Good only
      Medium culture: Minor or Major Good
      High culture:   Minor, Major, or Legendary Good (rare)

  If BAD → Challenging Event (severe but recoverable; 8 events)

Then, on the same advance (rollFate): the new era's hidden fate.
```

**Catastrophe odds:** see [Fated Dooms](#fated-dooms). The old passage chance,
(1 − good chance) × `catastropheChanceOnBadRoll` (0.30) = 18% / 15% / 12% at low / mid / high
faith, survives as the Last Passage's odds and, times `FateStrikeScale` (5), as a fated doom's
strike chance (90% / 75% / 60%).

**Faith matters:** Keeping faith above 75% of its cap improves your odds of a good epoch event
at every transition and lowers the chance a fated doom strikes, which rolls on the faith fill at
the moment it strikes. This is a new strategic reason to invest in Faith buildings and workers
beyond its existing morale/cohesion/diplomacy uses.

**Culture matters:** Culture level gates the tier of good events you can receive. High-culture
civilizations occasionally receive Legendary Good events: powerful outcomes unavailable at
lower culture levels.

### Good Epoch Events

**Minor Good (any culture level):**

1. **Age of Plenty**: all resource production ×2 for 72 real hours
2. **Population Surge**: all worker classes +15% count, immediately recruited
3. **Ancient Cache**: fills 40% of every resource's storage cap with current resources
4. **Trade Winds**: gold ×3 for 48h; all trade routes open regardless of requirements
5. **Cultural Festival**: culture +30%, faith +20% instantly; morale bonus for 48h

**Major Good (requires medium culture):**

6. **The Grand Discovery**: 3 free techs from current epoch's research tree
7. **Worker Innovation**: all worker output multipliers permanently +10% (stacks across events)
8. **The Architect's Gift**: 10 free buildings of any current-age type, instant, no resource cost
9. **Peaceful Century**: all negative random events suspended for 96h; all production +20%

**Legendary Good (requires high culture, rare):**

10. **Epoch Blessing**: permanent +15% production for current epoch's primary resource;
    one unique epoch wonder unlocked (exclusive to this event, not in the normal wonder list);
    recorded in civilization history as a golden age entry

### Challenging Bad Epoch Events (every bad outcome)

Severe but recoverable, with no reset and no modal. Applied immediately on epoch transition.

1. **The Famine**: food production -60% for 120 ticks; workers begin leaving if not corrected
2. **Merchant Betrayal**: gold -50%; all trade routes suspended for 72 ticks
3. **The Great Fire**: up to 8 random buildings destroyed (never wonders or storage); no targeted penalty on surrounding buildings
4. **Epidemic**: worker count -20%; food drain +15% for 96h; faith influences severity
5. **Resource Drought**: current epoch's primary output resource -70% for 90 ticks
6. **Political Instability**: faith -60%; military output -40%; knowledge production paused 60 ticks
7. **Economic Crash**: all gold halved; building costs +50% for 72h
8. **The Dark Age**: knowledge production and all research paused for 48h; one random tech gains a knowledge debt that must be cleared before it can be used

### Civilizational Catastrophe

This section describes what is built (updated 2026-09-30). Ideas from the original design that
were never built are kept under **Future ideas (not implemented)** at the end of it.

**Rules**

- **When it strikes.** At a fated moment inside an era, or at an advance (in the final epoch, a
  prestige) that would outrun it, never at a transition. See [Fated Dooms](#fated-dooms).
- **Iron-epoch gate.** No catastrophe before the epoch containing the Iron Age
  (`config.CatastropheGateEpoch = "iron_era"`). The Stone Era never has one.
  Good and challenging epoch events are unaffected. `config.FateAllowed` is the same gate, so the
  final epoch can be fated too: its doom is the Reality Tear (`cosmic_era`), and the Last Passage
  at prestige is a separate catastrophe beside it (see [The Last Passage](#the-last-passage)).
- **One doom per era per run.** Each era's fate (`FateSave`) rolls once, on entry, and resolves
  once (`Resolved`: `struck`, `spared` or `revealed`). Succumb, prestige and a new game clear it
  (`clearHarbingerRun`).
- **No direct player trigger.** `catastrophe invoke` was removed (2026-09-26). Choosing to face
  a catastrophe is the Harbinger's **Invite** (see [Harbinger](#harbinger)), which sets
  `FateSave.Invited` (and `Fated`, making a false prophet's invented doom real): the strike still
  comes at `StrikeTick`, or at an advance or prestige gate, but it is certain. Invite on the Last
  Passage's thread arms `catastropheInvited` instead (consumed at prestige, reset by
  Succumb/prestige, saved as `catastrophe_invited`, kept on load only in the final epoch).
  `CatastropheOutlook` reports probability 1 while either is set. The dev console's `/catastrophe`
  forces one for testing, respects the gate and ignores the fate (a fated doom still comes later,
  once the forced one is answered).
- **Pending blocks progress.** A catastrophe only sets `pendingCatastrophe`; nothing is destroyed
  until the player chooses. The game keeps running, but `AdvanceAge`, the plan's advance and
  `DoPrestige` refuse while one is pending, and a strike due meanwhile waits: one is never
  overwritten. Endure clears `ageReady`, so the next advance checks its requirements against the
  reduced stock. `resolveLastPassage` refuses while a catastrophe is pending ("The Reality Tear
  came first. Answer it before the Last Passage."), and the modal shows the pending catastrophe
  first (`pendingChoiceKey`).
- **No Defer.** The modal has two choices. Esc closes it without choosing; a status-bar badge
  shows the pending catastrophe and the bare `catastrophe` command reopens the modal. A save with a
  pending catastrophe shows the modal again on load.
- **Seeded.** Every roll (fate, strike, epoch event, destroyed buildings, ruins) comes from the
  run's seeded `GameEngine.rng`, drawing from pools built in sorted-key order.
- **Outlook.** `GameEngine.CatastropheOutlook()` (also `GameState.CatastropheOutlook`) reports the
  outlook as the player can know it and never reads the hidden fate. While a harbinger warns of
  this era's doom: `Warned`, `Possible` and the harbinger's tier and probability (the claim, for a
  false prophet). Otherwise `Possible` is true while the era is `FateAllowed`, nothing is pending
  and its doom has neither struck nor been spared (`fateSettled`), with probability 0: a fated era
  and a quiet one read the same until the harbinger comes. `FaithFill` is the current faith fill.
  It is read-only and lock-safe. `Passage` is `"epoch"` (`PassageEpoch`) or, in the final epoch,
  `"prestige"` (`PassagePrestige`, with `NextEpochKey` empty). There `Probability` and `Tier` are
  always the Last Passage's odds at its own thread's Appease
  (`appeaseMultiplierOf(lastPassageThread())`), with a coarse tier (none/low/medium/high: under
  14%, under 17%, 17% and up), and `Warned` says the Reality Tear's harbinger is speaking
  (`fateThread() != nil`); its warning is on `GameState.Harbinger`, and the UI shows both lines.
  See [The Last Passage](#the-last-passage).

**The modal** is a box floating over the dashboard, sized to its content:

```
╔════════════════════ ☄ Iron Era Catastrophe ════════════════════╗
║                    ☄ The Great Plague                            ║
║     A devastating plague sweeps your cities. The streets…        ║
║                                                                  ║
║ ── ENDURE: weather the catastrophe ──                            ║
║   • 20% of buildings destroyed (wonders and storage are spared)  ║
║   • All resources reduced to 15%                                 ║
║   • 25% of workers lost; workers of destroyed buildings go idle  ║
║   • Production -10% for 216 ticks, morale -10                    ║
║   ✓ Age, research, wonders and prestige preserved                ║
║   ✓ Survived marker on the epoch badge                           ║
║                                                                  ║
║ ── SUCCUMB: let civilization fall ──                             ║
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

### ENDURE: Consequences

- `floor(destroyable buildings / 5)` destroyed, at least 1 if any. Wonders and storage
  (`isDestroyable`, game/buildings.go) are neither destroyed nor counted. With a Harbinger Brace
  (`pendingBraceLevel` 1 / 2) it is 15% / 10%, same floor.
- Workers assigned to destroyed buildings return to the idle pool (same rule as selling).
- All unlocked resources drop to 15% of their stored amounts (30% / 45% braced).
- 25% of the single worker pool is lost; every building's assignment shrinks by the same share,
  whatever its worker domain.
- Reconstruction Effort: `production_all` −10% for 216 ticks. Morale −0.10.
- "Survived" marker on the epoch badge and a civilization-log entry.
- Research, wonders, age and prestige are untouched.
- The army's garrison softens the building and stock losses; see Army defense below.

### Army defense (the garrison)

The Defense Rating (`soldiers × 2 × (1 + military_power)`) is measured against the raid threat
of the age being played, `config.AgeThreat(order) = 160,000 × 2^order` (1.28M in the Iron Age),
and blunts `0.45 × defense / (defense + threat)` of a raid's losses (`config.DefenseMitigation`).
Constants and the reasoning for them are in `config/defense.go`; the engine side is
`game/defense.go`.

- **What it touches**: random events flagged `Raid` (their `steal_resource` and `worker_loss`,
  not their production debuffs); diplomacy war raids (a raid bigger than the stock still takes
  nothing, and the army never makes it land); Endure's building and stock losses. Nothing else:
  disasters, unrest, Endure's worker loss, the Reconstruction debuff and the Last Passage are
  unchanged.
- **Endure**: Brace first (`braceDestroyPct` / `braceKeepFrac`), then the garrison blunts its
  share of what is left, measured against the threat of the age the catastrophe strikes in (the
  current age for a harbinger preview: the doom strikes in it or later in the era). Brace and
  garrison together cut at most
  `EndureReductionCap` = 60% of the unbraced loss (at least 8% of buildings fall, at most 66% of
  stock is kept). `computeEndure` is the one function Endure, the catastrophe modal
  (`GameState.PendingEndure`) and the harbinger's Brace preview all call, so a preview can
  never promise more than Endure delivers. Buildings saved round down, so the garrison never
  saves more than its share.
- **No soldiers, no change**: with defense 0 every code path returns the pre-army numbers bit
  for bit (the steal, worker and keep arithmetic only runs when the share is above 0).
- **The incidental garrison**: age gates require military buildings and the lineage carries them
  forward, so a player who never thinks about the army still holds a garrison. The threat base
  was calibrated on it: the greedy smoke bot blunts 8-19% of a raid from the Iron Age on, and a
  deliberate army (the `-army=on` bot) 25-40% from the Industrial Age.
- **Endure timing**: a doom strikes at its fated moment anywhere in its era, so the garrison is
  measured against whatever age the player is in when it strikes, often one they have been
  building in for a while rather than one just entered. The harbinger's Brace preview measures
  against the current age, so a doom that strikes after an advance meets double the threat the
  preview assumed. The choice waits for the player and the modal preview is live, so training
  soldiers before choosing Endure still pays. Raids are where it works all age long.
- **No upkeep**: soldiers cost nothing ongoing. Stockpiling is bounded by soldier storage and
  saturates at the cap, and measured runs show no pacing gain from a garrison (see the Decision
  Log), so upkeep would be a tax with nothing to correct.
- **Tally**: `GameStats.Defense` (omitted while empty, so saves without a garrison are
  unchanged) records what the army saved this run, shown in the Army panel.

### SUCCUMB: Consequences

- Up to 8 destroyable buildings (neither wonders nor storage) become ruins (50% base output, no
  workers). Ruins persist across Succumb and prestige, capped at **24**: past the cap the
  lowest-value ruins (earliest `RequiredAge`, then lowest base output, then key) are dropped
  first. The cap also trims old saves on load.
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

Six legacies (Iron to Cosmic) are reachable, so Ancient Knowledge tops out at +150% (+175% with a
Stone Era legacy from an older save).

### Catastrophe vs Regular Prestige

| | Regular Prestige | Catastrophe Succumb |
|--|-----------------|---------------------|
| Trigger | Player-initiated from the Modern Age | A fated doom's strike: 27% of eras from the Iron Era on hold one, which then hits 60-90% by faith (lowered by Appease), or certain after the Harbinger's Invite |
| Reset scope | Full | Full; up to 8 new ruins (24 max) carry forward |
| Bonus pool | Prestige upgrade tree + points | Epoch legacy bonus + Ancient Knowledge |
| Repeatable | Yes | Once per era per run; bonuses once per epoch ever |
| Blocked by a pending catastrophe | Yes | n/a |
| Can bring a catastrophe | From the Cosmic Era: the Last Passage, after an open Reality Tear settles | n/a |

### Future ideas (not implemented)

Kept from the original design for reference. None of this exists in the game.

- **Endure lasting consequences:** building costs +20% and worker food drain +10% for 72 real
  hours; random events 20% harder during recovery.
- **Endure permanent rewards:** a Reconstruction tech branch (5 epoch-specific recovery techs); a
  **Monument to the Fallen** wonder (free, +2,000 culture, +500 faith/tick, morale bonus); titles
  such as "The Undying Iron Lords" shown in the Stats panel.
- **Succumb extras:** Ancient Knowledge scoped to the fallen epoch's tech tree; a Catastrophe Title;
  **Faster Return** (techs from 2 epochs below the catastrophe auto-complete next run); an
  **Exclusive Starting Event** in the first 10 ticks of the next run; ruins marked ☒ in the
  Economy panel.
- **Exclusive uniques per catastrophe:** Meteor Fragment wonder; Ancient Immunity passive (events
  15% less severe) and a Plague Doctor worker class; War Doctrine tech and Armistice Monument;
  Fallout Shelter tech and Nuclear Vault; Ghost Protocol tech and Dead Drop Network; Phoenix
  Protocol tech and Corporate Ruins wonder; Reality Anchor tech and Scar in Reality wonder
  (the Reality Tear legacy was also meant to boost antimatter + quantum_flux instead of dark_matter).

---

## Fated Dooms

A catastrophe strikes at a secret moment inside an era, never at a transition (changed
2026-09-30). Code: `game/fate.go` (the fate, the strike, the advance and prestige gates, the
false-prophet reveal), `game/catastrophe.go` (odds and outlook), `game/harbinger.go` (the threads,
the Cosmic Era's parked one). Player docs:
`site/docs/catastrophe.md` and `site/docs/harbinger.md`.

### The fate roll

- Every epoch rolls a `FateSave` on entry (`rollFate()`): `detectEpochTransition` calls it after
  `rollEpochEvent`, and `ensureFate()` covers a run's first tick (the Stone Era is never entered by
  a transition) and saves without a fate.
- `Fated = config.FateAllowed(epoch) && roll < FateChance` (0.27). `FateAllowed` is the Iron gate
  (`CatastropheAllowed`): Iron through Cosmic, the final epoch included, whose doom is the Reality
  Tear. Before the gate a fate can only hold a false prophet.
- A fated doom (or a false prophet) gets `StrikeTick = EntryTick + offset × Window`, with `offset`
  uniform in [0, 1) and `Window = expectedEraTicks(epoch)`, the sum of the era's
  `expectedAgeTicks` (Iron 10.5 h, Steel 21 h, Electric 31 h, Digital 42 h, Neon 60 h, Cosmic 96 h
  at 1x). The doom can fall in any of its ages, mid-age included.
- `LeadFrac` is uniform in [`harbingerLeadMin`, `harbingerLeadMax`] = [0.20, 0.60].
- Five `ge.rng` draws every time (fated, false prophet, offset, lead, claim), whatever the outcome,
  so the stream's shape never depends on the fate.
- `GameSave.Fate` persists it, so a reload cannot re-roll it.
- `expectedAgeTicks(age)` is the one measure of expected duration: the era's window, the lead, the
  "this age" forecast and the shortest warning. It is `config.AgeTargetTicks` today, so a pacing
  change, or a later per-age speed-up, carries through in one place.

### The harbinger's arrival

- Due at `fateArrivalTick() = StrikeTick − LeadFrac × expectedAgeTicks(current age)`: 30-90 minutes
  before the strike in the Iron Age, 1.6-4.8 hours in the Industrial Age, 2.4-7.2 hours in the
  Atomic Age. The lead follows the current age, so advancing into a longer age can bring the
  harbinger at that advance (`harbingerOnAgeAdvance` checks `fateHarbingerDue`).
- `fateTick()` brings it when due. An arrival tick before `EntryTick` only means it comes on
  entry: never before the era began.
- Shortest warning: `fateArrive` moves `StrikeTick` to at least
  `now + harbingerLeadMin × expectedAgeTicks(age)`, so a strike fated for the era's first moments
  is held until the shortest lead has passed.
- `FateSave.Arrived` is set on arrival: one doom's harbinger per era per run (`harbingerArrived`
  marks only the Last Passage thread). `fateHarbingerDue` waits only for another doom's thread: a
  live Last Passage thread is parked when the doom's harbinger comes (see
  [The Cosmic Era's two threads](#the-cosmic-eras-two-threads)).

### The strike

- At `StrikeTick`, with the doom's harbinger present (`fateThread()`) and nothing pending, `fateStrike` draws one
  `ge.rng` value against `strikeChance()` = `strikeBase() × harbingerAppeaseMultiplier()`, where
  `strikeBase = (1 − epochGoodChance()) × catastropheChanceOnBadRoll × FateStrikeScale`: 90% / 75%
  / 60% at low / mid / high faith (the old passage chance × 5), read at that moment. Certain when
  `Invited`.
- Hit: `Resolved = struck` and `triggerCatastrophe(epoch, source)` (source `invited` when invited).
  Miss: `Resolved = spared`. Either way `resolveHarbinger` speaks the verdict, records it and clears
  the thread.
- `StrikeChanceAt(fill, hasStorage, appease)` is the same rule as a pure function, for the smoke
  report.

### No outrunning it

`fateBeforeAdvance(next)` runs before every age advance: `AdvanceAge` (the `advance` command) and
`runPlan` (the plan's `plan advance`). It acts when the fate is open and the advance leaves the
era, or the figure speaking has said `WhenThisAge`. In the final epoch, whose passage is prestige,
`fateBeforePrestige()` runs in `DoPrestige` before `rollLastPassage()` and acts on any open fate.
Both hand over to `fateAtPassage(leaving, again, gerund)`:

1. No harbinger has come: `fateArrive()` now, and refuse with `errHarbingerAtGate(h, warning,
   again)` ("The Town Crier stands in your way, warning of impending doom before this age is out.
   Type 'harbinger' to answer, or advance again to meet it."; at a prestige, "... or confirm
   prestige again to meet it."). The plan puts its item back and tries again next tick,
   so a planned advance meets the strike one tick later. Not in the Stone Era, where nothing can
   strike: a false prophet who has not come by the era's end never comes.
2. A false prophet not invited: `revealFalseProphet(leaving)`, and the advance goes on.
3. A catastrophe already pending: refuse (`catastropheBlockErr(gerund)`).
4. Otherwise `fateStrike(true)`: a hit refuses the advance or prestige behind the pending
   catastrophe, a miss lets it through (at a prestige, on to the Last Passage roll). `AtAdvance`
   records it ("settled as you advanced" in the Epoch panel).

`detectEpochTransition` settles an open doom whose harbinger is here the same way for a direct
advance (tests, dev tools), and drops an era's thread at the era's end. Succumb, and a prestige
from an earlier era (the Digital or Neon Era), drop an open fate with the run
(`clearHarbingerRun`).

### The Cosmic Era's two threads

The final epoch runs two threads: its fated doom's (the Reality Tear, `TargetEpoch` `cosmic_era`)
and the Last Passage's (`TargetEpoch` `""`, started on entry by `maybeLastPassageArrive`).

- `fateThread()` is the live thread of the current era's doom, or nil; `lastPassageThread()` is the
  Last Passage's, live or parked.
- `fateArrive` parks a live Last Passage thread in `ge.parkedHarbinger` while the doom's thread
  speaks, and a Last Passage thread that starts while the doom's speaks is parked at once. Answers
  go to `ge.harbinger`, the thread speaking; the parked one keeps its levels, its invite and its
  figure until it resumes. `HarbingerView.LastPassageWaiting` makes the panel say "The Last
  Passage still waits at your next prestige. Your answers to it stand."
- When the doom resolves, `settleHarbinger` clears its thread and calls
  `resumeLastPassageThread()`: the parked thread is live again, handed off to the current age's
  figure (with its log line) if the age moved on.
- The Last Passage reads its own thread wherever it is: its odds through
  `appeaseMultiplierOf(lastPassageThread())`, its Endure share through `lastPassageBraceLevel()`,
  its invite state, and `settleHarbinger("", ...)` at prestige picks the live or parked thread.
- Both threads price at `cosmic_era`, each with its own levels. The doom's Brace preview shows the
  usual Endure numbers with the garrison; the Last Passage's shows the points share.
- `fateBeforePrestige` settles an open doom before `rollLastPassage`. If both are ever pending,
  `resolveLastPassage` refuses until the catastrophe is answered.
- `GameSave.ParkedHarbinger` persists a parked thread. On load it stays parked only while the
  doom's thread is live; otherwise it becomes the live thread.

### Offline

`applyOfflineProgress` cuts each step at `fateNextEventIn()` (the harbinger's arrival or the
strike; a false prophet's moment is not an event) and runs `harbingerTickCheck()` after it, so both
land at their own tick while the player is away. A struck doom waits, pending, for the player.

### What the player can see

Until the harbinger arrives nothing may reveal whether a doom is fated. `catastropheOutlook()` reads
only the harbinger present and what has resolved in the open (`fateSettled`: `struck` or `spared`;
a revealed false prophet reads like a quiet era). `EventFateRolled` and `EventFateResolved` carry
the hidden fate for tests and the smoke report; `ui/fate_guard_test.go` bans `ui/` and `mapmodel/`
from naming them, `FateSave` or the test accessors.

The `catastrophe` command and the Harbinger panel's outlook read "No harbinger has come: the Iron
Era is quiet, for now. A doom is always foretold before it strikes." in a quiet era; the warning
("The Oracle warns of doom before this age is out: medium risk of catastrophe (no figures this
early), faith 40% full.") while a harbinger is here; "No catastrophe can strike in the Stone Era."
there; and, once the doom has resolved, that nothing more will strike before the era ends. The
Epoch panel shows "Catastrophe: spared (nothing more will strike this era)" after a miss. In the
final epoch both show the Reality Tear's warning while its harbinger speaks ("Your future self
warns of doom before this age is out: 90% catastrophe chance (high), faith 0% full.") and, on a
line of its own, the Last Passage at its own odds; there is no "quiet, for now" line there.

### Timing forecast and severity

- `fateWhen(def)`: `WhenUntold` for a figure with `config.TimingNone` (before the Classical Age).
  Otherwise `WhenThisAge` in the era's last age, or when `StrikeTick` falls before the current
  age's end on the era's expected schedule (each age's `expectedAgeTicks` summed from
  `EntryTick`); else `WhenThisEra`. Set at arrival and at each handoff. Shown as "before this age
  is out" or "before the Iron Era ends" (`harbingerWhenText`, naming only the current era).
- `fateTierFor(p) = catastropheTierFor(p / FateStrikeScale)`: low under 70%, medium 70% to under
  85%, high 85% and up, so the three faith bands still map one-to-one (high faith → low, mid →
  medium, low → high). One Appease level brings any real doom to low. The Last Passage keeps
  `catastropheTierFor` (14% / 17%).

### False prophets

Rolled in `rollFate` only when nothing is fated: `falseRoll < HarbingerFor(ge.age).FalseProphetChance`,
the era's first age on entry (Primitive 8/64, Iron 5/64, Renaissance 2/64, 0 from the Industrial
Age). That is 12.5% of Stone Eras, 0.73 × 5/64 ≈ 5.7% of Iron Eras and 0.73 × 2/64 ≈ 2.3% of Steel
Eras. A false prophet draws `StrikeTick`, `LeadFrac` and a claimed tier (medium or high) like a real
doom and arrives the same way. At `StrikeTick` nothing happens (`lying()` skips the strike). It is
revealed at the advance that closes its foretold window: the age's end after `WhenThisAge`,
otherwise the era's end (figures with no timing included). In the Stone Era every harbinger is a
false prophet, and `harbingerPowerless` refuses all three answers there.

### Expected catastrophes

A first run to the Modern Age prestige lives through three eras that can be fated (Iron, Steel,
Electric; the Digital Era's doom rarely strikes before that prestige): 3 × 0.27 × 0.90 = 0.73
expected at low faith, 0.61 at mid, 0.49 at high. The transition roll this replaced met four
catastrophe-capable transitions per first run at 12-18% each: the smoke bot (low faith) expected
0.72 and measured 0.673 over 49 seeds; with fated dooms it measures 0.73 ± 0.12 over 49 seeds. A
deep run to a Quantum Age prestige meets six fated eras: 6 × 0.27 × 0.90 ≈ 1.46 expected at low
faith plus the Last Passage, against 6 × 0.18 = 1.08 plus the Last Passage under the transition
rolls (the first-run calibration makes each era's doom a bit likelier than one old transition
roll). The smoke progression report's Fated dooms section tracks the new numbers (fated eras,
strike offsets, warnings, struck, spared and outrun dooms).

### Saves from before fates

`restoreFateState` runs after the epoch is restored and before the harbinger. A save without
`fate` but with a live thread in an era that can be fated gets a fate carrying it over (a false
thread stays false unless invited) with `StrikeTick` at `MaxInt32`, so the era's final advance gate
strikes it (`Arrived` set, since its harbinger is already here), and `restoreHarbingerState`
points the thread's `TargetEpoch` at its own era. A Stone Era thread is dropped. A Last Passage
thread keeps its rules. Any other save, the Cosmic Era included, gets its era's fate rolled on the
first tick, from the current age.

### Test hooks and dev console

`ForceFateForTest`, `ForceQuietFateForTest`, `ForceFalseProphetForTest` and `FateForTest` are for
tests and the smoke suite only. The dev console's `/harbinger` (`summonHarbinger`) fates a doom
that strikes one lead from now and brings its harbinger (reopening the era's doom if it had
resolved); in the Stone Era it sends a false prophet. In the final epoch it starts the Last
Passage thread first if it has not come, and otherwise fates a Reality Tear and brings its
harbinger.

---

## Harbinger

Built 2026-09-26. A thread starts only for a fated doom (see [Fated Dooms](#fated-dooms)) or the
Last Passage. Code: `game/harbinger.go` (threads, answers, resolution), `game/fate.go` (when a
harbinger comes and what it says about when), `config/harbingers.go` (the 22-entry roster: `Name`,
`Description`, `AppeaseLabel`, `BraceLabel`, `InviteLabel`, plus the derived `ForecastPrecision`,
`ForecastTiming` and `FalseProphetChance`), `flavor/` (arrival, warning, action and outcome
lines). Player docs: `site/docs/harbinger.md`.

### Threads, start and handoff

A **thread** is one warning: an era's doom (`TargetEpoch` is its own `EpochKey`) or, in the Cosmic
Era, the Last Passage (`TargetEpoch` `""`; see [The Last Passage](#the-last-passage)). The Cosmic
Era can hold both at once (see [The Cosmic Era's two threads](#the-cosmic-eras-two-threads)). A
thread lasts until its doom resolves. The speaker is always the current age's roster figure, so all
22 figures can appear.

Hooks, all under the write lock:

- `harbingerTickCheck()`, near the top of `doTick` and after each offline step. In the final
  epoch it first starts the Last Passage thread if none has come (a Succumb, a prestige, a load),
  one check per epoch (the unpersisted `harbingerCheckedEpoch`). Then, in every epoch, it runs
  `fateTick()`: roll the era's fate if missing, bring the harbinger when due (`fateArrive`), strike
  at `StrikeTick`.
- `harbingerOnAgeAdvance()`, at the end of `advanceAge` (after `detectEpochTransition` and
  `fireAwakening`). The live thread of the current era (a doom's or the Last Passage's) is handed
  off (`harbingerHandoff`) if its `Age` differs from the new age; a parked Last Passage thread
  keeps its figure until it resumes. In the final epoch the Last Passage thread starts if it has not
  (`maybeLastPassageArrive`). Then the new age's longer lead may bring a fated doom's harbinger
  (`fateHarbingerDue`).
- `fateBeforeAdvance()` and, in the final epoch, `fateBeforePrestige()` bring the harbinger at the
  gate when none has come (see [No outrunning it](#no-outrunning-it)).
- `restoreHarbingerState()` starts nothing and draws nothing: a save loads to exactly the state it
  was written in, and the next tick brings whatever is due.

`fateArrive` builds the thread with the current figure, sets `When` (`fateWhen`), `AnnouncedTier`
(`harbingerDisplay`) and the lines, and for a false prophet the `ClaimFactor`. It sets
`FateSave.Arrived`, parks a live Last Passage thread, logs "⚑ The Oracle has come, warning of
impending doom before this age is out. Type 'harbinger' to answer." and publishes
`EventHarbingerArrived`. `maybeLastPassageArrive` requires no Last Passage thread, live or parked
(`lastPassageThread()`), and `harbingerArrived[currentEpoch]` unset; `harbingerArriveLastPassage`
then requires the Last Passage to be possible, and parks the new thread at once if a doom's thread
is speaking.

A **handoff** appends the new age to `HarbingerSave.Chain`, sets `Age`, re-derives `When` and
`AnnouncedTier` and draws a fresh arrival and warning line in the new voice. Levels, the invite,
`FalseProphet` and `ClaimFactor` stay with the thread. Each arrival or handoff logs, publishes
`EventHarbingerArrived` (payload `handoff` true on a handoff, which the toast renders as "takes up
the warning") and keeps the status-bar badge up while `GameState.Harbinger` is non-nil.
`HarbingerView.Earlier` lists the earlier figures for the panel's "Took up the warning from ..."
line. Nothing expires, and an era's thread cannot outlive its era.

### Forecast: timing and precision

- `ForecastTiming` (`timingFor`): `TimingNone` before the Classical Age, `TimingAge` from it. The
  Desert Prophet, who opens the first era that can be fated, still gives no timing, so the first
  doom a new player meets stays mysterious; the Oracle, whose whole trade is prophecy, is the first
  to name a time. `HarbingerView.When` and `WhenText` carry it; the Last Passage's is always
  `WhenUntold`.
- `ForecastPrecision` (`precisionFor`): numeric from the Industrial Age, where the panel prints
  `HarbingerView.Probability`. Before it the panel shows a severity only.

### False prophets and ClaimFactor

- An era's false prophet is rolled with its fate (see [False prophets](#false-prophets) above),
  against the entry age's `FalseProphetChance` (`(8 - ageIndex) / 64` before the Industrial Age, 0
  from it). The other ages' chances only apply when an old save rolls a fate mid-era. The Last
  Passage thread still draws its own roll when it starts (always 0 in the Cosmic Era).
- The claim is stored as `ClaimFactor = FateStrikeScale × claimBase[tier] / strikeBase()` at
  arrival, with `claimBase` 0.15 for medium and 0.18 for high, so it opens at 75% or 90%.
  `harbingerDisplay()` then reports `strikeChance() × ClaimFactor` (capped at 1, tier never below
  low), so Appease, a faith-band change or an Invite move the false claim exactly as they move a
  true one, and every figure repeats it. A false Steel Era thread that reaches the Newsboy prints
  the claimed figure.
- `HarbingerView` carries no false-prophet flag, and the UI's `catastrophe` outlook and Epoch
  panel show the thread's tier and figure while one is live.
- Old saves with `FalseProphet` but no `claim_factor` load with `ClaimFactor = 1`.

### Actions and era-based costs

Answers belong to the doom and carry across handoffs. Costs are pure functions of the thread's
epoch (`harbingerAppeaseCost(epochKey, level)`, `harbingerBraceCost(epochKey, level)`), rounded
up, level 2 at double, so the price is the same in every age of the era. Pricing off current
caps would make the era's first age (smallest caps) a discount.

- `harbingerAppeaseAges(epochKey)`: the epoch's ages, minus the game's last age.
- `harbingerHeldSinceStart(epochKey)`: resources unlocked (cumulative `UnlockResources`) by the
  epoch's first age, so every price is payable in every age of the epoch.
- `harbingerAdvanceAges(epochKey)`: the epoch's later ages plus the next epoch's first age.

| Action | Cost (level 1) | Effect | Cap |
|--------|----------------|--------|-----|
| Appease | `harbingerAppeaseIncomeShare` (1/4) of `config.FlowIncome` summed over `harbingerAppeaseAges` at their `AgeTargetTicks`, in faith, and in culture if culture is held since the start (Steel Era on); rounded up to 2 significant figures | Multiplies the real strike chance by `harbingerAppeaseFactor` = 0.6 per level (0.36 at 2) | 2 |
| Brace | 12% of `harbingerBraceBasis`: per resource held since the start, minus faith and culture, the largest `ResourceReqs` across `harbingerAdvanceAges` | Endure destroys 15% / 10% of destroyable buildings (neither wonders nor storage) and keeps 30% / 45% of resources (unbraced 20% / 15%) | 2 |
| Invite | free | Sets `HarbingerSave.Invited` and `FateSave.Invited` (and `Fated`); the strike stays at its fated moment. On the Last Passage's thread it arms `catastropheInvited`. Appease refuses afterwards, Brace does not | once per thread |

All three go to `ge.harbinger`, the thread speaking, and are refused for a thread in an era where
nothing can be fated (`harbingerPowerless`: the Stone Era's false prophets), and while the Last
Passage is pending.

Level-1 prices from the current config (the Stone Era's are never charged, since its answers are
refused; both Cosmic Era threads share the Cosmic price in [Costs](#costs), each with its own
levels):

| Thread | Appease | Brace |
|--------|---------|-------|
| Stone | 140 faith | 9,600 food, 4,800 wood, 2,400 knowledge |
| Iron | 14,000 faith | 26,400 knowledge, 26,400 stone, 6,360 iron, 21,600 gold |
| Steel | 190K faith, 2M culture | 3.6M knowledge, 1.8M gold, 288K steel |
| Electric | 3M faith, 42M culture | 56.4M steel, 924K oil, 3.96M electricity |
| Digital | 31M faith, 460M culture | 156M gold, 117.6B electricity, 19.2B data |
| Neon | 340M faith, 5.2B culture | 288B electricity, 46.8B data, 3B crypto |

Appease follows income, not storage (2026-09-27): faith is a flow resource at hand-set rates
with no market, so the old 15%-of-passage-storage price was out of reach in most threads once
the pacing rebalance shortened the ages. `TestAppeasePayableWithinThread` checks that the
modeled income over the era's ages reaches level 1 before its last age ends and levels 1 and 2 by
its end. A thread usually lasts only its lead (20-60% of an age), so in practice Appease is paid from faith
banked before the harbinger came. When storage cannot yet hold a price, `shortfall` adds "(your X
storage must reach N first)" to the refusal.

Appease is priced at the pacing targets, so it grows with the curve while faith and culture
storage stay as typed: the one-week curve (2026-10-01) raised the Appease prices about 2.6x.
The static smoke check (`smoke.StaticHarbingerPrices`, also run by `go test ./smoke`) holds
every Appease and Brace price, both levels, to the most storage buildable in its era's first
age; they all fit with room to spare.

- **Appease** is applied through `harbingerAppeaseMultiplier()` (the live thread's
  `appeaseMultiplierOf`), which scales both the strike roll (`strikeChance`) and the displayed odds
  (`harbingerDisplay`), so the two cannot disagree. The Last Passage's chance uses its own thread's
  levels, live or parked (`appeaseMultiplierOf(lastPassageThread())`), in `catastropheOutlook` and
  so in `rollLastPassage`. Worst case the faith spend drops
  the fill from the top band to the bottom (60% to 90% strike chance, x1.5), and x0.6 still leaves
  0.9x, so each level always lowers the odds.
- **Brace** lives on `HarbingerSave.BraceLevel` until resolution, then moves to
  `ge.pendingBraceLevel` if the catastrophe came. `Endure` reads and clears it
  (`braceDestroyPct`, `braceKeepFrac`). Succumb ignores it. Keeping it on the pending catastrophe
  means Esc-then-Endure and save/load both keep the discount. The panel's preview counts the
  garrison against the current age (`endurePreview(level, ge.age)`). The Last Passage's Endure
  share reads its own thread's level, live or parked (`lastPassageBraceLevel`).

### Resolution

`fateStrike` calls `resolveHarbinger(epochKey, came)` when an era's doom resolves (at `StrikeTick`
or at an advance gate), `came` telling whether it struck. `revealFalseProphet` calls
`settleHarbinger` with its own verdict line when a false prophet's window passes. `completePrestige`
resolves the Last Passage thread with `epochKey` `""`, live or parked. After an era's doom thread
is cleared, `resumeLastPassageThread()` brings back a parked Last Passage thread. The verdict is
spoken in the last figure's voice. Outcomes (`HarbingerRecord.Outcome`):

| Outcome | Condition |
|---------|-----------|
| `fulfilled` | came and invited (an invited false prophet included; its log line says the warning was invented) |
| `vindicated` | came, not invited |
| `spared` | the strike missed, true thread ("⚑ The doom the Oracle foretold passed you by. The warning was real, and you were spared.") |
| `discredited` | a false thread, revealed |

It logs the verdict (and the braced Endure numbers if relevant), draws one flavor line, appends a
`HarbingerRecord` (with `Age`/`Name` of the last figure, the full `Chain`, `When`, and for an
era's doom `EntryTick`, `Window`, `StrikeTick`, `ArrivedTick` and `AtAdvance`) to
`harbingerHistory`, and clears the live thread. The Epoch panel lists each record's chain of
figures, for example "The Oracle, The Town Crier → doom in the Iron Era: Spared (settled as you
advanced)".

### Determinism

All draws come from the seeded `ge.rng` under the write lock. The fate draws five values on entry
(see [The fate roll](#the-fate-roll)) and a strike one more. An arrival draws the arrival and
warning lines from the engine's flavor `Stream`; a Last Passage thread first draws its
false-prophet `Float64()` (always), then `Intn(2)` for the claimed tier only if the thread is
false. A handoff draws two lines; each action and the resolution draw one.

### Persistence and resets

`GameSave` fields, all `omitempty` (old saves load clean):

```go
Harbinger          *HarbingerSave    `json:"harbinger,omitempty"`
HarbingerArrived   map[string]bool   `json:"harbinger_arrived,omitempty"`
CatastropheInvited bool              `json:"catastrophe_invited,omitempty"`
PendingBraceLevel  int               `json:"pending_brace_level,omitempty"`
HarbingerHistory   []HarbingerRecord `json:"harbinger_history,omitempty"`
Fate               *FateSave         `json:"fate,omitempty"`
ParkedHarbinger    *HarbingerSave    `json:"parked_harbinger,omitempty"`
```

`HarbingerSave` carries `chain` (ages that have spoken, first to current), `when` and
`claim_factor` (false threads only); `HarbingerRecord` carries `chain`, `when` and the doom's
timing (`arrived_tick`, `entry_tick`, `window`, `strike_tick`, `at_advance`). `FateSave` holds
`epoch_key`, `fated`, `false_prophet`, `arrived`, `entry_tick`, `window`, `strike_tick`,
`lead_frac`, `claim`, `invited`, `resolved`, `resolved_tick` and `at_advance`. All `omitempty`.

On load, levels are clamped to 2, an empty `Chain` becomes `[Age]`, an unknown roster age drops
the thread, `pendingBraceLevel` is zeroed unless a catastrophe is pending, and a fate for another
era is dropped (the first tick rolls the current era's). A parked Last Passage thread stays parked
only while the doom's thread is live; otherwise it loads as the live thread. `clearHarbingerRun`
(live and parked threads, fate, arrivals, `harbingerCheckedEpoch`, invite, pending Brace) runs on
Succumb, prestige and reset, and
the next tick rolls the new run's Stone Era fate. `harbingerHistory` follows `epochEventHistory`:
kept by Succumb, cleared by prestige.

---

## The Last Passage

Built 2026-09-26. Code: `game/last_passage.go`, with the Cosmic thread's pricing in
`game/harbinger.go`. Player docs: `site/docs/prestige.md#the-last-passage`.

### Prestige is the Cosmic Era's passage

The Cosmic Era has no next epoch, so its passage is prestige; its own fated doom, the Reality Tear,
runs beside it (see [Fated Dooms](#fated-dooms)). `CatastropheOutlook.Passage` is `"prestige"`
there and `NextEpochKey` is empty; `Possible` is false once the Last Passage is pending. The Cosmic
Era gets a Last Passage thread on entry, at the Interstellar Age, with the four
cosmic figures handing off per age (Distress Beacon, Elder Relay, your future self, your unmade
self). Its figures are all past the Industrial Age, so no false prophets and numeric odds.

### The roll

`DoPrestige`, once confirmed, runs `fateBeforePrestige()` and then `rollLastPassage()` when
`lastPassageApplies()` (final epoch, past the Iron gate): an open Reality Tear settles first (see
[No outrunning it](#no-outrunning-it)), and a refusal there stops the prestige before the roll.
One `ge.rng` `Float64()` is always drawn, so the stream's shape does not depend on the odds or on
an invite. The chance is `catastropheOutlook().Probability`, the passage formula
`(1 - good chance) × catastropheChanceOnBadRoll (0.30) × appeaseMultiplierOf(lastPassageThread())`,
i.e. 18% / 15% / 12% by faith band, certain when `catastropheInvited` is armed (consumed on a hit).
Prestige from any earlier epoch never rolls.

- **Miss:** verdict `spared`, the RunEnding line, and prestige completes.
- **Hit:** `triggerLastPassage` sets `pendingLastPassage` and publishes a bus event for the toast.
  Prestige does not complete.

### Pending state

`pendingLastPassage` blocks only `DoPrestige` (`lastPassageBlockErr`); `AdvanceAge`, building
and everything else carry on. The dashboard shows the choice in the catastrophe modal
(`✦ The Last Passage`), Esc hides it, the status bar carries a `☄ LAST PASSAGE` badge and the bare
`catastrophe` command reopens it. The save list reports it as the pending choice. If a
catastrophe is pending beside it, that one is answered first: `pendingChoiceKey` shows it first and
`resolveLastPassage` refuses ("The Reality Tear came first. Answer it before the Last Passage.").
The dev console's `/lastpassage` (`forceLastPassage`) sets it for testing, final epoch only.

### Endure and Succumb

Both go through `resolveLastPassage` and complete the prestige; the level rises either way.

| Choice | Points from the run | Other effect |
|--------|---------------------|--------------|
| Endure | `floor(full × keep)`, minimum 0, where `keep` = `lastPassageKeepFrac[brace]` = 0.50 / 0.70 / 0.85 at Brace 0 / 1 / 2 | Verdict `vindicated` (`fulfilled` if invited) |
| Succumb | 0 | Sets `cosmicLegacy` |

Brace on the Last Passage's thread changes only the points share; nothing survives prestige to
destroy. Brace on the Reality Tear's thread works like any era's.
`recordLastPassageOutcome` appends a `catastropheHistory` entry carrying the Endured / Succumbed
markers, so the Stats tallies count it.

**Cosmic Legacy.** A one-time permanent flag. `cosmicLegacyModifiers()` emits
`production_all +CosmicLegacyProductionBonus` (0.10) into the resolver as source `cosmic_legacy`,
derived from the flag and never stored as a bonus value, like the derived Succumb research bonus
(`legacy`). It survives every prestige and Succumb; only a full wipe clears it. With the flag
held, the modal disables Succumb ("You already carry the Cosmic Legacy. Succumb is closed to
you."), so Endure is the only choice.

**Invite** on the Last Passage's thread arms `catastropheInvited` for the next prestige. It is the deliberate
path to the Cosmic Legacy; Appease is refused afterwards, Brace still raises the Endure share.

### The run's last lines

The prestige reset clears the log. `runEndingLines` writes the verdict, the Endure / Succumb
lines and one RunEnding flavor line aside (drawn in that order from `ge.rng`) and the new run's
log starts with them. `logRunEnding` runs on every prestige, from any age: the line is in the
voice of the age the run ended in, with the age's harbinger as the subject in the Cosmic Era.

### Costs

Fixed across the four Cosmic ages, level 2 at double, and shared by both Cosmic threads (each keeps
its own levels):

| Action | Level 1 | Basis |
|--------|---------|-------|
| Appease | 3.1B faith + 48B culture | a quarter of `config.FlowIncome` over the Interstellar, Galactic and Quantum Ages at their targets |
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

A Last Passage thread parked behind the Reality Tear's is saved as `ParkedHarbinger` (see
[Persistence and resets](#persistence-and-resets)).

### Tuning constants

- `game/last_passage.go`: `LastPassageKeep` (0.50, unbraced Endure share),
  `lastPassageKeepFrac` (0.50 / 0.70 / 0.85 by Brace), `CosmicLegacyProductionBonus` (0.10).
- `game/harbinger.go`: `harbingerAppeaseAges` (the Cosmic ages Appease is priced over).

---

## Epoch Event Pools

The existing universal events (drought, good harvest, plague, festival, trade windfall, etc.) remain
active in all epochs and scale their magnitude to current epoch resource production rates. Each epoch
adds 5 exclusive events that only appear during that epoch.

### Stone Era: Exclusive Events

1. **Sacred Grove Discovered**: an ancient forest is found; knowledge production +50% for 48 ticks
2. **Wandering Tribe**: nomad group joins; +80 pop, +6 workers across primitive classes
3. **Stone Idol**: workers uncover a carved idol; faith +300, morale +25% for 60 ticks
4. **Cave Paintings**: ancient art discovered; culture +500, +1 culture/tick permanently from next
   culture building built
5. **Bone Tools**: innovation event; wood production +100% for 30 ticks; one free early tech

### Iron Era: Exclusive Events

1. **The Spreading Plague**: population -15%, workers -10%; faith above 60% cap halves the losses
2. **Barbarian Horde**: military buildings take 40% damage unless military production > threshold
3. **Silk Road Opens**: all trade route gold income +80% for 90 ticks
4. **Philosopher's Academy**: knowledge production doubled for 60 ticks; one free knowledge tech
5. **Bronze Uprising**: production halved for 36 ticks unless faith > 40% cap

### Steel Era: Exclusive Events

1. **Industrial Accident**: 5 random Engineering buildings destroyed; surrounding output -20% for 48 ticks
2. **Worker Strike**: factory output halved until faith restored to >50% or 72 ticks elapse
3. **Colonial Gold Rush**: gold production ×3 for 60 ticks; 10 free Settler workers added
4. **Railroad Connection**: all trade route income +50% permanently until next epoch transition
5. **Colonial Revolt**: 20% of Colonial-era buildings damaged; gold income -30% for 60 ticks

### Electric Era: Exclusive Events

1. **Nuclear Test Fallout**: food production -30% for 120 ticks; faith -20% (public fear)
2. **Oil Crisis**: all electricity-dependent buildings offline for 36 ticks; oil building costs ×2 for 60 ticks
3. **Space Race Ignition**: knowledge +150% for 90 ticks; one free tech in knowledge tree
4. **Cold War Tension**: military production +60%, worker food drain +20% for 120 ticks (war footing)
5. **Power Grid Failure**: all Electric Era buildings offline for 18 ticks, then +50% electricity output
   on restoration (systems surge)

### Digital Era: Exclusive Events

1. **The Great Data Breach**: data -60%, crypto -30%; Hacker worker output -50% for 48 ticks
2. **AI Anomaly**: 3 random buildings swap their output resources for 60 ticks (chaos event)
3. **Biotech Breakthrough**: food production +100% for 90 ticks; biotech research branch available
4. **Silicon Drought**: titanium building costs +50% for 60 ticks; titanium production +20%
5. **Viral Memetic Storm**: culture and faith both halved for 48 ticks, then doubled for 48 ticks
   (net neutral but timing matters for players near faith thresholds)

### Neon Era: Exclusive Events

1. **Corporate War**: plasma production -40%, dark_matter +40%; military output +60% for 90 ticks
2. **Augmentation Rebellion**: 15% of workers revolt; food drain -10% permanently (fewer augmented workers)
3. **Fusion Breakthrough**: plasma production ×3 for 120 ticks; one free Neon Era tech
4. **Black Market Surge**: gold income +150% for 60 ticks; faith -20% (moral cost of dealings)
5. **Consciousness Upload**: 12% of population digitized; housing freed, knowledge +60% permanently

### Cosmic Era: Exclusive Events

1. **Alien Signal Received**: knowledge + data ×4 for 120 ticks; Xenology research branch unlocked
2. **Stellar Phenomena**: dark_matter ×2 for 60 ticks; antimatter disrupted -50% for 30 ticks
3. **Reality Distortion**: 5 random buildings swap output resources for 60 ticks (terrifying at scale)
4. **Quantum Resonance**: quantum_flux ×5 for 30 ticks (spike fills storage; plan for it)
5. **Dimensional Rift**: all production halted for 12 ticks, then ×3 for 60 ticks (terrifying/rewarding)

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
| Challenging bad epoch events (major epoch roll) | 8 | One fires per epoch transition (if bad) |
| Catastrophe events | 7 (6 reachable) | Iron Era on: at most one per era per run (a fated doom, inside the era) |
| **Total** | **88** | |

Note: The 10 good + 8 bad events are **epoch transition events**, separate from the regular
random event pool. They fire exactly once per epoch transition, replacing the normal "age advance"
announcement. Catastrophes are not transition events: they strike at a fated moment inside an era.

---

## UI: Epoch Badge

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
[yellow]✦ The Steel Era Dawns: The Age of Iron gives way to industry and empire.[-]
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
        ge.rollFate()                 // the new era's hidden fate (fate.go)
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
        ge.rollChallengingEpochEvent()   // never a catastrophe
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
Epoch transition events (good/bad) are a separate pool, fired once per epoch transition, NOT
drawn from the regular random event pool.

### Triggering (implementation)

In game/catastrophe.go:
- `triggerCatastrophe(epochKey, source)` sets `pendingCatastrophe`, appends an `EpochEventRecord`
  with `Outcome: "pending"`, and publishes `EventEpochEventFired` (`event_type: "catastrophe"`) so
  the dashboard toast fires. Sources: a fated doom's strike (`fateStrike`, source `invited` when
  the harbinger was invited) or the dev console's `forceCatastrophe` (`/catastrophe`, via
  `game.DevConsoleCommand`).
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
epoch. Catastrophes are limited to one per era per run by the era's fate (`GameSave.Fate`, see
[Fated Dooms](#fated-dooms)). `CatastropheFired` was
written by early builds of the overhaul; it stays in the struct only so those saves still verify.

---

## Decision Log

Epoch-system decisions. The project-wide log is in `README.md`.

| Date | Decision | Rationale |
|------|----------|-----------|
| 2026-09-30 | Catastrophes are fated in secret on entering an era (27%, Iron to Cosmic) and strike at a random moment across the era's expected length; a transition rolls only a good or challenging event. A harbinger comes only for a fated doom (or a false prophet, now only in an era with nothing fated, before the Industrial Age), 20-60% of the current age's target before the strike, and the doom cannot be outrun by advancing. The strike rolls 90/75/60% by faith (the passage chance ×5), ×0.6 per Appease | Adam's "randomize the when" (approved 2026-09-29): with catastrophes only at epoch boundaries and a harbinger every epoch, players could learn the schedule, which killed the dread and the idle randomness. A harbinger now means a doom is coming, so answering it matters; what it says sharpens with the figures (no timing before the Oracle, age or era from her, odds from the Newsboy); every window follows the age pacing targets, so a pacing change carries through in one place; `FateChance` is calibrated so a first run meets about as many catastrophes as before (0.73 expected at low faith, was 0.72). The Reality Tear stays, as the Cosmic Era's fated doom beside the Last Passage (decided 2026-10-01, so all six catastrophes remain for the planned badges) |
| 2026-09-26 | Prestige is the Cosmic Era's passage (the Last Passage), rolled once at confirmed prestige with the epoch odds; pending blocks only prestige; Endure keeps 50/70/85% of the run's points by Brace; Succumb grants a one-time Cosmic Legacy (+10% `production_all`, derived from a flag) | The Cosmic Era's four harbingers had nothing to warn of because the epoch has no transition out; prestige is the only passage it has, and points are what a prestige can lose, so Endure and Brace act on them |
| 2026-09-26 | Harbinger replaces invoke; epoch-long threads with the speaker changing each age; false prophets rolled once per thread (Stone, Iron, Steel Era); Appease x0.6 per level; Brace tiers (15%/30%, 10%/45%); costs priced off the passage | Choosing a catastrophe fits better as an answer to a warning than as a bare command; a thread gives the warning time to matter and uses 18 figures instead of 6; false prophets make early warnings worth doubting until the odds are printed; passage pricing keeps the price the same in every age so paying early is not a discount; x0.6 still lowers the odds after the worst faith-band drop the price can cause; Brace gives Endure-minded players something to buy without touching the odds |
