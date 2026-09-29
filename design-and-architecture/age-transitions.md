# AgeForge: Age Transition System

## Overview

When a player advances to a new age, the engine does not replace buildings on its own. It
offers each lineage building that has a next tier in the new age as a **pending upgrade**,
and the player decides when to pay for it with the `upgrade` command. Workers rename and take
the new age's stats at once, and the new age's buildings unlock for construction.

This design means:
- Old-tier buildings keep producing until the player upgrades them, so an age advance never
  costs production by itself
- Upgrading costs resources, so moving a large stock of buildings to the new tier is a
  spending decision rather than a free swap
- Workers always reflect the current age's class names

---

## Age Advance Pass (`advanceAge` in game/engine.go)

```
1. For each owned building with a lineage (wonders and storage are skipped):
      - Look up the lineage's next tier for the new age (config.BuildingNextTierForAge)
      - If one exists:
          • record a pending upgrade old key → new key (BuildingManager.SetPendingUpgrade)
          • mark the old building legacy: it keeps producing but can no longer be built
          • log a line naming the upgrade and the `upgrade <old key>` command
      - Any other owned lineage building whose lineage now has a higher unlocked tier is
        also marked legacy

2. Workers.SetAge(newAge): worker classes rename and take the new tier's food cost and
   output multiplier; counts and assignments are kept

3. New-age buildings unlock for construction

4. The age splash (ui/age_splash.go) lists the upgrades now on offer
   (AgeAdvanceSummary.BuildingsTransformed) and the buildings that went legacy
```

## Running Upgrades (`upgrade` command)

- `upgrade` with no arguments lists every pending upgrade: old key, new key, copies
  available, cost, and whether the player can afford it now
- `upgrade <building> [n|all]` converts n copies (default all) to the new tier
  (`GameEngine.UpgradeBuilding`)
- Cost per copy, per resource: the new copy's cost at the current new-tier count (with
  build_cost discounts) minus a refund of half the old copy's undiscounted cost, floored at
  zero (`BuildingManager.UpgradeCost`)
- The new tier's MaxCount still applies; the command upgrades as many copies as fit
- When every copy moves, worker assignments move to the new key. After a partial upgrade,
  workers beyond what the remaining old copies can hold follow the upgraded copies while the
  new building has room; the rest return to the idle pool
- Once no old copies remain, the pending upgrade is cleared
- Trying to build a legacy building that has a pending upgrade returns an error that points
  the player at `upgrade <key>`
- Pending upgrades and legacy flags are saved with the game

---

## Building Lineages

A lineage is a chain of age-specific incarnations of the same production role. Buildings in
a lineage share a domain, purpose, and worker type; only their name, stats, and age tier differ.

**Rules:**
- Each lineage has at most one entry per age
- A building belongs to exactly one lineage (or is "ageless": wonders, storage)
- On age advance, a lineage's next tier is offered as a pending upgrade; the player runs it
- If a lineage has no entry for the new age, the building stays as it is and no upgrade is offered

### Lineage Definitions

#### Housing
| Age | Building | Pop per building |
|-----|----------|-----------------|
| Primitive | Hut | +10 |
| Stone | Longhouse | +25 |
| Bronze | House | +50 |
| Iron | Townhouse | +80 |
| Classical | Villa | +120 |
| Medieval | Manor | +200 |
| Renaissance | Estate | +350 |
| Colonial | Settlement | +600 |
| Industrial | Tenement | +1,000 |
| Victorian | Row House | +1,800 |
| Electric | Apartment Block | +3,200 |
| Atomic | Housing Project | +5,500 |
| Modern | Apartment Tower | +10,000 |
| Information | Smart Condo | +18,000 |
| Digital | Megaplex | +32,000 |
| Cyberpunk | Arcology Pod | +55,000 |
| Fusion | Fusion Habitat | +100,000 |
| Space | Orbital Ring | +180,000 |
| Interstellar | Generation Ship | +320,000 |
| Galactic | Dyson Habitat | +600,000 |
| Quantum | Reality Fold | +1,000,000 |

