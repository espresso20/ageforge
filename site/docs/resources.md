# Resources

AgeForge has **24 resources** that unlock as you advance through the 22 ages. They accumulate on their own from your buildings and the workers in them. Stock has no limit: what you make piles up, and prices are the only brake. Research is not one of them: it is done by [scholars](knowledge.md), and nothing holds it.

Amounts print with up to three figures and a unit from a thousand up: K, M, B, T and Q, and past Q the units go on as Qi, Sx, Sp, Oc, No and Dc (1.5K, 876M, 11.8Qi).

---

## Full Resource Table

### Basic Resources

| Resource | Key | Unlocks | Notes |
|----------|-----|---------|-------|
| Food | `food` | Primitive Age | Every worker eats it every tick, so the rate must stay positive |
| Wood | `wood` | Primitive Age | The first building material |
| Faith | `faith` | Primitive Age | Never drains on its own; your faith strength (your faith buildings against a moderate set, times the share of your faith you have kept) sets the epoch and catastrophe odds, and faith income lifts morale |
| Stone | `stone` | Stone Age | Durable construction material |
| Iron | `iron` | Bronze Age | Metal for tools, weapons and early engineering |
| Gold | `gold` | Bronze Age | Pays for diplomacy and many buildings |
| Culture | `culture` | Classical Age | Spent on monuments, festivals and more; your culture strength decides which good epoch events you can get |

### Ores

