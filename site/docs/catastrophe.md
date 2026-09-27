# Catastrophe System

A catastrophe is a civilization-threatening event that forces a permanent choice: **Endure** or **Succumb**. Catastrophes strike when your civilization crosses into a new epoch, from the **Iron Era** on. You can't trigger one directly; the only way to choose one is to **Invite** it when a harbinger comes.

> **The Harbinger.** Every epoch whose transition can bring a catastrophe has harbingers who warn you of it, one figure per age, from the epoch's first age until the transition. While a harbinger is present you can **Appease** it (spend faith and culture to lower the real odds), **Brace** (spend resources so an Endure costs less) or **Invite** the catastrophe (guarantee it, for a deliberate Succumb). See [The Harbinger](harbinger.md).

When a catastrophe hits, nothing is destroyed yet. The game keeps running and the choice waits for you:

- A modal opens with the two choices. Each button shows its shortcut: **E** Endure, **S** Succumb. Tab or the arrow keys move between them and Enter picks the highlighted one. Endure is highlighted when the modal opens, so a stray Enter chooses Endure.
- **Esc** closes the modal without choosing, so you can look around first. The catastrophe stays pending.
- While a catastrophe is pending, the status bar shows a **☄ CATASTROPHE PENDING** badge, and **advancing to the next age and prestige are refused**. Type `catastrophe` to reopen the choice.
- The pending catastrophe is saved with your game. If you close the game, or come back after a long idle, the modal opens again when the save loads.

When the catastrophe rolls on the same advance that crosses into a new epoch, the age-advance splash shows first. The catastrophe modal opens once you dismiss the splash with any key, or when the splash times out after 20 seconds.

There is no "defer" button any more. Esc plus the `catastrophe` command does the same job without letting the choice be skipped.

---

## When It Triggers

There are 7 epochs across 22 ages (Cosmic spans 4 ages, the rest 3). Every run starts in the Stone Era, so a run has **6 epoch transitions**: into Iron, Steel, Electric, Digital, Neon and Cosmic. Each transition fires one **Epoch Event Roll**:

1. **Good or bad?** Faith fill decides the good-event chance: 40% below 25% faith, 50% between 25% and 75% (or when faith has no storage yet), 60% above 75%.
2. **On a bad roll**, a further **30% chance** escalates it to a catastrophe. Otherwise you get a Challenging event, applied immediately.

So the chance of a catastrophe at a transition is **18% at low faith, 15% at mid faith, 12% at high faith**.

Extra rules:

- **Iron Era gate.** No catastrophe before the epoch that contains the Iron Age. In practice the Stone Era never has one.
- **One per epoch per run.** Each epoch's transition rolls once per run. Succumb and prestige start a new run, so the epochs roll again.
- **Never overwritten.** A new catastrophe can't replace one that is still pending. You can't reach the next transition while one is pending anyway.
- **At most 6 per run**, one for each epoch from Iron to Cosmic.
- **The harbinger can change the odds.** Each level of Appease multiplies the chance by 0.6 (two levels at most). Invite makes it certain. See [The Harbinger](harbinger.md).

The `catastrophe` command (with nothing after it) shows the risk for your next transition when nothing is pending: the odds as a figure from the Industrial Age on, a low / medium / high severity before it. The Epoch tab shows the same line. While a harbinger is present, both repeat its warning, so they can't give away a false prophet. See [Epochs](epochs.md) for the event tables.

---

## The Catastrophes

Each epoch has a named catastrophe:

| Epoch | Catastrophe | Flavor |
|-------|-------------|--------|
| Stone Era | The Great Meteor | Not reachable: no catastrophes before the Iron Era. |
| Iron Era | The Great Plague | A devastating plague sweeps your cities. The streets fall silent. |
| Steel Era | The World War | Industrial warfare tears civilization apart. The factories are ash. |
| Electric Era | The Nuclear Exchange | Nations unleash the atom. Cities become glass. |
| Digital Era | The Great Hack | Every system falls silent. The AIs turn on their creators. |
| Neon Era | Corporate Armageddon | The megacorps end the world with a fusion bomb. |
| Cosmic Era | The Reality Tear | Exotic matter destabilizes spacetime. Reality cracks open. |

