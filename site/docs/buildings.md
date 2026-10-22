# Buildings

AgeForge uses a **14-lineage system**, split into **39 building lines**, with 302 total buildings spanning production lineages, storage, wonders, cultural monuments and four standalone buildings (two embassies, the Geographic Society and the Nano Foundry). Buildings belong to lineages that span the full 22-age arc, and each lineage is made of lines: a line makes one resource from its first building to its last. A building from an earlier age is an **older building**: it keeps producing, and you can upgrade it to its line's building for your current age at half the price of a new copy with the `upgrade` command.

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

A building from an earlier age is an **older building**. It keeps producing what it always made. You can upgrade it to its line's building for your current age, at half the price of a new copy, and it then counts as that building.

A building is older when its line has a newer building for the age you are in. Nothing is stored or marked: it is read from the line and the age, so it follows you through every advance. A Hut in the Iron Age is an older building, and it upgrades straight to the Iron Age's house, not one step at a time. A building only upgrades along its own line, to a building that makes the same resource (see [The 39 Building Lines](#the-39-building-lines)). A line that has no newer building for your age keeps its last building as its current one, so the last building of a line, such as the Sawmill, is never older, and a line that pauses for an age or two has nothing to upgrade to until it resumes.

Some buildings are in no line and are never older: storage (your stashes, storage pits and vaults keep counting toward your storage for the rest of the run, see [Storage Buildings](#storage-buildings-21-tiers)), knowledge buildings (they keep their tier, and each age's knowledge building is built new in its own age, see [Scholars and Research](knowledge.md#what-a-scholar-is)), wonders and monuments.

When you advance, the log names the buildings that became older (`↑ 12 Gathering Camps can upgrade to Forager Post. Type 'upgrade gathering_camp'.`). Leaving them as they are costs nothing: they keep producing at their old rate. You cannot build new copies of an older building; `build` says so and names the building it upgrades to.

### Upgrade commands

```
upgrade                    list the older buildings you hold
upgrade <building>         upgrade as many copies as you can pay for now
upgrade <building> <n>     upgrade up to n copies
upgrade <building> all     upgrade every copy you hold, as far as you can pay
```

Bare `upgrade` prints one line per older building you hold: the old and new keys, how many copies you hold, the price of the next copy, how many you can pay for now, and the reason if something blocks the next one (a tech to research, or a price larger than a store can hold).

Copies are upgraded one at a time, each at its own price, and the command stops at the first copy it cannot pay for. It says how many it upgraded, what they cost, and what the next copy costs or why it stopped. An upgrade takes effect at once: there is no build time. Upgrades are per building, so you control the order. An upgrade stops at the new building's max count, if it has one.

**Examples:**
```
upgrade gathering_camp
upgrade gathering_camp 3
upgrade gathering_camp all
```

### Upgrade cost

One upgraded copy costs **half** of what the next new copy of the building it becomes would cost right now, in every resource. The price is worked out the way a build is priced, with the same cost curve and the same build-cost cuts, and copies are priced one at a time: after each upgrade there is one more of the new building, so the next copy costs more. **Interchangeable Parts**, the Industrial Age's Craft capstone, takes another 15% off an upgrade's price.

The old copy is given up, and nothing is paid back for it. Selling an older building pays back half of its own price, but the copy stops producing, so an older building is better kept or upgraded than sold.

A price larger than a store can hold cannot be paid, however long you wait: the command refuses that copy and names the store. Build more storage first.

### Workers and upgrades

An upgraded copy keeps its workers. They move from the old building's slots to the new building's, one copy's share at a time. If the new building has fewer slots, the extra workers go back to the idle pool.

### Counting toward the next age

An upgraded copy is the building it became. It counts for the next age's building requirements, for milestones and badges, and for the price of the next copy. A requirement that names an older building (a milestone that asks for Wood Camps, a gate that asks for Stone Camps, a trade route that needs Markets) also counts the later buildings of the same line, so upgrading copies away never takes that progress back.

### Strategic advice

- **Older copies are a head start.** Upgrading a Gathering Camp costs half of a new Forager Post, and the upgraded copy counts toward the next age's gate like any other.
- **Upgrade before a food crunch.** A Forager Post makes 1.5 food/tick fully staffed against a Gathering Camp's 1. If a [harbinger](harbinger.md) warns that a catastrophe is coming, upgraded food buildings give you a wider safety margin.
- **You control the order.** You might upgrade your food line as soon as you advance and leave military buildings as they are until you have banked enough resources. Older buildings keep producing in the meantime.
- **A tech-gated building waits for its tech.** Where the new building needs a tech first (see [Buildings a tech opens](#buildings-a-tech-opens)), `upgrade` refuses and names the tech.

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

From the Modern Age on, every age's first-copy prices rise by **1.4 times** the previous age's. Modern prices are unchanged; Information is 1.4 times its authored price, Digital is 1.96 times, and the same curve continues through the Transcendent Age. Outputs do not change. Storage, wonders and upgrades read the resulting prices through their own rules.

In every age, a new copy of a building costs **1.29 times** what it was first written at. That holds for production, military, trade, housing, knowledge and storage buildings alike; wonders and monuments keep their prices. The stores were sized again to hold the dearer first copies, and an [upgrade](#older-buildings) costs half of a new copy, so upgrading an older building is the cheaper road. Where the factor would put one part of a first-copy price (the Coal Plant's coal, the Atomic Vault's stone, the Orbital Refinery's plasma) past what the buildings your age's gate asked for make of it in one target of the age, that part is held to that amount, so a town arriving with empty stores can always make its way to the building.

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

Most buildings (production, military, trade, diplomacy and monuments) cost **15% more** per copy than the last. Knowledge buildings cost **50% more**. Storage and housing buildings grow more gently, at **13%** per copy. Wonders have a flat price and are limited to one each. The scaling lineages still grow exponentially, so plan your production before bulk-buying.

### Build-cost reductions

A handful of **milestone rewards** (Master Builder, Grand Architect and others) each cut build costs by 3 to 5%, and two **techs** (Civil Engineering and Nanofabrication) by 3% each. The milestones' reductions add up (−19% with all of them) and multiply the scaled cost above; each tech's cut then multiplies what is left (×0.97 each). The cost never drops below 10% of what it would be without them. Stacked, the available reductions reach **−24%**. The cost `build` lists is the cost you are charged. You don't need to opt in; earning the milestone or completing the tech is enough.

**Example:** a Gathering Camp's first copy costs 20 wood and each copy costs 15% more. With none built or queued:
- 1st: 20 wood
- 2nd: 23 wood
- 5th: 36 wood
- Building 5 at once: 20+23+27+31+36 = **137 wood total** (not 20×5=100)

### Build times

Build times are capped by the age's pace: ordinary buildings take at most **1/6 of their age's target time**, wonders take at most **1/3**, and storage buildings take at most **1/48** of it (you can queue several copies at once, up to the cap). In practice that is 2m 30s in the Primitive Age, 2h 36m in the Renaissance and 10h 24m from the Interstellar Age on (storage: 18 seconds, 19m 30s and 1h 18m). The full per-age list is in the table under [How Production Rates Are Set](#how-production-rates-are-set).

These are first-run times. On known ground (an age a past run completed) every build time is divided by the age's [Era Mastery](prestige.md#era-mastery) speed, rounded up and never below one tick, and production and storage are multiplied by it. The time shown when you start a build is the real, shortened one.

### Buildings a tech opens

Most buildings unlock the moment you enter their age. Forty wait for a tech from that same age instead, so the age has something new partway through:

| Building | Age | Opened by |
|---|---|---|
| Standing Stones | Stone | Ritual |
| Altar | Bronze | Calendar, the age's keystone |
| Barracks | Bronze | Military Tactics |
| Legion Fort | Iron | Siege Warfare |
| Smelter | Iron | Iron Smelting |
| Forge | Classical | Metal Casting |
| Cathedral | Medieval | Theology, the age's keystone |
| Foundry | Renaissance | Blast Furnace |
| Colonial Steelworks | Colonial | Coke Smelting |
| Harbor | Colonial | Mercantilism |
| Embassy | Colonial | Embassies |
| Coal Plant | Industrial | Steam Power |
| Grand Embassy | Industrial | Concert of Nations, a capstone |
| Geographic Society | Industrial | Geographic Societies |
| Steam Works | Victorian | Electrification |
| Dynamo Hall | Electric | Power Distribution, the age's keystone |
| Nuclear Plant | Atomic | Civilian Reactors |
| Power Grid Hub | Modern | Advanced Electrics |
| Seaport | Modern | Containerization |
| Monument of Ages | Modern | Television |
| Smart Farm | Information | Internet of Things |
| Smart Complex | Information | Internet of Things |
| Cyber Command | Information | Cybersecurity |
| Microgrid Array | Information | Smart Grid |
| Container Terminal | Information | E-commerce |
| Quantum Battery Array | Digital | Grid Storage |
| Logistics Hub | Digital | Automated Logistics |
| Combat Aug Center | Cyberpunk | Cybernetics, the age's keystone |
| Holographic Theater | Cyberpunk | Holography |
| Dark Energy Tap | Cyberpunk | Dark Energy |
| Energy Exchange | Fusion | Maglev Transit |
| Stellar Core Drill | Interstellar | Stellar Core Mining |
| Pulsar Tap | Interstellar | Stellar Engineering |
| Neutron Star Mine | Galactic | Neutron Mining |
| Quasar Tap | Galactic | Antimatter Synthesis |
| Quantum Metal Works | Quantum | Quantum Metallurgy |
| Zero Point Generator | Quantum | Zero-Point Energy |
| Singularity Engine | Transcendent | Singularity Engineering |
| Omniversal War Council | Transcendent | Omniversal Command |
| Omniversal Bazaar | Transcendent | Omniversal Exchange |

Until the tech is done the building is missing from the build list, and `build` and `upgrade` say which tech it needs. Each of these buildings is its line's building for that age, so the older buildings of its line list it as their upgrade when you advance, but the upgrade itself waits for the tech. The build plan takes it early and waits. Copies you already have keep working. See [Technologies](technologies.md#tech-tree-by-age).

A tech only takes a building when the age keeps another way to make the same thing (the Mill beside the Foundry, the Port beside the Harbor), or when the tech is one a run researches anyway. A building that is its age's only source of what it makes is open from the first tick of its age: every age's knowledge and culture building and most of its military ones, the Steam Mine, the Uranium Mine, the Financial District, the Corporate HQ, the Petroleum Refinery, the Nano Foundry, the titanium smelters and mines and the data buildings from the Cyber Hub on among them. The Transcendent Age is the exception: nothing comes after it, so nothing can be stuck behind a tech, and its Singularity Engine, Omniversal Bazaar and Omniversal War Council each wait for one of its three last techs.

Wonders wait for a tech too, in their own way: from the Stone Age on each one needs its age's **keystone** before `build` will start it, but it is listed from the first tick of its age and its bank takes deposits and overflow the whole time. See [The keystone tech](wonders.md#the-keystone-tech).

---

## Lineage System

Each lineage starts at a specific age and gets a new tier in most ages after that. A lineage that makes the same resource all the way is one line. A lineage whose resource changes partway is split into one line for each resource, and a building only ever upgrades along its own line. When your civilization enters a new age:

1. Each line has one **current building** for your age: its newest building that the age has. Buildings of the line from earlier ages are **older buildings**.
2. Older buildings keep producing at their old rate but cannot be built again. The last building of a line has no newer building and is never older.
3. The current building is available to build (fresh copies) and to upgrade into (from older copies of its own line).

You keep all your progress and production. A Gathering Camp in the Food lineage upgrades to a Forager Post, then a Farm, then Field Works; use `upgrade <building>` after an advance to convert the copies you hold. See [Building Upgrades](#building-upgrades) for cost details and strategy.

---

## The 14 Production Lineages

| # | Lineage | Domain | Primary Output | Tiers | First Building | Final Building |
|---|---------|--------|----------------|-------|----------------|----------------|
| 1 | Housing | None | housing | 22 | Hut (+10 housing) | Transcendent Nexus (+15.7M housing) |
| 2 | Food | food | food | 21 | Gathering Camp | Quantum Cultivator |
| 3 | Organic Extraction | lumber | wood → coal → oil → nanobots → quantum flux | 21 | Wood Camp | Reality Harvester |
| 4 | Geological Extraction | masonry | stone → iron ore → uranium → titanium ore → dark matter crystals → antimatter | 21 | Stone Camp | Reality Excavator |
| 5 | Knowledge | none | research | 21 | Story Circle | Reality Academy |
| 6 | Faith | faith | faith | 21 | Shrine | Transcendence Hall |
| 7 | Military | military | soldiers | 22 | War Camp | Omniversal War Council |
| 8 | Trade | trade | gold | 20 | Market | Omniversal Bazaar |
| 9 | Engineering | engineering | iron → steel → electricity → plasma → dark matter → quantum flux | 20 | Smithy | Singularity Engine |
| 10 | Culture/Arts | None | culture | 17 | Amphitheater | Reality Art Engine |
| 11 | Metallurgy | metallurgy | iron → steel → titanium → dark matter → antimatter → quantum flux | 18 | Smelter | Quantum Metal Works |
| 12 | Energy | energy | coal → electricity → plasma → dark matter → quantum flux | 13 | Coal Plant | Zero Point Generator |
| 13 | Hacker/Digital | hacker | data | 9 | Mainframe Hall | Reality Processor |
| 14 | Harbor | trade | gold (+trade-route income) | 5 | Harbor | Logistics Hub |

> Lineages 8 to 14 start in later ages (Trade and Engineering in the Bronze Age, Culture/Arts in the Classical, Metallurgy in the Iron, Energy in the Industrial, Hacker/Digital in the Modern and Harbor in the Colonial), so they have fewer tiers. The Housing, Culture/Arts and Knowledge lineages have no worker domain: they work without workers (the Knowledge lineage comes with scholars). The **Harbor** lineage also raises trade income: besides producing gold, each harbor adds a percentage bonus to the income of *every* active trade route. See [Trade](trade.md#harbor-lineage-trade-route-income).

The **Knowledge** lineage is the one that comes with scholars. Each copy comes with its scholars, from 2 for a Story Circle to 20 for a Reality Academy, who take no workers, and each copy costs 50% more than the one before it. A knowledge building cannot be built past its cap, the count its age's gate asks for, 5 to 10 copies. [Scholars and Research](knowledge.md#what-a-scholar-is) has each building's scholars, work and cap.

The **Mainframe Hall** (Modern Age) is the first tier of the Hacker/Digital lineage and the first building that makes data. It uses `hacker` workers, the Modern Age's gate asks for eight of it, and its next tier is the Server Farm in the Information Age.

---

## The 39 Building Lines

A **line** is a run of buildings in one lineage that make the same resource. A line never switches resource: when a lineage moves on to a new resource, the new buildings start a new line, and an old building only ever upgrades to the next building of its own line. Eight lineages never change what they make, so each is one line: Housing, Food, Knowledge, Faith, Military, Culture/Arts, Hacker/Digital and Harbor. The other six are split into 31 lines, 39 in all.

A line can pause and pick up again. Energy's electricity line stops for the Modern Age, when the Oil Refinery makes oil, and picks up in the Information Age; Metallurgy's titanium line skips the Cyberpunk and Fusion Ages, when the Dark Matter Refinery and the Exotic Matter Forge make dark matter, and Trade's gold line skips the Cyberpunk Age, when the Black Market Hub makes crypto. A building whose line comes back is offered the next building of its line when its age arrives.

| Lineage | Makes | Ages | Buildings | First | Last |
|---------|-------|------|-----------|-------|------|
| Organic Extraction | wood | Primitive to Medieval | 6 | Wood Camp | Sawmill |
| Organic Extraction | coal | Renaissance to Industrial | 3 | Coal Mine | Steam Colliery |
| Organic Extraction | oil | Victorian to Information | 5 | Oil Derrick | Smart Refinery |
| Organic Extraction | nanobots | Digital to Fusion | 3 | Bio Fabrication Lab | Molecular Synthesizer |
| Organic Extraction | quantum flux | Space to Quantum | 4 | Quantum Organic Extractor | Reality Harvester |
| Geological Extraction | stone | Stone to Classical | 5 | Stone Camp | Marble Works |
| Geological Extraction | iron ore | Medieval to Industrial | 4 | Stonemason's Guild | Steam Mine |
| Geological Extraction | uranium | Victorian to Atomic | 3 | Uranium Mine | Uranium Processing Works |
| Geological Extraction | titanium ore | Modern to Digital | 3 | Titanium Mine | Nano Drill Complex |
| Geological Extraction | dark matter crystals | Cyberpunk to Space | 3 | Dark Crystal Mine | Asteroid Crystal Mine |
| Geological Extraction | antimatter | Interstellar to Quantum | 3 | Stellar Core Drill | Reality Excavator |
| Trade | gold | Bronze to Digital and Fusion to Transcendent | 19 | Market | Omniversal Bazaar |
| Trade | crypto | Cyberpunk | 1 | Black Market Hub | Black Market Hub |
| Engineering | iron | Bronze to Medieval | 4 | Smithy | Workshop |
| Engineering | steel | Renaissance to Victorian | 4 | Mill | Steam Works |
| Engineering | electricity | Electric to Digital | 5 | Power Station | Neural Grid |
| Engineering | plasma | Cyberpunk to Space | 3 | Augmentation Foundry | Launch Complex |
| Engineering | dark matter | Interstellar to Galactic | 2 | Warp Drive Plant | Dyson Assembly |
| Engineering | quantum flux | Quantum to Transcendent | 2 | Reality Forge | Singularity Engine |
| Metallurgy | iron | Iron to Medieval | 3 | Smelter | Ironmonger |
| Metallurgy | steel | Renaissance to Atomic | 6 | Foundry | Advanced Alloy Plant |
| Metallurgy | titanium | Modern to Digital and Space | 4 | Titanium Smelter | Orbital Refinery |
| Metallurgy | dark matter | Cyberpunk to Fusion | 2 | Dark Matter Refinery | Exotic Matter Forge |
| Metallurgy | antimatter | Interstellar to Galactic | 2 | Antimatter Forge | Stellar Metallurgy |
| Metallurgy | quantum flux | Quantum | 1 | Quantum Metal Works | Quantum Metal Works |
| Energy | coal | Industrial | 1 | Coal Plant | Coal Plant |
| Energy | electricity | Victorian to Atomic and Information to Cyberpunk | 6 | Steam Turbine | Dark Energy Tap |
| Energy | oil | Modern | 1 | Oil Refinery | Oil Refinery |
| Energy | plasma | Fusion to Interstellar | 3 | Tokamak Array | Pulsar Tap |
| Energy | dark matter | Galactic | 1 | Quasar Tap | Quasar Tap |
| Energy | quantum flux | Quantum | 1 | Zero Point Generator | Zero Point Generator |

---

## What Buildings Do

A building can have one or more of these effects. Knowing them helps you decide what to build.

| Effect | What It Does |
|--------|-------------|
| Production | Adds an amount of a resource each tick, scaled by how many workers it has (see below). Most production buildings do this. |
| Storage | Raises the storage of one resource, or storage for every resource (storage buildings). A knowledge building comes with scholars. |
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

For **construction resources** (anything the buildings of an age cost: wood, stone, iron, gold, steel, coal, electricity, data and so on), rates follow the **Payback Rule**: a producer's output is set so that, fully staffed, it earns back the price of its first copy in its age's **payback time**. The payback time is a share of how long the age is meant to take, and that share grows through the game: about 1/16 of the age in the Primitive Age, about 1/9 in the Iron Age, a sixth in the Renaissance (stretched to a third there; see below) and a fifth in the Industrial Age. From the Victorian Age on it grows faster: over a third of the age there, nearly a half in the Modern Age and seven tenths in the Interstellar. The table lists it before production bonuses, which multiply a late age's output six or seven times over. Every new copy costs **1.29 times** its authored price and outputs did not move, so each payback is 1.29 times what the shares above give; the table has the 1.29 in it. Later ages repay more slowly because every building you put up in earlier ages keeps producing alongside the new tier.

Output is valued at **price parity**. Each age has a price level for each resource (the typical first-copy price in that resource among the age's buildings), and resources are worth each other in the ratio of those levels. A building with two outputs splits its value between them, in proportion to what the age's gate asks of each. Because rates follow prices, they grow roughly 5 to 8x per age.

**Examples:**
- Wood Camp (Primitive): costs 20 wood and makes 0.569 wood/tick fully staffed, so it repays itself in 36 ticks (73 seconds).
- Stone Pit (Stone): costs 232 stone and 387 wood. At Stone Age parity (310 stone = 464 wood) that is worth 490 stone, and at 3.48 stone/tick it repays in about 4 minutes 40 seconds.

**Flow resources keep hand-set rates:** food, faith, culture and soldiers. They feed workers, set morale and epoch odds, fill culture storage or make up your army. **Nanobots** keep theirs too: they are a build material, asked for in the hundreds of thousands where an age's other prices are in the trillions, so a Nanobot Vat makes the 6.55K a tick it is listed at and a building that asks for nanobots is not valued by them. Research keeps fixed rates too: a knowledge building's work is a hand-set figure (see [Scholars and Research](knowledge.md#what-a-scholar-is)), and nothing is priced in it.

| Age | Target time | Payback (fully staffed) | Build-time cap | Storage build cap |
|-----|-------------|-------------------------|----------------|-------------------|
| Primitive | 15m | 1m 12s | 2m 30s | 18s |
| Stone | 45m | 5m | 7m 30s | 56s |
| Bronze | 3h 54m | 33m | 39m | 4m 52s |
| Iron | 6h 30m | 59m | 1h 5m | 8m 6s |
| Classical | 9h 6m | 1.6h | 1h 31m | 11m 22s |
| Medieval | 11h 42m | 2.3h | 1h 57m | 14m 36s |
| Renaissance | 15h 36m | 6.8h | 2h 36m | 19m 30s |
| Colonial | 18h 12m | 4.3h | 3h 2m | 22m 44s |
| Industrial | 20h 48m | 6.2h | 3h 28m | 26m |
| Victorian | 23h 24m | 10.7h | 3h 54m | 29m 14s |
| Electric | 26h | 14.4h | 4h 20m | 32m 30s |
| Atomic | 31h 12m | 18.2h | 5h 12m | 39m |
| Modern | 31h 12m | 18.6h | 5h 12m | 39m |
| Information | 36h 24m | 10.9h | 6h 4m | 45m 30s |
| Digital | 41h 36m | 29.3h | 6h 56m | 52m |
| Cyberpunk | 46h 48m | 36.0h | 7h 48m | 58m 30s |
| Fusion | 52h | 92.7h | 8h 40m | 1h 5m |
| Space | 57h 12m | 69.2h | 9h 32m | 1h 11m 30s |
| Interstellar | 62h 24m | 57.3h | 10h 24m | 1h 18m |
| Galactic | 62h 24m | 121.4h | 10h 24m | 1h 18m |
| Quantum | 62h 24m | 36.8h | 10h 24m | 1h 18m |
| Transcendent | 62h 24m | 33.8h | 10h 24m | 1h 18m |

The Renaissance's payback is 2x what the curve gives (about 3.4 hours): it is where gold income jumps, and at the curve's rate the age would run at barely half its target length. It was 1.7x until the Stone and Iron Eras gained their techs: a run now arrives with cheaper, quicker building, and the age had dropped to three quarters of its target. Its Exchange makes 1.38K gold, Mill 191 steel, Foundry 261 steel and Coal Mine 106 coal.

**From the Victorian Age on the curve is steeper.** The finished tech tree took research off the critical path of every age from the Victorian on: a wonder's keystone takes a small part of what the age's scholars do. What paces those ages is what their buildings make and what their requirements ask, and on the old curve a full run played the ages from the Modern to the Galactic at 0.38 to 0.56 of their targets (the Information Age aside, at 0.84). The Victorian, Electric and Atomic Ages had each been set by hand, at 1.7, 1.45 and 1.75 times the old curve; the steeper curve gives them 1.62, 1.67 and 1.71 and carries the Digital, Cyberpunk and Interstellar Ages as well. A Bessemer Plant makes 148K steel/tick (was 141K), an Electric Arc Furnace 690K (was 874K) and a Breeder Reactor 1.29M electricity (was 1.26M). The Bronze Age's payback is 1.1x (a Lumber Mill makes 7.91 wood/tick; a Quarry 4 stone): the age ran at under two thirds of its target.

Some ages sit off the curve, because the ages are not smooth: how quick one is depends mostly on how far its prices jump from the age before, which is what you arrive able to pay, and on how much of what it builds with the age before goes on making. An age answers to its own payback with about half the change, and passes about a third of it on to the next.

| Age | Against the curve | Why |
|-----|-------------------|-----|
| Industrial | 1.15x | A run that arrives with a large store plays it in under two thirds of its target. |
| Electric | 1.1x | It ran a little under two thirds of its target on the curve alone. |
| Modern | 1.1x | It ran a little under seven tenths of its target on the curve alone. |
| Information | 0.45x | The one late age that ran near its target. You buy your way into data and steel before you make anything, and you arrive from a Modern Age that repays more slowly than it did. It keeps the payback it had. |
| Fusion | 2.2x | It answers little to its own payback, and the Cyberpunk Age's producers go on making the electricity it builds with. |
| Space | 1.4x | |
| Galactic | 2.0x | The quickest age on the old curve, with the largest bonuses in the game by then. |
| Quantum, Transcendent | 0.5x | Where they were. No measured run plays them yet. |

Times are at the base tick of 2 seconds; game speed bonuses shorten them in real time. Every construction resource an age charges for has a maker in that age: see [The Maker Rule](#the-maker-rule).

---

## The Maker Rule

Every material an age charges for, in a building its gate asks for or in its wonder, has a maker in that age: a building of that age that makes it. Every maker is set on one standard, so a maker repays the price of its first copy in the same number of ticks as the age's other makers (the Payback Rule above). The rule covers building materials. It leaves out nanobots and the flow resources, food, faith, culture and soldiers.

One shortage is left as it is. The Electric Age has no maker of coal, and the older buildings you keep cover it. A building is never priced in a material nothing in its age makes any more: the Colonial Age's Settlement Block, Harbor and Colonial Warehouse are priced in stone, not wood. One exception is listed: the last age's wonder is paid from the Quantum Age's makers, because the Transcendent Age has none of its own.

There is no market that swaps one material for another, so a town that wants a material builds a maker of it. The output tables below give a building's second material after the word with.

Four raw materials are made and spent like wood or stone, in the ages that make them. In such an age every building and the wonder that cost the raw's worked form pay half of that part in the raw instead, one for one: marble with stone (Iron and Classical Age), iron ore with iron (Iron to Industrial Age), titanium ore with steel (Modern to Digital Age) and dark matter crystals with electricity (Cyberpunk to Space Age). The pair share one price level, so no price is larger or smaller overall. Housing, storage, the raws' own makers, the Mainframe Hall and the Black Market Hub pay no raw. See [What the ores are for](resources.md#what-the-ores-are-for).

---

## Lineage Output Progression

Several lineages change which resource they make as you move through the ages. The new resource starts a new line (see [The 39 Building Lines](#the-39-building-lines)): its first building is built fresh, the old line's buildings keep making the old resource, and nothing upgrades across from one to the other. A Sawmill never turns into a Coal Mine.

**Some tiers arrive before their resource does.** A resource gathers nothing until the age that unlocks it, whatever makes it. These buildings say `once unlocked` beside that output in their description:

| Building (age) | Output | Unlocks in |
|---|---|---|
| Colosseum (Iron) | culture | Classical Age |
| Uranium Mine (Victorian), Nuclear Extraction Plant (Electric) | uranium | Atomic Age |
| Titanium Smelter (Modern), Aerospace Foundry (Information), Nano Alloy Plant (Digital) | titanium | Space Age |
| Augmentation Foundry (Cyberpunk) | plasma (its electricity counts at once) | Fusion Age |
| Dark Matter Refinery (Cyberpunk), Exotic Matter Forge (Fusion) | dark matter | Interstellar Age |
| Quantum Organic Extractor (Space), Reality Matter Weaver (Interstellar), Cosmic Organic Works (Galactic) | quantum flux | Quantum Age |
| Stellar Core Drill, Antimatter Forge (Interstellar) | antimatter | Galactic Age |

Until then they count toward age requirements and milestones and make none of that resource.

### Organic Extraction (Lineage 3)

| Ages | Output Resource |
|------|----------------|
| Primitive to Medieval | wood |
| Renaissance to Industrial | coal |
| Victorian to Information | oil |
| Digital to Fusion | nanobots |
| Space to Quantum | quantum flux |

**Nano Foundry** (Modern Age) is a standalone nanobot producer outside the lineage chain. It gives nanobots a producer as soon as they unlock, two ages before the lineage's own Bio Fabrication Lab. It uses `engineering` workers and makes +80 nanobots/tick. The lineage's own makers are listed at 3.28K (Bio Fabrication Lab), 6.55K (Nanobot Vat) and 13.1K (Molecular Synthesizer) a tick and make exactly that: nanobots are a build material with hand-set rates.

### Geological Extraction (Lineage 4)

| Ages | Output Resource |
|------|----------------|
| Stone to Bronze | stone |
| Iron to Classical | marble and iron ore, with stone and iron |
| Medieval to Industrial | iron ore, with stone and iron |
| Victorian | uranium, with stone |
| Electric | uranium |
| Atomic | uranium, with stone |
| Modern to Digital | titanium ore |
| Cyberpunk to Space | dark matter crystals |
| Interstellar to Quantum | antimatter |

The quarries and mines from the Iron Age to the Industrial make stone and iron themselves, and the Uranium Mine and the Uranium Processing Works make stone. The titanium mines and the crystal mines make only their raw. Every raw is made at the standard rate and spent by the buildings of the ages that make it.

### Engineering (Lineage 9)

| Ages | Output Resource |
|------|----------------|
| Bronze to Medieval | iron |
| Renaissance to Victorian | steel |
| Electric to Digital | electricity |
| Cyberpunk to Space | plasma (with electricity in the Cyberpunk and Fusion Ages) |
| Interstellar | dark matter, with plasma |
| Galactic | dark matter |
| Quantum | quantum flux, with dark matter |
| Transcendent | quantum flux |

### Metallurgy (Lineage 11)

Metallurgy buildings produce metal directly; they do not use up ore.

| Ages | Output Resource |
|------|----------------|
| Iron to Medieval | iron |
| Renaissance to Industrial | steel |
| Victorian to Atomic | steel, with iron |
| Modern to Digital | steel, with titanium |
| Cyberpunk to Fusion | steel, with dark matter |
| Space | titanium, with steel |
| Interstellar to Galactic | antimatter, with titanium |
| Quantum | quantum flux |

---

## Older Buildings

A building is older when its line has a newer building for the age you are in (see [Building Upgrades](#building-upgrades)). Older buildings:

- Keep producing at their normal rate (you lose nothing when you advance)
- Cannot be built again; `build` names the building they upgrade to
- Can be upgraded, at half the price of a new copy of that building
- Count toward building totals in your civilization stats
- Can still have workers assigned and unassigned, and can be sold as normal

Your 20 Gathering Camps stay Gathering Camps, producing at their full rate, until you run `upgrade gathering_camp`. Production only improves once the upgrade is done, which is the reason to do it. See [Building Upgrades](#building-upgrades) for the full guide.

---

## Storage Buildings (21 tiers)

Storage buildings form their own lineage and add storage for **every** resource at once. They need no workers: build them and your storage goes up.

Storage buildings have **no copy limit**. Each copy costs **1.75 times** the one before it, so a tall stack gets dear fast. Their sizes follow one written rule: five copies of an age's storage building, added to the stores of the earlier ages, hold the dearest single price that age asks of a full town (eight copies of every building that makes something), with a quarter to spare. Every age gate asks for **five copies** of the storage building of the age you are leaving (see [The 22 Ages](ages.md#advancement-requirements)).

| Building | Age | Storage per copy (every resource) |
|----------|-----|--------|
| Stash | Primitive | +130 |
| Storage Pit | Stone | +1.5K |
| Warehouse | Bronze | +8.8K |
| Granary | Iron | +32K |
| Classical Vault | Classical | +270K |
| Strongroom | Medieval | +1.9M |
| Renaissance Vault | Renaissance | +7.8M |
| Colonial Warehouse | Colonial | +37M |
| Industrial Depot | Industrial | +110M |
| Victorian Vault | Victorian | +1.8B |
| Electric Warehouse | Electric | +20B |
| Atomic Vault | Atomic | +110B |
| Modern Depot | Modern | +430B |
| Info Vault | Information | +3T |
| Digital Archive | Digital | +9.4T |
| Cyber Vault | Cyberpunk | +110T |
| Fusion Vault | Fusion | +470T |
| Orbital Depot | Space | +2.4Q |
| Stellar Vault | Interstellar | +85Q |
| Galactic Vault | Galactic | +710Q |
| Quantum Vault | Quantum | +62Qi |

Pottery and Masonry add 10% each to every store, and Cloud Computing and Superconductors 8% each later.

### Storage is a wall

A building or an upgrade whose price in any one resource is larger than the store for that resource can hold **cannot be bought**. That holds for `build`, `build <n>`, `build max`, `upgrade` and the [build plan](plan.md). The game says so and names the storage building to build:

```
<Building> costs <amount> <resource>, more than your <resource> storage holds (<size>). Build more storage first: <Storage building>.
```

The age's wonder is the one exception: it is paid through its bank, a round at a time (see [Wonders](wonders.md#overflow)), so it may cost more than a store holds.

A storage building's next copy always fits: owning one raises your store to at least the price of the next.

The gate out of each age asks for five copies of its storage building, and all five can be bought.

A full store stops that resource piling up. What it cannot hold goes into the current wonder's bank while the wonder still needs it (unless you typed `wonder overflow off`), and is lost otherwise. It no longer pays toward the plan's next copies. So build storage first whenever you enter a new age.

Storage buildings **never upgrade** and are never offered as upgrades when you advance. Storage adds up: every storage building you have built keeps adding its capacity for the rest of the run, so the stashes from the Primitive Age still count in the Quantum Age. Like every other building, an older age's storage can no longer be built once you advance (the game tells you to build the current age's storage instead), so fill each tier while it is current.

Storage is also permanent: catastrophes never destroy it or turn it into ruins, and it can't be sold.

---

## Cultural Monuments (4)

Cultural Monuments are one-off structures (one copy each) that turn surplus **culture** into a lasting payoff. Each costs a large lump of culture plus other materials of its age, and gives a bonus to all production while it stands (an Endure can destroy it like any building but wonders, storage and knowledge buildings). Unlike wonders, they need no resource banking: build them with the normal `build <key>` command. Their bonuses add into the same pool as every other all-production bonus but a tech's. The pool applies in full up to +200% and a quarter of every point past it, so a monument built late still raises your output, by a quarter of what it lists. See [The all-production cap](resources.md#the-all-production-cap).

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

You can build as many as you like, and the pace scales with how many you've built and how fully you've staffed them. Game speed bonuses shorten the wait in real time; the **Factions** panel shows the countdown to the next dispatch. The Society sends the affordable short or long offer with the best haul per minute, never a grand offer. It uses that tier's slot only when it is free, and pays the offer's full resource cost. If your stores can't outfit one, it waits and sends it as soon as they can.

It keeps missions moving while you are away. Full details in [Army & Missions](military.md#automatic-dispatch-the-geographic-society).

---

## Wonders (22)

Wonders are unique buildings: you can build each wonder only once. You bank resources toward a wonder before you can build it:

```
wonder collect <resource|all> [amount|all|max]   bank resources toward the current wonder
build <wonder_key>                              start construction once it is fully funded
```

Full stores bank into the wonder on their own ([overflow](wonders.md#overflow)), and a wonder in your [build plan](plan.md) banks and starts by itself once what you hold covers the rest.

You must build each age's wonder before you can advance to the next age. A wonder's price adds up to 64 **price units** of its age, where one price unit is what a typical building of that age charges in a single resource (valued at [price parity](#how-production-rates-are-set)); each wonder keeps its own mix of resources. A wonder takes at most 1/3 of the age's target time to build.

Every wonder adds a little production of one or two resources, and many add more: a percentage to all production (the Crystal Palace, Hoover Dam and the last four wonders), to research or to expedition rewards, or housing or storage. All-production bonuses add into [the all-production pool](resources.md#the-all-production-cap), which applies in full up to +200% and a quarter of every point past it.

Endure and the Great Fire never destroy wonders, and Succumb never turns them into ruins. They stay for the rest of the run, like storage.

See [Wonders](wonders.md) for the full list with costs and effects.

---

## Ruins

Ruins are what is left of a previous civilization. You get them when you choose **Succumb** during a catastrophe: up to 8 of your buildings (never wonders, storage or knowledge buildings) become ruins and carry forward into your new run. You can hold at most 24 ruins; when a new batch goes over, the lowest-value ruins (earliest age, then lowest output) crumble first.

**How ruins differ from normal buildings:**

| Property | Normal Building | Ruin |
|----------|----------------|------|
| Production | `base × count × (0.20 + 0.80 × fill)` | `base × count × 0.50` |
| Workers | Needed for full output | Not needed (produces on its own) |
| Can build more? | Yes | No |
| Destroyed by Endure? | Yes (20% of buildings at random, fewer if you brace; never wonders, storage or knowledge buildings) | No |

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
