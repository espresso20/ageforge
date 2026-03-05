# Workers & Villagers

Workers are your civilization's workforce. They consume food each tick and drive production in every building they are assigned to. The old eight-type system (Worker/Shaman/Scholar/Soldier/Merchant/Engineer/Hacker/Astronaut) has been replaced by a **12-domain system** tied directly to building lineages.

---

## Overview

Workers are organized into 12 domains. Each domain is tied to specific building lineages, and workers in a domain can only be assigned to buildings that share that domain.

**Production formula:**

```
production = base_rate × building_count × (0.20 + 0.80 × assigned / total_capacity)
```

- `total_capacity = building_count × WorkerCapacity`
- **20% floor** — even completely unassigned buildings still produce at 20% efficiency
- **Full assignment** — assigning workers up to total capacity gives 100% efficiency
- Assigning more workers than the total capacity has no additional effect

---

## The 12 Domains

| Domain | Legacy Name | What They Do | Unlocked |
|--------|-------------|--------------|----------|
| food | Worker | Produce food; drain food for all workers | Primitive Age |
| knowledge | Scholar/Shaman | Generate knowledge for research | Primitive Age |
| lumber | — | Extract wood, coal, oil, nanobots, quantum flux | Stone Age |
| masonry | — | Extract stone, iron ore, uranium, titanium ore, dark matter crystals, antimatter | Stone Age |
| faith | Shaman | Generate faith | Medieval Age |
| military | Soldier | Train soldiers for expeditions | Iron Age |
| trade | Merchant | Generate gold | Bronze Age |
| metallurgy | — | Refine ore into iron, steel, titanium, dark matter | Iron Age |
| engineering | Engineer | Produce steel, electricity, plasma | Victorian Age |
| energy | — | Generate electricity, plasma, quantum flux | Victorian Age |
| hacker | Hacker | Generate data and crypto | Information Age |
| astronaut | Astronaut | Generate dark matter, antimatter | Space Age |

> **Note:** Culture buildings (Lineage 10) have no worker domain — they produce culture automatically each tick.

---

## Worker Class Names

Each domain has age-tiered class names. When you recruit a worker, they are always created at the current age's tier for that domain. New recruits join at the current tier; existing workers retain their tier and its food cost.

**Food drain scales geometrically:** `FoodCost = baseFoodCost × 1.5^tier`