The name is flavor. Endure and Succumb work the same way in every epoch; only the legacy bonus differs.

---

## Endure

Pay a cost and keep your civilization.

- **20% of your buildings destroyed**: `floor(non-wonder buildings / 5)`, at least 1 if you have any. Wonders are never destroyed and don't count toward the total. Brace lowers this to 15% or 10%.
- **Workers of destroyed buildings go idle** first, the same as when you sell a building.
- **All unlocked resources drop to 15%** of their stored amounts. Brace raises this to 30% or 45%.
- **25% of the worker pool is lost.** There is one worker pool, so every building loses the same share of its assigned workers, whatever its domain (food, knowledge, military and so on).
- **Reconstruction Effort**: production −10% for 216 ticks.
- **Morale −10 points.**
- **Survived** marker on the epoch badge and a line in the civilization log.

Age, research, wonders and prestige are untouched.

### Brace

If the harbinger warned you and you paid to **Brace**, Endure costs less. Brace only changes the two numbers below; worker loss, the debuff and the morale hit stay the same.

| Brace level | Buildings destroyed | Stored resources kept |
|-------------|---------------------|-----------------------|
| none | 20% | 15% |
| 1 | 15% | 30% |
| 2 | 10% | 45% |

Buildings destroyed are rounded down, with at least 1 if you have any. The Brace is attached to the pending catastrophe, so it still applies if you press Esc and Endure later, or save and load first. It does nothing for Succumb. See [Brace](harbinger.md#brace-soften-an-endure).

The −10% applies to every building, including the ones that survived, for the full 216 ticks. Negative production modifiers are floored at 10% of base, but a single −10% lands in full whatever other bonuses you hold. The same flooring covers per-resource rate modifiers and gather rate, and any active debuff shows in the Active Multipliers panel.

If morale was already low, the −10 can push it into the low band, where output is penalized. Morale drifts back toward 50% on its own; a food surplus speeds that up.

**When to Endure:** your civilization is big, and a full reset would cost more than the legacy bonus is worth. Also when you already hold this epoch's legacy bonus.

### How the destruction is picked

- Every built non-wonder building instance goes into a pool, in a fixed order (sorted by building key).
- The pool is shuffled with your run's seeded random generator and the first N are destroyed.
- Destroyed buildings are removed entirely; they don't become ruins.

Because the generator is seeded per run, the same run state destroys the same buildings.

**Recovery checklist:**

1. The log lists every lost building by name. Rebuild food and housing first.
2. Reassign or recruit workers for the rebuilt capacity.
3. Wait out the 216-tick reconstruction debuff; it can't be removed early.

---

## Succumb

Let the civilization fall, and keep something permanent.

- **Up to 8 buildings become ruins**, picked at random from your non-wonder buildings (same seeded, fixed-order pool as Endure). Ruins produce at 50% of base rate with no workers.
- **Ruin cap: 24.** Ruins carry across Succumb and prestige, but the total never exceeds 24. When new ruins push past the cap, the **lowest-value ruins crumble first**: earliest age first, then lowest base output. A late-game fall replaces primitive rubble instead of being thrown away. Saves from before the cap are trimmed the same way when loaded.
- **Legacy bonus** for the epoch (table below), permanent.
- **Ancient Knowledge**: +25% research speed for each distinct epoch you have succumbed in. Succumbing twice in the same epoch doesn't add another 25%.
- **Full reset** to the Primitive Age: buildings, resources, workers, research, milestones, events, build queue. You start with 15 food and 12 wood, plus prestige starting bonuses.
- **No prestige points are earned.** Your prestige level, points and upgrades are kept.
- A line in the civilization log.

Morale restarts at 50%.

### What carries forward

| Item | After Succumb |
|------|---------------|
| Prestige level, points and upgrades | Kept |
| Ruins | Kept, plus up to 8 new, capped at 24 |
| Legacy flags (bonuses and Ancient Knowledge) | Kept, plus this epoch |
| Civilization log | Kept |
| Epoch event history | Kept |
| Resources, buildings, workers | Reset |
| Research | Reset |
| Milestones and chains | Reset |

### Ancient Knowledge

The research bonus comes straight from your legacy flags: +25% per epoch flagged. It is recomputed whenever it's needed, so save/load, Succumb and prestige can't drop or double it. It shows as **Legacy** under Research Speed in the Active Multipliers panel.

From Iron to Cosmic there are 6 epochs you can succumb in, so the most you can earn now is **+150%**. A save that earned the Stone Era legacy before the Iron gate existed keeps it (+175% total).

**When to Succumb:** you don't yet hold this epoch's legacy, and the reset is cheap for you. A catastrophe always arrives right as you enter an epoch, so the question is how much of the run you'd be giving up.

---

## Legacy Bonus Table

| Epoch | Resources boosted | Bonus |
|-------|-------------------|-------|
| Stone Era | wood, stone | +20% each (only on saves that earned it before the Iron gate) |
| Iron Era | iron | +20% |
| Steel Era | steel, coal | +25% each |
| Electric Era | electricity, uranium | +25% each |
| Digital Era | data, titanium ore | +30% each |
| Neon Era | plasma, dark matter crystals | +30% each |
| Cosmic Era | dark matter | +35% |

Rate bonuses add to your other `<resource>_rate` bonuses and apply from tick 1 of every later run, including after prestige. A new game (wiping the save) clears them.

---

## Faith and the Odds

| Faith fill | Good event | Catastrophe at the transition |
|------------|------------|-------------------------------|
| under 25% | 40% | 18% |
| 25–75% (or no faith storage) | 50% | 15% |
| over 75% | 60% | 12% |

Over the 6 transitions of a run, high faith against low faith is roughly a third of a catastrophe fewer. More important, high faith also buys better good events. See [Faith](faith.md).

A harbinger's Appease multiplies these chances by 0.6 per level (0.36 at two levels), and Invite makes the catastrophe certain. See [The Harbinger](harbinger.md).

---

## Civilization Log

Every resolved catastrophe adds a line:

```
Tick N — Endured <Catastrophe> (<Epoch>). N buildings lost.
Tick N — Succumbed to <Catastrophe> (<Epoch>). Civilization reset. Legacy bonus earned.
```

The log survives Succumb and prestige. The Stats panel counts **Survived** (Endured) and **Succumbed** from these lines, so a pending catastrophe counts as neither.

The Epoch tab's history marks each past epoch's catastrophe as **Survived**, **Succumbed** or **Pending**. A catastrophe from an older save whose outcome was never stored shows as "outcome not recorded" rather than guessing.

---

## Strategy

### Endure beats Succumb when

- your civilization is large and deep into its run,
- you already hold this epoch's legacy (a repeat Succumb there adds nothing but ruins),
- you're close to a milestone chain that a reset would wipe,
- you braced when the harbinger warned you, so Endure costs less.

### Succumb beats Endure when

- the reset is cheap (early epochs, or a run you were going to prestige soon anyway),
- you don't hold this epoch's legacy yet,
- you have plenty of runs ahead to cash in the research bonus.

### Pending is a pause, not a dodge

Esc lets you check your buildings and resources before you commit, and the game keeps producing meanwhile. But you can't advance or prestige until you choose.

### Morale

Enduring costs 10 points of morale. Keep food positive, avoid over-militarizing (military above 30% of population drains morale), and build worship and culture buildings to climb back. See [Morale](morale.md).

### The long game

Succumbing once in each epoch from Iron to Cosmic, over several runs, collects all six reachable legacy bonuses and +150% research speed. You don't have to wait for the rolls: when a harbinger comes, **Invite** guarantees the catastrophe at that transition, so you can pick which epochs to fall in. Each run after that starts with those bonuses and up to 24 ruins producing from tick 1.
