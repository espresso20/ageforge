# Technologies

Research is your civilization's strongest long-term lever. 73 technologies span all 22 ages, and each one permanently changes your production, military strength, storage, or the pace of the game. Research is **sequential**: only one technology can be in progress at a time, and it must finish (or be canceled) before you can start the next.

---

## How Research Works

### Starting Research

Research costs **knowledge points (kp)**, deducted immediately when you start. There is no refund if you cancel: the knowledge is gone the moment you type the command.

Once started, the tech counts down in **ticks**. Each game tick decrements the counter by 1. When it hits zero, the effects are applied instantly and permanently.

**Formula for adjusted tick count:**
```
adjusted_ticks = max(1,  base_ticks × (1.0 − research_speed_bonus))
```

A `research_speed` bonus of `0.30` (30%) cuts the tick count to 70% of base. It cuts ticks, not time. Ticks can also run faster through `tick_speed` bonuses, and the two multiply, so a civilization with high research speed **and** high tick speed researches much faster.

The adjusted ticks are locked in at the moment you start the tech. Gaining more `research_speed` mid-research does not retroactively shorten the current countdown.

### Knowledge Cost is Upfront

Knowledge is removed from your stockpile when you issue the `research` command, before any ticks pass. If you don't have enough, the command fails. If your knowledge income drops to zero during a long research countdown, **research still completes**: the ticks count down whatever your knowledge income is, because the cost was already paid.

### Only One Slot

There is no queue. If you try to start a second tech while one is in progress, you get an error showing the active tech and how many ticks remain. Plan your research order in advance.

### When Research Completes

Effects are applied the tick the counter hits zero. You'll see a success message in the log. The tech is marked researched, and its bonuses apply to your rates from the next tick.

---

## Research Speed Sources

`research_speed` reduces the tick count when research starts. Its sources add together. The table also lists the prestige upgrade whose key (`research_speed`) suggests it belongs here:

| Source | How much | Notes |
|---|---|---|
| **Tech bonuses** | None in the current tree | No tech has a research speed effect today |
| **Ancient Knowledge** (Succumb) | +0.25 (25%) per epoch | Granted permanently for each distinct epoch you Succumb in (Iron to Neon in play, up to +125%); survives Succumb, prestige and save/load |
| **Prestige: Knowledge Production** | +0.05 per tier, max 5 tiers (+25%) | Raises knowledge output, not `research_speed` (see the note below) |

> **Note on Prestige "Knowledge Production":** Despite its key (`research_speed`), this prestige upgrade raises knowledge output (how fast you make knowledge), not the `research_speed` bonus that cuts tick counts. More knowledge lets you afford techs sooner, but each tech still takes the same number of ticks.

No tech in the current tree grants `research_speed`, so the tick reduction comes from Succumb's Ancient Knowledge. Many techs raise knowledge output instead (see [Knowledge output](#knowledge-output) below).

---

## Commands

```
research <tech_key>
```
Start researching a technology. Deducts knowledge cost immediately. Fails if: unknown key, already researched, another tech in progress, age requirement not met, prerequisites missing, or insufficient knowledge.

You can also type multi-word tech names with spaces; they are joined with underscores:
```
research bronze working
```
is equivalent to `research bronze_working`.

---

