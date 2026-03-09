# Trade & Diplomacy

Two interlocking systems power your economy beyond raw production: **resource exchange** (on-demand swaps) and **trade routes** (passive per-tick income). Layered on top, the **diplomacy** system lets you court six NPC factions — allied factions amplify your trade route yields, making the two systems deeply synergistic.

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

There is no cap on how many routes can run simultaneously — stack them all.

### Full Trade Routes Reference

| Key | Name | Min Age | Required Building | Export (per cycle) | Import (per cycle) | Cycle (ticks) |
|---|---|---|---|---|---|---|
| `local_barter` | Local Barter | Bronze | Market ×1 | 10 food | 8 wood | 10 |
| `stone_trade` | Stone Trade | Iron | Market ×2 | 15 wood | 12 stone | 12 |
| `gold_caravan` | Gold Caravan | Classical | Market ×3 | 50 stone | 5 gold | 15 |
| `silk_road` | Silk Road | Medieval | Market ×2 | 30 gold | 80 culture | 20 |
| `spice_trade` | Spice Trade | Colonial | Port ×1 | 100 gold | 200 food + 50 culture | 18 |
| `colonial_exports` | Colonial Exports | Colonial | Port ×2 | 500 food | 150 gold | 15 |
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

## Diplomacy

Six NPC factions exist in the game world. Each has an **opinion score** (-100 to +100) and a **diplomatic status**. Status determines whether you benefit from, are ignored by, or are penalised by that faction.

### Commands

```
diplomacy
diplomacy ally <faction_key>
diplomacy rival <faction_key>
diplomacy embargo <faction_key>
diplomacy gift <faction_key>
diplomacy neutral <faction_key>
```

`diplomacy` with no arguments opens the faction status overlay.

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

Two passive paths, one active:

1. **Trade routes:** Every completed trade route cycle calls `RecordTrade()`, which gives **+1 opinion to every discovered faction** simultaneously. Run more routes and they all climb together — passively.
2. **Gifting:** `diplomacy gift <faction_key>` costs 200 gold and gives **+15 opinion** to that faction. If opinion hits 25+, status auto-upgrades from neutral to friendly. Stack gifts to push a stubborn faction over 50 and then `diplomacy ally` them.
3. **Status transitions:** Once a faction is friendly (opinion ≥ 25), you can push to 50+ and spend 500 gold to ally. Going back to neutral is free (`diplomacy neutral`) but doesn't recover the 500 gold.

### Allied Bonuses

When allied, a faction applies its `TradeBonus` as a multiplier to all trade route imports of its specialty resource:

```
actual import = base import × (1.0 + trade_bonus)
```

Bonuses from multiple allied factions stack additively if they share a specialty (unlikely, but possible in theory).

| Faction | Specialty | Allied Bonus |
|---|---|---|
| Merchant Guild | gold | +20% |
| Artisan League | culture | +15% |
| Tech Consortium | data | +20% |
| Shadow Syndicate | crypto | +25% |
| Stellar Federation | dark\_matter | +20% |
| Quantum Collective | quantum\_flux | +30% |

### Faction Discovery

Factions are **auto-discovered** when you reach their minimum age — no manual action required. A log message announces each discovery. Until a faction is discovered, you cannot interact with them.

---

## Full Faction Reference

| Key | Name | Discovered At | Specialty | Allied Bonus | Notes |
|---|---|---|---|---|---|
| `merchant_guild` | Merchant Guild | Colonial Age | gold | +20% gold imports | First faction you'll meet. Most gold routes benefit immediately. |
| `artisan_league` | Artisan League | Industrial Age | culture | +15% culture imports | Silk Road + allied = potent culture engine. |
| `tech_consortium` | Tech Consortium | Information Age | data | +20% data imports | Pairs with Data Trade route for amplified gold conversion. |
| `shadow_syndicate` | Shadow Syndicate | Cyberpunk Age | crypto | +25% crypto imports | Highest early-late bonus. Crypto Market is already the best gold-per-tick route — this makes it obscene. |
| `stellar_federation` | Stellar Federation | Space Age | dark\_matter | +20% dark matter imports | Warp Commerce becomes a dark matter spigot with this allied. |
| `quantum_collective` | Quantum Collective | Quantum Age | quantum\_flux | +30% quantum flux imports | Largest single bonus in the game. Quantum Trade + allied = effectively unlimited gold. |

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

With all 15 routes active and all 6 factions allied, gold income from routes alone becomes enormous. At that point, resource exchange becomes redundant for most pairs — use it only for exotic resource-to-resource conversions that no route covers.

---

## Tips & Common Mistakes

- **The building must exist.** `trade route start <key>` will reject with a clear error if you don't have the required building at the required count. Check `trade route list` — routes with a red ✗ are waiting on buildings.
- **Market pressure decays naturally.** Don't panic if you see a rate penalty. Wait a few ticks and the pressure drops. No action required.
- **Embargo is not a punishment, it's a lockout — for you.** Declaring embargo prevents the faction's trade bonus from applying. There's almost no reason to ever embargo a faction; declaring rival is equally (and pointlessly) hostile but at least you could theoretically benefit from a mechanic around it later. Just leave factions at neutral if you don't want to invest in them.
- **Routes skip, not fail, when you're broke.** If your export stockpile is empty when a cycle fires, the cycle is silently skipped. Your route stays active. Top up exports and it resumes next cycle automatically.
- **Allied status costs 500 gold every time.** If you reset to neutral and want to re-ally, you pay 500 gold again. Opinion carries over — you don't lose it when going neutral — so you just need the gold.
- **Gift stacking is fast.** Six gifts (1,200 gold) takes a faction from 0 opinion to 90 opinion (+15 × 6), well past the 50 threshold. If you need a faction allied fast and have the gold, this is the quickest path.
- **Quantum Collective is worth prioritising.** The +30% quantum flux bonus is the largest single faction bonus in the game. Gift them to 50 opinion immediately when discovered and ally as soon as you can afford it — quantum flux is the gating resource for several late wonders.
