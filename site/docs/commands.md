# Commands

All commands are typed at the `>` prompt at the bottom of the screen. `↑`/`↓` step through your command history.

Counts (`recruit 5`, `build farm 3`, `sell hut 2`, ...) are whole numbers from 1 to 1,000,000, and amounts (`gather`, `trade`, `wonder collect`) are positive numbers. Anything else is refused with the command's usage line; the game doesn't guess.

Command names and the game keys they take (buildings, techs, resources, civilizations, themes, expeditions, trade routes, legacy kit items) are not case-sensitive: `sell Hut` is `sell hut`, and `research Tool Making` is `research tool_making`.

This page lists every command. Each section links to the page that explains the system behind it.

<figure class="screen" data-screen="help"><figcaption>The Help panel carries the same list inside the game.</figcaption></figure>

## The prompt

As you type, the best completion of the line shows in dim text after the cursor: type `adv` and `ance` appears after it. Completions come from the game, not a fixed list: `build` offers only the buildings you can build in this age, the ones you can afford first; `research` the techs you can start now, affordable first; `plan build` the same, then the next age's; `plan research` every unresearched tech on your tree, the ones you can start first; `assign` your built buildings with free worker slots first; `unassign` and `dismiss` buildings with workers in them; `workers share` the worker domains, the ones you have worker buildings in first, and `auto`; `sell` the buildings you have that it takes (not wonders or storage, and nothing in the Primitive Age); `trade` and `plan trade` what the market buys and sells; `diplomacy` the civilizations you have met, and `diplomacy accept` and `plan deal` a civilization's open deal numbers after it; `theme` the themes you have unlocked; `load` your saves.

| Key | What it does |
|---|---|
| `Tab` | Take the completion. Press it again for the next one (`Shift+Tab` goes back). When more can follow (`plan build `), a space comes with it. |
| `→` | Take the completion, when the cursor is at the end of the line |
| `Enter` | If the line is already a whole command, run it exactly as typed. If not, and the completion makes it one, run the completion (`adv` runs `advance`). Otherwise run the line as typed, and the game says what is wrong with it. |

Two more keys belong to the dashboard, with no panel open: `PgUp` and `PgDn` scroll the Buildings list, and `Ctrl+R` turns the Resources box to its next page when it has more resources than it can show. The Help panel lists every key.

Commands that can't be undone are never run from a completion: `Enter` on `plan cle` fills in `plan clear` and waits, and a second `Enter` runs it. These are `sell`, `dismiss`, `research cancel`, `plan clear`, `load <name>`, `prestige confirm yes`, `festival confirm yes`, `harbinger invite`, `diplomacy raid`, `quit`, and `account switch`, `import`, `recover` and `wipe`. Typed in full, they run on the first `Enter` like anything else.

## The log

The log in the main window says what each command did and what happened in the game: a build started, queued or finished, a gather, a sale, workers recruited or assigned, a worker share set, what the worker shares routine does (`Shares: recruited 3 workers (population 12/20), put 2 idle workers to work.`), a trade, a research started, a plan item added, and events, milestones, age advances, harbingers, warnings and errors. Each line starts with what happened, with no tick number. Routine confirmations are in the plain text color, so events, warnings and errors stand out.

The **Logs** panel (`logs`) shows the same lines with the tick each one happened on and a mark for its kind: `[*]` an event, `[+]` a success, `[i]` a note, `[!]` a warning, `[X]` an error, and `·` a routine confirmation.

---

## Timers and durations

The game shows every duration and countdown as an approximate wall-clock time, not a tick count: `~38s`, `~4m 44s`, `~1h 12m`. Expedition lengths, which are rolled at each launch, show as a range from the shortest to the longest possible roll. Readings carry two units of precision at most.

The `~` matters. A tick is 2 seconds of real time, and tick speed bonuses (from some techs, milestone speed boosts and some boons) make it shorter. Every reading is worked out from your **current** tick rate and changes when that rate does: finish a tech that grants tick speed and the countdown you were watching gets shorter. There is no speed setting; only those bonuses make ticks come faster.

