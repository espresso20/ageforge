# Prestige System

Prestige is the endgame reset loop. When you reach the **Modern Age** (Age 12), you can sacrifice your entire civilization to earn **Prestige Points** and purchase permanent upgrades that carry into every future run.

```
prestige confirm yes
```

> Prestige resets your age, resources, buildings, workers, and research. Prestige upgrades, legacy bonuses from Succumb, the Cosmic Legacy and ruins are **permanent**.

---

## When You Can Prestige

You can prestige from the **Modern Age (Age 12)** or any later age. There is no upper limit: if you push on to the Quantum Age before prestiging, you earn more points.

At 1x speed a run is paced to reach the Modern Age in about **3 days of game time** (the smoke-test bot gets there in about 2.4 days). The ages before it range from 15 minutes (Primitive) to 12 hours (Atomic); the Modern Age and the ages after it take 12 to 24 hours each. The game grants up to 24 hours of offline progress, so time away counts.

Prestige is refused while a [catastrophe](catastrophe.md) is pending. Type `catastrophe` and choose Endure or Succumb first. From the Digital or Neon Era, a doom fated for your era that hasn't struck yet ends with the run when you prestige.

In the Cosmic Era, confirming prestige first settles the era's own doom if one is still open, then can bring the [Last Passage](#the-last-passage). If it comes, the prestige waits until you choose Endure or Succumb.

To check your current prestige status:

```
prestige
```

This shows your current level, available points, points you would earn right now, and whether you have reached the age prestige needs. To view the upgrade shop without committing:

```
prestige shop
```

When you're ready:

```
prestige confirm yes
```

The double confirmation (`confirm yes`) is deliberate, because prestige can't be undone.

Every prestige, from any age, ends with one closing line in the log, written for the age the run ended in. From the Modern Age to the Space Age the civilization simply winds down: offices empty, the last tram runs, the orbital yards shut. In the Cosmic Era the age's harbinger is there at the end, and the lines turn to cosmic dread ("Your unmade self took your hand.").

---

## Prestige Points Formula

Points earned per prestige run are calculated as:

```
base      = age_index  (0 = Primitive, 1 = Stone, ..., 12 = Modern, 20 = Quantum, 21 = Transcendent)
bonus     = floor(milestones / 10) + floor(techs / 15) + floor(total_built / 50)
raw       = base + bonus
points    = floor(raw / sqrt(prestige_level + 1))
```

The `sqrt(level + 1)` divisor gives **diminishing returns**: each run pays fewer points for the same achievements as your prestige level grows. Every prestige pays at least 1 point.

### What contributes to points

| Source | Points per unit |
|--------|----------------|
| Age index (each age beyond Primitive) | 1 pt each |
| Every 10 milestones completed | +1 pt |
| Every 15 techs researched | +1 pt |
| Every 50 buildings constructed (lifetime) | +1 pt |

Reaching the Modern Age for the first time usually pays **4 to 8 points**, depending on how you played. Pushing to a late age (Quantum is index 20) before prestiging gives 20 or more before the divisor.

---

## The Last Passage

A [harbinger](harbinger.md) comes to warn of a doom fated inside your era, and the Cosmic Era can hold one too: the Reality Tear. But the Cosmic Era has no next epoch, so its passage is prestige itself: the **Last Passage**. A second harbinger thread warns of it from the moment you enter the era, with a new figure each age: the Distress Beacon, the Elder Relay, your future self, then your unmade self. While the Reality Tear's harbinger speaks, the Last Passage's thread waits behind it with its answers intact, and takes up the warning again once the doom has struck or passed you by.

Prestige from before the Cosmic Era never rolls for it.

### The roll

