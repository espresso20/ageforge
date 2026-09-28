# The Build Plan

AgeForge is paced for a player who checks in a few times a day. The build plan is how you put the hours between visits to work: a list of builds, techs, trades and an advance that the game starts for you, in order, as the resources come in. It runs while you play, and it runs while you are away (offline catch-up executes it as the time passes, not in one lump at the end).

Open it with `plan`. Add to it with commands:

```
plan build hut 10              # ten huts, one after another along the cost curve
plan build gathering_camp 5
plan research tool_making      # techs start one at a time, in plan order
plan trade gold stone 50000    # sell gold for stone as gold comes in, until 50K stone is bought
plan trade gold stone          # no amount: keep stone topped up
plan advance                   # advance as soon as the requirements are met
plan build longhouse 15        # the next age's buildings wait for the advance
```

`plan list` prints it; `plan remove <n>`, `plan up <n>`, `plan down <n>` and `plan clear` edit it (the panel does the same with keys). See [Commands](commands.md#build-plan).

---

## How it runs

- **Paid when it starts, not when you add it.** Adding an item costs nothing. Each copy is paid for at the moment it starts, at the price it has then, through the same checks as `build` and `research` (the age lock, a building's limit, one research at a time).
- **In order, every tick.** After each tick's production the game walks the plan from the top and starts everything it can afford. A build item with a count starts as many copies as the resources cover.
- **Waiting items hold their price.** An item that can't start yet doesn't block the items after it, but it holds back the price of its next copy. A later item only starts if it can be paid from what is left, so a cheap item lower down can never delay one above it, and resources the top items don't need aren't left idle. The order is your priority.
- **Some items hold nothing.** An item that can't start for a reason money won't fix holds nothing back: a price bigger than your storage (build storage first), a resource the current income won't bring in within a day (it needs the market or a producer first), or the next age's building before the advance.
- **A wonder pays its bank from what you hold.** A wonder's price in the plan is what its bank still lacks. Once what you hold (after what the items above it hold back) covers all of that, the plan banks it and starts construction, so a wonder whose stock sat in your stores no longer waits for you to `wonder collect`. While it waits it holds nothing back: it is a big bill, and holding it would stall everything below it. A part bigger than a full store can't be paid at once; deposits and [wonder overflow](wonders.md#overflow) fill it as before.
- **Techs queue.** Only the first research item in the plan can take the research slot when it frees up; later ones still hold their knowledge.
- **Trades are paced.** A trade item holds back what it will sell (what the items above leave, up to what it still wants) and sells once the market has recovered from its last sale, so the rate stays within a percent of the market's instead of sinking by selling every tick. It never buys more than the store has room for, and it needs a trade building standing, like `trade`.
- **Advance at its place.** An advance item advances the moment the requirements are met, when the walk reaches it: items above it go first, items below wait for the next tick, so they can't spend what the requirements count.
- **Dead items drop out.** A building of an earlier age after an advance, a building at its limit, a tech already researched, a tech whose prerequisite is neither researched nor planned before it, a trade the new age's market doesn't offer: each leaves the plan with a line in the log.
- **Staffed.** Copies the plan builds are staffed from your idle workers when they finish. The plan never recruits.

Each tick's starts are summed up in one log line (`Plan started: 3 × Hut, research Pottery`), and the welcome-back message says what the plan did while you were away.

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
       blocked  needs more knowledge storage

  ↑↓  select   U  up   D  down   X  remove   C  clear
```

**ready** starts on the next tick, **waiting** is saving up (the bar is how much of the next price is free after the items above), **blocked** says what it is waiting for. `C` asks for a second press before it clears.

## Offline

When you come back, the offline catch-up runs in one-minute steps: each step credits that minute's production (at the usual 50% offline rate, up to your caps), moves construction and research on, and lets the plan start what the step paid for. Buildings under construction and research finish while you're away, and the plan's queued techs start one after another. A day away resolves in a few milliseconds. With an empty plan and nothing under construction it pays exactly what it always did.

## Saving

The plan is saved with your game and comes back exactly as it was. Prestige, Succumb and a new game clear it.

## See also

- [Wonder overflow](wonders.md#overflow): what a full store would waste goes into the current wonder.
- [Storage](buildings.md#storage-buildings-21-tiers): each age's storage holds about an hour and a half of its production.
