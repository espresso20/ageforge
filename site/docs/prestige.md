# Prestige System

Prestige is the endgame reset loop. When you reach the **Transcendent Age**, you can sacrifice your entire civilization to earn **Prestige Points** and purchase permanent upgrades that carry into every future run.

```
prestige go
```

> ⚠ Prestige resets your age, resources, buildings, villagers, and research. Wonders, milestones, and prestige upgrades are **permanent**.

---

## How prestige works

1. Reach the **Transcendent Age** (Age 21)
2. Check your prestige panel (`F5: Stats → Prestige`) — it shows your available points
3. Type `prestige go` to trigger the reset
4. You respawn in the **Primitive Age** with all prestige upgrades active
5. Use accumulated points to buy upgrades with `prestige buy <key>`

Each prestige run is faster than the last because your permanent bonuses compound.

---

## Prestige upgrades

9 upgrades, each with 5 tiers. Costs are in Prestige Points.

| Upgrade | Key | Effect per Tier | Max Tier | Costs (T1→T5) |
|---|---|---|---|---|
| Gather Boost | `gather_boost` | +5% gather rate | 5 | 2 / 3 / 4 / 6 / 8 |
| Storage Bonus | `storage_bonus` | +20 all storage | 5 | 2 / 3 / 4 / 6 / 8 |
| Research Speed | `research_speed` | +5% knowledge rate | 5 | 2 / 3 / 5 / 8 / 10 |
| Military Power | `military_power` | +5% military power | 5 | 2 / 3 / 5 / 8 / 10 |
| Starting Food | `starting_food` | +25 starting food | 5 | 1 / 2 / 3 / 4 / 5 |
| Starting Wood | `starting_wood` | +25 starting wood | 5 | 1 / 2 / 3 / 4 / 5 |
| Population Cap | `population_cap` | +2 pop cap | 5 | 2 / 3 / 5 / 8 / 10 |
| Expedition Loot | `expedition_loot` | +5% expedition reward | 5 | 2 / 3 / 5 / 8 / 10 |
| Temporal Mastery | `tick_speed` | +5% tick speed | 5 | 6 / 10 / 17 / 23 / 33 |

```bash
prestige buy gather_boost
prestige buy tick_speed
prestige buy starting_food
```

---

## Recommended first-run priorities

| Priority | Upgrade | Why |
|---|---|---|
| 1 | `starting_food` + `starting_wood` | Makes the Primitive Age trivially fast (cheap, 1 point each) |
| 2 | `gather_boost` | All three starting villager types benefit immediately |
| 3 | `research_speed` | Knowledge snowballs — earlier techs = faster progression |
| 4 | `storage_bonus` | Prevents early resource caps from throttling growth |
| 5 | `tick_speed` | Best-in-slot late investment; very high cost but permanent |

---

## Prestige point formula

Points earned per prestige scale with how far you advanced and how many milestones you completed. Completing the Transcendent Age for the first time typically yields **4–8 points** depending on playstyle.

---

## Tips

- You can **buy prestige upgrades before triggering prestige** — spend any points from previous runs first
- **Temporal Mastery** is the most impactful but most expensive upgrade (33 points for Tier 5) — prioritise it on your 3rd or 4th run
- Milestones and completed chains are **preserved across prestige** — your civilization title carries over
- Each run, try to complete one more wonder than the last — wonder bonuses compound with prestige bonuses
