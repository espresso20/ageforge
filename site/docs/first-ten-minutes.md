# Your First Age

A walkthrough of the Primitive Age, your first age. It goes further than the Getting started guide that the game shows at the end of the Buildings list during your first 30 minutes of play in the Primitive and Stone Ages.

---

## Minute 0: You just launched AgeForge

On a first visit the game asks you to name your account, then opens its [main menu](how-to-play.md#the-main-menu) with New game first. Press `Enter` (or `n`), name your civilization, and the game starts.

You are in the Primitive Age. The screen shows:
- The status ribbon: your account name, `Primitive Age`, the year, `People 0/0`, `Morale 50%` and the next age
- The Ledger under it: a long bar to the Stone Age with what it is waiting on, and the four pillars
- The Resources box, the Under construction box (with the People box beside it on a big terminal) and the Log on the left
- The Buildings list in the middle, with the Getting started guide at the end of it while it shows
- The Panels list on the right, which names each panel (and, on a big terminal, what it is for)

Type `map` any time to open the [Map](map.md) (Esc closes it): it draws your town from your real buildings, so it grows as you build. Type `icons` if you want real icons on it. `minimap on` puts a short mini map above the Buildings list, on a terminal of 105 columns by 34 rows or larger.

<figure class="screen" data-screen="new-game"><figcaption>The screen a new game opens on, at 120 columns by 40 rows: almost nothing in stock, nothing built, and the first buildings waiting in the Buildings list.</figcaption></figure>

You have 25 food, 50 wood, no workers and no housing. Workers come on their own once you have housing and buildings for them to work in. Start by gathering:

```
gather wood 25
gather food 25
```

Each `gather` collects up to 25 of a resource (3 if you leave the amount out). `↑` recalls your last command, and `Tab` or `→` finishes a command you have started typing.

Hand-gathering works through the Medieval Age and is disabled from the Renaissance Age onward. By then your buildings and workers carry the economy.

---

## Step 1: Build a gathering camp and a wood camp

Food runs short before anything else, so get production going first. Both camps cost 22 wood, and you start with 50 (75 after the gathering above).

```
build gathering_camp
build wood_camp
```

A gathering camp makes +1.0 food a tick and a wood camp about +0.57 wood a tick when fully staffed (3 workers each). An unstaffed building still runs at 20%, so a staffed camp makes 5x what an empty one does. Workers come to staff them once you have housing (next step). The Stone Age asks for eight of each camp, so keep adding them as wood comes in.

---

## Step 2: Build huts

A hut costs 20 wood, takes 8 ticks (16 seconds) and adds 10 housing. More housing means more workers, and they come on their own. The Stone Age wants 10 huts, so keep adding them as wood comes in.

```
build hut
```

Builds run at the same time, not one after another, so start each one as soon as you can afford it. Each further copy of a building costs more than the last (a hut 13% more, a camp 15% more). The Under construction box shows what is being built and how far along it is.

---

## Step 3: Watch your workers arrive

Once a hut stands and your camps are built, workers arrive on their own. Your population grows toward your housing while food allows, and new workers fill your empty camp slots at once: the game places them every ten seconds. Workers past the slots wait in the reserve. The first time workers arrive, the log explains it.

```
people
```

`people` opens the People panel (`workers` does too). Its Overview tab shows your population over your homes, every kind of person, food use and what holds your population back. Its Work tab shows what auto-recruit is doing (growing, held by housing, held by food) and how your workers are spread across the work areas. Everyone eats food each tick (0.06 in the Primitive Age), and a new person is recruited only if your food income still covers what they eat, with a small margin to spare. More huts make room for more workers, and more buildings give them more slots to fill.

You do not pick a work area for a worker. It takes the area of the building it works in: a worker at a gathering camp works in Food, one at a wood camp in Lumber, one at a shrine in Faith. Type `status` for population, the reserve and food drain.

---

## Step 4: Steer them (optional)

By default every work area is on auto: your workers spread across your buildings in proportion to their worker slots, so they follow what you build. With every slot filled, that is a full crew everywhere: 3 workers on each gathering camp and wood camp, 2 on each shrine.

To put more of your workers on one kind of work, give it a share of the workforce:

```
roster faith 40
```

Now 40% of your workers go to faith, as far as your shrines have slots, and the other areas split the rest by their slots. `roster faith auto` undoes it, and `roster auto` puts every area back on auto. Do not set food to 0: food is what lets the game recruit.

You can also take over by hand. `assign` and `unassign` work as always, and the game does not take a worker out of a building afterward, except out of an area you set to 0. After any worker command it waits a minute before placing anyone, so it will not grab the workers you are moving. If you would rather recruit yourself, `workers auto-recruit off` stops automatic recruiting. Idle workers still go to work by your shares.

```
assign shrine 1
workers auto-recruit off
```

The People box shows your reserve: workers without a slot. They take any slot that opens within seconds, except in the minute after a worker command of yours. See [The Roster](workers-and-domains.md#the-roster) for the details.

---

## Step 5: Build stashes

A stash costs 50 wood and adds 1% to all production. There is no limit on copies, but each one costs 1.75 times the one before it.

```
build stash
```

Stock has no limit, so a stash holds nothing. It speeds the whole town up a little, and so do the other storage buildings you will meet. The Stone Age asks for five stashes.

---

## Step 6: Build story circles and a shrine

Story circles come with 2 scholars each. Scholars are not workers: they need no assignment, and they work on your research as soon as the circle stands. They are people, though: they live in your homes and eat like anyone. The Stone Age needs 5 story circles, and a circle past the fifth researches at half rate. A story circle costs 85 wood for the first, each copy costs 1.5 times the one before it, and it takes 75 ticks (2m 30s) to build, so start them early.

```
build story_circle
```

Also build shrines (85 wood for the first). The Stone Age asks for 8 shrines. Shrines produce faith and lift morale a little. Faith matters later: it lowers the odds that a catastrophe strikes and pays to appease a harbinger (see [Faith](faith.md)).

```
build shrine
```

---

## Step 7: Watch your rates

The Resources box shows a rate beside each stock. Food and wood should both be positive.
- Each staffed gathering camp adds 1.0 food a tick, and each person eats 0.06. If food is negative, build another gathering camp; workers come to staff it.
- Each staffed wood camp adds about 0.57 wood a tick.
- Each story circle adds 2 scholars, so more circles (up to five at full rate) finish research sooner.

---

## Step 8: Research your first tech

Once you have a story circle, its scholars can research:

```
research tool_making
```

It costs nothing to start. A bar fills as your scholars work, and the screens show its percent and the time left. With five story circles it takes about 14 minutes, and each circle you add shortens it. It gives +10% food and +10% wood for the rest of the run, and every `gather` brings 2 more. Get it here if you can: Stoneworking, the tech the Stone Age's wonder needs, stands on it. If you would rather advance first, queue it with `plan research tool_making` and the [build plan](plan.md) starts it when the research slot is free. Language (+10% research speed) and Fire Mastery (+10% food and +5% housing) are the age's other two techs. `research list` shows the time each would take your scholars now.

---

## Step 9: Build the Sacred Grove (required to advance)

Every age has a wonder, and you cannot advance until it stands. The Sacred Grove costs:
- 1.6K wood
- 800 food

Bank the resources, then build it (150 ticks, 5 minutes):

```
wonder collect all
build sacred_grove
```

`wonder collect all` banks as much of each needed resource as you have. Run it again whenever your stock refills. Nothing flows into the Grove's bank on its own.

The grove also speeds research and gives +0.05 food a tick permanently.

---

## Step 10: Check the long bar and advance

The long bar under the status ribbon runs to the next age, and the Waiting on line beside it (under it on a big terminal) names what is still missing, with your count beside each one. Type `charter` for the whole list. For the Stone Age the gate asks for:
- Gathering Camps: 8
- Huts: 10
- Shrines: 8
- Stashes: 5
- Story Circles: 5
- Wood Camps: 8
- The Sacred Grove built

No age asks you to hold an amount of a resource, and advancing spends none. See [Advancement Requirements](ages.md#advancement-requirements) for the rule, and [The Advance Panel](advance.md) for the screen. Keep building huts (new huts bring new workers) and keep your scholars researching. When the bar is full and the line says every pillar stands, type `advance`. The age changes only when you advance: with `advance`, with `Enter` in the Advance panel, or through a `plan advance` item. If you are stepping away, `plan advance` queues it, and the plan advances once everything is ready.

A `⚑ Harbinger` badge may appear on the status ribbon. Type `harbinger` to read it, or ignore it: it never blocks anything here, and no catastrophe can strike before the Iron Era. See [The Harbinger](harbinger.md).

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
- [ ] A positive food rate, so your workers keep arriving
- [ ] Tool Making researched (optional)

---

## What the Stone Age unlocks

When you reach the Stone Age you get:
- Stone: the age's resource
- Stone Camp: your first stone producer, built from wood alone
- Stone Pit: far more stone, but it costs stone to build
- Woodcutter Camp: a much better wood producer
- Forager Post: a better food producer
- Elders' Hall: a better knowledge building, with 2 scholars
- Standing Stones: a better faith building, once you research Ritual (it needs Language, a Primitive Age tech)
- Longhouse: bigger housing (+25 housing each)
- Storage Pit: +1% to all production a copy
- War Camp: an early military building that comes with 9 soldiers. They move in as food allows, and a garrison blunts raids (see [Army & Missions](military.md))
- Great Monolith: the Stone Age wonder, required to reach the Bronze Age. It needs Stoneworking researched before it can be built

Your older buildings do not change on their own. Type `upgrade` to see which ones can become their Stone Age version (gathering camps into forager posts, for example) at half the price of a new copy. An upgraded copy counts toward the next age's gate. See [Building Upgrades](buildings.md#building-upgrades).

Your first priority: start stone with Stone Camps (or `gather stone 25`), then build Stone Pits, Woodcutter Camps and Longhouses so workers have room to come and staff them. The Bronze Age asks for 8 each of Forager Posts, Stone Camps, Stone Pits, War Camps and Woodcutter Camps, 5 Elders' Halls, 5 Storage Pits, 15 Longhouses and the Great Monolith. See [How to Play](how-to-play.md) for the long game.

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
| See your people, the roster and what auto-recruit is doing | `people` |
| Put more workers on faith | `roster faith 40` |
| Put every area back on auto | `roster auto` |
| Place a worker by hand | `assign shrine 1` |
| Take a worker out of a building | `unassign gathering_camp 1` |
| Recruit by hand instead | `workers auto-recruit off`, then `recruit [count]` or `recruit max` |
| Start first research | `research tool_making` |
| Bank resources for the wonder | `wonder collect all` or `wonder collect <resource> [amount]` |
| Build the wonder | `build sacred_grove` |
| See what the next age needs | `charter` |
| Advance to the next age | `advance` |
| Check population, reserve and food drain | `status` |
| Check logs | `logs` |
| Open the map | `map` |
| Queue builds for while you are away | `plan build hut 10`, `plan advance` |
| Save and return to the main menu | `Esc` (with no panel open) |
| Come back to this game from the main menu | `Enter` on Continue, or `c` |

---

The game is idle, so you do not need to watch it. It recruits workers and puts them to work by your [roster](workers-and-domains.md#the-roster), while you are away too. Leave a [build plan](plan.md) before you step away: `plan build hut 10`, `plan research fire_mastery`, `plan advance`. The plan starts each item as the resources come in, while you play and while you are away, and pays for it only when it starts.