Balance values are *defined* in ticks (a tech costs so many ticks of research, an event lasts so many ticks), and this wiki quotes those tick figures where the tick count is the mechanic. The `dump` debug export prints the raw tick counts alongside the wall-clock readings.

---

## Building

| Command | Description |
|---|---|
| `build <building> [count\|max]` | Start constructing a building (default 1; `max` builds as many as you can afford, up to the building's limit) |
| `sell <building> [count]` | Demolish copies of a building and get back 50% of what they cost; their workers go back to idle. Wonders and storage can't be sold, nothing can be sold in the Primitive Age, and a building with a copy under construction can't be sold until it finishes. A refund your storage can't hold is lost. |
| `upgrade` | List the buildings you can upgrade now, with the cost of upgrading every copy |
| `upgrade <building> [count\|all]` | Turn copies of a building into the next age's tier of the same line. With no count it upgrades all copies. Each copy costs the new building's price minus what the old copy would sell for (half of what it cost), resource by resource and never below zero. It stops at the new building's max count. Storage buildings are never upgraded; older storage keeps counting. |
| `advance` | Advance to the next age once its requirements are met. The age never advances by itself (see `plan advance` to queue it) |
| `gather <food\|wood\|stone> [amount]` | Gather food, wood or stone by hand: 3 by default, at most 25 per use. With Tool Making researched every gather brings 2 more. Not available after the Medieval Age. |
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

See [Buildings](buildings.md) and [Ages](ages.md).

---

## Build Plan

A list of builds and techs the game starts for you, in order, as the resources come in: while you play and while you are away (offline catch-up runs it too). Each item is paid for when it starts, not when you add it. It holds up to 60 items, not counting research, and what a full store would throw away (after the wonder's share) is banked toward its items' next copies. Full rules on the [Build Plan](plan.md) page. With the prestige legacy kit's [Plan Template](prestige.md#plan-template), later runs add each age's part of the plan you wrote when they enter that age.

| Command | Description |
|---|---|
| `plan` | Open the **Plan** panel: the items, what each costs next and whether it can start. `↑`/`↓` select, `U`/`D` move the selected item, `X` removes it, `C` twice clears the plan. |
| `plan build <building> [count]` | Add copies of a building of this age or the next (default 1). The next age's buildings wait for the advance. Adding more of the building at the end of the plan adds to that item. |
| `plan research <tech>` (or `plan res`) | Add a tech and, before it, the techs it still needs (see [The research queue](plan.md#the-research-queue)). Techs start one at a time, in plan order. It takes any tech on your tech tree: a tech of a later age waits for its age |
| `plan trade <give> <get> [amount to get]` | Sell `<give>` for `<get>` at the market as it comes in, until that much `<get>` is bought. With no amount, it keeps buying `<get>` until you remove the item, never past your storage cap. Needs a trade building. |
| `plan advance` | Advance to the next age as soon as its requirements are met. The next age's buildings and techs can be planned too; they wait for the advance. |
| `plan deal <civ> <n>` | Take a civilization's trade deal `n` as soon as you can pay its price. While it waits it holds its price back from the items below, and it drops out if the offer rotates away. See [Factions & Diplomacy](factions.md). |
| `plan list` | Print the plan with each item's status |
| `plan remove <n>` (or `plan rm`) | Remove item `n` (what overflow banked for it goes back to your stores, up to their caps) |
| `plan up <n>` | Move item `n` one place up |
| `plan down <n>` | Move item `n` one place down |
| `plan clear` | Empty the plan |

```
plan build hut 10
plan build gathering_camp 5
plan res tool_making
plan trade gold stone 50000
plan advance
plan deal merchant_guild 2
plan up 3
plan rm 4
```

---

## Workers

| Command | Description |
|---|---|
| `recruit [count\|max]` | Recruit workers into free housing. New workers start idle; put them to work with `assign`, or leave them: a minute later any still idle go to work by your worker shares |
| `assign <building> [count\|all]` | Assign idle workers to a building (the building sets their domain) |
| `unassign <building> [count\|all]` | Take workers out of a building and back to idle |
| `dismiss <building> [count\|all]` | Remove workers from a building and from your population for good |
| `workers` | Open the **Workers** panel (summary, worker shares and what auto-recruit is doing, slot use, domain breakdown) |
| `workers share` | Print each domain's share of your workforce with its workers and slots (`Knowledge: 40% (set), 8 workers in 10 slots`), and whether auto-recruit is on |
| `workers share <domain> [percent\|auto]` | With a percent from 0 to 100, set that domain's share of your workforce; decimals and a `%` sign are allowed (`workers share lumber 12.5%`), and 0 keeps the domain empty. With `auto`, put the domain back on auto. With neither, print its line |
| `workers share auto` | Put every domain back on auto, the default: workers follow your buildings' worker slots |
| `workers auto-recruit [on\|off]` (or `workers autorecruit`) | Bare, show whether auto-recruit is on. `on` (the default) lets the game recruit into empty worker slots while housing and food allow; `off` leaves recruiting to you |

You don't have to recruit or assign workers yourself. With auto-recruit on (the default), the game recruits into empty worker slots while housing and food allow, and every 5 ticks (about 10 seconds) it puts idle workers to work by your **worker shares**: each domain's percent of your whole workforce, with domains on auto splitting the rest by their worker slots. It never takes a worker out of a building to keep to the shares, except out of a domain set to 0, so your own `assign` and `unassign` stick, and after any worker command it waits a minute before it places anyone. The full rules are in [Worker Shares](workers-and-domains.md#worker-shares).

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
workers share
workers share knowledge 40
workers share lumber 12.5%
workers share military 0
workers share knowledge auto
workers share auto
workers auto-recruit off
```

See [Workers](workers-and-domains.md) for the domain table and the efficiency formula.

---

## Research

| Command | Description |
|---|---|
| `research` | Open the **Research** panel: the tech tree as a map (see [Reading the Tech Tree](technologies.md#reading-the-tech-tree)) |
| `research tree [close\|far]` | Open the tree zoomed in on big badges (close) or out on the whole tree (far) |
| `research card <tech>` | Open the tree on a tech's card: what it does, costs and needs |
| `techs` | Opens the same **Research** panel |
| `research list` | List the techs you can research now, with their knowledge cost. A ★ marks the keystone its age's wonder needs |
| `research <tech>` | Start researching a tech |
| `research cancel` | Cancel the current research. Its progress is lost and the knowledge spent is not refunded |

```
research tool_making
research agriculture
research iron_smelting
```

The **Research** panel (`research`) is a map of the tech tree. Like the Map it leaves the command bar working and takes only the keys that print nothing:

| Key | Action |
|---|---|
| Arrows | Move to the nearest tech that way; the view follows |
| `Tab` / `Shift-Tab` | Next or previous tech you can start |
| `PgUp` / `PgDn` | Zoom out to a tech a line, and back in |
| `Home` | Back to your current age |
| `Enter` | Open the selected tech's card; on the card, do what its last line says (start the tech, or add it to the build plan) |
| `Esc` | Close the card, then the panel |

With something typed, `Tab` and `Enter` act on the prompt instead. `research list` prints the keys of the techs you can start. To queue techs, use `plan research`: it adds a tech with what it still needs first. The game never chooses what you research next, and with the prestige legacy kit's [Plan Template](prestige.md#plan-template) the techs you planned are planned again on later runs. See [Technologies](technologies.md) and [Knowledge](knowledge.md).

---

## Expeditions & Army

| Command | Description |
|---|---|
| `expedition` | Open the **Expeditions** panel for scouting missions (shorthand: `exp`) |
| `expedition list` | List the scouting expeditions available in your current age (shorthand: `exp list`) |
| `expedition <key>` | Send a scouting expedition. It costs resources, never soldiers (e.g. `expedition scout_party`; shorthand: `exp <key>`). Past the Scout Party it waits for the Exploration tech |
| `army` | Open the **Army** panel: soldier overview and military campaigns |
| `campaign list` | List the military campaigns available in your current age (`campaign` alone does the same) |
| `campaign <key>` | Wage a military campaign. It spends soldiers, plus any resource cost (e.g. `campaign raid_bandits`). Campaigns wait for the Military Tactics tech |

```
expedition
expedition list
expedition scout_party
expedition scout_ruins
army
campaign list
campaign raid_bandits
```

**Scouting** expeditions (`scout_party`, `scout_ruins`, `naval_expedition`) cost only resources and are available before soldiers appear in the Iron Age; they are how you meet other civilizations. **Military campaigns** (everything else) spend soldiers at launch, plus any resource cost. Either way you pay whether the mission succeeds or fails; only the reward differs. One scouting expedition and one military campaign can run at the same time, but not two of the same kind.

**Commands a tech opens.** Four commands wait for a tech: campaigns (Military Tactics), the Naval Expedition (Navigation), the black market (Mercantilism) and the Rail Freight route (Railroads). Used before the tech, the command is refused and names it: `Campaigns need Military Tactics first. Research it to send one.` See [Commands a Tech Opens](technologies.md#commands-a-tech-opens).

Keys with underscores can be typed with spaces: `expedition scout ruins` is the same as `expedition scout_ruins`. A campaign key given to `expedition` is refused with a pointer to `campaign <key>`, and the other way round. From the Industrial Age a **Geographic Society** sends scouting parties out by itself; a party you send by hand always goes first. See [Military](military.md).

---

## Trade

| Command | Description |
|---|---|
| `trade` | Open the **Trade** panel |
| `trade list` | List the market's exchange rates (build a market first) |
| `trade <give> <get> <amount to give>` | Sell that amount of `<give>` for `<get>` at the market rate (needs a market or a later trade building). `trade wood stone 100` sells 100 wood for stone |
| `trade route` (or `trade route list`) | List your active trade routes and the routes you can start |
| `trade route start <route>` | Start a trade route. Routes wait for The Wheel, a Bronze Age tech |
| `trade route stop <route>` | Stop an active trade route |
| `blackmarket` (or `bm`) | Show the black market's status: culture cost, payout odds and cooldown. Opens in the Colonial Age |
| `blackmarket <resource>` | Make a smuggling run: spend culture for a chance at a big haul of the chosen resource. Waits for the Mercantilism tech |
| `trade black [resource]` | Same as `blackmarket` |

```
trade list
trade wood stone 100
trade route start local_barter
trade route stop local_barter
blackmarket
```

Each trade pushes that pair's rate down a little, and the rate recovers over time. A route whose imports include a resource that a civilization you're at war with or have embargoed specializes in earns nothing until the conflict ends. **Harbors** raise the income of every active route. See [Trade](trade.md).

---

## Factions & Diplomacy

| Command | Description |
|---|---|
| `factions` | Open the **Factions** panel: active boons and setbacks, Geographic Society status, and the roster of civilizations (personality, backstory, strength, opinion, status, bonuses, wars and lent workers) |
| `diplomacy` (or `dip`) | Opens the same **Factions** panel |
| `diplomacy ally <civ>` | Ally with a civilization. Costs 500 gold (250 with the Global Village) and needs opinion 50 |
| `diplomacy rival <civ>` | Declare a rivalry |
| `diplomacy embargo <civ>` | Embargo a civilization. This is a provocation that can start a war |
| `diplomacy gift <civ>` | Send a gift of gold for +15 opinion, or +22 with Embassies: 200 gold, or 150 with Telecommunications (the Factions panel shows today's price and what it earns). Gifts, alliances, rivalries, embargoes and deals wait for Envoys, a Classical Age tech |
| `diplomacy neutral <civ>` | Set a civilization back to neutral |
| `diplomacy tribute <civ>` | Sue for peace with a civilization you are at war with (you pay gold and culture, scaled to its strength) |
| `diplomacy raid <civ>` | Raid a civilization's trade route (-20 opinion; a provocation that can start a war) |
| `diplomacy deals [civ]` | List a civilization's trade deals, numbered and worded from your side, e.g. `1. Buy: give 876M coal → get 966K food` (no civ: every civilization you have met). A Goodwill deal pays `+N opinion` |
| `diplomacy accept <civ> <n>` | Take trade deal `n` from a civilization: you give its price and get its goods, or opinion for a Goodwill deal |

```
factions                           # opens the Factions panel
diplomacy gift merchant_guild
diplomacy ally merchant_guild
diplomacy tribute ironhold_clans   # end a war you'd rather not fight
diplomacy deals merchant_guild     # what the Guild offers right now
diplomacy accept merchant_guild 1  # take its first deal
```

You meet civilizations by running scouting expeditions, not by building anything. Once you have met them, a staffed **Embassy** (Colonial Age) or **Grand Embassy** (Industrial Age) slowly raises the opinion of every civilization that isn't hostile to you. See [Factions & Diplomacy](factions.md).

---

## Wonders

| Command | Description |
|---|---|
| `wonder` | Show the current wonder's bank: each resource, banked / needed, and its keystone tech with whether it is researched |
| `wonder collect <res\|all> [amt\|all\|max]` (or `wonder bank …`) | Bank a resource toward the current wonder. With an amount, bank that much (at most what the wonder still needs). With `all`, `max` or no amount, bank as much as it still needs, up to what you have. `wonder collect all` does this for every resource the wonder still needs |
| `wonder overflow` | Show whether overflow is on |
| `wonder overflow on` | Bank what full stores would waste into the current wonder (the default) |
| `wonder overflow off` | Stop banking overflow into a wonder you have not planned. A wonder in your [build plan](plan.md#overflow-pays-the-plan) still takes overflow; without one, production over a storage cap is lost |
| `build <wonder>` | Build the wonder once its bank is full and its [keystone tech](wonders.md#the-keystone-tech) is researched |
| `wonders` | Open the **Wonders** panel: a progress bar for each required resource, and a sprite of each completed wonder |

```
wonder collect wood 1000
wonder bank food all       # as much food as it still needs, up to what you have
wonder bank all            # every resource it still needs, as far as your stores go
build great_monolith
```

A deposit says how much went in and the bank's new total. When nothing can go in, the command says why: the wonder is already built, it doesn't need that resource (and which ones it does), that part of the bank is already full, you have none on hand, or you asked for more than you have. `wonder bank all` banks what it can and lists what it skipped.

**Overflow.** While overflow is on, production that a full store would throw away goes into the current age's wonder bank instead, for every resource the wonder still needs and only up to what it still needs. It never takes from what you hold, and works during offline catch-up too. What the wonder doesn't need goes toward your build plan's queued copies (see [Build Plan](plan.md#overflow-pays-the-plan)). See [Wonders](wonders.md).

---

## Map

The Map draws your empire from your real game state: your buildings and the workers staffing them, your wonders, the civilizations you have met, your trade routes and wars. It has two styles: **roguelike** (the default), a glyph world seen from above with three zooms from the whole known world down to a named district, and **skyline**, your empire side-on as a panorama with one district for every age you have lived through. In both, an inspect cursor gives you the name and details of what it is on and the whole command to type for it (for example `build hut`). See **[The Map](map.md)** for both styles and all their keys.

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

`map style`, `map glyphs` and `minimap` are saved per account, like your theme. With no account loaded they last for the session. On terminals of 105 columns by 34 rows and larger the dashboard also shows a short **mini map** of the active style above the Buildings list; on smaller terminals it hides, and `minimap off` hides it anywhere.

The command bar keeps working while the Map is open: type commands as usual. The Map takes only the keys that print nothing, in every style:

| Key | Action |
|---|---|
| Arrows | Move the cursor or scroll |
| `Tab` / `Shift-Tab` | Next or previous thing to inspect |
| `PgUp` / `PgDn` | Zoom out and in (roguelike) or scroll half a screen (skyline) |
| `Enter` | Put the command for what the cursor is on into the prompt (not run: press Enter again to run it) |
| `Esc` | Close the Map |

With something typed, `Tab` and `Enter` act on the prompt instead. Each style lists its keys along its bottom edge.

---

## Prestige

| Command | Description |
|---|---|
| `prestige` | View prestige status and available points, how many legacy kit items you own, the points a prestige would pay now and what prestiging from the next age would pay (and how many more), the [Era Mastery](prestige.md#era-mastery) speed of the age you are in and the ages your next prestige would raise. Before the Modern Age it notes that a prestige now is an early taste. In the Cosmic Era it also shows the chance of the Last Passage and which figure is warning of it, and shows the Cosmic Legacy if you hold it |
| `prestige confirm` | Explain what confirming would do: the points you would earn, what resets, and what you keep (prestige points, the legacy kit and what it remembers, and Era Mastery). Before the Modern Age it says that this is an early taste, what it pays and what a full run pays. In the Cosmic Era this includes what Endure (share of this run's points) and Succumb (the Cosmic Legacy) would give you if the Last Passage comes |
| `prestige confirm yes` | Prestige now: a new run, keeping your points, the legacy kit and Era Mastery. It requires the Medieval Age, and before the Modern Age it is an early taste that pays little. In the Cosmic Era an open Reality Tear settles first, then it rolls the Last Passage; if either comes, prestige waits for your choice |
| `prestige shop` | View the legacy kit: each item's price, or "owned", and what the kit remembers from your runs (plan items and the ages they cover, worker shares, civilizations met) |
| `prestige buy <item>` | Buy a legacy kit item: `legacy_plan` (Plan Template, 9 points), `legacy_workers` (Worker Shares, 36) or `legacy_factions` (Old Friends, 54). It works at once, on the run you are in |
| `stats` | Open the **Stats** panel: empire statistics, active events, resource rates, active multipliers, your prestige points, what a prestige pays now and from the next age, your [Era Mastery](prestige.md#era-mastery) in the current age and the legacy kit items you own |

```
prestige confirm yes
prestige buy legacy_plan
prestige buy legacy_workers
```

A prestige pays for every age the run completed, and each era's ages pay three times what the era before paid: 9 points from the Medieval Age, 120 from the Modern Age, 363 for a run through the Digital Age. See [Prestige](prestige.md) for the [points](prestige.md#prestige-points-formula) and [the legacy kit](prestige.md#the-legacy-kit), and [The Last Passage](prestige.md#the-last-passage) for what can happen when you prestige from the Cosmic Era.

---

## Festival

| Command | Description |
|---|---|
| `festival` | Show festival status: culture cost, current culture, and the boost it grants |
| `festival confirm yes` | Hold a cultural festival now: spend culture for a temporary production boost. Festivals wait for Drama, a Classical Age tech |

A festival costs the larger of 2K culture or 5% of your culture storage, and gives **+20% to all production for 390 ticks** (~13 minutes). Festivals have a **780-tick cooldown** (~26 minutes). Three techs help: Baroque Arts makes the boost last 25% longer (487 ticks, about 16 minutes), Radio brings the next festival back 20% sooner and Social Media makes each one cost 20% less. The boost adds to the same pool as every other all-production bonus. It counts in full while that pool is under +200% and a quarter past it (see [The all-production cap](resources.md#the-all-production-cap)). When less than the full +20% would count, `festival` says so before you pay (`Held now, the festival counts a quarter past +200%: +5% now (see stats).`).

---

## Milestones & Epochs

| Command | Description |
|---|---|
| `milestones` (or `ms`) | Open the **Milestones** panel: chain progress, earned titles and active speed boosts |
| `epoch` | Open the **Epoch** panel (epoch events, catastrophe outlook, legacy bonuses and Era Mastery) |

See [Milestones](milestones.md) and [Epochs](epochs.md).

---

## Civilization History

| Command | Description |
|---|---|
| `history` | Open the Civilization History panel: braille line graphs of key metrics over time |

The History panel shows 7 graphs: **Population**, **Food rate**, **Knowledge rate**, **Faith**, **Morale**, **All production** and **Game speed**. It keeps the last 300 samples, one every 10 ticks, so each graph covers about the last 100 minutes. Age advances appear as `│` markers across all graphs. History is saved with your game. See [History](history.md).

---

## Catastrophe

| Command | Description |
|---|---|
| `catastrophe` (or `cat`) | Reopen the Endure / Succumb choice for a pending catastrophe or Last Passage. With nothing pending, show the outlook as you can know it: the doom a harbinger present foretells, or that no harbinger has come and the era is quiet, for now (in the Cosmic Era, the odds of the Last Passage at your next prestige take that line's place) |

A pending catastrophe blocks `advance` and `prestige confirm yes`, and holds up a `plan advance`, until you choose. A pending Last Passage blocks only `prestige confirm yes`. In the choice window, **E** endures, **S** succumbs and **Esc** closes it without choosing; the status bar shows a pending badge until you decide. There is no command to trigger a catastrophe directly; the harbinger's Invite (below) is the only way to choose one. See [Catastrophe](catastrophe.md).

---

## Harbinger

| Command | Description |
|---|---|
| `harbinger` (or `harb`) | Open the **Harbinger** panel. Also listed under Panels in the sidebar. With no harbinger present it says so, explains that one comes only when doom is on its way, some while before it strikes, and shows the outlook |
| `harbinger appease` | Buy the next Appease level: 15% of what a moderate faith economy makes in the age the harbinger arrived in, in faith, and the same for culture from the Steel Era on (level 2 double, or the same again in the Cosmic Era; the Last Passage's thread costs half of what the Interstellar Age makes, a little over three times a doom foretold there; see [Harbinger](harbinger.md#what-it-costs-by-epoch) for the prices). Each level multiplies the chance the doom strikes by 0.6. Two levels at most; refused after Invite |
| `harbinger brace` | Buy the next Brace level: 12% of the most the era asks of each resource you had when it began, except faith and culture (24% for level 2). In the Cosmic Era it is priced on the warning: a third of what the Interstellar Age makes of dark matter and titanium at a moderate economy on the Last Passage's thread (about 22 hours of income), a sixth on the Reality Tear's (about eleven), and level 2 costs the same again. An Endure then destroys 15% / 10% of buildings and keeps 30% / 45% of stored resources (against the Last Passage it keeps 70% / 85% of the run's prestige points instead of 50%). Two levels at most |
| `harbinger invite` | Guarantee the doom: it still strikes at its fated moment (on the Last Passage's thread, the Last Passage comes at your next prestige). Free, and can't be undone |

A harbinger comes only when a doom is fated in your era (Iron Era on), some while before it strikes, or before the Industrial Age as a false prophet. In the Stone Era, where nothing can strike, every harbinger is a false prophet and all three answers are refused. In the Cosmic Era a second thread warns of the Last Passage (your next prestige); while the Reality Tear's harbinger is speaking, the actions go to that thread. The actions work in any age of the thread, cost the same in each, and carry over between figures.

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

## Status & Logs

| Command | Description |
|---|---|
| `status` (or `s`) | Print a status summary: age and game time, every unlocked resource with amount, storage and rate, and your population by class with idle counts, food drain and assignments |
| `rates` | Print where each resource's rate comes from: buildings, the [worker output bonus](workers-and-domains.md#worker-output-bonuses), research, events, trade, bonuses, the [Cosmic Legacy](prestige.md#cosmic-legacy), [Era Mastery](prestige.md#era-mastery) and food drain. When a bonus pool applies less than it has earned it says both: `All production: +320% earned, +230% counts. Past +200% a bonus counts a quarter.` (see [The all-production cap](resources.md#the-all-production-cap)) |
| `logs` | Open the **Logs** panel: recent game log entries with their tick numbers (see [The log](#the-log)) |
| `dump` | Export logs to a file for debugging, in the `logs/` folder of your active account (`data/accounts/<id>/logs/`). The export prints raw tick counts alongside the wall-clock readings |
| `help` | Open the Help panel: full command reference and list of available panels |

---

## Saving & Loading

| Command | Description |
|---|---|
| `save` | Open an **Overwrite / Branch** prompt for your current run |
| `save <name>` | **Branch** a new save with that name off your current run (autosave then follows it) |
| `load` | Open the **Load Game** browser (your save tree) to pick which save or branch to load |
| `load <name>` | Load that save directly |
| `saves` | List all save files |
| `save list` | Same as `saves` |
| `Esc` | Close the open panel. With no panel open, save to your active save, stop the game and return to the main menu |
| `quit` | Save your game and quit |

The game autosaves to your active save every 60 seconds and again when `Esc` takes you to the main menu. Saves live in your active account's slot, `data/accounts/<id>/saves/`, inside the `data/` folder next to the `ageforge` binary.

**Keys in the Load Game browser:**

| Key | Action |
|---|---|
| `↑` / `↓` | Move the highlight between saves |
| `Enter` | Load the highlighted save |
| `d` / `r` / `c` | Delete (asks first), rename or duplicate the highlighted save |
| `Esc` | Return to where you opened it from: the main menu, or your current run if you opened it mid-game |

See [Saving & Loading](saving-and-loading.md) for branching, the save tree, row tags and save integrity.

---

## Accounts

| Command | Description |
|---|---|
| `account` | Show the active account's short ID and **recovery code** (restores identity, not progress) |
| `account list` | List the local accounts on this machine, marking the active one |
| `account badges` | List the active account's badges, earned and locked, with its points. The **Stats** panel shows the same list. See [Badges](account.md#badges) |
| `account switch <name>` | Switch to a local account by its name (changes which account's saves you see). During a game, saves the game to its own account and returns to the main menu |
| `account recover <code> [confirm]` | Restore your identity from a recovery code on a new machine or after a reinstall, and switch to it. It lands in its own slot and never overwrites an account. Add `confirm` when the game asks for it (when the account in use holds any progress) |
| `account export [path]` | Write a signed account backup **bound to its account ID** (unlocks, stats, badges, prefs). Default `account-<id8>-export.json` inside that account's slot, or a path you give |
| `account import <path> [replace]` | Bring a backup into **its own account slot**, found by the ID inside the file. It creates the account or **merges** into it; add `replace` to overwrite that account entirely. It never switches accounts: use `account switch <name>` for that |
| `account backup` | Copy the whole active account (`account.json`, `badges.json` and saves) to `data/backups/<name>-<id8>-<timestamp>/`. The Accounts panel's `b` key does the same for the highlighted account |
| `account wipe` | Tells you where to wipe an account: the **Accounts** panel's **Wipe Account** action, behind a type-the-name confirm. This command wipes nothing |

The game keeps several local accounts, one active at a time, each with its own saves; the **Accounts** entry on the main menu manages them. The recovery code carries your identity, not your progress: `account export` backs up progress, and `account backup` copies everything, saves included. Wiping an account deletes it and every save in its slot (after a backup). See [Account & Recovery](account.md).

---

## Themes

| Command | Description |
|---|---|
| `theme` | Open the **Themes** picker to browse palettes with live preview (`↑`/`↓` previews, `Enter` keeps, `Esc`/`q` reverts). Also on the main menu |
| `theme list` | List every theme by name and key, marking the active one, with each theme's light or dark variant, which are accessible, and how to unlock the ones you haven't |
| `theme <key>` | Switch directly to a theme by key (e.g. `theme high_contrast`) |

```
theme
theme list
theme high_contrast
```

Your theme is saved per account, not in any game save, so it carries across every save and new game. See [Themes & Accessibility](themes.md) for the themes and how to unlock them.

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

`plan research` and `plan remove` also have short forms, `plan res` and `plan rm`, and `wonder bank` is `wonder collect`.
