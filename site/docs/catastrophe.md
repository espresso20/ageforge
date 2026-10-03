# Catastrophe System

A catastrophe is a civilization-threatening event that forces a permanent choice: **Endure** or **Succumb**. From the **Iron Era** on, a doom may be fated in secret to strike at any moment inside an era, and a harbinger always comes to warn you first. You can't trigger one directly; the only way to choose one is to **Invite** it when a harbinger comes.

> **The Harbinger.** A doom never strikes unannounced. Some while before it does, a harbinger arrives, and each age's figure takes up the warning until the doom resolves. In the Cosmic Era a second thread also warns of the Last Passage at your next prestige. While a harbinger is present you can **Appease** it (spend faith and culture to lower the real odds), **Brace** (spend resources so an Endure costs less) or **Invite** the catastrophe (guarantee it, for a deliberate Succumb). See [The Harbinger](harbinger.md).

When a catastrophe hits, nothing is destroyed yet. The game keeps running and the choice waits for you.

A window opens with the two choices, Endure and Succumb. Press E or S, or move between them with Tab or the arrow keys and press Enter. Endure is highlighted when the window opens, so a stray Enter chooses Endure. Esc closes the window without choosing, so you can look around first; the catastrophe stays pending.

While a catastrophe is pending, the status bar shows a **☄ CATASTROPHE PENDING** badge, and the game refuses to advance to the next age or prestige. Type `catastrophe` to reopen the choice. The pending catastrophe is saved with your game, so if you close the game or come back after a long idle, the window opens again when the save loads.

If a catastrophe strikes while the age-advance splash is on screen, the splash comes first. The catastrophe window opens once you dismiss the splash with any key, or when the splash times out after 20 seconds.

---

## When It Triggers

There are 7 epochs across 22 ages (Cosmic spans 4 ages, the rest 3). Every run starts in the Stone Era. A catastrophe never comes with a new epoch: an epoch transition only rolls a good or a challenging event (see [Epochs](epochs.md)). Instead, each era from the Iron Era on, the Cosmic Era included, can hold one doom:

1. **Fated or not?** When you enter the era, a hidden roll decides: a doom is fated there 27% of the time. Nothing is ever fated in the Stone Era. The fate is saved with your game, so reloading can't re-roll it.
2. **When?** A fated doom strikes at a random moment anywhere in the era, mid-age included. The moment is drawn across the era's expected length from when you entered it: its ages' target times added up, 27.3 hours for the Iron Era at 1x, 54.6 for the Steel Era, 80.6 for the Electric Era, 109.2 for the Digital Era, 156 for the Neon Era and 249.6 for the Cosmic Era. On known ground each age counts its target divided by its [Era Mastery](prestige.md#era-mastery) speed, so a mastered era's doom still falls inside the shorter era.
3. **A warning first.** A harbinger arrives before the strike, 20% to 60% of the current age's target time ahead of it (divided by the age's speed on known ground, so warnings in a mastered era are shorter in proportion). Until then, nothing in the game tells a fated era from a quiet one: a quiet era is safe, for now. See [The Harbinger](harbinger.md).
4. **Does it hit?** At the fated moment the strike rolls on your faith fill at that moment: **90% below 25% faith, 75% between 25% and 75% (or when faith has no storage yet), 60% above 75%**. Each level of Appease multiplies the chance by 0.6 (two levels at most), and Invite makes it certain. A miss means you were spared.

A few more rules apply:

