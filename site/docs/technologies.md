# Technologies

Research is your civilization's strongest long-term lever. 77 technologies span all 22 ages, and each one permanently changes your production, military strength, storage, or the pace of the game. Only one technology is researched at a time, but the [build plan](plan.md) can queue the next ones and start each as soon as the slot frees up.

---

## How Research Works

### Starting Research

Research costs **knowledge points (kp)**, deducted immediately when you start. There is no refund if you cancel: the knowledge is gone the moment the tech starts.

Once started, the tech counts down in **ticks**. Each game tick takes 1 off the counter. When it hits zero, the effects are applied instantly and permanently.

**Tick count with research speed:**
```
ticks = base ticks × (1 − research speed), rounded down, never below 1
```

A research speed of +30% cuts the tick count to 70% of base. It cuts the number of ticks, not their length. A tick is 2 seconds of real time at base; game speed bonuses make each tick shorter (see [Game speed](#game-speed)). The two multiply, so a civilization with high research speed **and** high game speed researches much faster.

The tick count is locked in when you start the tech. Gaining more research speed mid-research does not shorten the current countdown.

**Era Mastery.** On known ground (an age a past run completed) the adjusted ticks are then divided by the age's [Era Mastery](prestige.md#era-mastery) speed, rounded up and never below one tick: 2x after one completion, up to 4.2x after ten. The Research panel shows the shortened times.

### Knowledge Cost is Upfront

Knowledge is removed from your stockpile when the tech starts, before any ticks pass. If you don't have enough, the command fails, and since you can't hold more knowledge than your storage, your knowledge storage must be at least the tech's cost. If your knowledge income drops to zero during a long research countdown, **research still completes**: the ticks count down whatever your knowledge income is, because the cost was already paid.

### One Slot, and the Plan as a Queue

There is one research slot. If you type `research` for a second tech while one is in progress, you get an error naming the active tech and the time it has left.

To line techs up, add them to the [build plan](plan.md) with `plan research <tech>` (or `plan res`). The plan starts each one as soon as the slot is free and you hold its knowledge, in plan order, and it keeps doing so while you are away:

- Only the first research item in the plan can take the slot when it frees up. A tech further down waits its turn even if it is cheaper.
- A planned tech holds back its knowledge cost while it waits, so plan items below it can't spend that knowledge.
- A tech's prerequisites must be researched, in progress, or planned above it. The plan refuses a tech that fails this, and drops a planned one whose prerequisite you remove from above it.
- You can plan the next age's techs; they wait for the advance.

See [How it runs](plan.md#how-it-runs) for the rest of the plan's rules.

### When Research Completes

Effects are applied the tick the counter hits zero. You'll see a success message in the log. The tech is marked researched, and its bonuses apply to your rates from the next tick.

---

## Research Speed Sources

Research speed reduces the tick count when research starts. Its sources add together:

| Source | Research speed | Notes |
|---|---|---|
| **Tech Pioneer** milestone (research 15 techs) | +5% | Scholar Chain |
| **Philosophes** milestone (35 techs, from the Classical Age) | +5% | |
| **Renaissance Mind** milestone (42 techs, from the Renaissance Age) | +10% | Scholar Chain, hidden until you get close |
| **Tech Master** milestone (50 techs, from the Information Age) | +10% | Scholar Chain, hidden; also +5% all production |
| **Tech Ascendant** milestone (all 77 techs, Transcendent Age) | +20% | Hidden. It arrives with your last tech, so it never shortens one |
| **Ancient Knowledge** (Succumb) | +25% per epoch | For each distinct epoch you Succumb in (Iron to Cosmic, up to +150%); kept through Succumb, prestige and save/load. See [Ancient Knowledge](catastrophe.md#ancient-knowledge) |
| **Techs** | none | No tech has a research speed effect |

Milestone research speed lasts for the run: milestones start over at prestige and at Succumb. Five milestones add +50% in total, +30% of it before your last tech. See [Milestones](milestones.md).

Nothing caps research speed except that a tech always takes at least 1 tick. Once your total reaches +100%, every tech finishes on the tick after you start it.

> **Note on the prestige upgrade Knowledge Production:** its key is `research_speed` (`prestige buy research_speed`), but it raises knowledge output by 5% per tier, not research speed. More knowledge lets you afford techs sooner; each tech still takes the same number of ticks. Many techs raise knowledge output too (see [Knowledge output](#knowledge-output) below).

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
plan research <tech_key>
```
Adds a tech to the build plan, which starts it when the research slot is free and you hold its knowledge (see [One Slot, and the Plan as a Queue](#one-slot-and-the-plan-as-a-queue)). `plan res` is the short form.

---

```
research list
```
Prints the techs you can start now (age reached, prerequisites done, not yet researched) with their keys and knowledge costs, then the active research and the approximate wall-clock time left on it (e.g. `~4m 44s`). See [Timers and durations](commands.md#timers-and-durations).

---

```
research cancel
```
Cancels the current research. **No refund.** The knowledge cost is lost. Only use this when pivoting is worth more than the sunk cost.

---

```
research
```
With no arguments, opens the **Research panel** (so does `techs`). It groups techs by age: researched techs show as complete, available ones are highlighted, and locked ones are dimmed. Ages past your next one are not listed: one line counts the techs they hold, so the tree never spoils an age you haven't reached.

---

**Shortcut:** `res` is an alias for `research`. All subcommands work identically.

---

## Tech Tree by Age

Prerequisites are listed by tech key.

Research time is capped at **one eighth of the tech's age target** (see [How Long Each Age Takes](ages.md#how-long-each-age-takes)), so the handful of techs each age offers fits inside it. From the Bronze to the Colonial Age most techs sit below that cap, at their own research times; in every other age each tech sits at the cap, so all techs of that age share the same tick count. Ticks and times below are base values (2 seconds a tick), before research speed and game speed bonuses. "+X% all production" effects add into one pool that is clamped at x3.0, so from about the Electric Age most of them no longer raise your output (see [The all-production cap](resources.md#the-all-production-cap) and [All production](#all-production) below).

**Mid-age unlocks.** Four techs open a building partway through their age instead of at its start: Civilian Reactors (Atomic, the Nuclear Plant), Internet of Things (Information, the Smart Farm and the Smart Complex), Holography (Cyberpunk, the Holographic Theater) and Maglev Transit (Fusion, the Energy Exchange). Their knowledge cost is what times them: each is priced so a player affords it about halfway through the age, after its other techs, so a long stretch of saving up has something new in the middle. In the Cyberpunk Age, Cybernetics is priced to finish about 26 hours in, with Holography some 11 hours later. In the Fusion Age, Plasma Physics, Superconductors and Maglev Transit run one after another, each needing the one before and priced to finish about 17, 24 and 32 hours in. A gated building stays hidden from the build list until its tech is done, and `build` and `upgrade` refuse it until then; you can still add it to your build plan, where it waits for the tech. Buildings you already have stay built. See [Buildings a tech opens](buildings.md#buildings-a-tech-opens).

### Primitive Age (~2 min/tech)

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

### Bronze Age (~23 to 29 min/tech)

| Key | Name | Cost | Ticks | Prerequisites | Effect |
|---|---|---|---|---|---|
| `bronze_working` | Bronze Working | 1.6K kp | 750 | `stoneworking` | +20% stone output, +10% worker output |
| `agriculture` | Agriculture | 12K kp | 700 | `animal_husbandry` | +0.5 food/tick |
| `currency` | Currency | 17.5K kp | 800 | `primitive_writing` | +30% gold output |
| `masonry` | Masonry | 13K kp | 700 | `stoneworking` | +50 storage for every resource |
| `military_tactics` | Military Tactics | 20K kp | 877 | `bronze_working` | +20% military power |

---

### Iron Age (~32 to 40 min/tech)

| Key | Name | Cost | Ticks | Prerequisites | Effect |
|---|---|---|---|---|---|
| `iron_smelting` | Iron Smelting | 30K kp | 1,100 | `bronze_working` | +40% iron output, +0.2 iron/tick |
| `road_building` | Road Building | 25K kp | 950 | `masonry` | +20% gold output, +10% worker output |
| `mathematics` | Mathematics | 37.5K kp | 1,200 | `primitive_writing`, `currency` | +20% knowledge output |
| `siege_warfare` | Siege Warfare | 35K kp | 1,005 | `military_tactics` | +30% military power |

---

### Classical Age (~43 to 53 min/tech)

| Key | Name | Cost | Ticks | Prerequisites | Effect |
|---|---|---|---|---|---|
| `philosophy` | Philosophy | 20K kp | 1,500 | `mathematics`, `primitive_writing` | +30% knowledge output, +0.2 culture/tick |
| `civil_engineering` | Civil Engineering | 18K kp | 1,300 | `masonry`, `road_building` | +100 storage for every resource, −5% build cost |
| `imperial_legions` | Imperial Legions | 22K kp | 1,600 | `siege_warfare`, `iron_smelting` | +40% military power |

---

### Medieval Age (~57 min to 1h 13m/tech)

| Key | Name | Cost | Ticks | Prerequisites | Effect |
|---|---|---|---|---|---|
| `steel_forging` | Steel Forging | 25K kp | 2,000 | `iron_smelting` | +0.25 steel/tick, +30% iron output |
| `theology` | Theology | 20K kp | 1,800 | `philosophy` | +0.3 faith/tick |
| `banking` | Banking | 30K kp | 2,100 | `currency`, `mathematics` | +50% gold output, +100 gold storage |
| `feudalism` | Feudalism | 22K kp | 1,700 | `military_tactics` | +5 housing |
| `alchemy` | Alchemy | 28K kp | 2,200 | `mathematics` | +15% knowledge output, +0.1 gold/tick |
| `chronometry` | Chronometry | 20K kp | 1,900 | none | +5% game speed |

---

### Renaissance Age (~1h 23m to 1h 46m/tech)

| Key | Name | Cost | Ticks | Prerequisites | Effect |
|---|---|---|---|---|---|
| `printing_press` | Printing Press | 50K kp | 3,000 | `theology`, `alchemy` | +40% knowledge output, +0.3 culture/tick |
| `navigation` | Navigation | 45K kp | 2,600 | `mathematics`, `road_building` | +50% gold output, +30% expedition reward |
| `gunpowder` | Gunpowder | 55K kp | 3,200 | `alchemy`, `siege_warfare` | +50% military power |
| `patronage` | Patronage | 40K kp | 2,500 | `banking` | +0.5 culture/tick, +0.12 knowledge/tick |

---

### Colonial Age (~2h 6m to 2h 16m/tech)

| Key | Name | Cost | Ticks | Prerequisites | Effect |
|---|---|---|---|---|---|
| `cartography` | Cartography | 80K kp | 4,000 | `navigation` | +50% expedition reward, +50% gold output |
| `mercantilism` | Mercantilism | 75K kp | 3,800 | `banking`, `navigation` | +2.0 gold/tick, +30% gold output |
| `colonialism` | Colonialism | 90K kp | 4,095 | `cartography`, `gunpowder` | +2.0 food/tick, +30% military power |

---

### Industrial Age (~2h 36m/tech)

| Key | Name | Cost | Ticks | Prerequisites | Effect |
|---|---|---|---|---|---|
| `steam_power` | Steam Power | 100K kp | 4,680 | `steel_forging` | +30% all production |
| `industrialization` | Industrialization | 120K kp | 4,680 | `steam_power` | +50% all production, +0.5 steel/tick |
| `railroads` | Railroads | 90K kp | 4,680 | `steam_power`, `road_building` | +100% gold output, +200 storage for every resource |
| `rifling` | Rifling | 80K kp | 4,680 | `gunpowder` | +50% military power |
| `clockwork_automation` | Clockwork Automation | 50K kp | 4,680 | `chronometry` | +10% game speed |

---

### Victorian Age (~2h 55m/tech)

| Key | Name | Cost | Ticks | Prerequisites | Effect |
|---|---|---|---|---|---|
| `electrification` | Electrification | 180K kp | 5,265 | `industrialization` | +1.0 electricity/tick, +20% all production |
| `telecommunications` | Telecommunications | 150K kp | 5,265 | `electrification` | +40% knowledge output, +50% gold output |
| `mass_production` | Mass Production | 200K kp | 5,265 | `industrialization`, `railroads` | +40% all production, +1.0 steel/tick |

---

### Electric Age (~3h 15m/tech)

| Key | Name | Cost | Ticks | Prerequisites | Effect |
|---|---|---|---|---|---|
| `power_distribution` | Power Distribution | 300K kp | 5,850 | `electrification` | +3.0 electricity/tick, +30% all production |
| `radio` | Radio | 250K kp | 5,850 | `telecommunications` | +2.0 culture/tick, +40% knowledge output |
| `chemical_engineering` | Chemical Engineering | 280K kp | 5,850 | `mass_production` | +1.0 oil/tick, +20% all production |

---

### Atomic Age (~3h 54m/tech)

| Key | Name | Cost | Ticks | Prerequisites | Effect |
|---|---|---|---|---|---|
| `nuclear_fission` | Nuclear Fission | 500K kp | 7,020 | `power_distribution`, `chemical_engineering` | +5.0 electricity/tick, +0.5 uranium/tick |
| `rocketry` | Rocketry | 400K kp | 7,020 | `rifling`, `chemical_engineering` | +100% military power, +50% expedition reward |
| `nuclear_deterrence` | Nuclear Deterrence | 600K kp | 7,020 | `nuclear_fission`, `rocketry` | +150% military power |
| `civilian_reactors` | Civilian Reactors | 3.4B kp | 7,020 | `nuclear_deterrence` | +5.0 electricity/tick, +0.5 uranium/tick; opens the Nuclear Plant |

---

### Modern Age (~3h 54m/tech)

| Key | Name | Cost | Ticks | Prerequisites | Effect |
|---|---|---|---|---|---|
| `electricity_tech` | Advanced Electrics | 800K kp | 7,020 | `nuclear_fission` | +50% all production, +5.0 electricity/tick |
| `computers` | Computers | 1M kp | 7,020 | `electricity_tech` | +80% knowledge output |
| `satellite_tech` | Satellite Technology | 1.2M kp | 7,020 | `rocketry`, `electricity_tech` | +1.0 data/tick, +60% knowledge output |
| `nanofabrication` | Nanofabrication | 1.1M kp | 7,020 | `computers` | −8% build cost |

---

### Information Age (~4h 33m/tech)

| Key | Name | Cost | Ticks | Prerequisites | Effect |
|---|---|---|---|---|---|
| `internet` | Internet | 2M kp | 8,190 | `computers`, `satellite_tech` | +3.0 data/tick, +120% knowledge output |
| `cybersecurity` | Cybersecurity | 1.8M kp | 8,190 | `computers` | +100% military power, +5K data storage |
| `social_media` | Social Media | 1.5M kp | 8,190 | `internet` | +5.0 culture/tick, +5.0 gold/tick |
| `medical_nanobots` | Medical Nanobots | 1.7M kp | 8,190 | `nanofabrication` | +10 housing, +8.0 food/tick |
| `internet_of_things` | Internet of Things | 4.5B kp | 8,190 | `social_media`, `cybersecurity`, `medical_nanobots` | +3.0 data/tick, +8.0 food/tick; opens the Smart Farm and the Smart Complex |

---

### Digital Age (~5h 12m/tech)

| Key | Name | Cost | Ticks | Prerequisites | Effect |
|---|---|---|---|---|---|
| `machine_learning` | Machine Learning | 3.5M kp | 9,360 | `internet`, `cybersecurity` | +5.0 data/tick, +50% all production |
| `cloud_computing` | Cloud Computing | 3M kp | 9,360 | `internet` | +8.0 data/tick, +10K storage for every resource |
| `self_replication` | Self-Replication | 3.2M kp | 9,360 | `medical_nanobots`, `machine_learning` | +200 nanobots/tick |

---

### Cyberpunk Age (~5h 51m/tech)

| Key | Name | Cost | Ticks | Prerequisites | Effect |
|---|---|---|---|---|---|
| `neural_interface` | Neural Interface | 6M kp | 10,530 | `machine_learning` | +30% worker output, +200% knowledge output |
| `blockchain` | Blockchain | 5M kp | 10,530 | `cybersecurity`, `cloud_computing` | +2.0 crypto/tick, +200% gold output |
| `cybernetics` | Cybernetics | 14B kp | 10,530 | `neural_interface` | +50% all production, +100% military power |
| `holography` | Holography | 6.3B kp | 10,530 | `cybernetics`, `blockchain` | +5.0 culture/tick, +2.0 crypto/tick; opens the Holographic Theater |

---

### Fusion Age (~6h 30m/tech)

| Key | Name | Cost | Ticks | Prerequisites | Effect |
|---|---|---|---|---|---|
| `fusion_power` | Fusion Power | 10M kp | 11,700 | `nuclear_fission`, `cybernetics` | +20.0 electricity/tick, +1.0 plasma/tick |
| `plasma_physics` | Plasma Physics | 9.5B kp | 11,700 | `fusion_power` | +3.0 plasma/tick, +30% all production |
| `superconductors` | Superconductors | 4.8B kp | 11,700 | `plasma_physics` | +50% all production, +50K storage for every resource |
| `maglev_transit` | Maglev Transit | 5.4B kp | 11,700 | `superconductors` | +1.0 plasma/tick, +5.0 gold/tick; opens the Energy Exchange |

---

### Space Age (~7h 9m/tech)

| Key | Name | Cost | Ticks | Prerequisites | Effect |
|---|---|---|---|---|---|
| `orbital_mechanics` | Orbital Mechanics | 20M kp | 12,870 | `rocketry`, `plasma_physics` | +1.0 titanium/tick, +100% expedition reward |
| `space_mining` | Space Mining | 18M kp | 12,870 | `orbital_mechanics` | +3.0 titanium/tick, +20.0 iron/tick |
| `zero_g_manufacturing` | Zero-G Manufacturing | 22M kp | 12,870 | `orbital_mechanics`, `superconductors` | +50% all production, +10.0 steel/tick |

---

### Interstellar Age (~7h 48m/tech)

| Key | Name | Cost | Ticks | Prerequisites | Effect |
|---|---|---|---|---|---|
| `warp_drive` | Warp Drive | 40M kp | 14,040 | `space_mining`, `zero_g_manufacturing` | +1.0 dark matter/tick, +200% expedition reward |
| `stellar_engineering` | Stellar Engineering | 45M kp | 14,040 | `warp_drive` | +10.0 plasma/tick, +100.0 electricity/tick |

---

### Galactic Age (~7h 48m/tech)

| Key | Name | Cost | Ticks | Prerequisites | Effect |
|---|---|---|---|---|---|
| `galactic_navigation` | Galactic Navigation | 80M kp | 14,040 | `warp_drive`, `stellar_engineering` | +50% all production, +5.0 dark matter/tick |
| `antimatter_synthesis` | Antimatter Synthesis | 90M kp | 14,040 | `galactic_navigation` | +2.0 antimatter/tick, +30% all production |

---

### Quantum Age (~7h 48m/tech)

| Key | Name | Cost | Ticks | Prerequisites | Effect |
|---|---|---|---|---|---|
| `quantum_mechanics` | Quantum Mechanics | 150M kp | 14,040 | `antimatter_synthesis` | +2.0 quantum flux/tick, +100% all production |
| `reality_manipulation` | Reality Manipulation | 200M kp | 14,040 | `quantum_mechanics` | +5.0 quantum flux/tick, +100% all production |
| `quantum_computing` | Quantum Computing | 150M kp | 14,040 | `clockwork_automation` | **+15% game speed** |

---

### Transcendent Age (~7h 48m)

| Key | Name | Cost | Ticks | Prerequisites | Effect |
|---|---|---|---|---|---|
| `transcendence` | Transcendence | 500M kp | 14,040 | `reality_manipulation` | +200% all production, +10.0 quantum flux/tick |

---

## Tech Effects Reference

### All production

Each of these adds a percentage to all production. The bonuses from every source (techs, wonders, milestones, monuments, events, boons, prestige) add into one pool, and the game multiplies every positive production rate by 1 + that pool, **clamped at x3.0** (+200%). See [The all-production cap](resources.md#the-all-production-cap).

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

The techs through the Electric Age add +190%; with the Crystal Palace and Hoover Dam wonders that passes the cap, and Advanced Electrics (Modern) takes the techs past it on their own. By the Transcendent Age they add +950% on paper. The Stats panel's Active Multipliers shows that raw sum, but output never goes past x3.0. From the Modern Age on, an all-production tech adds nothing you can see while you are over the cap; it still counts as a buffer, because a penalty (such as the Reconstruction Effort after you Endure) comes out of the raw pool first.

---

### Knowledge output

These techs raise your knowledge output, which pays for later techs:

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

Knowledge output bonuses have a pool of their own (shared with the Great Library and Global Network wonders, scholar milestones and the Knowledge Production prestige upgrade), clamped at +200% like all production. The techs through Radio add +195%, so with the Great Library you are at the cap by about the Electric Age, and Computers, Satellite Technology, Internet and Neural Interface add nothing you can see unless something pulls the pool down.

---

### Game speed

Three techs raise game speed: ticks come more often, so production, construction, research and every timer run faster in real time. Game speed isn't part of the all-production pool, so its cap doesn't apply. The bonuses add up with the game's other game speed sources:

| Source | Bonus | Lasts |
|---|---|---|
| Chronometry (Medieval) | +5% | the run |
| Clockwork Automation (Industrial) | +10% | the run |
| Quantum Computing (Quantum) | +15% | the run |
| Prestige upgrade Temporal Mastery | +5% per tier, +25% at tier 5 | every run |
| Prestige level | +1% per level | every run |
| Milestone chain boosts | +250% or +300% | a few minutes, once per chain |
| Time Dilation boons | +8% to +15%, before scaling | 1,950 to 3,900 ticks |

With the three techs and Temporal Mastery maxed, the game runs 55% faster (more with prestige levels), and that multiplies with research speed. See [Prestige](prestige.md#passive-prestige-bonuses), [Milestones](milestones.md) and [Boons](factions.md#boons).

---

### Military power

Military power lowers mission difficulty and raises your defense rating. See [Military Power Bonus](military.md#military-power-bonus).

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

Each resource's bonuses share a pool clamped at +200%, as knowledge's do. The gold techs from Currency to Mercantilism already add +230%, so from the Colonial Age on Railroads, Telecommunications and Blockchain add no gold you can see. Worker output bonuses have no such ceiling.

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
| Civilian Reactors | electricity | +5.0/tick, uranium +0.5/tick |
| Advanced Electrics | electricity | +5.0/tick |
| Radio | culture | +2.0/tick |
| Social Media | culture | +5.0/tick, gold +5.0/tick |
| Internet of Things | data | +3.0/tick, food +8.0/tick |
| Chemical Engineering | oil | +1.0/tick |
| Machine Learning | data | +5.0/tick |
| Cloud Computing | data | +8.0/tick |
| Blockchain | crypto | +2.0/tick |
| Holography | culture | +5.0/tick, crypto +2.0/tick |
| Fusion Power | electricity | +20.0/tick, plasma +1.0/tick |
| Plasma Physics | plasma | +3.0/tick |
| Maglev Transit | plasma | +1.0/tick, gold +5.0/tick |
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

## Paying for Research

Knowledge comes from the Knowledge lineage (Story Circle, Elders' Hall and on up to the Reality Academy), staffed by knowledge workers: **Shamans** in the Primitive Age, **Quantum Theorists** by the Quantum Age. Assign them with `assign <building_key> [count|all]`. The rates for every tier are on the [Knowledge](knowledge.md) page.

The seven mid-age techs of the Atomic to Fusion Ages (Civilian Reactors, Internet of Things, Cybernetics, Holography, Plasma Physics, Superconductors, Maglev Transit) cost billions, far more than the lineage makes: a fully staffed Research Campus makes about 102 knowledge a tick, under 6M over the whole Atomic Age. For those the market is the practical source: from the Industrial Age on, gold buys knowledge at a flat 5 knowledge per gold. See [Knowledge at the Market](knowledge.md#knowledge-at-the-market). `plan trade gold knowledge` buys it as gold comes in, and a `plan research` item behind it starts the tech once the knowledge is there.

Either way, your knowledge storage must hold a tech's full cost before you can pay for it.

The prestige **Knowledge Production** upgrade adds +5% knowledge output per tier, +25% at tier 5, from the start of every run.

---

## Strategy

### Raise Knowledge Output Early

Your first research bottleneck is knowledge income, not tick count. Rush `primitive_writing` → `mathematics` → `philosophy` to stack knowledge output bonuses in the first ages. Each percent you earn early pays off across every later tech.

### The Knowledge Snowball

Knowledge bonuses feed on themselves: each one makes the next tech arrive sooner. The chain looks like:
```
primitive_writing → mathematics → philosophy → printing_press → …
```
Each of these raises knowledge output, so the next tech arrives sooner in real time. The snowball stops at the +200% cap, which you reach by about the Electric Age (see [Knowledge output](#knowledge-output)).

### Game Speed: A Hidden Multiplier

`chronometry` (Medieval, no prerequisites) is one of the cheapest techs for what it does. +5% game speed runs everything that counts in ticks (production, building, research, expeditions) 5% faster. Research it early, then chain `clockwork_automation` in the Industrial Age for another +10%. Line them up with `plan research chronometry` as soon as you enter the Medieval Age.

### When to Cancel

Canceling costs you the full knowledge payment and the progress, with no refund. It only makes sense when a different tech gives a bonus you need now, badly enough to pay for the first one twice. As a rule: if you're more than halfway through the tick count, finish it, and put the urgent tech at the top of your plan so it starts next.

### The Grand Discovery Epoch Event

**The Grand Discovery**, a major good epoch event, instantly completes up to 3 unresearched techs of your current age or earlier, for free. It can come at an epoch transition when your culture is over 40% of its storage (see [Good Epoch Events](epochs.md#good-epoch-events)).

It doesn't weigh which techs are worth most: it takes the first three unresearched techs in alphabetical order of their keys, prerequisites or not. You can shape what it takes by researching the alphabetically early techs of your age first, which leaves the slots for the rest.

If Grand Discovery completes the tech you are researching, the research slot clears automatically (the knowledge you paid for it is not refunded).

### The Ancient Civilization Memory

A second way to skip the research grind comes only at the start of a new run, after a prestige or a Succumb, while you are still in the Primitive or Stone Age. An **ancient cache** has a ~40% chance (once per run) to offer one technology suited to your age that you haven't researched. Accepting it starts that tech at once, **free of prerequisites, the age requirement and knowledge cost**, but at **half research speed** (twice the normal tick count). The reachable tier scales with prestige level (one extra age of reach per two levels), so a high-prestige run can pull in a tech from an age it hasn't reached yet.

Unlike Grand Discovery, it never comes on your first-ever run: it needs prestige level 1 or higher. See [Prestige](prestige.md#ancient-civilization-memory) for full conditions.

### Late-Game Knowledge Scaling

Knowledge costs rise steeply, from 800 kp (Primitive) to hundreds of millions in the last ages, and the seven mid-age techs of the Atomic to Fusion Ages cost billions (Cybernetics 14B). Raise your knowledge storage ahead of them, and plan on buying most of that knowledge with gold (see [Paying for Research](#paying-for-research)).

---

## Tips and Common Mistakes

**Knowledge is deducted upfront.** Don't start a tech if your stockpile barely covers the cost. One bad event (The Dark Age cuts knowledge by 80% and cancels your active research) can set you back a long way.

**Prerequisites stack.** Before typing `research mathematics`, check that you have both `primitive_writing` and `currency`. `research list` shows only the techs you can start now; the Research panel (`research`) shows the locked ones dimmed. Or plan the whole chain: `plan research` takes a tech whose prerequisites are planned above it.

**The Dark Age epoch event** cancels your active research and drains 80% of your knowledge stockpile. If an epoch transition is close, consider whether to delay an expensive research start until after its event resolves.

**Prestige resets research** entirely, with every tech, its bonus and the milestones that gave research speed. The research benefits that survive a prestige are the **Ancient Knowledge** bonus (+25% research speed per epoch) from Succumbing, and the prestige upgrades Knowledge Production (knowledge output) and Temporal Mastery (game speed).

**Succumbing early is worth considering.** Succumbing to a catastrophe in the Iron Era (the earliest era one can strike in) costs you a run but grants +25% research speed permanently, and each further epoch you Succumb in adds another +25%. Players who Succumb at least once begin each later run with faster research from tick one; pair it with Knowledge Production so you can afford the techs as fast as they finish. See [Succumb](catastrophe.md#succumb).

**Don't overlook `civil_engineering`.** −5% build cost plus +100 storage for every resource is good value in the Classical Age and keeps helping for the rest of the run.

---

*See also: [Knowledge](knowledge.md) for making and buying knowledge; [The Build Plan](plan.md) for queuing research; [Epochs](epochs.md) for how Grand Discovery and the Dark Age event fire; [Prestige](prestige.md) for the Knowledge Production upgrade and the Ancient Civilization Memory; [Buildings](buildings.md) for the buildings techs open.*
