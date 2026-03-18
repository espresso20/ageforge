# Random Events

Random events are the unpredictable heartbeat of AgeForge — fires, windfalls, plagues, gold rushes, and quantum fluctuations that fire throughout your run, independent of any strategic choice you make. They keep the pace dynamic and reward players who pay attention and react.

> **Two event systems exist.** This page covers the **base/global random event system** — events that fire during normal gameplay based on game tick and weighted probability. The **epoch-exclusive event system** (events tied to specific epochs, governed by faith and culture) is a separate mechanism documented in [Epochs](epochs.md).

---

## How Events Fire

### The Timing Window

The EventManager enforces a **global cooldown** between all random events. No two events can fire within 150–600 ticks of each other (roughly 5–20 minutes at 1x speed). When an event fires, the next window opens at a random point in that range. This prevents event spam and keeps each event feeling like a notable moment.

The first event of any new game is also delayed by the same 150–600 tick window, giving you time to get established before the chaos begins.

### Weighted Random Selection

Each event has a `Weight` value. Higher weight = more likely to appear. When the engine checks for a new event, it builds a pool of all eligible events and runs a weighted random draw — so a Weight 15 event is five times more likely to be drawn than a Weight 3 event from the same pool.

Events are only eligible if:
- The current tick is past the event's `MinTick` (prevents certain events from firing before they're contextually relevant)
- The current age is at or past the event's minimum required age
- The event's individual cooldown has elapsed since it last fired
- The event is not currently active (a timed event cannot overlap with itself)

### Anti-Streak System

The engine tracks consecutive good and bad events and enforces soft limits:

| Situation | Rule |
|-----------|------|
| 3 good events in a row | Next draw is forced bad (3% lucky reset chance to skip this) |
| 2 bad events in a row | Next draw is forced good |
| Mixed event fires | Resets both streak counters |

This is why you won't see four droughts back to back, and also why a lucky streak always ends. The system intentionally prevents both brutal punishment spirals and lucky-streak sailing. After two rough events, a good one is mechanically guaranteed.

### Duration and Active Events

Events are either **instant** or **timed**:

- **Instant (Duration = 0):** Effect is applied once and done. No active event entry. The log message is all you'll see.
- **Timed (Duration > 0):** Effect persists for the listed number of ticks, tracked as an active event. When it expires, the engine fires an "ended" log message summarising any losses that accumulated during it (workers who fled, resources stolen tick by tick).

At 1x speed (default 2 seconds per tick):

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

Use the `logs` command (or press `L`) to open the Logs overlay and see your full event history. Events appear as log entries with their effect description and duration. When a timed event ends, the "ended" entry shows accumulated losses in yellow — useful for assessing actual damage.

---

## Base Event Reference

These 27 events have no `EpochKey` — they can fire in any epoch, throughout the entire game. They form the permanent background of random occurrences regardless of where you are in the run.

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
| Power Surge | `power_surge` | Victorian | 6 | +3.0 electricity/tick | 10 ticks | |
| Crypto Boom | `crypto_boom` | Cyberpunk | 7 | +5.0 crypto/tick | 15 ticks | |
| First Contact | `first_contact` | Space | 3 | +500 knowledge, +50 titanium | Instant | Rarest good event |
| Dark Matter Rift | `dark_matter_rift` | Interstellar | 4 | +3.0 dark matter/tick | 15 ticks | |
| Quantum Fluctuation | `quantum_fluctuation` | Quantum | 3 | +5.0 quantum flux/tick | 10 ticks | |

### Bad Events

| Name | Key | Min Age | Weight | Effect | Duration | Notes |
|------|-----|---------|--------|--------|----------|-------|
| Storm | `storm` | Primitive | 14 | Wood production -0.3/tick | 5 ticks | Most common bad event |
| Drought | `drought` | Primitive | 12 | Food production -0.5/tick | 10 ticks | |
| Bandit Raid | `bandit_raid` | Bronze | 10 | -10 food, -5 gold stolen | Instant | |
| Plague | `plague` | Stone | 6 | Food production -1.0/tick, -15% workers | 8 ticks | **Workers permanently lost** |
| Mine Collapse | `mine_collapse` | Iron | 7 | Iron production -0.5/tick, coal -0.3/tick, -5% workers | 8 ticks | **Workers permanently lost** |
| Heresy | `heresy` | Medieval | 5 | Faith production -0.5/tick | 12 ticks | |
| Pirate Attack | `pirate_attack` | Colonial | 7 | -50 gold, -30 food stolen | Instant | |
| Nuclear Scare | `nuclear_scare` | Atomic | 4 | Electricity -2.0/tick, knowledge -1.0/tick | 12 ticks | |
| Data Breach | `data_breach` | Information | 6 | -50 data, -100 gold stolen | Instant | |
| Industrial Accident | `industrial_accident` | Industrial | 8 | -10 steel, -15 oil stolen, -7% workers | Instant | **Workers permanently lost** |
| Crypto Winter | `crypto_winter` | Cyberpunk | 8 | -4.5 crypto stolen/tick | 14 ticks | Ongoing drain, not one-shot |