- **One doom per era per run.** Once an era's doom has struck or passed you by, nothing more strikes in that era. Succumb and prestige start a new run, and each era rolls again. That makes 6 catastrophes the most a run can have, one for each era from Iron to Cosmic, plus the Last Passage if you prestige from the Cosmic Era.
- **You can't outrun it.** If you advance out of the era before the doom's moment, the strike rolls at that advance, before it goes through: a hit makes the advance wait behind the catastrophe, a miss lets it through. The same goes for leaving an age once the harbinger has said the doom falls before this age is out, and in the Cosmic Era, whose passage is prestige, for confirming prestige. See [No Outrunning a Doom](harbinger.md#no-outrunning-a-doom).
- **Offline too.** A doom strikes at its moment while you are away and waits, pending, for you to come back. The log shows the warning and the strike.
- **Prestige from the Digital or Neon Era ends it.** A doom that hasn't struck there when you prestige is gone with the run. In the Cosmic Era an open doom settles before the Last Passage rolls.
- **Never on top of another.** A new catastrophe never replaces one that is still pending.

The `catastrophe` command (with nothing after it) shows the outlook as you can know it when nothing is pending. With a harbinger present it repeats the warning, for example "The Oracle warns of doom before this age is out: medium risk of catastrophe (no figures this early), faith 40% full." (the odds as a figure from the Industrial Age on, a low / medium / high severity before it). With none, it reads "No harbinger has come: the Iron Era is quiet, for now. A doom is always foretold before it strikes." In the Stone Era it says no catastrophe can strike there, and once the era's doom has struck or passed you by, that nothing more will strike before the era ends. In the Cosmic Era it shows the risk of the Last Passage, along with how full your faith is, and above it the Reality Tear's warning while that doom's harbinger speaks, for example "Your future self warns of doom before this age is out: 90% catastrophe chance (high), faith 0% full." The Epoch panel shows the same outlook. Neither can give away a false prophet. See [Epochs](epochs.md) for the event tables.

---

## The Last Passage

The Cosmic Era has no next epoch, so its passage is prestige. When you confirm prestige in the Cosmic Era, an open Reality Tear settles first: if its harbinger hasn't come, it comes now and the prestige waits for one more try; then its strike rolls, and a hit holds the prestige until you answer it. After that, the **Last Passage** rolls once: 18% / 15% / 12% by faith, ×0.6 per level of Appease on its own harbinger thread, certain if invited. If it comes, the prestige waits for your choice:

- **Endure** completes the prestige but keeps only 50% of the run's prestige points (70% or 85% if you braced).
- **Succumb** completes the prestige with no points from this run and grants the **Cosmic Legacy**, a permanent +10% production. You can earn it once; after that, Succumb is closed.

It behaves like a pending catastrophe: Esc closes the choice, a **☄ LAST PASSAGE** badge shows in the status bar, the bare `catastrophe` command reopens it, and it is saved with your game. Unlike a catastrophe, it blocks only prestige. Either choice adds a line to the civilization log and counts as Endured or Succumbed. If the Reality Tear and the Last Passage are ever pending at once, the Reality Tear is answered first: the window shows it first, and the Last Passage's choice is refused until it is ("The Reality Tear came first. Answer it before the Last Passage."). See [The Last Passage](prestige.md#the-last-passage).

---

## The Catastrophes

Each epoch has a named catastrophe:

| Epoch | Catastrophe | The story |
|-------|-------------|-----------|
| Stone Era | The Great Meteor | Never strikes: no catastrophes before the Iron Era. |
| Iron Era | The Great Plague | A plague empties your cities. |
| Steel Era | The World War | Industrial war levels the factories. |
| Electric Era | The Nuclear Exchange | Nuclear war turns cities to glass. |
| Digital Era | The Great Hack | Every system goes dark and the AIs turn on their makers. |
| Neon Era | Corporate Armageddon | The megacorps end the world with a fusion bomb. |
| Cosmic Era | The Reality Tear | Exotic matter cracks spacetime open. |

The name and story are flavor. Endure and Succumb work the same way in every epoch; only the legacy bonus differs.

---

## Endure

Pay a cost and keep your civilization. With no Brace and no soldiers, enduring costs you:

| Loss | Amount |
|------|--------|
| Buildings destroyed | 20% of your buildings other than wonders and storage, rounded down (`floor(those buildings / 5)`), at least 1 if you have any. Wonders and storage are never destroyed and don't count. Brace lowers this to 15% or 10%, and your garrison lowers it further. |
| Stored resources | Every unlocked resource drops to 15% of its stored amount. Brace raises this to 30% or 45%, and your garrison raises it further. |
| Workers | 25% of the worker pool is lost. |
| Production | Reconstruction Effort: all production −10% for 562 ticks (about 18m 44s). |
| Morale | −10 points. |

For how Brace and soldiers change these numbers, see [Brace](#brace) and [Your garrison](#your-garrison) below.

Workers of the destroyed buildings go idle first, the same as when you sell a building. There is one worker pool, so the 25% loss takes the same share of assigned workers from every building, whatever its domain (food, knowledge, military and so on). Age, research, wonders and prestige are untouched. The next age's requirements are checked again from scratch, since most of the stock they counted is gone. You earn the **Endured** marker on the epoch badge and a line in the civilization log.

### Brace

If the harbinger warned you and you paid to **Brace**, Endure costs less. Brace changes only the share of buildings destroyed and the share of stored resources kept; worker loss, the production debuff and the morale hit stay the same.

| Brace level | Buildings destroyed | Stored resources kept |
|-------------|---------------------|-----------------------|
| none | 20% | 15% |
| 1 | 15% | 30% |
| 2 | 10% | 45% |

Buildings destroyed are rounded down, with at least 1 if you have any. The Brace is attached to the pending catastrophe, so it still applies if you press Esc and Endure later, or save and load first. It does nothing for Succumb. See [Brace](harbinger.md#brace-soften-an-endure).

### Your garrison

Soldiers soften an Endure too. After Brace has done its part, your garrison blunts its share of what is left: the same share it would blunt of a raid, measured against the raid threat of the age the catastrophe strikes in. A pending catastrophe keeps you in that age until you choose. See [Defense: what your army blunts](military.md#7-defense-what-your-army-blunts) for how the share is worked out.

- **Buildings:** the braced share of buildings destroyed shrinks by the garrison's share. The number of buildings the garrison saves is rounded down, so it never saves more than its share, and at least 1 building still falls if you have any.
- **Stock:** the garrison keeps its share of the stock that Brace would have let go.
- Worker loss, the production debuff and the morale hit are unchanged.
- Your soldiers are measured before the blow lands. They are stock like everything else, so afterwards they drop with the rest of your resources.

**The combined cap.** Brace and garrison together can cut the unbraced loss by at most **60%**. However strong your army, at least **8%** of buildings fall and at most **66%** of stock is kept. The cap only bites at Brace level 2: level 2 already takes the building loss from 20% to 10%, so the garrison can add at most a fifth on top (10% down to 8%).

| Brace | Garrison share | Buildings destroyed | Stock kept |
|-------|----------------|---------------------|------------|
| none | none | 20% | 15% |
| none | 20% | 16% | 32% |
| 1 | none | 15% | 30% |
| 1 | 20% | 12% | 44% |
| 2 | 20% | 8% | 56% |
| 2 | 40% or more | 8% (cap) | 66% (cap) |

A player with no soldiers takes exactly the Brace-only numbers above. Few players have none: the military buildings the age gates require train a small garrison on their own, enough to blunt roughly 8-19% of a raid. See [The garrison you already have](military.md#the-garrison-you-already-have).

**What you see.** The catastrophe window's Endure section shows the real numbers, after Brace and garrison. Under them it adds a line for each: the Brace level with its own numbers ("before your garrison"), then how many buildings the garrison saves and how much stock it keeps compared with Brace alone. With no soldiers it says so and points you to the Army panel. If the cap cut in, a gray line says Brace and garrison together soften an Endure by at most 60%. After you Endure, the log adds:

```
Your garrison held the line: N buildings still stand that would have fallen, and you keep X% of your stock instead of Y%.
```

The buildings and stock it saved go into the **Saved this run** line of the Army panel (`army`).

The Last Passage is different: its Endure costs prestige points, not buildings and stock, and soldiers do not change it.

The −10% applies to every building, including the ones that survived, for the full 562 ticks. Negative production modifiers are floored at 10% of base, but a single −10% lands in full whatever other bonuses you hold. The same flooring covers per-resource rate modifiers and worker output, and any active debuff shows in the Active Multipliers panel.

If morale was already low, the −10 can push it into the low band, where output is penalized. Morale drifts back toward 50% on its own; a food surplus speeds that up.

**When to Endure:** your civilization is big, and a full reset would cost more than the legacy bonus is worth. Also when you already hold this epoch's legacy bonus.

### How the destruction is picked

Every built building except wonders and storage goes into a pool in a fixed order (sorted by building key). The pool is shuffled with your run's seeded random generator and the first N are destroyed. Destroyed buildings are removed entirely; they don't become ruins. Because the generator is seeded per run, the same run state destroys the same buildings.

Storage is spared because it is what raises your caps, and once its age has passed it can never be rebuilt. Losing some could leave your caps too small to pay for this age's storage, and then they could never rise again.

**Recovery checklist:**

1. The log lists every lost building by name. Rebuild food and housing first.
2. Reassign or recruit workers for the rebuilt capacity.
3. Wait out the 562-tick reconstruction debuff; it can't be removed early.

---

## Succumb

Let the civilization fall, and keep something permanent.

Up to 8 of your buildings other than wonders and storage become **ruins**, picked at random from the same seeded, fixed-order pool as Endure. Ruins produce at 50% of base rate with no workers. They carry across Succumb and prestige, but the total is capped at 24. When new ruins push past the cap, the lowest-value ruins crumble first (earliest age first, then lowest base output), so a late-game fall replaces primitive rubble. A save loaded with more than 24 ruins is trimmed the same way.

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

From Iron to Cosmic there are 6 eras you can succumb in, so the most you can earn is **+150%**. The Stone Era legacy can't be earned, since no catastrophe strikes there, but a save that already holds it keeps it (+175% total). Succumbing to the Last Passage grants the Cosmic Legacy instead, not an era legacy.

**When to Succumb:** you don't yet hold this epoch's legacy, and the reset is cheap for you. A catastrophe can strike anywhere in its era, early or late, so the question is how much of the run you'd be giving up at that moment.

---

## Legacy Bonus Table

| Epoch | Resources boosted | Bonus |
|-------|-------------------|-------|
| Stone Era | wood, stone | +20% each (can't be earned; kept by saves that hold it) |
| Iron Era | iron | +20% |
| Steel Era | steel, coal | +25% each |
| Electric Era | electricity, uranium | +25% each |
| Digital Era | data, titanium ore | +30% each |
| Neon Era | plasma, dark matter crystals | +30% each |
| Cosmic Era | dark matter | +35% |

Rate bonuses add to your other rate bonuses for that resource and apply from tick 1 of every later run, including after prestige. A new game (wiping the save) clears them.

---

## Faith and the Odds

| Faith fill | Good epoch event | Fated doom strikes | Last Passage |
|------------|------------------|--------------------|--------------|
| under 25% | 40% | 90% | 18% |
| 25 to 75% (or no faith storage) | 50% | 75% | 15% |
| over 75% | 60% | 60% | 12% |

A doom rolls on your faith fill at the moment it strikes, so the faith that counts is the faith you hold then, not when the harbinger came. A first run to a Modern Age prestige lives through three eras that can hold a doom (the Iron, Steel and Electric Eras; the Digital Era's doom rarely strikes before you prestige there), so it can expect about 0.73 catastrophes at low faith, 0.61 at mid faith and 0.49 at high faith. High faith saves you about a quarter of a catastrophe, and gives you more good events. A deep run to a Quantum Age prestige lives through all six such eras, about 1.46 catastrophes at low faith, plus the Last Passage. See [Faith](faith.md).

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
- you braced when the harbinger warned you, so Endure costs less,
- you keep a garrison that is strong for the age the catastrophe strikes in (check the Army panel's blunt share, or the Brace preview on the harbinger panel).

### Succumb beats Endure when

- the reset is cheap (early epochs, or a run you were going to prestige soon anyway),
- you don't hold this epoch's legacy yet,
- you have plenty of runs ahead to cash in the research bonus.

### Pending is a pause, not a dodge

Esc lets you check your buildings and resources before you commit, and the game keeps producing meanwhile. But you can't advance or prestige until you choose.

### Morale

Enduring costs 10 points of morale. Keep food positive, avoid over-militarizing (military above 30% of population drains morale), and build worship and culture buildings to climb back. See [Morale](morale.md).

### The long game

Succumbing once in each era from Iron to Cosmic, over several runs, collects all six reachable legacy bonuses and +150% research speed. You can't choose where a doom is fated, but when a harbinger comes, **Invite** guarantees its strike, so you can take the fall in that era. Each run after that starts with those bonuses and up to 24 ruins producing from tick 1.
