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

**Viewing workers:** type `status` (or `s`) or press `e` for the Economy tab. The population bar shows `total / max (idle: N, food drain: X.X/tick)`.

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
- Cutting food drain when food stocks are low (unassign expensive-domain buildings first)
- Freeing workers to send on expeditions (expedition losses come from the pool)

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

Each domain has one class per age it spans. Class name is cosmetic but also determines the food cost lookup for workers assigned to that domain's buildings. Food costs scale geometrically: `FoodCost = baseFoodCost × 1.5^tier`.

### food — base 0.06/tick, starts Primitive Age

| Age | Class Name | Food/tick |
|-----|-----------|-----------|
| Primitive | Forager | 0.060 |
| Stone | Farmhand | 0.090 |
| Bronze | Cultivator | 0.135 |
| Iron | Laborer | 0.203 |
| Classical | Peasant | 0.304 |
| Medieval | Serf | 0.456 |
| Renaissance | Plowman | 0.684 |
| Colonial | Colonial Farmer | 1.026 |
| Industrial | Factory Hand | 1.539 |
| Victorian | Agricultural Worker | 2.309 |
| Electric | Electric Farmer | 3.463 |
| Atomic | Atomic Agronomist | 5.194 |
| Modern | Modern Farmer | 7.791 |
| Information | Digital Cultivator | 11.687 |
| Digital | AI Agronomist | 17.530 |
| Cyberpunk | Aug Harvester | 26.295 |
| Fusion | Bio-Farmer | 39.443 |
| Space | Zero-G Farmer | 59.165 |
| Interstellar | Stellar Cultivator | 88.747 |
| Galactic | Galactic Farmer | 133.120 |
| Quantum | Quantum Harvester | 199.680 |

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
| Medieval | Acolyte | 2.0 |
| Renaissance | Monk | 3.0 |
| Colonial | Missionary | 4.5 |
| Industrial | Revivalist | 6.75 |
| Victorian | Parish Priest | ~10.1 |
| Electric | Evangelical | ~15.2 |
| Atomic | Atomic Priest | ~22.8 |
| Modern | Modern Shepherd | ~34.2 |
| Information | Digital Devotee | ~51.3 |
| Digital | Virtual Cleric | ~76.9 |
| Cyberpunk | Cyber Cleric | ~115.4 |
| Fusion | Plasma Prophet | ~173.1 |
| Space | Star Preacher | ~259.7 |
| Interstellar | Interstellar Mystic | ~389.5 |
| Galactic | Galactic High Priest | ~584.3 |
| Quantum | Quantum Sage | ~876.4 |

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
| Victorian | Tinker | 8.0 |
| Electric | Electrical Engineer | 12.0 |
| Atomic | Nuclear Engineer | 18.0 |
| Modern | Systems Engineer | ~27.0 |
| Information | Software Engineer | ~40.5 |
| Digital | AI Engineer | ~60.8 |
| Cyberpunk | Cyber Engineer | ~91.1 |
| Fusion | Plasma Engineer | ~136.7 |
| Space | Space Engineer | ~205.1 |
| Interstellar | Warp Engineer | ~307.6 |
| Galactic | Galactic Engineer | ~461.4 |
| Quantum | Quantum Engineer | ~692.1 |

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

Morale is a hidden engine multiplier (0.10–cap) that scales all worker-driven building output every tick.

**Output formula with morale:**

```
output = base_rate × building_count × (0.20 + 0.80 × assigned / total_capacity) × morale
```

### Default and floor

- Default morale: **1.0** (100% — no penalty or bonus)
- Floor: **0.10** (10% — maximum output reduction)
- Base cap: **1.0** — extended by wonders

### Raising the cap

Each wonder you build adds **+0.05** to the morale cap. With 4 wonders built, the cap is 1.20, allowing morale to reach 120% output.

### What raises morale

| Driver | Change per tick |
|--------|----------------|
| Food rate > 0 (surplus) | +0.002 |
| Age advancement | +0.08 (one-time) |
| Good event triggered | +0.04 (one-time) |
| Passive recovery (when below cap, food ≥ 0) | +0.001 |

### What lowers morale

| Driver | Change per tick / event |
|--------|------------------------|
| Food = 0 and food rate < 0 (starvation) | −0.005/tick |
| Military workers > 30% of total pop | −0.003 × (excess ratio × 10) per tick |
| Idle workers > 50% of total pop | −0.002/tick |
| Bad event triggered | −0.04 (one-time) |
| Endure a catastrophe | −0.10 (one-time) |

### Military ratio mechanic

If military-assigned workers exceed 30% of your total population, morale drains proportionally to how far over the threshold you are. At 40% military ratio (10 points over): −0.003 × 1.0 = −0.003/tick. At 50% (20 points over): −0.006/tick. Keep military assignments below 30% of pop to avoid this drain.

### Reset values

| Event | Morale set to |
|-------|--------------|
| New game / wipe | 1.0 |
| Prestige | 0.70 |
| Succumb (catastrophe) | 0.50 |

### Morale warning

When morale drops below **40%**, a log warning fires: *"⚠ Morale critical: N% — worker output severely reduced"*. The warning resets once morale recovers above 40%.

### Tips

- Keep food income positive — it is the most reliable morale driver
- Build wonders to raise the cap and allow morale to exceed 1.0
- Avoid over-militarising early; keep military assignments under 30% of pop
- After a Succumb or Prestige reset, morale starts below 1.0 — rebuild food production quickly to recover it
- Idle workers are a double penalty: zero extra production **and** a morale drain if they exceed 50% of pop

---

## Food Drain

Every worker costs food per tick. The amount is determined by the **food domain's class** for the current age — it is the same for all workers regardless of their building assignment:

```
total_food_drain = food_domain_food_cost × total_worker_count
```

Food drain scales with age. In primitive age, workers cost **0.06 food/tick** each — very manageable. By modern age the same worker costs **~7.8 food/tick**, so population size must be matched by food production capacity.

**Food drain table (all workers, by age tier):**

| Age | Food/tick per worker | 10 workers | 50 workers |
|-----|---------------------|-----------|-----------|
| Primitive | 0.060 | 0.6 | 3.0 |
| Stone | 0.090 | 0.9 | 4.5 |
| Bronze | 0.135 | 1.35 | 6.75 |
| Iron | 0.203 | 2.03 | 10.1 |
| Classical | 0.304 | 3.04 | 15.2 |
| Medieval | 0.456 | 4.56 | 22.8 |
| Victorian | 2.309 | 23.1 | 115.4 |
| Modern | 7.791 | 77.9 | 389.5 |

**What happens at food = 0:** food goes negative. Workers do not die automatically, but negative food income is a drain on all other food-producing systems. Unassign workers from buildings or build more food production to recover.

**Practical advice:**

- Always ensure food income > food drain before recruiting more workers
- The `status` command shows current food drain per tick
- Unassigning workers frees them but they still drain food while idle — the drain reduction only comes from reducing total population (no mechanic to "dismiss" workers permanently short of expedition losses or epidemic events)

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
- Unlock the military domain (Iron Age) by building `war_camp`; assign soldiers for expedition access
- Knowledge workers in libraries become critical — `scholars_haven` milestone requires 20 knowledge workers assigned to libraries

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
