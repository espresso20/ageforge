# Workers & Domains — Reference

Workers are your civilization's labour force. Every building that produces resources has a worker capacity; staffing it drives output from a 20% idle floor up to 100% full efficiency. Workers are managed through a **single generic pool** — any worker can go to any building. The domain is resolved automatically from the building, not from the worker.

For a gentler introduction see [Workers](villagers.md).

---

## Overview

- One pool of workers — all identical until assigned to a building
- Assigning a worker to a building gives them the **class name** for that building's domain at the current age (purely cosmetic / food-cost reference)
- **Production formula:** `base_rate × building_count × (0.20 + 0.80 × assigned / total_capacity)`
- 20% floor: unassigned buildings are never completely idle
- Food drain: all workers cost food per tick, scaled by current age tier

---

## The Worker Pool

There is a single pool of workers. Workers have no domain until they are assigned to a building. The total population is:

```
total_workers = idle_workers + sum of all building assignments
```

When you assign a worker to a building, they take on the **class name** for that building's domain at the current age — for example, a worker assigned to a `gathering_camp` in the primitive age becomes a **Forager**; the same worker assigned to a `story_circle` becomes a **Shaman**. The displayed class name is cosmetic and used to look up food cost scaling, but there are no separate domain pools.

**Key points:**

- You do not choose which domain a worker belongs to when recruiting
- Workers can be reassigned freely between any buildings with worker capacity
- The `status` command and Economy tab show total population, idle count, and current food drain
- Population is capped by your housing buildings (huts, longhouses, etc.)

### Class Name Quick Reference

The class name shown in the UI reflects the building's domain + current age. Selected examples:

| Domain | Primitive | Iron | Classical | Medieval | Victorian | Modern |
|--------|-----------|------|-----------|----------|-----------|--------|
| food | Forager | Laborer | Peasant | Serf | Agric. Worker | Modern Farmer |
| knowledge | Shaman | Scholar | Philosopher | Friar | Victorian Scholar | Modern Researcher |
| military | — | Soldier | Legionary | Knight | Victorian Guard | Modern Soldier |
| trade | — | Merchant | Trader | Nobleman | Victorian Trader | Corporate Trader |
| faith | Devotee | Celebrant | Initiate | Acolyte | Parish Priest | Modern Shepherd |
| engineering | — | Craftsman | Artisan | Engineer | Tinker | Systems Engineer |
| lumber | — | Lumberjack | Sawyer | Forester | Steam Logger | Petroleum Worker |
| masonry | — | Miner | Iron Extractor | Medieval Miner | Victorian Quarryman | Modern Geologist |
| metallurgy | — | Smelter | Ironworker | Medieval Smith | Steam Smelter | Modern Metallurgist |
| energy | — | — | — | — | Stoker | Power Engineer |
| hacker | — | — | — | — | — | — |
| astronaut | — | — | — | — | — | — |

