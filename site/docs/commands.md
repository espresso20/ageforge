# Commands

All commands are typed at the `>` prompt at the bottom of the screen. `↑`/`↓` step through your command history.

Counts (`recruit 5`, `build farm 3`, `sell hut 2`, ...) are whole numbers from 1 to 1,000,000, and amounts (`gather`, `trade`, `wonder collect`) are positive numbers. Anything else is refused with the command's usage line; the game doesn't guess.

Command names and the game keys they take (buildings, techs, resources, civilizations, themes, expeditions, trade routes, prestige upgrades) are not case-sensitive: `sell Hut` is `sell hut`, and `research Tool Making` is `research tool_making`.

## The prompt

As you type, the best completion of the line shows in dim text after the cursor: type `adv` and `ance` appears after it. Completions come from the game, not a fixed list: `build` offers only the buildings you can build in this age, the ones you can afford first; `research` the techs you can start now, affordable first; `plan build` and `plan research` the same, then the next age's; `assign` your built buildings with free worker slots first; `unassign` and `dismiss` buildings with workers in them; `sell` buildings you have; `trade` and `plan trade` what the market buys and sells; `diplomacy` the civilizations you have met, and `diplomacy accept` and `plan deal` a civilization's open deal numbers after it; `theme` the themes you have unlocked; `load` your saves.

| Key | What it does |
|---|---|
| `Tab` | Take the completion. Press it again for the next one (`Shift+Tab` goes back). When more can follow (`plan build `), a space comes with it. |
| `→` | Take the completion, when the cursor is at the end of the line |
| `Enter` | If the line is already a whole command, run it exactly as typed. If not, and the completion makes it one, run the completion (`adv` runs `advance`). Otherwise run the line as typed, and the game says what is wrong with it. |

Commands that can't be undone are never run from a completion: `Enter` on `plan cle` fills in `plan clear` and waits, and a second `Enter` runs it. These are `sell`, `dismiss`, `research cancel`, `plan clear`, `load <name>`, `prestige confirm yes`, `festival confirm yes`, `harbinger invite`, `diplomacy raid`, `quit`, and `account switch`, `import`, `recover` and `wipe`. Typed in full, they run on the first `Enter` like anything else.

## The log

The log in the main window keeps what is worth noticing: events, milestones, finished research, age advances, harbingers, warnings and errors. A routine confirmation of something a panel already shows goes only to the **Logs** panel (`logs`), marked with a `·`: a build started, queued or finished, a sale, an upgrade, workers recruited, assigned, unassigned or dismissed, a gather, a wonder deposit, a trade, a trade route started or stopped, a research started, an expedition sent, a gift, and a plan item added or started. A finished wonder still reaches the main log.

---

## Timers and durations

The game shows every duration and countdown as an approximate wall-clock time, not a tick count: `~38s`, `~4m 44s`, `~1h 12m`. Expedition lengths, which are rolled at each launch, show as a range from the shortest to the longest possible roll. Readings carry two units of precision at most.

The `~` matters. A tick is not a fixed amount of real time: tick speed bonuses shorten it, so every reading is worked out from your **current** tick rate and changes when that rate does. Finish a tech that grants tick speed and the countdown you were watching gets shorter. There is no speed setting: the game runs at 1x, and only those bonuses make ticks come faster.

Balance values are *defined* in ticks (a tech costs so many ticks of research, an event lasts so many ticks), and this wiki quotes those tick figures where the tick count is the mechanic. One tick is 2 seconds at 1x. The `dump` debug export prints the raw tick counts alongside the wall-clock readings.

---

## Building

