# Workers & Domains: Reference

Workers are your civilization's labor force. Every building that produces resources has worker slots; staffing it raises output from a 20% floor up to 100%. Workers come from a **single generic pool**, and any worker can go to any building. The domain comes from the building, not the worker.

For a gentler introduction see [Workers](villagers.md).

---

## Overview

- One pool of workers, all identical until assigned to a building
- An assigned worker shows the **class name** for that building's domain at the current age (a label only)
- **Production formula:** `base_rate × building_count × (0.20 + 0.80 × assigned / total_capacity)`
- 20% floor: a building with no workers still produces something
- Food drain: every worker eats the same amount per tick, set by the current age

---

## The Worker Pool

There is a single pool of workers. Workers have no domain until they are assigned to a building. Your population is:

```
total_workers = idle_workers + sum of all building assignments
```

An assigned worker takes on the **class name** for that building's domain at the current age. For example, a worker on a `gathering_camp` in the Primitive Age is a **Forager**; the same worker on a `story_circle` is a **Shaman**. The class name is only a label. There are no separate domain pools.

What to know:

- You don't choose a domain when recruiting.
- Workers can move freely between any buildings that have worker slots.
- The `status` command and the Economy panel show population, idle count and current food drain.
- Your population is limited by housing (huts, longhouses and so on).

### Class Name Quick Reference

The class name shown in the game comes from the building's domain and the current age. Selected examples:

| Domain | Primitive | Iron | Classical | Medieval | Victorian | Modern |
|--------|-----------|------|-----------|----------|-----------|--------|
| food | Forager | Laborer | Peasant | Serf | Agric. Worker | Modern Farmer |
| knowledge | Shaman | Scholar | Philosopher | Friar | Victorian Scholar | Modern Researcher |
| military | - | Soldier | Legionary | Knight | Victorian Guard | Modern Soldier |
| trade | - | Merchant | Trader | Nobleman | Victorian Trader | Corporate Trader |
| faith | Devotee | Celebrant | Initiate | Acolyte | Parish Priest | Modern Shepherd |
| engineering | - | Craftsman | Artisan | Engineer | Tinker | Systems Engineer |
| lumber | - | Lumberjack | Sawyer | Forester | Steam Logger | Petroleum Worker |
| masonry | - | Miner | Iron Extractor | Medieval Miner | Victorian Quarryman | Modern Geologist |
| metallurgy | - | Furnace Hand | Ironworker | Medieval Smith | Steam Smelter | Modern Metallurgist |
| energy | - | - | - | - | Stoker | Power Engineer |
| hacker | - | - | - | - | - | - |
| astronaut | - | - | - | - | - | - |