**Production scales geometrically:** `Multiplier = 2.0^tier` (vs. the building's base rate)

### Food Domain — base food cost 1.0, starts Primitive Age

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

### Knowledge Domain — base food cost 1.0, starts Primitive Age

Primitive: Shaman → Stone: Elder → Bronze: Scribe → Iron: Scholar → Classical: Philosopher → Medieval: Friar → Renaissance: Academician → Colonial: Naturalist → Industrial: Engineer-Scientist → Victorian: Victorian Scholar → Electric: Research Fellow → Atomic: Nuclear Scientist → Modern: Modern Researcher → Information: Data Scientist → Digital: AI Researcher → Cyberpunk: Cyber-Scholar → Fusion: Fusion Theorist → Space: Orbital Researcher → Interstellar: Stellar Scientist → Galactic: Galactic Researcher → **Quantum: Quantum Theorist**

### Lumber Domain — base food cost 1.0, starts Stone Age

Stone: Gatherer → Bronze: Woodcutter → Iron: Lumberjack → Classical: Sawyer → Medieval: Forester → Renaissance: Colonial Logger → Colonial: Mill Worker → Industrial: Coal Extractor → Victorian: Steam Logger → Electric: Electric Forester → Atomic: Fuel Extractor → Modern: Petroleum Worker → Information: Digital Forester → Digital: Bio-Extractor → Cyberpunk: Nano-Harvester → Fusion: Organic Engineer → Space: Biofield Harvester → Interstellar: Quantum Extractor → Galactic: Galactic Forester → **Quantum: Cosmic Extractor**

### Masonry Domain — base food cost 1.0, starts Stone Age

Stone: Quarryman → Bronze: Stone Cutter → Iron: Miner → Classical: Iron Extractor → Medieval: Medieval Miner → Renaissance: Renaissance Quarryman → Colonial: Colonial Miner → Industrial: Industrial Miner → Victorian: Victorian Quarryman → Electric: Electric Miner → Atomic: Uranium Miner → Modern: Modern Geologist → Information: Data Miner → Digital: Digital Excavator → Cyberpunk: Cyber Miner → Fusion: Plasma Driller → Space: Space Miner → Interstellar: Asteroid Miner → Galactic: Dark Matter Extractor → **Quantum: Crystal Miner**

### Faith Domain — base food cost 2.0, starts Medieval Age

Medieval: Acolyte → Renaissance: Monk → Colonial: Missionary → Industrial: Revivalist → Victorian: Parish Priest → Electric: Evangelical → Atomic: Atomic Priest → Modern: Modern Shepherd → Information: Digital Devotee → Digital: Virtual Cleric → Cyberpunk: Cyber Cleric → Fusion: Plasma Prophet → Space: Star Preacher → Interstellar: Interstellar Mystic → Galactic: Galactic High Priest → **Quantum: Quantum Sage**

### Military Domain — base food cost 2.0, starts Iron Age

Iron: Soldier → Classical: Legionary → Medieval: Knight → Renaissance: Musketeer → Colonial: Colonial Marine → Industrial: Industrial Rifleman → Victorian: Victorian Guard → Electric: Electric Trooper → Atomic: Atomic Soldier → Modern: Modern Soldier → Information: Information Warrior → Digital: Digital Soldier → Cyberpunk: Cyber Warrior → Fusion: Plasma Trooper → Space: Space Marine → Interstellar: Interstellar Commando → Galactic: Galactic Guardian → **Quantum: Quantum Soldier**

### Trade Domain — base food cost 1.0, starts Bronze Age

Bronze: Peddler → Iron: Merchant → Classical: Trader → Medieval: Nobleman → Renaissance: Banker → Colonial: Colonial Merchant → Industrial: Industrialist → Victorian: Victorian Trader → Electric: Electric Broker → Atomic: Atomic Trader → Modern: Corporate Trader → Information: Digital Trader → Digital: Crypto Broker → Cyberpunk: Cyber Dealer → Fusion: Plasma Merchant → Space: Space Trader → Interstellar: Interstellar Broker → Galactic: Galactic Merchant → **Quantum: Quantum Dealer**

### Metallurgy Domain — base food cost 2.0, starts Iron Age

Iron: Smelter → Classical: Ironworker → Medieval: Medieval Smith → Renaissance: Renaissance Metallurgist → Colonial: Foundry Worker → Industrial: Factory Worker → Victorian: Steam Smelter → Electric: Electric Smelter → Atomic: Atomic Metallurgist → Modern: Modern Metallurgist → Information: Digital Foundry Worker → Digital: Digital Smelter → Cyberpunk: Cyber Forge Worker → Fusion: Plasma Metallurgist → Space: Stellar Foundry Worker → Interstellar: Stellar Smelter → Galactic: Galactic Metallurgist → **Quantum: Quantum Smelter**

### Engineering Domain — base food cost 8.0, starts Victorian Age

Victorian: Tinker → Electric: Electrical Engineer → Atomic: Nuclear Engineer → Modern: Systems Engineer → Information: Software Engineer → Digital: AI Engineer → Cyberpunk: Cyber Engineer → Fusion: Plasma Engineer → Space: Space Engineer → Interstellar: Warp Engineer → Galactic: Galactic Engineer → **Quantum: Quantum Engineer**

### Energy Domain — base food cost 8.0, starts Victorian Age

Victorian: Stoker → Electric: Power Worker → Atomic: Reactor Technician → Modern: Power Engineer → Information: Grid Operator → Digital: Digital Power Manager → Cyberpunk: Cyber Energy Worker → Fusion: Fusion Technician → Space: Solar Engineer → Interstellar: Dark Energy Worker → Galactic: Antimatter Specialist → **Quantum: Zero-Point Engineer**

### Hacker Domain — base food cost 16.0, starts Information Age

Information: Script Kiddie → Digital: Coder → Cyberpunk: Black Hat → Fusion: AI Hacker → Space: Orbital Hacker → Interstellar: Interstellar Netrunner → Galactic: Galactic Hacker → **Quantum: Quantum Hacker**

### Astronaut Domain — base food cost 32.0, starts Space Age

Space: Cadet → Interstellar: Interstellar Pilot → Galactic: Galactic Explorer → **Quantum: Quantum Astronaut**

---

## Commands

**Recruit a worker:**

```
recruit <domain>
```

Examples: `recruit food`, `recruit knowledge`, `recruit military`

You can only recruit the current age's tier. Recruiting costs food upfront; the new worker begins draining food per tick immediately.

**Assign workers to a building:**

```
assign <domain> <building_key> <count>
```

Example: `assign food gathering_camp 5`

Workers must match the building's domain. A building with worker capacity 15, built 3 times, has 45 total slots.

**Unassign workers:**

```
unassign <domain> <building_key> <count>
unassign all <domain>
```

---

## Food Drain

Every worker in every domain costs food per tick. The exact amount is `baseFoodCost × 1.5^tier` for their current class tier. As you advance ages and recruit higher-tier workers, food drain rises substantially.

- Tier 0 food workers cost 1.0 food/tick each
- Tier 5 (Medieval Serf) costs `1.0 × 1.5^5 = ~7.6 food/tick`
- High-tier specialists (Hacker at tier 0 = 16.0, Astronaut at tier 0 = 32.0) are expensive — ensure food production keeps up before recruiting them

---

## Tips

- **Prioritize food workers early** — every other domain drains food. A food deficit stalls recruitment and can collapse your economy.
- **Faith workers before epoch transitions** — your faith level as a percentage of storage cap determines epoch roll odds. Keep faith workers assigned ahead of predicted epoch events.
- **Knowledge workers for fast research** — the knowledge domain multiplier doubles each tier, making late-game scholars vastly more productive than early shamans.
- **Metallurgy requires both extraction and refining** — assign masonry workers to geological extraction buildings to produce raw ore, then metallurgy workers to smelters to refine it.
- **Legacy class workers** — workers recruited in a previous age retain their tier's food cost and multiplier. New recruits always join at the current tier. Both coexist in the same domain pool.
- **Check idle count** — press `e` (Economy tab) to see the idle count in the status bar. Unassigned workers still drain food for zero extra production benefit.
