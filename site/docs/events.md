# Random events

Random events are the windfalls, droughts, raids, plagues and gold rushes that strike while you play. You do not choose them or when they come. Each one's size follows your town.

This page covers the base random events and the rules every random event shares. Each epoch adds five events of its own to the same pool, and every epoch transition rolls one transition event. Both are covered in [Epochs](epochs.md).

---

## How events fire

### The timer

A global timer spaces out all random events. After an event fires, the next one cannot fire for 150 to 600 ticks (about 5 to 20 minutes of real time), picked at random in that range. From the Bronze Age on the wait is 390 to 1,560 ticks (about 13 to 52 minutes). The first event of a game waits for the same window, which gives you time to get established.

Times on this page are at 2 seconds per tick. Tick speed bonuses make every one of them shorter (see [Timers and durations](commands.md#timers-and-durations)).

Time away fires no random events. Offline progress credits production, construction and research, but it rolls no events.

### Weighted random selection

Each event has a weight. The higher the weight, the more likely the event. When the timer runs out, the game builds a pool of all eligible events and draws one by weight, so a weight 15 event is five times as likely as a weight 3 event from the same pool. The pool holds the base events and the exclusive events of the epoch you are in.

An event is eligible only if:

- the run has reached the event's earliest tick,
- the current age is at or past the event's min age,
- the event's own cooldown has passed since it last fired (Plague, for example, has a 200-tick cooldown, 520 from the Bronze Age on),
- the event is not already active (a timed event cannot overlap itself).

### Anti-streak system

The game counts consecutive good and bad events and limits streaks:

| Situation | Rule |
|-----------|------|
| 3 good events in a row | The next draw is bad or mixed (3% of the time the limit is lifted and any event can come) |
| 2 bad events in a row | The next draw is good or mixed |
| A mixed event fires | Resets both counters |

So at most two bad events come in a row, and a run of good events ends after three. Base and epoch-exclusive events count alike.

Other timed effects join the active events list directly: milestone chain boosts, festivals, awakenings, epoch events and Endure's Reconstruction Effort. They skip the timer and the streak rules, and they count down in the active events list like any timed event.

### Duration and active events

Events are either instant or timed:

- An instant event applies its effect once, when it fires. The log line is all you see.
- A timed event stays active for its duration, and its per-tick effects apply the whole time. When it expires, the log prints a line like "Drought has ended."

Every duration and cooldown is set as a base length, which is exact in the Primitive and Stone Ages. From the Bronze Age on, a timed event lasts 2.6 times its base length (rounded to the nearest tick) and its cooldown is 2.6 times as long. The tables below give each event's duration as it runs. Events that cannot fire before the Bronze Age always run the stretched length. The three timed events that can fire earlier (Storm, Drought and Plague) show both.

### Seeing what happened

The main log reports each event as it fires: a sentence of its own, then what it did, in the amounts that were applied. For example `Lost 800 food and 1K wood.`, `Gained 2.4K food.` or `Food -8.2/tick for ~52s.` A loss counts only what actually left your stock. A research gain reads as a share of the tech in progress. With VibeMode on, the names and the sentences read in a different script, and the numbers, odds and durations stay the same (see [VibeMode](commands.md#vibemode)).

A timed event's log line says how long it lasts in the age it fired in. A Drought reads `for ~20s` in the Stone Age and `for ~52s` from the Bronze Age on.

The Stats panel (`stats`) lists every active timed event under "Active events", with the time left on it (for example `~2m 30s`) and its ongoing change to production, in green for a bonus and red for a penalty. A production boost shows as `+10% all production`. On the colorblind-safe and high-contrast [themes](themes.md), bonuses show blue and penalties orange or amber. A change to research speed is not listed there, and neither are instant effects (gifts, thefts, worker loss), since they already happened.

The Logs panel (`logs`) shows the last 50 lines with tick numbers. It also lists the active events with the time left.

---

## How big an event is

An event's size follows the town it happens to. Nothing is a fixed amount:

- A loss takes a share of what you hold of the resource, from 5% to 20% depending on the event.
- A gain gives minutes of your own income of the resource, from 3 to 15 minutes of what your town makes of it, all at once.
- A timed change is a share of your income of the resource for as long as the event lasts. A Drought takes half of your food production, and a Gold Rush doubles your gold production.

In the Primitive Age, for example, a Beast Stampede takes 80 of 1,000 food and 40 of 500.

Each amount is worked out when the event fires and held to a band that depends on the age. The band is counted in the age's typical income of the resource, which is what a moderate town of that age makes of it:

- A gain or a boost is sized on your income, but on at least a quarter and at most four times the typical income. A town that makes little of the resource still gets something, and a very large town gets no windfall out of scale with its age.
- A setback to production has no floor: a town that makes none of the resource loses none. Its ceiling is the same as a boost's.
- A loss is at least six seconds and at most half an hour of the typical income, and never more than a quarter of what you hold.

An effect on a resource you have not unlocked does nothing. The log line states the real amounts: what you lost, what you gained, each change per tick and how long it lasts.

---

## Base event reference

These 26 events belong to no epoch. They can fire in any epoch from their min age on. The 35 exclusive events (5 per epoch) bring the game's tables to 61 random events. One of those, AI Breakthrough, can never fire (see [Epochs](epochs.md#epoch-exclusive-random-events)).

### Good events

| Event | Min age | Weight | Effect | Duration | Notes |
|-------|---------|--------|--------|----------|-------|
| Bountiful Harvest | Primitive | 15 | +5 min of food income | Instant | Highest weight of any good event |
| Wandering Traders | Bronze | 12 | +5 min of gold income, +3 min of food income | Instant | |
| Skilled Immigrants | Stone | 10 | Research progress worth 5 min of your scholars' work | Instant | |
| Gold Rush | Bronze | 8 | Gold production +100% | 39 ticks (~1m 18s) | |
| Trade Boom | Medieval | 8 | Gold production +100% | 52 ticks (~1m 44s) | |
| Ancient Discovery | Iron | 6 | Research progress worth 10 min of your scholars' work | Instant | |
| Renaissance Fair | Renaissance | 10 | Culture production +50%, gold production +50% | 39 ticks (~1m 18s) | |
| Colonial Windfall | Colonial | 8 | +8 min of gold income, +5 min of culture income | Instant | |
| Power Surge | Victorian | 6 | Electricity production +100% | 26 ticks (~52s) | |
| Crypto Boom | Fusion | 7 | Crypto production +100% | 39 ticks (~1m 18s) | |
| First Contact | Space | 3 | Research progress worth 15 min of your scholars' work, +5 min of titanium income | Instant | |
| Dark Matter Rift | Interstellar | 4 | Dark matter production +100% | 39 ticks (~1m 18s) | |
| Quantum Fluctuation | Quantum | 3 | Quantum flux production +100% | 26 ticks (~52s) | |

### Bad events

| Event | Min age | Weight | Effect | Duration | Notes |
|-------|---------|--------|--------|----------|-------|
| Storm | Primitive | 14 | Wood production -50% | 5 ticks (~10s); 13 (~26s) from Bronze | Highest weight of any bad event |
| Drought | Primitive | 12 | Food production -50% | 10 ticks (~20s); 26 (~52s) from Bronze | |
| Bandit Raid | Bronze | 10 | Lose 6% of your food, 6% of your gold | Instant | Raid |
| Plague | Stone | 6 | -15% workers, food production -50% | 8 ticks (~16s); 21 (~42s) from Bronze | Takes workers |
| Mine Collapse | Iron | 7 | -5% workers, iron production -50% | 21 ticks (~42s) | Takes workers |
| Heresy | Medieval | 5 | Faith production -50% | 31 ticks (~1m 2s) | |
| Pirate Attack | Colonial | 7 | Lose 8% of your gold, 5% of your food | Instant | Raid |
| Nuclear Scare | Atomic | 4 | Electricity production -40%, research speed -30% | 31 ticks (~1m 2s) | |
| Data Breach | Information | 6 | Lose 10% of your data, 8% of your gold | Instant | Raid |
| Industrial Accident | Industrial | 8 | Lose 6% of your steel, 6% of your oil, -7% workers | Instant | Takes workers |
| Crypto Winter | Cyberpunk | 8 | Lose 12% of your crypto | Instant | |

### Mixed events

| Event | Min age | Weight | Effect | Duration | Notes |
|-------|---------|--------|--------|----------|-------|
| Earthquake | Stone | 5 | Lose 5% of your wood, +5 min of stone income | Instant | |
| Plasma Storm | Fusion | 5 | Electricity production -50%, plasma production +50% | 26 ticks (~52s) | Costs power, adds plasma |

---

## Raids and your garrison

Some bad events are raids: attacks by outsiders. If you have soldiers, your garrison blunts part of a raid. It cuts the resources the raid takes and the workers it drives off by the share shown in the Army panel (`army`), at most 45%. The share depends on your army's Defense against the raid Threat of your current age. See [Military](military.md). Some techs also cut what a raid takes. With no soldiers, your garrison blunts nothing.

The raids are:

| Event | Where | What the garrison blunts |
|-------|-------|--------------------------|
| Bandit Raid | base event, Bronze Age on | food and gold taken |
| Pirate Attack | base event, Colonial Age on | gold and food taken |
| Data Breach | base event, Information Age on | data and gold taken |
| Tribal Raid | Stone Era | food taken and workers who flee (not the food production penalty) |
| Beast Stampede | Stone Era | wood and food lost |
| The Great Breach | Digital Era | data taken (not the research speed penalty) |
| Corporate Espionage | Neon Era | gold and data taken |

A military building comes with its soldiers the moment it stands, and a War Camp can be built in the Stone Age. A garrison can blunt the two Stone Era raids and a Bandit Raid in the Bronze Age too (see [Military](military.md)). Only the stolen resources and lost workers are blunted. A raid's production penalty runs in full. Disasters and unrest are not raids, and soldiers do nothing against them. Plague, Mine Collapse, Industrial Accident, Crypto Winter, Earthquake and the rest take their full toll.

When the garrison blunts a raid, the log adds a green line under the event, for example:

```
Your garrison blunted about 28% of the raid: you kept 168 food and 210 gold.
```

A raid that the garrison blunts also costs the army. A red line reports the soldiers who fell defending the town: a twentieth of the soldiers at home, rounded up, and at least one. What the garrison kept is added to the "Saved this run" line in the Army panel.

---

## Effect types

| Effect | What it does |
|--------|--------------|
| Gain | Adds minutes of your own income of a resource once, when the event fires. |
| Research gain | Adds work to the tech in progress once, as minutes of your scholars' work. At most half that tech's price lands, and with no tech in progress it is lost. |
| Production change | Adds a share of your own income of one resource to its per-tick rate (or takes it off) for as long as the event lasts. The amount is worked out when the event fires and stays the same until it ends. |
| Research speed change | Adds or takes off a share of your scholars' work for as long as the event lasts. It holds whenever scholars are at work. |
| All production change | A percentage added to or taken off the all-production pool for as long as the event lasts. Epoch events, awakenings and Endure's Reconstruction Effort use it. The base random events do not. |
| Loss | Removes a share of what you hold of a resource once, when the event fires. |
| Worker loss | Removes a percentage of your workers when the event fires, from every building in proportion. The workers do not return when the event ends. |

Every random event also nudges morale when it fires: a good event lifts it by 4 points and a bad one lowers it by 4. Mixed events leave it alone. This applies to the base events and the epoch-exclusive events alike. Epoch transition events do not change morale. See [Morale](morale.md).

Worker loss is the only effect that is not undone when the event ends. Production changes end with the event, and lost stock builds up again as you produce. The event does not give the workers back, but your population grows again toward your housing while food allows.

---

## Epoch-exclusive random events

Each epoch adds 5 events to the random event pool while you are in that epoch. They follow the same timer, weights, cooldowns and streak rules as the base events. The epoch is the only difference: they fire only in it.

| Epoch | Events |
|-------|--------|
| Stone Era | Tribal Raid, Humming Grove, Beast Stampede, River Blessing, Wandering Sage |
| Iron Era | Iron Vein Strike, Locust Swarm, Conquered Village, Imperial Road, Oracle's Prophecy |
| Steel Era | Coal Seam Discovery, Workers' Uprising, Colonial Bounty, Steam Age Inventor, Industrial Blight |
| Electric Era | Grid Surge, Oil Strike, The Broadcast, Labor Movement, Nuclear Theory |
| Digital Era | The Great Breach, Viral Moment, Tech Monopoly, Server Outage, AI Breakthrough |
| Neon Era | Plasma Windfall, Void Rift, Neural Uprising, Corporate Espionage, Stellar Migration |
| Cosmic Era | Reality Fracture, Dimensional Harvest, Galactic Council, Entropy Wave, Transcendence Signal |

[Epochs](epochs.md#epoch-exclusive-random-events) lists each one's effect, weight, min age and duration.

Epoch-exclusive random events are not epoch transition events. A transition event fires once per epoch, at the boundary, and depends on your faith and culture. Faith and culture have no effect on whether a base or an epoch-exclusive random event fires. Only timing, weights, cooldowns and the streak rules apply.

---

## Age awakenings

Awakenings are one-time boosts, one per epoch, that fire the first time you enter an epoch's signature age: seven epochs, seven awakenings. An awakening always fires, always gives a production boost that fits its era, and never has a downside. The boost is temporary and ends after its duration. Each awakening fires at most once per run. The record of which ones fired is saved with your game, so a reload does not fire one again, and it clears when a run ends (a prestige or a Succumb), so the next run can earn them all again.

The log prints a line that reads "Awakening:" followed by the awakening's name, its effect and how long it lasts.

| Epoch | Trigger age | Awakening | Effect | Duration |
|-------|-------------|-----------|--------|----------|
| Stone Era ◈ | Stone Age | Pottery Mastery | +1.0 food/tick, +0.5 stone/tick | 250 ticks (~8m 20s) |
| Iron Era ⚔ | Iron Age | Discovery of Metallurgy | +2.0 iron/tick | 1,300 ticks (~43m 20s) |
| Steel Era ⚙ | Industrial Age | Steam Breakthrough | +25% all production | 520 ticks (~17m 20s) |
| Electric Era ⚡ | Victorian Age | The Grid Wakes | +2.0 electricity/tick, +10% all production | 780 ticks (~26m) |
| Digital Era ▣ | Modern Age | Networks Wake | +2.0 data/tick, +1.0 research/tick | 780 ticks (~26m) |
| Neon Era ◉ | Cyberpunk Age | Cybernetic Awakening | +20% all production | 650 ticks (~21m 40s) |
| Cosmic Era ✦ | Interstellar Age | First Contact Signal | +1.5 dark matter/tick, +10% all production | 1,040 ticks (~34m 40s) |

Every awakening after the Stone Age's fires from the Bronze Age on, so its duration already includes the 2.6 stretch. The "+% all production" awakenings add to the all-production pool, which applies in full however large it grows (see [The all-production pool](resources.md#the-all-production-pool)). An awakening appears in the active events list like any timed event and counts down over its duration.

---

## Responding to events

- Production penalties (Drought, Storm, Heresy, Nuclear Scare) are short and you cannot avoid them, so ride them out. Do not restructure your workers over a penalty that ends in about a minute.
- Losses (the raids and the losses in other events) are gone the moment the event fires. A loss is a share of what you hold, so a big stockpile loses more than a thin one. A garrison blunts the raids among them.
- Worker loss (Plague, Mine Collapse, Industrial Accident) takes workers from every building in proportion. With auto-recruit on (the default), the game fills the empty slots as your population grows back. With it off, use `recruit` and `assign`. See [People](workers-and-domains.md).
- Good events need no preparation. A boost such as Gold Rush or Crypto Boom adds a share of what you make, worked out when it fires, so let it run. A gain lands in your stock in full.
- Research windfalls (Skilled Immigrants, Ancient Discovery, First Contact and some epoch events) go to the tech in progress. They never give more than half of that tech's price, so they finish a tech only when its bar is already past halfway. With nothing in progress they are lost, and the log says "Your scholars had nothing to study, so the insight was lost." Start a tech before you hope for one.
- Faith and culture shape the epoch transition roll, not random events. Stacking faith does not make Bountiful Harvest more likely or Drought less likely. See [Epochs](epochs.md).
