# Workers & Domains: Reference

Workers are your civilization's labor force. Every building that produces resources has worker slots; staffing it raises output from a 20% floor up to 100%. Workers come from a **single generic pool**, and any worker can go to any building. The domain comes from the building, not the worker.

For a gentler introduction see [Workers](villagers.md).

---

## Overview

- One pool of workers, all identical until assigned to a building
- Workers arrive and go to work on their own, by your **worker shares** (every domain on auto by default, following your buildings' worker slots); see [Worker Shares](#worker-shares)
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
- Workers arrive on their own and go to work by your [worker shares](#worker-shares). You can also recruit and assign by hand.
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

## Worker Shares

You don't have to recruit or assign workers yourself. The game recruits into empty worker slots while housing and food allow (**auto-recruit**, on by default), and puts idle workers to work by your **worker shares**: how your workforce splits across the 12 domains. You can steer the split, or take over by hand. A worker you place by hand stays where you put it.

```
workers share
workers share <domain> [percent|auto]
workers share auto
workers auto-recruit [on|off]
```

| Command | Effect |
|---------|--------|
| `workers share` | List the shares: one line per domain that has worker buildings or a share set, and whether auto-recruit is on |
| `workers share knowledge` | Show one domain's line |
| `workers share knowledge 40` | Set knowledge to 40% of your workforce |
| `workers share lumber 12.5%` | A share is 0 to 100; decimals and a `%` sign are allowed |
| `workers share military 0` | Keep military empty |
| `workers share knowledge auto` | Put knowledge back on auto |
| `workers share auto` | Put every domain back on auto |
| `workers auto-recruit` | Show whether auto-recruit is on |
| `workers auto-recruit off` | Stop recruiting automatically; `on` starts it again (`workers autorecruit` works too) |

Each line of `workers share` reads like `Knowledge: 40% (set), 8 workers in 10 slots` or `Food: 34% (auto), 12 workers in 15 slots`. The Workers panel (`workers`) shows the same in its **Shares** section, under a line that says what auto-recruit is doing: on, off, waiting after your last worker command, no housing left, every worker slot is filled, waiting for food, or paused because food ran out.

### How the split works

- A **share** is a percent of your whole workforce, idle workers included, for one domain.
- Domains without a share are **on auto**: they split whatever the set shares leave, in proportion to their worker slots (the slots in that domain's built buildings). The default is every domain on auto, so your workers follow your buildings' slots.
- Set shares that add up to more than 100% are each scaled down to fit: knowledge 80% and military 40% work as about 67% and 33%. The auto domains then get nothing, except overflow: when every domain with a share is full, extra workers go to auto domains with free slots.
- A share of **0** keeps a domain empty. Its workers move to free slots elsewhere as room opens, and no new worker goes there. Setting food to 0 replies with a warning: "Food buildings get no workers now: watch your food."
- A domain never holds more workers than its slots. Workers its share can't place go to other domains with free slots. If you set a share for a domain with no buildings yet, the reply says the share applies once you build one.

**Example.** You have 30 worker slots in each of three domains and 40 workers:

| Domain | Slots | All on auto | With `workers share knowledge 50` | Workers then |
|--------|-------|-------------|-----------------------------------|--------------|
| food | 30 | 33% (auto) | 25% (auto) | 10 |
| lumber | 30 | 33% (auto) | 25% (auto) | 10 |
| knowledge | 30 | 33% (auto) | 50% (set) | 20 |

Knowledge takes its 50% first. Food and lumber split the other 50% by their slots, which are equal, so they get 25% each.

### The shares routine

Every 5 ticks (10 seconds at 1x), and once per step of the offline catch-up (one-minute steps), the game:

1. moves workers out of domains set to 0 into free slots elsewhere,
2. puts idle workers to work, each in the domain furthest below its share,
3. with auto-recruit on, recruits into the empty worker slots as far as housing allows, while the net food rate stays at or above a margin: one worker's food, or 5% of the food production if that is more. It never recruits more workers than there are empty slots, so it never adds idle workers.

Within a domain, buildings fill in this order: buildings that are not superseded before those that are (a building is superseded once a higher tier of its line is open), the newest age first, then the higher tier of its line.

**Food safety.** A recruit your food income can't feed can still come as a food worker, if a food building has a free slot and the worker grows more food than they eat. A staffed food slot adds 80% of the building's food divided by its slots: a worker in a Primitive gathering camp grows about 0.27 food a tick and eats 0.06. So a small food share never stops growth; the extra recruits go to food buildings. While food is short, idle workers go to food buildings first. Nothing is recruited while food is at 0 and workers are starving. The manual `recruit max` still does not check food.

**Your workers stay put.** The routine never takes a worker out of a building, except out of a domain set to 0, so your own `assign` and `unassign` stick. Setting or clearing a share (`workers share ...`) moves workers once to match the new shares: one at a time, from the domain furthest over its share to the one furthest under it that has a free slot. Food workers stay where moving them would make the food rate fall below zero.

**The minute's wait.** After a worker command (`recruit`, `assign`, `unassign`, `dismiss`, `sell`, `upgrade`) the routine waits a minute (30 ticks), so it never grabs workers you are moving by hand. After that, workers still idle go to work by your shares. So `unassign` alone may keep a building empty for only a minute; to keep a domain empty, set its share to 0.

**Auto-recruit off.** With `workers auto-recruit off`, nothing is recruited automatically, but idle workers still go to work by your shares after that minute. `workers auto-recruit on` starts recruiting again at once.

**Logs.** What the routine does is a routine line in the log (plain text color, and marked `·` in the **Logs** panel): `Shares: recruited 3 workers (population 12/20), put 2 idle workers to work.` The replies to setting a share or switching auto-recruit are routine lines too, except the warning for food at 0. The first time workers arrive in a run (when your population was 0), a note says: "Workers arrive on their own: the game recruits into empty worker slots while housing and food allow, and puts them to work by your worker shares. Type workers to see them; workers auto-recruit off to recruit by hand." After time away, the welcome back adds what the routine did, for example "While you were away, your worker shares recruited 12 workers (population 40/50), put 3 idle workers to work."

**Build plan.** A copy the [build plan](plan.md#how-it-runs) finishes is staffed from idle workers first: into the copy itself, or, with shares set, wherever the shares say. If idle workers run out, the plan moves workers out of superseded buildings, the same line's first; with shares set, only from the copy's own domain, so the split holds. It never moves food workers or workers in this age's buildings. The routine then recruits for what is still empty.

**Saving.** Shares, the auto-recruit switch and the minute's wait are saved with your game. Prestige and Succumb put every share back on auto. Auto-recruit is a preference and stays as you set it across prestige and Succumb, like wonder overflow. A new game resets both: every domain on auto, auto-recruit on. Older saves load with every domain on auto and auto-recruit on.

---

## Recruiting Workers

```
recruit [count|max]
```

Workers arrive on their own: with auto-recruit on (the default), the game recruits into empty worker slots while housing and food allow (see [Worker Shares](#worker-shares)). `recruit` still works when you want workers now, or when you turn auto-recruit off. Workers are recruited from free housing. You don't name a domain; workers are generic until assigned.

| Command | Effect |
|---------|--------|
| `recruit` | Recruit 1 worker |
| `recruit 5` | Recruit 5 workers |
| `recruit max` | Recruit as many as your housing allows |

New workers start idle. Put them to work with `assign`, or leave them: a minute after your last worker command, the shares routine puts any still idle to work by your shares.

**Food cost:** recruiting is free, but every worker eats food each tick from the moment it joins. The amount is the current age's rate for all workers, whatever building they staff: **0.06 food/tick** per worker in the Primitive Age. See [Food Drain](#food-drain).

**`recruit max`** fills all your free housing. It does not check your food income, so it can recruit more workers than your food supports and put you into a deficit. Auto-recruit does check it: it keeps a food margin and fills only empty worker slots.

**Viewing workers:** type `workers` to open the Workers panel (morale, summary, worker shares, building slots, workers by domain), or `status` for a text summary. The Workers panel shows net food/tick in green or red (blue or orange on the accessible [themes](commands.md#themes)), how many workers your food can sustain or a deficit warning, and per-building fill bars. The Workers box in the sidebar is always visible and shows population, idle, housing left, drain and net food at a glance.

---

## Assigning Workers to Buildings

```
assign <building> [count|all]
```

Assigns workers from the idle pool to the building. The domain comes from the building; you never specify one.

You don't have to assign: the shares routine puts idle workers to work every 5 ticks (10 seconds at 1x; see [Worker Shares](#worker-shares)). Assign by hand when you want workers in a particular building. Your assignments stick: the routine never takes a worker out of a building, except out of a domain whose share is 0. After `assign` it waits a minute before it places anyone, so it never grabs workers you are moving by hand.

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

Unassigned workers go back to the idle pool at once and can be reassigned elsewhere. They keep eating while idle. If you leave them idle, the shares routine puts them back to work by your shares a minute after your last worker command, possibly in the same domain.

When to unassign:

- To move workers from low-priority buildings to new, higher-tier ones after an age advance. Within a domain the routine fills the newer buildings first, so workers it places back in that domain go to the new ones.
- To pull workers out of a domain for good, set its share to 0 instead (`workers share military 0`): `unassign` alone lasts only until the routine places them again. Campaigns spend the `soldiers` resource, not workers, so once you've banked enough soldiers, military workers only cost food and morale.

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

With auto-recruit on, the game recruits into empty worker slots again once housing and food allow, the slots you just emptied included. To keep a building empty, sell it, set its domain's share to 0, or turn auto-recruit off (`workers auto-recruit off`).

---

## Starvation

When food runs out while your workers are still eating, they begin dying:

- A warning is logged on the first tick without food
- **1 worker dies every 5 ticks** (10 seconds at 1x) until food recovers
- Deaths are logged in red
- As soon as there is food in stock again, deaths stop and a recovery message is logged

Overpopulating without the food to back it up costs you workers. Auto-recruit keeps a food margin and recruits nothing while workers are starving, but `recruit max`, a food share set too low, or the higher food rate after an age advance can still put you into a deficit. To recover:

1. `dismiss` workers from low-priority buildings to cut drain at once
2. Build more food buildings. While food is short, the shares routine sends idle workers to food buildings first.
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

- Make sure food income exceeds food drain before recruiting more workers. Auto-recruit does this for you (it keeps a margin of one worker's food, or 5% of food production if that is more); `recruit` by hand does not.
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

Worker slots are **per building**. If you have built 3 libraries, you have 3 × 4 = 12 slots. The shares routine fills them as workers come, or use `assign library all` to fill them from your idle workers at once.

For a full per-lineage building list see [Buildings](buildings.md).

---

## Strategy

### Early game (Primitive / Stone Age)

Build 2-3 gathering camps and a few huts. Workers arrive on their own and staff the camps. Each camp holds 3 workers and auto-recruit fills only empty slots, so build camps as you build housing. Foragers eat 0.06 food/tick each, and a fully staffed camp produces 1.0 food/tick, so it covers its own three workers' drain (0.18) several times over.

```
build gathering_camp
build hut
build gathering_camp
build hut
workers
```

Once food is stable, build story circles (knowledge) to start on research; workers come to staff them too. For faster research, give knowledge a bigger share: `workers share knowledge 40`.

### Mid game (Bronze / Iron Age)

- Lumber and masonry buildings are staffed as they finish, as far as housing and food allow. Stone and wood go into most building costs, so keep your housing ahead of your worker slots.
- A `war_camp` produces the `soldiers` resource (from the Iron Age) that you spend on campaigns. On auto its slots draw workers like any other building's; if you don't need soldiers yet, `workers share military 0` keeps those workers for your producers.
- Knowledge workers in libraries matter more and more. The `scholars_haven` milestone needs 50 knowledge workers and 3 libraries.

### Late game

- As food buildings get more productive, you need fewer food workers. Give food a small share (not 0) to move the rest to other domains. The routine still sends extra recruits to food buildings when the food income can't feed them.
- Every worker eats the age's rate, and that rate keeps rising, so a large late-game population needs a large food income. Scale food before you scale headcount.
- After an age advance, the new buildings fill from idle workers and new recruits. When housing is full, `unassign <old building> all` frees workers, and a minute later the routine puts them to work by your shares, newer buildings first. Or `assign <new building> all` places them yourself.

### Milestone: scholars_haven

Needs **50 knowledge workers** assigned to knowledge buildings and **3 Libraries** built. Track progress with `milestones` or `ms`.

```
assign library 50
```

Or give knowledge a large share (`workers share knowledge 60`) and let the routine fill your knowledge buildings as workers come.

### Specialize vs. spread

Focusing on one domain can finish milestone chains faster. Spreading across domains protects you from resource shortages but delays milestones. Worker shares are the lever: auto spreads your workers by worker slots, and a share focuses them (`workers share knowledge 50`). Early on, food and knowledge are enough; add military when campaigns open up and trade when gold becomes the bottleneck.

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
| `workers ` | `share`, `auto-recruit` |
| `workers share ` | the worker domains, the ones you have worker buildings in (or a share set for) first, then the rest, and `auto` |
| `workers share kn` | `knowledge` |
| `workers share knowledge ` | `auto` |
| `workers auto-recruit ` | `off`, `on` |

For `assign` the prompt only suggests buildings you have **built**, and for `unassign` and `dismiss` only buildings that have workers to remove.

---

## What Does Not Exist

These command forms are **not valid** and return an error:

```
recruit food            # INVALID: no domain argument
recruit military 5      # INVALID: no domain argument
assign food gathering_camp  # INVALID: the domain comes from the building
unassign all food       # INVALID: name a building
workers share gathering_camp 40  # INVALID: a share is for a domain, not a building
workers share food 150  # INVALID: a share is a percent from 0 to 100
```

---

## See Also

- [Workers](villagers.md): introductory guide to the worker system
- [Buildings](buildings.md): building lineages and worker slots per building
- [Resources](resources.md): which resources each lineage produces in which age
- [Epochs](epochs.md): how epoch transitions affect building output resources
- [Milestones](milestones.md): milestone chains that reward domain specialization
