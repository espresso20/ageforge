# Morale

Morale is a civilization-wide stat that reflects the mood and motivation of your population. It is shown as a percentage and acts as a **two-way dial centred on 50%**: keep it high and all production is rewarded with a bonus; let it crater and all production is penalised. Unlike Faith or Knowledge it is not a stored resource you spend — it is a multiplier on everything your workers do, and it must be actively managed.

For how morale interacts with the worker production formula, see [Workers & Domains](workers-and-domains.md#morale).

---

## Why Morale Matters

Morale multiplies **all worker-driven building output** every tick — food, knowledge, trade, soldiers, everything. The same buildings and the same worker assignments produce more when morale is high and less when it is low:

```
output = base_rate × building_count × (0.20 + 0.80 × assigned / total_capacity) × morale_multiplier
```

It is the only stat that touches every domain at once, which makes it one of the highest-leverage things to keep an eye on.

---

## The Scale

- Morale is internally a float, displayed as a **percentage**.
- It **starts at 50% (neutral)** for a new civilization.
- It has a hard **floor of 10%** — it can sink low, but never to zero.
- Its **ceiling is 100% + 5% per Wonder built**. With no wonders the cap is 100%; build wonders and the cap rises, letting morale climb higher and unlock a larger production bonus.

So a civilization with 4 wonders built can push morale as high as **120%**, while one with none tops out at **100%**.

---

## The Three Bands

Morale is centred on 50%. What it does depends on which band it sits in:

| Band | Morale Range | Effect on All Production |
|------|--------------|--------------------------|
| **Low** | below 25% | **Penalty** — ramps down to **×0.50** (half production) at the 10% floor |
| **Neutral** | 25% – 75% | **No effect** — production runs at its normal rate |
| **High** | above 75% | **Bonus** — ramps up to **+20%** as morale approaches the cap |

The effect scales **smoothly** within the high and low bands — it is not a flat step. Just above 75% the bonus is tiny; the full +20% is only reached as morale nears its cap. Likewise, just below 25% the penalty is mild; the full ×0.50 only applies at the 10% floor.

The neutral band is wide on purpose. Normal play sits comfortably inside 25%–75% with no morale effect at all — you only feel morale once you actively push it high or let it collapse.

---

## Drift to Neutral

Morale **drifts gently back toward 50% every tick**. This is the most important thing to understand about it:

- A high-morale **bonus must be earned and sustained**. If you stop building morale-restoring buildings, the bonus bleeds away as morale drifts back down to neutral. It is never a permanent freebie.
- A low-morale **penalty is self-healing**. Once you remove whatever was dragging morale down (fix the food deficit, shed excess military workers), the drift pulls morale back up toward neutral on its own.

In short: neutral is the resting state. You have to spend effort to live above it, and the game forgives you for dipping below it as long as you stop the bleeding.

---

## What Raises Morale

| Source | Effect |
|--------|--------|
| **Morale-restoring buildings** | The main lever. Worship buildings (shrines, temples, and their later-age equivalents) and culture/entertainment buildings lift morale **each tick just by existing** — no workers required. |
| **Good events** | A favourable event lifts morale. |
| **Advancing to a new age** | Reaching a new age gives a one-time morale boost. |

## What Lowers Morale

| Source | Effect |
|--------|--------|
| **Food starvation** | Running out of food drains morale each tick. |
| **Over-militarisation** | Military workers exceeding **30% of your population** drain morale — the further over the threshold, the faster the drain. |
| **Idle workforce** | More than **50% of your workers sitting idle** drains morale each tick. |
| **Bad events & catastrophes** | A negative event lowers morale; **enduring a catastrophe** costs morale. |

---

## The Key Lever: Morale-Restoring Buildings

Because morale always drifts back to 50%, the only way to **actively push it up into the bonus zone and keep it there** is to build **morale-restoring buildings** — the era-appropriate worship buildings (shrine/temple line) and culture/entertainment buildings. They raise morale a little every tick simply by standing, with no workers assigned.

Stack enough of them and their per-tick lift outpaces the drift-to-neutral, parking your morale in the high band for a sustained **+20% to all production**. Stop building them and morale slides back to neutral. Everything else on the "raises morale" list (good events, age advances) is a one-time nudge — morale-restoring buildings are the steady, controllable source.

---

## Where Morale Is Shown

- **Workers panel** (`workers`) — a coloured morale bar.
- **Villager sidebar** — the same coloured bar, always visible at a glance.
- **Load Game browser** — each save's detail pane shows its morale, so you can size up a civilization before loading it.

The bar is **green when morale is boosting production** (high band), **neutral in the middle**, and **red when it is penalising production** (low band).

---

## Managing Morale — Strategy

1. **Ignore it until you want the bonus.** Normal play sits in the neutral band. There is no penalty for being at 50%, so early on, spend your effort elsewhere.
2. **Keep food positive.** Starvation is the most common cause of a morale slide into the penalty band. Fix the food deficit and the drift heals the rest.
3. **Don't over-militarise.** Keep military workers comfortably under 30% of population — past that, morale drains faster the more lopsided your army gets. (See [Military](military.md).)
4. **Don't park idle workers.** More than half your population sitting idle drains morale on top of wasting their food. Assign them or `dismiss` them.
5. **Build worship and culture buildings to go positive.** When you want the +20% production swing, build morale-restoring buildings faster than morale drifts back to neutral. Sustaining the high band is an active choice, not a one-time purchase.
6. **Build wonders to raise the ceiling.** Each wonder adds +5% to the cap, so the more wonders you have, the higher the bonus morale can reach. (See [Wonders](wonders.md).)

---

## See Also

- [Workers & Domains](workers-and-domains.md#morale) — the production formula and how morale fits the worker system
- [Military](military.md) — the military-ratio morale drain in detail
- [Buildings](buildings.md) — which worship and culture buildings restore morale
- [Wonders](wonders.md) — how wonders raise the morale ceiling
