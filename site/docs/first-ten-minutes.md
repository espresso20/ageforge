# Your First Age

A guided walkthrough of the Primitive Age, your first age. The steps follow the **Getting started** guide the game shows under the Buildings list.

---

## Minute 0: You just launched AgeForge

The game opens on its [main menu](how-to-play.md#the-main-menu). On a first visit that is the forge and a title page, with **New game** first: press `Enter` (or `n`), name your civilization, and the game starts.

You're in the **Primitive Age**. The screen shows:
- The status bar: your account name, `Primitive Age`, `Pop: 0/0` and `Morale 50%`
- The Next Age row: what the Stone Age asks for, each item marked ✓ or ✗
- The Economy panel, which is always on screen: Resources, Under construction and the Log on the left, the Buildings list on the right, with the Getting started guide under it

Type `map` any time to open the [Map](map.md) (Esc closes it): it draws your town from your real buildings, so it grows as you build. Type `icons` if you want real icons on it. `minimap on` puts a short mini map above the Buildings list, on a terminal of 105 columns by 34 rows or larger.

<figure class="screen" data-screen="new-game"><figcaption>The screen a new game opens on, at 120 columns by 40 rows: stores nearly empty, nothing built, and the first buildings waiting in the Buildings list.</figcaption></figure>

You have no workers, no housing and 50 wood. Workers come on their own once you have housing and buildings for them to work in. Start by gathering:

```
gather wood 25
gather food 25
```

Each `gather` collects up to **25** of a resource (3 if you leave the amount out). `↑` recalls your last command, and `Tab` or `→` finishes a command you've started typing.

Hand-gathering is an early-game crutch. It works through the **Medieval Age** and is disabled from the **Renaissance Age** onward. By then your buildings and workers should carry the economy.

---

## Step 1: Build a gathering camp and a wood camp

Food runs short before anything else, so get production going first. Both camps cost **24 wood**, and you start with 50.

```
build gathering_camp
build wood_camp
```

A gathering camp makes **+1.0 food/tick** and a wood camp about **+0.57 wood/tick** when fully staffed (3 workers each). An unstaffed building still runs at 20%, so a staffed camp makes **5x** what an empty one does. Workers come to staff them once you have housing (next step). The Stone Age asks for **eight of each camp**, so keep adding them as wood comes in.

---

## Step 2: Build huts

A hut costs **21 wood**, takes 8 ticks (16 seconds) and adds **10 housing**. More housing means more workers, and they come on their own. The Stone Age wants 10 huts, so keep adding them as wood comes in.

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

`workers` opens the Workers panel: your population, idle count and food use, and the **Roster** section, which says what auto-recruit is doing (recruiting, no housing left, every worker slot filled, waiting for food) and how your workers are spread across the domains. Every worker eats food each tick, so the game recruits only while your food rate stays positive with a margin to spare. More huts make room for more workers, and more buildings give them more slots to fill.

You don't pick a domain for a worker. It takes the class of the building it works in: a worker on a `gathering_camp` is a Forager, one on a `story_circle` a Shaman. Type `status` for population, idle count and food drain.

---

## Step 4: Steer them (optional)

By default every domain is on auto: your workers spread across your buildings in proportion to their worker slots, so they follow what you build. With every slot filled, that is a full crew everywhere: 3 workers on each gathering camp and wood camp, 2 on each shrine.

To put more of your workers on one kind of work, give it a share of the workforce:

```
roster faith 40
```

Now 40% of your workers go to faith, as far as your shrines have slots, and the other domains split the rest by their slots. `roster faith auto` undoes it, and `roster auto` puts every domain back on auto. Don't set food to 0: food is what lets the game recruit.

You can also take over by hand. `assign` and `unassign` work as always, and the game never moves a worker you placed. After any worker command it waits a minute before placing anyone, so it won't grab the workers you are moving. If you'd rather recruit yourself, `workers auto-recruit off` stops automatic recruiting; idle workers still go to work by your shares.

```
assign story_circle 1
workers auto-recruit off
```

The Workers box in the sidebar should show **Idle: 0**. The game recruits only into empty slots and puts idle workers to work within seconds; after a worker command of yours, it waits a minute first. See [The Roster](workers-and-domains.md#the-roster) for the details.

---

## Step 5: Build stashes

Food starts with only 50 storage and wood with 75. Storage fills fast, and anything past it is wasted. A stash costs **52 wood** and gives **+150 storage** for every resource. There is no limit on copies, but each one costs 1.75 times the one before it.

```
build stash
```

Build two early so there is room to hold what comes in. The Stone Age asks for **five stashes**.

---

## Step 6: Build story circles and a shrine

Story circles come with **2 scholars** each. Scholars are not workers: they need no housing, no food and no assignment, and they start research the moment the circle stands. The Stone Age needs **5 story circles**, and a circle cannot be built past 5. A story circle costs 90 wood and takes 75 ticks (2m 30s) to build, so start them early.

```
build story_circle
```

Also build **shrines** (90 wood for the first). The Stone Age asks for **8 shrines**. Shrines produce faith and lift morale a little. Faith matters later: it lowers the odds that a catastrophe strikes and pays to appease a harbinger (see [Faith](faith.md)).

```
build shrine
```

---

## Step 7: Watch your rates

The Economy panel is always on screen. Look at the rate column:
- `food: +N/t` should be positive (each staffed gathering camp adds +1.0; aim for at least +2)
- `wood: +N/t` should be positive (each staffed wood camp adds about +0.57)

If food is negative, build another gathering camp; workers come to staff it. If research is slow to finish, build more story circles: each adds 2 scholars.

---

## Step 8: Research your first tech

Once you have a story circle, its scholars can research:

```
research tool_making
```

It costs nothing to start. A bar fills as your scholars work, and the screens show its percent and the time left. With five story circles it takes about 13 minutes; each circle you add shortens it. It gives +10% food and +10% wood for the rest of the run, and every `gather` brings 2 more. Get it here if you can: Stoneworking, the tech the Stone Age's wonder needs, stands on it. If you'd rather advance first, queue it with `plan research tool_making` and the [build plan](plan.md) starts it when the research slot is free. Language (+10% research speed) and Fire Mastery (+10% food and +5% housing) are the age's other two techs. `research list` shows the time each would take your scholars now.

---

## Step 9: Build the Sacred Grove (required to advance)

Every age has a wonder, and you can't advance until it stands. The Sacred Grove costs:
- 1.6K wood
- 800 food

Bank the resources, then build it (75 ticks, 2m 30s):

```
wonder collect all
build sacred_grove
```

`wonder collect all` banks as much of each needed resource as you have. Run it again whenever your stock refills; your storage limits how much you can bank at once. **Wonder overflow** helps too: it is on by default, and production your full stores would waste goes into the Grove's bank on its own.

The grove also speeds research and gives +0.05 food/tick permanently.

---

## Step 10: Check the Next Age row and advance

The Next Age row under the status bar always shows what you need for the **next age**. For the Stone Age:
- Gathering Camps: 8
- Huts: 10
- Shrines: 8
- Stashes: 5
- Story Circles: 5
- Wood Camps: 8
- The Sacred Grove built

No age asks you to hold an amount of a resource, and advancing spends none. See [Advancement Requirements](ages.md#advancement-requirements) for the rule. Keep building huts (new huts bring new workers) and keep your scholars researching. When every item has its ✓, type `advance`. The age never advances on its own. If you're stepping away, `plan advance` queues it and the plan advances for you once everything is ready.

> **A harbinger?** A `⚑ Harbinger` badge may appear on the status bar. Type `harbinger` to read it, or ignore it: it never blocks anything, and no catastrophe can strike before the Iron Era. See [The Harbinger](harbinger.md).

---

## Checklist at the Stone Age transition

Before you advance, you should have:
- [ ] 8 gathering camps
- [ ] 8 wood camps
- [ ] 10 huts
- [ ] 8 shrines
- [ ] 5 stashes
- [ ] 5 story circles
- [ ] The Sacred Grove built
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
- **Elders' Hall**: a better knowledge building, with 2 scholars
- **Standing Stones**: a better faith building, once you research Ritual (it needs Language, a Primitive Age tech)
- **Longhouse**: bigger housing (+25 housing each)
- **Storage Pit**: +1.2K storage for every resource
- **War Camp**: an early military building. It trains soldiers from the day it stands, and a garrison blunts the raids of the Stone and Bronze Ages (see [Military](military.md))
- **Great Monolith**: the Stone Age wonder, required to reach the Bronze Age

Your older buildings don't change on their own. Type `upgrade` to see which ones can become their Stone Age version (gathering camps into forager posts, for example); see [Building Upgrades](buildings.md#building-upgrades).

Your first priority: start stone with **Stone Camps** (or `gather stone 25`), then build **Stone Pits**, **Woodcutter Camps** and **Longhouses** so workers have room to come and staff them. The Bronze Age asks for 8 each of Elders' Halls, Forager Posts, Stone Camps, Stone Pits, War Camps and Woodcutter Camps, 5 Storage Pits, 15 Longhouses and the Great Monolith. See [How to Play](how-to-play.md) for the long game.

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
| See your workers, the roster and what auto-recruit is doing | `workers` |
| Put more workers on faith | `roster faith 40` |
| Put every domain back on auto | `roster auto` |
| Place a worker by hand | `assign shrine 1` |
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
| Come back to this game from the main menu | `Enter` on **Continue**, or `c` |

---

> **Tip:** The game is idle, so you don't need to babysit it. It recruits workers and puts them to work by your [roster](workers-and-domains.md#the-roster), while you are away too. Leave a [build plan](plan.md) before you step away: `plan build hut 10`, `plan research fire_mastery`, `plan advance`. The plan starts each item as the resources come in, while you play and while you are away, and pays for it only when it starts.
