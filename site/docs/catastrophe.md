# Catastrophe System

A catastrophe is a civilization-threatening event that forces a permanent choice: **Endure** or **Succumb**. Catastrophes strike when your civilization crosses into a new epoch, from the **Iron Era** on. You can't trigger one directly; the only way to choose one is to **Invite** it when a harbinger comes.

> **The Harbinger.** Every epoch whose transition can bring a catastrophe, and the Cosmic Era, whose passage is prestige, has harbingers who warn you of it, one figure per age, from the epoch's first age until the transition. While a harbinger is present you can **Appease** it (spend faith and culture to lower the real odds), **Brace** (spend resources so an Endure costs less) or **Invite** the catastrophe (guarantee it, for a deliberate Succumb). See [The Harbinger](harbinger.md).

When a catastrophe hits, nothing is destroyed yet. The game keeps running and the choice waits for you.

A window opens with the two choices, Endure and Succumb. Press E or S, or move between them with Tab or the arrow keys and press Enter. Endure is highlighted when the window opens, so a stray Enter chooses Endure. Esc closes the window without choosing, so you can look around first; the catastrophe stays pending.

While a catastrophe is pending, the status bar shows a **☄ CATASTROPHE PENDING** badge, and the game refuses to advance to the next age or prestige. Type `catastrophe` to reopen the choice. The pending catastrophe is saved with your game, so if you close the game or come back after a long idle, the window opens again when the save loads.

When the catastrophe rolls on the same advance that crosses into a new epoch, the age-advance splash shows first. The catastrophe window opens once you dismiss the splash with any key, or when the splash times out after 20 seconds.

---

## When It Triggers

There are 7 epochs across 22 ages (Cosmic spans 4 ages, the rest 3). Every run starts in the Stone Era, so a run has **6 epoch transitions**: into Iron, Steel, Electric, Digital, Neon and Cosmic. Each transition fires one **Epoch Event Roll**:

1. **Good or bad?** Faith fill decides the good-event chance: 40% below 25% faith, 50% between 25% and 75% (or when faith has no storage yet), 60% above 75%.
2. **On a bad roll**, a further **30% chance** escalates it to a catastrophe. Otherwise you get a Challenging event, applied immediately.

So the chance of a catastrophe at a transition is **18% at low faith, 15% at mid faith, 12% at high faith**.

A few more rules apply. No catastrophe can strike before the epoch that contains the Iron Age, so in practice the Stone Era never has one. Each epoch's transition rolls once per run; Succumb and prestige start a new run, so the epochs roll again. That makes 6 catastrophes the most a run can have, one for each epoch from Iron to Cosmic. A new catastrophe never replaces one that is still pending, and you can't reach the next transition while one is pending anyway. Finally, the harbinger can change the odds: each level of Appease multiplies the chance by 0.6 (two levels at most), and Invite makes it certain. See [The Harbinger](harbinger.md).

The `catastrophe` command (with nothing after it) shows the risk for your next transition when nothing is pending: the odds as a figure from the Industrial Age on, a low / medium / high severity before it. In the Cosmic Era it shows the risk of the Last Passage instead, along with how full your faith is. The Epoch panel shows the same line. While a harbinger is present, both repeat its warning, so they can't give away a false prophet. See [Epochs](epochs.md) for the event tables.

---

## The Last Passage

The Cosmic Era has no next epoch, so its passage is prestige. When you confirm prestige in the Cosmic Era, the **Last Passage** rolls with the same odds (18% / 15% / 12% by faith, ×0.6 per level of Appease, certain if invited). If it comes, the prestige waits for your choice:

- **Endure** completes the prestige but keeps only 50% of the run's prestige points (70% or 85% if you braced).
- **Succumb** completes the prestige with no points from this run and grants the **Cosmic Legacy**, a permanent +10% production. You can earn it once; after that, Succumb is closed.

