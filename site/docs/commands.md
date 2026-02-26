# Commands

All commands are typed at the `>` prompt at the bottom of the screen. Press `↑`/`↓` to navigate history.

---

## Building

| Command | Description |
|---|---|
| `build <key>` | Start constructing a building |
| `build cancel` | Cancel the current build in the queue |
| `gather <resource> [amount]` | Manually gather food, wood, or stone (max 10) |

**Example:**
```
build hut
build lumber_mill
build cancel
gather wood 5
```

---

## Villagers

| Command | Description |
|---|---|
| `recruit villager` | Recruit a new idle villager (costs food) |
| `assign <type> <resource> <n>` | Assign N villagers of a type to gather a resource |
| `unassign <type> <resource> <n>` | Free up N assigned villagers |

**Villager types:** `worker`, `shaman`, `scholar`, `soldier`, `merchant`
**Assignable resources vary by type** — workers can gather food/wood/stone; scholars gather knowledge; etc.

```
recruit villager
assign worker food 3
assign shaman knowledge 1
unassign worker food 1
```

---

## Research

| Command | Description |
|---|---|
| `research <key>` | Start researching a technology |
| `research cancel` | Cancel active research (progress is lost) |

```
research basic_tools
research agriculture
research smelting
```

Tech keys are shown in the **F2: Research** tab (dim grey when locked, gold circle when available).

---

## Military

| Command | Description |
|---|---|
| `expedition <key>` | Launch a military expedition |
| `speed <1–5>` | Set game tick speed (1=slow, 5=fast) |

```
expedition small_raid
expedition ruins_delve
speed 3
```

Only one expedition can be active at a time. Check **F3: Military** for available expeditions and soldier requirements.

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

Active trade routes run for a fixed duration. Check **F4: Trade** for rates and faction standings.

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

Wonders are shown in **F6: Wonders** with progress bars for each required resource.

---

## Prestige

| Command | Description |
|---|---|
| `prestige go` | Trigger prestige reset (requires Transcendent Age) |
| `prestige buy <key>` | Purchase a prestige upgrade |

```
prestige go
prestige buy production_mastery
prestige buy vault_keeper
```

Available upgrades and costs are shown in **F5: Stats → Prestige** panel.

---

## Save / System

| Command | Description |
|---|---|
| `save <name>` | Save the game with a specific name |
| `Esc` | Quick-save (autosave slot) |
| `export` | Export save as a portable token (copied to clipboard) |
| `import <token>` | Import a save token |
| `wipe` | Delete all saves (requires typing `wipe` twice) |
| `help` | Show a quick command summary |

---

## Tab shortcuts

Type a single letter to jump straight to a tab:

| Key | Tab |
|---|---|
| `e` | Economy (F1) |
| `r` | Research (F2) |
| `m` | Military (F3) |
| `t` | Trade (F4) |
| `s` | Stats (F5) |
| `w` | Wonders (F6) |
| `l` | Logs (F7) |
