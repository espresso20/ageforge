# First 10 Minutes

A guided walkthrough of your first game. The Primitive Age is tuned to take about **15 minutes** at 1x speed; follow this and you'll reach the Stone Age in about that time.

---

## Minute 0: You just launched AgeForge

You're in the **Primitive Age**. The screen shows:
- Status bar: `Primitive Age  Tick: 0  |  Pop: 0/0  Morale: 50%`
- Age progress bar: requirements for the Stone Age
- The Economy panel, which is always on screen
- On a large terminal (about 120x40 or bigger), a **mini map** above the Buildings list

The mini map draws your town from your real buildings, so it grows as you build. On smaller terminals it hides to leave the Buildings list room. Type `map` any time to open the full [Map](map.md) (Esc closes it), and `icons` if you want real icons on it.

You have no workers, no housing and very few resources. Start by gathering.

```
gather wood 25
gather food 25
```

Repeat until you have gathered enough wood to build what you need. Each `gather` grants up to **25** of a resource. ↑ recalls your last command.

Hand-gathering is an early-game crutch. It works through the **Medieval Age** and is disabled from the **Renaissance Age** onward. By then your buildings and workers should carry the economy.

---

## Step 1: Build a gathering camp and a wood camp

Food runs out before anything else, so get production going first. Both camps cost **16 wood**, and you start with 50.

```
build gathering_camp
build wood_camp
```

A gathering camp makes **+1.0 food/tick** and a wood camp about **+0.57 wood/tick** when fully staffed (3 workers each). An unstaffed building still runs at 20%, so a staffed camp makes **5x** what an empty one does. You'll staff them in Step 5.

---

## Step 2: Build housing

Each hut adds 10 housing, and the Stone Age wants 10 huts. More housing means more workers.

```
build hut
build hut
build hut
```

While those build (8 ticks each), watch the build queue progress bar in the Economy panel. Queue a new one as soon as each finishes and you have gathered enough resources to build more. **Aim for 5 huts in your first minute.**

---

## Step 3: Build a stash

Food and wood start with only 50 storage each. Storage fills fast, and anything past it is wasted. Your first stash costs **35 wood**; gather to that and build it early. Each stash gives **+500 storage** for every resource, and you can build up to 50.

```
build stash
```

Build two stashes alongside huts: you need room for the 1K food and 1K wood the Stone Age asks for.

---

## Step 4: Recruit workers

Once your huts are up, recruit some workers:

```
recruit 5
```

You don't pick a domain when recruiting. Workers become whatever you **assign** them to. Recruiting is free, but every worker eats food each tick from then on, so check that your food rate stays positive.

---

## Step 5: Assign everyone

This is the most important step. **Idle workers produce nothing and still eat food.**

```
assign gathering_camp 3
assign story_circle 1
assign shrine 1
```

A good split for the early game:
- 3 workers on each gathering_camp (food; 3 is a full crew)
- 2 workers on each story_circle (knowledge)
- 1 worker on the shrine (faith)
- Up to 3 workers on each wood_camp (wood)

Fill a building's slots before spreading workers thin: a fully staffed camp makes 5x an empty one.

A worker's class name comes from what it's assigned to: a worker on `gathering_camp` is a Forager, one on `story_circle` a Shaman. Type `workers` to see the breakdown by domain, or `status` for population, idle count and food drain.

The Workers box in the sidebar should show **Idle: 0**. If it doesn't, keep assigning.

---

## Step 6: Build story circles and a shrine

Story circles produce knowledge: **+0.2 knowledge/tick** each when staffed. The Stone Age needs **5 story circles** and **150 knowledge**, and knowledge is also what you spend on research.

```
build story_circle
```

Story circles cost 60 wood. Build all five early.

Also build your first **shrine**. Shrines produce faith, which helps prevent civilization-level disasters later on and feeds into epoch events.

```
build shrine
```

---

## Step 7: Watch your rates

The Economy panel is always on screen. Look at the rate column:
- `food: +N/t` should be positive (each staffed gathering camp adds +1.0; aim for at least +2)
- `wood: +N/t` should be positive (each staffed wood camp adds about +0.57)
- `knowledge: +N/t` should be positive (five staffed story circles make +1.0)

