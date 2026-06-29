# Technologies

Research is your civilization's most powerful long-term lever. 52 technologies span all 22 ages — each one permanently alters your production rates, military strength, storage caps, or the pace of the game itself. Research is **sequential**: only one technology can be in progress at a time, and it must run to completion (or be deliberately cancelled) before you can start the next.

---

## How Research Works

### Starting Research

Research costs **knowledge points (kp)**, deducted immediately when you start. There is no refund if you cancel — the knowledge is gone the moment you type the command.

Once started, the tech counts down in **ticks**. Each game tick decrements the counter by 1. When it hits zero, the effects are applied instantly and permanently.

**Formula for adjusted tick count:**
```
adjusted_ticks = max(1,  base_ticks × (1.0 − research_speed_bonus))
```

A `research_speed` bonus of `0.30` (30%) cuts tick count to 70% of base — not time, but ticks. Since ticks can also run faster via `tick_speed` bonuses, both compound. A civilization with high research speed **and** high tick speed researches dramatically faster.

The adjusted ticks are locked in at the moment you start the tech. Gaining more `research_speed` mid-research does not retroactively shorten the current countdown.

### Knowledge Cost is Upfront

Knowledge is removed from your stockpile when you issue the `research` command — before any ticks pass. If you don't have enough, the command fails. If your knowledge income drops to zero during a long research countdown, **research still completes** — the ticks count down regardless of your current knowledge income. The cost was already paid.

### Only One Slot

There is no queue. If you try to start a second tech while one is in progress, you get an error showing the active tech and how many ticks remain. Plan your research order in advance.

### When Research Completes

Effects are applied the tick the counter hits zero. You'll see a success message in the log. The tech is now marked researched and its bonuses feed into `recalculateRates` on the next tick.

---

## Research Speed Sources

`research_speed` reduces tick count at research start. All three sources add together before being applied:

| Source | How much | Notes |
|---|---|---|
| **Tech bonuses** | Varies — see tech list | Accumulate permanently as techs complete |
| **Ancient Knowledge** (Succumb) | +0.25 (25%) | Granted permanently when you Succumb to an epoch catastrophe; survives all future resets |
| **Prestige: Research Speed** | +0.05 per tier, max 5 tiers (+25%) | Boosts `knowledge_rate`, not `research_speed` directly — see note below |

> **Note on Prestige "Research Speed":** Despite its name, the prestige upgrade boosts `knowledge_rate` (how fast you generate knowledge), not the `research_speed` tick-reduction multiplier. More knowledge income means you can afford more techs faster, but it doesn't reduce tick counts. The two mechanics are complementary, not the same thing.

No tech in the current tree directly grants `research_speed` as a bonus target — the tick-reduction multiplier comes primarily from Succumb's Ancient Knowledge. Knowledge income, however, is boosted by many techs (see §5 below).

---

## Commands

```
research <tech_key>
```
Start researching a technology. Deducts knowledge cost immediately. Fails if: unknown key, already researched, another tech in progress, age requirement not met, prerequisites missing, or insufficient knowledge.

You can also enter multi-word tech names with spaces — they are automatically joined with underscores:
```
research bronze working
```
is equivalent to `research bronze_working`.

---

```
research list
```
Lists all technologies available in the current age with their status (researched, in progress, available, locked by prerequisite). Also shows the currently active research and ticks remaining.

---

```
research cancel
```
Cancels the current research. **No refund.** The knowledge cost is lost. Only use this when pivoting is worth more than the sunk cost.

---

```
research
```
With no arguments, opens the **Research overlay panel** (same as the Research overlay panel). The overlay groups techs by age with visual indicators: researched techs show as complete, available ones are highlighted, and locked ones are dimmed.

---

**Shortcut:** `res` is an alias for `research`. All subcommands work identically.

---

## Tech Tree by Age

Prerequisites are listed using tech keys. "—" means no prerequisite.

### Primitive Age (~1 min/tech at 1× speed)

| Key | Name | Cost | Ticks | Prerequisites | Effect |
|---|---|---|---|---|---|
| `tool_making` | Tool Making | 800 kp | 200 | — | +15% gather rate |
| `fire_mastery` | Fire Mastery | 1,000 kp | 200 | `tool_making` | +0.1 food/tick |

---

### Stone Age (~2 min/tech)