```
research list
```
Lists all technologies available in the current age with their status (researched, in progress, available, locked by prerequisite). Also shows the active research and the approximate wall-clock time left on it (e.g. `~4m 44s`). See [Timers and durations](commands.md#timers-and-durations).

---

```
research cancel
```
Cancels the current research. **No refund.** The knowledge cost is lost. Only use this when pivoting is worth more than the sunk cost.

---

```
research
```
With no arguments, opens the **Research panel**. It groups techs by age: researched techs show as complete, available ones are highlighted, and locked ones are dimmed. Ages past your next one are not listed: one line counts the techs they hold, so the tree never spoils an age you haven't reached.

---

**Shortcut:** `res` is an alias for `research`. All subcommands work identically.

---

## Tech Tree by Age

Prerequisites are listed by tech key.

Research time is capped at **one eighth of the tech's age target** (see [How Long Each Age Takes](ages.md#how-long-each-age-takes)), so the handful of techs each age offers fits inside it. Every tech in the current tree sits at that cap, which is why all techs of one age share the same tick count. Ticks below are at 1× speed (one tick is 2 seconds), before any `research_speed` bonus.

### Primitive Age (~2 min/tech at 1× speed)

| Key | Name | Cost | Ticks | Prerequisites | Effect |
|---|---|---|---|---|---|
| `tool_making` | Tool Making | 800 kp | 56 | none | +15% worker output |
| `fire_mastery` | Fire Mastery | 1K kp | 56 | `tool_making` | +0.1 food/tick |

---

### Stone Age (~6 min/tech)

| Key | Name | Cost | Ticks | Prerequisites | Effect |
|---|---|---|---|---|---|
| `stoneworking` | Stoneworking | 6K kp | 168 | `tool_making` | +20% stone output |
| `animal_husbandry` | Animal Husbandry | 7.5K kp | 168 | `fire_mastery` | +0.2 food/tick |
| `pottery` | Pottery | 5K kp | 168 | `fire_mastery` | +25 storage for every resource |
| `primitive_writing` | Primitive Writing | 10K kp | 168 | `pottery` | +10% knowledge output |

---

### Bronze Age (~11 min/tech)

| Key | Name | Cost | Ticks | Prerequisites | Effect |
|---|---|---|---|---|---|
| `bronze_working` | Bronze Working | 1.6K kp | 337 | `stoneworking` | +20% stone output, +10% worker output |
| `agriculture` | Agriculture | 12K kp | 337 | `animal_husbandry` | +0.5 food/tick |
| `currency` | Currency | 17.5K kp | 337 | `primitive_writing` | +30% gold output |
| `masonry` | Masonry | 13K kp | 337 | `stoneworking` | +50 storage for every resource |
| `military_tactics` | Military Tactics | 20K kp | 337 | `bronze_working` | +20% military power |

---

### Iron Age (~19 min/tech)

| Key | Name | Cost | Ticks | Prerequisites | Effect |
|---|---|---|---|---|---|
| `iron_smelting` | Iron Smelting | 30K kp | 562 | `bronze_working` | +40% iron output, +0.2 iron/tick |
| `road_building` | Road Building | 25K kp | 562 | `masonry` | +20% gold output, +10% worker output |
| `mathematics` | Mathematics | 37.5K kp | 562 | `primitive_writing`, `currency` | +20% knowledge output |
| `siege_warfare` | Siege Warfare | 35K kp | 562 | `military_tactics` | +30% military power |

---

### Classical Age (~26 min/tech)

| Key | Name | Cost | Ticks | Prerequisites | Effect |
|---|---|---|---|---|---|
| `philosophy` | Philosophy | 20K kp | 787 | `mathematics`, `primitive_writing` | +30% knowledge output, +0.2 culture/tick |
| `civil_engineering` | Civil Engineering | 18K kp | 787 | `masonry`, `road_building` | +100 storage for every resource, −5% build cost |
| `imperial_legions` | Imperial Legions | 22K kp | 787 | `siege_warfare`, `iron_smelting` | +40% military power |

---

### Medieval Age (~34 min/tech)

| Key | Name | Cost | Ticks | Prerequisites | Effect |
|---|---|---|---|---|---|
| `steel_forging` | Steel Forging | 25K kp | 1,012 | `iron_smelting` | +0.25 steel/tick, +30% iron output |
| `theology` | Theology | 20K kp | 1,012 | `philosophy` | +0.3 faith/tick |
| `banking` | Banking | 30K kp | 1,012 | `currency`, `mathematics` | +50% gold output, +100 gold storage |
| `feudalism` | Feudalism | 22K kp | 1,012 | `military_tactics` | +5 housing |
| `alchemy` | Alchemy | 28K kp | 1,012 | `mathematics` | +15% knowledge output, +0.1 gold/tick |
| `chronometry` | Chronometry | 20K kp | 1,012 | none | +5% tick speed |

---

### Renaissance Age (~45 min/tech)

| Key | Name | Cost | Ticks | Prerequisites | Effect |
|---|---|---|---|---|---|
| `printing_press` | Printing Press | 50K kp | 1,350 | `theology`, `alchemy` | +40% knowledge output, +0.3 culture/tick |
| `navigation` | Navigation | 45K kp | 1,350 | `mathematics`, `road_building` | +50% gold output, +30% expedition reward |
| `gunpowder` | Gunpowder | 55K kp | 1,350 | `alchemy`, `siege_warfare` | +50% military power |
| `patronage` | Patronage | 40K kp | 1,350 | `banking` | +0.5 culture/tick, +0.12 knowledge/tick |

---

### Colonial Age (~52 min/tech)

| Key | Name | Cost | Ticks | Prerequisites | Effect |
|---|---|---|---|---|---|
| `cartography` | Cartography | 80K kp | 1,575 | `navigation` | +50% expedition reward, +50% gold output |
| `mercantilism` | Mercantilism | 75K kp | 1,575 | `banking`, `navigation` | +2.0 gold/tick, +30% gold output |
| `colonialism` | Colonialism | 90K kp | 1,575 | `cartography`, `gunpowder` | +2.0 food/tick, +30% military power |

---

### Industrial Age (~1 hr/tech)

| Key | Name | Cost | Ticks | Prerequisites | Effect |
|---|---|---|---|---|---|
| `steam_power` | Steam Power | 100K kp | 1,800 | `steel_forging` | +30% all production |
| `industrialization` | Industrialization | 120K kp | 1,800 | `steam_power` | +50% all production, +0.5 steel/tick |
| `railroads` | Railroads | 90K kp | 1,800 | `steam_power`, `road_building` | +100% gold output, +200 storage for every resource |
| `rifling` | Rifling | 80K kp | 1,800 | `gunpowder` | +50% military power |
| `clockwork_automation` | Clockwork Automation | 50K kp | 1,800 | `chronometry` | +10% tick speed |

---

### Victorian Age (~1.1 hr/tech)

| Key | Name | Cost | Ticks | Prerequisites | Effect |
|---|---|---|---|---|---|
| `electrification` | Electrification | 180K kp | 2,025 | `industrialization` | +1.0 electricity/tick, +20% all production |
| `telecommunications` | Telecommunications | 150K kp | 2,025 | `electrification` | +40% knowledge output, +50% gold output |
| `mass_production` | Mass Production | 200K kp | 2,025 | `industrialization`, `railroads` | +40% all production, +1.0 steel/tick |

---

### Electric Age (~1.25 hr/tech)

| Key | Name | Cost | Ticks | Prerequisites | Effect |
|---|---|---|---|---|---|
| `power_distribution` | Power Distribution | 300K kp | 2,250 | `electrification` | +3.0 electricity/tick, +30% all production |
| `radio` | Radio | 250K kp | 2,250 | `telecommunications` | +2.0 culture/tick, +40% knowledge output |
| `chemical_engineering` | Chemical Engineering | 280K kp | 2,250 | `mass_production` | +1.0 oil/tick, +20% all production |

---

### Atomic Age (~1.5 hr/tech)

| Key | Name | Cost | Ticks | Prerequisites | Effect |
|---|---|---|---|---|---|
| `nuclear_fission` | Nuclear Fission | 500K kp | 2,700 | `power_distribution`, `chemical_engineering` | +5.0 electricity/tick, +0.5 uranium/tick |
| `rocketry` | Rocketry | 400K kp | 2,700 | `rifling`, `chemical_engineering` | +100% military power, +50% expedition reward |
| `nuclear_deterrence` | Nuclear Deterrence | 600K kp | 2,700 | `nuclear_fission`, `rocketry` | +150% military power |

---

### Modern Age (~1.5 hr/tech)

| Key | Name | Cost | Ticks | Prerequisites | Effect |
|---|---|---|---|---|---|
| `electricity_tech` | Advanced Electrics | 800K kp | 2,700 | `nuclear_fission` | +50% all production, +5.0 electricity/tick |
| `computers` | Computers | 1M kp | 2,700 | `electricity_tech` | +80% knowledge output |
| `satellite_tech` | Satellite Technology | 1.2M kp | 2,700 | `rocketry`, `electricity_tech` | +1.0 data/tick, +60% knowledge output |
| `nanofabrication` | Nanofabrication | 1.1M kp | 2,700 | `computers` | −8% build cost |

---

### Information Age (~1.75 hr/tech)

| Key | Name | Cost | Ticks | Prerequisites | Effect |
|---|---|---|---|---|---|
| `internet` | Internet | 2M kp | 3,150 | `computers`, `satellite_tech` | +3.0 data/tick, +120% knowledge output |
| `cybersecurity` | Cybersecurity | 1.8M kp | 3,150 | `computers` | +100% military power, +5K data storage |
| `social_media` | Social Media | 1.5M kp | 3,150 | `internet` | +5.0 culture/tick, +5.0 gold/tick |
| `medical_nanobots` | Medical Nanobots | 1.7M kp | 3,150 | `nanofabrication` | +10 housing, +8.0 food/tick |

---

### Digital Age (~2 hr/tech)

| Key | Name | Cost | Ticks | Prerequisites | Effect |
|---|---|---|---|---|---|
| `machine_learning` | Machine Learning | 3.5M kp | 3,600 | `internet`, `cybersecurity` | +5.0 data/tick, +50% all production |
| `cloud_computing` | Cloud Computing | 3M kp | 3,600 | `internet` | +8.0 data/tick, +10K storage for every resource |
| `self_replication` | Self-Replication | 3.2M kp | 3,600 | `medical_nanobots`, `machine_learning` | +200 nanobots/tick |

---

### Cyberpunk Age (~2.25 hr/tech)

| Key | Name | Cost | Ticks | Prerequisites | Effect |
|---|---|---|---|---|---|
| `neural_interface` | Neural Interface | 6M kp | 4,050 | `machine_learning` | +30% worker output, +200% knowledge output |
| `blockchain` | Blockchain | 5M kp | 4,050 | `cybersecurity`, `cloud_computing` | +2.0 crypto/tick, +200% gold output |
| `cybernetics` | Cybernetics | 5.5M kp | 4,050 | `neural_interface` | +50% all production, +100% military power |

---

### Fusion Age (~2.5 hr/tech)

| Key | Name | Cost | Ticks | Prerequisites | Effect |
|---|---|---|---|---|---|
| `fusion_power` | Fusion Power | 10M kp | 4,500 | `nuclear_fission`, `cybernetics` | +20.0 electricity/tick, +1.0 plasma/tick |
| `plasma_physics` | Plasma Physics | 9M kp | 4,500 | `fusion_power` | +3.0 plasma/tick, +30% all production |
| `superconductors` | Superconductors | 11M kp | 4,500 | `fusion_power` | +50% all production, +50K storage for every resource |

---

### Space Age (~2.75 hr/tech)

| Key | Name | Cost | Ticks | Prerequisites | Effect |
|---|---|---|---|---|---|
| `orbital_mechanics` | Orbital Mechanics | 20M kp | 4,950 | `rocketry`, `plasma_physics` | +1.0 titanium/tick, +100% expedition reward |
| `space_mining` | Space Mining | 18M kp | 4,950 | `orbital_mechanics` | +3.0 titanium/tick, +20.0 iron/tick |
| `zero_g_manufacturing` | Zero-G Manufacturing | 22M kp | 4,950 | `orbital_mechanics`, `superconductors` | +50% all production, +10.0 steel/tick |

---

### Interstellar Age (~3 hr/tech)

| Key | Name | Cost | Ticks | Prerequisites | Effect |
|---|---|---|---|---|---|
| `warp_drive` | Warp Drive | 40M kp | 5,400 | `space_mining`, `zero_g_manufacturing` | +1.0 dark matter/tick, +200% expedition reward |
| `stellar_engineering` | Stellar Engineering | 45M kp | 5,400 | `warp_drive` | +10.0 plasma/tick, +100.0 electricity/tick |

---

### Galactic Age (~3 hr/tech)

| Key | Name | Cost | Ticks | Prerequisites | Effect |
|---|---|---|---|---|---|
| `galactic_navigation` | Galactic Navigation | 80M kp | 5,400 | `warp_drive`, `stellar_engineering` | +50% all production, +5.0 dark matter/tick |
| `antimatter_synthesis` | Antimatter Synthesis | 90M kp | 5,400 | `galactic_navigation` | +2.0 antimatter/tick, +30% all production |

---

### Quantum Age (~3 hr/tech)

| Key | Name | Cost | Ticks | Prerequisites | Effect |
|---|---|---|---|---|---|
| `quantum_mechanics` | Quantum Mechanics | 150M kp | 5,400 | `antimatter_synthesis` | +2.0 quantum flux/tick, +100% all production |
| `reality_manipulation` | Reality Manipulation | 200M kp | 5,400 | `quantum_mechanics` | +5.0 quantum flux/tick, +100% all production |
| `quantum_computing` | Quantum Computing | 150M kp | 5,400 | `clockwork_automation` | **+15% tick speed** |

---

### Transcendent Age (~3 hr)

| Key | Name | Cost | Ticks | Prerequisites | Effect |
|---|---|---|---|---|---|
| `transcendence` | Transcendence | 500M kp | 5,400 | `reality_manipulation` | +200% all production, +10.0 quantum flux/tick |

---

## Tech Effects Reference

### All production

These are the biggest single techs in the game. Each adds a percentage to all production. The bonuses from every source add together, and the total multiplies every positive production rate.

| Tech | Bonus |
|---|---|
| Steam Power | +0.30 |
| Industrialization | +0.50 |
| Electrification | +0.20 |
| Mass Production | +0.40 |
| Power Distribution | +0.30 |
| Chemical Engineering | +0.20 |
| Advanced Electrics | +0.50 |
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

By the Transcendent Age, the all-production bonuses from techs alone add up to +9.5 (+950%), before wonders and other sources.

---

### Knowledge output

These techs multiply the output of every knowledge building, which pays for the expensive late-game techs:

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

### Tick speed

Three techs make ticks fire more often. They stack with each other and with the Prestige "Temporal Mastery" upgrade (+5% per tier, 5 tiers):

| Source | Bonus |
|---|---|
| Chronometry (Medieval) | +5% |
| Clockwork Automation (Industrial) | +10% |
| Quantum Computing (Quantum) | +15% |
| Prestige: Temporal Mastery (max) | +25% |
| **Total (all maxed)** | **+55%** |

Tick speed multiplies with research speed: with a clock 55% faster, late-game techs that would take hours finish much sooner in wall-clock time.

---

### Military power

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

### Resource output and worker output

These raise the positive production of one resource, or worker output:

| Tech | Target | Bonus |
|---|---|---|
| Bronze Working | stone | +20% |
| Bronze Working | worker output | +10% |
| Currency | gold | +30% |
| Road Building | gold | +20% |
| Road Building | worker output | +10% |
| Iron Smelting | iron | +40% |
| Banking | gold | +50% |
| Navigation | gold | +50% |
| Cartography | gold | +50% |
| Mercantilism | gold | +30% |
| Railroads | gold | +100% |
| Telecommunications | gold | +50% |
| Blockchain | gold | +200% |
| Stoneworking | stone | +20% |
| Steel Forging | iron | +30% |
| Tool Making | worker output | +15% |
| Neural Interface | worker output | +30% |

---

### Expedition reward

| Tech | Bonus |
|---|---|
| Navigation | +30% |
| Cartography | +50% |
| Rocketry | +50% |
| Orbital Mechanics | +100% |
| Warp Drive | +200% |

---

### Storage

These add a flat amount of storage for every resource (for Banking and Cybersecurity, just one resource):

| Tech | Target | Amount |
|---|---|---|
| Pottery | every resource | +25 |
| Masonry | every resource | +50 |
| Civil Engineering | every resource | +100 |
| Banking | gold | +100 |
| Railroads | every resource | +200 |
| Cybersecurity | data | +5K |
| Cloud Computing | every resource | +10K |
| Superconductors | every resource | +50K |

---

### Build cost

Two techs reduce build cost:

| Tech | Bonus |
|---|---|
| Civil Engineering | −5% |
| Nanofabrication | −8% |

The reduction multiplies every build cost by `(1 + total build cost bonus)` (never below 10% of base), together with the build-cost milestone rewards, and the costs shown in the Buildings panel include it. It stacks with those milestones toward a total of roughly −32%, so it is worth taking when you're building dozens of structures. See [Buildings](buildings.md#build-cost-reductions).

---

### Housing

| Tech | Bonus |
|---|---|
| Feudalism | +5 housing |
| Medical Nanobots | +10 housing |

---

### Flat production

These techs add a flat amount to a resource's production each tick. Unlike the percentage bonuses above, they don't depend on your buildings or workers:

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
| Steel Forging | steel | +0.25/tick |
| Mercantilism | gold | +2.0/tick |
| Electrification | electricity | +1.0/tick |
| Power Distribution | electricity | +3.0/tick |
| Nuclear Fission | electricity | +5.0/tick, uranium +0.5/tick |
| Advanced Electrics | electricity | +5.0/tick |
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

Knowledge building rates, per fully staffed copy: the early lineage is set by hand (Story Circle 0.2, Elders' Hall 0.6, Scriptorium 2.0, Agora 1.6, Library 3.2 knowledge/tick). In the Medieval, Renaissance and Colonial ages, where knowledge is also a building material, the rate is derived from the building's price like any other producer (Monastery Library 78.3, University 208, Natural Philosophy Hall 772). From the Industrial Age on, knowledge buildings follow `rate = 0.05 × 2^tier` (Research Institute 12.8, Academy 25.6, and so on). A fully staffed high-tier knowledge building produces far more per tick than several low-tier ones. Upgrade your knowledge lineage early and put workers in the highest-tier building you can afford.

The prestige **Knowledge Production** upgrade adds +5% knowledge output per tier. Five tiers give your knowledge buildings a permanent +25% from the start of each run.

---

## Strategy

### Raise Knowledge Output Early

Your first research bottleneck is knowledge income, not tick count. Rush `primitive_writing` → `mathematics` → `philosophy` to stack knowledge output bonuses in the first ages. Each percent you earn early pays off across every later tech.

### The Research Speed Snowball

Knowledge bonuses feed on themselves: each one makes the next tech arrive sooner. The chain looks like:
```
primitive_writing → mathematics → philosophy → printing_press → …
```
Each of these raises knowledge output, so the next tech arrives sooner in wall-clock time.

### Tick Speed: A Hidden Multiplier

`chronometry` (Medieval, no prerequisites) is one of the cheapest techs for what it does. +5% tick speed makes everything that runs on ticks (research, building, expeditions) finish 5% faster. Research it early, then chain `clockwork_automation` in the Industrial Age for another +10%.

### When to Cancel

Canceling costs you the full knowledge payment, with no refund. It only makes sense when:
- You've reached a new age and a different tech gives a bonus you need now.
- An epoch event is about to fire and you want to pivot to a prerequisite for something the event might complete for free (see Grand Discovery below).

As a rule: if you're more than halfway through the tick count, finish it.

### The Grand Discovery Epoch Event

The **Grand Discovery** (`good_major` epoch event) instantly completes up to 3 available, unresearched technologies from your current age, for free. It fires during positive epoch events when your culture is high enough to unlock major events.

You can't choose which 3 techs it picks, but you can shape the pool by researching first the techs you don't want it to spend a slot on. If you have 3 desirable expensive techs you haven't started yet when the event fires, all three can complete in a single event.

If Grand Discovery completes the tech you are researching, the research slot clears automatically.

### The Ancient Civilization Memory

A second way to skip the research grind exists, but only at the very start of a fresh prestige run. While you are still in the Primitive or Stone age, an **ancient cache** has a ~40% chance (once per run) to offer one technology suited to your age that you haven't researched. Accepting it starts that tech at once, **free of prerequisites, the age requirement and knowledge cost**, but at **half research speed** (twice the normal tick count). The reachable tier scales with prestige level (one extra age of reach per two levels), so a high-prestige run can pull in a tech from an age it hasn't reached yet.

Unlike Grand Discovery, this is a prestige-run mechanic and never fires on your first-ever run (it needs prestige level 1 or higher). See [Prestige](prestige.md#ancient-civilization-memory) for full conditions.

### Late-Game Knowledge Scaling

Knowledge costs rise steeply: from 800 kp (Primitive) to 500M kp (Transcendent). In the Space and Interstellar ages, single techs cost tens of millions of kp. Build up your knowledge lineage and take every knowledge output tech before you reach those ages, or the wait gets very long.

---

## Tips and Common Mistakes

**Knowledge is deducted upfront.** Don't start a tech if your stockpile barely covers the cost. One bad event (The Dark Age cuts knowledge by 80% and cancels your active research) can set you back a long way.

**Prerequisites stack.** Before typing `research mathematics`, check that you have both `primitive_writing` and `currency`. Use `research list` to see what's locked and why.

**The Dark Age epoch event** cancels your active research and drains 80% of your knowledge stockpile. If an epoch transition is close, consider whether to delay an expensive research start until after its event resolves.

**Prestige resets research** entirely, with every tech and its bonus. The only research benefits that survive a prestige are the **Ancient Knowledge** bonus (+25% research speed per epoch) from Succumbing and the knowledge output bonus from the prestige upgrade shop.

**Succumbing early is worth considering.** Succumbing to a catastrophe in the Iron Era (the earliest era one can strike in) costs you a run but grants +25% research speed permanently, and each further epoch you Succumb in adds another +25%. Players who Succumb at least once and invest in the Research Speed prestige upgrade begin each subsequent run with noticeably faster research from tick one.

**Don't overlook `civil_engineering`.** −5% build cost plus +100 storage for every resource is good value in the Classical Age and keeps helping for the rest of the run.

---

*See also: [Epochs](epochs.md) for how Grand Discovery and the Dark Age event fire; [Prestige](prestige.md) for the Knowledge Production upgrade and the Ancient Civilization Memory; [Buildings](buildings.md) for the knowledge lineage.*
