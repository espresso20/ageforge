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
5. **Catastrophe is offered** — the Civilizational Catastrophe modal appears (see below)
6. **UI epoch badge updates** — the epoch badge near the age indicator changes color and icon
7. **New worker domains may unlock** — Hacker (Digital Era), Astronaut (Neon Era)

---

## Civilizational Catastrophe System

At each epoch boundary, the game presents a **Civilizational Catastrophe** — a world-scale event
that tests your civilization. You are given a choice: endure and survive, or let civilization fall
and rise again, stronger.

This system is separate from the regular prestige system. Both can be used in the same run and
their bonuses stack.

### The Choice Modal

```
╔══════════════════════════════════════════════════════════╗
║  ☄ THE GREAT METEOR                                      ║
║  A celestial body has struck your settlement.            ║
║  The sky burns. Your people scatter.                     ║
╠══════════════════════════════════════════════════════════╣
║                                                          ║
║  [ENDURE]                    [SUCCUMB]                   ║
║                                                          ║
║  Weather the catastrophe.    Let civilization fall.      ║
║                                                          ║
║  Your people survive, but:   Everything resets. But:     ║
║  • 30% buildings destroyed   • Epoch Legacy Bonus        ║
║  • Resources wiped to 15%    • 8 Ruins carry forward     ║
║  • 25% workers lost          • Ancient Knowledge kept    ║
║  • Building costs +20%       • Catastrophe title earned  ║
║    for 72h (reconstruction)  • Stone Legacy: +20%        ║
║                                wood/stone production     ║
║  Research kept. Progress      All research kept as       ║
║  kept. The scars remain.      "Ancient Knowledge."       ║
║                               History remembers you.     ║
╚══════════════════════════════════════════════════════════╝
```

