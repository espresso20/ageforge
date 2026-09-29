# Resources

AgeForge has **26 resources** that unlock progressively as you advance through the 22 ages. Resources accumulate passively based on your buildings and worker assignments. Each resource has a small base storage, so you need Storage lineage buildings to hold meaningful quantities.

---

## Full Resource Table

### Basic Resources

| Resource | Key | Unlocks | Base Storage | Notes |
|----------|-----|---------|--------------|-------|
| Food | `food` | Primitive Age | 50 | Feeds all workers every tick, so the rate must stay positive |
| Wood | `wood` | Primitive Age | 50 | Primary early building material |
| Knowledge | `knowledge` | Primitive Age | 30 | Pays for all research; storage starts low, so build the Knowledge lineage early |
| Faith | `faith` | Primitive Age | 50 | Does not drain on its own; its fill % of storage sets the epoch roll odds |
| Stone | `stone` | Stone Age | 50 | Durable construction material |
| Iron | `iron` | Bronze Age | 50 | Metal for tools, weapons, and early engineering |
| Gold | `gold` | Bronze Age | 50 | Currency for trade, diplomacy, and mid-game buildings |

### Knowledge & Faith

Both have mechanics beyond simple accumulation. See their sections below.

### Culture

| Resource | Key | Unlocks | Base Storage | Notes |
|----------|-----|---------|--------------|-------|
| Culture | `culture` | Classical Age | 50 | Spent on monuments, festivals and more; its fill % of storage decides which good epoch events you can get |

See the section below.

### Ore Intermediates

These resources are produced by Geological Extraction buildings and consumed by the Metallurgy lineage. They are intermediate materials, not used directly in buildings or tech.

| Resource | Key | Unlocks | Base Storage | Notes |
|----------|-----|---------|--------------|-------|
| Marble | `marble` | Iron Age | 30 | Refined stone for monumental construction |
| Iron Ore | `iron_ore` | Iron Age | 30 | Raw ore before smelting; feeds the Metallurgy lineage |
| Titanium Ore | `titanium_ore` | Space Age | 20 | Raw titanium ore; the Metallurgy lineage refines it into titanium |
| Dark Matter Crystals | `dark_matter_crystals` | Cyberpunk Age | 10 | Crystallized dark matter; refines into dark matter |
| Nanobots | `nanobots` | Modern Age | 20 | Microscopic machines. Built by the **Nano Foundry** (Modern Age) and by Organic Extraction from the Digital Era on; several digital and cyberpunk buildings use them as a build material |

### Industrial Resources

| Resource | Key | Unlocks | Base Storage | Notes |
|----------|-----|---------|--------------|-------|
| Coal | `coal` | Renaissance Age | 50 | Fuel for smelting; Organic Extraction output in the Steel Era |
| Steel | `steel` | Medieval Age | 30 | Refined metal for advanced construction |
| Oil | `oil` | Industrial Age | 50 | Fuel for machines; Organic Extraction output in the Electric Era |
| Electricity | `electricity` | Victorian Age | 50 | Powers modern infrastructure |
| Uranium | `uranium` | Atomic Age | 30 | Radioactive reactor fuel |

### Advanced Resources

| Resource | Key | Unlocks | Base Storage | Notes |
|----------|-----|---------|--------------|-------|
| Data | `data` | Modern Age | 50 | Digital information and analytics; produced by the Hacker domain |
| Crypto | `crypto` | Cyberpunk Age | 50 | Decentralized digital currency |
| Plasma | `plasma` | Fusion Age | 30 | Superheated ionized gas for energy |
| Titanium | `titanium` | Space Age | 30 | Lightweight refined metal for space construction |
| Dark Matter | `dark_matter` | Interstellar Age | 20 | Exotic refined matter for warp technology |
| Antimatter | `antimatter` | Galactic Age | 20 | Annihilation fuel for megastructures |
| Quantum Flux | `quantum_flux` | Quantum Age | 10 | Unstable quantum energy for reality manipulation |

### Military

| Resource | Key | Unlocks | Base Storage | Notes |
|----------|-----|---------|--------------|-------|
| Soldiers | `soldiers` | Iron Age | Sum of military buildings' soldier caps | Produced by military buildings when workers are assigned; spent on campaigns |

