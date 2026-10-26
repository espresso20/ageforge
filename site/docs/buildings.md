# Buildings

AgeForge has 302 buildings. 251 lineage buildings sit in 14 production lineages across the 22 ages, and each lineage is made of lines: a line makes one resource from its first building to its last. The 14 lineages hold 39 building lines. The other 51 buildings are 21 storage buildings, 22 wonders, 4 cultural monuments and 4 standalone buildings (the Embassy, the Grand Embassy, the Geographic Society and the Nano Foundry).

A building from an earlier age is an older building. It keeps producing, and you can upgrade it to its line's building for your current age at half the price of a new copy (see [Building upgrades](#building-upgrades)). For the full wonder list see [Wonders](wonders.md).

---

## Why buildings matter

Buildings make most of your resources, and the workers in them turn that capacity into output. More buildings give you:

- More production. Each copy adds its rate to your total.
- More worker slots. Slots grow with the number of copies, which raises how much your workers can make.
- More output from everything. Each storage building adds 1% to all production (see [Storage buildings](#storage-buildings-21-tiers)).
- Progress toward the next age. Each age gate asks for counts of certain buildings and for the age's wonder.

---

## Building commands

```
build                      list your unlocked buildings, with the next copy's cost and the count built
build <building>           queue one copy
build <building> <count>   queue up to that many copies
build <building> max       queue as many as you can pay for
```

`b` is short for `build`, and `<building>` is the key the Buildings panel shows next to the name (`gathering_camp`). The price is paid when you queue a copy. To build copies as the resources come in, add them to the [build plan](plan.md) with `plan build <building> [count]`.

You can build only the current age's buildings, and a building that waits for a tech needs that tech first (see [Buildings a tech opens](#buildings-a-tech-opens)). Wonders and cultural monuments allow one copy each. Every other building has no limit. Queued copies are built side by side, each taking the building's own build time (see [Build times](#build-times)).

<figure class="screen" data-screen="buildings"><figcaption>The Buildings panel in the Bronze Age, scrolled to the current age: each building with the key that build takes, what its next copy costs and how its worker slots are filled.</figcaption></figure>

With a count or `max`, each copy is priced at its own step of the cost curve and the command stops at the first copy you cannot pay for. `max` does not divide your stock by the price of the first copy. Wonders are the exception: they are paid through a bank (see [Wonders](#wonders-22)).

## Building upgrades

A building from an earlier age is an older building. It keeps producing what it always made. You can upgrade it to its line's building for your current age, at half the price of a new copy, and it then counts as that building.

Whether a building is older is worked out from its line and your age, and nothing is stored, so it follows you through every advance. A Hut in the Iron Age is an older building, and it upgrades straight to the Iron Age's house, not one step at a time. A building only upgrades along its own line, to a building that makes the same resource (see [The 39 building lines](#the-39-building-lines)). A line with no newer building for your age keeps its last building as its current one. So the last building of a line, such as the Sawmill, is never older, and a line that pauses for an age or two has nothing to upgrade to until it resumes.

Some buildings are in no line and are never older: storage (your stashes, storage pits and vaults keep adding their boost to production for the rest of the run, see [Storage buildings](#storage-buildings-21-tiers)), knowledge buildings (each age's knowledge building is built new in its own age, see [Scholars and research](knowledge.md#what-a-scholar-is)), wonders, monuments and the four standalone buildings.

When you advance, the log names the buildings that became older (`↑ 12 Gathering Camps can upgrade to Forager Post. Type 'upgrade gathering_camp'.`). The Buildings panel marks each one `(legacy)`. Leaving them as they are costs nothing: they keep producing at their old rate. You cannot build new copies of an older building. `build` says so and names the building it upgrades to.

### Upgrade commands

```
upgrade                    list the older buildings you hold
upgrade <building>         upgrade as many copies as you can pay for now
upgrade <building> <n>     upgrade up to n copies
upgrade <building> all     upgrade every copy you hold, as far as you can pay
```

Bare `upgrade` prints one line per older building you hold: the old and new keys, how many copies you hold, the price of the next copy, how many you can pay for now, and the reason if something blocks the next one (a tech to research). To upgrade copies as you can pay for them, add them to the [build plan](plan.md) with `plan upgrade <building> [count|all]`.

Copies are upgraded one at a time, each at its own price, and the command stops at the first copy it cannot pay for. It says how many it upgraded, what they cost, and what the next copy costs or why it stopped. An upgrade takes effect at once, with no build time. Upgrades are per building, so you control the order.

**Examples:**
```
upgrade gathering_camp
upgrade gathering_camp 3
upgrade gathering_camp all
```

### Upgrade cost

One upgraded copy costs half of what the next new copy of the building it becomes would cost right now, in every resource. The price is worked out the way a build is priced, with the same cost curve and the same build-cost cuts. Copies are priced one at a time: after each upgrade there is one more of the new building, so the next copy costs more. Interchangeable Parts, the Industrial Age's Craft capstone, takes another 15% off an upgrade's price.

The old copy is given up, and nothing is paid back for it. Selling an older building pays back half of its own price, but the copy stops producing, so an older building is better kept or upgraded than sold.

An upgrade is refused when you cannot pay for it or when the new building's tech is not researched yet.

### Workers and upgrades

An upgraded copy keeps its workers. They move from the old building's slots to the new building's, one copy's share at a time. If the new building has fewer slots, the extra workers go back to the idle pool.

### Counting toward the next age

An upgraded copy is the building it became. It counts for the next age's building requirements, for milestones and badges, and for the price of the next copy. A requirement that names an older building also counts the later buildings of the same line, so upgrading copies away never takes that progress back. This holds for a milestone that asks for Wood Camps, a gate that asks for Stone Camps and a trade route that needs Markets.

### Advice

- Older copies are a head start. Upgrading a Gathering Camp costs half of a new Forager Post, and the upgraded copy counts toward the next age's gate like any other.
- A Forager Post makes 1.5 food/tick fully staffed against a Gathering Camp's 1, so upgrading your food buildings is a cheap way to widen your food margin.
- You choose the order. Upgrade your food line as soon as you advance and leave other lines as they are until you have the stock. Older buildings keep producing in the meantime.
- Where the new building needs a tech first (see [Buildings a tech opens](#buildings-a-tech-opens)), `upgrade` refuses and names the tech.

---

## Selling buildings

```
sell <building>           demolish 1 copy, recover 50% of its price
sell <building> <count>   demolish that many copies (most expensive first)
```

Selling removes the most expensive copies first, and each copy refunds 50% of its listed price at its step of the cost curve, before any build-cost cuts. If selling leaves fewer worker slots than workers assigned, the extra workers go back to the idle pool. They are not dismissed, so your population stays the same.

Some sales are refused:

- Nothing can be sold in the Primitive Age.
- Wonders cannot be sold.
- Storage cannot be sold, and neither can knowledge buildings. Once their age has passed, they can never be rebuilt.
- A building with any copy in the build queue cannot be sold until it is built.

---

## Cost scaling

Each building has a first-copy price, and every copy after that costs a fixed percentage more than the one before. The price of the next copy depends on how many are already built plus how many are in the build queue, rounded down in each resource:

```
cost of next copy = floor(first-copy price × scale ^ (built + queued))
```

When you buy several at once (`build <building> <count>` or `build <building> max`), each copy in the batch is priced at its own step and the total is the sum, not a flat multiple:

```
total cost = sum of floor(first-copy price × scale ^ (built + queued + i))
             for i = 0 to count-1
```

So a batch always costs more per copy than the first copy alone, and queued buildings raise the price of your next purchase before they finish. The scale depends on the kind of building:

| Buildings | Each copy costs |
|---|---|
| Production, faith, military, trade and diplomacy buildings | 15% more than the last |
| Housing | 13% more |
| Knowledge buildings | 50% more |
| Storage buildings | 1.75 times the last (see [Storage buildings](#storage-buildings-21-tiers)) |
| Wonders and cultural monuments | A flat price, one copy each |

The scaling lineages still grow exponentially, so plan your production before bulk-buying.

Prices also climb with the age after the Modern Age. Each of those ages has first-copy prices 1.4 times the previous age's: the Information Age is 1.4 times the Modern Age, the Digital Age 1.96 times, and the same step continues through the Transcendent Age. What buildings make does not change with it.

The first-copy prices the game lists for every building but a wonder or a monument already include a factor of 1.43. An upgrade costs half of a new copy, so upgrading an older building is the cheaper road. Four price parts are held down to what the buildings your age's gate asked for can make of them in a set time, so a town arriving with empty stores can still reach the building: the coal of the Agricultural Works and the Coal Plant, the stone of the Atomic Vault and the plasma of the Orbital Refinery.

### Build-cost reductions

Five milestone rewards cut build costs: Early Builder, Forge Master and Seasoned Builder by 3% each, and Master Builder and Grand Architect by 5% each. Together they cut 19%, and the result multiplies the scaled cost above.

Eleven techs cut costs too, by 2 to 4% each: Civil Engineering, Guilds, Architecture, Surveying, Interchangeable Parts, Prefabrication, Nanofabrication, Space Elevator, Hull Printing, Dyson Engineering and Reality Engineering. Each tech's cut multiplies what the others left.

With every milestone and tech, a building costs about 58% of the scaled cost, a cut of 42%. The cost never drops below 10% of what it would be without them. The cost `build` lists is the cost you are charged. You don't need to opt in: earning the milestone or finishing the tech is enough. Wonders are paid at their listed price, so these cuts do not touch them.

**Example:** a Gathering Camp's first copy lists 22 wood and each copy costs 15% more. With none built or queued and no cuts:
- 1st: 22 wood
- 2nd: 26 wood
- 5th: 40 wood
- Building 5 at once: 22+26+30+34+40 = **152 wood total** (not 22×5 = 110)

### Build times

Each building has its own build time, listed in ticks and shown when you start the build. A tick is 2 seconds at the base speed, so a Hut takes 16 seconds. Later buildings take longer, and a wonder takes the longest of its age. A game speed bonus shortens every build time in real time.

Some techs shorten build times, never to less than 40% of the listed time, and some also shorten wonder construction. On an age you have mastered, [Era Mastery](prestige.md#era-mastery) divides every build time by the age's speed, rounded up and never below one tick. The time shown when you start a build is the real, shortened one.

### Buildings a tech opens

Most buildings unlock the moment you enter their age. Forty wait for a tech from that same age instead:

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

Until the tech is done, `build` leaves the building out of its list and refuses it, naming the tech, and `upgrade` does the same. The Buildings panel and the dashboard's Buildings list show it greyed out with the research it needs. Each of these buildings is its line's building for that age, so the older buildings of its line list it as their upgrade when you advance, but the upgrade itself waits for the tech. The build plan takes it early and waits. See [Technologies](technologies.md#tech-tree-by-age).

Every other building is open from the start of its age. Wonders wait for a tech in their own way. From the Stone Age on, each one needs its age's keystone before `build` will start it, but it is listed from the start of its age and its bank takes deposits the whole time. See [The keystone tech](wonders.md#the-keystone-tech).

---

## Lineage system

Each lineage starts at a specific age and gets a new building in most ages after that. A lineage that makes the same resource all the way is one line. A lineage whose resource changes partway is split into one line for each resource, and a building only ever upgrades along its own line. When you enter a new age:

1. Each line has one current building for your age: its newest building that the age has. Buildings of the line from earlier ages are older buildings.
2. Older buildings keep producing at their old rate but cannot be built again. The last building of a line has no newer building and is never older.
3. The current building is available to build (fresh copies) and to upgrade into (from older copies of its own line).

You keep all your progress and production. The Food lineage runs Gathering Camp, Forager Post, Farm, Field Works and so on, and a Gathering Camp upgrades to whichever of them is current for your age. Use `upgrade <building>` after an advance to convert the copies you hold. See [Building upgrades](#building-upgrades) for cost details and advice.

---

## The 14 Production Lineages

| # | Lineage | Domain | Primary output | Tiers | First building | Final building |
|---|---------|--------|----------------|-------|----------------|----------------|
| 1 | Housing | None | housing | 22 | Hut (+10 housing) | Transcendent Nexus (+15.7M housing) |
| 2 | Food | food | food | 21 | Gathering Camp | Quantum Cultivator |
| 3 | Organic Extraction | lumber | wood → coal → oil → nanobots → quantum flux | 21 | Wood Camp | Reality Harvester |
| 4 | Geological Extraction | masonry | stone → iron ore → uranium → titanium ore → dark matter crystals → antimatter | 21 | Stone Camp | Reality Excavator |
| 5 | Knowledge | None | research | 21 | Story Circle | Reality Academy |
| 6 | Faith | faith | faith | 21 | Shrine | Transcendence Hall |
| 7 | Military | None | its soldiers, no resource | 22 | War Camp | Omniversal War Council |
| 8 | Trade | trade | gold (crypto in the Cyberpunk Age) | 20 | Market | Omniversal Bazaar |
| 9 | Engineering | engineering | iron → steel → electricity → plasma → dark matter → quantum flux | 20 | Smithy | Singularity Engine |
| 10 | Culture/Arts | None | culture | 17 | Amphitheater | Reality Art Engine |
| 11 | Metallurgy | metallurgy | iron → steel → titanium → dark matter → titanium → antimatter → quantum flux | 18 | Smelter | Quantum Metal Works |
| 12 | Energy | energy | coal → electricity → oil → electricity → plasma → dark matter → quantum flux | 13 | Coal Plant | Zero Point Generator |
| 13 | Hacker/Digital | hacker | data | 9 | Mainframe Hall | Reality Processor |
| 14 | Harbor | trade | gold (+trade-route income) | 5 | Harbor | Logistics Hub |

Food, Organic Extraction, Knowledge, Faith and Housing start in the Primitive Age, and Geological Extraction and Military in the Stone Age. The rest start later: Trade and Engineering in the Bronze Age, Metallurgy in the Iron, Culture/Arts in the Classical, Harbor in the Colonial, Energy in the Industrial and Hacker/Digital in the Modern. That is why they have fewer tiers.

Housing, Knowledge, Military and Culture/Arts have no worker slots. The Harbor lineage also raises trade income: besides producing gold, each harbor adds a percentage bonus to the income of every active trade route. See [Trade](trade.md#harbor-lineage-trade-route-income).

Each Knowledge building comes with its scholars, from 2 for a Story Circle to 20 for a Reality Academy, and the scholars take no worker slots. Each copy costs 50% more than the one before it. A knowledge building researches at full rate up to the number of copies its age's gate asks for, between 5 and 10. A copy past that count still builds, and it researches at half rate. [Scholars and research](knowledge.md#what-a-scholar-is) has each building's scholars, work and full-rate count.

The Mainframe Hall (Modern Age) is the first building of the Hacker/Digital lineage and the first that makes data. It uses `hacker` workers, the Modern Age's gate asks for eight of it, and its next building is the Server Farm in the Information Age.

---

## The 39 building lines

A line is a run of buildings in one lineage that make the same resource. A line never switches resource: when a lineage moves on to a new resource, the new buildings start a new line, and an older building only ever upgrades to the next building of its own line. Eight lineages never change what they make, so each is one line: Housing, Food, Knowledge, Faith, Military, Culture/Arts, Hacker/Digital and Harbor. The other six are split into 31 lines, 39 in all.

A line can pause and pick up again. Energy's electricity line stops for the Modern Age, when the Oil Refinery makes oil, and picks up in the Information Age. Metallurgy's titanium line skips the Cyberpunk and Fusion Ages, when the Dark Matter Refinery and the Exotic Matter Forge make dark matter. Trade's gold line skips the Cyberpunk Age, when the Black Market Hub makes crypto. A building whose line comes back is offered the next building of its line when its age arrives.

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

## What buildings do

A building can have one or more of these effects:

| Effect | What it does |
|--------|-------------|
| Production | Adds an amount of a resource each tick, scaled by how many workers it has (see below). Most production buildings do this. |
| Housing | Raises the number of people your homes hold (housing buildings, and some wonders). |
| Research | Knowledge buildings come with scholars who add research work to the tech in progress. Some wonders add research too. |
| Soldiers | A military building comes with soldiers instead of producing a resource. See [Army and missions](military.md). |
| Bonus | A percentage bonus to a rate, such as a wonder's or monument's bonus to all production, or the 1% each storage building adds (which add into [the all-production pool](resources.md#the-all-production-pool)). |
| Trade route income | Raises the income of every trade route (Harbor lineage). |
| Morale | Restores a little morale every tick (Faith and Culture/Arts buildings). |
| Opinion | Raises opinion with the civilizations you have met that are not hostile, each tick, scaled by workers (the two embassies; see [Factions and diplomacy](factions.md#embassy-buildings)). |

The Faith lineage (shrines, temples and their later buildings) and the Culture/Arts lineage restore a little morale every tick, and they need no workers for it. See [Morale](morale.md).

### Worker scaling formula

A building with worker slots makes:

```
actual_rate = base_rate × count × (0.20 + 0.80 × workers_assigned / total_capacity)
```

At 0 workers a building still produces 20% of its base rate (the idle floor). At full staffing it produces 100%. Workers beyond the slots add nothing. An empty building isn't wasted, only throttled.

Buildings with no worker slots (Culture/Arts buildings and wonders) produce exactly `base_rate × count` whatever your workers are doing.

### How production rates are set

`base_rate` is the fully staffed rate, and it is the number each building's description shows.

For construction resources (anything the buildings of an age cost: wood, stone, iron, gold, steel, coal, electricity, data and so on), the makers of an age follow the Payback Rule: fully staffed, a maker repays the price of its first copy in the same number of ticks as the age's other makers. That number is set for each age.

Output is valued at price parity. Each age has a price level for each resource: the median first-copy price in that resource among the age's buildings. Resources are worth each other in the ratio of those levels. A building with two outputs splits its value between them, in proportion to what the age's gate asks of each.

**Examples:**
- Wood Camp (Primitive): costs 22 wood (22.88 before the game rounds it down) and makes 0.569 wood/tick fully staffed, so it repays itself in about 40 ticks (80 seconds).
- Stone Pit (Stone): costs 257 stone and 429 wood. At Stone Age parity (343 stone = 515 wood) that is worth 543 stone, and at 3.48 stone/tick it repays in about 5 minutes.

Food, faith, culture and nanobots keep hand-set rates. Food feeds your workers, faith sets morale and catastrophe odds, and culture pays for festivals and monuments. Nanobots are a build material asked for in the tens or hundreds of thousands where an age's other prices are in the trillions, so a Nanobot Vat makes the 6.55K a tick it is listed at, and a building that asks for nanobots is not valued by them. A military building comes with its soldiers and makes no resource.

Times are at the base tick of 2 seconds. A game speed bonus shortens them in real time. Every construction resource an age charges for has a maker in that age: see [The maker rule](#the-maker-rule).

---

## The maker rule

Every material an age charges for, in a building its gate asks for or in its wonder, has a maker in that age: a building of that age that makes it. Every maker is set on one standard, so a maker repays the price of its first copy in the same number of ticks as the age's other makers (the Payback Rule above). The rule covers building materials. It leaves out nanobots and the flow resources, food, faith and culture.

One shortage is left as it is. The Electric Age has no building that makes coal, and a few more of your older buildings make up the difference. The Colonial Age makes no wood, so its Settlement Block, Harbor and Colonial Warehouse are priced in stone. One exception is listed: the last age's wonder is paid from the Quantum Age's makers, because the Transcendent Age has none of its own.

There is no market that swaps one material for another, so a town that wants a material builds a maker of it. The output tables below give a building's second material after the word "with".

Four raw materials are made and spent like wood or stone, in the ages that make them. In such an age every building and the wonder that cost the raw's worked form pay half of that part in the raw instead, one for one: marble with stone (Iron and Classical Ages), iron ore with iron (Iron to Industrial Ages), titanium ore with steel (Modern to Digital Ages) and dark matter crystals with electricity (Cyberpunk to Space Ages). The pair share one price level, so no price is larger or smaller overall. Housing, storage, the raws' own makers, the Mainframe Hall and the Black Market Hub pay no raw. See [What the ores are for](resources.md#what-the-ores-are-for).

---

## Lineage output progression

Several lineages change which resource they make as you move through the ages. The new resource starts a new line (see [The 39 building lines](#the-39-building-lines)): its first building is built fresh, the old line's buildings keep making the old resource, and nothing upgrades across from one to the other. A Sawmill never turns into a Coal Mine.

Some buildings arrive before their resource does. A resource gathers nothing until the age that unlocks it, whatever makes it. These buildings say `once unlocked` beside that output in their description:

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

### Organic Extraction (lineage 3)

| Ages | Output resource |
|------|----------------|
| Primitive to Medieval | wood |
| Renaissance to Industrial | coal |
| Victorian to Information | oil |
| Digital to Fusion | nanobots |
| Space to Quantum | quantum flux |

The Nano Foundry (Modern Age) is a standalone nanobot producer outside the lineage chain. It gives nanobots a producer as soon as they unlock, two ages before the lineage's own Bio Fabrication Lab. It uses `engineering` workers and makes +80 nanobots/tick. The lineage's own makers are listed at 3.28K (Bio Fabrication Lab), 6.55K (Nanobot Vat) and 13.1K (Molecular Synthesizer) a tick.

### Geological Extraction (lineage 4)

| Ages | Output resource |
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

The quarries and mines from the Iron Age to the Industrial make stone and iron themselves, and the Uranium Mine and the Uranium Processing Works make stone. The titanium mines and the crystal mines make only their raw.

### Engineering (lineage 9)

| Ages | Output resource |
|------|----------------|
| Bronze to Medieval | iron |
| Renaissance to Victorian | steel (with electricity in the Victorian Age) |
| Electric to Digital | electricity |
| Cyberpunk to Space | plasma (with electricity in the Cyberpunk and Fusion Ages) |
| Interstellar | dark matter, with plasma |
| Galactic | dark matter |
| Quantum | quantum flux, with dark matter |
| Transcendent | quantum flux |

### Metallurgy (lineage 11)

| Ages | Output resource |
|------|----------------|
| Iron to Medieval | iron |
| Renaissance to Industrial | steel |
| Victorian to Atomic | steel, with iron |
| Modern to Digital | titanium, with steel |
| Cyberpunk to Fusion | dark matter, with steel |
| Space | titanium, with steel |
| Interstellar to Galactic | antimatter, with titanium |
| Quantum | quantum flux |

---

## Storage buildings (21 tiers)

Storage buildings form their own lineage. Stock has no limit, so there is nothing for a storage building to hold. Each copy you own adds 1% to all production instead. They need no workers: build them and everything you make rises.

The boost goes into the same pool as the wonders' and monuments' bonuses to all production (see [the all-production pool](resources.md#the-all-production-pool)). It adds to theirs, shows as "Storage" among the sources in the Stats panel, and counts in full. Every copy counts the same, whatever its age: five copies of each of the 21 buildings add up to 105%.

Storage buildings have no copy limit. Each copy costs 1.75 times the one before it, so a tall stack gets dear fast. Every age gate asks for five copies of the storage building of the age you are leaving (see [The 22 ages](ages.md#advancement-requirements)). The experimental measured gate, which is off by default, replaces building counts with population, income and a tech of the age (see [The measured gate](ages.md#the-measured-gate-experimental)).

| Building | Age |
|----------|-----|
| Stash | Primitive |
| Storage Pit | Stone |
| Warehouse | Bronze |
| Granary | Iron |
| Classical Vault | Classical |
| Strongroom | Medieval |
| Renaissance Vault | Renaissance |
| Colonial Warehouse | Colonial |
| Industrial Depot | Industrial |
| Victorian Vault | Victorian |
| Electric Warehouse | Electric |
| Atomic Vault | Atomic |
| Modern Depot | Modern |
| Info Vault | Information |
| Digital Archive | Digital |
| Cyber Vault | Cyberpunk |
| Fusion Vault | Fusion |
| Orbital Depot | Space |
| Stellar Vault | Interstellar |
| Galactic Vault | Galactic |
| Quantum Vault | Quantum |

### No storage wall

Stock has no limit, so no building is refused for the size of its price. A building, an upgrade or a [build plan](plan.md) item is refused only for what it costs against what you hold, by `build`, `build <building> <count>`, `build <building> max`, `upgrade` and the plan alike. If you hold the stock, you can buy it.

Nothing you make is wasted, and nothing overflows. The age's wonder is still paid through its bank, which you fill by hand (see [Wonders](wonders.md)).

No price reads a store size. The festival, the black market and faction deals follow your income, and the harbinger's Brace follows what a moderate town of the age makes.

Storage buildings never upgrade and are never offered as upgrades when you advance. Every storage building you have built keeps adding its 1% for the rest of the run, so the stashes from the Primitive Age still count in the Quantum Age. Like every other building, an older age's storage can no longer be built once you advance (the game tells you to build the current age's storage instead), so fill each tier while it is current.

Storage is also permanent: catastrophes never destroy it or turn it into ruins, and it can't be sold.

---

## Cultural monuments (4)

Cultural monuments are one-off structures (one copy each) that turn surplus culture into a lasting payoff. Each costs a large lump of culture plus other materials of its age, and gives a bonus to all production while it stands. An Endure can destroy one like any building but wonders, storage and knowledge buildings. Unlike wonders, monuments need no banking: build them with the normal `build` command. Their bonuses add into the same pool as every other all-production bonus but a tech's. The pool applies in full, however large it grows, so a monument built late still raises your output by all of what it lists. See [The all-production pool](resources.md#the-all-production-pool).

| Monument | Age | Culture cost | Bonus |
|----------|-----|--------------|-----------------|
| Cultural Obelisk | Classical | 710 | +1% all production |
| Grand Amphitheater | Medieval | 7.1K | +2% all production |
| Eternal Library | Industrial | 140K | +3% all production |
| Monument of Ages | Modern | 7.1M | +5% all production |

Monuments are one way to spend culture, and the `festival` command is another. See [Culture](resources.md#culture).

---

## The Geographic Society

The Geographic Society is the one building that plays part of the game for you. It is opened in the Industrial Age by the Geographic Societies tech (which needs Cartography). It costs gold, steel and coal, holds 8 military workers, and produces no resource. Instead it sends out scouting expeditions on its own, so an empire left to run keeps exploring and keeps meeting the world's civilizations, which is where civilization boons come from.

You can build as many as you like. The wait between parties shortens with how many you have built and how fully you have staffed them. A game speed bonus shortens the wait in real time, and the Factions panel shows when the next party goes. The Society sends only short missions: it takes the affordable short offer with the best haul per minute. It pays the offer's full resource cost and sends a party only when the soldiers at home can make one up. If your stores cannot outfit one, it waits and sends it as soon as they can.

It keeps missions moving while you are away. Full details are in [Army and missions](military.md#automatic-dispatch-the-geographic-society).

---

## Wonders (22)

Wonders are unique buildings: you can build each wonder only once, and there is one for each age. You bank resources toward a wonder before you can build it:

```
wonder                                          show the bank of the current wonder
wonder collect <resource|all> [amount|all|max]  bank resources toward the current wonder
build <wonder_key>                              start construction once it is fully banked
```

`build` takes no resources from your stock for a wonder. A wonder in your [build plan](plan.md) banks and starts by itself once what you hold covers the rest.

You must build each age's wonder before you can advance to the next age. A wonder's price is worked out from what a typical building of its age charges, and each wonder keeps its own mix of resources. Build-cost cuts do not apply to it. A wonder takes the longest to build of any building in its age, and the techs that shorten wonder construction cut it further.

Each wonder has effects of its own. Some produce one or two resources, some add research, housing, a bonus to expedition rewards or to the knowledge rate, and six add a percentage to all production: the Crystal Palace, Hoover Dam and the last four wonders. All-production bonuses add into [the all-production pool](resources.md#the-all-production-pool), which applies in full, however large it grows.

Endure and the Great Fire never destroy wonders, and Succumb never turns them into ruins. They stay for the rest of the run, like storage.

The Wonders panel (`wonders`) shows the bank and the wonders you have built. Press `G` or `Tab` there for the gallery, a list of the wonders you have built, oldest first. See [Wonders](wonders.md) for the full list with costs and effects.

---

## Ruins

Ruins are what is left of a previous civilization. You get them when you choose Succumb during a catastrophe: up to 8 of your buildings (never wonders, storage or knowledge buildings) become ruins and carry forward into your new run, through prestige too. You can hold at most 24 ruins. When a new batch goes over, the lowest-value ruins (earliest age, then lowest output) crumble first.

| Property | Normal building | Ruin |
|----------|----------------|------|
| Production | `base × count × (0.20 + 0.80 × fill)` | `base × count × 0.50` |
| Workers | Needed for full output | Not needed (produces on its own) |
| Can build more? | Yes | No |
| Destroyed by Endure? | Yes (20% of buildings at random, fewer if you brace; never wonders, storage or knowledge buildings) | No |

Ruins give your new run a head start. They produce at 50% of base rate with no workers, so food, wood and stone (or whatever the ruined building made) trickle in from the first tick without any setup.

Ruins build up across runs: each Succumb adds up to 8 more on top of the ones you already have, up to the cap of 24.

---

## Worker capacity

Every building that takes workers has a number of worker slots per copy:

```
total_capacity = building_count × slots per copy
```

More copies mean more worker slots and a higher production ceiling. Building more copies of the current building is how you grow output within an age. A building with no workers produces only 20% of its base rate.

Your population grows toward your housing while food allows (auto-recruit is on by default), and the game puts idle workers to work by your roster. To do it by hand, recruit workers with `recruit [count|max]` and assign them with `assign <building> [count|all]`. `unassign` takes workers out of a building and they go idle, and `dismiss` removes them from your population. See [Workers and domains](workers-and-domains.md#the-roster).

---

## Culture buildings (lineage 10)

Culture buildings have no worker slots. They produce culture every tick at their base rate, with or without workers. Build them to earn culture passively.

They also restore morale every tick (see [Morale](morale.md)), so they do two jobs: culture for epoch events and morale for a production bonus. Your culture strength sets which good events your epoch changes can bring (see [Culture strength](epochs.md#culture-strength)).

---

## The Buildings list and panel

Two places list your buildings. The Buildings list on the dashboard shows the current age's buildings. Each entry has a mark (`✓` when you can pay for the next copy, `✗` when you cannot, `MAX` at the limit), the count you hold, the cost of the next copy, what the building does, and its worker slots filled over the total. Buildings a tech still locks show greyed out, with the research they need. Scroll the list with PgUp and PgDn (Option+↑ and Option+↓ on a Mac keyboard). When the Resources box has more resources than it can show, Ctrl+R turns it to its next page.

The Buildings panel (`buildings`) lists every building you have unlocked, age by age, oldest first, up to your current age. Each entry has the name and key, how many you hold, the `(legacy)` mark on an older building, its worker slots filled over the total, the cost of the next copy and what the building does. It is a browser: build with `build`.

Use `rates` to print the rate breakdown for your resources, and `status` for a detailed status summary.
