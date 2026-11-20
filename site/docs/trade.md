# Trade

Trade routes and the Black Market move resources around your economy. There are 21 trade routes. A running route is a flow in your income: it takes some of a resource and brings another every tick, and the flow grows with the buildings you have for it, by one gate's worth for each doubling of them. Each minute a route runs also raises the opinion of you held by every civilization you have met (except one at war with you). Harbors raise what the routes bring. The Black Market turns culture into a resource of your choice, with a risk of one deal in six.

The game has no market where one resource swaps for another. Every material an age charges for comes from a building (see [The maker rule](buildings.md#the-maker-rule)).

The other civilizations (meeting them, opinion, deals, alliances, war) have their own page: [Factions & Diplomacy](factions.md).

---

## Trade routes

A running route takes some of one resource from you every tick and brings you another. Its flow shows in the Resources box as a line of its own, as "trade". A route needs its required building. At the building count its listing asks, it sends 10% of what the town makes of its export (before any trade) and brings in the import, worth 10% more than what it sent; with more of the building it moves more, but with diminishing returns: each doubling of the line adds one more gate's flow (flow x1 at the gate, x2 at twice the buildings, x3 at four times). A route only sends what the town can spare: it never takes the stock below what the town itself uses, so a route that exports food leaves what the people eat.

<figure class="screen" data-screen="trade"><figcaption>The Trade panel in a window too small for the board: a card for each route, with Local Barter running at its full flow.</figcaption></figure>

### Commands

```
trade
trade route list
trade route start <route|all>
trade route stop <route>
```

`trade` on its own opens the Trade panel, a board that fills the screen. Your open routes stand in a column at the left, each with its state, what it trades, what it takes and brings each second, its share of your own income of the import (it can pass 100% when the import is worth far more than what the town makes of it), and a percentage when it is held back for want of stock. A running route is bright, a suspended one red, a route you could start plain and a locked one faint. A short route says what it still needs. A line at the top gives the total trade in and out each second. Beside the routes a goods table lists what the running routes take out and bring in each minute, anything that bears on trade (a lock, a blockade, an ally's bonus) and the way to the Factions panel. Below the table is the list of every open offer from the civilizations you have met, grouped under each one with an Accept button on the offer. A civilization that won't trade says why instead. A long list scrolls with `PgUp` and `PgDn`, with a line at each end counting the offers out of sight and one saying how many you can pay now. `↑` and `↓` move from route to route and on through the offers. `Enter` runs the picked entry's command: `stop` (pause) or `start` (resume) on a route, accept on an offer, which first shows what the lot would add to the town's stock. A click picks an entry, and a click on its button runs the command. What the command did shows on the panel. A window too small for the board shows the routes and the offers as a list of cards. `t` works in place of `trade`.

`trade route start all` starts every route you could start in one go. Routes that can't start are left alone.

`trade route list` shows your active routes (running at what share of their flow, or paused) and the routes you could start (a green checkmark if you have the required building, a red X if not). `trade route start <route>` starts a route. It fails if the required building isn't built or you haven't reached the route's age. `trade route stop <route>` (or `trade route pause <route>`) pauses a route: it stays on the list at no flow until you start it again, and a paused route is saved as paused. There's no limit on how many routes can run at once.

A route whose building count falls below its gate while it runs is held: it stays on the list and moves nothing, and the log says so. Rebuild the building and it runs again by itself.

Stock has no limit, so what a route brings in is never cut short by a full store. The one thing that slows a route is not having what it gives away. For each resource the routes export, add up what they would take in a tick and compare it with what the town can spare: what it makes of it less what it uses, plus the stock, all but a twentieth of the town's make that always stays. If the stock cannot cover it, every route exporting that resource runs at the same share of its flow, both what it takes and what it brings, so the stock is shared and never goes below nothing. The panel shows the share as a percentage and the log says so once when a route is slowed and once when it is back to full.

Every minute a route runs also raises opinion by +1 with every civilization you have met that isn't at war with you. See [What routes are worth](#what-routes-are-worth).

### Trade disruption (war & embargo)

If you are at war with a civilization, or you have put it under embargo, every route that brings in that civilization's specialty resource is disrupted. It takes nothing and gives nothing while the conflict lasts. The route isn't stopped, so it resumes by itself once the conflict is over. End a war with tribute or by waiting it out, and lift an embargo with `diplomacy neutral`. See [War & Peace](factions.md#war-amp-peace).

On the Trade routes panel a disrupted route's card is marked suspended, with a note naming the blockaded resource, and a warning above the cards lists every blockaded resource. The log also notes when a disrupted route is held.

Disruption follows the war and embargo state directly, so there's nothing separate to track. Before you embargo or provoke a civilization, check which of your routes bring in its specialty.

The panel lists each civilization only as the head of its offers. The one other diplomacy effect it shows is an Allied bonuses note while an ally is boosting a resource (see [Allied Bonuses](factions.md#allied-bonuses)). The Factions panel (`factions`) has opinion, status and every diplomacy verb.

### Full trade routes reference

Every route opens in the Bronze Age or later. A route at the building count it asks for sends 10% of what the town makes of its export (before any trade) and brings in goods worth 10% more than what it sent, by the same price levels deals use (a route with two imports splits that worth between them). Where an age has no price for a good, the route uses the price levels of the latest earlier age that has one for all of them; research a route brings is not valued yet and is not paid.

| Key | Name | Min Age | Required Building | You give | You get |
|---|---|---|---|---|---|
| `local_barter` | Local Barter | Bronze | Market ×1 | food | wood |
| `stone_trade` | Stone Trade | Iron | Market ×2 | wood | stone |
| `gold_caravan` | Gold Caravan | Classical | Market ×3 | stone | gold |
| `silk_road` | Silk Road | Medieval | Market ×2 | gold | culture |
| `mercantile_convoy` | Mercantile Convoy | Renaissance | Exchange ×1 | stone + wood | gold |
| `spice_trade` | Spice Trade | Colonial | Port ×1 | gold | food + culture |
| `colonial_exports` | Colonial Exports | Colonial | Port ×2 | food | gold |
| `triangular_trade` | Triangular Trade | Colonial | Harbor ×1 | food + gold | culture + research progress |
| `tea_clippers` | Tea Clippers | Colonial | Harbor ×2 | gold | food + culture |
| `coal_barges` | Coal Barges | Industrial | Harbor ×2 | coal | gold + iron |
| `cotton_exchange` | Cotton Exchange | Industrial | Stock Exchange ×1 | gold | culture + research progress |
| `steamship_line` | Steamship Line | Industrial | Harbor Authority ×2 | steel + coal | gold |
| `rail_freight` | Rail Freight | Industrial | Integrated Steelworks ×1 | iron | gold + coal |
| `oil_pipeline` | Oil Pipeline | Victorian | Oil Derrick ×2 | oil | gold |
| `power_exchange` | Power Exchange | Electric | Power Station ×1 | electricity | gold |
| `data_trade` | Data Trade | Information | Server Farm ×1 | data | gold |
| `crypto_market` | Crypto Market | Cyberpunk | Black Market Hub ×1 | crypto | gold |
| `fusion_export` | Fusion Export | Fusion | Fusion Reactor ×1 | electricity | gold |
| `warp_commerce` | Warp Commerce | Interstellar | Warp Drive Plant ×1 | gold | dark matter |
| `stellar_exchange` | Stellar Freight | Galactic | Galactic Trade Hub ×1 | dark matter | gold |
| `quantum_trade` | Quantum Trade | Quantum | Reality Processor ×1 | quantum flux | gold |

What a route gives is boosted by your harbors and by an ally whose specialty it brings in: what it brings × (1 + harbor bonus + ally bonus).

### Techs and routes

Trade routes wait for The Wheel, a Bronze Age tech (it needs Woodworking). Until it is researched, `trade route start` is refused and names it, and the Trade routes panel says so. Rail Freight waits for Railroads in the same way: `trade route start rail_freight` is refused until the tech is researched, with its name in the refusal. Warp Commerce waits for Interstellar Trade, an Interstellar Age tech that needs Asteroid Claims and Warp Drive. A game that was already running either route keeps it for the rest of that run.

Boatbuilding, Containerization and Asteroid Claims each make every route bring in 10% more (33% with all three). A harbor's and an ally's share are added to that, each as its own share of what it brings. Road Building, Railroads and Automated Logistics each make every route's flow 1/0.85 greater, and Wormholes 1/0.8. With all four, a route's flow is 1/0.49 of the listed one, a little over twice.

---

## Harbor lineage: trade-route income

Markets and banks (the trade lineage) make gold directly. Harbors add a flat percentage to what every active route gives you, and the bonuses add up across tiers and copies. Four of the five wait for a tech of their age: the Harbor for Mercantilism, the Seaport for Containerization, the Container Terminal for E-commerce and the Logistics Hub for Automated Logistics. Harbors also produce gold themselves, which usually earns far more than the route bonus, because that bonus is a percentage of small fixed amounts (see [What routes are worth](#what-routes-are-worth)).

The harbor bonus adds to an ally's bonus: a route bringing in an ally's specialty, with a fleet of harbors, pays `base × (1 + harbor bonus + ally bonus)`.

| Tier | Key | Name | Min Age | Route Income Bonus | Workers |
|---|---|---|---|---|---|
| 0 | `harbor` | Harbor | Colonial | +5% | 4 |
| 1 | `harbor_authority` | Harbor Authority | Industrial | +10% | 5 |
| 2 | `seaport` | Seaport | Modern | +15% | 6 |
| 3 | `container_terminal` | Container Terminal | Information | +20% | 8 |
| 4 | `logistics_hub` | Logistics Hub | Digital | +25% | 10 |

Harbors use the trade worker domain, the same workers who staff markets and embassies, so a big harbor fleet competes with your markets for hands. Three routes (`triangular_trade`, `tea_clippers`, `coal_barges`) need harbors and one (`steamship_line`) needs Harbor Authorities, rather than ports.

---

## Black Market

From the Colonial Age, once Mercantilism is researched, you can make smuggling runs on the Black Market. A run spends a lump of culture for the same worth in a resource you choose. Five runs in six go through. The sixth time the smugglers vanish with the culture and you get nothing.

```text
blackmarket              # show cost, odds and cooldown
blackmarket <resource>   # make a smuggling run for the chosen resource
trade black <resource>   # the same, through the trade command
```

`bm` works in place of `blackmarket`. In the Trade panel the Black Market is a smugglers' stop on the map, off the roads (a card in a window too small for the map). Select it to see what a run costs, how much culture you hold, the odds and whether it is ready or lying low, with a chip for each resource a run would pay in and how much. A click on a chip makes that run at once, without a question, and the chips are off while the stop is not ready or you are short of culture. Before the Colonial Age the stop is not there; without Mercantilism it is quiet and says what it needs.

A run costs the larger of 5K culture and a number of hours of your own culture income. That number is about 200 in the Colonial Age, about 166 in the Industrial Age, and 200 in every age from the Victorian Age on. A town that makes twice the culture pays twice.

A run goes through five times in six. When it does, it pays the same worth whichever resource you take it in: the minutes of culture income you staked, as the same minutes of your own income of the resource you chose. A run that cost 200 hours of culture income pays 200 hours of your income of iron, of oil or of gold alike. One time in six the smugglers vanish with the culture and pay nothing.

The smugglers deal in gold and nine resources (iron, coal, steel, oil, electricity, data, crypto, dark matter and quantum flux), and only in those your age already makes. Ask for anything else and they refuse, and nothing is spent. Your income of the resource counts as between a quarter and four times what a moderate town of the age makes of it, so a town that makes none is still paid, and one that makes a great deal is not paid without limit. The culture is spent up front, whether the run goes through or not. After each run there's a cooldown of 624 ticks (about 21 minutes).

The Black Market turns a culture surplus into whatever you're short on, if you can live with the one run in six that is lost.

---

## Strategy

### What routes are worth

A route is sized by the town: it sends 10% of what the town makes of its export, so it grows with the age and the economy. Each doubling of the route's building adds one more such flow. Harbor and ally bonuses add a percentage to what it brings, and a route moves nothing while the town makes none of its export (the board says so). A route is worth a fraction of the income it draws on, so build for the opinion first.

What a route pays all run long is opinion. Each minute a route runs adds +1 opinion with every civilization you have met that isn't at war with you, all at once. A route that runs adds 60 opinion an hour with each of them on its own, however big it is, more than any personality drift takes away. Opinion is what opens better deals, alliances and worker loans (see [Factions & Diplomacy](factions.md#opinion-and-status)), and mercantile civilizations warm to you just for having a route running.

In practice:

- Start `local_barter` as soon as you have a market and The Wheel, and leave it running. Add the cheap routes of each age as you meet their buildings. A cheap route raises opinion as fast as an expensive one.
- Prefer routes whose goods you can spare. A route that eats a resource you're short on costs more than the opinion is worth.
- More routes means more minutes of trade, and each counts with every civilization you know. Once opinion sits near +100 with everyone, more routes add nothing.

---

## Tips and common mistakes

- A route needs its building. `trade route start <route>` refuses with an error if you don't have the required building at the required count. In `trade route list`, routes with a red ✗ are waiting on buildings.
- An embargo mostly hurts you. It cancels that civilization's ally bonus, stops its deals, disrupts your routes that bring in its specialty and counts as a provocation. See [Factions & Diplomacy](factions.md#war-amp-peace).
- Routes slow down, they don't fail, when you run short. If you don't have what a route gives away, it runs at the share the stock can cover and the route stays active. Top up the stockpile and it carries on at its full flow.
