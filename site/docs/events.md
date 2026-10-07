# Random Events

Random events are the fires, windfalls, plagues, gold rushes and quantum fluctuations that happen throughout your run, independent of the choices you make. They keep the pace changing and reward players who pay attention and react.

> **Two kinds of events come from the epoch system, and this page covers neither in depth.** This page covers the **base random events**, which fire during normal play by timer and weighted chance. Each epoch also adds its own **epoch-exclusive random events** to the same pool, and every epoch transition fires one **epoch transition event** shaped by your faith and culture. Both are documented in [Epochs](epochs.md).

---

## How Events Fire

### The Timing Window

A **global timer** spaces out all random events. After an event fires, the next one can't fire for 150 to 600 ticks (about 5 to 20 minutes of real time), picked at random in that range. From the Bronze Age on, where each age runs 2.6 times as long, the wait is 390 to 1,560 ticks (about 13 to 52 minutes), so an age still sees about as many events. The first event of a new game waits for the same window, giving you time to get established.

Real times on this page are at 2 seconds per tick; tick speed bonuses make every one of them shorter (see [Timers and durations](commands.md#timers-and-durations)).

### Weighted Random Selection

Each event has a weight. The higher the weight, the more likely it is. When the timer runs out, the game builds a pool of all eligible events and draws one by weight, so a weight 15 event is five times more likely to be drawn than a weight 3 event from the same pool.

An event is eligible only if:
- the run has reached the event's earliest tick (some events can't fire before they make sense),
- the current age is at or past the event's minimum age,
- the event's own cooldown has passed since it last fired (`plague`, for example, has a 200-tick cooldown, 520 ticks from the Bronze Age on),
- the event is not currently active (a timed event cannot overlap with itself).

### Anti-Streak System

The game tracks consecutive good and bad events and limits streaks:

| Situation | Rule |
|-----------|------|
| 3 good events in a row | Next draw is bad or mixed (a 3% chance skips this and allows any) |
| 2 bad events in a row | Next draw is good or mixed |
| Mixed event fires | Resets both streak counters |

So you won't see four droughts back to back, and a lucky streak always ends. After two rough events, the next one is guaranteed to be good or mixed.

Milestone chain boosts, festivals, epoch transition events and Endure's reconstruction penalty add their timed effects directly, without any of these checks. They show in the active events list and count down normally, but they don't use up the global timer or count toward streaks.

### Duration and Active Events

Events are either **instant** or **timed**:

- **Instant:** the effect is applied once, when the event fires. The log line is all you'll see.
- **Timed:** the event stays active for its duration and its per-tick effects apply the whole time. When it expires, the log prints an "ended" line that repeats what it cost you when it fired (workers who fled, resources stolen), in yellow.

Every duration and cooldown is set as a base length, which is exact in the Primitive and Stone Ages. From the Bronze Age on, a timed event lasts 2.6 times its base length (rounded to the nearest tick) and its cooldown is 2.6 times as long; instant amounts don't change. The tables below give each event's duration as it runs: events that can't fire before the Bronze Age always run the stretched length, and the three timed events that can fire earlier (Storm, Drought, Plague) show both.

### Seeing what happened

The main log reports each event as it fires, with its effect. When an event takes resources or workers, the next line says exactly what you lost (`You lost 10 food and 5 gold.` or `You lost 8 food and 3 workers.`), counting only what actually left your stores. The **Logs** panel (`logs`) keeps the history with tick numbers, so check it after time away.

A timed event's log line says how long it lasts in the age it fired in (Gold Rush reads `for ~30s` in the Stone Age and `for ~1m 18s` from the Bronze Age on), and the Stats panel counts it down.

An instant grant goes into your storage, and storage only takes what it has room for. When a full store cuts a grant short, the next line says what fit (`Storage was nearly full: only 25 food (of 250) fit.`). Milestone rewards and boon lumps do the same.

The **Stats panel** (`stats`) lists every active timed event under "Active Events" with the time left on it (e.g. `~2m 30s`) and its ongoing effect, color-coded: **green** for a bonus, **red** for a penalty. A production-boost event shows `all production +10%` in green. (On the colorblind-safe and high-contrast [themes](themes.md), bonuses show **blue** and penalties **orange**, with `▲`/`▼` marking the sign.) Instant effects (resource grants, thefts, worker loss) are not listed there, since they already happened.

Event log lines carry a little personality (wandering traders "smelling of cabbage and opportunity"), but the joke always sits alongside the mechanical summary, never in place of it.

---

## Base Event Reference

These 26 events belong to no epoch. They can fire in any epoch, throughout the whole game. Together with the 35 epoch-exclusive events (5 per epoch), that makes 61 random events.

### Good Events

| Name | Key | Min Age | Weight | Effect | Duration | Notes |
|------|-----|---------|--------|--------|----------|-------|
| Bountiful Harvest | `bountiful_harvest` | Primitive | 15 | +250 food | Instant | Most common good event early |
| Wandering Traders | `wandering_traders` | Bronze | 12 | +15 gold, +10 food | Instant | |
| Skilled Immigrants | `skilled_immigrants` | Stone | 10 | +10 knowledge | Instant | |
| Gold Rush | `gold_rush` | Bronze | 8 | +1.0 gold/tick | 39 ticks (~1m 18s) | |
| Trade Boom | `trade_boom` | Medieval | 8 | +2.0 gold/tick | 52 ticks (~1m 44s) | |
| Ancient Discovery | `ancient_discovery` | Iron | 6 | +50 knowledge | Instant | |
| Renaissance Fair | `renaissance_fair` | Renaissance | 10 | +0.5 culture/tick, +0.5 gold/tick | 39 ticks (~1m 18s) | |
| Colonial Windfall | `colonial_windfall` | Colonial | 8 | +100 gold, +30 culture | Instant | |
| Power Surge | `power_surge_base` | Victorian | 6 | +3.0 electricity/tick | 26 ticks (~52s) | |
| Crypto Boom | `crypto_boom` | Cyberpunk | 7 | +5.0 crypto/tick | 39 ticks (~1m 18s) | |
| First Contact | `first_contact` | Space | 3 | +500 knowledge, +50 titanium | Instant | Rarest good event |
| Dark Matter Rift | `dark_matter_rift` | Interstellar | 4 | +3.0 dark matter/tick | 39 ticks (~1m 18s) | |
| Quantum Fluctuation | `quantum_fluctuation` | Quantum | 3 | +5.0 quantum flux/tick | 26 ticks (~52s) | |

### Bad Events

| Name | Key | Min Age | Weight | Effect | Duration | Notes |
|------|-----|---------|--------|--------|----------|-------|
| Storm | `storm` | Primitive | 14 | Wood -0.3/tick | 5 ticks (~10s); 13 (~26s) from Bronze | Most common bad event |
| Drought | `drought` | Primitive | 12 | Food -0.5/tick | 10 ticks (~20s); 26 (~52s) from Bronze | |
| Bandit Raid | `bandit_raid` | Bronze | 10 | -10 food, -5 gold stolen | Instant | **Raid** (garrison blunts) |
| Plague | `plague` | Stone | 6 | Food -1.0/tick, -15% workers | 8 ticks (~16s); 21 (~42s) from Bronze | **Workers permanently lost** |
| Mine Collapse | `mine_collapse` | Iron | 7 | Iron -0.5/tick, -5% workers | 21 ticks (~42s) | **Workers permanently lost** |
| Heresy | `heresy` | Medieval | 5 | Faith -0.5/tick | 31 ticks (~1m 2s) | |
| Pirate Attack | `pirate_attack` | Colonial | 7 | -50 gold, -30 food stolen | Instant | **Raid** (garrison blunts) |
| Nuclear Scare | `nuclear_scare` | Atomic | 4 | Electricity -2.0/tick, knowledge -1.0/tick | 31 ticks (~1m 2s) | |
| Data Breach | `data_breach` | Information | 6 | -50 data, -100 gold stolen | Instant | **Raid** (garrison blunts) |
| Industrial Accident | `industrial_accident` | Industrial | 8 | -10 steel, -15 oil stolen, -7% workers | Instant | **Workers permanently lost** |
| Crypto Winter | `crypto_winter` | Cyberpunk | 8 | -4.5 crypto stolen | 36 ticks (~1m 12s) | The theft happens once, when it fires |

### Mixed Events

| Name | Key | Min Age | Weight | Effect | Duration | Notes |
|------|-----|---------|--------|--------|----------|-------|
| Earthquake | `earthquake` | Stone | 5 | -15 wood stolen, +20 stone | Instant | Trade-off |
| Plasma Storm | `plasma_storm` | Fusion | 5 | Electricity -5.0/tick, plasma +3.0/tick | 26 ticks (~52s) | Hurts power, helps plasma |

### Raids and your garrison

Some bad events are **raids**: attacks by outsiders. If you have soldiers, your garrison blunts part of a raid: it cuts the resources the raid steals and the workers it drives off by the share shown in the Army panel (`army`), at most 45%. The share depends on your army's Defense Rating against the raid threat of your current age; see [Military](military.md). With no soldiers a raid hits exactly as listed, but from the Iron Age on most players have some: the military buildings the age gates require train a garrison that blunts roughly 8-19% of a raid without any effort.

The raids are:

| Event | Key | Where | What the garrison blunts |
|-------|-----|-------|--------------------------|
| Bandit Raid | `bandit_raid` | base event, Bronze Age on | food and gold stolen |
| Pirate Attack | `pirate_attack` | base event, Colonial Age on | gold and food stolen |
| Data Breach | `data_breach` | base event, Information Age on | data and gold stolen |
| Tribal Raid | `tribal_raid` | Stone Era | food stolen and workers who flee (not the food production penalty) |
| Beast Stampede | `beast_stampede` | Stone Era | wood and food lost |
| The Great Breach | `epoch_data_breach` | Digital Era | data stolen (not the knowledge production penalty) |
| Corporate Espionage | `corporate_espionage` | Neon Era | gold and data stolen |

Soldiers only exist from the Iron Age, so the two Stone Era raids (and a Bandit Raid in the Bronze Age) still hit in full in a normal run. Only the stolen resources and lost workers are blunted; a raid's production penalty runs in full. Disasters and unrest are not raids, and soldiers do nothing against them: `plague`, `mine_collapse`, `industrial_accident`, `crypto_winter`, `earthquake` and the rest take their full toll.

When the garrison blunts a raid, the log adds a green line under the event, for example:

```
Your garrison blunted about 28% of the raid: you kept 3 food and 1 gold.
```

What it kept is added to the **Saved this run** line in the Army panel.

---

## Effect Types Explained

| Effect | What it does |
|--------|--------------|
| Resource grant | Adds a fixed amount of a resource once, when the event fires. |
| Production change | Adds a flat amount to one resource's per-tick rate (or subtracts it) for as long as the event lasts. It is not a percentage: +3.0 electricity/tick is +3.0 no matter how much you already make. |
| Theft | Removes a fixed amount of a resource once, when the event fires (never more than you have). For a timed event, the amount is repeated in the "ended" line. |
| Worker loss | Removes a percentage of your workers **permanently**, when the event fires, from every building in proportion and from the idle pool alike. They do not return when the event ends. |

Every random event also nudges **morale** when it fires: a good event lifts it by 4 points and a bad one lowers it by 4 (mixed events leave it alone). This applies to the base events above and to the epoch-exclusive events alike. Epoch transition events don't change morale. See [Morale](morale.md).

**Worker loss is the only permanent damage** in the random event system. Production changes and thefts are temporary or one-time, and morale drifts back toward 50% on its own.

---

## Epoch-Exclusive Random Events (Brief Reference)

Each epoch has 5 more events that enter the random event pool only while you're in that epoch. They follow the same weights, cooldowns, global timer and anti-streak rules as the base events; the only difference is that they are limited to their epoch.

| Epoch | Epoch Key |
|-------|-----------|
| Stone Era | `stone_era` |
| Iron Era | `iron_era` |
| Steel Era | `steel_era` |
| Electric Era | `electric_era` |
| Digital Era | `digital_era` |
| Neon Era | `neon_era` |
| Cosmic Era | `cosmic_era` |

For the full list with effects and durations, see [Epochs](epochs.md).

> **Epoch-exclusive random events are not epoch transition events.** Transition events fire once per epoch, at the boundary, and depend on your faith and culture. Faith and culture have no effect on whether base or epoch-exclusive random events fire; only timing, weights, cooldowns and the anti-streak rules apply.

---

## Age Awakenings

Awakenings are one-time boosts, one per epoch, that fire the first time you enter an epoch's signature age: seven epochs, seven awakenings. Unlike the epoch transition roll (which depends on faith and culture and can come up good or bad), an awakening always fires, always gives a modest production boost that fits its era, and never has a downside. The boost is **temporary**: it works like any timed event and ends after its listed duration. Each awakening fires at most once per prestige run. The record of which ones fired is saved with your game (a reload won't fire one again) and clears on prestige or reset, so the next run can earn them all again.