| Key | Name | Cost | Ticks | Prerequisites | Effect |
|---|---|---|---|---|---|
| `stoneworking` | Stoneworking | 6,000 kp | 500 | `tool_making` | +20% stone rate |
| `animal_husbandry` | Animal Husbandry | 7,500 kp | 550 | `fire_mastery` | +0.2 food/tick |
| `pottery` | Pottery | 5,000 kp | 450 | `fire_mastery` | +25 all storage |
| `primitive_writing` | Primitive Writing | 10,000 kp | 600 | `pottery` | +10% knowledge rate |

---

### Bronze Age (~3 min/tech)

| Key | Name | Cost | Ticks | Prerequisites | Effect |
|---|---|---|---|---|---|
| `bronze_working` | Bronze Working | 15,000 kp | 750 | `stoneworking` | +20% iron rate, +10% gather rate |
| `agriculture` | Agriculture | 12,000 kp | 700 | `animal_husbandry` | +0.5 food/tick |
| `currency` | Currency | 17,500 kp | 800 | `primitive_writing` | +30% gold rate |
| `masonry` | Masonry | 13,000 kp | 700 | `stoneworking` | +50 all storage |
| `military_tactics` | Military Tactics | 20,000 kp | 900 | `bronze_working` | +20% military power |

---

### Iron Age (~4 min/tech)

| Key | Name | Cost | Ticks | Prerequisites | Effect |
|---|---|---|---|---|---|
| `iron_smelting` | Iron Smelting | 30,000 kp | 1,100 | `bronze_working` | +40% iron rate, +0.2 iron/tick |
| `road_building` | Road Building | 25,000 kp | 950 | `masonry` | +20% gold rate, +10% gather rate |
| `mathematics` | Mathematics | 37,500 kp | 1,200 | `primitive_writing`, `currency` | +20% knowledge rate |
| `siege_warfare` | Siege Warfare | 35,000 kp | 1,005 | `military_tactics` | +30% military power |

---

### Classical Age (~5 min/tech)

| Key | Name | Cost | Ticks | Prerequisites | Effect |
|---|---|---|---|---|---|
| `philosophy` | Philosophy | 20,000 kp | 1,500 | `mathematics`, `primitive_writing` | +30% knowledge rate, +0.2 culture/tick |
| `civil_engineering` | Civil Engineering | 18,000 kp | 1,300 | `masonry`, `road_building` | +100 all storage, −5% build cost |
| `imperial_legions` | Imperial Legions | 22,000 kp | 1,600 | `siege_warfare`, `iron_smelting` | +40% military power |

---

### Medieval Age (~7 min/tech)

| Key | Name | Cost | Ticks | Prerequisites | Effect |
|---|---|---|---|---|---|
| `steel_forging` | Steel Forging | 25,000 kp | 2,000 | `iron_smelting` | +0.1 steel/tick, +30% iron rate |
| `theology` | Theology | 20,000 kp | 1,800 | `philosophy` | +0.3 faith/tick |
| `banking` | Banking | 30,000 kp | 2,100 | `currency`, `mathematics` | +50% gold rate, +100 gold storage |
| `feudalism` | Feudalism | 22,000 kp | 1,700 | `military_tactics` | +5 population capacity |
| `alchemy` | Alchemy | 28,000 kp | 2,200 | `mathematics` | +15% knowledge rate, +0.1 gold/tick |
| `chronometry` | Chronometry | 20,000 kp | 1,900 | — | +5% tick speed |

---

### Renaissance Age (~10 min/tech)

| Key | Name | Cost | Ticks | Prerequisites | Effect |
|---|---|---|---|---|---|
| `printing_press` | Printing Press | 50,000 kp | 3,000 | `theology`, `alchemy` | +40% knowledge rate, +0.3 culture/tick |
| `navigation` | Navigation | 45,000 kp | 2,600 | `mathematics`, `road_building` | +50% gold rate, +30% expedition reward |
| `gunpowder` | Gunpowder | 55,000 kp | 3,200 | `alchemy`, `siege_warfare` | +50% military power |
| `patronage` | Patronage | 40,000 kp | 2,500 | `banking` | +0.5 culture/tick, +0.12 knowledge/tick |

---

### Colonial Age (~14 min/tech)

