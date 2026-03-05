# How to Play

AgeForge is a **command-driven idle empire builder**. Resources accumulate passively, buildings construct over time, and you steer the direction by issuing commands at the `>` prompt.

---

## The core loop

```
Build → Recruit → Assign → Research → Advance → Repeat
```

1. **Build** structures to increase capacity and production
2. **Recruit** villagers to grow your workforce
3. **Assign** villagers to gather specific resources
4. **Research** technologies that multiply your output
5. **Advance** to the next age when requirements are met
6. **Prestige** when you reach the Transcendent Age

---

## The screen layout

```
┌─ Status bar ──────────────────────────────────────────────────────────────────┐
│ 🏛 Stone Age  [P0]  "Founder"         Tick: 1,247  Pop: 18/30  ×1  F1-F7 Esc │
├─ Age progress ────────────────────────────────────────────────────────────────┤
│ Next: Bronze Age  food:3102/8000 ████░░  stone:890/4000 ███░░  wood:1240/8000 │
├─ Tab bar ─────────────────────────────────────────────────────────────────────┤
│ F1:Economy  F2:Research  F3:Military  F4:Trade  F5:Stats  F6:Wonders  F7:Logs │
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

- **F1–F7** — switch tabs
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
Queue another while the first is building. Population cap unlocks more villagers.

### 2. Build stashes (storage)
Resources cap out quickly. Add storage:
```
build stash
```

### 3. Recruit your first workers
```
recruit food
recruit knowledge
```

### 4. Assign food workers to a building
Idle workers produce nothing. Assign them:
```
assign food gathering_camp 2
```

### 5. Assign knowledge workers to a building
Knowledge workers produce knowledge. Assign one:
```
assign knowledge story_circle 1
```

### 6. Build a gathering camp
Gathering camps boost food and wood production:
```
build gathering_camp
```

### 7. Research basic tools
Once you have enough knowledge (50 kp):
```
research basic_tools
```
This gives a permanent +20% worker output bonus.

### 8. Watch the age bar
The second row shows what you need for the next age. Keep building and assigning until the requirements fill up. The game advances automatically.

---

## Resource management tips

- Resources cap at their storage limit — once capped, production is wasted
- **Food drain** = total pop × 0.5/tick. Always keep food positive
- **Knowledge** is the most important resource early — prioritise knowledge workers
- Watch the `Rate` column in the Economy tab; negative rates will drain you

---

## Workers

Workers are organized into 12 domains, each tied to specific buildings. Recruit workers with `recruit <domain>` and assign them with `assign <domain> <building_key> <count>`.

**Core domains**: food, knowledge, faith, military, trade, engineering
**Production domains**: lumber, masonry, metallurgy, energy
**Late-game domains**: hacker, astronaut

Workers assigned to buildings boost production. A building at 0 assigned workers still produces at 20% efficiency (the floor). Full assignment = 100% efficiency.

Food workers are special — they produce food but all workers across all domains consume food per tick. Keep food production above total worker drain.

---

## Building priorities by age

| Age | Priority buildings |
|---|---|
| Primitive | Hut, Stash, Altar, Sacred Grove |
| Stone Age | Gathering Camp, Woodcutter Camp, Stone Pit, Firepit |
| Bronze Age | Farm, Lumber Mill, Quarry, Mine, Library |
| Iron Age | Coal Mine, Smithy, Barracks, Granary |
| Classical | Forum, Aqueduct, Amphitheater |
| Medieval | University, Cathedral, Castle |

---

## What are wonders?

Wonders are unique mega-structures (22 total) that grant **permanent civilization bonuses**. They cost enormous amounts of resources that you must **bank** before building:

```
wonder collect wood 500
wonder collect stone 500
build stonehenge
```

Each wonder can only be built once. They appear in the **F6: Wonders** tab.
