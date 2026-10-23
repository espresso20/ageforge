# How to Play

AgeForge is a **command-driven idle civilization game**. Resources accumulate on their own, buildings construct over time, and you steer by typing commands at the `>` prompt. A first run takes about a week of real time, so the game is built for checking in a few times a day.

---

## The main menu

The game opens on its main menu, and `Esc` brings you back to it from a game.

<figure class="screen" data-screen="menu"><figcaption>The main menu of an account with a game, at 120 columns by 40 rows: its own town behind the menu, Continue first, and a caption that names the town and its age.</figcaption></figure>

Once your account has a game, the menu is drawn over **your own town**: the map of your current game, in your map style, glyph set and theme, at the age that game is in, moving as the [Map](map.md) does. It is a look at the save and nothing more. The game is not loaded or run, no time away is counted, and nothing is written, so looking at the menu costs you nothing and earns you nothing. **Continue** is first and already selected: press `Enter` and you are back in that game. Under it are the town, its age and its people, and beside it the save's name.

<figure class="screen" data-screen="menu-first"><figcaption>The main menu on a first visit: the wordmark as iron on the fire, the menu as a table of contents, and the forge on its plate.</figcaption></figure>

An account with no game yet sees the forge and a title page, with **New game** first. The same page shows if your save cannot be read.

