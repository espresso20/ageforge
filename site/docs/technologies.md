# Technologies

Research is your civilization's strongest long-term lever. 202 technologies span all 22 ages, and each one changes your production, military strength, storage, a command you can use, or the pace of the game, for the rest of the run. A tech's bonus is small and always counts in full: no cap holds it back (see [How Tech Bonuses Stack](#how-tech-bonuses-stack)). Only one technology is researched at a time, but the [build plan](plan.md) can queue the next ones and start each as soon as the slot frees up.

<figure class="screen" data-screen="research"><figcaption>The Research panel in the Classical Age: the tech tree as a map. Philosophy is part way through, Civil Engineering can start, and the Medieval Age waits below the dotted line.</figcaption></figure>

---

## How Research Works

### Starting Research

Research costs **knowledge points (kp)**, deducted immediately when you start. There is no refund if you cancel: the knowledge is gone the moment the tech starts.

Once started, the tech counts down in **ticks**. Each game tick takes 1 off the counter. When it hits zero, the effects are applied at once and last for the run.

**Tick count with research speed:**
```
ticks = base ticks × (1 − research speed), rounded down, never below 1
```

A research speed of +30% cuts the tick count to 70% of base. It cuts the number of ticks, not their length. A tick is 2 seconds of real time at base; game speed bonuses make each tick shorter (see [Game speed](#game-speed)). The two multiply, so a civilization with high research speed **and** high game speed researches much faster.

The tick count is locked in when you start the tech. Gaining more research speed mid-research does not shorten the current countdown.

A tech's card on the tech tree gives the time it would take if you started it now, with research speed, Ancient Knowledge and Era Mastery counted, and the Stats panel says what your research speed does (`Research speed +30%: techs take 70% of their base time.`).

**Ancient Knowledge.** For each distinct epoch you have [Succumbed](catastrophe.md#ancient-knowledge) in, the adjusted ticks are multiplied by 0.8, rounded down and never below one tick: ×0.8 after one epoch, ×0.64 after two, ×0.26 with all six. It multiplies what research speed leaves, so it never brings a tech to a single tick on its own. The Stats panel has a line for it (`Ancient Knowledge: research time ×0.64. The times on the tech tree include it.`).

**Era Mastery.** On known ground (an age a past run completed) the ticks are then divided by the age's [Era Mastery](prestige.md#era-mastery) speed, rounded up and never below one tick: 2x after one completion, up to 4.2x after ten. A tech's card shows the shortened time.

**Research techs.** Thirteen techs cut research time: the four Knowledge and Computing capstones (Scholasticism, Big Science, General AI and the Timeless Archive) by 6% each, Scientific Method by 4%, and Printing Press, Public Education, Computers, Search Engines, Machine Learning, Unified Theory, Xenology and Cosmology by 3% each. Their cuts multiply what research speed leaves (×0.587 with all thirteen), rounded down and never below one tick. However many techs cut it, research never takes less than 50% of its time on their account. The Stats panel has a line for it (`Research techs: research time ×0.94. The times on the tech tree include it.`).

All together:

```
ticks = base ticks × (1 − research speed) × the techs' cut × 0.8 ^ epochs succumbed in ÷ Era Mastery speed
```

### Knowledge Cost is Upfront

Knowledge is removed from your stockpile when the tech starts, before any ticks pass. If you don't have enough, the command fails, and since you can't hold more knowledge than your storage, your knowledge storage must be at least the tech's cost. If your knowledge income drops to zero during a long research countdown, **research still completes**: the ticks count down whatever your knowledge income is, because the cost was already paid.

### Keystones, the Spine and the Rest

Every tech has a **kind**, and its kind decides how long it takes and what it costs.

- A **keystone** is the tech an age's wonder needs before it can be built. There is one in each age from the Stone Age on, and since the next age needs the wonder, the keystone is the one tech an age asks of you. See [Keystone Techs](ages.md#keystone-techs) for the list.
- The **spine** is every tech a keystone stands on, all the way down: Tool Making under Stoneworking, Steam Power under Industrialization. Keystones and the spine are the only techs a run has to research, 48 of the 202.
- A **capstone** ends a lane for its era: a bigger step that takes longer and costs more than any other tech of its age. No run has to research one. Each era from the Iron Era on has two, twelve in all: Scholasticism and Guilds in the Medieval Age, the Concert of Nations and Interchangeable Parts in the Industrial Age, Big Science and the Military-Industrial Complex in the Atomic Age, the Global Village and General AI in the Digital Age, Stellar Cartography and the Space Elevator in the Space Age, the Timeless Archive and Reality Engineering in the Quantum Age.
- Everything else is **optional**: yours to take or leave.

The tech tree draws each keystone with a ★ on a double frame and each capstone in half blocks, a keystone's card names the wonder that waits for it, and `research list` marks it too (`★ keystone: Colosseum`).

### How Research is Priced

No tech has a price or a time of its own. Both follow from its age and its kind.

**Time.** Each age has a research cap: one eighth of the age's target length. A tech takes a share of its age's cap:

| Kind | Share of the age's research cap |
|---|---|
| Spine | 40% |
| Keystone | 50% |
| Optional | 50% |
| Capstone | 80% |

So a tech you must research takes a sixteenth of its age at most, and a capstone, the long one, a tenth.

**Cost.** An age's techs share one knowledge budget: 90% of the knowledge a well-run civilization makes in the age's target time (50% in the Primitive Age, 30% in the Transcendent Age). The budget is split by weight:

| Kind | Weight |
|---|---|
| Spine | 0.6 |
| Keystone | 0.8 |
| Optional | 1.0 |
| Capstone | 1.6 |

A tech costs the budget times its weight, divided by the weights of all the techs in its age. So the techs you must research are the cheap ones, and a keystone always costs less than an optional tech of its age. Researching everything an age offers takes most of the knowledge the age makes, so research lasts the whole age: there is nearly always a tech coming within reach. When an update adds techs to an age, each tech there gets a little cheaper and the age's total stays where it is. Every age now holds the techs the tree was drawn for: three in the Primitive Age, four in the Transcendent, and six to thirteen in each age between.

A keystone costs a twentieth to a seventh of what its age makes of knowledge: a seventh in the Stone Age, about a tenth from the Bronze Age to the Electric Age, and a twentieth to a twelfth from the Atomic Age on, where the ages hold the most techs. The techs a keystone stands on in earlier ages are cheaper still, but they add up if you skipped them, so pick up each age's spine as you go.

The costs and ticks in the tables below are base values. They used to be typed by hand, tech by tech: early techs cost more knowledge than their own age made (the Stone Age's four cost 28.5K against about 4K made), and from the Classical Age on the whole list cost next to nothing but for a few techs priced to land mid-age. Now the early techs are within reach in their own ages and the later ones are a real part of each age.

### One Slot, and the Plan as a Queue

There is one research slot. If you type `research` for a second tech while one is in progress, you get an error naming the active tech and the time it has left.

To line techs up, add them to the [build plan](plan.md) with `plan research <tech>` (or `plan res`). The plan starts each one as soon as the slot is free and you hold its knowledge, in plan order, and it keeps doing so while you are away:

- Only the first research item in the plan can take the slot when it frees up. A tech further down waits its turn even if it is cheaper.
- A planned tech holds back its knowledge cost while it waits, so plan items below it can't spend that knowledge.
- **A tech brings what it needs.** `plan research imperial_legions` adds every prerequisite that is not researched, in progress or planned yet, each after its own, and then Imperial Legions. One command queues a whole chain toward a goal, and the answer lists what went in, in order. For an either-or group the plan takes the branch you already hold or have planned, else the one that costs the least knowledge to add.
- **Any tech you can see.** The plan takes a tech of any age on your [tech tree](#reading-the-tech-tree): up to the next age on a first run, and every age your account has reached on a later one. A tech of a later age waits for its age (`waits for the Iron Age`) without holding up this age's techs below it.
- A tech already in the plan that loses a prerequisite (you removed it, moved it below, canceled it mid-research, or a game update changed what the tech needs) is not dropped: it waits, and the plan shows what it needs (`needs Philosophy first`). While it waits it holds neither the slot's turn nor its knowledge, so you can plan the missing tech below it and that one starts first.
- On the tree, a planned tech shows its place in the plan (`▸3`), and the second `Enter` on a locked tech's card plans it with its chain.

The plan is the only research queue, and the game never chooses what you research next. The techs you plan are remembered with the rest of your plan: with the prestige legacy kit's [Plan Template](prestige.md#plan-template) owned, they are planned again on later runs, in the age you planned them in. Techs you start by hand with `research` are not remembered.

See [The research queue](plan.md#the-research-queue) and [How it runs](plan.md#how-it-runs) for the rest of the plan's rules.

### When Research Completes

Effects are applied the tick the counter hits zero. You'll see a success message in the log. The tech is marked researched, and its bonuses apply to your rates from the next tick.

---

## Research Speed Sources

Research speed reduces the tick count when research starts. Its sources add together:

| Source | Research speed | Notes |
|---|---|---|
| **Tech Pioneer** milestone (research 22 techs) | +5% | Scholar Chain |
| **Philosophes** milestone (63 techs, from the Classical Age) | +5% | |
| **Renaissance Mind** milestone (82 techs, from the Renaissance Age) | +10% | Scholar Chain, hidden until you get close |
| **Tech Master** milestone (113 techs, from the Information Age) | +10% | Scholar Chain, hidden; also +5% all production |
| **Tech Ascendant** milestone (all 202 techs, Transcendent Age) | +20% | Hidden. It arrives with your last tech, so it never shortens one |
| **Ancient Knowledge** (Succumb) | not research speed: research time ×0.8 per epoch | For each distinct epoch you Succumb in (Iron to Cosmic, ×0.26 with all six). It multiplies the time research speed leaves, so it is not in this pool and no cap holds it. Kept through Succumb, prestige and save/load. See [Ancient Knowledge](catastrophe.md#ancient-knowledge) |
| Thirteen techs, from **Scholasticism** to the **Timeless Archive** (see [Building costs, construction time and research time](#building-costs-construction-time-and-research-time)) | not research speed: research time ×0.94 to ×0.97 each | They multiply the time research speed leaves, ×0.587 with all thirteen, and never take it under 50% between them. They are not in this pool |

Milestone research speed lasts for the run: milestones start over at prestige and at Succumb. Five milestones add +50% in total, +30% of it before your last tech. See [Milestones](milestones.md).

Nothing caps research speed except that a tech always takes at least 1 tick. At +100% every tech would finish on the tick after you start it, and anything past +100% would add nothing. No run reaches that: the five milestones are the only sources and add +50% between them. (Ancient Knowledge used to be +25% research speed per epoch and reached +100% alone with the fourth; it is a multiplier on research time now.) The Stats panel would mark the Research speed line as capped, and the log would say so when a research speed bonus is past it.

> **Research speed is not knowledge output.** Fifteen techs raise knowledge output (see [Output of one resource](#output-of-one-resource) below). More knowledge lets you afford techs sooner; each tech still takes the same number of ticks.

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
Adds a tech to the build plan, with the techs it still needs before it. The plan starts each when the research slot is free and you hold its knowledge (see [One Slot, and the Plan as a Queue](#one-slot-and-the-plan-as-a-queue)). It takes any tech on your tree, the next age's included. `plan res` is the short form.

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
With no arguments, opens the **Research panel**, the tech tree as a map (so do `techs` and `research tree`). See [Reading the Tech Tree](#reading-the-tech-tree).

---

```
research tree close
research tree far
```
Opens the tree at that zoom: big badges with full names (close, the default), or one-line pills that fit every lane on screen (far). `PgUp` and `PgDn` do the same from inside the panel.

---

```
research card <tech_key>
```
Opens the tree on that tech's card. It works for any tech you can see on the map, and never for one in an age still hidden.

---

**Shortcut:** `res` is an alias for `research`. All subcommands work identically.

---

## Reading the Tech Tree

`research` opens the tree as a map. It covers the whole screen but the command bar, and it does not have to fit: the view follows the tech you select.

- **Lanes run across, ages run down.** Each lane is a column with its name at the top (Knowledge, Trade, Craft, and so on); a lane appears once one of its techs is in sight. When lanes are off screen, an arrow at that end of the lane row counts them (`2▸`), and a wide terminal names the nearest on the left (`◂ FAITH`). Each age is a band, named in the gutter on the left with how many of its techs you hold.
- **A tech is a badge with its name over it.** The glyph in the middle is the tech's emblem, or its lane's.
- **Lines show what a tech needs.** A line leaves the notch under a badge and comes down on the name of the tech that needs it. A solid line is needed outright. Dashed lines marked `or` are an either-or group: one of them will do. Selecting a tech lights the chain that leads to it.
- **One line under the map** names the selected tech, its state, its price and time, and what it opens. A key bar sits under that.

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
| A shaded block `░░░░░░░` | A tech of the next age: you can read its card and plan it |
| A `◇` at the bottom right | It opens a command or a building |

Zoomed out, a tech is one line: `✓` before the emblem is researched, `⟳` is in progress, `( )` round it can start, `┆ ┆` waits, `░ ░` is the next age, `▸` before it is in your plan, and a `★` after the letterhead is a keystone.

Ages past the next one are not drawn. One line at the foot of the map counts their techs, and nothing names them: not the map, not a card, not `research card`.

### The card

`Enter` on a tech opens its card over the map. It says what the tech does in plain sentences, what it opens, what it costs and how long it takes at your current research speed, what it builds on (techs you hold in green) and how many later techs build on it.

<figure class="screen" data-screen="research-card"><figcaption>Civil Engineering's card. Research is busy, so Enter would add it to the build plan.</figcaption></figure>

The card is also the confirmation. Research spends knowledge for good, so nothing happens on the first `Enter`. The last line of the card says what a second `Enter` does: start the tech if it can start now, or add it to your [build plan](plan.md) if the research slot is busy, the knowledge is not there yet, or the tech still waits for something. A tech that needs others goes in with them, in order, and the line names them first (`adds it to your plan, after Military Tactics, which it needs first`). `Esc` closes the card.

### Keys

The panel takes only keys that print nothing. Letters always go to the command bar, so every command works with the tree open.

| Key | On the map | With a card open |
|---|---|---|
| Arrows | Move to the nearest tech that way; the view follows | Nothing |
| `Tab`, `Shift-Tab` | The next or previous tech you can start | Nothing |
| `PgUp`, `PgDn` | Zoom out and in | Nothing |
| `Home` | Jump back to your current age | Nothing |
| `Enter` | Open the selected tech's card | Do what the card says |
| `Esc` | Close the panel | Close the card |

With something typed in the command bar, `Tab` and `Enter` belong to the command, as on the Map.

The glyphs follow your `map glyphs` setting: with `map glyphs ascii` the tree is drawn in plain ASCII, with letters for emblems. What research adds up to (every bonus your techs give together, and what research speed does to the times) is on the Stats panel, under **Research bonuses**.

---

## How Tech Bonuses Stack

A tech's bonus is applied in a layer of its own, after every other bonus and after the pools' soft cap.

The game's other bonuses (milestone rewards, wonders, monuments, festivals, events, boons) add up in pools, and a pool has a soft cap: all production and each resource's own output apply in full up to +200%, and a quarter of every point past it (see [The all-production cap](resources.md#the-all-production-cap)). So a pool that has earned +320% applies +230%. Techs join no pool. A tech's output bonus multiplies what the pools make, where [Era Mastery](prestige.md#era-mastery) does, so it counts in full whatever else you hold.

```
what a resource makes = buildings and workers × pools (in full to +200%, a quarter past it) × (1 + tech bonuses)
```

- **Bonuses on the same thing add up.** Two +5% techs on the same resource give +10%, not ×1.1025. A tech's bonus on all production adds to its bonuses on one resource: with +20% stone and +5% all production, stone runs at ×1.25.
- **Nothing caps the layer.** The tree is finite, and all of its techs together come to +99% knowledge, +119% food, +42% steel, +31% gold and +10% all production. The last tech you take does its full step.
- **Flat amounts stay amounts.** Steel Forging's +0.25 steel a tick and Satellite Technology's +1 data are the first source of a resource, not a bonus, and the layer does not multiply them.
- **Cuts multiply, and each has a floor.** A tech that cuts a price or a time takes its share of what the cuts before it left: two 3% cuts on building costs leave 0.97 × 0.97. Building costs never fall under 10% of the listed price, construction under 40% of its time, or research under 50% of its time, however many cuts stack. All of today's techs together stay far from every floor.
- **Storage and housing are percentages** of everything you hold: +10% storage is 10% more of every store, and +5% housing is 5% more housing, rounded up to a whole person.
- **Game speed, military power and expedition rewards** are added to the pools of those names, shared with milestones and wonders. None of the three has the soft cap.
- The `rates` command shows what the techs add to a resource on its **Research** line, and the Stats panel lists what all your techs come to together, under Research bonuses.

A tech's bonus is never permanent. Prestige, Succumb and a new game wipe every tech and everything it gave.

---

## Commands a Tech Opens

Some commands wait for a tech. Until it is researched, the command is refused with the name of the tech that opens it (`Campaigns need Military Tactics first. Research it to send one.`). The tech's effect list says what it opens, and the log says it again when the research finishes (`Military Tactics opens campaigns.`).

| Command | Opened by | Age |
|---|---|---|
| Trade routes (`trade route start <route>`) | The Wheel | Bronze |
| Campaigns (`campaign <key>`) | Military Tactics | Bronze |
| Expeditions past the Scout Party (`expedition scout_ruins`) | Exploration | Iron |
| Diplomacy: gifts, alliances, rivalries, embargoes and deals (`diplomacy ...`, `plan deal`) | Envoys | Classical |
| Festivals (`festival confirm yes`) | Drama | Classical |
| The Naval Expedition (`expedition naval_expedition`) | Navigation, after Exploration | Renaissance |
| The black market (`blackmarket <resource>`) | Mercantilism | Colonial |
| The Rail Freight route (`trade route start rail_freight`) | Railroads | Industrial |
| The Warp Commerce route (`trade route start warp_commerce`) | Interstellar Trade | Interstellar |

The Scout Party needs no tech: it is a walk into the hills, open from the first age through the Bronze Age. Scout Nearby Ruins is listed from the Bronze Age and waits for Exploration, an Iron Age tech, which needs Map Making or Boatbuilding. The market needs no tech either, only its building. The panels say which tech a command waits for (the Trade, Expeditions, Army and Factions panels, and `festival`). Every lock has its tech now: Warp Commerce's, Interstellar Trade, was the last to arrive.

**Saves from before a lock.** A game saved before a command waited for its tech keeps every command open in the age it was in, and keeps for the rest of that run any command it was already using: a campaign under way, a route running, an ally, a festival held. In that age its buildings and its wonder wait for no tech either. One line in the log says so when the save is first loaded. The locks start with the next age for what that run never used, and with the next run for the rest. This covers the four commands and the seven buildings that found their tech with the Stone and Iron Eras' new techs, Stonehenge's keystone, the nine buildings that found theirs with the Steel and Electric Eras' techs, and the Warp Commerce route and the nineteen buildings that found theirs with the Digital, Neon and Cosmic Eras' (a game saved in one of those ages keeps its age's buildings open until it advances, and a Warp Commerce route it was running stays open for the run).

---

## Tech Tree by Age

Prerequisites are listed by tech key. A tech needs every prerequisite it lists.

**Changed prerequisites.** Thirteen techs need different techs than they used to. Fire Mastery needs nothing now, and Rocketry follows Aviation, no longer Rifling and Chemical Engineering. Primitive Writing follows Language, Feudalism follows The Plough, Road Building needs The Wheel as well as Masonry, and Navigation needs Exploration as well as Mathematics. Mathematics no longer needs Currency, Philosophy no longer lists Primitive Writing (Mathematics already needs it), Civil Engineering no longer lists Masonry (Road Building already needs it), Navigation no longer needs Road Building, and Mass Production no longer needs Railroads. Alchemy needs Philosophy instead of Mathematics, and Quantum Computing needs Quantum Mechanics as well as Clockwork Automation. A save from before the change keeps everything it had: a tech you already researched stays researched and still opens the techs that need it, a research in progress finishes, and a tech in your plan waits for what it now needs.

Research time is a share of the age's research cap, one eighth of the age's target: 40% for a spine tech, 50% for a keystone or an optional one, 80% for a capstone (see [How Research is Priced](#how-research-is-priced)). Ticks and times below are base values (2 seconds a tick), before research speed and game speed bonuses. Every effect in the tables counts in full in every age (see [How Tech Bonuses Stack](#how-tech-bonuses-stack)).

**Mid-age unlocks.** Four techs open a building partway through their age instead of at its start: Civilian Reactors (Atomic, the Nuclear Plant), Internet of Things (Information, the Smart Farm and the Smart Complex), Holography (Cyberpunk, the Holographic Theater) and Maglev Transit (Fusion, the Energy Exchange). What times them is where they stand in the tree: each needs most of its age's other techs first, so it comes after them. Civilian Reactors stands behind Nuclear Deterrence, Internet of Things behind three Information Age techs, Holography behind Cybernetics and Blockchain, and in the Fusion Age, Plasma Physics, Superconductors and Maglev Transit run one after another. A gated building stays hidden from the build list until its tech is done, and `build` and `upgrade` refuse it until then; you can still add it to your build plan, where it waits for the tech. Buildings you already have stay built. See [Buildings a tech opens](buildings.md#buildings-a-tech-opens). A wonder's keystone works the same way for building, but the wonder is listed and its bank is open from the start of the age.

**Buildings that wait for a tech.** Forty buildings wait for a tech of their own age in the same way; [Buildings a tech opens](buildings.md#buildings-a-tech-opens) lists them all, from the Standing Stones (Ritual) to the Omniversal Bazaar (Omniversal Exchange). Each either has another building of its age that makes the same thing, or waits for a tech no run leaves the age without, so no age is ever stuck behind an optional tech. A building that is its age's only source of what it makes stays open from the start of its age: every age's knowledge and culture building and most of its military ones, the Steam Mine, the Uranium Mine, the Nano Foundry, the titanium smelters and mines, the Cyber Hub and the data buildings after it, and the gold buildings of the ages that have only one. The Transcendent Age is the exception: nothing comes after it, so nothing can be stuck, and its three last buildings each wait for one of its three last techs.

**Either-or.** Exploration is the tree's one tech with two ways in: it needs Map Making or Boatbuilding, whichever you hold. The tree draws both lines dashed with an `or`, and `plan research exploration` takes the one you have or, with neither, the cheaper.

### Primitive Age (~46s to 56s/tech)

| Key | Name | Kind | Cost | Ticks | Prerequisites | Effect |
|---|---|---|---|---|---|---|
| `language` | Language | spine | 38 kp | 23 | none | +10% knowledge production |
| `fire_mastery` | Fire Mastery | optional | 63 kp | 28 | none | +10% food production, +5% housing |
| `tool_making` | Tool Making | spine | 38 kp | 23 | none | +10% food production, +10% wood production, gathering by hand brings 2 more |

---

### Stone Age (~2m 16s to 2m 48s/tech)

| Key | Name | Kind | Cost | Ticks | Prerequisites | Effect |
|---|---|---|---|---|---|---|
| `ritual` | Ritual | spine | 409 kp | 68 | `language` | Opens the Standing Stones, +10% faith production |
| `primitive_writing` | Primitive Writing | spine | 409 kp | 68 | `language` | +10% knowledge production |
| `pottery` | Pottery | optional | 682 kp | 84 | `fire_mastery` | +10% storage |
| `animal_husbandry` | Animal Husbandry | optional | 682 kp | 84 | `fire_mastery` | +10% food production |
| `woodworking` | Woodworking | optional | 682 kp | 84 | `tool_making` | +10% wood production |
| `stoneworking` | Stoneworking | **keystone** (Great Monolith) | 545 kp | 84 | `tool_making` | +10% stone production |

---

### Bronze Age (~11m to 14m/tech)

| Key | Name | Kind | Cost | Ticks | Prerequisites | Effect |
|---|---|---|---|---|---|---|
| `calendar` | Calendar | **keystone** (Stonehenge) | 4.77K kp | 439 | `ritual` | Opens the Altar, +10% faith production |
| `map_making` | Map Making | optional | 5.97K kp | 439 | `primitive_writing` | +10% knowledge production |
| `currency` | Currency | spine | 3.58K kp | 351 | `primitive_writing` | +10% gold production, market fee 3 points lower |
| `boatbuilding` | Boatbuilding | optional | 5.97K kp | 439 | `woodworking` | +5% food production, trade routes bring in 10% more |
| `agriculture` | Agriculture | optional | 5.97K kp | 439 | `animal_husbandry` | +10% food production |
| `the_wheel` | The Wheel | optional | 5.97K kp | 439 | `woodworking` | Opens trade routes, construction takes 5% less time |
| `masonry` | Masonry | optional | 5.97K kp | 439 | `stoneworking` | +10% storage |
| `bronze_working` | Bronze Working | spine | 3.58K kp | 351 | `stoneworking` | +10% stone production, +10% iron production |
| `military_tactics` | Military Tactics | optional | 5.97K kp | 439 | `bronze_working` | Opens campaigns, opens the Barracks, +15% military power |

---

### Iron Age (~19m to 24m/tech)

| Key | Name | Kind | Cost | Ticks | Prerequisites | Effect |
|---|---|---|---|---|---|---|
| `priesthood` | Priesthood | optional | 43.9K kp | 731 | `calendar` | Morale can rise 5 points higher |
| `mathematics` | Mathematics | **keystone** (Colosseum) | 35.1K kp | 731 | `primitive_writing` | +8% knowledge production |
| `exploration` | Exploration | spine | 26.3K kp | 585 | one of `map_making` or `boatbuilding` | Opens expeditions past the Scout Party |
| `irrigation` | Irrigation | optional | 43.9K kp | 731 | `agriculture` | +8% food production, +5% housing |
| `road_building` | Road Building | optional | 43.9K kp | 731 | `masonry`, `the_wheel` | +8% gold production, trade routes take 15% less time |
| `iron_smelting` | Iron Smelting | spine | 26.3K kp | 585 | `bronze_working` | Opens the Smelter, +8% iron production |
| `siege_warfare` | Siege Warfare | optional | 43.9K kp | 731 | `military_tactics` | Opens the Legion Fort, +15% military power |

---

### Classical Age (~34m/tech)

| Key | Name | Kind | Cost | Ticks | Prerequisites | Effect |
|---|---|---|---|---|---|---|
| `drama` | Drama | optional | 181K kp | 1,024 | `priesthood` | Opens festivals |
| `philosophy` | Philosophy | **keystone** (Parthenon) | 145K kp | 1,024 | `mathematics` | +8% knowledge production |
| `envoys` | Envoys | optional | 181K kp | 1,024 | `exploration` | Opens diplomacy |
| `the_plough` | The Plough | optional | 181K kp | 1,024 | `irrigation` | +8% food production |
| `civil_engineering` | Civil Engineering | optional | 181K kp | 1,024 | `road_building` | Buildings cost 3% less |
| `metal_casting` | Metal Casting | optional | 181K kp | 1,024 | `iron_smelting` | Opens the Forge, +8% iron production |
| `imperial_legions` | Imperial Legions | optional | 181K kp | 1,024 | `siege_warfare`, `iron_smelting` | +15% military power, raids on you take 10% less |

---

### Medieval Age (~35m to 1h 10m/tech)

| Key | Name | Kind | Cost | Ticks | Prerequisites | Effect |
|---|---|---|---|---|---|---|
| `theology` | Theology | **keystone** (Great Library) | 1.62M kp | 1,316 | `philosophy` | Opens the Cathedral, +8% faith production |
| `alchemy` | Alchemy | optional | 2.03M kp | 1,316 | `philosophy` | +8% knowledge production |
| `scholasticism` | Scholasticism | capstone | 3.24M kp | 2,106 | `alchemy`, `theology` | Research takes 6% less time |
| `banking` | Banking | spine | 1.22M kp | 1,053 | `currency`, `mathematics` | +8% gold production, market fee 3 points lower |
| `feudalism` | Feudalism | optional | 2.03M kp | 1,316 | `the_plough` | +8% housing |
| `chronometry` | Chronometry | optional | 2.03M kp | 1,316 | none | +5% game speed |
| `guilds` | Guilds | capstone | 3.24M kp | 2,106 | `civil_engineering`, `metal_casting` | Buildings cost 4% less, construction takes 8% less time |
| `steel_forging` | Steel Forging | spine | 1.22M kp | 1,053 | `iron_smelting` | +0.25 steel/tick, +8% iron production |
| `fortification` | Fortification | optional | 2.03M kp | 1,316 | `imperial_legions` | Raids on you take 15% less, +10% military power |

---

### Renaissance Age (~46m to 58m/tech)

| Key | Name | Kind | Cost | Ticks | Prerequisites | Effect |
|---|---|---|---|---|---|---|
| `patronage` | Patronage | **keystone** (Sistine Chapel) | 15.8M kp | 1,755 | `banking` | +6% culture production |
| `printing_press` | Printing Press | optional | 19.7M kp | 1,755 | `alchemy`, `theology` | +6% knowledge production, research takes 3% less time |
| `navigation` | Navigation | spine | 11.8M kp | 1,404 | `exploration`, `mathematics` | Opens the Naval Expedition, +10% expedition rewards |
| `crop_rotation` | Crop Rotation | optional | 19.7M kp | 1,755 | `feudalism` | +6% food production |
| `architecture` | Architecture | optional | 19.7M kp | 1,755 | `civil_engineering` | Buildings cost 3% less, wonders take 25% less time to build |
| `blast_furnace` | Blast Furnace | optional | 19.7M kp | 1,755 | `steel_forging` | Opens the Foundry, +6% steel production |
| `gunpowder` | Gunpowder | optional | 19.7M kp | 1,755 | `alchemy`, `siege_warfare` | +12% military power |

---

### Colonial Age (~1h 8m/tech)

| Key | Name | Kind | Cost | Ticks | Prerequisites | Effect |
|---|---|---|---|---|---|---|
| `baroque_arts` | Baroque Arts | optional | 52.1M kp | 2,048 | `patronage` | Festivals last 25% longer |
| `scientific_method` | Scientific Method | optional | 52.1M kp | 2,048 | `printing_press` | Research takes 4% less time |
| `cartography` | Cartography | **keystone** (Grand Lighthouse) | 41.7M kp | 2,048 | `navigation` | +10% expedition rewards, expeditions take 10% less time |
| `mercantilism` | Mercantilism | optional | 52.1M kp | 2,048 | `banking`, `navigation` | Opens the black market, opens the Harbor |
| `embassies` | Embassies | optional | 52.1M kp | 2,048 | `envoys` | Opens the Embassy, gifts raise opinion 50% more |
| `new_world_crops` | New World Crops | optional | 52.1M kp | 2,048 | `crop_rotation` | +6% food production, +4% housing |
| `surveying` | Surveying | optional | 52.1M kp | 2,048 | `architecture` | Construction takes 5% less time, buildings cost 2% less |
| `coke_smelting` | Coke Smelting | optional | 52.1M kp | 2,048 | `blast_furnace` | Opens the Colonial Steelworks, +6% steel production, +6% coal production |
| `colonialism` | Colonialism | optional | 52.1M kp | 2,048 | `cartography`, `gunpowder` | +12% military power |

---

### Industrial Age (~1h 2m to 2h 4m/tech)

| Key | Name | Kind | Cost | Ticks | Prerequisites | Effect |
|---|---|---|---|---|---|---|
| `romanticism` | Romanticism | optional | 126M kp | 2,340 | `baroque_arts` | +6% culture production |
| `encyclopedia` | Encyclopedia | optional | 126M kp | 2,340 | `scientific_method` | +6% knowledge production |
| `railroads` | Railroads | optional | 126M kp | 2,340 | `steam_power`, `road_building` | Opens the Rail Freight route, trade routes take 15% less time |
| `geographic_societies` | Geographic Societies | optional | 126M kp | 2,340 | `cartography` | Opens the Geographic Society |
| `concert_of_nations` | Concert of Nations | capstone | 202M kp | 3,744 | `embassies`, `geographic_societies` | Opens the Grand Embassy, alliances give 25% more |
| `seed_drill` | Seed Drill | optional | 126M kp | 2,340 | `new_world_crops` | +6% food production |
| `industrialization` | Industrialization | **keystone** (Crystal Palace) | 101M kp | 2,340 | `steam_power` | +5% all production |
| `clockwork_automation` | Clockwork Automation | optional | 126M kp | 2,340 | `chronometry` | +10% game speed |
| `interchangeable_parts` | Interchangeable Parts | capstone | 202M kp | 3,744 | `industrialization`, `clockwork_automation` | Buildings cost 4% less, upgrades cost 15% less |
| `steam_pumps` | Steam Pumps | optional | 126M kp | 2,340 | `coke_smelting` | +6% iron ore production, +6% coal production |
| `rifling` | Rifling | optional | 126M kp | 2,340 | `gunpowder` | +12% military power |
| `steam_power` | Steam Power | spine | 75.8M kp | 1,872 | `steel_forging` | Opens the Coal Plant, +6% steel production, +6% coal production |

---

### Victorian Age (~1h 10m to 1h 27m/tech)

| Key | Name | Kind | Cost | Ticks | Prerequisites | Effect |
|---|---|---|---|---|---|---|
| `museums` | Museums | optional | 327M kp | 2,633 | `romanticism` | +5% culture production |
| `public_education` | Public Education | optional | 327M kp | 2,633 | `encyclopedia` | +5% knowledge production, research takes 3% less time |
| `telecommunications` | Telecommunications | optional | 327M kp | 2,633 | `electrification` | Deals refresh 30% sooner, gifts cost 25% less |
| `sanitation` | Sanitation | optional | 327M kp | 2,633 | `seed_drill` | +6% housing |
| `mass_production` | Mass Production | **keystone** (Eiffel Tower) | 262M kp | 2,633 | `industrialization` | Construction takes 8% less time |
| `geology` | Geology | optional | 327M kp | 2,633 | `steam_pumps` | +5% iron ore production, +5% coal production |
| `general_staff` | General Staff | optional | 327M kp | 2,633 | `rifling` | Campaigns take 15% less time |
| `electrification` | Electrification | spine | 196M kp | 2,106 | `industrialization` | Opens the Steam Works, +5% electricity production |

---

### Electric Age (~1h 18m to 1h 37m/tech)

| Key | Name | Kind | Cost | Ticks | Prerequisites | Effect |
|---|---|---|---|---|---|---|
| `radio` | Radio | optional | 644M kp | 2,925 | `telecommunications` | Festivals come back 20% sooner |
| `modern_physics` | Modern Physics | optional | 644M kp | 2,925 | `public_education` | +5% knowledge production |
| `wire_transfers` | Wire Transfers | optional | 644M kp | 2,925 | `telecommunications` | Market fee 2 points lower |
| `fertilizers` | Fertilizers | optional | 644M kp | 2,925 | `sanitation` | +5% food production |
| `assembly_line` | Assembly Line | optional | 644M kp | 2,925 | `mass_production` | Construction takes 5% less time |
| `chemical_engineering` | Chemical Engineering | spine | 386M kp | 2,340 | `mass_production` | +5% oil production, +5% steel production |
| `mechanized_warfare` | Mechanized Warfare | optional | 644M kp | 2,925 | `general_staff` | +10% military power |
| `power_distribution` | Power Distribution | **keystone** (Hoover Dam) | 515M kp | 2,925 | `electrification` | Opens the Dynamo Hall, +5% electricity production |
| `aviation` | Aviation | spine | 386M kp | 2,340 | `mass_production` | Expeditions take 10% less time |

---

### Atomic Age (~1h 33m to 3h 7m/tech)

| Key | Name | Kind | Cost | Ticks | Prerequisites | Effect |
|---|---|---|---|---|---|---|
| `cinema` | Cinema | optional | 625M kp | 3,510 | `radio` | +5% culture production |
| `big_science` | Big Science | capstone | 999M kp | 5,616 | `modern_physics` | Research takes 6% less time |
| `corporations` | Corporations | optional | 625M kp | 3,510 | `mercantilism`, `wire_transfers` | +5% gold production |
| `green_revolution` | Green Revolution | optional | 625M kp | 3,510 | `fertilizers` | +5% food production, +4% housing |
| `prefabrication` | Prefabrication | optional | 625M kp | 3,510 | `assembly_line` | Buildings cost 2% less, +5% storage |
| `plastics` | Plastics | optional | 625M kp | 3,510 | `chemical_engineering` | +5% oil production |
| `nuclear_deterrence` | Nuclear Deterrence | optional | 625M kp | 3,510 | `nuclear_fission`, `rocketry` | +10% military power, raids on you take 10% less |
| `military_industrial_complex` | Military-Industrial Complex | capstone | 999M kp | 5,616 | `mechanized_warfare`, `nuclear_deterrence` | Room for 20% more soldiers, campaigns bring back 20% more |
| `nuclear_fission` | Nuclear Fission | **keystone** (Particle Accelerator) | 500M kp | 3,510 | `power_distribution`, `chemical_engineering` | +5% uranium production, +5% electricity production |
| `civilian_reactors` | Civilian Reactors | optional | 625M kp | 3,510 | `nuclear_deterrence` | Opens the Nuclear Plant |
| `rocketry` | Rocketry | spine | 375M kp | 2,808 | `aviation` | +10% expedition rewards, +10% military power |

---

### Modern Age (~1h 33m to 1h 57m/tech)

| Key | Name | Kind | Cost | Ticks | Prerequisites | Effect |
|---|---|---|---|---|---|---|
| `television` | Television | optional | 593M kp | 3,510 | `cinema` | Opens the Monument of Ages |
| `information_theory` | Information Theory | optional | 593M kp | 3,510 | `modern_physics` | +5% knowledge production |
| `containerization` | Containerization | optional | 593M kp | 3,510 | `corporations` | Opens the Seaport, trade routes bring in 10% more |
| `suburbs` | Suburbs | optional | 593M kp | 3,510 | `green_revolution` | +5% housing |
| `nanofabrication` | Nanofabrication | optional | 593M kp | 3,510 | `computers` | Buildings cost 3% less |
| `titanium_alloys` | Titanium Alloys | optional | 593M kp | 3,510 | `chemical_engineering` | +5% steel production |
| `special_forces` | Special Forces | optional | 593M kp | 3,510 | `nuclear_deterrence` | Campaigns take 15% less time |
| `electricity_tech` | Advanced Electrics | spine | 356M kp | 2,808 | `nuclear_fission` | Opens the Power Grid Hub, +5% electricity production |
| `satellite_tech` | Satellite Technology | **keystone** (Space Program) | 474M kp | 3,510 | `rocketry`, `electricity_tech` | +1 data/tick, +10% expedition rewards |
| `computers` | Computers | spine | 356M kp | 2,808 | `electricity_tech` | +5% knowledge production, research takes 3% less time |

---

### Information Age (~1h 49m to 2h 16m/tech)

| Key | Name | Kind | Cost | Ticks | Prerequisites | Effect |
|---|---|---|---|---|---|---|
| `social_media` | Social Media | optional | 913M kp | 4,095 | `internet` | +5% culture production, festivals cost 20% less |
| `search_engines` | Search Engines | optional | 913M kp | 4,095 | `internet` | Research takes 3% less time |
| `e_commerce` | E-commerce | optional | 913M kp | 4,095 | `containerization`, `internet` | Opens the Container Terminal, market fee 2 points lower |
| `medical_nanobots` | Medical Nanobots | optional | 913M kp | 4,095 | `nanofabrication` | +5% housing, +5% food production |
| `internet_of_things` | Internet of Things | optional | 913M kp | 4,095 | `social_media`, `cybersecurity`, `medical_nanobots` | Opens the Smart Farm, opens the Smart Complex |
| `embedded_systems` | Embedded Systems | optional | 913M kp | 4,095 | `nanofabrication` | Construction takes 4% less time |
| `precision_mining` | Precision Mining | optional | 913M kp | 4,095 | `titanium_alloys` | +5% steel production |
| `cybersecurity` | Cybersecurity | spine | 548M kp | 3,276 | `computers` | Opens the Cyber Command, +10% military power |
| `smart_grid` | Smart Grid | optional | 913M kp | 4,095 | `electricity_tech` | Opens the Microgrid Array, +5% electricity production |
| `space_stations` | Space Stations | optional | 913M kp | 4,095 | `satellite_tech` | +8% expedition rewards |
| `internet` | Internet | **keystone** (Global Network) | 731M kp | 4,095 | `computers`, `satellite_tech` | +5% data production |

---

### Digital Age (~2h 36m to 4h 9m/tech)

| Key | Name | Kind | Cost | Ticks | Prerequisites | Effect |
|---|---|---|---|---|---|---|
| `virtual_reality` | Virtual Reality | optional | 1.34B kp | 4,680 | `social_media` | Morale can rise 5 points higher |
| `open_science` | Open Science | optional | 1.34B kp | 4,680 | `search_engines` | +5% knowledge production |
| `automated_logistics` | Automated Logistics | optional | 1.34B kp | 4,680 | `e_commerce` | Opens the Logistics Hub, trade routes take 15% less time |
| `global_village` | Global Village | capstone | 2.14B kp | 7,488 | `e_commerce`, `social_media` | Every civilization offers 1 more deal, alliances cost 50% less |
| `gene_editing` | Gene Editing | optional | 1.34B kp | 4,680 | `medical_nanobots` | +5% food production |
| `self_replication` | Self-Replication | optional | 1.34B kp | 4,680 | `medical_nanobots`, `machine_learning` | +10% nanobots production, construction takes 5% less time |
| `nano_alloys` | Nano Alloys | optional | 1.34B kp | 4,680 | `precision_mining` | +5% steel production |
| `drone_warfare` | Drone Warfare | optional | 1.34B kp | 4,680 | `cybersecurity` | +10% military power |
| `grid_storage` | Grid Storage | optional | 1.34B kp | 4,680 | `smart_grid` | Opens the Quantum Battery Array, +5% storage |
| `reusable_launchers` | Reusable Launchers | optional | 1.34B kp | 4,680 | `space_stations` | Expeditions take 10% less time |
| `machine_learning` | Machine Learning | **keystone** (World Simulation) | 1.07B kp | 4,680 | `internet`, `cybersecurity` | +5% data production, research takes 3% less time |
| `cloud_computing` | Cloud Computing | optional | 1.34B kp | 4,680 | `internet` | +8% storage |
| `general_ai` | General AI | capstone | 2.14B kp | 7,488 | `machine_learning`, `cloud_computing` | Research takes 6% less time |

---

### Cyberpunk Age (~2h 20m to 2h 55m/tech)

| Key | Name | Kind | Cost | Ticks | Prerequisites | Effect |
|---|---|---|---|---|---|---|
| `holography` | Holography | optional | 2.91B kp | 5,265 | `cybernetics`, `blockchain` | Opens the Holographic Theater |
| `neural_interface` | Neural Interface | spine | 1.75B kp | 4,212 | `machine_learning` | +4% knowledge production |
| `blockchain` | Blockchain | optional | 2.91B kp | 5,265 | `cybersecurity`, `cloud_computing` | Market fee 2 points lower |
| `synthetic_food` | Synthetic Food | optional | 2.91B kp | 5,265 | `gene_editing` | +4% food production |
| `cybernetics` | Cybernetics | **keystone** (Neon Citadel) | 2.33B kp | 5,265 | `neural_interface` | Opens the Combat Aug Center, +10% military power |
| `dark_crystal_mining` | Dark Crystal Mining | optional | 2.91B kp | 5,265 | `nano_alloys` | +4% dark matter crystals production |
| `augmented_soldiers` | Augmented Soldiers | optional | 2.91B kp | 5,265 | `cybernetics`, `drone_warfare` | +10% military power, room for 10% more soldiers |
| `dark_energy` | Dark Energy | optional | 2.91B kp | 5,265 | `grid_storage` | Opens the Dark Energy Tap, +4% electricity production |
| `lunar_outposts` | Lunar Outposts | optional | 2.91B kp | 5,265 | `reusable_launchers` | +8% expedition rewards |
| `darknets` | Darknets | optional | 2.91B kp | 5,265 | `cloud_computing` | +4% data production |

---

### Fusion Age (~2h 36m to 3h 15m/tech)

| Key | Name | Kind | Cost | Ticks | Prerequisites | Effect |
|---|---|---|---|---|---|---|
| `neural_art` | Neural Art | optional | 3.98B kp | 5,850 | `holography` | +4% culture production |
| `unified_theory` | Unified Theory | optional | 3.98B kp | 5,850 | `neural_interface`, `plasma_physics` | Research takes 3% less time |
| `maglev_transit` | Maglev Transit | optional | 3.98B kp | 5,850 | `superconductors` | Opens the Energy Exchange |
| `closed_biospheres` | Closed Biospheres | optional | 3.98B kp | 5,850 | `synthetic_food` | +5% housing |
| `molecular_assembly` | Molecular Assembly | optional | 3.98B kp | 5,850 | `cybernetics`, `self_replication` | Construction takes 5% less time |
| `superconductors` | Superconductors | spine | 2.39B kp | 4,680 | `plasma_physics` | +8% storage |
| `plasma_weapons` | Plasma Weapons | optional | 3.98B kp | 5,850 | `augmented_soldiers`, `plasma_physics` | +10% military power |
| `fusion_power` | Fusion Power | **keystone** (Stellar Cradle) | 3.18B kp | 5,850 | `nuclear_fission`, `cybernetics` | +4% electricity production, +4% plasma production |
| `plasma_physics` | Plasma Physics | spine | 2.39B kp | 4,680 | `fusion_power` | +4% plasma production |
| `fusion_drives` | Fusion Drives | optional | 3.98B kp | 5,850 | `fusion_power`, `lunar_outposts` | Expeditions take 10% less time |
| `quantum_networking` | Quantum Networking | optional | 3.98B kp | 5,850 | `darknets` | +4% data production |

---

### Space Age (~2h 51m to 5h 43m/tech)

| Key | Name | Kind | Cost | Ticks | Prerequisites | Effect |
|---|---|---|---|---|---|---|
| `overview_effect` | Overview Effect | optional | 3.67B kp | 6,435 | `neural_art` | Morale can rise 5 points higher |
| `deep_space_astronomy` | Deep Space Astronomy | optional | 3.67B kp | 6,435 | `unified_theory` | +4% knowledge production |
| `asteroid_claims` | Asteroid Claims | optional | 3.67B kp | 6,435 | `maglev_transit` | Trade routes bring in 10% more |
| `stellar_cartography` | Stellar Cartography | capstone | 5.87B kp | 10,296 | `asteroid_claims`, `deep_space_astronomy` | Expeditions take 20% less time, +10% expedition rewards |
| `hydroponics` | Hydroponics | optional | 3.67B kp | 6,435 | `closed_biospheres` | +4% food production |
| `zero_g_manufacturing` | Zero-G Manufacturing | spine | 2.2B kp | 5,148 | `orbital_mechanics`, `superconductors` | Construction takes 6% less time |
| `asteroid_refining` | Asteroid Refining | optional | 3.67B kp | 6,435 | `space_mining` | +4% titanium production |
| `orbital_defense` | Orbital Defense | optional | 3.67B kp | 6,435 | `plasma_weapons`, `orbital_mechanics` | Raids on you take 10% less |
| `orbital_solar` | Orbital Solar | optional | 3.67B kp | 6,435 | `plasma_physics` | +4% electricity production, +4% plasma production |
| `orbital_mechanics` | Orbital Mechanics | **keystone** (Dyson Scaffold) | 2.93B kp | 6,435 | `rocketry`, `plasma_physics` | +10% expedition rewards |
| `space_mining` | Space Mining | spine | 2.2B kp | 5,148 | `orbital_mechanics` | +4% titanium production, +4% steel production |
| `space_elevator` | Space Elevator | capstone | 5.87B kp | 10,296 | `zero_g_manufacturing`, `space_mining` | Buildings cost 4% less, +8% storage |
| `orbital_relays` | Orbital Relays | optional | 3.67B kp | 6,435 | `quantum_networking`, `orbital_mechanics` | +4% data production |

---

### Interstellar Age (~3h 7m to 3h 54m/tech)

| Key | Name | Kind | Cost | Ticks | Prerequisites | Effect |
|---|---|---|---|---|---|---|
| `void_contemplation` | Void Contemplation | optional | 5.8B kp | 7,020 | `overview_effect` | Appeasing a harbinger costs 20% less |
| `xenology` | Xenology | optional | 5.8B kp | 7,020 | `deep_space_astronomy` | Research takes 3% less time |
| `interstellar_trade` | Interstellar Trade | optional | 5.8B kp | 7,020 | `asteroid_claims`, `warp_drive` | Opens the Warp Commerce route |
| `protein_synthesis` | Protein Synthesis | optional | 5.8B kp | 7,020 | `hydroponics` | +4% food production, +4% housing |
| `hull_printing` | Hull Printing | optional | 5.8B kp | 7,020 | `zero_g_manufacturing` | Buildings cost 2% less, construction takes 4% less time |
| `stellar_core_mining` | Stellar Core Mining | optional | 5.8B kp | 7,020 | `asteroid_refining` | Opens the Stellar Core Drill, +4% titanium production |
| `fleet_doctrine` | Fleet Doctrine | optional | 5.8B kp | 7,020 | `orbital_defense` | +10% military power |
| `stellar_engineering` | Stellar Engineering | spine | 3.48B kp | 5,616 | `warp_drive` | Opens the Pulsar Tap, +4% plasma production, +4% electricity production |
| `warp_drive` | Warp Drive | **keystone** (Warp Nexus) | 4.64B kp | 7,020 | `space_mining`, `zero_g_manufacturing` | +10% expedition rewards, +4% dark matter production |
| `galactic_network` | Galactic Network | optional | 5.8B kp | 7,020 | `orbital_relays` | +4% data production |

---

### Galactic Age (~3h 7m to 3h 54m/tech)

| Key | Name | Kind | Cost | Ticks | Prerequisites | Effect |
|---|---|---|---|---|---|---|
| `galactic_memory` | Galactic Memory | optional | 5.67B kp | 7,020 | `void_contemplation` | +4% culture production |
| `cosmology` | Cosmology | optional | 5.67B kp | 7,020 | `xenology` | Research takes 3% less time |
| `federation_charter` | Federation Charter | optional | 5.67B kp | 7,020 | `interstellar_trade`, `xenology` | Alliances give 25% more, every civilization offers 1 more deal |
| `matter_conversion` | Matter Conversion | optional | 5.67B kp | 7,020 | `protein_synthesis` | +4% food production, +4% housing |
| `dyson_engineering` | Dyson Engineering | optional | 5.67B kp | 7,020 | `stellar_engineering`, `hull_printing` | Buildings cost 2% less |
| `neutron_mining` | Neutron Mining | optional | 5.67B kp | 7,020 | `stellar_core_mining` | Opens the Neutron Star Mine, +4% antimatter production |
| `armada_command` | Armada Command | optional | 5.67B kp | 7,020 | `fleet_doctrine` | +10% military power |
| `antimatter_synthesis` | Antimatter Synthesis | spine | 3.4B kp | 5,616 | `galactic_navigation` | Opens the Quasar Tap, +4% antimatter production |
| `galactic_navigation` | Galactic Navigation | **keystone** (Cosmic Beacon) | 4.54B kp | 7,020 | `warp_drive`, `stellar_engineering` | +4% dark matter production, +10% expedition rewards |
| `terraforming` | Terraforming | optional | 5.67B kp | 7,020 | `warp_drive` | +6% housing |
| `mind_uploading` | Mind Uploading | optional | 5.67B kp | 7,020 | `galactic_network` | +4% data production |

---

### Quantum Age (~3h 7m to 6h 14m/tech)

| Key | Name | Kind | Cost | Ticks | Prerequisites | Effect |
|---|---|---|---|---|---|---|
| `reality_art` | Reality Art | optional | 4.9B kp | 7,020 | `galactic_memory` | +4% culture production |
| `quantum_mechanics` | Quantum Mechanics | **keystone** (Reality Anchor) | 3.92B kp | 7,020 | `antimatter_synthesis` | +4% quantum flux production |
| `timeless_archive` | Timeless Archive | capstone | 7.84B kp | 11,232 | `cosmology`, `quantum_mechanics` | Research takes 6% less time |
| `probability_markets` | Probability Markets | optional | 4.9B kp | 7,020 | `federation_charter` | Market fee 2 points lower |
| `quantum_cultivation` | Quantum Cultivation | optional | 4.9B kp | 7,020 | `matter_conversion` | +4% food production |
| `reality_manipulation` | Reality Manipulation | spine | 2.94B kp | 5,616 | `quantum_mechanics` | +4% quantum flux production |
| `reality_engineering` | Reality Engineering | capstone | 7.84B kp | 11,232 | `reality_manipulation`, `dyson_engineering` | Buildings cost 4% less, construction takes 8% less time |
| `quantum_metallurgy` | Quantum Metallurgy | optional | 4.9B kp | 7,020 | `neutron_mining` | Opens the Quantum Metal Works, +4% antimatter production |
| `probability_warfare` | Probability Warfare | optional | 4.9B kp | 7,020 | `armada_command` | +10% military power, campaigns bring back 15% more |
| `zero_point_energy` | Zero-Point Energy | optional | 4.9B kp | 7,020 | `antimatter_synthesis` | Opens the Zero Point Generator |
| `wormholes` | Wormholes | optional | 4.9B kp | 7,020 | `galactic_navigation` | Trade routes take 20% less time, expeditions take 15% less time |
| `quantum_computing` | Quantum Computing | optional | 4.9B kp | 7,020 | `clockwork_automation`, `quantum_mechanics` | +15% game speed |

---

### Transcendent Age (~3h 54m)

| Key | Name | Kind | Cost | Ticks | Prerequisites | Effect |
|---|---|---|---|---|---|---|
| `transcendence` | Transcendence | **keystone** (Singularity Core) | 4.53B kp | 7,020 | `reality_manipulation` | +5% all production |
| `omniversal_exchange` | Omniversal Exchange | optional | 5.67B kp | 7,020 | `probability_markets` | Opens the Omniversal Bazaar |
| `singularity_engineering` | Singularity Engineering | optional | 5.67B kp | 7,020 | `reality_manipulation` | Opens the Singularity Engine |
| `omniversal_command` | Omniversal Command | optional | 5.67B kp | 7,020 | `probability_warfare` | Opens the Omniversal War Council |

---

## Tech Effects Reference

Every number here counts in full in every age: see [How Tech Bonuses Stack](#how-tech-bonuses-stack).

### Output of one resource

Each raises what one resource's buildings and workers make. Bonuses on the same resource add up.

| Resource | Techs | Together |
|---|---|---|
| Food | Fire Mastery +10%, Tool Making +10%, Animal Husbandry +10%, Boatbuilding +5%, Agriculture +10%, Irrigation +8%, The Plough +8%, Crop Rotation +6%, New World Crops +6%, Seed Drill +6%, Fertilizers +5%, Green Revolution +5%, Medical Nanobots +5%, Gene Editing +5%, Synthetic Food +4%, Hydroponics +4%, Protein Synthesis +4%, Matter Conversion +4%, Quantum Cultivation +4% | +119% |
| Knowledge | Language +10%, Primitive Writing +10%, Map Making +10%, Mathematics +8%, Philosophy +8%, Alchemy +8%, Printing Press +6%, Encyclopedia +6%, Public Education +5%, Modern Physics +5%, Information Theory +5%, Computers +5%, Open Science +5%, Neural Interface +4%, Deep Space Astronomy +4% | +99% |
| Steel | Blast Furnace +6%, Coke Smelting +6%, Steam Power +6%, Chemical Engineering +5%, Titanium Alloys +5%, Precision Mining +5%, Nano Alloys +5%, Space Mining +4% | +42% |
| Electricity | Electrification +5%, Power Distribution +5%, Nuclear Fission +5%, Advanced Electrics +5%, Smart Grid +5%, Dark Energy +4%, Fusion Power +4%, Orbital Solar +4%, Stellar Engineering +4% | +41% |
| Culture | Patronage +6%, Romanticism +6%, Museums +5%, Cinema +5%, Social Media +5%, Neural Art +4%, Galactic Memory +4%, Reality Art +4% | +39% |
| Iron | Bronze Working +10%, Iron Smelting +8%, Metal Casting +8%, Steel Forging +8% | +34% |
| Gold | Currency +10%, Road Building +8%, Banking +8%, Corporations +5% | +31% |
| Data | Internet +5%, Machine Learning +5%, Darknets +4%, Quantum Networking +4%, Orbital Relays +4%, Galactic Network +4%, Mind Uploading +4% | +30% |
| Faith | Ritual +10%, Calendar +10%, Theology +8% | +28% |
| Coal | Coke Smelting +6%, Steam Pumps +6%, Steam Power +6%, Geology +5% | +23% |
| Wood | Tool Making +10%, Woodworking +10% | +20% |
| Stone | Stoneworking +10%, Bronze Working +10% | +20% |
| Plasma | Fusion Power +4%, Plasma Physics +4%, Orbital Solar +4%, Stellar Engineering +4% | +16% |
| Titanium | Asteroid Refining +4%, Space Mining +4%, Stellar Core Mining +4% | +12% |
| Antimatter | Neutron Mining +4%, Antimatter Synthesis +4%, Quantum Metallurgy +4% | +12% |
| Iron ore | Steam Pumps +6%, Geology +5% | +11% |
| Oil | Chemical Engineering +5%, Plastics +5% | +10% |
| Nanobots | Self-Replication +10% | +10% |
| Dark matter | Warp Drive +4%, Galactic Navigation +4% | +8% |
| Quantum flux | Quantum Mechanics +4%, Reality Manipulation +4% | +8% |
| Uranium | Nuclear Fission +5% | +5% |
| Dark matter crystals | Dark Crystal Mining +4% | +4% |

### All production

| Tech | Age | Bonus |
|---|---|---|
| Industrialization | Industrial | +5% |
| Transcendence | Transcendent | +5% |

Together +10%, on top of each resource's own bonus. Nothing here joins the all-production pool that milestones and wonders fill, so that pool's soft cap never touches it.

### First sources

These two techs make a resource before any building of their age does. The amount is flat: neither the pools nor the tech layer multiply it.

| Tech | Age | Makes |
|---|---|---|
| Steel Forging | Medieval | +0.25 steel/tick |
| Satellite Technology | Modern | +1 data/tick |

### Storage and housing

| Tech | Age | Bonus |
|---|---|---|
| Pottery | Stone | +10% storage |
| Masonry | Bronze | +10% storage |
| Prefabrication | Atomic | +5% storage |
| Grid Storage | Digital | +5% storage |
| Cloud Computing | Digital | +8% storage |
| Superconductors | Fusion | +8% storage |
| Space Elevator | Space | +8% storage |
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

Storage bonuses add up to +54% of every store, and housing bonuses to +61%. Housing is rounded up to a whole person, so even the first +5% on a small village houses one more.

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
| Currency | Bronze | Market fee 3 points lower |
| Banking | Medieval | Market fee 3 points lower |
| Wire Transfers | Electric | Market fee 2 points lower |
| E-commerce | Information | Market fee 2 points lower |
| Blockchain | Cyberpunk | Market fee 2 points lower |
| Probability Markets | Quantum | Market fee 2 points lower |
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
| Military-Industrial Complex | Atomic | Room for 20% more soldiers |
| Augmented Soldiers | Cyberpunk | Room for 10% more soldiers |
| Military-Industrial Complex | Atomic | Campaigns bring back 20% more |
| Probability Warfare | Quantum | Campaigns bring back 15% more |
| Global Village | Digital | Every civilization offers 1 more deal |
| Federation Charter | Galactic | Every civilization offers 1 more deal |
| Global Village | Digital | Alliances cost 50% less |
| Void Contemplation | Interstellar | Appeasing a harbinger costs 20% less |

- **Market fee.** The market keeps 20 points of every trade (see [Trade](trade.md)). Currency, Banking, Wire Transfers, E-commerce, Blockchain and Probability Markets take 14 points off between them, so with all six the market keeps 6 and pays 17.5% more on every trade. The fee never falls under 5 points. Faction deals keep their edge over the market's rate as it stands.
- **Trade route time, expedition time, deal refresh, festival cooldown.** A cut shortens the timer when it is next set: a route's next run, the next expedition sent, the next set of offers, the next festival. Cuts multiply (two 15% cuts leave 72% of a route's time) and none takes a timer under 40% of its length: the four cuts of a route's time leave 49% of it, and the six of an expedition's 45%. Expedition time is for scouting expeditions, and the Army panel lists their times with it; campaigns have a number of their own.
- **Campaign time, campaign loot, soldier storage.** A General Staff and Special Forces each cut the time of every campaign sent after them by 15%. The Military-Industrial Complex makes every campaign bring back 20% more of its listed loot, won or lost, on top of your expedition rewards, and Probability Warfare 15% more again; the Military-Industrial Complex gives room for 20% more soldiers and Augmented Soldiers 10% more of that. See [Military](military.md).
- **Wonder construction.** With Architecture a wonder takes three quarters of its listed time to build, and the cuts of all construction then apply to what is left. It is the one cut a wonder has to itself. See [Wonders](wonders.md).
- **Festival length.** With Baroque Arts a festival's bonus lasts 25% longer; its price and the wait for the next are other numbers. See [Festival](commands.md#festival).
- **Gifts and alliances.** With Embassies a gift earns 22 opinion where it earned 15 (half as much again, rounded down). With the Concert of Nations every ally adds 25% more to its specialty, and with the Federation Charter 25% more again: a +20% bonus becomes +25%, then +31.25%, and the Factions panel lists it so. See [Factions](factions.md).
- **Deals and alliances.** The Global Village and the Federation Charter each add one deal to every set a civilization offers, and the Global Village halves what an alliance costs (250 gold).
- **Appease cost.** With Void Contemplation every Appease level costs 20% less, in every harbinger's thread, the Last Passage's included. See [Harbingers](harbinger.md).
- **Upgrade cost.** With Interchangeable Parts, `upgrade` asks for 15% less of every resource, after the cut of building costs on the new copy. Building new is not an upgrade.
- **Raid losses.** The cut comes off what a raid would take before your garrison meets it, so it stacks with the garrison's own share (see [Military](military.md)). Cuts multiply: with Imperial Legions, Fortification, Nuclear Deterrence and Orbital Defense a raid takes 0.9 × 0.85 × 0.9 × 0.9, about 62% of what it would.
- **Trade route income.** Boatbuilding, Containerization and Asteroid Claims each make every run of every route bring in 10% more of what the route lists, 33% with all three. A harbor's and an ally's share are added to that, each as its own share of the listed amount (see [Trade](trade.md)).
- **Morale ceiling.** Priesthood, Virtual Reality and the Overview Effect each let morale rise 5 points higher than 100% plus what your wonders add (see [Morale](morale.md)).
- **Hand gathering.** Every `gather` brings 2 more than you asked for, the bare `gather` (3) and the largest (25) alike.

### Game speed

Three techs raise game speed: ticks come more often, so production, construction, research and every timer run faster in real time. Game speed is a pool of its own, shared with milestone chain boosts and boons; the soft cap is not on it. The bonuses add together:

| Source | Bonus | Lasts |
|---|---|---|
| Chronometry (Medieval) | +5% | the run |
| Clockwork Automation (Industrial) | +10% | the run |
| Quantum Computing (Quantum) | +15% | the run |
| Milestone chain boosts | +250% or +300% | a few minutes, once per chain |
| Time Dilation boons | +8% to +15%, before scaling | 1,950 to 3,900 ticks |

With all three techs the game runs 30% faster, and that multiplies with research speed. See [Milestones](milestones.md) and [Boons](factions.md#boons). [Era Mastery](prestige.md#era-mastery) is a different thing: it leaves the tick alone and makes each tick of an age you know do more.

### Military power

Military power lowers mission difficulty and raises your defense rating. See [Military Power Bonus](military.md#military-power-bonus). The bonuses add together, with the military milestones' and with nothing to cap them.

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

Together +96%, added to the wonders' and milestones' expedition rewards.

---

## Paying for Research

Knowledge comes from the Knowledge lineage (Story Circle, Elders' Hall and on up to the Reality Academy), staffed by knowledge workers: **Shamans** in the Primitive Age, **Quantum Theorists** by the Quantum Age. Assign them with `assign <building_key> [count|all]`. The rates for every tier are on the [Knowledge](knowledge.md) page.

From the Atomic Age on a tech costs from 500M to a few billion (a capstone up to 3.47B), far more than the lineage makes: a fully staffed Research Campus makes about 102 knowledge a tick, under 6M over the whole Atomic Age. For those the market is the practical source: from the Industrial Age on, gold buys knowledge at a flat 5 knowledge per gold. See [Knowledge at the Market](knowledge.md#knowledge-at-the-market). `plan trade gold knowledge` buys it as gold comes in, and a `plan research` item behind it starts the tech once the knowledge is there.

Either way, your knowledge storage must hold a tech's full cost before you can pay for it.

---

## Strategy

### Raise Knowledge Output Early

Your first research bottleneck is knowledge income, not tick count. Rush `primitive_writing` → `mathematics` → `philosophy` to stack knowledge output bonuses in the first ages. Each percent you earn early pays off across every later tech.

### The Knowledge Snowball

Knowledge bonuses feed on themselves: each one makes the next tech arrive sooner. The chain looks like:
```
language → primitive_writing → mathematics → philosophy → printing_press → encyclopedia → …
```
Each of these raises knowledge output, so the next tech arrives sooner in real time. The steps are small (+4% to +10%), they add up to +99% with all fifteen, and no cap stops them (see [Output of one resource](#output-of-one-resource)).

### Game Speed: A Hidden Multiplier

`chronometry` (Medieval, no prerequisites) is one of the cheapest techs for what it does. +5% game speed runs everything that counts in ticks (production, building, research, expeditions) 5% faster. Research it early, then chain `clockwork_automation` in the Industrial Age for another +10%. Line them up with `plan research chronometry` as soon as you enter the Medieval Age.

### When to Cancel

Canceling costs you the full knowledge payment and the progress, with no refund. It only makes sense when a different tech gives a bonus you need now, badly enough to pay for the first one twice. As a rule: if you're more than halfway through the tick count, finish it, and put the urgent tech at the top of your plan so it starts next.

### The Grand Discovery Epoch Event

**The Grand Discovery**, a major good epoch event, instantly completes up to 3 unresearched techs of your current age or earlier, for free. It can come at an epoch transition when your culture strength is over 40% (see [Good Epoch Events](epochs.md#good-epoch-events)).

It doesn't weigh which techs are worth most: it takes the first three unresearched techs in alphabetical order of their keys, prerequisites or not. You can shape what it takes by researching the alphabetically early techs of your age first, which leaves the slots for the rest.

If Grand Discovery completes the tech you are researching, the research slot clears automatically (the knowledge you paid for it is not refunded).

### The Ancient Civilization Memory

A second way to skip the research grind comes only at the start of a new run, after a prestige or a Succumb, while you are still in the Primitive or Stone Age. An **ancient cache** has a ~40% chance (once per run) to offer one technology suited to your age that you haven't researched. Accepting it starts that tech at once, **free of prerequisites, the age requirement and knowledge cost**, but at **half research speed** (twice the normal tick count). The reachable tier scales with prestige level (one extra age of reach per two levels), so a high-prestige run can pull in a tech from an age it hasn't reached yet.

Unlike Grand Discovery, it never comes on your first-ever run: it needs prestige level 1 or higher. See [Prestige](prestige.md#ancient-civilization-memory) for full conditions.

### Late-Game Knowledge Scaling

Knowledge costs rise steeply, from 38 kp for Tool Making to billions in the late ages (Cybernetics 1.61B, the Timeless Archive 3.47B). Raise your knowledge storage ahead of them, and plan on buying most of that knowledge with gold (see [Paying for Research](#paying-for-research)).

---

## Tips and Common Mistakes

**Knowledge is deducted upfront.** Don't start a tech if your stockpile barely covers the cost. One bad event (The Dark Age cuts knowledge by 80% and cancels your active research) can set you back a long way.

**Prerequisites stack.** Before typing `research banking`, check that you have both `currency` and `mathematics`. `research list` shows only the techs you can start now; on the tech tree a tech that still waits has a dashed frame, and its card names what it builds on, with the techs you lack dimmed.

**The Dark Age epoch event** cancels your active research and drains 80% of your knowledge stockpile. If an epoch transition is close, consider whether to delay an expensive research start until after its event resolves.

**Prestige resets research** entirely, with every tech, its bonus and the milestones that gave research speed. What survives a prestige is **Ancient Knowledge** (research time ×0.8 per epoch) from Succumbing, faster research times in the ages a past run completed ([Era Mastery](prestige.md#era-mastery)), and, with the legacy kit's [Plan Template](prestige.md#plan-template), the techs you put in your build plan, which are planned again in the age you planned them in.

**Succumbing early is worth considering.** Succumbing to a catastrophe in the Iron Era (the earliest era one can strike in) costs you a run but cuts research time to ×0.8 permanently, and each further epoch you Succumb in multiplies it by 0.8 again. The fallen run's completed ages also gain an [Era Mastery](prestige.md#era-mastery) level, so the rebuild is quicker. Players who Succumb at least once begin each later run with faster research from tick one. See [Succumb](catastrophe.md#succumb).

**Don't overlook `civil_engineering`.** Buildings cost 3% less from the Classical Age on, for the rest of the run, and the cut multiplies with every build-cost milestone reward you earn after it.

---

*See also: [Knowledge](knowledge.md) for making and buying knowledge; [The Build Plan](plan.md) for queuing research; [Epochs](epochs.md) for how Grand Discovery and the Dark Age event fire; [Prestige](prestige.md) for the Plan Template and the Ancient Civilization Memory; [Buildings](buildings.md) for the buildings techs open.*
