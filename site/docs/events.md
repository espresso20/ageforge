# Random Events

Random events are the fires, windfalls, plagues, gold rushes and quantum fluctuations that happen throughout your run, independent of the choices you make. They keep the pace changing and reward players who pay attention and react.

> **Two kinds of events come from the epoch system, and this page covers neither in depth.** This page covers the **base random events**, which fire during normal play based on game tick and weighted chance. Each epoch also adds its own **epoch-exclusive random events** to the same pool, and every epoch transition fires one **epoch transition event** shaped by your faith and culture. Both are documented in [Epochs](epochs.md).

---

## How Events Fire

### The Timing Window

A **global timer** spaces out all random events. After an event fires, the next one can't fire for 150 to 600 ticks (roughly 5 to 20 minutes at 1x speed), picked at random in that range. This prevents event spam and keeps each event feeling like a notable moment.

The first event of any new game waits for the same 150 to 600 tick window, giving you time to get established first.

### Weighted Random Selection

Each event has a weight. The higher the weight, the more likely it is. When the game checks for a new event, it builds a pool of all eligible events and draws one by weight, so a weight 15 event is five times more likely to be drawn than a weight 3 event from the same pool.

An event is eligible only if:
- the run has reached the event's earliest tick (some events can't fire before they make sense),
- the current age is at or past the event's minimum age,
- the event's own cooldown has passed since it last fired,
- the event is not currently active (a timed event cannot overlap with itself).

### Anti-Streak System

The game tracks consecutive good and bad events and limits streaks:

| Situation | Rule |
|-----------|------|
| 3 good events in a row | Next draw is bad or mixed (a 3% chance skips this and allows any) |
| 2 bad events in a row | Next draw is good or mixed |
| Mixed event fires | Resets both streak counters |

This is why you won't see four droughts back to back, and why a lucky streak always ends. After two rough events, the next one is guaranteed to be good or mixed.

### Duration and Active Events

Events are either **instant** or **timed**:

- **Instant (duration 0):** The effect is applied once. No active event entry; the log message is all you'll see.
- **Timed (duration above 0):** The event stays active for the listed number of ticks. Its per-tick effects apply the whole time. When it expires, the log prints an "ended" message that sums up what it cost you when it fired (workers who fled, resources stolen).

At 1x speed (2 seconds per tick):

| Duration | Real-time equivalent |
|----------|---------------------|
| 5 ticks | ~10 seconds |
| 8 ticks | ~16 seconds |
| 10 ticks | ~20 seconds |
| 12 ticks | ~24 seconds |
| 14 ticks | ~28 seconds |
| 15 ticks | ~30 seconds |
| 20 ticks | ~40 seconds |

### Viewing Events

Use the `logs` command to open the Logs panel and see your full event history. Events appear as log entries with their effect and duration. When a timed event ends, the "ended" entry shows its losses in yellow, which is useful for judging the actual damage.

> **A note on tone.** Event log lines are written with a little personality: a wandering merchant "smelling of cabbage and opportunity," a crypto boom turning your least competent citizen into a thought leader. The flavor is cosmetic and always sits *alongside* the mechanical summary (the resources gained or lost, and the duration), never in place of it, so you can read the joke and still know exactly what happened to your economy. Milestone completions and other notable log moments (a building finishing, an age turning over, a research breakthrough, a famine starting or ending, enduring a catastrophe) carry the same light touch.

