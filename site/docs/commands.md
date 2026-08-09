# Commands

All commands are typed at the `>` prompt at the bottom of the screen. Press `↑`/`↓` to navigate history.

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
| `build cancel` | Cancel the current build in the queue |
| `sell <building> [count]` | Demolish a building and recover 50% of its build cost. Workers are returned to idle. |
| `upgrade <building> [count\|all]` | Convert building copies to the next-age tier equivalent, paying only the cost delta (new copy cost minus 50% of the old copy's sell value). Defaults to all copies if no count given. |
| `gather <resource> [amount]` | Manually gather food, wood, or stone (max 25 per command). Disabled from the Renaissance Age onward — works through the Medieval Age only. |

**Example:**
```
build hut
build lumber_mill
build cancel
sell lumber_mill
sell gathering_camp 3
upgrade forager_post
upgrade forager_post 3
upgrade forager_post all
gather wood 5
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
| `trade start <key>` | Activate a trade route |
| `trade stop <key>` | Cancel an active trade route |
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
trade start coastal_market
trade stop coastal_market
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
| `build <wonder_key>` | Build the wonder once its bank is full |

```
wonder collect wood 1000
wonder collect stone 500
build great_monolith
```

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
| `prestige` | View prestige status and available points |
| `prestige confirm yes` | Trigger prestige reset (requires Modern Age) |
| `prestige shop` | View prestige upgrade list |
| `prestige buy <key>` | Purchase a prestige upgrade |

```
prestige confirm yes
prestige buy gather_boost
prestige buy tick_speed
```

Available upgrades and costs are shown in **Stats** overlay (`stats`).

---

## Festival

| Command | Description |
|---|---|
| `festival` | Show festival status — culture cost, current culture, and the buff it grants |
| `festival confirm yes` | Hold a cultural festival now — spends culture for a temporary production boost |

Spend a lump of **culture** (max(2,000, 5% of your culture storage cap)) for **+20% to all production for 150 ticks** (~5 minutes). There's a **300-tick cooldown** (~10 minutes) between festivals, so it stays a rare, deliberate boost rather than a per-tick reflex. This is one of the culture sinks — see [Resources](resources.md#culture). Prestige gates remain the primary long-term culture sink.

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
| `catastrophe invoke` | Trigger a voluntary catastrophe for the current epoch |

Voluntary catastrophes let you force an epoch event outside the normal roll. Use with caution — catastrophes can destroy buildings (Endure) or reset your run (Succumb). See [Epochs](epochs.md).

---

## Save / System

| Command | Description |
|---|---|
| `save` | Open an **Overwrite / Branch** prompt for your current run |
| `save <name>` | **Branch** a new save with that name off your current run (autosave then follows it) |
| `load` | Open the **Load Game** browser (your save tree) to pick which save/branch to load |
| `load <name>` | Load that save directly |
| `saves` | List all save files |
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
| `theme list` | List every theme by name and key, marking the active one and noting which are accessible |
| `theme <key>` | Switch directly to a theme by key (e.g. `theme high_contrast`) |
| `dump` | Export logs to a file for debugging — the one player-reachable place that still prints raw tick counts, alongside the wall-clock reading |
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
| `theme` | Open the **Themes** picker — browse palettes with live preview (`↑`/`↓` previews, `Enter` keeps, `Esc` reverts) |
| `theme list` | List every theme by name and key, marking the active one and the lock status of any theme you haven't unlocked |
| `theme <key>` | Switch directly to a theme by key (e.g. `theme high_contrast`) |

```
theme
theme list
theme high_contrast
```

Your theme choice **persists per account** (saved in `account.json`, not in any game save), so it carries across every save and new game. There are 9 themes — the default **Forge** look, three always-unlocked **accessibility** themes (colorblind-safe + high-contrast), and five **flavor** themes you unlock by reaching later ages.

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
