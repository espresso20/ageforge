# Commands

All commands are typed at the `>` prompt at the bottom of the screen. `Enter` runs what you typed, `Tab` takes the highlighted suggestion, and `↑`/`↓` navigate history.

Counts (`recruit 5`, `build farm 3`, `sell hut 2`, ...) are whole numbers from 1 to 1,000,000, and amounts (`gather`, `trade`, `wonder collect`) are positive numbers. Anything else is refused with the command's usage line rather than guessed at.

---

## Timers and durations

Every duration and countdown in the game reads as an approximate wall-clock time rather than a tick count: `~38s`, `~4m 44s`, `~1h 12m`. Durations that are rolled per launch — expedition lengths — read as a range, `~2m – 3m 20s`. Readings carry two units of precision at most.

The `~` is doing real work. A tick is not a fixed amount of real time: `tick_speed` bonuses and the speed multiplier both shorten it, so every reading is computed from your **current** tick rate and moves as that rate does. Finish a tech that grants tick speed and the countdown you were already watching gets shorter.

Balance values are still *defined* in ticks — a tech costs so many ticks of research, an event lasts so many ticks — and this wiki quotes those tick figures where the tick count is the mechanic. The game shows you the wall-clock conversion of them. Raw tick counts survive in exactly one place a player can reach: the `dump` debug export, which prints ticks alongside the wall-clock reading.

---

## Building