These four come from Geological Extraction mines. In the ages that make one, buildings are priced in it; see [What the ores are for](#what-the-ores-are-for).

| Resource | Key | Unlocks | Notes |
|----------|-----|---------|-------|
| Marble | `marble` | Iron Age | Made by the Marble Quarry and the Marble Works. Bought in the Iron and Classical Age, with stone |
| Iron Ore | `iron_ore` | Iron Age | Made from the Iron to the Industrial Age. Bought in those ages, with iron |
| Titanium Ore | `titanium_ore` | Modern Age | Made by the Modern to Digital Age titanium mines. Bought in those ages, with steel |
| Dark Matter Crystals | `dark_matter_crystals` | Cyberpunk Age | Made from the Cyberpunk to the Space Age. Bought in those ages, with electricity |

### Industrial Resources

| Resource | Key | Unlocks | Notes |
|----------|-----|---------|-------|
| Steel | `steel` | Medieval Age | Refined metal for advanced construction; made from the Renaissance Age |
| Coal | `coal` | Renaissance Age | Fuel; Organic Extraction makes it from the Renaissance to the Industrial Age |
| Oil | `oil` | Industrial Age | Fuel for machines; made from the Victorian Age |
| Electricity | `electricity` | Victorian Age | Powers modern infrastructure |
| Uranium | `uranium` | Atomic Age | Reactor fuel, mined by Geological Extraction |

### Advanced Resources

| Resource | Key | Unlocks | Notes |
|----------|-----|---------|-------|
| Data | `data` | Modern Age | Made by Hacker buildings from the Modern Age, starting with the Mainframe Hall |
| Nanobots | `nanobots` | Modern Age | Made by the **Nano Foundry** (Modern Age) and by Organic Extraction from the Digital to the Fusion Age; Digital and Cyberpunk buildings cost them. A build material with hand-set rates |
| Crypto | `crypto` | Cyberpunk Age | Made by the Black Market Hub, with a trickle from the Neon Citadel wonder. Blockchain adds 4% to it |
| Plasma | `plasma` | Fusion Age | Superheated gas for energy |
| Titanium | `titanium` | Space Age | Light metal for space construction |
| Dark Matter | `dark_matter` | Interstellar Age | Exotic matter for warp technology |
| Antimatter | `antimatter` | Galactic Age | Annihilation fuel for megastructures |
| Quantum Flux | `quantum_flux` | Quantum Age | Unstable quantum energy |

Soldiers are not on this list. They are people, not a resource: the army is the soldiers at the posts of your military buildings. See [Soldiers](#soldiers) below and [Military](military.md) for the full mechanics.

---

## Food

Food is the most critical resource. Every worker eats food every tick, whatever building it staffs, and every worker eats the same amount: the current age's rate. That is **0.06 food/tick** per worker in the Primitive Age, rising 12% with each age to about 0.23 in the Modern Age and 0.58 in the Quantum Age. When you advance, the new rate applies to every worker at once. See [Food Drain](workers-and-domains.md#food-drain) for the full table.

If food runs out, one worker starves every 5 ticks (10 seconds) until there is food in stock again. See [Starvation](workers-and-domains.md#starvation).

**How to increase food:**
- Build more Food lineage buildings (`gathering_camp`, `forager_post`, `farm`, ...). Workers come to staff them on their own.
- Give food a bigger [share of the roster](workers-and-domains.md#the-roster), or `assign gathering_camp 3` by hand
- Research Tool Making, Fire Mastery, Animal Husbandry and Agriculture: +10% food each, and they add up

**Watch the food rate** in the Economy panel. Auto-recruit keeps a food margin for you; `recruit max` does not.

---

## Faith

Faith comes from Faith lineage buildings (Shrine, Standing Stones, Altar, ...), and much faster with workers in them. It never drains on its own. It goes down only when something spends or removes it: see [Faith](faith.md).

Faith does two jobs:

- **Odds.** Your faith strength sets the odds of a good epoch event and of a fated catastrophe striking. It measures what your faith buildings have made against a moderate set (eight staffed copies of each), times the share of your faith you have kept. A typical town sits in the bottom band; more faith buildings and fuller crews move it up. It is not a share of what you hold: stock has no limit, so the odds read your faith buildings and the share you have kept, not a pile. See [Faith Threshold Bands](faith.md#faith-threshold-bands).
- **Morale.** Your faith income (faith per tick, not the stockpile) lifts morale a little every tick, up to a limit, and worship buildings lift it too. See [Morale](morale.md).

**Advice:** Keep faith workers in your faith buildings well before an epoch transition. The buildings make only a trickle with no one in them.

---

## Culture

Culture comes from the Culture/Arts lineage (Amphitheater onward). These buildings take no workers and produce culture on their own, so you can build them while your workers stay elsewhere. Each one also lifts morale.

Culture has two jobs. Your **culture strength** decides which tier of good epoch event you can get at an epoch transition: over 40% makes Major events eligible, and over 75% adds a 15% chance at the Legendary one. It is what your own culture buildings have made against a moderate set, times the share of your culture you have kept, so it takes about twice a moderate set to open Major events, and culture you spend or buy does not help (see [Culture strength](epochs.md#culture-strength)). And culture is **spent**, on the things below. Prestige resets culture along with every other resource.

### Culture sinks

**Cultural Monuments** are four one-off buildings (one of each) built with the normal `build <key>` command. Each costs a large lump of culture plus other materials and gives a small **permanent** all-production bonus once built (it adds into [the all-production pool](#the-all-production-pool)).

| Monument | Age | Culture Cost | Permanent Bonus |
|----------|-----|--------------|-----------------|
| Cultural Obelisk | Classical | 710 | +1% all production |
| Grand Amphitheater | Medieval | 7.1K | +2% all production |
| Eternal Library | Industrial | 140K | +3% all production |
| Monument of Ages | Modern | 7.1M | +5% all production |

**The `festival` command** spends a lump of culture (2K, or if that is more a number of minutes of your own culture income, set by the age) for **+20% all production for 13 minutes** (390 ticks), then waits **26 minutes** (780 ticks) before the next one (Baroque Arts makes the boost last 25% longer, Radio shortens the wait by 20%, and Social Media the price by 20%). See [Commands](commands.md).

Culture also pays for:

- **Smuggling runs** on the black market (Colonial Age on): a culture stake that may pay out a haul of the resource you pick. See [Trade](trade.md).
- **Appease**, together with faith, when a [harbinger](harbinger.md) warns of a catastrophe.
- **Tribute**, together with gold, to end a war with a civilization. See [Factions](factions.md).
- The **Sistine Chapel** wonder in the Renaissance Age.

---

## What the ores are for

In the ages that make one, each of the four is a material like wood or stone. A building of that age pays half of the part that would have been its worked form in the raw instead, one for one. Marble stands in for half the stone (Iron and Classical Age), iron ore for half the iron (Iron to Industrial Age), titanium ore for half the steel (Modern to Digital Age) and dark matter crystals for half the electricity (Cyberpunk to Space Age). The age's wonder is paid the same way. 92 buildings cost a raw, wonders among them.

A raw and its worked form share one price level, so a price in units is what it was: a building that cost 6K stone costs 3K stone and 3K marble. Housing and storage pay no raw, nor do the mines that make it. The Mainframe Hall and the Black Market Hub pay none either, because each makes what its age's mine is priced in.

Where a raw is made:

| Raw | Made by | Bought in |
|-----|---------|-----------|
| Marble | Marble Quarry, Marble Works | Iron and Classical Age |
| Iron Ore | Marble Quarry, Marble Works, Stonemason's Guild, Iron Mine, Deep Iron Mine, Steam Mine | Iron to Industrial Age |
| Titanium Ore | Titanium Mine, Precision Mine, Nano Drill Complex | Modern to Digital Age |
| Dark Matter Crystals | Dark Crystal Mine, Exotic Mineral Extractor, Asteroid Crystal Mine | Cyberpunk to Space Age |

Research raises each raw like any other material, and every tech bonus on one adds up with the rest (the full list is under [Output of one resource](technologies.md#output-of-one-resource)):

| Raw | Techs | Together |
|-----|-------|----------|
| Marble | Quarrying +8% | +8% |
| Iron Ore | Quarrying +8%, Surveying +6%, Steam Pumps +6%, Geology +5% | +25% |
| Titanium Ore | Titanium Alloys +5%, Precision Mining +5% | +10% |
| Dark Matter Crystals | Dark Crystal Mining +4%, Asteroid Refining +4%, Space Mining +4% | +12% |

Stone, titanium and dark matter have research too: stone +38% (Stoneworking, Masonry, Bronze Working, Civil Engineering), titanium +16% (Quantum Metallurgy joins its three) and dark matter +16% (Cosmology and Zero-Point Energy join its two).

A raw can also be a deal good in a faction trade, in the ages that price it (see [Trade deals](factions.md#trade-deals)).

The ores count toward two milestones:

- **Industrial Titan** (Industrial Age): stockpile 10,000 coal and 5,000 iron ore.
- **Mining Syndicate**: build 22 stone pits and 10 iron mines.

The quarries and mines of the Iron to the Industrial Age make stone, iron and iron ore, and the Marble Quarry and Marble Works make marble too. The titanium mines and the crystal mines make only their raw, at the standard rate. Within a domain, workers fill the newest buildings first, so a mine that makes only a raw draws masonry workers away from your older quarries.

---

## Resource Storage

**Stock has no limit.** Nothing you make is wasted, and no store is ever full. Production, hand gathering, loot, event gains, milestone rewards, black market wins, trade runs, faction deals, boons and the refund for a sold building all add their whole amount, however much you hold. The same goes for the time you spend away: what your town makes while the game is closed is added in full. Rising prices are the only thing that holds you back.

**Nothing is refused for costing more than a store.** A building, an upgrade or a [build plan](plan.md) item is refused only for what it costs against what you have. If you hold the stock, you can buy it. The game no longer names a storage building to build first.

**No rule reads a store size.** A few prices used to be a share of one: the [festival](#culture), the [black market](trade.md#black-market), [faction deals](factions.md#trade-deals) and the harbinger's [Brace](harbinger.md#brace-soften-an-endure). They follow your income now. The festival, the black market and a deal's size are minutes of what your own town makes of the resource, so a town that makes twice the culture pays twice. Brace is priced on what a moderate town of the age makes, and capped at the same figure for every town.

**Storage buildings no longer hold anything.** There is one per age, from the Stash to the Quantum Vault, with no limit on copies, each copy costing 1.75 times the one before it. Each copy you own adds **1% to all production**, in the same pool as the wonders' and monuments' bonuses to all production (see [the all-production pool](#the-all-production-pool)). See [Storage Buildings](buildings.md#storage-buildings-21-tiers) for the list. Every age gate still asks for five copies of the storage building of the age you are leaving.

**Storage is permanent.**

- It never upgrades. When you advance, your storage stays as it is and keeps counting.
- An earlier age's storage can't be built once its age has passed, so build each age's storage while you can.
- It can't be sold.
- Catastrophes don't destroy it.

When you advance to a new age, the stock you carry over is trimmed to a few copies of the cheapest building of the new age that uses it. That is a head start: see [Ages](ages.md).

---

## The all-production pool

Every "+X% all production" bonus in the game adds into one pool, except a technology's: wonders, milestones, cultural monuments, [storage buildings](buildings.md#storage-buildings-21-tiers) (1% a copy), events (festivals included) and faction boons. Penalties, like the -10% Reconstruction Effort after you [Endure](catastrophe.md#endure) a catastrophe, come out of the same pool. The game multiplies your output by **1 + the pool**, and never by less than **x0.1**.

**There is no ceiling.** The pool applies in full, however large it grows. Nothing you earn counts for less because of what you already hold, and a bonus is worth the same number of points at any point in a run.

> **A worked example.** Your milestones, wonders and monuments have earned +320% all production, so your output is multiplied by 4.2. Then you finish a milestone worth +10%. The pool is now +330% and your output is multiplied by 4.3: the milestone added 10 points.

Milestone rewards fill most of the pool, and the five copies of every storage building the age gates ask for add 5% for each age. By the last age a full set of milestones, wonders and monuments, with five copies of every storage building, has earned +766%, and your output is multiplied by 8.66.

| The pool holds | Output is multiplied by |
|---|---|
| +100% | 2 |
| +200% | 3 |
| +320% | 4.2 |
| +400% | 5 |
| +766% | 8.66 |

Timed bonuses follow the same rule. A festival's +20% or an Age of Plenty's +100% adds exactly that to the pool, whatever it holds. So do penalties: a -10% setback costs 10 points.

Two things sit outside the pool and are applied after it, so they count in full in every age however much the pool holds:

- **Technologies.** A tech's "+X% production" multiplies your output after the pool has been applied, and tech bonuses add up among themselves with no cap of their own. See [How Tech Bonuses Stack](technologies.md#how-tech-bonuses-stack).
- The [Cosmic Legacy](prestige.md#cosmic-legacy), +10% on everything a resource makes.

Bonuses to one resource ("+10% gold production" from a milestone or a wonder) have their own pool for that resource, with the same rule: in full, however large, never under x0.1. The fullest in any age, the scholars' work (scholar output), holds +145% with every scholar milestone and both research wonders.

**The game shows the pool as it is.** The few limits that remain are named where they bite:

- The Stats panel's Active Multipliers shows the whole pool as its total, with the sources beside it. It adds a note only when a limit holds a pool back: the x0.1 floor, or research speed past +100%.
- The `rates` command adds a line only for a pool a limit holds below what it has earned: `All production: -140% earned, -90% counts.`
- The Milestones and Wonders panels list every all-production reward and wonder effect as it is, with no note. A reward in a pool a limit holds says what it would add before you work for it: `+5% research speed (capped: no effect now)`.
- The tech tree never shows one: no cap holds a tech's bonus.
- The log adds a line only when a bonus you just earned is held back by a limit.

<figure class="screen" data-screen="rates"><figcaption>The foot of the Stats panel in the Bronze Age: the rate of every resource, then the multipliers in play and where each one comes from.</figcaption></figure>

[Worker output](workers-and-domains.md#worker-output-bonuses) bonuses are not in either pool and have no cap.

[Era Mastery](prestige.md#era-mastery) is not part of either pool. On known ground it multiplies every resource's net rate by the age's speed after the pools and after food drain, in full.

---

## Production Rates

Every building rate in the game is the **fully staffed** rate shown in the building's description. For **construction resources** (anything the buildings of your current age cost) the rate follows the **Payback Rule**: fully staffed, a producer earns back the price of its first copy within the age's payback time. That time is a growing share of the age: under a minute in the Primitive Age, 45 minutes in the Iron Age, 5 hours in the Renaissance. Later ages repay more slowly because your older buildings keep producing too. The **flow resources** (food, faith, culture and nanobots) keep hand-set rates, and the requirements that ask for them are sized to match. Details and the full table are on [Buildings](buildings.md#how-production-rates-are-set).

---

## Where the Hard Ones Come From

There is no market to buy a material from. A material an age charges for comes from a building of that age that makes it (see [The Maker Rule](buildings.md#the-maker-rule)). Several of these makers sit in a lineage you might not look in:

- **Stone** from the Iron Age on: the Marble Quarry through the Steam Mine, then the Uranium Mine and the Uranium Processing Works (all Geological Extraction).
- **Iron**: the Marble Quarry through the Steam Mine in Geological Extraction, then the Bessemer Plant, the Electric Arc Furnace and the Advanced Alloy Plant in Metallurgy.
- **Marble** in the Iron and Classical Age: the Marble Quarry and the Marble Works. **Iron ore** from the Iron Age to the Industrial: the Marble Quarry, the Marble Works, the Stonemason's Guild, the Iron Mine, the Deep Iron Mine and the Steam Mine.
- **Steel** from the Modern Age on: the Titanium Smelter, Aerospace Foundry, Nano Alloy Plant, Dark Matter Refinery, Exotic Matter Forge and Orbital Refinery (Metallurgy).
- **Titanium** in the Space Age from the Orbital Refinery, in the Interstellar Age from the Antimatter Forge and in the Galactic Age from Stellar Metallurgy. The earlier titanium smelters make only a trickle.
- **Titanium ore** from the Modern Age to the Digital: the Titanium Mine, the Precision Mine and the Nano Drill Complex. **Dark matter crystals** from the Cyberpunk Age to the Space Age: the Dark Crystal Mine, the Exotic Mineral Extractor and the Asteroid Crystal Mine.
- **Data** from the Modern Age: the Mainframe Hall, then the Server Farm and the rest of the Hacker lineage.
- **Crypto** in the Cyberpunk Age: the Black Market Hub, with a trickle from the Neon Citadel wonder.
- **Plasma** in the Interstellar Age: the Warp Drive Plant. **Dark matter** in the Quantum Age: the Reality Forge.

One material has no maker of its own age: coal in the Electric Age. The older buildings you keep cover it. A material nothing in an age makes any more is no longer charged there: the Colonial Age's buildings are priced in stone, not wood.

---

## Gold

Gold pays for diplomacy gifts and many mid-to-late game buildings.

**How to increase gold:**
- Build Trade lineage buildings (Market, Trading Post, Merchant Quarter, ...) and staff them
- Research Currency, Road Building, Banking, Wire Transfers, Corporations and Probability Markets: +10%, +8%, +8%, +5%, +5% and +4% gold, and they add up to +40%
- Trade routes pay small fixed amounts on top (see [Trade](trade.md))

---

## Soldiers

Soldiers are not a resource: they are the people at the posts of your military buildings. The first stand when a military building does (a War Camp can be built in the Stone Age).

**Where soldiers come from:** each copy of a military building (War Camp, Barracks and their successors) has a fixed number of posts, which fill from the reserve and then with new arrivals as the food allows. Nothing produces or stores them, and a military building has no worker slots. Two technologies, the Military-Industrial Complex and Augmented Soldiers, give every military building a tenth more posts each. A copy that is lost takes its soldiers with it.

**What soldiers need:** they are people, so they take a place in your homes and eat the same food as a worker. If your homes are full when a military building stands, its soldiers still arrive (as far as the food margin allows) and your reserve stops growing until there is room.

**How soldiers are spent:** every mission takes a party out of the soldiers at home, and some of the party fall on the way (see [Parties and losses](military.md#parties-and-losses)). A raid that lands against the garrison costs soldiers too. An Endure leaves the soldiers at their posts.

Soldiers are one of the four kinds of [people](workers-and-domains.md#the-people-of-your-town), so the People box lists them (with how many are out on missions) rather than the Resources box, and the People panel's Army tab shows them building by building.

See [Military](military.md) for the soldiers each building comes with, and the full campaign and expedition table.