| Epoch | Trigger Age | Awakening | Temporary Effect | Duration |
|-------|-------------|-----------|------------------|----------|
| Stone Era ◈ | Stone Age | Pottery Mastery | +1.0 food/tick, +0.5 stone/tick | 250 ticks (~8m 20s) |
| Iron Era ⚔ | Iron Age | Discovery of Metallurgy | +2.0 iron/tick | 1,300 ticks (~43m 20s) |
| Steel Era ⚙ | Industrial Age | Steam Breakthrough | +25% to all production | 520 ticks (~17m 20s) |
| Electric Era ⚡ | Victorian Age | The Grid Wakes | +2.0 electricity/tick, +10% all production | 780 ticks (~26 min) |
| Digital Era ▣ | Modern Age | Networks Wake | +2.0 data/tick, +1.0 knowledge/tick | 780 ticks (~26 min) |
| Neon Era ◉ | Cyberpunk Age | Cybernetic Awakening | +20% to all production | 650 ticks (~21m 40s) |
| Cosmic Era ✦ | Interstellar Age | First Contact Signal | +1.5 dark matter/tick, +10% all production | 1,040 ticks (~34m 40s) |

Every awakening after the Stone Age's fires from the Bronze Age on, so its duration is already 2.6 times the base length, and its log line quotes the real length. The "+% all production" awakenings add to the all-production pool, which applies in full up to +200% and a quarter of every point past it (see [The all-production cap](resources.md#the-all-production-cap)). Awakenings appear in the active events list like any timed event and count down over their duration.

