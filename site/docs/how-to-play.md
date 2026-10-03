# How to Play

AgeForge is a **command-driven idle civilization game**. Resources accumulate on their own, buildings construct over time, and you steer by typing commands at the `>` prompt. A first run takes about a week of real time, so the game is built for checking in a few times a day.

---

## The core loop

```
Build → Staff → Research → Advance → Repeat
```

1. **Build** structures to add housing, storage and production
2. **Staff** them: workers arrive on their own while housing and food allow and go to work by your [worker shares](workers-and-domains.md#worker-shares). Steer the split with `workers share`, or `recruit` and `assign` by hand
3. **Research** technologies that multiply your output
4. **Advance** to the next age with `advance` once its requirements are met and the age's wonder is built
5. **Prestige** once you reach the Modern Age (or later) for permanent upgrades. Every age the run completed then runs faster on the next run ([Era Mastery](prestige.md#era-mastery))

New to the game? [Your First Age](first-ten-minutes.md) walks you through the Primitive Age step by step, from the first `gather` to the Sacred Grove.

---

## The screen layout

```
┌──────────────────────────────────────────────────────────────────────────┐
│ Name · Stone Age  ◈ Stone Era  |  Pop: 18/30  Morale 54%  |  hint        │  status bar
│ Next Age: Bronze Age  ✗ Food 3.1K/4K  ✓ Knowledge 1.5K/1.5K  ✗ Stone ... │  next age row
├─ Resources ──────────┬─ Buildings ──────────────────────┬─ Panels ───────┤
│ amounts and rates    │ (mini map on a large terminal)   │ panel names    │
├─ Under construction ─┤ what you can build, what you own │ ...            │
│ progress bars        │ PgUp / PgDn scroll               │                │
├─ Log ────────────────┤                                  ├─ Workers ──────┤
│ ...                  │                                  │ pop, idle,     │
│                      │                                  │ housing, food  │
├─ Command ────────────┴──────────────────────────────────┴────────────────┤
│ > _                                                                      │
└──────────────────────────────────────────────────────────────────────────┘
```

The status bar also carries your prestige level and civilization title once you have them, and badges for a waiting harbinger or catastrophe. The Next Age row marks each requirement ✓ or ✗, and names the age's wonder until it stands. Panels you open (`research`, `trade`, `map` and so on) open over the dashboard, and the prompt keeps working while they are open.

---

## Navigation

- Type a panel's name to open it: `research`, `army`, `trade`, `factions`, `map`, `help` and so on.
- `help` opens the Help panel: a full command reference plus the list of every panel you can open.
- **Tab** or **→** takes the completion shown in dim text after the cursor; press **Tab** again for the next one.
- **PgUp / PgDn** scroll the Buildings list.
- **↑ / ↓** step through your command history.
- **Esc** closes the open panel. With no panel open, it saves the game, stops it and returns to the main menu.
- Common commands have short names: `b` is `build`, `r` is `recruit`, `a` is `assign`, `s` is `status`, `t` is `trade`. See [Command shortcuts](commands.md#command-shortcuts) for the full list.

---

## A week-long first run

A first run takes about **a week of real time** to reach the Modern Age, where you can prestige for the first time. The first hour is quick: the Primitive Age takes about 15 minutes and the Stone Age about 45. From the Bronze Age on, every age takes hours: the Bronze Age 3h 54m, the Iron Age 6h 30m, and longer from there. See [How Long Each Age Takes](ages.md#how-long-each-age-takes) for the full curve. A tick is 2 seconds of real time.

That is the first run. After a prestige, every age the run completed runs faster the next time, twice as fast after one completion and up to 4.2x after ten, so later runs reach the Modern Age much sooner (see [Era Mastery](prestige.md#era-mastery)).

You don't need to watch it. Check in a few times a day, spend what has built up, and set the game up for the hours you're away:

- **Offline progress.** When you load a save, the game credits the time since you saved it, up to **24 hours**, at **50%** of your production. It runs that time in steps, so construction and research finish, the build plan starts what the income pays for, and your worker shares recruit and staff as the hours pass. The welcome-back lines in the log say what happened.
- **The build plan.** Queue up to 60 builds, techs, trades and an advance with `plan build`, `plan research`, `plan trade` and `plan advance`. The game starts each item, in order, as the resources come in, and pays for it only when it starts. See [The Build Plan](plan.md).
- **Overflow.** Production a full store would waste goes into your age's wonder bank (`wonder overflow`, on by default), and what the wonder doesn't need goes toward the plan's queued copies.
- **Storage.** From the Bronze Age on, an age's storage built out in full holds about 4.5 hours of that age's typical income. Past that, anything overflow doesn't take is lost, so build storage before a long absence.

---

## Catastrophes and the harbinger

From the Iron Era on, each era may hide a fated catastrophe. It can strike at any moment in the era, but never unannounced: a **harbinger** comes first, with a `⚑ Harbinger` badge on the status bar. Type `harbinger` to read it, and answer if you like: **Appease** lowers the odds, **Brace** softens the blow, **Invite** makes it certain. When a catastrophe strikes, nothing is lost until you choose: **Endure** (lose part of your buildings, stores and workers, and keep your age) or **Succumb** (fall back to the Primitive Age and keep a permanent legacy bonus). Higher faith lowers the odds of a strike. See [Catastrophes](catastrophe.md) and [The Harbinger](harbinger.md).

---

## Resource management tips

- Each resource has a storage limit. Once a resource is full, anything more it would produce is wasted, unless wonder overflow banks it.
- **Food drain:** every worker eats the same amount, whatever building it staffs. That amount is set by your age: 0.06 food/tick per worker in the Primitive Age, rising 12% with each age. Keep food production above total drain. If food hits zero, one worker dies every 5 ticks (10 seconds) until food recovers. See [Food drain](workers-and-domains.md#food-drain) and [Starvation](workers-and-domains.md#starvation).
- **Food** runs out first early on. Keep your gathering camps staffed before anything else, then build up knowledge.
- Watch the rates in the Resources box. A negative rate is draining you.

## Morale

Morale is a percentage that multiplies **all worker output**. It **starts at 50%**, which is neutral. Above 50% your workers get a bonus that grows to **+20%** at the morale cap (100%, plus 5% for every wonder you build); below 50% they take a penalty that deepens to **x0.50** at the 10% floor. Morale drifts back toward 50% each tick, so a bonus has to be kept up.

Worship and culture buildings, your faith income, good events and age advances raise it. Starvation, a large army, many idle workers, bad events and catastrophes lower it. It shows as `Morale NN%` on the status bar, green for a bonus and red for a penalty, and as a bar in the Workers panel. See [Morale](morale.md) for the full system.

---

## Workers

Workers fall into 12 worker domains (food, lumber, masonry, knowledge, faith and so on), each tied to its own buildings. A worker takes the domain of the building it works in. Workers arrive on their own while housing and food allow and go to work by your worker shares (`workers share`; every domain is on auto by default). You can also recruit with `recruit [count|max]` and assign with `assign <building> [count|all]`.

Assigned workers raise a building's output. A building with no workers still produces at 20% (the floor), and a fully staffed one at 100%. Only food workers make food, but every worker eats it. See [Workers & Domains](workers-and-domains.md) for the full guide.

---

## Building priorities by age

| Age | Priority buildings |
|---|---|
| Primitive | Gathering Camp, Wood Camp, Hut, Stash, Story Circle, Shrine, Sacred Grove |
| Stone Age | Stone Camp, Stone Pit, Woodcutter Camp, Forager Post, Longhouse, Elders' Hall, Storage Pit, Great Monolith |
| Bronze Age | House, Farm, Lumber Mill, Quarry, Smithy, Scriptorium, Market, Warehouse, Stonehenge |
| Iron Age | Townhouse, Field Works, Smelter, Trading Post, Agora, Granary, Colosseum |
| Classical | Villa, Terrace Farm, Library, Merchant Quarter, Forge, Aqueduct, Classical Vault, Parthenon |
| Medieval | Manor, Demesne, Guildhall, Monastery Library, Cathedral, Strongroom, Great Library |

### Upgrading buildings after an age advance

Advancing an age does **not** convert your buildings. Instead, each one that has a newer version gets a pending upgrade marker, and a gold hint in the Economy panel names the target building. Use `upgrade <building>` to convert your existing copies to the new tier. For example, `upgrade gathering_camp` after entering the Stone Age turns your Gathering Camps into Forager Posts. An upgrade costs the new building's price minus what the old copy would sell for (half what it cost), takes effect at once and keeps the building's workers. Upgrade your food buildings first after every advance. Storage buildings never upgrade: the ones you built keep counting, so start on the new age's storage. See [Buildings](buildings.md#building-upgrades) for the full guide.

---

## What are wonders?

Wonders are unique mega-structures (22 wonders in all, one per age) that grant **permanent civilization bonuses**. You need your age's wonder to **advance** to the next age. Each one costs about the same share of its age's economy, and you **bank** the resources before you build it. Stonehenge, the Bronze Age wonder, needs about 34K stone, 19K wood and 3.4K iron:

```
wonder collect all
build stonehenge
```

`wonder collect all` banks as much of every needed resource as you have; repeat it until the bank is full, then build. Wonder overflow fills the bank too, from production your full stores would waste. Each wonder can be built only once. They are listed in the **Wonders** panel (`wonders`). See [Wonders](wonders.md).

---

## The Map

Type `map` to open the **Map**: your empire drawn from your real game state (your buildings and the workers staffing them, your wonders, the civilizations you have met, your trade routes). It has two styles:

- **Roguelike** (the default): a glyph world seen from above. Your town grows out of your real buildings, quarter by quarter, with streets, walls and wonders. It has three zooms: **PgUp** zooms out to the region, the known world with the civilizations you have met, and **PgDn** zooms in to a district with building names.
- **Skyline**: your empire side-on as a panorama, one district for every age you have lived through. Windows light only where workers are staffed, and trade routes travel by land, sea or air. `←` and `→` scroll it, and PgUp and PgDn move half a screen.

Switch with `map style skyline` or `map style roguelike` (or the shorter `style skyline`). `map glyphs` chooses the glyph set; both are saved with your account.

The prompt keeps working while the Map is open. Letters you type go to the prompt, and the Map takes only the keys that print nothing: the arrows, PgUp and PgDn, Home and End, and, while the prompt is empty, Tab, Shift+Tab and Enter. In both styles an inspect cursor tells you what it is on and the command to type for it; Tab jumps between buildings and wonders, and Enter puts that command into the prompt. Esc closes the Map. On a large terminal the dashboard also shows a **mini map** above the Buildings list. `citymap` and `worldmap` still work and open the Map. See [The Map](map.md) for all the keys.
