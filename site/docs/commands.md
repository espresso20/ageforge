# Commands

All commands are typed at the `>` prompt at the bottom of the screen. Press `↑`/`↓` to navigate history.

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

---

## Trade

| Command | Description |
|---|---|
| `trade start <key>` | Activate a trade route |
| `trade stop <key>` | Cancel an active trade route |
| `diplomacy gift <faction>` | Send a gift to improve faction opinion |

```
trade start coastal_market
trade stop coastal_market
diplomacy gift forest_clan
```

Active trade routes run for a fixed duration. Check **Trade** overlay (`trade`) for rates and faction standings.

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

Wonders are shown in **Wonders** overlay (`wonders`) with progress bars for each required resource. Each completed wonder now displays a colour sprite thumbnail next to its name in the Wonders overlay. Built wonders also appear as a row of sprites along the bottom strip of the City Map (see [City Map](#city-map) below).

---

## City Map

| Command | Description |
|---|---|
| `map` | Open the City Map overlay — shows your civilization rendered as pixel art on an age-appropriate satellite background |

```
map
```

The City Map renders your civilization as pixel art buildings packed organically around a central palace, displayed over an AI-generated age-themed satellite background (Civilization 1 style, viewed from 80,000 feet).

- **Palace** — sits at the centre; the sprite changes with each of the 22 ages (roundhouse → pyramid → cathedral → megaplex, etc.)
- **Buildings** — fill the map organically; packing density increases as you advance through later ages
- **Wonders** — completed wonders appear as a sprite row along the bottom edge of the map
- **Background** — each age has a distinct satellite-style backdrop in a Civilization 1 pixel art aesthetic

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
| `account` | Show your account's short ID and **recovery code** (restores identity, not progress) |
| `account recover <code>` | Restore your identity from a recovery code on a new machine/reinstall |
| `account export [path]` | Write a signed **progress** backup (unlocks, stats, achievements) to a file (default `data/account-export.json`) |
| `account import <path> [replace]` | Restore a progress backup file — **merges** by default; add `replace` to overwrite wholesale |
| `account wipe` | Points you to the main-menu **Wipe Account** action — the actual (permanent) wipe lives behind a type-your-name confirm, not this command |
| `theme` | Open the **Themes** picker — browse palettes with live preview (also on the main menu) |
| `theme list` | List every theme by name and key, marking the active one and noting which are accessible |
| `theme <key>` | Switch directly to a theme by key (e.g. `theme high_contrast`) |
| `dump` | Export logs to a file for debugging |
| `help` | Open the Help panel — full command reference and list of available panels |

Save files live in `./data/saves/*.json`, relative to the directory you launch the game from. The `save <name>` and `load <name>` commands above work with the same files as the **Load Game** browser below.

`account` (no arguments) prints your account's short ID and its **recovery code** — a short `AGEF-…` string that restores your **identity** (your account ID) across machines and reinstalls. The code restores **identity only, not earned progress** (theme unlocks and lifetime stats are separate — back those up with `account export`). Write the code down to keep your identity; it is a convenience identifier, not a password. To restore on another machine, run `account recover <code>`. If the local account already has unlocked progress, recovery asks you to confirm with `account recover <code> confirm` first, since recovering replaces the local identity and the code does not carry your unlocks.

Your **progress** (theme unlocks, lifetime stats, achievements, prefs) is backed up separately from the recovery code. `account export` writes a signed `account-export.json` (or a path you give it); `account import <path>` restores it. Import **merges** by default — unioning unlocks and achievements and taking the higher of each lifetime stat, so re-importing an old backup never drops something you've earned since — or add `replace` to overwrite your progress wholesale. A missing or tampered file is rejected and your account is left unchanged. With no server, progress recovery only works if you exported it first. See [Account & Recovery](account.md) for the full model.

**Wiping your account** is permanent and deletes your identity, theme unlocks, lifetime stats, and achievements — it does **not** touch your game saves. Because it's irreversible, it lives on the **main menu** (Esc to reach it) behind a type-your-account-name confirm, not as a plain command; typing `account wipe` just points you there. After a wipe you start over by naming a fresh account. See [Account & Recovery](account.md#wiping-your-account).

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

From the main menu, choosing **Load Game** opens a save browser that lists every save in `./data/saves/` (most-recent first). Load Game is always available — if you have no saves yet, the browser shows a "No saved games found — start a new game" message instead of an empty list.

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
