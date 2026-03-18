# Catastrophe System

A catastrophe is a civilization-threatening event that forces you to make a permanent, irreversible choice. Catastrophes occur at epoch transitions — every 3 ages, when your civilization crosses into a new epoch — and can also be invoked voluntarily at any time during an epoch.

When a catastrophe fires, the game pauses progression and presents a modal with three choices: **Endure**, **Succumb**, or **Defer**. You cannot advance ages until the catastrophe is resolved.

---

## When It Triggers

Epochs span 3 ages each (7 epochs total across 22 ages). At every epoch transition, an **Epoch Event Roll** fires. The roll can land good or bad based on your faith level:

- A **bad roll** (60/50/40% depending on faith) → 70% chance of a Challenging event, **30% chance of a Catastrophe**

So a catastrophe is not guaranteed at every epoch — it is a weighted roll outcome. Higher faith reduces bad-roll probability, which reduces the chance a catastrophe appears at all.

Additionally: only **one catastrophe per epoch** can occur, whether random or voluntary. Once one has been resolved (or invoked), no further catastrophe can trigger for the remainder of that epoch.

See [Epochs](epochs.md) for full event tables and faith thresholds.

---

## The Seven Catastrophes

Each epoch has a named catastrophe with its own flavor:

| Epoch | Catastrophe Name | Flavor |
|-------|-----------------|--------|
| Stone Era | The Great Meteor | A celestial body has struck your settlement. The sky burns. Your people scatter. |
| Iron Era | The Great Plague | A devastating plague sweeps your cities. The streets fall silent. |
| Steel Era | The World War | Industrial warfare tears civilization apart. The factories are ash. |
| Electric Era | The Nuclear Exchange | Nations unleash the atom. Cities become glass. |
| Digital Era | The Great Hack | Every system falls silent. The AIs turn on their creators. |
| Neon Era | Corporate Armageddon | The megacorps end the world with a fusion bomb. |
| Cosmic Era | The Reality Tear | Exotic matter destabilizes spacetime. Reality cracks open. |

---

## The Three Choices

### Endure

Pay a cost and keep your civilization intact.

- **20% of buildings destroyed** — chosen uniformly at random from all non-wonder buildings; wonders are never destroyed
- **All resources reduced to 15% of current stored amounts** — every unlocked resource drops to 15% simultaneously
- **Workers reduced by 25%** — a quarter of your workforce is lost
- **Production debuff −10% for 216 ticks** — a timed `Reconstruction Effort` event applies −10% to all production for the recovery period
- **Survived marker** recorded in your civilization history for that epoch

Endure is painful but survivable. A well-developed civilization with many buildings loses only a fraction of its output — the production penalty is temporary, buildings grow back, and workers can be re-recruited.

**Morale impact:** Endure also deals a **−0.10 morale hit** on top of the building destruction. If morale was already low heading into a catastrophe, this can push it toward the 0.10 floor and significantly slow recovery. Prioritise food surplus after Enduring to stabilise morale alongside rebuilding.

**When to choose Endure:** When your civilization is large and a full reset would cost you more than the legacy bonus is worth. Late in an epoch, with 20+ ages progressed and significant building counts, Endure preserves enormous progress that Succumb would erase.

---

### Succumb

Accept total reset and gain permanent power.

- **Full civilization reset** — resources, buildings, workers, research, milestones, and build queue all reset to zero
- **8 ruins generated** from your current buildings and carried into the new run (produced at 50% base rate, no workers required)
- **Ancient Knowledge** — permanent +25% research speed for all future runs (stacks additively with each Succumb)
- **Epoch Legacy Bonus** — permanent production multiplier for the primary resources of this epoch, active from tick 1 of every future run
- **Epoch Succumbed** marker recorded in civilization history

The full reset is real — you return to Primitive Age with 15 food and 12 wood. But the permanent bonuses, ruins, and prestige upgrades all survive.

**Morale impact:** After Succumb, morale **resets to 0.50** — below the default 1.0 of a fresh run. You begin the new civilization at half morale output and must recover through food surplus ticks (+0.002/tick), age advances (+0.08 each), and passive recovery (+0.001/tick). Plan early food production accordingly.

**What carries forward after Succumb:**

| Item | Survives? |
|------|----------|
| Prestige level and upgrade tiers | Yes |
| Ruins (up to 8 new from current run) | Yes |
| Epoch legacy bonuses (all prior + new) | Yes |
| Ancient Knowledge bonus | Yes (+25% stacked) |
| Catastrophe/civilization history log | Yes |
| Resources, buildings, workers | No — reset to zero |
| Research / tech tree | No — reset |
| Milestones and chains | No — reset |
| Epoch event history | Yes (carried forward) |

