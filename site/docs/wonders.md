# Wonders

22 unique wonders can be built exactly once per run, one in each age. Each grants civilization-wide bonuses that last until you prestige or Succumb, and every age advance needs its age's wonder. From the Stone Age on, each wonder needs one technology first: its keystone.

<figure class="screen" data-screen="wonders"><figcaption>The Wonders panel in the Bronze Age: the bank of Stonehenge part filled, and the wonders already built below it.</figcaption></figure>

---

## How to build a wonder

Wonders require you to **bank resources** before construction begins. Resources in the wonder bank are reserved and cannot be used elsewhere.

```
wonder                               # the bank: each resource, banked / needed
wonder collect <resource> <amount>   # bank that much (never more than it still needs)
wonder collect <resource> all        # as much as it still needs, up to what you have
wonder collect all                   # the same for every resource it needs
build <wonder_key>                   # start construction once the bank is full
```

`build` does not take resources for a wonder from your stock: bank the full price with `wonder collect` first, then `build`.

### The keystone tech

From the Stone Age on, a wonder needs one technology of its own age before it can be built: its **keystone** (Stoneworking for the Great Monolith, Mathematics for the Colosseum; each wonder below lists its own). The bank does not wait for it: deposits and overflow go in from the first tick of the age, so you can fill the bank while the keystone is being researched. Only `build <wonder_key>` is refused until the tech is done, and it names the tech:

```
Great Monolith needs Stoneworking first. Research it to build here.
```

The Wonders panel and `wonder` show the keystone with the bank: `✗ Keystone: Stoneworking, not researched (research stoneworking)` until it is done, then `✓ Keystone: Stoneworking, researched`. A full bank says to research it first when it is still missing, and the log tells you when the research has opened the wonder. The Sacred Grove needs no tech, and Stonehenge has no keystone yet. See [Keystone Techs](ages.md#keystone-techs) for the whole list and what each keystone stands on.

In a game saved before keystones existed, the age you were in is exempt: its wonder builds without the tech, and the lock starts when you next advance.

`wonder bank` is the same command as `wonder collect` (`wonder bank food all`), `max` means the same as `all`, and leaving the amount off (`wonder bank food`) means `all` too. Each deposit says how much went in. When nothing can go in, the command says why: the wonder is already built, it doesn't need that resource (and lists the ones it does), that part of the bank is full, you have none on hand, or a set amount is more than you have (that is refused rather than banked in part; use `all` to bank what you have).

A [build plan](plan.md) can hold the wonder too (`plan build <wonder_key>`): it banks what the wonder still lacks from what you hold as soon as that covers all of it, and starts construction. A planned wonder whose keystone is not researched waits for it (`waits for Stoneworking`) and starts once it is.

### Overflow

**Wonder overflow** is on by default. When a resource the current age's wonder still needs would be clamped at its storage limit, the part the limit would cut off goes into the wonder's bank instead of being lost, up to what the wonder still needs of it. It never takes anything you hold: only what production was about to waste. It works during offline catch-up too. What the wonder doesn't need then goes toward your [build plan](plan.md#overflow-pays-the-plan)'s queued copies, with or without wonder overflow, so the wonder always comes first.

A wonder you have put in your build plan is one of those queued copies. With wonder overflow off it still takes overflow, in plan order, into its own bank, so a planned wonder's bank fills while you are away either way. Turning wonder overflow off only stops a wonder you have not planned from taking it.

```
wonder overflow        # is it on?
wonder overflow off    # only a wonder in your build plan takes overflow
wonder overflow on
```

The log says so when overflow finishes a resource's part of the bank (and when the bank is full), and the welcome-back summary adds up what it banked while you were away. The switch is saved with your game and survives prestige. The Wonders panel shows whether it is on.

The **Wonders** panel (`wonders`) shows progress with a bar for each resource. Each completed wonder there has a small color thumbnail, a 2-character half-block icon taken from the wonder's sprite, so you can tell them apart at a glance. Locked wonders of ages past your next one are counted in one line, not listed, so the panel doesn't name ages you haven't reached.

---

## Viewing Wonders on the Map

Completed wonders appear on the [Map](map.md) (`map` command) as landmarks, each drawn in its era's look. In the roguelike style they stand among your streets, and Tab jumps the cursor between buildings and wonders. In the skyline they stand among the age districts. Point the inspect cursor at one to see its name, the age it belongs to, and its details.

This age's wonder shows up only once you start it: bank some of its cost or queue its construction, and the roguelike style marks its plot with scaffolding (the cursor shows how much is banked). Until then the Map shows nothing for it.

---

## Wonders and Age Advancement

Completing a wonder is **required** to advance to the next age. Each age unlocks one wonder, and `advance` is refused until that wonder is built.

The wonder requirement appears in the **age progress bar** at the top of the screen alongside your other advancement requirements. If the wonder is still missing, you will see a red notice: `✗ Wonder: <name>`, followed by `✗ Keystone: <tech>` while its keystone tech is still to research.

If you try to advance before completing your age's wonder, the game will tell you which wonder is blocking and remind you to research its keystone if that is missing, then to use `wonder collect` and `build <key>`.

Once completed, the wonder appears as a landmark on the Map (`map` command).

---

## What wonders cost

Because every wonder stands between you and the next age, each one costs about the same share of its age's economy: **40 price units** of that age. A price unit is the median price of one resource in that age (the typical first-copy price of the age's buildings in that resource). Each wonder has its own resource mix.