| Command | Description |
|---|---|
| `build <building> [count\|max]` | Start constructing a building (default 1; `max` builds as many as you can afford, up to the building's limit) |
| `sell <building> [count]` | Demolish a building and get back 50% of its build cost. Its workers go back to idle. |
| `upgrade` | List the buildings you can upgrade now, with the cost of upgrading every copy |
| `upgrade <building> [count\|all]` | Turn copies of a building into the next age's tier of the same line. With no count it upgrades all copies. You pay only the difference in cost: the new copy's cost minus half of what the old copy would sell for. It stops at the new building's max count. Storage buildings are never upgraded; older storage keeps counting. |
| `advance` | Advance to the next age once its requirements are met. The age never advances by itself (see `plan advance` to queue it) |
| `gather <food\|wood\|stone> [amount]` | Gather food, wood or stone by hand: 3 by default, at most 25 per use. Not available after the Medieval Age. |
| `buildings` | Open the **Buildings** panel |

**Example:**
```
build hut
build lumber_mill
build hut max
sell lumber_mill
sell gathering_camp 3
upgrade
upgrade forager_post
upgrade forager_post 3
gather wood 5
```

---

## Build Plan

A list of builds and techs the game starts for you, in order, as the resources come in: while you play and while you are away (offline catch-up runs it too). Each item is paid for when it starts, not when you add it. Full rules on the [Build Plan](plan.md) page.

| Command | Description |
|---|---|
| `plan` | Open the **Plan** panel: the items, what each costs next and whether it can start. `↑`/`↓` select, `U`/`D` move the selected item, `X` removes it, `C` twice clears the plan. |
| `plan build <building> [count]` | Add copies of a building of this age or the next (default 1). The next age's buildings wait for the advance. Adding more of the building at the end of the plan adds to that item. |
| `plan research <tech>` | Add a tech. Techs start one at a time, in plan order; a prerequisite can be planned before it. |
| `plan trade <give> <get> [amount to get]` | Sell `<give>` for `<get>` at the market as it comes in, until that much `<get>` is bought. With no amount, it keeps buying `<get>` until you remove the item, never past your storage cap. Needs a trade building. |
| `plan advance` | Advance to the next age as soon as its requirements are met. The next age's buildings and techs can be planned too; they wait for the advance. |
| `plan deal <civ> <n>` | Take a civilization's trade deal `n` as soon as you can pay its price. While it waits it holds its price back from the items below, and it drops out if the offer rotates away. See [Trade deals](trade.md#trade-deals). |
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
plan deal merchant_guild 2
plan up 3
```

---

## Workers

| Command | Description |
|---|---|
| `recruit [count\|max]` | Recruit workers into free housing. New workers start idle; put them to work with `assign` |
| `assign <building> [count\|all]` | Assign idle workers to a building (the building sets their domain) |
| `unassign <building> [count\|all]` | Take workers out of a building and back to idle |
| `dismiss <building> [count\|all]` | Remove workers from a building and from your population for good |
| `workers` | Open the **Workers** panel (summary, slot use, domain breakdown) |
| `status` (or `s`) | Print a status summary: age and tick, every unlocked resource with amount, storage and rate, and your population by class with idle counts and assignments |
| `rates` | Print where each resource's rate comes from: buildings, workers, research, events, trade, bonuses and food drain |

You recruit workers into free housing and assign them to buildings, where they take that building's domain class (Gatherer, Lumberjack, etc.). `unassign` returns workers to idle; `dismiss` lowers your population.

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
| `research list` | List the techs you can research now, with their knowledge cost |
| `research <tech>` | Start researching a tech |
| `research cancel` | Cancel the current research. Its progress is lost and the knowledge spent is not refunded |

```
research tool_making
research agriculture
research iron_smelting
```

Tech keys are shown in the **Research** panel (`research`): dim gray when locked, a gold circle when available.

---

## Expeditions & Army

| Command | Description |
|---|---|
| `expedition` | Open the **Expeditions** panel for scouting missions (shorthand: `exp`) |
| `expedition list` | List the scouting expeditions available in your current age (shorthand: `exp list`) |
| `expedition <key>` | Send a scouting expedition. It costs resources, never soldiers (e.g. `expedition scout_ruins`; shorthand: `exp <key>`) |
| `army` | Open the **Army** panel: soldier overview and military campaigns |
| `campaign list` | List the military campaigns available in your current age (`campaign` alone does the same) |
| `campaign <key>` | Wage a military campaign. It spends soldiers, plus any resource cost (e.g. `campaign raid_bandits`) |

```
expedition
expedition list
expedition scout_party
expedition scout_ruins
army
campaign list
campaign raid_bandits
```

Timed missions come in two kinds, on two panels. **Scouting** expeditions (`scout_party`, `scout_ruins`, `naval_expedition`) cost only resources, **0 soldiers**, and are available early, before soldiers appear in the Iron Age. You send these with `expedition <key>` from the **Expeditions** panel. **Military campaigns** (everything else) **spend soldiers** at launch, plus any resource cost. You wage these with `campaign <key>` from the **Army** panel. Either way you pay the cost whether the mission succeeds or fails; only the reward differs. One scouting expedition **and** one military campaign can run at the same time, but not two of the same kind.

Keys with underscores can be typed with spaces: `expedition scout ruins` is the same as `expedition scout_ruins`. If you give a campaign key to `expedition`, the game refuses and points you to `campaign <key>`; a scouting key given to `campaign` points you to `expedition <key>`.

From the **Industrial Age** you can build a **Geographic Society**, which sends scouting parties out by itself, no command needed. It uses the same single scouting slot and waits whenever you have a party in the field, so a party you send by hand always goes first. The more Societies you build and staff, the shorter the wait between automatic parties, though it stays slower than sending them yourself. See [Automatic dispatch](military.md#automatic-dispatch-the-geographic-society).

---

## Trade

| Command | Description |
|---|---|
| `trade` | Open the **Trade** panel |
| `trade list` | List the market's exchange rates (build a market first) |
| `trade <give> <get> <amount to give>` | Sell that amount of `<give>` for `<get>` at the market rate (needs a market). `trade wood stone 100` sells 100 wood for stone |
| `trade route` (or `trade route list`) | List your active trade routes and the routes you can start |
| `trade route start <route>` | Start a trade route |
| `trade route stop <route>` | Stop an active trade route |
| `blackmarket` (or `bm`) | Show the black market's status: culture cost, payout odds and cooldown. Opens in the Colonial Age |
| `blackmarket <resource>` | Make a smuggling run: spend culture for a chance at a big haul of the chosen resource |
| `trade black [resource]` | Same as `blackmarket` |
| `factions` | Open the **Factions** panel: active boons and setbacks, Geographic Society status, and the roster of 11 civilizations (personality, backstory, strength, opinion, status, bonuses, wars and lent workers) |
| `diplomacy` (or `dip`) | Opens the same **Factions** panel |
| `diplomacy ally <civ>` | Ally with a civilization. Costs 500 gold and needs opinion 50 |
| `diplomacy rival <civ>` | Declare a rivalry |
| `diplomacy embargo <civ>` | Embargo a civilization. This is a provocation that can start a war |
| `diplomacy gift <civ>` | Send a gift: 200 gold for +15 opinion |
| `diplomacy neutral <civ>` | Set a civilization back to neutral |
| `diplomacy tribute <civ>` | Sue for peace with a civilization you are at war with (you pay gold and culture, scaled to its strength) |
| `diplomacy raid <civ>` | Raid a civilization's trade route (-20 opinion; a provocation that can start a war) |
| `diplomacy deals [civ]` | List a civilization's trade deals, numbered and worded from your side, e.g. `1. Buy: give 876M coal → get 966K food` (no civ: every civilization you have met). A Goodwill deal pays `+N opinion` |
| `diplomacy accept <civ> <n>` | Take trade deal `n` from a civilization: you give its price and get its goods, or opinion for a Goodwill deal. See [Trade deals](trade.md#trade-deals) |

```
trade list
trade wood stone 100
trade route start coastal_market
trade route stop coastal_market
factions                   # opens the Factions panel
diplomacy                  # the same panel
diplomacy gift merchant_guild
diplomacy ally merchant_guild
diplomacy tribute ironhold_clans   # end a war you'd rather not fight
diplomacy deals merchant_guild     # what the Guild offers right now
diplomacy accept merchant_guild 1  # take its first deal
```

Active trade routes run for a fixed duration. A route whose imports include a resource that a civilization you're **at war with or have embargoed** specializes in is **disrupted** (no income) until the conflict ends; see [Trade Disruption](trade.md#trade-disruption-war-amp-embargo). **Harbors** (`harbor` to `logistics_hub`) raise the income of every active route. The **Trade** panel (`trade`) shows rates. The **Factions** panel (`factions`, or `diplomacy` / `dip`) shows active boons and setbacks, Geographic Society status and each civilization's opinion of you; see [The Factions panel](trade.md#the-factions-panel). Bare `diplomacy` opens that panel; add an action (`ally`/`rival`/`embargo`/`gift`/`neutral`) to act on a civilization directly. You meet civilizations by **running scouting expeditions**, not by building anything. Once you have met them, a staffed **Embassy** (Colonial Age) or **Grand Embassy** (Industrial Age) slowly raises the opinion of every civilization that isn't hostile to you; see [Trade & Diplomacy](trade.md#embassy-buildings).

---

## Wonders

| Command | Description |
|---|---|
| `wonder` | Show the current wonder's bank: each resource, banked / needed |
| `wonder collect <res\|all> [amt\|all\|max]` (or `wonder bank …`) | Bank a resource toward the current wonder. With an amount, bank that much (at most what the wonder still needs). With `all`, `max` or no amount, bank as much as it still needs, up to what you have. `wonder collect all` does this for every resource the wonder still needs |
| `wonder overflow` | Show whether overflow is on |
| `wonder overflow on` | Bank what full stores would waste into the current wonder (the default) |
| `wonder overflow off` | Let production over a storage cap be lost instead |
| `build <wonder>` | Build the wonder once its bank is full |

```
wonder collect wood 1000
wonder bank food all       # as much food as it still needs, up to what you have
wonder bank all            # every resource it still needs, as far as your stores go
build great_monolith
```

A deposit says how much went in and the bank's new total. When nothing can go in, the command says why: the wonder is already built, it doesn't need that resource (and which ones it does), that part of the bank is already full, you have none on hand, or you asked for more than you have. `wonder bank all` banks what it can and lists what it skipped. Autocomplete after `wonder bank` offers only the resources the wonder still needs.

**Overflow.** While overflow is on, production that a full store would throw away goes into the current age's wonder bank instead, for every resource the wonder still needs and only up to what it still needs. It never takes from what you hold, works during offline catch-up too, and says so in the log when it finishes a resource's part of the bank. See [Wonders](wonders.md#overflow).

The **Wonders** panel (`wonders`) shows a progress bar for each required resource, and a color sprite thumbnail next to each completed wonder. Completed wonders also appear on the Map as landmarks drawn in their era's look, and the cursor jumps to them with Tab (see [Map](#map) below).

---

## Map

The Map draws your empire from your real game state: your buildings and the workers staffing them, your wonders, the civilizations you have met, your trade routes and wars. It has two styles, both built on the same map model: **roguelike** (the default), a glyph world seen from above with three zooms from the whole known world down to a named district, and **skyline**, your empire side-on as a panorama with one district for every age you have lived through. In both, an inspect cursor gives you the name and details of what it is on and the whole command to type for it (for example `build hut`). See **[The Map](map.md)** for both styles and all their keys.

| Command | Description |
|---|---|
| `map` | Open the Map panel. It covers everything but the command bar, which keeps working while it is open |
| `map style [roguelike\|skyline]` | Bare, show the current style. With a name, switch to it (default roguelike). Saved per account |
| `map flows [on\|off]` | The Map's flows overlay (full stores, understaffed buildings, idle workers): on, off, or bare to switch it. Lasts for the session |
| `style [roguelike\|skyline]` | Same as `map style` |
| `minimap [on\|off]` | Bare, show whether the dashboard's mini map is on. `off` hides it so the Buildings list gets the whole column; `on` brings it back (default on). Saved per account |
| `map glyphs [ascii\|unicode\|nerd]` | Bare, show the current glyph set. With a name, switch to it (default unicode). `nerd` needs a Nerd Font in your terminal; every icon has a Unicode fallback. Saved per account |
| `citymap` | Same as `map` |
| `worldmap` | Same as `map`, opened on the known world (the roguelike style's region zoom) |
| `icons` | A guided check: do you see Nerd Font icons? If yes it sets `map glyphs nerd`; if not it offers to install JetBrains Mono Nerd Font for your user (pinned v3.5.1, checksum-verified, no admin rights) and tells you how to select it in your terminal |

```
map
map style skyline
style roguelike
map glyphs ascii
map flows on
minimap off
worldmap
icons
```

`map style`, `map glyphs` and `minimap` are saved per account, like your theme: they carry across saves and new games. With no account loaded they last for the session. On terminals of about 120x40 and larger the dashboard also shows a short **mini map** of the active style above the Buildings list; on smaller terminals it hides, and `minimap off` hides it anywhere.

The command bar keeps working while the Map is open: type commands as usual. The Map takes only the keys that print nothing, in every style:

| Key | Action |
|---|---|
| Arrows | Move the cursor or scroll |
| `Tab` / `Shift-Tab` | Next or previous thing to inspect |
| `PgUp` / `PgDn` | Zoom out and in (roguelike) or scroll half a screen (skyline) |
| `Enter` | Put the command for what the cursor is on into the prompt (not run: press Enter again to run it) |
| `Esc` | Close the Map |

With something typed, `Tab` and `Enter` act on the prompt instead.

Each style lists its keys along its bottom edge, and [The Map](map.md) has the full tables.

---

## Prestige

| Command | Description |
|---|---|
| `prestige` | View prestige status and available points. In the Cosmic Era it also shows the chance of the Last Passage and which figure is warning of it, and shows the Cosmic Legacy if you hold it |
| `prestige confirm` | Explain what confirming would do. In the Cosmic Era this includes what Endure (share of this run's points) and Succumb (the Cosmic Legacy) would give you if the Last Passage comes |
| `prestige confirm yes` | Prestige now (requires the Modern Age). In the Cosmic Era it first rolls the Last Passage; if it comes, prestige waits for your choice |
| `prestige shop` | List the prestige upgrades |
| `prestige buy <key>` | Buy a prestige upgrade |

```
prestige confirm yes
prestige buy gather_boost
prestige buy tick_speed
```

The **Stats** panel (`stats`) also shows the available upgrades and their costs.

See [The Last Passage](prestige.md#the-last-passage) for what can happen when you prestige from the Cosmic Era.

---

## Festival

| Command | Description |
|---|---|
| `festival` | Show festival status: culture cost, current culture, and the boost it grants |
| `festival confirm yes` | Hold a cultural festival now: spend culture for a temporary production boost |

A festival costs the larger of 2K culture or 5% of your culture storage, and gives **+20% to all production for 150 ticks** (~5 minutes). Festivals have a **300-tick cooldown** (~10 minutes), so they stay an occasional, deliberate boost. Culture also pays for cultural monuments, black-market smuggling runs, harbinger Appease (from the Steel Era) and tribute to end a war; see [Resources](resources.md#culture).

---

## Milestones & Epochs

| Command | Description |
|---|---|
| `milestones` (or `ms`) | Open the **Milestones** panel: chain progress, earned titles and active speed boosts |
| `epoch` | Open the **Epoch** panel |

See [Milestones](milestones.md) and [Epochs](epochs.md).

---

## Civilization History

| Command | Description |
|---|---|
| `history` | Open the Civilization History panel: braille line graphs of key metrics over time |

The History panel shows 7 graphs: **Population**, **Food Rate**, **Knowledge**, **Faith**, **Morale**, **Prod Bonus** and **Tick Speed**. Each covers up to about 100 minutes of history at 1x (300 samples, one every 10 ticks). Age advances appear as `│` markers across all graphs, so you can line up changes with your advances.

The graphs appear once two samples exist (about 40 seconds at 1x). History is saved with your game. See [History](history.md) for details.

---

## Catastrophe

| Command | Description |
|---|---|
| `catastrophe` (or `cat`) | Reopen the Endure / Succumb choice for a pending catastrophe or Last Passage. With nothing pending, show the catastrophe odds for the next epoch transition (in the Cosmic Era, for the Last Passage at your next prestige) |

A pending catastrophe blocks `advance` and `prestige confirm yes` until you choose. A pending Last Passage blocks only `prestige confirm yes`. In the choice window, **E** endures, **S** succumbs and **Esc** closes it without choosing; the status bar shows a pending badge until you decide. There is no Defer option and no command to trigger a catastrophe directly; the harbinger's Invite (below) is the only way to choose one. See [Catastrophe](catastrophe.md).

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
| `load` | Open the **Load Game** browser (your save tree) to pick which save or branch to load |
| `load <name>` | Load that save directly |
| `saves` | List all save files |
| `save list` | Same as `saves` |
| `Esc` | Close the open panel. With no panel open, save to your active save, stop the game and return to the main menu |
| `account` | Show the active account's short ID and **recovery code** (restores identity, not progress) |
| `account list` | List the local accounts on this machine, marking the active one |
| `account switch <name>` | Switch to a local account by its name (changes which account's saves you see) |
| `account recover <code> [confirm]` | Restore your identity from a recovery code on a new machine or after a reinstall. Add `confirm` when the game asks for it |
| `account export [path]` | Write a signed account backup **bound to its account ID** (unlocks, stats, achievements, prefs). Default `account-<id8>-export.json` inside that account's slot, or a path you give |
| `account import <path> [replace]` | Bring a backup into **its own account slot**, found by the ID inside the file. It creates the account or **merges** into it; add `replace` to overwrite that account entirely. Then switches to it |
| `account backup` | Copy the whole active account (`account.json` + saves) to `data/backups/<name>-<id8>-<timestamp>/`. The Accounts panel's `b` key does the same for the highlighted account |
| `account wipe` | Tells you where to wipe an account: the **Accounts** panel's **Wipe Account** action, behind a type-the-name confirm. This command wipes nothing |
| `theme` | Open the **Themes** picker to browse palettes with live preview (also on the main menu) |
| `theme list` | List every theme by name and key, marking the active one and noting each theme's light/dark variant and which are accessible |
| `theme <key>` | Switch directly to a theme by key (e.g. `theme high_contrast`) |
| `quit` | Save your game and quit |
| `logs` | Open the **Logs** panel: recent game log entries, routine confirmations included (see [The log](#the-log)) |
| `dump` | Export logs to a file for debugging, in the `logs/` folder of your active account (`data/accounts/<id>/logs/`). The export prints raw tick counts alongside the wall-clock readings |
| `help` | Open the Help panel: full command reference and list of available panels |

Save files live in your **active account's** slot, `data/accounts/<id>/saves/*.json`, inside the `data/` folder next to the `ageforge` binary (saves are per-account). The `save <name>` and `load <name>` commands above use the same files as the **Load Game** browser below. See [Saving & Loading](saving-and-loading.md) and [Account & Recovery](account.md).

`account` (no arguments) prints your account's short ID and its **recovery code**, a short `AGEF-…` string that restores your **identity** (your account ID) across machines and reinstalls. The code restores **identity only, not earned progress**: theme unlocks and lifetime stats are separate, and you back them up with `account export`. Write the code down to keep your identity; it is an identifier, not a password. To restore on another machine, run `account recover <code>`. If the local account already has unlocked themes, recovery asks you to run `account recover <code> confirm` first, since recovering replaces the local identity and the code does not carry your unlocks.

The game keeps **multiple local accounts**, one active at a time. The **Accounts** entry on the main menu lists them and is where you switch between them, create new ones and back them up. `account list` prints the same list at the prompt, and `account switch <name>` makes a different account active, which changes which saves you see.

Each account's **progress** (theme unlocks, lifetime stats, achievements, prefs) is backed up separately from the recovery code. `account export` writes a signed backup **bound to its account ID** (by default `account-<id8>-export.json` inside that account's slot, or a path you give it); `account import <path>` brings one back. Import always lands in the slot of the **account whose ID is inside the backup**. It **creates** that account if it doesn't exist locally, or **merges** into it if it does, and then switches to it. A merge combines unlocks and achievements and keeps the higher value of each lifetime stat, so importing an old backup never drops something you've earned since. Add `replace` to overwrite that account entirely. Because import goes by the ID inside the file, it **can't overwrite a different account**. A missing or tampered file is refused and your accounts stay unchanged. With no server, you can recover progress only if you exported it first. See [Account & Recovery](account.md) for the full model.

A **backup** is different from an export: it is a full copy of an account's slot on disk, its `account.json` **plus a recursive copy of its `saves/`**, written to `data/backups/<name>-<id8>-<timestamp>/`. `account backup` backs up the active account (the Accounts panel's `b` does the same for the highlighted one). **Wiping or exporting an account also makes a full backup**, so a wipe always leaves a recoverable copy. The game keeps the **last 10 backups per account** and deletes older ones. To restore, copy a backup folder's `account.json` and `saves/` back into `data/accounts/<id>/`. See [Backups](account.md#backups).

**Wiping an account** is permanent and deletes that account's identity, theme unlocks, lifetime stats and achievements. It does **not** touch your game saves. Because it can't be undone, it lives in the **Accounts** panel (`w` on the highlighted account) behind a type-the-account-name confirm, not in a typed command; `account wipe` only points you there. See [Account & Recovery](account.md#wiping-an-account).

A bare `save` (no name) opens a prompt. **Overwrite** writes your current run to its **active** save right now. **Branch new** starts a new save (with a suggested name you can edit) whose parent is your current save, then moves autosave onto the new branch, so the old save stays as it was at the branch point. `save <name>` branches straight to that name. The active save is the one you most recently named or loaded (a new game has you name it up front). Autosave writes to it every 60 seconds, and the game saves it again when `Esc` takes you back to the main menu, so the file on disk is never more than about a minute behind your game. See [Saving & Loading](saving-and-loading.md) for the full model.

A bare `load` (no name) opens the **Load Game** browser so you can pick which save or branch to load from your save tree; it never picks a slot for you. You can open it mid-game; `Esc` returns you to your current run without loading anything. `load <name>` skips the browser and loads that save directly.

See [Saving & Loading](saving-and-loading.md) for the full save system.

---

## Themes

AgeForge's interface colors come from swappable themes. Open the picker with the **Themes** entry on the main menu, or type `theme` at the prompt in game.

| Command | Description |
|---|---|
| `theme` | Open the **Themes** picker to browse palettes with live preview (`↑`/`↓` previews, `Enter` keeps, `Esc`/`q` reverts) |
| `theme list` | List every theme by name and key, marking the active one, each theme's light/dark variant, and the lock status of any theme you haven't unlocked |
| `theme <key>` | Switch directly to a theme by key (e.g. `theme high_contrast`) |

```
theme
theme list
theme high_contrast
```

Your theme choice **is saved per account** (in `account.json`, not in any game save), so it carries across every save and new game. There are 11 themes, dark and light: two **Standard** themes (the default dark **Forge** and the light **Daylight**), four always-unlocked **accessibility** themes (colorblind-safe, plus high-contrast in dark and light), and five **flavor** themes you unlock by reaching later ages. Every theme paints its own background, so it looks right whatever your terminal's colors.

See [Themes & Accessibility](themes.md) for the full list and unlock conditions.

---

## Saving & Loading

### The Load Game browser

From the main menu, **Load Game** opens a save browser that lists every save belonging to your **active account**, under `data/accounts/<id>/saves/` (most recent first). Load Game is always available; if you have no saves yet, the browser says so and suggests starting a new game.

Highlighting a save fills a **detail pane** on the side with what you need to judge that save before loading it: its earned title, age and epoch, the size of the civilization (population, buildings, wonders, milestones, techs, soldiers), prestige level and points, [morale](morale.md), a ⚠ warning if a catastrophe is pending, and the exact save time.

**Keys inside the browser:**

| Key | Action |
|---|---|
| `↑` / `↓` | Move the highlight between saves |
| `Enter` | Load the highlighted save |
| `d` | Delete the highlighted save (asks you to confirm first) |
| `r` | Rename the highlighted save (type a new name; names that match an existing save or contain path characters are refused) |
| `c` | Duplicate the highlighted save (creates `<name>-copy`) |
| `Esc` | Return to the main menu |

**Row tags:** these symbols are also explained on screen in a bordered **Key** box between the save list and the detail pane, so you don't have to leave the browser to look them up.

| Tag | Meaning |
|---|---|
| ★ auto | The autosave slot |
| ● active | The save your game is autosaving into |
| ⚠ modified | The save file was edited outside the game (integrity check failed) |
| ⚠ corrupt | The file could not be read. It is still listed but dimmed, and cannot be loaded |

---

## Command shortcuts

These commands have a shorter name. The short name takes everything the full one does: `b hut 3` is `build hut 3`, `r max` is `recruit max`.

| Shortcut | Command |
|---|---|
| `g` | `gather` |
| `b` | `build` |
| `r` | `recruit` |
| `a` | `assign` |
| `u` | `unassign` |
| `res` | `research` |
| `exp` | `expedition` |
| `t` | `trade` |
| `bm` | `blackmarket` |
| `dip` | `diplomacy` |
| `cat` | `catastrophe` |
| `harb` | `harbinger` |
| `s` | `status` |
| `exportlogs` | `dump` |
| `h`, `?` | `help` |
| `acct` | `account` |
| `ms` | `milestones` |
| `citymap`, `worldmap` | `map` |
