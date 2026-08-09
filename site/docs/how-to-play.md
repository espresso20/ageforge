# How to Play

AgeForge is a **command-driven idle empire builder**. Resources accumulate passively, buildings construct over time, and you steer the direction by issuing commands at the `>` prompt.

---

## The core loop

```
Build → Recruit → Assign → Research → Advance → Repeat
```

1. **Build** structures to increase capacity and production
2. **Recruit** workers to grow your workforce
3. **Assign** workers to gather specific resources
4. **Research** technologies that multiply your output
5. **Advance** to the next age when requirements are met
6. **Prestige** when you reach the Modern Age (or later) for permanent bonuses

---

## The screen layout

```
┌─ Status bar ──────────────────────────────────────────────────────────────────┐
│ 🏛 Stone Age  [P0]  "Founder"         Tick: 1,247  Pop: 18/30  ×1  Panels Esc │
├─ Age progress ────────────────────────────────────────────────────────────────┤
│ Next: Bronze Age  food:3102/8000 ████░░  stone:890/4000 ███░░  wood:1240/8000 │
├─ Tab bar ─────────────────────────────────────────────────────────────────────┤
│ Type a command name to open its overlay panel. Esc to close. │
├─ Tab content (scrollable) ────────────────────────────────────────────────────┤
│  Resources             │  Buildings                                           │
│  ...                   │  ...                                                 │
├─ Mini status ─────────────────────────────────────────────────────────────────┤
│ Recent logs            │ Pop/Idle/Food  Tech progress                         │
├───────────────────────────────────────────────────────────────────────────────┤
│ > _                                                                           │
└───────────────────────────────────────────────────────────────────────────────┘
```

---

## Navigation

- overlay commands (type `research`, `army`, `trade`, `factions`, `citymap`, `worldmap`, `help`, etc.) — open panels
- `help` opens the Help panel: a full command reference plus the list of every panel you can open
- **PgUp / PgDn** — scroll the active tab
- **↑ / ↓** — navigate command history
- **Esc** — save game
- Single-letter shortcuts: `e` `r` `m` `t` `s` `w` `l` switch tabs directly

---

## Step-by-step: First 15 minutes

### 1. Build huts (pop cap)
Your starting pop cap is too low. Build huts to raise it:
```
build hut
```
Queue another while the first is building. Population cap unlocks more workers.

### 2. Build stashes (storage)
Resources cap out quickly. Add storage:
```
build stash
```

### 3. Recruit your first workers
```
recruit 2
```
Workers are recruited generically — no domain needed. Assign them to buildings to give them a domain class.

### 4. Assign food workers to a building
Idle workers produce nothing. Assign them:
```
assign gathering_camp 2
```

### 5. Assign knowledge workers to a building
Knowledge workers produce knowledge. Assign one:
```
assign story_circle 1
```

### 6. Build a gathering camp
Gathering camps boost food and wood production:
```
build gathering_camp
```

### 7. Research tool making
Once you have enough knowledge (800 kp):
```
research tool_making
```
This gives a permanent +15% gather rate bonus.

### 8. Watch the age bar
The second row shows what you need for the next age. Keep building and assigning until the requirements fill up. The game advances automatically.

---

## Resource management tips

- Resources cap at their storage limit — once capped, production is wasted
- **Food drain** = `baseFoodCost × 1.12^tier /tick` per worker. Food domain workers start at 0.06/tick (Primitive Age) — very cheap. Other domains start at 1.0–32.0. Always keep food production positive — if food hits zero, workers die at 1 per 5 ticks
- **Knowledge** is the most important resource early — prioritise knowledge workers
- Watch the `Rate` column in the Economy tab; negative rates will drain you

## Morale

Morale is a percentage multiplier applied to **all worker output**: `production = base × count × (0.20 + 0.80 × assigned/cap) × morale`. It **starts at 50%** (neutral), bottoms out at a **10% floor**, and is capped at **100% + 5% per wonder** built.

It is a two-way dial with three bands: below **25%** penalises production (ramping down to **×0.50** at the floor), **25–75%** has no effect, and above **75%** rewards production (ramping up to **+20%**). Morale drifts back toward 50% each tick, so the bonus has to be sustained.

**What raises it:** worship & culture buildings, good events, age advances.
**What lowers it:** starvation, military workers over 30% of population, idle workers over 50% of population, bad events and catastrophes.

**Key levers:** Keep food positive, build worship/culture buildings to reach the **+20%** bonus, don't over-militarise, and build wonders to raise the cap.

**Where to see it:** A colored bar in the Workers panel and the villager sidebar — green when boosting, neutral in the middle, red when penalised.

For the full morale system see [Morale](morale.md).

