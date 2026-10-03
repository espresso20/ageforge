# Prestige System

Prestige is the endgame reset loop. When you reach the **Modern Age** (Age 12), you can give up your entire civilization to earn **Prestige Points** and buy permanent upgrades that carry into every future run. Every age the run completed also runs faster from then on: see [Era Mastery](#era-mastery).

```
prestige confirm yes
```

> Prestige resets your age, resources, buildings, workers, and research. Prestige upgrades, Era Mastery, legacy bonuses from Succumb, the Cosmic Legacy and ruins are **permanent**.

---

## When You Can Prestige

You can prestige from the **Modern Age (Age 12)** or any later age. There is no upper limit: if you push on to the Quantum Age before prestiging, you earn more points.

A first run is paced to reach the Modern Age in about **a week of real time** (the smoke-test bot gets there in about 5.3 days). The ages before it range from 15 minutes (Primitive) to 31h 12m (Atomic); the Modern Age and the ages after it take 31 to 62 hours each. See [How Long Each Age Takes](ages.md#how-long-each-age-takes). The game keeps playing while you are away: offline progress runs for up to 24 hours, at 50% of your normal production. That is the first run. Later runs are faster, because the ages a past run completed run 2x to 4.2x as fast (see [Era Mastery](#era-mastery)).

Prestige is refused while a [catastrophe](catastrophe.md) is pending. Type `catastrophe` and choose Endure or Succumb first. From the Digital or Neon Era, a doom fated for your era that hasn't struck yet ends with the run when you prestige.

In the Cosmic Era, confirming prestige first settles the era's own doom if one is still open, then can bring the [Last Passage](#the-last-passage). If it comes, the prestige waits until you choose Endure or Succumb.

To check your current prestige status:

```
prestige
```

This shows your current level, available points, points you would earn right now, whether you have reached the age prestige needs, and your [Era Mastery](#era-mastery): how fast the age you are in runs and which ages your next prestige would raise. To view the upgrade shop without committing:

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

Points earned per prestige are calculated as:

```
base      = age index  (0 = Primitive, 1 = Stone, ..., 12 = Modern, 20 = Quantum, 21 = Transcendent)
bonus     = floor(milestones / 10) + floor(techs / 15) + floor(structures / 50)
raw       = base + bonus
points    = floor(raw / sqrt(prestige level + 1))
```

Milestones, techs and structures are this run's: all three start over when you prestige or Succumb. Structures are every building you finished this run, wonders included; upgrades don't count, and selling or losing a building doesn't take it back.

The `sqrt(level + 1)` divisor gives **diminishing returns**. It uses your prestige level before this prestige, so your first prestige (level 0) is divided by 1, your second by about 1.41, your third by about 1.73 and your fifth by about 2.24. Every prestige pays at least 1 point.

### What contributes to points

| Source | Points |
|--------|--------|
| Age index (each age beyond Primitive) | 1 each (12 at the Modern Age) |
| Every 10 milestones completed this run | +1 |
| Every 15 techs researched this run | +1 |
| Every 50 structures built this run | +1 |

### A first run, worked through

A first run that prestiges as soon as it enters the Modern Age has usually researched every tech up to the Atomic Age (about 45) and completed 35 to 40 milestones. Say 45 techs, 37 milestones and 475 structures:

```
base   = 12                      (Modern Age)
bonus  = floor(37 / 10)  = 3
       + floor(45 / 15)  = 3
       + floor(475 / 50) = 9
raw    = 12 + 3 + 3 + 9  = 27
points = floor(27 / sqrt(0 + 1)) = 27
```

That is what the smoke-test bot earns on its first prestige: 27 or 28 points. Structures are the part you control most: every 50 more you build is another point, so a run that builds 1,000 structures earns 10 more than one that builds 500.

The same achievements at prestige level 1 pay floor(27 / 1.41) = 19, and at level 2, floor(27 / 1.73) = 15.

Pushing past the Modern Age adds 1 point per age, plus whatever milestones, techs and structures the extra ages bring: usually a point or two per age, for 31 to 62 hours of play each. A run that reaches the Quantum Age (index 20) has a base of 20 before the bonus.

---

## Prestige Upgrades

9 upgrades, each with 5 tiers. Costs are in Prestige Points. All upgrades persist across every reset, including prestige and Succumb.

| Upgrade | Key | Effect per Tier | Max Tier | Cost (T1 to T5) | Total |
|---------|-----|-----------------|----------|-----------------|-------|
| Gather Boost | `gather_boost` | +5% worker output | 5 | 2 / 3 / 4 / 6 / 8 | 23 |
| Storage Bonus | `storage_bonus` | +20 storage for every resource | 5 | 2 / 3 / 4 / 6 / 8 | 23 |
| Knowledge Production | `research_speed` | +5% knowledge production | 5 | 2 / 3 / 5 / 8 / 10 | 28 |
| Military Power | `military_power` | +5% military power | 5 | 2 / 3 / 5 / 8 / 10 | 28 |
| Starting Food | `starting_food` | +25 starting food | 5 | 1 / 2 / 3 / 4 / 5 | 15 |
| Starting Wood | `starting_wood` | +25 starting wood | 5 | 1 / 2 / 3 / 4 / 5 | 15 |
| Housing Bonus | `population_cap` | +2 housing | 5 | 2 / 3 / 5 / 8 / 10 | 28 |
| Expedition Loot | `expedition_loot` | +5% expedition rewards | 5 | 2 / 3 / 5 / 8 / 10 | 28 |
| Temporal Mastery | `tick_speed` | +5% game speed | 5 | 6 / 10 / 17 / 23 / 33 | 89 |

Buying every tier of every upgrade costs 277 points.

```
prestige shop                # view available upgrades and costs
prestige buy gather_boost    # buy the next tier of Gather Boost
prestige buy tick_speed      # buy the next tier of Temporal Mastery
prestige buy starting_food   # buy the next tier of Starting Food
```

You can buy prestige upgrades at any time, not only right after a prestige. Points left over from earlier runs can be spent as soon as you log in.

### Effect Types

| Kind | Upgrades | What each tier does |
|------|----------|---------------------|
| Percentage | Gather Boost, Knowledge Production, Military Power, Expedition Loot, Temporal Mastery | Adds a percentage to that rate. Gather Boost at tier 3 is +15% worker output. |
| Flat | Storage Bonus, Housing Bonus | Adds a flat amount. Storage Bonus at tier 5 is +100 storage for every resource. |
| Starting resource | Starting Food, Starting Wood | Adds to what you start each run with. At tier 5 you begin with +125 food or wood. |

**What each one is worth.**

- **Temporal Mastery** makes ticks come faster: +5% game speed per tier, +25% at tier 5, on top of the game speed from techs and milestone chains. Production, construction, research and every timer in ticks run that much faster in real time. Game speed isn't part of the all-production pool, so [the all-production cap](resources.md#the-all-production-cap) doesn't limit it.
- **Gather Boost** adds a share of your workers' base output on top of everything else. It sits outside the all-production cap too. Late in a run, when other bonuses have multiplied output, the same +5% of base output is a smaller share of the total.
- **Knowledge Production** raises knowledge output, which pays for techs. It does not shorten research times.
- **Military Power** raises your Defense Rating, so your garrison blunts more of a raid and of an Endure (see [Your garrison](catastrophe.md#your-garrison)).
- **Starting Food, Starting Wood, Storage Bonus and Housing Bonus** are flat amounts. They help in the first minutes of a run and barely register after that.
- **Expedition Loot** raises expedition rewards, which are fixed amounts, small next to a late-game economy.

---

## Era Mastery

Every age remembers how many of your runs completed it. That count is the age's **mastery**, from 0 to 10, and an age you have mastered runs faster: its production, storage, build times and research times all move **k** times as fast, where k = 1 + √mastery.

### How mastery grows

- **Prestige raises it.** Each prestige adds one level to every age below the run's furthest age: the age you prestige from, or a deeper one if the run went deeper. A Modern Age prestige completes the Primitive to the Atomic Age, so each of those gains a level. Push on to the Information Age first and the Modern Age gains one too.
- **It stops at 10.** An age at mastery 10 gains nothing more.
- **It is fixed during a run.** Mastery changes only at prestige, so an age's speed never changes while you play through it.
- **Succumb leaves it alone.** A Succumb neither raises nor lowers mastery.
- **It stays with your save.** Mastery is saved with your game and survives every prestige and Succumb. A new game starts from zero.

### How much faster

| Mastery | 0 | 1 | 2 | 3 | 4 | 5 | 6 | 7 | 8 | 9 | 10 |
|---|---|---|---|---|---|---|---|---|---|---|---|
| Speed (k) | 1 | 2 | 2.41 | 2.73 | 3 | 3.24 | 3.45 | 3.65 | 3.83 | 4 | 4.16 |
| Shown in game | 1x | 2x | 2.4x | 2.7x | 3x | 3.2x | 3.4x | 3.6x | 3.8x | 4x | 4.2x |

The first completion doubles an age's speed, and each one after it adds less.

An age at mastery 0 is the **frontier**: no run has completed it, and it runs at 1x. A first run is all frontier, so it plays exactly as it always did.

### What runs faster

On **known ground** (an age running faster than 1x, through mastery or [catch-up](#catch-up)):

- **Production.** Every resource's net rate is multiplied by k. This comes last, after the ×3 production cap and every other bonus, food drain included, so a balanced food supply stays balanced. The `rates` breakdown shows it as its own **Era Mastery** part.
- **Storage.** Every cap is multiplied by k too, so a store holds the same hours of income it would at 1x.
- **Build times** are divided by k, rounded up, never below one tick.
- **Research times** are divided by k after your research speed is applied, rounded up, never below one tick. The Research and Wonders panels show the shortened times.
- **Catastrophe timing.** A fated doom's window and the harbinger's lead shrink with each age's speed (see [Catastrophes on known ground](#catastrophes-on-known-ground)).

What does not change:

- **The clock.** Mastery doesn't make ticks come faster. Each tick produces more, and builds and research need fewer ticks.
- **Timers on the clock.** Events, raids, trade routes, expeditions, campaigns and cooldowns keep their length, so a mastered age, being shorter, holds fewer of them.
- **Prestige points and the upgrade shop.** Both work as described on this page.

### Catch-up

Your **record** is the deepest age you have ever entered, in any run. An age **6 or more ages behind your record** runs at least **4x**, or at its own speed if that is higher (4.2x at mastery 10). With a record in the Interstellar Age, every age up to the Modern Age runs at 4x or better, whatever its mastery.

Catch-up applies after a Succumb too, even on a run that has never prestiged: the rebuild runs the ages 6 or more behind your record at least 4x. Fall in the Steel Era from the Industrial Age, and the next run's Primitive, Stone and Bronze Ages run at 4x.

### The grace rule

When your speed drops, as you step from known ground onto new ground or out of catch-up, storage shrinks with it. Stock already above the new cap is not cut:

- it stays until you spend it;
- production adds nothing to that resource while it is over the cap;
- once it falls under the cap, the grace ends and the resource fills normally again.

Graced stock is saved with your game.

### Catastrophes on known ground

A fated doom's window and the harbinger's lead are measured in each age's target time divided by its speed. In a mastered era the doom and its warning both fall inside the shorter era, and warnings are shorter in the same proportion: at 4x the harbinger's lead in the Iron Age is about 20 minutes to an hour instead of 1.3 to 3.9 hours. Mastery is fixed for the run, so a fate rolled when you enter an era never shifts under you. See [The Harbinger](harbinger.md#when-the-harbinger-arrives).

### Saves from before Era Mastery

The first time a save from before Era Mastery loads, it gets the mastery its past prestiges earned. Every prestige back then was a Modern Age prestige, which completes the Primitive to the Atomic Age, so a save at prestige level L gets mastery L (10 at most) in each of those ages. Its record becomes the deepest of: the age it is in, the Modern Age, the Interstellar Age if it holds the Cosmic Legacy, and the first age of every epoch it succumbed in. This happens once, and the log says so:

```
Era Mastery: ages you have completed now run faster. Primitive to Atomic: mastery 5 (3.2x).
```

A save at prestige level 0 gains nothing.

**Example.** A level-5 save, mid-run in the Industrial Age, loads with mastery 5 (k = 3.24) in every age from the Primitive to the Atomic Age and a record in the Modern Age. The Industrial Age runs 3.2x faster at once. The run's next Modern Age prestige takes those ages to mastery 6 (k = 3.45).

### What you see

- `prestige` and the **Stats** panel show the age you are in: `Era Mastery: known ground, 2.4x faster (mastery 2)`, `known ground, 4x faster (catching up to your record)` or `new ground, 1x`. Under it is what your next prestige adds, for example `Next prestige: the Primitive Age to the Atomic Age gain a mastery level each.`
- The **Epoch** panel (`epoch`) has an **Era Mastery** section: every age up to your record with its mastery and speed (catch-up marked), and the next prestige's gains. It never names an age past your record.
- Entering an age logs its speed: `Known ground: the Bronze Age runs 2.4x faster (mastery 2).`, or `Known ground: the Bronze Age runs 4x faster while you catch up to your record.` Stepping from known ground onto an age no run has completed logs `New ground: the Modern Age runs at 1x until a prestige completes it.`
- A prestige logs the ages that gained, for example `Era Mastery: the Primitive Age to the Atomic Age gained a mastery level each and will run faster.`

### How much faster a run gets

Measured with the smoke-test bot, a near-perfect player, on the one-week curve (median of three seeds):

| Run | To the Modern Age |
|-----|-------------------|
| First run | 5.3 days |
| Second run | 2.3 days: the first run's ages go by 2.3x faster, and the run ends an age deeper |
| Veteran (mastery 10 through the Space Age, record in the Interstellar Age) | 1.3 days |

On known ground the early ages are short: at 4x the Primitive and Stone Ages are paced at about 15 minutes together and the Bronze Age at under an hour. The start of a run rewards checking in often.

---

## Recommended Upgrade Priorities

These follow from the costs and effects above. A first prestige pays about 27 points; later ones pay less as the divisor grows.

### First prestige (about 27 points)

| Priority | Upgrade | Cost | Why |
|----------|---------|------|-----|
| 1st | `tick_speed` tiers 1-2 | 16 | +10% game speed: the whole run, timers included, goes faster, and no cap limits it |
| 2nd | `gather_boost` tiers 1-3 | 9 | +15% worker output, cheap, outside the all-production cap |
| 3rd | `starting_food` and `starting_wood` tier 1 | 2 | Spend the leftovers; they speed up the first minutes |

### Second and third prestige (about 15 to 19 points each)

| Priority | Upgrade | Cost | Why |
|----------|---------|------|-----|
| 1st | `tick_speed` tier 3 | 17 | Another +5% game speed |
| 2nd | `gather_boost` tiers 4-5 | 14 | Finishes Gather Boost at +25% |
| 3rd | `research_speed` tier 1 | 2 | More knowledge for techs |

### Later runs

| Priority | Upgrade | Cost | Why |
|----------|---------|------|-----|
| 1st | `tick_speed` tiers 4-5 | 56 | Temporal Mastery at +25% |
| 2nd | `research_speed` tiers 2-5 | 26 | Knowledge Production at +25% |
| 3rd | `military_power` | 28 for all 5 | Worth it if raids or Endures cost you; skip it if they don't |

Storage Bonus, Housing Bonus and Expedition Loot come last: flat storage and housing stop mattering early in a run, and expedition rewards are small fixed amounts.

---

## What Resets vs Persists

### Resets on Prestige
- All resources (reset to starting amounts: 15 food, 12 wood, plus your Starting Food and Starting Wood upgrades)
- All buildings and build queue
- All workers (recruited and assigned)
- All research (tech tree reverts)
- Milestones and milestone chains
- The run's structure count
- Current epoch and epoch event history
- Age (returns to Primitive Age)
- The run's timers: the ready-to-advance notice, a famine in progress and the Geographic Society's survey countdown

### Persists Across Prestige
- Prestige level and all purchased upgrade tiers
- Era Mastery: every age's mastery, and your record (the deepest age you have ever entered)
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

Each Succumb to a catastrophe grants that epoch's **legacy bonus**, a permanent production bonus for the epoch's main resources, and **Ancient Knowledge**, +25% research speed per distinct epoch succumbed. Both survive prestige and apply from tick 1 of every run. See [Succumb](catastrophe.md#succumb) and the [Legacy Bonus Table](catastrophe.md#legacy-bonus-table).

---

## Ruins at Prestige

When you prestige, any ruins you've accumulated from Succumb events carry forward. Ruins produce at 50% of the base rate and need no workers, so they give free production from tick 1.

On a fresh prestige run with accumulated ruins, your food, wood, or other resources may already be ticking up before you've built a single building. Each Succumb adds up to 8 ruins, and the collection is capped at 24; past the cap the lowest-value (earliest-age) ruins crumble first, so later falls upgrade the collection.

See [Catastrophe](catastrophe.md#succumb) for how ruins are made.

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

## The Last Passage

A [harbinger](harbinger.md) comes to warn of a doom fated inside your era, and the Cosmic Era can hold one too: the Reality Tear. But the Cosmic Era has no next epoch, so its passage is prestige itself: the **Last Passage**. A second harbinger thread warns of it from the moment you enter the era, with a new figure each age: the Distress Beacon, the Elder Relay, your future self, then your unmade self. While the Reality Tear's harbinger speaks, the Last Passage's thread waits behind it with its answers intact, and takes up the warning again once the doom has struck or passed you by.

Prestige from before the Cosmic Era never rolls for it.

### The roll

When you type `prestige confirm yes` in the Cosmic Era, an open Reality Tear settles first (see [The Reality Tear comes first](#the-reality-tear-comes-first)). Then the Last Passage rolls once: 18%, 15% or 12% from low to high faith, by your faith fill at that moment (see [Faith Threshold Bands](faith.md#faith-threshold-bands)). Each level of Appease on the Last Passage's thread multiplies the chance by 0.6 (two levels at most). Invite makes it certain.

`prestige` shows the current chance and which figure is warning of it, for example `☄ The Last Passage: 18% chance (high) when you prestige.` `prestige confirm` spells out what Endure and Succumb would give you before you commit.

- **Nothing comes.** The verdict is Spared, and prestige completes as normal.
- **It comes.** Prestige does **not** complete yet. A choice opens, titled **✦ The Last Passage**, in the same style as the catastrophe choice.

### The Reality Tear comes first

Prestige is the Cosmic Era's passage, so its fated doom can't be outrun past it. If that doom is still open when you confirm prestige:

- **No harbinger yet.** It comes at the prestige, and the prestige waits for one more try: "... Type 'harbinger' to answer, or confirm prestige again to meet it."
- **Then the strike rolls,** before the Last Passage. A hit holds the prestige behind the pending Reality Tear: Endure it, then prestige again, which rolls the Last Passage (or Succumb, which resets the run with no prestige). A miss (spared) lets the same confirm go on to the Last Passage roll.
- **Both pending at once.** If the two are ever pending together, the Reality Tear is answered first: the choice window shows it first, and the Last Passage's Endure and Succumb are refused until it is ("The Reality Tear came first. Answer it before the Last Passage.").

### While it is pending

- **Esc** closes the choice. The status bar shows a **☄ LAST PASSAGE** warning telling you to type `catastrophe` (or `cat`), and a bare `catastrophe` reopens the choice.
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

The result is rounded down, so a small run can keep 0 points. Here Brace changes only the points share: buildings and resources reset anyway, and your soldiers don't change it. The log records a Vindicated verdict, or Fulfilled if you invited it.

**Succumb** earns no points from this run and grants the [Cosmic Legacy](#cosmic-legacy). If you already carry it, Succumb is closed ("You already carry the Cosmic Legacy. Succumb is closed to you.") and Endure is the only choice.

### Choosing it on purpose

Inviting the Last Passage's thread is how you take the Cosmic Legacy on purpose. Invite is free, can't be undone and closes Appease; your next prestige brings the Last Passage. Brace levels still raise the Endure share, in case you change your mind. See [The Harbinger](harbinger.md).

---

## Cosmic Legacy

A one-time, permanent reward for Succumbing to the Last Passage.

- **+10% all production**, active from tick 1 of every run. It adds into the same all-production pool as everything else, so it is limited by [the x3 cap](resources.md#the-all-production-cap) too.
- Shows as **Cosmic Legacy** in the Stats panel, under Active Multipliers and in the Legacy Bonuses list. `prestige` shows `Cosmic Legacy: +10% production (permanent)`.
- Survives every prestige and every Succumb. Only wiping the game clears it.
- You earn it once. While you hold it, Succumb is closed at the Last Passage.

---

## Tips

- Spend banked points from earlier runs as soon as you log in. You don't need to prestige to spend them.
- Milestones and the structure count reset on prestige. Each run earns them again, and they count toward that run's points.
- Build wide before you prestige: every 50 structures is a point, and structures are usually the biggest part of the bonus.
- Pushing past the Modern Age pays in mastery more than in points. Each extra age usually adds a point or two and takes a day or more on a first run, but every age the run completes gains a mastery level, so the next run is faster further in.
- Prestiging from the Cosmic Era is a gamble. Appease before you confirm to lower the odds, Brace if you would Endure, Invite if you want the Cosmic Legacy.
