# AgeForge — Worker Class System

## Overview

Workers are age-tiered. Every age unlocks a new class of worker for each relevant domain.
Higher-tier workers cost more food per tick but produce significantly more per assignment.
Old-tier workers remain after an age advance but cannot be newly recruited — only the
current age's tier is available for recruitment. This creates a workforce transition dynamic:
do you keep cheap legacy workers or invest in feeding expensive new ones?

## Mechanics

- Each building has a `worker_domain` (which class fills it) and `worker_capacity` (how many)
- Buildings produce at **20% base rate** with no workers assigned
- Each assigned worker of the matching domain adds `worker_output / worker_capacity` to that building
- Full capacity = 100% production. Over-capacity is not possible.
- A Serf assigned to a Gathering Camp (Worker domain building) produces nothing — domains are strict
- Higher-tier workers have a higher `output_multiplier` — a Space Marine does more per slot than a Hunter

## Tier Stats (design targets, to be tuned)

Each tier step:
- Food cost per worker: × 1.5 over previous tier
- Output per worker: × 2.0 over previous tier
- Net efficiency per food spent: × 1.33 improvement per tier (worth upgrading)

---

## Worker Class Tables

### Raw Materials Domain
Fills: gathering_camp, woodcutter_camp, stone_pit, farm, quarry, mine, coal_mine, oil_well,
       plantation, fusion_reactor (raw fuel), antimatter_plant

| Age | Class Name | Food/tick | Output Multiplier |
|-----|-----------|-----------|-------------------|
| Primitive | Gatherer | 0.10 | 1.0× |
| Stone | Tribesman | 0.15 | 2.0× |
| Bronze | Laborer | 0.22 | 4.0× |
| Iron | Bondsman | 0.33 | 8.0× |
| Classical | Plebeian | 0.50 | 16× |
| Medieval | Serf | 0.75 | 32× |
| Renaissance | Artisan | 1.10 | 64× |
| Colonial | Settler | 1.65 | 128× |
| Industrial | Factory Hand | 2.50 | 256× |
| Victorian | Mill Worker | 3.75 | 512× |
| Electric | Tradesman | 5.60 | 1,024× |
| Atomic | Plant Operator | 8.40 | 2,048× |
| Modern | Contractor | 12.6 | 4,096× |
| Information | Gig Worker | 19.0 | 8,192× |
| Digital | Drone | 28.5 | 16,384× |
| Cyberpunk | Augment | 42.7 | 32,768× |
| Fusion | Reactor Hand | 64.0 | 65,536× |
| Space | Colonist | 96.0 | 131,072× |
| Interstellar | Void Walker | 144 | 262,144× |
| Galactic | Star Miner | 216 | 524,288× |
| Quantum | Harvester | 325 | 1,048,576× |

---

### Knowledge Domain
Fills: altar, sacred_grove, firepit (primitive), library, great_library, cathedral,
       university, research_lab, ai_lab, quantum_computer

| Age | Class Name | Food/tick | Output Multiplier |
|-----|-----------|-----------|-------------------|
| Primitive | Shaman | 0.50 | 1.0× |
| Stone | Elder | 0.75 | 2.0× |
| Bronze | Scribe | 1.10 | 4.0× |
| Iron | Philosopher | 1.65 | 8.0× |
| Classical | Rhetorician | 2.50 | 16× |
| Medieval | Monk | 3.75 | 32× |
| Renaissance | Polymath | 5.60 | 64× |
| Colonial | Naturalist | 8.40 | 128× |
| Industrial | Inventor | 12.6 | 256× |
| Victorian | Academic | 19.0 | 512× |
| Electric | Physicist | 28.5 | 1,024× |
| Atomic | Nuclear Scientist | 42.7 | 2,048× |
| Modern | Researcher | 64.0 | 4,096× |
| Information | Data Scientist | 96.0 | 8,192× |
| Digital | AI Engineer | 144 | 16,384× |
| Cyberpunk | Ghost | 216 | 32,768× |
| Fusion | Plasma Theorist | 325 | 65,536× |
| Space | Astrophysicist | 487 | 131,072× |
| Interstellar | Xenologist | 730 | 262,144× |
| Galactic | Cosmic Scholar | 1,096 | 524,288× |
| Quantum | Reality Theorist | 1,644 | 1,048,576× |

---

### Military Domain
Fills: barracks, keep, castle, bunker, missile_silo (all military production buildings)

| Age | Class Name | Food/tick | Output Multiplier |
|-----|-----------|-----------|-------------------|
| Primitive | Hunter | 0.25 | 1.0× |
| Stone | Warrior | 0.37 | 2.0× |
| Bronze | Spearman | 0.56 | 4.0× |
| Iron | Legionnaire | 0.84 | 8.0× |
| Classical | Hoplite | 1.26 | 16× |
| Medieval | Knight | 1.90 | 32× |
| Renaissance | Musketeer | 2.85 | 64× |
| Colonial | Militiaman | 4.27 | 128× |
| Industrial | Infantry | 6.40 | 256× |
| Victorian | Rifleman | 9.60 | 512× |
| Electric | Doughboy | 14.4 | 1,024× |
| Atomic | Grunt | 21.6 | 2,048× |
| Modern | Special Ops | 32.4 | 4,096× |
| Information | Cyber Warrior | 48.6 | 8,192× |
| Digital | Drone Pilot | 72.9 | 16,384× |
| Cyberpunk | Street Samurai | 109 | 32,768× |
| Fusion | Plasma Trooper | 164 | 65,536× |
| Space | Space Marine | 246 | 131,072× |
| Interstellar | Void Ranger | 369 | 262,144× |
| Galactic | Stellar Guardian | 553 | 524,288× |
| Quantum | Probability Assassin | 830 | 1,048,576× |