If food is negative, build another gathering camp and staff it. If knowledge is zero, assign workers to story_circle or build more of them.

---

## Step 8: Research your first tech

Once you have 800 knowledge (knowledge storage starts at 30, so you need stashes to hold that much):

```
research tool_making
```

This takes 56 ticks (just under 2 minutes) and gives a permanent +15% worker output. It costs more knowledge than the Stone Age asks for, so if you're short, advance first and research it later.

While that's researching, queue another building. Keep the build queue busy.

---

## Step 9: Build the Sacred Grove (required to advance)

Every age has a wonder, and you can't advance until it stands. The Sacred Grove costs:
- 1K wood
- 500 food

Bank the resources, then build it (75 ticks, 2m 30s):

```
wonder collect all
build sacred_grove
```

`wonder collect all` banks as much of each needed resource as you have. Run it again whenever your stock refills; your storage limits how much you can bank at once.

The grove also gives +0.02 knowledge/tick and +0.05 food/tick permanently.

---

## Step 10: Check the age bar and advance

The second row always shows what you need for the **next age**. For the Stone Age:
- Food: 1K
- Wood: 1K
- Knowledge: 150
- Huts: 10
- Story Circles: 5
- The Sacred Grove built

Keep building huts and story circles, keep assigning workers and keep knowledge flowing. When every bar is full, type `advance`. The age never advances on its own. If you're stepping away, `plan advance` queues it and the plan advances for you once everything is ready.

---

## Checklist at the Stone Age transition

Before you advance, you should have:
- [ ] 10 huts
- [ ] 5 story circles
- [ ] The Sacred Grove built
- [ ] 2-3 stashes
- [ ] Staffed gathering camps and wood camps
- [ ] 10+ workers, all assigned
- [ ] Food rate positive by at least +3/tick
- [ ] Tool Making researched (optional)

---

## What the Stone Age unlocks

When you reach the Stone Age:
- **Stone** resource unlocks (needed for the Bronze Age)
- **Stone Pit**: produces stone
- **Stone Camp**: early masonry building
- **Woodcutter Camp**: dedicated wood building
- **Forager Post**: upgraded food building
- **Standing Stones**: better faith building
- **Elders' Hall**: upgraded knowledge building
- **Longhouse**: bigger housing (+25 housing each)
- **War Camp**: early military building
- **Great Monolith**: the Stone Age wonder, required to reach the Bronze Age

Your first priority: build **Stone Pits** and **Woodcutter Camps**, and assign workers to them. The Stone Age is tuned to take about **45 minutes**, and the Bronze Age needs 4K food, 8K wood, 4K stone, 1.5K knowledge, 15 longhouses, 5 stone pits, 5 elders' halls, and the Great Monolith.

---

## Quick command reference for the first age

| What to do | Command |
|---|---|
| Build a gathering camp | `build gathering_camp` |
| Build a wood camp | `build wood_camp` |
| Build a hut | `build hut` |
| Build a stash | `build stash` |
| Build a story circle | `build story_circle` |
| Build a shrine | `build shrine` |
| Recruit workers | `recruit [count]` or `recruit max` |
| Assign workers to gathering camp | `assign gathering_camp 3` |
| Assign workers to story circle | `assign story_circle 1` |
| Unassign a worker from a building | `unassign gathering_camp 1` |
| Start first research | `research tool_making` |
| Bank resources for the wonder | `wonder collect all` or `wonder collect <resource> [amount]` |
| Build the wonder | `build sacred_grove` |
| Advance to the next age | `advance` |
| Check population, idle and food drain | `status` |
| Check worker breakdown | `workers` |
| Check logs | `logs` |
| Open the map | `map` |
| Queue builds for while you're away | `plan build hut 10`, `plan advance` |
| Save and return to the main menu | `Esc` (with no panel open) |

---

> **Tip:** The game is idle, so you don't need to babysit it. Set up your assignments, then leave a [build plan](plan.md) before you step away: `plan build hut 10`, `plan research fire_mastery`, `plan advance`. The plan starts each item as the resources come in, while you play and while you are away, and pays for it only when it starts.