#### Raw Production (food / wood / stone; worker determines which resource flows)
| Age | Building | Worker Capacity | Base Rate (20% floor) |
|-----|----------|----------------|----------------------|
| Primitive | Gathering Camp | 3 | 0.1/tick per worker |
| Stone | Forager Post | 4 | 0.1/tick per worker |
| Bronze | Farm | 5 | 0.1/tick per worker |
| Iron | Ironworks Camp | 5 | 0.1/tick per worker |
| Classical | Field Estate | 6 | 0.1/tick per worker |
| Medieval | Serfdom | 6 | 0.1/tick per worker |
| Renaissance | Workshop | 7 | 0.1/tick per worker |
| Colonial | Plantation | 8 | 0.1/tick per worker |
| Industrial | Factory | 10 | 0.1/tick per worker |
| Victorian | Mill | 10 | 0.1/tick per worker |
| Electric | Processing Plant | 12 | 0.1/tick per worker |
| Atomic | Automated Factory | 12 | 0.1/tick per worker |
| Modern | Industrial Complex | 15 | 0.1/tick per worker |
| Information | Smart Factory | 15 | 0.1/tick per worker |
| Digital | Nano-Factory | 18 | 0.1/tick per worker |
| Cyberpunk | Augmented Works | 20 | 0.1/tick per worker |
| Fusion | Fusion Forge | 20 | 0.1/tick per worker |
| Space | Orbital Platform | 25 | 0.1/tick per worker |
| Interstellar | Asteroid Mine | 25 | 0.1/tick per worker |
| Galactic | Stellar Processor | 30 | 0.1/tick per worker |
| Quantum | Reality Harvester | 30 | 0.1/tick per worker |

> **Note:** Base rate per worker scales with the worker's OutputMultiplier (see workers.md).
> The 0.1/tick figure above is for Tier-1 workers. A Serf (Medieval, 32× multiplier) produces
> 3.2/tick per slot in the same building.

#### Knowledge Production
| Age | Building | Worker Capacity |
|-----|----------|----------------|
| Primitive | Altar | 2 |
| Stone | Standing Stones | 2 |
| Bronze | Scriptorium | 3 |
| Iron | Agora | 3 |
| Classical | Forum | 4 |
| Medieval | Cathedral | 4 |
| Renaissance | University | 5 |
| Colonial | Academy | 5 |
| Industrial | Institute | 6 |
| Victorian | Museum | 6 |
| Electric | Laboratory | 7 |
| Atomic | Research Center | 7 |
| Modern | Think Tank | 8 |
| Information | Innovation Hub | 8 |
| Digital | AI Research Lab | 10 |
| Cyberpunk | Neuro-Lab | 10 |
| Fusion | Theoretical Institute | 12 |
| Space | Space Observatory | 12 |
| Interstellar | Xenology Center | 15 |
| Galactic | Cosmic Library | 15 |
| Quantum | Reality Institute | 20 |

#### Military
| Age | Building | Worker Capacity |
|-----|----------|----------------|
| Primitive | Hunting Lodge | 3 |
| Stone | War Camp | 4 |
| Bronze | Barracks | 5 |
| Iron | Legion Fort | 6 |
| Classical | Military Academy | 6 |
| Medieval | Castle Keep | 7 |
| Renaissance | Fortress | 7 |
| Colonial | Fort | 8 |
| Industrial | Military Base | 10 |
| Victorian | Garrison | 10 |
| Electric | Command Post | 12 |
| Atomic | Bunker | 12 |
| Modern | Special Ops HQ | 14 |
| Information | Cyber Command | 15 |
| Digital | Drone Hub | 16 |
| Cyberpunk | Combat Augmentation Center | 18 |
| Fusion | Plasma Command | 20 |
| Space | Space Force Base | 20 |
| Interstellar | Fleet Command | 25 |
| Galactic | Stellar Armada HQ | 25 |
| Quantum | Probability War Room | 30 |

#### Trade & Commerce
| Age | Building | Worker Capacity |
|-----|----------|----------------|
| Primitive | Barter Post | 2 |
| Stone | Trade Camp | 2 |
| Bronze | Market | 3 |
| Iron | Trading Post | 3 |
| Classical | Merchant Quarter | 4 |
| Medieval | Guildhall | 4 |
| Renaissance | Exchange | 5 |
| Colonial | Port | 5 |
| Industrial | Stock Exchange | 6 |
| Victorian | Bank | 6 |
| Electric | Financial District | 7 |
| Atomic | Corporate HQ | 7 |
| Modern | Investment Firm | 8 |
| Information | Venture Hub | 8 |
| Digital | Crypto Exchange | 10 |
| Cyberpunk | Black Market | 10 |
| Fusion | Energy Exchange | 12 |
| Space | Asteroid Market | 12 |
| Interstellar | Galactic Bazaar | 15 |
| Galactic | Stellar Exchange | 15 |
| Quantum | Probability Market | 18 |

