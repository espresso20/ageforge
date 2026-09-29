# How to Play

AgeForge is a **command-driven idle empire builder**. Resources accumulate passively, buildings construct over time, and you steer by typing commands at the `>` prompt.

---

## The core loop

```
Build → Recruit → Assign → Research → Advance → Repeat
```

1. **Build** structures to add housing, storage and production
2. **Recruit** workers to grow your population
3. **Assign** workers to buildings so they produce
4. **Research** technologies that multiply your output
5. **Advance** to the next age with `advance` once its requirements are met and the age's wonder is built
6. **Prestige** once you reach the Modern Age (or later) for permanent bonuses

---

## The screen layout

```
┌─ Status bar ──────────────────────────────────────────────────────────┐
│ Stone Age  Tick: 1247  |  Pop: 18/30  Morale: 54%  |  panel hint      │
├─ Age progress ────────────────────────────────────────────────────────┤
│ Next: Bronze Age  food:3102/4000 █████░  stone:890/4000 ███░░  ...    │
├─ Economy panel (PgUp / PgDn scroll) ─────────────┬─ Panels ───────────┤
│  Resources, buildings, build queue               │  panel names       │
│  ...                                             │  ...               │
│  Log                                             ├─ Workers ──────────┤
│  ...                                             │  pop, idle, drain  │
├──────────────────────────────────────────────────┴────────────────────┤
│ > _                                                                   │
└───────────────────────────────────────────────────────────────────────┘
```

---

## Navigation

