# Workers

Workers are your civilization's workforce. They consume food each tick and drive production in every building they are assigned to. Workers are organized into a **12-domain system** tied directly to building lineages.

---

## Overview

Workers come from a **single generic pool** — all workers are identical until assigned to a building. When you assign a worker to a building, they take on the class name for that building's domain at the current age. There are no separate domain pools; any worker can go to any building.

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

| Domain | What They Do | Unlocked |
|--------|--------------|----------|
| food | Produce food; drain food for all workers | Primitive Age |
| knowledge | Generate knowledge for research | Primitive Age |
| lumber | Extract wood, coal, oil, nanobots, quantum flux | Stone Age |
| masonry | Extract stone, iron ore, uranium, titanium ore, dark matter crystals, antimatter | Stone Age |
| faith | Generate faith | Primitive Age (early tiers); formal domain from Medieval Age |
| military | Train soldiers for expeditions | Iron Age |
| trade | Generate gold | Bronze Age |
| metallurgy | Refine ore into iron, steel, titanium, dark matter | Iron Age |
| engineering | Produce steel, electricity, plasma | Bronze Age (early tiers); high-cost tier from Victorian Age |
| energy | Generate electricity, plasma, quantum flux | Victorian Age |
| hacker | Generate data and crypto | Information Age |
| astronaut | Generate dark matter, antimatter | Space Age |

> **Note:** Culture buildings (Lineage 10) have no worker domain — they produce culture automatically each tick.

---

## Worker Class Names

Each domain has age-tiered class names. Workers are recruited generically and take on a domain class when assigned to a building. When assigned, they join at the current age's tier for that domain. Existing workers retain their tier and its food cost when you advance ages.

**Food drain scales geometrically:** `FoodCost = baseFoodCost × 1.12^tier`

