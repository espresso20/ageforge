# AgeForge — Age Transition System

## Overview

When a player advances to a new age, a **transformation pass** automatically upgrades their
civilization. Buildings transform to their next-tier equivalent, workers rename and gain new
stats, and new construction options unlock. The player never manually migrates — the game
handles it as part of the age advance event.

This design means:
- The Economy tab always shows **only current-age buildings** (no overflow, no age grouping)
- Workers always reflect the current age's class names
- Players feel the civilization genuinely advancing, not just unlocking more buildings on top

---

## Transformation Pass (on age advance)

```
1. For each building in player inventory:
      - Look up the building's lineage chain
      - If a next-tier entry exists for the new age: transform in-place
          • building key, name, stats update
          • count is preserved exactly
          • assigned workers remain (still valid for the same domain)
      - If no next-tier entry: building becomes "legacy" (functional, not upgradeable)

2. For each worker class in player inventory:
      - Find the new-age tier for their domain
      - Rename the class (Tribesman → Laborer)
      - Apply new food cost and output multiplier immediately
      - Count is preserved — no re-recruitment needed

3. Unlock new-age buildings for fresh construction

4. Display "Age Advance" summary screen:
      - List all buildings that transformed (old name → new name, count)
      - List all worker classes that renamed
      - List newly unlocked buildings
      - Net production change (before/after comparison)
```

---

## Building Lineages

A lineage is a chain of age-specific incarnations of the same production role. Buildings in
a lineage share a domain, purpose, and worker type — only their name, stats, and age tier differ.

**Rules:**
- Each lineage has at most one entry per age
- A building belongs to exactly one lineage (or is "ageless" — wonders, storage)
- Lineage advancement is automatic on age advance
- If a lineage has no entry for the new age, the building becomes legacy

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

#### Raw Production (food / wood / stone — worker determines which resource flows)
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

#### Storage (no lineage transformation — storage buildings are age-specific, standalone)
Storage buildings do **not** transform on age advance. They stack additionally.
A player keeps their stashes AND can build Stone Age storage pits on top.
This is intentional: storage growth is cumulative and should feel like infrastructure investment.
See economy.md Law 1 (Storage Covenant) for capacity requirements per age.

#### Wonders (ageless — never transform)
Wonders are permanent landmarks. A Great Monolith built in Stone Age stays a Great Monolith
in the Quantum Age. They do not transform, cannot be rebuilt, and are never demolished.
This makes wonders feel like historical monuments rather than upgradeable units.

---

## Legacy Buildings

Buildings with no next-tier lineage entry become **legacy** on age advance:
- Still produce at their current stats
- Cannot be built again (grayed out in Economy tab with legacy tag)
- Cannot be upgraded
- Do not disappear — they remain as long-standing infrastructure
- Example: `firepit` (stone age) has no bronze equivalent → stays as legacy on bronze advance

Legacy buildings fade in relevance naturally (their fixed stats fall behind the new age's
production curve) without punishing the player by removing them.

---

## Worker Transformation

On age advance, all worker classes in the player's workforce rename and restat:

- The **count** is preserved exactly
- The **food cost per worker** updates to the new tier's value immediately
  (could be a net increase — player may need to adjust food production)
- The **output multiplier** updates to the new tier value
- **Assignment is preserved** — workers stay in whatever buildings they were in
- The domain stays the same — a worker assigned to a building stays assigned

**Example:**
Stone Age advance to Bronze Age:
- 200 Tribesmen (Raw Materials) → 200 Laborers
- Food drain: 200 × 0.15 → 200 × 0.22 (net +14 food/tick drain)
- Output: each worker produces 2× more per assignment slot

The player sees a net production gain but also higher food cost. This creates a moment of
"do I have enough food production for Bronze Age Laborers?" — a satisfying decision point.

---

## Age Advance UI Summary Screen

When the age advances, show a modal/toast sequence:

```
╔══════════════════════════════════════════════╗
║  ✦ Bronze Age Dawns ✦                        ║
╠══════════════════════════════════════════════╣
║  Buildings Transformed:                       ║
║    75 Gathering Camps → Farms               ║
║     3 Woodcutter's Camps → Farms            ║
║    23 Altars → Scriptoriums                 ║
║     5 War Camps → Barracks                  ║
║                                              ║
║  Workers Upgraded:                           ║
║    280 Tribesmen → Laborers                 ║
║      5 Elders → Scribes                     ║
║                                              ║
║  Unlocked for Construction:                  ║
║    Smithy, Market, Library, House, Warehouse ║
║                                              ║
║  ⚠ Food drain +14/tick — assign more farmers ║
╚══════════════════════════════════════════════╝
```

---

## Implementation Notes

- `BuildingDef` needs a `LineageKey string` field (e.g. `"raw_production"`, `"housing"`)
  and a `LineageTier int` (0 = primitive, 1 = stone, etc.)
- `AgeDef` triggers transformation pass in the engine's `advanceAge()` function
- Save format: buildings saved by key; on load, if a key is not recognized (old save with
  pre-transformation building names), the engine maps via a migration table
- The transformation pass fires synchronously inside `advanceAge()` before the tick resumes