---

## Workers

Workers are organized into 12 domains, each tied to specific buildings. Recruit workers with `recruit [count|max]` (no domain argument needed) and assign them with `assign <building_key> [count|all]`. Workers take on the domain class of the building they are assigned to.

**Core domains**: food, knowledge, faith, military, trade, engineering
**Production domains**: lumber, masonry, metallurgy, energy
**Late-game domains**: hacker, astronaut

Workers assigned to buildings boost production. A building at 0 assigned workers still produces at 20% efficiency (the floor). Full assignment = 100% efficiency.

Food workers are special — they produce food but all workers across all domains consume food per tick. Keep food production above total worker drain.

---

## Building priorities by age

| Age | Priority buildings |
|---|---|
| Primitive | Hut, Stash, Gathering Camp, Story Circle, Shrine, Sacred Grove |
| Stone Age | Stone Pit, Woodcutter Camp, Forager Post, Longhouse, Storage Pit |
| Bronze Age | Farm, Lumber Mill, Quarry, Scriptorium, Market, Smithy, Warehouse |
| Iron Age | Smelter, Hunting Lodge, Granary, Trading Post |
| Classical | Agora, Library, Forge, Amphitheater, Aqueduct |
| Medieval | University, Cathedral, Castle Keep, Great Library |

### Upgrading buildings after an age advance

Buildings are **not** automatically transformed when you advance an age. Instead, they gain a pending upgrade marker and a gold hint appears in the Economy tab showing the target building. Use `upgrade <building>` to convert your existing copies to the new tier — for example, `upgrade gathering_camp` after entering the Stone Age converts your Gathering Camps into Forager Posts. Each upgrade costs only the price difference between old and new (with 50% of the old building's value credited back), so it is always cheaper than demolishing and rebuilding. Upgrade your food and storage lineages first after every advance. See [Buildings](buildings.md#building-upgrades) for the full guide.

---

## What are wonders?

Wonders are unique mega-structures (22 total) that grant **permanent civilization bonuses**. They cost enormous amounts of resources that you must **bank** before building:

```
wonder collect wood 500
wonder collect stone 500
build stonehenge
```

Each wonder can only be built once. They appear in the **Wonders** overlay (`wonders`).

---

## Maps

There are two map views: a close-up of your own settlement and a zoomed-out view of the wider world. The **City Map** retints live with your active theme; the **World Map** is drawn in an era-specific cartographic medium that changes as you advance.

**City Map** (`citymap`, or `map`) — A theme-aware, procedurally generated **top-down pixel-art city**: you look straight down at the roofs, streets and squares of one living settlement, which re-skins to the current era as you advance. See [The City Map](city-map.md) for how the city's look evolves each age. The **streets are the gaps between wards** (a connected web of thin lanes); the heart is a dressed **town square**, and built **wonders** anchor the centre as dominant complexes (a **ziggurat** in the ancient ages, a **cathedral/keep** in the medieval ages). The city re-skins by age — thatch huts on dirt lanes (Primitive) → clay-tile mudbrick behind a **mudbrick wall with gates** (Bronze/Iron/Classical) → slate-roofed timber behind a **stone wall with towers** (Medieval/Renaissance) → open, wall-less grids and towers (Industrial and beyond). Every colour is drawn from your active theme, so switching themes retints the whole city; there is no terrain on this view (that lives on the World Map) and all greenery is built.

The map's structure changes by era — from a clustered organic village (Primitive/Stone), to roads radiating from the palace (Bronze/Iron/Classical), a walled castle with quarters (Medieval/Renaissance), a zoned road grid (Colonial/Industrial/Victorian), regular city blocks (Electric/Atomic/Modern), campus clusters (Information/Digital/Cyberpunk/Fusion), and finally concentric orbital rings (Space/Galactic/Quantum and beyond). Buildings are drawn as small 2.5D volumes (a lit roof, a shaded wall, and a drop shadow) so the city reads as dimensional, and each production lineage gets its own distinct colour.

**World Map** (`worldmap`) — A single **seeded world**: one continent with elevation, biomes, coastlines and rivers, the **same land every game** on your account. What changes each age is the **cartographic medium** it's drawn in — a charcoal cave-sketch (Primitive), inked parchment with a compass rose (Medieval), a satellite mosaic (Modern), a neon holo-grid (Cyberpunk), and so on through all 17 planetary ages. Once you leave the planet it becomes a **strategic star-map** of your empire versus the rival diplomacy factions, standings shown by signal colour (at-war red, ally green, mercantile gold, neutral steel-blue). See [The World Map](world-map.md). Open either map at any time by typing `citymap` or `worldmap` at the prompt.
