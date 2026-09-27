# First 10 Minutes

A guided walkthrough of your first game. The Primitive Age is tuned to take about **15 minutes** at 1x speed; follow this and you'll hit the Stone Age in about that time.

---

## Minute 0: You just launched AgeForge

You're in the **Primitive Age**. The screen shows:
- Top bar: `🏛 Primitive Age  Tick: 0  Pop: 1/6`
- Age progress bar: requirements for Stone Age
- Economy tab open by default

You have 0 workers and minimal resources. Let's fix that.

```
gather wood 25
gather food 25
``` 

Repeat until you have gathered enough wood to build a hut. Each `gather` grants up to **25** of a resource.

*hint* You can use the up arrow to recall previous commands and spam that gathering!

*note* Hand-gathering is an early-game crutch. It works through the **Medieval Age** and is disabled from the **Renaissance Age** onward — by then your buildings and workers should carry the economy.

---

## Step 1: Build a gathering camp and a wood camp

Food runs out before anything else, so get production going before you do anything fancy. Both camps cost **16 wood**, and you start with 50.

```
build gathering_camp
build wood_camp
```

A gathering camp makes **+1.0 food/tick** and a wood camp about **+0.57 wood/tick** when fully staffed (3 workers each). An unstaffed building still runs at 20%, so a staffed camp makes **5x** what an empty one does. You'll staff them in Step 5.

---

## Step 2: Build housing

Each hut adds 10 to your pop cap, and the Stone Age wants 10 of them. More pop means more workers.

```
build hut
build hut
build hut
```

While those build (8 ticks each), watch the build queue progress bar in the Economy tab. Queue a new one as soon as each finishes and you have gathered enough resources to build more!. **Aim for 5 huts in your first minute.**

---

## Step 3: Build a stash

Your food and wood caps are tiny (50 each). Storage fills fast and wastes production. Your first stash costs **35 wood** — gather to that and build it early to break the cap. Each stash gives **+500 storage** and you can build up to 50 before they're capped out.

```
build stash
```

Build two stashes alongside huts: you need room for the 1,000 food and 1,000 wood the Stone Age asks for.

---

## Step 4: Recruit workers

Once your huts are up (pop cap raised), recruit some workers:

```
recruit 5
```

Workers are recruited generically — there's no domain when recruiting. They become what you **assign** them to. Each worker costs food upfront — check your food rate stays positive after recruiting.

---

## Step 5: Assign everyone

This is the most important step. **Idle workers produce nothing and still consume food.**

```
assign gathering_camp 3
assign story_circle 1
assign shrine 1
```

General rule for early game:
- 3 workers → each gathering_camp (food production; 3 is a full crew)
- 2 workers → each story_circle (knowledge)
- 1 worker → shrine (faith)
- Up to 3 workers → each wood_camp (wood production)

Fill a building's slots before spreading workers thin: a fully staffed camp makes 5x an empty one.

Workers derive their role from what they're assigned to — a worker on `gathering_camp` becomes a Forager; one on `story_circle` becomes a Shaman. Check the Economy tab (type `status` or view the Economy tab) to see your worker breakdown by domain.

Check the status bar shows **Idle: 0**. If it doesn't, keep assigning.

---

## Step 6: Build story circles and a shrine

Story circles produce knowledge: **+0.2 knowledge/tick** each when staffed. The Stone Age needs **5 story circles** and **150 knowledge**, and knowledge is also what lets you research techs.

```
build story_circle
```

Story circles cost 60 wood — build all five early.

Also build your first **shrine**. Shrines produce faith, which helps prevent long-term civilisation-level disasters and contributes to epoch events.

```
build shrine
```

---

## Step 7: Watch your rates

Open the Economy tab (or just wait — you're already there). Look at the rate column:
- `food: +N/t` — should be positive (each staffed gathering camp adds +1.0; aim for at least +2)
- `wood: +N/t` — should be positive (each staffed wood camp adds about +0.57)
- `knowledge: +N/t` — should be positive (five staffed story circles make +1.0)

If food is negative, build another gathering camp and staff it. If knowledge is zero, assign workers to story_circle or build more of them.

---

## Step 8: Research your first tech

Once you hit 800 knowledge (watch the knowledge bar; knowledge storage starts at 30, so you need stashes to hold that much):

```
research tool_making
```

This takes 56 ticks (just under 2 minutes) and gives +15% gather rate — permanently. It costs more knowledge than the Stone Age asks for, so if you're short, advance first and research it later.

While that's researching, queue another building — keep the build queue busy constantly.

---

## Step 9: Build the Sacred Grove (required to advance)

Every age has a wonder, and you can't advance until it stands. The Sacred Grove costs:
- 1,000 wood
- 500 food

Bank the resources, then build it (75 ticks, 2m 30s):

```
wonder collect wood 1000
wonder collect food 500
build sacred_grove
```

It also gives +0.02 knowledge/t and +0.05 food/t permanently. Bank in chunks as your storage allows.

---

## Step 10: Check the age bar

The second row always shows what you need for the **next age**. For Stone Age:
- Food: 1,000
- Wood: 1,000
- Knowledge: 150
- Huts: 10
- Story Circles: 5
- The Sacred Grove built

Keep building huts and story circles. Keep assigning workers. Keep knowledge flowing. The age advances **automatically** when all bars fill.

---

## Checklist at the Stone Age transition

Before you advance, you should have:
- [ ] 10 huts
- [ ] 5 story circles
- [ ] The Sacred Grove built
- [ ] 2–3 stashes
- [ ] Staffed gathering camps and wood camps
- [ ] 10+ pop, all assigned
- [ ] Food rate positive by at least +3/t
- [ ] Tool Making researched (optional)

---

## What the Stone Age unlocks

When you reach Stone Age:
- **Stone** resource unlocks (needed for Bronze Age)
- **Stone Pit** — produces stone passively
- **Stone Camp** — early masonry production building
- **Woodcutter Camp** — dedicated wood building
- **Forager Post** — upgraded food building
- **Standing Stones** — better faith building
- **Elders' Hall** — upgraded knowledge building
- **Longhouse** — bigger housing (+25 pop cap each)
- **War Camp** — early military building
- **Great Monolith** — the Stone Age wonder, required to reach the Bronze Age

Your first priority: build **Stone Pits** and **Woodcutter Camps**, and assign workers to them. The Stone Age is tuned to take about **45 minutes**, and the Bronze Age needs 4,000 food, 8,000 wood, 4,000 stone, 1,500 knowledge, 15 longhouses, 5 stone pits, 5 elders' halls, and the Great Monolith.

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
| Bank resources for the wonder | `wonder collect <resource> <amount>` |
| Build the wonder | `build sacred_grove` |
| Check economy tab | `status` |
| Check worker breakdown | `status` |
| Check logs | `logs` |
| Queue builds for while you're away | `plan build hut 10`, `plan advance` |
| Save the game | `Esc` |

---

> **Tip:** The game is idle — you don't need to babysit it. Set up your assignments, then leave a [build plan](plan.md) before you step away: `plan build hut 10`, `plan research fire_mastery`, `plan advance`. The plan starts each item as the resources come in, while you play and while you are away, and pays for it only when it starts.