The flow resources (food, faith, culture) aren't priced that way. Those parts are set by hand to what your buildings actually make: the Sacred Grove takes **500 food**, the Great Monolith **1.5K food**, and the Sistine Chapel **20K faith**.

Wonders are banked a deposit at a time, but each part of a wonder's price fits in the most storage you can build in its age, so you never have to bank one resource in several rounds while sitting at its storage limit.

No wonder takes longer to build than a sixth of its age's target length (see [How Long Each Age Takes](ages.md#how-long-each-age-takes)). Build times below are in ticks, with the real time beside them; one tick is 2 seconds. These are first-run times: on known ground [Era Mastery](prestige.md#era-mastery) divides them by the age's speed, and the Wonders panel shows the shorter time. **Architecture**, a Renaissance Age tech, takes a quarter off the time of every wonder you start after it, and the techs that cut all construction then take their share of what is left.

---

## Wonders and the all-production cap

Six wonders add to all production: the Crystal Palace (+15%), Hoover Dam (+20%), Warp Nexus (+80%), Cosmic Beacon (+50%), Reality Anchor (+50%) and Singularity Core (+200%). Every all-production bonus in the game, from wonders, milestones, monuments, events and boons alike, adds into one pool (a tech's bonus and the Cosmic Legacy are applied after it). The pool applies in full up to +200%, and past that every further point counts a quarter (see [The all-production cap](resources.md#the-all-production-cap)). The first five wonders alone come to +215%, and milestones usually take the pool past +200% well before the last of them, so a late wonder adds a quarter of what it lists. An example: with +400% earned, the pool applies +250%. Building the Warp Nexus (+80%) takes it to +480% earned and +270% applied, 20 points more. The Wonders panel shows that figure beside the bonus before you build (`+80% all production (counts a quarter past +200%: +20% now)`). Their flat outputs (dark matter, antimatter, quantum flux and the rest) count in full.

---

## Wonder list

### 🌿 Sacred Grove
**Age:** Primitive · **Key:** `sacred_grove` · **Build:** 75 ticks (2m 30s) · **Keystone:** none

| Resource | Cost |
|---|---|
| Wood | 1K |
| Food | 500 |

**Bonus:** +0.02 knowledge/t · +0.05 food/t

---

### 🗿 Great Monolith
**Age:** Stone · **Key:** `great_monolith` · **Build:** 225 ticks (7m 30s) · **Keystone:** Stoneworking (`stoneworking`)

| Resource | Cost |
|---|---|
| Stone | 6.3K |
| Wood | 5K |
| Food | 1.5K |

**Bonus:** +0.05 knowledge/t · +5K storage for every resource

---

### ⭕ Stonehenge
**Age:** Bronze · **Key:** `stonehenge` · **Build:** 1,170 ticks (39m) · **Keystone:** Calendar

| Resource | Cost |
|---|---|
| Stone | 34K |
| Wood | 19K |
| Iron | 3.4K |

**Bonus:** +0.8 knowledge/t · +0.6 faith/t

---

### 🏟 Colosseum
**Age:** Iron · **Key:** `colosseum` · **Build:** 1,950 ticks (1h 5m) · **Keystone:** Mathematics (`mathematics`)

| Resource | Cost |
|---|---|
| Stone | 320K |
| Iron | 71K |
| Gold | 63K |

**Bonus:** +100 housing · +2.0 culture/t

---

### 🏛 Parthenon
**Age:** Classical · **Key:** `parthenon` · **Build:** 2,500 ticks (1h 23m 20s) · **Keystone:** Philosophy (`philosophy`)

| Resource | Cost |
|---|---|
| Stone | 1M |
| Gold | 440K |
| Iron | 440K |

**Bonus:** +2.0 culture/t · +1.2 knowledge/t

---

### 📚 Great Library
**Age:** Medieval · **Key:** `great_library` · **Build:** 3,510 ticks (1h 57m) · **Keystone:** Theology (`theology`)

| Resource | Cost |
|---|---|
| Stone | 3.5M |
| Gold | 2.7M |
| Knowledge | 840K |

**Bonus:** +2.0 knowledge/t · **+30% knowledge output**

---

### 🎨 Sistine Chapel
**Age:** Renaissance · **Key:** `sistine_chapel` · **Build:** 4,680 ticks (2h 36m) · **Keystone:** Patronage (`patronage`)

| Resource | Cost |
|---|---|
| Stone | 24M |
| Gold | 24M |
| Faith | 20K |
| Culture | 8M |

**Bonus:** +3.5 culture/t · +1.8 faith/t

---

### 🏮 Grand Lighthouse
**Age:** Colonial · **Key:** `grand_lighthouse` · **Build:** 5,460 ticks (3h 2m) · **Keystone:** Cartography (`cartography`)

| Resource | Cost |
|---|---|
| Stone | 160M |
| Gold | 120M |
| Steel | 23M |

**Bonus:** +5.0 gold/t · **+80% expedition reward**, the biggest expedition bonus of any wonder

---

### 🏗 Crystal Palace
**Age:** Industrial · **Key:** `crystal_palace` · **Build:** 6,240 ticks (3h 28m) · **Keystone:** Industrialization (`industrialization`)

| Resource | Cost |
|---|---|
| Steel | 450M |
| Iron | 400M |
| Gold | 310M |
| Coal | 400M |

**Bonus:** **+15% all production** · +8.0 gold/t

---

### 🗼 Eiffel Tower
**Age:** Victorian · **Key:** `eiffel_tower` · **Build:** 7,020 ticks (3h 54m) · **Keystone:** Mass Production (`mass_production`)

| Resource | Cost |
|---|---|
| Steel | 4.7B |
| Iron | 3.8B |
| Gold | 5.1B |

**Bonus:** +5.0 culture/t · +2.0 knowledge/t

---

### 🌊 Hoover Dam
**Age:** Electric · **Key:** `hoover_dam` · **Build:** 7,800 ticks (4h 20m) · **Keystone:** Power Distribution (`power_distribution`)

| Resource | Cost |
|---|---|
| Steel | 71B |
| Stone | 50B |
| Electricity | 20B |

**Bonus:** +10.0 electricity/t · **+20% all production**

---

### ⚛️ Particle Accelerator
**Age:** Atomic · **Key:** `particle_accelerator` · **Build:** 9,360 ticks (5h 12m) · **Keystone:** Nuclear Fission (`nuclear_fission`)

| Resource | Cost |
|---|---|
| Steel | 100B |
| Electricity | 140B |
| Uranium | 16B |

**Bonus:** +10.0 knowledge/t · +1.5 uranium/t

---

### 🚀 Space Program
**Age:** Modern · **Key:** `space_program` · **Build:** 9,360 ticks (5h 12m) · **Keystone:** Satellite Technology (`satellite_tech`)

| Resource | Cost |
|---|---|
| Steel | 770B |
| Gold | 690B |
| Electricity | 430B |
| Knowledge | 600B |

**Bonus:** +6.0 knowledge/t · +8.0 culture/t

---

### 🌐 Global Network
**Age:** Information · **Key:** `global_network` · **Build:** 10,920 ticks (6h 4m) · **Keystone:** Internet (`internet`)

| Resource | Cost |
|---|---|
| Steel | 4.8T |
| Data | 580B |
| Electricity | 960B |
| Gold | 2.4T |

**Bonus:** +30.0 data/t · **+30% knowledge output**

---

### 💻 World Simulation
**Age:** Digital · **Key:** `world_simulation` · **Build:** 12,480 ticks (6h 56m) · **Keystone:** Machine Learning (`machine_learning`)

| Resource | Cost |
|---|---|
| Steel | 34T |
| Data | 1T |
| Electricity | 20T |

**Bonus:** +60.0 data/t · +15.0 knowledge/t

---

### 🌆 Neon Citadel
**Age:** Cyberpunk · **Key:** `neon_citadel` · **Build:** 14,040 ticks (7h 48m) · **Keystone:** Cybernetics (`cybernetics`)

| Resource | Cost |
|---|---|
| Steel | 100T |
| Electricity | 77T |
| Crypto | 12T |
| Data | 6.4T |

**Bonus:** +10.0 crypto/t · **+500 housing**

---

### ☀️ Stellar Cradle
**Age:** Fusion · **Key:** `stellar_cradle` · **Build:** 15,600 ticks (8h 40m) · **Keystone:** Fusion Power (`fusion_power`)

| Resource | Cost |
|---|---|
| Steel | 430T |
| Plasma | 340T |
| Electricity | 460T |

**Bonus:** +15.0 plasma/t · +200.0 electricity/t

---

### 🛰 Dyson Scaffold
**Age:** Space · **Key:** `dyson_scaffold` · **Build:** 17,160 ticks (9h 32m) · **Keystone:** Orbital Mechanics (`orbital_mechanics`)

| Resource | Cost |
|---|---|
| Titanium | 910T |
| Plasma | 660T |
| Steel | 5.5Q |

**Bonus:** +200.0 electricity/t · +30.0 plasma/t

---

### 🌀 Warp Nexus
**Age:** Interstellar · **Key:** `warp_nexus` · **Build:** 18,720 ticks (10h 24m) · **Keystone:** Warp Drive (`warp_drive`)

| Resource | Cost |
|---|---|
| Titanium | 31Q |
| Dark Matter | 2.2Q |
| Plasma | 29Q |

**Bonus:** +8.0 dark matter/t · **+80% all production**

---

### 🌌 Cosmic Beacon
**Age:** Galactic · **Key:** `cosmic_beacon` · **Build:** 18,720 ticks (10h 24m) · **Keystone:** Galactic Navigation (`galactic_navigation`)

| Resource | Cost |
|---|---|
| Dark Matter | 23Q |
| Antimatter | 17Q |
| Titanium | 65Q |

**Bonus:** +10.0 antimatter/t · **+50% all production**

---

### ⚡ Reality Anchor
**Age:** Quantum · **Key:** `reality_anchor` · **Build:** 18,720 ticks (10h 24m) · **Keystone:** Quantum Mechanics (`quantum_mechanics`)

| Resource | Cost |
|---|---|
| Quantum Flux | 24Q |
| Antimatter | 39Q |
| Dark Matter | 44Q |

**Bonus:** +15.0 quantum flux/t · **+50% all production**

---

### ✨ Singularity Core
**Age:** Transcendent · **Key:** `singularity_core` · **Build:** 18,720 ticks (10h 24m) · **Keystone:** Transcendence (`transcendence`)

| Resource | Cost |
|---|---|
| Quantum Flux | 360Q |
| Antimatter | 370Q |
| Dark Matter | 370Q |

**Bonus:** **+200% all production** · +20.0 quantum flux/t

> The Singularity Core is the last wonder and has the biggest bill in the game. Its +200% all production is the largest bonus in the game. By the Transcendent Age the pool is far past +200%, where a bonus counts a quarter, so it adds 50 points to what the pool applies.

---

## Wonders and Morale

Every wonder you build raises your **morale cap** by 5 points. Morale is a civilization-wide multiplier on production. It sits on a continuous curve around 50%: above 50% you get a bonus that grows the higher morale climbs, up to **+20%** at the cap, and below 50% a penalty. The base cap with no wonders is **100%**.

- 5 wonders built: morale cap **125%**
- 10 wonders built: morale cap **150%**
- 22 wonders built: morale cap **210%**

These figures are the morale **ceiling**, how high the morale percentage can climb, not an output multiplier. The bonus still tops out at +20%, reached at the cap. A higher cap spreads that same bonus over a wider range of morale, so you need more morale to reach the full +20%. Morale is raised with worship and culture buildings, good events and age advances.

---

## Tips

- Bank resources over several deposits as they come in, rather than waiting to cover the whole price at once. Overflow does some of this for you.
- **Great Library** and **Global Network** each give +30% knowledge output, and the two stack.
- **Grand Lighthouse** and the Rocketry tech both raise expedition rewards, and they stack.
- Build each age's wonder early. It is required for `advance`, and its bonuses last for the rest of the run.
- **Crystal Palace** (+15% all production) lands in the Industrial Age, before the all-production pool usually reaches its cap, so it counts in full.
- Each wonder also raises the morale cap by 5 points (100% base, 210% with all 22). See [Morale](morale.md).