The full class progression for every domain is in [Worker Class Names by Age](#worker-class-names-by-age) below.

---

## Recruiting Workers

```
recruit [count|max]
```

Workers are recruited from free housing. You don't name a domain; workers are generic until assigned.

| Command | Effect |
|---------|--------|
| `recruit` | Recruit 1 worker |
| `recruit 5` | Recruit 5 workers |
| `recruit max` | Recruit as many as your housing allows |

**Food cost:** recruiting is free, but every worker eats food each tick from the moment it joins. The amount is the current age's rate for all workers, whatever building they staff: **0.06 food/tick** per worker in the Primitive Age. See [Food Drain](#food-drain).

**`recruit max`** fills all your free housing. It does not check your food income, so it can recruit more workers than your food supports and put you into a deficit.

**Viewing workers:** type `workers` to open the Workers panel (morale, summary, slot utilization, domain breakdown), or `status` for a text summary. The Workers panel shows net food/tick in green or red (blue or orange on the accessible [themes](commands.md#themes)), how many workers your food can sustain or a deficit warning, and per-building fill bars. The Workers box in the sidebar is always visible and shows population, idle, housing left, drain and net food at a glance.

---

## Assigning Workers to Buildings

```
assign <building> [count|all]
```

Assigns workers from the idle pool to the building. The domain comes from the building; you never specify one.

| Command | Effect |
|---------|--------|
| `assign gathering_camp 3` | Assign 3 workers to gathering_camp |
| `assign library all` | Assign all idle workers to library |
| `assign barracks 5` | Assign 5 workers to barracks |
| `assign` (no arguments) | Shows the usage line |

Rules:

- You must have built at least one of the building. You can't assign to a building you haven't built.
- Assignment is capped at the building count times its worker slots.
- You can't assign more workers than you have idle.
- Type `assign ` and the prompt suggests your built buildings that take workers, the ones with a free slot first (`Tab` takes the suggestion).

**Production scaling formula:**

```
output = base_rate × building_count × (0.20 + 0.80 × assigned / total_capacity)
```

Here `total_capacity` is the building count times the building's worker slots.

| Workers Assigned | Efficiency |
|-----------------|------------|
| 0 (none) | 20% |
| 25% of slots | 40% |
| 50% of slots | 60% |
| 75% of slots | 80% |
| 100% (full staff) | 100% |

The 0.20 floor means an unstaffed building always contributes something. Only full staffing reaches the full rate.

`base_rate` is the fully staffed rate shown in the building's description. For construction resources it is set by the Payback Rule (fully staffed, a producer earns back its first copy's price within the age's payback time); food, faith, culture and soldiers keep hand-set rates. See [How Production Rates Are Set](buildings.md#how-production-rates-are-set).

**Example: gathering_camp**

- Base rate 1.0 food/tick, 3 worker slots
- Built 2×: 6 slots in total, base output 1.0 × 2 = 2.0 food/tick
- With 3 workers assigned (50% fill): `2.0 × (0.20 + 0.80 × 0.50)` = **1.2 food/tick**
- With all 6 slots filled (100%): **2.0 food/tick**

**Example: library (Classical Age, knowledge domain)**

- Base rate 3.2 knowledge/tick, 4 worker slots
- Built 3×: 12 slots in total, base output 3.2 × 3 = 9.6 knowledge/tick
- With `assign library 8` (67% fill): `9.6 × (0.20 + 0.80 × 0.67)` ≈ **7.0 knowledge/tick** (73% efficiency)
- With all 12 slots filled: **9.6 knowledge/tick**

---

## Unassigning Workers

```
unassign <building> [count|all]
```

Removes workers from a building and returns them to the idle pool.

| Command | Effect |
|---------|--------|
| `unassign gathering_camp 2` | Remove 2 workers from gathering_camp |
| `unassign barracks all` | Remove all workers from barracks |
| `unassign library 5` | Remove 5 workers from library |

Unassigned workers go back to the idle pool at once and can be reassigned elsewhere. They keep eating while idle.

When to unassign:

- To move workers from low-priority buildings to new, higher-tier ones after an age advance.
- To pull military workers off military buildings once you've banked enough soldiers. Campaigns spend the `soldiers` resource, not workers, so idle military workers only cost food and morale.

---

## Dismissing Workers

```
dismiss <building> [count|all]
```

`dismiss` permanently removes workers from a building **and** from your population. Unlike `unassign`, dismissed workers are gone; population drops at once.

| Command | Effect |
|---------|--------|
| `dismiss gathering_camp 2` | Remove 2 workers from gathering_camp and reduce population by 2 |
| `dismiss barracks all` | Dismiss all workers assigned to barracks |

When to dismiss:

- **Food deficit.** Moving workers to idle with `unassign` doesn't help, because idle workers still eat. `dismiss` is the only way to cut drain without changing food production.
- **Housing pressure.** Free housing for workers you'd rather have elsewhere.
- **Late-game cleanup.** Remove workers from early buildings that no longer earn their housing.

---

## Starvation

When food runs out while your workers are still eating, they begin dying:

- A warning is logged on the first tick without food
- **1 worker dies every 5 ticks** (10 seconds at 1x) until food recovers
- Deaths are logged in red
- As soon as there is food in stock again, deaths stop and a recovery message is logged

Overpopulating without the food to back it up costs you workers. To recover:

1. `dismiss` workers from low-priority buildings to cut drain at once
2. Build more food buildings and assign workers to them
3. `unassign` workers from other buildings and `assign` them to food buildings

---

## The 12 Worker Domains

| Domain Key | Display Name | Primary Output | Example Buildings | Unlocks |
|------------|--------------|---------------|-------------------|---------|
| `food` | Food | food | gathering_camp, forager_post, farm, field_works | Primitive Age |
| `knowledge` | Knowledge | knowledge | story_circle, elders_hall, scriptorium, library | Primitive Age |
| `lumber` | Organic Extraction | wood / coal / oil / quantum_flux | logging_camp, lumber_mill, sawmill | Stone Age |
| `masonry` | Geological Extraction | stone / iron_ore / uranium / titanium_ore | stone_pit, quarry, iron_mine | Stone Age |
| `faith` | Faith | faith | shrine, altar, temple, cathedral | Primitive Age (early class) |
| `military` | Military | soldiers | war_camp, barracks, hunting_lodge, legion_fort | Iron Age |
| `trade` | Trade | gold | market_stall, bazaar, trading_post | Bronze Age |
| `engineering` | Engineering | iron / steel / electricity / plasma | smithy, workshop, forge, factory | Bronze Age (early class) |
| `metallurgy` | Metallurgy | iron / steel / titanium / dark_matter | smelter, blast_furnace, steel_mill | Iron Age |
| `energy` | Energy | electricity / plasma / quantum_flux | coal_plant, power_station, fusion_reactor | Victorian Age |
| `hacker` | Hacker | data / crypto | server_farm, darknet_hub, quantum_core | Information Age |
| `astronaut` | Astronaut | dark_matter / antimatter | launch_pad, space_station | Space Age |

> **Culture** (Lineage 10) has no worker domain. Culture buildings produce automatically and take no workers.

---

## Worker Class Names by Age

Each domain has one class per age it spans. The class name is a label: it doesn't change what a worker produces or eats. Food drain depends only on the age (see [Food Drain](#food-drain)).

### food (starts Primitive Age)

The food domain's column doubles as the food each worker eats per tick in that age, since that rate applies to every worker.

| Age | Class Name | Food/tick per worker |
|-----|-----------|-----------|
| Primitive | Forager | 0.060 |
| Stone | Farmhand | 0.067 |
| Bronze | Cultivator | 0.075 |
| Iron | Laborer | 0.084 |
| Classical | Peasant | 0.094 |
| Medieval | Serf | 0.106 |
| Renaissance | Plowman | 0.118 |
| Colonial | Colonial Farmer | 0.133 |
| Industrial | Factory Hand | 0.149 |
| Victorian | Agricultural Worker | 0.166 |
| Electric | Electric Farmer | 0.186 |
| Atomic | Atomic Agronomist | 0.209 |
| Modern | Modern Farmer | 0.234 |
| Information | Digital Cultivator | 0.262 |
| Digital | AI Agronomist | 0.293 |
| Cyberpunk | Aug Harvester | 0.328 |
| Fusion | Bio-Farmer | 0.368 |
| Space | Zero-G Farmer | 0.412 |
| Interstellar | Stellar Cultivator | 0.461 |
| Galactic | Galactic Farmer | 0.517 |
| Quantum | Quantum Harvester | 0.579 |

### knowledge (starts Primitive Age)

Primitive: Shaman → Stone: Elder → Bronze: Scribe → Iron: Scholar → Classical: Philosopher → Medieval: Friar → Renaissance: Academician → Colonial: Naturalist → Industrial: Engineer-Scientist → Victorian: Victorian Scholar → Electric: Research Fellow → Atomic: Nuclear Scientist → Modern: Modern Researcher → Information: Data Scientist → Digital: AI Researcher → Cyberpunk: Cyber-Scholar → Fusion: Fusion Theorist → Space: Orbital Researcher → Interstellar: Stellar Scientist → Galactic: Galactic Researcher → **Quantum: Quantum Theorist**

### lumber (starts Stone Age)

Stone: Gatherer → Bronze: Woodcutter → Iron: Lumberjack → Classical: Sawyer → Medieval: Forester → Renaissance: Renaissance Logger → Colonial: Mill Worker → Industrial: Coal Extractor → Victorian: Steam Logger → Electric: Electric Forester → Atomic: Fuel Extractor → Modern: Petroleum Worker → Information: Digital Forester → Digital: Bio-Extractor → Cyberpunk: Nano-Harvester → Fusion: Organic Engineer → Space: Biofield Harvester → Interstellar: Quantum Extractor → Galactic: Galactic Forester → **Quantum: Cosmic Extractor**

### masonry (starts Stone Age)

Stone: Quarryman → Bronze: Stone Cutter → Iron: Miner → Classical: Iron Extractor → Medieval: Medieval Miner → Renaissance: Renaissance Quarryman → Colonial: Colonial Miner → Industrial: Industrial Miner → Victorian: Victorian Quarryman → Electric: Electric Miner → Atomic: Uranium Miner → Modern: Modern Geologist → Information: Precision Miner → Digital: Digital Excavator → Cyberpunk: Cyber Miner → Fusion: Plasma Driller → Space: Space Miner → Interstellar: Asteroid Miner → Galactic: Dark Matter Extractor → **Quantum: Crystal Miner**

### faith (early tiers from Primitive, formal tiers from Medieval)

The early tiers cover shrines and altars before the formal domain starts at the Medieval Age.

| Age | Class Name |
|-----|-----------|
| Primitive | Devotee |
| Stone | Believer |
| Bronze | Worshipper |
| Iron | Celebrant |
| Classical | Initiate |
| Medieval | Acolyte |
| Renaissance | Monk |
| Colonial | Missionary |
| Industrial | Revivalist |
| Victorian | Parish Priest |
| Electric | Evangelical |
| Atomic | Atomic Priest |
| Modern | Modern Shepherd |
| Information | Digital Devotee |
| Digital | Virtual Cleric |
| Cyberpunk | Cyber Cleric |
| Fusion | Plasma Prophet |
| Space | Star Preacher |
| Interstellar | Interstellar Mystic |
| Galactic | Galactic High Priest |
| Quantum | Quantum Sage |

### military (starts Iron Age)

Iron: Soldier → Classical: Legionary → Medieval: Knight → Renaissance: Musketeer → Colonial: Colonial Marine → Industrial: Industrial Rifleman → Victorian: Victorian Guard → Electric: Electric Trooper → Atomic: Atomic Soldier → Modern: Modern Soldier → Information: Information Warrior → Digital: Digital Soldier → Cyberpunk: Cyber Warrior → Fusion: Plasma Trooper → Space: Space Marine → Interstellar: Interstellar Commando → Galactic: Galactic Guardian → **Quantum: Quantum Soldier**

### trade (starts Bronze Age)

Bronze: Peddler → Iron: Merchant → Classical: Trader → Medieval: Nobleman → Renaissance: Banker → Colonial: Colonial Merchant → Industrial: Industrialist → Victorian: Victorian Trader → Electric: Electric Broker → Atomic: Atomic Trader → Modern: Corporate Trader → Information: Digital Trader → Digital: Crypto Broker → Cyberpunk: Cyber Dealer → Fusion: Plasma Merchant → Space: Space Trader → Interstellar: Interstellar Broker → Galactic: Galactic Merchant → **Quantum: Quantum Dealer**

### engineering (early tiers from Bronze, later tiers from Victorian)

| Age | Class Name |
|-----|-----------|
| Bronze | Apprentice |
| Iron | Craftsman |
| Classical | Artisan |
| Medieval | Engineer |
| Renaissance | Master Engineer |
| Colonial | Mechanic |
| Industrial | Machinist |
| Victorian | Tinker |
| Electric | Electrical Engineer |
| Atomic | Nuclear Engineer |
| Modern | Systems Engineer |
| Information | Software Engineer |
| Digital | AI Engineer |
| Cyberpunk | Cyber Engineer |
| Fusion | Plasma Engineer |
| Space | Space Engineer |
| Interstellar | Warp Engineer |
| Galactic | Galactic Engineer |
| Quantum | Quantum Engineer |

### metallurgy (starts Iron Age)

Iron: Furnace Hand → Classical: Ironworker → Medieval: Medieval Smith → Renaissance: Renaissance Metallurgist → Colonial: Foundry Worker → Industrial: Factory Worker → Victorian: Steam Smelter → Electric: Electric Smelter → Atomic: Atomic Metallurgist → Modern: Modern Metallurgist → Information: Digital Foundry Worker → Digital: Digital Smelter → Cyberpunk: Cyber Forge Worker → Fusion: Plasma Metallurgist → Space: Stellar Foundry Worker → Interstellar: Stellar Smelter → Galactic: Galactic Metallurgist → **Quantum: Quantum Smelter**

### energy (starts Victorian Age)

Victorian: Stoker → Electric: Power Worker → Atomic: Reactor Technician → Modern: Power Engineer → Information: Grid Operator → Digital: Digital Power Manager → Cyberpunk: Cyber Energy Worker → Fusion: Fusion Technician → Space: Solar Engineer → Interstellar: Dark Energy Worker → Galactic: Antimatter Specialist → **Quantum: Zero-Point Engineer**

### hacker (starts Information Age)

Information: Script Kiddie → Digital: Coder → Cyberpunk: Black Hat → Fusion: AI Hacker → Space: Orbital Hacker → Interstellar: Interstellar Netrunner → Galactic: Galactic Hacker → **Quantum: Quantum Hacker**

### astronaut (starts Space Age)

Space: Cadet → Interstellar: Interstellar Pilot → Galactic: Galactic Explorer → **Quantum: Quantum Astronaut**

---

## Morale

Morale is a civilization-wide percentage that multiplies **all** worker-driven building output every tick.

**Output formula with morale:**

```
output = base_rate × building_count × (0.20 + 0.80 × assigned / total_capacity) × morale_multiplier
```

### The scale

- A new civilization **starts at 50%** (neutral)
- Floor: **10%**
- Cap: **100% + 5% per wonder built** (100% with no wonders)

### The production curve

Morale's effect is a continuous curve centered on 50%. Any move away from 50% has an effect, small near the center and larger toward the ends:

| Morale | Effect |
| ------ | -------- |
| 10% (the floor) | **×0.50**, half output |
| Below 50% | Production **penalty**, easing linearly from ×0.50 at the floor to ×1.00 at 50% |
| Exactly 50% | **×1.00**, normal output |
| Above 50% | Production **bonus**, rising linearly from ×1.00 at 50% to **+20%** at the cap |

Morale **drifts gently back toward 50% every tick**. A bonus has to be earned and kept up, and it fades if you stop; a penalty heals itself once you remove the cause.

### What raises morale

- **Worship and culture buildings.** The shrine and temple line and the culture buildings of each age lift morale every tick once built; they need no workers.
- **Your faith production rate**, up to a per-tick limit
- **Good events**
- **Advancing to a new age**

### What lowers morale

- **Starvation**
- **Military workers over 30% of population.** The further over, the faster the drain.
- **More than 50% of workers idle**
- **Bad events and catastrophes.** Enduring a catastrophe costs morale.

### Where it shows

Morale appears as a **colored bar** in the Workers panel (`workers`) and as `Morale: NN%` in the status bar: green when it boosts production, red when it penalizes it. Each save's detail pane in the Load Game browser also shows it.

### Tips

- Keep food income positive.
- Keep military workers under 30% of population.
- Don't leave half your population idle.
- Build worship and culture buildings to push morale toward the **+20% bonus**. This is the main lever.
- Build wonders to raise the cap.

See [Morale](morale.md) for the full morale page.

---

## Food Drain

Every worker eats food each tick. The amount is set by the current age and is the same for every worker, whatever building it staffs:

```
total_food_drain = food_per_worker(current age) × population
```

In the Primitive Age a worker eats **0.06 food/tick**. The rate rises 12% with each age, to about **0.23** in the Modern Age and **0.58** in the Quantum Age. When you advance, the new rate applies to every worker at once. The full per-age list is in the food table under [Worker Class Names by Age](#worker-class-names-by-age).

**Food drain table (by population):**

| Age | Food/tick per worker | 10 workers | 50 workers |
|-----|---------------------|-----------|-----------|
| Primitive | 0.060 | 0.60 | 3.00 |
| Stone | 0.067 | 0.67 | 3.36 |
| Bronze | 0.075 | 0.75 | 3.76 |
| Iron | 0.084 | 0.84 | 4.22 |
| Classical | 0.094 | 0.94 | 4.72 |
| Medieval | 0.106 | 1.06 | 5.29 |
| Victorian | 0.166 | 1.66 | 8.32 |
| Modern | 0.234 | 2.34 | 11.69 |

**What happens at food = 0:** starvation begins. One worker dies every 5 ticks until there is food in stock again. A warning is logged on the first tick without food and each death is shown in red; when food recovers, a recovery message is logged and deaths stop.

Practical advice:

- Make sure food income exceeds food drain before recruiting more workers.
- The `status` command and the Workers panel both show food drain per tick and net food/tick.
- `unassign` returns workers to the idle pool, where they still eat. Use `dismiss` to remove workers from your population and reduce drain at once.

---

## Worker Building Reference

Key buildings that take workers, grouped by domain:

| Building | Key | Domain | Worker Slots | Available From |
|----------|-----|--------|----------------|----------------|
| Gathering Camp | `gathering_camp` | food | 3 | Primitive Age |
| Forager Post | `forager_post` | food | 4 | Stone Age |
| Farm | `farm` | food | 5 | Bronze Age |
| Story Circle | `story_circle` | knowledge | 2 | Primitive Age |
| Elders' Hall | `elders_hall` | knowledge | 2 | Stone Age |
| Library | `library` | knowledge | 4 | Classical Age |
| War Camp | `war_camp` | military | 3 | Stone Age |
| Barracks | `barracks` | military | 4 | Bronze Age |
| Hunting Lodge | `hunting_lodge` | military | 5 | Iron Age |

Worker slots are **per building**. If you have built 3 libraries, you have 3 × 4 = 12 slots. Use `assign library all` to fill them.

For a full per-lineage building list see [Buildings](buildings.md).

---

## Strategy

### Early game (Primitive / Stone Age)

Recruit 5-10 workers right away and assign them to `gathering_camp`. Each camp holds 3 workers, so build 2-3 camps before recruiting past 9. Foragers eat 0.06 food/tick each, and a fully staffed camp produces 1.0 food/tick, so it covers its own three workers' drain (0.18) several times over.

```
build gathering_camp
recruit 3
assign gathering_camp 3
build gathering_camp
recruit 3
assign gathering_camp 3
```

Once food is stable, assign some workers to `story_circle` (knowledge) to start on research.

### Mid game (Bronze / Iron Age)

- Staff lumber and masonry buildings as soon as they are built. Stone and wood go into most building costs.
- Build a `war_camp` and staff it with workers. It produces the `soldiers` resource (from the Iron Age) that you spend on campaigns.
- Knowledge workers in libraries matter more and more. The `scholars_haven` milestone needs 50 knowledge workers and 3 libraries.

### Late game

- As food buildings get more productive, move food workers to other domains.
- Every worker eats the age's rate, and that rate keeps rising, so a large late-game population needs a large food income. Scale food before you scale headcount.
- Use `unassign <building> all` and then `assign <new building> all` to move your workforce quickly after an age advance.

### Milestone: scholars_haven

Needs **50 knowledge workers** assigned to knowledge buildings and **3 Libraries** built. Track progress with `milestones` or `ms`.

```
assign library 50
```

### Specialize vs. spread

Focusing on one domain can finish milestone chains faster. Spreading across domains protects you from resource shortages but delays milestones. Early on, food and knowledge are enough; add military when campaigns open up and trade when gold becomes the bottleneck.

---

## Completion Reference

The prompt suggests completions for all worker commands, shown dim after the cursor; `Tab` takes one and cycles to the next (see [The prompt](commands.md#the-prompt)):

| Typed | Suggestions |
|-------|----------------|
| `assign ` | your built buildings that have worker slots, the ones with a free slot first |
| `assign gathering_camp ` | `all` |
| `assign lib` | built buildings whose key starts with `lib`, such as `library` |
| `unassign ` | only buildings that currently have workers assigned |
| `unassign barracks ` | `all` |
| `dismiss ` | only buildings that currently have workers assigned |
| `dismiss barracks ` | `all` |
| `recruit ` | `max` |

For `assign` the prompt only suggests buildings you have **built**, and for `unassign` and `dismiss` only buildings that have workers to remove.

---

## What Does Not Exist

These command forms are **not valid** and return an error:

```
recruit food            # INVALID: no domain argument
recruit military 5      # INVALID: no domain argument
assign food gathering_camp  # INVALID: the domain comes from the building
unassign all food       # INVALID: name a building
```

---

## See Also

- [Workers](villagers.md): introductory guide to the worker system
- [Buildings](buildings.md): building lineages and worker slots per building
- [Resources](resources.md): which resources each lineage produces in which age
- [Epochs](epochs.md): how epoch transitions affect building output resources
- [Milestones](milestones.md): milestone chains that reward domain specialization