**Production scales geometrically:** `Multiplier = 2.0^tier` (vs. the building's base rate)

### Food Domain — base food cost 0.06, starts Primitive Age

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

### Faith Domain — base food cost 0.08 (early tiers, Primitive–Classical); 2.0 (formal tier from Medieval Age)

Early tiers (Primitive: Devotee · Stone: Believer · Bronze: Worshipper · Iron: Celebrant · Classical: Initiate) cover Shrine/Altar buildings at lower food costs. Formal tier begins at Medieval Age.

Medieval: Acolyte → Renaissance: Monk → Colonial: Missionary → Industrial: Revivalist → Victorian: Parish Priest → Electric: Evangelical → Atomic: Atomic Priest → Modern: Modern Shepherd → Information: Digital Devotee → Digital: Virtual Cleric → Cyberpunk: Cyber Cleric → Fusion: Plasma Prophet → Space: Star Preacher → Interstellar: Interstellar Mystic → Galactic: Galactic High Priest → **Quantum: Quantum Sage**

### Military Domain — base food cost 2.0, starts Iron Age

Iron: Soldier → Classical: Legionary → Medieval: Knight → Renaissance: Musketeer → Colonial: Colonial Marine → Industrial: Industrial Rifleman → Victorian: Victorian Guard → Electric: Electric Trooper → Atomic: Atomic Soldier → Modern: Modern Soldier → Information: Information Warrior → Digital: Digital Soldier → Cyberpunk: Cyber Warrior → Fusion: Plasma Trooper → Space: Space Marine → Interstellar: Interstellar Commando → Galactic: Galactic Guardian → **Quantum: Quantum Soldier**

### Trade Domain — base food cost 1.0, starts Bronze Age

Bronze: Peddler → Iron: Merchant → Classical: Trader → Medieval: Nobleman → Renaissance: Banker → Colonial: Colonial Merchant → Industrial: Industrialist → Victorian: Victorian Trader → Electric: Electric Broker → Atomic: Atomic Trader → Modern: Corporate Trader → Information: Digital Trader → Digital: Crypto Broker → Cyberpunk: Cyber Dealer → Fusion: Plasma Merchant → Space: Space Trader → Interstellar: Interstellar Broker → Galactic: Galactic Merchant → **Quantum: Quantum Dealer**

### Metallurgy Domain — base food cost 2.0, starts Iron Age

Iron: Smelter → Classical: Ironworker → Medieval: Medieval Smith → Renaissance: Renaissance Metallurgist → Colonial: Foundry Worker → Industrial: Factory Worker → Victorian: Steam Smelter → Electric: Electric Smelter → Atomic: Atomic Metallurgist → Modern: Modern Metallurgist → Information: Digital Foundry Worker → Digital: Digital Smelter → Cyberpunk: Cyber Forge Worker → Fusion: Plasma Metallurgist → Space: Stellar Foundry Worker → Interstellar: Stellar Smelter → Galactic: Galactic Metallurgist → **Quantum: Quantum Smelter**

### Engineering Domain — early tiers base 0.50/tick (Bronze–Industrial), formal tiers base 8.0 from Victorian Age

Bronze: Apprentice → Iron: Craftsman → Classical: Artisan → Medieval: Engineer → Renaissance: Master Eng. → Colonial: Mechanic → Industrial: Machinist → Victorian: Tinker → Electric: Electrical Engineer → Atomic: Nuclear Engineer → Modern: Systems Engineer → Information: Software Engineer → Digital: AI Engineer → Cyberpunk: Cyber Engineer → Fusion: Plasma Engineer → Space: Space Engineer → Interstellar: Warp Engineer → Galactic: Galactic Engineer → **Quantum: Quantum Engineer**

### Energy Domain — base food cost 8.0, starts Victorian Age

Victorian: Stoker → Electric: Power Worker → Atomic: Reactor Technician → Modern: Power Engineer → Information: Grid Operator → Digital: Digital Power Manager → Cyberpunk: Cyber Energy Worker → Fusion: Fusion Technician → Space: Solar Engineer → Interstellar: Dark Energy Worker → Galactic: Antimatter Specialist → **Quantum: Zero-Point Engineer**

### Hacker Domain — base food cost 16.0, starts Information Age

Information: Script Kiddie → Digital: Coder → Cyberpunk: Black Hat → Fusion: AI Hacker → Space: Orbital Hacker → Interstellar: Interstellar Netrunner → Galactic: Galactic Hacker → **Quantum: Quantum Hacker**

### Astronaut Domain — base food cost 32.0, starts Space Age

Space: Cadet → Interstellar: Interstellar Pilot → Galactic: Galactic Explorer → **Quantum: Quantum Astronaut**

---

## Commands

**Recruit workers:**

```
recruit [count|max]
```

Examples: `recruit`, `recruit 5`, `recruit max`

Workers are recruited generically from available housing capacity — no domain argument is needed. A freshly recruited worker has no domain class yet; they take on the domain class of the building they are first assigned to (e.g. assigned to a gathering_camp → becomes a Forager; assigned to a story_circle → becomes a Shaman). You can only recruit up to your current housing cap. Recruiting costs food upfront; each new worker begins draining food per tick immediately.

**Assign workers to a building:**

```
assign <building_key> [count|all]
```

Example: `assign gathering_camp 5`

The domain is inferred automatically from the building. A building with worker capacity 15, built 3 times, has 45 total slots.

**Unassign workers:**

```
unassign <building_key> [count|all]
```

Returns workers to the idle pool. They still drain food while idle.

**Permanently dismiss workers:**

```
dismiss <building_key> [count|all]
```

`dismiss` removes workers from a building AND from the population pool entirely — it reduces total population, not just reassigns them to idle. Use it to free up housing capacity or cut food drain when you have more workers than you can sustain. Unlike `unassign`, dismissed workers are gone.

**View worker status:**

```
workers
```

Opens a three-panel overlay:
- **Summary** — total pop / max, idle count, housing remaining, food drain/tick, net food/tick (color-coded), and a break-even or starvation warning
- **Slot Utilization** — filled vs. total slots across all worker buildings, fill bar, top buildings by vacancy
- **Domain Breakdown** — per-domain class name and food cost, with per-building assignment bars showing assigned/capacity

A worker mini-box is also always visible in the sidebar showing pop/idle/housing/drain/net food at a glance.

---

## Starvation

When food hits 0 and net food income is negative, workers begin dying:

- A warning is logged on the first tick of deficit
- 1 worker is killed every 5 ticks until food recovers
- Deaths are logged in red
- When food returns to positive net income, starvation stops and a recovery message is shown

This is intentional Malthusian gameplay — overpopulating without food infrastructure is punished. The fastest recovery is to `dismiss` workers from low-priority buildings to immediately reduce food drain, or to `unassign` expensive-domain workers and redirect them to food buildings.

---

## Morale

Morale is a percentage multiplier on all worker-driven building output. It **starts at 50%** (neutral), bottoms out at a **10% floor**, and is capped at **100% + 5% per wonder** built.

```
output = base_rate × building_count × (0.20 + 0.80 × assigned / capacity) × morale_multiplier
```

It is a two-way dial: below **25%** it penalises production (ramping down to **×0.50** at the floor), **25–75%** has no effect, and above **75%** it rewards production (ramping up to **+20%**). Morale drifts back toward 50% each tick, so the bonus must be sustained.

- **Rises from:** worship and culture (morale-restoring) buildings, good events, age advances
- **Falls from:** starvation, over-militarisation (military > 30% of pop), too many idle workers (> 50% of pop), bad events and catastrophes

Morale shows as a colored bar in the Workers panel and the villager sidebar.

See [Morale](morale.md) for the full details.

---

## Food Drain

Every worker in every domain costs food per tick. The exact amount is `baseFoodCost × 1.12^tier` for their current class tier. As you advance ages and recruit higher-tier workers, food drain rises — but the curve is much gentler than it used to be (tier-20 is ~9.6× base, not 3325×).

- Tier 0 food workers (Foragers) cost **0.06 food/tick** each — extremely cheap
- Tier 5 food workers (Medieval Serf) cost `0.06 × 1.12^5 ≈ 0.11 food/tick`
- Other domain base costs: knowledge/lumber/masonry/trade = 1.0, faith/military/metallurgy = 2.0, engineering/energy = 8.0, hacker = 16.0, astronaut = 32.0
- High-tier specialists (Hacker at tier 0 = 16.0, Astronaut at tier 0 = 32.0) are expensive — ensure food production keeps up before recruiting them

---

## Tips

- **Prioritize food workers early** — every other domain drains food. A food deficit stalls recruitment and can collapse your economy.
- **Faith workers before epoch transitions** — your faith level as a percentage of storage cap determines epoch roll odds. Keep faith workers assigned ahead of predicted epoch events.
- **Knowledge workers for fast research** — the knowledge domain multiplier doubles each tier, making late-game knowledge workers vastly more productive than early-tier ones.
- **Metallurgy requires both extraction and refining** — assign masonry workers to geological extraction buildings to produce raw ore, then metallurgy workers to smelters to refine it.
- **Single pool** — all workers are in one pool regardless of which buildings they are assigned to. Reassign freely between any buildings at any time.
- **Check idle count** — press `e` (Economy tab) to see the idle count in the status bar. Unassigned workers still drain food for zero extra production benefit.
- **Use `dismiss` to cut population** — if you are in a food deficit and can't produce enough, `dismiss` workers from non-critical buildings to permanently reduce drain. `unassign` alone does not help; idle workers still cost food.
- **Use `workers` for a full picture** — the overlay shows net food/tick and per-domain breakdown, making it easy to spot which domain is draining the most.
