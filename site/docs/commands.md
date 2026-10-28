# Commands

You run the game by typing commands at the prompt at the bottom of the screen. `↑` and `↓` step through your last 50 commands.

Counts (`recruit 5`, `build farm 3`, `sell hut 2`) are whole numbers from 1 to 1,000,000. Amounts (`gather`, `wonder collect`) are positive numbers. Anything else is refused with the command's usage line. The game does not guess. A command it does not know gets a "Did you mean" suggestion when a known command is close.

Command names and the game keys they take (buildings, techs, resources, civilizations, themes, expeditions and trade routes) are not case-sensitive: `sell Hut` is `sell hut`, and `research Tool Making` is `research tool_making`.

This page lists every command. Each section links to the page that explains the system behind it.

<figure class="screen" data-screen="help"><figcaption>The Help panel lists the same commands and keys inside the game.</figcaption></figure>

## The prompt

As you type, the best completion of the line shows in dim text after the cursor: type `adv` and `ance` appears after it. Completions come from your game, not a fixed list.

- `build` offers the buildings you can build in this age, the ones you can afford first.
- `research` offers the techs you can start.
- `plan build` offers the same buildings, then the next age's.
- `plan research` offers the techs you have not researched or planned yet, the ones you can start first.
- `upgrade` and `plan upgrade` offer the buildings that have an upgrade.
- `assign` offers your built buildings that take workers, the ones with a free slot first.
- `unassign` and `dismiss` offer buildings with workers in them.
- `roster` offers the worker domains, the ones you have worker buildings in first, and `auto`.
- `sell` offers the buildings you have that it takes (not wonders, storage or knowledge buildings, and nothing in the Primitive Age).
- `wonder collect` offers the resources the wonder still needs.
- `trade route start` and `trade route stop` offer the routes you can start and the routes running.
- `diplomacy` offers the civilizations you have met. `diplomacy accept` and `plan deal` offer the open deal numbers of the civilization you typed.
- `theme` offers the themes you have unlocked.
- `load` offers your saves, and `account switch` your local accounts.

| Key | What it does |
|---|---|
| `Tab` | Take the completion. Press it again for the next one (`Shift+Tab` goes back). When more can follow (`plan build `), a space comes with it. |
| `→` | Take the completion, when the cursor is at the end of the line. |
| `Enter` | If the line is already a whole command, run it exactly as typed. If not, and the completion makes it one, run the completion (`adv` runs `advance`). Otherwise run the line as typed, and the game says what is wrong with it. |
| `↑` / `↓` | Step back and forward through the last 50 commands you ran. |

Commands that can't be undone are never run from a completion. `Enter` on `plan cle` fills in `plan clear` and waits, and a second `Enter` runs it. These are `sell`, `dismiss`, `research cancel`, `plan clear`, `load <name>`, `prestige confirm yes`, `festival confirm yes`, `harbinger invite`, `diplomacy raid`, `quit`, and `account switch`, `import`, `recover` and `wipe`. Typed in full, they run on the first `Enter` like anything else.

Most panels leave the prompt working while they are open, and take only the keys that print nothing. The Wonders, Harbinger and Plan panels use letter keys, so press `Esc` to close one before you type a command. In a panel that is longer than the screen, `↑`, `↓`, `PgUp`, `PgDn`, `Home` and `End` scroll it while the prompt is empty.

With no panel open, `PgUp` and `PgDn` turn the Buildings list a page of whole entries (the box says how many are above and below). `Ctrl+R` turns the Resources box to its next page when it has more resources than it can show. `Esc` saves your game to its own save and returns to the main menu. The Help panel lists every key.