### Mixed Events

| Name | Key | Min Age | Weight | Effect | Duration | Notes |
|------|-----|---------|--------|--------|----------|-------|
| Earthquake | `earthquake` | Stone | 5 | -15 wood stolen, +20 stone | Instant | Trade-off |
| Plasma Storm | `plasma_storm` | Fusion | 5 | Electricity -5.0/tick, plasma +3.0/tick | 10 ticks | Hurts power, helps plasma |

---

## Effect Types Explained

| Effect Type | What It Does |
|-------------|-------------|
| `instant_resource` | One-time addition to a resource. Applied once; no active event tracked. |
| `production` | Multiplier bonus or penalty to a specific resource's production rate. Persists for the event duration. Positive = boost, negative = penalty. |
| `steal_resource` | Removes a fixed amount from a resource. For instant events, applied once. For timed events, can drain per-tick — check the expired log for total losses. |
| `worker_loss` | Removes a percentage of your total worker pool **permanently**. Workers lost this way do not return when the event ends. |
| `morale` | Instant one-time adjustment to morale. Good epoch events grant **+0.04 morale**; bad epoch events apply **−0.04 morale**. These reflect the mood and motivation of the population — a windfall lifts spirits, while disasters shake confidence. |

**Worker loss is the only permanent structural effect** in the random event system. Morale changes from events are real but recoverable — food surplus, passive recovery, and age advances will restore morale over time. All production modifiers and resource steals are temporary or one-time. If an event has `worker_loss`, treat it as a permanent cost, not a debuff.

> **Note:** The morale effect applies to **epoch transition events** (the faith/culture roll system documented in [Epochs](epochs.md)), not to the base or epoch-exclusive random events listed in the tables above. A good epoch transition event lifts morale by +0.04; a bad one shaves −0.04. This is separate from catastrophe morale penalties (Endure −0.10, Succumb resets to 0.50).

---

## Epoch-Exclusive Random Events (Brief Reference)

Each epoch has 5 additional events that enter the random event pool only while you're in that epoch. They fire through the same weighted-random, cooldown-respecting, anti-streak system as base events — the only difference is eligibility is restricted to the matching epoch.

| Epoch | Epoch Key |
|-------|-----------|
| Stone Era | `stone_era` |
| Iron Era | `iron_era` |
| Steel Era | `steel_era` |
| Electric Era | `electric_era` |
| Digital Era | `digital_era` |
| Neon Era | `neon_era` |
| Cosmic Era | `cosmic_era` |