**When to choose Succumb:** When your civilization is still small (early in an epoch) and the reset cost is low relative to the permanent bonus you'll gain. The Ancient Knowledge +25% research speed compounds across every single future run — an early Succumb in the Stone Era costs little but pays dividends forever. Multiple Succumbs across different epochs stack legacy bonuses for dramatic long-term acceleration.

---

### Defer

Close the modal without deciding.

- The catastrophe stays **pending** — nothing changes immediately
- You will be prompted again the next time you open the game
- **Age advancement is blocked** until you resolve the pending catastrophe
- You can continue gathering resources and playing otherwise

Defer has no cost and no timer — you can hold a catastrophe pending indefinitely. This is useful if you want to accumulate more resources, push a few more milestone goals, or simply need to think through the choice.

**When to choose Defer:** When you're mid-milestone chain, when you want to ensure faith is high enough to benefit from Endure or Succumb more optimally, or when you'd prefer to make the choice on your next session.

> Age advancement is blocked until the catastrophe is resolved. Defer is a pause, not an escape — you must eventually Endure or Succumb.

---

## Endure in Detail

The 20% building destruction is handled by `DestroyRandom`:

- A pool of all current building instances is assembled (each copy of each building is a separate entry)
- Wonders are explicitly excluded — they cannot be destroyed
- The pool is shuffled randomly
- The first N = `floor(total_buildings / 5)` entries are destroyed (minimum 1 if any buildings exist)
- The destroyed buildings are removed entirely — they do not become ruins

After destruction, the production debuff event (`endure_reconstruction`) is injected:

```
Effect: production_all -10% for 216 ticks
```

This is separate from the building loss. Even buildings that survived are producing at 90% for the reconstruction period. Plan for both effects simultaneously.

**Recovery checklist after Endure:**
1. Check which buildings were destroyed (the log lists every lost building by name)
2. Rebuild the most impactful lost buildings first (usually food and worker capacity)
3. Re-recruit workers if needed to fill rebuilt capacity
4. Wait out the 216-tick reconstruction debuff (there is no way to remove it early)
5. Resume normal progression once the timed event expires

---

## Succumb in Detail

The 8 ruins are generated before the reset:

- A pool of all current building instances is assembled (excluding wonders)
- The pool is shuffled randomly
- Up to 8 instances are selected and moved to the ruins map
- Their counts are removed from the active building count (so workers can no longer be assigned to them)
- After the reset, the ruins are restored into the new civilization's building state

Ruins produce at **50% of base rate with no worker scaling**:

```
ruin_production = base_rate × ruin_count × 0.50
```

No workers are needed. No assignment is required. They produce automatically from tick 1.

**After reset, the following sequence applies:**
1. Epoch legacy bonus is applied to `permanentBonuses` map (affects all future production calculations)
2. Ancient Knowledge is stacked: `permanentBonuses["research_speed"] += 0.25`
3. All prior legacy bonuses from previous Succumbs are re-applied
4. Age unlocks for Primitive Age are applied
5. Starting resources: 15 food, 12 wood (+ prestige starting_food/starting_wood bonuses)
6. Ruins are loaded into the building state

---

## Defer in Detail

Defer has no mechanical cost on the engine side. When you choose Defer:

- `pendingCatastrophe` remains set to the current epoch key
- The UI hides the modal
- All normal gameplay continues (resource production, ticking, building, etc.)
- Age advancement is blocked by a check against `pendingCatastrophe != ""`

You can Defer indefinitely. There is no counter, no escalating penalty, no second catastrophe that triggers if you wait too long. The only cost of Defer is the age advancement block.

---

## Legacy Bonus Table

Epoch legacy bonuses earned from Succumb are additive with all other production multipliers:

| Epoch | Resources Boosted | Bonus |
|-------|-----------------|-------|
| Stone Era | wood, stone | +20% each |
| Iron Era | iron | +20% |
| Steel Era | steel, coal | +25% each |
| Electric Era | electricity, uranium | +25% each |
| Digital Era | data, titanium ore | +30% each |
| Neon Era | plasma, dark matter crystals | +30% each |
| Cosmic Era | dark matter | +35% |

These bonuses apply to the production rate calculation in every future tick, regardless of prestige resets. They are stored in `legacyBonuses` (a map of epoch key → true) and re-applied via `reapplyLegacyBonuses()` after every reset.

A player who has Succumbed in every epoch has stacked production bonuses across all primary resources, making mid-game resource phases almost instant.