A keyboard with no Page Up, Page Down, Home or End (a Mac's, for one) is covered too. `Option+↑` and `Option+↓` work as `PgUp` and `PgDn`, and `Option+←` and `Option+→` as `Home` and `End`, on the dashboard and in every panel. Other keyboards can use `Alt` for the same chords. Some terminals, Apple Terminal by default, send Option+← and Option+→ as `Esc b` and `Esc f`. The game reads those as `Home` and `End` while the prompt is empty, and with something typed they jump a word, as the prompt always does. If your terminal does not pass Option+arrows through, set it to send Option as Meta (`Esc+`). In the Map, `-` and `=` zoom with no modifier at all.

## The log

The log in the main window says what each command did and what happened in the game: a build started, queued or finished, a gather, a sale, workers recruited or assigned, a share of the roster set, what the roster routine does (`Roster: recruited 3 workers (population 12/20), put 2 idle workers to work.`), a trade, a research started, a plan item added, and events, milestones, age advances, harbingers, warnings and errors. Each line starts with what happened, with no tick number. Routine confirmations are in the plain text color, so events, warnings and errors stand out.

The Logs panel (`logs`) shows the last 50 of the same lines with the tick each one happened on and a mark for its kind: `[*]` an event, `[+]` a success, `[i]` a note, `[!]` a warning, `[X]` an error, and `·` a routine confirmation. Under the lines it shows the tick, age, population, food use and game speed, then your active events, build queue, research and worker assignments.

---

## Timers and durations

The game shows every duration and countdown as an approximate wall-clock time, not a tick count: `~38s`, `~4m 44s`, `~1h 12m`. A length that can vary shows as a range from the shortest to the longest. A reading has two units of precision at most.

The `~` matters. A tick is 2 seconds of real time, and tick speed bonuses (from techs, milestones and boons) make it shorter. Every reading is worked out from your current tick rate and changes when that rate does: finish a tech that grants tick speed and the countdown you were watching gets shorter. There is no speed setting. Only those bonuses make ticks come faster.

Balance values are defined in ticks (a tech costs so many ticks of research, an event lasts so many ticks), and this Compendium quotes those tick figures where the tick count is the mechanic. The `dump` export keeps the raw tick counts.

---

## Building

| Command | Description |
|---|---|
| `build <building> [count\|max]` | Start building copies of a building (default 1). A count builds as many as you can pay for, up to that number. `max` builds as many as you can pay for, up to the building's limit. A building from an earlier age is refused (see `upgrade`), and one that waits for a tech names it. The age's wonder is the exception: it is paid through its bank. Bare `build` lists the buildings you have unlocked, with how many you have, what each costs and whether you can afford it. |
| `sell <building> [count]` | Demolish copies of a building and get back 50% of what they cost. Their workers go back to idle. Wonders, storage and knowledge buildings can't be sold, nothing can be sold in the Primitive Age, and a building with a copy under construction can't be sold until it finishes. |
| `upgrade` | List the older buildings you hold: copies, the building each becomes, the price of the next copy, how many you can pay for now, and what blocks the next one. |
| `upgrade <building> [count\|all]` | Turn copies of an older building into its line's building for your age, one copy at a time, each at half the price of a new copy of that building. With a count it upgrades that many. With no count, or with `all`, it goes through every copy you hold. It stops at the first copy it cannot pay for or at the new building's max count, and says how many it upgraded and why it stopped. A building that waits for a tech refuses and names it. The copy keeps its workers. Storage and knowledge buildings are never upgraded. |
| `advance` | Advance to the next age once its requirements are met. The age never advances by itself (see `plan advance` to queue it). When the requirements are not met it says so and points at `charter`. |
| `charter` | Open the Advance panel: the four pillars the next age asks for, the march of ages with a calendar year, and what the next age brings. See [The Advance Panel](advance.md). |
| `gather <food\|wood\|stone> [amount]` | Gather food, wood or stone by hand: 3 by default, at most 25 per use. With Tool Making researched every gather brings 2 more. Gathering by hand ends after the Medieval Age. |
| `buildings` | Open the Buildings panel: every building of the ages you have reached, grouped by age, with how many you have built and, for a building that waits for a tech, the research it needs. |

In the Advance panel, `↑` and `↓` pick a pillar, `Enter` advances once all four stand (the same as `advance`) and `Esc` closes it. The prompt keeps working while it is open.

```
build hut
build lumber_mill
build hut max
sell lumber_mill
sell gathering_camp 3
upgrade
upgrade gathering_camp
upgrade gathering_camp 3
gather wood 5
```

See [Buildings](buildings.md) and [Ages](ages.md).

---

## Build plan

The plan is a list of builds and techs that the game starts for you, in order, as the resources come in. It runs while you play and while you are away (offline catch-up runs it too). Each item is paid for when it starts, not when you add it. An item that cannot start yet does not hold up the ones after it, but it keeps its next price back from them. The plan holds up to 60 items, not counting research. Full rules are on the [Build Plan](plan.md) page. After the first prestige, [Plan Template](prestige.md#plan-template) adds each age's part of the plan you wrote when you enter that age.

| Command | Description |
|---|---|
| `plan` | Open the Plan panel: the items, what each costs next and whether it can start. |
| `plan build <building> [count]` | Add copies of a building of this age or the next (default 1). The next age's buildings wait for the advance. Adding more of the building at the end of the plan adds to that item. |
| `plan upgrade <building> [count\|all]` | Plan upgrades for copies you hold (default 1). Each copy upgrades as soon as its price can be paid. |
| `plan research <tech>` (or `plan res`) | Add a tech and, before it, the techs it still needs (see [The research queue](plan.md#the-research-queue)). Techs start one at a time, in plan order. It takes any tech you can see on your tech tree: a tech of a later age waits for its age. |
| `plan advance` | Advance to the next age as soon as its requirements are met. The next age's buildings and techs can be planned too. They wait for the advance. |
| `plan deal <civ> <n>` | Take a civilization's trade deal `n` as soon as you can pay its price. While it waits it holds its price back from the items below, and it drops out if the offer rotates away. See [Factions & Diplomacy](factions.md). |
| `plan list` | Print the plan with each item's status. |
| `plan remove <n>` (or `plan rm`) | Remove item `n`. |
| `plan optimize` (or `plan reorder`) | Reorder the planned buildings for the fastest growth, mixing homes and producers, and say how much sooner the plan finishes. It only keeps an order that is faster. See [Optimizing the order](plan.md#optimizing-the-order). |
| `plan up <n>` | Move item `n` one place up. |
| `plan down <n>` | Move item `n` one place down. |
| `plan clear` | Empty the plan. |

```
plan build hut 10
plan build gathering_camp 5
plan res tool_making
plan advance
plan deal merchant_guild 2
plan up 3
plan rm 4
plan optimize
```

In the Plan panel:

| Key | Action |
|---|---|
| `↑` / `↓` (or `K` / `J`) | Pick an item |
| `U` / `D` | Move the picked item up or down |
| `X` (or `Delete`) | Remove the picked item |
| `C` | Clear the plan. Press a capital `C` twice. |
| `Esc` | Close the panel |

---

## Workers

| Command | Description |
|---|---|
| `recruit [count\|max]` | Recruit workers into free housing (default 1). A free worker slot is not needed. New workers start idle, in the reserve. Put them to work with `assign`, or leave them: a minute later any still idle go to work by your roster. |
| `assign <building> [count\|all]` | Assign idle workers to a building (the building sets their domain). |
| `unassign <building> [count\|all]` | Take workers out of a building and back to idle. |
| `dismiss <building> [count\|all]` | Remove workers from a building and from your population for good. |
| `people` | Open the People panel: your homes and every kind of person in them, then tabs for your work areas, your army and your scholars. `workers` opens it too. |
| `workers` | The same as `people`. |
| `roster` | Print each domain's share of your workforce with its workers and slots (`Faith: 40% (set), 8 workers in 10 slots`), and whether auto-recruit is on. |
| `roster <domain> [percent\|auto]` | With a percent from 0 to 100, set that domain's share of your workforce. Decimals and a `%` sign are allowed (`roster lumber 12.5%`), and 0 keeps the domain empty. With `auto`, put the domain back on auto. With neither, print its line. |
| `roster auto` | Put every domain back on auto, the default: workers follow your buildings' worker slots. |
| `workers auto-recruit [on\|off]` (or `workers autorecruit`) | Bare, show whether auto-recruit is on. `on` (the default) lets your population grow toward your housing while food allows, filling empty worker slots at once. `off` leaves recruiting to you. |

`workers share` takes the same arguments as `roster`.

You don't have to recruit or assign workers yourself. With auto-recruit on (the default), your population grows toward your housing while food allows. Empty worker slots fill at once, and past them the population grows by 1% every 5 ticks (at least one worker), with the newcomers waiting in the reserve. Every 5 ticks (about 10 seconds) the game puts idle workers to work by your roster: each domain's percent of your whole workforce, with domains on auto splitting the rest by their worker slots. It never takes a worker out of a building to keep to the shares, except out of a domain set to 0, so your own `assign` and `unassign` stick. After a worker command (recruit, assign, unassign, dismiss, sell or upgrade) it waits a minute before it places anyone. The full rules are in [Roster](workers-and-domains.md#the-roster).

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
roster
roster faith 40
roster lumber 12.5%
roster military 0
roster faith auto
roster auto
workers auto-recruit off
```

The People panel leaves the prompt working. `←` and `→` switch between its Overview, Work, Army and Scholars tabs, `↑` and `↓` pick a row, and `Esc` closes it.

See [People](workers-and-domains.md) for the domain table and the efficiency formula.

---

## Research

| Command | Description |
|---|---|
| `research` | Open the Research panel: the tech tree as a map (see [Reading the Tech Tree](technologies.md#reading-the-tech-tree)). |
| `research tree [close\|far]` | Open the tree zoomed in on big badges (`close`) or out on the whole tree (`far`). |
| `research card <tech>` | Open the tree on a tech's card: what it does, how long it takes and what it needs. It opens only on a tech you can already see. |
| `techs` | Opens the same Research panel. |
| `research list` | List the techs you can research now, with the time each would take your scholars. A ★ marks the keystone its age's wonder needs. Under the list: the tech in progress with its bar, percent and time left, the queue in order, and your scholars. |
| `research <tech>` | Start researching a tech. It costs no resources, only scholars' time. It needs its prerequisites and its age. |
| `research cancel` | Cancel the current research. The work done on it is lost. |
| `scholars` | Show your scholars: the kinds and counts, how many research and how many teach, and what teaching adds to what workers make. |
| `scholars <percent>` | Set the share of your scholars that research, a whole percent from 0 to 100 (100 by default). The rest teach. |

```
research tool_making
research agriculture
research iron_smelting
```

The Research panel (`research`) is a map of the tech tree. Like the Map it leaves the prompt working and takes only the keys that print nothing:

| Key | Action |
|---|---|
| Arrows | Move to the nearest tech that way; the view follows |
| `Tab` / `Shift+Tab` | Next or previous tech you can start |
| `PgUp` / `PgDn` (`Option+↑` / `Option+↓`) | Zoom out to a tech a line, and back in |
| `Home` | Back to your current age |
| `Enter` | Open the selected tech's card; on the card, do what its last line says (start the tech, or add it to the build plan) |
| `Esc` | Close the card, then the panel |

With something typed, `Tab` and `Enter` act on the prompt instead. `research list` prints the keys of the techs you can start. To queue techs, use `plan research`: it adds a tech with what it still needs first. The game never chooses what you research next, and after the first prestige [Plan Template](prestige.md#plan-template) plans the techs again on later runs. See [Technologies](technologies.md) and [Scholars and Research](knowledge.md).

---

## Expeditions and army

| Command | Description |
|---|---|
| `expedition` | Open the Expeditions panel: the Ready Room and the Expedition Report. |
| `expedition list` | Print the mission board. |
| `expedition <name>` | Send the offer with that name. A part of the name will do, and the game asks which when several fit. |
| `campaign` | Same as `expedition`. |
| `campaign list` | Same as `expedition list`. |
| `campaign <name>` | Same as `expedition <name>`. |
| `army` | Open the Army panel: the soldiers at your posts, your defense rating and your campaigns. |

```
expedition
expedition list
expedition copper hills
campaign skirmish
campaign list
army
```

The board has short offers in the Stone Era, long offers added in the Iron Era and grand offers added in the Steel Era. A Veteran army sees the next tier an era early. Each offer shows as a pin on the Ready Room's map with its length, odds, haul and cost. Every offer takes a party of soldiers, and your army decides how many missions are out at once. A grand offer is military or civil. Once built, the Geographic Society sends offers by itself. See [Military](military.md#the-expeditions-panel).

The Expeditions panel leaves the prompt working:

| Key | Action |
|---|---|
| Arrows | Hop from pin to pin on the Ready Room; pick a party on the Report |
| `Tab` / `Shift+Tab` | Switch between the Ready Room and the Expedition Report |
| `Enter` | Ask to send the selected offer, and `Enter` again sends it |
| `Esc` | Take back the question, then close the panel |

---

## Trade

| Command | Description |
|---|---|
| `trade` | Open the Trade routes panel. |
| `trade route` (or `trade route list`) | List your active trade routes and the routes you can start. |
| `trade route start <route>` | Start a trade route. Routes wait for The Wheel, a Bronze Age tech. The Rail Freight route also waits for Railroads, and the Warp Commerce route for Interstellar Trade. |
| `trade route stop <route>` | Stop an active trade route. |
| `blackmarket` (or `bm`) | Show the black market's status: its culture cost, the odds and the wait between runs. It opens in the Colonial Age. |
| `blackmarket <resource>` | Make a smuggling run: spend culture for the same worth in the chosen resource. Five runs in six go through. The sixth time the smugglers vanish with the culture. It waits for the Mercantilism tech. |
| `trade black [resource]` | Same as `blackmarket`. |

```
trade route start local_barter
trade route stop local_barter
blackmarket
blackmarket gold
```

A run costs the larger of 5K culture or a number of hours of your own culture income, up to 200 hours. After a run the smugglers lie low for 624 ticks (about 21 minutes). They deal in gold, iron, coal, steel, oil, electricity, data, crypto, dark matter and quantum flux.

A route whose imports include a resource that a civilization you are at war with or have embargoed specializes in earns nothing until the conflict ends. Harbors raise the income of every active route. See [Trade](trade.md).

---

## Factions and diplomacy

| Command | Description |
|---|---|
| `factions` | Open the Factions panel: boons and setbacks, the Geographic Society, and the civilizations you have met (personality, backstory, strength, opinion, status, bonuses, wars and lent workers). |
| `diplomacy` (or `dip`) | Opens the same Factions panel. |
| `diplomacy ally <civ>` | Ally with a civilization. Costs 500 gold (250 with the Global Village) and needs opinion 50. |
| `diplomacy rival <civ>` | Declare a rivalry. |
| `diplomacy embargo <civ>` | Embargo a civilization. This is a provocation that can start a war. |
| `diplomacy gift <civ>` | Send a gift of gold for +15 opinion, or +22 with Embassies: 200 gold, or 150 with Telecommunications (the Factions panel shows today's price and what it earns). |
| `diplomacy neutral <civ>` | Set a civilization back to neutral. |
| `diplomacy tribute <civ>` | Sue for peace with a civilization you are at war with (you pay gold and culture, scaled to its strength). |
| `diplomacy raid <civ>` | Raid a civilization's trade route (-20 opinion; a provocation that can start a war). |
| `diplomacy deals [civ]` | List a civilization's trade deals, numbered and worded from your side, such as `1. Buy: give 876M coal → get 966K food` (with no civ: every civilization you have met). A Goodwill deal pays `+N opinion`. |
| `diplomacy accept <civ> <n>` | Take trade deal `n` from a civilization: you give its price and get its goods, or opinion for a Goodwill deal. |

Gifts, alliances, rivalries, embargoes and deals wait for Envoys, a Classical Age tech.

```
factions                           # opens the Factions panel
diplomacy gift merchant_guild
diplomacy ally merchant_guild
diplomacy tribute ironhold_clans   # end a war you'd rather not fight
diplomacy deals merchant_guild     # what the Guild offers right now
diplomacy accept merchant_guild 1  # take its first deal
```

You meet civilizations by running scouting expeditions, not by building anything. Once you have met them, a staffed Embassy (Colonial Age) or Grand Embassy (Industrial Age) slowly raises the opinion of every civilization that isn't hostile to you. See [Factions & Diplomacy](factions.md).

---

## Wonders

| Command | Description |
|---|---|
| `wonder` | Show the current wonder's bank: each resource, banked and needed, and its keystone tech with whether it is researched. |
| `wonder collect <resource\|all> [amount\|all\|max]` (or `wonder bank`) | Bank a resource toward the current wonder. With an amount, bank that much (at most what the wonder still needs). With `all`, `max` or no amount, bank as much as it still needs, up to what you have. `wonder collect all` does this for every resource the wonder still needs. |
| `build <wonder>` | Build the wonder once its bank is full and its [keystone tech](wonders.md#the-keystone-tech) is researched. |
| `wonders` | Open the Wonders panel: a progress bar for each resource this age's wonder needs, and the list of wonders with the ones you have built. |

```
wonder collect wood 1000
wonder bank food all       # as much food as it still needs, up to what you have
wonder bank all            # every resource it still needs, as far as you hold it
build great_monolith
```

A deposit says how much went in and the bank's new total. When nothing can go in, the command says why: the wonder is already built, it doesn't need that resource (and which ones it does), that part of the bank is already full, you have none on hand, or you asked for more than you have. `wonder bank all` banks what it can and lists what it skipped.

The Wonders panel uses letter keys, so it keeps the keyboard:

| Key | Action |
|---|---|
| `↑` / `↓`, `PgUp` / `PgDn` | Scroll the list |
| `G` or `Tab` | Open the [gallery](wonders.md#the-gallery) of every wonder you have built. The same keys close it. |
| `←` / `→` | In the gallery, step through the wonders, oldest first (`↑`, `↓`, `PgUp` and `PgDn` step too) |
| `Home` / `End` | In the gallery, the first wonder and the last |
| `Esc` | In the gallery, back to the list. From the list, close the panel. |

---

## Map

The Map draws your empire from your real game state: your buildings and the workers staffing them, and your wonders. It is a glyph world seen from above, with three zooms: the whole town, a close-up and one kind of building's district. An inspect cursor gives you the name and details of what it is on and the whole command to type for it (for example `build hut`). See [The Map](map.md) for all its keys.

| Command | Description |
|---|---|
| `map` (or `citymap`) | Open the Map panel. The prompt keeps working while it is open. |
| `worldmap` | Same as `map`, opened on the whole town (the region zoom). |
| `map flows [on\|off]` | The Map's flows overlay (stocks running dry, understaffed buildings, idle workers): on, off, or bare to switch it. It lasts for the session. |
| `map glyphs [ascii\|unicode\|nerd]` | Bare, show the current glyph set. With a name, switch to it (default unicode). `nerd` needs a Nerd Font in your terminal. Saved per account. |
| `minimap [on\|off]` | Bare, show whether the dashboard's mini map is on. `on` shows it above the Buildings list. `off` hides it so the list gets the whole column (default on). Saved per account. |
| `icons` | A guided check: do you see Nerd Font icons? If yes it sets `map glyphs nerd`. If not it offers to install JetBrains Mono Nerd Font for your user (pinned v3.5.1, checksum-verified, no admin rights) and tells you how to select it in your terminal. |

```
map
map glyphs ascii
map flows on
minimap off
worldmap
icons
```

`map glyphs` and `minimap` are saved per account, like your theme. With no account loaded they last for the session. The mini map is on until you type `minimap off`. On terminals of 105 columns by 34 rows and larger, the dashboard shows a short mini map above the Buildings list. On smaller terminals it hides, and `minimap off` hides it anywhere.

The prompt keeps working while the Map is open: type commands as usual. The Map takes only the keys that print nothing:

| Key | Action |
|---|---|
| Arrows | Move the cursor or scroll |
| `Tab` / `Shift+Tab` | Next or previous building or wonder |
| `PgUp` / `PgDn` (`Option+↑` / `Option+↓`), or `-` / `=` | Zoom out and in a step: the whole town, the close-up, a district |
| `Enter` | Put the command for what the cursor is on into the prompt (not run: press `Enter` again to run it) |
| `Esc` | Close the Map |

While the prompt is empty, `-` zooms out and `=` and `+` zoom in, so a keyboard with no Page keys can zoom. With something typed, `Tab` and `Enter` act on the prompt instead. The Map lists its keys along its bottom edge.

---

## Prestige

| Command | Description |
|---|---|
| `prestige` | Show your prestige level, the points you have earned, your production multiplier (and the one a prestige would bring), Era Mastery, and the carry-overs: Plan Template, Roster and Old Friends. In the Cosmic Era it also shows the Last Passage outlook. |
| `prestige confirm` | Explain the points and production multiplier a prestige would add, what resets, and what persists. In the Cosmic Era this includes the Last Passage outcome. |
| `prestige confirm yes` | Prestige. It requires the Medieval Age. Points stay earned, and the three carry-overs are active from the first prestige. After a prestige, the next one opens only once you pass the furthest age an earlier run reached. |
| `stats` | Open the Stats panel: empire statistics, the ages reached and the time spent in each age of this run, active events, the prestige section (lasting multiplier, carry-overs and [Era Mastery](prestige.md#era-mastery)), research bonuses, resource rates and active multipliers. It also shows the prestige cue when a new run would pass the current age sooner. |

```
prestige confirm yes
```

A prestige pays for every age the run completed, and each era's ages pay three times what the era before paid: 9 points from the Medieval Age, 120 from the Modern Age, 363 for a run through the Digital Age. See [Prestige](prestige.md) for the [points](prestige.md#prestige-points-formula), the [carry-overs](prestige.md#carry-overs), and [The Last Passage](prestige.md#the-last-passage) for what can happen when you prestige from the Cosmic Era.

---

## Festival

| Command | Description |
|---|---|
| `festival` | Show the festival's status: its culture cost, your culture, the boost, and the cooldown. |
| `festival confirm` | Show the cost and effect of a festival held now, and ask you to confirm. |
| `festival confirm yes` | Hold a cultural festival: spend culture for a temporary production boost. Festivals wait for Drama, a Classical Age tech. |

A festival costs the larger of 2K culture or a number of hours of your own culture income, and gives +20% to all production for 390 ticks (about 13 minutes). Festivals have a 780-tick cooldown (about 26 minutes). The hours depend on the age: roughly 8 hours in the Classical Age, rising to 100 hours by the Industrial Age and staying at 100 in every age from the Victorian Age on, so the price stays in step with what you make. Three techs help: Baroque Arts makes the boost last 25% longer (487 ticks, about 16 minutes), Radio brings the next festival back 20% sooner and Social Media makes each one cost 20% less. The boost adds to the same pool as every other all-production bonus (see [The all-production pool](resources.md#the-all-production-pool)). If a limit holds that pool, `festival` says how much of the boost counts.

---

## Milestones and epochs

| Command | Description |
|---|---|
| `milestones` (or `ms`) | Open the Milestones panel: chain progress, earned titles and active speed boosts. |
| `epoch` | Open the Epoch panel: the current epoch and its event, the catastrophe outlook, epoch history, harbingers, legacy bonuses, Era Mastery and the civilization log. |

See [Milestones](milestones.md) and [Epochs](epochs.md).

---

## Civilization history

| Command | Description |
|---|---|
| `history` | Open the Civilization History panel: braille line graphs of key metrics over time. |

The History panel shows 7 graphs: Population, Food rate, Research rate, Faith, Morale, All production and Game speed. It keeps the last 300 samples, one every 10 ticks, so each graph covers about the last 100 minutes. Age advances appear as `│` markers across all graphs. History is saved with your game. See [History](history.md).

---

## Catastrophe

| Command | Description |
|---|---|
| `catastrophe` (or `cat`) | Reopen the Endure or Succumb choice for a pending catastrophe or Last Passage. With nothing pending, show the outlook as you can know it: the doom a harbinger present foretells, or that no harbinger has come and the era is quiet, for now (in the Cosmic Era, the odds of the Last Passage at your next prestige take that line's place). |

A pending catastrophe blocks `advance` and `prestige confirm yes`, and holds up a `plan advance`, until you choose. A pending Last Passage blocks only `prestige confirm yes`. In the choice window, `E` endures, `S` succumbs and `Esc` closes it without choosing. The status bar shows a pending badge until you decide. The harbinger's Invite (below) is the only way to choose a catastrophe yourself. See [Catastrophe](catastrophe.md).

---

## Harbinger

| Command | Description |
|---|---|
| `harbinger` (or `harb`) | Open the Harbinger panel. With no harbinger present it says so, explains that one comes only when doom is on its way, some while before it strikes, and shows the outlook. |
| `harbinger appease` | Buy the next Appease level. It costs faith and culture (whichever your era has unlocked when it begins), and level 2 costs the same as level 1. Each level multiplies the chance the doom strikes by 0.6. Two levels at most. Refused after Invite. |
| `harbinger brace` | Buy the next Brace level. It costs materials of the era, and level 2 costs the same as level 1. Each level softens an Endure. Two levels at most. |
| `harbinger invite` | Guarantee the doom: it still strikes at its fated moment (on the Last Passage's thread, the Last Passage comes at your next prestige). Free, and can't be undone. |

A harbinger comes only when a doom is fated in your era (Iron Era on), some while before it strikes, or before the Industrial Age as a false prophet. In the Stone Era, where nothing can strike, every harbinger is a false prophet and all three answers are refused. In the Cosmic Era a second thread warns of the Last Passage (your next prestige). While the Reality Tear's harbinger is speaking, the actions go to that thread. The prices by epoch are in [Harbinger](harbinger.md#what-it-costs-by-epoch).

The Harbinger panel uses letter keys, so it keeps the keyboard:

| Key | Action |
|---|---|
| `A` | Appease |
| `B` | Brace |
| `I` | Invite. Press `I` twice to confirm, since it can't be undone. |
| `↑` / `↓` | Scroll |
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

## Status and logs

| Command | Description |
|---|---|
| `status` (or `s`) | Print a status summary: age and game time, every unlocked resource with amount and rate, and your population by class with idle counts, food drain and assignments. |
| `rates` | Print where each resource's rate comes from: buildings, the [worker output bonus](workers-and-domains.md#worker-output-bonuses), research, events, trade, bonuses, the [Cosmic Legacy](prestige.md#cosmic-legacy), [Era Mastery](prestige.md#era-mastery), prestige and food drain, plus what teaching adds. When a limit holds a bonus pool below what it has earned it says both: `All production: -140% earned, -90% counts.` (see [The all-production pool](resources.md#the-all-production-pool)). |
| `logs` | Open the Logs panel: recent game log entries with their tick numbers (see [The log](#the-log)). |
| `dump` | Export the logs to a file for debugging, in the `logs/` folder of your active account (`data/accounts/<id>/logs/`). The export has the raw tick counts. |
| `help` | Open the Help panel: the keys, every command and the list of panels. |

---

## Saving and loading

| Command | Description |
|---|---|
| `save` | Open an Overwrite or Branch prompt for your current run. |
| `save <name>` | Branch a new save with that name off your current run (autosave then follows it). A save name is one word. |
| `load` | Open the Load Game browser (your save tree) to pick which save or branch to load. |
| `load <name>` | Load that save directly. |
| `saves` | List all save files. |
| `save list` | Same as `saves`. |
| `quit` | Save your game to its own save and quit. `Ctrl+C` leaves the same way. |

With no panel open, `Esc` saves to your active save, stops the game and returns to the main menu. With a panel open, it closes the panel.

The game autosaves to your active save every 60 seconds and again when `Esc` takes you to the main menu. Saves live in your active account's slot, `data/accounts/<id>/saves/`, inside the `data/` folder next to the `ageforge` binary.

In the Load Game browser:

| Key | Action |
|---|---|
| `↑` / `↓` | Move between saves |
| `Enter` | Load the selected save |
| `d` / `r` / `c` | Delete (asks first), rename or duplicate the selected save |
| `m` | Mark the selected save as your main game (the one Continue on the main menu opens), or unmark it |
| `Esc` | Return to where you opened it from: the main menu, or your current run if you opened it mid-game |

See [Saving & Loading](saving-and-loading.md) for branching, the save tree, row tags and save integrity.

---

## Badges

| Command | Description |
|---|---|
| `badges` | Open the badge case: every badge your account holds or can earn. The prompt keeps working while it is open. |
| `badges <family>` | Open the case on one family's tab, by its name: `ages`, `wonders`, `lineages`, `ladders`, `specials` and the rest. A family you have nothing in sight of yet has no tab to open. |
| `badges all` | Open the case on the All tab, which shows every family. |
| `badges next` | Open the case on the badges you are closest to earning, nearest first. |
| `badges <name>` | Open one badge's detail by its name, or a part of it (`badges hut hoarder`, `badges rock`). Only badges you can already see are found. |
| `title` | List the titles your account holds, and mark the one it wears. |
| `title <name>` | Wear a title you hold. The start of its name is enough. |
| `title default` | Go back to the title your badge score holds. |

```
badges
badges lineages
badges next
badges rock solid
title
title survivor
```

The badge case leaves the prompt working and takes only the keys that print nothing:

| Key | Action |
|---|---|
| Arrows | Move from badge to badge; the grid follows |
| `Tab` / `Shift+Tab` | The next or the previous family's tab |
| `PgUp` / `PgDn` (`Option+↑` / `Option+↓`) | A page up and down |
| `Home` / `End` | The first badge and the last |
| `Enter` | Open the selected badge's detail, and close it |
| `Esc` | Close the detail, then the case |

With something typed, `Tab` and `Enter` act on the prompt instead. The case draws in your theme and in the glyph set you chose with `map glyphs`. A few badges move, and `motion off` holds them still. See [Badges](badges.md) for the badges themselves, their ladders, titles and the themes they give, and [the badge case](account.md#the-badge-case) for what the case shows. The title you wear belongs to your account and needs one to be loaded.

---

## Accounts

| Command | Description |
|---|---|
| `account` | Show the active account's short ID and what a backup is (the recovery code restores identity, not progress). |
| `account show code` | Show the active account's recovery code, and nothing else, in a small window that `Esc` closes. The code is not written to the game log, so it is not in the Logs panel, a save or a dump. |
| `account list` | List the local accounts on this machine, marking the active one. |
| `account badges` | List the active account's badges as text, family by family: its points and title, each family's earned badges and the first four still to earn. The Stats panel has a shorter digest, and `badges` opens the badge case, which shows them all. See [Badges](badges.md). |
| `account switch <name>` | Switch to a local account by its name (changes which account's saves you see). During a game, it saves the game to its own account and returns to the main menu. |
| `account recover <code> [confirm]` | Restore your identity from a recovery code on a new machine or after a reinstall, and switch to it. It lands in its own slot and never overwrites an account. Add `confirm` when the game asks for it (when the account in use holds any progress). |
| `account export [path]` | Write a signed account backup bound to its account ID (unlocks, stats, badges, prefs). The default is `account-<id8>-export.json` inside that account's slot, or a path you give. It also takes a full backup of the account. |
| `account import <path> [replace]` | Bring a backup into its own account slot, found by the ID inside the file. It creates the account or merges into it. Add `replace` to overwrite that account entirely. It never switches accounts: use `account switch <name>` for that. |
| `account backup` | Copy the whole active account (`account.json`, `badges.json` and saves) to `data/backups/<name>-<id8>-<timestamp>/`. The 10 most recent backups per account are kept. The Accounts panel's `b` key does the same for the selected account. |
| `account wipe` | Tells you where to wipe an account: the Accounts panel on the main menu, behind a type-the-name confirm (press `w` on the account). This command wipes nothing. |

The game keeps several local accounts, one active at a time, each with its own saves. The Accounts entry on the main menu manages them. The recovery code carries your identity, not your progress: `account export` backs up progress, and `account backup` copies everything, saves included. Wiping an account deletes it and every save in its slot (after a backup). See [Accounts & Recovery](account.md).

---

## Themes

| Command | Description |
|---|---|
| `theme` | Open the Themes picker to browse palettes with live preview (`↑` and `↓` preview, `Enter` keeps, `Esc` or `q` reverts). It is also on the main menu. |
| `theme list` | List every theme by name and key, marking the active one, with each theme's light or dark variant, which are accessible, and how to unlock the ones you haven't. |
| `theme <key>` | Switch directly to a theme by key (for example `theme high_contrast`). |
| `motion [on\|off]` | Bare, show whether motion is on. `off` holds still everything that moves on its own: the maps, the main menu, the badges that move in the badge case, and a theme's ambient effect. `on` brings it back (default on). Saved per account. |

```
theme
theme list
theme high_contrast
motion off
```

Your theme is saved per account, not in any game save, so it carries across every save and new game. See [Themes & Accessibility](themes.md) for the themes and the [badges](badges.md#themes-from-badges) that unlock them. `motion` is saved per account too. With no account loaded it lasts for the session.

---

## VibeMode

| Command | Description |
|---|---|
| `vibemode [on\|off]` | Bare, show whether VibeMode is on. `on` swaps the game's words for a second script, written in GenZ. `off` brings the plain words back. The words change at once, with no restart (default off). Saved per account. |

```
vibemode on
vibemode off
```

VibeMode changes words and nothing else. Costs, rates, rules, saves and badges are the same with it on or off, and a save made in one script loads in the other. The script covers the names and descriptions of buildings, techs, civilizations, wonders and ages, events, milestones and badges, the names of your people, the log and the refusals, the panels and the help, the map, and the stories the log tells (encounters, expeditions, raids, harbinger warnings and the end of a run). A line that has no VibeMode version reads in plain words. A line already in the log keeps the words it was written in, and new lines come in the script you have on. The setting is saved per account. With no account loaded it lasts for the session.

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
| `achievements` | `badges` |
| `h`, `?` | `help` |
| `acct` | `account` |
| `ms` | `milestones` |
| `citymap`, `worldmap` | `map` |

`plan research`, `plan remove` and `plan optimize` also have short forms, `plan res`, `plan rm` and `plan reorder`, and `wonder bank` is `wonder collect`.
