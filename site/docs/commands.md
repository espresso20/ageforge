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

## Military

| Command | Description |
|---|---|
| `expedition list` | List expeditions available in your current age (shorthand: `exp list`) |
| `expedition <key>` | Launch an expedition — spends its soldier cost (shorthand: `exp <key>`) |
| `speed [multiplier]` | Set game speed (1.0, 1.5, 2.0 … +0.5 per wonder built) |

```
expedition list
expedition scout_party
expedition scout_ruins
speed 1.5
```

Expeditions **spend the `soldiers` resource** at launch (and sometimes a resource cost too) — the cost is deducted whether the run succeeds or fails; only the reward differs. Before the Iron Age, `scout_party` costs 0 soldiers (just food + wood). Only one expedition can be active at a time. Check the **Military** overlay (`army`) for available expeditions and their soldier costs.

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

The History overlay shows 6 live graphs: **Population**, **Food Rate**, **Knowledge Rate**, **Faith**, **Production Bonus**, and **Tick Speed**. Each graph covers up to ~1 hour of rolling history (300 samples, one every 10 ticks). Age advances appear as `│` markers across all graphs so you can correlate events with metric changes.

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
| `save [name]` | Save the game (optional name; defaults to autosave) |
| `load [name]` | Load a saved game |
| `saves` | List all save files |
| `Esc` | Quick-save (autosave slot) |
| `dump` | Export logs to a file for debugging |
| `help` | Open the Help panel — full command reference and list of available panels |

Save files live in `./data/saves/*.json`, relative to the directory you launch the game from. The `save <name>` and `load <name>` commands above work with the same files as the **Load Game** browser below.

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

**Row tags:**

| Tag | Meaning |
|---|---|
| ★ auto | The autosave slot |
| ⭐ elite | An elite save |
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
