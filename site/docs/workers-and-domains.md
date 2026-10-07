# Workers

Workers are your civilization's labor force. Every building that produces resources has worker slots, and staffing it raises its output from a 20% floor up to 100%. Workers come from a **single generic pool**: any worker can go to any building. There are 12 worker domains, and a worker's domain comes from the building it staffs, not from the worker.

---

## Quick Start

- **Workers come on their own.** Build housing and buildings with worker slots. With auto-recruit on (the default), the game recruits into empty slots while housing and food allow, and puts idle workers to work by your [worker shares](#worker-shares).
- **Steer, or take over.** `workers share knowledge 40` sends more workers to knowledge; `workers share military 0` keeps military empty. `recruit` and `assign` still work by hand, and a worker you place stays where you put it.
- **Every worker eats.** Each worker eats the current age's food rate every tick, whatever it does. Keep the food rate positive (see [Food Drain](#food-drain)).
- **Look at the panel.** `workers` opens the Workers panel: morale, food, shares, open slots and who works where.

The first age, step by step, is in [Your First Ten Minutes](first-ten-minutes.md).

---

## The Worker Pool

There is a single pool of workers. Workers have no domain until they are assigned to a building. Your population is:

```
total_workers = idle_workers + sum of all building assignments
```

An assigned worker takes on the **class name** for that building's domain at the current age. A worker on a `gathering_camp` in the Primitive Age is a **Forager**; the same worker on a `story_circle` is a **Shaman**. The class name is only a label: it changes nothing about what the worker produces or eats. There are no separate domain pools.

- You don't choose a domain when recruiting.
- Workers can move freely between any buildings that have worker slots.
- Your population is limited by housing (huts, longhouses and so on).
- Idle workers eat and produce nothing.

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

Every 5 ticks (10 seconds), and once per step of the offline catch-up (one-minute steps), the game:

1. moves workers out of domains set to 0 into free slots elsewhere,
2. puts idle workers to work, each in the domain furthest below its share,
3. with auto-recruit on, recruits into the empty worker slots as far as housing allows, while the net food rate stays at or above a margin: one worker's food, or 5% of the food production if that is more. It never recruits more workers than there are empty slots, so it never adds idle workers. On known ground ([Era Mastery](prestige.md#era-mastery)) the margin is measured before the age's speed-up, so the same buildings feed the same workforce at any speed.

Within a domain, buildings fill in this order: buildings that are not superseded before those that are (a building is superseded once a higher tier of its line is open), the newest age first, then the higher tier of its line.

**Food safety.** A recruit your food income can't feed can still come as a food worker, if a food building has a free slot and the worker grows more food than they eat. A staffed food slot adds 80% of the building's food divided by its slots: a worker in a Primitive gathering camp grows about 0.27 food a tick and eats 0.06. So a small food share never stops growth; the extra recruits go to food buildings. While food is short, idle workers go to food buildings first. Nothing is recruited while food is at 0 and workers are starving. The manual `recruit max` still does not check food.

**Your workers stay put.** The routine never takes a worker out of a building, except out of a domain set to 0, so your own `assign` and `unassign` stick. Setting or clearing a share (`workers share ...`) moves workers once to match the new shares: one at a time, from the domain furthest over its share to the one furthest under it that has a free slot. Food workers stay where moving them would make the food rate fall below zero.

**The minute's wait.** After a worker command (`recruit`, `assign`, `unassign`, `dismiss`, `sell`, `upgrade`) the routine waits a minute (30 ticks), so it never grabs workers you are moving by hand. After that, workers still idle go to work by your shares. So `unassign` alone may keep a building empty for only a minute; to keep a domain empty, set its share to 0.

**Auto-recruit off.** With `workers auto-recruit off`, nothing is recruited automatically, but idle workers still go to work by your shares after that minute. `workers auto-recruit on` starts recruiting again at once.

**Logs.** What the routine does is a routine line in the log (plain text color, and marked `·` in the **Logs** panel): `Shares: recruited 3 workers (population 12/20), put 2 idle workers to work.` The replies to setting a share or switching auto-recruit are routine lines too, except the warning for food at 0. The first time workers arrive in a run (when your population was 0), a note says: "Workers arrive on their own: the game recruits into empty worker slots while housing and food allow, and puts them to work by your worker shares. Type workers to see them; workers auto-recruit off to recruit by hand." After time away, the welcome back adds what the routine did, for example "While you were away, your worker shares recruited 12 workers (population 40/50), put 3 idle workers to work."

**Build plan.** A copy the [build plan](plan.md#how-it-runs) finishes is staffed from idle workers first: into the copy itself, or, with shares set, wherever the shares say. If idle workers run out, the plan moves workers out of superseded buildings, the same line's first; with shares set, only from the copy's own domain, so the split holds. It never moves food workers or workers in this age's buildings. The routine then recruits for what is still empty.

**Saving.** Shares, the auto-recruit switch and the minute's wait are saved with your game. Prestige and Succumb put every share back on auto, unless you own **Worker Shares** from the prestige [legacy kit](prestige.md#worker-shares) (`prestige buy legacy_workers`, 36 points): then the shares you had when the run ended carry over into the new run, and its log says so ("Worker Shares: your shares carry over.", with the split). The game remembers your shares at every prestige and Succumb even before you buy it, so bought later, it sets them at once (unless you have set some since). Auto-recruit is a preference and stays as you set it across prestige and Succumb, like wonder overflow. A new game resets both: every domain on auto, auto-recruit on. Older saves load with every domain on auto and auto-recruit on.

---

## Commands

| Command | Shortcut | Effect |
|---------|----------|--------|
| `workers` | | Open the Workers panel |
| `workers share ...` | | Set or show worker shares (see [Worker Shares](#worker-shares)) |
| `workers auto-recruit on`, `off` | | Turn auto-recruit on or off; bare `workers auto-recruit` shows it |
| `recruit [count]` | `r` | Recruit workers into free housing (default 1); `recruit max` fills all of it |
| `assign <building> [count]` | `a` | Put idle workers to work in a building (default 1, or `all`) |
| `unassign <building> [count]` | `u` | Take workers out of a building; they go idle (default 1, or `all`) |
| `dismiss <building> [count]` | | Remove workers from a building and from your population (default 1, or `all`) |
| `status` | `s` | Text summary, including population, idle workers and food drain |

### Recruiting

Workers arrive on their own: with auto-recruit on, the game recruits into empty worker slots while housing and food allow. `recruit` still works when you want workers now, or when you turn auto-recruit off. Workers are recruited from free housing, and you don't name a domain.

| Command | Effect |
|---------|--------|
| `recruit` | Recruit 1 worker |
| `recruit 5` | Recruit 5 workers |
| `recruit max` | Recruit as many as your housing allows |

New workers start idle. Put them to work with `assign`, or leave them: a minute after your last worker command, the shares routine puts any still idle to work by your shares.

Recruiting is free, but every worker eats from the moment it joins (see [Food Drain](#food-drain)). **`recruit max`** fills all your free housing and does not check your food income, so it can put you into a deficit. Auto-recruit does check it: it keeps a food margin and fills only empty worker slots.

### Assigning

Assigns workers from the idle pool to a building. The domain comes from the building; you never name one.

| Command | Effect |
|---------|--------|
| `assign gathering_camp 3` | Assign 3 workers to gathering_camp |
| `assign library all` | Assign all idle workers to library, up to its free slots |
| `assign barracks 5` | Assign 5 workers to barracks |
| `assign` (no arguments) | Shows the usage line |

- You must have built at least one of the building.
- Assignment is capped at the building count times its worker slots.
- You can't assign more workers than you have idle.

You don't have to assign: the shares routine puts idle workers to work every 5 ticks (10 seconds). Assign by hand when you want workers in a particular building. Your assignments stick: the routine never takes a worker out of a building, except out of a domain whose share is 0.

### Unassigning

Removes workers from a building and returns them to the idle pool, where they keep eating.

| Command | Effect |
|---------|--------|
| `unassign gathering_camp 2` | Remove 2 workers from gathering_camp |
| `unassign barracks all` | Remove all workers from barracks |

If you leave them idle, the shares routine puts them back to work a minute after your last worker command, possibly in the same domain. Within a domain it fills the newest buildings first, so `unassign <old building> all` after an advance moves workers to your new buildings. To pull workers out of a domain for good, set its share to 0 instead (`workers share military 0`).

### Dismissing

`dismiss` removes workers from a building **and** from your population. Dismissed workers are gone.

| Command | Effect |
|---------|--------|
| `dismiss gathering_camp 2` | Remove 2 workers from gathering_camp and reduce population by 2 |
| `dismiss barracks all` | Dismiss all workers assigned to barracks |

Use it in a food deficit (idle workers still eat, so `unassign` doesn't help) or to free housing. With auto-recruit on, the game recruits into the emptied slots again once housing and food allow. To keep a building empty, sell it, set its domain's share to 0, or turn auto-recruit off.

### The Workers panel

`workers` opens the Workers panel. Before your first worker it says how workers arrive and shows your shares. After that it has five sections:

- **Morale**: your morale as a colored bar (green for a bonus, red for a penalty), its cap once wonders have raised it above 100%, and what it does to production.
- **Summary**: population over housing, idle workers and housing left; food use per tick (and each worker's share of it) and net food per tick, green when rising and red when falling; then how many workers your food production can feed, or a warning that food is falling.
- **Shares**: what auto-recruit is doing (recruiting as slots open, off, waiting after your last worker command, no housing left, every worker slot filled, waiting for food, or paused because food ran out), then each domain's share (set, or auto) with a bar of its workers over its slots.
- **Building slots**: filled slots over total slots across every worker building, with a bar and a percent, and the four buildings with the most open slots (or "All slots filled").
- **By domain**: each domain with workers, its current class name and head count, and a bar per building showing workers over slots.

<figure class="screen" data-screen="workers"><figcaption>The Workers panel in the Bronze Age: morale, food and housing, then the share of each domain with its workers over its slots.</figcaption></figure>

On the accessible [themes](themes.md) the panel's greens and reds are blue and orange. The Workers box in the sidebar is always visible: population over housing, idle workers, housing left, food use and net food.

### Completion

The prompt suggests completions for all worker commands, shown dim after the cursor; `Tab` takes one and cycles to the next (see [The prompt](commands.md#the-prompt)):

| Typed | Suggestions |
|-------|----------------|
| `assign ` | your built buildings that have worker slots, the ones with a free slot first |
| `assign gathering_camp ` | `all` |
| `assign lib` | built buildings whose key starts with `lib`, such as `library` |
| `unassign ` and `dismiss ` | only buildings that currently have workers assigned |
| `unassign barracks ` | `all` |
| `recruit ` | `max` |
| `workers ` | `share`, `auto-recruit` |
| `workers share ` | the worker domains, the ones you have worker buildings in (or a share set for) first, then the rest, and `auto` |
| `workers share knowledge ` | `auto` |
| `workers auto-recruit ` | `off`, `on` |

### What does not work

These forms return an error:

```
recruit food                     # no domain argument: workers are generic
assign food gathering_camp       # the domain comes from the building
unassign all food                # name a building
workers share gathering_camp 40  # a share is for a domain, not a building
workers share food 150           # a share is a percent from 0 to 100
```

---

## Production and Efficiency

```
output = base_rate × building_count × (0.20 + 0.80 × assigned / total_capacity)
```

`total_capacity` is the building count times the building's worker slots. `base_rate` is the fully staffed rate shown in the building's description. For construction resources it is set by the Payback Rule; food, faith, culture and soldiers keep hand-set rates (see [How Production Rates Are Set](buildings.md#how-production-rates-are-set)).

| Workers Assigned | Efficiency |
|-----------------|------------|
| 0 (none) | 20% |
| 25% of slots | 40% |
| 50% of slots | 60% |
| 75% of slots | 80% |
| 100% (full staff) | 100% |

An unstaffed building still contributes 20%. Only full staffing reaches the full rate, and workers beyond the slots add nothing. [Morale](#morale) then multiplies the result.

### Worker output bonuses

**Worker output** is the part of a building's output its workers add: the 80% that staffing brings, on top of the 20% an empty building makes. A worker output bonus raises that part:

```
worker output       = base_rate × building_count × 0.80 × assigned / total_capacity × morale_multiplier
worker output bonus = worker output × (sum of your worker output bonuses)
```

So with +15% worker output a fully staffed building makes 12% more than its listed rate (15% of the 80% its crew adds), and an empty building gains nothing. One thing gives it today: the Jacked In milestone (+15%, from the Cyberpunk Age). No tech does any more: Tool Making, Bronze Working, Road Building and Neural Interface raise the output of a resource instead (see [Technologies](technologies.md#output-of-one-resource)).

The bonus is added on top of your other production bonuses, not multiplied by them. It is outside [the all-production cap](resources.md#the-all-production-cap) and has no cap of its own, so it keeps counting when all production is capped. `rates` shows it as its own **Worker bonus** part, and the Stats panel lists the total as **Worker output** under Active Multipliers.

**Example: gathering_camp**

- Base rate 1.0 food/tick, 3 worker slots
- Built 2×: 6 slots in total, base output 1.0 × 2 = 2.0 food/tick
- With 3 workers (50% fill): `2.0 × (0.20 + 0.80 × 0.50)` = **1.2 food/tick**
- With all 6 slots filled: **2.0 food/tick**

**Example: library (Classical Age, knowledge domain)**

- Base rate 3.2 knowledge/tick, 4 worker slots
- Built 3×: 12 slots in total, base output 3.2 × 3 = 9.6 knowledge/tick
- With 8 workers (67% fill): `9.6 × (0.20 + 0.80 × 0.67)` ≈ **7.0 knowledge/tick**
- With all 12 slots filled: **9.6 knowledge/tick**

### Early worker buildings

Worker slots are **per building**: 3 libraries give 3 × 4 = 12 slots.

| Building | Key | Domain | Worker Slots | Available From |
|----------|-----|--------|----------------|----------------|
| Gathering Camp | `gathering_camp` | food | 3 | Primitive Age |
| Story Circle | `story_circle` | knowledge | 2 | Primitive Age |
| Wood Camp | `wood_camp` | lumber | 3 | Primitive Age |
| Shrine | `shrine` | faith | 2 | Primitive Age |
| Forager Post | `forager_post` | food | 4 | Stone Age |
| Elders' Hall | `elders_hall` | knowledge | 2 | Stone Age |
| Stone Camp | `stone_camp` | masonry | 3 | Stone Age |
| War Camp | `war_camp` | military | 3 | Stone Age |
| Farm | `farm` | food | 5 | Bronze Age |
| Smithy | `smithy` | engineering | 4 | Bronze Age |
| Market | `market` | trade | 3 | Bronze Age |
| Barracks | `barracks` | military | 4 | Bronze Age |
| Library | `library` | knowledge | 4 | Classical Age |

For every building see [Buildings](buildings.md).

---

## Food Drain

Every worker eats food each tick. The amount is set by the current age and is the same for every worker, whatever building it staffs, idle workers included:

```
total_food_drain = food_per_worker(current age) × population
```

A worker eats **0.06 food/tick** in the Primitive Age. The rate rises 12% with each age, to about **0.23** in the Modern Age and **0.58** in the Quantum Age (the Transcendent Age keeps the Quantum rate). When you advance, the new rate applies to every worker at once. The full per-age list is the food column under [Worker Class Names by Age](#worker-class-names-by-age).

| Age | Food/tick per worker | 10 workers | 50 workers |
|-----|---------------------|-----------|-----------|
| Primitive | 0.060 | 0.60 | 3.00 |
| Stone | 0.067 | 0.67 | 3.36 |
| Bronze | 0.075 | 0.75 | 3.76 |
| Iron | 0.084 | 0.84 | 4.21 |
| Classical | 0.094 | 0.94 | 4.72 |
| Medieval | 0.106 | 1.06 | 5.29 |
| Victorian | 0.166 | 1.66 | 8.32 |
| Modern | 0.234 | 2.34 | 11.69 |
| Quantum | 0.579 | 5.79 | 28.94 |

- Make sure food income exceeds food drain before recruiting by hand. Auto-recruit does this for you: it keeps a margin of one worker's food, or 5% of food production if that is more.
- The Workers panel shows food use and net food per tick; `status` shows the drain.
- `unassign` returns workers to the idle pool, where they still eat. `dismiss` is the way to cut drain at once.

---

## Starvation

When food runs out while your workers are still eating, they begin dying:

- A warning is logged on the first tick without food
- **1 worker dies every 5 ticks** (10 seconds) until food recovers
- Deaths are logged in red
- As soon as there is food in stock again, deaths stop and a recovery message is logged

Starvation also drags morale down. Auto-recruit keeps a food margin and recruits nothing while workers are starving, but `recruit max`, a food share set too low, or the higher food rate after an age advance can still put you into a deficit. To recover:

1. `dismiss` workers from low-priority buildings to cut drain at once
2. Build more food buildings. While food is short, the shares routine sends idle workers to food buildings first.
3. `unassign` workers from other buildings and `assign` them to food buildings

---

## Morale

Morale multiplies the output of every worker-driven building:

```
output = base_rate × building_count × (0.20 + 0.80 × assigned / total_capacity) × morale_multiplier
```

It starts at 50%, where it has no effect, and the further it moves from 50% the bigger the effect: half output at the 10% floor, up to +20% at its cap. Starvation, military workers over 30% of your population and more than half your workers idle pull it down; worship and culture buildings, your faith income and good events push it up. The Workers panel shows it as a colored bar. Everything else is on [Morale](morale.md).

---

## The 12 Worker Domains

A building's domain decides which worker class staffs it. "First Building" is the age of the earliest building that takes workers in that domain.

| Domain | Lineage | Makes | Example Buildings | First Building |
|--------|---------|-------|-------------------|----------------|
| `food` | Food | food | `gathering_camp`, `forager_post`, `farm`, `field_works` | Primitive Age |
| `knowledge` | Knowledge | knowledge | `story_circle`, `elders_hall`, `scriptorium`, `library` | Primitive Age |
| `lumber` | Organic Extraction | wood, then coal, oil, nanobots and quantum flux | `wood_camp`, `woodcutter_camp`, `lumber_mill`, `coal_mine` | Primitive Age |
| `faith` | Faith | faith | `shrine`, `standing_stones`, `altar`, `temple` | Primitive Age |
| `masonry` | Geological Extraction | stone, then [ores](resources.md#what-the-ores-are-for), uranium and antimatter | `stone_camp`, `stone_pit`, `quarry`, `uranium_mine` | Stone Age |
| `military` | Military | soldiers (from the Iron Age) | `war_camp`, `barracks`, `hunting_lodge`, `legion_fort` | Stone Age |
| `trade` | Trade, Harbor | gold | `market`, `trading_post`, `merchant_quarter`, `harbor` | Bronze Age |
| `engineering` | Engineering | iron, steel, electricity, plasma, dark matter, quantum flux | `smithy`, `ironworks`, `mill`, `power_station`, `fusion_reactor` | Bronze Age |
| `metallurgy` | Metallurgy | iron, steel, titanium, dark matter, antimatter, quantum flux | `smelter`, `forge`, `foundry`, `steel_mill`, `orbital_refinery` | Iron Age |
| `energy` | Energy | coal, oil, electricity, plasma, dark matter, quantum flux | `coal_plant`, `steam_turbine`, `nuclear_reactor`, `fusion_reactor_array` | Industrial Age |
| `hacker` | Hacker | data | `server_farm`, `data_center`, `cyber_hub` | Information Age |
| `astronaut` | none | nothing | none | No building uses it yet |

- The Nano Foundry (nanobots) takes engineering workers. The Embassy and Grand Embassy take trade workers, and the Geographic Society takes military workers.
- Metallurgy doesn't refine ore: its buildings consume nothing, like every other producer.
- **Culture** has no worker domain. Culture buildings produce on their own and take no workers.

---

## Worker Class Names by Age

Each domain has one class per age it spans. Before a domain's first class age, a worker shows the domain's first class: a `wood_camp` worker in the Primitive Age is a Gatherer, a `war_camp` worker before the Iron Age a Soldier, and a `coal_plant` worker in the Industrial Age a Stoker. The Transcendent Age keeps the Quantum Age classes. The class name is a label: what a worker eats depends only on the age (see [Food Drain](#food-drain)).

### Quick reference

| Domain | Primitive | Iron | Classical | Medieval | Victorian | Modern |
|--------|-----------|------|-----------|----------|-----------|--------|
| food | Forager | Laborer | Peasant | Serf | Agricultural Worker | Modern Farmer |
| knowledge | Shaman | Scholar | Philosopher | Friar | Victorian Scholar | Modern Researcher |
| lumber | Gatherer | Lumberjack | Sawyer | Forester | Steam Logger | Petroleum Worker |
| masonry | - | Miner | Iron Extractor | Medieval Miner | Victorian Quarryman | Modern Geologist |
| faith | Devotee | Celebrant | Initiate | Acolyte | Parish Priest | Modern Shepherd |
| military | - | Soldier | Legionary | Knight | Victorian Guard | Modern Soldier |
| trade | - | Merchant | Trader | Nobleman | Victorian Trader | Corporate Trader |
| engineering | - | Craftsman | Artisan | Engineer | Tinker | Systems Engineer |
| metallurgy | - | Furnace Hand | Ironworker | Medieval Smith | Steam Smelter | Modern Metallurgist |
| energy | - | - | - | - | Stoker | Power Engineer |

### food (starts Primitive Age)

The food column doubles as the food each worker eats per tick in that age, since that rate applies to every worker.

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

### faith (starts Primitive Age)

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

### engineering (starts Bronze Age)

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

### astronaut (starts Space Age; no building uses it yet)

Space: Cadet → Interstellar: Interstellar Pilot → Galactic: Galactic Explorer → **Quantum: Quantum Astronaut**

---

## Strategy

### Early game (Primitive and Stone Age)

Build gathering camps and huts together. Each camp holds 3 workers and auto-recruit fills only empty slots, so build camps as you build housing. A worker eats 0.06 food/tick, and a fully staffed camp makes 1.0 food/tick, so a camp covers its own three workers' drain (0.18) several times over. Once food is stable, build story circles to start research; for faster research, give knowledge a bigger share (`workers share knowledge 40`). The full walkthrough is [Your First Ten Minutes](first-ten-minutes.md).

### Mid game (Bronze and Iron Age)

- Lumber and masonry buildings are staffed as they finish, as far as housing and food allow. Stone and wood go into most building costs, so keep your housing ahead of your worker slots.
- A `war_camp` produces the `soldiers` resource from the Iron Age, spent on campaigns. On auto its slots draw workers like any other building's; if you don't need soldiers yet, `workers share military 0` keeps those workers for your producers.
- From the Iron Age, this age's masonry building mines ore that nothing spends, and the routine fills the newest buildings first. Build ore mines only for their milestones (see [What the ores are for](resources.md#what-the-ores-are-for)).

### Late game

- As food buildings get more productive, you need fewer food workers. Give food a small share (not 0) to move the rest elsewhere. The routine still sends extra recruits to food buildings when the food income can't feed them.
- Every worker eats the age's rate, and that rate keeps rising, so a large late-game population needs a large food income. Scale food before headcount.
- After an age advance, the new buildings fill from idle workers and new recruits. When housing is full, `unassign <old building> all` frees workers, and a minute later the routine puts them to work by your shares, newer buildings first. Or `assign <new building> all` places them yourself.

### Milestone: scholars_haven

Needs **50 knowledge workers** and **3 Libraries** built. Three libraries hold only 12 workers, so the other 38 have to be in your other knowledge buildings. Give knowledge a large share (`workers share knowledge 60`) and let the routine fill them as workers come. Track progress with `milestones` or `ms`.

### Specialize or spread

Focusing on one domain can finish milestone chains faster; spreading protects you from shortages. Worker shares are the lever: auto spreads your workers by worker slots, and a share focuses them (`workers share knowledge 50`). Early on, food and knowledge are enough; add military when campaigns open up and trade when gold becomes the bottleneck.

### Tips

- **Let the shares do the busywork.** Set a share when you want more workers in one domain or none in it, and `workers share auto` to go back.
- **Food first.** Auto-recruit recruits only while food allows, so early on your food buildings set the pace of growth.
- **Faith workers before epoch transitions.** Your faith as a share of faith storage sets the epoch and catastrophe odds (see [Faith](faith.md#faith-threshold-bands)).
- **Watch the idle count.** `status` (or `s`) shows it. The game puts idle workers to work every 5 ticks, so workers who stay idle have no free slot to go to: build more worker buildings.
- **`dismiss` cuts drain, `unassign` doesn't.** Idle workers still eat.

---

## See Also

- [Buildings](buildings.md): building lineages and worker slots per building
- [Resources](resources.md): what each lineage produces in which age
- [Morale](morale.md): the multiplier on everything your workers make
- [Build Plan](plan.md): builds that start as resources come in, staffed by the shares routine
- [Milestones](milestones.md): milestone chains that reward domain specialization
