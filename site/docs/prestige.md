# Prestige System

Prestige is the reset loop. From the **Medieval Age** (Age 5) on, you can give up your entire civilization to earn **Prestige Points**, and the deeper the run went, the more it pays: every age the run completed adds points, and each era's ages are worth three times the era before. Points buy the **legacy kit**, three items that carry your run's automation into every future run: your build plan, your worker shares and the civilizations you met. Every age the run completed also runs faster from then on: see [Era Mastery](#era-mastery).

```
prestige confirm yes
```

> Prestige resets your age, resources, buildings, workers, and research. Prestige points, the legacy kit and what it remembers, Era Mastery, legacy bonuses from Succumb, the Cosmic Legacy and ruins are **permanent**.

---

## When You Can Prestige

You can prestige from the **Medieval Age (Age 5)** or any later age. A prestige from the Medieval Age to the Atomic Age is an **early taste**: it is optional and pays little (see [Early Tastes and Full Runs](#early-tastes-and-full-runs)). From the **Modern Age (Age 12)** on, a prestige is a full run. There is no upper limit: every age you complete before prestiging adds points, and each era's ages are worth three times the last era's.

A first run is paced to reach the Medieval Age in about 20 hours and the Modern Age in about **a week of real time** (the smoke-test bot gets there in about 4.9 days). The ages before the Modern Age range from 15 minutes (Primitive) to 31h 12m (Atomic); the Modern Age and the ages after it take 31 to 62 hours each. See [How Long Each Age Takes](ages.md#how-long-each-age-takes). The game keeps playing while you are away: offline progress runs for up to 24 hours, at 50% of your normal production. That is the first run. Later runs are faster, because the ages a past run completed run 2x to 4.2x as fast (see [Era Mastery](#era-mastery)).

Prestige is refused while a [catastrophe](catastrophe.md) is pending. Type `catastrophe` and choose Endure or Succumb first. Before the Cosmic Era, a doom fated for your era that hasn't struck yet ends with the run when you prestige, so a taste in the Medieval Age escapes an Iron Era doom, at the price of the run.

In the Cosmic Era, confirming prestige first settles the era's own doom if one is still open, then can bring the [Last Passage](#the-last-passage). If it comes, the prestige waits until you choose Endure or Succumb.

To check your current prestige status:

```
prestige
```

This shows your current level, available points, how many legacy kit items you own, the points you would earn right now and what prestiging from the next age would pay instead (and how many more), and your [Era Mastery](#era-mastery): how fast the age you are in runs and which ages your next prestige would raise. Before the Modern Age it adds a note that a prestige now is an early taste. Before the Medieval Age it says where prestige opens and what a prestige there pays. To view the shop without committing:

<figure class="screen" data-screen="prestige"><figcaption>Prestige has no panel of its own: the command answers in the log, read here in the Logs panel in the Medieval Age, where an early taste would pay 9 points.</figcaption></figure>

```
prestige shop
```

When you're ready:

```
prestige confirm yes
```

`prestige confirm` on its own first says how many points you would earn, that everything else resets, and what you keep: prestige points, the legacy kit and what it remembers, and Era Mastery. Before the Modern Age it also says that this is an early taste, what it pays and what a full run pays (see [Early Tastes and Full Runs](#early-tastes-and-full-runs)). The double confirmation (`confirm yes`) is deliberate, because prestige can't be undone.

The **Stats** panel (`stats`) has a Prestige section too: your level, Era Mastery, points, what a prestige pays now and from the next age, and the kit items you own.

Every prestige, from any age, ends with one closing line in the log, written for the age the run ended in. Before the Cosmic Era the civilization simply winds down: in the Modern Age the offices empty and the last tram runs, in the Space Age the orbital yards shut. In the Cosmic Era the age's harbinger is there at the end, and the lines turn to cosmic dread ("Your unmade self took your hand.").

---

## Prestige Points Formula

A prestige pays **depth points**. Every age the run completed (each age before the one you prestige from) adds its weight, and the weight triples with each era:

```
weight of an age = 3 ^ era    (era 0 = Stone Era, 1 = Iron Era, ..., 6 = Cosmic Era)
points           = the weights of every age before the one you prestige from, added up
```

Nothing else counts. Milestones, techs and structures add no points, and there is no divisor: your prestige level changes nothing, so your tenth prestige from the Modern Age pays the same 120 points as your first.

### What contributes to points

| Era | Ages | Points for each age completed |
|-----|------|-------------------------------|
| Stone Era | Primitive, Stone, Bronze | 1 |
| Iron Era | Iron, Classical, Medieval | 3 |
| Steel Era | Renaissance, Colonial, Industrial | 9 |
| Electric Era | Victorian, Electric, Atomic | 27 |
| Digital Era | Modern, Information, Digital | 81 |
| Neon Era | Cyberpunk, Fusion, Space | 243 |
| Cosmic Era | Interstellar, Galactic, Quantum, Transcendent | 729 |

The 22 ages weigh 4,008 in all. The last age can't be completed, so the most a prestige pays is 3,279, from the Transcendent Age.

### A first run, worked through

A first run that prestiges as soon as it enters the Modern Age has completed the twelve ages before it:

```
Stone Era      3 ages × 1   =   3
Iron Era       3 ages × 3   =   9
Steel Era      3 ages × 9   =  27
Electric Era   3 ages × 27  =  81
points                      = 120
```

Stop earlier or push on, and the points by the age you prestige from are:

| Era | Points |
|-----|--------|
| Iron Era | Medieval 9 |
| Steel Era | Renaissance 12, Colonial 21, Industrial 30 |
| Electric Era | Victorian 39, Electric 66, Atomic 93 |
| Digital Era | Modern 120, Information 201, Digital 282 |
| Neon Era | Cyberpunk 363, Fusion 606, Space 849 |
| Cosmic Era | Interstellar 1,092, Galactic 1,821, Quantum 2,550, Transcendent 3,279 |

Pushing past the Modern Age adds 81 points for each Digital Era age you complete, 31 to 42 hours of play each on a first run. Each Neon Era age adds 243, and each Cosmic Era age 729. A prestige from the Cyberpunk Age is a run through the Digital Age (363 points), and one from the Interstellar Age a run through the Space Age (1,092).

In the Cosmic Era the [Last Passage](#the-last-passage) can take part of a run's points: Endure keeps 50%, 70% or 85% of them, and Succumb keeps none.

---

## Early Tastes and Full Runs

Prestige opens at the Medieval Age, but the points reward depth. A prestige from the Medieval Age to the Atomic Age is an **early taste**; from the Modern Age on it is a **full run**.

- **A taste pays little.** On a first run, a prestige pays about 12 points per day of play at the Medieval Age, 24 at the Modern Age and 38 for a run through the Digital Age (from the smoke-test bot's first-run times). Going deeper pays far more per day.
- **A second taste adds little.** A Medieval Age reset right after a Modern Age run adds 9 points to that run's 120 (7.5%).
- **What a taste is for.** Its 9 points buy the [Plan Template](#plan-template), the first kit item. It raises [Era Mastery](#era-mastery) like any prestige: every age below the one you prestiged from gains a level, so a Medieval Age taste speeds up the Primitive to the Classical Age. And it ends an Iron Era doom that hasn't struck yet, though it costs you the run.
- **The game says so, before and after.** `prestige confirm` before the Modern Age carries a line like this one, and the new run's log repeats it in the past tense right after "Prestige complete":

  ```
  This is an early taste: a prestige from the Medieval Age pays 9 prestige points, for the 5 ages this run completed. Going deeper pays far more: each era's ages are worth 3 times the era before, and a full run, 7 ages further on, pays 120 prestige points.
  ```

  Once you have seen the Modern Age, in this run or an earlier one, the line names it: "and a run to the Modern Age pays 120 prestige points."
- **Your account tells them apart.** Each prestige is recorded under the age it was made from (see [Lifetime stats & achievements](account.md#lifetime-stats-amp-achievements)). Total Prestiges still counts every prestige, tastes included.

---

## The Legacy Kit

The prestige shop sells the **legacy kit**: three items that carry your run's automation across every prestige, and across a Succumb too. Each is bought once and kept for good. Costs are in Prestige Points.

| Item | Key | Cost | What it does |
|------|-----|------|--------------|
| Plan Template | `legacy_plan` | 9 | Your build plan carries over: each age's part of the plan you wrote is added again when you enter that age |
| Worker Shares | `legacy_workers` | 36 | Your worker shares carry over to each new run |
| Old Friends | `legacy_factions` | 54 | Civilizations you have met are met again as soon as your age reaches theirs, at neutral opinion |

The whole kit costs 99 points: a Medieval Age taste (9 points) buys the Plan Template, and a first Modern Age run (120) buys the rest and leaves 21.

```
prestige shop                  # the kit: each item's price (or "owned") and what the kit remembers
prestige buy legacy_plan       # buy the Plan Template
prestige buy legacy_workers    # buy Worker Shares
```

You can buy kit items at any time, not only right after a prestige. Points left over from earlier runs can be spent as soon as you log in.

### The kit remembers before you buy it

From your first prestige or Succumb on, the game remembers your runs whether or not you own any kit items: the plan you wrote, your worker shares and every civilization you met. An item bought later puts that memory to work at once, on the run you are in: the Plan Template adds the current age's part of the template to the plan (unless you have already planned something in this age), your remembered worker shares are set (unless you have set some), and remembered civilizations within reach are met.

`prestige shop` lists what the kit remembers under the items, for example:

```
  What the kit remembers from your runs
  Plan: 41 items over 9 ages.
  Worker shares: 3 domains.
  Civilizations met: 4.
```

Before your first prestige the shop says: "The kit remembers your plan, worker shares and the civilizations you meet. Nothing is remembered yet: it starts with your first prestige."

### Plan Template

**What it records.** As you write your [build plan](plan.md), the game records each item with the age you added it in: builds, techs, trades and advances (`plan build`, `plan research`, `plan trade`, `plan advance`). A build counts the copies you added, and removing an item takes back the copies it never started, so a build's count ends as what you started plus what is still waiting. A trade you remove after it has bought something stays recorded, as a trade for the amount it bought. Deals (`plan deal`) are not recorded, because a civilization's offers end with the run. Each age records up to 60 items.

**When it becomes the template.** At each prestige and Succumb, age by age. An age the run wrote in takes this run's part (with the Plan Template owned, so does every age the run entered). Ages the run never reached keep the part an older run wrote, so a short run never wipes a deeper one: a Medieval Age taste after a Modern Age run keeps that run's plan for the Renaissance to the Atomic Age.

**What it does.** With the Plan Template owned, the start of every run and every advance add that age's part of the template to the plan, the advance item included if you planned one. So a plan written once chains ages while you are away: the plan advances, the next age's part goes in, the plan works through it and advances again. Each item goes through the same checks as the plan commands and the plan's 60-item limit, and one log line says what went in:

```
Plan Template: added 12 items for the Bronze Age.
```

If the plan is full, the line also says how many items waited out, and an item that can't be planned yet (a building at its limit, say) is skipped and counted. The items the template adds count as written again, so the template carries forward from run to run, and what you add or remove changes it for next time.

**Your techs come along.** The techs you queue with `plan research` are recorded like any other item, so with the Plan Template they are planned again on later runs, in the age you planned them in. That is your own research path: the game never chooses what you research next. Only the techs in your plan start without you, and a tech you start by hand with `research` is not recorded.

### Worker Shares

With Worker Shares owned, the [worker shares](workers-and-domains.md#worker-shares) you set carry over into each new run, and into the rebuild after a Succumb, instead of going back to auto. The new run's log says so ("Worker Shares: your shares carry over.", with the split). Auto-recruit was already kept across prestige.

The kit remembers the shares you had when the run ended. If a run ends with every domain on auto, the kit keeps the shares it remembered before, unless you own Worker Shares, in which case the next run starts on auto too.

### Old Friends

Every civilization you meet is remembered, from every run. With Old Friends owned, each is met again as soon as your age reaches its own, with no mission and no wait for the two-age fallback, at neutral opinion, as at first contact. The log says so:

```
Old friends: the Riverlands Tribes remember your people and make contact again.
```

A civilization you have never met still has to be found by sending missions (see [Meeting civilizations](factions.md#meeting-civilizations)).

---

## The Old Shop and Its Refund

Before the legacy kit, the shop sold nine perks bought in tiers: Gather Boost, Storage Bonus, Knowledge Production, Military Power, Starting Food, Starting Wood, Housing Bonus, Expedition Loot and Temporal Mastery. They are retired: hidden, with no effect, and they can't be bought (`prestige buy` refuses them and says their points were refunded).

The first time a save from before the legacy kit loads, the game refunds the old perks once, after checking the save's signature. The save is marked, so it never happens twice.

1. Your **old points** are what you spent on the perks, at their old prices, plus the points you had left.
2. They are converted at the new rate: each past prestige counts as one Modern Age run (120 points), or your old points × 4.44 (120 new points for every 27 old: a first Modern Age run used to pay 27), if that is more.
3. Your available points and your lifetime total both become the refund, and the old tiers go to 0. Your prestige level is kept.

One line in the log says what you got:

```
The prestige shop changed. Your old perks were refunded as 600 points (5 prestiges at 120 each).
```

When your old points at the new rate are the bigger number, the line ends "(your N old points at the new rate)" instead.

| Save | Old points | Refund |
|------|------------|--------|
| Level 5: 83 points spent on perks, 3 left | 86 | 5 × 120 = **600** (86 × 4.44 = 382 is less) |
| Level 38, every perk at its top tier (277 points spent) | 277 or more | 38 × 120 = **4,560** |
| Level 0 | none | nothing |

The level-5 save's 600 points buy the whole kit (99) and leave 501. Its next Modern Age prestige pays 120 points; its last one under the old formula paid 12.

---

## Era Mastery

Every age remembers how many of your runs completed it and then ended, in a prestige or a Succumb. That count is the age's **mastery**, from 0 to 10, and an age you have mastered runs faster: its production, storage, build times and research times all move **k** times as fast, where k = 1 + √mastery.

### How mastery grows

- **Prestige raises it.** Each prestige adds one level to every age below the run's furthest age: the age you prestige from, or a deeper one if the run went deeper. A Modern Age prestige completes the Primitive to the Atomic Age, so each of those gains a level. Push on to the Information Age first and the Modern Age gains one too. An early taste counts the same way: a Medieval Age prestige raises the Primitive to the Classical Age.
- **It stops at 10.** An age at mastery 10 gains nothing more.
- **A Succumb raises it too.** A Succumb ends the run, so every age that run completed gains a level, exactly as at a prestige, and the rebuild starts on known ground. A first run that falls in the Renaissance Age comes back with mastery 1 in the Primitive to the Medieval Age. It still earns no prestige points.
- **No age is counted twice for one run.** The count starts over with each new run. The prestige after a Succumb raises only the ages the rebuilt run completed: fall in the Renaissance Age, rebuild, prestige from the Medieval Age, and the Primitive to the Classical Age are at mastery 2 while the Medieval Age, completed once, stays at 1.
- **It is fixed during a run.** Mastery changes only when a run ends, so an age's speed never changes while you play through it.
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
- **Prestige points and the legacy kit.** Both work as described on this page.

### Catch-up

Your **record** is the deepest age you have ever entered, in any run. An age **6 or more ages behind your record** runs at least **4x**, or at its own speed if that is higher (4.2x at mastery 10). With a record in the Interstellar Age, every age up to the Modern Age runs at 4x or better, whatever its mastery.

Catch-up applies after a Succumb too, even on a run that has never prestiged: the rebuild runs the ages 6 or more behind your record at least 4x. Fall in the Steel Era from the Industrial Age, and the next run's Primitive, Stone and Bronze Ages run at 4x; the Iron to the Colonial Age, which the fallen run completed, run at 2x.

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
- The **Epoch** panel (`epoch`) has an **Era Mastery** section: every age up to your record with its mastery and speed (catch-up marked), and the next prestige's gains. It never names an age past your record. Before any age has mastery it shows only the current age's speed and a one-line hint.
- Entering an age logs its speed: `Known ground: the Bronze Age runs 2.4x faster (mastery 2).`, or `Known ground: the Bronze Age runs 4x faster while you catch up to your record.` Stepping from known ground onto an age no run has completed logs `New ground: the Modern Age runs at 1x until a run completes it.`
- A prestige or a Succumb logs the ages that gained, for example `Era Mastery: the Primitive Age to the Atomic Age gained a mastery level each and will run faster.`

### How much faster a run gets

Measured with the smoke-test bot, a near-perfect player, on the one-week curve (medians: eight seeds for the first two runs, three for the others):

| Run | To the Modern Age |
|-----|-------------------|
| First run | 4.9 days |
| Second run | 2.2 days: the first run's ages go by 2.2x faster, and a run played as long as the first ends an age deeper |
| Third run | 1.9 days |
| Veteran (mastery 10 through the Space Age, record in the Interstellar Age) | 1.2 days |

On known ground the early ages are short: at 4x the Primitive and Stone Ages are paced at about 15 minutes together and the Bronze Age at under an hour. The start of a run rewards checking in often.

---

## Recommended Kit Priorities

The kit is priced in the order most players should buy it. A Medieval Age taste pays for the first item, and a first Modern Age run pays for the rest.

| Priority | Item | Cost | Why |
|----------|------|------|-----|
| 1st | `legacy_plan` | 9 | The plan you wrote comes back age by age, so a later run keeps building, researching and advancing while you are away |
| 2nd | `legacy_workers` | 36 | Your worker split is in place from the first worker of every run. If you leave every domain on auto, buy it last |
| 3rd | `legacy_factions` | 54 | Every civilization you have met is back the moment you reach its age, with its trade deals, and no missions spent finding it |

Once you own all three, the shop has nothing more to sell, and the points you earn stay banked.

---

## What Resets vs Persists

### Resets on Prestige
- All resources (reset to starting amounts: 15 food, 12 wood)
- All buildings and build queue
- The build plan (with the Plan Template owned, the first age's part of the template goes back in at once)
- All workers (recruited and assigned). Worker shares go back to auto, unless you own Worker Shares
- All research (tech tree reverts)
- Milestones and milestone chains
- The run's structure count
- The civilizations you have met, with their opinion of you, and your trade routes (with Old Friends owned, each civilization is met again when you reach its age)
- Current epoch and epoch event history
- Age (returns to Primitive Age)
- The run's timers: the ready-to-advance notice, a famine in progress and the Geographic Society's survey countdown

### Persists Across Prestige
- Prestige level, prestige points and the legacy kit items you own
- What the kit remembers: the plan template (the techs you planned included), your worker shares and the civilizations you have met (it is used only by the items you own)
- Your auto-recruit and wonder overflow settings
- Era Mastery: every age's mastery, and your record (the deepest age you have ever entered)
- Ruins (from past Succumb events), which carry into the new run
- Legacy bonuses (from Succumb events), active from tick 1
- The Cosmic Legacy, if you have earned it
- Ancient Knowledge (research time ×0.8 for each distinct epoch succumbed in), which comes from your legacy bonuses, so prestige never drops it
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

When you type `prestige confirm yes` in the Cosmic Era, an open Reality Tear settles first (see [The Reality Tear comes first](#the-reality-tear-comes-first)). Then the Last Passage rolls once: 18%, 15% or 12% from low to high faith, by your faith fill at that moment (see [Faith Threshold Bands](faith.md#faith-threshold-bands)). Each level of Appease on the Last Passage's thread multiplies the chance by 0.6 (two levels at most; the first costs 1B faith and 16B culture, the second double: see [What it costs](harbinger.md#what-it-costs-by-epoch)). Invite makes it certain.

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

Bracing for the Last Passage takes effort. Its price is set by what the Interstellar Age makes, not by what the era's advances ask: level 1 costs 20Q dark matter and 24Q titanium, about 22 hours of a moderate economy's income, and level 2 double. Appease level 1 (1B faith, 16B culture, about 32 hours) lowers the chance instead. For what each is worth, see [Brace or Appease against the Last Passage](harbinger.md#brace-or-appease-against-the-last-passage).

**Succumb** earns no points from this run and grants the [Cosmic Legacy](#cosmic-legacy). If you already carry it, Succumb is closed ("You already carry the Cosmic Legacy. Succumb is closed to you.") and Endure is the only choice.

### Choosing it on purpose

Inviting the Last Passage's thread is how you take the Cosmic Legacy on purpose. Invite is free, can't be undone and closes Appease; your next prestige brings the Last Passage. Brace levels still raise the Endure share, in case you change your mind. See [The Harbinger](harbinger.md).

---

## Cosmic Legacy

A one-time, permanent reward for Succumbing to the Last Passage.

- **+10% all production**, active from tick 1 of every run, in every age. It is not part of the all-production pool: everything a resource makes is multiplied by 1.1 after [the x3 cap](resources.md#the-all-production-cap) and every other bonus, so no cap can hold it back. (It used to add into the pool, where it did nothing once the pool was full, from about the Victorian Age on.)
- It multiplies what you make, before the food your workers eat is taken off, so it never deepens a food shortage.
- Shows as **Cosmic Legacy** in the Stats panel, under Active Multipliers (`Cosmic Legacy ×1.10`, beside the pool, like morale) and in the Legacy Bonuses list. `rates` shows what it adds to each resource as its own **Cosmic Legacy** part. `prestige` shows `Cosmic Legacy: +10% production (permanent)`.
- Survives every prestige and every Succumb. Only wiping the game clears it.
- You earn it once. While you hold it, Succumb is closed at the Last Passage.

---

## Tips

- Spend banked points from earlier runs as soon as you log in. You don't need to prestige to spend them, and a kit item works the moment you buy it.
- Write your plan the way you want the next run to play it, `plan advance` included. The Plan Template replays each age's part when you enter that age, so a plan that ends in an advance chains into the next age's part while you are away.
- Queue your techs with `plan research` instead of starting them by hand. Only planned techs are recorded, so they are the ones the Plan Template lines up again on the next run.
- Don't rush the first prestige. A Medieval Age taste pays about 12 points per day of play, a Modern Age prestige about 24 and a run through the Digital Age about 38.
- Take a taste for what it does, not for its points: it buys the Plan Template early, or escapes an Iron Era doom.
- Milestones and the structure count reset on prestige, and neither changes the points.
- Pushing past the Modern Age pays in points and in mastery. Each Digital Era age the run completes adds 81 points, and every age it completes gains a mastery level, so the next run is faster further in.
- Prestiging from the Cosmic Era is a gamble. Appease before you confirm to lower the odds, Brace if you would Endure, Invite if you want the Cosmic Legacy.
