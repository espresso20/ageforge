# Epochs

Your run passes through 22 ages, and those ages are grouped into 7 epochs. Ages are about what you build. Epochs are about what happens to you. Each epoch has its own primary and energy resources and its own random events. Entering an epoch rolls one transition event, good or bad, and from the Iron Era on each epoch can hold one doom.

You advance an age by meeting its requirements (see [The Advance Panel](advance.md)). You do not choose an epoch event. The roll depends on your faith and your culture.

---

## The 7 epochs

| # | Icon | Epoch | Ages | Primary resource | Energy resource | Catastrophe |
|---|------|-------|------|------------------|-----------------|-------------|
| 1 | ◈ | Stone Era | Primitive, Stone, Bronze | wood | food | none (catastrophes start in the Iron Era) |
| 2 | ⚔ | Iron Era | Iron, Classical, Medieval | iron | coal | The Great Plague |
| 3 | ⚙ | Steel Era | Renaissance, Colonial, Industrial | steel | coal | The World War |
| 4 | ⚡ | Electric Era | Victorian, Electric, Atomic | steel | electricity | The Nuclear Exchange |
| 5 | ▣ | Digital Era | Modern, Information, Digital | data | electricity | The Great Hack |
| 6 | ◉ | Neon Era | Cyberpunk, Fusion, Space | plasma | plasma | Corporate Armageddon |
| 7 | ✦ | Cosmic Era | Interstellar, Galactic, Quantum, Transcendent | dark matter | antimatter | The Reality Tear (and the Last Passage at prestige) |

