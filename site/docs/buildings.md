# Buildings

AgeForge uses a **14-lineage system** with 301 total buildings spanning production lineages, storage, wonders, cultural monuments and a handful of one-off administrative buildings. Buildings belong to lineages that span the full 22-age arc. When you advance an age, buildings that have a next-tier equivalent do **not** change on their own. They get a pending upgrade marker, and you upgrade them at your own pace with the `upgrade` command.

For the full wonder list see [Wonders](wonders.md).

---

## Why Buildings Matter

Every resource you use comes from buildings, and the workers assigned to production buildings turn that capacity into output. More buildings give you:

- **More production.** Each extra building of the same tier adds its base rate to your total.
- **More worker slots.** Worker capacity grows with building count, which raises your production ceiling.
- **More storage.** Storage buildings add storage for every resource at once.
- **Progress toward the next age.** Many ages require a minimum building count or specific buildings.

Most of the game comes down to this: build more buildings and fill them with workers.

---

## Building Commands

```
build <key>              queue one building (you must have enough resources)
build <key> <count>      queue that many copies
build <key> max          build as many as you can afford right now
build                    list all available buildings with costs and count built
```

**`max`** builds as many copies as you can afford along the full cost curve and stops as soon as the next copy would cost more than you have. It does **not** divide your resources by the first copy's price; each copy is priced at its own step of the curve before the game decides to continue.

## Building Upgrades

When you advance to a new age, buildings that have a next-tier equivalent do **not** change on their own. Each one gets a **pending upgrade** marker and keeps producing at its old rate. You choose when to upgrade, one building type at a time.

