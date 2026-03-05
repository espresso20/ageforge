# Catastrophe System

A catastrophe is a civilization-threatening event that forces you to make a permanent, irreversible choice. Catastrophes occur at epoch transitions — every 3 ages, when your civilization crosses into a new epoch — and can also be invoked voluntarily at any time during an epoch.

When a catastrophe fires, the game pauses progression and presents a modal with three choices. You cannot advance ages until the catastrophe is resolved.

---

## When It Triggers

Epochs span 3 ages each (7 epochs total across 22 ages). At every epoch transition, an **Epoch Event Roll** fires. The roll can land good or bad based on your faith level:

- A **bad roll** (60/50/40% depending on faith) → 70% chance of a Challenging event, **30% chance of a Catastrophe**

So a catastrophe is not guaranteed at every epoch — it is a weighted roll outcome. Higher faith reduces bad-roll probability, which reduces the chance a catastrophe appears at all.

See [Epochs](epochs.md) for full event tables and faith thresholds.

---

## The Three Choices

### Endure

Pay a cost and keep your civilization intact.

- 20% of your buildings destroyed (chosen randomly)
- Resources reduced to 15% of current amounts
- Workers reduced by 25%
- Production debuff −10% for 216 ticks
- Epoch marked as "Survived" (✓) in your history

**When to choose Endure:** When your civilization is large and a full reset would be extremely costly. The resource and building losses are painful but survivable, especially with high storage and redundant buildings.

---

### Succumb

Accept total reset and gain permanent power.

- Full civilization reset — resources, buildings, workers, and research all reset to zero
- 8 **ruins** placed in your world — ancient remnants that produce at 50% base rate with no workers required
- **Ancient Knowledge** — permanent +25% research speed for all future runs (including prestige resets)
- **Legacy Bonus** — permanent production boost for this epoch's primary resource(s), stacking across multiple runs
- Epoch marked as "Succumbed" in your history

**When to choose Succumb:** When you want to lock in the legacy bonus and your civilization is still small enough that a reset isn't devastating. The ruins and Ancient Knowledge bonuses compound across all future runs — early Succumbs often yield better long-term returns than enduring.

---

### Defer

Close the modal without deciding.

- The catastrophe stays **pending**
- You will be prompted again on your next login
- **You cannot advance ages while a catastrophe is pending**

**When to choose Defer:** When you need time to think, or want to continue accumulating resources before choosing. Be aware that age advancement is blocked until you resolve the choice.

---

## Faith & The Odds

Your **faith level** directly controls the probability that any epoch transition rolls good or bad:

| Faith Level | Chance of Good Event |
|-------------|---------------------|
| 0–25% (No Faith / Dim Faith) | 40% good |
| 26–75% (Low Faith / Mid Faith) | 50% good |
| 76–100% (Strong Faith / Faith Full) | 60% good |

A bad roll doesn't guarantee a catastrophe — it gives a 70% chance of a Challenging event and only a 30% chance of a Catastrophe. Maintaining 76%+ faith heading into an epoch transition gives you the best possible odds against bad outcomes.

> Invest in Faith lineage buildings before every epoch transition. The difference between 40% and 60% good-event chance is enormous over a full run.

See [Faith](faith.md) for the full Faith system and building lineage.

---

## Legacy Bonuses

Legacy bonuses are earned by Succumbing to a catastrophe during a given epoch. They are **permanent** — they survive prestige resets and stack across multiple Succumb runs.

| Epoch | Legacy Bonus |
|-------|-------------|
| Stone Era | wood +20%, stone +20% |
| Iron Era | iron +20% |
| Steel Era | steel +20%, coal +20% |
| Electric Era | electricity +20% |
| Digital Era | data +20%, oil +20% |
| Neon Era | dark_matter +20% |
| Cosmic Era | dark_matter +35% |

Each epoch's bonus applies to the primary resources of that epoch's progression. Because they stack, a player who Succumbs in multiple epochs accumulates compounding production multipliers across all future runs.

---

## Voluntary Catastrophe

You can trigger a catastrophe yourself at any time during an epoch:

```
catastrophe invoke
```

This bypasses the epoch transition roll entirely and immediately presents the Endure / Succumb / Defer modal. One catastrophe per epoch maximum — you cannot invoke a second one after resolving the first.

**Why invoke voluntarily?**

- You want to Succumb early while your civilization is still small, minimizing the reset cost
- You want to lock in a legacy bonus before reaching the natural epoch transition
- You are in a poor-faith, early-epoch state where the reset cost is low

> The best time to voluntarily Succumb is during the Stone Era when your civilization is small. The legacy bonus (wood +20%, stone +20%) is earned at minimal cost, and Ancient Knowledge compounds across the entire run.

---

## Tips

- **Faith first** — Build Faith lineage buildings before every epoch transition. 76%+ faith gives the best odds against catastrophes appearing at all.
- **Succumb early, not late** — Resetting a small civilization is far less painful than resetting a large one. If you plan to Succumb, do it early in the epoch.
- **Ruins produce passively** — After a Succumb, your 8 ruins work with zero workers. This gives you a production head start in your new run before you rebuild your worker base.
- **Ancient Knowledge stacks with other bonuses** — The +25% research speed from Succumb combines with Great Library, Global Network, and prestige research upgrades for dramatic tech acceleration.
- **You can Defer indefinitely** — If a catastrophe appears at an inconvenient time, Defer and prepare. Just remember age advancement is blocked until you choose.