The **Stats panel** lists every currently active timed event under "Active Events" with the approximate wall-clock time left on it (e.g. `~2m 30s`; see [Timers and durations](commands.md#timers-and-durations)). Under each event it shows the ongoing per-tick or percentage effect, color-coded: **green** for a bonus, **red** for a penalty. A Famine shows `food -3.0/t` in red; a production-boost event shows `all production +10%` in green. (On the colorblind-safe and high-contrast [themes](commands.md#themes), bonuses show **blue** and penalties **orange**, with `▲`/`▼` glyphs marking the sign.) Instant effects (resource grants, theft, worker loss) are not listed there, since they already happened when the event fired.

---

## Base Event Reference

These 26 events belong to no epoch. They can fire in any epoch, throughout the whole game, and form the permanent background of random events. Together with the 35 epoch-exclusive events (5 per epoch), that makes 61 random events.

### Good Events

| Name | Key | Min Age | Weight | Effect | Duration | Notes |
|------|-----|---------|--------|--------|----------|-------|
| Bountiful Harvest | `bountiful_harvest` | Primitive | 15 | +250 food | Instant | Most common good event early |
| Wandering Traders | `wandering_traders` | Bronze | 12 | +15 gold, +10 food | Instant | |
| Skilled Immigrants | `skilled_immigrants` | Stone | 10 | +10 knowledge | Instant | |
| Gold Rush | `gold_rush` | Bronze | 8 | +1.0 gold/tick | 15 ticks | |
| Trade Boom | `trade_boom` | Medieval | 8 | +2.0 gold/tick | 20 ticks | |
| Ancient Discovery | `ancient_discovery` | Iron | 6 | +50 knowledge | Instant | |
| Renaissance Fair | `renaissance_fair` | Renaissance | 10 | +0.5 culture/tick, +0.5 gold/tick | 15 ticks | |
| Colonial Windfall | `colonial_windfall` | Colonial | 8 | +100 gold, +30 culture | Instant | |
| Power Surge | `power_surge_base` | Victorian | 6 | +3.0 electricity/tick | 10 ticks | |
| Crypto Boom | `crypto_boom` | Cyberpunk | 7 | +5.0 crypto/tick | 15 ticks | |
| First Contact | `first_contact` | Space | 3 | +500 knowledge, +50 titanium | Instant | Rarest good event |
| Dark Matter Rift | `dark_matter_rift` | Interstellar | 4 | +3.0 dark matter/tick | 15 ticks | |
| Quantum Fluctuation | `quantum_fluctuation` | Quantum | 3 | +5.0 quantum flux/tick | 10 ticks | |

### Bad Events

| Name | Key | Min Age | Weight | Effect | Duration | Notes |
|------|-----|---------|--------|--------|----------|-------|
| Storm | `storm` | Primitive | 14 | Wood -0.3/tick | 5 ticks | Most common bad event |
| Drought | `drought` | Primitive | 12 | Food -0.5/tick | 10 ticks | |
| Bandit Raid | `bandit_raid` | Bronze | 10 | -10 food, -5 gold stolen | Instant | |
| Plague | `plague` | Stone | 6 | Food -1.0/tick, -15% workers | 8 ticks | **Workers permanently lost** |
| Mine Collapse | `mine_collapse` | Iron | 7 | Iron -0.5/tick, coal -0.3/tick, -5% workers | 8 ticks | **Workers permanently lost** |
| Heresy | `heresy` | Medieval | 5 | Faith -0.5/tick | 12 ticks | |
| Pirate Attack | `pirate_attack` | Colonial | 7 | -50 gold, -30 food stolen | Instant | |
| Nuclear Scare | `nuclear_scare` | Atomic | 4 | Electricity -2.0/tick, knowledge -1.0/tick | 12 ticks | |
| Data Breach | `data_breach` | Information | 6 | -50 data, -100 gold stolen | Instant | |
| Industrial Accident | `industrial_accident` | Industrial | 8 | -10 steel, -15 oil stolen, -7% workers | Instant | **Workers permanently lost** |
| Crypto Winter | `crypto_winter` | Cyberpunk | 8 | -4.5 crypto stolen | 14 ticks | The theft happens once, when it fires |

### Mixed Events

| Name | Key | Min Age | Weight | Effect | Duration | Notes |
|------|-----|---------|--------|--------|----------|-------|
| Earthquake | `earthquake` | Stone | 5 | -15 wood stolen, +20 stone | Instant | Trade-off |
| Plasma Storm | `plasma_storm` | Fusion | 5 | Electricity -5.0/tick, plasma +3.0/tick | 10 ticks | Hurts power, helps plasma |

---

## Effect Types Explained

| Effect Type | What It Does |
|-------------|-------------|
| `instant_resource` | Adds a fixed amount of a resource once, when the event fires. |
| `production` | Adds a flat amount to one resource's per-tick rate (or subtracts it, if negative) for as long as the event lasts. It is not a percentage: +3.0 electricity/tick is +3.0 no matter how much you already make. |
| `steal_resource` | Removes a fixed amount of a resource once, when the event fires (never more than you have). For a timed event, the amount is repeated in the "ended" message. |
| `worker_loss` | Removes a percentage of your total workers **permanently**, when the event fires. They do not return when the event ends. |

Every random event also nudges **morale** when it fires: a good event lifts it by 4 points and a bad one lowers it by 4 (mixed events leave it alone). This applies to the base events above and to the epoch-exclusive events alike. Epoch transition events don't change morale. Catastrophes are separate: enduring one costs 10 points of morale, and a Succumb reset returns morale to the 50% baseline. See [Morale](morale.md) for how morale affects production.

**Worker loss is the only permanent damage** in the random event system. Morale drifts back toward 50% over time, and worship and culture buildings, good events and age advances raise it. All production changes and resource thefts are temporary or one-time. If an event has worker loss, treat it as a permanent cost, not a debuff.

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

For the full list of epoch-exclusive events with effects, durations and strategy notes, see [Epochs: Epoch-Exclusive Random Events](epochs.md#epoch-exclusive-random-events).

> **Epoch-exclusive random events are not epoch transition events.** Transition events fire once per epoch, at the boundary, and depend on your faith and culture. Epoch-exclusive random events fire during normal play within the epoch and follow the same rules as the base events. Faith and culture have no effect on whether base or epoch-exclusive random events fire; only timing, cooldowns and the anti-streak system apply.

---

## Age Awakenings

Awakenings are one-time boosts, one per epoch, that fire the first time you enter an epoch's signature age: seven epochs, seven awakenings. Unlike the epoch transition roll (which depends on faith and culture and can come up good or bad), an awakening always fires, always gives a modest production boost that fits its era, and never has a downside. The boost is **temporary**: it works like any timed event and ends after its listed duration. Each awakening fires at most once per prestige run. The record of which ones fired is saved with your game (a reload won't fire one again) and clears on prestige or reset, so the next run can earn them all again.

| Epoch | Trigger Age | Awakening | Temporary Effect | Duration |
|-------|-------------|-----------|------------------|----------|
| Stone Era ◈ | Stone Age | Pottery Mastery | +1.0 food/tick, +0.5 stone/tick | 250 ticks (~8 min) |
| Iron Era ⚔ | Iron Age | Discovery of Metallurgy | +2.0 iron/tick | 500 ticks (~16 min) |
| Steel Era ⚙ | Industrial Age | Steam Breakthrough | +25% to all production | 200 ticks (~6.5 min) |
| Electric Era ⚡ | Victorian Age | Electrification | +2.0 electricity/tick, +10% all production | 300 ticks (~10 min) |
| Digital Era ▣ | Modern Age | Information Age Dawns | +2.0 data/tick, +1.0 knowledge/tick | 300 ticks (~10 min) |
| Neon Era ◉ | Cyberpunk Age | Cybernetic Awakening | +20% to all production | 250 ticks (~8 min) |
| Cosmic Era ✦ | Interstellar Age | First Contact Signal | +1.5 dark matter/tick, +10% all production | 400 ticks (~13 min) |

Awakenings appear in the active events list like any timed event and count down over their duration, so you can see how long the boost has left. Each one also prints a short line of flavor text in the log when it fires.

---

## Managing Events

### Viewing Your Event Log

```
logs
```

Opens the Logs panel. All event messages appear here with timestamps. Check it regularly: events fire while you're doing other things, and the log is the only record of what happened and what was lost.

When a timed event ends, its "ended" entry shows the losses in yellow (e.g. `3 workers fled`, `45 food stolen`).

### Responding to Bad Events

**Production penalty events** (`drought`, `storm`, `heresy`, `nuclear_scare`): these are short, flat per-tick penalties that you can't avoid, so ride them out. If food goes negative during a drought, you want food banked before the event hits. The Stats panel shows how long an active event has left.

**Resource theft events** (`bandit_raid`, `pirate_attack`, `data_breach`): the resources are gone the moment the event fires, and there is nothing to do afterwards. Keep reserves of gold and food, the resources taken most often. Storage buildings are underrated insurance against these.

**Worker loss events** (`plague`, `mine_collapse`, `industrial_accident`): the most dangerous kind, because lost workers are gone for good.
- Keep a few idle (unassigned) workers at all times rather than recruiting to the limit and assigning everyone.
- After a worker loss, check your assignments: some buildings may now be understaffed.
- Use `recruit` to replace lost workers as soon as your food allows.

**Crypto Winter** (`crypto_winter`): takes 4.5 crypto when it fires. The event then stays listed for 14 ticks.

### Making the Most of Good Events

**Production boost windows:** `gold_rush`, `trade_boom`, `power_surge_base`, `crypto_boom` and the other boost events add a flat amount per tick. The bonus doesn't grow with your workers, so there's no need to reshuffle assignments to catch it; just let it run.

**Knowledge windfalls:** `skilled_immigrants`, `ancient_discovery` and `first_contact` give instant knowledge. If you're saving up for a tech, a windfall can cover the last of the cost. Research is paid in full when you start it, so check whether you can now afford the next one.

**Renaissance Fair:** adds culture and gold per tick for 15 ticks. It is small, so treat it as a bonus rather than something to plan around.

---

## Anti-Streak and Cooldown Details

For players who want the full mechanics:

**Global timer:** After any event fires, the next event can't fire for 150 to 600 ticks (random in that range). At 2 seconds per tick that is 5 to 20 minutes of real time. The first event of a new game has the same delay.

**Per-event cooldown:** Each event also has its own cooldown. Even after the global timer runs out, a specific event cannot come back until its own cooldown has passed since it last fired. For example, `plague` has a 200-tick cooldown, so it cannot fire again for 200 ticks after its last occurrence.

**Anti-streak rule:**
- After 3 or more good events in a row, the next draw is limited to bad or mixed events (a 3% chance skips this and allows any).
- After 2 or more bad events in a row, the next draw is limited to good or mixed events.

Mixed events (like `earthquake` or `plasma_storm`) reset both streak counters, which is why they can break a lucky or unlucky run's rhythm.

**No overlap:** A timed event cannot fire again while it is still active. If `drought` is active, another drought cannot start until the current one expires and its cooldown has passed.

**Effects added directly:** Milestone chain boosts, festivals, epoch transition events and Endure's reconstruction debuff add their timed effects directly, without any of these checks. They show in the active events list and count down normally, but they don't use up the global timer or count toward streaks.

---

## Tips

- **Check `logs` regularly.** Events fire in the background. A plague or mine collapse you didn't notice may already have taken workers, and the "ended" entry shows the actual losses.

- **Bad timed events are temporary (except worker loss).** A drought takes 0.5 food/tick for 10 ticks, which is 20 seconds at 1x. Don't make permanent decisions (like restructuring worker assignments) because of a short debuff.

- **Bank resources before the late game.** Pirate attacks, data breaches and corporate espionage (an epoch-exclusive event) grow as you progress. A surplus absorbs the hit. Early thefts are small; by the Neon Era, Corporate Espionage takes 10K gold at once.

- **Keep idle workers at all times.** Worker loss from `plague`, `mine_collapse` or `industrial_accident` is the only permanent damage in this system. Assigning every worker with none idle is the riskiest setup. Even 5 to 10 unassigned workers give you room.

- **Good events need no preparation.** A gold rush fires whether you're ready or not, and its flat bonus is the same either way.

- **Faith and culture affect epoch transition events, not random events.** Stacking faith doesn't make `bountiful_harvest` more likely or `drought` less likely. The random event system runs on ticks, weights and cooldowns. Faith and culture only matter for the epoch transition roll. See [Epochs](epochs.md).
