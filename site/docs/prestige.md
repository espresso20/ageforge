# Prestige System

Prestige is the endgame reset loop. When you reach the **Modern Age** (Age 12), you can sacrifice your entire civilization to earn **Prestige Points** and purchase permanent upgrades that carry into every future run.

```
prestige confirm yes
```

> Prestige resets your age, resources, buildings, workers, and research. Prestige upgrades, legacy bonuses from Succumb, and ruins are **permanent**.

---

## When You Can Prestige

You unlock the ability to prestige at **Modern Age (Age 12)** and any age beyond. There is no hard cap — if you push all the way to Quantum Age before prestiiging, you'll earn substantially more points.

To check your current prestige status:

```
prestige
```

This shows your current level, available points, points you would earn right now, and whether the prestige threshold has been met. To view the upgrade shop without committing:

```
prestige shop
```

When you're ready:

```
prestige confirm yes
```

The double confirmation (`confirm yes`) is intentional — prestige is irreversible.

---

## Prestige Points Formula

Points earned per prestige run are calculated as:

```
base      = age_index  (0 = Primitive, 1 = Stone, ..., 12 = Modern, 21 = Quantum)
bonus     = floor(milestones / 10) + floor(techs / 15) + floor(total_built / 50)
raw       = base + bonus
points    = floor(raw / sqrt(prestige_level + 1))
```

**Diminishing returns** apply via the `sqrt(level + 1)` divisor — each run yields fewer raw points per achievement as your prestige level grows. A minimum of 1 point is guaranteed if you've reached Medieval Age (index ≥ 5) and your level isn't too high.

### What contributes to points

| Source | Points per unit |
|--------|----------------|
| Age index (each age beyond Primitive) | 1 pt each |
| Every 10 milestones completed | +1 pt |
| Every 15 techs researched | +1 pt |
| Every 50 buildings constructed (lifetime) | +1 pt |

Reaching Modern Age for the first time typically yields **4–8 points** depending on playstyle. Pushing to late ages (Quantum = index 21) before prestiiging yields 20+ before the divisor.

---

## Prestige Upgrades

9 upgrades, each with 5 tiers. Costs are in Prestige Points. All upgrades persist across every reset, including prestige and Succumb.

| Upgrade | Key | Effect per Tier | Max Tier | Cost (T1 → T5) |
|---------|-----|-----------------|----------|----------------|
| Gather Boost | `gather_boost` | +5% gather rate | 5 | 2 / 3 / 4 / 6 / 8 |
| Storage Bonus | `storage_bonus` | +20 all storage | 5 | 2 / 3 / 4 / 6 / 8 |
| Research Speed | `research_speed` | +5% knowledge rate | 5 | 2 / 3 / 5 / 8 / 10 |
| Military Power | `military_power` | +5% military power | 5 | 2 / 3 / 5 / 8 / 10 |
| Starting Food | `starting_food` | +25 starting food | 5 | 1 / 2 / 3 / 4 / 5 |
| Starting Wood | `starting_wood` | +25 starting wood | 5 | 1 / 2 / 3 / 4 / 5 |
| Population Cap | `population_cap` | +2 population cap | 5 | 2 / 3 / 5 / 8 / 10 |
| Expedition Loot | `expedition_loot` | +5% expedition reward | 5 | 2 / 3 / 5 / 8 / 10 |
| Temporal Mastery | `tick_speed` | +5% tick speed | 5 | 6 / 10 / 17 / 23 / 33 |

```
prestige shop                — view available upgrades and costs
prestige buy gather_boost    — buy next tier of Gather Boost
prestige buy tick_speed      — buy next tier of Temporal Mastery
prestige buy starting_food   — buy next tier of Starting Food
```

You can buy prestige upgrades **before triggering prestige** — spend any points accumulated from prior runs as soon as you log in. There is no reason to wait.

### Effect Types

- **`rate_bonus`** upgrades (Gather Boost, Research Speed, Military Power, Expedition Loot, Temporal Mastery) — each tier adds a fractional multiplier to the named rate. `gather_boost` at tier 3 = +15% gather rate.
- **`flat_bonus`** upgrades (Storage Bonus, Population Cap) — each tier adds a flat value. Storage Bonus at tier 5 = +100 all storage.
- **`starting_resource`** upgrades (Starting Food, Starting Wood) — each tier adds to the starting amount of that resource on reset. At tier 5, you begin every run with +125 food or wood.

---

## Passive Prestige Bonuses

Beyond the purchased upgrades, you gain **passive bonuses** just from having a higher prestige level:

```
+2% production (all resources) per prestige level
+1% tick speed per prestige level
```

These stack on top of your purchased upgrade bonuses. A prestige level 5 player has +10% production and +5% tick speed before spending a single prestige point on upgrades.

---

## What Resets vs Persists

### Resets on Prestige
- All resources (reset to starting amounts: 15 food, 12 wood + prestige bonuses)
- All buildings and build queue
- All workers (recruited and assigned)
- All research (tech tree reverts)
- Milestones and milestone chains
- Current epoch and epoch event history
- Age (returns to Primitive Age)