For the full list of epoch-exclusive events including effects, durations, and strategy notes, see [Epochs — Epoch-Exclusive Random Events](epochs.md#epoch-exclusive-random-events).

> **Important distinction:** Epoch-exclusive random events are NOT the same as epoch transition events. Epoch transition events (the faith/culture roll) fire once per epoch at the boundary and are governed entirely by faith and culture. Epoch-exclusive random events fire during normal gameplay within the epoch, following the same rules as base events. Faith and culture have no effect on whether base or epoch-exclusive random events fire — only timing, cooldowns, and the anti-streak system apply.

---

## Managing Events

### Viewing Your Event Log

```
logs
```

Opens the Logs overlay. All event messages appear here with timestamps. Check it regularly — events fire while you're doing other things, and the log is the only record of what happened and what was lost.

When a timed event ends, the expiry log entry shows accumulated losses in yellow (e.g. `3 workers fled`, `45 food stolen`). This tells you the actual cost of the event, not just the stated effect magnitude.

### Responding to Bad Events

**Production penalty events** (`drought`, `storm`, `heresy`, `nuclear_scare`):
These are temporary production debuffs. Timed events you cannot avoid — ride them out. If food production goes negative during a drought, make sure you have food reserves banked before the event hits. Check `logs` when you see the event fire so you know how many ticks remain.

**Resource steal events** (`bandit_raid`, `pirate_attack`, `data_breach`):
Instant events — the resources are gone. Nothing to do post-fire. Defensively: keep deep reserves of gold and food (the most commonly targeted resources). Storage buildings are underrated insurance against these.

**Worker loss events** (`plague`, `mine_collapse`, `industrial_accident`):
The most dangerous category. Workers lost are gone permanently. Strategies:
- Maintain a buffer of idle (unassigned) workers at all times. Don't recruit-to-max and assign everything.
- After a worker_loss event, check your domain assignments — capacity caps may have changed, and some buildings might now be over-assigned.
- Use `recruit` to replace lost workers as soon as your food reserves allow.

**Crypto Winter** (`crypto_winter`):
A drain event that removes crypto per tick over 14 ticks rather than all at once. If you see it fire and you have substantial crypto reserves, you can't stop the drain — but the total loss is bounded by the event duration.

### Maximising Good Events

**Production boost windows:**
When `gold_rush`, `trade_boom`, `power_surge`, `crypto_boom`, or similar boost events fire, this is the time to recruit and assign workers to that resource domain. Boosted production multiplied by more workers compounds significantly. Check your rates during the window.

**Knowledge windfalls:**
`skilled_immigrants`, `ancient_discovery`, and `first_contact` give instant knowledge. If you're close to finishing a research, queue the most expensive tech you can afford — the event can tip you over the threshold.

**Renaissance Fair:**
Culture and gold both boost simultaneously for 15 ticks. If you're building toward a culture-gated milestone or epoch transition, this window is worth pushing hard into culture production infrastructure.

---

## Anti-Streak and Cooldown Details

For players who want to understand the mechanics fully:

**Global cooldown:** After any event fires, the next event cannot fire for 150–600 ticks (random in that range). At 2 seconds per tick this is 5–20 minutes of real time. The first event of a new game has the same delay.

**Per-event cooldown:** Each event definition has its own `Cooldown` field. Even after the global cooldown expires, a specific event cannot reappear until its individual cooldown has elapsed since it last fired. For example, `plague` has a 200-tick cooldown — even if the global cooldown expired, plague cannot fire again for 200 ticks after its last occurrence.

**Anti-streak rule:**
- Fires after ≥ 3 good events: next draw is forced to `bad` or `mixed` only (3% chance to skip and allow any)
- Fires after ≥ 2 bad events: next draw is forced to `good` or `mixed` only

Mixed events (like `earthquake` or `plasma_storm`) reset both streak counters, which is why they can sometimes break a lucky or unlucky run's rhythm.

**Cannot overlap:** A timed event cannot fire a second instance of itself while already active. If `drought` is active, another drought cannot trigger until the current one expires and the individual cooldown elapses.

**InjectEvent:** Some systems bypass all of these rules. Milestone chain boosts and certain epoch event side-effects use `InjectEvent`, which adds an active event directly with no cooldown or eligibility checks. These injected events still appear in the logs and tick down normally, but they do not consume the global cooldown window or affect streak counters.

---

## Tips

- **Check `logs` regularly.** Events fire in the background. A plague or mine collapse you didn't notice could have already taken workers. The expiry log shows actual losses.

- **Bad timed events are temporary (except worker loss).** A 50% food production penalty for 10 ticks sounds alarming, but at 2 seconds per tick that's 20 seconds. Don't make permanent decisions (like restructuring worker assignments) because of a short-duration debuff.

- **Bank resources before late-game.** Pirate attacks, data breaches, and corporate espionage (epoch event) scale in magnitude as you progress. A surplus buffer absorbs the hit. Early-game resource steal amounts are small; by the Neon Era they're 10,000 gold at a time.

- **Maintain idle workers at all times.** Worker loss from `plague`, `mine_collapse`, or `industrial_accident` is the only permanent damage in this system. Running every worker assigned with zero idle is the riskiest configuration. Even 5–10 unassigned workers gives you breathing room.

- **Good events don't require preparation — but reward it.** A gold rush fires whether you're ready or not. If you have workers ready to assign to gold buildings, you can capitalise on the boost window. Players who react fast earn more from good events than those who don't notice until the window closes.

- **Faith and culture affect EPOCH events, not base random events.** Stacking faith doesn't make `bountiful_harvest` more likely or `drought` less likely. The base event system is purely tick-based, weighted-random, with cooldowns. Faith and culture are levers for the epoch transition roll only. See [Epochs](epochs.md) for that system.
