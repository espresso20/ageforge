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
gather wood 5
gather food 5
``` 

repeat until you have gathered enough wood to build a hut

---

## Step 1: Build housing immediately

Your pop cap is 6. You need more. Build huts first — realize those worker gains.

```
build hut
build hut
build hut
```

While those build (3 ticks each), watch the build queue progress bar in the Economy tab. Queue a new one as soon as each finishes and you have gathered enough resources to build more!. **Aim for 5 huts in your first minute.**

---

## Step 2: Build a stash

Your food and wood caps are tiny (50 each). Storage fills fast and wastes production. Each stash gives **+300 storage** and you can build up to 30 before they're capped out.

```
build stash
```

Build one or two stashes alongside huts.

---

## Step 3: Recruit workers

Once your huts are up (pop cap raised), recruit:

```
recruit food
recruit food
recruit knowledge
```

Keep recruiting until you're at 8–10 pop. Each worker costs food upfront — check your food rate stays positive.

---

## Step 4: Assign everyone

This is the most important step. **Idle workers produce nothing.**

```
assign gathering_camp 2
assign gathering_camp 1
assign story_circle 1
```

General rule for early game:
- Half your workers → food (assign to gathering_camp)
- Remaining workers → wood (also assign to gathering_camp for now)
- All knowledge workers → story_circle

Check the status bar shows **Idle: 0**. If it doesn't, assign the rest.

---

## Step 5: Build an altar

Altars produce knowledge passively. Knowledge is what lets you research techs that multiply everything.

```
build altar
build altar
```

Altars are cheap. Build as many as you can afford early.

---

## Step 6: Watch your rates

Press `F1` (or just wait — you're already there). Look at the rate column:
- `food: +N/t` — should be positive (at least +2)
- `wood: +N/t` — should be positive (at least +1)
- `knowledge: +N/t` — should be positive (at least +1)

If food is negative, add more food workers. If knowledge is zero, assign more knowledge workers or build more altars.

---

## Step 7: Research your first tech

Once you hit 2,500 knowledge (watch the knowledge bar):

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
- Food: 8,000
- Wood: 5,200
- Knowledge: 1,400
- Huts: 20
- Altars: 10

Keep building huts and altars. Keep assigning workers. Keep knowledge flowing. The age advances **automatically** when all bars fill.

---

## Checklist at the Stone Age transition

Before you advance, you should have:
- [ ] 15–20 huts
- [ ] 8–10 altars
- [ ] 3–4 stashes
- [ ] 10+ pop, all assigned
- [ ] Tool Making researched
- [ ] Food rate positive by at least +3/t

---

## What the Stone Age unlocks

When you reach Stone Age:
- **Stone** resource unlocks (needed for Bronze Age)
- **Gathering Camp** — boosts food and wood
- **Stone Pit** — produces stone passively
- **Woodcutter Camp** — dedicated wood building
- **Firepit** — better knowledge building
- **Great Monolith** — first major wonder

Your first priority: build **Stone Pits** and assign lumber workers to wood production, because Bronze Age needs a lot of both.

---

## Quick command reference for the first age

| What to do | Command |
|---|---|
| Build a hut | `build hut` |
| Build a stash | `build stash` |
| Build an altar | `build altar` |
| Recruit a food worker | `recruit food` |
| Recruit a knowledge worker | `recruit knowledge` |
| Assign food workers to gathering camp | `assign gathering_camp 3` |
| Assign knowledge workers to story circle | `assign story_circle 1` |
| Start first research | `research tool_making` |
| Check economy tab | `e` or `F1` |
| Check logs | `l` or `F7` |
| Save the game | `Esc` |

---

> **Tip:** The game is idle — you don't need to babysit it. Set up your assignments, queue a few buildings, start a research, then step away for a few minutes and come back to see the progress.