- Type a panel's name to open it: `research`, `army`, `trade`, `factions`, `citymap`, `worldmap`, `help` and so on.
- `help` opens the Help panel: a full command reference plus the list of every panel you can open.
- **PgUp / PgDn** scroll the Economy panel.
- **↑ / ↓** step through your command history.
- **Esc** closes the open panel. With no panel open, it saves the game, stops it and returns to the main menu.
- Common commands have short names: `b` is `build`, `r` is `recruit`, `a` is `assign`, `s` is `status`, `t` is `trade`. See [Command shortcuts](commands.md#command-shortcuts) for the full list.

---

## Step-by-step: First 15 minutes

The Primitive Age is tuned to take about **15 minutes** at 1x speed, the Stone Age about **45 minutes**, and the Bronze Age about **1.5 hours**. See [How Long Each Age Takes](ages.md#how-long-each-age-takes) for the full curve.

### 1. Build a gathering camp and a wood camp
Food runs out before anything else. Get production going first:
```
build gathering_camp
build wood_camp
```
A fully staffed gathering camp makes **+1.0 food/tick**; a staffed wood camp about **+0.57 wood/tick**.

### 2. Build huts (housing)
You start with no housing, so you can't recruit anyone yet. Each hut gives +10 housing:
```
build hut
```
Queue another while the first is building. More housing means more workers.

### 3. Build stashes (storage)
Resources hit their storage limit quickly. Add storage:
```
build stash
```

### 4. Recruit your first workers
```
recruit 2
```
You don't pick a domain when you recruit. A worker takes the domain of the building you assign it to.

### 5. Assign food workers to a building
Idle workers produce nothing. Assign them:
```
assign gathering_camp 3
```
A staffed camp makes **5x** what an empty one does (an unstaffed building runs at 20%), so fill each camp's 3 slots.

### 6. Assign knowledge workers to a building
Knowledge workers produce knowledge. Build a story circle and assign one:
```
build story_circle
assign story_circle 1
```
A fully staffed story circle makes **+0.2 knowledge/tick**.

### 7. Research tool making
Once you have 800 knowledge:
```
research tool_making
```
This gives a permanent +15% worker output.

### 8. Build the age's wonder
Every age has a wonder, and you can't advance until it stands. For the Primitive Age that's the Sacred Grove (1K wood, 500 food):
```
wonder collect all
build sacred_grove
```
`wonder collect all` banks every resource the wonder still needs, as far as your stock goes. Run it again as your stock refills.

### 9. Watch the age bar and advance
The second row shows what you need for the next age. The Stone Age asks for 1K food, 1K wood, 150 knowledge, 10 huts and 5 story circles, plus the Sacred Grove. Keep building and assigning until the requirements fill up, then type `advance`. The age never advances by itself; if you'll be away, queue `plan advance` and the plan advances for you once everything is ready.

---

## Resource management tips

- Each resource has a storage limit. Once a resource is full, anything more it would produce is wasted.
- **Food drain:** every worker eats the same amount, whatever building it staffs. That amount is set by your age: 0.06 food/tick per worker in the Primitive Age, rising 12% with each age to about 0.58 in the Quantum Age. Keep food production above total drain. If food hits zero, one worker dies every 5 ticks (10 seconds at 1x) until food recovers.
- **Food** runs out first early on. Keep your gathering camps staffed before anything else, then build up knowledge.
- Watch the `Rate` column in the Economy panel. A negative rate is draining you.

## Morale

Morale is a percentage multiplier on **all worker output**: `production = base × count × (0.20 + 0.80 × assigned/capacity) × morale`. It **starts at 50%** (neutral), has a **10% floor**, and is capped at **100% + 5% per wonder** built.

The effect is a continuous curve centered on 50%. At exactly 50% production is normal. Above 50% production gets a bonus that grows steadily up to **+20%** at the cap. Below 50% it gets a penalty that deepens to **×0.50** at the 10% floor. Morale drifts back toward 50% each tick, so you have to keep a bonus going.

**What raises it:** worship and culture buildings, your faith production rate, good events, age advances.
**What lowers it:** starvation, military workers over 30% of population, idle workers over 50% of population, bad events and catastrophes.

**How to use it:** keep food positive, build worship and culture buildings to reach the **+20%** bonus, keep the army under 30% of population, and build wonders to raise the cap.

**Where to see it:** a colored bar in the Workers panel (`workers`) and the `Morale: NN%` figure in the status bar. Green means a bonus, red a penalty.

For the full morale system see [Morale](morale.md).

---

## Workers

Workers fall into 12 domains, each tied to specific buildings. Recruit workers with `recruit [count|max]` and assign them with `assign <building> [count|all]`. A worker takes the domain of the building it is assigned to.

**Core domains**: food, knowledge, faith, military, trade, engineering
**Production domains**: lumber, masonry, metallurgy, energy
**Late-game domains**: hacker, astronaut

Assigned workers raise a building's output. A building with no workers still produces at 20% (the floor), and a fully staffed one at 100%.

Only food workers make food, but every worker eats it. Keep food production above total worker drain.

---

## Building priorities by age

| Age | Priority buildings |
|---|---|
| Primitive | Gathering Camp, Wood Camp, Hut, Stash, Story Circle, Shrine, Sacred Grove |
| Stone Age | Stone Pit, Woodcutter Camp, Forager Post, Longhouse, Elders' Hall, Storage Pit, Great Monolith |
| Bronze Age | Farm, Lumber Mill, Quarry, Scriptorium, Market, Smithy, Warehouse, Stonehenge |
| Iron Age | Smelter, Agora, Trading Post, Hunting Lodge, Granary, Colosseum |
| Classical | Library, Military Academy, Merchant Quarter, Forge, Amphitheater, Aqueduct, Parthenon |
| Medieval | Guildhall, Castle Keep, Monastery Library, Cathedral, Great Library |

### Upgrading buildings after an age advance

Advancing an age does **not** convert your buildings. Instead, each one that has a newer version gets a pending upgrade marker, and a gold hint in the Economy panel names the target building. Use `upgrade <building>` to convert your existing copies to the new tier. For example, `upgrade gathering_camp` after entering the Stone Age turns your Gathering Camps into Forager Posts. An upgrade costs only the difference between the old and new price (with 50% of the old building's value credited back), so it is always cheaper than demolishing and rebuilding. Upgrade your food buildings first after every advance. Storage buildings never upgrade: the ones you built keep counting, so start on the new age's storage. See [Buildings](buildings.md#building-upgrades) for the full guide.

---

## What are wonders?

Wonders are unique mega-structures (22 total, one per age) that grant **permanent civilization bonuses**. You need your age's wonder to **advance** to the next age. Each one costs about the same share of its age's economy, and you **bank** the resources before you build it. Stonehenge, the Bronze Age wonder, needs about 34K stone, 19K wood and 3.4K iron:

```
wonder collect all
build stonehenge
```

`wonder collect all` banks as much of every needed resource as you have; repeat it until the bank is full, then build. Each wonder can be built only once. They are listed in the **Wonders** panel (`wonders`).

---

## Maps

There are two map views: a close-up of your own settlement and a zoomed-out view of the wider world.

**City Map** (`citymap`, or `map`) is a procedurally generated **top-down pixel-art city**. You look straight down at the roofs, streets and squares of one settlement drawn from your actual buildings. Built **wonders** anchor the center, and the city re-skins to the current era as you advance, from thatch huts on dirt lanes in the Primitive Age to mudbrick and then stone city walls from the Bronze to the Renaissance Age, and open grids and towers from the Industrial Age on. Every color comes from your active theme, so switching themes retints the whole city. There is no terrain on this view (that lives on the World Map), and every bit of greenery is built. See [The City Map](city-map.md) for how the city's look changes each age.

**World Map** (`worldmap`) shows a single **seeded world**: one continent with elevation, biomes, coastlines and rivers, the **same land every game** on your account. What changes each age is the **cartographic medium** it's drawn in: a charcoal cave-sketch (Primitive), inked parchment with a compass rose (Medieval), a satellite mosaic (Modern), a neon holo-grid (Cyberpunk), and so on through all 17 planetary ages. Once you leave the planet it becomes a **strategic star-map** of your empire and the rival civilizations, each colored by its stance toward you (at war red, ally green, mercantile gold, neutral steel blue). See [The World Map](world-map.md).

Open either map at any time by typing `citymap` or `worldmap` at the prompt.
