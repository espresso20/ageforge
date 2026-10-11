# Epochs

Epochs group your run into eras. You progress through 22 ages in order, and those ages are grouped into 7 epochs. Each epoch is an era of human (and post-human) history with its own resources, events and catastrophe.

Ages and epochs do different jobs: **ages are about what you build; epochs are about what happens to you**. You push through an age transition by paying its cost. An epoch transition is a dice roll, shaped by your faith and culture, that fires exactly one large event with lasting consequences.

---

## The 7 Epochs

| # | Icon | Epoch | Ages | Primary Resource | Energy Resource | Catastrophe |
|---|------|-------|------|-----------------|-----------------|-------------|
| 1 | ◈ | Stone Era | Primitive → Stone → Bronze | wood | food | none (catastrophes start in the Iron Era) |
| 2 | ⚔ | Iron Era | Iron → Classical → Medieval | iron | coal | The Great Plague |
| 3 | ⚙ | Steel Era | Renaissance → Colonial → Industrial | steel | coal | The World War |
| 4 | ⚡ | Electric Era | Victorian → Electric → Atomic | steel | electricity | The Nuclear Exchange |
| 5 | ▣ | Digital Era | Modern → Information → Digital | data | electricity | The Great Hack |
| 6 | ◉ | Neon Era | Cyberpunk → Fusion → Space | plasma | plasma | Corporate Armageddon |
| 7 | ✦ | Cosmic Era | Interstellar → Galactic → Quantum → Transcendent | dark matter | antimatter | The Reality Tear (and the Last Passage at prestige) |

The Cosmic Era is the only epoch with 4 ages instead of 3.

On a first run the Modern Age, where a full run's prestige comes, arrives after about a week (the smoke-test bot takes about 4.9 days); prestige opens earlier, at the Medieval Age, as an early taste. From the Bronze Age on, ages and the clocks inside them (event durations, cooldowns, awakenings) run 2.6 times as long as on the earlier three-day curve. The game keeps playing while you are away: offline progress runs for up to 24 hours, at 50% of your normal production.

### What each epoch is like

**◈ Stone Era.** Your settlement scratches out survival. Food is the bottleneck and wood is the building block. Events are small and local: river floods, wandering sages, tribal raids. Shrines are all the faith you have. No catastrophe can strike here.

**⚔ Iron Era.** Iron matters most. Your armies grow, trade routes lengthen and faith starts to carry weight. The Great Plague is the first catastrophe that can strike, and the cheapest one to Succumb to, since the run you give up is still short.

**⚙ Steel Era.** Production reaches industrial scale. A Workers' Uprising costs you 8% of your workers and 15% of your faith, and Coal Seam Discovery adds 30% to your coal production for 468 ticks.

**⚡ Electric Era.** Grid Surge, Oil Strike and Nuclear Theory give small pushes to production and research. This era's catastrophe is The Nuclear Exchange. Nuclear scares and labor movements are short, bearable setbacks.

**▣ Digital Era.** Data replaces iron as the bottleneck. Server Outage, The Great Breach and AI Breakthrough move data and knowledge around in amounts that are small next to Digital Era output. The Great Hack is this era's catastrophe.

**◉ Neon Era.** Plasma is both the primary and the energy resource. Neural Uprising is the nastiest event outside the transition roll: it removes 20% of your workers, steals 500 food and drains food at the same time. Corporate Espionage takes 10K gold and 8K data, a small sum by now.

**✦ Cosmic Era.** Dark matter and antimatter arrive in amounts that make earlier resources look small. Reality fractures and entropy waves are manageable if your production is strong. The Reality Tear catastrophe is the hardest reset decision of the run. With no epoch after it, the Cosmic Era's passage is prestige itself: the [Last Passage](prestige.md#the-last-passage), which a second harbinger thread warns of from the Interstellar Age on.

---

## Age Awakenings (One Per Epoch)

Each epoch has a single **Awakening**: a one-time production boost that fires the first time you enter that epoch's signature age. It is separate from the epoch-event roll described below. An awakening always fires and is always positive (the roll can come up bad), but the boost is temporary and decays. An awakening fires at most once per prestige run and resets on prestige, so the next run can earn it again.