| Command | Description |
|---|---|
| `build <key>` | Start constructing a building |
| `sell <building> [count]` | Demolish a building and recover 50% of its build cost. Workers are returned to idle. |
| `upgrade <building> [count\|all]` | Convert building copies to the next-age tier equivalent, paying only the cost delta (new copy cost minus 50% of the old copy's sell value). Defaults to all copies if no count given. Stops at the new building's max count. Storage buildings are never upgraded — older storage keeps counting. |
| `gather <resource> [amount]` | Manually gather food, wood, or stone (max 25 per command). Disabled from the Renaissance Age onward — works through the Medieval Age only. |
| `buildings` | Open the **Buildings** panel |

**Example:**
```
build hut
build lumber_mill
sell lumber_mill
sell gathering_camp 3
upgrade forager_post
upgrade forager_post 3
upgrade forager_post all
gather wood 5
```

---

## Build Plan

A list of builds and techs the game starts for you, in order, as the resources come in: while you play and while you are away (offline catch-up runs it too). Each item is paid for when it starts, not when you add it. Full rules on the [Build Plan](plan.md) page.

| Command | Description |
|---|---|
| `plan` | Open the **Plan** panel: the items, what each costs next and whether it can start. `↑`/`↓` select, `U`/`D` move the selected item, `X` removes it, `C` twice clears the plan. |
| `plan build <building> [count]` | Add copies of a building of this age (default 1). Adding more of the building at the end of the plan adds to that item. |
| `plan research <tech>` | Add a tech. Techs start one at a time, in plan order; a prerequisite can be planned before it. |
| `plan trade <from> <to> [amount]` | Sell `from` for `to` at the market as it comes in, until `amount` of `to` is bought; with no amount, keep `to` topped up until you remove the item. Needs a trade building to sell. |
| `plan advance` | Advance to the next age as soon as its requirements are met. The next age's buildings and techs can be planned too; they wait for the advance. |
| `plan list` | Print the plan with each item's status |
| `plan remove <n>` | Remove item `n` |
| `plan up <n>` | Move item `n` one place up |
| `plan down <n>` | Move item `n` one place down |
| `plan clear` | Empty the plan |

```
plan build hut 10
plan build gathering_camp 5
plan research tool_making
plan trade gold stone 50000
plan advance
plan up 3
```

---

## Workers

| Command | Description |
|---|---|
| `recruit [count\|max]` | Recruit one or more workers from available housing capacity |
| `assign <building_key> [count\|all]` | Assign workers to a building (domain inferred from building) |
| `unassign <building_key> [count\|all]` | Unassign workers from a building (returns them to idle pool) |
| `dismiss <building_key> [count\|all]` | Permanently remove workers from a building and from the population pool entirely |
| `workers` | Open the worker status overlay (summary, slot utilization, domain breakdown) |
| `status` (or `s`) | Print a status summary: age and tick, every unlocked resource with amount, storage and rate, and your population by class with idle counts and assignments |
| `rates` | Print a per-resource rate breakdown — buildings, workers, research, events, trade, bonuses and food drain |

Workers are recruited generically from available housing capacity and assigned to buildings, where they become that building's domain class (Gatherer, Lumberjack, etc.). `unassign` returns workers to idle; `dismiss` reduces total population.

```
recruit
recruit 3
recruit max
assign gathering_camp 3
assign library all
unassign shrine
unassign shrine all
dismiss shrine 2
dismiss barracks all
workers
```

See [Workers & Domains (Reference)](workers-and-domains.md) for the full domain table and efficiency formula.

---

## Research

| Command | Description |
|---|---|
| `research` | Open the **Research** panel |
| `techs` | Opens the same **Research** panel |
| `research list` | List the technologies you can research now, with their knowledge cost |
| `research <key>` | Start researching a technology |
| `research cancel` | Cancel active research (progress is lost) |

```
research tool_making
research agriculture
research iron_smelting
```

Tech keys are shown in the **Research** overlay (`research`) (dim grey when locked, gold circle when available).

---

## Expeditions & Army

| Command | Description |
|---|---|
| `expedition` | Open the **Expeditions** panel — scouting missions (shorthand: `exp`) |
| `expedition list` | List scouting expeditions available in your current age (shorthand: `exp list`) |
| `expedition <key>` | Send a scouting expedition — costs resources, never soldiers (e.g. `expedition scout_ruins`; shorthand: `exp <key>`) |
| `army` | Open the **Army** panel — soldier overview and military campaigns |
| `campaign list` | List military campaigns available in your current age (`campaign` alone does the same) |
| `campaign <key>` | Wage a military campaign — spends soldiers, plus any resource cost (e.g. `campaign raid_bandits`) |
| `speed [multiplier]` | Set game speed (1.0, 1.5, 2.0 … +0.5 per wonder built) |

```
expedition
expedition list
expedition scout_party
expedition scout_ruins
army
campaign list
campaign raid_bandits
speed 1.5
```

Timed missions come in two kinds, split across two panels. **Scouting** expeditions (`scout_party`, `scout_ruins`, `naval_expedition`) cost only resources — **0 soldiers** — and are available early, before the `soldiers` resource exists at the Iron Age. You go on these with `expedition <key>` from the **Expeditions** panel. **Military campaigns** (everything else) **spend the `soldiers` resource** at launch, plus any resource cost. You wage these with `campaign <key>` from the **Army** panel. Either way the cost is deducted whether the run succeeds or fails; only the reward differs. One scouting expedition **and** one military campaign can run at the same time (one of each category), but not two of the same category.

Keys with underscores can be typed with spaces: `expedition scout ruins` is equivalent to `expedition scout_ruins`. If you run a campaign key through `expedition` the game refuses and points you to `campaign <key>`; likewise running a scouting key through `campaign` redirects you to `expedition <key>`.

From the **Industrial Age** you can build a **Geographic Society**, which sends scouting parties out by itself — no command needed. It uses the same single scouting slot and waits whenever you have a party in the field, so dispatching by hand always takes priority. The more Societies you build and staff, the shorter the wait between automatic parties, though it stays slower than sending them yourself. See [Automatic dispatch](military.md#automatic-dispatch-the-geographic-society).

---

## Trade

| Command | Description |
|---|---|
| `trade` | Open the **Trade** panel |
| `trade list` | List the market's exchange rates (build a market first) |
| `trade <from> <to> <amount>` | Exchange resources at the market rate (needs a market), e.g. `trade wood stone 100` |
| `trade route` (or `trade route list`) | List your active trade routes and the routes you can start |
| `trade route start <key>` | Activate a trade route |
| `trade route stop <key>` | Cancel an active trade route |
| `blackmarket` (or `bm`) | Show black-market status — culture cost, payout odds, and cooldown (Colonial Age+) |
| `blackmarket <resource>` | Run a high-risk culture deal for a chance at a big haul of the chosen resource |
| `trade black <resource>` | Alias for `blackmarket <resource>` |
| `factions` | Open the **Factions** panel — live favours and setbacks, Geographic Society status, and the full 11-civ roster (personality, backstory, strength, opinion, status, bonuses, war + lent-worker state) |
| `diplomacy` (or `dip`) | Opens the same **Factions** panel |
| `diplomacy ally <civ>` | Ally with a civilization (opinion ≥ 50, costs 500 gold) |
| `diplomacy rival <civ>` | Declare a rivalry |
| `diplomacy embargo <civ>` | Embargo a civilization (counts as a war provocation) |
| `diplomacy gift <civ>` | Send a gift to improve opinion (+15, costs 200 gold) |
| `diplomacy neutral <civ>` | Reset a civilization to neutral |
| `diplomacy tribute <civ>` | Sue for peace with a civilization at war (pays gold + culture, scaled to its strength) |
| `diplomacy raid <civ>` | Raid a civilization's trade route (-20 opinion; a war provocation) |

```
trade list
trade wood stone 100
trade route start coastal_market
trade route stop coastal_market
factions                   # opens the Factions panel
diplomacy                  # the same panel, under its older name
diplomacy gift merchant_guild
diplomacy ally merchant_guild
diplomacy tribute ironhold_clans   # end a war you'd rather not fight
```

Active trade routes run for a fixed duration. Routes whose imports include a resource specialised in by a civilization you're **at war with or have embargoed** are **disrupted** (no income) until the conflict ends — see [Trade Disruption](trade.md#trade-disruption-war-amp-embargo). **Harbours** (`harbor` → `logistics_hub`) boost the income of every active route. Check the **Trade** overlay (`trade`) for rates, and the **Factions** panel (`factions`, or the older `diplomacy` / `dip`) for live favours, Geographic Society status and civ standings — see [The Factions panel](trade.md#the-factions-panel). Bare `diplomacy` opens that panel; add an action (`ally`/`rival`/`embargo`/`gift`/`neutral`) to act on a faction directly. You meet civilizations by **running scouting expeditions**, not by building anything — but once you have met them, a staffed **Embassy** (Colonial Age) or **Grand Embassy** (Industrial Age) passively raises opinion with your non-hostile factions — see the [Trade & Diplomacy](trade.md#embassy-buildings) wiki page.

---

## Wonders

| Command | Description |
|---|---|
| `wonder collect <resource> <amount>` | Bank resources toward a wonder |
| `wonder overflow` | Show whether overflow is on |
| `wonder overflow on` | Bank what full stores would waste into the current wonder (the default) |
| `wonder overflow off` | Let production over a storage cap be lost instead |
| `build <wonder_key>` | Build the wonder once its bank is full |

```
wonder collect wood 1000
wonder collect stone 500
build great_monolith
```

**Overflow.** While overflow is on, production that a full store would throw away goes into the current age's wonder bank instead, for every resource the wonder still needs and only up to what it still needs. It never takes from what you hold, works during offline catch-up too, and says so in the log when it finishes a resource's part of the bank. See [Wonders](wonders.md#overflow).

Wonders are shown in **Wonders** overlay (`wonders`) with progress bars for each required resource. Each completed wonder now displays a colour sprite thumbnail next to its name in the Wonders overlay. Completed wonders also appear on the City Map as the largest, most ornate central complexes — an era-appropriate silhouette (a ziggurat in the ancient ages, a cathedral/keep in the medieval ages) in muted, in-family colours (see [City Map](#city-map) below).

---

## Map Views

There are two map views — a close-up of **your own settlement** and a zoomed-out view of the **wider world**. The **City Map** is theme-aware and retints live when you switch themes; the **World Map** is drawn in an era-specific **cartographic medium** that evolves as you advance.

| Command | Description |
|---|---|
| `citymap` | Open the **City Map** — a theme-aware procedural rendering of your settlement, with per-age layouts, roads, and your actual buildings drawn as lineage-coloured markers |
| `map` | Alias for `citymap` (kept for muscle memory) |
| `worldmap` | Open the **World Map** — a seeded continent (elevation, biomes, coastlines, rivers) redrawn each age in that era's cartographic medium; beyond the planet it becomes a strategic star-map of your empire and the rival factions |

```
citymap
worldmap
```

### City Map

The City Map (`citymap`, also `map`) renders your civilization as a **top-down pixel-art city** — you look straight down at the roofs, streets and squares of one living settlement, and it re-skins to the current era as you advance. Every colour is drawn from your **active colour theme**, so switching themes retints the whole city instantly. There is no world terrain on this view (the biome map lives on the **World Map**); the ground is a quiet, era-tinted surface and every green thing — gardens, ponds, street-trees — is **built**.

- **Layout** — the city is a compact cluster of **wards** (blocks); the **streets are the gaps between them**, a connected web of thin lanes. Towns come in four **forms** picked per civ + era — rambling **organic**, **radial** (a hub with a ring road), **grid**, and **ribbon** (strung along a road) — so no two civs look alike. The whole city always fits the panel and densifies as you grow (near 1:1 building-to-roof at low counts, packed-but-legible at high counts). Layout is stable and grows in place: new buildings slot into the existing fabric.
- **City Center & wonders** — the heart is a dressed **town square** (paved ground + era props). Built **wonders** are the central anchors the town hugs, each drawn as a dominant, unmistakable complex — a **ziggurat** in the ancient ages, a **cathedral/keep** in the medieval ages.
- **Per-era re-skin** — roofs, ground, streets, walls and props all restyle by age while the bones persist: earthy **thatch huts** on winding dirt lanes (Primitive/Stone) → **clay-tile mudbrick** town ringed by a **mudbrick wall with gates** (Bronze/Iron/Classical) → **slate-roofed timber** town ringed by a **stone wall with towers and a gatehouse** (Medieval/Renaissance) → open, wall-less **rowhouse grids** and, later, **towers**, **arcologies** and **domes** (Industrial and beyond). Walled ages leave **gates** where the main streets exit; industrial-and-later cities are open sprawl.
- **Buildings & labels** — every distinct building you've built appears as **count-scaled top-down roofs**, drawn by an **atlas** (huts round, longhouses elongated, temples ornate, camps as tents, workshops flat, wonders grand) in the era's roof material with a subtle per-lineage tint, so a domain reads by roof shape and hue without a wall of text. Only **key landmarks** are labelled (the City Center, wonders, and a promoted hero when you have no civic building yet) — as soft pill banners that stay readable over any roof.

### World Map

The World Map (`worldmap`) is a single **seeded world** — one continent with elevation, biomes, coastlines and rivers — that is the **same land every game** on your account. What changes as you advance is the **cartographic medium** it is drawn in: a charcoal cave-sketch in the Primitive Age, inked parchment with a compass rose in the Medieval, a satellite mosaic in the Modern, a neon holo-grid in the Cyberpunk. Ages 1–17 each get their own medium.

Once you leave the planet (Space onward) the World Map stops being a map of land and becomes a **strategic star-map** — your empire against the rival diplomacy factions competing for control. Standings read the same everywhere through signal colours: **at-war red, ally green, mercantile gold, neutral steel-blue**, with your own seat as the command hub. The five cosmic ages each get their own strategic view, from a home star-cluster up to an ascension lattice.

See **[The World Map](world-map.md)** for the full per-age breakdown.

---

## Prestige

| Command | Description |
|---|---|
| `prestige` | View prestige status and available points. In the Cosmic Era it also shows the chance of the Last Passage and which figure is warning of it, and shows the Cosmic Legacy if you hold it |
| `prestige confirm` | Explain what confirming would do. In the Cosmic Era this includes what Endure (share of this run's points) and Succumb (the Cosmic Legacy) would give you if the Last Passage comes |
| `prestige confirm yes` | Trigger prestige reset (requires Modern Age). In the Cosmic Era it first rolls the Last Passage; if it comes, prestige waits for your choice |
| `prestige shop` | View prestige upgrade list |
| `prestige buy <key>` | Purchase a prestige upgrade |

```
prestige confirm yes
prestige buy gather_boost
prestige buy tick_speed
```

Available upgrades and costs are shown in **Stats** overlay (`stats`).

See [The Last Passage](prestige.md#the-last-passage) for what can happen when you prestige from the Cosmic Era.

---

## Festival

| Command | Description |
|---|---|
| `festival` | Show festival status — culture cost, current culture, and the buff it grants |
| `festival confirm yes` | Hold a cultural festival now — spends culture for a temporary production boost |

Spend a lump of **culture** (max(2,000, 5% of your culture storage cap)) for **+20% to all production for 150 ticks** (~5 minutes). There's a **300-tick cooldown** (~10 minutes) between festivals, so it stays a rare, deliberate boost rather than a per-tick reflex. This is one of the culture sinks — see [Resources](resources.md#culture). Prestige gates remain the primary long-term culture sink.

---

## Milestones & Epochs

| Command | Description |
|---|---|
| `milestones` (or `ms`) | Open the **Milestones** panel — chain progress, earned titles and active speed boosts |
| `epoch` | Open the **Epoch** panel |

See [Milestones](milestones.md) and [Epochs](epochs.md).

---

## Civilization History

| Command | Description |
|---|---|
| `history` | Open the Civilization History overlay — braille line graphs of key metrics over time |

The History overlay shows 7 live graphs: **Population**, **Food Rate**, **Knowledge Rate**, **Faith**, **Morale**, **Production Bonus**, and **Tick Speed**. Each graph covers up to ~1 hour of rolling history (300 samples, one every 10 ticks). Age advances appear as `│` markers across all graphs so you can correlate events with metric changes.

Graphs appear after ~30 seconds of play. History is saved and restored automatically. See [History](history.md) for details.

---

## Catastrophe

| Command | Description |
|---|---|
| `catastrophe` (or `cat`) | Reopen the Endure / Succumb choice for a pending catastrophe or Last Passage. With nothing pending, show the catastrophe odds for the next epoch transition (in the Cosmic Era, for the Last Passage at your next prestige) |

A pending catastrophe blocks `advance` and `prestige confirm yes` until you choose. A pending Last Passage blocks only `prestige confirm yes`. In the choice modal, **E** endures, **S** succumbs and **Esc** closes it without choosing; the status bar shows a pending badge until you decide. There is no Defer option, and no command to trigger a catastrophe directly; the harbinger's Invite (below) is the only way to choose one. See [Catastrophe](catastrophe.md).

---

## Harbinger

| Command | Description |
|---|---|
| `harbinger` (or `harb`) | Open the **Harbinger** panel. Also listed under Panels in the sidebar. With no harbinger present it says so and describes the outlook for your next epoch transition |
| `harbinger appease` | Buy the next Appease level: a quarter of what a moderate faith economy makes over the epoch's ages, in faith, and the same for culture from the Steel Era on (level 2 double; see [Harbinger](harbinger.md#what-it-costs-by-epoch) for the prices). Each level multiplies the real catastrophe chance by 0.6. Two levels at most; refused after Invite |
| `harbinger brace` | Buy the next Brace level: 12% of the most the epoch still asks of each resource you had when it began, except faith and culture (24% for level 2). An Endure then destroys 15% / 10% of buildings and keeps 30% / 45% of stored resources (in the Cosmic Era it keeps 70% / 85% of the run's prestige points instead of 50%). Two levels at most |
| `harbinger invite` | Guarantee the catastrophe at this transition (in the Cosmic Era, the Last Passage at your next prestige). Free, and can't be undone |

Every epoch whose transition can bring a catastrophe (Stone to Neon Era) has a harbinger thread, and so does the Cosmic Era, whose thread warns of the Last Passage (your next prestige). Each thread runs from its first age until its passage, with each age's figure taking up the warning in turn. The actions work in any age of the thread, cost the same in each, and carry over between figures.

**Keys in the Harbinger panel:**

| Key | Action |
|---|---|
| `A` | Appease |
| `B` | Brace |
| `I` | Invite. Press `I` twice to confirm, since it can't be undone |
| `Esc` | Close the panel |

```
harbinger
harb
harbinger appease
harbinger brace
harbinger invite
```

See [The Harbinger](harbinger.md) for the roster, false prophets and verdicts.

---

## Save / System

| Command | Description |
|---|---|
| `save` | Open an **Overwrite / Branch** prompt for your current run |
| `save <name>` | **Branch** a new save with that name off your current run (autosave then follows it) |
| `load` | Open the **Load Game** browser (your save tree) to pick which save/branch to load |
| `load <name>` | Load that save directly |
| `saves` | List all save files |
| `save list` | Same as `saves` |
| `Esc` | Quick-save to your active save |
| `account` | Show the active account's short ID and **recovery code** (restores identity, not progress) |
| `account list` | List the local accounts on this machine, marking the active one |
| `account switch <name>` | Switch to a local account by its name (changes which account's saves you see) |
| `account recover <code>` | Restore your identity from a recovery code on a new machine/reinstall |
| `account export [path]` | Write a signed, **ID-bound** account backup (unlocks, stats, achievements, prefs). Default `account-<id8>-export.json` inside that account's slot, or a path you give |
| `account import <path> [replace]` | Bring a backup into **its own account slot** (keyed by the embedded ID — creates it or **merges**; add `replace` to overwrite that account wholesale). Switches to it |
| `account backup` | Full snapshot of the active account (`account.json` + saves) saved to `data/backups/<name>-<id8>-<timestamp>/`. The Accounts panel's `b` action does the same for the highlighted account |
| `account wipe` | Points you to the **Accounts** panel's **Wipe Account** action — the actual (permanent) wipe lives there behind a type-the-name confirm, not this command |
| `theme` | Open the **Themes** picker — browse palettes with live preview (also on the main menu) |
| `theme list` | List every theme by name and key, marking the active one and noting each theme's light/dark variant and which are accessible |
| `theme <key>` | Switch directly to a theme by key (e.g. `theme high_contrast`) |
| `dump` | Export logs to a file for debugging, in `logs/` inside your active account's folder (`data/accounts/<id>/logs/`) — the one player-reachable place that still prints raw tick counts, alongside the wall-clock reading |
| `help` | Open the Help panel — full command reference and list of available panels |

Save files live under your **active account's** slot — `data/accounts/<id>/saves/*.json`, relative to the directory you launch the game from (saves are per-account). The `save <name>` and `load <name>` commands above work with the same files as the **Load Game** browser below. See [Saving & Loading](saving-and-loading.md) and [Account & Recovery](account.md).

`account` (no arguments) prints your account's short ID and its **recovery code** — a short `AGEF-…` string that restores your **identity** (your account ID) across machines and reinstalls. The code restores **identity only, not earned progress** (theme unlocks and lifetime stats are separate — back those up with `account export`). Write the code down to keep your identity; it is a convenience identifier, not a password. To restore on another machine, run `account recover <code>`. If the local account already has unlocked progress, recovery asks you to confirm with `account recover <code> confirm` first, since recovering replaces the local identity and the code does not carry your unlocks.

The game keeps **multiple local accounts**, one active at a time; the **Accounts** entry on the main menu lists them and is where you switch between them, create new ones, and back them up. `account list` prints the same list from the prompt, and `account switch <name>` makes a different account active — changing which account's saves you see.

Each account's **progress** (theme unlocks, lifetime stats, achievements, prefs) is backed up separately from the recovery code. `account export` writes a signed backup that is **bound to its account ID** (default `account-<id8>-export.json` inside that account's slot, or a path you give it); `account import <path>` brings one back. Import is keyed by the **account ID embedded in the backup** and always lands in **that account's own slot** — it **creates** that account if it doesn't exist locally, or **merges** into it if it does (unioning unlocks and achievements and taking the higher of each lifetime stat, so re-importing an old backup never drops something you've earned since), and then switches to it. Add `replace` to overwrite that account wholesale. Because it's keyed by the embedded ID, importing **can't clobber a different account** — at worst it updates the one the backup belongs to. A missing or tampered file is rejected and your accounts are left unchanged. With no server, progress recovery only works if you exported it first. See [Account & Recovery](account.md) for the full model.

Separately from the export blob, a **backup** is a full on-disk snapshot of an account's slot — its `account.json` **plus a recursive copy of its `saves/`** — written to `data/backups/<name>-<id8>-<timestamp>/`. `account backup` snapshots the active account (the Accounts panel's `b` does the same for the highlighted one), and **wiping or exporting an account auto-creates a full backup first**, so a wipe always leaves a recoverable copy. Only the **last 10 backups per account** are kept; older ones are pruned automatically. To restore, copy a backup folder's `account.json` and `saves/` back into `data/accounts/<id>/`. See [Backups](account.md#backups).

**Wiping an account** is permanent and deletes that account's identity, theme unlocks, lifetime stats, and achievements — it does **not** touch your game saves. Because it's irreversible, it lives in the **Accounts** panel (`w` on the highlighted account) behind a type-the-account-name confirm, not as a plain command; typing `account wipe` just points you there. See [Account & Recovery](account.md#wiping-an-account).

A bare `save` (no name) opens a prompt: **Overwrite** writes your current run to its **active** save right now, while **Branch new** forks a fresh save (suggested name, editable) whose parent is your current save and then moves autosave onto the new branch — leaving the old save frozen at the branch point. `save <name>` branches straight to that name. The active save is the one you most recently named or loaded (a new game has you name it up front); the periodic autosave and `Esc` continuously overwrite it, so your current game is always kept current on disk. See [Saving & Loading](saving-and-loading.md) for the full model.

A bare `load` (no name) opens the **Load Game** browser so you can pick which save/branch to load from your save tree — it never assumes a slot. You can open it mid-game; `Esc` returns you to your current run without loading anything. `load <name>` skips the browser and loads that save directly.

See [Saving & Loading](saving-and-loading.md) for the full save system.

---

## Themes

AgeForge's interface colors are driven by a set of swappable themes. Open the picker with the **Themes** entry on the main menu, or type `theme` from the in-game prompt.

| Command | Description |
|---|---|
| `theme` | Open the **Themes** picker — browse palettes with live preview (`↑`/`↓` previews, `Enter` keeps, `Esc`/`q` reverts) |
| `theme list` | List every theme by name and key, marking the active one, each theme's light/dark variant, and the lock status of any theme you haven't unlocked |
| `theme <key>` | Switch directly to a theme by key (e.g. `theme high_contrast`) |

```
theme
theme list
theme high_contrast
```

Your theme choice **persists per account** (saved in `account.json`, not in any game save), so it carries across every save and new game. There are 11 themes, dark and light — two **Standard** themes (the default dark **Forge** and the light **Daylight**), four always-unlocked **accessibility** themes (colorblind-safe, plus high-contrast in dark and light), and five **flavor** themes you unlock by reaching later ages. Every theme paints its own background, so it looks right whatever your terminal's colors.

See [Themes & Accessibility](themes.md) for the full list and unlock conditions.

---

## Saving & Loading

### The Load Game browser

From the main menu, choosing **Load Game** opens a save browser that lists every save belonging to your **active account** — under `data/accounts/<id>/saves/` (most-recent first). Load Game is always available — if you have no saves yet, the browser shows a "No saved games found — start a new game" message instead of an empty list.

Highlighting a save updates a **detail pane** on the side with everything you need to size up that save before loading it: its earned title, age and epoch, civilization scale (population, buildings, wonders, milestones, techs, soldiers), prestige level and points, [morale](morale.md), a ⚠ warning if a catastrophe is pending, and the exact save time.

**Keys inside the browser:**

| Key | Action |
|---|---|
| `↑` / `↓` | Move the highlight between saves |
| `Enter` | Load the highlighted save |
| `d` | Delete the highlighted save (asks you to confirm first) |
| `r` | Rename the highlighted save (type a new name; names that collide with an existing save or contain path characters are rejected) |
| `c` | Duplicate the highlighted save (creates `<name>-copy`) |
| `Esc` | Return to the main menu |

**Row tags:** these symbols are also explained on-screen in a bordered **Key** box between the save list and the detail pane, so you don't have to leave the browser to look them up.

| Tag | Meaning |
|---|---|
| ★ auto | The autosave slot |
| ⚠ modified | The save file was edited outside the game (cheater badge) |
| ⚠ corrupt | The file could not be read. It is still listed but dimmed, and cannot be loaded |

---

## Tab shortcuts

Type a single letter to jump straight to a tab:

| Key | Tab |
|---|---|
| `e` | Economy tab |
| `r` | Research overlay (`research`) |
| `m` | City Map overlay |
| `t` | Trade overlay (`trade`) |
| `s` | Stats overlay (`stats`) |
| `w` | Wonders overlay (`wonders`) |
| `l` | Logs overlay (`logs`) |
| `history` | Civilization History overlay |
| `workers` | Worker status overlay (summary / slot utilization / domain breakdown) |
