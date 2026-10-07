# Resources

AgeForge has **26 resources** that unlock as you advance through the 22 ages. They accumulate on their own from your buildings and the workers in them. Every resource starts with a small base storage, so storage buildings decide how much of anything you can hold.

---

## Full Resource Table

### Basic Resources

| Resource | Key | Unlocks | Base Storage | Notes |
|----------|-----|---------|--------------|-------|
| Food | `food` | Primitive Age | 50 | Every worker eats it every tick, so the rate must stay positive |
| Wood | `wood` | Primitive Age | 50 | The first building material |
| Knowledge | `knowledge` | Primitive Age | 30 | Pays for research; buildings also cost it from the Medieval to the Colonial Age |
| Faith | `faith` | Primitive Age | 50 | Never drains on its own; its fill % of storage sets the epoch and catastrophe odds, and faith income lifts morale |
| Stone | `stone` | Stone Age | 50 | Durable construction material |
| Iron | `iron` | Bronze Age | 50 | Metal for tools, weapons and early engineering |
| Gold | `gold` | Bronze Age | 50 | Currency for the market, diplomacy and many buildings |
| Culture | `culture` | Classical Age | 50 | Spent on monuments, festivals and more; its fill % of storage decides which good epoch events you can get |

### Ores

These four come from Geological Extraction mines. Nothing in the game spends them; see [What the ores are for](#what-the-ores-are-for).

| Resource | Key | Unlocks | Base Storage | Notes |
|----------|-----|---------|--------------|-------|
| Marble | `marble` | Iron Age | 30 | Mined by the Iron and Classical Age marble quarries |
| Iron Ore | `iron_ore` | Iron Age | 30 | Mined from the Iron to the Industrial Age |
| Titanium Ore | `titanium_ore` | Space Age | 20 | Mined by Modern to Digital Age mines, which produce nothing until it unlocks |
| Dark Matter Crystals | `dark_matter_crystals` | Cyberpunk Age | 10 | Mined from the Cyberpunk to the Space Age |

### Industrial Resources

| Resource | Key | Unlocks | Base Storage | Notes |
|----------|-----|---------|--------------|-------|
| Steel | `steel` | Medieval Age | 30 | Refined metal for advanced construction; made from the Renaissance Age |
| Coal | `coal` | Renaissance Age | 50 | Fuel; Organic Extraction makes it from the Renaissance to the Industrial Age |
| Oil | `oil` | Industrial Age | 50 | Fuel for machines; made from the Victorian Age |
| Electricity | `electricity` | Victorian Age | 50 | Powers modern infrastructure |
| Uranium | `uranium` | Atomic Age | 30 | Reactor fuel, mined by Geological Extraction |

### Advanced Resources

| Resource | Key | Unlocks | Base Storage | Notes |
|----------|-----|---------|--------------|-------|
| Data | `data` | Modern Age | 50 | Made by Hacker buildings from the Information Age; in the Modern Age the market sells it |
| Nanobots | `nanobots` | Modern Age | 20 | Made by the **Nano Foundry** (Modern Age) and by Organic Extraction from the Digital to the Fusion Age; Digital and Cyberpunk buildings cost them |
| Crypto | `crypto` | Cyberpunk Age | 50 | No building makes it: it comes from the Neon Citadel wonder, the Blockchain tech and the market |
| Plasma | `plasma` | Fusion Age | 30 | Superheated gas for energy |
| Titanium | `titanium` | Space Age | 30 | Light metal for space construction |
| Dark Matter | `dark_matter` | Interstellar Age | 20 | Exotic matter for warp technology |
| Antimatter | `antimatter` | Galactic Age | 20 | Annihilation fuel for megastructures |
| Quantum Flux | `quantum_flux` | Quantum Age | 10 | Unstable quantum energy |

### Military

| Resource | Key | Unlocks | Base Storage | Notes |
|----------|-----|---------|--------------|-------|
| Soldiers | `soldiers` | Iron Age | 0 | Made by military buildings when workers staff them; spent on campaigns |

Soldiers are a true stockpiled resource (the 26th). See [Soldiers](#soldiers) below and [Military](military.md) for the full mechanics.

---

## Food

Food is the most critical resource. Every worker eats food every tick, whatever building it staffs, and every worker eats the same amount: the current age's rate. That is **0.06 food/tick** per worker in the Primitive Age, rising 12% with each age to about 0.23 in the Modern Age and 0.58 in the Quantum Age. When you advance, the new rate applies to every worker at once. See [Food Drain](workers-and-domains.md#food-drain) for the full table.

If food runs out, one worker starves every 5 ticks (10 seconds) until there is food in stock again. See [Starvation](workers-and-domains.md#starvation).

**How to increase food:**
- Build more Food lineage buildings (`gathering_camp`, `forager_post`, `farm`, ...). Workers come to staff them on their own.
- Give food a bigger [worker share](workers-and-domains.md#worker-shares), or `assign gathering_camp 3` by hand
- Research Fire Mastery, Animal Husbandry, Agriculture and the other food techs

**Watch the food rate** in the Economy panel. Auto-recruit keeps a food margin for you; `recruit max` does not.

---

## Knowledge

Knowledge pays for research. Every technology costs knowledge, and from the Medieval to the Colonial Age buildings cost it too. The Modern Age's wonder, the Space Program, costs 600B knowledge.

**How to increase knowledge:**
- Build Knowledge lineage buildings (Story Circle, Elders' Hall, Scriptorium, ...). Workers come to staff them.
- Give knowledge a bigger [worker share](workers-and-domains.md#worker-shares): `workers share knowledge 40`
- Buy it at the market with gold (see below)

Knowledge storage starts at 30 and grows with your storage buildings, like every other resource. A tech priced above your knowledge storage can't start until you build more storage.

Each Knowledge building has its own fully staffed rate: 0.2 knowledge/tick for a Story Circle, 0.6 for an Elders' Hall, 2.0 for a Scriptorium, 1.6 for an Agora and 3.2 for a Library. While buildings cost knowledge (Medieval to Colonial), it counts as a construction resource and its producers follow the [Payback Rule](buildings.md#how-production-rates-are-set): a Monastery Library makes 30.1/tick, a University 61.1. See [Knowledge](knowledge.md) for the full lineage.

**Gold for knowledge.** From the Medieval Age on, the market sells knowledge for gold. From the Medieval to the Colonial Age it trades at parity like any construction resource (1 gold buys 0.2 knowledge in the Medieval Age, 0.07 in the Colonial Age). From the Industrial Age on the rate is a fixed **5 knowledge per gold**, the practical way to pay for something like the Space Program.

---

## Faith

Faith comes from Faith lineage buildings (Shrine, Standing Stones, Altar, ...), and much faster with workers in them. It never drains on its own. It goes down only when something spends or removes it: see [Faith](faith.md).

Faith does two jobs:

- **Odds.** Your faith as a percentage of your faith storage sets the odds of a good epoch event and of a fated catastrophe striking. See [Faith Threshold Bands](faith.md#faith-threshold-bands).
- **Morale.** Your faith income (faith per tick, not the stockpile) lifts morale a little every tick, up to a limit, and worship buildings lift it too. See [Morale](morale.md).

Faith can't be bought at the market.

**Advice:** Keep faith workers in your faith buildings well before an epoch transition. The buildings make only a trickle with no one in them.

---

## Culture

Culture comes from the Culture/Arts lineage (Amphitheater onward). These buildings take no workers and produce culture on their own, so you can build them while your workers stay elsewhere. Each one also raises culture storage and lifts morale.

Culture has two jobs. Its **fill % of storage** decides which tier of good epoch event you can get at an epoch transition: over 40% makes Major events eligible, and over 75% adds a 15% chance at the Legendary one (see [Epochs](epochs.md)). And culture is **spent**, on the things below. Prestige resets culture along with every other resource.

### Culture sinks

**Cultural Monuments** are four one-off buildings (one of each) built with the normal `build <key>` command. Each costs a large lump of culture plus other materials and gives a small **permanent** all-production bonus once built (it counts toward [the all-production cap](#the-all-production-cap)).

| Monument | Age | Culture Cost | Permanent Bonus |
|----------|-----|--------------|-----------------|
| Cultural Obelisk | Classical | 710 | +1% all production |
| Grand Amphitheater | Medieval | 7.1K | +2% all production |
| Eternal Library | Industrial | 140K | +3% all production |
| Monument of Ages | Modern | 7.1M | +5% all production |

**The `festival` command** spends a lump of culture (2K, or 5% of your culture storage if that is more) for **+20% all production for 13 minutes** (390 ticks), then waits **26 minutes** (780 ticks) before the next one. See [Commands](commands.md).

Culture also pays for:

- **Smuggling runs** on the black market (Colonial Age on): a culture stake that may pay out a haul of the resource you pick. See [Trade](trade.md).
- **Appease**, together with faith, when a [harbinger](harbinger.md) warns of a catastrophe.
- **Tribute**, together with gold, to end a war with a civilization. See [Factions](factions.md).
- The **Sistine Chapel** wonder in the Renaissance Age.

---

## What the ores are for

Geological Extraction mines produce marble, iron ore, titanium ore and dark matter crystals. **Nothing spends them.** No building, technology, wonder, age requirement or market pair asks for any of the four, and the Metallurgy lineage doesn't refine them: metallurgy buildings make iron, steel, titanium, dark matter, antimatter and quantum flux and consume nothing, like every other producer.

The ores count toward two milestones:

- **Industrial Titan** (Industrial Age): stockpile 10,000 coal and 5,000 iron ore.
- **Mining Syndicate**: build 25 stone pits and 10 iron mines.

What Geological Extraction is good for is stone (up to the Bronze Age), uranium (from the Atomic Age) and antimatter (from the Galactic Age). Its other buildings, from the Iron Age's marble quarry to the Space Age's crystal mine, make only ore. Within a domain, workers fill the newest buildings first, so an ore mine draws masonry workers away from your older stone quarries. Build ore mines for the milestones, not for production. No building of the Iron Age or later makes stone: your older quarries keep making it, and the market sells it (see [Buying at the Market](#buying-at-the-market)). For the same reason, don't `upgrade` your quarries in the Iron Age: the upgrade turns them into marble quarries, which make ore instead of stone.

---

## Resource Storage

Every resource has a small base storage (10 to 50). **Storage buildings** raise the cap of every resource at once: one per age from the Stash to the Quantum Vault, up to 25 copies each (50 Stashes). A few technologies raise every cap too, culture buildings add culture storage, and military buildings add soldier storage. See [Storage Buildings](buildings.md#storage-buildings-21-tiers) for the full table.

**How much it holds.** A full stack of an age's storage, with every earlier age's, holds at least **4.5 hours** of that age's typical production of each resource its buildings cost, from the Bronze Age on. The Primitive and Stone Ages hold an hour and a half: they fill fast and are meant to. A player who checks in every few hours loses little to a full store. On known ground, [Era Mastery](prestige.md#era-mastery) multiplies every cap by the age's speed along with production, so a store still holds the same hours of income.

**Storage is permanent.**

- It never upgrades. When you advance, your storage stays as it is and keeps counting.
- An earlier age's storage can't be built once its age has passed, so build each age's storage while you can.
- It can't be sold.
- Catastrophes don't destroy it.

**Production a full store would waste goes somewhere useful first.** It goes into the current age's [wonder](wonders.md#overflow) bank, up to what the wonder still needs, then into your [build plan](plan.md#overflow-pays-the-plan), toward the next copy of each queued building in plan order. Only what neither needs is lost. This works during offline catch-up too.

**Priority:** build this age's storage early in every age. Late resources (plasma, titanium, dark matter, antimatter, quantum flux) start at 10 to 30 storage and fill almost at once without it.

---

## The all-production cap

Every "+X% all production" bonus in the game adds into one pool: wonders, technologies, milestones, cultural monuments, events (festivals included) and faction boons. Penalties, like the -10% Reconstruction Effort after you [Endure](catastrophe.md#endure) a catastrophe, come out of the same pool. The game multiplies your output by **1 + that pool**, clamped between **x0.1 and x3.0**.

Techs and wonders alone reach the x3.0 cap from about the Electric Age. After that, a late "+X% all production" bonus adds nothing you can see while you are over the cap. It isn't wasted: a penalty comes out of the raw pool first, so the surplus absorbs it, and the bonus matters again whenever a setback pulls the pool back under the cap.

One bonus sits outside the pool: the [Cosmic Legacy](prestige.md#cosmic-legacy). Its +10% is applied after the cap, to everything a resource makes, so it counts in every age however full the pool is.

Bonuses to one resource ("+30% gold") have their own pool for that resource, with the same x0.1 to x3.0 clamp. Gold's pool fills in the Colonial Age and knowledge's in the Electric Age, from techs alone.

**The game tells you when a cap is holding a bonus back.** Nothing is hidden above the cap:

- The Stats panel's Active Multipliers shows what counts, with a note when a pool is past its cap: `All production +200% capped at +200%: +405% earned`. The sources beside it still list everything you earned.
- The Research panel's **Research bonuses** list carries the same note, and so does every researched tech in a capped pool. A tech you can still research says what the cap would leave of it, before you spend the knowledge: `(capped: no effect now)` or `(capped: +5% of it counts now)`.
- The Milestones and Wonders panels put the same notes beside rewards and wonder effects.
- The `festival` command warns you before you pay when the cap would swallow the festival.
- The log adds a line when a bonus you just earned is capped: a tech, a milestone, a wonder, an awakening, an epoch event, a festival or a boon.

[Worker output](workers-and-domains.md#worker-output-bonuses) bonuses are not in either pool and have no cap.

[Era Mastery](prestige.md#era-mastery) is not part of either pool. On known ground it multiplies every resource's net rate by the age's speed after the cap and after food drain, so a mastered age is not held to x3.

---

## Production Rates

Every building rate in the game is the **fully staffed** rate shown in the building's description. For **construction resources** (anything the buildings of your current age cost) the rate follows the **Payback Rule**: fully staffed, a producer earns back the price of its first copy within the age's payback time. That time is a growing share of the age: under a minute in the Primitive Age, 58 minutes in the Iron Age, 5 hours in the Renaissance. Later ages repay more slowly because your older buildings keep producing too. The **flow resources** (food, faith, culture and soldiers) keep hand-set rates, and the requirements that ask for them are sized to match. Details and the full table are on [Buildings](buildings.md#how-production-rates-are-set).

---

## Buying at the Market

Once you own any trade-lineage building (a `market` or its successors), the market trades **any two construction resources of your current age** at their **price parity less a 20% fee**. Each age has a price level per resource (the typical first-copy price in that resource), and one unit buys `0.8 × (price level of what you buy) / (price level of what you sell)`. In the Space Age, for example, steel and titanium are priced 180T and 245T, so 1 steel buys 1.09 titanium and 1 titanium buys 0.588 steel. Any pair of construction resources works (steel → titanium in the Space Age, data → crypto in the Cyberpunk Age). A listed pair with a resource that isn't a construction resource of the age (food, culture, faith, or knowledge outside the Medieval to Colonial Ages) trades at its fixed rate. Repeated trades of one pair push its rate down (floor 50%), and a round trip always loses value, so trading never beats building. See [Trade](trade.md#resource-exchange).

Some construction resources have **no producing building in certain ages**, and the market is where you get them. Producers you built in earlier ages keep running, but the current age has none to build:

- **Stone** from the Iron Age on
- **Iron** from the Renaissance Age on
- **Steel** from the Modern Age on
- **Titanium** from the Interstellar Age on (in the Space Age the Orbital Refinery makes it; earlier titanium smelters make only a trickle)
- **Crypto** in the Cyberpunk Age (apart from the Neon Citadel wonder and the Blockchain tech)

The same goes for wood in the Colonial Age, coal in the Electric Age, data in the Modern Age, plasma in the Galactic Age, dark matter from the Quantum Age on and antimatter in the Transcendent Age. Faith can't be bought at the market at all. A [build plan](plan.md) trade item (`plan trade <give> <get>`) buys these as the resources come in, while you are away too.

---

## Gold

Gold pays for the market, diplomacy gifts and many mid-to-late game buildings.

**How to increase gold:**
- Build Trade lineage buildings (Market, Trading Post, Merchant Quarter, ...) and staff them
- Research Currency, Mercantilism and the other gold techs
- Trade routes pay small fixed amounts on top (see [Trade](trade.md))

---

## Soldiers

Soldiers are a stockpiled resource (the 26th) that unlocks at the **Iron Age**. Only military buildings produce them.

**How soldiers are produced:** each military building (War Camp, Barracks and their successors) produces soldiers every tick when military-domain workers staff it, the same way food or wood production scales with workers. A fully staffed military building produces its soldier cap ÷ 50 soldiers per tick (at least 0.1/tick). War Camps and Barracks come before the Iron Age but make no soldiers until it.

**How soldiers are stored:** each military building adds its soldier cap to soldier storage: a War Camp 10, a Barracks 20, doubling with each tier up the lineage. Your shared storage (storage buildings and storage techs) counts toward the soldier cap too.

**How soldiers are spent:** launching a campaign costs soldiers, taken from your stockpile at launch whether the campaign succeeds or fails. Scouting expeditions cost none. Enduring a catastrophe also cuts your soldiers along with every other stored resource.

See [Military](military.md) for production formulas, per-building soldier caps, and the full campaign and expedition table.
