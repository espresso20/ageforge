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
  Harbinger reads it at arrival and for its panel.

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

Built 2026-09-26. Code: `game/harbinger.go` (mechanic), `config/harbingers.go` (the 22-entry
roster: `Name`, `Description`, `AppeaseLabel`, `BraceLabel`, `InviteLabel`, plus the derived
`ForecastPrecision` and `FalseProphetChance`), `flavor/` (arrival, warning, action and outcome
lines). Player docs: `site/docs/harbinger.md`.

### Arrival

- `advanceAge` calls `maybeHarbingerArrive()` after `detectEpochTransition` and `fireAwakening`.
  `restoreHarbingerState` calls it again after a load.
- It arrives only when the player stands in the **last age of the current epoch**, no harbinger
  is live, `harbingerArrived[currentEpoch]` is unset, and `catastropheOutlook().Possible` is true
  (the next epoch is past the Iron gate and has not rolled this run). So it visits the Bronze,
  Medieval, Industrial, Atomic, Digital and Space Ages only. The Cosmic Era is last and gets none.
  The other 16 roster entries exist but never appear under this rule.
- Non-blocking: a log entry plus the arrival and warning lines, `EventHarbingerArrived` on the bus
  (toast), and a status-bar badge while `GameState.Harbinger` is non-nil. It never expires.
- The dev console's `/harbinger` (`summonHarbinger`) skips the last-age and once-per-epoch checks
  but still needs `Possible`.

### False prophets and precision

- `FalseProphetChance` is `(8 - ageIndex) / 64` before the Industrial Age (index 8) and 0 from
  it: 6/64 at Bronze, 3/64 at Medieval, 0 for every later visitor.
- A false prophet's `AnnouncedTier` is medium or high (rolled), whatever the real tier. The
  displayed tier is `announced + (liveReal - ArrivalRealTier)`, clamped to low..high, so Appease
  and faith-band changes move it like a real one. `HarbingerView` carries no false-prophet flag.
- `ForecastPrecision` is numeric from the Industrial Age: the panel prints
  `HarbingerView.Probability` (after Appease). Vague ages show the tier only. Numeric and
  false-prophet ages never overlap, by construction of the two curves.

### Actions

Costs are fractions of the **current storage cap** (`harbingerLevelCost`), rounded up; level 2
costs double. Locked resources and resources with zero storage are skipped. Caps track the
economy at every age, never read zero (rates can), and bound what the next age's requirements
can be, so a fraction of the cap is always a slice of what it takes to advance. The harbinger
never expires, so an idle player can let storage fill and pay later.

| Action | Cost (level 1) | Effect | Cap |
|--------|----------------|--------|-----|
| Appease | 15% of faith cap + 15% of culture cap (culture once unlocked, Classical) | Multiplies the real catastrophe chance by `harbingerAppeaseFactor` = 0.6 per level (0.36 at 2) | 2 |
| Brace | 12% of the cap of each resource in the next age's `ResourceReqs`, minus faith and culture (falls back to the epoch's primary resource) | Endure destroys 15% / 10% of non-wonder buildings and keeps 30% / 45% of resources (unbraced 20% / 15%) | 2 |
| Invite | free | Sets `HarbingerSave.Invited` and arms `catastropheInvited`; Appease refuses afterwards, Brace does not | once |

- **Appease** is applied through `harbingerAppeaseMultiplier()`, which scales
  `catastropheChanceOnBadRoll` in both `rollEpochEvent` (the real roll) and `catastropheOutlook`
  (the displayed odds), so the two cannot disagree. It lowers the real odds even for a false
  prophet. Spending 15% of the faith cap can drop the faith band; worst case the base chance goes
  from 12% to 18% (x1.5), and x0.6 still leaves 0.9x, so each level always lowers the odds.
- **Brace** lives on `HarbingerSave.BraceLevel` until resolution, then moves to
  `ge.pendingBraceLevel` if the catastrophe came. `Endure` reads and clears it
  (`braceDestroyPct`, `braceKeepFrac`). Succumb ignores it. Keeping it on the pending catastrophe
  means Esc-then-Endure and save/load both keep the discount.

### Resolution

`detectEpochTransition` calls `resolveHarbinger(newEpoch, came)` right after `rollEpochEvent`,
where `came` is "no catastrophe was pending before the roll and one is pending for this epoch
now". Verdicts (`HarbingerRecord.Outcome`):

| Outcome | Condition |
|---------|-----------|
| `fulfilled` | came and invited |
| `vindicated` | came, not invited (a false prophet here gets a note that the warning was invented) |
| `spared` | did not come, true harbinger |
| `discredited` | did not come, false prophet |

It logs the verdict (and the braced Endure numbers if relevant), draws one flavor line, appends a
`HarbingerRecord` to `harbingerHistory` and clears the live harbinger.

### Determinism

All draws come from the seeded `ge.rng` under the write lock. Arrival draws, in order: the
false-prophet `Float64()` (always, whatever the age's chance, so the stream's shape does not
depend on it), then `Intn(2)` for the fake tier only if the harbinger is false, then the arrival
and warning lines from the engine's flavor `Stream`. Each action and the resolution draw one
flavor line.

### Persistence and resets

`GameSave` fields, all `omitempty` (old saves load clean):

```go
Harbinger          *HarbingerSave    `json:"harbinger,omitempty"`
HarbingerArrived   map[string]bool   `json:"harbinger_arrived,omitempty"`
CatastropheInvited bool              `json:"catastrophe_invited,omitempty"`
PendingBraceLevel  int               `json:"pending_brace_level,omitempty"`
HarbingerHistory   []HarbingerRecord `json:"harbinger_history,omitempty"`
```

On load, levels are clamped to 2, an unknown roster age drops the harbinger, and
`pendingBraceLevel` is zeroed unless a catastrophe is pending. `clearHarbingerRun` (live
harbinger, arrivals, invite, pending Brace) runs on Succumb, prestige and reset.
`harbingerHistory` follows `epochEventHistory`: kept by Succumb, cleared by prestige.

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
| 2026-09-26 | Harbinger replaces invoke; false prophets pre-industrial; Appease x0.6 per level; Brace tiers (15%/30%, 10%/45%) | Choosing a catastrophe fits better as an answer to a warning than as a bare command; false prophets make early warnings worth doubting until the odds are printed; x0.6 per level with storage-cap pricing always lowers the odds even when the faith spend drops a band; Brace gives Endure-minded players something to buy without touching the odds |
