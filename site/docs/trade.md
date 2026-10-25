# Trade

Trade routes and the black market move resources around your economy. **Trade routes** run in the background, trading a small fixed bundle of goods every few ticks, and every completed cycle raises the opinion of you held by every civilization you have met (except one at war with you). **Harbors** raise what the routes pay, and the **black market** turns culture into a resource of your choice, with a risk of one deal in six.

The game has no market where one resource swaps for another. Every material an age charges for comes from a building: see [The maker rule](buildings.md#the-maker-rule).

The other civilizations (meeting them, opinion, deals, alliances, war) have their own page: [Factions & Diplomacy](factions.md).

---

## Trade Routes

Trade routes run in the background. Every few ticks a route takes a set of resources from you and gives you others in return. They only need the required buildings and enough of what they take in stock.

<figure class="screen" data-screen="trade"><figcaption>The Trade routes panel, with Local Barter part way through a cycle.</figcaption></figure>

### Commands

```
trade
trade route list
trade route start <route>
trade route stop <route>
```

`trade` on its own opens the **Trade routes** panel: your active routes, the routes you could start, and any allied bonuses.

`trade route list` shows your active routes (with the approximate time left on the current cycle and the number of cycles completed) and the routes you could start (a green checkmark if you have the required building, a red X if not). `trade route start <route>` starts a route; it fails if the required building isn't built or you haven't reached the route's age. `trade route stop <route>` stops a route at once, mid-cycle. There's no limit on how many routes can run at once.

A route **stops itself** if you fall below its required building count while it runs, and the log says so. Rebuild the building and `trade route start` it again.

If a cycle comes round and you don't have enough of what the route gives away, that cycle is skipped with no penalty (the log says so once per shortage), and the route tries again next cycle.

Stock has no limit, so a run brings in all of its import however much you hold. The one thing that stops a run is not having what it gives away: then the cycle is skipped, silently, and the route tries again next cycle. Only the amounts that actually moved count toward the route's totals.

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
| `triangular_trade` | Triangular Trade | Colonial | Harbor ×1 | 400 food + 60 gold | 120 culture + research progress | 47 |
| `tea_clippers` | Tea Clippers | Colonial | Harbor ×2 | 250 gold | 600 food + 90 culture | 52 |
| `coal_barges` | Coal Barges | Industrial | Harbor ×2 | 300 coal | 220 gold + 150 iron | 36 |
| `cotton_exchange` | Cotton Exchange | Industrial | Stock Exchange ×1 | 400 gold | 200 culture + research progress | 42 |
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

**Techs and routes.** Trade routes wait for **The Wheel**, a Bronze Age tech (it needs Woodworking): until it is researched `trade route start` is refused and names it, and the Trade panel says so. Boatbuilding, Containerization and Asteroid Claims each make every run bring in 10% more of what the route lists (33% with all three); a harbor's and an ally's share are added to that, each as its own share of the listed amount. Road Building, Railroads and Automated Logistics each make every route run 15% faster, and Wormholes 20% (with all four a route's next run takes 49% of the ticks listed). And **Rail Freight waits for Railroads**: `trade route start rail_freight` is refused until the tech is researched, with its name in the refusal. **Warp Commerce waits for Interstellar Trade** in the same way, an Interstellar Age tech that needs Asteroid Claims and Warp Drive. A game that was already running either route when its rule arrived keeps it for the rest of that run.

---

## Harbor lineage: trade-route income

Markets and banks (the **trade** lineage) make gold directly. **Harbors** (four of them wait for a tech of their age: the Harbor for Mercantilism, the Seaport for Containerization, the Container Terminal for E-commerce and the Logistics Hub for Automated Logistics) add a flat percentage to what **every active route** gives you, and the bonuses add up across tiers and copies. Harbors also produce gold themselves, which usually earns far more than the route bonus, because that bonus is a percentage of small fixed amounts (see [What routes are worth](#what-routes-are-worth)).

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

From the **Colonial Age**, once **Mercantilism** is researched, you can make smuggling runs on the black market. A run spends a lump of **culture** for the same worth in a resource you choose. Five runs in six go through. The sixth time the smugglers vanish with the culture and you get nothing.

```text
blackmarket              # show cost, odds and cooldown
blackmarket <resource>   # make a smuggling run for the chosen resource
trade black <resource>   # the same, through the trade command
```

`bm` works in place of `blackmarket`.

A run costs the larger of 5K culture and a number of hours of your own culture income: about 200 in the Colonial Age, about 166 in the Industrial Age, and 200 in every age from the Victorian Age on. The price grows as you do, and a town that makes twice the culture pays twice. A run goes through **five times in six**. When it does, it pays the same worth whichever resource you take it in: the minutes of culture income you staked, as the same minutes of your own income of the resource you chose. So a run that cost 200 hours of culture income pays 200 hours of your income of iron, of oil or of gold alike. One time in six the smugglers vanish with the culture and pay nothing. The smugglers deal in gold and nine resources (iron, coal, steel, oil, electricity, data, crypto, dark matter and quantum flux), and only in those your age already makes; ask for anything else and they refuse, and nothing is spent. Your income of the resource counts between a quarter and four times what a moderate town of the age makes of it, so a town that makes none is still paid and one that makes a great deal is not paid without limit. The culture is spent up front, whether the run goes through or not. After each run there's a cooldown of 624 ticks (about 21 minutes).

Use it to turn a culture surplus into whatever you're short on, if you can live with the one run in six that is lost.

---

## Strategy

### What routes are worth

Route payouts are fixed amounts, set per route, and they don't grow with the age. They were sized for a much smaller economy: Colonial Exports pays 150 gold a cycle in an age where a Port costs about 30M gold, and Quantum Trade pays 5K gold in an age where no building costs gold at all (none has since the Digital Age). Harbor and ally bonuses add a percentage to those amounts, which keeps them small. A route's goods are a small help in the Bronze and Iron Ages and close to nothing after that, so don't build for them.

What a route pays all run long is **opinion**. Each completed cycle adds +1 opinion with every civilization you have met that isn't at war with you, all at once. Local Barter, which costs 10 food a cycle, adds about 70 opinion an hour with each of them on its own, more than any personality drift takes away. Opinion is what opens better deals, alliances and worker loans (see [Factions & Diplomacy](factions.md#opinion-and-status)), and mercantile civilizations warm to you just for having a route running.

So:

- Start `local_barter` as soon as you have a market and leave it running. Add the cheap routes of each age as you meet their buildings; a cheap route raises opinion as fast as an expensive one with the same cycle.
- Prefer routes whose goods you can spare. A route that eats a resource you're short on costs more than the opinion is worth.
- More routes means more cycles, and every cycle counts with every civilization you know. Once opinion sits near +100 with everyone, more routes add nothing.

---

## Tips & Common Mistakes

- A route needs its building. `trade route start <route>` refuses with an error if you don't have the required building at the required count. In `trade route list`, routes with a red ✗ are waiting on buildings.
- An embargo mostly hurts you. It cancels that civilization's ally bonus, stops its deals, disrupts your routes that bring in its specialty and counts as a provocation. See [Factions & Diplomacy](factions.md#war-amp-peace).
- Routes skip a cycle, they don't fail, when you run short. If you don't have what a route gives away when its cycle comes round, that cycle is skipped and the route stays active. Top up the stockpile and it carries on next cycle.