- Every run starts in the Stone Era, so it has no transition roll, and no catastrophe can strike there.
- The Cosmic Era has four ages. The other six epochs have three.
- The Cosmic Era has no epoch after it. Its passage is prestige itself, the [Last Passage](prestige.md#the-last-passage). A second harbinger thread warns of it from the Interstellar Age on.
- The Resource Drought event hits the primary resource of the epoch you enter.

---

## Age awakenings

Each epoch has one awakening: a boost that fires the first time you enter that epoch's signature age. An awakening is not part of the transition roll. It always fires, it is always good, and it ends when its duration runs out. It fires at most once per run. The record of which ones have fired is saved with your game and clears when a run ends (a prestige or a Succumb), so the next run earns each one again.

| Epoch | Awakening | Triggers on | Effect |
|-------|-----------|-------------|--------|
| ◈ Stone Era | Pottery Mastery | Stone Age | +1 food/tick, +0.5 stone/tick |
| ⚔ Iron Era | Discovery of Metallurgy | Iron Age | +2 iron/tick |
| ⚙ Steel Era | Steam Breakthrough | Industrial Age | +25% all production |
| ⚡ Electric Era | The Grid Wakes | Victorian Age | +2 electricity/tick, +10% all production |
| ▣ Digital Era | Networks Wake | Modern Age | +2 data/tick, +1 research/tick |
| ◉ Neon Era | Cybernetic Awakening | Cyberpunk Age | +20% all production |
| ✦ Cosmic Era | First Contact Signal | Interstellar Age | +1.5 dark matter/tick, +10% all production |

[Random events](events.md#age-awakenings) lists each awakening's duration and shows how it appears in the log and the Stats panel.

---

## How the transition roll works

### The roll

The first age advance that crosses into an epoch rolls that epoch's transition event once. The roll never repeats for the same epoch in the same run.

<figure class="screen" data-screen="epoch-arrival"><figcaption>Entering an epoch, here the Iron Era: the era's name takes the screen first, twice the height of an age's where it fits, and its heavy blow lights the whole screen.</figcaption></figure>

The advance that opens an epoch plays a longer [arrival](ages.md) than any other advance. The era's name comes before the age's, two blows build up to a heavy one, and your theme's ambient effect plays if it has one. The event the roll brought is listed with what the age opens.

Faith strength sets the odds of a good event:

| Faith strength | Chance of a good event |
|----------------|------------------------|
| Under 25% | 40% |
| 25% to 75% | 50% |
| Over 75% | 60% |

A moderate town reads about 22%, so it rolls at 40%. Doing better takes more faith buildings than a moderate set, kept staffed. See [Faith threshold bands](faith.md#faith-threshold-bands).

A bad roll brings one of the 8 [challenging events](#challenging-epoch-events), picked at random and applied at once. You make no choice. A transition never brings a catastrophe.

A good roll brings one event from a pool that your culture strength sets. Every event in the pool is equally likely.

| Culture strength | Good events in the pool |
|------------------|-------------------------|
| 40% or less | Minor |
| Over 40% | Minor and Major |
| Over 75% | Minor and Major, and on 15% of rolls the Legendary event too |

The Legendary tier holds one event, Epoch Blessing. On the other 85% of rolls at that strength the pool is the same as for Major. The Cultural Festival is left out of the pool while culture is locked, which it is on entering the Iron Era (culture unlocks in the Classical Age).

From the Iron Era on, entering an epoch also rolls in secret whether a doom is fated somewhere inside it. It is fated 27% of the time. A [harbinger](harbinger.md) always warns before a doom strikes. See [When it triggers](catastrophe.md#when-it-triggers).

### Culture strength

Culture strength works like [faith strength](faith.md#faith-threshold-bands). It is the culture you hold that your own culture buildings made, measured against what four and a half moderate sets of culture buildings would have made over the same run. A moderate set is eight copies of every culture building your age has, the Amphitheater onward. Culture buildings take no workers.

```
devotion         = what your culture buildings have made this run ÷ what a moderate set would have made
share kept       = culture you hold ÷ the culture your income has made this run   (at most 100%)
culture strength = devotion × share kept ÷ 4.5                                    (at most 100%)
```

| Your culture buildings | Culture strength (nothing spent) | Good events open |
|------------------------|----------------------------------|------------------|
| none | 0% | Minor |
| a moderate set | 22% | Minor |
| twice that | 44% | Minor and Major |
| three times | 67% | Minor and Major |
| three and a half times | 78% | Minor and Major, and the Legendary on 15% of rolls |

- Major events open at a devotion over 1.8. The Legendary opens over 3.375.
- Spending culture lowers your strength. A festival, a Black Market deal, a monument or an Appease paid in culture takes its share of the strength with it: spend half the culture your income has made and your strength halves.
- Culture you did not make counts for nothing. A wonder's or a tech's culture or an event's gift can make up for culture you spent, but it never lifts your strength above what your culture buildings earned.
- Culture strength follows the whole run, not the age. Nothing jumps when you advance, so the roll reads what the Epoch panel showed you before you advanced.
- Culture is locked until the Classical Age. On entering the Iron Era your culture strength is 0%, so only Minor events can come.

Faith sets the odds of a good event. Culture sets which good events can come.

### Cooldown and streaks

The transition event has no cooldown and no streak rule. It fires once per epoch per run. The random events that fire during normal play have their own cooldowns and an [anti-streak rule](events.md#anti-streak-system).

---

## The Epoch panel

Type `epoch` to open the Epoch panel. From the top it shows:

- Your current epoch: its icon and name, its ages, and its primary and energy resources. An age you cannot see yet is counted ("1 more to come"), not named.
- The epoch event your transition roll brought, with its duration and a line about it, or "No event this epoch".
- Once culture is unlocked and an epoch remains to come, "Good events at the next era", for example "Minor only", with your culture strength, devotion and share kept under it.
- Your catastrophe status for this era: none so far, foretold, spared, pending, endured or succumbed. In the Stone Era it reads "none in the Stone Era".
- The outlook: the doom a present harbinger foretells, with what it says about when and how likely, or "no harbinger has come. Quiet, for now." The panel never names an era you have not reached. In the Cosmic Era the next passage is prestige (the Last Passage). Its odds get a line of their own, below the Reality Tear's warning while that doom's harbinger speaks, and the panel shows the Last Passage as pending while its choice waits.
- The current [harbinger](harbinger.md), if one is present, with its Appease and Brace levels.
- Harbingers: the ones that have come and gone this run, each with its chain of figures and the verdict.
- Epoch history: each epoch you have entered, with the transition event it brought (or "no event") and how its catastrophe ended.
- Legacy bonuses from Succumb.
- Era Mastery: every age up to your record (the deepest age you have ever entered) with its mastery and speed, catch-up marked, and what your next prestige adds. Before any age has mastery it shows only the current age's speed. See [Era Mastery](prestige.md#era-mastery).
- The civilization log: your Endure and Succumb decisions.

---

## Good epoch events

There are 10 good events in three tiers. A good roll brings exactly one. The log prints the event's name and what it does, and lines under it with the amounts that applied.

### Minor events (any culture strength)

| Event | Effect | Duration |
|-------|--------|----------|
| Age of Plenty | All production +100% (double) | 562 ticks (~18m 44s) |
| Population Surge | +15% workers added | Instant |
| Ancient Cache | You gain 15 minutes of your town's income of every unlocked resource, all of it landing in your stock. In research it lands on the tech in progress: at most half that tech's price, at most the price of the cheapest tech of the age you entered, and lost with nothing in progress | Instant |
| Trade Winds | Gold production +50% | 374 ticks (~12m 28s) |
| Cultural Festival | +30% of the culture you hold and +20% of the faith you hold at once, then culture production +50% and faith production +50%. Not drawn while culture is locked | 374 ticks (~12m 28s) |

### Major events (culture strength over 40%)

| Event | Effect | Duration |
|-------|--------|----------|
| The Grand Discovery | Completes up to 3 techs you can start (age reached, prerequisites done) for free, cheapest first | Instant |
| Worker Innovation | All production +10% for the rest of the run | Rest of run |
| The Architect's Gift | 10 free copies of the non-wonder building you have the most of | Instant |
| Peaceful Century | All production +20% | 749 ticks (~24m 58s) |

### Legendary event (culture strength over 75%, on 15% of rolls)

| Event | Effect | Duration |
|-------|--------|----------|
| Epoch Blessing | All production +15% for the rest of the run | Rest of run |

Every event in the pool is equally likely, so even with the Legendary tier open Epoch Blessing is one draw in ten.

Worker Innovation and Epoch Blessing last until you Succumb or prestige. Like Age of Plenty and Peaceful Century, they add to the all-production pool, which applies in full however large it grows (see [The all-production pool](resources.md#the-all-production-pool)). Each adds its whole percentage to the pool, so each is worth the same number of points whenever it lands.

---

## Challenging epoch events

There are 8 bad events. Every bad transition roll brings one, picked at random and applied at once, with no choice.

| Event | Effect | Duration |
|-------|--------|----------|
| The Famine | Food production -30% | 312 ticks (~10m 24s) |
| Merchant Betrayal | Lose half your gold, then gold production -30% | 187 ticks (~6m 14s) |
| The Great Fire | Up to 8 random buildings destroyed (never wonders, storage or knowledge buildings). Their workers go back to the idle pool | Instant |
| Epidemic | Food production -20%, and 20% of your workers die | 468 ticks (~15m 36s) |
| Resource Drought | Production of the epoch's primary resource -40% | 234 ticks (~7m 48s) |
| Political Instability | Lose 60% of your faith, then research speed -30% | 156 ticks (~5m 12s) |
| Economic Crash | Lose half your gold, then gold production -40% | 562 ticks (~18m 44s) |
| The Dark Age | Current research canceled and its work lost, then research speed -40% | 374 ticks (~12m 28s) |

The gold and faith losses are a share of what you hold, so a bigger stockpile loses more. The production drains are a share of what you make, worked out when the event fires. The log line under the event says what that comes to per tick.

Lost workers do not come back when an event ends. Your population grows again toward your housing while food allows. With auto-recruit on (the default) the game fills empty worker slots as it grows (see [The roster](workers-and-domains.md#the-roster)).

---

## Epoch-exclusive random events

An event's size follows your town: a loss is a share of what you hold, a gain is minutes of your own income, and a timed change is a share of your income for as long as it lasts. [Random events](events.md#how-big-an-event-is) has the rule and its limits.

Apart from the transition roll, each epoch lists 5 events that enter the random event pool only while you are in that epoch. They fire during normal play, not at transitions. They follow the same timer, weights, cooldowns and streak rules as the base events.

Min age is the earliest age an event can fire in. An event fires only in its own epoch's ages, so one whose min age is the epoch's last age fires only there. Weight sets how likely the event is against the other events in the pool.

The Stone Era's timed events run the listed time in the Primitive and Stone Ages and 2.6 times as long from the Bronze Age on. Every later epoch's events fire from the Iron Age on, so their durations already include that stretch.

### Stone Era ◈

| Event | Kind | Min age | Weight | Effect | Duration |
|-------|------|---------|--------|--------|----------|
| Tribal Raid | Bad, raid | Primitive | 10 | Lose 8% of your food, -10% workers, food production -20% | 60 ticks (~2m); 156 (~5m 12s) from Bronze |
| Humming Grove | Good | Primitive | 10 | +5 min of wood income, faith production +50% | 120 ticks (~4m); 312 (~10m 24s) from Bronze |
| Beast Stampede | Bad, raid | Primitive | 8 | Lose 10% of your wood, 8% of your food | Instant |
| River Blessing | Good | Primitive | 12 | Food production +25% | 144 ticks (~4m 48s); 374 (~12m 28s) from Bronze |
| Wandering Sage | Good | Primitive | 7 | research progress worth 10 min of your scholars' work, +10 min of faith income | Instant |

### Iron Era ⚔

| Event | Kind | Min age | Weight | Effect | Duration |
|-------|------|---------|--------|--------|----------|
| Iron Vein Strike | Good | Iron | 10 | Iron production +30% | 468 ticks (~15m 36s) |
| Locust Swarm | Bad | Iron | 9 | -12% workers, food production -30% | 312 ticks (~10m 24s) |
| Conquered Village | Good | Iron | 8 | +10 min of gold income | Instant |
| Imperial Road | Good | Iron | 8 | Gold production +20% | 562 ticks (~18m 44s) |
| Oracle's Prophecy | Good | Iron | 7 | Faith production +30%, research speed +15% | 374 ticks (~12m 28s) |

### Steel Era ⚙

| Event | Kind | Min age | Weight | Effect | Duration |
|-------|------|---------|--------|--------|----------|
| Coal Seam Discovery | Good | Industrial | 10 | Coal production +30% | 468 ticks (~15m 36s) |
| Workers' Uprising | Bad | Industrial | 9 | Lose 15% of your faith, -8% workers, food production -15% | 312 ticks (~10m 24s) |
| Colonial Bounty | Good | Colonial | 8 | +12 min of gold income | Instant |
| Steam Age Inventor | Good | Industrial | 7 | research progress worth 8 min of your scholars' work, research speed +20% | 374 ticks (~12m 28s) |
| Industrial Blight | Bad | Industrial | 9 | Lose 10% of your faith, food production -20% | 374 ticks (~12m 28s) |

### Electric Era ⚡

| Event | Kind | Min age | Weight | Effect | Duration |
|-------|------|---------|--------|--------|----------|
| Grid Surge | Good | Victorian | 10 | Electricity production +30% | 374 ticks (~12m 28s) |
| Oil Strike | Good | Victorian | 8 | +8 min of gold income, oil production +40% | 468 ticks (~15m 36s) |
| The Broadcast | Good | Victorian | 8 | +10 min of culture income, faith production +20% | 468 ticks (~15m 36s) |
| Labor Movement | Bad | Victorian | 9 | Food production -15%, gold production -15% | 156 ticks (~5m 12s) |
| Nuclear Theory | Good | Atomic | 6 | research progress worth 10 min of your scholars' work, research speed +25% | 468 ticks (~15m 36s) |

### Digital Era ▣

| Event | Kind | Min age | Weight | Effect | Duration |
|-------|------|---------|--------|--------|----------|
| The Great Breach | Bad, raid | Information | 9 | Lose 20% of your data, research speed -20% | 312 ticks (~10m 24s) |
| Viral Moment | Good | Information | 10 | +15 min of culture income | Instant |
| Tech Monopoly | Good | Information | 8 | Gold production +30% | 468 ticks (~15m 36s) |
| Server Outage | Bad | Information | 9 | Data production -40% | 312 ticks (~10m 24s) |
| AI Breakthrough | Good | Cyberpunk | 6 | Research speed +30%, data production +20% | 562 ticks (~18m 44s) |

AI Breakthrough's min age, the Cyberpunk Age, belongs to the Neon Era, but the event fires only in the Digital Era, so it never fires.

### Neon Era ◉

| Event | Kind | Min age | Weight | Effect | Duration |
|-------|------|---------|--------|--------|----------|
| Plasma Windfall | Good | Fusion | 10 | Plasma production +40%, electricity production +30% | 468 ticks (~15m 36s) |
| Void Rift | Good | Fusion | 7 | +10 min of dark matter crystals income | Instant |
| Neural Uprising | Bad | Fusion | 9 | Lose 12% of your food, -20% workers, food production -15% | 312 ticks (~10m 24s) |
| Corporate Espionage | Bad, raid | Fusion | 8 | Lose 15% of your gold, 15% of your data | Instant |
| Stellar Migration | Mixed | Space | 8 | +5 min of food income, food production -15% | 374 ticks (~12m 28s) |

### Cosmic Era ✦

| Event | Kind | Min age | Weight | Effect | Duration |
|-------|------|---------|--------|--------|----------|
| Reality Fracture | Bad | Quantum | 9 | Quantum flux production -40%, research speed -10% | 312 ticks (~10m 24s) |
| Dimensional Harvest | Good | Quantum | 8 | +8 min of antimatter income, +8 min of quantum flux income | Instant |
| Galactic Council | Good | Quantum | 6 | +10 min of gold income, gold production +20% | 562 ticks (~18m 44s) |
| Entropy Wave | Bad | Quantum | 8 | Quantum flux production -20%, research speed -20% | 374 ticks (~12m 28s) |
| Transcendence Signal | Good | Quantum | 4 | research progress worth 15 min of your scholars' work, +10 min of culture income | Instant |

The raids among these are blunted by your garrison. See [Raids and your garrison](events.md#raids-and-your-garrison).

---

## Endure and Succumb

A catastrophe strikes at its fated moment inside an era, not at a transition, and waits for you to choose. **Endure** costs you a share of your buildings, stock and workers and keeps the run. **Succumb** resets the run to the Primitive Age for ruins, the epoch's legacy bonus and Ancient Knowledge. The rules and numbers live on one page: see [Endure](catastrophe.md#endure), [Succumb](catastrophe.md#succumb) and the [Legacy Bonus Table](catastrophe.md#legacy-bonus-table).

Each distinct epoch you Succumb in multiplies your research time by 0.8 (Ancient Knowledge: ×0.8, ×0.64, down to ×0.26 with all six). Repeat Succumbs in the same epoch add nothing.

---

## Working with the roll

- Faith strength sets the odds of a good roll, and culture strength sets which good events can come. Both read the whole run, so buildings put up just before an advance change little.
- Spending lowers either strength by the share spent. An Appease costs faith, and a festival or a monument costs culture. Check the Epoch panel's "Good events at the next era" line, and the faith row in the Economy panel, before you advance. The roll reads the strengths they show.
- The transition roll picks its event at random from the pool. A bad roll takes half your gold in Merchant Betrayal and Economic Crash and 60% of your faith in Political Instability, so a bigger stockpile loses more.
- A fated doom strikes with a chance that depends on your faith strength at the strike: 90% under 25%, 75% from 25% to 75%, and 60% over 75% (see [Faith](faith.md#faith-threshold-bands)).

---

## What carries over at a transition

At a transition (a normal age advance into the next epoch):

- Your buildings, workers, research and wonders carry on.
- Your stockpiles are trimmed, as at every age advance: each resource the next age builds with is capped at 8 times the cost of the cheapest building that uses it, resources no building of the age uses keep 10%, and faith is untouched. See [What happens on age advance](ages.md#what-happens-on-age-advance).
- A doom still open in the era you leave settles first, at the advance (see [When it triggers](catastrophe.md#when-it-triggers)).
- The transition event fires once.
- From the Iron Era on, the era's fate is rolled in secret: a doom is fated somewhere inside it 27% of the time.
- Timed events from the epoch you left keep counting down.
- The random event pool switches to the exclusive events of the epoch you enter.
- The Epoch panel records the transition and its outcome.

A Succumb or a prestige starts the next run in the Primitive Age. It clears which transition rolls and awakenings have fired, so the next run rolls them again. A prestige also clears the Epoch panel's history and harbinger record. A Succumb keeps both. The civilization log carries across both. See [What carries forward](catastrophe.md#what-carries-forward) and [What resets vs persists](prestige.md#what-resets-vs-persists).

A prestige is refused while a catastrophe is pending.
