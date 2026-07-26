# Trade & Diplomacy

Two interlocking systems power your economy beyond raw production: **resource exchange** (on-demand swaps) and **trade routes** (passive per-tick income). Layered on top, the **diplomacy** system lets you encounter an **11-civilization roster** of NPC powers — allied civilizations amplify your trade route yields, peaceful ones lend you workers, and provoked ones can declare war. The systems are deeply synergistic.

---

## Resource Exchange

Instant, one-off swaps between two resource types. You need at least one `market` or `port` built before any exchange is possible.

```
trade <from> <to> <amount>
trade list
```

`trade list` shows all rates currently available to your age, including any active market pressure penalties.

### Exchange Rates

Each pair has a **base rate** — the units of the target resource you receive per unit sold at zero market pressure. Rates are unlocked by age:

| Age | Available Pairs |
|---|---|
| Bronze | food↔wood, food↔stone, wood↔stone, gold→food/wood/stone |
| Iron | iron↔gold, iron→stone |
| Medieval | gold→knowledge, gold→culture, faith→culture |
| Colonial | gold↔coal |
| Industrial | steel→gold, oil→gold |
| Electric | electricity→gold |
| Information | data↔gold |
| Cyberpunk | crypto↔gold |
| Space | dark\_matter→gold |
| Quantum | quantum\_flux→gold |

Full base rates for notable pairs:

| From | To | Base Rate |
|---|---|---|
| gold | food | 50 |
| gold | wood | 40 |
| gold | stone | 30 |
| iron | gold | 2.0 |
| iron | stone | 3.0 |
| gold | knowledge | 5.0 |
| faith | culture | 2.0 |
| data | gold | 5.0 |
| crypto | gold | 20.0 |
| dark\_matter | gold | 50.0 |
| quantum\_flux | gold | 100.0 |

### Market Pressure

Every exchange you execute on the same pair increases **supply pressure** on that pair. Pressure reduces the effective rate:

```
effective rate = base rate × (1 − pressure × 0.30)
```

Pressure caps at 1.0 (a 30% rate reduction). There is also a hard floor at 50% of base rate — you can never be squeezed below half the listed rate.

Pressure **decays 2% per tick**, multiplicatively. Leave a pair alone and it recovers fully on its own. Having more markets and ports helps too: each market or port you own reduces how much pressure a single trade adds (formula: `+0.10 / (1 + market_count × 0.20)`).

**When to exchange:** Use exchanges to convert surplus resources into something you're running short on, or to buy a specific resource you can't produce yet. Don't use the same pair repeatedly in quick succession — you'll hammer the rate. Spread trades across different pairs, or wait a few ticks between repeat swaps on the same pair.

---

## Trade Routes

Trade routes run automatically in the background, consuming a set of **export** resources every N ticks and delivering **import** resources in return. Unlike exchanges, routes don't suffer market pressure — they just need the required buildings and enough exports in stock to fire.

### Commands

```
trade route list
trade route start <key>
trade route stop <key>
```

- `trade route list` — shows active routes (with ticks left and cycles completed) and available routes (green checkmark if you have the required building, red X if not).
- `trade route start <key>` — activates a route. Fails if the required building isn't built or if you haven't reached the route's minimum age.
- `trade route stop <key>` — deactivates a route immediately, mid-cycle.

A route **auto-suspends** if its required building is demolished while it's running — you'll see a log message. Rebuild the building and `trade route start` again to resume it.

If a route fires but you don't have enough export resources, it silently skips that cycle (no penalty) and tries again next cycle. Keep your export stockpiles healthy.

### Trade Disruption (War & Embargo)

Conflict has a price. If you are **at war** with a civilization, or you have placed it under **embargo**, every route whose **imports** include that civ's **specialty resource** is **disrupted**: it earns nothing and consumes nothing while the conflict lasts. The route isn't stopped — its timer keeps running — so it **resumes automatically** the moment peace returns (end the war via tribute or wait it out; lift the embargo with `diplomacy neutral`).

Disrupted routes are flagged in the Trade overlay with a red ✖ and a "DISRUPTED — shipments blockaded" note naming the affected resource, and a banner at the top of the routes panel lists every blockaded resource. You'll also see a log line each time a disrupted route would have fired.

This reuses the existing diplomacy war/embargo state — there's no separate disruption mechanic to track. The practical lesson: before you embargo or provoke the **gold** specialist (Merchant Guild) or the **culture** specialist (Artisan League), check which of your routes import those goods.

There is no cap on how many routes can run simultaneously — stack them all.