---

## Responding to Events

**Production penalties** (`drought`, `storm`, `heresy`, `nuclear_scare`): short, flat per-tick penalties you can't avoid, so ride them out. A drought takes 0.5 food a tick for 26 ticks from the Bronze Age on, 13 food in all. Don't restructure your workers over a penalty that lasts under a minute.

**Thefts** (`bandit_raid`, `pirate_attack`, `data_breach` and the thefts in other events): the resources are gone the moment the event fires. The amounts are fixed and never grow, while your income does, so they are small by the time each can fire: a Pirate Attack takes 50 gold and 30 food, and the largest theft in the game, Corporate Espionage in the Neon Era, takes 10K gold and 8K data. Normal reserves cover them; there is nothing to stockpile for. Many thefts are raids (see [Raids and your garrison](#raids-and-your-garrison)), so a garrison takes a share off them.

**Worker loss** (`plague`, `mine_collapse`, `industrial_accident`): the only lasting damage. Lost workers come out of every building in proportion, idle ones included, so keeping workers idle doesn't shield anything. With auto-recruit on (the default), the game refills the empty slots as housing and food allow; with it off, `recruit` and `assign` to replace them. See [Workers](workers-and-domains.md).

**Good events** need no preparation. Boosts like `gold_rush` and `crypto_boom` add a flat amount per tick that doesn't grow with your workers, so let them run. Knowledge windfalls (`skilled_immigrants`, `ancient_discovery`, `first_contact`) can cover the last of a tech's cost; research is paid in full when you start it, so check whether you can now afford the next one.

**Faith and culture affect epoch transition events, not random events.** Stacking faith doesn't make `bountiful_harvest` more likely or `drought` less likely. See [Epochs](epochs.md).
