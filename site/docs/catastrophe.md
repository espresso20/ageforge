# Catastrophes

A catastrophe is a doom that threatens your civilization and forces a permanent choice: **Endure** it or **Succumb** to it. From the Iron Era on, a doom may be fated in secret to strike at some moment inside an era, and a harbinger always comes to warn you first. You can't trigger a catastrophe directly. The only way to choose one is to **Invite** it when a harbinger comes.

> **The Harbinger.** A doom never strikes unannounced. Some while before it does, a harbinger arrives to warn you, and you can answer it: lower the odds, soften an Endure, or invite the fall. [The Harbinger](harbinger.md) covers the warning, the answers and what they cost. This page covers what fate decides and what a catastrophe does.

When a catastrophe strikes, nothing is destroyed yet. The game keeps running and the choice waits for you.

A window opens with two buttons, Endure and Succumb. Press `E` or `S`, or move between the buttons with `Tab` or the arrow keys and press `Enter`. Endure is selected when the window opens, so a stray `Enter` chooses Endure. `Esc` closes the window without choosing, so you can look around first. The catastrophe stays pending.

<figure class="screen" data-screen="catastrophe"><figcaption>The choice a catastrophe brings, here the Great Plague in the Iron Era: what Endure would cost with this garrison, and what Succumb would keep.</figcaption></figure>

While a catastrophe is pending, the status ribbon shows a **☄ Catastrophe** badge (on a wide terminal it reads "☄ Catastrophe pending. Type catastrophe to choose."). The game refuses to advance to the next age or to prestige until you choose. Type `catastrophe` to reopen the choice. The pending catastrophe is saved with your game, so if you close the game and come back, the window opens again when the save loads.

If a catastrophe strikes while the arrival screen of an age advance is up, the arrival comes first. The catastrophe window opens once you close that screen (any key moves on to what the age opens, and a second key closes it), or when it closes by itself after 20 seconds.

---

## When it triggers

There are 7 epochs across 22 ages. The Cosmic Era has 4 ages and each of the others has 3. Every run starts in the Stone Era. A catastrophe never comes with a new epoch: an epoch transition only rolls a good or a challenging event (see [Epochs](epochs.md)). Instead, each era from the Iron Era on, the Cosmic Era included, can hold one doom.