---

### Trade Domain
Fills: market, port, bank, amphitheater, forum, art_studio, colony, black_market,
       galactic_hub, trade-related wonders

| Age | Class Name | Food/tick | Output Multiplier |
|-----|-----------|-----------|-------------------|
| Primitive | Trader | 0.20 | 1.0× |
| Stone | Peddler | 0.30 | 2.0× |
| Bronze | Merchant | 0.45 | 4.0× |
| Iron | Caravanner | 0.68 | 8.0× |
| Classical | Agora Keeper | 1.02 | 16× |
| Medieval | Guilder | 1.53 | 32× |
| Renaissance | Financier | 2.30 | 64× |
| Colonial | Merchant Captain | 3.45 | 128× |
| Industrial | Industrialist | 5.17 | 256× |
| Victorian | Magnate | 7.76 | 512× |
| Electric | Tycoon | 11.6 | 1,024× |
| Atomic | Executive | 17.5 | 2,048× |
| Modern | Broker | 26.2 | 4,096× |
| Information | Venture Capitalist | 39.3 | 8,192× |
| Digital | Crypto Whale | 59.0 | 16,384× |
| Cyberpunk | Fixer | 88.5 | 32,768× |
| Fusion | Energy Baron | 133 | 65,536× |
| Space | Asteroid Trader | 199 | 131,072× |
| Interstellar | Trade Emissary | 299 | 262,144× |
| Galactic | Galactic Merchant | 448 | 524,288× |
| Quantum | Probability Trader | 672 | 1,048,576× |

---

### Engineering Domain
Unlocks at Bronze Age. Fills: lumber_mill, quarry (advanced), smithy, factory, oil_well,
power_grid, reactor, power_plant, fusion_reactor (infrastructure), plasma_forge, launch_pad

| Age | Class Name | Food/tick | Output Multiplier |
|-----|-----------|-----------|-------------------|
| Bronze | Toolmaker | 0.22 | 1.0× |
| Iron | Smith | 0.33 | 2.0× |
| Classical | Architect | 0.50 | 4.0× |
| Medieval | Master Builder | 0.75 | 8.0× |
| Renaissance | Engineer | 1.10 | 16× |
| Colonial | Mechanist | 1.65 | 32× |
| Industrial | Machinist | 2.50 | 64× |
| Victorian | Steam Engineer | 3.75 | 128× |
| Electric | Electrical Engineer | 5.60 | 256× |
| Atomic | Nuclear Engineer | 8.40 | 512× |
| Modern | Systems Engineer | 12.6 | 1,024× |
| Information | Software Engineer | 19.0 | 2,048× |
| Digital | Systems Architect | 28.5 | 4,096× |
| Cyberpunk | Augmentation Specialist | 42.7 | 8,192× |
| Fusion | Fusion Engineer | 64.0 | 16,384× |
| Space | Rocket Scientist | 96.0 | 32,768× |
| Interstellar | Warp Engineer | 144 | 65,536× |
| Galactic | Dyson Architect | 216 | 131,072× |
| Quantum | Reality Engineer | 325 | 262,144× |

---

### Hacker Domain
Unlocks at Information Age. Fills: server_farm, fiber_hub, data_center, ai_lab,
smart_grid, digital_archive, cyber_vault

| Age | Class Name | Food/tick | Output Multiplier |
|-----|-----------|-----------|-------------------|
| Information | Programmer | 0.30 | 1.0× |
| Digital | Hacker | 0.45 | 2.0× |
| Cyberpunk | Ghost | 0.68 | 4.0× |
| Fusion | Data Architect | 1.02 | 8.0× |
| Space | Signal Breaker | 1.53 | 16× |
| Interstellar | Memetic Hacker | 2.30 | 32× |
| Galactic | Consciousness Coder | 3.45 | 64× |
| Quantum | Quantum Cryptographer | 5.17 | 128× |

---

### Astronaut Domain
Unlocks at Space Age. Fills: space_station, orbital_habitat, warp_gate, colony_ship,
star_forge, stellar_cradle, dyson_scaffold, cosmic_beacon, reality_anchor, singularity_core

| Age | Class Name | Food/tick | Output Multiplier |
|-----|-----------|-----------|-------------------|
| Space | Astronaut | 0.40 | 1.0× |
| Interstellar | Cosmonaut | 0.60 | 2.0× |
| Galactic | Void Traveler | 0.90 | 4.0× |
| Quantum | Dimension Walker | 1.35 | 8.0× |

---

## Recruitment Rules

1. **Only current-age tier** can be newly recruited. Advancing an age immediately unlocks
   the new tier for each domain.
2. **Legacy workers persist** — existing workers don't disappear, but cost their original
   food rate. Players may keep them or dismiss them.
3. **Dismissal**: workers can be dismissed (removed from population) to free up food budget.
4. **Assignment is cross-tier**: a building's worker capacity accepts any tier of the
   correct domain. Higher tiers produce more per slot.
5. **Over-assignment is impossible**: can't assign more workers to a building than its
   `worker_capacity` allows.

## UI Implications

- Population panel shows active tier prominently; legacy tiers shown collapsed or grayed
- Each worker class name appears in the UI (seeing "Serf" instead of "Worker" rewards lore engagement)
- Worker assignment uses domain name, not individual class name, to avoid UI churn on advance
  (you always "assign to Raw Materials" — the class name is flavor on top)
