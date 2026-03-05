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

## Workers

| Command | Description |
|---|---|
| `recruit <domain>` | Recruit a worker in the given domain (current age's tier) |
| `assign <building_key> [count\|all]` | Assign workers to a building (domain inferred from building) |
| `unassign <building_key> [count\|all]` | Unassign workers from a building |
| `unassign all <domain>` | Unassign all workers in a domain |

**Worker domains:** `food`, `knowledge`, `faith`, `military`, `trade`, `engineering`, `lumber`, `masonry`, `metallurgy`, `energy`, `hacker`, `astronaut`

Workers are recruited by domain. Assignment targets a building directly — the domain is inferred automatically from the building.

```
recruit food 5
recruit faith 3
assign gathering_camp 3
assign library all
unassign shrine
unassign all military
```

See [Workers & Domains (Reference)](workers-and-domains.md) for the full domain table and efficiency formula.

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
| `prestige go` | Trigger prestige reset (requires Modern Age) |
| `prestige buy <key>` | Purchase a prestige upgrade |

```
prestige go
prestige buy production_mastery
prestige buy vault_keeper
```

Available upgrades and costs are shown in **F5: Stats → Prestige** panel.

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
