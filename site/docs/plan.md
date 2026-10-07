# The Build Plan

AgeForge is paced for a player who checks in a few times a day. The build plan is how you put the hours between visits to work: a list of builds, techs, trades, deals with other civilizations and an advance that the game starts for you, in order, as the resources come in. It runs while you play, and it runs while you are away (offline catch-up executes it as the time passes, not in one lump at the end). What your full stores would throw away goes toward the plan's next copies instead of being lost (see [Overflow pays the plan](#overflow-pays-the-plan)), and the plan holds up to **60 items**.

Open it with `plan`. Add to it with commands:

```
plan build hut 10              # ten huts, one after another along the cost curve
plan build gathering_camp 5
plan research tool_making      # techs start one at a time, in plan order
plan trade gold stone 50000    # sell gold for stone as gold comes in, until 50K stone is bought
plan trade gold stone          # no amount: keeps selling until you remove it
plan advance                   # advance as soon as the requirements are met
plan deal merchant_guild 2     # take the Merchant Guild's deal 2 once its price is there
plan build longhouse 15        # a next-age building waits for the advance
```

`plan list` prints it; `plan remove <n>`, `plan up <n>`, `plan down <n>` and `plan clear` edit it (the panel does the same with keys). See [Commands](commands.md#build-plan). The plan holds at most 60 items; a full plan refuses a new one until you remove one (adding more of the building at the end of the plan still adds to that item). One build item holds at most 1,000 copies.

---

## How it runs

- **Paid when it starts, not when you add it.** Adding an item costs nothing. Each copy is paid for at the moment it starts, at the price it has then, through the same checks as `build` and `research` (the age lock, a building's limit, one research at a time).
- **In order, every tick.** After each tick's production the game walks the plan from the top and starts everything it can afford. A build item with a count starts as many copies as the resources cover.
- **Waiting items hold their price.** An item that can't start yet doesn't block the items after it, but it holds back the price of its next copy. A later item only starts if it can be paid from what is left, so a cheap item lower down can never delay one above it, and resources the top items don't need aren't left idle. The order is your priority.
- **Some items hold nothing.** An item that can't start for a reason money won't fix holds nothing back: a price bigger than your storage (build storage first, or let overflow bank the part over it), a resource the current income won't bring in within a day (it needs the market or a producer first), or the next age's building before the advance.
- **A wonder pays its bank from what you hold.** A wonder's price in the plan is what its bank still lacks. Once what you hold (after what the items above it hold back) covers all of that, the plan banks it and starts construction, so you don't need to `wonder collect` for it. While it waits it holds nothing back: it is a big bill, and holding it would stall everything below it. A part bigger than a full store can't be paid at once; deposits and overflow fill it over time, and a wonder in the plan takes overflow whether [wonder overflow](wonders.md#overflow) is on or off.
- **Techs queue.** Only the first research item in the plan can take the research slot when it frees up; later ones still hold their knowledge.
- **Trades are paced.** A trade item holds back what it will sell (what the items above leave, up to what it still wants) and sells once the market has recovered from its last sale, so the rate stays within a percent of the market's instead of sinking by selling every tick. It never buys more than the store has room for, and like `trade` it needs a trade building. With an amount, the item drops out once it has bought that much. With no amount it is an open-ended sell order: it stays until you remove it, and while the store it buys into is full it pauses and holds nothing back.
- **Deals wait for their price.** A deal item (`plan deal <civ> <n>`, see [Trade deals](factions.md#trade-deals)) is a one-off purchase at a fixed price, so it runs like a single build: it holds its price back while it waits and takes the deal once what is left covers it. It holds nothing while the goods wouldn't fit in their store or the civilization won't trade (war, embargo, rivalry, hostility). It drops out once taken, or when the offer is gone: the civilization's offers rotated or you took it by hand. Offers don't rotate while you are away, so a deal planned before you leave is still there for the plan to take.
- **Advance at its place.** An advance item advances the moment the requirements are met, when the walk reaches it: items above it go first, items below wait for the next tick, so they can't spend what the requirements count.
- **Dead items drop out.** A building of an earlier age after an advance, a building at its limit, a tech already researched, a trade the new age's market doesn't offer: each leaves the plan with a line in the log.
- **A tech missing a prerequisite waits.** The plan only takes a tech whose prerequisites are researched, in progress or planned above it. If one of them later goes missing (you removed it from above, you canceled it mid-research, or a game update changed what the tech needs), the tech stays in the plan and shows what it needs (`needs Philosophy first`). While it waits it holds neither the research slot's turn nor its knowledge, so the missing tech can be planned below it and still start first.
- **Staffed.** Copies the plan builds are staffed when they finish, from idle workers first: into the copy, or, with [worker shares](workers-and-domains.md#worker-shares) set, wherever the shares say. If those run out, the plan moves workers out of buildings an advance superseded (a higher tier of their line is open), the same line's first; with shares set, only from the copy's own domain, so the split holds. It never moves food workers or workers in this age's buildings. Then the shares routine recruits for what is still empty (with auto-recruit on, the default), as housing and food allow, so the next age's producers built after `plan advance` don't sit empty until you come back.

Each tick's starts are summed up in one log line (`Plan started: 3 × Hut, research Pottery`), and the welcome-back message says what the plan did while you were away.

## Overflow pays the plan

When a store is full, the production it would throw away goes toward your plan instead of being lost:

- **The wonder first.** [Wonder overflow](wonders.md#overflow) takes what this age's wonder still needs, as before.
- **Then the plan, in order.** What is left goes into the bank of the first build item that needs that resource, up to the price of its next copy, then the next item's, and so on down the plan. Whatever no item needs is lost as before.
- **The bank pays first.** When the copy starts, its bank pays its share of the price and your stores pay the rest. While it waits, the item holds back only what its bank doesn't cover, so the items below get more of what you hold.
- **Over the cap.** A copy priced above your storage, which could never start before, starts once overflow has banked the part over the cap and your stores hold the rest. Until then it shows as blocked, with what it has banked.
- **A queued wonder banks too.** A wonder in your plan takes overflow in plan order like any other item, into its own bank (the one `wonder collect` fills), up to what that bank still lacks. With wonder overflow on, this age's wonder has had first call already; with it off, putting the wonder in the plan is what lets its bank fill while you are away.
- **Only what you queued.** Overflow only pays for copies in your plan; it never starts anything on its own. Only this age's buildings bank: the next age's, its wonder included, wait for the advance.
- **Banks go back.** Removing an item, clearing the plan, an item dropping out, and advancing put its bank back into your stores, up to their caps (the rest was overflow and still doesn't fit). The advance does this before it trims your stockpiles, so a bank never carries more into the next age than your stores could.

It works during offline catch-up too, and the welcome back says how much overflow banked toward your plan. The Plan panel shows each item's bank. There is no switch: a plan with nothing in it banks nothing.

## The Plan panel

```
═══ Build Plan ═══
 ▸  1. Hut ×3
       ready    starts next tick for 14 wood
    2. Story Circle
       blocked  needs more wood storage (12 wood banked from overflow)
    3. Wood Camp
       ready    starts next tick for 16 wood
    4. Stash ×2
       waiting  ████░░░░  57%  wood 20 / 35 (30 held for items above)
    5. research Tool Making
       blocked  needs more knowledge storage

  ↑↓  select   U  up   D  down   X  remove   C  clear
```

**ready** starts on the next tick, **waiting** is saving up (the bar is how much of the next price its bank and what is free after the items above cover), **blocked** says what it is waiting for. What an item has banked from overflow shows on its line. `C` asks for a second press before it clears.

## Offline

Offline progress covers up to **24 hours**; time away beyond that isn't credited. When you come back, the offline catch-up runs in one-minute steps: each step credits that minute's production (at the usual 50% offline rate, up to your storage), moves construction and research on, and lets the plan start what the step paid for. Buildings under construction and research finish while you're away, and the plan's queued techs start one after another. What a full store would waste during a step goes to the wonder and then toward the plan, as in live play. Each step also runs the [worker shares](workers-and-domains.md#worker-shares) routine: it puts idle workers to work by your shares and, with auto-recruit on, recruits into empty worker slots as housing and food allow, so what the plan builds gets staffed while you're away. The welcome back says what it did, for example "While you were away, your worker shares recruited 12 workers (population 40/50), put 3 idle workers to work." A day away resolves in a few milliseconds. With an empty plan, nothing under construction and no workers for the shares routine to place or recruit, the steps pay the same total as one lump payment would.

## The Plan Template

The game remembers the plan you write, age by age, for the prestige [legacy kit](prestige.md#the-legacy-kit). Each item is recorded with the age you added it in: builds, techs, trades and advances. A build counts the copies you added, and removing an item takes back the copies it never started; a trade removed after it bought something stays recorded for the amount it bought. Deals aren't recorded, since a civilization's offers end with the run.

At each prestige and Succumb the record becomes the **template**. An age the run wrote in takes this run's part; an age the run never reached keeps what an older run wrote there, so a short run never wipes a deeper one.

With the kit's **Plan Template** (`prestige buy legacy_plan`, 9 points) owned, the start of every run and every advance add that age's part of the template to the plan, the advance item included if you planned one. A plan written once then chains ages while you are away: the plan advances, the next age's part goes in, and the plan works through it. The items go through the same checks as the plan commands and the 60-item limit, and one log line says what went in (`Plan Template: added 12 items for the Bronze Age.`). The items it adds count as written again, so the template carries forward from run to run. See [Plan Template](prestige.md#plan-template) for the details.

The techs you planned come along with everything else: a `plan research` item is planned again in the age you added it in. Only planned techs carry over this way. The game never chooses what you research next, and a tech you started by hand with `research` is not part of the template.

## Saving

The plan is saved with your game, banks included, and comes back exactly as it was. Prestige, Succumb and a new game clear it; with the Plan Template owned, a new run (or the rebuild after a Succumb) starts with the first age's part of your template.

## See also

- [Wonder overflow](wonders.md#overflow): what a full store would waste goes into the current wonder first.
- [Storage](buildings.md#storage-buildings-21-tiers): each age's storage holds at least 4.5 hours of its typical production from the Bronze Age on (an hour and a half in the Primitive and Stone Ages).
