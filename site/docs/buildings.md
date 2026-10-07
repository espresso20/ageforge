# Buildings

AgeForge uses a **14-lineage system** with 301 total buildings spanning production lineages, storage, wonders, cultural monuments and four standalone buildings (two embassies, the Geographic Society and the Nano Foundry). Buildings belong to lineages that span the full 22-age arc. When you advance an age, buildings that have a next-tier equivalent do **not** change on their own. They get a pending upgrade marker, and you upgrade them at your own pace with the `upgrade` command.

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

`b` is short for `build`. To build copies as the resources come in, rather than now, add them to the [build plan](plan.md) with `plan build <key> [count]`.

<figure class="screen" data-screen="buildings"><figcaption>The Buildings panel in the Bronze Age, scrolled to the current age: each building with the key that build takes, what its next copy costs and how its worker slots are filled.</figcaption></figure>

**`max`** builds as many copies as you can afford along the full cost curve and stops as soon as the next copy would cost more than you have. It does **not** divide your resources by the first copy's price; each copy is priced at its own step of the curve before the game decides to continue.

## Building Upgrades

When you advance to a new age, buildings that have a next-tier equivalent do **not** change on their own. Each one gets a **pending upgrade** marker and keeps producing at its old rate. You choose when to upgrade, one building type at a time.