| Key | Name | Cost | Ticks | Prerequisites | Effect |
|---|---|---|---|---|---|
| `cartography` | Cartography | 80,000 kp | 4,000 | `navigation` | +50% expedition reward, +50% gold rate |
| `mercantilism` | Mercantilism | 75,000 kp | 3,800 | `banking`, `navigation` | +2.0 gold/tick, +30% gold rate |
| `colonialism` | Colonialism | 90,000 kp | 4,400 | `cartography`, `gunpowder` | +2.0 food/tick, +30% military power |

---

### Industrial Age (~18 min/tech)

| Key | Name | Cost | Ticks | Prerequisites | Effect |
|---|---|---|---|---|---|
| `steam_power` | Steam Power | 100,000 kp | 5,200 | `steel_forging` | +30% all production |
| `industrialization` | Industrialization | 120,000 kp | 6,000 | `steam_power` | +50% all production, +0.5 steel/tick |
| `railroads` | Railroads | 90,000 kp | 5,000 | `steam_power`, `road_building` | +100% gold rate, +200 all storage |
| `rifling` | Rifling | 80,000 kp | 4,800 | `gunpowder` | +50% military power |
| `clockwork_automation` | Clockwork Automation | 50,000 kp | 5,400 | `chronometry` | +10% tick speed |

---

### Victorian Age (~23 min/tech)

| Key | Name | Cost | Ticks | Prerequisites | Effect |
|---|---|---|---|---|---|
| `electrification` | Electrification | 180,000 kp | 7,000 | `industrialization` | +1.0 electricity/tick, +20% all production |
| `telecommunications` | Telecommunications | 150,000 kp | 6,600 | `electrification` | +40% knowledge rate, +50% gold rate |
| `mass_production` | Mass Production | 200,000 kp | 7,400 | `industrialization`, `railroads` | +40% all production, +1.0 steel/tick |

---

### Electric Age (~32 min/tech)

| Key | Name | Cost | Ticks | Prerequisites | Effect |
|---|---|---|---|---|---|
| `power_distribution` | Power Distribution | 300,000 kp | 9,500 | `electrification` | +3.0 electricity/tick, +30% all production |
| `radio` | Radio | 250,000 kp | 9,000 | `telecommunications` | +2.0 culture/tick, +40% knowledge rate |
| `chemical_engineering` | Chemical Engineering | 280,000 kp | 9,800 | `mass_production` | +1.0 oil/tick, +20% all production |

---

### Atomic Age (~45 min/tech)

| Key | Name | Cost | Ticks | Prerequisites | Effect |
|---|---|---|---|---|---|
| `nuclear_fission` | Nuclear Fission | 500,000 kp | 13,050 | `power_distribution`, `chemical_engineering` | +5.0 electricity/tick, +0.5 uranium/tick |
| `rocketry` | Rocketry | 400,000 kp | 12,000 | `rifling`, `chemical_engineering` | +100% military power, +50% expedition reward |
| `nuclear_deterrence` | Nuclear Deterrence | 600,000 kp | 15,000 | `nuclear_fission`, `rocketry` | +150% military power |

---

### Modern Age (~1.1 hr/tech)

| Key | Name | Cost | Ticks | Prerequisites | Effect |
|---|---|---|---|---|---|
| `electricity_tech` | Electricity | 800,000 kp | 18,000 | `nuclear_fission` | +50% all production, +5.0 electricity/tick |
| `computers` | Computers | 1,000,000 kp | 20,000 | `electricity_tech` | +80% knowledge rate |
| `satellite_tech` | Satellite Technology | 1,200,000 kp | 19,000 | `rocketry`, `electricity_tech` | +1.0 data/tick, +60% knowledge rate |
| `nanofabrication` | Nanofabrication | 1,100,000 kp | 19,000 | `computers` | −8% build cost |

---

### Information Age (~1.5 hr/tech)

| Key | Name | Cost | Ticks | Prerequisites | Effect |
|---|---|---|---|---|---|
| `internet` | Internet | 2,000,000 kp | 26,000 | `computers`, `satellite_tech` | +3.0 data/tick, +120% knowledge rate |
| `cybersecurity` | Cybersecurity | 1,800,000 kp | 24,000 | `computers` | +100% military power, +5,000 data storage |
| `social_media` | Social Media | 1,500,000 kp | 23,000 | `internet` | +5.0 culture/tick, +5.0 gold/tick |
| `medical_nanobots` | Medical Nanobots | 1,700,000 kp | 24,000 | `nanofabrication` | +10 population cap, +8.0 food/tick |