#### Engineering / Industry
Unlocks at Bronze Age.
| Age | Building | Worker Capacity |
|-----|----------|----------------|
| Bronze | Smithy | 4 |
| Iron | Ironworks | 5 |
| Classical | Aqueduct | 5 |
| Medieval | Workshop | 6 |
| Renaissance | Mill | 6 |
| Colonial | Dockyard | 7 |
| Industrial | Iron Works | 8 |
| Victorian | Steam Works | 9 |
| Electric | Power Station | 10 |
| Atomic | Nuclear Plant | 11 |
| Modern | Power Grid | 12 |
| Information | Smart Grid | 13 |
| Digital | Neural Grid | 14 |
| Cyberpunk | Augmentation Foundry | 15 |
| Fusion | Fusion Reactor | 18 |
| Space | Launch Complex | 20 |
| Interstellar | Warp Drive Plant | 22 |
| Galactic | Dyson Assembly | 25 |
| Quantum | Reality Forge | 30 |

#### Storage (age-specific, standalone, never upgraded)
Storage buildings are never offered as upgrades on age advance. They stack additionally.
A player keeps their stashes AND can build this age's storage on top.
This is intentional: storage growth is cumulative and should feel like infrastructure investment.
Each age's storage can only be built in that age, so trading a copy in for a capped slot of the
next tier would only lower the most the player can ever store.
See economy.md Law 1 (Storage Covenant) for capacity requirements per age.

#### Wonders (ageless, never upgraded)
Wonders are permanent landmarks. A Great Monolith built in Stone Age stays a Great Monolith
in the Quantum Age. They are never upgraded, cannot be rebuilt, and are never demolished.
This makes wonders feel like historical monuments rather than upgradeable units.

---

## Legacy Buildings

A building becomes **legacy** on age advance when its lineage has a newer tier: either the
tier offered as a pending upgrade or a higher tier already unlocked.
- Still produces at its current stats
- Cannot be built again (the build error points at `upgrade` when an upgrade is on offer)
- Can be upgraded with `upgrade <key>` while a pending upgrade exists
- Does not disappear; copies the player never upgrades stay as long-standing infrastructure

Legacy buildings fade in relevance naturally (their fixed stats fall behind the new age's
production curve) without punishing the player by removing them.

---

## Worker Renames

On age advance, all worker classes in the player's workforce rename and restat:

- The **count** is preserved exactly
- The **food cost per worker** updates to the new tier's value immediately
  (could be a net increase, so the player may need to adjust food production)
- The **output multiplier** updates to the new tier value
- **Assignment is preserved**: workers stay in whatever buildings they were in
- The domain stays the same: a worker assigned to a building stays assigned

**Example:**
Stone Age advance to Bronze Age:
- 200 Tribesmen (Raw Materials) → 200 Laborers
- Food drain: 200 × 0.15 → 200 × 0.22 (net +14 food/tick drain)
- Output: each worker produces 2× more per assignment slot

The player sees a net production gain but also higher food cost. This creates a moment of
"do I have enough food production for Bronze Age Laborers?", a real decision point.

---

## Age Advance UI Summary Screen

The age splash (ui/age_splash.go) shows, among other things, the upgrades now on offer and
the buildings that went legacy. Nothing on it has changed yet: the listed buildings move to
the new tier only when the player runs `upgrade`.

```
── Buildings Transformed ──
  Gathering Camp → Farm        x75
  Altar → Scriptorium          x23
  War Camp → Barracks          x5
── Legacy Buildings ──
  Firepit
```

The section heading still reads "Buildings Transformed" (the `BuildingsTransformed` field of
`AgeAdvanceSummary`), although each line is an offer, not a completed change.

---

## Implementation Notes

- `BuildingDef` carries `LineageKey string` (e.g. `"housing"`) and `LineageTier int`
- `advanceAge()` in game/engine.go records pending upgrades and legacy flags; it does not
  change building counts
- `GameEngine.UpgradeBuilding` does the conversion when the player runs `upgrade`, through
  `BuildingManager.UpgradeCost` and `BuildingManager.PartialTransform`
- `GameEngine.GetAvailableUpgrades` feeds the `upgrade` listing (`cmdUpgrade` in ui/input.go)
- Save format: buildings saved by key, with pending upgrades and legacy flags alongside;
  on load, if a key is not recognized (old save with renamed building keys), the engine maps
  it via a migration table
