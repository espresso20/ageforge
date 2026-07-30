# First 10 Minutes

A guided walkthrough of your first game. Follow this and you'll hit the Stone Age within 10–15 minutes of real time.

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

## Step 1: Build housing immediately

Your pop cap is 10 per hut. You need more. Build huts first — realize those worker gains.

```
build hut
build hut
build hut
```

While those build (8 ticks each), watch the build queue progress bar in the Economy tab. Queue a new one as soon as each finishes and you have gathered enough resources to build more!. **Aim for 5 huts in your first minute.**

---

## Step 2: Build a stash

Your food and wood caps are tiny (50 each). Storage fills fast and wastes production. Your first stash costs **35 wood** — gather to that and build it early to break the cap. Each stash gives **+500 storage** and you can build up to 50 before they're capped out.

```
build stash
```

Build one or two stashes alongside huts.

---

## Step 3: Recruit workers

Once your huts are up (pop cap raised), recruit some workers:

```
recruit 5
```

Workers are recruited generically — there's no domain when recruiting. They become what you **assign** them to. Each worker costs food upfront — check your food rate stays positive after recruiting.

---

## Step 4: Assign everyone

This is the most important step. **Idle workers produce nothing and still consume food.**

```
assign gathering_camp 3
assign story_circle 1
assign shrine 1
```

General rule for early game:
- 3–4 workers → gathering_camp (food production)
- 1–2 workers → story_circle (knowledge)
- 1 worker → shrine (faith)
- If you have a wood_camp, assign 1–2 there

Workers derive their role from what they're assigned to — a worker on `gathering_camp` becomes a Forager; one on `story_circle` becomes a Shaman. Check the Economy tab (type `status` or view the Economy tab) to see your worker breakdown by domain.

Check the status bar shows **Idle: 0**. If it doesn't, keep assigning.

---

## Step 5: Build a story_circle and a shrine

Story circles produce knowledge passively. Knowledge is what lets you research techs that multiply everything.

```
build story_circle
```

Story circles are cheap — build as many as you can afford early.

Also build your first **shrine**. Shrines produce faith, which helps prevent long-term civilisation-level disasters and contributes to epoch events.

```
build shrine
```

---

## Step 6: Watch your rates

Open the Economy tab (or just wait — you're already there). Look at the rate column:
- `food: +N/t` — should be positive (at least +2)
- `wood: +N/t` — should be positive (at least +1)
- `knowledge: +N/t` — should be positive (at least +1)

If food is negative, assign more workers to gathering_camp. If knowledge is zero, assign workers to story_circle or build more of them.

---

## Step 7: Research your first tech

Once you hit 800 knowledge (watch the knowledge bar):

```
research tool_making
```

This takes 300 ticks but gives +15% gather rate to all workers — permanently. It pays for itself within a minute.

While that's researching, queue another building — keep the build queue busy constantly.

---

## Step 8: Build a sacred grove (optional early wonder)

The Sacred Grove wonder is surprisingly cheap:
- 8,000 wood
- 5,000 food

It gives +knowledge/t and +food/t permanently. If you're ahead on resources, start banking for it:

```
wonder collect wood 1000
wonder collect food 500
```

Don't force it — just bank when you have surplus.

---

## Step 9: Check the age bar

The second row always shows what you need for the **next age**. For Stone Age:
- Food: 16,000
- Wood: 10,000
- Knowledge: 2,800
- Huts: 20
- Story Circles: 5

Keep building huts and story circles. Keep assigning workers. Keep knowledge flowing. The age advances **automatically** when all bars fill.

---

## Checklist at the Stone Age transition

Before you advance, you should have:
- [ ] 20 huts
- [ ] 5 story circles
- [ ] 3–4 stashes
- [ ] 10+ pop, all assigned
- [ ] Tool Making researched
- [ ] Food rate positive by at least +3/t

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
- **Longhouse** — bigger housing (+20 pop cap each)
- **War Camp** — early military building
- **Great Monolith** — first major wonder

Your first priority: build **Stone Pits** and **Woodcutter Camps**, and assign workers to them. Bronze Age needs a lot of stone and wood.

---

## Quick command reference for the first age

| What to do | Command |
|---|---|
| Build a hut | `build hut` |
| Build a stash | `build stash` |
| Build a story circle | `build story_circle` |
| Build a shrine | `build shrine` |
| Recruit workers | `recruit [count]` or `recruit max` |
| Assign workers to gathering camp | `assign gathering_camp 3` |
| Assign workers to story circle | `assign story_circle 1` |
| Unassign a worker from a building | `unassign gathering_camp 1` |
| Start first research | `research tool_making` |
| Check economy tab | `status` |
| Check worker breakdown | `status` |
| Check logs | `logs` |
| Save the game | `Esc` |

---

> **Tip:** The game is idle — you don't need to babysit it. Set up your assignments, queue a few buildings, start a research, then step away for a few minutes and come back to see the progress.
