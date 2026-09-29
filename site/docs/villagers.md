# Workers

Workers are your civilization's workforce. They eat food each tick and drive production in every building they are assigned to. Workers are organized into a **12-domain system** tied directly to building lineages.

---

## Overview

Workers come from a **single generic pool**: every worker is identical until you assign it to a building. An assigned worker takes on the class name for that building's domain at the current age. There are no separate domain pools; any worker can go to any building.

**Production formula:**

```
production = base_rate × building_count × (0.20 + 0.80 × assigned / total_capacity)
```

- `total_capacity` is the building count times the building's worker slots.
- A building with no workers still produces at 20% (the floor).
- Filling every slot gives 100%.
- Assigning more workers than there are slots adds nothing.

---

## The 12 Domains

| Domain | What They Do | Unlocked |
|--------|--------------|----------|
| food | Produce food | Primitive Age |
| knowledge | Generate knowledge for research | Primitive Age |
| lumber | Extract wood, coal, oil, nanobots, quantum flux | Stone Age |
| masonry | Extract stone, iron ore, uranium, titanium ore, dark matter crystals, antimatter | Stone Age |
| faith | Generate faith | Primitive Age (early tiers); formal domain from Medieval Age |
| military | Train soldiers for campaigns | Iron Age |
| trade | Generate gold | Bronze Age |
| metallurgy | Refine ore into iron, steel, titanium, dark matter | Iron Age |
| engineering | Produce steel, electricity, plasma | Bronze Age (early tiers); later tiers from Victorian Age |
| energy | Generate electricity, plasma, quantum flux | Victorian Age |
| hacker | Generate data and crypto | Information Age |
| astronaut | Generate dark matter, antimatter | Space Age |

> **Note:** Culture buildings (Lineage 10) have no worker domain. They produce culture automatically each tick.

---

## Worker Class Names

Each domain has a class name for every age it spans. Workers are recruited generically and take on a class when assigned to a building: the class for that building's domain at the current age. When you advance, assigned workers take the new age's class name.