Storage buildings never get an upgrade: your stashes, storage pits and vaults keep counting toward your storage for the rest of the run (see [Storage Buildings](#storage-buildings-21-tiers)).

### The upgrade hint

In the Economy panel's building list, a building with an upgrade waiting shows a gold hint line:

```
↑ Upgrade available → Forager Post   type: upgrade gathering_camp
```

The log also announces each upgrade when you advance. Until you upgrade, the old building keeps producing normally, with no penalty for leaving it pending. The reason to upgrade is the higher production rate of the new tier.

### Upgrade commands

```
upgrade <building>         upgrade ALL copies of that building (default)
upgrade <building> <n>     upgrade exactly n copies
upgrade <building> all     same as no count argument
```

There is no global `upgrade all` command. Upgrades are per building, so you control the order and pacing.

An upgrade stops at the new building's max count. If the target is capped and has room for fewer copies than you asked for, only that many are upgraded and the rest stay as they are; if it is already full, the upgrade is refused.

**Examples:**
```
upgrade gathering_camp
upgrade forager_post 3
upgrade forager_post all
```

### Upgrade cost

Each copy you upgrade costs the price of a new copy minus half of what the old copy cost. In other words, **you trade in the old building at 50% of its price toward the new one.** Old copies are traded in from the most expensive down, and new copies are priced up the new building's cost curve as usual. Build-cost reductions apply to the new copy's price but not to the trade-in value.

The cost of a copy never goes below zero, and upgrading is always cheaper than selling the old building and building the new one from scratch.

At high building counts (10 or more), the trade-in value of the most expensive old copies can cover the whole price of the new copy, so those upgrades are free. This is intentional: it rewards civilizations that invested heavily in a lineage before advancing.

### Workers and upgrades

- **Full upgrade** (all copies upgraded): workers move to the new building automatically. No reassignment needed.
- **Partial upgrade** (only some copies upgraded): workers stay on the old building as far as its remaining copies can hold them. Any extra workers move to the new building while it has room, and the rest go back to the idle pool.

### Legacy buildings (pending upgrade)

Once a building has a pending upgrade:

- It **keeps producing** at its current rate, with no penalty.
- You **cannot build more** copies of the old tier; `build` refuses it.
- You **can still** `assign`, `unassign` and `sell` copies of the old tier as normal.
- It counts toward your civilization stats as usual.

Don't let pending upgrades pile up during high-demand periods. Upgrading your food lineage before a resource squeeze is almost always the right call.

### Strategic advice

- **Upgrade high-count buildings early.** The most expensive old copies give the biggest trade-in, so a civilization with 15 Gathering Camps gets proportionally cheaper upgrades than one with 3.
- **Upgrade before a food crunch.** A Forager Post makes 1.5 food/tick fully staffed against a Gathering Camp's 1. If an epoch catastrophe is coming, upgraded food buildings give you a wider safety margin.
- **You control the order.** You might upgrade your food lineage as soon as you advance and leave military or knowledge buildings pending until you've banked enough resources. There is no time pressure, because pending buildings still produce.
- **Don't sell pending buildings for cash.** The 50% sell refund is already built into the upgrade price, so you get that value back when you upgrade. Selling instead throws away the upgrade discount.

---

### Selling Buildings

```
sell <key>           demolish 1 copy, recover 50% of its price
sell <key> <count>   demolish that many copies (most expensive first)
```

You can sell any building except wonders and storage from the **Stone Age onward**. Selling removes the most expensive copies first, and each copy refunds 50% of what it cost at its step of the cost curve.

**Worker handling:** if selling leaves fewer worker slots than workers assigned, the extra workers are unassigned and go back to the idle pool. They are not dismissed, so your population stays the same.

**Restrictions:**
- Wonders cannot be sold.
- Storage cannot be sold: once its age has passed, it can never be rebuilt.
- You cannot sell a building while any copy of it is in the build queue.
- Sell is not available in the Primitive Age.

### Cost Scaling

Each building has a first-copy price, and every copy after that costs a fixed percentage more than the one before. The price of the next copy depends on how many are already **built** plus how many are **in the build queue**:

```
cost of next copy = floor(first-copy price × scale ^ (built + queued))
```

When you buy several at once (`build <key> <count>` or `build <key> max`), each copy in the batch is priced at its own step and the total is the sum, not a flat multiple:

```
total cost = sum of floor(first-copy price × scale ^ (built + queued + i))
             for i = 0 to count-1
```

So a batch always costs more per copy than the first copy alone, and queued buildings raise the price of your next purchase before they finish.

Most buildings (production, military, research, trade, diplomacy and monuments) cost **15% more** per copy than the last. Storage and housing buildings grow more gently, at **13%** per copy. Wonders have a flat price and are limited to one each. The scaling lineages still grow exponentially, so plan your production before bulk-buying.

### Build-cost reductions

A handful of **milestone rewards** (Master Builder, Grand Architect and others) and two **techs** (Civil Engineering −5%, Nanofabrication −8%) each cut build costs by 3 to 8%. The reductions add up and multiply the scaled cost above, and the cost never drops below 10% of what it would be without them. Stacked, the available reductions reach **−32%**. The cost `build` lists is the cost you are charged. You don't need to opt in; earning the milestone or completing the tech is enough.

**Example:** a Gathering Camp's first copy costs 16 wood and each copy costs 15% more. With none built or queued:
- 1st: 16 wood
- 2nd: 18 wood
- 5th: 27 wood
- Building 5 at once: 16+18+21+24+27 = **106 wood total** (not 16×5=80)

### Build times

Build times are capped by the age's pace: nothing takes longer to build than **1/6 of its age's target time** (wonders included), and storage buildings take at most **1/48** of it (you can queue several copies at once, up to the cap). In practice that is 2 minutes in the Primitive Age, 1 hour in the Renaissance and 4 hours from the Interstellar Age on (storage: 19 seconds, 8 minutes and 30 minutes). The full per-age list is in the table under [How Production Rates Are Set](#how-production-rates-are-set).

---

## Lineage System

Each lineage starts at a specific age and gets a new tier in most ages after that. When your civilization enters a new age:

1. Buildings in active lineages that have a next tier get a **pending upgrade** marker.
2. The old tier becomes **legacy**: it keeps producing at its old rate but cannot be built again.
3. The new tier becomes available to build (fresh copies) and to upgrade into (from old copies).

You keep all your progress and production. A Gathering Camp in the Food lineage upgrades to a Forager Post, then a Farm, then Field Works; use `upgrade <building>` after each advance to convert your existing copies. See [Building Upgrades](#building-upgrades) for cost details and strategy.

---

## The 14 Production Lineages

| # | Lineage | Domain | Primary Output | Tiers | First Building | Final Building |
|---|---------|--------|----------------|-------|----------------|----------------|
| 1 | Housing | None | housing | 22 | Hut (+10 housing) | Transcendent Nexus (+15.7M housing) |
| 2 | Food | food | food | 21 | Gathering Camp | Quantum Cultivator |
| 3 | Organic Extraction | lumber | wood → coal → oil → nanobots → quantum flux | 21 | Wood Camp | Reality Harvester |
| 4 | Geological Extraction | masonry | stone → iron ore → uranium → titanium ore → dark matter crystals → antimatter | 21 | Stone Camp | Reality Excavator |
| 5 | Knowledge | knowledge | knowledge | 21 | Story Circle | Reality Academy |
| 6 | Faith | faith | faith | 21 | Shrine | Transcendence Hall |
| 7 | Military | military | soldiers | 22 | War Camp | Omniversal War Council |
| 8 | Trade | trade | gold | 20 | Market | Omniversal Bazaar |
| 9 | Engineering | engineering | iron → steel → electricity → plasma → dark matter → quantum flux | 20 | Smithy | Singularity Engine |
| 10 | Culture/Arts | None | culture | 17 | Amphitheater | Reality Art Engine |
| 11 | Metallurgy | metallurgy | iron → steel → titanium → dark matter → antimatter → quantum flux | 18 | Smelter | Quantum Metal Works |
| 12 | Energy | energy | coal → electricity → plasma → dark matter → quantum flux | 13 | Coal Plant | Zero Point Generator |
| 13 | Hacker/Digital | hacker | data | 8 | Server Farm | Reality Processor |
| 14 | Harbor | trade | gold (+trade-route income) | 5 | Harbor | Logistics Hub |

> Lineages 8 to 14 start in later ages (Trade and Engineering in the Bronze Age, Culture/Arts in the Classical, Metallurgy in the Iron, Energy in the Industrial, Hacker/Digital in the Information and Harbor in the Colonial), so they have fewer tiers. The Housing and Culture/Arts lineages have no worker domain: they work without workers. The **Harbor** lineage also raises trade income: besides producing gold, each harbor adds a percentage bonus to the income of *every* active trade route. See [Trade & Diplomacy](trade.md#harbor-lineage-trade-route-income).

---

## What Buildings Do

A building can have one or more of these effects. Knowing them helps you decide what to build.

| Effect | What It Does |
|--------|-------------|
| Production | Adds an amount of a resource each tick, scaled by how many workers it has (see below). Most production buildings do this. |
| Storage | Raises the storage of one resource, or storage for every resource (storage buildings). |
| Housing | Raises your housing, the number of workers you can have (housing buildings). |
| Bonus | A percentage bonus to a rate, such as the wonders' bonus to all production. |
| Morale | Restores a little morale every tick (Faith and Culture/Arts buildings). |
| Opinion | Raises opinion with every civilization you are not hostile with, each tick, scaled by workers (embassies). |

The **Faith lineage** (shrines, temples and their later tiers) and the **Culture/Arts lineage** restore civilization morale every tick just by existing; they don't need workers for it. That makes them your main way to push morale above neutral, where it raises all worker output. See [Morale](morale.md).

### Worker Scaling Formula

Production buildings with a worker domain use:

```
actual_rate = base_rate × count × (0.20 + 0.80 × workers_assigned / total_capacity)
```

At **0 workers** a building still produces **20% of its base rate** (the idle floor). At **full staffing** it produces **100%**. An empty building isn't wasted, only throttled.

Buildings without a worker domain (Housing, Culture/Arts) produce exactly `base_rate × count` whatever your workers are doing.

### How Production Rates Are Set

`base_rate` is the **fully staffed** rate, and it is the number each building's description shows.

For **construction resources** (anything the buildings of an age cost: wood, stone, iron, gold, steel, coal, electricity, data and so on), rates follow the **Payback Rule**: a producer's output is set so that, fully staffed, it earns back the price of its first copy in its age's **payback time**. The payback time is a share of how long the age is meant to take, and that share grows through the game: about 1/16 of the age in the Primitive Age, about 1/7 in the Iron Age, a quarter in the Renaissance (stretched to about a third there; see below), about a third in the Victorian Age and about two thirds in the Space Age. Later ages repay more slowly because every building you put up in earlier ages keeps producing alongside the new tier.

Output is valued at **price parity**. Each age has a price level for each resource (the typical first-copy price in that resource among the age's buildings), and resources are worth each other in the ratio of those levels. A building with two outputs splits its value between them. Because rates follow prices, they grow roughly 5 to 8x per age.

**Examples:**
- Wood Camp (Primitive): costs 16 wood and makes 0.569 wood/tick fully staffed, so it repays itself in 28 ticks (56 seconds).
- Stone Pit (Stone): costs 180 stone and 300 wood. At Stone Age parity (240 stone = 360 wood) that is worth 380 stone, and at 3.14 stone/tick it repays in about 4 minutes.

**Flow resources keep hand-set rates:** food, faith, culture and soldiers. They feed workers, set morale and epoch odds, fill culture storage or make up your army, and the requirements that ask for them are sized to those rates. Resources nothing in the age costs (marble and iron ore, for example, or knowledge outside the Medieval to Colonial Ages) also keep fixed rates.

| Age | Target time | Payback (fully staffed) | Build-time cap | Storage build cap |
|-----|-------------|-------------------------|----------------|-------------------|
| Primitive | 15m | 56s | 2m | 19s |
| Stone | 45m | 4m | 8m | 56s |
| Bronze | 1.5h | 11m | 15m | 2m |
| Iron | 2.5h | 22m | 25m | 3m |
| Classical | 3.5h | 38m | 35m | 4m |
| Medieval | 4.5h | 58m | 45m | 6m |
| Renaissance | 6h | 1.9h | 1h | 8m |
| Colonial | 7h | 2h | 1.2h | 9m |
| Industrial | 8h | 2.5h | 1.3h | 10m |
| Victorian | 9h | 3.2h | 1.5h | 11m |
| Electric | 10h | 3.9h | 1.7h | 12m |
| Atomic | 12h | 5.1h | 2h | 15m |
| Modern | 12h | 5.6h | 2h | 15m |
| Information | 14h | 7.1h | 2.3h | 18m |
| Digital | 16h | 8.7h | 2.7h | 20m |
| Cyberpunk | 18h | 10.6h | 3h | 22m |
| Fusion | 20h | 12.6h | 3.3h | 25m |
| Space | 22h | 14.7h | 3.7h | 28m |
| Interstellar | 24h | 17.1h | 4h | 30m |
| Galactic | 24h | 18.1h | 4h | 30m |
| Quantum | 24h | 19.1h | 4h | 30m |
| Transcendent | 24h | 20.2h | 4h | 30m |

The Renaissance's payback is 1.3x what the curve gives (1.5 hours): it is where gold income jumps, and at the curve's rate the age would run at barely half its target speed. Its University makes 208 knowledge/tick, Exchange 3.76K gold, Mill 520 steel, Foundry 711 steel and Coal Mine 289 coal.

Times are game time at 1x speed (1 tick = 2 seconds). Some construction resources have no producer in certain ages (stone after the Bronze Age, for example); the market sells them at parity instead. See [Resources](resources.md#buying-at-the-market).

---

## Lineage Output Progression

Several lineages change which resource they produce as you move through the ages. Each new tier simply produces the new resource, so you don't need to rebuild or reconfigure anything.

### Organic Extraction (Lineage 3)

| Ages | Output Resource |
|------|----------------|
| Primitive to Medieval | wood |
| Renaissance to Industrial | coal |
| Victorian to Information | oil |
| Digital to Fusion | nanobots |
| Space to Quantum | quantum flux |

**Nano Foundry** (Modern Age) is a standalone nanobot producer outside the lineage chain. It gives nanobots a producer as soon as they unlock, two ages before the lineage's own Bio Fabrication Lab. It uses `engineering` workers and makes +80 nanobots/tick.

### Geological Extraction (Lineage 4)

| Ages | Output Resource |
|------|----------------|
| Stone to Bronze | stone |
| Iron to Classical | marble and iron ore |
| Medieval to Industrial | iron ore |
| Victorian to Atomic | uranium |
| Modern to Digital | titanium ore |
| Cyberpunk to Space | dark matter crystals |
| Interstellar to Quantum | antimatter |

### Engineering (Lineage 9)

| Ages | Output Resource |
|------|----------------|
| Bronze to Medieval | iron |
| Renaissance to Victorian | steel |
| Electric to Digital | electricity |
| Cyberpunk to Space | plasma (with electricity in the Cyberpunk and Fusion Ages) |
| Interstellar to Galactic | dark matter |
| Quantum to Transcendent | quantum flux |

### Metallurgy (Lineage 11)

Metallurgy buildings produce metal directly; they do not use up ore.

| Ages | Output Resource |
|------|----------------|
| Iron to Medieval | iron |
| Renaissance to Atomic | steel |
| Modern to Digital | titanium |
| Cyberpunk to Fusion | dark matter |
| Space | titanium |
| Interstellar to Galactic | antimatter |
| Quantum | quantum flux |

---

## Legacy Buildings

When an age advance gives a lineage tier a pending upgrade, those buildings become legacy. Legacy buildings:

- Keep producing at their normal rate (you lose nothing when you advance)
- Cannot be built again
- Show a gold `↑ Upgrade available` hint with the target building's name in the Economy panel
- Count toward building totals in your civilization stats
- Can still have workers assigned and unassigned, and can be sold as normal

Your 20 Gathering Camps stay Gathering Camps, producing at their full rate, until you run `upgrade gathering_camp`. Production only improves once the upgrade is done, which is the reason to do it. See [Building Upgrades](#building-upgrades) for the full guide.

---

## Storage Buildings (21 tiers)

Storage buildings form their own lineage and add storage for **every** resource at once. They need no workers: build them and your storage goes up.

Every storage building is **capped at 25 copies** (Stash at 50). The cap is deliberate. Each tier's storage per copy is at least its largest build cost, so a full stack always adds more storage than it costs, and a storage building never becomes an unaffordable wall. The cost curve would eventually outrun even that, so the stack stops just short of that point. Build the next age's storage instead of over-stacking.

| Building | Age | Storage per copy (every resource) | Max |
|----------|-----|--------|-----|
| Stash | Primitive | +500 | 50 |
| Storage Pit | Stone | +2.2K | 25 |
| Warehouse | Bronze | +11K | 25 |
| Granary | Iron | +26K | 25 |
| Classical Vault | Classical | +110K | 25 |
| Strongroom | Medieval | +410K | 25 |
| Renaissance Vault | Renaissance | +2.7M | 25 |
| Colonial Warehouse | Colonial | +33M | 25 |
| Industrial Depot | Industrial | +170M | 25 |
| Victorian Vault | Victorian | +1.1B | 25 |
| Electric Warehouse | Electric | +3.5B | 25 |
| Atomic Vault | Atomic | +20B | 25 |
| Modern Depot | Modern | +90B | 25 |
| Info Vault | Information | +790B | 25 |
| Digital Archive | Digital | +1.5T | 25 |
| Cyber Vault | Cyberpunk | +8T | 25 |
| Fusion Vault | Fusion | +30T | 25 |
| Orbital Depot | Space | +200T | 25 |
| Stellar Vault | Interstellar | +2Q | 25 |
| Galactic Vault | Galactic | +20Q | 25 |
| Quantum Vault | Quantum | +200Q | 25 |

A full stack of an age's storage (with every earlier age's) holds at least an hour and a half of that age's typical production of each resource it builds with, so a player who checks in every hour or so loses nothing to full storage. For longer absences the [build plan](plan.md) spends income as it arrives and [wonder overflow](wonders.md#overflow) banks what full storage would waste.

> **Tip:** Stash is capped at 50. Build them out before you leave the Primitive Age, then start on Storage Pits as soon as you enter the Stone Age. Full storage stops all resource accumulation, so build storage first whenever you enter a new age.

Storage buildings **never upgrade** and are never offered as upgrades when you advance. Storage adds up: every storage building you have built keeps adding its capacity for the rest of the run, so the stashes from the Primitive Age still count in the Quantum Age. Like every other building, an older age's storage can no longer be built once you advance (the game tells you to build the current age's storage instead), so fill each tier while it is current.

Storage is also permanent: catastrophes never destroy it or turn it into ruins, and it can't be sold.

---

## Cultural Monuments (4)

Cultural Monuments are one-off structures (one copy each) that turn surplus **culture** into a permanent payoff. Each costs a large lump of culture plus other materials of its age, and gives a permanent bonus to all production once built. Unlike wonders, they need no resource banking: build them with the normal `build <key>` command.

| Monument | Age | Culture Cost | Permanent Bonus |
|----------|-----|--------------|-----------------|
| Cultural Obelisk | Classical | 710 | +1% all production |
| Grand Amphitheater | Medieval | 7.1K | +2% all production |
| Eternal Library | Industrial | 140K | +3% all production |
| Monument of Ages | Modern | 7.1M | +5% all production |

Monuments are one way to spend culture; the `festival` command is another. See [Culture](resources.md#culture).

---

## The Geographic Society

Unlocked in the **Industrial Age**, the Geographic Society is the one building that plays part of the game for you. It costs gold, steel and coal, holds **8 military workers**, and produces no resource. Instead it **sends out scouting expeditions on its own**, so an empire left to run keeps exploring and keeps meeting the world's civilizations, which is where civilization boons come from.

You can build as many as you like, and the pace scales with how many you've built and how fully you've staffed them: one unstaffed Society sends a party roughly every 900 ticks (about **30 minutes** at 1x speed), and six fully staffed ones bottom out at around one every 100 ticks (about **3m 20s**). These times shrink as your tick speed rises; the **Factions** panel shows the countdown to the next dispatch in wall-clock time. It sends **scouting parties only**, never military campaigns. It never runs two at once or bypasses the single scouting slot, and it pays the full resource cost of every party, skipping the cycle if you can't cover it.

It is deliberately **slower than sending expeditions yourself**: even fully built and staffed, it runs at about 60% of the pace of a player chaining expeditions by hand. Treat it as a floor under your exploration rather than a substitute for it. Full details in [Military & Expeditions](military.md#automatic-dispatch-the-geographic-society).

---

## Wonders (22)

Wonders are unique buildings: you can build each wonder only once. You bank resources toward a wonder before you can build it:

```
wonder collect <resource> <amount|all>   bank resources toward the active wonder
build <wonder_key>                       start construction once it is fully funded
```

You must build each age's wonder before you can advance to the next age. A wonder's price adds up to 40 **price units** of its age, where one price unit is what a typical building of that age charges in a single resource (valued at [price parity](#how-production-rates-are-set)); each wonder keeps its own mix of resources. A wonder takes at most 1/6 of the age's target time to build.

Wonders give civilization-wide bonuses.

Endure and the Great Fire never destroy wonders, and Succumb never turns them into ruins. They stay for the rest of the run, like storage.

See [Wonders](wonders.md) for the full list with costs and effects.

---

## Ruins

Ruins are what is left of a previous civilization. You get them when you choose **Succumb** during a catastrophe: up to 8 of your buildings (never wonders or storage) become ruins and carry forward into your new run. You can hold at most 24 ruins; when a new batch goes over, the lowest-value ruins (earliest age, then lowest output) crumble first.

**How ruins differ from normal buildings:**

| Property | Normal Building | Ruin |
|----------|----------------|------|
| Production | `base × count × (0.20 + 0.80 × fill)` | `base × count × 0.50` |
| Workers | Needed for full output | Not needed (produces on its own) |
| Can build more? | Yes | No |
| Destroyed by Endure? | Yes (20% of buildings at random, fewer if you brace; never wonders or storage) | No |

Ruins give your new run a free head start. They produce at 50% of base rate with no workers, so food, wood and stone (or whatever the ruined building made) trickle in from the first tick without any setup.

Ruins build up across runs: each Succumb adds up to 8 new ruins on top of the ones you already have, up to the cap of 24.

---

## Worker Capacity

Every production building has a number of worker slots per copy. The worker slots a building type gives you are:

```
total_capacity = building_count × slots per copy
```

More buildings mean more worker slots and a higher production ceiling. Building more copies of the current tier is how you grow output within an age.

Recruit workers with `recruit [count|max]` and assign them with `assign <building> [count|all]`. A building with no workers produces only 20% of its base rate.

---

## Culture Buildings (Lineage 10)

Culture buildings have no worker domain. They produce culture every tick at their base rate, with or without workers, and each one also adds culture storage. Build them to earn culture passively.

They also **restore morale every tick** (see [Morale](morale.md)), so they do two jobs: culture for epoch outcomes and morale for a production bonus.

How full your culture storage is decides which good epoch events you can get, so a civilization rich in culture comes through epoch transitions better.

---

## Economy Panel

The Economy panel is always visible behind the other panels. Its right side lists your buildings grouped by age, with their workers and any upgrade hint; the left side shows your resource rates and the construction queue. Scroll the building list with PgUp and PgDn.

Use `rates` to print the current production and consumption rates for all resources in the command output, and `status` for a detailed status summary.