Storage buildings never get an upgrade: your stashes, storage pits and vaults keep counting toward your storage for the rest of the run (see [Storage Buildings](#storage-buildings-21-tiers)).

### The upgrade hint

In the Economy panel's building list, a building with an upgrade waiting shows a gold hint line under it:

```
↑ Upgrade available: Forager Post. Type: upgrade gathering_camp
```

The log also lists each upgrade when you advance (`↑ 12 Gathering Camps can upgrade to Forager Post. Type 'upgrade gathering_camp'.`). Until you upgrade, the old building keeps producing normally, with no penalty for leaving it pending. The reason to upgrade is the higher production rate of the new tier.

### Upgrade commands

```
upgrade                    list the upgrades you can make
upgrade <building>         upgrade ALL copies of that building (default)
upgrade <building> <n>     upgrade exactly n copies
upgrade <building> all     same as no count argument
```

Bare `upgrade` prints one line per building with an upgrade waiting: the old and new keys, how many copies, the cost to upgrade all of them, and a check mark if you can afford it (a cross if not).

There is no global `upgrade all` command. Upgrades are per building, so you control the order and pacing. An upgrade takes effect at once: there is no build time.

An upgrade stops at the new building's max count. If the target is capped and has room for fewer copies than you asked for, only that many are upgraded and the rest stay as they are; if it is already full, the upgrade is refused.

**Examples:**
```
upgrade gathering_camp
upgrade forager_post 3
upgrade forager_post all
```

### Upgrade cost

Each copy you upgrade costs the price of a new copy minus the old copy's sell value (half of what it cost). In other words, **you trade in the old building at 50% of its price toward the new one.** The sum is done per resource, and no resource goes below zero. Old copies are traded in from the most expensive down, and new copies are priced up the new building's cost curve as usual. Build-cost reductions apply to the new copy's price but not to the trade-in value. **Interchangeable Parts**, the Industrial Age's Craft capstone, then takes 15% off what is left to pay, in every resource.

Because it is done per resource, the trade-in only counts in resources the new building costs. If the old building cost wood and the new one costs none, that part of its value is lost; so is any trade-in beyond the new price in a resource. An upgrade therefore never costs less than selling the old copy and building a new one, and sometimes a little more. What it saves is time: it is instant, with no build time, and the workers stay.

How much the trade-in covers depends on how many old copies you have. With 10, the most expensive copy typically covers 10% to 40% of the first new copy's price. Around 20 to 25 copies it can cover all of it, if the two buildings cost the same resources; those upgrades are free.

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

**Upgrade before you advance again.** Upgrade offers are made only when you advance, one tier at a time. Copies you carry through a second advance can still upgrade into the tier they were offered, but the copies that upgrade produces may never be offered another upgrade (Wood Camps upgraded to Woodcutter Camps in the Bronze Age, for example, get no offer in the Iron Age or after). Build the newer tier fresh instead, and clear each age's upgrades before you leave it.

### Strategic advice

- **Upgrade high-count buildings early.** The most expensive old copies give the biggest trade-in, so a civilization with 15 Gathering Camps gets proportionally cheaper upgrades than one with 3.
- **Upgrade before a food crunch.** A Forager Post makes 1.5 food/tick fully staffed against a Gathering Camp's 1. If a [harbinger](harbinger.md) warns that a catastrophe is coming, upgraded food buildings give you a wider safety margin.
- **You control the order.** You might upgrade your food lineage as soon as you advance and leave military or knowledge buildings pending until you've banked enough resources. Within the age there is no time pressure, because pending buildings still produce; just finish before your next advance (see above).
- **Keep pending buildings until you upgrade them.** Selling a pending copy pays back its 50% at once, about what it brings as a trade-in, but the copy stops producing. Sell only when you need those resources right now.
- **A tech-gated tier waits for its tech.** In the ages where the new tier needs a tech first (see [Buildings a tech opens](#buildings-a-tech-opens)), `upgrade` refuses until the tech is done, even though the hint already shows.

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

A handful of **milestone rewards** (Master Builder, Grand Architect and others) each cut build costs by 3 to 5%, and two **techs** (Civil Engineering and Nanofabrication) by 3% each. The milestones' reductions add up (−19% with all of them) and multiply the scaled cost above; each tech's cut then multiplies what is left (×0.97 each). The cost never drops below 10% of what it would be without them. Stacked, the available reductions reach **−24%**. The cost `build` lists is the cost you are charged. You don't need to opt in; earning the milestone or completing the tech is enough.

**Example:** a Gathering Camp's first copy costs 16 wood and each copy costs 15% more. With none built or queued:
- 1st: 16 wood
- 2nd: 18 wood
- 5th: 27 wood
- Building 5 at once: 16+18+21+24+27 = **106 wood total** (not 16×5=80)

### Build times

Build times are capped by the age's pace: nothing takes longer to build than **1/6 of its age's target time** (wonders included), and storage buildings take at most **1/48** of it (you can queue several copies at once, up to the cap). In practice that is 2m 30s in the Primitive Age, 2h 36m in the Renaissance and 10h 24m from the Interstellar Age on (storage: 18 seconds, 19m 30s and 1h 18m). The full per-age list is in the table under [How Production Rates Are Set](#how-production-rates-are-set).

These are first-run times. On known ground (an age a past run completed) every build time is divided by the age's [Era Mastery](prestige.md#era-mastery) speed, rounded up and never below one tick, and production and storage are multiplied by it. The time shown when you start a build is the real, shortened one.

### Buildings a tech opens

Most buildings unlock the moment you enter their age. Twenty-one wait for a tech from that same age instead, so the age has something new partway through:

| Building | Age | Opened by |
|---|---|---|
| Standing Stones | Stone | Ritual |
| Altar | Bronze | Calendar, the age's keystone |
| Barracks | Bronze | Military Tactics |
| Smelter | Iron | Iron Smelting |
| Legion Fort | Iron | Siege Warfare |
| Forge | Classical | Metal Casting |
| Cathedral | Medieval | Theology, the age's keystone |
| Foundry | Renaissance | Blast Furnace |
| Colonial Steelworks | Colonial | Coke Smelting |
| Harbor | Colonial | Mercantilism |
| Embassy | Colonial | Embassies |
| Coal Plant | Industrial | Steam Power |
| Geographic Society | Industrial | Geographic Societies |
| Grand Embassy | Industrial | Concert of Nations, a capstone |
| Steam Works | Victorian | Electrification |
| Dynamo Hall | Electric | Power Distribution, the age's keystone |
| Nuclear Plant | Atomic | Civilian Reactors |
| Smart Farm | Information | Internet of Things |
| Smart Complex | Information | Internet of Things |
| Holographic Theater | Cyberpunk | Holography |
| Energy Exchange | Fusion | Maglev Transit |

Until the tech is done the building is missing from the build list, and `build` and `upgrade` say which tech it needs. Each of these buildings is its lineage's tier for that age, so the upgrade hint for the old tier shows when you advance, but the upgrade itself waits for the tech. The build plan takes it early and waits. Copies you already have keep working. See [Technologies](technologies.md#tech-tree-by-age).

A tech only takes a building when the age keeps another way to make the same thing (the Mill beside the Foundry, the Port beside the Harbor), or when the tech is one a run researches anyway. A building that is its age's only source of what it makes is open from the first tick of its age: every age's knowledge, culture and military building, the Steam Mine, the Uranium Mine, the Financial District, the Corporate HQ and the Petroleum Refinery among them.

Wonders wait for a tech too, in their own way: from the Stone Age on each one needs its age's **keystone** before `build` will start it, but it is listed from the first tick of its age and its bank takes deposits and overflow the whole time. See [The keystone tech](wonders.md#the-keystone-tech).

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

> Lineages 8 to 14 start in later ages (Trade and Engineering in the Bronze Age, Culture/Arts in the Classical, Metallurgy in the Iron, Energy in the Industrial, Hacker/Digital in the Information and Harbor in the Colonial), so they have fewer tiers. The Housing and Culture/Arts lineages have no worker domain: they work without workers. The **Harbor** lineage also raises trade income: besides producing gold, each harbor adds a percentage bonus to the income of *every* active trade route. See [Trade](trade.md#harbor-lineage-trade-route-income).

---

## What Buildings Do

A building can have one or more of these effects. Knowing them helps you decide what to build.

| Effect | What It Does |
|--------|-------------|
| Production | Adds an amount of a resource each tick, scaled by how many workers it has (see below). Most production buildings do this. |
| Storage | Raises the storage of one resource, or storage for every resource (storage buildings). |
| Housing | Raises your housing, the number of workers you can have (housing buildings). |
| Bonus | A percentage bonus to a rate, such as a wonder's or monument's bonus to all production (which adds into [the all-production pool](resources.md#the-all-production-cap)). |
| Trade route income | Raises the income of every trade route (Harbor lineage). |
| Morale | Restores a little morale every tick (Faith and Culture/Arts buildings). |
| Opinion | Raises opinion with the non-hostile civilizations you have met, each tick, scaled by workers (the two embassies; see [Factions & Diplomacy](factions.md#embassy-buildings)). |

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

For **construction resources** (anything the buildings of an age cost: wood, stone, iron, gold, steel, coal, electricity, data and so on), rates follow the **Payback Rule**: a producer's output is set so that, fully staffed, it earns back the price of its first copy in its age's **payback time**. The payback time is a share of how long the age is meant to take, and that share grows through the game: about 1/16 of the age in the Primitive Age, about 1/9 in the Iron Age, a sixth in the Renaissance (stretched to more than a quarter there, and shortened in the Information and Cyberpunk Ages; see below), about 2/9 in the Victorian Age and about a third in the Space Age. Later ages repay more slowly because every building you put up in earlier ages keeps producing alongside the new tier.

Output is valued at **price parity**. Each age has a price level for each resource (the typical first-copy price in that resource among the age's buildings), and resources are worth each other in the ratio of those levels. A building with two outputs splits its value between them. Because rates follow prices, they grow roughly 5 to 8x per age.

**Examples:**
- Wood Camp (Primitive): costs 16 wood and makes 0.569 wood/tick fully staffed, so it repays itself in 28 ticks (56 seconds).
- Stone Pit (Stone): costs 180 stone and 300 wood. At Stone Age parity (240 stone = 360 wood) that is worth 380 stone, and at 3.48 stone/tick it repays in about 3 minutes 40 seconds.

**Flow resources keep hand-set rates:** food, faith, culture and soldiers. They feed workers, set morale and epoch odds, fill culture storage or make up your army, and the requirements that ask for them are sized to those rates. Resources nothing in the age costs (marble and iron ore, for example, or knowledge outside the Medieval to Colonial Ages) also keep fixed rates.

| Age | Target time | Payback (fully staffed) | Build-time cap | Storage build cap |
|-----|-------------|-------------------------|----------------|-------------------|
| Primitive | 15m | 56s | 2m 30s | 18s |
| Stone | 45m | 4m | 7m 30s | 56s |
| Bronze | 3h 54m | 25m | 39m | 4m 52s |
| Iron | 6h 30m | 45m | 1h 5m | 8m 6s |
| Classical | 9h 6m | 1.2h | 1h 31m | 11m 22s |
| Medieval | 11h 42m | 1.8h | 1h 57m | 14m 36s |
| Renaissance | 15h 36m | 5.2h | 2h 36m | 19m 30s |
| Colonial | 18h 12m | 3.4h | 3h 2m | 22m 44s |
| Industrial | 20h 48m | 4.2h | 3h 28m | 26m |
| Victorian | 23h 24m | 8.7h | 3h 54m | 29m 14s |
| Electric | 26h | 8.8h | 4h 20m | 32m 30s |
| Atomic | 31h 12m | 13.7h | 5h 12m | 39m |
| Modern | 31h 12m | 8.3h | 5h 12m | 39m |
| Information | 36h 24m | 8.2h | 6h 4m | 45m 30s |
| Digital | 41h 36m | 12.4h | 6h 56m | 52m |
| Cyberpunk | 46h 48m | 11.7h | 7h 48m | 58m 30s |
| Fusion | 52h | 17.1h | 8h 40m | 1h 5m |
| Space | 57h 12m | 19.7h | 9h 32m | 1h 11m 30s |
| Interstellar | 62h 24m | 22.5h | 10h 24m | 1h 18m |
| Galactic | 62h 24m | 23.4h | 10h 24m | 1h 18m |
| Quantum | 62h 24m | 24.4h | 10h 24m | 1h 18m |
| Transcendent | 62h 24m | 25.3h | 10h 24m | 1h 18m |

The Renaissance's payback is 2x what the curve gives (about 2.6 hours): it is where gold income jumps, and at the curve's rate the age would run at barely half its target length. It was 1.3x while the age's requirement also asked for 30M knowledge; that requirement is gone (see [Keystone Techs](ages.md#keystone-techs)), and the payback carries its share. It was 1.7x until the Stone and Iron Eras gained their techs: a run now arrives with a fifth more knowledge and cheaper, quicker building, and the age had dropped to three quarters of its target. Its University makes 76.3 knowledge/tick, Exchange 1.38K gold, Mill 191 steel, Foundry 261 steel and Coal Mine 106 coal.

The Electric Era's ages repay more slowly too, since the Steel and Electric Eras gained their techs: Victorian 1.7x, Electric 1.45x and Atomic 1.75x what the curve gives. With nine techs an age where there were three, the keystone costs a tenth of what the age makes of knowledge, where it cost a quarter to a half, and research no longer holds those ages up; their buildings and their requirements do. A Bessemer Plant makes 141K steel/tick (was 240K), an Electric Arc Furnace 874K (was 1.27M) and a Breeder Reactor 1.26M electricity (was 2.21M). The Bronze Age's payback is 1.1x (a Lumber Mill makes 7.91 wood/tick, was 8.7; a Quarry 4 stone, was 4.4): the age ran at under two thirds of its target.

The Information and Cyberpunk Ages go the other way, at 0.8x: they ran 1.2 to 1.5x their targets, and the extra time was spent waiting.

Times are at the base tick of 2 seconds; game speed bonuses shorten them in real time. Some construction resources have no producer in certain ages (stone after the Bronze Age, for example); the market sells them at parity instead. See [Resources](resources.md#buying-at-the-market).

---

## Lineage Output Progression

Several lineages change which resource they produce as you move through the ages. Each new tier simply produces the new resource, so you don't need to rebuild or reconfigure anything.

**Some tiers arrive before their resource does.** A resource gathers nothing until the age that unlocks it, whatever makes it. These buildings say `once unlocked` beside that output in their description:

| Building (age) | Output | Unlocks in |
|---|---|---|
| War Camp (Stone), Barracks (Bronze) | soldiers | Iron Age |
| Colosseum (Iron) | culture | Classical Age |
| Uranium Mine (Victorian), Nuclear Extraction Plant (Electric) | uranium | Atomic Age |
| Titanium Mine (Modern), Precision Mine (Information), Nano Drill Complex (Digital) | titanium ore | Space Age |
| Titanium Smelter (Modern), Aerospace Foundry (Information), Nano Alloy Plant (Digital) | titanium | Space Age |
| Augmentation Foundry (Cyberpunk) | plasma (its electricity counts at once) | Fusion Age |
| Dark Matter Refinery (Cyberpunk), Exotic Matter Forge (Fusion) | dark matter | Interstellar Age |
| Quantum Organic Extractor (Space), Reality Matter Weaver (Interstellar), Cosmic Organic Works (Galactic) | quantum flux | Quantum Age |
| Stellar Core Drill, Antimatter Forge (Interstellar) | antimatter | Galactic Age |

Until then they count toward age requirements and milestones and make nothing.

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
| Storage Pit | Stone | +2.75K | 25 |
| Warehouse | Bronze | +21K | 25 |
| Granary | Iron | +35K | 25 |
| Classical Vault | Classical | +170K | 25 |
| Strongroom | Medieval | +710K | 25 |
| Renaissance Vault | Renaissance | +4.6M | 25 |
| Colonial Warehouse | Colonial | +60M | 25 |
| Industrial Depot | Industrial | +340M | 25 |
| Victorian Vault | Victorian | +1.3B | 25 |
| Electric Warehouse | Electric | +5.5B | 25 |
| Atomic Vault | Atomic | +23B | 25 |
| Modern Depot | Modern | +110B | 25 |
| Info Vault | Information | +2.3T | 25 |
| Digital Archive | Digital | +2.7T | 25 |
| Cyber Vault | Cyberpunk | +22T | 25 |
| Fusion Vault | Fusion | +38T | 25 |
| Orbital Depot | Space | +200T | 25 |
| Stellar Vault | Interstellar | +2Q | 25 |
| Galactic Vault | Galactic | +34Q | 25 |
| Quantum Vault | Quantum | +200Q | 25 |

A full stack of an age's storage (with every earlier age's) holds at least **4.5 hours** of that age's typical production of each resource it builds with from the Bronze Age on, and an hour and a half in the Primitive and Stone Ages, which fill fast and are meant to. (That is with the storage techs researched: Pottery and Masonry add 10% each to every store, and Cloud Computing and Superconductors 8% each later.) A player who checks in every few hours loses little to full storage. For longer absences the [build plan](plan.md) spends income as it arrives, and what full storage would still waste goes to the [wonder](wonders.md#overflow) and then [toward the plan's next copies](plan.md#overflow-pays-the-plan).

> **Tip:** Stash is capped at 50. Build them out before you leave the Primitive Age, then start on Storage Pits as soon as you enter the Stone Age. A full store stops that resource piling up: what it would waste goes into the current wonder's bank while the wonder still needs it (unless you typed `wonder overflow off`), then toward your plan's next copies, and anything neither needs is lost. So build storage first whenever you enter a new age.

Storage buildings **never upgrade** and are never offered as upgrades when you advance. Storage adds up: every storage building you have built keeps adding its capacity for the rest of the run, so the stashes from the Primitive Age still count in the Quantum Age. Like every other building, an older age's storage can no longer be built once you advance (the game tells you to build the current age's storage instead), so fill each tier while it is current.

Storage is also permanent: catastrophes never destroy it or turn it into ruins, and it can't be sold.

---

## Cultural Monuments (4)

Cultural Monuments are one-off structures (one copy each) that turn surplus **culture** into a lasting payoff. Each costs a large lump of culture plus other materials of its age, and gives a bonus to all production while it stands (an Endure can destroy it like any building but wonders and storage). Unlike wonders, they need no resource banking: build them with the normal `build <key>` command. Their bonuses add into the same pool as every other all-production bonus but a tech's. The pool applies in full up to +200% and a quarter of every point past it, so a monument built late still raises your output, by a quarter of what it lists. See [The all-production cap](resources.md#the-all-production-cap).

| Monument | Age | Culture Cost | Bonus |
|----------|-----|--------------|-----------------|
| Cultural Obelisk | Classical | 710 | +1% all production |
| Grand Amphitheater | Medieval | 7.1K | +2% all production |
| Eternal Library | Industrial | 140K | +3% all production |
| Monument of Ages | Modern | 7.1M | +5% all production |

Monuments are one way to spend culture; the `festival` command is another. See [Culture](resources.md#culture).

---

## The Geographic Society

Unlocked in the **Industrial Age** by the **Geographic Societies** tech (which needs Cartography), the Geographic Society is the one building that plays part of the game for you. It costs gold, steel and coal, holds **8 military workers**, and produces no resource. Instead it **sends out scouting expeditions on its own**, so an empire left to run keeps exploring and keeps meeting the world's civilizations, which is where civilization boons come from.

You can build as many as you like, and the pace scales with how many you've built and how fully you've staffed them: one unstaffed Society sends a party about every 2,340 ticks (**1h 18m** at the base tick), and six fully staffed ones bottom out at around one every 260 ticks (about **8m 40s**). Game speed bonuses shorten these in real time; the **Factions** panel shows the countdown to the next dispatch in wall-clock time. It sends **scouting parties only**, never military campaigns. It uses your one scouting slot only when it is free, and it pays the full resource cost of every party: if your stores can't outfit one, it waits and sends it as soon as they can.

It is deliberately **slower than sending expeditions yourself**: even fully built and staffed, it runs at about 60% of the pace of a player chaining expeditions by hand. Treat it as a floor under your exploration rather than a substitute for it. Full details in [Army & Missions](military.md#automatic-dispatch-the-geographic-society).

---

## Wonders (22)

Wonders are unique buildings: you can build each wonder only once. You bank resources toward a wonder before you can build it:

```
wonder collect <resource|all> [amount|all|max]   bank resources toward the current wonder
build <wonder_key>                              start construction once it is fully funded
```

Full stores bank into the wonder on their own ([overflow](wonders.md#overflow)), and a wonder in your [build plan](plan.md) banks and starts by itself once what you hold covers the rest.

You must build each age's wonder before you can advance to the next age. A wonder's price adds up to 40 **price units** of its age, where one price unit is what a typical building of that age charges in a single resource (valued at [price parity](#how-production-rates-are-set)); each wonder keeps its own mix of resources. A wonder takes at most 1/6 of the age's target time to build.

Every wonder adds a little production of one or two resources, and many add more: a percentage to all production (the Crystal Palace, Hoover Dam and the last four wonders), to knowledge output or to expedition rewards, or housing or storage. All-production bonuses add into [the all-production pool](resources.md#the-all-production-cap), which applies in full up to +200% and a quarter of every point past it.

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

The Economy panel is always visible behind the other panels. Its right side lists your buildings grouped by age, with their workers and any upgrade hint; the left side shows your resource rates and the construction queue. Scroll the building list with PgUp and PgDn, and turn the Resources box to its next page with Ctrl+R when it has more resources than it can show.

Use `rates` to print the current production and consumption rates for all resources in the command output, and `status` for a detailed status summary.