Full class progression for every domain is in the [Worker Class Names by Age](#worker-class-names-by-age) section below.

---

## Recruiting Workers

```
recruit [count|max]
```

Workers are recruited from available housing capacity. No domain argument — workers are generic until assigned.

| Command | Effect |
|---------|--------|
| `recruit` | Recruit 1 worker |
| `recruit 5` | Recruit 5 workers |
| `recruit max` | Recruit as many as housing cap allows |

**Food cost:** each worker drains food per tick immediately on recruitment, based on the current age's tier for the food domain. In primitive age this is **0.06 food/tick** per worker — very cheap. The food cost is always determined by the food domain's class for the current age, regardless of which building the worker is assigned to.

**`max` behaviour:** attempts to fill all remaining housing slots (up to population cap). It does not check your food income — you can recruit more workers than your food supports, putting you into a deficit.

**Viewing workers:** type `workers` for the full three-panel overlay (summary, slot utilization, domain breakdown), or press `e` for the Economy tab. The `workers` overlay shows net food/tick color-coded green/red (blue/orange on the accessible [themes](commands.md#themes)), a break-even or starvation warning, and per-building fill bars. A worker mini-box is always visible in the sidebar showing pop/idle/housing/drain/net at a glance.

---

## Assigning Workers to Buildings

```
assign <building_key> [count|all]
```

Assigns workers from the idle pool to the specified building. The domain is resolved automatically from the building definition — you never specify a domain.

| Command | Effect |
|---------|--------|
| `assign gathering_camp 3` | Assign 3 workers to gathering_camp |
| `assign library all` | Assign all idle workers to library |
| `assign barracks 5` | Assign 5 workers to barracks |
| `assign` (no args) | Shows usage error |

**Rules:**

- The building must have been **built at least once** (`Count > 0`) — you cannot assign to unbuilt buildings
- Assignment is capped at `building_count × WorkerCapacity` — the total staffing capacity
- You cannot assign more workers than you have idle
- Press TAB after `assign ` to autocomplete available worker-accepting building keys

**Production scaling formula:**

```
output = base_rate × building_count × (0.20 + 0.80 × assigned / total_capacity)
```

Where `total_capacity = building_count × building.WorkerCapacity`.

| Workers Assigned | Efficiency |
|-----------------|------------|
| 0 (none) | 20% |
| 25% of capacity | 40% |
| 50% of capacity | 60% |
| 75% of capacity | 80% |
| 100% (full staff) | 100% |

The 0.20 floor means idle buildings always contribute something. Full staffing is required to reach full rate.

**Example — gathering_camp:**

- Base rate: 0.50 food/tick, WorkerCapacity: 3
- Built 2× → total capacity 6, base output: 0.50 × 2 = 1.0 food/tick
- With 3 workers assigned (50% fill): `1.0 × (0.20 + 0.80 × 0.50)` = **0.60 food/tick**
- With all 6 workers assigned (100% fill): **1.0 food/tick**

**Example — library (classical_age, knowledge domain):**

- WorkerCapacity: 4
- Built 3× → 12 total slots
- Base output: knowledge/tick
- Assign workers with `assign library 8` to reach 67% efficiency

---

## Unassigning Workers

```
unassign <building_key> [count|all]
```

Removes workers from a building and returns them to the idle pool.

| Command | Effect |
|---------|--------|
| `unassign gathering_camp 2` | Remove 2 workers from gathering_camp |
| `unassign barracks all` | Remove all workers from barracks |
| `unassign library 5` | Remove 5 workers from library |

Unassigned workers return to the idle pool immediately and can be reassigned elsewhere. They continue to drain food while idle.

**When to unassign:**

- Reassigning workers from low-priority buildings to new higher-tier ones after an age advance
- Pulling military workers off military buildings once you've banked enough soldiers — expeditions spend the `soldiers` resource, not workers, so idle military workers are pure food/morale drain

---

## Dismissing Workers

```
dismiss <building_key> [count|all]
```

`dismiss` permanently removes workers from a building **and** from the total population pool. Unlike `unassign`, dismissed workers are gone — population decreases immediately.

| Command | Effect |
|---------|--------|
| `dismiss gathering_camp 2` | Remove 2 workers from gathering_camp and reduce total pop by 2 |
| `dismiss barracks all` | Dismiss all workers assigned to barracks |

**When to dismiss:**

- Food deficit — cutting idle workers with `unassign` doesn't help because idle workers still drain food. `dismiss` is the only way to reduce drain without food production changes.
- Housing pressure — free up housing slots for more productive workers in a different domain
- Late-game cleanup — remove early-tier cheap workers that are no longer worth their housing slot

---

## Starvation

When food hits 0 and net food income is negative, workers begin dying:

- A warning is logged on the first tick of deficit
- **1 worker is killed every 5 ticks** until food net income recovers
- Deaths are logged in red
- When food returns to positive net income, deaths stop and a recovery message is shown

Starvation is intentional — overpopulating without food infrastructure is punished. Recovery options:

1. `dismiss` workers from low-priority buildings to immediately cut drain
2. Build more food production buildings and assign workers to them
3. `unassign` from expensive-domain buildings and `assign` them to food buildings instead

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

> **Culture** (Lineage 10) has no worker domain — culture buildings produce automatically and cannot be assigned workers.

---

## Worker Class Names by Age

Each domain has one class per age it spans. Class name is cosmetic but also determines the food cost lookup for workers assigned to that domain's buildings. Food costs scale geometrically: `FoodCost = baseFoodCost × 1.12^tier`.

### food — base 0.06/tick, starts Primitive Age

| Age | Class Name | Food/tick |
|-----|-----------|-----------|
| Primitive | Forager | 0.060 |
| Stone | Farmhand | 0.067 |
| Bronze | Cultivator | 0.075 |
| Iron | Laborer | 0.084 |
| Classical | Peasant | 0.095 |
| Medieval | Serf | 0.106 |
| Renaissance | Plowman | 0.119 |
| Colonial | Colonial Farmer | 0.133 |
| Industrial | Factory Hand | 0.149 |
| Victorian | Agricultural Worker | 0.167 |
| Electric | Electric Farmer | 0.187 |
| Atomic | Atomic Agronomist | 0.210 |
| Modern | Modern Farmer | 0.235 |
| Information | Digital Cultivator | 0.263 |
| Digital | AI Agronomist | 0.295 |
| Cyberpunk | Aug Harvester | 0.330 |
| Fusion | Bio-Farmer | 0.370 |
| Space | Zero-G Farmer | 0.415 |
| Interstellar | Stellar Cultivator | 0.465 |
| Galactic | Galactic Farmer | 0.521 |
| Quantum | Quantum Harvester | 0.583 |

### knowledge — base 1.0/tick, starts Primitive Age

Primitive: Shaman → Stone: Elder → Bronze: Scribe → Iron: Scholar → Classical: Philosopher → Medieval: Friar → Renaissance: Academician → Colonial: Naturalist → Industrial: Engineer-Scientist → Victorian: Victorian Scholar → Electric: Research Fellow → Atomic: Nuclear Scientist → Modern: Modern Researcher → Information: Data Scientist → Digital: AI Researcher → Cyberpunk: Cyber-Scholar → Fusion: Fusion Theorist → Space: Orbital Researcher → Interstellar: Stellar Scientist → Galactic: Galactic Researcher → **Quantum: Quantum Theorist**

### lumber — base 1.0/tick, starts Stone Age

Stone: Gatherer → Bronze: Woodcutter → Iron: Lumberjack → Classical: Sawyer → Medieval: Forester → Renaissance: Colonial Logger → Colonial: Mill Worker → Industrial: Coal Extractor → Victorian: Steam Logger → Electric: Electric Forester → Atomic: Fuel Extractor → Modern: Petroleum Worker → Information: Digital Forester → Digital: Bio-Extractor → Cyberpunk: Nano-Harvester → Fusion: Organic Engineer → Space: Biofield Harvester → Interstellar: Quantum Extractor → Galactic: Galactic Forester → **Quantum: Cosmic Extractor**

### masonry — base 1.0/tick, starts Stone Age

Stone: Quarryman → Bronze: Stone Cutter → Iron: Miner → Classical: Iron Extractor → Medieval: Medieval Miner → Renaissance: Renaissance Quarryman → Colonial: Colonial Miner → Industrial: Industrial Miner → Victorian: Victorian Quarryman → Electric: Electric Miner → Atomic: Uranium Miner → Modern: Modern Geologist → Information: Data Miner → Digital: Digital Excavator → Cyberpunk: Cyber Miner → Fusion: Plasma Driller → Space: Space Miner → Interstellar: Asteroid Miner → Galactic: Dark Matter Extractor → **Quantum: Crystal Miner**

### faith — early tiers 0.08→0.40/tick (Primitive–Classical), formal tiers base 2.0 from Medieval

Early tiers cover shrine/altar buildings at lower food costs before the formal domain kicks in at Medieval Age.

| Age | Class Name | Food/tick |
|-----|-----------|-----------|
| Primitive | Devotee | 0.08 |
| Stone | Believer | 0.12 |
| Bronze | Worshipper | 0.18 |
| Iron | Celebrant | 0.27 |
| Classical | Initiate | 0.40 |
| Medieval | Acolyte | 2.00 |
| Renaissance | Monk | 2.24 |
| Colonial | Missionary | 2.51 |
| Industrial | Revivalist | 2.81 |
| Victorian | Parish Priest | 3.15 |
| Electric | Evangelical | 3.52 |
| Atomic | Atomic Priest | 3.95 |
| Modern | Modern Shepherd | 4.42 |
| Information | Digital Devotee | 4.95 |
| Digital | Virtual Cleric | 5.54 |
| Cyberpunk | Cyber Cleric | 6.21 |
| Fusion | Plasma Prophet | 6.95 |
| Space | Star Preacher | 7.79 |
| Interstellar | Interstellar Mystic | 8.72 |
| Galactic | Galactic High Priest | 9.77 |
| Quantum | Quantum Sage | 10.94 |

### military — base 2.0/tick, starts Iron Age

Iron: Soldier → Classical: Legionary → Medieval: Knight → Renaissance: Musketeer → Colonial: Colonial Marine → Industrial: Industrial Rifleman → Victorian: Victorian Guard → Electric: Electric Trooper → Atomic: Atomic Soldier → Modern: Modern Soldier → Information: Information Warrior → Digital: Digital Soldier → Cyberpunk: Cyber Warrior → Fusion: Plasma Trooper → Space: Space Marine → Interstellar: Interstellar Commando → Galactic: Galactic Guardian → **Quantum: Quantum Soldier**

### trade — base 1.0/tick, starts Bronze Age

Bronze: Peddler → Iron: Merchant → Classical: Trader → Medieval: Nobleman → Renaissance: Banker → Colonial: Colonial Merchant → Industrial: Industrialist → Victorian: Victorian Trader → Electric: Electric Broker → Atomic: Atomic Trader → Modern: Corporate Trader → Information: Digital Trader → Digital: Crypto Broker → Cyberpunk: Cyber Dealer → Fusion: Plasma Merchant → Space: Space Trader → Interstellar: Interstellar Broker → Galactic: Galactic Merchant → **Quantum: Quantum Dealer**

### engineering — early tiers 0.50→5.60/tick (Bronze–Industrial), formal tiers base 8.0 from Victorian

| Age | Class Name | Food/tick |
|-----|-----------|-----------|
| Bronze | Apprentice | 0.50 |
| Iron | Craftsman | 0.75 |
| Classical | Artisan | 1.10 |
| Medieval | Engineer | 1.65 |
| Renaissance | Master Eng. | 2.50 |
| Colonial | Mechanic | 3.75 |
| Industrial | Machinist | 5.60 |
| Victorian | Tinker | 8.00 |
| Electric | Electrical Engineer | 8.96 |
| Atomic | Nuclear Engineer | 10.04 |
| Modern | Systems Engineer | 11.24 |
| Information | Software Engineer | 12.59 |
| Digital | AI Engineer | 14.10 |
| Cyberpunk | Cyber Engineer | 15.79 |
| Fusion | Plasma Engineer | 17.69 |
| Space | Space Engineer | 19.81 |
| Interstellar | Warp Engineer | 22.19 |
| Galactic | Galactic Engineer | 24.85 |
| Quantum | Quantum Engineer | 27.83 |

### metallurgy — base 2.0/tick, starts Iron Age

Iron: Smelter → Classical: Ironworker → Medieval: Medieval Smith → Renaissance: Renaissance Metallurgist → Colonial: Foundry Worker → Industrial: Factory Worker → Victorian: Steam Smelter → Electric: Electric Smelter → Atomic: Atomic Metallurgist → Modern: Modern Metallurgist → Information: Digital Foundry Worker → Digital: Digital Smelter → Cyberpunk: Cyber Forge Worker → Fusion: Plasma Metallurgist → Space: Stellar Foundry Worker → Interstellar: Stellar Smelter → Galactic: Galactic Metallurgist → **Quantum: Quantum Smelter**

### energy — base 8.0/tick, starts Victorian Age

Victorian: Stoker → Electric: Power Worker → Atomic: Reactor Technician → Modern: Power Engineer → Information: Grid Operator → Digital: Digital Power Manager → Cyberpunk: Cyber Energy Worker → Fusion: Fusion Technician → Space: Solar Engineer → Interstellar: Dark Energy Worker → Galactic: Antimatter Specialist → **Quantum: Zero-Point Engineer**

### hacker — base 16.0/tick, starts Information Age

Information: Script Kiddie → Digital: Coder → Cyberpunk: Black Hat → Fusion: AI Hacker → Space: Orbital Hacker → Interstellar: Interstellar Netrunner → Galactic: Galactic Hacker → **Quantum: Quantum Hacker**

### astronaut — base 32.0/tick, starts Space Age

Space: Cadet → Interstellar: Interstellar Pilot → Galactic: Galactic Explorer → **Quantum: Quantum Astronaut**

---

## Morale

Morale is a civilization-wide percentage that acts as a multiplier on **all** worker-driven building output every tick.

**Output formula with morale:**

```
output = base_rate × building_count × (0.20 + 0.80 × assigned / total_capacity) × morale_multiplier
```

### The scale

- A new civilization **starts at 50%** (neutral)
- Floor: **10%** — production never sinks below the low-morale ramp
- Cap: **100% + 5% per Wonder built** — with 0 wonders the ceiling is 100%

### The three bands

Morale is a two-way dial centred on 50%. The middle band is dead; only the extremes matter:

| Band | Morale | Effect |
| ------ | -------- | -------- |
| Low | below 25% | Production **penalty**, ramping smoothly down to **×0.50** (half output) at the 10% floor |
| Neutral | 25%–75% | **No effect** — production runs at the normal rate |
| High | above 75% | Production **bonus**, ramping smoothly up to **+20%** as morale approaches the cap |

Morale **drifts gently back toward 50% every tick**. The high bonus must therefore be earned and sustained — it bleeds away if you stop maintaining it — while a low-morale penalty is self-healing once you remove the cause.

### What raises morale

- **Morale-restoring buildings** — era-appropriate worship buildings (the shrine/temple line) and culture/entertainment buildings lift morale each tick simply by existing; no workers needed
- **Good events**
- **Advancing to a new age**

### What lowers morale

- **Food starvation**
- **Military workers exceeding 30% of population** — the further over, the faster the drain
- **More than 50% of workers idle**
- **Bad events and catastrophes** — enduring a catastrophe costs morale

### Where it shows

Morale appears as a **colored bar** in the Workers panel (`workers`) and the villager sidebar — green when boosting, neutral in the middle band, red when penalising. It also appears in each save's detail in the Load Game browser.

### Tips

- Keep food income positive
- Don't over-militarise — keep military workers under 30% of pop
- Don't leave half your population idle
- Build worship and culture buildings to push morale into the **+20% bonus** zone — this is the key lever
- Build wonders to raise the cap

See [Morale](morale.md) for the full morale page.

---

## Food Drain

Every worker costs food per tick. The amount is determined by the **food domain's class** for the current age — it is the same for all workers regardless of their building assignment:

```
total_food_drain = food_domain_food_cost × total_worker_count
```

Food drain scales with age. In primitive age, workers cost **0.06 food/tick** each — very manageable. By modern age a worker costs **~0.24 food/tick** (down significantly from older builds), keeping late-game populations affordable as long as food production scales too.

**Food drain table (all workers, by age tier):**

| Age | Food/tick per worker | 10 workers | 50 workers |
|-----|---------------------|-----------|-----------|
| Primitive | 0.060 | 0.60 | 3.0 |
| Stone | 0.067 | 0.67 | 3.35 |
| Bronze | 0.075 | 0.75 | 3.75 |
| Iron | 0.084 | 0.84 | 4.2 |
| Classical | 0.095 | 0.95 | 4.75 |
| Medieval | 0.106 | 1.06 | 5.3 |
| Victorian | 0.167 | 1.67 | 8.35 |
| Modern | 0.235 | 2.35 | 11.75 |

**What happens at food = 0:** starvation begins. Workers start dying — 1 killed every 5 ticks — until food net income returns to positive. A warning is logged on the first deficit tick; deaths are shown in red. When food recovers, a recovery message is shown and deaths stop.

**Practical advice:**

- Always ensure food income > food drain before recruiting more workers
- The `status` command and `workers` overlay both show current food drain per tick and net food/tick
- `unassign` returns workers to the idle pool — they still drain food while idle. Use `dismiss` to permanently remove workers from the population pool and immediately reduce total drain
- `dismiss <building_key> [count|all]` cuts population directly; unlike `unassign`, dismissed workers are gone entirely

---

## Worker Building Reference

Key buildings that accept workers, grouped by domain:

| Building | Key | Domain | Worker Capacity | Available From |
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

Worker capacity is **per building instance**. If you have built 3 libraries, total capacity is 3 × 4 = 12 slots. Use `assign library all` to fill all available slots.

For a full per-lineage building list see [Buildings](buildings.md).

---

## Strategy

### Early game (Primitive / Stone Age)

Recruit 5–10 workers immediately and assign them to `gathering_camp`. Each camp holds 3 workers; build 2–3 camps before recruiting past 9. Foragers at 0.06 food/tick each, and camps produce 0.50 food/tick base — a fully staffed camp more than covers its own drain.

```
build gathering_camp
recruit 3
assign gathering_camp 3
build gathering_camp
recruit 3
assign gathering_camp 3
```

Once food is stable, assign some workers to `story_circle` (knowledge) to begin unlocking research.

### Mid game (Bronze / Iron Age)

- Assign workers to lumber and masonry buildings as soon as they are built — stone and wood extraction gate most building costs
- Unlock the military domain (Iron Age) by building `war_camp`, then staff it with military workers — it produces the `soldiers` resource you spend on expeditions
- Knowledge workers in libraries become critical — `scholars_haven` milestone requires 50 knowledge workers assigned to libraries

### Late game

- Reassign food workers to more productive domains as food buildings scale up in efficiency and storage
- Hackers (Information Age, base 16.0 food/tick) and astronauts (Space Age, base 32.0 food/tick) require massive food income — scale food first
- Use `unassign <building> all` + `assign <new_building> all` to quickly pivot your workforce after an age advance

### Milestone: scholars_haven

Requires **50 knowledge workers** assigned to knowledge-domain buildings and **3 Libraries** built. Track progress with `milestones` or `ms`.

```
assign library 50
```

### Specialise vs. spread

Focusing one domain can unlock milestone chains faster. Spreading across domains provides resilience against resource shortages but delays milestones. In the early game, food + knowledge is sufficient; add military when expeditions unlock and trade when gold becomes the bottleneck.

---

## Autocomplete Reference

The command input supports TAB autocomplete for all worker commands:

| Typed | TAB Suggestions |
|-------|----------------|
| `assign ` | all unlocked building keys with WorkerCapacity > 0 (alphabetical) |
| `assign gathering_camp ` | `all` |
| `assign lib` | `library`, `library_of_congress`, … (filtered by prefix) |
| `unassign ` | only building keys that currently have workers assigned |
| `unassign barracks ` | `all` |
| `dismiss ` | only building keys that currently have workers assigned |
| `dismiss barracks ` | `all` |
| `recruit ` | `max` |

Autocomplete only shows buildings you have **unlocked** for assign, and buildings with **active assignments** for unassign — it won't suggest buildings you haven't reached yet or that have no workers to remove.

---

## What Does Not Exist

These command forms are **not valid** and will return an error:

```
recruit food            # INVALID — no domain arg
recruit military 5      # INVALID — no domain arg
assign food gathering_camp  # INVALID — domain is auto-resolved
unassign all food       # INVALID — must specify a building key
```

---

## See Also

- [Workers](villagers.md) — introductory guide to the worker system
- [Buildings](buildings.md) — building lineages and per-building capacity values
- [Resources](resources.md) — which resources each lineage produces at which age
- [Epochs](epochs.md) — how epoch transitions affect building output resources
- [Milestones](milestones.md) — milestone chains that reward domain specialisation