---

### Digital Age (~1.8 hr/tech)

| Key | Name | Cost | Ticks | Prerequisites | Effect |
|---|---|---|---|---|---|
| `machine_learning` | Machine Learning | 3,500,000 kp | 34,000 | `internet`, `cybersecurity` | +5.0 data/tick, +50% all production |
| `cloud_computing` | Cloud Computing | 3,000,000 kp | 32,000 | `internet` | +8.0 data/tick, +10,000 all storage |
| `self_replication` | Self-Replication | 3,200,000 kp | 33,000 | `medical_nanobots`, `machine_learning` | +200 nanobots/tick |

---

### Cyberpunk Age (~2.8 hr/tech)

| Key | Name | Cost | Ticks | Prerequisites | Effect |
|---|---|---|---|---|---|
| `neural_interface` | Neural Interface | 6,000,000 kp | 48,000 | `machine_learning` | +30% gather rate, +200% knowledge rate |
| `blockchain` | Blockchain | 5,000,000 kp | 45,000 | `cybersecurity`, `cloud_computing` | +2.0 crypto/tick, +200% gold rate |
| `cybernetics` | Cybernetics | 5,500,000 kp | 50,000 | `neural_interface` | +50% all production, +100% military power |

---

### Fusion Age (~3.7 hr/tech)

| Key | Name | Cost | Ticks | Prerequisites | Effect |
|---|---|---|---|---|---|
| `fusion_power` | Fusion Power | 10,000,000 kp | 65,000 | `nuclear_fission`, `cybernetics` | +20.0 electricity/tick, +1.0 plasma/tick |
| `plasma_physics` | Plasma Physics | 9,000,000 kp | 62,000 | `fusion_power` | +3.0 plasma/tick, +30% all production |
| `superconductors` | Superconductors | 11,000,000 kp | 70,000 | `fusion_power` | +50% all production, +50,000 all storage |

---

### Space Age (~5 hr/tech)

| Key | Name | Cost | Ticks | Prerequisites | Effect |
|---|---|---|---|---|---|
| `orbital_mechanics` | Orbital Mechanics | 20,000,000 kp | 85,000 | `rocketry`, `plasma_physics` | +1.0 titanium/tick, +100% expedition reward |
| `space_mining` | Space Mining | 18,000,000 kp | 82,000 | `orbital_mechanics` | +3.0 titanium/tick, +20.0 iron/tick |
| `zero_g_manufacturing` | Zero-G Manufacturing | 22,000,000 kp | 92,000 | `orbital_mechanics`, `superconductors` | +50% all production, +10.0 steel/tick |

---

### Interstellar Age (~7 hr/tech)

| Key | Name | Cost | Ticks | Prerequisites | Effect |
|---|---|---|---|---|---|
| `warp_drive` | Warp Drive | 40,000,000 kp | 120,000 | `space_mining`, `zero_g_manufacturing` | +1.0 dark matter/tick, +200% expedition reward |
| `stellar_engineering` | Stellar Engineering | 45,000,000 kp | 130,000 | `warp_drive` | +10.0 plasma/tick, +100.0 electricity/tick |

---

### Galactic Age (~9 hr/tech)

| Key | Name | Cost | Ticks | Prerequisites | Effect |
|---|---|---|---|---|---|
| `galactic_navigation` | Galactic Navigation | 80,000,000 kp | 160,000 | `warp_drive`, `stellar_engineering` | +50% all production, +5.0 dark matter/tick |
| `antimatter_synthesis` | Antimatter Synthesis | 90,000,000 kp | 180,000 | `galactic_navigation` | +2.0 antimatter/tick, +30% all production |

---

### Quantum Age (~12 hr/tech)

| Key | Name | Cost | Ticks | Prerequisites | Effect |
|---|---|---|---|---|---|
| `quantum_mechanics` | Quantum Mechanics | 150,000,000 kp | 220,000 | `antimatter_synthesis` | +2.0 quantum flux/tick, +100% all production |
| `reality_manipulation` | Reality Manipulation | 200,000,000 kp | 250,000 | `quantum_mechanics` | +5.0 quantum flux/tick, +100% all production |
| `quantum_computing` | Quantum Computing | 150,000,000 kp | 200,000 | `clockwork_automation` | **+15% tick speed** |

