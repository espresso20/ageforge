# Trade

The market and the routes move resources around your economy. The **resource exchange** swaps one resource for another at your age's prices, on demand. **Trade routes** run in the background, swapping a small fixed bundle of goods every few ticks, and every completed cycle raises the opinion of you held by every civilization you have met (except one at war with you). **Harbors** raise what the routes pay, and the **black market** gambles culture on a haul.

The other civilizations (meeting them, opinion, deals, alliances, war) have their own page: [Factions & Diplomacy](factions.md).

---

## Resource Exchange

An exchange is an instant, one-off swap between two resources. You need at least one trade building first: a `market` or anything later in the trade lineage (trading post, merchant quarter, guildhall, exchange, port, stock exchange, bank, ...). Upgraded buildings count too, so upgrading your markets keeps the exchange open. If you skipped the market in the Bronze Age, the Iron Age trading post costs only stone and iron, so you can always open the exchange.

```
trade <give> <get> <amount to give>
trade list
plan trade <give> <get> [amount to get]
```

- `trade` makes one exchange now. The amount counts what you **give**.
- `trade list` shows every rate open to your age, including any market pressure penalty. Bare `trade` opens the **Trade** panel, which shows the same rates and your routes.
- `plan trade` puts a **sell order** in the [build plan](plan.md): as `give` comes in, the plan sells it for `get` at the market rate. Here the amount counts what you **get**. With an amount, the order drops out once it has bought that much. With no amount it keeps selling until you remove it, and it pauses while the `get` store is full. It sells in lumps once the pair's rate has recovered from its last sale, so it keeps within about 1% of the market rate, and it only sells what the plan items above it leave free. It also drops out if the pair stops trading (an age advance can do that).

