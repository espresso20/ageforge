# How to play

AgeForge is a **command-driven idle civilization game**. Resources accumulate on their own, buildings construct over time, and you steer by typing commands at the prompt at the foot of the screen. The game keeps going while you are away (see [While you are away](#while-you-are-away)), so you can check in whenever you like.

---

## The main menu

The game opens on its main menu, and `Esc` brings you back to it from a game.

<figure class="screen" data-screen="menu"><figcaption>The main menu of an account with a game, at 120 columns by 40 rows: its own town behind the menu, Continue first, and a caption that names the town and its age.</figcaption></figure>

Once your account has a game, the menu is drawn over **your own town**: the map of your current game, in your glyph set and theme, at the age that game is in, moving as the [Map](map.md) does. It is a look at the save and nothing more. The game is not loaded or run, no time away is counted, and nothing is written, so looking at the menu costs you nothing and earns you nothing. **Continue** is first and already selected: press `Enter` and you are back in that game. Under it are the town, its age and its people, and beside it the save's name.

<figure class="screen" data-screen="menu-first"><figcaption>The main menu on a first visit: the wordmark as iron on the fire, the menu as a table of contents, and the forge on its plate.</figcaption></figure>

An account with no game yet sees the forge and a title page, with **New game** first. The same page shows if your save cannot be read.

| Entry | Key | What it does |
|---|---|---|
| **Continue** | `c` | Opens your current game. It is on the menu once the account has a save |
| **New game** | `n` | Asks you to name a civilization, then starts it (see [Starting a new game](saving-and-loading.md#starting-a-new-game)) |
| **Load game** | `l` | Opens the [Load Game browser](saving-and-loading.md#the-load-game-browser), where you can also mark your main game |
| **Badges** | `b` | Shows how many badges you hold and the title you wear, and opens the [Ranks screen](account.md#ranks). No game needs to be loaded |
| **Settings** | `s` | Opens a page of display settings, one row each with its value at the right: **Themes** (opens the [theme](themes.md) picker), **Map glyphs**, **Motion**, **Minimap** and **Discord status**, and in the game's own window **Font**, **Text size** (a key each way, the same as `Ctrl` or `Cmd` with `+` and `-`) and **Fullscreen** (the same as `F11`) (your age and town on your Discord profile while Discord is running; see [Account settings](account.md#account-settings)). A setting that is on or off is a slide switch: a click on the row flips it, and so does `Enter` on the picked row. The pointer lights the row under it and the picked row has a bar at its left. Each row's letter is drawn as a key at its right. `Enter` flips a switch, and `←` and `→` step the map glyphs through ascii and unicode. A change applies and saves at once, as the matching command does. `Esc` returns to the menu |
| **Accounts** | `a` | Opens the [Accounts panel](account.md#the-accounts-panel): switch, create, back up and wipe accounts, and delete an account's saves |
| **Check for updates** | `u` | Asks for the latest release (see [Updating](getting-started.md#updating)) |
| **Quit** | `q` | Leaves the game |

`↑` and `↓` (or `Tab` and `Shift+Tab`) move the selection, `Enter` or `Space` opens it, and each entry's letter opens it directly. With motion on, choosing an entry strikes the title: a flare and a burst of sparks, then the entry opens. The version is on the page, and so is **update available (u)** when a newer release is out.

In the window the mouse works on the menu too. The entry under the pointer is the selected one, and a click opens it as `Enter` does, on the Settings page as well. The pages the menu opens (Load game, Badges, Themes, Accounts) take it as well: a click picks a row or presses a button. So do the game's dialogs (a confirmation, a notice, the update check): a click on a button presses it.

**Which game Continue opens.** Your account remembers the game you played last, and that is your current game. If you keep several games and want one of them on the menu whatever you played last, mark it as your **main game** with `m` in the Load Game browser. See [Continue and your main game](saving-and-loading.md#continue-and-your-main-game).

**Motion and size.** `motion off` holds the menu to one still picture, like the maps (see [Motion](themes.md#motion)). The menu fits any window size from 80 columns by 24 rows up. With `map glyphs ascii` it is drawn without block characters. On a light theme the fire and the hot iron are ink on paper: hotter is darker.

---

## The core loop

```
Build → Staff → Research → Advance → Repeat
```

1. Build structures to add housing and production.
2. Staff them. Workers move in on their own toward your housing while food allows, and go to work by your [roster](workers-and-domains.md#the-roster). Steer the split with `roster`, or `recruit` and `assign` by hand.
3. Research technologies. Each one changes your production, army, housing or the commands you can use.
4. Advance to the next age with `advance` once the buildings it asks for stand and the age's wonder is built (see [Advancement Requirements](ages.md#advancement-requirements)).
5. Prestige for a lasting production bonus. After the first prestige, [Plan Template, Roster and Old Friends](prestige.md#carry-overs) carry into later runs. Prestige opens at the Medieval Age, but a run to the Modern Age (or later) pays far more. Every age the run completed then runs faster on the next run: twice as fast after one completion and up to 4.2x after ten, and 4x at once for ages 6 or more behind your record (see [Era Mastery](prestige.md#era-mastery)).

New to the game? [Your First Age](first-ten-minutes.md) walks you through the Primitive Age step by step, from the first `gather` to the Sacred Grove.

---

## The screen layout

```
┌─ Ada · Magistrate ─────────────────────── Research complete ── Esc menu ─┐
│ Bronze Age  Year c. 2,300 BC  │  People 906/1,200   Morale 54%           │  the header: now
│ To the Iron Age ████████░░ 49%  Knowledge ✓  People 2/3  Industry 1/2    │  the way to the next age
│ Waiting on: Scriptorium 3/5 · Library 0/2        Short of:  Stone        │  what it waits on
└──────────────────────────────────────────────────────────────────────────┘
┌─ Resources ───────────────────────────┬─ Buildings ───────┬─ Panels ─────┐
│ amounts and rates                     │ what you can      │ TOWN         │
│                                       │ build, and what   │ people ...   │
│                                       │ you own           │ PROGRESS     │
│                                       │                   │ research ... │
│                                       │                   │ WORLD        │
├─ Under construction ───┬─ People ─────┤                   │ trade ...    │
│ progress bars          │ pop and bar  │                   │ YOU          │
│                        │ four kinds,  │                   │ badges ...   │
│                        │ held back by │                   │              │
├─ Log ──────────────────┴──────────────┤                   │              │
│ events and news                       │                   │              │
│                                       │                   │              │
├─ Command ─────────────────────────────┴───────────────────┴──────────────┤
│ ❯ _                                                                      │
└──────────────────────────────────────────────────────────────────────────┘
```

That is the layout on a window of 120 columns by 36 rows or more. On a smaller one the Under construction box has its row to itself, and the People box sits under the Panels list.

The top of the screen is the **header**, one framed box that says the same things about your progress at every size.

- **The top edge** has your account's name when you have one, with the title you wear, and your civilization title in quotes. At its right is a **notice** for a few seconds of news (an age advance, a finished tech, a built wonder, a milestone, a new epoch, a harbinger, a catastrophe or an epoch event) and an `Esc` `menu` button. A long name gives way to the notice. A badge you earn opens a card of its own at the top right instead (see [Badges](account.md#badges)).
- **The first row** is now: your age, the calendar year, your people over your homes, morale and what it does to production, then your prestige level and a `DEV` mark in developer mode. A pending catastrophe, a pending Last Passage and a waiting harbinger show at its right as chips (a cheater badge shows there too). On a narrow window the row gives things up in a fixed order: what morale does to production, the long words of the chips, the year, the long figures of the people count (`8.31K` for `8,310`), your prestige level and `DEV`, and last the chips' words. The age, your people, morale and the chips' marks always stay.
- **The long bar** runs to the next age, with its percent. On a window of 100 columns by 30 rows or more the four pillars (Knowledge, People, Industry and Monument) sit beside it, each as a small meter with how many of its requirements stand (`2/3`), or a check once it stands. The bar's percent and the year are the same ones the [Advance panel](advance.md) shows, and `charter` opens it.
- **What it waits on.** A **Waiting on** line names what is still missing: each building the gate asks for with your count beside it, the keystone tech and the wonder's bank, and then the wonder itself. When the list is longer than the room, the line counts the rest (`+4 more`). When every pillar stands it says so and tells you to type `advance`. On a window under 100 columns by 30 rows the line shares the bar's row; from there up it has its own row, with a **Short of** list at its right. That list names the resources that a waiting building, or the wonder's bank, costs more of than you hold. The Resources box below lists every resource with its amount and its rate.

The header is four rows on a window under 100 columns by 30 rows and five from there up. On the plain glyph tier (`map glyphs ascii`) the bars and marks are drawn in plain characters. The [Advance](advance.md) page also covers the experimental gate (the environment variable `AGEFORGE_GATES=v2`), which asks for population, income and a key tech in place of building counts.

Every part of the header takes a click: People and Morale open People, the bar, the pillars and the waiting line open the Advance panel, an alert chip opens its panel (`catastrophe` or `harbinger`), and the `Esc` `menu` button presses `Esc`. The pointer is a hand over each, and the part under it lights up.

Every panel you open (`research`, `trade`, `map`, `help` and so on) takes the whole screen but the command bar. Its name is on the top edge, the keys it takes are on the line along its foot, and `Esc` closes it back to the dashboard. The prompt keeps working in most panels. On the panels that are text (Milestones, Army, Trade routes, Factions, Statistics, Logs, Epoch, Civilization History, Buildings and Help), `↑` `↓` `PgUp` `PgDn` `Home` and `End` scroll the panel while the prompt is empty, and anything else you type goes to the prompt. The Wonders, Harbinger and Build Plan panels have keys that are letters (`G`, `A` `B` `I`, `U` `D` `X` `C`), so they keep the keyboard: press `Esc` before you type a command.

In the window, the mouse wheel scrolls a text panel too, whatever is in the prompt, and a click in the command bar puts the cursor at the character you clicked. Over the dashboard's Buildings list the wheel moves the list one building at a time, and over the Resources box it turns the box's page when it has more than one. While a dialog or a floating window is open, the mouse belongs to it: a click beside it does nothing. On the dashboard, a click on a name in the Panels list opens that panel, as typing the name does; what you have typed in the prompt stays as it is. A click on a building in the Buildings list writes its `build` command in the prompt, ready for `Enter`, and a double click builds one. A click on a box opens the panel that is that box in full: the mini map opens the Map panel, the People box People, the Log the full record and Under construction the build plan. On the Load game, Accounts and Themes pages the key buttons along the foot take clicks: a click on one presses its key, so `Esc Back` leaves the page. The panels do the same: the keys along a panel's foot are buttons, so a click on `Esc close` closes it, and on a cap that shows two keys (`↑↓`) each key is its own button. The `×` in a panel's top right corner closes it too. The Research, Badges and Map panels do not have these buttons yet. A right click means Back: it does what `Esc` does on a panel, a dialog or a page. On the dashboard with no panel open it does nothing, so it never takes you out of a game. A building in the Buildings list and a name in the Panels list light up while the pointer is on them. The pointer turns into a hand over anything a click acts on (a menu entry, a button, a row, a badge, a tech, a tab) and into a text bar over a line you can type in.

<figure class="screen" data-screen="dashboard"><figcaption>The dashboard in the Bronze Age, on a window of 120 columns by 40 rows: the header across the top; resources at the top left with construction and the People box side by side under them, and the log below; the Buildings list in the middle; the Panels list on the right, with the mini map under it when the window is tall enough.</figcaption></figure>

The dashboard fits any window size from 80 columns by 24 rows up, and no row on it wraps. Each resource has one row in the Resources box: its amount, its rate and a marker when it is falling (`▼`). Culture and faith add a bar toward their next step. On a narrow window the rows give up the bar first. On a short window the Panels list is set in two columns, or packed several names to a row on the smallest screens. When the Resources box cannot show every resource it shows a page of them and says so: **Ctrl+R** turns the page. The mini map is on until you type `minimap off`. It is a box under the Panels list, square when there is room, and shows from 80 columns by 36 rows.

Resting the pointer on a thing for about half a second shows a small tooltip, which goes on any move, key or click. On the dashboard, an off Build button gives the reason it is off, a card's price lists each resource as have and need with the short ones marked, and a line in the Resources box lists what its rate is made of. Everything a tooltip says is also on the screen or in a command; the tooltip only explains it.

**The Under construction box.** It has a line for each building under way (copies of one building share a line), with a bar and the time left. A short box sums up instead (`9 under way, the first in ~4m 38s`).

**The People box.** From 120 columns by 36 rows the People box is the right half of the Under construction row, and the row is as tall as the box. It shows the headcount over your homes (with how full they are when there is room to say so), one bar across the homes in the colors of the four kinds of person (workers, the reserve, soldiers and scholars), a line for each kind in the age's own word, and what holds your population back. The soldiers' line shows the soldiers at their posts over all the posts they have while some wait on food, and how many are out on a mission. `Held by: food` carries the net food a tick and how many more it would feed once the window is 42 rows tall. When nothing holds the population back the box says `Room to grow` and counts the empty homes. On a smaller window the box is the short one under the Panels list, with the same lines in less room. The [People panel](workers-and-domains.md#the-people-of-your-town) has the whole picture.

**The Panels list.** On the same windows the Panels column is wider, and the panels are set in four groups, each with a few words on what it is for: Town (`people`, `plan`, `wonders`, `map`), Progress (`milestones`, `research`, `charter`, `epoch`, `history`), World (`expedition`, `army`, `trade`, `factions`, `harbinger`) and You (`badges`, `stats`, `logs`, `help`). The panel that is open is marked. On a smaller window, or when the column is too short for the groups, the list is the panel names alone, and it never drops one.

---

## Navigation

- Type a panel's name to open it: `research`, `army`, `trade`, `factions`, `map`, `help` and so on.
- `help` opens the Help panel: a full command reference plus the list of every panel you can open.
- **Tab** or **→** takes the completion shown in dim text after the cursor, and **Tab** again takes the next one (**Shift+Tab** goes back). **Enter** runs a completion that makes a whole command. A command that cannot be undone (`sell`, `dismiss`, `load` and `quit`, for example) is put in the prompt instead, and a second **Enter** runs it.
- **PgUp / PgDn** turn the Buildings list a page (with a text panel open, they scroll the panel). A Mac keyboard has no Page keys: **Option+↑ / Option+↓** do the same everywhere, and **Option+←** and **Option+→** are Home and End. The list never shows part of an entry: an entry that does not fit whole waits for the next page, and the box says how many are above and below.
- **Ctrl+R** turns the Resources box to its next page, when it has more resources than it can show.
- **↑ / ↓** step through your last 50 commands (with a text panel open and the prompt empty, they scroll the panel).
- **Esc** closes the open panel. With no panel open, it saves the game, stops it and returns to the [main menu](#the-main-menu), where **Continue** takes you back in.
- Common commands have short names: `b` is `build`, `r` is `recruit`, `a` is `assign`, `s` is `status`, `t` is `trade`. See [Command shortcuts](commands.md#command-shortcuts) for the full list.
- Display settings you can change at any time are `theme`, `motion`, `minimap` and `map glyphs` (see [Themes](themes.md)).

---

## While you are away

You don't need to watch the game. Spend what has built up when you check in, and set the game up for the hours you're away. At the base speed a tick is 2 seconds of real time.

- **Offline progress.** When you load a save, the game credits the time since you saved it, up to **24 hours**, at **50%** of your production. It runs that time in steps, so construction and research finish, the build plan starts what the income pays for, and your roster recruits and staffs as the time passes. The welcome-back lines in the log say what happened.
- **The build plan.** Queue up to 60 items (builds, upgrades, deals and an advance) with `plan build`, `plan upgrade`, `plan deal` and `plan advance`, and queue techs with `plan research`, which take no room in the 60. The game starts each item, in order, as the resources come in, and pays for it only when it starts. See [The build plan](plan.md).
- **Nothing is lost.** Stock has no limit, so everything you make while you are away is waiting when you come back. Nothing flows into a wonder's bank on its own: fill it with `wonder collect`, or let a planned wonder pay its bank from what you hold.
- **Storage.** Storage buildings hold nothing: each copy adds 1% to all production. They have no copy limit, each copy costs 1.75 times the one before it, and every age gate asks for five copies of the storage building of the age you are leaving.

---

## Catastrophes and the harbinger

From the Iron Era on, each era may hide a fated catastrophe. It can strike at any moment in the era, but never unannounced: a **harbinger** comes first, with a `⚑ Harbinger` chip in the header. Type `harbinger` to read it, and answer if you like: **Appease** lowers the odds, **Brace** softens the blow, **Invite** makes it certain. When a catastrophe strikes, nothing is lost until you choose: **Endure** (lose part of your buildings, stock and workers, and keep your age) or **Succumb** (fall back to the Primitive Age and keep a permanent legacy bonus). Higher faith lowers the odds of a strike. See [Catastrophes](catastrophe.md) and [The Harbinger](harbinger.md).

---

## Resource management tips

- Stock has no limit: what you make piles up, and a price is the only thing that holds you back. Each age's storage building adds 1% to all production a copy.
- **Food drain:** everyone eats the same amount, whatever building they work in. Workers, the reserve, scholars and soldiers all count. That amount is set by your age: 0.06 food/tick per person in the Primitive Age, rising 12% with each age. Keep food production above total drain. If food hits zero, one worker dies every 5 ticks (10 seconds at the base speed) until food recovers. See [Food drain](workers-and-domains.md#food-drain) and [Starvation](workers-and-domains.md#starvation).
- **Food** runs out first early on. Keep your gathering camps staffed before anything else, then build a Story Circle: its scholars start research with no workers.
- Watch the rates in the Resources box. A negative rate is draining you.

## Morale

Morale is a percentage that multiplies your buildings' output and your scholars' research work. A new game **starts at 50%**, which is neutral (a prestige run starts at 70%). Above 50% you get a bonus that grows to **+20%** at the morale ceiling (100%, plus 5 points for each wonder you build and for three techs). Below 50% you take a penalty that deepens to **x0.50** at the 10% floor. Morale drifts back toward 50% each tick, so a bonus has to be kept up.

Worship and culture buildings, your faith income, good events and age advances raise it. Starvation, a large army, many idle workers, bad events and enduring a catastrophe lower it. It shows as `Morale NN%` in the header, green for a bonus and red for a penalty, and on the People panel. See [Morale](morale.md) for the full system.

---

## Workers

Workers fall into 11 worker domains (food, lumber, masonry, faith and so on), each tied to its own buildings. A worker takes the domain of the building it works in. Workers move in on their own toward your housing while food allows, and go to work by your roster (`roster`; every domain is on auto by default). You can also recruit with `recruit [count|max]` and assign with `assign <building> [count|all]`.

Assigned workers raise a building's output. A building with no workers still produces at 20% (the floor), and a fully staffed one at 100%. Food comes from food buildings, but everyone eats it. Scholars come with knowledge buildings and soldiers with military buildings. See [People](workers-and-domains.md) for the full guide.

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

Advancing an age does **not** convert your buildings. Each one whose line has a newer building for the new age becomes an older building: it keeps producing, and you can upgrade it. Use `upgrade <building>` to convert your existing copies to the building of your age. For example, `upgrade gathering_camp` after entering the Stone Age turns your Gathering Camps into Forager Posts. An upgraded copy costs half of what a new copy of the new building costs and takes effect at once. Its workers move with it as far as the new building has room, and it counts as a copy of the new building toward the buildings the age gate asks for. Upgrade your food buildings first after every advance. Storage buildings never upgrade: the ones you built keep counting, so start on the new age's storage. See [Buildings](buildings.md#building-upgrades) for the full guide.

---

## What are wonders?

Wonders are unique mega-structures (22 wonders in all, one per age) that give civilization-wide bonuses. You need your age's wonder to **advance** to the next age. From the Stone Age on, each wonder also needs its keystone tech before you can build it. You **bank** the resources before you build the wonder. Stonehenge, the Bronze Age wonder, needs 54.4K stone, 30.4K wood and 5.44K iron, and the Calendar tech:

```
wonder collect all
build stonehenge
```

`wonder collect all` banks as much of every needed resource as you have. Repeat it until the bank is full, then build. Each wonder can be built only once. They are listed in the **Wonders** panel (`wonders`), where `G` or `Tab` opens a gallery of the ones you have built. See [Wonders](wonders.md).

---

## The Map

Type `map` to open the **Map**: your empire drawn from your real game state (your buildings and the workers staffing them, and your wonders). It is a glyph world seen from above. Your town grows out of your real buildings, quarter by quarter, along roads that grow from the town square. It has three zooms: the whole town, the close-up and a district. **PgUp** (or **-**) zooms out a step, and **PgDn** (or **=**) zooms in a step.

`map glyphs` chooses the glyph set (`ascii` or `unicode`), and it is saved with your account.

The prompt keeps working while the Map is open. Letters you type go to the prompt, and the Map takes only the keys that print nothing: the arrows, PgUp and PgDn, Home and End, and, while the prompt is empty, Tab, Shift+Tab, Enter and the zoom keys `-` and `=`. An inspect cursor tells you what it is on and the command to type for it. Tab jumps between buildings and wonders, and Enter puts that command into the prompt. Esc closes the Map. On a large window the dashboard also shows a **mini map** under the Panels list (`minimap off` hides it). `citymap` and `worldmap` also open the Map. See [The Map](map.md) for all the keys.