Class names are labels. What a worker produces depends on the building, and what it eats depends on the age (see [Food Drain](#food-drain)).

### Food domain (starts Primitive Age)

| Age | Class Name |
|-----|-----------|
| Primitive | Forager |
| Stone | Farmhand |
| Bronze | Cultivator |
| Iron | Laborer |
| Classical | Peasant |
| Medieval | Serf |
| Renaissance | Plowman |
| Colonial | Colonial Farmer |
| Industrial | Factory Hand |
| Victorian | Agricultural Worker |
| Electric | Electric Farmer |
| Atomic | Atomic Agronomist |
| Modern | Modern Farmer |
| Information | Digital Cultivator |
| Digital | AI Agronomist |
| Cyberpunk | Aug Harvester |
| Fusion | Bio-Farmer |
| Space | Zero-G Farmer |
| Interstellar | Stellar Cultivator |
| Galactic | Galactic Farmer |
| Quantum | Quantum Harvester |

### Knowledge domain (starts Primitive Age)

Primitive: Shaman → Stone: Elder → Bronze: Scribe → Iron: Scholar → Classical: Philosopher → Medieval: Friar → Renaissance: Academician → Colonial: Naturalist → Industrial: Engineer-Scientist → Victorian: Victorian Scholar → Electric: Research Fellow → Atomic: Nuclear Scientist → Modern: Modern Researcher → Information: Data Scientist → Digital: AI Researcher → Cyberpunk: Cyber-Scholar → Fusion: Fusion Theorist → Space: Orbital Researcher → Interstellar: Stellar Scientist → Galactic: Galactic Researcher → **Quantum: Quantum Theorist**

### Lumber domain (starts Stone Age)

Stone: Gatherer → Bronze: Woodcutter → Iron: Lumberjack → Classical: Sawyer → Medieval: Forester → Renaissance: Colonial Logger → Colonial: Mill Worker → Industrial: Coal Extractor → Victorian: Steam Logger → Electric: Electric Forester → Atomic: Fuel Extractor → Modern: Petroleum Worker → Information: Digital Forester → Digital: Bio-Extractor → Cyberpunk: Nano-Harvester → Fusion: Organic Engineer → Space: Biofield Harvester → Interstellar: Quantum Extractor → Galactic: Galactic Forester → **Quantum: Cosmic Extractor**

### Masonry domain (starts Stone Age)

Stone: Quarryman → Bronze: Stone Cutter → Iron: Miner → Classical: Iron Extractor → Medieval: Medieval Miner → Renaissance: Renaissance Quarryman → Colonial: Colonial Miner → Industrial: Industrial Miner → Victorian: Victorian Quarryman → Electric: Electric Miner → Atomic: Uranium Miner → Modern: Modern Geologist → Information: Data Miner → Digital: Digital Excavator → Cyberpunk: Cyber Miner → Fusion: Plasma Driller → Space: Space Miner → Interstellar: Asteroid Miner → Galactic: Dark Matter Extractor → **Quantum: Crystal Miner**

### Faith domain (early tiers from Primitive Age, formal tiers from Medieval Age)

The early tiers (Primitive: Devotee · Stone: Believer · Bronze: Worshipper · Iron: Celebrant · Classical: Initiate) cover shrines and altars. The formal tiers begin at the Medieval Age.

Medieval: Acolyte → Renaissance: Monk → Colonial: Missionary → Industrial: Revivalist → Victorian: Parish Priest → Electric: Evangelical → Atomic: Atomic Priest → Modern: Modern Shepherd → Information: Digital Devotee → Digital: Virtual Cleric → Cyberpunk: Cyber Cleric → Fusion: Plasma Prophet → Space: Star Preacher → Interstellar: Interstellar Mystic → Galactic: Galactic High Priest → **Quantum: Quantum Sage**

### Military domain (starts Iron Age)

Iron: Soldier → Classical: Legionary → Medieval: Knight → Renaissance: Musketeer → Colonial: Colonial Marine → Industrial: Industrial Rifleman → Victorian: Victorian Guard → Electric: Electric Trooper → Atomic: Atomic Soldier → Modern: Modern Soldier → Information: Information Warrior → Digital: Digital Soldier → Cyberpunk: Cyber Warrior → Fusion: Plasma Trooper → Space: Space Marine → Interstellar: Interstellar Commando → Galactic: Galactic Guardian → **Quantum: Quantum Soldier**

### Trade domain (starts Bronze Age)

Bronze: Peddler → Iron: Merchant → Classical: Trader → Medieval: Nobleman → Renaissance: Banker → Colonial: Colonial Merchant → Industrial: Industrialist → Victorian: Victorian Trader → Electric: Electric Broker → Atomic: Atomic Trader → Modern: Corporate Trader → Information: Digital Trader → Digital: Crypto Broker → Cyberpunk: Cyber Dealer → Fusion: Plasma Merchant → Space: Space Trader → Interstellar: Interstellar Broker → Galactic: Galactic Merchant → **Quantum: Quantum Dealer**

### Metallurgy domain (starts Iron Age)

Iron: Smelter → Classical: Ironworker → Medieval: Medieval Smith → Renaissance: Renaissance Metallurgist → Colonial: Foundry Worker → Industrial: Factory Worker → Victorian: Steam Smelter → Electric: Electric Smelter → Atomic: Atomic Metallurgist → Modern: Modern Metallurgist → Information: Digital Foundry Worker → Digital: Digital Smelter → Cyberpunk: Cyber Forge Worker → Fusion: Plasma Metallurgist → Space: Stellar Foundry Worker → Interstellar: Stellar Smelter → Galactic: Galactic Metallurgist → **Quantum: Quantum Smelter**

### Engineering domain (early tiers from Bronze Age, later tiers from Victorian Age)

Bronze: Apprentice → Iron: Craftsman → Classical: Artisan → Medieval: Engineer → Renaissance: Master Eng. → Colonial: Mechanic → Industrial: Machinist → Victorian: Tinker → Electric: Electrical Engineer → Atomic: Nuclear Engineer → Modern: Systems Engineer → Information: Software Engineer → Digital: AI Engineer → Cyberpunk: Cyber Engineer → Fusion: Plasma Engineer → Space: Space Engineer → Interstellar: Warp Engineer → Galactic: Galactic Engineer → **Quantum: Quantum Engineer**

### Energy domain (starts Victorian Age)

Victorian: Stoker → Electric: Power Worker → Atomic: Reactor Technician → Modern: Power Engineer → Information: Grid Operator → Digital: Digital Power Manager → Cyberpunk: Cyber Energy Worker → Fusion: Fusion Technician → Space: Solar Engineer → Interstellar: Dark Energy Worker → Galactic: Antimatter Specialist → **Quantum: Zero-Point Engineer**

### Hacker domain (starts Information Age)

Information: Script Kiddie → Digital: Coder → Cyberpunk: Black Hat → Fusion: AI Hacker → Space: Orbital Hacker → Interstellar: Interstellar Netrunner → Galactic: Galactic Hacker → **Quantum: Quantum Hacker**

### Astronaut domain (starts Space Age)

Space: Cadet → Interstellar: Interstellar Pilot → Galactic: Galactic Explorer → **Quantum: Quantum Astronaut**

---

## Commands

**Recruit workers:**

```
recruit [count|max]
```

Examples: `recruit`, `recruit 5`, `recruit max`

Workers are recruited from free housing, and you don't name a domain. A new worker has no class until you assign it; it then takes the class of that building's domain (a worker on a gathering_camp becomes a Forager, one on a story_circle a Shaman). You can recruit only up to your housing. Recruiting is free, but each new worker starts eating food the next tick.

**Assign workers to a building:**

```
assign <building> [count|all]
```

Example: `assign gathering_camp 5`

The domain comes from the building. A building with 15 worker slots, built 3 times, has 45 slots in total.

**Unassign workers:**

```
unassign <building> [count|all]
```

Returns workers to the idle pool. Idle workers still eat.

**Permanently dismiss workers:**

```
dismiss <building> [count|all]
```

`dismiss` removes workers from a building **and** from your population. Unlike `unassign`, dismissed workers are gone. Use it to free housing or cut food drain when you have more workers than you can feed.

**View worker status:**

```
workers
```

Opens the Workers panel:
- **Morale**: the current morale as a colored bar
- **Summary**: population and housing, idle count, housing left, food drain/tick, net food/tick (color-coded), and either how many workers your food can sustain or a food-deficit warning
- **Slot Utilization**: filled vs. total slots across all worker buildings, a fill bar, and the buildings with the most free slots
- **Domain Breakdown**: each domain's class name, with a bar per building showing assigned/slots

The Workers box in the sidebar is always visible and shows population, idle, housing left, drain and net food at a glance.

---

## Starvation

When food runs out while your workers are still eating, they begin dying:

- A warning is logged on the first tick without food
- 1 worker dies every 5 ticks (10 seconds at 1x) until food recovers
- Deaths are logged in red
- As soon as there is food in stock again, starvation stops and a recovery message is logged

Overpopulating without the food to back it up costs you workers. The fastest fix is to `dismiss` workers from low-priority buildings, which cuts drain at once, or to move workers onto food buildings with `unassign` and `assign`.

---

## Morale

Morale is a percentage multiplier on all worker-driven building output. It **starts at 50%** (neutral), has a **10% floor**, and is capped at **100% + 5% per wonder** built.

```
output = base_rate × building_count × (0.20 + 0.80 × assigned / capacity) × morale_multiplier
```

The effect is a continuous curve centered on 50%. At exactly 50% output is normal. Above 50% it gets a bonus that grows steadily to **+20%** at the cap; below 50% it gets a penalty that deepens to **×0.50** at the 10% floor. Morale drifts back toward 50% each tick, so a bonus has to be kept up.

- **Rises from:** worship and culture buildings, your faith production rate, good events, age advances
- **Falls from:** starvation, too many soldiers (military workers over 30% of population), too many idle workers (over 50% of population), bad events and catastrophes

Morale shows as a colored bar in the Workers panel and as `Morale: NN%` in the status bar.

See [Morale](morale.md) for the full details.

---

## Food Drain

Every worker eats the same amount of food per tick, whatever building it staffs. The amount depends only on your current age: it starts at **0.06 food/tick** per worker in the Primitive Age and rises 12% with each age.

```
total_food_drain = food_per_worker(current age) × population
```

- Primitive Age: **0.06 food/tick** per worker
- Medieval Age: about **0.11 food/tick** per worker
- Modern Age: about **0.23 food/tick** per worker
- Quantum Age and later: about **0.58 food/tick** per worker

When you advance, the new rate applies to every worker at once. See [Workers & Domains](workers-and-domains.md#food-drain) for the full table.

---

## Tips

- **Prioritize food workers early.** Every worker eats. A food deficit stalls recruitment and can collapse your economy.
- **Faith workers before epoch transitions.** Your faith level as a percentage of faith storage sets the epoch roll odds. Keep faith workers assigned ahead of predicted epoch events.
- **Metallurgy needs both extraction and refining.** Assign masonry workers to extraction buildings to produce raw ore, then metallurgy workers to smelters to refine it.
- **One pool.** All workers are in one pool regardless of which buildings they staff. Reassign freely between any buildings at any time.
- **Check the idle count.** Type `status` (or `s`) to see it. Idle workers eat food and produce nothing.
- **Use `dismiss` to cut population.** In a food deficit you can't fix, `dismiss` workers from non-critical buildings to reduce drain permanently. `unassign` alone does not help; idle workers still eat.
- **Use `workers` for the full picture.** The panel shows net food/tick, how many workers your food can sustain, and the per-domain breakdown.
