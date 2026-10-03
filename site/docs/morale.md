# Morale

Morale is a civilization-wide stat for the mood of your population. It is shown as a percentage and works as a **two-way dial centered on 50%**: keep it high and production gets a bonus of up to +20%; let it crater and production can fall to half. Unlike faith or knowledge it is not a stored resource you spend. It is a multiplier on what your buildings produce, and you manage it with buildings.

---

## Why Morale Matters

Morale multiplies **all worker-driven building output** every tick: food, knowledge, trade, soldiers, everything. The same buildings and the same worker assignments produce more when morale is high and less when it is low:

```
output = base rate × buildings × (0.20 + 0.80 × workers assigned / worker slots) × morale multiplier
```

Morale is **not** part of the all-production pool. That pool (techs, wonders, milestones, events and the rest) is capped at x3.0 and usually reaches the cap from about the Electric Age (see [The all-production cap](resources.md#the-all-production-cap)). Morale multiplies building output separately, so its bonus keeps counting after the pool has hit the cap.

For how morale fits the rest of the worker system, see [Workers](workers-and-domains.md).

---

## The Scale

- Morale is shown as a **percentage**.
- A new game **starts at 50% (neutral)**. A prestige run starts at **70%**, and a Succumb reset puts it back to 50%.
- It has a hard **floor of 10%**. It can sink low, but never to zero.
- Its **ceiling is 100% + 5% per wonder built**. With no wonders the cap is 100%; each wonder raises it.

So a civilization with 4 wonders built can push morale as high as **120%**, while one with none tops out at **100%**.

---

## The Production Curve

Morale's effect on production is a **continuous curve centered on 50%**, with no neutral "dead zone". The multiplier moves as soon as morale leaves 50%:

| Morale | Effect on production |
|--------|----------------------|
| **At the 10% floor** | **×0.50**, half production (the worst case) |
| **Below 50%** | **Penalty**, easing linearly from ×0.50 (at the floor) to ×1.00 (at 50%) |
| **Exactly 50%** | **×1.00**, the normal baseline |
| **Above 50%** | **Bonus**, rising linearly from ×1.00 (at 50%) to **+20%** (at the morale cap) |
| **At the cap** | **+20%**, the full bonus (the best case) |

The bonus is a share of the way from 50% to your cap: with no wonders, 75% morale is halfway and gives +10%, and a prestige run's opening 70% gives +8%. Each wonder stretches the cap, so the same morale gives a little less bonus until you climb to the new cap.

The two sides are **not symmetric**. Below 50% the multiplier loses 1.25 points of production for every point of morale (from ×1.00 at 50% to ×0.50 at 10%); above 50% it gains only 0.4 points per point with no wonders. A point of morale lost below 50% costs about three times what a point above it earns.

---

## Drift to Neutral

Every tick, morale moves **0.08 points back toward 50%**, after everything else that tick has pushed it. This is the number to beat:

- **Above 50%**, your buildings and faith have to add more than 0.08 points a tick for morale to climb. Add less and it slides back and settles at 50%. A prestige run's 70% drifts down to 50% in about 8 minutes if nothing holds it.
- **Below 50%**, the drift pulls morale up by 0.08 points a tick once you remove whatever was dragging it down. From the 10% floor that is 500 ticks, under 17 minutes, back to neutral.

Neutral is the resting state. Living above it takes a steady source of lift, and the game forgives a dip below it once you stop the cause.

---

## What Raises Morale

| Source | Effect |
|--------|--------|
| **Worship and culture buildings** | The main lever. Each built copy adds a fixed amount every tick, with or without workers (table below). Older tiers keep counting. |
| **Faith production rate** | +0.02 points a tick for each faith per tick you produce, at most +0.4 points a tick (reached at 20 faith per tick). A faith rate of 4 per tick on its own matches the drift. Your faith *rate* counts, not your stored faith. |
| **Good events** | +4 points when a good random event fires. |
| **Advancing to a new age** | +8 points, once per advance. |

## What Lowers Morale

| Source | Effect |
|--------|--------|
| **Starvation** | -0.5 points a tick while your food store is empty and food is falling. |
| **Too large an army** | Military workers above **30% of your population** drain 0.3 points a tick for every 10 percentage points over (at 40% military, -0.3 a tick; at 50%, -0.6). |
| **Idle workforce** | -0.2 points a tick while more than **50% of your workers** sit idle. |
| **Bad events** | -4 points when a bad random event fires. |
| **Enduring a catastrophe** | -10 points. |

When morale falls below 40% the log warns you once; the warning resets when morale climbs back to 40%.

---

## The Key Lever: Worship and Culture Buildings

Because the drift takes 0.08 points a tick, your buildings and faith together have to add more than that before morale climbs at all. Past that point the surplus is what moves it: two Shrines (0.12 a tick) beat the drift by 0.04 points a tick, enough to climb from 50% to 100% in about 40 minutes. Five Shrines (0.30) climb the same distance in under 8 minutes. Once your lift is comfortably above the drift, morale sits at the cap and you get the full **+20%**.

**Morale lift per built copy, points per tick:**

| Age | Worship (faith line) | Lift | Culture line | Lift |
|-----|----------------------|------|--------------|------|
| Primitive | Shrine | 0.06 | | |
| Stone | Standing Stones | 0.06 | | |
| Bronze | Altar | 0.07 | | |
| Iron | Temple | 0.06 | | |
| Classical | Oracle House | 0.09 | Amphitheater | 0.05 |
| Medieval | Cathedral | 0.10 | Great Hall | 0.06 |
| Renaissance | Basilica | 0.12 | Art Studio | 0.07 |
| Colonial | Mission | 0.13 | Concert Hall | 0.08 |
| Industrial | Church | 0.15 | Opera House | 0.09 |
| Victorian | Grand Cathedral | 0.18 | Grand Museum | 0.10 |
| Electric | Revival Hall | 0.20 | Radio Station | 0.12 |
| Atomic | Spiritual Center | 0.23 | Cinema | 0.13 |
| Modern | Meditation Center | 0.27 | TV Studio | 0.15 |
| Information | Digital Temple | 0.31 | Media Center | 0.18 |
| Digital | Cyber Shrine | 0.35 | VR Studio | 0.20 |
| Cyberpunk | Neon Sanctuary | 0.41 | Holographic Theater | 0.23 |
| Fusion | Quantum Chapel | 0.47 | Neural Art Complex | 0.27 |
| Space | Orbital Sanctuary | 0.54 | Zero G Gallery | 0.31 |
| Interstellar | Void Monastery | 0.62 | Cultural Beacon | 0.35 |
| Galactic | Stellar Shrine | 0.71 | Civilization Archive | 0.41 |
| Quantum | Transcendence Hall | 0.82 | Reality Art Engine | 0.47 |

The worship buildings also make faith, which adds its own lift once your faith rate grows, and the culture line needs no workers at all. Good events and age advances are one-time nudges; these buildings are the steady source you control.

---

## Where Morale Is Shown

- **Status bar**: `Morale 62% (production +5%)`, green when morale boosts production and red when it costs production. At exactly 50% it reads just `Morale 50%`.
- **Workers panel** (`workers`): a colored morale bar with the same reading, for example `▲ Morale 62% (production +5%)`, or `Morale 50% (production steady)` at neutral.
- **Stats panel** (`stats`): when morale is off neutral it appears under **Active Multipliers** in the **All Production** line as a `Morale ×N.NN` factor, next to your research, wonder, milestone and active-event bonuses. Each source is listed and colored on its own, so a penalty is never hidden by a bonus on the same line.
- **Load Game browser**: each save's detail pane shows its morale.

On the colorblind-safe and high-contrast [themes](themes.md) the boost and penalty colors are blue and orange instead of green and red.

---

## Managing Morale: Strategy

1. **Build two Shrines early.** One Shrine (0.06) can't beat the drift; two can, and morale starts climbing toward the +20% bonus in the first age.
2. **Keep food positive.** Starvation (-0.5 a tick) is more than six times the drift. Fix the food deficit and the drift heals the rest. See [Starvation](workers-and-domains.md#starvation).
3. **Keep the army in proportion.** Hold military workers under 30% of population; past that, morale drains faster the more lopsided your army gets. See [Military](military.md).
4. **Don't leave workers idle.** More than half your population sitting idle drains morale on top of wasting food. The worker shares routine puts idle workers to work when there are free slots; with no free slots, build worker buildings or `dismiss` them.
5. **Keep a faith economy going.** From about 4 faith per tick, faith alone holds morale up, and at 20 per tick it adds five times the drift.
6. **Build wonders to raise the ceiling.** Each wonder adds +5% to the cap, which keeps the full +20% within reach as long as your lift can carry morale up to it. See [Wonders](wonders.md).

---

## See Also

- [Workers](workers-and-domains.md): the production formula and the worker system
- [Faith](faith.md): faith production, which also lifts morale
- [Military](military.md): the military-ratio morale drain in detail
- [Buildings](buildings.md): the worship and culture lines
- [Wonders](wonders.md): how wonders raise the morale ceiling