---

## Faith and the Odds

Your **faith level** directly controls the probability that any epoch transition rolls good or bad:

| Faith Level | Threshold | Chance of Good Event |
|-------------|-----------|---------------------|
| 0–24.9% | No Faith / Dim Faith | 40% good |
| 25–75% | Low Faith / Mid Faith | 50% good |
| 75.1–100% | Strong Faith / Faith Full | 60% good |

A bad roll doesn't guarantee a catastrophe — it gives a 70% chance of a Challenging event and only a 30% chance of a Catastrophe. Maintaining 76%+ faith heading into an epoch transition gives you the best possible odds:

- At 60% good: only 40% bad. Of those bad rolls, 30% escalate → **12% catastrophe chance**
- At 40% good: 60% bad. Of those, 30% escalate → **18% catastrophe chance**

That 6% difference across 7 epoch transitions means roughly half a catastrophe more or less across a full run.

> Invest in Faith lineage buildings before every epoch transition. A full faith bar before each transition is the single best risk-reduction measure available.

See [Faith](faith.md) for the full Faith system and building lineage.

---

## Voluntary Catastrophe

You can trigger a catastrophe yourself at any time during an epoch:

```
catastrophe invoke
```

This bypasses the epoch transition roll entirely and immediately presents the Endure / Succumb / Defer modal. One catastrophe per epoch maximum — you cannot invoke a second one after resolving the first, and a voluntary invocation counts toward that limit.

**Why invoke voluntarily?**

- You want to Succumb while your civilization is small, minimizing reset cost
- You want to lock in a legacy bonus before reaching the natural epoch transition
- You're still in the Stone Era with a tiny civilization — resetting now costs almost nothing
- You want to test Endure consequences without waiting for an epoch roll

> The canonical optimal play: invoke catastrophe voluntarily in the Stone Era (first available epoch), choose Succumb. The legacy bonus (wood +20%, stone +20%) and Ancient Knowledge (+25% research speed) are earned when your civilization is at its smallest, making the reset nearly costless. These bonuses then compound through every subsequent run.

---

## Civilization History

Every catastrophe resolution is logged in your civilization history. The entry records:

```
Tick N — Endured/Succumbed to <Catastrophe Name> (<Epoch Name>). N buildings lost. / Civilization reset. Legacy bonus earned.
```

This history survives all resets, including prestige. You can review it to track your run lineage and see which epochs you've resolved catastrophes in.

---

## Strategy Guide

### When Endure beats Succumb

- You're late in a full run with 100+ buildings and 30+ ages progressed
- You've already Succumbed in this epoch during a previous life and claimed the legacy bonus
- Your civilization is large enough that the ancient knowledge +25% research speed marginal gain is small relative to the reset cost
- You're close to a milestone chain completion that resets on Succumb

### When Succumb beats Endure

- Your civilization is small (under 30 buildings, early in the epoch)
- You haven't yet claimed the legacy bonus for this epoch
- You have a high number of ruins already, making the head start on reset very strong
- You're behind on your expected progression and want to leverage compounding permanent bonuses

### Managing morale through catastrophes

Both Endure and Succumb impose morale costs that compound with the other penalties. The recommended recovery path in either case is the same: keep food in surplus, avoid over-militarising (military workers above 30% of population drain morale further), and let age advances (+0.08 each) do the heavy lifting. Surviving multiple catastrophes in a single run without attending to morale recovery is a common reason civilizations stall — low morale suppresses all worker output until the food situation stabilises.

See [Workers & Domains](workers-and-domains.md) for the full morale system.

### When Defer is correct

- You're mid-build of an important wonder that would be lost to Endure
- You want to drain your food/wood into starting resource prestige upgrades before resetting
- You're one age advance away from a milestone and want to complete it first
- It's literally just bad timing and you need 10 more minutes

### The Optimal Long-Term Path

The highest long-term output comes from collecting all 7 legacy bonuses by Succumbing once per epoch across different runs, then running full prestige loops with all bonuses stacked. After 7+ runs you'll have:

- +20% wood, +20% stone (Stone Era)
- +20% iron (Iron Era)
- +25% steel, +25% coal (Steel Era)
- +25% electricity, +25% uranium (Electric Era)
- +30% data, +30% titanium ore (Digital Era)
- +30% plasma, +30% dark matter crystals (Neon Era)
- +35% dark matter (Cosmic Era)
- +25% research speed per Succumb × (number of Succumbs)

This makes the mid-game effectively instant and late-game resources trivial to accumulate.
