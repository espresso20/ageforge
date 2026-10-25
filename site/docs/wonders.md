# Wonders

22 unique wonders can be built exactly once per run, one in each age. Each grants bonuses to your whole civilization that last until you prestige or Succumb, and every advance to the next age needs the wonder of the age you are in. From the Stone Age on, each wonder also needs one technology first: its keystone.

<figure class="screen" data-screen="wonders"><figcaption>The Wonders panel in the Bronze Age: the bank of Stonehenge part filled, and the wonders already built below it.</figcaption></figure>

---

## How to build a wonder

A wonder is paid for through its bank. You deposit resources into the bank first, and you can start construction once the bank holds the full price. Banked resources are reserved for the wonder and cannot be used elsewhere.

```
wonder                               # the bank: each resource, banked / needed
wonder collect <resource> <amount>   # bank that much (never more than it still needs)
wonder collect <resource> all        # as much as it still needs, up to what you have
wonder collect all                   # the same for every resource it needs
build <wonder_key>                   # start construction once the bank is full
```

`build` does not take resources for a wonder from your stock: bank the full price with `wonder collect` first, then `build`.

`wonder bank` is the same command as `wonder collect` (`wonder bank food all`), `max` means the same as `all`, and leaving the amount off (`wonder bank food`) means `all` too. Each deposit says how much went in. When nothing can go in, the command says why: the wonder is already built, it doesn't need that resource (and lists the ones it does), that part of the bank is full, you have none on hand, or a set amount is more than you have. A set amount you cannot cover is refused rather than banked in part, so use `all` to bank what you have.

A [build plan](plan.md) can hold the wonder too (`plan build <wonder_key>`): it banks what the wonder still lacks from what you hold as soon as that covers all of it, and starts construction. A planned wonder whose keystone is not researched waits for it (`waits for Stoneworking`) and starts once it is.

### The keystone tech

From the Stone Age on, a wonder needs one technology of its own age before it can be built: its keystone (Stoneworking for the Great Monolith, Mathematics for the Colosseum; each wonder in the list below names its own). The bank does not wait for it. Deposits go in from the first tick of the age, so you can fill the bank while the keystone is being researched. Only `build <wonder_key>` is refused until the tech is done, and it names the tech:

```
Great Monolith needs Stoneworking first. Research it to build here.
```