### Full Trade Routes Reference

| Key | Name | Min Age | Required Building | Export (per cycle) | Import (per cycle) | Cycle (ticks) |
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
| `crypto_market` | Crypto Market | Cyberpunk | Black Market ×1 | 50 crypto | 1,000 gold | 8 |
| `fusion_export` | Fusion Export | Fusion | Fusion Reactor ×1 | 200 electricity | 1,000 gold | 12 |
| `warp_commerce` | Warp Commerce | Space | Warp Drive Plant ×1 | 500 gold | 200 dark matter | 15 |
| `stellar_exchange` | Stellar Exchange | Galactic | Galactic Trade Hub ×1 | 100 dark matter | 2,000 gold | 20 |
| `quantum_trade` | Quantum Trade | Quantum | Reality Processor ×1 | 50 quantum flux | 5,000 gold | 10 |

---

## Harbour Lineage — Trade-Route Income

Markets and banks (the **trade** lineage) make gold directly. **Harbours** are different: they make your **trade routes** more profitable. Each built harbour adds a flat percentage bonus to the **imports of every active route**, stacking additively across tiers and instances. They also produce a little gold themselves, so an idle harbour still earns its keep.

The bonus stacks with allied-civ trade bonuses: a route importing a specialty resource you're allied for, run through a fleet of harbours, pays out `base × (1 + harbour_bonus + ally_bonus)`.

| Tier | Key | Name | Min Age | Route Income Bonus | Workers |
|---|---|---|---|---|---|
| 0 | `harbor` | Harbor | Colonial | +5% | 4 |
| 1 | `harbor_authority` | Harbor Authority | Industrial | +10% | 5 |
| 2 | `seaport` | Seaport | Modern | +15% | 6 |
| 3 | `container_terminal` | Container Terminal | Information | +20% | 8 |
| 4 | `logistics_hub` | Logistics Hub | Digital | +25% | 10 |

Harbours use the **trade** worker domain — the same recruits that staff markets and embassies — so a big harbour fleet competes with your markets for hands. Several Colonial-era routes (`triangular_trade`, `tea_clippers`, `coal_barges`) require harbours rather than ports, giving the colonial→industrial economy something fresh to build toward.

---

## Black Market

Once you reach the **Colonial Age**, smuggling networks open up. The black market is a **high-risk, high-reward culture sink**: you spend a lump of **culture** on a deal that *might* pay out a large haul of a resource of your choice — or vanish with your culture and deliver nothing.

```text
blackmarket              # show cost, odds, and cooldown
blackmarket <resource>   # run a deal for the chosen resource
trade black <resource>   # the same thing, via the trade command
```

