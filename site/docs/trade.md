# Trade & Diplomacy

Two systems add to your economy beyond raw production: **resource exchange** (swap one resource for another on demand) and **trade routes** (automatic income every few ticks). On top of those, **diplomacy** lets you meet an **11-civilization roster** of other powers. Allied civilizations boost production of their specialty resource, peaceful ones lend you workers, and provoked ones can declare war. Trade routes raise opinion, and allies raise route income.

---

## Resource Exchange

An exchange is an instant, one-off swap between two resources. You need at least one trade building first: a `market` or anything later in the trade lineage (trading post, merchant quarter, guildhall, exchange, port, stock exchange, bank, ...). Upgraded buildings count too, so upgrading your markets keeps the exchange open. If you skipped the market in the Bronze Age, the Iron Age trading post costs only stone and iron, so you can always open the exchange (and it is the Iron Age's gold producer).

```
trade <give> <get> <amount to give>
trade list
plan trade <give> <get> [amount to get]
```

`trade list` shows every rate open to your age, including any market pressure penalty. `plan trade` queues an exchange in the [build plan](plan.md). Rates follow your age: see [Exchange Rates](#exchange-rates).

### Exchange Rates

Each pair has a **base rate**: how many units of the resource you get per unit you give, at zero market pressure. The rate of a pair depends on what the two resources are in your current age.

**Construction resources trade at parity.** A *construction resource* of an age is any resource that one of that age's buildings costs (wonders aside), except the flow resources: food, faith, culture and soldiers. Each age has a **price level** for each construction resource: the median first-copy price in that resource among the age's buildings. Any two construction resources of your current age trade at the ratio of their price levels, less a **20% fee**:

```
base rate (give A, get B) = price level of B ÷ price level of A × 0.8
```

That covers every pair of them, including pairs that were never on the list: steel for titanium in the Space Age, data for crypto in the Cyberpunk Age, gold for stone. Price levels change with each age, so these rates do too. Because of the fee a round trip always loses value (0.8 × 0.8 keeps 64%), so trading never beats building. The exchange is how you get the resources no building of your age makes: stone after the Bronze Age, iron after the Medieval Age, steel from the Modern Age on, titanium, crypto.

Examples from each age (base rates, before pressure; "wood → stone 0.48" means you get 0.48 stone per wood):

| Age | Examples |
|---|---|
| Bronze | wood → stone 0.48, stone → wood 1.33, gold → iron 0.792, iron → gold 0.808 |
| Iron | gold → stone 1.5, wood → gold 0.711, iron → wood 1.07 |
| Classical | iron → stone 3.67, gold → iron 0.533, wood → stone 1.96 |
| Medieval | knowledge → gold 3.2, gold → knowledge 0.2, iron → stone 3.93 |
| Renaissance | coal → steel 1.41, gold → steel 0.267, knowledge → steel 2 |
| Colonial | wood → iron 7.08, steel → gold 1.78, knowledge → steel 4.11 |
| Industrial | coal → steel 2.1, iron → steel 2.8, stone → gold 1.2 |
| Victorian | oil → steel 1.8, coal → oil 0.457, gold → steel 1.31 |
| Electric | oil → steel 4, electricity → gold 1.7, coal → electricity 0.64 |
| Atomic | uranium → steel 8, gold → uranium 0.109, electricity → iron 0.94 |
| Modern | data → steel 20.5, oil → electricity 3.67, gold → data 0.0963 |
| Information | data → gold 17.2, electricity → steel 1.33, steel → data 0.048 |
| Digital | data → steel 9.25, electricity → data 0.0985, steel → electricity 0.562 |
| Cyberpunk | data → crypto 4.42, steel → crypto 0.46, crypto → electricity 1.95 |
| Fusion | plasma → steel 3.2, electricity → plasma 0.267, steel → electricity 0.6 |
| Space | steel → titanium 1.11, plasma → titanium 1.67, titanium → steel 0.576 |
| Interstellar | electricity → titanium 6.86, dark\_matter → plasma 4, plasma → dark\_matter 0.16 |
| Galactic | plasma → titanium 16.9, antimatter → dark\_matter 4, dark\_matter → antimatter 0.16 |
| Quantum | titanium → quantum\_flux 0.198, quantum\_flux → antimatter 232, dark\_matter → antimatter 0.96 |
| Transcendent | dark\_matter → antimatter 0.96, antimatter → dark\_matter 0.667 |

The **Trade** panel (`trade list`) lists every pair open to you in your current age with its live rate.

**Listed pairs.** The market also keeps its original list of pairs, unlocked by age:

| Age | Available Pairs |
|---|---|
| Bronze | food↔wood, food↔stone, wood↔stone, gold→food/wood/stone |
| Iron | iron↔gold, iron→stone |
| Medieval | gold→knowledge, gold→culture, faith→culture |
| Colonial | gold↔coal |
| Industrial | steel→gold, oil→gold |
| Electric | electricity→gold |
| Modern | data↔gold |
| Cyberpunk | crypto↔gold |
| Space | dark\_matter→gold |
| Quantum | quantum\_flux→gold |

Where both sides of a listed pair are construction resources of your current age, it trades at parity like any other pair (gold → wood is 3.43 in the Bronze Age, 0.9 in the Iron Age). Otherwise it keeps a fixed rate. That is always the case for pairs with food, faith or culture, and for pairs with knowledge, coal, stone and the rest in ages where no building costs them. The fixed rates:

| You give | You get | Fixed Rate |
|---|---|---|
| gold | food | 50 |
| gold | wood | 40 |
| gold | stone | 30 |
| gold | culture | 3.0 |
| gold | coal | 10 |
| coal | gold | 0.08 |
| iron | gold | 2.0 |
| iron | stone | 3.0 |
| gold | knowledge | 5.0 |
| faith | culture | 2.0 |
| oil | gold | 3.0 |
| steel | gold | 5.0 |
| electricity | gold | 0.5 |
| data | gold | 5.0 |
| crypto | gold | 20.0 |
| dark\_matter | gold | 50.0 |
| quantum\_flux | gold | 100.0 |

### Market Pressure

Every exchange you make on a pair adds **supply pressure** to that pair, and pressure lowers the rate:

```
effective rate = base rate × (1 − pressure × 0.30)
```

Pressure tops out at 1.0, so repeated trading can cut a pair's rate by at most 30%.

Pressure **decays 2% per tick**, multiplicatively, so a pair you leave alone recovers fully on its own. More trade buildings help too: each trade-lineage building you own (market, port, bank, ...) shrinks the pressure a single trade adds (`+0.10 ÷ (1 + trade buildings × 0.20)`).

**When to exchange:** turn a surplus into something you're short on, or buy a resource you can't produce yet or that no building of your age produces. Don't hit the same pair over and over in quick succession, because you'll drive its rate down. Spread trades across different pairs, or wait a few ticks between repeat swaps on one pair.

---

## Trade Routes

Trade routes run in the background. Every few ticks a route takes a set of resources from you and gives you others in return. Routes don't suffer market pressure; they only need the required buildings and enough of what they take in stock.

### Commands

```
trade route list
trade route start <route>
trade route stop <route>
```

`trade route list` shows your active routes (with the approximate time left on the current cycle and the number of cycles completed) and the routes you could start (a green checkmark if you have the required building, a red X if not). `trade route start <route>` starts a route; it fails if the required building isn't built or you haven't reached the route's age. `trade route stop <route>` stops a route at once, mid-cycle. There's no limit on how many routes can run at once, so run them all.

A route **stops itself** if you fall below its required building count while it runs, and the log says so. Rebuild the building and `trade route start` it again.

If a cycle comes round and you don't have enough of what the route gives away, that cycle is skipped with no penalty, and the route tries again next cycle. Keep those stockpiles topped up.

### Trade Disruption (War & Embargo)

If you are **at war** with a civilization, or you have put it under **embargo**, every route that brings in that civilization's **specialty resource** is **disrupted**: it takes nothing and gives nothing while the conflict lasts. The route isn't stopped and its timer keeps running, so it **resumes by itself** once the conflict is over (end a war with tribute or by waiting it out; lift an embargo with `diplomacy neutral`).

On the Trade panel a disrupted route has a red ✖ and a note naming the blockaded resource, and a banner above the routes lists every blockaded resource. The log also notes each cycle a disrupted route misses.

Disruption follows the war and embargo state directly; there's nothing separate to track. Before you embargo or provoke the **gold** specialist (Merchant Guild) or the **culture** specialist (Artisan League), check which of your routes bring in those goods.

The Trade panel doesn't list civilizations. The one diplomacy effect it shows is an **Allied Bonuses** block under the routes while an ally is boosting a resource (see [Allied Bonuses](#allied-bonuses)). Its last line points you to the **Factions** panel (`factions`) for opinion and diplomacy actions.

### Full Trade Routes Reference

| Key | Name | Min Age | Required Building | You give (per cycle) | You get (per cycle) | Cycle (ticks) |
|---|---|---|---|---|---|---|
| `local_barter` | Local Barter | Bronze | Market ×1 | 10 food | 8 wood | 10 |
| `stone_trade` | Stone Trade | Iron | Market ×2 | 15 wood | 12 stone | 12 |
| `gold_caravan` | Gold Caravan | Classical | Market ×3 | 50 stone | 5 gold | 15 |
| `silk_road` | Silk Road | Medieval | Market ×2 | 30 gold | 80 culture | 20 |
| `mercantile_convoy` | Mercantile Convoy | Renaissance | Exchange ×1 | 300 stone + 200 wood | 90 gold | 16 |
| `spice_trade` | Spice Trade | Colonial | Port ×1 | 100 gold | 200 food + 50 culture | 18 |
| `colonial_exports` | Colonial Exports | Colonial | Port ×2 | 500 food | 150 gold | 15 |
| `triangular_trade` | Triangular Trade | Colonial | Harbor ×1 | 400 food + 60 gold | 120 culture + 80 knowledge | 18 |
| `tea_clippers` | Tea Clippers | Colonial | Harbor ×2 | 250 gold | 600 food + 90 culture | 20 |
| `coal_barges` | Coal Barges | Industrial | Harbor ×2 | 300 coal | 220 gold + 150 iron | 14 |
| `cotton_exchange` | Cotton Exchange | Industrial | Seaport ×1 | 400 gold | 200 culture + 150 knowledge | 16 |
| `steamship_line` | Steamship Line | Industrial | Seaport ×2 | 250 steel + 200 coal | 900 gold | 18 |
| `rail_freight` | Rail Freight | Industrial | Steam Works ×1 | 200 iron | 100 gold + 50 coal | 12 |
| `oil_pipeline` | Oil Pipeline | Victorian | Oil Derrick ×2 | 100 oil | 300 gold | 15 |
| `power_exchange` | Power Exchange | Electric | Power Station ×1 | 500 electricity | 200 gold | 10 |
| `data_trade` | Data Trade | Information | Server Farm ×1 | 100 data | 500 gold | 10 |
| `crypto_market` | Crypto Market | Cyberpunk | Black Market ×1 | 50 crypto | 1K gold | 8 |
| `fusion_export` | Fusion Export | Fusion | Fusion Reactor ×1 | 200 electricity | 1K gold | 12 |
| `warp_commerce` | Warp Commerce | Space | Warp Drive Plant ×1 | 500 gold | 200 dark matter | 15 |
| `stellar_exchange` | Stellar Freight | Galactic | Galactic Trade Hub ×1 | 100 dark matter | 2K gold | 20 |
| `quantum_trade` | Quantum Trade | Quantum | Reality Processor ×1 | 50 quantum flux | 5K gold | 10 |

A few routes appear in the list before you can build what they need. Cotton Exchange and Steamship Line need a Seaport (Modern Age), Rail Freight needs a Steam Works (Victorian Age), and Warp Commerce needs a Warp Drive Plant (Interstellar Age).

---

## Harbor lineage: trade-route income

Markets and banks (the **trade** lineage) make gold directly. **Harbors** make your **trade routes** pay more instead. Each harbor building you own adds a flat percentage to what **every active route** gives you, and the bonuses add up across tiers and copies. Harbors also produce a little gold themselves, so an unused harbor still earns something.

The harbor bonus adds to an ally's bonus: a route bringing in an ally's specialty, with a fleet of harbors, pays `base × (1 + harbor bonus + ally bonus)`.

| Tier | Key | Name | Min Age | Route Income Bonus | Workers |
|---|---|---|---|---|---|
| 0 | `harbor` | Harbor | Colonial | +5% | 4 |
| 1 | `harbor_authority` | Harbor Authority | Industrial | +10% | 5 |
| 2 | `seaport` | Seaport | Modern | +15% | 6 |
| 3 | `container_terminal` | Container Terminal | Information | +20% | 8 |
| 4 | `logistics_hub` | Logistics Hub | Digital | +25% | 10 |

Harbors use the **trade** worker domain, the same workers who staff markets and embassies, so a big harbor fleet competes with your markets for hands. Three routes (`triangular_trade`, `tea_clippers`, `coal_barges`) need harbors rather than ports.

---

## Black Market

From the **Colonial Age** you can make smuggling runs on the black market. A run spends a lump of **culture** for a chance at a large haul of a resource you choose. If it fails, the culture is gone and you get nothing.

```text
blackmarket              # show cost, odds and cooldown
blackmarket <resource>   # make a smuggling run for the chosen resource
trade black <resource>   # the same, through the trade command
```

A run costs `max(5K, 10% of your culture storage)` culture, so the price grows as you do. It has a **55% chance** to pay out. A win gives you the chosen resource worth **2.5×** the culture you staked, valued at that resource's gold exchange rate. The culture is spent up front, win or lose. After each run there's a cooldown of about 240 ticks (about 8 minutes).

Use it to turn a culture surplus into whatever you're short on, if you can live with the odds.

---

## Diplomacy: Civilization Encounters

The game has an **11-civilization roster**. You meet them by **sending missions**: an age only makes a civilization *eligible*, and a resolved **scouting expedition or campaign** is what turns it up. A late fallback means even a player who never explores meets everyone eventually. Each civilization has an **opinion** of you (-100 to +100), a **diplomatic status**, a **personality** and a **backstory**. Status decides whether it helps you, ignores you or works against you. Personality drives how its opinion drifts and whether it lends workers or goes to war.

### Commands

```
factions                    # open the Factions panel
diplomacy                   # the same panel, under its older name
diplomacy ally <civ>
diplomacy rival <civ>
diplomacy embargo <civ>
diplomacy gift <civ>
diplomacy neutral <civ>
diplomacy tribute <civ>     # pay to end a war
diplomacy raid <civ>        # raid their trade route: -20 opinion, and a provocation
diplomacy deals [civ]       # list trade deals (one civilization, or every one you have met)
diplomacy accept <civ> <n>  # take deal n
plan deal <civ> <n>         # take deal n once you can pay for it
```

### The Factions panel

`factions` opens the **Factions** panel, one screen for everything the other civilizations are doing to you. `diplomacy` and `dip` with no arguments open the same panel, and `diplomacy <action> <civ>` still performs an action directly. The panel is also in the sidebar Panels list, the prompt's completions and the in-game `help` panel.

From top to bottom it shows:

1. The title, how many civilizations you have met and how many are still undiscovered.
2. **Boons and setbacks**: every timed effect a civilization has on you right now. Encounters hand out timed **boons** (marked `✦`) and, when a run goes badly, timed **setbacks** (marked `⚠`). Each line names the civilization, the effect, its size (`+13% food`, `+8% all prod`, `+9% tick speed`) and the time left. The section header shows how full both pools are (`boons 2/5 · setbacks 1/3`), because both are hard limits: **five** boons and **three** setbacks at once (see [First Contact & Discovery](#first-contact-amp-discovery) for what happens at the limit). Workers on loan from another civilization are listed here too; they have no expiry clock, but they are a live effect all the same.
3. The **Geographic Society**, in one of three states: nothing built (a prompt to build one, Industrial Age); starved (a party is due but your stores can't outfit it); or running, with the number of Societies, their staffing and fill, the dispatch interval, and a countdown with a progress bar to the next party. See [Automatic dispatch](military.md#automatic-dispatch-the-geographic-society).
4. A card for each civilization you have met: name, personality, specialty, a **strength rating** of 1-5 stars, a line of backstory, the opinion bar, the status (color-coded) with any active trade bonus, how far it is to the next threshold (e.g. *+8 to friendly*, *ally-eligible: 500g*), a war banner, lent workers, and a line for any boon or setback that civilization is applying right now, so the card agrees with the section at the top. The card also lists the civilization's current **trade deals**, one numbered line per offer, worded from your side (`2. Sell: give 225K iron → get 342K gold   +16% vs market`). A price you can't pay yet shows in red, deals you've taken are dimmed, and the card says when the next set arrives. A civilization that won't trade says why instead (for example, because it is at war with you). See [Trade deals](#trade-deals).
5. A short list of civilizations you haven't met yet: one line each with the name, strength, the age you must reach, specialty and personality. It shows six at most, and the rest collapse into a `… N more` line.

Every duration on the panel (boon and setback time left, the Society's interval and countdown) shows as approximate wall-clock time, not ticks. See [Timers and durations](commands.md#timers-and-durations).

### Trade deals

Every civilization you have met offers a small, rotating set of **trade deals**. Nothing is written by hand for each civilization: its specialty, personality and strength, and its opinion of you, decide what it offers. The Factions panel shows each civilization's deals on its card. `diplomacy deals [civ]` lists them, `diplomacy accept <civ> <n>` takes one, and `plan deal <civ> <n>` queues one in the [build plan](plan.md).

Every deal is worded from your side, the same way on the panel, in `diplomacy deals`, in the plan and in the log: `<kind>: give <price> → get <goods>`, for example `Buy: give 876M coal → get 966K food`.

| Kind | You give | You get |
|---|---|---|
| **Buy** | one of your construction resources | the civilization's specialty, sized by the goods |
| **Sell** | one of the resources you hold most of (fullest store first), which the civilization wants | the civilization's specialty, a little more than a Buy pays |
| **Goodwill** | one of the resources you hold most of | **+5 opinion** instead of goods |
| **Rare** | one of your construction resources | a construction resource of the **next** age that this age's market doesn't sell, at a steep price (0.6 of its next-age parity) |

Next to each line the panel shows how much better than the market the deal pays (`+15% vs market`), or `not sold at the market` when the market doesn't trade the pair.

Where the market trades the pair, a deal pays **5% to 25% better than the market** (it trades at 0.84 to 1.0 of parity, where the market pays 0.8). Where the market doesn't trade the pair, the rate comes from the two resources' price levels (food and culture are valued by what the age's producers of them make). A deal never pays better than parity, so trading still never beats building. Deals ignore market pressure: each one is a fixed contract.

A civilization's personality sets how many deals it offers and which kinds:

| Personality | Deals |
|---|---|
| **peaceful** | 2 when neutral, 3 friendly, 4 allied; mostly Buy deals |
| **mercantile** | one more than peaceful, 2 points better rates, 20% bigger lots, and more Sell deals (it asks for your goods) |
| **aggressive** | one fewer than peaceful (at least one), 4 points worse rates, and more Goodwill deals |
| **isolationist** | one deal (two when allied): a **Rare** deal when the next age has goods for it to sell, its specialty otherwise |

Opinion matters too. Friendly civilizations (friendly status, or opinion 25+) and allies offer more deals, at better rates (0.88 of parity when neutral, 0.92 friendly, 0.96 allied, plus 0.03 on a Sell) and in bigger lots. A civilization **at war** with you, under your **embargo**, your **rival**, or with opinion **-50 or lower** offers nothing and won't honor the offers it already made until that changes. Taking a Buy, Sell or Rare deal adds +1 opinion and counts as a trade; a Goodwill deal adds +5. Deals can raise opinion only up to **50**: they can bring a civilization to the edge of an alliance but not past it, and they can't end a war (no one at war trades with you; tribute or waiting it out still ends it).

A deal moves about 1.5 median building prices of the age (×0.9 plus 0.1 per point of the civilization's strength, ×1.25 friendly, ×1.5 allied, ×1.2 mercantile, ×2 for a Rare deal, and a roll between ×0.75 and ×1.25). It's capped so the goods fit in half your storage and the price in 80% of it. Amounts are rounded to three figures, never in your favor.

Offers rotate after **an hour of play at 1x** (1,800 ticks) and when you advance an age. The timer only runs while the game does: offline catch-up doesn't advance it, so when you come back you find the offers you left, and a deal you planned is still there for the plan to take while you're away. The panel shows the time to the next set.

Some examples, at neutral opinion and at each civilization's own age:

- *Riverlands Tribes* (peaceful, food), Bronze Age: `Sell: give 2.1K stone → get 3.0K food` (+13% vs market), `Goodwill: give 727 iron → get +5 opinion`.
- *Merchant Guild* (mercantile, gold), Colonial Age: `Sell: give 2.2M knowledge → get 23M gold` (+16%), `Buy: give 5.4M wood → get 22M gold` (+12%), and a third.
- *Ironhold Clans* (aggressive, iron), Medieval Age: one deal, `Sell: give 84.5K knowledge → get 179K iron` (+8%).
- *Atomic Directorate* (isolationist, steel), Atomic Age: `Rare: give 25.7B electricity → get 3.3B oil`, oil being the Modern Age's goods.

### Personalities

Every civilization has one of four personalities, which sets how its opinion drifts and how it behaves toward you:

| Personality | Opinion drift | Behavior |
|---|---|---|
| **peaceful** | trends **up** over time | Lends you workers when its opinion is 40 or more (see Worker Lending) |
| **aggressive** | trends **down** over time | Can be provoked into **war** when it strongly dislikes you |
| **mercantile** | rises while any of your trade routes is running, cools when none are | Rewards active trade routes |
| **isolationist** | trends toward **neutral** (0) | Slow to befriend, slow to anger |

Drift is gradual (1 point every 25 ticks) and stays within -100 to +100. It runs alongside the rival/embargo drain and the natural drift toward zero.

### Worker Lending

A peaceful civilization with opinion 40 or more now and then **lends you** 3 to 6 workers, announced in the log. Lent workers join your pool at once (they may take you over your housing for a while) and go home after 200 ticks. If the civilization's opinion is **above 80**, the loan is **permanent**: the workers stay. A civilization lends one batch at a time, and loans show on the Factions panel as *↳ N workers on loan*.

### War & Peace

A civilization declares **war** only when **both** conditions hold: its opinion is **below -75**, *and* you have provoked it twice. **Raiding its trade route** (`diplomacy raid`) is one provocation and **embargoing it** is another, so a raid plus an embargo, or two embargoes, while it's deeply hostile starts a war. Anger alone never starts a war, and provocations while you're on good terms don't either.

While at war, the civilization **raids** you every 40 ticks and takes 50 × its strength (1-5) of its specialty resource. War is purely a matter of these raids; there's no tactical combat. Each raid is logged with what you lost and a short account of it.

There are two ways to make **peace**:

1. **Tribute.** `diplomacy tribute <civ>` pays 300 gold and 50 culture per point of the civilization's strength and ends the war at once. Opinion rises by 25, but never above 0.
2. **Wait it out.** A war ends by itself after 300 ticks without a new provocation. Stop provoking them and it burns out.

`dip` works in place of `diplomacy` everywhere.

### Opinion and Status

| Status | Meaning |
|---|---|
| `neutral` | The default. No bonuses, no penalties. |
| `friendly` | Set when a gift or deal brings opinion to 25 or more. Friendly civilizations offer more deals at better rates, give bigger boons and gentler setbacks. |
| `allied` | Needs opinion 50+ and costs 500 gold. Adds the civilization's bonus to your whole production of its specialty resource and to route income of it (see [Allied Bonuses](#allied-bonuses)). |
| `rival` | Free to declare. Opinion drops an extra 5 every 50 ticks. No trade bonus and no deals. |
| `embargo` | Free to declare. The same opinion drain as rival, no trade bonus, no deals, and it counts as a provocation. Your routes that bring in its specialty are disrupted. |

**Natural drift:** every 100 ticks, opinion moves 1 point toward zero. For an ally this is a slow leak, so keep trading to make up for it.

**Rival/embargo drain:** -5 opinion every 50 ticks. A civilization at -100 stays there. One at +80 that you declare a rival will eventually fall below the ally threshold.

### Raising Opinion

| Source | Opinion gain | Notes |
|---|---|---|
| Trade routes | +1 per completed route cycle | Applies to every civilization you have met that isn't at war with you, all at once. More routes, faster gains. |
| Embassies | a steady trickle per tick | Spread across your non-hostile civilizations (see Embassy Buildings below). |
| Gifts | +15 per gift | `diplomacy gift <civ>` costs 200 gold. At 25+ a neutral civilization becomes friendly. |
| Deals | +1 per Buy, Sell or Rare deal; +5 per Goodwill deal | Only up to opinion 50. |
| Personality | +1 every 25 ticks | Peaceful civilizations, and mercantile ones while a route is running. |

Once a civilization is at 50 or more, spend 500 gold to ally with it. Going back to neutral (`diplomacy neutral`) is free but doesn't refund the 500 gold.

### Embassy Buildings

Two diplomacy buildings turn workers into a steady source of opinion. Staff an embassy and, each tick, it raises opinion with every **non-hostile** civilization (neutral, friendly or allied; rivals and embargoed civilizations get nothing). Output follows the same `0.20 + 0.80 × fill` staffing curve as production buildings, so a fully staffed embassy runs at full rate and an empty one still trickles. Opinion tops out at +100.

| Building | Unlocks | Cost | Workers | Opinion Rate |
|---|---|---|---|---|
| **Embassy** | Colonial Age | gold + iron | 5 (trade domain) | +0.05 opinion / worker / tick |
| **Grand Embassy** | Industrial Age | gold + steel | 8 (trade domain) | +0.10 opinion / worker / tick (2× the Embassy) |

Embassies use the **trade** worker domain, the same workers who staff markets, so every worker in an embassy is one fewer in a market. The gain each tick is split across all the non-hostile civilizations you have met, so embassies matter most once several are in play (Industrial Age onward).

### Allied Bonuses

An ally multiplies your production of its specialty resource:

```
production rate = normal rate × (1 + ally bonus)
route income    = base income × (1 + harbor bonus + ally bonus)
```

The bonus applies to your whole per-tick rate of that resource (buildings, workers and everything else), and to that resource on every trade route that brings it in. Two allies with the same specialty would add their bonuses together.

The bonus also shows in the **Active Multipliers** section of the Stats panel as a `Diplomacy` line on the affected resource, so you can see which of your rates an alliance is raising. The figure there is the same `1 + ally bonus` the game applies.

The Trade panel lists the same bonuses under **Allied Bonuses**, one line per ally (for example *Merchant Guild: +20% gold*). An ally that is at war with you gives nothing and is left off the list.

| Civilization | Specialty | Allied Bonus |
|---|---|---|
| Riverlands Tribes | food | +15% |
| Ironhold Clans | iron | +20% |
| Merchant Guild | gold | +20% |
| Artisan League | culture | +15% |
| Atomic Directorate | steel | +20% |
| Tech Consortium | data | +20% |
| Shadow Syndicate | crypto | +25% |
| Plasma Nomads | plasma | +22% |
| Stellar Federation | dark\_matter | +20% |
| Void Reavers | antimatter | +28% |
| Quantum Collective | quantum\_flux | +30% |

### First Contact & Discovery

Three rules decide when you meet a civilization. Reaching its first age makes it *eligible*; it doesn't meet you on its own. Sending missions is what finds it: whenever a **scouting expedition** or **campaign** resolves, the game rolls a chance to **encounter** a civilization. An encounter discovers a new eligible civilization (first contact) or, once you know everyone within reach, meets a known one again. Scouting finds civilizations more often than campaigns, and success more often than failure (see [Military & Expeditions](military.md#civilization-encounters) for the odds). Finally, if you never send a mission, each civilization is discovered anyway about **two ages after** its first age, far later than an explorer would meet it.

On first contact the log introduces the civilization's name, personality and backstory. Until then it appears only in the not-yet-met list on the Factions panel, and you can't deal with it. The founding civilizations (Riverlands Tribes, Ironhold Clans) appear early; the rest turn up across the eras up to the Cosmic Era.

**Encounter boons.** An encounter, whether first contact or a repeat meeting, can also grant a **boon**, rolled from a shared catalog: a timed boost to one resource (often the civilization's **specialty**) or to knowledge, an all-production or tick-speed surge, an instant lump of resources, or a gang of temporary workers. The roll depends on the civilization's character. A peaceful civilization leans toward gentle production and knowledge gifts, an aggressive one toward tick speed and spoils, a mercantile one toward caches of gold, and an isolationist one toward rare large hoards. It also depends on the civilization's **opinion** of you: the higher it is, the bigger the boon, and an **ally** can give you the rare tier that neutral civilizations never offer. Stronger civilizations give bigger gifts. Instant gifts such as a supply caravan or a lost vault also **scale with your age**, so a crate of goods that was a fortune in the Bronze Age is still a fortune in the Quantum Age. The log names the civilization and the reward.

**You can hold five boons at once.** Only *timed* rewards take a slot; a gift used on arrival holds nothing, so the limit only ever turns away timed rewards. While all five slots are full, an encounter that rolls a **timed** reward comes back empty (the envoys are thanked and sent home with their crates unopened), but one that rolls an **instant lump of resources or a gang of temporary workers still delivers it**, since those need no slot. A full set of boons costs you buffs, not goods.

Each timed boon lasts **750-3000 ticks** (about **25 minutes to 1h 40m** at 1x; the panel shows the time left), so slots free up steadily. A busy explorer gets a reward from about four encounters in five, has at least one boon running almost all the time, and has all five slots full only **12-18%** of the time.

**Encounters can go badly.** A **setback** takes the place of a boon whenever the mission **failed**, on about **one in three** meetings with a civilization you are **at war** with, and on about **one in four** timed rewards turned away because all five boon slots are full. A wartime meeting that doesn't turn violent is a **standoff**: the two parties see each other and withdraw, and you come home with nothing but the sighting. A civilization you are fighting never gives you a gift. Setbacks come from their own table: a handful of workers lost on the way home (to fever or bad water), part of one resource stockpile spoiled, stolen or written off, a temporary drop in one resource's output, or a production dip across your civilization while word of the expedition spreads. Setbacks get worse with the civilization's **strength** and milder with its **opinion** of you, so an ally's bad news is gentler than a rival's, and a war with a strong civilization is the worst case. They are limited: at most three timed setbacks run at once (against five boon slots), a setback expires sooner than a boon of the same size, and a spoilage takes part of a store, never all of it.

Encounters are one more reason to keep missions running after you've met everyone, with the catch that a failed run brings back a bill and a run that resolves with all five boon slots full brings back goods rather than buffs. See [Military & Expeditions](military.md#civilization-encounters).

---

## Full Civilization Reference

| Key | Name | Eligible From | Personality | Strength | Specialty | Allied Bonus |
|---|---|---|---|---|---|---|
| `riverlands_tribes` | Riverlands Tribes | Bronze Age | peaceful | 1 | food | +15% food |
| `ironhold_clans` | Ironhold Clans | Medieval Age | aggressive | 3 | iron | +20% iron |
| `merchant_guild` | Merchant Guild | Colonial Age | mercantile | 2 | gold | +20% gold |
| `artisan_league` | Artisan League | Industrial Age | peaceful | 1 | culture | +15% culture |
| `atomic_directorate` | Atomic Directorate | Atomic Age | isolationist | 4 | steel | +20% steel |
| `tech_consortium` | Tech Consortium | Information Age | mercantile | 2 | data | +20% data |
| `shadow_syndicate` | Shadow Syndicate | Cyberpunk Age | aggressive | 3 | crypto | +25% crypto |
| `plasma_nomads` | Plasma Nomads | Fusion Age | peaceful | 2 | plasma | +22% plasma |
| `stellar_federation` | Stellar Federation | Space Age | isolationist | 4 | dark\_matter | +20% dark matter |
| `void_reavers` | Void Reavers | Galactic Age | aggressive | 5 | antimatter | +28% antimatter |
| `quantum_collective` | Quantum Collective | Quantum Age | isolationist | 5 | quantum\_flux | +30% quantum flux |

The peaceful civilizations (Riverlands Tribes, Plasma Nomads, Artisan League) are the ones that lend workers; keep their opinion above 80 for permanent loans. The aggressive ones (Ironhold Clans, Shadow Syndicate, Void Reavers) drift hostile and will declare war if you provoke them while they dislike you; the Void Reavers (strength 5) raid hardest. The isolationists (Atomic Directorate, Stellar Federation, Quantum Collective) stay near neutral, hard to befriend and hard to anger. The mercantile ones (Merchant Guild, Tech Consortium) warm up while you keep trade routes running.

---

## Strategy

### Trade routes: priority order

**Early (Bronze to Medieval):** start `local_barter` as soon as you build a market; it runs indefinitely and costs almost nothing. Add `stone_trade` and `gold_caravan` once you have 2 and 3 markets. `silk_road` in the Medieval Age is the best culture-per-tick route for most of that era; put it ahead of raw gold unless you have plenty.

**Mid-game (Colonial to Industrial):** `spice_trade` and `colonial_exports` both need a port, so build one early in the Colonial Age. They pull in opposite directions: spice trade gives you food and culture for gold, and colonial exports turn food into gold. Run both once you have 2 ports. `rail_freight` pays well but eats iron, so only start it if you have iron to spare.

**Late-game (Electric onward):** each later route pays much more gold. Run all of them as fast as you can build what they need. `crypto_market` has the shortest cycle (8 ticks) and pays 1K gold, the best gold per tick until `quantum_trade` comes online.

### Diplomacy: when to ally

Ally when you can afford the 500 gold and you depend on the civilization's specialty. For example, with `crypto_market` bringing in 1K gold every 8 ticks, allying with the Merchant Guild (+20% gold) adds 200 gold to every cycle, so the 500 gold pays for itself in three cycles. The same 20% also applies to all the rest of your gold production.

There's little point rushing an alliance before you produce or import that resource. Opinion climbs on its own from route cycles, so spend your gold on construction or research until then.

**Gifts** (200 gold, +15 opinion) are most useful when a civilization is just short of the 50 needed to ally. Starting from 0, three gifts bring it to 45; one Goodwill deal (+5) or five route cycles take it to 50. Four gifts (800 gold) take it straight to 60. You can only gift civilizations you have met.

### Market pressure

- Don't use the same pair back to back. Check `trade list` before a repeat trade: a `↓` marker with a percentage means you're paying the penalty.
- If you need a lot of one resource, give different resources for it (for example, iron for gold, then wood for gold) rather than selling iron over and over.
- Every trade-lineage building reduces how fast pressure builds. By the Colonial Age, five or so trade buildings (ports, exchanges, banks, ...) make pressure a minor issue for occasional exchanges.

### Routes and diplomacy together

Every completed route cycle raises opinion with every civilization you have met by 1, so five running routes add 5 opinion per round of cycles. More routes bring earlier alliances, alliances raise both production and route income of the ally's specialty, and the extra income pays for more routes. Plan routes and diplomacy together.

### End-game

With all 21 routes running and your civilizations allied, gold from routes alone is enormous. Resource exchange stays useful even then: it's how you get construction resources your age has no producer for (titanium, crypto, steel from the Modern Age on), and how you rebalance between the ones you do make.

---

## Tips & Common Mistakes

- A route needs its building. `trade route start <route>` refuses with an error if you don't have the required building at the required count. In `trade route list`, routes with a red ✗ are waiting on buildings.
- Market pressure fades by itself. If a rate looks low, wait a few ticks and it recovers.
- An embargo mostly hurts you. It cancels that civilization's ally bonus, stops its deals, disrupts your routes that bring in its specialty and counts as a provocation. If you don't want to invest in a civilization, leave it neutral.
- Routes skip a cycle, they don't fail, when you run short. If you don't have what a route gives away when its cycle comes round, that cycle is skipped and the route stays active. Top up the stockpile and it carries on next cycle.
- An alliance costs 500 gold each time. If you drop to neutral and want to ally again, you pay 500 gold again. Opinion stays when you go neutral, so the gold is all you need.
- Gifts add up fast. Six gifts (1.2K gold) take a civilization from 0 to 90 opinion, well past 50. If you need an ally quickly and have the gold, this is the fastest way.
- The Quantum Collective's +30% quantum flux is the largest ally bonus in the game. Gift it to 50 opinion as soon as you meet it and ally when you can afford to.