| Epoch | Awakening | Triggers On | Effect |
|-------|-----------|-------------|--------|
| ◈ Stone Era | Pottery Mastery | Stone Age | +1 food/tick, +0.5 stone/tick for ~8 min |
| ⚔ Iron Era | Discovery of Metallurgy | Iron Age | +2 iron/tick for ~43 min |
| ⚙ Steel Era | Steam Breakthrough | Industrial Age | +25% all production for ~17 min |
| ⚡ Electric Era | The Grid Wakes | Victorian Age | +2 electricity/tick, +10% all production for ~26 min |
| ▣ Digital Era | Networks Wake | Modern Age | +2 data/tick, +1 knowledge/tick for ~26 min |
| ◉ Neon Era | Cybernetic Awakening | Cyberpunk Age | +20% all production for ~22 min |
| ✦ Cosmic Era | First Contact Signal | Interstellar Age | +1.5 dark matter/tick, +10% all production for ~35 min |

See [Events](events.md#age-awakenings) for exact durations and how awakenings show up in the active-events panel.

---

## How Epoch Events Work

### The Roll

Every time you cross into a **new epoch** (the first age advance that crosses an epoch boundary), the game rolls your epoch transition event exactly once. The roll never repeats for the same epoch in the same run.

<figure class="screen" data-screen="epoch-arrival"><figcaption>Entering a new epoch, here the Iron Era: the era's name takes the screen first, twice the height of an age's, and its heavy blow lights the whole screen.</figcaption></figure>

The advance that opens an epoch gets a bigger [arrival](ages.md) than any other: two blows build up to a heavy one, the era's name comes before the age's, and your theme's ambient effect plays in the sky if it has one. The event the roll brought is listed with what the age opens.

The outcome is decided in steps.

**Step 1: good or bad?** Your faith strength sets the odds of a good event: 40% below a strength of 25%, 50% from 25% to 75%, 60% above 75%. A typical town reads about 22% and rolls at 40%; it takes more faith buildings than a moderate set, kept staffed, to do better. See [Faith Threshold Bands](faith.md#faith-threshold-bands).

**Step 2: if the roll is bad,** you get a Challenging event, applied at once with no choice. A transition never brings a catastrophe. Entering an era from the Iron Era on also rolls, in secret, whether a doom is fated somewhere inside it; a [harbinger](harbinger.md) always warns before it strikes. See [When It Triggers](catastrophe.md#when-it-triggers).

**Step 3: if the roll is good,** your **culture strength** decides which tiers you can draw from:

| Culture strength | Eligible Tiers |
|------------------|----------------|
| 40% or less | Minor only |
| Over 40% | Minor + Major |
| Over 75% (15% chance) | All tiers (Legendary eligible) |

The Legendary draw is a 15% chance inside the over-75% bracket, so it doesn't fire every time even at full strength. The other 85% of those rolls draw from Minor + Major.

#### Culture strength

Culture strength works like [faith strength](faith.md#faith-threshold-bands). It is the culture you hold that your own culture buildings made, measured against what four and a half **moderate sets** of culture buildings would have made over the same run. A moderate set is five copies of every culture building your age has (the Amphitheater onward; culture buildings take no workers).

```
devotion         = what your culture buildings have made this run ÷ what a moderate set would have made
share kept       = culture you hold ÷ the culture your income has made this run   (at most 100%)
culture strength = devotion × share kept ÷ 4.5                                    (at most 100%)
```

| Your culture buildings | Culture strength (nothing spent) | Good events open |
|------------------------|----------------------------------|------------------|
| none | 0% | Minor |
| a moderate set | 22% | Minor |
| twice that | 44% | Minor + Major |
| three times | 67% | Minor + Major |
| three and a half times | 78% | Minor + Major, and the Legendary on 15% of rolls |

- **Major events open at a devotion over 1.8**, the Legendary over 3.375.
- **Spending culture lowers it.** A festival, a Black Market deal, a monument or an Appease paid in culture takes its share of the strength with it: spend half the culture your income has made and your strength halves.
- **Culture you did not make counts for nothing past what you spent.** Culture bought at the market, an event's gift or a wonder's and a tech's culture can make up for culture you spent; it never raises your strength above what your culture buildings earned.
- It follows the whole run, not the age: nothing jumps when you advance, so the roll reads what the Epochs panel showed you before you advanced. The panel prints it: "Good events at the next era: Minor only", with your culture strength, devotion and share kept under it.

The tier used to read how full culture's storage was. Culture has no store of its own to speak of: it is kept in the general one, which every storage building raises. No town filled it by making culture. A moderate town's whole run of culture, never spent, came to under 8% of the smallest store an era's gate leaves you holding, and under 1% from the Modern Age on; three and a half times the culture buildings reached 19% at best, at the Iron Era's end. The one way over 40% was the market, which sells 3 culture for a gold: from the Medieval Age to the Digital Age, six to twelve minutes of a moderate economy's gold filled a typical store that far. Buying culture no longer opens a tier; building for it does.

| Leaving | Old rule: a moderate town's culture, of its store | Old rule: three and a half times the buildings | Now: moderate | Now: twice | Now: three and a half times |
|---------|---------------------------------------------------|-----------------------------------------------|---------------|------------|-----------------------------|
| the Iron Era (Medieval Age) | 7.5% | 19% | 22%, Minor | 44%, Major | 78%, Legendary |
| the Steel Era (Industrial Age) | 2.5% | 7.9% | 22%, Minor | 44%, Major | 78%, Legendary |
| the Electric Era (Atomic Age) | 0.05% | 0.2% | 22%, Minor | 44%, Major | 78%, Legendary |
| the Digital Era (Digital Age) | 0.01% | 0.04% | 22%, Minor | 44%, Major | 78%, Legendary |
| the Neon Era (Space Age) | 0.002% | 0.007% | 22%, Minor | 44%, Major | 78%, Legendary |

> Faith sets your odds of a good event; culture sets how good it can be. You need both to get the best outcomes reliably.

### Cooldown and Anti-Streak

The epoch transition event has no cooldown; it fires once per epoch. The **regular random events** (the ones that fire during normal play, not at transitions) have a per-event cooldown and an anti-streak rule:

- After 3 good events in a row, the next is bad or mixed (a 3% chance lifts the limit).
- After 2 bad events in a row, the next is good or mixed.
- Each event has its own cooldown (minimum ticks between two occurrences of that event).

This keeps normal play from running into long lucky streaks or long runs of punishment.

### Reading the Epoch panel (`epoch`)

Type **`epoch`** to open the Epoch panel. It shows:

- Current epoch name, icon, and primary/energy resources, and its ages. An age past your next one is counted ("1 more to come"), not named
- The result of your last epoch transition roll
- Your catastrophe status for this era (none so far, foretold, spared, pending, endured or succumbed) and the outlook: the doom a harbinger present foretells, with what it says about when and how likely, or "no harbinger has come. Quiet, for now." The panel never names an era you haven't reached. In the Cosmic Era the next passage is prestige (the Last Passage): the panel shows its odds on a line of their own, below the Reality Tear's warning while that doom's harbinger speaks, and shows **THE LAST PASSAGE** while its choice is pending
- The current [harbinger](harbinger.md), if one is present, and for past harbinger threads this run the chain of figures and the verdict
- Full epoch event history for the current run
- Your legacy bonuses earned across all runs
- **Era Mastery**: every age up to your record (the deepest age you have ever entered) with its mastery and speed, catch-up marked, and what your next prestige adds. Before any age has mastery it shows only the current age's speed. See [Era Mastery](prestige.md#era-mastery)
- The civilization history log (catastrophe decisions, Succumb/Endure records)

---

## Good Epoch Events

10 events across three tiers. You get exactly one per epoch transition when the roll is good.

### Minor events (any culture level)

| Event | Effect | Duration |
|-------|--------|----------|
| Age of Plenty | +100% all production (double) | 562 ticks (~18m 44s) |
| Population Surge | +15% workers added | Instant |
| Ancient Cache | Adds 40% of each resource's storage to that resource | Instant |
| Trade Winds | Gold production +50% | 374 ticks (~12m 28s) |
| Cultural Festival | +30% of your culture and +20% of your faith at once, then culture production +50% and faith production +50%. Never rolled on entering the Iron Era: culture unlocks in the Classical Age | 374 ticks |

### Major events (culture strength over 40%)

| Event | Effect | Duration |
|-------|--------|----------|
| The Grand Discovery | Up to 3 techs from your current age or earlier completed for free (fewer if fewer are left) | Instant |
| Worker Innovation | +10% all production for the rest of the run | Rest of run |
| The Architect's Gift | 10 free copies of your most-built non-wonder building | Instant |
| Peaceful Century | +20% all production | 749 ticks (~24m 58s) |

### Legendary event (culture strength over 75%, 15% chance)

| Event | Effect | Duration |
|-------|--------|----------|
| Epoch Blessing | +15% all production for the rest of the run, recorded in history | Rest of run |

> Worker Innovation and Epoch Blessing last until you Succumb or prestige. Like Age of Plenty and Peaceful Century, they add to all production, and every all-production bonus but a tech's shares one pool. The pool applies in full up to +200% and a quarter of every point past it (see [The all-production cap](resources.md#the-all-production-cap)). A full set of milestones passes +200% from about the Electric Age, so these events lift output most when they land early in a run, and a quarter as much once the pool is past +200%.

---

## Challenging Epoch Events

8 bad events. Every bad transition roll brings one, applied at once, with no choice.

| Event | Effect | Duration |
|-------|--------|----------|
| The Famine | Food production -30% | 312 ticks |
| Merchant Betrayal | Lose half your gold, then gold production -30% | 187 ticks |
| The Great Fire | Up to 8 random buildings destroyed (never wonders or storage) | Instant |
| Epidemic | Lose 20% of workers, then food production -20% | 468 ticks |
| Resource Drought | Epoch's primary resource: production -40% | 234 ticks |
| Political Instability | Lose 60% of your faith, then knowledge production -30% | 156 ticks |
| Economic Crash | Lose half your gold, then gold production -40% | 562 ticks |
| The Dark Age | Current research canceled, lose 80% of your knowledge, then knowledge production -40% | 374 ticks |

The Great Fire and Epidemic are the two you least want to see. Eight lost buildings in the early game set you back a long way. Lost workers come back on their own: with auto-recruit on (the default), the game recruits into the empty worker slots as housing and food allow (see [The roster](workers-and-domains.md#the-roster)), but output dips until they do.

---

## Epoch-Exclusive Random Events

An event's size follows your town: a loss is a share of what you hold, a gain is minutes of your own income, and a timed change is a share of your income for as long as it lasts. [Random Events](events.md#how-big-an-event-is) has the rule and its limits.

Apart from the transition roll, each epoch has 5 events that only enter the random event pool while you're in that epoch. They fire during normal play, not at transitions.

Durations are what each event runs in its own epoch. The Stone Era's events run the listed time in the Primitive and Stone Ages and 2.6 times as long in the Bronze Age; every later era's durations already include that stretch.

### Stone Era ◈

| Event | Sentiment | Effect |
|-------|-----------|--------|
| Tribal Raid | Bad | Lose 8% of your food, -10% workers, food production -20% for 60 ticks |
| Humming Grove | Good | +5 min of wood income, faith production +50% for 120 ticks |
| Beast Stampede | Bad | Lose 10% of your wood, 8% of your food |
| River Blessing | Good | Food production +25% for 144 ticks |
| Wandering Sage | Good | +10 min of knowledge income, +10 min of faith income |

### Iron Era ⚔

| Event | Sentiment | Effect |
|-------|-----------|--------|
| Iron Vein Strike | Good | Iron production +30% for 468 ticks |
| Locust Swarm | Bad | -12% workers, food production -30% for 312 ticks |
| Conquered Village | Good | +10 min of gold income |
| Imperial Road | Good | Gold production +20% for 562 ticks |
| Oracle's Prophecy | Good | Faith production +30%, knowledge production +15% for 374 ticks |

### Steel Era ⚙

| Event | Sentiment | Effect |
|-------|-----------|--------|
| Coal Seam Discovery | Good | Coal production +30% for 468 ticks |
| Workers' Uprising | Bad | Lose 15% of your faith, -8% workers, food production -15% for 312 ticks |
| Colonial Bounty | Good | +12 min of gold income |
| Steam Age Inventor | Good | +8 min of knowledge income, knowledge production +20% for 374 ticks |
| Industrial Blight | Bad | Lose 10% of your faith, food production -20% for 374 ticks |

### Electric Era ⚡

| Event | Sentiment | Effect |
|-------|-----------|--------|
| Grid Surge | Good | Electricity production +30% for 374 ticks |
| Oil Strike | Good | +8 min of gold income, oil production +40% for 468 ticks |
| The Broadcast | Good | +10 min of culture income, faith production +20% for 468 ticks |
| Labor Movement | Bad | Food production -15%, gold production -15% for 156 ticks |
| Nuclear Theory | Good | +10 min of knowledge income, knowledge production +25% for 468 ticks |

### Digital Era ▣

| Event | Sentiment | Effect |
|-------|-----------|--------|
| The Great Breach | Bad | Lose 20% of your data, knowledge production -20% for 312 ticks |
| Viral Moment | Good | +15 min of culture income |
| Tech Monopoly | Good | Gold production +30% for 468 ticks |
| Server Outage | Bad | Data production -40% for 312 ticks |
| AI Breakthrough | Good | Knowledge production +30%, data production +20% for 562 ticks |

### Neon Era ◉

| Event | Sentiment | Effect |
|-------|-----------|--------|
| Plasma Windfall | Good | Plasma production +40%, electricity production +30% for 468 ticks |
| Void Rift | Good | +10 min of dark matter crystals income |
| Neural Uprising | Bad | Lose 12% of your food, -20% workers, food production -15% for 312 ticks |
| Corporate Espionage | Bad | Lose 15% of your gold, 15% of your data |
| Stellar Migration | Mixed | +5 min of food income, food production -15% for 374 ticks |

### Cosmic Era ✦

| Event | Sentiment | Effect |
|-------|-----------|--------|
| Reality Fracture | Bad | Quantum flux production -40%, knowledge production -10% for 312 ticks |
| Dimensional Harvest | Good | +8 min of antimatter income, +8 min of quantum flux income |
| Galactic Council | Good | +10 min of gold income, gold production +20% for 562 ticks |
| Entropy Wave | Bad | Quantum flux production -20%, knowledge production -20% for 374 ticks |
| Transcendence Signal | Good | +15 min of knowledge income, +10 min of culture income |

---

## Endure vs Succumb

A catastrophe strikes at its fated moment inside an era, not at a transition, and waits for you to choose: **Endure** (lose a share of your buildings, stock and workers, and keep the run) or **Succumb** (reset to the Primitive Age for ruins, the epoch's legacy bonus and Ancient Knowledge). The rules and numbers live on one page: see [Endure](catastrophe.md#endure), [Succumb](catastrophe.md#succumb) and the [Legacy Bonus Table](catastrophe.md#legacy-bonus-table).

---

## Random Event Types (Reference)

These are the kinds of effect events can apply:

| Effect | What It Does |
|--------|-------------|
| Gain | Adds minutes of your own income of a resource once (no duration, no active-event entry) |
| Production | Adds a share of your own income of one resource to its rate while the event lasts (negative for a penalty). The amount is worked out when the event fires |
| All production | A percentage bonus or penalty to all production (used by the Endure debuff and by epoch events), added to the all-production pool |
| Loss | Removes a share of what you hold of a resource once |
| Worker loss | Removes a percentage of your workers |

A duration of 0 means the effect happens once. A duration above 0 means the event appears in the active-events panel and ticks down until it expires.

---

## Gameplay Strategy by Playstyle

### The Balanced Approach

Keep about twice a moderate set of faith buildings, fully staffed, for a faith strength in the middle band (about 44%). Invest in culture buildings at a moderate pace. Take transition events as they come without over-optimizing.

In the middle band the odds are a coin flip. With about twice a moderate set of culture buildings (a culture strength over 40%) you're eligible for Major events. You won't hit Legendary, but The Grand Discovery and Worker Innovation are both strong. A first run to the Modern Age crosses 4 epoch transitions (into the Iron, Steel, Electric and Digital Eras), so at a steady 50% good rate you can expect about 2 good events and 2 challenging ones. Catastrophes come separately: at mid faith a first run can expect about 0.6 of them.

Best for: a first or second run, players who don't want to commit hard to one strategy, relaxed sessions.

### Faith Maximizer

Build faith production aggressively: about three and a half times a moderate set of faith buildings, fully staffed, and spend as little faith as you can. That holds your faith strength above 75% going into every epoch transition, and through each era, since a fated doom rolls on your faith strength when it strikes.

Going from the middle band to the top band moves your odds from 50/50 to 60/40. Over the 4 transitions of a first run that's about 0.4 extra good events compared to a neutral run. It also cuts your catastrophe exposure: a fated doom strikes 60% of the time instead of 75% (see [Faith](faith.md#faith-threshold-bands)).

Trade-offs: every worker on a faith building is one not growing food, so you compete with your gathering and farming capacity. Don't let food go critical in the early Stone Era chasing faith.

Best for: players who want steady income and dislike variance. It also suits runs where you plan to Endure rather than Succumb, because good events let you build strength before the hit.

### Culture Rusher

Put culture buildings first: three and a half times a moderate set, kept from the Classical Age on, and spend as little culture as you can, aiming for a culture strength over 75% before each epoch boundary.

The Legendary tier is the goal. Epoch Blessing (+15% all production) and Worker Innovation (+10%) are the two strongest transition events, and both last the rest of the run. They pay off most when they land at the Iron or Steel Era transition: from about the Electric Age a full set of milestones takes the all-production pool past +200%, where a bonus counts a quarter, and a late Blessing then adds about 4 points.

Culture also opens Major events (a culture strength over 40%, about twice a moderate set of culture buildings), which are strictly better than Minor events. The Grand Discovery (up to 3 free techs) and The Architect's Gift (10 free buildings) can skip you ahead a long way.

Trade-offs: culture buildings cost a lot, and culture you spend on festivals, Black Market deals or Appease comes off your strength. This approach is slower to build production in the early epochs but speeds up in the mid-game once Major and Legendary events start landing.

Best for: research-focused runs, players who know the tech tree well, longer sessions where the late-game payoff matters.

### Speed Runner

Spend as little as possible on faith and culture. Build production only. Accept bad events as the cost of faster ages.

Time spent on faith and culture buildings is time not spent on production buildings. Challenging events (not catastrophes) are mostly temporary drains. A Resource Drought or Merchant Betrayal is annoying but recoverable. If your production is high enough, you shrug off events that would cripple a weaker economy.

Worker losses are what bite later on. A Neural Uprising takes 20% of your workers in the Neon Era; auto-recruit refills them as housing and food allow, but output dips until it does. A loss is a share of what you hold, so it is as large late as it is early.

Best for: experienced players who know the age costs, speedrun-minded sessions, players who plan to Succumb quickly anyway.

### Epoch Endurance (Farming)

Delay advancing to the next age on purpose. You stay in the current epoch longer to collect more random events and resources, and more time to build, before the next transition roll.

When is this worth it?
- You're approaching a transition with low faith and low culture. Stay in the epoch, build faith and culture, then cross the boundary when you're ready.
- You're one or two techs away from a building that will improve your odds for the next epoch.
- The next epoch introduces a resource you can't yet produce at scale (for example, entering the Digital Era without data production in place).

When is it a trap?
- Later ages have better production buildings. Staying in an earlier age means slower accumulation overall.
- The longer you stay, the more random events fire, and bad events hurt a civilization that has stopped growing more than one that is still growing.
- Your current epoch's exclusive events matter less once you've outgrown their amounts.

The sweet spot is usually one extra age's worth of time (enough to build a few faith or culture buildings and fill their storage), not three.

---

## Tips and Common Mistakes

**Know which faith band you are in.** Below a faith strength of 25%, your good-event odds are 40% and a fated doom strikes 90% of the time. That is where a typical town sits, and the game is balanced for it. Storage has nothing to do with it: faith strength measures your faith buildings against a moderate set, and how much of your faith you have kept. If you want better odds, build past the moderate set early and keep the buildings staffed. Check the faith row before you advance into a new epoch: the roll reads the strength it shows.

**Culture storage matters as much as culture production.** The tier check uses the share of storage, not the raw amount. A small culture storage at 90% beats a large one at 10%. Don't build more culture storage than you can fill before the transition.

**The anti-streak rule is your safety net during normal play.** After two bad events in a row, the next one is good or mixed; after three good ones, the next is usually bad or mixed. See [Events](events.md#anti-streak-system).

**Know how Challenging events hit.** Merchant Betrayal and Economic Crash take half of whatever gold you hold, and Political Instability takes 60% of your faith, so a bigger stockpile loses more in absolute terms. The per-tick drains (The Famine, Resource Drought, the gold and knowledge drains) are a share of what you make, worked out when the event fires: the log line under the event says what that comes to per tick.

**Know your epoch's primary resource before you cross in.** The Digital Era wants data production running before you arrive. The Neon Era wants plasma reactors. Don't cross an epoch boundary and then find you can't produce the era's core resource.

**Succumb early, Endure late.** In the Iron and Steel Eras the reset costs little, and the legacy bonus, ruins and faster research pay off over many more epochs. The ages the fallen run completed also gain an [Era Mastery](prestige.md#era-mastery) level, so the rebuild takes about half the time. In the Neon and Cosmic Eras your civilization is a huge investment, and Enduring is usually worth the hit.

**Each epoch you Succumb in multiplies research time by 0.8 permanently** (Ancient Knowledge: ×0.8, ×0.64, down to ×0.26 with all six). Repeat Succumbs in the same epoch add nothing, so the value is in collecting different epochs. This is the main argument for a deliberate early Succumb in the Iron Era.

**Listen to the harbinger.** It gives you only part of an age before its doom strikes, so answer promptly. See [The Harbinger](harbinger.md).

---

## Epoch Transitions: What Carries, What Resets

**At epoch transitions (a normal age advance into a new epoch):**
- Your buildings, workers, research and wonders carry on
- Your stockpiles are trimmed, as at every age advance: each resource the new age builds with is capped at 8 times the cost of the cheapest new building that uses it, resources no new building uses keep 10%, and faith is untouched. Hoarding before a boundary doesn't pay. See [What Happens on Age Advance](ages.md#what-happens-on-age-advance)
- A doom still open in the era you leave rolls first, at the advance (see [When It Triggers](catastrophe.md#when-it-triggers))
- The epoch event fires once (the transition roll)
- From the Iron Era on, the new era's fate is rolled in secret: a doom is fated somewhere inside it 27% of the time
- Active events from the previous epoch keep ticking down
- The random event pool shifts to include the new epoch's exclusive events
- The Epoch panel records the transition and its outcome
- Your status bar icon and color change

**After Succumb:** the run starts over in the Primitive Age with your ruins, legacy bonuses, Ancient Knowledge, prestige points, [legacy kit](prestige.md#the-legacy-kit) and [Era Mastery](prestige.md#era-mastery). The kit remembers the fallen run's plan, roster and civilizations, and the items you own work on the rebuild. Every age the fallen run completed gains a mastery level, as at a prestige, so the rebuild runs them at least 2x, and ages 6 or more behind your record run at least 4x ([catch-up](prestige.md#catch-up)). See [What carries forward](catastrophe.md#what-carries-forward).

**After Prestige:** like Succumb, but chosen, from the Medieval Age on, and it earns prestige points to spend on the [legacy kit](prestige.md#the-legacy-kit): every age the run completed pays, and each era's ages pay three times what the era before paid, so a prestige before the Modern Age is an early taste that pays little. It is refused while a catastrophe is pending. Before the Cosmic Era a doom that hasn't struck yet ends with the run, and from the Cosmic Era it can bring the [Last Passage](prestige.md#the-last-passage). Every age below the run's furthest age gains a mastery level, so it runs faster on the next run. The per-run epoch event history is cleared; the civilization log carries. See [What Resets vs Persists](prestige.md#what-resets-vs-persists).

Each run builds on the last. Three runs in, you have ruins producing for free, stacked research speed, legacy bonuses on the resources that matter most, and the ages your runs completed going 2.4 to 4 times as fast, and you still play through all 22 ages.