---

### Transcendent Age (~18 hr)

| Key | Name | Cost | Ticks | Prerequisites | Effect |
|---|---|---|---|---|---|
| `transcendence` | Transcendence | 500,000,000 kp | 320,000 | `reality_manipulation` | +200% all production, +10.0 quantum flux/tick |

---

## Tech Effects Reference

### `production_all` — Multiplier on All Positive Rates

These are the biggest single techs in the game. Each adds a fractional multiplier that stacks additively across all sources before being applied to every positive production rate.

| Tech | Bonus |
|---|---|
| Steam Power | +0.30 |
| Industrialization | +0.50 |
| Electrification | +0.20 |
| Mass Production | +0.40 |
| Power Distribution | +0.30 |
| Chemical Engineering | +0.20 |
| Electricity | +0.50 |
| Machine Learning | +0.50 |
| Cybernetics | +0.50 |
| Plasma Physics | +0.30 |
| Superconductors | +0.50 |
| Zero-G Manufacturing | +0.50 |
| Galactic Navigation | +0.50 |
| Antimatter Synthesis | +0.30 |
| Quantum Mechanics | +1.00 |
| Reality Manipulation | +1.00 |
| Transcendence | +2.00 |

By the Transcendent Age, accumulated `production_all` bonuses run well past +10.0 (1,000%). Every percentage point compounds against your entire production base.

---

### `knowledge_rate` — Knowledge Income Multiplier

Knowledge rate techs multiply the output of all knowledge-producing buildings. Since more knowledge income means you can fund more expensive late-game techs:

| Tech | Bonus |
|---|---|
| Primitive Writing | +10% |
| Mathematics | +20% |
| Philosophy | +30% |
| Alchemy | +15% |
| Printing Press | +40% |
| Telecommunications | +40% |
| Radio | +40% |
| Computers | +80% |
| Satellite Technology | +60% |
| Internet | +120% |
| Neural Interface | +200% |

---

### `tick_speed` — Faster Game Clock

Three techs accelerate how often ticks fire. They stack with each other and with the Prestige "Temporal Mastery" upgrade (+5% per tier, 5 tiers):

| Source | Bonus |
|---|---|
| Chronometry (Medieval) | +5% |
| Clockwork Automation (Industrial) | +10% |
| Quantum Computing (Quantum) | +15% |
| Prestige: Temporal Mastery (max) | +25% |
| **Total (all maxed)** | **+55%** |

Tick speed compounds with research speed — a +55% faster clock means late-game techs that would take hours complete considerably sooner in wall-clock time.

---

### `military_power` — Combat Strength Multiplier

Applied to expedition success and military calculations.

| Tech | Bonus |
|---|---|
| Military Tactics | +20% |
| Siege Warfare | +30% |
| Imperial Legions | +40% |
| Gunpowder | +50% |
| Rifling | +50% |
| Colonialism | +30% |
| Rocketry | +100% |
| Nuclear Deterrence | +150% |
| Cybersecurity | +100% |
| Cybernetics | +100% |

---

### `gold_rate` / `iron_rate` / `stone_rate` / `gather_rate` — Per-Resource Multipliers

Applied to positive production rates of the named resource or worker-gathered rates:

| Tech | Target | Bonus |
|---|---|---|
| Bronze Working | iron_rate | +20% |
| Bronze Working | gather_rate | +10% |
| Currency | gold_rate | +30% |
| Road Building | gold_rate | +20% |
| Road Building | gather_rate | +10% |
| Iron Smelting | iron_rate | +40% |
| Banking | gold_rate | +50% |
| Navigation | gold_rate | +50% |
| Cartography | gold_rate | +50% |
| Mercantilism | gold_rate | +30% |
| Railroads | gold_rate | +100% |
| Telecommunications | gold_rate | +50% |
| Blockchain | gold_rate | +200% |
| Stoneworking | stone_rate | +20% |
| Steel Forging | iron_rate | +30% |
| Tool Making | gather_rate | +15% |
| Neural Interface | gather_rate | +30% |

---

### `expedition_reward` — Expedition Loot Multiplier

| Tech | Bonus |
|---|---|
| Navigation | +30% |
| Cartography | +50% |
| Rocketry | +50% |
| Orbital Mechanics | +100% |
| Warp Drive | +200% |