### Persists Across Prestige
- Prestige level and all purchased upgrade tiers
- Ruins (from past Succumb events) — carry into the new run
- Legacy bonuses (from Succumb events) — active from tick 1
- Ancient Knowledge bonus (+25% research speed per Succumb) — stacks across all Succumbs
- Civilization history / catastrophe log

### Morale on Prestige

Morale resets to **0.70** on prestige — not the default 1.0 of a completely fresh run, but not the punishing 0.50 of a Succumb reset either. The 0.70 floor represents institutional memory: your rebuilt civilization begins slightly behind peak output but recovers faster than one starting from scratch.

In practice this means prestige runs begin at roughly 70% worker efficiency and must recover to 1.0 (or higher, with wonders built) through food surplus ticks (+0.002/tick), passive recovery (+0.001/tick), and age advances (+0.08 each). With ruins producing from tick 1 and starting food/wood prestige bonuses, keeping food in surplus from the first few ticks is usually straightforward — morale will tick back toward 1.0 within a few dozen ticks if food stays positive.

### Culture on Prestige

Culture is reduced to **20% of its current value** rather than fully reset. Any unlocked culture threshold bonuses remain permanently unlocked. This means a player who accumulated culture across multiple runs retains meaningful culture bonuses even after prestige.

---

## Legacy Bonuses

Legacy bonuses are earned by choosing **Succumb** during a catastrophe event. They are separate from prestige upgrades but interact with them on every subsequent run.

Each Succumb grants:
- **Ancient Knowledge** — permanent +25% research speed (stacks additively per Succumb)
- **Epoch Legacy Bonus** — permanent production multiplier for the primary resources of that epoch

| Epoch | Legacy Production Bonus |
|-------|------------------------|
| Stone Era | wood +20%, stone +20% |
| Iron Era | iron +20% |
| Steel Era | steel +25%, coal +25% |
| Electric Era | electricity +25%, uranium +25% |
| Digital Era | data +30%, titanium ore +30% |
| Neon Era | plasma +30%, dark matter crystals +30% |
| Cosmic Era | dark matter +35% |

These bonuses apply from **tick 1** of every new run, including after prestige. A player who has Succumbed in the Stone Era and Iron Era starts every run with wood, stone, and iron production already multiplied.

Multiple Succumbs in different epochs stack independently. There is no cap on how many legacy bonuses you can accumulate across runs.

Legacy bonuses survive prestige the same way ruins do — they are part of your permanent meta-state.

---

## Ruins at Prestige

When you prestige, any ruins you've accumulated from Succumb events carry forward. Ruins produce at 50% base rate with no worker requirement — they're free production from tick 1.

On a fresh prestige run with accumulated ruins, your food, wood, or other resources may already be ticking up before you've built a single building. The more catastrophes you've Succumbed to, the stronger your ruins collection.

See [Catastrophe](catastrophe.md) for how ruins are generated.

---

## Recommended Upgrade Priorities

### First prestige (4–8 points)

| Priority | Upgrade | Why |
|----------|---------|-----|
| 1st | `starting_food` + `starting_wood` | Makes Primitive Age trivially fast; 1 pt each at tier 1 |
| 2nd | `gather_boost` tier 1–2 | Accelerates early resource collection in every run |
| 3rd | `research_speed` tier 1 | Knowledge snowballs; earlier techs mean faster ages |

### Second and third prestige (10–20 points)

| Priority | Upgrade | Why |
|----------|---------|-----|
| 1st | `storage_bonus` tier 1–2 | Prevents early resource caps from throttling growth |
| 2nd | `research_speed` tier 2–3 | Compound gains become significant at higher tiers |
| 3rd | `population_cap` tier 1 | +2 pop is small but early housing pressure is real |

### Late game (20+ points available)

| Priority | Upgrade | Why |
|----------|---------|-----|
| 1st | `tick_speed` — start buying tiers | The single most impactful upgrade long-term |
| 2nd | Max out `research_speed` | Full 5 tiers = +25% knowledge rate, stacks with Ancient Knowledge |
| 3rd | `expedition_loot` | Late-game resource acceleration via expeditions |

**Temporal Mastery** (`tick_speed`) is the most expensive upgrade (33 points for tier 5) but also the most powerful — each tier makes the entire game tick 5% faster. At tier 5 you're running at 1.25× base speed before passive bonuses.

---

## Faith Bonus at Prestige

If your faith resource is at **100% of its storage cap** at the moment you prestige, you receive a faith prestige multiplier — additional prestige points on top of the normal formula. Stacking faith buildings before triggering prestige is a legitimate optimization.

---

## Tips

- **Buy upgrades before prestiging** — spend any banked points from prior runs the moment you log in; you don't need to trigger prestige to spend points
- **Milestones are permanent** — your civilization title and all completed milestone chains carry over; milestone progress toward chains is preserved
- **Each run should complete one more wonder than the last** — wonder bonuses compound with prestige bonuses for dramatic acceleration
- **Don't rush the first prestige** — reaching further ages (through Classical, Medieval, even Renaissance) before your first prestige gives substantially more points than resetting at the minimum threshold
- **The passive bonus compounds** — at prestige level 10, you have +20% production all and +10% tick speed before spending a single upgrade point