Soldiers are a true stockpiled resource (the 26th), produced and stored by your military buildings. See [Soldiers](#soldiers) below and [Military](military.md) for the full mechanics.

---

## Food

Food is the most critical resource. Every worker in every domain eats food each tick, and the amount scales with their class tier (base food cost × 1.12^tier). If food hits zero, one worker dies every 5 ticks until food is restored.

**How to increase food:**
- Assign food-domain workers to Food lineage buildings (`assign gathering_camp 5`)
- Build more Food lineage buildings to increase total worker capacity
- Research Fire Mastery, Animal Husbandry, Agriculture and related techs for extra food per tick

**Watch the food rate** in the Economy panel. Keep the rate (`+N/t`) positive before recruiting new workers into any domain.

---

## Knowledge

Knowledge pays for all research. Technologies cost knowledge to unlock, and your knowledge storage sets how much you can bank.

**How to increase knowledge:**
- Assign knowledge-domain workers to Knowledge lineage buildings
- Build Knowledge lineage buildings (Story Circle → Library → University → ...)
- Knowledge storage starts at 30, so build Knowledge buildings early to raise it

Each Knowledge building has its own fully staffed rate: 0.2 knowledge/tick for a Story Circle, 0.6 for an Elders' Hall, 2.0 for a Scriptorium, 1.6 for an Agora and 3.2 for a Library. From the Medieval to the Colonial Age buildings also cost knowledge, so it counts as a construction resource there and its producers follow the [Payback Rule](buildings.md#how-production-rates-are-set) (a Monastery Library makes 78.3/tick, a University 208). See [Knowledge](knowledge.md) for the full lineage.

---

## Faith

Faith accumulates from Faith lineage buildings and faith-domain workers. It does not drain on its own. It goes down only when something spends or removes it: the Political Instability epoch event (removes 60% of your faith), a few random events (Heresy, Workers' Uprising, Industrial Blight), Appeasing a [harbinger](harbinger.md), building the Sistine Chapel, and Enduring a catastrophe. The Renaissance Age asks for faith but only checks it; advancing does not spend it. The faith resource unlocks at the Primitive Age. Early faith workers (Devotee, Believer, Worshipper, Celebrant, Initiate) exist from the Primitive through the Classical Age with lower food costs. The formal, costlier Faith domain tier (Acolyte, base 2.0 food/tick) begins at the Medieval Age.

Your faith level as a **percentage of your storage** sets the epoch event roll odds:

| Faith % of storage | Epoch Roll (Good Chance) |
|----------------|--------------------------|
| Less than 25% | 40% |
| 25% to 75% | 50% |
| More than 75% | 60% |

**Advice:** Assign Faith workers (Acolytes and their successors) well before you expect an epoch transition. Faith lineage buildings produce a little on their own, but workers speed it up a lot.

---

## Culture

Culture comes from the Culture/Arts lineage (Amphitheater onward). These buildings produce culture without workers, so you can build them while your workers stay on other domains.

Culture has two jobs. Its **fill % of storage** decides which tier of good epoch event you can get at an epoch transition: over 40% makes Major events eligible, and over 75% adds a 15% chance at the Legendary one (see [Epochs](epochs.md)). And culture is **spent**, on the things below. Prestige resets culture along with every other resource.

### Culture sinks

**Cultural Monuments** are four one-off buildings (one of each) built with the normal `build <key>` command. Each costs a large lump of culture plus other materials and gives a small **permanent** all production bonus once built.

| Monument | Age | Culture Cost | Permanent Bonus |
|----------|-----|--------------|-----------------|
| Cultural Obelisk | Classical | 710 | +1% all production |
| Grand Amphitheater | Medieval | 7.1K | +2% all production |
| Eternal Library | Industrial | 140K | +3% all production |
| Monument of Ages | Modern | 7.1M | +5% all production |

**The `festival` command** spends a lump of culture (2K, or 5% of your culture storage if that is more) for **+20% all production for 150 ticks**, then goes on a **300-tick cooldown**. See [Commands](commands.md).

Culture also pays for:

- **Smuggling runs** on the black market (Colonial Age on): a culture stake that may pay out a haul of the resource you pick. See [Trade & Diplomacy](trade.md).
- **Appease**, together with faith, when a [harbinger](harbinger.md) warns of a catastrophe.
- **Tribute**, together with gold, to end a war with a civilization.
- The **Sistine Chapel** wonder in the Renaissance Age.

---

## Ore Processing Chain

The metallurgy pipeline requires two lineages working together:

```
Geological Extraction buildings  →  raw ore  →  Metallurgy buildings  →  refined metal
     (masonry workers)                              (metallurgy workers)
```

| Raw Ore | Refined Output | Notes |
|---------|---------------|-------|
| Iron Ore | Iron / Steel | Iron Age ore; Metallurgy lineage starts in the Iron Age |
| Titanium Ore | Titanium | Space Age ore |
| Dark Matter Crystals | Dark Matter | Cyberpunk Age ore |

You need both the Geological Extraction buildings (to produce ore) and the Metallurgy buildings (to refine it), staffed with the right workers, to keep metal flowing. Ore with no smelters, or smelters with no ore, both produce nothing.

---

## Resource Storage

The 25 civilian resources share storage from the Storage lineage buildings. Each Storage tier raises storage for every one of them at once. The base storage is very small (10 to 50 units), so you will hit it early. Soldiers are the exception: their storage comes from your military buildings, not the Storage lineage (see [Soldiers](#soldiers)).

**Priority:** Build a new Storage lineage building as your first or second action when entering any new age.

Late-game resources (Plasma, Titanium, Dark Matter, Antimatter, Quantum Flux) have especially low base storage (10 to 30 units) and fill up almost at once without dedicated storage. See [Buildings](buildings.md) for the full Storage lineage progression.

---

## Production Rates

Every building rate in the game is the **fully staffed** rate shown in the building's description. For **construction resources** (anything the buildings of your current age cost) the rate follows the **Payback Rule**: fully staffed, a producer earns back the price of its first copy within the age's payback time. That time is a growing share of the age: under a minute in the Primitive Age, 22 minutes in the Iron Age, 1.9 hours in the Renaissance. Later ages repay more slowly because your older buildings keep producing too. The **flow resources** (food, faith, culture and soldiers) keep hand-set rates, and the requirements that ask for them are sized to match. Details and the full table are on [Buildings](buildings.md#how-production-rates-are-set).

---

## Buying at the Market

Once you own any trade-lineage building (a `market` or its successors), the market trades **any two construction resources of your current age** at their **price parity less a 20% fee**. Each age has a price level per resource (the typical first-copy price in that resource), and one unit buys `0.8 × (price level of what you buy) / (price level of what you sell)`. In the Space Age, for example, steel and titanium are priced 180T and 250T, so 1 steel buys 1.11 titanium and 1 titanium buys 0.576 steel. Any pair of construction resources works (steel → titanium in the Space Age, data → crypto in the Cyberpunk Age, gold → stone). Fixed pairs involving food, culture, faith and similar resources keep their listed rates. Repeated trades of one pair push its rate down (floor 50%), and a round trip always loses value, so trading never beats building. See [Trade & Diplomacy](trade.md#resource-exchange).

Some construction resources have **no producing building in certain ages**, and the market is where you get them. Producers you built in earlier ages keep running, but the current age has none to build:

- **Stone** from the Iron Age on
- **Iron** from the Renaissance Age on
- **Steel** from the Modern Age on
- **Titanium** from the Interstellar Age on (in the Space Age the Orbital Refinery makes it; earlier titanium smelters make only a trickle)
- **Crypto** in the Cyberpunk Age (apart from the Neon Citadel wonder and the Blockchain tech)

The same goes for wood in the Colonial Age, coal in the Electric Age and data in the Modern Age. Faith cannot be bought at the market at all.

---

## Gold

Gold pays for trade, diplomacy gifts, and many mid-to-late game buildings.

**How to increase gold:**
- Assign trade-domain workers to Trade lineage buildings (Market, Bank, Stock Exchange, ...)
- Complete trade routes that export surplus resources
- Research Currency, Mercantilism, and related technologies

---

## Soldiers

Soldiers are a stockpiled resource (the 26th) that unlocks at the **Iron Age**. Unlike civilian resources, soldiers are both **produced** and **stored** only by your military buildings.

**How soldiers are produced:** Each military building you own (War Camp, Barracks, and successors) produces soldiers every tick when military-domain workers are assigned to it, the same way food or wood production scales with assigned workers. A fully staffed military building produces roughly its soldier cap ÷ 50 soldiers per tick (minimum 0.1/tick).

**How soldiers are stored:** Your soldier storage is the **sum of every military building's soldier cap**. The Storage lineage adds nothing here; building more (and higher-tier) military buildings is the only way to raise it. A single War Camp holds 10 soldiers, a Barracks adds 20, and caps double per tier up the military lineage.

**How soldiers are spent:** Launching a campaign costs soldiers, taken from your stockpile at launch whether the campaign succeeds or fails. Scouting expeditions cost none. Enduring a catastrophe also cuts your soldiers along with every other stored resource.

See [Military](military.md) for production formulas, per-building soldier caps, and the full campaign and expedition table.