It behaves like a pending catastrophe: Esc closes the choice, a **☄ LAST PASSAGE** badge shows in the status bar, the bare `catastrophe` command reopens it, and it is saved with your game. Unlike a catastrophe, it blocks only prestige. Either choice adds a line to the civilization log and counts as Endured or Succumbed. See [The Last Passage](prestige.md#the-last-passage).

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

Pay a cost and keep your civilization. Enduring costs you:

| Loss | Amount |
|------|--------|
| Buildings destroyed | 20% of your non-wonder buildings, rounded down (`floor(non-wonder buildings / 5)`), at least 1 if you have any. Wonders are never destroyed and don't count. |
| Stored resources | Every unlocked resource drops to 15% of its stored amount. |
| Workers | 25% of the worker pool is lost. |
| Production | Reconstruction Effort: all production −10% for 216 ticks. |
| Morale | −10 points. |

Workers of the destroyed buildings go idle first, the same as when you sell a building. There is one worker pool, so the 25% loss takes the same share of assigned workers from every building, whatever its domain (food, knowledge, military and so on). Age, research, wonders and prestige are untouched. You earn the **Endured** marker on the epoch badge and a line in the civilization log.

### Brace

If the harbinger warned you and you paid to **Brace**, Endure costs less. Brace changes only the share of buildings destroyed and the share of stored resources kept; worker loss, the production debuff and the morale hit stay the same.

| Brace level | Buildings destroyed | Stored resources kept |
|-------------|---------------------|-----------------------|
| none | 20% | 15% |
| 1 | 15% | 30% |
| 2 | 10% | 45% |

Buildings destroyed are rounded down, with at least 1 if you have any. The Brace is attached to the pending catastrophe, so it still applies if you press Esc and Endure later, or save and load first. It does nothing for Succumb. See [Brace](harbinger.md#brace-soften-an-endure).

The −10% applies to every building, including the ones that survived, for the full 216 ticks. Negative production modifiers are floored at 10% of base, but a single −10% lands in full whatever other bonuses you hold. The same flooring covers per-resource rate modifiers and worker output, and any active debuff shows in the Active Multipliers panel.

If morale was already low, the −10 can push it into the low band, where output is penalized. Morale drifts back toward 50% on its own; a food surplus speeds that up.

**When to Endure:** your civilization is big, and a full reset would cost more than the legacy bonus is worth. Also when you already hold this epoch's legacy bonus.

### How the destruction is picked

Every built non-wonder building goes into a pool in a fixed order (sorted by building key). The pool is shuffled with your run's seeded random generator and the first N are destroyed. Destroyed buildings are removed entirely; they don't become ruins. Because the generator is seeded per run, the same run state destroys the same buildings.

**Recovery checklist:**

1. The log lists every lost building by name. Rebuild food and housing first.
2. Reassign or recruit workers for the rebuilt capacity.
3. Wait out the 216-tick reconstruction debuff; it can't be removed early.

---

## Succumb

Let the civilization fall, and keep something permanent.

Up to 8 of your non-wonder buildings become **ruins**, picked at random from the same seeded, fixed-order pool as Endure. Ruins produce at 50% of base rate with no workers. They carry across Succumb and prestige, but the total is capped at 24. When new ruins push past the cap, the lowest-value ruins crumble first (earliest age first, then lowest base output), so a late-game fall replaces primitive rubble. Older saves over the cap are trimmed the same way when loaded.

You also get the epoch's **legacy bonus** (table below), permanently, and **Ancient Knowledge**: +25% research speed for each distinct epoch you have succumbed in. Succumbing twice in the same epoch doesn't add another 25%.

Then the civilization resets to the Primitive Age: buildings, resources, workers, research, milestones, events and the build queue. You start with 15 food and 12 wood, plus prestige starting bonuses. No prestige points are earned, but your prestige level, points and upgrades are kept. Morale restarts at 50%, and the civilization log gets a line.

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

From Iron to Cosmic there are 6 epochs you can succumb in, so the most you can earn is **+150%**. A save that earned the Stone Era legacy before catastrophes were limited to the Iron Era on keeps it (+175% total).

**When to Succumb:** you don't yet hold this epoch's legacy, and the reset is cheap for you. A catastrophe always arrives right as you enter an epoch, so the question is how much of the run you'd be giving up.

---

## Legacy Bonus Table

| Epoch | Resources boosted | Bonus |
|-------|-------------------|-------|
| Stone Era | wood, stone | +20% each (only on older saves that earned it) |
| Iron Era | iron | +20% |
| Steel Era | steel, coal | +25% each |
| Electric Era | electricity, uranium | +25% each |
| Digital Era | data, titanium ore | +30% each |
| Neon Era | plasma, dark matter crystals | +30% each |
| Cosmic Era | dark matter | +35% |

Rate bonuses add to your other rate bonuses for that resource and apply from tick 1 of every later run, including after prestige. A new game (wiping the save) clears them.

---

## Faith and the Odds

| Faith fill | Good event | Catastrophe at the transition |
|------------|------------|-------------------------------|
| under 25% | 40% | 18% |
| 25 to 75% (or no faith storage) | 50% | 15% |
| over 75% | 60% | 12% |

Over the 6 transitions of a run, high faith against low faith is roughly a third of a catastrophe fewer. High faith also gives you more good events. See [Faith](faith.md).

A harbinger's Appease multiplies these chances by 0.6 per level (0.36 at two levels), and Invite makes the catastrophe certain. See [The Harbinger](harbinger.md).

---

## Civilization Log

Every resolved catastrophe adds a line to the civilization log with the tick, the outcome (Endured or Succumbed), the catastrophe and its epoch, and then either the number of buildings lost (Endure) or a note that the civilization reset (Succumb). Endure and Succumb at the Last Passage add a line of their own. The log survives Succumb and prestige. The Stats panel counts **Endured** and **Succumbed** from these lines, so a pending catastrophe counts as neither.

The Epoch panel's history marks each past epoch's catastrophe as **Endured**, **Succumbed** or **Pending**. A catastrophe from an older save whose outcome was never stored shows as "outcome not recorded" rather than guessing.

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