Rates follow your age: see [Exchange Rates](#exchange-rates).

### Exchange Rates

Each pair has a **base rate**: how many units of the resource you get per unit you give, at zero market pressure. The rate of a pair depends on what the two resources are in your current age.

**Construction resources trade at parity.** A *construction resource* of an age is any resource that one of that age's buildings costs (wonders aside), except the flow resources: food, faith, culture and soldiers. Each age has a **price level** for each construction resource: the median first-copy price in that resource among the age's buildings. Any two construction resources of your current age trade at the ratio of their price levels, less a **20% fee**:

```
base rate (give A, get B) = price level of B ÷ price level of A × 0.8
```

**Three techs lower the fee.** Currency and Banking take 3 points off each and Blockchain 2, so with all three the market keeps 12% and pays 10% more than the rates on this page (every market rate is scaled the same way, the listed pairs included). The fee never falls under 5%.

That covers every pair of them, including pairs that were never on the list: steel for titanium in the Space Age, data for crypto in the Cyberpunk Age, gold for stone. Price levels change with each age, so these rates do too. Because of the fee a round trip always loses value (0.8 × 0.8 keeps 64%), so trading never beats building. The exchange is how you get the resources no building of your age makes: stone after the Bronze Age, iron after the Medieval Age, steel from the Modern Age on, titanium, crypto.

Examples from each age (base rates, before pressure; "wood → stone 0.48" means you get 0.48 stone per wood):

| Age | Examples |
|---|---|
| Bronze | wood → stone 0.48, stone → wood 1.33, gold → iron 0.792, iron → gold 0.808 |
| Iron | gold → stone 1.6, wood → gold 0.667, iron → wood 1.07 |
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
| Space | steel → titanium 1.09, plasma → titanium 1.63, titanium → steel 0.588 |
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

Where both sides of a listed pair are construction resources of your current age, it trades at parity like any other pair (gold → wood is 3.43 in the Bronze Age, 0.96 in the Iron Age). Otherwise it keeps a fixed rate. That is always the case for pairs with food, faith or culture, and for pairs with wood, stone, iron, coal and the rest in ages where no building costs them. The fixed rates:

| You give | You get | Fixed Rate |
|---|---|---|
| food | wood | 1.0 |
| wood | food | 1.0 |
| food | stone | 0.8 |
| stone | food | 1.25 |
| wood | stone | 0.9 |
| stone | wood | 1.1 |
| gold | food | 50 |
| gold | wood | 40 |
| gold | stone | 30 |
| iron | gold | 2.0 |
| gold | iron | 0.4 |
| iron | stone | 3.0 |
| gold | knowledge | 5.0 |
| gold | culture | 3.0 |
| faith | culture | 2.0 |
| gold | coal | 10 |
| coal | gold | 0.08 |
| steel | gold | 5.0 |
| oil | gold | 3.0 |
| electricity | gold | 0.5 |
| data | gold | 5.0 |
| gold | data | 0.15 |
| crypto | gold | 20.0 |
| gold | crypto | 0.04 |
| dark\_matter | gold | 50.0 |
| quantum\_flux | gold | 100.0 |

### Market Pressure

Every exchange you make on a pair adds **supply pressure** to that pair, and pressure lowers the rate:

```
effective rate = base rate × (1 − pressure × 0.30)
```

Pressure tops out at 1.0, so repeated trading can cut a pair's rate by at most 30%.

Pressure **decays 2% per tick**, multiplicatively, so a pair you leave alone recovers fully on its own. More trade buildings help too: each trade-lineage building you own (market, port, bank, ...) shrinks the pressure a single trade adds (`+0.10 ÷ (1 + trade buildings × 0.20)`). Every trade adds the same step whatever its size, so one large trade costs less in pressure than many small ones.

**When to exchange:** turn a surplus into something you're short on, or buy a resource you can't produce yet or that no building of your age produces. Don't hit the same pair over and over in quick succession, because you'll drive its rate down. Spread trades across different pairs, wait a few ticks between repeat swaps on one pair, or let a `plan trade` order do the timing for you.

---

## Trade Routes

Trade routes run in the background. Every few ticks a route takes a set of resources from you and gives you others in return. Routes don't suffer market pressure; they only need the required buildings and enough of what they take in stock.

<figure class="screen" data-screen="trade"><figcaption>The Trade panel, scrolled to its routes: the last of the market rates, then Local Barter part way through a cycle.</figcaption></figure>

### Commands

```
trade route list
trade route start <route>
trade route stop <route>
```

`trade route list` shows your active routes (with the approximate time left on the current cycle and the number of cycles completed) and the routes you could start (a green checkmark if you have the required building, a red X if not). `trade route start <route>` starts a route; it fails if the required building isn't built or you haven't reached the route's age. `trade route stop <route>` stops a route at once, mid-cycle. There's no limit on how many routes can run at once.

A route **stops itself** if you fall below its required building count while it runs, and the log says so. Rebuild the building and `trade route start` it again.

If a cycle comes round and you don't have enough of what the route gives away, that cycle is skipped with no penalty (the log says so once per shortage), and the route tries again next cycle.

Every completed cycle also raises opinion by +1 with every civilization you have met that isn't at war with you. See [What routes are worth](#what-routes-are-worth).

### Trade Disruption (War & Embargo)

If you are **at war** with a civilization, or you have put it under **embargo**, every route that brings in that civilization's **specialty resource** is **disrupted**: it takes nothing and gives nothing while the conflict lasts. The route isn't stopped and its timer keeps running, so it **resumes by itself** once the conflict is over (end a war with tribute or by waiting it out; lift an embargo with `diplomacy neutral`). See [War & Peace](factions.md#war-amp-peace).

On the Trade panel a disrupted route has a red ✖ and a note naming the blockaded resource, and a banner above the routes lists every blockaded resource. The log also notes each cycle a disrupted route misses.

Disruption follows the war and embargo state directly; there's nothing separate to track. Before you embargo or provoke a civilization, check which of your routes bring in its specialty.

The Trade panel doesn't list civilizations. The one diplomacy effect it shows is an **Allied bonuses** block under the routes while an ally is boosting a resource (see [Allied Bonuses](factions.md#allied-bonuses)). Its last line points you to the **Factions** panel (`factions`) for opinion and deals.

### Full Trade Routes Reference

Every route opens in the Bronze Age or later. From the Bronze Age on, ages and their timers run 2.6 times their base length, and the cycles below already include that.

| Key | Name | Min Age | Required Building | You give (per cycle) | You get (per cycle) | Cycle (ticks) |
|---|---|---|---|---|---|---|
| `local_barter` | Local Barter | Bronze | Market ×1 | 10 food | 8 wood | 26 |
| `stone_trade` | Stone Trade | Iron | Market ×2 | 15 wood | 12 stone | 31 |
| `gold_caravan` | Gold Caravan | Classical | Market ×3 | 50 stone | 5 gold | 39 |
| `silk_road` | Silk Road | Medieval | Market ×2 | 30 gold | 80 culture | 52 |
| `mercantile_convoy` | Mercantile Convoy | Renaissance | Exchange ×1 | 300 stone + 200 wood | 90 gold | 42 |
| `spice_trade` | Spice Trade | Colonial | Port ×1 | 100 gold | 200 food + 50 culture | 47 |
| `colonial_exports` | Colonial Exports | Colonial | Port ×2 | 500 food | 150 gold | 39 |
| `triangular_trade` | Triangular Trade | Colonial | Harbor ×1 | 400 food + 60 gold | 120 culture + 80 knowledge | 47 |
| `tea_clippers` | Tea Clippers | Colonial | Harbor ×2 | 250 gold | 600 food + 90 culture | 52 |
| `coal_barges` | Coal Barges | Industrial | Harbor ×2 | 300 coal | 220 gold + 150 iron | 36 |
| `cotton_exchange` | Cotton Exchange | Industrial | Stock Exchange ×1 | 400 gold | 200 culture + 150 knowledge | 42 |
| `steamship_line` | Steamship Line | Industrial | Harbor Authority ×2 | 250 steel + 200 coal | 900 gold | 47 |
| `rail_freight` | Rail Freight | Industrial | Integrated Steelworks ×1 | 200 iron | 100 gold + 50 coal | 31 |
| `oil_pipeline` | Oil Pipeline | Victorian | Oil Derrick ×2 | 100 oil | 300 gold | 39 |
| `power_exchange` | Power Exchange | Electric | Power Station ×1 | 500 electricity | 200 gold | 26 |
| `data_trade` | Data Trade | Information | Server Farm ×1 | 100 data | 500 gold | 26 |
| `crypto_market` | Crypto Market | Cyberpunk | Black Market Hub ×1 | 50 crypto | 1K gold | 21 |
| `fusion_export` | Fusion Export | Fusion | Fusion Reactor ×1 | 200 electricity | 1K gold | 31 |
| `warp_commerce` | Warp Commerce | Interstellar | Warp Drive Plant ×1 | 500 gold | 200 dark matter | 39 |
| `stellar_exchange` | Stellar Freight | Galactic | Galactic Trade Hub ×1 | 100 dark matter | 2K gold | 52 |
| `quantum_trade` | Quantum Trade | Quantum | Reality Processor ×1 | 50 quantum flux | 5K gold | 26 |

What a route gives is boosted by your harbors and by an ally whose specialty it brings in: `amount × (1 + harbor bonus + ally bonus)`.

**Techs and routes.** Trade routes wait for **The Wheel**, a Bronze Age tech (it needs Woodworking): until it is researched `trade route start` is refused and names it, and the Trade panel says so. The market itself needs no tech. Boatbuilding makes every run bring in 10% more of what the route lists; a harbor's and an ally's share are added to that, each as its own share of the listed amount. Road Building and Railroads each make every route run 15% faster (28% with both: a route's next run takes 0.85 × 0.85 of the ticks listed). And **Rail Freight waits for Railroads**: `trade route start rail_freight` is refused until the tech is researched, with its name in the refusal. A game that was already running the route when this rule arrived keeps it for the rest of that run.

---

## Harbor lineage: trade-route income

Markets and banks (the **trade** lineage) make gold directly. **Harbors** add a flat percentage to what **every active route** gives you, and the bonuses add up across tiers and copies. Harbors also produce gold themselves, which usually earns far more than the route bonus, because that bonus is a percentage of small fixed amounts (see [What routes are worth](#what-routes-are-worth)).

The harbor bonus adds to an ally's bonus: a route bringing in an ally's specialty, with a fleet of harbors, pays `base × (1 + harbor bonus + ally bonus)`.

| Tier | Key | Name | Min Age | Route Income Bonus | Workers |
|---|---|---|---|---|---|
| 0 | `harbor` | Harbor | Colonial | +5% | 4 |
| 1 | `harbor_authority` | Harbor Authority | Industrial | +10% | 5 |
| 2 | `seaport` | Seaport | Modern | +15% | 6 |
| 3 | `container_terminal` | Container Terminal | Information | +20% | 8 |
| 4 | `logistics_hub` | Logistics Hub | Digital | +25% | 10 |

Harbors use the **trade** worker domain, the same workers who staff markets and embassies, so a big harbor fleet competes with your markets for hands. Three routes (`triangular_trade`, `tea_clippers`, `coal_barges`) need harbors and one (`steamship_line`) needs Harbor Authorities, rather than ports.

---

## Black Market

From the **Colonial Age**, once **Mercantilism** is researched, you can make smuggling runs on the black market. A run spends a lump of **culture** for a chance at a large haul of a resource you choose. If it fails, the culture is gone and you get nothing.

```text
blackmarket              # show cost, odds and cooldown
blackmarket <resource>   # make a smuggling run for the chosen resource
trade black <resource>   # the same, through the trade command
```

`bm` works in place of `blackmarket`.

A run costs `max(5K, 10% of your culture storage)` culture, so the price grows as you do. It has a **55% chance** to pay out. A win pays **2.5×** the culture you staked, counted as gold, in the resource you chose, converted at that resource's fixed rate to gold from the table above. So the smugglers deal only in gold and the resources with a fixed price in gold: iron, coal, steel, oil, electricity, data, crypto, dark matter and quantum flux. For example, a 5K culture stake that wins pays 12.5K gold, 6,250 iron or 156K coal. A haul bigger than the room in that store is cut to fit. The culture is spent up front, win or lose. After each run there's a cooldown of 624 ticks (about 21 minutes).

Use it to turn a culture surplus into whatever you're short on, if you can live with the odds.

---

## Strategy

### What routes are worth

Route payouts are fixed amounts, set per route, and they don't grow with the age. They were sized for a much smaller economy: Colonial Exports pays 150 gold a cycle in an age where a Port costs about 23M gold, and Quantum Trade pays 5K gold in an age where no building costs gold at all (none has since the Digital Age). Harbor and ally bonuses add a percentage to those amounts, which keeps them small. A route's goods are a small help in the Bronze and Iron Ages and close to nothing after that, so don't build for them.

What a route pays all run long is **opinion**. Each completed cycle adds +1 opinion with every civilization you have met that isn't at war with you, all at once. Local Barter, which costs 10 food a cycle, adds about 70 opinion an hour with each of them on its own, more than any personality drift takes away. Opinion is what opens better deals, alliances and worker loans (see [Factions & Diplomacy](factions.md#opinion-and-status)), and mercantile civilizations warm to you just for having a route running.

So:

- Start `local_barter` as soon as you have a market and leave it running. Add the cheap routes of each age as you meet their buildings; a cheap route raises opinion as fast as an expensive one with the same cycle.
- Prefer routes whose goods you can spare. A route that eats a resource you're short on costs more than the opinion is worth.
- More routes means more cycles, and every cycle counts with every civilization you know. Once opinion sits near +100 with everyone, more routes add nothing.

### Market pressure

- Don't use the same pair back to back. Check `trade list` before a repeat trade: a `↓` marker with a percentage means you're paying the penalty.
- If you need a lot of one resource, give different resources for it (for example, iron for gold, then wood for gold) rather than selling iron over and over, or make one large trade instead of several small ones.
- Every trade-lineage building reduces how fast pressure builds. By the Colonial Age, five or so trade buildings (ports, exchanges, banks, ...) make pressure a minor issue for occasional exchanges.
- For a steady need (stone after the Bronze Age, say), a `plan trade` order sells as the resource comes in and waits out the pressure for you, including while you're away.

### End-game

Resource exchange stays useful to the end: it's how you get construction resources your age has no producer for (titanium, crypto, steel from the Modern Age on), and how you rebalance between the ones you do make. A [faction deal](factions.md#trade-deals) for the same pair pays 5% to 25% better than the market, so check the Factions panel before a big exchange.

---

## Tips & Common Mistakes

- A route needs its building. `trade route start <route>` refuses with an error if you don't have the required building at the required count. In `trade route list`, routes with a red ✗ are waiting on buildings.
- Market pressure fades by itself. If a rate looks low, wait a few ticks and it recovers.
- An embargo mostly hurts you. It cancels that civilization's ally bonus, stops its deals, disrupts your routes that bring in its specialty and counts as a provocation. See [Factions & Diplomacy](factions.md#war-amp-peace).
- Routes skip a cycle, they don't fail, when you run short. If you don't have what a route gives away when its cycle comes round, that cycle is skipped and the route stays active. Top up the stockpile and it carries on next cycle.
- `trade` and `plan trade` count the amount from different sides: `trade wood gold 500` gives 500 wood, `plan trade wood gold 500` buys 500 gold.