The Wonders panel and `wonder` show the keystone with the bank: `✗ Keystone: Stoneworking, not researched (research stoneworking)` until it is done, then `✓ Keystone: Stoneworking, researched`. A full bank says to research the keystone first when it is still missing, and the log tells you when the research has opened the wonder. The Sacred Grove needs no tech. See [Keystone Techs](ages.md#keystone-techs) for the whole list and what each keystone stands on.

In a game saved before keystones existed, the age you were in is exempt: its wonder builds without the tech, and the lock starts when you next advance.

### No overflow

Stock has no limit, so nothing is ever cut off, and nothing flows into a wonder's bank on its own. The bank moves only when you deposit into it (`wonder collect`), or when a wonder in your build plan pays what its bank lacks from what you hold. When the last part is in, the log says so once and gives the next step, for example `Great Monolith is fully banked. Research Stoneworking, then type 'build great_monolith' to start construction.`

### The Wonders panel

The Wonders panel (`wonders`) opens on this age's wonder: its effects, its keystone line, and a bar for each resource in the bank with what is banked and what it needs. Below it, the list runs through the wonders in age order:

- A wonder you have built shows a small color icon (a 2-character half-block sprite of the wonder), its name, its age, its description and its effects.
- This age's wonder shows the share of its price that is banked, and says when the bank is full and the wonder is ready to build or waiting for its keystone.
- A wonder of your next age is listed as locked, with the age but not the wonder's name. Wonders of ages beyond that are counted in one line, so the panel never names an age you have not reached.

The share is weighted by what each resource is worth in the age's prices, so a resource named in huge amounts does not count for more than a dear one. The Ledger, the Advance panel and the [Map](map.md) show the same share. `↑` `↓` scroll the panel, `PgUp` `PgDn` page through it, and `Esc` closes it. The panel has letter keys of its own, so it keeps the keyboard: press `Esc` before you type a command. Once you have built a wonder, the panel's first lines say that `G` opens the gallery.

### The gallery

Press `G` (or `Tab`) in the Wonders panel for the gallery: a picture of every wonder you have built, one at a time, drawn the way the skyline map draws it and as large as the panel. A line above the picture counts the wonders you have built (`2 of 2 built`). Under the picture are the wonder's name, its age, the year that age began and its description. The gallery opens on the newest wonder. `←` `→` (or `↑` `↓`, `PgUp` `PgDn`, `Option+↑` `Option+↓`) step through the wonders in age order. `Home` and `End` (or `Option+←` and `Option+→`) go to the first and the last. `G` or `Tab` goes back to the list, and so does `Esc` (a second `Esc` closes the panel). A wonder you have not built is not in the gallery. It works in every glyph set and theme, at 80 columns by 24 rows and larger.

<figure class="screen" data-screen="wonders-gallery"><figcaption>The gallery in the Bronze Age: a wonder built earlier, drawn at dusk with its lamps lit.</figcaption></figure>

The game does not record the day a wonder was finished, so the gallery gives the year its age began.

---

## Viewing wonders on the Map

Built wonders appear on the [Map](map.md) (`map` command) as landmarks, each drawn in its era's look. The roguelike style puts every wonder you have built on its own plot among your streets, and `Tab` jumps the cursor between buildings and wonders. The skyline shows one wonder at a time: this age's, and none from the ages behind you (the [gallery](#the-gallery) has those). Point the inspect cursor at a wonder to see its name, the age it belongs to and its details. `Enter` then types `wonder collect all` while the wonder is rising, and `wonders` otherwise.

This age's wonder shows up only once you start it: bank some of its cost or queue its construction, and the Map marks its plot with scaffolding. On the skyline the scaffolding climbs as the bank fills, and in the roguelike style the cursor shows the percent banked. Until then the Map shows nothing for it.

---

## Wonders and age advancement

You need the wonder of your age to advance to the next one. Each age unlocks one wonder, and `advance` is refused until it is built. The refusal names the wonder that is blocking and says what to do, including researching the keystone when it is missing:

```
Build the Great Monolith wonder before advancing: research Stoneworking, its keystone, bank its cost with 'wonder collect', then type 'build great_monolith'. Type 'charter' to see everything the advance needs.
```

The Ledger at the top of the dashboard lists the wonder among what the advance is still waiting on: the keystone tech, the wonder's bank with its percent, and the wonder itself. The [Advance panel](advance.md) (`charter`) sorts the same rows into its pillars: the keystone leads Knowledge, and the bank and the wonder are the Monument's.

---

## What wonders cost

Every wonder stands between you and the next age, so each costs about the same share of its age's economy: about 45 price units of that age. A price unit is the median price of one resource in that age (the typical first-copy price of the age's buildings in that resource). Each wonder has its own mix of resources. A wonder keeps its authored price while a new copy of an ordinary building costs 1.43 times its authored price, which is why it counts as about 45 units and not the 64 it was authored at.

From the Modern Age on, the age price level rises by 1.4 times per age. Each wonder follows that price level, and its construction time follows the same curve, so its share of the age stays the same.

In the ages that make a raw material, a wonder pays half of its stone, iron, steel or electricity part in the raw instead, one for one: marble, iron ore, titanium ore or dark matter crystals. The Colosseum, the Parthenon, the Crystal Palace, the Space Program, the Global Network, the World Simulation, the Neon Citadel and the Stellar Cradle are paid that way, and the wonder's total in price units is unchanged. See [What the ores are for](resources.md#what-the-ores-are-for).

The flow resources (food, faith, culture) aren't priced that way. Those parts are set by hand to what your buildings actually make: the Sacred Grove takes 800 food, the Great Monolith 2.4K food, and the Sistine Chapel 32K faith and 960K culture.

A wonder is paid through its bank, a deposit at a time. Stock has no limit (see [No storage wall](buildings.md#no-storage-wall)), so the whole price can sit in your stock until you bank it.

Build times in the list above are in ticks, with the real time beside them at the base tick of 2 seconds. They are first-run times: on known ground [Era Mastery](prestige.md#era-mastery) divides them by the age's speed. Architecture, a Renaissance Age tech, takes a quarter off the time of every wonder you start after it, and the techs that cut all construction time then take their share of what is left. When you start a wonder, the log says how long it will take.

---

## Wonders and the all-production pool

Six wonders add to all production: the Crystal Palace (+15%), Hoover Dam (+20%), Warp Nexus (+80%), Cosmic Beacon (+50%), Reality Anchor (+50%) and Singularity Core (+200%). Every all-production bonus in the game, from wonders, milestones, monuments, events and boons alike, adds into one pool (a tech's bonus and the Cosmic Legacy are applied after it). The pool applies in full, however large it grows (see [The all-production pool](resources.md#the-all-production-pool)), so a late wonder adds all of what it lists. For example, with +400% earned, building the Warp Nexus (+80%) takes the pool to +480%, which is 80 points more.

---

## Wonder list

Amounts are the bank's price, as the Wonders panel shows it. They use the game's units: K, M, B and T, then Q (quadrillion) and Qi (quintillion). See [Resources](resources.md).

### 🌿 Sacred Grove
Age: Primitive Age · Key: `sacred_grove` · Build: 150 ticks (5m) · Keystone: none

| Resource | Cost |
|---|---|
| Wood | 1.6K |
| Food | 800 |

Bonus: speeds research · +0.05 food/tick

---

### 🗿 Great Monolith
Age: Stone Age · Key: `great_monolith` · Build: 450 ticks (15m) · Keystone: Stoneworking (`stoneworking`)

| Resource | Cost |
|---|---|
| Stone | 10.1K |
| Wood | 8K |
| Food | 2.4K |

Bonus: speeds research

---

### ⭕ Stonehenge
Age: Bronze Age · Key: `stonehenge` · Build: 2340 ticks (1h 18m) · Keystone: Calendar (`calendar`)

| Resource | Cost |
|---|---|
| Stone | 54.4K |
| Wood | 30.4K |
| Iron | 5.44K |

Bonus: speeds research · +0.6 faith/tick

---

### 🏟 Colosseum
Age: Iron Age · Key: `colosseum` · Build: 3900 ticks (2h 10m) · Keystone: Mathematics (`mathematics`)

| Resource | Cost |
|---|---|
| Stone | 256K |
| Marble | 256K |
| Iron | 56.8K |
| Iron Ore | 56.8K |
| Gold | 101K |

Bonus: +100 housing · +2 culture/tick

---

### 🏛 Parthenon
Age: Classical Age · Key: `parthenon` · Build: 5000 ticks (2h 46m) · Keystone: Philosophy (`philosophy`)

| Resource | Cost |
|---|---|
| Stone | 800K |
| Marble | 800K |
| Gold | 704K |
| Iron | 352K |
| Iron Ore | 352K |

Bonus: +2 culture/tick · speeds research

---

### 📚 Great Library
Age: Medieval Age · Key: `great_library` · Build: 7020 ticks (3h 54m) · Keystone: Theology (`theology`)

| Resource | Cost |
|---|---|
| Stone | 10.5M |
| Gold | 8.07M |

Bonus: speeds research · +30% scholar output

---

### 🎨 Sistine Chapel
Age: Renaissance Age · Key: `sistine_chapel` · Build: 9360 ticks (5h 12m) · Keystone: Patronage (`patronage`)

| Resource | Cost |
|---|---|
| Stone | 38.4M |
| Gold | 38.4M |
| Faith | 32K |
| Culture | 960K |

Bonus: +3.5 culture/tick · +1.8 faith/tick

---

### 🏮 Grand Lighthouse
Age: Colonial Age · Key: `grand_lighthouse` · Build: 10920 ticks (6h 4m) · Keystone: Cartography (`cartography`)

| Resource | Cost |
|---|---|
| Stone | 320M |
| Gold | 192M |
| Steel | 36.8M |

Bonus: +5 gold/tick · +80% expedition rewards (the only wonder that raises them)

---

### 🏗 Crystal Palace
Age: Industrial Age · Key: `crystal_palace` · Build: 12480 ticks (6h 56m) · Keystone: Industrialization (`industrialization`)

| Resource | Cost |
|---|---|
| Steel | 720M |
| Iron | 320M |
| Iron Ore | 320M |
| Gold | 496M |
| Coal | 640M |

Bonus: +15% all production · +8 gold/tick

---

### 🗼 Eiffel Tower
Age: Victorian Age · Key: `eiffel_tower` · Build: 14040 ticks (7h 48m) · Keystone: Mass Production (`mass_production`)

| Resource | Cost |
|---|---|
| Steel | 7.52B |
| Iron | 6.08B |
| Gold | 8.16B |

Bonus: +5 culture/tick · speeds research

---

### 🌊 Hoover Dam
Age: Electric Age · Key: `hoover_dam` · Build: 15600 ticks (8h 40m) · Keystone: Power Distribution (`power_distribution`)

| Resource | Cost |
|---|---|
| Steel | 114B |
| Electricity | 32B |

Bonus: +10 electricity/tick · +20% all production

---

### ⚛️ Particle Accelerator
Age: Atomic Age · Key: `particle_accelerator` · Build: 18720 ticks (10h 24m) · Keystone: Nuclear Fission (`nuclear_fission`)

| Resource | Cost |
|---|---|
| Steel | 186B |
| Electricity | 260B |
| Uranium | 29.8B |

Bonus: speeds research · +1.5 uranium/tick

---

### 🚀 Space Program
Age: Modern Age · Key: `space_program` · Build: 18720 ticks (10h 24m) · Keystone: Satellite Technology (`satellite_tech`)

| Resource | Cost |
|---|---|
| Steel | 660B |
| Titanium Ore | 660B |
| Gold | 1.19T |
| Electricity | 738B |

Bonus: speeds research · +8 culture/tick

---

### 🌐 Global Network
Age: Information Age · Key: `global_network` · Build: 30576 ticks (16h 59m) · Keystone: Internet (`internet`)

| Resource | Cost |
|---|---|
| Steel | 5.38T |
| Titanium Ore | 5.38T |
| Data | 1.3T |
| Electricity | 2.15T |
| Gold | 5.38T |

Bonus: +30 data/tick · +30% scholar output

---

### 💻 World Simulation
Age: Digital Age · Key: `world_simulation` · Build: 48922 ticks (1d 3h) · Keystone: Machine Learning (`machine_learning`)

| Resource | Cost |
|---|---|
| Steel | 53.3T |
| Titanium Ore | 53.3T |
| Data | 3.14T |
| Electricity | 62.7T |

Bonus: +60 data/tick · speeds research

---

### 🌆 Neon Citadel
Age: Cyberpunk Age · Key: `neon_citadel` · Build: 77052 ticks (1d 18h) · Keystone: Cybernetics (`cybernetics`)

| Resource | Cost |
|---|---|
| Steel | 439T |
| Electricity | 169T |
| Dark Matter Crystals | 169T |
| Crypto | 52.7T |
| Data | 28.1T |

Bonus: +10 crypto/tick · +500 housing

---

### ☀️ Stellar Cradle
Age: Fusion Age · Key: `stellar_cradle` · Build: 119858 ticks (2d 18h) · Keystone: Fusion Power (`fusion_power`)

| Resource | Cost |
|---|---|
| Steel | 2.64Q |
| Plasma | 2.09Q |
| Electricity | 1.41Q |
| Dark Matter Crystals | 1.41Q |

Bonus: +15 plasma/tick · +200 electricity/tick

---

### 🛰 Dyson Scaffold
Age: Space Age · Key: `dyson_scaffold` · Build: 184582 ticks (4d 6h) · Keystone: Orbital Mechanics (`orbital_mechanics`)

| Resource | Cost |
|---|---|
| Titanium | 7.83Q |
| Plasma | 5.68Q |
| Steel | 47.3Q |

Bonus: +200 electricity/tick · +30 plasma/tick

---

### 🌀 Warp Nexus
Age: Interstellar Age · Key: `warp_nexus` · Build: 281906 ticks (6d 12h) · Keystone: Warp Drive (`warp_drive`)

| Resource | Cost |
|---|---|
| Titanium | 373Q |
| Dark Matter | 26.5Q |
| Plasma | 349Q |

Bonus: +8 dark matter/tick · +80% all production

---

### 🌌 Cosmic Beacon
Age: Galactic Age · Key: `cosmic_beacon` · Build: 394668 ticks (9d 3h) · Keystone: Galactic Navigation (`galactic_navigation`)

| Resource | Cost |
|---|---|
| Dark Matter | 388Q |
| Antimatter | 287Q |
| Titanium | 1.1Qi |

Bonus: +10 antimatter/tick · +50% all production

---

### ⚡ Reality Anchor
Age: Quantum Age · Key: `reality_anchor` · Build: 552536 ticks (12d 18h) · Keystone: Quantum Mechanics (`quantum_mechanics`)

| Resource | Cost |
|---|---|
| Quantum Flux | 826Q |
| Antimatter | 1.34Qi |
| Dark Matter | 1.52Qi |

Bonus: +15 quantum flux/tick · +50% all production

---

### ✨ Singularity Core
Age: Transcendent Age · Key: `singularity_core` · Build: 773550 ticks (17d 21h) · Keystone: Transcendence (`transcendence`)

| Resource | Cost |
|---|---|
| Quantum Flux | 11.9Qi |
| Antimatter | 12.2Qi |
| Dark Matter | 12.2Qi |

Bonus: +200% all production · +20 quantum flux/tick

> The Singularity Core is the last wonder, and it has the biggest numbers. Its +200% all production is the largest all-production bonus of any wonder, and it adds all 200 points to the pool. The Transcendent Age has no makers of its own, so the Singularity Core is paid from the Quantum Age's makers (see [The Maker Rule](buildings.md#the-maker-rule)).

---

## Wonders and morale

Every wonder you build raises your morale ceiling by 5 points. Morale multiplies production on a continuous curve around 50%: above 50% you get a bonus that grows the higher morale climbs, up to +20% at the ceiling, and below 50% you take a penalty. With no wonders the ceiling is 100%. Three techs add 5 points each: Priesthood (Iron Age), Virtual Reality (Digital Age) and the Overview Effect (Space Age).

- 5 wonders built: ceiling 125%
- 10 wonders built: ceiling 150%
- 22 wonders built: ceiling 210%, or 225% with the three techs

These figures are the morale ceiling, how high the morale percentage can climb, not an output multiplier. The bonus still tops out at +20%, reached at the ceiling. A higher ceiling spreads that same bonus over a wider range of morale, so you need more morale to reach the full +20%. Morale is raised with worship and culture buildings, good events and age advances. See [Morale](morale.md).

---

## Tips

- Bank resources over several deposits as they come in, rather than waiting to cover the whole price at once.
- Great Library and Global Network each give +30% scholar output, and the two stack.
- Grand Lighthouse and the Rocketry tech both raise expedition rewards, and they stack.
- Build each age's wonder early. It is required for `advance`, and its bonuses last for the rest of the run.
- Every wonder raises the morale ceiling by 5 points (100% base, 210% with all 22, 225% with the three morale techs as well).
