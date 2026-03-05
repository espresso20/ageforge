# Workers & Domains — Reference

Workers are the active labour force in AgeForge. Each worker belongs to a **domain** tied to a specific set of building lineages. Assigning workers to buildings increases production efficiency; leaving buildings unstaffed still produces at 20% base rate.

For a gentler introduction see [Villagers](villagers.md).

---

## Overview

- 12 worker domains, each covering one or more building lineages
- Output formula: `output = base_rate × building_count × (0.20 + 0.80 × assigned/total_capacity)`
- Food drain: `domain_food_cost × worker_count` per tick, scaling geometrically (×1.5 per class tier)
- Recruitment: `recruit <domain>` — always recruits the current age's tier; cost scales with tier
- Multiple class tiers can coexist in a domain (older workers keep their tier's food cost and multiplier)

---

## Domain Reference

| Domain | Legacy Key | Building Lineage | Primary Output |
|--------|------------|------------------|----------------|
| food | worker | Food lineage | food |
| knowledge | scholar | Knowledge lineage | knowledge |
| faith | shaman | Faith lineage | faith |
| military | soldier | Military lineage | soldiers |
| trade | merchant | Trade lineage | gold |
| engineering | engineer | Engineering lineage | iron / steel / electricity / plasma |
| lumber | — | Organic Extraction lineage | wood / coal / oil / nanobots / quantum_flux |
| masonry | — | Geological Extraction lineage | stone / iron_ore / uranium / titanium_ore |
| metallurgy | — | Metallurgy lineage | iron / steel / titanium / dark_matter / antimatter |
| energy | — | Energy lineage | coal / electricity / plasma / dark_matter |
| hacker | hacker | Hacker / Digital lineage | data |
| astronaut | astronaut | Space buildings | dark_matter / antimatter |

> Culture has no worker domain — it auto-produces passively and cannot be boosted by assignment.

---

## Domain Unlock Ages

| Domain | Unlocks At |
|--------|------------|
| food | Primitive Age |
| knowledge | Primitive Age |
| lumber | Stone Age |
| masonry | Stone Age |
| military | Iron Age |
| trade | Bronze Age |
| faith | Medieval Age |
| metallurgy | Iron Age |
| engineering | Victorian Age |
| energy | Victorian Age |
| hacker | Information Age |
| astronaut | Space Age |

---

## Worker Class Progression

Each domain has one class per age it spans. Recruit the current age's tier with `recruit <domain>`. The first class in each domain is always tier 0 (multiplier ×1.0); each subsequent tier doubles the multiplier and adds ×1.5 to food cost.

### food (primitive → quantum, 21 tiers)

| Tier | Age | Class Name |
|------|-----|------------|
| 0 | Primitive | Forager |
| 1 | Stone | Farmhand |
| 2 | Bronze | Cultivator |
| 3 | Iron | Laborer |
| 4 | Classical | Peasant |
| 5 | Medieval | Serf |
| 6 | Renaissance | Plowman |
| 7 | Colonial | Colonial Farmer |
| 8 | Industrial | Factory Hand |
| 9 | Victorian | Agricultural Worker |
| 10 | Electric | Electric Farmer |
| 11 | Atomic | Atomic Agronomist |
| 12 | Modern | Modern Farmer |
| 13 | Information | Digital Cultivator |
| 14 | Digital | AI Agronomist |
| 15 | Cyberpunk | Aug Harvester |
| 16 | Fusion | Bio-Farmer |
| 17 | Space | Zero-G Farmer |
| 18 | Interstellar | Stellar Cultivator |
| 19 | Galactic | Galactic Farmer |
| 20 | Quantum | Quantum Harvester |

### knowledge (primitive → quantum, 21 tiers)

| Tier | Age | Class Name |
|------|-----|------------|
| 0 | Primitive | Shaman |
| 1 | Stone | Elder |
| 2 | Bronze | Scribe |
| 3 | Iron | Scholar |
| 4 | Classical | Philosopher |
| 5 | Medieval | Friar |
| 6 | Renaissance | Academician |
| 7 | Colonial | Naturalist |
| 8 | Industrial | Engineer-Scientist |
| 9 | Victorian | Victorian Scholar |
| 10 | Electric | Research Fellow |
| 11 | Atomic | Nuclear Scientist |
| 12 | Modern | Modern Researcher |
| 13 | Information | Data Scientist |
| 14 | Digital | AI Researcher |
| 15 | Cyberpunk | Cyber-Scholar |
| 16 | Fusion | Fusion Theorist |
| 17 | Space | Orbital Researcher |
| 18 | Interstellar | Stellar Scientist |
| 19 | Galactic | Galactic Researcher |
| 20 | Quantum | Quantum Theorist |

### faith (medieval → quantum, 16 tiers)

| Tier | Age | Class Name |
|------|-----|------------|
| 0 | Medieval | Acolyte |
| 1 | Renaissance | Monk |
| 2 | Colonial | Missionary |
| 3 | Industrial | Revivalist |
| 4 | Victorian | Parish Priest |
| 5 | Electric | Evangelical |
| 6 | Atomic | Atomic Priest |
| 7 | Modern | Modern Shepherd |
| 8 | Information | Digital Devotee |
| 9 | Digital | Virtual Cleric |
| 10 | Cyberpunk | Cyber Cleric |
| 11 | Fusion | Plasma Prophet |
| 12 | Space | Star Preacher |
| 13 | Interstellar | Interstellar Mystic |
| 14 | Galactic | Galactic High Priest |
| 15 | Quantum | Quantum Sage |

### military (iron → quantum, 18 tiers)

| Tier | Age | Class Name |
|------|-----|------------|
| 0 | Iron | Soldier |
| 1 | Classical | Legionary |
| 2 | Medieval | Knight |
| 3 | Renaissance | Musketeer |
| 4 | Colonial | Colonial Marine |
| 5 | Industrial | Industrial Rifleman |
| 6 | Victorian | Victorian Guard |
| 7 | Electric | Electric Trooper |
| 8 | Atomic | Atomic Soldier |
| 9 | Modern | Modern Soldier |
| 10 | Information | Information Warrior |
| 11 | Digital | Digital Soldier |
| 12 | Cyberpunk | Cyber Warrior |
| 13 | Fusion | Plasma Trooper |
| 14 | Space | Space Marine |
| 15 | Interstellar | Interstellar Commando |
| 16 | Galactic | Galactic Guardian |
| 17 | Quantum | Quantum Soldier |

### trade (bronze → quantum, 19 tiers)

| Tier | Age | Class Name |
|------|-----|------------|
| 0 | Bronze | Peddler |
| 1 | Iron | Merchant |
| 2 | Classical | Trader |
| 3 | Medieval | Nobleman |
| 4 | Renaissance | Banker |
| 5 | Colonial | Colonial Merchant |
| 6 | Industrial | Industrialist |
| 7 | Victorian | Victorian Trader |
| 8 | Electric | Electric Broker |
| 9 | Atomic | Atomic Trader |
| 10 | Modern | Corporate Trader |
| 11 | Information | Digital Trader |
| 12 | Digital | Crypto Broker |
| 13 | Cyberpunk | Cyber Dealer |
| 14 | Fusion | Plasma Merchant |
| 15 | Space | Space Trader |
| 16 | Interstellar | Interstellar Broker |
| 17 | Galactic | Galactic Merchant |
| 18 | Quantum | Quantum Dealer |

### engineering (victorian → quantum, 12 tiers)

| Tier | Age | Class Name |
|------|-----|------------|
| 0 | Victorian | Tinker |
| 1 | Electric | Electrical Engineer |
| 2 | Atomic | Nuclear Engineer |
| 3 | Modern | Systems Engineer |
| 4 | Information | Software Engineer |
| 5 | Digital | AI Engineer |
| 6 | Cyberpunk | Cyber Engineer |
| 7 | Fusion | Plasma Engineer |
| 8 | Space | Space Engineer |
| 9 | Interstellar | Warp Engineer |
| 10 | Galactic | Galactic Engineer |
| 11 | Quantum | Quantum Engineer |

### metallurgy (iron → quantum, 18 tiers)

| Tier | Age | Class Name |
|------|-----|------------|
| 0 | Iron | Smelter |
| 1 | Classical | Ironworker |
| 2 | Medieval | Medieval Smith |
| 3 | Renaissance | Renaissance Metallurgist |
| 4 | Colonial | Foundry Worker |
| 5 | Industrial | Factory Worker |
| 6 | Victorian | Steam Smelter |
| 7 | Electric | Electric Smelter |
| 8 | Atomic | Atomic Metallurgist |
| 9 | Modern | Modern Metallurgist |
| 10 | Information | Digital Foundry Worker |
| 11 | Digital | Digital Smelter |
| 12 | Cyberpunk | Cyber Forge Worker |
| 13 | Fusion | Plasma Metallurgist |
| 14 | Space | Stellar Foundry Worker |
| 15 | Interstellar | Stellar Smelter |
| 16 | Galactic | Galactic Metallurgist |
| 17 | Quantum | Quantum Smelter |

### energy (victorian → quantum, 12 tiers)

| Tier | Age | Class Name |
|------|-----|------------|
| 0 | Victorian | Stoker |
| 1 | Electric | Power Worker |
| 2 | Atomic | Reactor Technician |
| 3 | Modern | Power Engineer |
| 4 | Information | Grid Operator |
| 5 | Digital | Digital Power Manager |
| 6 | Cyberpunk | Cyber Energy Worker |
| 7 | Fusion | Fusion Technician |
| 8 | Space | Solar Engineer |
| 9 | Interstellar | Dark Energy Worker |
| 10 | Galactic | Antimatter Specialist |
| 11 | Quantum | Zero-Point Engineer |

### lumber (stone → quantum, 20 tiers)

| Tier | Age | Class Name |
|------|-----|------------|
| 0 | Stone | Gatherer |
| 1 | Bronze | Woodcutter |
| 2 | Iron | Lumberjack |
| 3 | Classical | Sawyer |
| 4 | Medieval | Forester |
| 5 | Renaissance | Colonial Logger |
| 6 | Colonial | Mill Worker |
| 7 | Industrial | Coal Extractor |
| 8 | Victorian | Steam Logger |
| 9 | Electric | Electric Forester |
| 10 | Atomic | Fuel Extractor |
| 11 | Modern | Petroleum Worker |
| 12 | Information | Digital Forester |
| 13 | Digital | Bio-Extractor |
| 14 | Cyberpunk | Nano-Harvester |
| 15 | Fusion | Organic Engineer |
| 16 | Space | Biofield Harvester |
| 17 | Interstellar | Quantum Extractor |
| 18 | Galactic | Galactic Forester |
| 19 | Quantum | Cosmic Extractor |

### masonry (stone → quantum, 20 tiers)

| Tier | Age | Class Name |
|------|-----|------------|
| 0 | Stone | Quarryman |
| 1 | Bronze | Stone Cutter |
| 2 | Iron | Miner |
| 3 | Classical | Iron Extractor |
| 4 | Medieval | Medieval Miner |
| 5 | Renaissance | Renaissance Quarryman |
| 6 | Colonial | Colonial Miner |
| 7 | Industrial | Industrial Miner |
| 8 | Victorian | Victorian Quarryman |
| 9 | Electric | Electric Miner |
| 10 | Atomic | Uranium Miner |
| 11 | Modern | Modern Geologist |
| 12 | Information | Data Miner |
| 13 | Digital | Digital Excavator |
| 14 | Cyberpunk | Cyber Miner |
| 15 | Fusion | Plasma Driller |
| 16 | Space | Space Miner |
| 17 | Interstellar | Asteroid Miner |
| 18 | Galactic | Dark Matter Extractor |
| 19 | Quantum | Crystal Miner |

### hacker (information → quantum, 8 tiers)

| Tier | Age | Class Name |
|------|-----|------------|
| 0 | Information | Script Kiddie |
| 1 | Digital | Coder |
| 2 | Cyberpunk | Black Hat |
| 3 | Fusion | AI Hacker |
| 4 | Space | Orbital Hacker |
| 5 | Interstellar | Interstellar Netrunner |
| 6 | Galactic | Galactic Hacker |
| 7 | Quantum | Quantum Hacker |

### astronaut (space → quantum, 4 tiers)

| Tier | Age | Class Name |
|------|-----|------------|
| 0 | Space | Cadet |
| 1 | Interstellar | Interstellar Pilot |
| 2 | Galactic | Galactic Explorer |
| 3 | Quantum | Quantum Astronaut |

---

## Recruitment

```
recruit <domain>
```

Recruits one worker at the current age's class tier for that domain. Recruitment cost scales with tier. You can only recruit the tier matching your current age — you cannot recruit a future tier or retroactively recruit a past tier.

**Examples:**
```
recruit food
recruit knowledge
recruit military
recruit hacker
```

---

## Assignment Commands

```
assign <domain> <building_key> <count>
unassign <domain> <building_key> <count>
unassign all <domain>
```

**Examples:**
```
assign food gathering_camp 5
assign knowledge library 3
unassign food gathering_camp 2
unassign all military
```

---

## Assignment Rules

- The domain must match the building's `WorkerDomain` field — you cannot assign food workers to a library or knowledge workers to a farm
- `count` must not exceed the domain's current idle worker pool
- `total_capacity` = `building_count × per_building_capacity` (shown in the Economy tab)
- Assigning more workers than capacity allows is rejected — fill more buildings first

---

## Efficiency Formula

```
production_per_tick = base_rate × building_count × efficiency_multiplier
efficiency_multiplier = 0.20 + 0.80 × (total_assigned / total_capacity)
```

The 0.20 floor means unstaffed buildings are not dead weight — they always contribute a baseline. Full staffing is required to reach 100% efficiency.

| Fill Level | Efficiency |
|------------|------------|
| 0% (none assigned) | 20% |
| 25% capacity | 40% |
| 50% capacity | 60% |
| 75% capacity | 80% |
| 100% capacity | 100% |

---

## Legacy Workers (Age Advancement)

When you advance to a new age, existing workers are not removed. They retain their tier's food cost and production multiplier. You can then recruit new workers at the new (higher) tier. Both tiers coexist in the same domain:

- The UI shows the current (highest) class name prominently
- Older-tier workers continue to produce at their multiplier but cost their original food rate
- Over time, higher-tier workers make up a larger share of your workforce as you recruit more

This means transitioning ages does not break your production — you gradually upgrade your workforce.

---

## Food Drain Management

Total food drain per tick = sum of `(domain_food_cost × worker_count)` for every domain.

The Economy tab (`e` or `F1`) shows per-domain food drain in the villager panel. Key tips:

- High-base-cost domains (engineering 8.0, energy 8.0, hacker 16.0, astronaut 32.0) compound quickly with tier scaling — recruit carefully
- Ensure your food production rate (from food-domain workers and buildings) exceeds your total drain before recruiting in expensive domains
- `unassign all <domain>` combined with `recruit` in a cheaper domain is a valid rebalancing strategy
- Faith workers (base 2.0) are cheaper than military (base 2.0 same) but unlock later — plan food accordingly

---

## See Also

- [Villagers](villagers.md) — introductory guide to worker types and early-game assignments
- [Buildings](buildings.md) — building lineages and per-building capacity values
- [Resources](resources.md) — which resources each lineage produces at which age
- [Epochs](epochs.md) — how epoch transitions affect building output resources