The player can also select a third option: **[DEFER]** — the catastrophe is skipped entirely.
No bonuses, no consequences. The option to defer can only be used once per catastrophe; if deferred,
the catastrophe will not re-appear in this run (you've committed to surviving without it).

### ENDURE — Consequences and Rewards

**Immediate damage (all applied on choice):**
- 30% of buildings randomly destroyed (shown to player as a list: "73 Farms lost, 12 Monasteries lost")
- All resources drop to 15% of current stored amount
- 25% of workers removed (distributed evenly across all domains)
- Research, milestones, wonders, and age unlocks are fully preserved

**Lasting consequences (fade after 72 real hours):**
- All building costs +20% (reconstruction premium — materials are scarce)
- Worker food drain +10% (survivors need more care in the aftermath)
- Random events hit 20% harder during recovery window

**Permanent rewards (never fade):**
- "Survived" marker added to epoch badge (visual distinction in UI)
- Unlock **Reconstruction** tech branch for this epoch (5 epoch-specific recovery techs)
- Unique Wonder unlocked: **Monument to the Fallen** — costs nothing to build, provides massive
  culture (+2,000) + faith (+500/tick permanently) + morale bonus
- Title earned: "The Undying [Epoch]" (e.g., "The Undying Iron Lords") — shown in Stats tab

### SUCCUMB — Civilizational Reset

**Immediate reset:**
- Full civilization reset: all buildings → 0, all resources → 0, all workers → 0
- Age resets to Primitive; epoch resets to Stone Era
- **8 Ruins carry forward**: random buildings from your previous civilization remain as Ruins.
  Ruins produce at 50% output with no workers assigned. Cannot be rebuilt if destroyed. Cannot
  be built again (they're relics of the fallen age). Shown with a ☒ marker in the Economy tab.
- All current epoch's research carries as **Ancient Knowledge**: permanent +25% research speed
  for that epoch's tech tree in all future runs
- Civilization history log gains a lore entry for the catastrophe

**Permanent rewards (stack across runs, never lost):**
- **Epoch Legacy Bonus**: epoch-specific production multiplier that applies to every future run
- **Catastrophe Title**: recorded in civilization history permanently
- **Faster Return**: all techs from 2 epochs below the catastrophe point auto-complete in next run
- **Exclusive Starting Event**: a unique event only available to civilizations that experienced
  this specific catastrophe — appears in the first 10 ticks of the new run

### The 7 Catastrophes

| Epoch | Catastrophe Name | Flavor Text |
|-------|----------------|-------------|
| Stone Era | **The Great Meteor** | A celestial body strikes your settlement. The sky burns. |
| Iron Era | **The Great Plague** | A devastating plague sweeps your cities. The streets are silent. |
| Steel Era | **The World War** | Industrial warfare tears civilization apart. The factories are ash. |
| Electric Era | **The Nuclear Exchange** | Nations unleash the atom. Cities become glass. |
| Digital Era | **The Great Hack** | Every system falls silent. The AIs turn on their creators. |
| Neon Era | **Corporate Armageddon** | The megacorps end the world with a fusion bomb. |
| Cosmic Era | **The Reality Tear** | Exotic matter destabilizes spacetime. Reality cracks open. |

### SUCCUMB Legacy Bonuses by Epoch

| Epoch Catastrophe | Epoch Legacy Bonus | Exclusive Unique |
|-------------------|--------------------|-----------------|
| The Great Meteor | +20% wood + stone production (permanent) | **Meteor Fragment** wonder available |
| The Great Plague | +20% iron production; Ancient Immunity passive (events 15% less severe) | **Plague Doctor** worker class unlocked |
| The World War | +25% steel + coal production; War Doctrine tech | **Armistice Monument** wonder |
| The Nuclear Exchange | +25% electricity + uranium production; Fallout Shelter tech | **Nuclear Vault** building |
| The Great Hack | +30% data + titanium production; Ghost Protocol tech | **Dead Drop Network** building |
| Corporate Armageddon | +30% plasma + dark_matter production; Phoenix Protocol tech | **Corporate Ruins** wonder |
| The Reality Tear | +35% antimatter + quantum_flux production; Reality Anchor tech | **Scar in Reality** wonder |

Legacy Bonuses stack across runs. A civilization that has succumbed to all 7 catastrophes receives
all 7 legacy bonuses simultaneously and has access to all 7 exclusive buildings and wonders.

### Catastrophe vs Regular Prestige

| | Regular Prestige | Catastrophe Succumb |
|--|-----------------|---------------------|
| Trigger | Player-initiated anytime | Offered at epoch boundary |
| Reset scope | Full | Full + 8 Ruins carry forward |
| Bonus pool | Prestige upgrade tree (9 slots) | Epoch Legacy Bonuses (7 slots) |
| Lore / narrative | None | Civilization history records it |
| Repeatable | Yes, unlimited | Once per epoch per run |
| Stack with each other | Yes | Yes — fully compatible |

A player who has both prestiged and succumbed to catastrophes holds both bonus types simultaneously.
They are designed to reward different play styles: efficiency players use prestige; narrative players
build catastrophe histories.

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
| **Total** | **63** | |

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
- `GameEngine` tracks: `currentEpoch string`, `catastropheOffered map[string]bool`, `legacyBonuses map[string]bool`,
  `catastropheHistory []string` (for civilization log)

### Epoch Transition Trigger

```
advanceAge() {
    // ... existing age advance logic ...
    newEpoch := epochForAge(ge.currentAge)
    if newEpoch != ge.currentEpoch {
        ge.currentEpoch = newEpoch
        ge.bus.Publish(EpochAdvanced, newEpoch)
        // trigger catastrophe offer UI
        if !ge.catastropheOffered[newEpoch] {
            ge.pendingCatastrophe = newEpoch
        }
    }
}
```

### Event Pool Filtering

EventManager's random event trigger should filter the available event pool:
```
candidateEvents = universalEvents + epochExclusiveEvents[currentEpoch]
```
Previous epoch exclusive events are permanently removed from the pool on epoch transition.

### Catastrophe Save State

In `GameSave`:
```go
LegacyBonuses      map[string]bool   `json:"legacy_bonuses,omitempty"`
CatastropheHistory []string          `json:"catastrophe_history,omitempty"`
SurvivedEpochs     map[string]bool   `json:"survived_epochs,omitempty"`
PendingCatastrophe string            `json:"pending_catastrophe,omitempty"`
Ruins               []RuinState      `json:"ruins,omitempty"`
```

Ruins persist across runs (they're part of your civilization's identity). Legacy bonuses are
permanent and never removed.