---

### `storage` — Flat Storage Increases

These add a flat amount to all resource storage caps (or, for Banking, just gold):

| Tech | Target | Amount |
|---|---|---|
| Pottery | all | +25 |
| Masonry | all | +50 |
| Civil Engineering | all | +100 |
| Banking | gold | +100 |
| Railroads | all | +200 |
| Cybersecurity | data | +5,000 |
| Cloud Computing | all | +10,000 |
| Superconductors | all | +50,000 |

---

### `build_cost` — Construction Cost Reduction

Two techs reduce build cost:

| Tech | Bonus |
|---|---|
| Civil Engineering | −5% |
| Nanofabrication | −8% |

This reduction is live: it multiplies your cumulative build cost by `(1 + Σ build_cost)` (floored at 10% of base) alongside the build-cost milestone rewards, and the saving is reflected in the cost the build menu shows. Small but permanent, it stacks with those milestones toward the current ceiling of roughly −32%, so it's worth taking when you're building dozens of structures. See [Buildings](buildings.md#build-cost-reductions).

---

### `capacity` — Population Cap

| Tech | Bonus |
|---|---|
| Feudalism | +5 population capacity |

---

### Special — `production` (flat per-tick addition)

These techs add a flat amount directly to a resource's per-tick production rate. Unlike `bonus` effects (which are multipliers), `production` effects are additive flat increases — the values below are added to your rate each tick regardless of building count or workers:

| Tech | Resource | Flat Bonus |
|---|---|---|
| Fire Mastery | food | +0.1/tick |
| Animal Husbandry | food | +0.2/tick |
| Agriculture | food | +0.5/tick |
| Colonialism | food | +2.0/tick |
| Iron Smelting | iron | +0.2/tick |
| Patronage | knowledge | +0.12/tick |
| Patronage | culture | +0.5/tick |
| Theology | faith | +0.3/tick |
| Alchemy | gold | +0.1/tick |
| Steel Forging | steel | +0.1/tick |
| Mercantilism | gold | +2.0/tick |
| Electrification | electricity | +1.0/tick |
| Power Distribution | electricity | +3.0/tick |
| Nuclear Fission | electricity | +5.0/tick, uranium +0.5/tick |
| Electricity | electricity | +5.0/tick |
| Radio | culture | +2.0/tick |
| Social Media | culture | +5.0/tick, gold +5.0/tick |
| Chemical Engineering | oil | +1.0/tick |
| Machine Learning | data | +5.0/tick |
| Cloud Computing | data | +8.0/tick |
| Blockchain | crypto | +2.0/tick |
| Fusion Power | electricity | +20.0/tick, plasma +1.0/tick |
| Plasma Physics | plasma | +3.0/tick |
| Orbital Mechanics | titanium | +1.0/tick |
| Space Mining | titanium | +3.0/tick, iron +20.0/tick |
| Zero-G Manufacturing | steel | +10.0/tick |
| Warp Drive | dark matter | +1.0/tick |
| Stellar Engineering | plasma | +10.0/tick, electricity +100.0/tick |
| Galactic Navigation | dark matter | +5.0/tick |
| Antimatter Synthesis | antimatter | +2.0/tick |
| Quantum Mechanics | quantum flux | +2.0/tick |
| Reality Manipulation | quantum flux | +5.0/tick |
| Transcendence | quantum flux | +10.0/tick |

---

## Knowledge Workers

The knowledge domain lineage produces all your research fuel. Workers in knowledge buildings are called **Shamans** in the Primitive Age, eventually becoming **Quantum Theorists** in the Quantum Age. Assign them using:

```
assign <building_key> [count|all]
```

Knowledge buildings scale as: `rate = 0.002 × 2^tier`. A fully staffed high-tier knowledge building produces dramatically more per tick than multiple low-tier ones. Prioritise upgrading your knowledge lineage and assigning workers to the highest-tier building you can afford.

The prestige **Research Speed** upgrade adds +5% per tier to `knowledge_rate` — five tiers gives your knowledge buildings a permanent +25% output multiplier from the very start of each run.

---

## Strategy

### Prioritise Knowledge Rate Early

Your first research bottleneck is knowledge income, not tick count. Rush `primitive_writing` → `mathematics` → `philosophy` to stack knowledge rate bonuses in the first three ages. Every percent of knowledge rate you earn early pays off across hundreds of future techs.

### The Research Speed Snowball

There is a natural research compound loop: research rate bonuses that make future research cheaper and faster, then spend that efficiency on the next one. The chain looks like:
```
primitive_writing → mathematics → philosophy → printing_press → …
```
Each of these improves `knowledge_rate`, meaning the next tech arrives faster in wall-clock time.

### Tick Speed: A Hidden Multiplier

`chronometry` (Medieval, no prerequisites) is one of the cheapest techs relative to its impact. +5% tick speed means every future tick-based process — research, building, expeditions — completes 5% faster. Research it early, then chain `clockwork_automation` in the Industrial Age for another +10%.

### When to Cancel

Cancelling costs you the full knowledge payment — no refund. Cancelling is only sensible when:
- You've unlocked a new age and realised a different tech path gates a critical resource you need now.
- An epoch event is about to fire and you want to pivot to a prerequisite for something the event might complete for free (see Grand Discovery below).
- You started a very expensive tech before realising you can't sustain knowledge income through its duration.

As a rule: if you're more than halfway through the tick count, finish it.

### The Grand Discovery Epoch Event

The **Grand Discovery** (`good_major` epoch event) instantly completes up to 3 available, unresearched technologies from your current age — for free, bypassing knowledge costs. It fires during positive epoch events when your culture is high enough to unlock major events.

You can't control exactly which 3 techs get selected, but you can influence the pool by pre-researching the techs you don't want to "waste" a slot on. If you have 3 desirable expensive techs you haven't started yet when the event fires, all three can complete in a single event.

Any tech currently in progress that gets completed by Grand Discovery is handled cleanly — the in-progress slot clears automatically.

### The Ancient Civilization Memory

A second way to skip the research grind exists, but only at the very start of a fresh prestige run. While you are still in the Primitive or Stone age, an **ancient cache** has a ~40% chance (once per run) to offer one age-appropriate technology you haven't researched. Accepting it researches that tech immediately — **free of prerequisites, the age gate, and knowledge cost** — but at **half research speed** (2× the normal tick count). The reachable tier scales with prestige level (one extra age of reach per two levels), so a high-prestige run can pull in a tech from an age it hasn't reached yet.

Unlike Grand Discovery, this is a prestige-run mechanic and never fires on your first-ever run (it requires prestige level ≥ 1). See [Prestige](prestige.md#ancient-civilization-memory) for full conditions.

### Late-Game Knowledge Scaling

Knowledge costs scale steeply: from 800 kp (Primitive) to 500,000,000 kp (Transcendent). In the Space and Interstellar ages, individual techs cost tens of millions of kp. Focus your knowledge lineage build and max out all knowledge-rate techs before reaching those ages or the wait becomes prohibitive.

---

## Tips & Common Mistakes

**Knowledge is deducted upfront.** Don't start a tech if your stockpile barely covers the cost — one bad event (The Dark Age cuts knowledge by 80% and cancels your active research) could set you back significantly.

**Prerequisites stack.** Before typing `research mathematics`, check that you have both `primitive_writing` and `currency`. Use `research list` to see what's gated and why.

**The Dark Age epoch event** cancels your active research and drains 80% of your knowledge stockpile. If an epoch catastrophe is imminent, consider whether to delay an expensive research start until after the event resolves.

**Prestige resets research** entirely. All techs, all bonuses — gone. The only persistent research benefit across a prestige reset is the **Ancient Knowledge** bonus (+25% research_speed) that comes from Succumbing, and the knowledge-rate bonus from the prestige upgrade shop.

**Succumbing early is worth considering.** Succumbing to an epoch catastrophe in the Stone Era costs you a run but grants +25% research speed permanently. Players who Succumb at least once and invest in the Research Speed prestige upgrade begin each subsequent run with noticeably faster research from tick one.

**Don't overlook `civil_engineering`** — −5% build cost plus +100 all storage is exceptionally good value in the Classical Age and helps throughout the rest of the run.

---

*See also: [Epochs](epochs.md) for how Grand Discovery and the Dark Age event fire — [Prestige](prestige.md) for the Research Speed upgrade and the Ancient Civilization Memory — [Buildings](buildings.md) for knowledge lineage construction.*
