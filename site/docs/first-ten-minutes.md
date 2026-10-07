# Your First Age

A guided walkthrough of the Primitive Age, your first age. It is tuned to take about **15 minutes** of real time; follow this and you'll reach the Stone Age in about that time. The steps follow the **Getting started** guide the game shows under the Buildings list.

---

## Minute 0: You just launched AgeForge

You're in the **Primitive Age**. The screen shows:
- The status bar: your account name, `Primitive Age`, the epoch (`◈ Stone Era`), `Pop: 0/0` and `Morale 50%`
- The Next Age row: what the Stone Age asks for, each item marked ✓ or ✗
- The Economy panel, which is always on screen: Resources, Under construction and the Log on the left, the Buildings list on the right, with the Getting started guide under it
- On a large terminal (about 120x40 or bigger), a **mini map** above the Buildings list

The mini map draws your town from your real buildings, so it grows as you build. On smaller terminals it hides to leave the Buildings list room. Type `map` any time to open the full [Map](map.md) (Esc closes it), and `icons` if you want real icons on it.

You have no workers, no housing and 50 wood. Workers come on their own once you have housing and buildings for them to work in. Start by gathering:

```
gather wood 25
gather food 25
```

Each `gather` collects up to **25** of a resource (3 if you leave the amount out). `↑` recalls your last command, and `Tab` or `→` finishes a command you've started typing.

Hand-gathering is an early-game crutch. It works through the **Medieval Age** and is disabled from the **Renaissance Age** onward. By then your buildings and workers should carry the economy.

---

## Step 1: Build a gathering camp and a wood camp

Food runs short before anything else, so get production going first. Both camps cost **16 wood**, and you start with 50.

```
build gathering_camp
build wood_camp
```

A gathering camp makes **+1.0 food/tick** and a wood camp about **+0.57 wood/tick** when fully staffed (3 workers each). An unstaffed building still runs at 20%, so a staffed camp makes **5x** what an empty one does. Workers come to staff them once you have housing (next step).

---

## Step 2: Build huts

A hut costs **14 wood**, takes 8 ticks (16 seconds) and adds **10 housing**. More housing means more workers, and they come on their own. The Stone Age wants 10 huts, so keep adding them as wood comes in.

```
build hut
```

Builds run at the same time, not one after another, so start each one as soon as you can afford it. Each further copy of a building costs a little more than the last. The Under construction box shows what is being built and how far along it is.

---

## Step 3: Watch your workers arrive

Once a hut stands and your camps are built, workers arrive on their own. The game recruits into empty worker slots while housing and food allow, and puts each new worker to work: within seconds your camps are staffed. The first time, the log says so.

```
workers
```

`workers` opens the Workers panel: your population, idle count and food use, and the **Shares** section, which says what auto-recruit is doing (recruiting, no housing left, every worker slot filled, waiting for food) and how your workers are spread across the domains. Every worker eats food each tick, so the game recruits only while your food rate stays positive with a margin to spare. More huts make room for more workers, and more buildings give them more slots to fill.

You don't pick a domain for a worker. It takes the class of the building it works in: a worker on a `gathering_camp` is a Forager, one on a `story_circle` a Shaman. Type `status` for population, idle count and food drain.

---

## Step 4: Steer them (optional)

By default every domain is on auto: your workers spread across your buildings in proportion to their worker slots, so they follow what you build. With every slot filled, that is a full crew everywhere: 3 workers on each gathering camp and wood camp, 2 on each story circle and shrine.

To put more of your workers on one kind of work, give it a share of the workforce:

```
workers share knowledge 40
```

Now 40% of your workers go to knowledge, as far as your story circles have slots, and the other domains split the rest by their slots. `workers share knowledge auto` undoes it, and `workers share auto` puts every domain back on auto. Don't set food to 0: food is what lets the game recruit.

You can also take over by hand. `assign` and `unassign` work as always, and the game never moves a worker you placed. After any worker command it waits a minute before placing anyone, so it won't grab the workers you are moving. If you'd rather recruit yourself, `workers auto-recruit off` stops automatic recruiting; idle workers still go to work by your shares.

```
assign story_circle 1
workers auto-recruit off
```

