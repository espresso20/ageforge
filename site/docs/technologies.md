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

The Research panel lists every tech at the time it would take if you started it now, with research speed, Ancient Knowledge and Era Mastery counted, and its header says what your research speed does (`Research speed +30%: techs take 70% of their base time.`).

**Ancient Knowledge.** For each distinct epoch you have [Succumbed](catastrophe.md#ancient-knowledge) in, the adjusted ticks are multiplied by 0.8, rounded down and never below one tick: ×0.8 after one epoch, ×0.64 after two, ×0.26 with all six. It multiplies what research speed leaves, so it never brings a tech to a single tick on its own. The Research panel's header has a line for it (`Ancient Knowledge: research time ×0.64. The times below include it.`).

**Era Mastery.** On known ground (an age a past run completed) the ticks are then divided by the age's [Era Mastery](prestige.md#era-mastery) speed, rounded up and never below one tick: 2x after one completion, up to 4.2x after ten. The Research panel shows the shortened times.

All together:

```
ticks = base ticks × (1 − research speed) × 0.8 ^ epochs succumbed in ÷ Era Mastery speed
```

### Knowledge Cost is Upfront

Knowledge is removed from your stockpile when the tech starts, before any ticks pass. If you don't have enough, the command fails, and since you can't hold more knowledge than your storage, your knowledge storage must be at least the tech's cost. If your knowledge income drops to zero during a long research countdown, **research still completes**: the ticks count down whatever your knowledge income is, because the cost was already paid.

### Keystones, the Spine and the Rest

Every tech has a **kind**, and its kind decides how long it takes and what it costs.

- A **keystone** is the tech an age's wonder needs before it can be built. There is one in each age from the Stone Age on (the Bronze Age's arrives with a later update), and since the next age needs the wonder, the keystone is the one tech an age asks of you. See [Keystone Techs](ages.md#keystone-techs) for the list.
- The **spine** is every tech a keystone stands on, all the way down: Tool Making under Stoneworking, Steam Power under Industrialization. Keystones and the spine are the only techs a run has to research, 43 of the 77 today.
- Everything else is **optional**: yours to take or leave.

The Research panel marks each keystone with a ★ and the wonder it opens (`★ keystone: Colosseum`), and so does `research list`.

### How Research is Priced

No tech has a price or a time of its own. Both follow from its age and its kind.

**Time.** Each age has a research cap: one eighth of the age's target length (see [How Long Each Age Takes](ages.md#how-long-each-age-takes)). A tech takes a share of its age's cap:

| Kind | Share of the age's research cap |
|---|---|
| Spine | 40% |
| Keystone | 50% |
| Optional | 50% |

So no tech takes longer than a sixteenth of its age, and the techs you must research are the quick ones.

**Cost.** An age's techs share one knowledge budget: 30% of the knowledge a well-run civilization makes in the age's target time (50% in the Primitive Age, 90% in the Renaissance). The budget is split by weight:

| Kind | Weight |
|---|---|
| Spine | 0.6 |
| Keystone | 0.8 |
| Optional | 1.0 |

A tech costs the budget times its weight, divided by the weights of all the techs in its age. So the techs you must research are the cheap ones, a keystone always costs less than an optional tech of its age, and researching everything an age offers takes under a third of what the age makes. When a later update adds techs to an age, each tech there gets a little cheaper and the age's total stays where it is.

The Renaissance is the exception. Knowledge is what paces that age (its exchanges buy nearly everything else), and its old requirement asked for 30M of it. Patronage, the Sistine Chapel's keystone, carries that weight now: the Renaissance's four techs cost 80M knowledge together, 18.8M of it for Patronage, about as much as the Colonial Age's. Plan on four Renaissance Vaults before its cheapest tech fits your knowledge storage.

The costs and ticks in the tables below are base values. They used to be typed by hand, tech by tech: early techs cost more knowledge than their own age made (the Stone Age's four cost 28.5K against about 4K made), and from the Classical Age on the whole list cost next to nothing. Now the early techs are within reach in their own ages and the later ones are a real part of each age.

### One Slot, and the Plan as a Queue

There is one research slot. If you type `research` for a second tech while one is in progress, you get an error naming the active tech and the time it has left.

To line techs up, add them to the [build plan](plan.md) with `plan research <tech>` (or `plan res`). The plan starts each one as soon as the slot is free and you hold its knowledge, in plan order, and it keeps doing so while you are away:

- Only the first research item in the plan can take the slot when it frees up. A tech further down waits its turn even if it is cheaper.
- A planned tech holds back its knowledge cost while it waits, so plan items below it can't spend that knowledge.
- A tech's prerequisites must be researched, in progress, or planned above it, and the plan refuses a tech that fails this. A tech already in the plan that loses a prerequisite (you removed it from above, canceled it mid-research, or a game update changed what the tech needs) is not dropped: it waits, and the plan shows what it needs (`needs Philosophy first`). While it waits it holds neither the slot's turn nor its knowledge, so you can plan the missing tech below it and that one starts first.
- You can plan the next age's techs; they wait for the advance.

The plan is the only research queue, and the game never chooses what you research next. The techs you plan are remembered with the rest of your plan: with the prestige legacy kit's [Plan Template](prestige.md#plan-template) owned, they are planned again on later runs, in the age you planned them in. Techs you start by hand with `research` are not remembered.

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
| **Ancient Knowledge** (Succumb) | not research speed: research time ×0.8 per epoch | For each distinct epoch you Succumb in (Iron to Cosmic, ×0.26 with all six). It multiplies the time research speed leaves, so it is not in this pool and no cap holds it. Kept through Succumb, prestige and save/load. See [Ancient Knowledge](catastrophe.md#ancient-knowledge) |
| **Techs** | none | No tech has a research speed effect |

Milestone research speed lasts for the run: milestones start over at prestige and at Succumb. Five milestones add +50% in total, +30% of it before your last tech. See [Milestones](milestones.md).

Nothing caps research speed except that a tech always takes at least 1 tick. At +100% every tech would finish on the tick after you start it, and anything past +100% would add nothing. No run reaches that: the five milestones are the only sources and add +50% between them. (Ancient Knowledge used to be +25% research speed per epoch and reached +100% alone with the fourth; it is a multiplier on research time now.) The Stats panel would mark the Research speed line as capped, and the log would say so when a research speed bonus is past it.

> **Research speed is not knowledge output.** Many techs raise knowledge output (see [Knowledge output](#knowledge-output) below). More knowledge lets you afford techs sooner; each tech still takes the same number of ticks.

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

Prerequisites are listed by tech key. A tech needs every prerequisite it lists.

**Changed prerequisites.** Eleven techs need different techs than they used to. Fire Mastery, Primitive Writing, Feudalism and Rocketry need nothing now. Mathematics no longer needs Currency, Philosophy no longer lists Primitive Writing (Mathematics already needs it), Civil Engineering no longer lists Masonry (Road Building already needs it), Navigation no longer needs Road Building, and Mass Production no longer needs Railroads. Alchemy needs Philosophy instead of Mathematics, and Quantum Computing needs Quantum Mechanics as well as Clockwork Automation. A save from before the change keeps everything it had: a tech you already researched stays researched and still opens the techs that need it, a research in progress finishes, and a tech in your plan waits for what it now needs.

Research time is a share of the age's research cap, one eighth of the age's target: 40% for a spine tech, 50% for a keystone or an optional one (see [How Research is Priced](#how-research-is-priced)). Ticks and times below are base values (2 seconds a tick), before research speed and game speed bonuses. "+X% all production" effects add into one pool that is clamped at x3.0, so from about the Electric Age most of them no longer raise your output (see [The all-production cap](resources.md#the-all-production-cap) and [All production](#all-production) below).

**Mid-age unlocks.** Four techs open a building partway through their age instead of at its start: Civilian Reactors (Atomic, the Nuclear Plant), Internet of Things (Information, the Smart Farm and the Smart Complex), Holography (Cyberpunk, the Holographic Theater) and Maglev Transit (Fusion, the Energy Exchange). What times them is where they stand in the tree: each needs most of its age's other techs first, so it comes after them. Civilian Reactors stands behind Nuclear Deterrence, Internet of Things behind three Information Age techs, Holography behind Cybernetics and Blockchain, and in the Fusion Age, Plasma Physics, Superconductors and Maglev Transit run one after another. A gated building stays hidden from the build list until its tech is done, and `build` and `upgrade` refuse it until then; you can still add it to your build plan, where it waits for the tech. Buildings you already have stay built. See [Buildings a tech opens](buildings.md#buildings-a-tech-opens). A wonder's keystone works the same way for building, but the wonder is listed and its bank is open from the start of the age.

### Primitive Age (~46s to 56s/tech)

| Key | Name | Kind | Cost | Ticks | Prerequisites | Effect |
|---|---|---|---|---|---|---|
| `tool_making` | Tool Making | spine | 61 kp | 23 | none | +15% worker output |
| `fire_mastery` | Fire Mastery | optional | 102 kp | 28 | none | +0.1 food/tick |

---

### Stone Age (~2m 16s to 2m 48s/tech)

| Key | Name | Kind | Cost | Ticks | Prerequisites | Effect |
|---|---|---|---|---|---|---|
| `stoneworking` | Stoneworking | **keystone** (Great Monolith) | 302 kp | 84 | `tool_making` | +20% stone output |
| `animal_husbandry` | Animal Husbandry | optional | 377 kp | 84 | `fire_mastery` | +0.2 food/tick |
| `pottery` | Pottery | optional | 377 kp | 84 | `fire_mastery` | +25 storage for every resource |
| `primitive_writing` | Primitive Writing | spine | 226 kp | 68 | none | +10% knowledge output |

---

### Bronze Age (~11m to 14m/tech)

| Key | Name | Kind | Cost | Ticks | Prerequisites | Effect |
|---|---|---|---|---|---|---|
| `bronze_working` | Bronze Working | spine | 1.02K kp | 351 | `stoneworking` | +20% stone output, +10% worker output |
| `agriculture` | Agriculture | optional | 1.7K kp | 439 | `animal_husbandry` | +0.5 food/tick |
| `currency` | Currency | spine | 1.02K kp | 351 | `primitive_writing` | +30% gold output |
| `masonry` | Masonry | optional | 1.7K kp | 439 | `stoneworking` | +50 storage for every resource |
| `military_tactics` | Military Tactics | optional | 1.7K kp | 439 | `bronze_working` | +20% military power |

---

### Iron Age (~19m to 24m/tech)

| Key | Name | Kind | Cost | Ticks | Prerequisites | Effect |
|---|---|---|---|---|---|---|
| `iron_smelting` | Iron Smelting | spine | 13.8K kp | 585 | `bronze_working` | +40% iron output, +0.2 iron/tick |
| `road_building` | Road Building | optional | 22.9K kp | 731 | `masonry` | +20% gold output, +10% worker output |
| `mathematics` | Mathematics | **keystone** (Colosseum) | 18.4K kp | 731 | `primitive_writing` | +20% knowledge output |
| `siege_warfare` | Siege Warfare | optional | 22.9K kp | 731 | `military_tactics` | +30% military power |

---

### Classical Age (~34m/tech)

| Key | Name | Kind | Cost | Ticks | Prerequisites | Effect |
|---|---|---|---|---|---|---|
| `philosophy` | Philosophy | **keystone** (Parthenon) | 115K kp | 1,024 | `mathematics` | +30% knowledge output, +0.2 culture/tick |
| `civil_engineering` | Civil Engineering | optional | 143K kp | 1,024 | `road_building` | +100 storage for every resource, −5% build cost |
| `imperial_legions` | Imperial Legions | optional | 143K kp | 1,024 | `siege_warfare`, `iron_smelting` | +40% military power |

---

### Medieval Age (~35m to 43m/tech)

| Key | Name | Kind | Cost | Ticks | Prerequisites | Effect |
|---|---|---|---|---|---|---|
| `steel_forging` | Steel Forging | spine | 548K kp | 1,053 | `iron_smelting` | +0.25 steel/tick, +30% iron output |
| `theology` | Theology | **keystone** (Great Library) | 730K kp | 1,316 | `philosophy` | +0.3 faith/tick |
| `banking` | Banking | spine | 548K kp | 1,053 | `currency`, `mathematics` | +50% gold output, +100 gold storage |
| `feudalism` | Feudalism | optional | 913K kp | 1,316 | none | +5 housing |
| `alchemy` | Alchemy | optional | 913K kp | 1,316 | `philosophy` | +15% knowledge output, +0.1 gold/tick |
| `chronometry` | Chronometry | optional | 913K kp | 1,316 | none | +5% game speed |

---

### Renaissance Age (~46m to 58m/tech)

| Key | Name | Kind | Cost | Ticks | Prerequisites | Effect |
|---|---|---|---|---|---|---|
| `printing_press` | Printing Press | optional | 23.5M kp | 1,755 | `theology`, `alchemy` | +40% knowledge output, +0.3 culture/tick |
| `navigation` | Navigation | spine | 14.1M kp | 1,404 | `mathematics` | +50% gold output, +30% expedition reward |
| `gunpowder` | Gunpowder | optional | 23.5M kp | 1,755 | `alchemy`, `siege_warfare` | +50% military power |
| `patronage` | Patronage | **keystone** (Sistine Chapel) | 18.8M kp | 1,755 | `banking` | +0.5 culture/tick, +0.12 knowledge/tick |

---

### Colonial Age (~1h 8m/tech)

| Key | Name | Kind | Cost | Ticks | Prerequisites | Effect |
|---|---|---|---|---|---|---|
| `cartography` | Cartography | **keystone** (Grand Lighthouse) | 21.8M kp | 2,048 | `navigation` | +50% expedition reward, +50% gold output |
| `mercantilism` | Mercantilism | optional | 27.3M kp | 2,048 | `banking`, `navigation` | +2.0 gold/tick, +30% gold output |
| `colonialism` | Colonialism | optional | 27.3M kp | 2,048 | `cartography`, `gunpowder` | +2.0 food/tick, +30% military power |

---

### Industrial Age (~1h 2m to 1h 18m/tech)

| Key | Name | Kind | Cost | Ticks | Prerequisites | Effect |
|---|---|---|---|---|---|---|
| `steam_power` | Steam Power | spine | 55.3M kp | 1,872 | `steel_forging` | +30% all production |
| `industrialization` | Industrialization | **keystone** (Crystal Palace) | 73.7M kp | 2,340 | `steam_power` | +50% all production, +0.5 steel/tick |
| `railroads` | Railroads | optional | 92.2M kp | 2,340 | `steam_power`, `road_building` | +100% gold output, +200 storage for every resource |
| `rifling` | Rifling | optional | 92.2M kp | 2,340 | `gunpowder` | +50% military power |
| `clockwork_automation` | Clockwork Automation | optional | 92.2M kp | 2,340 | `chronometry` | +10% game speed |

---

### Victorian Age (~1h 10m to 1h 27m/tech)

| Key | Name | Kind | Cost | Ticks | Prerequisites | Effect |
|---|---|---|---|---|---|---|
| `electrification` | Electrification | spine | 165M kp | 2,106 | `industrialization` | +1.0 electricity/tick, +20% all production |
| `telecommunications` | Telecommunications | optional | 275M kp | 2,633 | `electrification` | +40% knowledge output, +50% gold output |
| `mass_production` | Mass Production | **keystone** (Eiffel Tower) | 220M kp | 2,633 | `industrialization` | +40% all production, +1.0 steel/tick |

---

### Electric Age (~1h 18m to 1h 37m/tech)

| Key | Name | Kind | Cost | Ticks | Prerequisites | Effect |
|---|---|---|---|---|---|---|
| `power_distribution` | Power Distribution | **keystone** (Hoover Dam) | 481M kp | 2,925 | `electrification` | +3.0 electricity/tick, +30% all production |
| `radio` | Radio | optional | 601M kp | 2,925 | `telecommunications` | +2.0 culture/tick, +40% knowledge output |
| `chemical_engineering` | Chemical Engineering | spine | 361M kp | 2,340 | `mass_production` | +1.0 oil/tick, +20% all production |

---

### Atomic Age (~1h 33m to 1h 57m/tech)

| Key | Name | Kind | Cost | Ticks | Prerequisites | Effect |
|---|---|---|---|---|---|---|
| `nuclear_fission` | Nuclear Fission | **keystone** (Particle Accelerator) | 449M kp | 3,510 | `power_distribution`, `chemical_engineering` | +5.0 electricity/tick, +0.5 uranium/tick |
| `rocketry` | Rocketry | spine | 337M kp | 2,808 | none | +100% military power, +50% expedition reward |
| `nuclear_deterrence` | Nuclear Deterrence | optional | 562M kp | 3,510 | `nuclear_fission`, `rocketry` | +150% military power |
| `civilian_reactors` | Civilian Reactors | optional | 562M kp | 3,510 | `nuclear_deterrence` | +5.0 electricity/tick, +0.5 uranium/tick; opens the Nuclear Plant |

---

### Modern Age (~1h 33m to 1h 57m/tech)

| Key | Name | Kind | Cost | Ticks | Prerequisites | Effect |
|---|---|---|---|---|---|---|
| `electricity_tech` | Advanced Electrics | spine | 412M kp | 2,808 | `nuclear_fission` | +50% all production, +5.0 electricity/tick |
| `computers` | Computers | spine | 412M kp | 2,808 | `electricity_tech` | +80% knowledge output |
| `satellite_tech` | Satellite Technology | **keystone** (Space Program) | 549M kp | 3,510 | `rocketry`, `electricity_tech` | +1.0 data/tick, +60% knowledge output |
| `nanofabrication` | Nanofabrication | optional | 686M kp | 3,510 | `computers` | −8% build cost |

---

### Information Age (~1h 49m to 2h 16m/tech)

| Key | Name | Kind | Cost | Ticks | Prerequisites | Effect |
|---|---|---|---|---|---|---|
| `internet` | Internet | **keystone** (Global Network) | 582M kp | 4,095 | `computers`, `satellite_tech` | +3.0 data/tick, +120% knowledge output |
| `cybersecurity` | Cybersecurity | spine | 436M kp | 3,276 | `computers` | +100% military power, +5K data storage |
| `social_media` | Social Media | optional | 727M kp | 4,095 | `internet` | +5.0 culture/tick, +5.0 gold/tick |
| `medical_nanobots` | Medical Nanobots | optional | 727M kp | 4,095 | `nanofabrication` | +10 housing, +8.0 food/tick |
| `internet_of_things` | Internet of Things | optional | 727M kp | 4,095 | `social_media`, `cybersecurity`, `medical_nanobots` | +3.0 data/tick, +8.0 food/tick; opens the Smart Farm and the Smart Complex |

---

### Digital Age (~2h 36m/tech)

| Key | Name | Kind | Cost | Ticks | Prerequisites | Effect |
|---|---|---|---|---|---|---|
| `machine_learning` | Machine Learning | **keystone** (World Simulation) | 1.6B kp | 4,680 | `internet`, `cybersecurity` | +5.0 data/tick, +50% all production |
| `cloud_computing` | Cloud Computing | optional | 2B kp | 4,680 | `internet` | +8.0 data/tick, +10K storage for every resource |
| `self_replication` | Self-Replication | optional | 2B kp | 4,680 | `medical_nanobots`, `machine_learning` | +200 nanobots/tick |

---

### Cyberpunk Age (~2h 20m to 2h 55m/tech)

| Key | Name | Kind | Cost | Ticks | Prerequisites | Effect |
|---|---|---|---|---|---|---|
| `neural_interface` | Neural Interface | spine | 1.11B kp | 4,212 | `machine_learning` | +30% worker output, +200% knowledge output |
| `blockchain` | Blockchain | optional | 1.85B kp | 5,265 | `cybersecurity`, `cloud_computing` | +2.0 crypto/tick, +200% gold output |
| `cybernetics` | Cybernetics | **keystone** (Neon Citadel) | 1.48B kp | 5,265 | `neural_interface` | +50% all production, +100% military power |
| `holography` | Holography | optional | 1.85B kp | 5,265 | `cybernetics`, `blockchain` | +5.0 culture/tick, +2.0 crypto/tick; opens the Holographic Theater |

---

### Fusion Age (~2h 36m to 3h 15m/tech)

| Key | Name | Kind | Cost | Ticks | Prerequisites | Effect |
|---|---|---|---|---|---|---|
| `fusion_power` | Fusion Power | **keystone** (Stellar Cradle) | 2.03B kp | 5,850 | `nuclear_fission`, `cybernetics` | +20.0 electricity/tick, +1.0 plasma/tick |
| `plasma_physics` | Plasma Physics | spine | 1.53B kp | 4,680 | `fusion_power` | +3.0 plasma/tick, +30% all production |
| `superconductors` | Superconductors | spine | 1.53B kp | 4,680 | `plasma_physics` | +50% all production, +50K storage for every resource |
| `maglev_transit` | Maglev Transit | optional | 2.54B kp | 5,850 | `superconductors` | +1.0 plasma/tick, +5.0 gold/tick; opens the Energy Exchange |

---

### Space Age (~2h 51m to 3h 34m/tech)

| Key | Name | Kind | Cost | Ticks | Prerequisites | Effect |
|---|---|---|---|---|---|---|
| `orbital_mechanics` | Orbital Mechanics | **keystone** (Dyson Scaffold) | 3.36B kp | 6,435 | `rocketry`, `plasma_physics` | +1.0 titanium/tick, +100% expedition reward |
| `space_mining` | Space Mining | spine | 2.52B kp | 5,148 | `orbital_mechanics` | +3.0 titanium/tick, +20.0 iron/tick |
| `zero_g_manufacturing` | Zero-G Manufacturing | spine | 2.52B kp | 5,148 | `orbital_mechanics`, `superconductors` | +50% all production, +10.0 steel/tick |

---

### Interstellar Age (~3h 7m to 3h 54m/tech)

| Key | Name | Kind | Cost | Ticks | Prerequisites | Effect |
|---|---|---|---|---|---|---|
| `warp_drive` | Warp Drive | **keystone** (Warp Nexus) | 5.23B kp | 7,020 | `space_mining`, `zero_g_manufacturing` | +1.0 dark matter/tick, +200% expedition reward |
| `stellar_engineering` | Stellar Engineering | spine | 3.92B kp | 5,616 | `warp_drive` | +10.0 plasma/tick, +100.0 electricity/tick |

---

### Galactic Age (~3h 7m to 3h 54m/tech)

| Key | Name | Kind | Cost | Ticks | Prerequisites | Effect |
|---|---|---|---|---|---|---|
| `galactic_navigation` | Galactic Navigation | **keystone** (Cosmic Beacon) | 5.23B kp | 7,020 | `warp_drive`, `stellar_engineering` | +50% all production, +5.0 dark matter/tick |
| `antimatter_synthesis` | Antimatter Synthesis | spine | 3.92B kp | 5,616 | `galactic_navigation` | +2.0 antimatter/tick, +30% all production |

---

### Quantum Age (~3h 7m to 3h 54m/tech)

| Key | Name | Kind | Cost | Ticks | Prerequisites | Effect |
|---|---|---|---|---|---|---|
| `quantum_mechanics` | Quantum Mechanics | **keystone** (Reality Anchor) | 3.03B kp | 7,020 | `antimatter_synthesis` | +2.0 quantum flux/tick, +100% all production |
| `reality_manipulation` | Reality Manipulation | spine | 2.27B kp | 5,616 | `quantum_mechanics` | +5.0 quantum flux/tick, +100% all production |
| `quantum_computing` | Quantum Computing | optional | 3.79B kp | 7,020 | `clockwork_automation`, `quantum_mechanics` | **+15% game speed** |

---

### Transcendent Age (~3h 54m)

| Key | Name | Kind | Cost | Ticks | Prerequisites | Effect |
|---|---|---|---|---|---|---|
| `transcendence` | Transcendence | **keystone** (Singularity Core) | 9.1B kp | 7,020 | `reality_manipulation` | +200% all production, +10.0 quantum flux/tick |

---

## Tech Effects Reference

### All production

Each of these adds a percentage to all production. The bonuses from every source (techs, wonders, milestones, monuments, events, boons) add into one pool, and the game multiplies every positive production rate by 1 + that pool, **clamped at x3.0** (+200%). See [The all-production cap](resources.md#the-all-production-cap). The [Cosmic Legacy](prestige.md#cosmic-legacy) is the one all-production bonus outside the pool: it multiplies production by 1.1 after the cap.

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

The techs through the Electric Age add +190%; with the Crystal Palace and Hoover Dam wonders that passes the cap, and Advanced Electrics (Modern) takes the techs past it on their own. By the Transcendent Age they add +950% on paper. Output never goes past x3.0. The Stats panel's Active Multipliers shows what counts and says how much you earned (`+200% capped at +200%: +950% earned`), and the Research panel puts a "capped" note beside every tech the cap holds back, before you buy it and after. From the Modern Age on, an all-production tech adds nothing you can see while you are over the cap; it still counts as a buffer, because a penalty (such as the Reconstruction Effort after you Endure) comes out of the raw pool first.

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

Knowledge output bonuses have a pool of their own (shared with the Great Library and Global Network wonders and scholar milestones), clamped at +200% like all production. The techs through Radio add +195%, so with the Great Library you are at the cap by about the Electric Age, and Computers, Satellite Technology, Internet and Neural Interface add nothing you can see unless something pulls the pool down.

---

### Game speed

Three techs raise game speed: ticks come more often, so production, construction, research and every timer run faster in real time. Game speed isn't part of the all-production pool, so its cap doesn't apply. The bonuses add up with the game's other game speed sources:

| Source | Bonus | Lasts |
|---|---|---|
| Chronometry (Medieval) | +5% | the run |
| Clockwork Automation (Industrial) | +10% | the run |
| Quantum Computing (Quantum) | +15% | the run |
| Milestone chain boosts | +250% or +300% | a few minutes, once per chain |
| Time Dilation boons | +8% to +15%, before scaling | 1,950 to 3,900 ticks |

With all three techs the game runs 30% faster, and that multiplies with research speed. See [Milestones](milestones.md) and [Boons](factions.md#boons). [Era Mastery](prestige.md#era-mastery) is a different thing: it leaves the clock alone and makes each tick of a mastered age produce more, with builds and research needing fewer ticks.

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

Each resource's bonuses share a pool clamped at +200%, as knowledge's do. The gold techs from Currency to Mercantilism already add +230%, so from the Colonial Age on Railroads, Telecommunications and Blockchain add no gold you can see. The Research panel says so beside each one (`capped: no effect now`).

Worker output bonuses have no such ceiling. They raise what your workers add to the buildings they staff, which is 80% of a fully staffed building's listed rate, and they are added on top of the other production bonuses. With all four techs (+65%) a fully staffed building makes 52% of its listed rate more. See [Worker output bonuses](workers-and-domains.md#worker-output-bonuses).

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

**Prerequisites stack.** Before typing `research banking`, check that you have both `currency` and `mathematics`. `research list` shows only the techs you can start now; the Research panel (`research`) shows the locked ones dimmed. Or plan the whole chain: `plan research` takes a tech whose prerequisites are planned above it.

**The Dark Age epoch event** cancels your active research and drains 80% of your knowledge stockpile. If an epoch transition is close, consider whether to delay an expensive research start until after its event resolves.

**Prestige resets research** entirely, with every tech, its bonus and the milestones that gave research speed. What survives a prestige is **Ancient Knowledge** (research time ×0.8 per epoch) from Succumbing, faster research times in the ages a past run completed ([Era Mastery](prestige.md#era-mastery)), and, with the legacy kit's [Plan Template](prestige.md#plan-template), the techs you put in your build plan, which are planned again in the age you planned them in.

**Succumbing early is worth considering.** Succumbing to a catastrophe in the Iron Era (the earliest era one can strike in) costs you a run but cuts research time to ×0.8 permanently, and each further epoch you Succumb in multiplies it by 0.8 again. The fallen run's completed ages also gain an [Era Mastery](prestige.md#era-mastery) level, so the rebuild is quicker. Players who Succumb at least once begin each later run with faster research from tick one. See [Succumb](catastrophe.md#succumb).

**Don't overlook `civil_engineering`.** −5% build cost plus +100 storage for every resource is good value in the Classical Age and keeps helping for the rest of the run.

---

*See also: [Knowledge](knowledge.md) for making and buying knowledge; [The Build Plan](plan.md) for queuing research; [Epochs](epochs.md) for how Grand Discovery and the Dark Age event fire; [Prestige](prestige.md) for the Plan Template and the Ancient Civilization Memory; [Buildings](buildings.md) for the buildings techs open.*
