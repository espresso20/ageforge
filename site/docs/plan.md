# The build plan

The build plan puts the hours between visits to work. It is a list of builds, upgrades, techs, deals with other civilizations and an advance that the game starts for you, in order, as the resources come in. It runs while you play, and it runs while you are away (offline catch-up executes it as the time passes, not in one lump at the end). An item waits only for what it costs against what you hold (see [Big prices and the plan](#big-prices-and-the-plan)). The plan holds up to **60 items**, not counting research.

Open it with `plan`. Add to it with commands:

```
plan build hut 10              # ten huts, one after another along the cost curve
plan build gathering_camp 5
plan upgrade gathering_camp all # upgrade every camp you hold, as each price is ready
plan research tool_making      # techs start one at a time, in plan order
plan research imperial_legions # a tech that needs others: they go in first, in order
plan advance                   # advance as soon as the requirements are met
plan deal merchant_guild 2     # take the Merchant Guild's deal 2 once its price is there
plan build longhouse 15        # a next-age building waits for the advance
plan optimize                  # reorder the buildings for the fastest growth
```

`plan list` prints it. `plan remove <n>` (or `plan rm <n>`), `plan up <n>`, `plan down <n>` and `plan clear` edit it, and the panel does the same with keys. `plan optimize` (or `plan reorder`) reorders the buildings for you (see [Optimizing the order](#optimizing-the-order)). `plan research` can be written `plan res`. See [Commands](commands.md#build-plan).

The plan holds at most 60 items. A full plan refuses a new one until you remove one, but adding more of the building at the end of the plan still adds to that item. Research takes no room: the 60 are for builds, upgrades, deals and the advance, and a research queue is bounded by the tree, each tech once. One build or upgrade item holds at most 1,000 copies. You can plan a building of your current age or of the next age. A wonder, or a building with a limit, is planned only while it has room.

---

## How it runs

- **Paid when it starts, not when you add it.** Adding an item costs nothing. Each copy is paid for at the moment it starts, at the price it has then, through the same checks as `build` and `research` (the age lock, a building's limit, one research at a time).
- **In order, every tick.** After each tick's production the game walks the plan from the top and starts everything it can afford. A build item with a count starts as many copies as the resources cover.
- **Waiting items hold their price.** An item that can't start yet doesn't block the items after it, but it holds back the price of its next copy. A later item only starts if it can be paid from what is left, so a cheap item lower down can never delay one above it, and resources the top items don't need aren't left idle. The order is your priority.
- **Some items hold nothing.** An item that can't start for a reason money won't fix holds nothing back: a resource the current income won't bring in within a day (it needs a producer first), or the next age's building before the advance.
- **A wonder pays its bank from what you hold.** A wonder's price in the plan is what its bank still lacks. Once what you hold (after what the items above it hold back) covers all of that, the plan banks it and starts construction, so you don't need to `wonder collect` for it. While it waits it holds nothing back: it is a big bill, and holding it would stall everything below it. Stock has no limit, so the whole price can build up in your stores. A wonder whose [keystone tech](wonders.md#the-keystone-tech) is not researched yet waits for it (`waits for Stoneworking`) and starts once it is.
- **Techs queue.** Only the first research item in the plan can take the research slot when it frees up, and a tech has no price to hold. See [The research queue](#the-research-queue) for what `plan research` adds and which techs it takes.
- **Deals wait for their price.** A deal item (`plan deal <civ> <n>`, see [Trade deals](factions.md#trade-deals)) is a one-off purchase at a fixed price, so it runs like a single build: it holds its price back while it waits and takes the deal once what is left covers it. It holds nothing while the civilization won't trade (war, embargo, rivalry, hostility). It drops out once taken, or when the offer is gone because the civilization's offers rotated or you took it by hand. Offers don't rotate while you are away, so a deal planned before you leave is still there for the plan to take.
- **Upgrades wait for their price.** `plan upgrade <building> [count|all]` adds upgrades for copies you hold. The plan upgrades one copy at a time, as soon as each can be paid, and drops the item if the upgrade is no longer available.
- **Advance at its place.** An advance item advances the moment the requirements are met, when the walk reaches it: items above it go first, and items below wait for the next tick. Like the `advance` command, it waits behind a pending catastrophe. If a doom's harbinger comes at the plan's advance, the plan tries again on the next tick (see [No outrunning a doom](harbinger.md#no-outrunning-a-doom)).
- **Dead items drop out.** A building of an earlier age after an advance, a building at its limit, a tech already researched: each leaves the plan with a line in the log.
- **A tech missing a prerequisite waits.** Planning a tech plans what it needs above it, so this only happens when one of those later goes missing (you removed it, you moved it below, you canceled it mid-research, or a game update changed what the tech needs). The tech stays in the plan and shows what it needs (`needs Philosophy first`). While it waits it does not hold the research slot's turn, so the missing tech can be planned below it and still start first.
- **A tech of a later age waits for the age.** It shows `waits for the Iron Age`, does not hold the slot's turn, and never drops out for waiting. This age's techs planned below it run as if it were not there.
- **Staffed.** Copies the plan builds are staffed when they finish, from idle workers first: into the copy, or, with a [roster](workers-and-domains.md#the-roster) set, wherever the shares say. If those run out, the plan moves workers out of buildings an advance superseded (a higher tier of their line is open), the same line's first. With shares set, it takes them only from the copy's own domain, so the split holds. It never moves food workers or workers in this age's buildings. Then the roster routine recruits for what is still empty (with auto-recruit on, the default), as housing and food allow, so the next age's producers built after `plan advance` don't sit empty until you come back.

Each tick's starts are summed up in one log line (`Plan: started building 3 Huts, started researching Pottery.`), and the welcome-back message says what the plan did while you were away.

## Optimizing the order

Which building to plan first is a puzzle: a hut gives a camp someone to work it, a camp gives the hut something to be for, and food keeps them both. `plan optimize` (or `plan reorder`) solves it for you. If you ran `plan build hut 10`, `plan build wood_camp 10` and `plan build gathering_camp 10`, it mixes them: a few camps, a hut or two for the people they need, more camps, and so on, in the order that gets the town growing fastest.

```
> plan optimize
Reordered the plan for faster growth. The last building should start in about 7m 55s instead of 13m 5s.
  1. 6 Wood Camps
  2. 4 Huts
  3. 3 Gathering Camps
  4. 4 Wood Camps
  5. 6 Huts
  6. 7 Gathering Camps
```

- **It plays your town forward.** It uses the prices, the crews and the homes the game itself uses: a camp makes 20% of its output with nobody in it and up to the full amount with a full crew, the crews are the people your homes can hold and your food can feed, and each new copy costs more than the last. Every step it buys the building that pays itself back soonest, counting the wait to afford it, the time it takes to build, and the income it adds. A producer that has no one to staff it brings the home or the food it needs ahead of it.
- **The reply shows the new order** (the first 12 items, then how many more) and how long the plan should take to start its last building, before and after. Both come from the same simulation, which plays the plan by the rules above, so the estimate moves with your income and your prices. If some copies can't start until the town makes what they cost, the reply says how many.
- **It only keeps a faster order.** If no order it finds is faster, the plan stays as it was and the reply says so. Running it again on a plan it has already reordered changes nothing. If no building in the plan changes how the town grows, the reply says there is nothing to reorder.
- **Nothing is added or removed.** Every building keeps its total. Items are split and merged to fit the new order, and never go over the plan's 60 items (when the best order would need more, it buys in bigger steps).
- **Other items keep their place.** Research, deals and upgrades stay where they were among the builds, counted in copies, and keep their order. So do storage, knowledge buildings, the army, wonders and anything that cannot start yet (another age's buildings, a building whose tech is not researched). Only buildings that make something or house people move.
- **An advance ends the reordering.** The buildings before it are what it waits for, so they are reordered among themselves. The advance and everything after it stay as they are.
- **It plays the first 2,000 copies.** A plan can hold 1,000 copies of one building. Past the first 2,000 copies of the plan, the rest keep their order after the reordered ones, and the reply says how many.
- **The same town gives the same order,** every time.

## The research queue

The plan is how you queue research. One command lays out a whole chain toward a goal:

```
> plan research imperial_legions
Planned research: Imperial Legions, after the 2 techs it needs first (Military Tactics, then Siege Warfare). Techs start one at a time, in plan order.
```

- **What a tech needs goes in first.** `plan research <tech>` adds every prerequisite that is not researched, being researched or already in the plan, each after its own, and then the tech. What is already on its way is not added twice.
- **Either-or groups.** Where a tech needs one of two or more techs, the plan takes the one you already hold, are researching or have planned. If none is, it takes the branch whose techs cost the least in all (the first listed when they cost the same). Plan the other branch yourself first if you want it instead.
- **It always fits.** Research does not count against the plan's 60 items, so a chain goes in whole however full the plan is. The Plan panel counts the two apart (`42 of 60 items, and 9 techs queued (research takes no room)`).
- **Any tech in sight.** The plan takes a tech of any age you may see on the [tech tree](technologies.md#reading-the-tech-tree): every age up to the next one, and on a later run every age your account has reached before. So on ground you know, you can lay out a run's research in one sitting. A tech of an age still hidden is answered like a key the game does not know.
- **The order is yours.** `plan up` and `plan down` reorder techs like any item. A tech moved above something it needs waits for it, and nothing is dropped.
- **On the tree.** Each planned tech carries its place in the plan on the tree (`▸3`). On a tech's open card, the second `Enter` plans a tech that cannot start now, chain included. A tech that can start now is started instead.

`research <tech>` starts one tech now and refuses one whose prerequisites you do not hold.

## Big prices and the plan

Stock has no limit, so a plan item is never held up by how large its price is. An item waits only for what it costs against what you hold, after the items above it hold back their share. Nothing overflows into the plan or into a wonder's bank.

Research is never held up by stock. It starts in plan order, so the techs behind the first one read `after the research above it`, and the first reads `research slot busy` while another tech is in progress.

## Planned expeditions

`plan expedition <mission>` queues a mission that is on the expedition board. It leaves when its price is covered and a party of soldiers is at home, the same as sending it yourself, and it holds back its price from the items below it. While the party is away the item waits and holds nothing back. If the offer leaves the board before the plan reaches it, the item drops out and the log says so. It behaves the same while you are away.

Every item also reports when it starts: the amount it needs, what you hold, what the items above hold back, the income, and how long until the slowest resource is ready. A resource with no income reads `never`, and a wonder reads `needs you`, since it is paid by hand.

## The Plan panel

Type `plan` to open the Build Plan panel. It takes the whole screen but the command bar, and its keys are letters, so press `Esc` before you type a command.

```
═══ Build plan ═══
 ▸  1. research Currency
       blocked  research slot busy
    2. Lumber Mill ×3 (1 started)
       waiting  █░░░░░░░  20%  wood 1.22K / 6.04K
    3. Quarry ×5
       waiting  ░░░░░░░░   0%  wood 0 / 3.91K (1.22K held for items above)
    4. advance to the Iron Age
       blocked  waiting for the requirements

 3 of 60 items, and 1 tech queued (research takes no room)

 ↑↓  select   U  up   D  down   X  remove   C  clear
```

Each item shows one of three states. **ready** starts on the next tick. **waiting** is saving up (the bar is how much of the next price what is free after the items above it covers). **blocked** says what the item is waiting for.

| Key | What it does |
|---|---|
| `↑` `↓` (or `k` `j`) | Select an item |
| `U` / `D` | Move the selected item up or down |
| `X` (or `Delete`) | Remove the selected item |
| `C` | Clear the plan. It asks for a second press, and the key is the capital letter |
| `Esc` | Close the panel |

<figure class="screen" data-screen="plan"><figcaption>The Plan panel with four items: a tech waiting for the research slot, two builds saving up for wood, and an advance waiting for the age's requirements.</figcaption></figure>

## Offline

Offline progress covers up to **24 hours**. Time away beyond that isn't credited. When you come back, the offline catch-up runs in steps of 30 ticks (a minute at the base tick speed). Each step credits that step's production at the usual 50% offline rate, moves construction and research on, and lets the plan start what the step paid for. Buildings under construction and research finish while you're away, and the plan's queued techs start one after another. Each step also runs the [roster](workers-and-domains.md#the-roster) routine: it puts idle workers to work by your shares and, with auto-recruit on, recruits into empty worker slots as housing and food allow, so what the plan builds gets staffed while you're away. The welcome back says what it did, for example "While you were away, your roster recruited 12 workers (population 40/50), put 3 idle workers to work." With an empty plan, nothing under construction and no workers for the roster routine to place or recruit, the steps pay the same total as one lump payment would.

## The Plan Template

The game remembers the plan you write, age by age, for [Plan Template](prestige.md#plan-template) after your first prestige. Each item is recorded with the age you added it in: builds, upgrades, techs and advances. A build counts the copies you added, and removing an item takes back the copies it never started. Deals aren't recorded, since a civilization's offers end with the run.

At each prestige and Succumb the record becomes the **template**. An age the run wrote in takes this run's part. An age the run never reached keeps what an older run wrote there, so a short run never wipes a deeper one.

From the first prestige, **Plan Template** adds that age's part of the template to the plan at the start of every run and every advance, including an advance item you planned. A plan written once then chains ages while you are away: the plan advances, the next age's part goes in, and the plan works through it. The items go through the same checks as the plan commands and the 60-item limit, and one log line says what went in (`Plan Template: added 12 items for the Bronze Age.`). If the plan is full, the line says how many items waited out, and it counts any that can't be planned yet. The items it adds count as written again, so the template carries forward from run to run. See [Plan Template](prestige.md#plan-template) for the details.

The techs you planned come along with everything else: a `plan research` item is planned again in the age you added it in, and so is each tech the command added before it. Only planned techs carry over this way. The game never chooses what you research next, and a tech you started by hand with `research` is not part of the template.

## Saving

The plan is saved with your game and comes back exactly as it was. Prestige, Succumb and a new game clear it. With the Plan Template owned, a new run (or the rebuild after a Succumb) starts with the first age's part of your template.

## See also

- [Storage](buildings.md#storage-buildings-21-tiers): each copy adds 1% to all production, with no copy limit and 1.75 times the price for each copy.