When you type `prestige confirm yes` in the Cosmic Era, an open Reality Tear settles first (see [The Reality Tear comes first](#the-reality-tear-comes-first)). Then the Last Passage rolls once, with odds set by your faith fill (see [Faith and the Odds](catastrophe.md#faith-and-the-odds)):

| Faith fill | Chance of the Last Passage |
|------------|----------------------------|
| under 25% | 18% |
| 25 to 75% | 15% |
| over 75% | 12% |

Each level of Appease on the Last Passage's thread multiplies the chance by 0.6 (two levels at most). Invite makes it certain.

`prestige` shows the current chance and which figure is warning of it, for example `☄ The Last Passage: 18% chance (high) when you prestige.` `prestige confirm` spells out what Endure and Succumb would give you before you commit.

- **Nothing comes.** The verdict is Spared, and prestige completes as normal.
- **It comes.** Prestige does **not** complete yet. A choice opens, titled **✦ The Last Passage**, in the same style as the catastrophe choice.

### The Reality Tear comes first

Prestige is the Cosmic Era's passage, so its fated doom can't be outrun past it. If that doom is still open when you confirm prestige:

- **No harbinger yet.** It comes at the prestige, and the prestige waits for one more try: "... Type 'harbinger' to answer, or confirm prestige again to meet it."
- **Then the strike rolls,** before the Last Passage. A hit holds the prestige behind the pending Reality Tear: Endure it, then prestige again, which rolls the Last Passage (or Succumb, which resets the run with no prestige). A miss (spared) lets the same confirm go on to the Last Passage roll.
- **Both pending at once.** If the two are ever pending together, the Reality Tear is answered first: the choice window shows it first, and the Last Passage's Endure and Succumb are refused until it is ("The Reality Tear came first. Answer it before the Last Passage.").

### While it is pending

- **Esc** closes the choice. The status bar shows a **☄ LAST PASSAGE** warning telling you to type `catastrophe`, and a bare `catastrophe` reopens the choice.
- Only prestige is blocked. You can still advance ages, build and play on.
- It is saved with your game, and the Load Game browser lists it as the pending choice.

### Endure or Succumb

Both finish the prestige and raise your prestige level. Both add a line to the civilization log and count toward the Stats panel's tally of endured and succumbed catastrophes.

**Endure** keeps part of this run's prestige points. Brace raises the share:

| Brace level | Points kept |
|-------------|-------------|
| none | 50% |
| 1 | 70% |
| 2 | 85% |

The result is rounded down, so a small run can keep 0 points. Here Brace changes only the points share: buildings and resources reset anyway. The log records a Vindicated verdict, or Fulfilled if you invited it.

**Succumb** earns no points from this run and grants the [Cosmic Legacy](#cosmic-legacy). If you already carry it, Succumb is closed ("You already carry the Cosmic Legacy. Succumb is closed to you.") and Endure is the only choice.

### Choosing it on purpose

Inviting the Last Passage's thread is how you take the Cosmic Legacy on purpose. Invite is free, can't be undone and closes Appease; your next prestige brings the Last Passage. Brace levels still raise the Endure share, in case you change your mind. See [Invite](harbinger.md#invite-choose-the-catastrophe).

---

## Cosmic Legacy

A one-time, permanent reward for Succumbing to the Last Passage.

- **+10% production** (all resources), active from tick 1 of every run.
- Shows as **Cosmic Legacy** in the Stats panel, under Active Multipliers and in the Legacy Bonuses list. `prestige` shows `Cosmic Legacy: +10% production (permanent)`.
- Survives every prestige and every Succumb. Only wiping the game clears it.
- You earn it once. While you hold it, Succumb is closed at the Last Passage.

---

## Prestige Upgrades

9 upgrades, each with 5 tiers. Costs are in Prestige Points. All upgrades persist across every reset, including prestige and Succumb.

| Upgrade | Key | Effect per Tier | Max Tier | Cost (T1 → T5) |
|---------|-----|-----------------|----------|----------------|
| Gather Boost | `gather_boost` | +5% worker output | 5 | 2 / 3 / 4 / 6 / 8 |
| Storage Bonus | `storage_bonus` | +20 storage for every resource | 5 | 2 / 3 / 4 / 6 / 8 |
| Knowledge Production | `research_speed` | +5% knowledge production | 5 | 2 / 3 / 5 / 8 / 10 |
| Military Power | `military_power` | +5% military power | 5 | 2 / 3 / 5 / 8 / 10 |
| Starting Food | `starting_food` | +25 starting food | 5 | 1 / 2 / 3 / 4 / 5 |
| Starting Wood | `starting_wood` | +25 starting wood | 5 | 1 / 2 / 3 / 4 / 5 |
| Housing Bonus | `population_cap` | +2 housing | 5 | 2 / 3 / 5 / 8 / 10 |
| Expedition Loot | `expedition_loot` | +5% expedition rewards | 5 | 2 / 3 / 5 / 8 / 10 |
| Temporal Mastery | `tick_speed` | +5% tick speed | 5 | 6 / 10 / 17 / 23 / 33 |

```
prestige shop                # view available upgrades and costs
prestige buy gather_boost    # buy the next tier of Gather Boost
prestige buy tick_speed      # buy the next tier of Temporal Mastery
prestige buy starting_food   # buy the next tier of Starting Food
```

You can buy prestige upgrades **before you prestige again**. Points left over from earlier runs can be spent as soon as you log in, so there is no reason to wait.

### Effect Types

| Kind | Upgrades | What each tier does |
|------|----------|---------------------|
| Percentage | Gather Boost, Knowledge Production, Military Power, Expedition Loot, Temporal Mastery | Adds a percentage to that rate. Gather Boost at tier 3 is +15% worker output. |
| Flat | Storage Bonus, Housing Bonus | Adds a flat amount. Storage Bonus at tier 5 is +100 storage for every resource. |
| Starting resource | Starting Food, Starting Wood | Adds to what you start each run with. At tier 5 you begin with +125 food or wood. |

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
- The run's timers: the ready-to-advance notice, a famine in progress and the Geographic Society's survey countdown

### Persists Across Prestige
- Prestige level and all purchased upgrade tiers
- Ruins (from past Succumb events), which carry into the new run
- Legacy bonuses (from Succumb events), active from tick 1
- The Cosmic Legacy, if you have earned it
- Ancient Knowledge bonus (+25% research speed per distinct epoch succumbed), which comes from your legacy bonuses, so prestige never drops it
- Civilization history and catastrophe log

### Morale on Prestige

On prestige, morale is set to **70%**. A brand-new game starts at 50%. Morale works on a continuous curve around 50%: above it production gets a bonus that grows the higher morale goes, and below it a penalty that grows the lower it goes. So a prestige run opens with a small production bonus (about +8%, since no wonders are built yet), and morale then drifts slowly back toward 50% on its own. To keep a bonus, or reach the full +20%, you raise morale the usual ways: worship and culture buildings, good events and age advances.

See [Morale](morale.md) for the full curve.

### Culture on Prestige

Culture resets with every other resource. A new run starts with none.

---

## Legacy Bonuses

Legacy bonuses are earned by choosing **Succumb** during a catastrophe event. They are separate from prestige upgrades but interact with them on every subsequent run.

Each Succumb grants:
- **Ancient Knowledge**: a permanent +25% research speed per distinct epoch succumbed (a second Succumb in the same epoch adds nothing; +150% at most, Iron to Cosmic)
- **Epoch Legacy Bonus**: a permanent production multiplier for the main resources of that epoch

| Epoch | Legacy Production Bonus |
|-------|------------------------|
| Stone Era | wood +20%, stone +20% |
| Iron Era | iron +20% |
| Steel Era | steel +25%, coal +25% |
| Electric Era | electricity +25%, uranium +25% |
| Digital Era | data +30%, titanium ore +30% |
| Neon Era | plasma +30%, dark matter crystals +30% |
| Cosmic Era | dark matter +35% |

These bonuses apply from **tick 1** of every new run, including after prestige. A player who has Succumbed in the Iron Era and Steel Era starts every run with iron, steel and coal production already multiplied, and +50% research speed.

Succumbs in different epochs stack; a second Succumb in an epoch you already hold adds nothing. Catastrophes strike from the Iron Era on, so six legacy bonuses are reachable (the Stone Era one only exists on saves that already hold it). The Cosmic Era's comes from Succumbing to the Reality Tear; Succumbing to the Last Passage grants the Cosmic Legacy instead.

Legacy bonuses survive prestige the same way ruins do.

---

## Ruins at Prestige

When you prestige, any ruins you've accumulated from Succumb events carry forward. Ruins produce at 50% of the base rate and need no workers, so they give free production from tick 1.

On a fresh prestige run with accumulated ruins, your food, wood, or other resources may already be ticking up before you've built a single building. Each Succumb adds up to 8 ruins, and the collection is capped at 24; past the cap the lowest-value (earliest-age) ruins crumble first, so later falls upgrade the collection.

See [Catastrophe](catastrophe.md) for how ruins are generated.

---

## Ancient Civilization Memory

Early in a fresh prestige run, you may find a relic of the civilization you just gave up: an **ancient cache** that remembers something your predecessor knew. Accepting it lets you skip the normal research grind for a single technology, at the cost of a slow rebuild.

> You have discovered an old cache. It appears to contain memories of a now-extinct civilization.

When the cache surfaces, you get an **Accept / Decline** choice offering one technology suited to your current age. Press **A** to accept or **D** to decline.

- **Accept** starts research on the offered tech at once, **free of prerequisites**, skipping the usual age requirement and knowledge cost. It completes at **half research speed** (twice the normal tick count).
- **Decline** and the cache crumbles to dust. Nothing else happens.

Either way, the run's single cache chance is spent the moment the cache is offered.

### When It Fires

| Condition | Requirement |
| --------- | ----------- |
| Timing | **Early** in a new run, while you are still in the **Primitive** or **Stone** age |
| Probability | **~40%** chance, rolled at most **once per prestige run** |
| Prestige level | **Level 1 or higher**. Your very first run has no predecessor to remember, so the cache never appears |
| Already known | Never offers a technology you have already researched |

### One Memory Per Run

There is exactly **one** cache chance per run, and it is used up the moment the cache is offered, **not** when you accept. So:

- **Declining does not let you re-roll.** The chance is spent whether you accept or decline.
- **Saving and reloading does not re-roll it.** Your save remembers that the chance is spent.
- The chance **comes back on the next prestige, Succumb, or reset**, so every new run gets a fresh roll.

### Tech Tier Scales With Prestige Level

Which technology the cache can offer depends on your **prestige level**. At low prestige it offers a tech from near your current age. Higher prestige reaches further: **one extra age of reach per two prestige levels**. A high-prestige player can be offered a tech from an age they haven't reached yet this run, well ahead of the normal curve.

The half-speed penalty applies however advanced the offered tech is, so an out-of-age tech from the cache takes a while to finish. It still skips the prerequisites and knowledge cost it would normally need.

See [Technologies](technologies.md) for the full tech tree and the other free-research path (Grand Discovery).

---

## Recommended Upgrade Priorities

### First prestige (4 to 8 points)

| Priority | Upgrade | Why |
|----------|---------|-----|
| 1st | `starting_food` + `starting_wood` | Makes the Primitive Age much faster; 1 pt each at tier 1 |
| 2nd | `gather_boost` tier 1-2 | Faster early gathering in every run |
| 3rd | `research_speed` tier 1 | More knowledge early means earlier techs and faster ages |

### Second and third prestige (10 to 20 points)

| Priority | Upgrade | Why |
|----------|---------|-----|
| 1st | `storage_bonus` tier 1-2 | Keeps early storage limits from holding back growth |
| 2nd | `research_speed` tier 2-3 | The gains add up at higher tiers |
| 3rd | `population_cap` tier 1 | +2 housing is small, but housing runs short early on |

### Late game (20+ points available)

| Priority | Upgrade | Why |
|----------|---------|-----|
| 1st | `tick_speed`: start buying tiers | The strongest upgrade over many runs |
| 2nd | Max out `research_speed` | All 5 tiers give +25% knowledge output and stack with Ancient Knowledge |
| 3rd | `expedition_loot` | More resources from late-game expeditions |

**Temporal Mastery** (`tick_speed`) is the most expensive upgrade (33 points for tier 5) and also the strongest: each tier makes the whole game tick 5% faster. At tier 5 you're running at 1.25× base speed before passive bonuses.

---

## Tips

- Spend banked points from earlier runs as soon as you log in. You don't need to prestige to spend them.
- Milestones reset on prestige. Each run earns them again, and they count toward that run's points.
- Aim to finish one more wonder each run than the last. Wonder bonuses stack with prestige bonuses.
- Don't rush the first prestige. Reaching later ages (Information, Digital and beyond) first gives many more points than resetting at the Modern Age.
- Prestiging from the Cosmic Era is a gamble. Appease before you confirm to lower the odds, Brace if you would Endure, Invite if you want the Cosmic Legacy.
- The passive bonus grows with every level. At prestige level 10 you have +20% all production and +10% tick speed before spending a single upgrade point.