| Entry | Key | What it does |
|---|---|---|
| **Continue** | `c` | Opens your current game. It is on the menu once the account has a save |
| **New game** | `n` | Asks you to name a civilization, then starts it (see [Starting a new game](saving-and-loading.md#starting-a-new-game)) |
| **Load game** | `l` | Opens the [Load Game browser](saving-and-loading.md#the-load-game-browser), where you can also mark your main game |
| **Badges** | `b` | Shows how many badges you hold and the title you wear, and opens the [badge case](account.md#the-badge-case). No game needs to be loaded |
| **Themes** | `t` | Opens the [theme](themes.md) picker |
| **Accounts** | `a` | Opens the [Accounts panel](account.md#the-accounts-panel): switch, create, back up and wipe accounts, and delete an account's saves |
| **Check for updates** | `u` | Asks for the latest release (see [Updating](getting-started.md#updating)) |
| **Quit** | `q` | Leaves the game |

`↑` and `↓` move the selection and `Enter` opens it; each entry's letter opens it directly. Choosing an entry strikes the title: a flare and a burst of sparks, then the entry opens. The version is on the page, and so is **update available (u)** when a newer release is out.

**Which game Continue opens.** Your account remembers the game you played last, and that is your current game. If you keep several games and want one of them on the menu whatever you played last, mark it as your **main game** with `m` in the Load Game browser. See [Continue and your main game](saving-and-loading.md#continue-and-your-main-game).

**Motion and size.** `motion off` holds the menu to one still picture, like the maps (see [Motion](themes.md#motion)). The menu fits any terminal from 80 columns by 24 rows up. With `map glyphs ascii` it is drawn without block characters. On a light theme the fire and the hot iron are ink on paper: hotter is darker.

---

## The core loop

```
Build → Staff → Research → Advance → Repeat
```

1. **Build** structures to add housing and production
2. **Staff** them: workers move in on their own toward your housing while food allows, and go to work by your [roster](workers-and-domains.md#the-roster). Steer the split with `roster`, or `recruit` and `assign` by hand
3. **Research** technologies that multiply your output
4. **Advance** to the next age with `advance` once the buildings it asks for stand and the age's wonder is built (see [Advancement Requirements](ages.md#advancement-requirements))
5. **Prestige** for a lasting production bonus. After the first prestige, [Plan Template, Roster and Old Friends](prestige.md#carry-overs) carry into later runs. Prestige opens at the Medieval Age, but a run to the Modern Age (or later) pays far more. Every age the run completed then runs faster on the next run ([Era Mastery](prestige.md#era-mastery))

New to the game? [Your First Age](first-ten-minutes.md) walks you through the Primitive Age step by step, from the first `gather` to the Sacred Grove.

---

## The screen layout

```
┌──────────────────────────────────────────────────────────────────────────┐
│ Stone Age · Year 3,200 BC · People 18/30 · Morale 54% · Next: Bronze Age │  status ribbon
│ To the Bronze Age ████████░░░░░░  52%  Waiting on: Elders' Hall 3/5 · ...│  the Ledger: the long bar
│ KNOWLEDGE ████ 80%   PEOPLE ███ 60%   INDUSTRY ██ 40%   MONUMENT ░ 0%    │  and the four pillars
│ Food 2.1K +4/t ▲ rising    Wood 340 +12/t ▲ rising    ...                │  and a card for each resource
├─ Resources ───────────────────────────┬─ Buildings ───────┬─ Panels ─────┤
│ amounts and rates                     │ what you can      │ TOWN         │
│                                       │ build, and what   │ people ...   │
│                                       │ you own           │ PROGRESS     │
│                                       │ (minimap on: a    │ research ... │
│                                       │ map above it)     │ WORLD        │
├─ Under construction ───┬─ People ─────┤                   │ trade ...    │
│ progress bars          │ pop and bar  │                   │ YOU          │
│                        │ four kinds,  │                   │ badges ...   │
│                        │ held back by │                   │              │
├─ Log ──────────────────┴──────────────┤                   │              │
│ events and news                       ├─ Skyline ─────────┤              │
│                                       │ your skyline      │              │
├─ Command ─────────────────────────────┴───────────────────┴──────────────┤
│ > _                                                                      │
└──────────────────────────────────────────────────────────────────────────┘
```

That is the layout on a terminal of 120 columns by 36 rows or more. On a smaller one the Under construction box has its row to itself, the People box sits under the Panels list, and there is no Skyline strip (see below).

The top of the screen is the **status ribbon** and the **Ledger**, and they say the same things about your progress at every size.

The **status ribbon** is one line: your age, the calendar year, your people over your homes, morale and the next age. It also carries your prestige level and civilization title once you have them, and badges for a waiting harbinger or catastrophe. On a narrow terminal it gives things up in a fixed order: the hint at its end, then the next age, then the words of the badges, what morale does to production, your title, your account's name, the rest of the badges' words, and last the year. The age, the badges, your people and morale always stay.

The **Ledger** is under it:

- **The long bar** runs to the next age, with its percent. Beside it, or under it on a big terminal, a **Waiting on** line names what is still missing: each building the gate asks for with your count beside it, the keystone tech and the wonder's bank, and then the wonder itself. When the list is longer than the room, the line counts the rest (`+4 more`). When every pillar stands it says so and tells you to type `advance`. The bar's percent and the year are the same ones the [Advance panel](advance.md) shows, and `charter` opens it.
- **The four pillars** (Knowledge, People, Industry and Monument) are small meters, each with the same percent as in the Advance panel, or a check once it stands. On a big terminal each has a line under its name saying how many of its requirements are met.
- **The resource cards** show each resource that has something going on: its name, the amount, the rate a tick, a trend line of the last few minutes of the amount, and whether it is rising, falling or steady. A resource with nothing held, no rate and no part in the gate has no card. The cards stay where they are as the numbers change.
- **The bottleneck** is marked. A resource a waiting building or the wonder's bank is short of has its name lit, the one that will take longest has an arrow before it, and the **Short of** line names them. A resource the gate is short of always has a card, even when you hold none of it.

The Ledger comes in three sizes. Under 100 columns by 30 rows it is two rows: the long bar with what it waits on, and one row of cards (a few of them, with the ones the gate is short of first and a count of the rest). From 100 columns by 30 rows it adds the four pillars and a card row with trend lines. On the big layout (120 columns by 36 rows) the bar and the waiting line have a row each, the pillars have two, and the cards fill one to three rows by the height of the terminal and how many resources your age has; those rows change when you advance an age and at no other time. On the plain glyph tier (`map glyphs ascii`) the bars and trend lines are drawn in plain characters.

Panels you open (`research`, `trade`, `map` and so on) open over the dashboard, and the prompt keeps working while they are open.

<figure class="screen" data-screen="dashboard"><figcaption>The dashboard in the Bronze Age, on a terminal of 120 columns by 40 rows: the status ribbon and the Ledger across the top; then resources, construction and the People box side by side, and the log on the left; the Buildings list in the middle; the Panels list on the right.</figcaption></figure>

The dashboard fits any terminal from 80 columns by 24 rows up, and no row on it wraps. In the Resources box each resource has one row: its amount, its rate and a marker when it is falling (`▼`); culture and faith add a bar toward their next step. On a narrow terminal the rows give up the bar first, and at 80 columns `/tick` is written `/t`. On a short terminal the Panels list is set in two columns, and when the Resources box cannot show every resource it shows a page of them and says so: **Ctrl+R** turns the page. The mini map is off until you type `minimap on`; it then shows from 105 columns by 36 rows.

**The People box.** From 120 columns by 36 rows the People box is the right half of the Under construction row, and the row is as tall as the box. It shows the headcount over your homes (with how full they are when there is room to say so), one bar across the homes in the colors of the four kinds of person (workers, the reserve, soldiers and scholars), a line for each kind in the age's own word, and what holds your population back. The soldiers' line shows the soldiers at their posts over all the posts they have while some wait on food, and how many are out on a mission. `Held by: food` carries the net food a tick and how many more it would feed once the terminal is 42 rows tall. When nothing holds the population back the box says so and counts the empty homes. On a smaller terminal the box is the short one under the Panels list, with the same lines in less room. The [People panel](workers-and-domains.md#the-people-of-your-town) has the whole picture.

**The Panels list.** On the same terminals the Panels column is wider, and the panels are set in four groups, each with a few words on what it is for: Town (`people`, `plan`, `wonders`, `map`), Progress (`milestones`, `research`, `charter`, `epoch`, `history`), World (`expedition`, `army`, `trade`, `factions`, `harbinger`) and You (`badges`, `stats`, `logs`, `help`). The panel that is open is marked. On a smaller terminal, or when the column is too short for the groups, the list is the panel names alone, and it never drops one.

**The skyline strip.** Under the Buildings list a strip titled Skyline draws your town's skyline, whatever your map style is. It is a picture: it takes no keys and has no cursor, it moves with the Map when the motion setting is on and holds still when it is off, and `map` still opens the full Map. It is as wide as the Buildings list and half as tall as the Log box beside it, and it needs 4 rows inside its border, so it shows when the Log box is 12 rows or more: a terminal of 120 columns needs about 56 rows (more once your age has many resources), and one under 120 columns about 43. A terminal under 100 columns by 30 rows never shows it. Where it hides, the Buildings list has the whole column.

---

## Navigation

- Type a panel's name to open it: `research`, `army`, `trade`, `factions`, `map`, `help` and so on.
- `help` opens the Help panel: a full command reference plus the list of every panel you can open.
- **Tab** or **→** takes the completion shown in dim text after the cursor; press **Tab** again for the next one.
- **PgUp / PgDn** turn the Buildings list a page. The list never shows part of an entry: an entry that does not fit whole waits for the next page, and the box says how many are above and below.
- **Ctrl+R** turns the Resources box to its next page, when it has more resources than it can show.
- **↑ / ↓** step through your command history.
- **Esc** closes the open panel. With no panel open, it saves the game, stops it and returns to the [main menu](#the-main-menu), where **Continue** takes you back in.
- Common commands have short names: `b` is `build`, `r` is `recruit`, `a` is `assign`, `s` is `status`, `t` is `trade`. See [Command shortcuts](commands.md#command-shortcuts) for the full list.

---

## A week-long first run

A first run takes about **a week of real time** to reach the Modern Age, where a full run's prestige comes (prestige opens earlier, at the Medieval Age, as an early taste that pays little). A tick is 2 seconds of real time.

That is the first run. After a prestige, every age the run completed runs faster the next time, twice as fast after one completion and up to 4.2x after ten (4x at once for ages 6 or more behind your record), so later runs reach the Modern Age much sooner (see [Era Mastery](prestige.md#era-mastery)).

You don't need to watch it. Check in a few times a day, spend what has built up, and set the game up for the hours you're away:

- **Offline progress.** When you load a save, the game credits the time since you saved it, up to **24 hours**, at **50%** of your production. It runs that time in steps, so construction and research finish, the build plan starts what the income pays for, and your roster recruits and staffs as the hours pass. The welcome-back lines in the log say what happened.
- **The build plan.** Queue up to 60 builds, techs, deals and an advance with `plan build`, `plan research`, `plan deal` and `plan advance`. The game starts each item, in order, as the resources come in, and pays for it only when it starts. See [The Build Plan](plan.md).
- **Nothing is lost.** Stock has no limit, so everything you make while you are away is waiting when you come back. Nothing flows into a wonder's bank on its own: fill it with `wonder collect`, or let a planned wonder pay its bank from what you hold.
- **Storage.** Storage buildings hold nothing: each copy adds 1% to all production. They have no copy limit, each copy costs 1.75 times the one before it, and every age gate asks for five copies of the storage building of the age you are leaving.

---

## Catastrophes and the harbinger

From the Iron Era on, each era may hide a fated catastrophe. It can strike at any moment in the era, but never unannounced: a **harbinger** comes first, with a `⚑ Harbinger` badge on the status bar. Type `harbinger` to read it, and answer if you like: **Appease** lowers the odds, **Brace** softens the blow, **Invite** makes it certain. When a catastrophe strikes, nothing is lost until you choose: **Endure** (lose part of your buildings, stores and workers, and keep your age) or **Succumb** (fall back to the Primitive Age and keep a permanent legacy bonus). Higher faith lowers the odds of a strike. See [Catastrophes](catastrophe.md) and [The Harbinger](harbinger.md).

---

## Resource management tips

- Stock has no limit: what you make piles up, and a price is the only thing that holds you back. Each age's storage building adds 1% to all production a copy.
- **Food drain:** everyone eats the same amount, whatever building a worker staffs, scholars and the reserve included. That amount is set by your age: 0.06 food/tick per worker in the Primitive Age, rising 12% with each age. Keep food production above total drain. If food hits zero, one worker dies every 5 ticks (10 seconds) until food recovers. See [Food drain](workers-and-domains.md#food-drain) and [Starvation](workers-and-domains.md#starvation).
- **Food** runs out first early on. Keep your gathering camps staffed before anything else, then build a Story Circle: its scholars start research with no workers.
- Watch the rates in the Resources box. A negative rate is draining you.

## Morale

Morale is a percentage that multiplies **all worker output**. It **starts at 50%**, which is neutral. Above 50% your workers get a bonus that grows to **+20%** at the morale cap (100%, plus 5% for every wonder you build); below 50% they take a penalty that deepens to **x0.50** at the 10% floor. Morale drifts back toward 50% each tick, so a bonus has to be kept up.

Worship and culture buildings, your faith income, good events and age advances raise it. Starvation, a large army, many idle workers, bad events and catastrophes lower it. It shows as `Morale NN%` on the status bar, green for a bonus and red for a penalty, and on the People panel. See [Morale](morale.md) for the full system.

---

## Workers

Workers fall into 11 worker domains (food, lumber, masonry, faith and so on), each tied to its own buildings. A worker takes the domain of the building it works in. Workers move in on their own toward your housing while food allows, and go to work by your roster (`roster`; every domain is on auto by default). You can also recruit with `recruit [count|max]` and assign with `assign <building> [count|all]`.

Assigned workers raise a building's output. A building with no workers still produces at 20% (the floor), and a fully staffed one at 100%. Only food workers make food, but every worker eats it. See [People](workers-and-domains.md) for the full guide.

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

Advancing an age does **not** convert your buildings. Each one whose line has a newer building for the new age becomes an older building: it keeps producing, and you can upgrade it. Use `upgrade <building>` to convert your existing copies to the building of your age. For example, `upgrade gathering_camp` after entering the Stone Age turns your Gathering Camps into Forager Posts. An upgraded copy costs half of what a new copy of the new building costs, takes effect at once, keeps the building's workers, and counts toward the next age's building requirements. Upgrade your food buildings first after every advance. Storage buildings never upgrade: the ones you built keep counting, so start on the new age's storage. See [Buildings](buildings.md#building-upgrades) for the full guide.

---

## What are wonders?

Wonders are unique mega-structures (22 wonders in all, one per age) that grant **permanent civilization bonuses**. You need your age's wonder to **advance** to the next age. Each one costs about the same share of its age's economy, and you **bank** the resources before you build it. Stonehenge, the Bronze Age wonder, needs about 34K stone, 19K wood and 3.4K iron:

```
wonder collect all
build stonehenge
```

`wonder collect all` banks as much of every needed resource as you have; repeat it until the bank is full, then build. Each wonder can be built only once. They are listed in the **Wonders** panel (`wonders`). See [Wonders](wonders.md).

---

## The Map

Type `map` to open the **Map**: your empire drawn from your real game state (your buildings and the workers staffing them, your wonders, the civilizations you have met, your trade routes). It has two styles:

- **Roguelike** (the default): a glyph world seen from above. Your town grows out of your real buildings, quarter by quarter, with streets, walls and wonders. It has three zooms: **PgUp** zooms out to the region, the known world with the civilizations you have met, and **PgDn** zooms in to a district with building names.
- **Skyline**: your empire side-on as a panorama, one district for every age you have lived through. Windows light only where workers are staffed, and trade routes travel by land, sea or air. `←` and `→` scroll it, and PgUp and PgDn move half a screen.

Switch with `map style skyline` or `map style roguelike` (or the shorter `style skyline`). `map glyphs` chooses the glyph set; both are saved with your account.

The prompt keeps working while the Map is open. Letters you type go to the prompt, and the Map takes only the keys that print nothing: the arrows, PgUp and PgDn, Home and End, and, while the prompt is empty, Tab, Shift+Tab and Enter. In both styles an inspect cursor tells you what it is on and the command to type for it; Tab jumps between buildings and wonders, and Enter puts that command into the prompt. Esc closes the Map. On a large terminal the dashboard also shows a **mini map** above the Buildings list. `citymap` and `worldmap` still work and open the Map. See [The Map](map.md) for all the keys.
