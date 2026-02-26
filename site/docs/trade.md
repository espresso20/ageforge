# Trade & Diplomacy

The trade system lets you exchange surplus resources through trade routes. Diplomacy with 6 NPC factions unlocks opinion-based trade bonuses.

---

## Exchange rates

The Economy shows live exchange rates for all resource pairs. Rates shift based on supply and demand — if you flood a resource, its sell price drops. Rates shown in the **F4: Trade** tab.

---

## Trade routes

Trade routes run for a fixed duration and exchange resources automatically. Only available routes appear (based on age and required buildings).

```
trade start <key>
trade stop <key>
```

| Route | Key | Unlock | Requires | Export | Import | Duration |
|---|---|---|---|---|---|---|
| Local Barter | `local_barter` | Bronze | Market ×1 | 10 food | 8 wood | 10t |
| Stone Trade | `stone_trade` | Iron | Market ×2 | 15 wood | 12 stone | 12t |
| Gold Caravan | `gold_caravan` | Classical | Market ×3 | 50 stone | 5 gold | 15t |
| Silk Road | `silk_road` | Medieval | Market ×2 | 30 gold | 80 culture | 20t |
| Spice Trade | `spice_trade` | Colonial | Port ×1 | 100 gold | 200 food + 50 culture | 18t |
| Colonial Exports | `colonial_exports` | Colonial | Port ×2 | 500 food | 150 gold | 15t |
| Rail Freight | `rail_freight` | Industrial | Train Station ×1 | 200 iron | 100 gold + 50 coal | 12t |
| Oil Pipeline | `oil_pipeline` | Victorian | Oil Well ×2 | 100 oil | 300 gold | 15t |
| Power Exchange | `power_exchange` | Electric | Power Grid ×1 | 500 electricity | 200 gold | 10t |
| Data Trade | `data_trade` | Information | Fiber Hub ×1 | 100 data | 500 gold | 10t |
| Crypto Market | `crypto_market` | Cyberpunk | Black Market ×1 | 50 crypto | 1,000 gold | 8t |
| Fusion Export | `fusion_export` | Fusion | Fusion Reactor ×1 | 200 electricity | 1,000 gold | 12t |
| Warp Commerce | `warp_commerce` | Space | Warp Gate ×1 | 500 gold | 200 dark matter | 15t |
| Stellar Exchange | `stellar_exchange` | Galactic | Galactic Hub ×1 | 100 dark matter | 2,000 gold | 20t |
| Quantum Trade | `quantum_trade` | Quantum | Quantum Computer ×1 | 50 quantum flux | 5,000 gold | 10t |

---

## Diplomacy & Factions

Six NPC factions can be discovered throughout the game. Each has an **opinion** score (-100 to +100) and a **status**:

- `neutral` — default
- `allied` — opinion 60+ — trade bonus applies
- `rival` — opinion -40 or below — trade penalties

```
diplomacy gift <faction_key>
```

Gifting costs gold and raises opinion. Some story events also shift faction opinion automatically.

### Factions

| Faction | Key | Appears | Specialty | Allied bonus |
|---|---|---|---|---|
| Merchant Guild | `merchant_guild` | Colonial Age | Gold | +20% gold rate |
| Artisan League | `artisan_league` | Industrial Age | Culture | +15% culture production |
| Tech Consortium | `tech_consortium` | Information Age | Data | +20% data production |
| Shadow Syndicate | `shadow_syndicate` | Cyberpunk Age | Crypto | +25% crypto production |
| Stellar Federation | `stellar_federation` | Space Age | Dark Matter | +20% dark matter production |
| Quantum Collective | `quantum_collective` | Quantum Age | Quantum Flux | +30% quantum flux production |

---

## Tips

- **Silk Road is highly efficient** in the Medieval Age — gold→culture is cheap if you have surplus gold
- **Crypto Market** is the best gold-per-tick route in the Cyberpunk Age — pair it with Blockchain tech for explosive gold gains
- **Stellar Exchange** and **Quantum Trade** generate so much gold late-game that gold effectively becomes unlimited
- Gift the **Quantum Collective** as soon as it appears — +30% quantum flux is enormous when you're accumulating it for the Singularity Core wonder
- Faction bonuses stack with everything else — getting all 6 to **allied** status gives massive across-the-board production boosts
