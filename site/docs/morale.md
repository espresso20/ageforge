# Morale

Morale is a civilization-wide stat for the mood of your population. It is shown as a percentage and works as a **two-way dial centered on 50%**: keep it high and all production gets a bonus; let it crater and all production takes a penalty. Unlike faith or knowledge it is not a stored resource you spend. It is a multiplier on everything your workers do, and you have to manage it.

For how morale fits the worker production formula, see [Workers & Domains](workers-and-domains.md#morale).

---

## Why Morale Matters

Morale multiplies **all worker-driven building output** every tick: food, knowledge, trade, soldiers, everything. The same buildings and the same worker assignments produce more when morale is high and less when it is low:

```
output = base_rate × building_count × (0.20 + 0.80 × assigned / total_capacity) × morale_multiplier
```

It is the only stat that touches every domain at once, so it is one of the most useful things to keep an eye on.

---

## The Scale

- Morale is shown as a **percentage**.
- It **starts at 50% (neutral)** for a new civilization.
- It has a hard **floor of 10%**. It can sink low, but never to zero.
- Its **ceiling is 100% + 5% per wonder built**. With no wonders the cap is 100%; each wonder raises it, letting morale climb higher and earn a larger production bonus.

So a civilization with 4 wonders built can push morale as high as **120%**, while one with none tops out at **100%**.

---

## The Production Curve

Morale's effect on production is a **continuous curve centered on 50%**, with no neutral "dead zone". The multiplier moves as soon as morale leaves 50%:

| Morale | Effect on All Production |
|--------|--------------------------|
| **At the 10% floor** | **×0.50**, half production (the worst case) |
| **Below 50%** | **Penalty**, easing linearly from ×0.50 (at the floor) to ×1.00 (at 50%) |
| **Exactly 50%** | **×1.00**, the normal baseline |
| **Above 50%** | **Bonus**, rising linearly from ×1.00 (at 50%) to **+20%** (at the morale cap) |
| **At the cap** | **+20%**, the full bonus (the best case) |

The effect scales **smoothly**. Just off 50% it is tiny (at 52% the bonus is well under +1%), and it grows with distance, reaching the endpoints only at the extremes: +20% at the cap, ×0.50 at the 10% floor.

The two sides are **not symmetric**. The 10% floor is closer to 50% than the cap is, so the downside ramp is **steeper than the upside**: a point of morale lost below 50% costs more than a point gained above it earns. Every level of morale has some effect; staying near 50% keeps that effect small.

---

## Drift to Neutral

Morale **drifts gently back toward 50% every tick**. This is the most important thing to understand about it:

- A high-morale **bonus has to be earned and kept up**. If you stop building morale-restoring buildings, the bonus fades as morale drifts back to neutral.
- A low-morale **penalty heals itself**. Once you remove whatever was dragging morale down (fix the food deficit, shed excess military workers), the drift pulls morale back up toward neutral on its own.

Neutral is the resting state. Living above it takes effort, and the game forgives a dip below it once you stop the cause.

---

## What Raises Morale

| Source | Effect |
|--------|--------|
| **Morale-restoring buildings** | The main lever. Worship buildings (shrines, temples and their later-age equivalents) and culture/entertainment buildings lift morale **each tick once built**, with no workers needed. |
| **Faith production rate** | A small morale lift scales with your **faith produced per tick** (your faith *rate*, not your stored faith), so the more faith you are actively generating, the more it nudges morale up. The per-tick lift is **capped**, so a huge late-game faith rate can't max out morale in a single step. |
| **Good events** | A favorable event lifts morale. |
| **Advancing to a new age** | Reaching a new age gives a one-time morale boost. |

## What Lowers Morale

| Source | Effect |
|--------|--------|
| **Starvation** | Running out of food drains morale each tick. |
| **Too large an army** | Military workers above **30% of your population** drain morale; the further over, the faster the drain. |
| **Idle workforce** | More than **50% of your workers sitting idle** drains morale each tick. |
| **Bad events & catastrophes** | A negative event lowers morale, and **enduring a catastrophe** costs morale. |

---

## The Key Lever: Morale-Restoring Buildings

Because morale always drifts back to 50%, the only way to **push it into the bonus zone and keep it there** is to build **morale-restoring buildings**: the worship buildings of each age (the shrine and temple line) and culture/entertainment buildings. They raise morale a little every tick once built, with no workers assigned.

Build enough of them and their per-tick lift outpaces the drift to neutral, holding morale high for a lasting bonus of up to **+20% to all production**. Stop building them and morale slides back to neutral. Good events and age advances are one-time nudges; morale-restoring buildings are the steady source you control.

---

## Where Morale Is Shown

- **Workers panel** (`workers`): a colored morale bar.
- **Status bar**: `Morale: NN%`, colored by effect, with a `+NN%` / `-NN%` tag when it is boosting or penalizing production.
- **Stats panel** (`stats`): when morale is off neutral it appears under **Active Multipliers** in the **All Production** line as a `Morale ×N.NN` factor, next to your research, wonder, prestige and active-event bonuses. A **green** headline is a net bonus, a **red** one a net penalty, and a **white** headline marks a line that shows only because opposing sources cancel out. Each source is listed and colored on its own, so a penalty (a famine event, say) is never hidden by a bonus on the same line. The panel lists **rate multipliers only**; housing and storage bonuses are shown elsewhere.
- **Load Game browser**: each save's detail pane shows its morale, so you can size up a civilization before loading it.

The bar is **green when morale is above 50%** (boosting production), **neutral at exactly 50%**, and **red below 50%** (penalizing production). On the colorblind-safe and high-contrast [themes](commands.md#themes) the boost and penalty colors are blue and orange instead.

---

## Managing Morale: Strategy

1. **50% is the safe baseline.** At exactly 50% there is no bonus and no penalty, so early on you can leave morale alone and spend your effort elsewhere. The curve applies as soon as morale drifts off 50%, small at first and larger the further it goes.
2. **Keep food positive.** Starvation is the most common reason morale slides into a penalty. Fix the food deficit and the drift heals the rest.
3. **Keep the army in proportion.** Hold military workers comfortably under 30% of population; past that, morale drains faster the more lopsided your army gets. (See [Military](military.md).)
4. **Don't leave workers idle.** More than half your population sitting idle drains morale on top of wasting their food. Assign them or `dismiss` them.
5. **Build worship and culture buildings to go positive.** For the +20% production bonus, build morale-restoring buildings faster than morale drifts back to neutral. Holding it there is ongoing work, not a one-time purchase.
6. **Build wonders to raise the ceiling.** Each wonder adds +5% to the cap, so the more wonders you have, the higher morale and its bonus can go. (See [Wonders](wonders.md).)

---

## See Also

- [Workers & Domains](workers-and-domains.md#morale): the production formula and how morale fits the worker system
- [Military](military.md): the military-ratio morale drain in detail
- [Buildings](buildings.md): which worship and culture buildings restore morale
- [Wonders](wonders.md): how wonders raise the morale ceiling