1. When you enter the era, a hidden roll decides whether a doom is fated there. It is fated 27% of the time. Nothing is ever fated in the Stone Era. The fate is saved with your game, so reloading can't re-roll it.
2. A fated doom strikes at a random moment anywhere in the era, mid-age included. The moment is drawn evenly across the era's expected length, counted from when you entered it: the target times of its ages added up. On known ground each age counts its target divided by its [Era Mastery](prestige.md#era-mastery) speed, so a mastered era's doom still falls inside the shorter era. The moment is counted in ticks, so anything that makes ticks come faster brings it sooner in real time.
3. A harbinger arrives before the strike, between 20% and 60% of an age ahead of it. That share is measured on the same length the Appease and Brace prices are set from (the age's target time, with the price climb from the Information Age on), divided by the age's speed on known ground, so warnings in a mastered era are shorter in proportion. Until it arrives, nothing in the game tells a fated era from a quiet one: a quiet era is safe, for now. Before the Industrial Age a quiet era can also send a false prophet, a warning with no doom behind it. See [The Harbinger](harbinger.md).
4. At the fated moment the strike rolls on your faith strength at that moment: 90%, 75% or 60% from low to high faith (see [Faith threshold bands](faith.md#faith-threshold-bands)). A typical town rolls at 90%, and it takes more faith buildings than a moderate set, kept staffed, to lower it. Each level of Appease multiplies the chance by 0.6 (two levels at most), and Invite makes it certain. A miss means you were spared.

More rules apply:

- Each era holds at most one doom per run. Once an era's doom has struck or passed you by, nothing more strikes in that era. Succumb and prestige start a new run, and each era rolls again. A run can meet at most 6 catastrophes, one for each era from Iron to Cosmic, plus the [Last Passage](prestige.md#the-last-passage) if you prestige from the Cosmic Era.
- You can't outrun a doom. If you advance out of the era before the doom's moment, the strike rolls at that advance, before it goes through: a hit makes the advance wait behind the catastrophe, and a miss lets it through. The same goes for leaving an age once the harbinger has said the doom falls before this age is out, and, in the Cosmic Era, whose passage is prestige, for confirming prestige. If no harbinger has come by then, one comes at that advance and the advance waits so you can answer it. Advance again and the strike rolls. See [No outrunning a doom](harbinger.md#no-outrunning-a-doom).
- A doom strikes at its moment while you are away and waits, pending, until you come back. The log shows the warning and the strike.
- Prestige before the Cosmic Era leaves an open doom behind with the run. Prestige opens at the Medieval Age, so an early taste there escapes an Iron Era doom, but it costs you the run. In the Cosmic Era an open doom settles before the Last Passage rolls.
- A new catastrophe never replaces one that is still pending.

The `catastrophe` command (alias `cat`) reopens a pending choice. With nothing pending it shows the outlook as you can know it:

- With a harbinger here, it repeats the warning, for example "The Oracle warns of doom before this age is out: medium risk of catastrophe (no figures this early), faith strength 44% (devotion 2.0x, 100% of your faith kept)." From the Industrial Age on the odds show as a figure. Before it you get a low, medium or high severity.
- With none, it reads "No harbinger has come: the Iron Era is quiet, for now. A doom is always foretold before it strikes."
- In the Stone Era it says no catastrophe can strike there. Once the era's doom has struck or passed you by, it says nothing more will strike before the era ends.
- In the Cosmic Era it shows the risk of the Last Passage with your faith strength. While the Reality Tear's harbinger speaks, its warning shows above that, for example "Your future self warns of doom before this age is out: 90% catastrophe chance (high), faith strength 18% (devotion 0.8x, 100% of your faith kept)."

The Epoch panel shows the same outlook. Neither tells a false prophet from a true one. See [Epochs](epochs.md) for the event tables.

---

## The Last Passage

The Cosmic Era has no next epoch, so its passage is prestige. A prestige from the Cosmic Era can bring the **Last Passage**: a choice between keeping part of the run's prestige points (Endure) and the permanent Cosmic Legacy (Succumb). It blocks only prestige. Like a catastrophe it waits behind a badge on the status ribbon until you answer it with `catastrophe`. See [The Last Passage](prestige.md#the-last-passage).

The roll happens when you confirm the prestige. It reads the same faith bands as a doom, at 18%, 15% or 12% from low to high faith. A harbinger warns of it from the moment you enter the era, and its Appease lowers the odds like any other (×0.6 per level, two levels at most). Invite makes it certain. Brace does not touch the odds. It raises the share of the run's prestige points that an Endure keeps, from 50% to 70% at level 1 and 85% at level 2.

A Succumb at the Last Passage gives no points from the run and grants the Cosmic Legacy, +10% to all production, counted in full after every other bonus. You earn it once. Once you carry it, only Endure is left at the Last Passage.

What the answers cost depends on the age the harbinger arrives in, so the prices are in one table: [What it costs, by epoch](harbinger.md#what-it-costs-by-epoch). [Brace or Appease against the Last Passage](harbinger.md#brace-or-appease-against-the-last-passage) compares the two.

---

## The catastrophes

Each era has a named catastrophe.

| Era | Catastrophe | The story |
|-----|-------------|-----------|
| Stone Era | The Great Meteor | Never strikes: no catastrophe comes before the Iron Era. |
| Iron Era | The Great Plague | A plague sweeps your cities and the streets fall silent. |
| Steel Era | The World War | Industrial warfare tears civilization apart. |
| Electric Era | The Nuclear Exchange | Nations turn the atom on each other and cities become glass. |
| Digital Era | The Great Hack | Every system falls silent and the AIs turn on their creators. |
| Neon Era | Corporate Armageddon | The megacorps end the world with a fusion bomb. |
| Cosmic Era | The Reality Tear | Exotic matter destabilizes spacetime and reality cracks open. |

The name and story are flavor. Endure and Succumb work the same way in every era, and only the legacy bonus differs. The Reality Tear is the Cosmic Era's fated doom like any other era's. The Last Passage at a Cosmic Era prestige is a separate roll (see [The Last Passage](#the-last-passage)).

---

## Appease, Brace and Invite

While a harbinger is here you can answer it from the Harbinger panel (`harbinger`) or with `harbinger appease`, `harbinger brace` and `harbinger invite`. The answers belong to the doom, not to the figure, so the levels you buy carry over when the next figure takes up the warning.

- Appease lowers the chance the doom strikes. Each of its two levels multiplies the chance by 0.6. It costs faith and, from the Steel Era on, culture. It is refused after Invite.
- Brace softens an Endure if the doom strikes. Each of its two levels costs materials the age makes, from a written list for each era. The levels are shown under [Brace](#brace). It is still allowed after Invite.
- Invite is free and can't be undone. The doom then strikes for certain when its moment comes, even if a false prophet foretold it.

Nothing is refunded. Appease and Brace are spent whether or not the doom comes. In the Stone Era all three are refused, because nothing can strike there. The prices depend on the era and on the age the harbinger arrives in, so they are in one table: [What it costs, by epoch](harbinger.md#what-it-costs-by-epoch).

---

## Endure

Pay a cost and keep your civilization. With no Brace and no garrison, enduring costs you:

| Loss | Amount |
|------|--------|
| Buildings destroyed | 20% of your buildings other than wonders, storage and knowledge buildings, rounded down (`floor(those buildings / 5)`), and at least 1 if you have any. Wonders, storage and knowledge buildings are never destroyed and don't count. Brace lowers this to 15% or 10%, and your garrison lowers it further. |
| Stock | Every unlocked resource drops to 15% of what you hold. Brace raises this to 30% or 45%, and your garrison raises it further. |
| Workers | 25% of your workers are lost, rounded down. |
| Production | Reconstruction Effort: all production -10% for 562 ticks (about 18m 44s at the base tick). |
| Morale | -10 points. |

For how Brace and soldiers change these numbers, see [Brace](#brace) and [Your garrison](#your-garrison) below.

Workers of the destroyed buildings go idle first, the same as when you sell a building. There is one worker pool, so the 25% loss takes the same share of assigned workers from every building, whatever its domain (food, faith, military and so on). Age, research, wonders and prestige are untouched. The next age's requirements are checked again from scratch, since most of the stock they counted is gone. The Epoch panel marks the era as Endured, the civilization log gets a line, and the Endure counts toward your Dooms endured badges (see [Badges](badges.md)).

### Brace

If the harbinger warned you and you paid to Brace, Endure costs less. Brace changes only the share of buildings destroyed and the share of stock kept. The worker loss, the production debuff and the morale hit stay the same.

| Brace level | Buildings destroyed | Stock kept |
|-------------|---------------------|------------|
| none | 20% | 15% |
| 1 | 15% | 30% |
| 2 | 10% | 45% |

Buildings destroyed are rounded down, with at least 1 if you have any. The Brace is attached to the pending catastrophe, so it still applies if you press `Esc` and Endure later, or save and load first. It does nothing for Succumb. See [What it costs, by epoch](harbinger.md#what-it-costs-by-epoch) for the prices.

### Your garrison

Soldiers soften an Endure too. After Brace has done its part, your garrison blunts its share of what is left: the same share it would blunt of a raid, measured against the raid threat of the age the catastrophe strikes in. A pending catastrophe keeps you in that age until you choose. See [Defense: what your army blunts](military.md#defense-what-your-army-blunts) for how the share is worked out.

- The braced share of buildings destroyed shrinks by the garrison's share. The number of buildings the garrison saves is rounded down, so it never saves more than its share, and at least 1 building still falls if you have any.
- The garrison keeps its share of the stock that Brace would have let go.
- The worker loss, the production debuff and the morale hit are unchanged.
- Your garrison is measured before the blow lands. Military buildings can be among the ones destroyed, and a destroyed copy takes its soldiers with it.

Brace and garrison together can cut the unbraced loss by at most 60%. However strong your army, at least 8% of buildings fall and at most 66% of stock is kept. The cap only bites at Brace level 2. Level 2 already takes the building loss from 20% to 10%, so the garrison can add at most a fifth on top (10% down to 8%).

| Brace | Garrison share | Buildings destroyed | Stock kept |
|-------|----------------|---------------------|------------|
| none | none | 20% | 15% |
| none | 20% | 16% | 32% |
| 1 | none | 15% | 30% |
| 1 | 20% | 12% | 44% |
| 2 | 20% | 8% | 56% |
| 2 | 40% or more | 8% (cap) | 66% (cap) |

A player with no soldiers takes exactly the Brace-only numbers above. Few players have none: the military buildings the age gates require come with a small garrison. See [The garrison you already have](military.md#the-garrison-you-already-have).

The catastrophe window shows the real numbers, after Brace and garrison. Under them it adds a line for each: the Brace level with its own numbers ("before your garrison"), then how many buildings the garrison saves and how much stock it keeps compared with Brace alone. With no soldiers it says so and points you to the Army panel. If the cap cut in, a gray line says Brace and garrison together soften an Endure by at most 60%. After you Endure, the log adds:

```
Your garrison held the line: N buildings still stand that would have fallen, and you keep X% of your stock instead of Y%.
```

The buildings and stock it saved go into the **Saved this run** line of the Army panel (`army`).

The Last Passage is different. Its Endure costs prestige points, not buildings and stock, and soldiers do not change it.

The -10% from Reconstruction Effort applies to every building, including the ones that survived, for the full 562 ticks. It goes into the all-production pool, whose multiplier never drops below x0.1 (see [The all-production pool](resources.md#the-all-production-pool)). The debuff lands in full whatever the pool holds: it costs 10 points. The pool of each resource's own output and the worker output pool have the same floor. Any active debuff shows in the Active multipliers section of the Stats panel.

Morale below 50% lowers production a little more the lower it goes, and it drifts back toward 50% on its own. See [Morale](morale.md).

### How the destruction is picked

Every built building except wonders, storage and knowledge buildings goes into a pool in a fixed order (sorted by building key). The pool is shuffled with your run's seeded random generator and the first N are destroyed. Destroyed buildings are removed entirely. They don't become ruins. Because the generator is seeded per run, the same run state destroys the same buildings.

Storage is never destroyed by an Endure (nor by The Great Fire epoch event). Every copy adds to all production for the rest of the run, and once its age has passed it can never be rebuilt. Knowledge buildings are spared for a similar reason: they come with your scholars, and an older age's copy can never be rebuilt.

After an Endure:

1. The log lists every lost building by name. Rebuild food and housing first.
2. Let the workers come back. With auto-recruit on (the default), the game recruits into empty worker slots as housing and food allow. See [The roster](workers-and-domains.md#the-roster).
3. Wait out the 562-tick reconstruction debuff. It can't be removed early.

---

## Succumb

Let the civilization fall, and keep something permanent.

A Succumb ends the run, and the ending plays on the whole screen as a [prestige's](prestige.md) does, in a lower voice: the town goes cold and dark, what the fall leaves you is read out, and the land comes up out of the dark. Any key skips to the next part.

Up to 8 of your buildings other than wonders, storage and knowledge buildings become **ruins**, picked at random from the same seeded, fixed-order pool as Endure. Ruins produce at 50% of base rate with no workers. They carry across Succumb and prestige, up to 24 in all. When new ruins push past the cap, the lowest-value ruins crumble first (earliest age first, then lowest base output), so a late-game fall replaces primitive rubble. A save loaded with more than 24 ruins is trimmed the same way.

You also get the era's **legacy bonus** (table below), permanently, and **Ancient Knowledge**: research takes 0.8 times as long for each distinct epoch you have succumbed in. Succumbing twice in the same era doesn't add another step.

The run also ends as a prestige run does for [Era Mastery](prestige.md#era-mastery): every age it completed gains a mastery level, so the rebuild runs on known ground. The catastrophe window lists it when there is a level to gain.

Then the civilization resets to the Primitive Age: buildings, resources, workers, research, milestones, events, trade routes, civilizations met, the build queue and the build plan. You start with 15 food and 12 wood. No prestige points are earned, but your prestige level, points, upgrades and [Era Mastery](prestige.md#era-mastery) are kept. After the first prestige, Plan Template puts the first age's part of your plan back and Roster sets your shares. The ages the fallen run completed have gained a mastery level, so they run at least 2x on the rebuild, and every age 6 or more behind your record runs at least 4x ([catch-up](prestige.md#catch-up)). The festival and black market cooldowns start over with the run. Morale restarts at 50%, and the civilization log gets a line.

### What carries forward

| Item | After Succumb |
|------|---------------|
| Prestige level, points and upgrades | Kept |
| What Plan Template, Roster and Old Friends remember | Kept, plus the fallen run's plan (the techs you planned included), roster and civilizations met |
| Era Mastery and your record (the deepest age you have ever entered) | Kept, and every age the fallen run completed gains a mastery level |
| Ruins | Kept, plus up to 8 more, capped at 24 |
| Legacy bonuses and Ancient Knowledge | Kept, plus this era |
| Civilization log | Kept |
| Epoch panel history | Kept |
| Resources, buildings, workers | Reset |
| Research | Reset |
| Milestones and chains | Reset |
| Trade routes and civilizations met | Reset |

### Ancient Knowledge

The research bonus comes straight from your legacy flags: research time ×0.8 for each era flagged. It is worked out whenever it's needed, so saving, loading, Succumb and prestige can't drop or double it.

It multiplies. The research speed pool is added up and taken off a tech's listed time. The techs' own cut and Ancient Knowledge then multiply what is left, and [Era Mastery](prestige.md#era-mastery) divides that. Every era you succumb in counts:

| Eras succumbed in | 1 | 2 | 3 | 4 | 5 | 6 |
|---|---|---|---|---|---|---|
| Research time | ×0.8 | ×0.64 | ×0.51 | ×0.41 | ×0.33 | ×0.26 |

It is not part of the research speed pool, so it has a line of its own: **Ancient Knowledge** in the Stats panel, under Legacy bonuses and again under Research bonuses, where the listed tech times already include it.

From Iron to Cosmic there are 6 eras you can succumb in, so the most you can earn is ×0.26. The Stone Era legacy can't be earned, since no catastrophe strikes there, but a save that already holds it keeps it (a seventh step, ×0.21). A Succumb at the Last Passage grants the Cosmic Legacy instead, not an era legacy.

---

## Legacy bonus table

| Era | Resources boosted | Bonus |
|-----|-------------------|-------|
| Stone Era | wood, stone | +20% each (can't be earned; kept by saves that hold it) |
| Iron Era | iron | +20% |
| Steel Era | steel, coal | +25% each |
| Electric Era | electricity, uranium | +25% each |
| Digital Era | data, titanium ore | +30% each |
| Neon Era | plasma, dark matter crystals | +30% each |
| Cosmic Era | dark matter | +35% |

Rate bonuses add to your other rate bonuses for that resource and apply from tick 1 of every later run, including after prestige. A new game (wiping the save) clears them.

---

## Faith and the odds

Your faith strength when a fated doom's moment comes sets the chance it strikes: 90%, 75% or 60% from low to high faith. A Cosmic Era prestige rolls the Last Passage on it the same way, at 18%, 15% or 12%. The full table, and what high faith saves you over a run, are on the [Faith](faith.md#faith-threshold-bands) page. A harbinger's Appease multiplies these chances by 0.6 per level, and Invite makes the catastrophe certain.

---

## Civilization log

Every resolved catastrophe adds a line to the civilization log with the tick, the outcome (Endured or Succumbed), the catastrophe and its era, and then either the number of buildings lost (Endure) or the words "Civilization reset" with whether the legacy bonus was earned or already held (Succumb). An Endure or a Succumb at the Last Passage adds a line of its own. The log survives Succumb and prestige, and the Epoch panel shows it. The Stats panel counts **Endured** and **Succumbed** from these lines, so a pending catastrophe counts as neither.

The Epoch panel's history marks each past era's catastrophe as **Endured**, **Succumbed** or **Pending**.

---

## Strategy

Endure tends to be the better choice when:

- your civilization is large and deep into its run,
- you already hold this era's legacy, since a repeat Succumb there adds no legacy bonus and no step of Ancient Knowledge,
- you are close to a milestone chain that a reset would wipe,
- you braced when the harbinger warned you, so Endure costs less,
- your garrison is strong for the age the catastrophe strikes in (the Army panel shows its blunt share, and the Harbinger panel's Brace preview counts it).

Succumb tends to be the better choice when:

- the reset is cheap, in the early eras where the run you give up is still short. A Succumb pays no prestige points, so think twice once the run is deep enough for a prestige to pay well (see [Prestige](prestige.md)),
- you don't hold this era's legacy yet,
- you have plenty of runs ahead to cash in the research bonus.

A pending catastrophe is a pause, not a dodge. `Esc` lets you check your buildings and resources before you commit, and the game keeps producing meanwhile. You can't advance or prestige until you choose.

An Endure costs 10 points of morale. See [Morale](morale.md) for what lifts it again.

Succumbing once in each era from Iron to Cosmic, over several runs, collects all six reachable legacy bonuses and Ancient Knowledge at ×0.26 research time. You can't choose where a doom is fated, but when a harbinger comes, Invite guarantees its strike, so you can take the fall in that era. Each run after that starts with those bonuses and up to 24 ruins producing from tick 1.