The Workers box in the sidebar should show **Idle: 0**. The game recruits only into empty slots and puts idle workers to work within seconds; after a worker command of yours, it waits a minute first. See [Worker Shares](workers-and-domains.md#worker-shares) for the details.

---

## Step 5: Build stashes

Food and wood start with only 50 storage each, and knowledge with 30. Storage fills fast, and anything past it is wasted. A stash costs **35 wood** and gives **+500 storage** for every resource, and you can build up to 50.

```
build stash
```

Build two early. The Stone Age asks for 1K food and 1K wood and the Sacred Grove for 1K wood, and you need the room to hold them. A stash also lifts your knowledge storage past the 61 knowledge Tool Making costs.

---

## Step 6: Build story circles and a shrine

Story circles produce knowledge: **+0.2 knowledge/tick** each when staffed (2 workers). The Stone Age needs **5 story circles**, and knowledge is what you spend on research. A story circle costs 60 wood and takes 75 ticks (2m 30s) to build, so start them early.

```
build story_circle
```

Also build your first **shrine** (60 wood). Shrines produce faith and lift morale a little. Faith matters later: it lowers the odds that a catastrophe strikes and pays to appease a harbinger (see [Faith](faith.md)).

```
build shrine
```

---

## Step 7: Watch your rates

The Economy panel is always on screen. Look at the rate column:
- `food: +N/t` should be positive (each staffed gathering camp adds +1.0; aim for at least +2)
- `wood: +N/t` should be positive (each staffed wood camp adds about +0.57)
- `knowledge: +N/t` should be positive (five staffed story circles make +1.0)

If food is negative, build another gathering camp; workers come to staff it. If knowledge is low, build more story circles, or give knowledge a bigger share with `workers share knowledge 40`.

---

## Step 8: Research your first tech

Once you have 61 knowledge (knowledge storage starts at 30, so you need a stash to hold that much):

```
research tool_making
```

It costs the 61 knowledge up front, takes 23 ticks (46 seconds) and gives a permanent +15% to what your workers produce. Get it here if you can: Stoneworking, the tech the Stone Age's wonder needs, stands on it. If you're short, advance first and research it later, or queue it with `plan research tool_making` and the [build plan](plan.md) starts it once the knowledge is there. Fire Mastery (102 knowledge, +0.1 food/tick) is the age's other tech.

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

`wonder collect all` banks as much of each needed resource as you have. Run it again whenever your stock refills; your storage limits how much you can bank at once. **Wonder overflow** helps too: it is on by default, and production your full stores would waste goes into the Grove's bank on its own.

The grove also gives +0.02 knowledge/tick and +0.05 food/tick permanently.

---

## Step 10: Check the Next Age row and advance

The Next Age row under the status bar always shows what you need for the **next age**. For the Stone Age:
- Food: 1K
- Wood: 1K
- Knowledge: 150
- Huts: 10
- Story Circles: 5
- The Sacred Grove built

Keep building huts and story circles (new huts bring new workers) and keep knowledge flowing. When every item has its ✓, type `advance`. The age never advances on its own. If you're stepping away, `plan advance` queues it and the plan advances for you once everything is ready.

> **A harbinger?** A `⚑ Harbinger` badge may appear on the status bar. Type `harbinger` to read it, or ignore it: it never blocks anything, and no catastrophe can strike before the Iron Era. See [The Harbinger](harbinger.md).

---

## Checklist at the Stone Age transition

Before you advance, you should have:
- [ ] 10 huts
- [ ] 5 story circles
- [ ] The Sacred Grove built
- [ ] 2-3 stashes
- [ ] Staffed gathering camps and wood camps
- [ ] 10+ workers, none idle
- [ ] Food rate positive by at least +3/tick
- [ ] Tool Making researched (optional)

---

## What the Stone Age unlocks

When you reach the Stone Age:
- **Stone**: a new resource, needed for the Bronze Age
- **Stone Camp**: your first stone producer, built from wood alone
- **Stone Pit**: far more stone, but it costs stone to build
- **Woodcutter Camp**: a much better wood producer
- **Forager Post**: a better food producer
- **Elders' Hall**: a better knowledge building
- **Standing Stones**: a better faith building
- **Longhouse**: bigger housing (+25 housing each)
- **Storage Pit**: +2.2K storage for every resource (up to 25)
- **War Camp**: an early military building. Soldiers unlock in the Iron Age, so it can wait
- **Great Monolith**: the Stone Age wonder, required to reach the Bronze Age

Your older buildings don't change on their own. Type `upgrade` to see which ones can become their Stone Age version (gathering camps into forager posts, for example); see [Building Upgrades](buildings.md#building-upgrades).

Your first priority: start stone with **Stone Camps** (or `gather stone 25`), then build **Stone Pits**, **Woodcutter Camps** and **Longhouses** so workers have room to come and staff them. The Stone Age is tuned to take about **45 minutes**, and the Bronze Age needs 4K food, 8K wood, 4K stone, 1.5K knowledge, 15 longhouses, 5 stone pits, 5 elders' halls and the Great Monolith. From the Bronze Age on, each age takes hours: see [How to Play](how-to-play.md) for the long game.

---

## Quick command reference for the first age

| What to do | Command |
|---|---|
| Gather by hand | `gather wood 25`, `gather food 25` |
| Build a gathering camp | `build gathering_camp` |
| Build a wood camp | `build wood_camp` |
| Build a hut | `build hut` |
| Build a stash | `build stash` |
| Build a story circle | `build story_circle` |
| Build a shrine | `build shrine` |
| See your workers, their shares and what auto-recruit is doing | `workers` |
| Put more workers on knowledge | `workers share knowledge 40` |
| Put every domain back on auto | `workers share auto` |
| Place a worker by hand | `assign story_circle 1` |
| Take a worker out of a building | `unassign gathering_camp 1` |
| Recruit by hand instead | `workers auto-recruit off`, then `recruit [count]` or `recruit max` |
| Start first research | `research tool_making` |
| Bank resources for the wonder | `wonder collect all` or `wonder collect <resource> [amount]` |
| Build the wonder | `build sacred_grove` |
| Advance to the next age | `advance` |
| Check population, idle and food drain | `status` |
| Check logs | `logs` |
| Open the map | `map` |
| Queue builds for while you're away | `plan build hut 10`, `plan advance` |
| Save and return to the main menu | `Esc` (with no panel open) |

---

> **Tip:** The game is idle, so you don't need to babysit it. It recruits workers and puts them to work by your [worker shares](workers-and-domains.md#worker-shares), while you are away too. Leave a [build plan](plan.md) before you step away: `plan build hut 10`, `plan research fire_mastery`, `plan advance`. The plan starts each item as the resources come in, while you play and while you are away, and pays for it only when it starts.