- **Cost:** `max(5,000, 10% of your culture storage cap)` culture per deal, scaling with your progression.
- **Odds:** a **55% chance** of a payout. On a win you receive the chosen resource worth **2.5×** the culture stake (valued via that resource's gold exchange rate). On a loss the culture is simply gone.
- **Cooldown:** ~240 ticks (about 8 minutes) between deals, so it can't be spammed.
- **Always-spent:** the culture is consumed up front, win or lose — that's the gamble.

It's a way to convert a culture surplus into a swing of whatever resource you're short on, if you're willing to ride the variance.

---

## Diplomacy — Civilization Encounters

The game world holds an **11-civilization roster**. You meet them by **running expeditions** — an age only makes a civ *eligible*; it's a resolved **scouting expedition or military campaign** that actually turns someone up, with a generous late fallback so even a player who never explores meets everyone eventually. Each civ has an **opinion score** (-100 to +100), a **diplomatic status**, a **personality**, and a **backstory**. Status determines whether you benefit from, are ignored by, or are penalised by that civ; personality drives how its opinion drifts and whether it lends workers or goes to war.

### Commands

```
diplomacy                       # opens the Diplomacy overlay
diplomacy ally <civ_key>
diplomacy rival <civ_key>
diplomacy embargo <civ_key>
diplomacy gift <civ_key>
diplomacy neutral <civ_key>
diplomacy tribute <civ_key>     # sue for peace with a civ at war
diplomacy raid <civ_key>        # raid their trade route (provocation — tanks opinion)
```

`diplomacy` (or `dip`) with no arguments opens the **Diplomacy overlay** — a full-screen panel listing every civilization with its personality, a backstory snippet, opinion bars, color-coded status, the active trade-rate bonus, a threshold indicator (e.g. *+8 to friendly*, *ally-eligible — 500g*), war banners, and lent-worker status. `diplomacy <action> <civ_key>` still performs the action directly.

### Personalities

Every civilization has one of four personalities that shapes its passive opinion drift and its behaviour toward you:

| Personality | Opinion drift | Behaviour |
|---|---|---|
| **peaceful** | trends **up** over time | Lends you workers when standing is high (see Worker Lending) |
| **aggressive** | trends **down** over time | Provocable into **war** when deeply hostile |
| **mercantile** | rises when you **trade**, cools when you don't | Trade-focused; reward active trade routes |
| **isolationist** | trends toward **neutral** (0) | Slow to befriend, slow to anger |

Drift is gradual (±1 on a periodic cadence) and clamped to the -100..+100 range. It runs alongside the existing rival/embargo decay and the natural drift toward zero.

### Worker Lending

Peaceful civilizations with healthy opinion (40+) occasionally **lend you workers** via an event — a backstory-flavoured *"+N workers from the &lt;civ&gt;"* message. Lent workers join your pool immediately (they may temporarily exceed your population cap) and stay for a fixed window before returning home. If the lending civ's opinion is **above 80**, the loan is **permanent** — the workers choose to stay. Loans are tracked per-civ and surface in the overlay as *↳ N workers on loan*.

### War & Peace

A civilization declares **war** only when **both** conditions are met: its opinion is **below -75** *and* a **provocation threshold** is crossed. Provocations are tracked per-civ — **raiding their trade route** (`diplomacy raid`) counts as one, and **embargoing them** counts as one. Two provocations (e.g. a raid + an embargo, or two embargoes) while deeply hostile trips the war. Anger alone never starts a war, and provocations while on good terms don't either.

While at war, the civ launches periodic **raid events** that drain resources — severity scales with the civ's **strength** (1-5). War is purely event-driven; there is no tactical combat.

To make **peace**, you have two options:

1. **Tribute** — `diplomacy tribute <civ>` pays gold + culture (scaled to the civ's strength) to end the war immediately and restore a wary truce.
2. **Wait them out** — a war auto-ends after a stretch of provocation-free ticks. Stop poking them and the war burns out on its own.

The shorthand `dip` works in place of `diplomacy` everywhere.

### Opinion and Status

| Status | Meaning |
|---|---|
| `neutral` | Default state. No bonuses, no penalties. |
| `friendly` | Reached automatically when opinion hits 25+. No mechanical effect yet, but you're close to allied. |
| `allied` | Requires opinion ≥ 50 and costs 500 gold. Grants the faction's trade bonus to your imports. |
| `rival` | Free to declare. Opinion decays an extra -5 every 50 ticks (on top of natural drift). No trade bonuses. |
| `embargo` | Free to declare. Same opinion drain as rival. Cuts you off from that faction's trade bonus entirely. |

**Natural opinion drift:** Every 100 ticks, opinion nudges 1 point toward zero. At allied status with positive opinion, this is a slow bleed — keep trading to offset it.

**Rival/embargo drain:** -5 opinion every 50 ticks. A faction at -100 is stuck there; a faction at +80 with rival declared will fall off allied threshold eventually.

### Raising Opinion

Three passive paths, one active:

1. **Trade routes:** Every completed trade route cycle calls `RecordTrade()`, which gives **+1 opinion to every discovered faction** simultaneously. Run more routes and they all climb together — passively.
2. **Embassies:** Staffed Embassy and Grand Embassy buildings passively generate opinion every tick, spread across your non-hostile factions (see **Embassy Buildings** below).
3. **Gifting:** `diplomacy gift <faction_key>` costs 200 gold and gives **+15 opinion** to that faction. If opinion hits 25+, status auto-upgrades from neutral to friendly. Stack gifts to push a stubborn faction over 50 and then `diplomacy ally` them.
4. **Status transitions:** Once a faction is friendly (opinion ≥ 25), you can push to 50+ and spend 500 gold to ally. Going back to neutral is free (`diplomacy neutral`) but doesn't recover the 500 gold.

### Embassy Buildings

Two diplomacy buildings turn your workforce into a steady source of opinion. Assign workers to an embassy and, each tick, it raises opinion with every **non-hostile** faction (neutral, friendly, or allied — rivals and embargoed factions get nothing). Output scales with the worker fill ratio using the same `0.20 + 0.80 × fill` curve as production buildings, so a fully-staffed embassy runs at full rate while an empty one still trickles. Opinion is capped at +100 per faction.

| Building | Unlocks | Cost | Workers | Opinion Rate |
|---|---|---|---|---|
| **Embassy** | Colonial Age | gold + iron | 5 (trade domain) | +0.05 opinion / worker / tick |
| **Grand Embassy** | Industrial Age | gold + steel | 8 (trade domain) | +0.10 opinion / worker / tick (2× the Embassy) |

Embassies use the **trade** worker domain (the same Colonial Merchant / Industrialist classes that staff markets), so embassy and market workers draw from the same pool — staffing one means fewer hands for the other. Because the per-tick gain is split across all discovered non-hostile factions, embassies are most effective once several factions are in play (Industrial Age onward).

### Allied Bonuses

When allied, a faction applies its `TradeBonus` as a multiplier to all trade route imports of its specialty resource:

```
actual import = base import × (1.0 + trade_bonus)
```

Bonuses from multiple allied factions stack additively if they share a specialty (unlikely, but possible in theory).

The allied trade bonus also surfaces in the **Active Multipliers** panel (Stats overlay) as a `Diplomacy` line on the affected resource's rate, so you can see at a glance which of your production rates an alliance is amplifying. The number shown is the same `1 + trade_bonus` factor described above — the panel and the applied bonus read from the same source, so they can't drift.

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

Discovery follows a **floor + trigger + fallback** model:

- **Age is a floor.** Reaching a civilization's minimum age makes it *eligible* to be met — it does **not** discover it on its own.
- **Expeditions are the trigger.** Whenever a **scouting expedition** or **military campaign** resolves, the game rolls a chance to **encounter** a civilization. An encounter discovers a new eligible civ (first contact) — or, once you already know everyone within reach, re-encounters a known one. So you find new civilizations by *running expeditions*: scouting turns them up more readily than military campaigns, and success beats failure (see [Military & Expeditions](military.md#faction-encounters) for the odds).
- **Late fallback.** Never run an expedition and you still meet everyone eventually: about **two ages past** a civ's minimum age it is auto-discovered anyway — just far later than an active explorer would have met it.

A flavour log message introduces each civ's name, personality, and backstory on first contact. Until a civilization is discovered, it shows in the overlay as a locked teaser (*??? — reach the X Age*) and you cannot interact with it. The founding civs (Riverlands Tribes, Ironhold Clans) appear early; the rest are met across the eras through the Cosmic Era.

**Encounter boons.** An encounter — first contact *or* a re-encounter of a civ you already know — can also grant a temporary **boon**: a boost to production of that civilization's **specialty resource**, currently **+8% to +20% for 3,000–6,000 ticks**. A flavour line names the civ and the boost, e.g. *"The Ironhold Clans share their forge-craft — +15% iron for 4,200 ticks."* It's one more reason to keep expeditions running even after you've met everyone — see [Military & Expeditions](military.md#faction-encounters).

---

## Full Civilization Reference

| Key | Name | Eligible From | Personality | Strength | Specialty | Allied Bonus |
|---|---|---|---|---|---|---|
| `riverlands_tribes` | Riverlands Tribes | Bronze Age | peaceful | 1 | food | +15% food imports |
| `ironhold_clans` | Ironhold Clans | Medieval Age | aggressive | 3 | iron | +20% iron imports |
| `merchant_guild` | Merchant Guild | Colonial Age | mercantile | 2 | gold | +20% gold imports |
| `artisan_league` | Artisan League | Industrial Age | peaceful | 1 | culture | +15% culture imports |
| `atomic_directorate` | Atomic Directorate | Atomic Age | isolationist | 4 | steel | +20% steel imports |
| `tech_consortium` | Tech Consortium | Information Age | mercantile | 2 | data | +20% data imports |
| `shadow_syndicate` | Shadow Syndicate | Cyberpunk Age | aggressive | 3 | crypto | +25% crypto imports |
| `plasma_nomads` | Plasma Nomads | Fusion Age | peaceful | 2 | plasma | +22% plasma imports |
| `stellar_federation` | Stellar Federation | Space Age | isolationist | 4 | dark\_matter | +20% dark matter imports |
| `void_reavers` | Void Reavers | Galactic Age | aggressive | 5 | antimatter | +28% antimatter imports |
| `quantum_collective` | Quantum Collective | Quantum Age | isolationist | 5 | quantum\_flux | +30% quantum flux imports |

**Notable behaviours:**

- **Riverlands Tribes / Plasma Nomads / Artisan League** (peaceful) are your worker-lending civs — keep their opinion high (80+ for permanent loans).
- **Ironhold Clans / Shadow Syndicate / Void Reavers** (aggressive) drift hostile and will declare war if you provoke them while deeply disliked. The Void Reavers (strength 5) raid hardest.
- **Atomic Directorate / Stellar Federation / Quantum Collective** (isolationist) sit near neutral — hard to befriend, hard to anger.
- **Merchant Guild / Tech Consortium** (mercantile) warm up the more trade routes you run.

---

## Strategy

### Trade Routes: Priority Order

**Early (Bronze–Medieval):** Start `local_barter` the moment you build a market — it runs indefinitely and costs almost nothing. Add `stone_trade` and `gold_caravan` as soon as you have Markets ×2 and ×3. The `silk_road` in Medieval is the best culture-per-tick route for most of that era; prioritise it over raw gold unless you're flush.

**Mid-game (Colonial–Industrial):** `spice_trade` and `colonial_exports` both require a port, so build one early in Colonial. They pull in opposite directions — spice trade gives food and culture, colonial exports flip food into gold. Run both simultaneously once you have Port ×2. `rail_freight` in Industrial pays well but eats iron; only start it if you have iron to spare.

**Late-game (Electric onward):** Every subsequent route has dramatically increasing gold yields. Stack all of them as fast as you can build the required infrastructure. `crypto_market` has the shortest cycle (8 ticks) and pays 1,000 gold — it's the most gold-efficient route until `quantum_trade` comes online.

### Diplomacy: When to Ally vs Stay Neutral

Ally when you can afford the 500 gold and you're actively running the faction's specialty route. The math is simple: if `Crypto Market` is generating 1,000 gold per 8 ticks, a 25% bonus from Shadow Syndicate is an extra 250 gold per cycle — the 500 gold cost pays back in two cycles.

Don't bother rushing allies before you have the relevant route running. Opinion will climb on its own from trade route cycles; save your gold for construction or research until the route is active.

**Gift usage:** Gifts (200 gold, +15 opinion) are most valuable when you're stuck just below the 50-threshold needed to ally. Three gifts push a faction from 0 to 45 — one more trade cycle tips them over. Don't gift factions you haven't discovered or factions you're embargoing.

### Market Pressure Management

- Never spam the same exchange pair back-to-back. Use `trade list` to check the pressure indicator before repeat trades — a `↓` marker with percentage means you're paying the penalty.
- If you need large amounts of one resource, spread exchanges across different source resources (e.g., convert iron to gold, then wood to gold separately, rather than repeatedly selling iron).
- Building more markets and ports directly reduces how fast pressure builds per trade. In Colonial Age+, having Port ×2 and Market ×3 makes pressure largely a non-issue for occasional exchanges.

### Diplomacy + Trade Synergy

Every completed trade route cycle raises opinion with all discovered factions by 1. Running five active routes means every faction gains 5 opinion per cycle collectively. This compounds: more routes → faster opinion gain → earlier alliance → bigger import bonuses → faster resource accumulation → more builds → more routes.

The loop is self-reinforcing. Don't think of routes and diplomacy as separate systems — they're one flywheel.

### End-game State

With all 15 routes active and your discovered civilizations allied, gold income from routes alone becomes enormous. At that point, resource exchange becomes redundant for most pairs — use it only for exotic resource-to-resource conversions that no route covers.

---

## Tips & Common Mistakes

- **The building must exist.** `trade route start <key>` will reject with a clear error if you don't have the required building at the required count. Check `trade route list` — routes with a red ✗ are waiting on buildings.
- **Market pressure decays naturally.** Don't panic if you see a rate penalty. Wait a few ticks and the pressure drops. No action required.
- **Embargo is not a punishment, it's a lockout — for you.** Declaring embargo prevents the faction's trade bonus from applying. There's almost no reason to ever embargo a faction; declaring rival is equally (and pointlessly) hostile but at least you could theoretically benefit from a mechanic around it later. Just leave factions at neutral if you don't want to invest in them.
- **Routes skip, not fail, when you're broke.** If your export stockpile is empty when a cycle fires, the cycle is silently skipped. Your route stays active. Top up exports and it resumes next cycle automatically.
- **Allied status costs 500 gold every time.** If you reset to neutral and want to re-ally, you pay 500 gold again. Opinion carries over — you don't lose it when going neutral — so you just need the gold.
- **Gift stacking is fast.** Six gifts (1,200 gold) takes a faction from 0 opinion to 90 opinion (+15 × 6), well past the 50 threshold. If you need a faction allied fast and have the gold, this is the quickest path.
- **Quantum Collective is worth prioritising.** The +30% quantum flux bonus is the largest single faction bonus in the game. Gift them to 50 opinion immediately when discovered and ally as soon as you can afford it — quantum flux is the gating resource for several late wonders.
