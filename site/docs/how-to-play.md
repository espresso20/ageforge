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

1. **Build** structures to add housing, storage and production
2. **Staff** them: workers arrive on their own while housing and food allow and go to work by your [worker shares](workers-and-domains.md#worker-shares). Steer the split with `workers share`, or `recruit` and `assign` by hand
3. **Research** technologies that multiply your output
4. **Advance** to the next age with `advance` once its requirements are met and the age's wonder is built
5. **Prestige** for points that buy the [legacy kit](prestige.md#the-legacy-kit), which carries your plan, worker shares and civilizations into later runs. Prestige opens at the Medieval Age, but a run to the Modern Age (or later) pays far more. Every age the run completed then runs faster on the next run ([Era Mastery](prestige.md#era-mastery))

New to the game? [Your First Age](first-ten-minutes.md) walks you through the Primitive Age step by step, from the first `gather` to the Sacred Grove.

---

## The screen layout

```
┌──────────────────────────────────────────────────────────────────────────┐
│ Name · Stone Age  ◈ Stone Era  |  Pop: 18/30  Morale 54%  |  hint        │  status bar
│ Next Age: Bronze Age  ✗ Food 3.1K/4K  ✓ Knowledge 1.5K/1.5K  ✗ Stone ... │  next age row
├─ Resources ──────────┬─ Buildings ──────────────────────┬─ Panels ───────┤
│ amounts and rates    │ (mini map above it: minimap on)  │ panel names    │
├─ Under construction ─┤ what you can build, what you own │ ...            │
│ progress bars        │ PgUp / PgDn turn the page        │                │
├─ Log ────────────────┤                                  ├─ Workers ──────┤
│ ...                  │                                  │ pop, idle,     │
│                      │                                  │ housing, food  │
├─ Command ────────────┴──────────────────────────────────┴────────────────┤
│ > _                                                                      │
└──────────────────────────────────────────────────────────────────────────┘
```

The status bar also carries your prestige level and civilization title once you have them, and badges for a waiting harbinger or catastrophe. The Next Age row marks each requirement ✓ or ✗, and names the age's wonder until it stands. Panels you open (`research`, `trade`, `map` and so on) open over the dashboard, and the prompt keeps working while they are open.

<figure class="screen" data-screen="dashboard"><figcaption>The dashboard in the Bronze Age, on a terminal of 120 columns by 40 rows: resources, construction and the log on the left, the Buildings list in the middle, the panel names and the Workers box on the right.</figcaption></figure>

The dashboard fits any terminal from 80 columns by 24 rows up, and no row on it wraps. Each resource has one row: its amount over its cap, its rate, a bar, and a marker when the store is nearly full (`◈`) or falling (`▼`). On a narrow terminal a row gives up the bar first, then the cap, and at 80 columns `/tick` is written `/t`. On a short terminal the Panels list is set in two columns, and when the Resources box cannot show every resource it shows a page of them and says so: **Ctrl+R** turns the page. The mini map is off until you type `minimap on`; it then shows from 105 columns by 34 rows.

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

A first run takes about **a week of real time** to reach the Modern Age, where a full run's prestige comes (prestige opens earlier, at the Medieval Age about 20 hours in, as an early taste that pays little). A tick is 2 seconds of real time.

That is the first run. After a prestige, every age the run completed runs faster the next time, twice as fast after one completion and up to 4.2x after ten (4x at once for ages 6 or more behind your record), so later runs reach the Modern Age much sooner (see [Era Mastery](prestige.md#era-mastery)).

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
