# Technologies

Research gives your civilization lasting bonuses. 203 technologies span all 22 ages, and each one changes your production, military power, housing, building costs or times, research time, a number in one of the game's mechanics, or what you can build or do, for the rest of the run. A tech's bonus is small and counts in full: no cap holds it back (see [How tech bonuses stack](#how-tech-bonuses-stack)). Only one technology is researched at a time, but the [build plan](plan.md) can queue the next ones and start each as soon as the slot frees up.

<figure class="screen" data-screen="research"><figcaption>The Research panel in the Classical Age: the tech tree as a map. Philosophy is part way through, Civil Engineering can start, and the Medieval Age waits below the dotted line.</figcaption></figure>

---

## How research works

### Starting research

Starting a tech costs nothing. It needs its age and its prerequisites, and the research slot must be free. Once started, your [scholars](knowledge.md) fill its bar with the work they do each tick, and the tech is done when the bar reads 100%. Its effects apply at once and last for the run. If you cancel it, the work done on it is lost.

A tech has a price, the work it needs, and a time, which is what is left of the price over the work the researching scholars do each tick:

```
time = work still to do ÷ work a tick
```

A tick is 2 seconds of real time at base; game speed makes each tick shorter (see [Game speed](#game-speed)). The time is worked out from the scholars researching now, so it changes when they do: a new knowledge building, a bonus, or a change to the research share with `scholars <percent>` changes the time left on the tech in progress and the time every other tech shows. The times on the tree, in `research list` and on a tech's card are the ones your scholars would take if you started it now. With no scholar researching, they say the tech needs scholars researching.

The work a tick is:

```
work a tick = researching scholars' work × output bonuses × Era Mastery speed ÷ (1 − research speed) ÷ the techs' cut ÷ 0.8 ^ epochs succumbed in
```

The scholars' work is the research of your knowledge buildings (each copy at full rate up to the building's full-rate count, then at half rate), shared out by how many of your scholars research. Output bonuses include morale, the all-production and knowledge-output pools, and the techs that raise research output. The three divisors are cuts of time, and between them they never bring a tech under 5% of its time.

- Ancient Knowledge: for each distinct epoch you have [Succumbed](catastrophe.md#ancient-knowledge) in, the time is multiplied by 0.8: ×0.8 after one epoch, ×0.64 after two, ×0.26 with all six. The Stats panel has a line for it (`Ancient Knowledge: research time ×0.64. The times on the tech tree include it.`).
- Era Mastery: on known ground (an age a past run completed) the scholars' work is multiplied by the age's [Era Mastery](prestige.md#era-mastery) speed: 2x after one completion, up to 4.2x after ten. A tech's card shows the shortened time.
- Research techs: thirteen techs cut research time. The four Knowledge and Computing capstones (Scholasticism, Big Science, General AI and the Timeless Archive) cut it by 6% each, Scientific Method by 4%, and Printing Press, Public Education, Computers, Search Engines, Machine Learning, Unified Theory, Xenology and Cosmology by 3% each. Their cuts multiply (×0.587 with all thirteen), and research never takes less than 50% of its time on their account. The Stats panel has a line for it (`Research techs: research time ×0.94. The times on the tech tree include it.`).

### The bar and the time

The tech in progress shows a bar that fills to 100%, its percent and the time left, in the header of the Research panel, on its card, in `research list` and in the Logs panel. The Scholars tab of the People panel shows its percent and the time left. The Stats panel says what your research speed does (`Research speed +30%: techs take 70% of their base time.`).

Events, loot, milestone rewards, the Ancient Cache and two trade routes can add work to the tech in progress, at most half its price. See [Scholars and Research](knowledge.md).

### Keystones, the spine and the rest

Every tech has a kind, and its kind says what it is for.

- A keystone is the tech an age's wonder needs before it can be built. There is one in each age from the Stone Age on. The next age needs the wonder, so the keystone is the one tech an age asks of you, together with the techs it stands on. See [Keystone Techs](ages.md#keystone-techs) for the list.
- The spine is every tech a keystone stands on, all the way down: Tool Making under Stoneworking, Steam Power under Industrialization. Keystones and the spine are the only techs a run has to research, 48 of the 203.
- A capstone ends a lane for its era: a bigger step that is one of its age's two dearest. No run has to research one. Each era from the Iron Era on has two, twelve in all: Scholasticism and Guilds in the Medieval Age, the Concert of Nations and Interchangeable Parts in the Industrial Age, Big Science and the Military-Industrial Complex in the Atomic Age, the Global Village and General AI in the Digital Age, Stellar Cartography and the Space Elevator in the Space Age, the Timeless Archive and Reality Engineering in the Quantum Age.
- Everything else is optional: yours to take or leave.

The tech tree draws each keystone with a ★ on a double frame and each capstone in half blocks. A keystone's card names the wonder that waits for it, and `research list` marks it too (`★ keystone: Colosseum`).

### How research is priced

A tech's price follows its age's ladder: inside an age each tech costs 1.3 times the one before it, from the cheapest to the dearest. The tables below give every tech's time for a full set of scholars at its age: every knowledge building up to that age built to its full-rate count, all researching, before any bonus. Your own times are the ones the tree shows. The Transcendent Age has no knowledge building of its own, so its techs are researched by the scholars of the ages before it.

Research simmers. From the Stone Age on, the techs of an age together ask for 1.2 times the work that the scholars of a typical town do while the town stays in that age. Researching all of an age takes more work than those scholars do in it, and what you have not finished carries forward into the ages after it. It is an idle game: you can get every tech, but it takes some time. The Primitive Age is the exception. Its techs together ask for 0.8 times that work, so a new player sees a tech finish early. The keystone is the other exception. It is priced to finish inside its age, because its wonder needs it before you can leave.

### One slot, and the plan as a queue

There is one research slot. If you type `research` for a second tech while one is in progress, you get an error naming the active tech.

To line techs up, add them to the [build plan](plan.md) with `plan research <tech>` (or `plan res`). The plan starts each one as soon as the slot is free, in plan order, and it keeps doing so while you are away:

- Only the first research item in the plan can take the slot when it frees up. A tech further down waits its turn even if it is cheaper.
- A tech brings what it needs. `plan research imperial_legions` adds every prerequisite that is not researched, in progress or planned yet, each after its own, and then Imperial Legions. One command queues a whole chain toward a goal, and the answer lists what went in, in order. For an either-or group the plan takes the branch you already hold or have planned, else the one that costs the least work to add.
- The plan takes a tech of any age you can see on your [tech tree](#reading-the-tech-tree): up to the next age on a first run, and every age your account has reached on a later one. A tech of a later age waits for its age (`waits for the Iron Age`) without holding up this age's techs below it.
- A tech already in the plan that loses a prerequisite (you removed it, moved it below, canceled it mid-research, or a game update changed what the tech needs) is not dropped: it waits, and the plan shows what it needs (`needs Philosophy first`). While it waits it does not hold the slot's turn, so you can plan the missing tech below it and that one starts first.
- On the tree, a planned tech shows its place in the plan (`▸3`), and the second `Enter` on a locked tech's card plans it with its chain.

The plan is the only research queue, and the game never chooses what you research next. The techs you plan are remembered with the rest of your plan: [Plan Template](prestige.md#plan-template) plans them again on later runs after your first prestige, in the age you planned them in. Techs you start by hand with `research` are not remembered.

See [The research queue](plan.md#the-research-queue) and [How it runs](plan.md#how-it-runs) for the rest of the plan's rules.

### When research completes

The log says `Research complete: <tech>.` and the tech's effects apply at once. A tech that opens a command says so (`Drama opens festivals.`), and a keystone says which wonder it opens.

---

## Research speed sources

Research speed cuts the time of every tech, the one in progress included. Its sources add together. A tech takes this long:

```
ticks = base × (1 − research speed) × the techs' cut × Ancient Knowledge
```

| Source | Research speed | Notes |
|---|---|---|
| Tech Pioneer milestone (research 22 techs) | +5% | In the Scholar Chain |
| Philosophes milestone (64 techs, from the Classical Age) | +5% | |
| Renaissance Mind milestone (83 techs, from the Renaissance Age) | +10% | In the Scholar Chain; hidden |
| Tech Master milestone (114 techs, from the Information Age) | +10% | In the Scholar Chain; hidden; also +5% all production |
| Tech Ascendant milestone (all 203 techs, Transcendent Age) | +20% | Hidden. It arrives with your last tech, so it never shortens one |
| Ancient Knowledge (Succumb) | not research speed: research time ×0.8 per epoch | For each distinct epoch you Succumb in (Iron to Cosmic, ×0.26 with all six). It multiplies the time research speed leaves, so it is not in this pool and no cap holds it. Kept through Succumb, prestige and save/load. See [Ancient Knowledge](catastrophe.md#ancient-knowledge) |
| Thirteen techs, from Scholasticism to the Timeless Archive (see [Building costs, construction time and research time](#building-costs-construction-time-and-research-time)) | not research speed: research time ×0.94 to ×0.97 each | They multiply the time research speed leaves, ×0.587 with all thirteen, and never take it under 50% between them. They are not in this pool |

Milestone research speed lasts for the run: milestones start over at prestige and at Succumb. Five milestones add +50% in total, +30% of it before your last tech. Hidden milestones show once you are well on the way to them. See [Milestones](milestones.md).

Research speed counts up to +100%, because a tech always takes a tick at least. The sources above come to +50% at most.

> Research speed is not research output. Fifteen techs raise research output, the work your scholars do. Their effect lists call it scholar output (see [Output of one resource](#output-of-one-resource) below). More work lets the tech in progress fill its bar sooner; research speed cuts the time the same work takes. Both shorten the time on the screens. A civilization's Enlightenment boon is worded as research speed, but it raises the work your scholars do, so it counts as output.

---

## Commands

```
research <tech_key>
```
Start researching a technology. It costs nothing. Fails if: unknown key, already researched, another tech in progress, age requirement not met, or prerequisites missing.

You can also type multi-word tech names with spaces; they are joined with `_`:
```
research bronze working
```
is equivalent to `research bronze_working`.

---

```
plan research <tech_key>
```
Adds a tech to the build plan, with the techs it still needs before it. The plan starts each when the research slot is free (see [One slot, and the plan as a queue](#one-slot-and-the-plan-as-a-queue)). It takes any tech on your tree, the next age's included. `plan res` is the short form.

---

```
research list
```
Prints the techs you can start now (age reached, prerequisites done, not yet researched) with their keys and the time each would take your scholars now (`~4m 44s`, or that it needs scholars researching), then the active research with its bar and the time left, the queue in order, and your scholars. See [Timers and durations](commands.md#timers-and-durations).

---

```
research cancel
```
Cancels the current research. The work done on it is lost.

---

```
research
```
With no arguments, opens the Research panel, the tech tree as a map (so do `techs` and `research tree`). See [Reading the tech tree](#reading-the-tech-tree).

---

```
research tree close
research tree far
```
Opens the tree at that zoom: big badges with the tech's name over each (close, the default), or one-line pills so more lanes fit on screen (far). `PgUp` and `PgDn` (`Option+↑` and `Option+↓` on a Mac keyboard) do the same from inside the panel.

---

```
research card <tech_key>
```
Opens the tree on that tech's card. It works for any tech you can see on the map, and never for one in an age still hidden.

---

Shortcut: `res` is an alias for `research`. All subcommands work identically.

---

## Reading the tech tree

`research` opens the tree as a map. It covers the whole screen but the command bar, and it does not have to fit: the view follows the tech you select.

- The top bar counts the techs you know out of those in the ages you have reached (`RESEARCH 23/33`). While a tech is in progress it shows the tech's name, bar, percent and time left, and your scholars on the right.
- Lanes run across, ages run down. Each lane is a column with its name in capitals at the top. From left to right they are Faith & Culture, Knowledge, Trade & Exploration, Agriculture, Craft & Engineering, Materials, Military, Energy, Flight & Space and Computing; a lane appears once one of its techs is in sight. When lanes are off screen, an arrow at that end of the lane row counts them (`2▸`), and a wide terminal names the nearest on the left (`◂ FAITH`). Each age is a band, named in the gutter on the left with how many of its techs you hold.
- A tech is a badge with its name over it. The glyph in the middle is the tech's emblem, or its lane's.
- Lines show what a tech needs. A line leaves the notch under a badge and comes down on the name of the tech that needs it. A solid line is needed outright. Dashed lines marked `or` are an either-or group: one of them will do. Selecting a tech lights the chain that leads to it.
- One line under the map names the selected tech, its state, its time and what it opens. A key bar sits under that.

<figure class="screen" data-screen="research-far"><figcaption>The same tree zoomed out with PgUp: every lane on screen, a tech a line. A tick is researched, round brackets can start, dashed bars wait for something, and a shaded pair is the next age. The two techs led by ▸ are in the build plan: one plan research for Fortification, a tech of the next age, queued Imperial Legions before it.</figcaption></figure>

### What the frames and marks mean

The shape says it as well as the color, so the tree reads in every theme.

| You see | It means |
|---|---|
| A rounded frame `╭─────╮` | An optional tech |
| A double frame `╔═════╗` | A tech on the spine: a keystone stands on it |
| A double frame with a `★` | A keystone: its age's wonder cannot be built without it |
| Half blocks `▄▄▄▄▄▄▄` | A capstone: the long, dear tech that ends a lane for its era. The blocks are broken (`▄ ▄ ▄ ▄`) while it waits |
| A `✓` at the top right, the frame in the lane's color | Researched |
| A `⟳`, and the bottom edge filling as a bar `╚▓▓░░░╝` | Being researched |
| A solid, bright frame and no mark | You can start it |
| `▸1` at the top right | It is in your build plan, at that place |
| A dashed frame `╭┄┄┄┄┄╮` | It waits for a tech it needs |
| A shaded block `░░░░░░░` | A tech of an age you have not reached yet but may see: you can read its card and plan it |
| A `◇` at the bottom right | It opens a command or a building |

Zoomed out, a tech is one line: `✓` before the emblem is researched, `⟳` is in progress, `( )` round it can start, `┆ ┆` waits, `░ ░` is the next age, `▸` before it is in your plan, and a `★` after the letterhead is a keystone.

Ages beyond what you may see are not drawn: the next age on a first run, and further on a later one. One line at the foot of the map counts their techs, and nothing names them: not the map, not a card, not `research card`.

### The card

`Enter` on a tech opens its card over the map. It says what the tech does in plain sentences, what it opens, how long it takes your scholars now, what it builds on (techs you hold in green) and how many later techs build on it.

<figure class="screen" data-screen="research-card"><figcaption>Civil Engineering's card. Research is busy, so Enter would add it to the build plan.</figcaption></figure>

The first `Enter` only opens the card. The last line of the card says what a second `Enter` does: start the tech if it can start now, or add it to your [build plan](plan.md) if the research slot is busy or the tech still waits for something. A tech that needs others goes in with them, in order, and the line names them first (`adds it to your plan, after Military Tactics, which it needs first`). `Esc` closes the card.

### Keys

The panel takes only keys that print nothing. Letters always go to the command bar, so every command works with the tree open.

| Key | On the map | With a card open |
|---|---|---|
| Arrows | Move to the nearest tech that way; the view follows | Nothing |
| `Tab`, `Shift-Tab` | The next or previous tech you can start | Nothing |
| `PgUp`, `PgDn` (`Option+↑`, `Option+↓`) | Zoom out and in | Nothing |
| `Home` | Jump to the tech in progress, or else the first tech you can start in your current age | Nothing |
| `Enter` | Open the selected tech's card | Do what the card says |
| `Esc` | Close the panel | Close the card |

With something typed in the command bar, `Tab` and `Enter` belong to the command, as on the Map.

The glyphs follow your `map glyphs` setting: with `map glyphs ascii` the tree is drawn in plain ASCII, with letters for emblems. What research adds up to (every bonus your techs give together, and what research speed does to the times) is on the Stats panel, under Research bonuses.

---

## How tech bonuses stack

A tech's bonus is applied in a layer of its own, after every other bonus and after the pools.

The game's other bonuses (milestone rewards, wonders, monuments, festivals, events, boons) add up in pools, and a production pool has no ceiling: all production and each resource's own output apply in full, however large the pool grows (see [The all-production pool](resources.md#the-all-production-pool)). So a pool that has earned +320% applies +320%. Techs join no pool. A tech's output bonus multiplies what the pools make, where [Era Mastery](prestige.md#era-mastery) does, so it counts in full whatever else you hold.

```
what a resource makes = buildings and workers × pools (in full, at any size) × (1 + tech bonuses)
```

- Bonuses on the same thing add up. Two +5% techs on the same resource give +10%, not ×1.1025. A tech's bonus on all production adds to its bonuses on one resource: with +20% stone and +5% all production, stone runs at ×1.25.
- Nothing caps the layer. The tree is finite, and all of its techs together come to +99% scholar output, +119% food, +42% steel, +40% gold, +4% crypto and +14.2% all production. The last tech you take does its full step.
- Flat amounts stay amounts. Steel Forging's +0.25 steel a tick and Satellite Technology's +1 data are the first source of a resource, not a bonus, and the layer does not multiply them.
- Cuts multiply, and each has a floor. A tech that cuts a price or a time takes its share of what the cuts before it left: two 3% cuts on building costs leave 0.97 × 0.97. Building costs never fall under 10% of the listed price, construction under 40% of its time, or research under 50% of its time, however many cuts stack. All of today's techs together stay far from every floor.
- Housing is a percentage of the housing you have: +5% housing is 5% more housing, rounded up to a whole person.
- Game speed, military power and expedition rewards are added to the pools of those names, shared with milestones and wonders.
- The `rates` command shows what the techs add to a resource on its Research line, and the Stats panel lists what all your techs come to together, under Research bonuses.

A tech's bonus lasts for the run. Prestige, Succumb and a new game wipe every tech and everything it gave.

---

## Commands a tech opens

Some commands wait for a tech. Until it is researched, the command is refused with the name of the tech that opens it (`Festivals need Drama first. Research it to hold one.`). The tech's effect list says what it opens, and the log says it again when the research finishes (`Drama opens festivals.`).

| Command | Opened by | Age |
|---|---|---|
| Trade routes (`trade route start <route>`) | The Wheel | Bronze |
| Diplomacy: gifts, alliances, rivalries, embargoes and deals (`diplomacy ...`, `plan deal`) | Envoys | Classical |
| Festivals (`festival confirm yes`) | Drama | Classical |
| The black market (`blackmarket <resource>`) | Mercantilism | Colonial |
| The Rail Freight route (`trade route start rail_freight`) | Railroads | Industrial |
| The Warp Commerce route (`trade route start warp_commerce`) | Interstellar Trade | Interstellar |

The Trade and Factions panels say which tech a command waits for, and so does `festival`.

---

## Tech tree by age

Prerequisites are listed by tech key. A tech needs every prerequisite it lists, and one of an either-or group where the row says so. If a game update changes what a tech needs, a save keeps what it holds: a researched tech stays researched, a research in progress finishes, and a tech in your plan waits for what it now needs.

Time is what a tech takes a full set of scholars at its age: every knowledge building up to that age built to its full-rate count, all researching (see [How research is priced](#how-research-is-priced)). It is before any bonus and at 2 seconds a tick; your own times are on the tree. Every effect in the tables counts in full in every age (see [How tech bonuses stack](#how-tech-bonuses-stack)).

Forty buildings wait for a tech of their own age instead of opening with the age. [Buildings a tech opens](buildings.md#buildings-a-tech-opens) lists them all, from the Standing Stones (Ritual) to the Omniversal Bazaar (Omniversal Exchange). A gated building stays hidden from the build list until its tech is done, and `build` and `upgrade` refuse it until then; you can still add it to your build plan, where it waits for the tech. Buildings you already have stay built. The gate out of an age asks for such a building only when you cannot leave the age without its tech (see [Ages](ages.md)), so an optional tech never holds an age shut. A wonder's keystone works the same way for building, but the wonder is listed and its bank is open from the start of the age.

Some of those techs sit deep in their age's chain, so their building comes late in the age: Civilian Reactors (Atomic Age, the Nuclear Plant), Internet of Things (Information Age, the Smart Farm and the Smart Complex), Holography (Cyberpunk Age, the Holographic Theater) and Maglev Transit (Fusion Age, the Energy Exchange). Civilian Reactors stands behind Nuclear Deterrence, Internet of Things behind three Information Age techs, Holography behind Cybernetics and Blockchain, and in the Fusion Age, Plasma Physics, Superconductors and Maglev Transit run one after another.

Exploration is the tree's one tech with two ways in: it needs Map Making or Boatbuilding, whichever you hold. The tree draws both lines dashed with an `or`, and `plan research exploration` takes the one you have or, with neither, the cheaper.

### Primitive Age (~10m to 17m/tech)

| Key | Name | Kind | Time | Prerequisites | Effect |
|---|---|---|---|---|---|
| `language` | Language | spine | 10m | none | +10% scholar output |
| `fire_mastery` | Fire Mastery | optional | 17m | none | +10% food production, +5% housing |
| `tool_making` | Tool Making | spine | 13m | none | +10% food production, +10% wood production, gathering by hand brings 2 more |

---

### Stone Age (~8m 14s to 30m/tech)

| Key | Name | Kind | Time | Prerequisites | Effect |
|---|---|---|---|---|---|
| `ritual` | Ritual | spine | 8m 14s | `language` | Opens the Standing Stones, +10% faith production |
| `primitive_writing` | Primitive Writing | spine | 10m | `language` | +10% scholar output |
| `pottery` | Pottery | optional | 18m | `fire_mastery` | +0.8% all production |
| `animal_husbandry` | Animal Husbandry | optional | 23m | `fire_mastery` | +10% food production |
| `woodworking` | Woodworking | optional | 30m | `tool_making` | +10% wood production |
| `stoneworking` | Stoneworking | **keystone** (Great Monolith) | 13m | `tool_making` | +10% stone production |

---

### Bronze Age (~8m to 1h 5m/tech)

| Key | Name | Kind | Time | Prerequisites | Effect |
|---|---|---|---|---|---|
| `calendar` | Calendar | **keystone** (Stonehenge) | 13m | `ritual` | Opens the Altar, +10% faith production |
| `map_making` | Map Making | optional | 17m | `primitive_writing` | +10% scholar output |
| `currency` | Currency | spine | 8m | `primitive_writing` | +10% gold production |
| `boatbuilding` | Boatbuilding | optional | 22m | `woodworking` | +5% food production, trade routes bring in 10% more |
| `agriculture` | Agriculture | optional | 29m | `animal_husbandry` | +10% food production |
| `the_wheel` | The Wheel | optional | 38m | `woodworking` | Opens trade routes, construction takes 5% less time |
| `masonry` | Masonry | optional | 50m | `stoneworking` | +10% stone production, +0.8% all production |
| `bronze_working` | Bronze Working | spine | 10m | `stoneworking` | +10% stone production, +10% iron production |
| `military_tactics` | Military Tactics | optional | 1h 5m | `bronze_working` | Opens the Barracks, +15% military power |

---

### Iron Age (~18m to 1h 52m/tech)

| Key | Name | Kind | Time | Prerequisites | Effect |
|---|---|---|---|---|---|
| `priesthood` | Priesthood | optional | 39m | `calendar` | Morale can rise 5 points higher |
| `mathematics` | Mathematics | **keystone** (Colosseum) | 30m | `primitive_writing` | +8% scholar output |
| `exploration` | Exploration | spine | 18m | one of `map_making` or `boatbuilding` | +10% expedition rewards |
| `irrigation` | Irrigation | optional | 51m | `agriculture` | +8% food production, +5% housing |
| `road_building` | Road Building | optional | 1h 6m | `masonry`, `the_wheel` | +8% gold production, trade routes take 15% less time |
| `iron_smelting` | Iron Smelting | spine | 23m | `bronze_working` | Opens the Smelter, +8% iron production |
| `quarrying` | Quarrying | optional | 1h 26m | `bronze_working` | +8% marble production, +8% iron ore production |
| `siege_warfare` | Siege Warfare | optional | 1h 52m | `military_tactics` | Opens the Legion Fort, +15% military power |

---

### Classical Age (~39m to 3h 11m/tech)

| Key | Name | Kind | Time | Prerequisites | Effect |
|---|---|---|---|---|---|
| `drama` | Drama | optional | 51m | `priesthood` | Opens festivals |
| `philosophy` | Philosophy | **keystone** (Parthenon) | 39m | `mathematics` | +8% scholar output |
| `envoys` | Envoys | optional | 1h 6m | `exploration` | Opens diplomacy |
| `the_plough` | The Plough | optional | 1h 27m | `irrigation` | +8% food production |
| `civil_engineering` | Civil Engineering | optional | 1h 53m | `road_building` | +8% stone production, buildings cost 3% less |
| `metal_casting` | Metal Casting | optional | 2h 27m | `iron_smelting` | Opens the Forge, +8% iron production |
| `imperial_legions` | Imperial Legions | optional | 3h 11m | `siege_warfare`, `iron_smelting` | +15% military power, raids on you take 10% less |

---

### Medieval Age (~30m to 4h 11m/tech)

| Key | Name | Kind | Time | Prerequisites | Effect |
|---|---|---|---|---|---|
| `theology` | Theology | **keystone** (Great Library) | 52m | `philosophy` | Opens the Cathedral, +8% faith production |
| `alchemy` | Alchemy | optional | 1h 7m | `philosophy` | +8% scholar output |
| `scholasticism` | Scholasticism | capstone | 3h 12m | `alchemy`, `theology` | Research takes 6% less time |
| `banking` | Banking | spine | 30m | `currency`, `mathematics` | +8% gold production |
| `feudalism` | Feudalism | optional | 1h 27m | `the_plough` | +8% housing |
| `chronometry` | Chronometry | optional | 1h 54m | none | +5% game speed |
| `guilds` | Guilds | capstone | 4h 11m | `civil_engineering`, `metal_casting` | Buildings cost 4% less, construction takes 8% less time |
| `steel_forging` | Steel Forging | spine | 40m | `iron_smelting` | +0.25 steel/tick, +8% iron production |
| `fortification` | Fortification | optional | 2h 28m | `imperial_legions` | Raids on you take 15% less, +10% military power |

---

### Renaissance Age (~1h 56m to 9h 20m/tech)

| Key | Name | Kind | Time | Prerequisites | Effect |
|---|---|---|---|---|---|
| `patronage` | Patronage | **keystone** (Sistine Chapel) | 2h 31m | `banking` | +6% culture production |
| `printing_press` | Printing Press | optional | 3h 16m | `alchemy`, `theology` | +6% scholar output, research takes 3% less time |
| `navigation` | Navigation | spine | 1h 56m | `exploration`, `mathematics` | +10% expedition rewards |
| `crop_rotation` | Crop Rotation | optional | 4h 15m | `feudalism` | +6% food production |
| `architecture` | Architecture | optional | 5h 32m | `civil_engineering` | Buildings cost 3% less, wonders take 25% less time to build |
| `blast_furnace` | Blast Furnace | optional | 7h 10m | `steel_forging` | Opens the Foundry, +6% steel production |
| `gunpowder` | Gunpowder | optional | 9h 20m | `alchemy`, `siege_warfare` | +12% military power |

---

### Colonial Age (~1h 38m to 13h 24m/tech)

| Key | Name | Kind | Time | Prerequisites | Effect |
|---|---|---|---|---|---|
| `baroque_arts` | Baroque Arts | optional | 2h 8m | `patronage` | Festivals last 25% longer |
| `scientific_method` | Scientific Method | optional | 2h 46m | `printing_press` | Research takes 4% less time |
| `cartography` | Cartography | **keystone** (Grand Lighthouse) | 1h 38m | `navigation` | +10% expedition rewards, expeditions take 10% less time |
| `mercantilism` | Mercantilism | optional | 3h 37m | `banking`, `navigation` | Opens the black market, opens the Harbor |
| `embassies` | Embassies | optional | 4h 41m | `envoys` | Opens the Embassy, gifts raise opinion 50% more |
| `new_world_crops` | New World Crops | optional | 6h 6m | `crop_rotation` | +6% food production, +4% housing |
| `surveying` | Surveying | optional | 7h 56m | `architecture` | +6% iron ore production, construction takes 5% less time, buildings cost 2% less |
| `coke_smelting` | Coke Smelting | optional | 10h 19m | `blast_furnace` | Opens the Colonial Steelworks, +6% steel production, +6% coal production |
| `colonialism` | Colonialism | optional | 13h 24m | `cartography`, `gunpowder` | +12% military power |

---

### Industrial Age (~39m to 11h 49m/tech)

| Key | Name | Kind | Time | Prerequisites | Effect |
|---|---|---|---|---|---|
| `romanticism` | Romanticism | optional | 1h 6m | `baroque_arts` | +6% culture production |
| `encyclopedia` | Encyclopedia | optional | 1h 26m | `scientific_method` | +6% scholar output |
| `railroads` | Railroads | optional | 1h 53m | `steam_power`, `road_building` | Opens the Rail Freight route, trade routes take 15% less time |
| `geographic_societies` | Geographic Societies | optional | 2h 27m | `cartography` | Opens the Geographic Society |
| `concert_of_nations` | Concert of Nations | capstone | 9h 5m | `embassies`, `geographic_societies` | Opens the Grand Embassy, alliances give 25% more |
| `seed_drill` | Seed Drill | optional | 3h 11m | `new_world_crops` | +6% food production |
| `industrialization` | Industrialization | **keystone** (Crystal Palace) | 51m | `steam_power` | +5% all production |
| `clockwork_automation` | Clockwork Automation | optional | 4h 8m | `chronometry` | +10% game speed |
| `interchangeable_parts` | Interchangeable Parts | capstone | 11h 49m | `industrialization`, `clockwork_automation` | Buildings cost 4% less, upgrades cost 15% less |
| `steam_pumps` | Steam Pumps | optional | 5h 22m | `coke_smelting` | +6% iron ore production, +6% coal production |
| `rifling` | Rifling | optional | 6h 59m | `gunpowder` | +12% military power |
| `steam_power` | Steam Power | spine | 39m | `steel_forging` | Opens the Coal Plant, +6% steel production, +6% coal production |

---

### Victorian Age (~3h 24m to 21h 22m/tech)

| Key | Name | Kind | Time | Prerequisites | Effect |
|---|---|---|---|---|---|
| `museums` | Museums | optional | 5h 45m | `romanticism` | +5% culture production |
| `public_education` | Public Education | optional | 7h 29m | `encyclopedia` | +5% scholar output, research takes 3% less time |
| `telecommunications` | Telecommunications | optional | 9h 43m | `electrification` | Deals refresh 30% sooner, gifts cost 25% less |
| `sanitation` | Sanitation | optional | 12h 39m | `seed_drill` | +6% housing |
| `mass_production` | Mass Production | **keystone** (Eiffel Tower) | 4h 25m | `industrialization` | Construction takes 8% less time |
| `geology` | Geology | optional | 16h 26m | `steam_pumps` | +5% iron ore production, +5% coal production |
| `general_staff` | General Staff | optional | 21h 22m | `rifling` | Campaigns take 15% less time |
| `electrification` | Electrification | spine | 3h 24m | `industrialization` | Opens the Steam Works, +5% electricity production |

---

### Electric Age (~4h 20m to 35h 29m/tech)

| Key | Name | Kind | Time | Prerequisites | Effect |
|---|---|---|---|---|---|
| `radio` | Radio | optional | 9h 33m | `telecommunications` | Festivals come back 20% sooner |
| `modern_physics` | Modern Physics | optional | 12h 25m | `public_education` | +5% scholar output |
| `wire_transfers` | Wire Transfers | optional | 16h 10m | `telecommunications` | +5% gold production |
| `fertilizers` | Fertilizers | optional | 20h 57m | `sanitation` | +5% food production |
| `assembly_line` | Assembly Line | optional | 27h 15m | `mass_production` | Construction takes 5% less time |
| `chemical_engineering` | Chemical Engineering | spine | 4h 20m | `mass_production` | +5% oil production, +5% steel production |
| `mechanized_warfare` | Mechanized Warfare | optional | 35h 29m | `general_staff` | +10% military power |
| `power_distribution` | Power Distribution | **keystone** (Hoover Dam) | 7h 21m | `electrification` | Opens the Dynamo Hall, +5% electricity production |
| `aviation` | Aviation | spine | 5h 39m | `mass_production` | Expeditions take 10% less time |

---

### Atomic Age (~3h 14m to 44h 37m/tech)

| Key | Name | Kind | Time | Prerequisites | Effect |
|---|---|---|---|---|---|
| `cinema` | Cinema | optional | 5h 28m | `radio` | +5% culture production |
| `big_science` | Big Science | capstone | 34h 25m | `modern_physics` | Research takes 6% less time |
| `corporations` | Corporations | optional | 7h 7m | `mercantilism`, `wire_transfers` | +5% gold production |
| `green_revolution` | Green Revolution | optional | 9h 15m | `fertilizers` | +5% food production, +4% housing |
| `prefabrication` | Prefabrication | optional | 11h 59m | `assembly_line` | Buildings cost 2% less, +0.4% all production |
| `plastics` | Plastics | optional | 15h 35m | `chemical_engineering` | +5% oil production |
| `nuclear_deterrence` | Nuclear Deterrence | optional | 20h 23m | `nuclear_fission`, `rocketry` | +10% military power, raids on you take 10% less |
| `military_industrial_complex` | Military-Industrial Complex | capstone | 44h 37m | `mechanized_warfare`, `nuclear_deterrence` | Military buildings have 10% more soldier posts, campaigns bring back 20% more |
| `nuclear_fission` | Nuclear Fission | **keystone** (Particle Accelerator) | 4h 12m | `power_distribution`, `chemical_engineering` | +5% uranium production, +5% electricity production |
| `civilian_reactors` | Civilian Reactors | optional | 26h 23m | `nuclear_deterrence` | Opens the Nuclear Plant |
| `rocketry` | Rocketry | spine | 3h 14m | `aviation` | +10% expedition rewards, +10% military power |

---

### Modern Age (~4h 52m to 51h 44m/tech)

| Key | Name | Kind | Time | Prerequisites | Effect |
|---|---|---|---|---|---|
| `television` | Television | optional | 10h 43m | `cinema` | Opens the Monument of Ages |
| `information_theory` | Information Theory | optional | 13h 58m | `modern_physics` | +5% scholar output |
| `containerization` | Containerization | optional | 18h 7m | `corporations` | Opens the Seaport, trade routes bring in 10% more |
| `suburbs` | Suburbs | optional | 23h 32m | `green_revolution` | +5% housing |
| `nanofabrication` | Nanofabrication | optional | 30h 36m | `computers` | Buildings cost 3% less |
| `titanium_alloys` | Titanium Alloys | optional | 39h 50m | `chemical_engineering` | +5% steel production, +5% titanium ore production |
| `special_forces` | Special Forces | optional | 51h 44m | `nuclear_deterrence` | Campaigns take 15% less time |
| `electricity_tech` | Advanced Electrics | spine | 4h 52m | `nuclear_fission` | Opens the Power Grid Hub, +5% electricity production |
| `satellite_tech` | Satellite Technology | **keystone** (Space Program) | 8h 15m | `rocketry`, `electricity_tech` | +1 data/tick, +10% expedition rewards |
| `computers` | Computers | spine | 6h 20m | `electricity_tech` | +5% scholar output, research takes 3% less time |

---

### Information Age (~3h 1m to 41h 52m/tech)

| Key | Name | Kind | Time | Prerequisites | Effect |
|---|---|---|---|---|---|
| `social_media` | Social Media | optional | 5h 7m | `internet` | +5% culture production, festivals cost 20% less |
| `search_engines` | Search Engines | optional | 6h 39m | `internet` | Research takes 3% less time |
| `e_commerce` | E-commerce | optional | 8h 40m | `containerization`, `internet` | Opens the Container Terminal |
| `medical_nanobots` | Medical Nanobots | optional | 11h 14m | `nanofabrication` | +5% housing, +5% food production |
| `internet_of_things` | Internet of Things | optional | 14h 41m | `social_media`, `cybersecurity`, `medical_nanobots` | Opens the Smart Farm, opens the Smart Complex |
| `embedded_systems` | Embedded Systems | optional | 19h 3m | `nanofabrication` | Construction takes 4% less time |
| `precision_mining` | Precision Mining | optional | 24h 45m | `titanium_alloys` | +5% steel production, +5% titanium ore production |
| `cybersecurity` | Cybersecurity | spine | 3h 1m | `computers` | Opens the Cyber Command, +10% military power |
| `smart_grid` | Smart Grid | optional | 32h 11m | `electricity_tech` | Opens the Microgrid Array, +5% electricity production |
| `space_stations` | Space Stations | optional | 41h 52m | `satellite_tech` | +8% expedition rewards |
| `internet` | Internet | **keystone** (Global Network) | 3h 56m | `computers`, `satellite_tech` | +5% data production |

---

### Digital Age (~4h 41m to 109h 29m/tech)

| Key | Name | Kind | Time | Prerequisites | Effect |
|---|---|---|---|---|---|
| `virtual_reality` | Virtual Reality | optional | 6h 6m | `social_media` | Morale can rise 5 points higher |
| `open_science` | Open Science | optional | 7h 54m | `search_engines` | +5% scholar output |
| `automated_logistics` | Automated Logistics | optional | 10h 17m | `e_commerce` | Opens the Logistics Hub, trade routes take 15% less time |
| `global_village` | Global Village | capstone | 84h 25m | `e_commerce`, `social_media` | Every civilization offers 1 more deal, alliances cost 50% less |
| `gene_editing` | Gene Editing | optional | 13h 23m | `medical_nanobots` | +5% food production |
| `self_replication` | Self-Replication | optional | 17h 24m | `medical_nanobots`, `machine_learning` | +10% nanobots production, construction takes 5% less time |
| `nano_alloys` | Nano Alloys | optional | 22h 41m | `precision_mining` | +5% steel production |
| `drone_warfare` | Drone Warfare | optional | 29h 28m | `cybersecurity` | +10% military power |
| `grid_storage` | Grid Storage | optional | 38h 19m | `smart_grid` | Opens the Quantum Battery Array, +0.4% all production |
| `reusable_launchers` | Reusable Launchers | optional | 49h 47m | `space_stations` | Expeditions take 10% less time |
| `machine_learning` | Machine Learning | **keystone** (World Simulation) | 4h 41m | `internet`, `cybersecurity` | +5% data production, research takes 3% less time |
| `cloud_computing` | Cloud Computing | optional | 64h 42m | `internet` | +0.6% all production |
| `general_ai` | General AI | capstone | 109h 29m | `machine_learning`, `cloud_computing` | Research takes 6% less time |

---

### Cyberpunk Age (~26h 43m to 283h 53m/tech)

| Key | Name | Kind | Time | Prerequisites | Effect |
|---|---|---|---|---|---|
| `holography` | Holography | optional | 45h 14m | `cybernetics`, `blockchain` | Opens the Holographic Theater |
| `neural_interface` | Neural Interface | spine | 26h 43m | `machine_learning` | +4% scholar output |
| `blockchain` | Blockchain | optional | 58h 45m | `cybersecurity`, `cloud_computing` | +4% crypto production |
| `synthetic_food` | Synthetic Food | optional | 76h 31m | `gene_editing` | +4% food production |
| `cybernetics` | Cybernetics | **keystone** (Neon Citadel) | 34h 48m | `neural_interface` | Opens the Combat Aug Center, +10% military power |
| `dark_crystal_mining` | Dark Crystal Mining | optional | 99h 21m | `nano_alloys` | +4% dark matter crystals production |
| `augmented_soldiers` | Augmented Soldiers | optional | 128h 58m | `cybernetics`, `drone_warfare` | +10% military power, military buildings have 10% more soldier posts |
| `dark_energy` | Dark Energy | optional | 167h 51m | `grid_storage` | Opens the Dark Energy Tap, +4% electricity production |
| `lunar_outposts` | Lunar Outposts | optional | 218h 28m | `reusable_launchers` | +8% expedition rewards |
| `darknets` | Darknets | optional | 283h 53m | `cloud_computing` | +4% data production |

---

### Fusion Age (~46h 51m to 644h 7m/tech)

| Key | Name | Kind | Time | Prerequisites | Effect |
|---|---|---|---|---|---|
| `neural_art` | Neural Art | optional | 102h 51m | `holography` | +4% culture production |
| `unified_theory` | Unified Theory | optional | 134h 1m | `neural_interface`, `plasma_physics` | Research takes 3% less time |
| `maglev_transit` | Maglev Transit | optional | 174h 1m | `superconductors` | Opens the Energy Exchange |
| `closed_biospheres` | Closed Biospheres | optional | 226h 29m | `synthetic_food` | +5% housing |
| `molecular_assembly` | Molecular Assembly | optional | 294h | `cybernetics`, `self_replication` | Construction takes 5% less time |
| `superconductors` | Superconductors | spine | 46h 51m | `plasma_physics` | +0.6% all production |
| `plasma_weapons` | Plasma Weapons | optional | 382h 19m | `augmented_soldiers`, `plasma_physics` | +10% military power |
| `fusion_power` | Fusion Power | **keystone** (Stellar Cradle) | 79h 28m | `nuclear_fission`, `cybernetics` | +4% electricity production, +4% plasma production |
| `plasma_physics` | Plasma Physics | spine | 60h 46m | `fusion_power` | +4% plasma production |
| `fusion_drives` | Fusion Drives | optional | 497h 7m | `fusion_power`, `lunar_outposts` | Expeditions take 10% less time |
| `quantum_networking` | Quantum Networking | optional | 644h 7m | `darknets` | +4% data production |

---

### Space Age (~46h 55m to 1092h 20m/tech)

| Key | Name | Kind | Time | Prerequisites | Effect |
|---|---|---|---|---|---|
| `overview_effect` | Overview Effect | optional | 102h 45m | `neural_art` | Morale can rise 5 points higher |
| `deep_space_astronomy` | Deep Space Astronomy | optional | 133h 54m | `unified_theory` | +4% scholar output |
| `asteroid_claims` | Asteroid Claims | optional | 173h 57m | `maglev_transit` | Trade routes bring in 10% more |
| `stellar_cartography` | Stellar Cartography | capstone | 837h 27m | `asteroid_claims`, `deep_space_astronomy` | Expeditions take 20% less time, +10% expedition rewards |
| `hydroponics` | Hydroponics | optional | 226h 9m | `closed_biospheres` | +4% food production |
| `zero_g_manufacturing` | Zero-G Manufacturing | spine | 46h 55m | `orbital_mechanics`, `superconductors` | Construction takes 6% less time |
| `asteroid_refining` | Asteroid Refining | optional | 293h 43m | `space_mining` | +4% titanium production, +4% dark matter crystals production |
| `orbital_defense` | Orbital Defense | optional | 381h 54m | `plasma_weapons`, `orbital_mechanics` | Raids on you take 10% less |
| `orbital_solar` | Orbital Solar | optional | 497h 37m | `plasma_physics` | +4% electricity production, +4% plasma production |
| `orbital_mechanics` | Orbital Mechanics | **keystone** (Dyson Scaffold) | 79h 17m | `rocketry`, `plasma_physics` | +10% expedition rewards |
| `space_mining` | Space Mining | spine | 60h 41m | `orbital_mechanics` | +4% titanium production, +4% steel production, +4% dark matter crystals production |
| `space_elevator` | Space Elevator | capstone | 1092h 20m | `zero_g_manufacturing`, `space_mining` | Buildings cost 4% less, +0.6% all production |
| `orbital_relays` | Orbital Relays | optional | 647h 18m | `quantum_networking`, `orbital_mechanics` | +4% data production |

---

### Interstellar Age (~238h 16m to 2526h 50m/tech)

| Key | Name | Kind | Time | Prerequisites | Effect |
|---|---|---|---|---|---|
| `void_contemplation` | Void Contemplation | optional | 402h 8m | `overview_effect` | Appeasing a harbinger costs 20% less |
| `xenology` | Xenology | optional | 522h 10m | `deep_space_astronomy` | Research takes 3% less time |
| `interstellar_trade` | Interstellar Trade | optional | 681h 13m | `asteroid_claims`, `warp_drive` | Opens the Warp Commerce route |
| `protein_synthesis` | Protein Synthesis | optional | 885h 17m | `hydroponics` | +4% food production, +4% housing |
| `hull_printing` | Hull Printing | optional | 1149h 22m | `zero_g_manufacturing` | Buildings cost 2% less, construction takes 4% less time |
| `stellar_core_mining` | Stellar Core Mining | optional | 1494h 29m | `asteroid_refining` | Opens the Stellar Core Drill, +4% titanium production |
| `fleet_doctrine` | Fleet Doctrine | optional | 1941h 38m | `orbital_defense` | +10% military power |
| `stellar_engineering` | Stellar Engineering | spine | 238h 16m | `warp_drive` | Opens the Pulsar Tap, +4% plasma production, +4% electricity production |
| `warp_drive` | Warp Drive | **keystone** (Warp Nexus) | 309h 6m | `space_mining`, `zero_g_manufacturing` | +10% expedition rewards, +4% dark matter production |
| `galactic_network` | Galactic Network | optional | 2526h 50m | `orbital_relays` | +4% data production |

---

### Galactic Age (~120h 10m to 1656h 27m/tech)

| Key | Name | Kind | Time | Prerequisites | Effect |
|---|---|---|---|---|---|
| `galactic_memory` | Galactic Memory | optional | 203h 6m | `void_contemplation` | +4% culture production |
| `cosmology` | Cosmology | optional | 264h 10m | `xenology` | +4% dark matter production, research takes 3% less time |
| `federation_charter` | Federation Charter | optional | 344h 16m | `interstellar_trade`, `xenology` | Alliances give 25% more, every civilization offers 1 more deal |
| `matter_conversion` | Matter Conversion | optional | 446h 3m | `protein_synthesis` | +4% food production, +4% housing |
| `dyson_engineering` | Dyson Engineering | optional | 580h 18m | `stellar_engineering`, `hull_printing` | Buildings cost 2% less |
| `neutron_mining` | Neutron Mining | optional | 753h 31m | `stellar_core_mining` | Opens the Neutron Star Mine, +4% antimatter production |
| `armada_command` | Armada Command | optional | 980h 52m | `fleet_doctrine` | +10% military power |
| `antimatter_synthesis` | Antimatter Synthesis | spine | 120h 10m | `galactic_navigation` | Opens the Quasar Tap, +4% antimatter production |
| `galactic_navigation` | Galactic Navigation | **keystone** (Cosmic Beacon) | 156h 20m | `warp_drive`, `stellar_engineering` | +4% dark matter production, +10% expedition rewards |
| `terraforming` | Terraforming | optional | 1275h 21m | `warp_drive` | +6% housing |
| `mind_uploading` | Mind Uploading | optional | 1656h 27m | `galactic_network` | +4% data production |

---

### Quantum Age (~105h 13m to 1898h 11m/tech)

| Key | Name | Kind | Time | Prerequisites | Effect |
|---|---|---|---|---|---|
| `reality_art` | Reality Art | optional | 178h 28m | `galactic_memory` | +4% culture production |
| `quantum_mechanics` | Quantum Mechanics | **keystone** (Reality Anchor) | 137h 12m | `antimatter_synthesis` | +4% quantum flux production |
| `timeless_archive` | Timeless Archive | capstone | 1454h 35m | `cosmology`, `quantum_mechanics` | Research takes 6% less time |
| `probability_markets` | Probability Markets | optional | 232h 6m | `federation_charter` | +4% gold production |
| `quantum_cultivation` | Quantum Cultivation | optional | 302h 15m | `matter_conversion` | +4% food production |
| `reality_manipulation` | Reality Manipulation | spine | 105h 13m | `quantum_mechanics` | +4% quantum flux production |
| `reality_engineering` | Reality Engineering | capstone | 1898h 11m | `reality_manipulation`, `dyson_engineering` | Buildings cost 4% less, construction takes 8% less time |
| `quantum_metallurgy` | Quantum Metallurgy | optional | 393h 2m | `neutron_mining` | Opens the Quantum Metal Works, +4% antimatter production, +4% titanium production |
| `probability_warfare` | Probability Warfare | optional | 510h 39m | `armada_command` | +10% military power, campaigns bring back 15% more |
| `zero_point_energy` | Zero-Point Energy | optional | 663h 20m | `antimatter_synthesis` | Opens the Zero Point Generator, +4% dark matter production |
| `wormholes` | Wormholes | optional | 862h 26m | `galactic_navigation` | Trade routes take 20% less time, expeditions take 15% less time |
| `quantum_computing` | Quantum Computing | optional | 1124h 28m | `clockwork_automation`, `quantum_mechanics` | +15% game speed |

---

### Transcendent Age (~1123h 38m to 2469h 42m)

| Key | Name | Kind | Time | Prerequisites | Effect |
|---|---|---|---|---|---|
| `transcendence` | Transcendence | **keystone** (Singularity Core) | 1123h 38m | `reality_manipulation` | +5% all production |
| `omniversal_exchange` | Omniversal Exchange | optional | 1458h 43m | `probability_markets` | Opens the Omniversal Bazaar |
| `singularity_engineering` | Singularity Engineering | optional | 1891h 59m | `reality_manipulation` | Opens the Singularity Engine |
| `omniversal_command` | Omniversal Command | optional | 2469h 42m | `probability_warfare` | Opens the Omniversal War Council |

---

## Tech effects reference

Every number here counts in full in every age: see [How Tech Bonuses Stack](#how-tech-bonuses-stack).

### Output of one resource

Each raises what one resource's buildings and workers make. Bonuses on the same resource add up.

| Resource | Techs | Together |
|---|---|---|
| Food | Fire Mastery +10%, Tool Making +10%, Animal Husbandry +10%, Boatbuilding +5%, Agriculture +10%, Irrigation +8%, The Plough +8%, Crop Rotation +6%, New World Crops +6%, Seed Drill +6%, Fertilizers +5%, Green Revolution +5%, Medical Nanobots +5%, Gene Editing +5%, Synthetic Food +4%, Hydroponics +4%, Protein Synthesis +4%, Matter Conversion +4%, Quantum Cultivation +4% | +119% |
| Research (scholar output) | Language +10%, Primitive Writing +10%, Map Making +10%, Mathematics +8%, Philosophy +8%, Alchemy +8%, Printing Press +6%, Encyclopedia +6%, Public Education +5%, Modern Physics +5%, Information Theory +5%, Computers +5%, Open Science +5%, Neural Interface +4%, Deep Space Astronomy +4% | +99% |
| Steel | Blast Furnace +6%, Coke Smelting +6%, Steam Power +6%, Chemical Engineering +5%, Titanium Alloys +5%, Precision Mining +5%, Nano Alloys +5%, Space Mining +4% | +42% |
| Electricity | Electrification +5%, Power Distribution +5%, Nuclear Fission +5%, Advanced Electrics +5%, Smart Grid +5%, Dark Energy +4%, Fusion Power +4%, Orbital Solar +4%, Stellar Engineering +4% | +41% |
| Gold | Currency +10%, Road Building +8%, Banking +8%, Wire Transfers +5%, Corporations +5%, Probability Markets +4% | +40% |
| Culture | Patronage +6%, Romanticism +6%, Museums +5%, Cinema +5%, Social Media +5%, Neural Art +4%, Galactic Memory +4%, Reality Art +4% | +39% |
| Stone | Stoneworking +10%, Masonry +10%, Bronze Working +10%, Civil Engineering +8% | +38% |
| Iron | Bronze Working +10%, Iron Smelting +8%, Metal Casting +8%, Steel Forging +8% | +34% |
| Data | Internet +5%, Machine Learning +5%, Darknets +4%, Quantum Networking +4%, Orbital Relays +4%, Galactic Network +4%, Mind Uploading +4% | +30% |
| Faith | Ritual +10%, Calendar +10%, Theology +8% | +28% |
| Iron ore | Quarrying +8%, Surveying +6%, Steam Pumps +6%, Geology +5% | +25% |
| Coal | Coke Smelting +6%, Steam Pumps +6%, Steam Power +6%, Geology +5% | +23% |
| Wood | Tool Making +10%, Woodworking +10% | +20% |
| Plasma | Fusion Power +4%, Plasma Physics +4%, Orbital Solar +4%, Stellar Engineering +4% | +16% |
| Titanium | Asteroid Refining +4%, Space Mining +4%, Stellar Core Mining +4%, Quantum Metallurgy +4% | +16% |
| Dark matter | Warp Drive +4%, Cosmology +4%, Galactic Navigation +4%, Zero-Point Energy +4% | +16% |
| Dark matter crystals | Dark Crystal Mining +4%, Asteroid Refining +4%, Space Mining +4% | +12% |
| Antimatter | Neutron Mining +4%, Antimatter Synthesis +4%, Quantum Metallurgy +4% | +12% |
| Oil | Chemical Engineering +5%, Plastics +5% | +10% |
| Titanium ore | Titanium Alloys +5%, Precision Mining +5% | +10% |
| Nanobots | Self-Replication +10% | +10% |
| Marble | Quarrying +8% | +8% |
| Quantum flux | Quantum Mechanics +4%, Reality Manipulation +4% | +8% |
| Uranium | Nuclear Fission +5% | +5% |
| Crypto | Blockchain +4% | +4% |

### All production

| Tech | Age | Bonus |
|---|---|---|
| Pottery | Stone | +0.8% |
| Masonry | Bronze | +0.8% |
| Industrialization | Industrial | +5% |
| Prefabrication | Atomic | +0.4% |
| Grid Storage | Digital | +0.4% |
| Cloud Computing | Digital | +0.6% |
| Superconductors | Fusion | +0.6% |
| Space Elevator | Space | +0.6% |
| Transcendence | Transcendent | +5% |

Together +14.2%, on top of each resource's own bonus. Nothing here joins the all-production pool that milestones and wonders fill.

### First sources

These two techs make a resource before any building of their age does. The amount is flat: neither the pools nor the tech layer multiply it.

| Tech | Age | Makes |
|---|---|---|
| Steel Forging | Medieval | +0.25 steel/tick |
| Satellite Technology | Modern | +1 data/tick |

### Housing

| Tech | Age | Bonus |
|---|---|---|
| Fire Mastery | Primitive | +5% housing |
| Irrigation | Iron | +5% housing |
| Feudalism | Medieval | +8% housing |
| New World Crops | Colonial | +4% housing |
| Sanitation | Victorian | +6% housing |
| Green Revolution | Atomic | +4% housing |
| Suburbs | Modern | +5% housing |
| Medical Nanobots | Information | +5% housing |
| Closed Biospheres | Fusion | +5% housing |
| Protein Synthesis | Interstellar | +4% housing |
| Matter Conversion | Galactic | +4% housing |
| Terraforming | Galactic | +6% housing |

Housing bonuses add up to +61%. Housing is rounded up to a whole person, so even the first +5% on a small village houses one more.

### Building costs, construction time and research time

| Tech | Age | Cut |
|---|---|---|
| Civil Engineering | Classical | buildings cost 3% less |
| Guilds | Medieval | buildings cost 4% less |
| Architecture | Renaissance | buildings cost 3% less |
| Surveying | Colonial | buildings cost 2% less |
| Interchangeable Parts | Industrial | buildings cost 4% less |
| Prefabrication | Atomic | buildings cost 2% less |
| Nanofabrication | Modern | buildings cost 3% less |
| Space Elevator | Space | buildings cost 4% less |
| Hull Printing | Interstellar | buildings cost 2% less |
| Dyson Engineering | Galactic | buildings cost 2% less |
| Reality Engineering | Quantum | buildings cost 4% less |
| The Wheel | Bronze | construction takes 5% less time |
| Guilds | Medieval | construction takes 8% less time |
| Surveying | Colonial | construction takes 5% less time |
| Mass Production | Victorian | construction takes 8% less time |
| Assembly Line | Electric | construction takes 5% less time |
| Embedded Systems | Information | construction takes 4% less time |
| Self-Replication | Digital | construction takes 5% less time |
| Molecular Assembly | Fusion | construction takes 5% less time |
| Zero-G Manufacturing | Space | construction takes 6% less time |
| Hull Printing | Interstellar | construction takes 4% less time |
| Reality Engineering | Quantum | construction takes 8% less time |
| Scholasticism | Medieval | research takes 6% less time |
| Printing Press | Renaissance | research takes 3% less time |
| Scientific Method | Colonial | research takes 4% less time |
| Public Education | Victorian | research takes 3% less time |
| Big Science | Atomic | research takes 6% less time |
| Computers | Modern | research takes 3% less time |
| Search Engines | Information | research takes 3% less time |
| Machine Learning | Digital | research takes 3% less time |
| General AI | Digital | research takes 6% less time |
| Unified Theory | Fusion | research takes 3% less time |
| Xenology | Interstellar | research takes 3% less time |
| Cosmology | Galactic | research takes 3% less time |
| Timeless Archive | Quantum | research takes 6% less time |

Cuts of the same kind multiply. With every tech, buildings cost 28.5% less (the floor is 10% of the listed price, shared with the build-cost milestone rewards, which multiply in too), construction takes 47.8% less time (the floor is 40% of the time) and research takes 41.3% less time (the floor is 50%). The costs and times the panels show include them.

### One number in a mechanic

These techs move a single number of a game mechanic.

| Tech | Age | Effect |
|---|---|---|
| Tool Making | Primitive | Gathering by hand brings 2 more |
| Road Building | Iron | Trade routes take 15% less time |
| Railroads | Industrial | Trade routes take 15% less time |
| Automated Logistics | Digital | Trade routes take 15% less time |
| Wormholes | Quantum | Trade routes take 20% less time |
| Cartography | Colonial | Expeditions take 10% less time |
| Aviation | Electric | Expeditions take 10% less time |
| Reusable Launchers | Digital | Expeditions take 10% less time |
| Fusion Drives | Fusion | Expeditions take 10% less time |
| Stellar Cartography | Space | Expeditions take 20% less time |
| Wormholes | Quantum | Expeditions take 15% less time |
| Imperial Legions | Classical | Raids on you take 10% less |
| Fortification | Medieval | Raids on you take 15% less |
| Nuclear Deterrence | Atomic | Raids on you take 10% less |
| Orbital Defense | Space | Raids on you take 10% less |
| Telecommunications | Victorian | Deals refresh 30% sooner |
| Telecommunications | Victorian | Gifts cost 25% less |
| Radio | Electric | Festivals come back 20% sooner |
| Social Media | Information | Festivals cost 20% less |
| Boatbuilding | Bronze | Trade routes bring in 10% more |
| Containerization | Modern | Trade routes bring in 10% more |
| Asteroid Claims | Space | Trade routes bring in 10% more |
| Priesthood | Iron | Morale can rise 5 points higher |
| Virtual Reality | Digital | Morale can rise 5 points higher |
| Overview Effect | Space | Morale can rise 5 points higher |
| Architecture | Renaissance | Wonders take 25% less time to build |
| Baroque Arts | Colonial | Festivals last 25% longer |
| Embassies | Colonial | Gifts raise opinion 50% more |
| Concert of Nations | Industrial | Alliances give 25% more |
| Federation Charter | Galactic | Alliances give 25% more |
| Interchangeable Parts | Industrial | Upgrades cost 15% less |
| General Staff | Victorian | Campaigns take 15% less time |
| Special Forces | Modern | Campaigns take 15% less time |
| Military-Industrial Complex | Atomic | Military buildings have 10% more soldier posts |
| Augmented Soldiers | Cyberpunk | Military buildings have 10% more soldier posts |
| Military-Industrial Complex | Atomic | Campaigns bring back 20% more |
| Probability Warfare | Quantum | Campaigns bring back 15% more |
| Global Village | Digital | Every civilization offers 1 more deal |
| Federation Charter | Galactic | Every civilization offers 1 more deal |
| Global Village | Digital | Alliances cost 50% less |
| Void Contemplation | Interstellar | Appeasing a harbinger costs 20% less |

- Trade route time, expedition time, deal refresh and festival cooldown: a cut shortens the timer when it is next set (a route's next run, the next expedition sent, the next set of offers, the next festival). Cuts multiply (two 15% cuts leave 72% of a route's time) and none takes a timer under 40% of its length: the four cuts of a route's time leave 49% of it, and the six of an expedition's 45%. Expedition time is for short, long and civil grand missions, the scouting group; military grand missions have a campaign time of their own.
- Campaign time, campaign loot and soldier posts: General Staff and Special Forces each cut the time of every military grand mission (campaign) sent after them by 15%. The Military-Industrial Complex makes every campaign bring back 20% more of its listed loot, won or lost, on top of your expedition rewards, and Probability Warfare 15% more again. The Military-Industrial Complex and Augmented Soldiers each give every military building a tenth more soldier posts (two techs make it a fifth more), and the new posts fill as the food allows. See [Military](military.md).
- Wonder construction: with Architecture a wonder takes three quarters of its listed time to build, and the cuts of all construction then apply to what is left. It is the one cut a wonder has to itself. See [Wonders](wonders.md).
- Festival length: with Baroque Arts a festival's bonus lasts 25% longer; its price and the wait for the next are other numbers. See [Festival](commands.md#festival).
- Gifts and alliances: with Embassies a gift earns 22 opinion where it earned 15 (half as much again, rounded down). With the Concert of Nations every ally adds 25% more to its specialty, and with the Federation Charter 25% more again: a +20% bonus becomes +25%, then +31.25%, and the Factions panel lists it so. See [Factions](factions.md).
- Deals and alliances: the Global Village and the Federation Charter each add one deal to every set a civilization offers, and the Global Village halves what an alliance costs (250 gold).
- Appease cost: with Void Contemplation every Appease level costs 20% less, in every harbinger's thread, the Last Passage's included. See [Harbingers](harbinger.md).
- Upgrade cost: with Interchangeable Parts, `upgrade` asks for 15% less of every resource, on top of the half price, after the cut of building costs on the new copy. Building new is not an upgrade.
- Raid losses: the cut comes off what a raid would take before your garrison meets it, so it stacks with the garrison's own share (see [Military](military.md)). Cuts multiply: with Imperial Legions, Fortification, Nuclear Deterrence and Orbital Defense a raid takes 0.9 × 0.85 × 0.9 × 0.9, about 62% of what it would.
- Trade route income: Boatbuilding, Containerization and Asteroid Claims each make every run of every route bring in 10% more of what the route lists, 33% with all three. A harbor's and an ally's share are added to that, each as its own share of the listed amount (see [Trade](trade.md)).
- Morale ceiling: Priesthood, Virtual Reality and the Overview Effect each let morale rise 5 points higher than 100% plus what your wonders add (see [Morale](morale.md)).
- Hand gathering: every `gather` brings 2 more than you asked for, whether it is the default amount (3) or the largest (25).

### Game speed

Three techs raise game speed: ticks come more often, so everything that counts in ticks (production, construction, research, expeditions) runs faster in real time. Game speed is a pool of its own, shared with milestone chain boosts and boons. The bonuses add together:

| Source | Bonus | Lasts |
|---|---|---|
| Chronometry (Medieval) | +5% | the run |
| Clockwork Automation (Industrial) | +10% | the run |
| Quantum Computing (Quantum) | +15% | the run |
| Milestone chain boosts | +250% or +300% | 150 or 180 ticks, once per chain; longer from the Bronze Age on |
| Time Dilation boons | +8% to +15% | 1,950 to 3,900 ticks |

With all three techs the game runs 30% faster, and that multiplies with research speed. A tick is never shorter than 0.2 seconds, however much game speed you hold. See [Milestones](milestones.md) and [Boons](factions.md#boons). [Era Mastery](prestige.md#era-mastery) is a different thing: it leaves the tick alone and makes each tick of an age you know do more.

### Military power

Military power raises the odds of military grand offers and feeds your defense rating. See [Military power bonus](military.md#military-power-bonus). The bonuses add together, with the military milestones' and with nothing to cap them.

| Tech | Age | Bonus |
|---|---|---|
| Military Tactics | Bronze | +15% |
| Siege Warfare | Iron | +15% |
| Imperial Legions | Classical | +15% |
| Fortification | Medieval | +10% |
| Gunpowder | Renaissance | +12% |
| Colonialism | Colonial | +12% |
| Rifling | Industrial | +12% |
| Mechanized Warfare | Electric | +10% |
| Nuclear Deterrence | Atomic | +10% |
| Rocketry | Atomic | +10% |
| Cybersecurity | Information | +10% |
| Drone Warfare | Digital | +10% |
| Cybernetics | Cyberpunk | +10% |
| Augmented Soldiers | Cyberpunk | +10% |
| Plasma Weapons | Fusion | +10% |
| Fleet Doctrine | Interstellar | +10% |
| Armada Command | Galactic | +10% |
| Probability Warfare | Quantum | +10% |

Together +201%.

### Expedition rewards

| Tech | Age | Bonus |
|---|---|---|
| Exploration | Iron | +10% |
| Navigation | Renaissance | +10% |
| Cartography | Colonial | +10% |
| Rocketry | Atomic | +10% |
| Satellite Technology | Modern | +10% |
| Space Stations | Information | +8% |
| Lunar Outposts | Cyberpunk | +8% |
| Stellar Cartography | Space | +10% |
| Orbital Mechanics | Space | +10% |
| Warp Drive | Interstellar | +10% |
| Galactic Navigation | Galactic | +10% |

Together +106%, added to the wonders' and milestones' expedition rewards.

---

## Paying for research

Research costs no resource. It is done by your [scholars](knowledge.md), who come with the knowledge buildings (Story Circle, Elders' Hall and on up to the Reality Academy): Shamans in the Primitive Age, Quantum Theorists in the Quantum Age. They need no assignment. The work of every tier is on the [Scholars and Research](knowledge.md) page.

Research is not a resource you hold or trade, so your knowledge buildings are what you plan around. The tables give each tech's time for a full set of scholars.

---

## Strategy

### Raise research output early

Your first research bottleneck is the work your scholars do. The techs that raise research output start with `language`, `primitive_writing`, `map_making`, `mathematics` and `philosophy`. Each percent you earn early shortens every later tech.

### The research snowball

Research bonuses feed on themselves: each one makes the next tech arrive sooner. One chain through the tree runs:

```
language → primitive_writing → mathematics → philosophy → alchemy → printing_press → scientific_method → encyclopedia
```

Each of these raises research output except Scientific Method, which cuts research time. The output steps are small (+4% to +10%), they add up to +99% with all fifteen, and no cap stops them (see [Output of one resource](#output-of-one-resource)).

### Getting game speed early

`chronometry` (Medieval Age, no prerequisites) gives +5% game speed, which runs everything that counts in ticks (production, building, research, expeditions) 5% faster. `clockwork_automation` in the Industrial Age follows it for another +10%. Line them up with `plan research chronometry` as soon as you enter the Medieval Age.

### When to cancel

Canceling costs you the work done on the tech. It makes sense only when a different tech gives a bonus you need now, badly enough to do the first one's work twice. If the bar is more than halfway, finishing it is usually cheaper, and you can put the urgent tech at the top of your plan so it starts next.

### The Grand Discovery epoch event

The Grand Discovery, a major good epoch event, instantly completes up to 3 techs you could start now (their age reached and their prerequisites done), cheapest first, for free. It can come at an epoch transition when your culture strength is over 40% (see [Good Epoch Events](epochs.md#good-epoch-events)).

A tech whose prerequisites you have not all researched, or whose either-or group you have none of, is not one of them. If fewer than three can be started, it completes the ones there are. If it completes the tech you are researching, the research slot clears.

### The Ancient Civilization Memory

A second way to skip research comes only at the start of a new run, after a prestige or a Succumb, in the Primitive Age. An ancient cache has a 40% chance (once per run) to offer one technology that you have not researched. Accepting it starts that tech at once, free of prerequisites and the age requirement, but at half the work a tick (twice the normal time). The offer reaches one extra age for every two prestige levels, so a high-prestige run can be offered a tech from an age it has not reached yet.

Unlike Grand Discovery, it never comes on your first-ever run: it needs prestige level 1 or higher. See [Prestige](prestige.md#ancient-civilization-memory) for the full conditions.

### Late-game research scaling

Prices rise steeply from the first age to the last. The work of your scholars rises with them, so build your age's knowledge building up to its full-rate count; a copy past that count still helps, at half rate.

---

## Tips and common mistakes

- Canceling loses the work. The Dark Age epoch event also cancels your active research and loses its work, and takes 40% off the work your scholars do for a while. If an epoch transition is close, consider whether to delay a long research start until after its event resolves.
- Prerequisites stack. Before typing `research banking`, check that you have both `currency` and `mathematics`. `research list` shows only the techs you can start now; on the tech tree a tech that still waits has a dashed frame, and its card names what it builds on, with the techs you lack dimmed.
- Prestige resets research entirely, with every tech, its bonus and the milestones that gave research speed. What survives a prestige is Ancient Knowledge (research time ×0.8 per epoch) from Succumbing, faster research times in the ages a past run completed ([Era Mastery](prestige.md#era-mastery)), and, after the first prestige, the techs [Plan Template](prestige.md#plan-template) reads from your build plan.
- A Succumb ends the run but keeps Ancient Knowledge. Succumbing to a catastrophe in the Iron Era (the earliest era one can strike in) cuts research time to ×0.8 permanently, and each further epoch you Succumb in multiplies it by 0.8 again. The fallen run's completed ages also gain an [Era Mastery](prestige.md#era-mastery) level. Every later run begins with the faster research from the first tick. See [Succumb](catastrophe.md#succumb).
- Civil Engineering (`civil_engineering`) makes buildings cost 3% less from the Classical Age on, for the rest of the run, and the cut multiplies with every build-cost milestone reward you earn after it. It also adds 8% to stone output.

---

*See also: [Scholars and Research](knowledge.md) for the scholars who research; [The Build Plan](plan.md) for queuing research; [Epochs](epochs.md) for how Grand Discovery and the Dark Age event fire; [Prestige](prestige.md) for the Plan Template and the Ancient Civilization Memory; [Buildings](buildings.md) for the buildings techs open.*
