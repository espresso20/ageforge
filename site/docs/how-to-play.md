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

- overlay commands (type `research`, `army`, `trade`, etc.) — open panels
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

Morale is a multiplier applied to **all worker output**: `production = base × count × (0.20 + 0.80 × assigned/cap) × morale`. Default morale is 1.0; the floor is 0.10 (near standstill). The base cap is also 1.0, but each built wonder raises it by +0.05 — so building wonders is the only way to push worker output above 100%.

**What drives morale:**
- Food surplus: **+0.002/tick** while food production is positive
- Food deficit at zero food: **−0.005/tick**
- Military workers over 30% of population: **−0.003/tick per 10% over the threshold**
- Idle workers over 50% of population: **−0.002/tick**
- Passive recovery: **+0.001/tick** always
- Age advance: **+0.08** (one-time)

**Key levers:** Keep food in surplus, don't over-militarize (target soldiers below 25–28% of total population), and build wonders to raise the cap.

**Where to see it:** The morale bar appears in the villager panel. The status bar shows a color-coded percentage: green (≥ 80%), yellow (50–79%), red (< 50%).

For full morale details including tick driver math and event interactions, see [Workers & Domains](workers-and-domains.md).

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
