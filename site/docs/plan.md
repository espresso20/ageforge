# The Build Plan

AgeForge is paced for a player who checks in a few times a day. The build plan is how you put the hours between visits to work: a list of builds, techs, deals with other civilizations and an advance that the game starts for you, in order, as the resources come in. It runs while you play, and it runs while you are away (offline catch-up executes it as the time passes, not in one lump at the end). A price larger than a store cannot be paid, so the plan will not start it (see [Full stores and the plan](#full-stores-and-the-plan)). The plan holds up to **60 items**, not counting research.

Open it with `plan`. Add to it with commands:

```
plan build hut 10              # ten huts, one after another along the cost curve
plan build gathering_camp 5
plan research tool_making      # techs start one at a time, in plan order
plan research imperial_legions # a tech that needs others: they go in first, in order
plan advance                   # advance as soon as the requirements are met
plan deal merchant_guild 2     # take the Merchant Guild's deal 2 once its price is there
plan build longhouse 15        # a next-age building waits for the advance
```

`plan list` prints it; `plan remove <n>`, `plan up <n>`, `plan down <n>` and `plan clear` edit it (the panel does the same with keys). See [Commands](commands.md#build-plan). The plan holds at most 60 items; a full plan refuses a new one until you remove one (adding more of the building at the end of the plan still adds to that item). Research takes no room: the 60 are for builds, deals and the advance, and a research queue is bounded by the tree, each tech once. One build item holds at most 1,000 copies.

---

## How it runs

- **Paid when it starts, not when you add it.** Adding an item costs nothing. Each copy is paid for at the moment it starts, at the price it has then, through the same checks as `build` and `research` (the age lock, a building's limit, one research at a time).
- **In order, every tick.** After each tick's production the game walks the plan from the top and starts everything it can afford. A build item with a count starts as many copies as the resources cover.
- **Waiting items hold their price.** An item that can't start yet doesn't block the items after it, but it holds back the price of its next copy. A later item only starts if it can be paid from what is left, so a cheap item lower down can never delay one above it, and resources the top items don't need aren't left idle. The order is your priority.
- **Some items hold nothing.** An item that can't start for a reason money won't fix holds nothing back: a price bigger than your storage (it reads `blocked  needs more <resource> storage`: build storage first), a resource the current income won't bring in within a day (it needs a producer first), or the next age's building before the advance.
- **A wonder pays its bank from what you hold.** A wonder's price in the plan is what its bank still lacks. Once what you hold (after what the items above it hold back) covers all of that, the plan banks it and starts construction, so you don't need to `wonder collect` for it. While it waits it holds nothing back: it is a big bill, and holding it would stall everything below it. A part bigger than a full store can't be paid at once; deposits and overflow fill it over time, and the plan's wonder takes overflow only while [wonder overflow](wonders.md#overflow) is on. A wonder whose [keystone tech](wonders.md#the-keystone-tech) is not researched yet waits for it (`waits for Stoneworking`) and starts once it is; overflow fills its bank in the meantime while wonder overflow is on.
- **Techs queue.** Only the first research item in the plan can take the research slot when it frees up; a tech has no price to hold. See [The research queue](#the-research-queue) for what `plan research` adds and which techs it takes.
- **Deals wait for their price.** A deal item (`plan deal <civ> <n>`, see [Trade deals](factions.md#trade-deals)) is a one-off purchase at a fixed price, so it runs like a single build: it holds its price back while it waits and takes the deal once what is left covers it. It holds nothing while the goods wouldn't fit in their store or the civilization won't trade (war, embargo, rivalry, hostility). It drops out once taken, or when the offer is gone: the civilization's offers rotated or you took it by hand. Offers don't rotate while you are away, so a deal planned before you leave is still there for the plan to take.
- **Advance at its place.** An advance item advances the moment the requirements are met, when the walk reaches it: items above it go first, and items below wait for the next tick.
- **Dead items drop out.** A building of an earlier age after an advance, a building at its limit, a tech already researched: each leaves the plan with a line in the log.
- **Sell orders are gone.** The market no longer swaps resources, so a plan saved with sell orders loses them when it loads. See [Saving and Loading](saving-and-loading.md#saves-from-before-the-market-ended).
- **A tech missing a prerequisite waits.** Planning a tech plans what it needs above it, so this only happens when one of those later goes missing (you removed it, you moved it below, you canceled it mid-research, or a game update changed what the tech needs). The tech stays in the plan and shows what it needs (`needs Philosophy first`). While it waits it does not hold the research slot's turn, so the missing tech can be planned below it and still start first.
- **A tech of a later age waits for the age.** It shows `waits for the Iron Age`, does not hold the slot's turn, and never drops out for waiting. This age's techs planned below it run as if it were not there.
- **Staffed.** Copies the plan builds are staffed when they finish, from idle workers first: into the copy, or, with a [roster](workers-and-domains.md#the-roster) set, wherever the shares say. If those run out, the plan moves workers out of buildings an advance superseded (a higher tier of their line is open), the same line's first; with shares set, only from the copy's own domain, so the split holds. It never moves food workers or workers in this age's buildings. Then the roster routine recruits for what is still empty (with auto-recruit on, the default), as housing and food allow, so the next age's producers built after `plan advance` don't sit empty until you come back.

Each tick's starts are summed up in one log line (`Plan started: 3 × Hut, research Pottery`), and the welcome-back message says what the plan did while you were away.

## The research queue

The plan is how you queue research. One command lays out a whole chain toward a goal:

```
> plan research imperial_legions
Planned research: Imperial Legions, after the 2 techs it needs first (Military Tactics, then Siege Warfare). Techs start one at a time, in plan order.
```

- **What a tech needs goes in first.** `plan research <tech>` adds every prerequisite that is not researched, being researched or already in the plan, each after its own, and then the tech. What is already on its way is not added twice.
- **Either-or groups.** Where a tech needs one of two or more techs, the plan takes the one you already hold, are researching or have planned. If none is, it takes the branch that costs the least work to add (the first listed when they cost the same). Plan the other branch yourself first if you want it instead.
- **It always fits.** Research does not count against the plan's 60 items, so a chain goes in whole however full the plan is. The Plan panel counts the two apart (`42 of 60 items, and 9 techs queued`).
- **Any tech in sight.** The plan takes a tech of any age you may see on the [tech tree](technologies.md#reading-the-tech-tree): every age up to the next one, and on a later run every age your account has reached before. So on ground you know, you can lay out a run's research in one sitting. A tech of an age still hidden is answered like a key the game does not know.
- **The order is yours.** `plan up` and `plan down` reorder techs like any item. A tech moved above something it needs waits for it; nothing is dropped.
- **On the map.** Each planned tech carries its place in the plan on the tree (`▸3`), and the second `Enter` on a tech's card plans it the same way, chain included.

`research <tech>` itself is unchanged: it starts one tech now and refuses one whose prerequisites you do not hold.

## Full stores and the plan

What a full store cannot hold goes into this age's [wonder bank](wonders.md#overflow) while wonder overflow is on, and is otherwise lost. It does not pay toward the plan's items.

A plan item whose price is larger than a store cannot start, however long you wait. It shows as blocked and reads `blocked  needs more <resource> storage` until you build more storage (see [Storage is a wall](buildings.md#storage-is-a-wall)). The age's wonder is the one exception: it is paid through its bank.

A tech has no price to pay, so it is never blocked by a store. Research starts in plan order, so the techs behind the first one read `after the research above it`, and the first reads `research slot busy` while another tech is in progress.

A save from before this rule had banked part of its items' prices in the plan. Those banks are returned to your stores when the save loads (see [Old saves](saving-and-loading.md#saves-from-before-the-storage-wall)).

## The Plan panel

```
═══ Build Plan ═══
 ▸  1. Hut ×3
       ready    starts next tick for 14 wood
    2. Story Circle
       blocked  needs more wood storage
    3. Wood Camp
       ready    starts next tick for 16 wood
    4. Stash ×2
       waiting  ████░░░░  57%  wood 20 / 35 (30 held for items above)
    5. research Tool Making
       blocked  research slot busy

  ↑↓  select   U  up   D  down   X  remove   C  clear
```

**ready** starts on the next tick, **waiting** is saving up (the bar is how much of the next price what is free after the items above covers), **blocked** says what it is waiting for. `C` asks for a second press before it clears.

<figure class="screen" data-screen="plan"><figcaption>The Plan panel with five items: houses saving up for their next copy, a tech held back by the research slot, two builds waiting their turn for wood, and an advance that goes once the next age is ready.</figcaption></figure>

## Offline

Offline progress covers up to **24 hours**; time away beyond that isn't credited. When you come back, the offline catch-up runs in one-minute steps: each step credits that minute's production (at the usual 50% offline rate, up to your storage), moves construction and research on, and lets the plan start what the step paid for. Buildings under construction and research finish while you're away, and the plan's queued techs start one after another. What a full store would waste during a step goes to the wonder while wonder overflow is on, and is otherwise lost, as in live play. Each step also runs the [roster](workers-and-domains.md#the-roster) routine: it puts idle workers to work by your shares and, with auto-recruit on, recruits into empty worker slots as housing and food allow, so what the plan builds gets staffed while you're away. The welcome back says what it did, for example "While you were away, your roster recruited 12 workers (population 40/50), put 3 idle workers to work." A day away resolves in a few milliseconds. With an empty plan, nothing under construction and no workers for the roster routine to place or recruit, the steps pay the same total as one lump payment would.

## The Plan Template

The game remembers the plan you write, age by age, for [Plan Template](prestige.md#plan-template) after your first prestige. Each item is recorded with the age you added it in: builds, techs and advances. A build counts the copies you added, and removing an item takes back the copies it never started. Deals aren't recorded, since a civilization's offers end with the run.

At each prestige and Succumb the record becomes the **template**. An age the run wrote in takes this run's part; an age the run never reached keeps what an older run wrote there, so a short run never wipes a deeper one.

From the first prestige, **Plan Template** adds that age's part of the template to the plan at the start of every run and every advance, including an advance item you planned. A plan written once then chains ages while you are away: the plan advances, the next age's part goes in, and the plan works through it. The items go through the same checks as the plan commands and the 60-item limit, and one log line says what went in (`Plan Template: added 12 items for the Bronze Age.`). The items it adds count as written again, so the template carries forward from run to run. See [Plan Template](prestige.md#plan-template) for the details.

The techs you planned come along with everything else: a `plan research` item is planned again in the age you added it in, and so is each tech the command added before it. Only planned techs carry over this way. The game never chooses what you research next, and a tech you started by hand with `research` is not part of the template.

## Saving

The plan is saved with your game and comes back exactly as it was. Prestige, Succumb and a new game clear it; with the Plan Template owned, a new run (or the rebuild after a Succumb) starts with the first age's part of your template.

## See also

- [Wonder overflow](wonders.md#overflow): what a full store would waste goes into the current wonder while overflow is on.
- [Storage](buildings.md#storage-buildings-21-tiers): no copy limit, 1.75 times more for each copy, and a price larger than a store cannot be bought.
