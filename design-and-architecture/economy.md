# AgeForge: Economy Design

## Overview

The AgeForge economy has four interlocking systems. **They must be designed together, not
independently.** Designing any one system in isolation produces the bugs and imbalances that
prompted this document.

```
Housing → Population → Workers → Production
    ↑                                  ↓
    └─────── Building Costs ←──────────┘
                    ↑
              Storage Capacity
```

---

## The Four Laws

These are non-negotiable design constraints. Every balance change must be checked against them.

### Law 1: The Storage Covenant
> At any stage of the game, a player's maximum possible storage for a resource **must be
> ≥ 2× the cost** of the most expensive building they are expected to build at that stage.

Violation of this law makes buildings literally unbuildable (costs exceed what can ever be
accumulated). This was the root cause of hut #68+ being impossible with 50 stashes.

For age gates this is made exact by the **Gate Covenant** (see Storage Design): the last
required copy of every required building costs at most half the storage buildable by then,
with no discounts, and a unit test enforces it.

The covenant has a second, time-based clause (2026-09-27):

> The most storage buildable in an age must hold at least **`StorageHoldHours` (1.5) hours
> of the age's typical production** of each of its construction resources.

Typical production is `config.TypicalIncome(res, age)`, `FlowIncome`'s definition applied
to every resource: `FlowCopies` (5) fully staffed copies of every producer up to the age,
the earlier wonders and the techs, times the `production_all` bonus held by then. Prices
size storage from below (the Gate Covenant); this clause sizes it to time. Without it a
full Renaissance to Victorian store filled in 5 to 15 minutes, so a player away for three
hours lost nearly all of it.

**Why 1.5 hours.** The smoke bot ends an age making one to three times the typical income
(it keeps more than five copies of the best producers and has the bonuses), so 1.5 hours of
typical income is about one hour of a well-built economy: a player who checks in hourly
loses nothing at a cap. Longer absences are the build plan's and wonder overflow's job,
which put the income to work instead of holding it, so storage keeps its pressure. At 2
hours (every age's storage about a third bigger again) 8-hour check-ins improved by about
5% and 3-hour ones not at all (measured with an earlier bot), so the extra storage wasn't
earning its keep. The lever is the age's storage per
copy, raised only where an age falls short; there is no flat multiplier. `smoke.StaticStorage`
checks it, `TestStorageCovenant` fails `go test ./...` when a change breaks it, and
`TestStorageCovenantCatchesBrokenStorage` feeds it the old numbers.

### Law 2: The Pacing Curve
> Every age has a target length at 1x, and the economy is derived from it. Nothing is priced
> or rated by guesswork against "normal play pace".

AgeForge is a terminal idle game that players check a few times a day, and it grants up to
24 hours of offline progress, so game time is close to calendar time. The targets
(`config.AgeTargets`, 2026-09-27):

| Age | Target | Age | Target | Age | Target |
|-----|--------|-----|--------|-----|--------|
| Primitive | 15 min | Renaissance | 6 h | Information | 14 h |
| Stone | 45 min | Colonial | 7 h | Digital | 16 h |
| Bronze | 1.5 h | Industrial | 8 h | Cyberpunk | 18 h |
| Iron | 2.5 h | Victorian | 9 h | Fusion | 20 h |
| Classical | 3.5 h | Electric | 10 h | Space | 22 h |
| Medieval | 4.5 h | Atomic | 12 h | Interstellar, Galactic, Quantum | 24 h each |
| | | Modern | 12 h | | |

That is about three days to the Modern Age, where the first prestige unlocks, and about
twelve days to the Quantum Age. **The first prestige is a three-day goal, not a several-week
one**: the old "several weeks of calendar time" target, and the per-building build-time
table that went with it (hours to weeks per building from the Victorian Age on), described
a game no one could finish; no run reached the Modern Age at all.

Derived from the targets:
- **Payback** (Law 3) sets production.
- **Construction time** is at most 1/6 of the age's target, wonders included (the next
  advance needs them), and 1/48 for storage buildings, which are bought many times an
  age.
- **Research time** is at most 1/8 of the tech's age target, so the handful of techs an age
  offers fit in it one at a time.
- The per-building wait falls out of these: with a dozen producers each repaying in the
  payback time, the next copy is minutes away early and an hour or two late.

**Measuring it.** The smoke harness plays a greedy bot on fixed seeds and reports the time
spent in each age; `smoke/targets.go` holds the same table (a test keeps the two equal) and
`-pacing enforce` fails a set of runs whose median for a first-cycle age leaves 0.5x-2x of it. CI runs enforced:
the per-PR fast tier grades the Primitive and Stone Ages, the nightly every age to the first
prestige, and a weekly deep run every age to the Galactic (five seeds to a Quantum Age
prestige). The bot has to play like a reasonable person for this to
measure the game rather than the bot (see the bot's strategy comment in `smoke/bot.go`).

**Idle targets.** The targets describe an attentive player. The player AgeForge is built
for checks in a few times a day, so the check-in player has targets of its own, in
`smoke/idle_targets.go`: the first prestige within **3.5 days at 1-hour check-ins, 5 days
at 3-hour and 8 days at 8-hour**, the median of three seeds. The bot plays those
(`Bot.CheckIn`, `Bot.planAhead`): at each visit it spends and builds storage, then leaves a
build plan for the hours until the next one, with wonder overflow on. The nightly's `idle`
scenario enforces them; the per-age greedy targets don't apply, since a check-in player
can only act at a visit.

### Law 3: The Payback Rule
> A production building, fully staffed, earns back the price of its first copy in its age's
> **payback time**, valued at the age's **price parity**. Its output is derived from its
> price; no rate is typed in for a construction resource.

Definitions:
- A **construction resource** of an age is one that appears in the first-copy price of at
  least one of that age's buildings (wonders aside), except the flow resources below.
- Its **price level** is the median of those first-copy prices. The **parity** of two
  resources is the ratio of their price levels. A **price unit** is one price level: a
  building costing the median in each of three resources costs 3 price units.
- The **payback time** is `target × epochProgress^1.25 / 16`, where `epochProgress` counts
  epochs of three ages continuously (1 in the Primitive Age, 2 in the Iron Age, 3 in the
  Renaissance). That is 1/16 of the target in the Primitive Age, about 1/7 in the Iron Age,
  1/4 in the Renaissance, 1/3 in the Victorian and 2/3 in the Space Age. `PaybackAdjust`
  stretches it for an age the smooth curve leaves too fast; today only the Renaissance
  (1.3x, see the appendix of 2026-09-28).

Then `rate = priceUnits(first copy) × priceLevel(output) / payback / n`, where a building
with n construction-resource outputs splits its value between them. Staffing still scales it
from 20% to 100% (Law 4).

Why the payback grows through the game: a Primitive player has nothing but what they build
that age, so producers have to repay in minutes. From the Bronze Age on every age also runs
on all the older buildings, which keep producing forever, and there are more of them each
age; if new producers repaid as fast, the later ages would fly by (at a flat 1/20 share the
smoke bot finished the Renaissance in under an hour and the whole run to Modern in a day).
Counting progress per age rather than per epoch keeps the first age of an epoch from
inheriting buildings that repaid much faster than its own.

Why this replaced "5-10x per age": rates used to double each age while prices grew 5-8x,
so by the Renaissance a new producer took weeks to repay and gates needed months of
production. Deriving rates from prices makes production track prices by construction; the
old 5-10x target now holds because prices grow that fast.

**Flow resources** (food, faith, culture, soldiers) are exempt: their amounts drive other
systems (food feeds workers, faith sets morale and catastrophe odds, culture fills its own
caps and pays for festivals and monuments, soldiers are an army). Their producers keep
hand-set rates, and requirements on them, including wonder prices, are sized to what the
age produces of them. Resources no building of an age costs (iron ore, marble, knowledge
before the Medieval Age) also keep their typed rates.

"What the age produces" has one definition, **`config.FlowIncome(res, age)`**: the output
of `FlowCopies` (5) fully staffed copies of every producer of the resource from the
Primitive Age up to that age, plus every earlier age's wonder (each advance requires its
age's wonder, so they stand), plus the flat output of the techs up to that age. It is a
moderate investment, not a maximum, and it leaves out bonuses and worker upkeep. The Gate
Covenant's flow check and the harbinger's Appease price are both sized to it.

**Market parity (corollary).** The market trades any two construction resources of the
player's current age at parity less a 20% fee, whether or not the pair is listed in
`config.BaseExchangeRates`. So a round trip always loses value and trading never beats
building, and every construction resource has a source as long as one of them is produced:
stone after the Bronze Age, iron after the Medieval Age, steel from the Modern Age on,
crypto, and titanium outside the Space Age (whose Orbital Refinery makes it) come from the
market. Listed pairs with a flow resource (or knowledge where it isn't a
construction resource) keep their fixed rates.

**Wonders** cost `WonderPriceUnits` (40) price units of their age, in their typed resource
proportions, plus hand-sized flow parts. They are part of every gate: the current age's
wonder must stand before `advance`.

### Law 4: The Coupling Law
> Every production building has a **worker capacity** (how many workers of a given type it
> can employ) and a **worker output** (production per assigned worker per tick). Buildings
> produce at **20% base rate without workers**. Full production requires workers.

This means:
- Housing → Population → Workers → Production is a real chain, not two parallel systems.
- Players must invest in housing to unlock production potential.
- Workers and buildings are complementary, not competing.
- The "idle" base rate (20%) means the game keeps running if you step away and forget to
  assign, but assigned workers provide a 5× multiplier, which makes assignment worth the effort.

---

## Worker Domain Mapping

Worker types map to production domains, not individual buildings. One worker type (per age tier)
serves all buildings in its domain. **See [workers.md](workers.md) for the full tier table**
with age-specific class names (Gatherer → Tribesman → Laborer → Serf → ... → Harvester).

| Domain | Unlocks | Example Buildings |
|--------|---------|-------------------|
| Raw Materials | Primitive | gathering_camp, stone_pit, farm, quarry, mine, oil_well |
| Knowledge | Primitive | altar, sacred_grove, library, cathedral, university |
| Military | Primitive | barracks, keep, bunker, missile_silo |
| Trade | Primitive | market, port, bank, amphitheater |
| Engineering | Bronze | smithy, factory, reactor, plasma_forge, launch_pad |
| Hacker | Information | server_farm, data_center, ai_lab, smart_grid |
| Astronaut | Space | space_station, warp_gate, star_forge, reality_anchor |

New buildings must be assigned to an existing domain, or a new domain must be justified
and added with full tier coverage across all relevant ages.

---

## Housing Scale

Since workers must fill building slots across potentially hundreds of buildings, housing must
scale dramatically per tier. Each housing age tier should provide roughly **5× more population
capacity per building** than the previous tier.

| Housing Building | Age | Pop per building | Worker capacity implication |
|-----------------|-----|-----------------|----------------------------|
| Hut | Primitive | +10 | ~50 huts → 500 pop |
| Longhouse | Stone | +25 | ~30 buildings → 750 pop |
| House | Bronze | +50 | ~20 buildings → 1,000 pop |
| Manor | Medieval | +150 | ~15 buildings → 2,250 pop |
| Apartment | Industrial | +500 | ~10 buildings → 5,000 pop |
| Skyscraper | Modern | +2,000 | ~8 buildings → 16,000 pop |

> **Note:** These are design targets, not final tuned numbers. The key property is that
> population capacity should never be the bottleneck that prevents filling building worker
> slots. Housing should be the *first* thing a player builds, not an afterthought.

---

## Cost Scaling Formula

Prices are the given; production is derived from them (Law 3). The direction used to be the
other way round (`building_cost = production_rate × target_build_time`), but the rates were
never set to match, which is how prices outran production by 10^5 by the late game. A tick is
2 seconds at 1x (`game.BaseTickInterval`, mirrored by `config.TickSeconds`).

### Cost Scale Factors

Scale factors control how much more expensive each subsequent copy of a building becomes.
They should be chosen so that the Nth building takes roughly the same real time as the 1st,
given that production has also grown.

General guidelines:
- **Storage buildings**: 1.15-1.20 (slow scale, you want many of these)
- **Production buildings**: 1.25-1.35 (moderate, you want several but not infinite)
- **Housing buildings**: 1.10-1.15 (keep cheap: housing is a prerequisite, not a sink)
- **Specialist buildings** (research, military): 1.35-1.50 (expensive per unit, few needed)
- **Late-game buildings** (space, quantum): 1.45-1.60 (steep is fine, everything takes days)

---

## Storage Design

Storage is the **resource pressure** mechanism: it stops players from walking away for a year
and coming back to infinite resources.

### Rules
- Every resource has a storage cap, set by the sum of storage buildings.
- Storage buildings have **MaxCount** (the only building type that does).
- At each age tier, the available storage buildings must satisfy **Law 1**: their maximum
  combined capacity must be ≥ 2× the most expensive building cost in that tier.
- Storage should scale roughly with production: a player at Bronze Age should have roughly
  5-10× more storage than at Primitive Age.

### Storage Buildings by Age

Every age has one storage building, each capped at **25 copies** (the stash at 50). Storage
never transforms or upgrades: the stashes you built in the Primitive Age still count in the
Quantum Age, and because of the age lock they can never be rebuilt later. The live numbers
are in `config/buildings.go`; the smoke report's "Tightest requirement per advance" table
shows how much headroom each advance has.

Storage is also permanent within a run, like wonders: no catastrophe takes it (Endure, the
Great Fire and Succumb's ruins draw only on other buildings, see `isDestroyable` in
`game/buildings.go`) and `sell` refuses it. Storage is what raises caps (techs add only a
little), the age lock means an older age's storage can never be rebuilt, and the first
copy of an age's storage can cost more than the storage left after a loss (a first
Victorian Vault is about 210M steel). The nightly smoke run found the soft-lock this made:
seed 1's second run entered the Victorian Age with 478M of storage, a Nuclear Exchange
(Endure) destroyed both Industrial Depots and 11 older storage copies, and the 130M left
could never hold a Victorian Vault. The caps could never rise again, so nothing the age
required (academy, Bessemer plant, steam turbine) could ever be afforded, and a player had
no way out but a wipe. Selling storage had the same risk: sell this age's storage just
before advancing and you arrive unable to build the next age's.

### The Gate Covenant (Law 1 applied to age gates)

Law 1 as written ("storage ≥ 2× the most expensive building expected at that stage") left
"expected" open, and every age gate eventually broke it. The Gate Covenant pins it down. For
every advance from age A to age B, against the **most storage buildable by the end of A**
(every storage building up to A at its MaxCount, plus storage techs), with **no build_cost
discounts assumed**:

1. **Buildable:** every building B requires can be built in A. The age lock only lets you
   build the current age's buildings, so a requirement naming an older building (Medieval
   asking for 30 Bronze Age barracks) can only be met by copies built ages earlier, and an
   `upgrade` in between destroys them for good. Requirements name the building the lineage
   has in A.
2. **Storable (buildings):** the last required copy of each required building costs at most
   **half** that storage in every resource it costs. This is the 2× of Law 1.
3. **Storable (resources):** each resource requirement is at most **80%** of that storage
   (1.25× headroom), so the gate never needs the last storage slot of every age.
4. **Storable (wonder):** each part of A's wonder, which every advance also requires, costs
   at most that storage. Wonders are banked a deposit at a time (`wonder collect`), so no
   margin is asked for, but a part bigger than a full store means banking at the cap in
   rounds.
5. **Sourced, from a cold start:** every resource the gate asks for, directly, in a
   required building's price or in the wonder's, can be had **in A** by a player who
   skipped every building no gate required. What A reaches is a fixed point: start from
   what carries over (below), buy every building of A whose first copy that pays for, add
   what those buildings make, open the market once a trade building stands or can be
   bought (the exchange needs one), and add what the market sells (parity between A's
   construction resources, the listed pairs); repeat. The gate's total must then come
   from A itself: A's producers, hand gathering (food, wood, stone, through the Medieval
   Age), the market, techs whose flat output alone covers the whole amount within 48 hours
   at 1x, or the carried stock.

   **What carries over into A**, and nothing else:
   - every building an earlier gate required, and every earlier age's wonder (an upgrade
     keeps the lineage and its output);
   - the resources the gate into A required, held at the advance, up to the carryover cap
     (8 copies of A's cheapest building priced in it; faith whole);
   - a steady supply of every resource an earlier gate required as a resource requirement,
     unless it could be hand-gathered then or the stock carried into that age met it:
     thousands of iron can't be banked without making it, and whatever made it (a producer,
     or a trade building and the market) still stands. Prices of required buildings imply
     no supply: a few scriptoriums' gold can come from events.

   Carried producers are sized to an older age's prices (the Stellar Cradle cost 940T
   uranium in the Fusion Age, where only old Atomic Age mines made any), so they only pay
   for a first copy of A's buildings and seed market parity; they never supply a gate's
   total. A producer that costs its own output (the Bronze Age smithy and iron, the
   Renaissance mill and steel) doesn't count until something else supplies the first
   batch. The rule used to count market parity without asking whether a trade building
   could stand, which let the **Iron Age gold trap** through: the trading post cost gold,
   and it was both the Iron Age's only gold producer and its only trade building, so a
   player who had skipped the optional Bronze Age market could never get gold there (one
   smoke seed spent 2.5 days in the Iron Age). The trading post no longer costs gold.
6. **Flow within the target:** every flow resource the gate asks for (requirement, required
   copies and wonder together) is made within A's pacing target at `FlowIncome`, or the rest
   costs at most **10 price units** of A at the market's listed pairs (gold → food, gold →
   culture). Faith has no market, so a faith requirement must fit A's own output: the
   Renaissance's old 44K faith at about 1.5 faith/tick and the Sistine Chapel's old 6M faith
   both fail it.
7. **Storage ladder:** the first copy of B's storage building fits, with **1.25×** to spare
   (`GateLadderMargin`), in the **least storage the gate into B forces** a player to hold,
   not the most buildable. That is the gate's biggest single price: a resource requirement
   (held all at once) or the last required copy of a building buildable in A (paid all at
   once), undiscounted. The wonder is banked a deposit at a time and forces nothing. Every
   storage building raises every cap alike, so one figure stands for all caps. Storage is
   never lost (see Storage Buildings by Age), so that forced storage is the least anyone
   enters B with, and the age lock leaves B's storage as the only storage they can build
   there. Today every advance passes with room: the tightest is the Classical Age (the first
   Classical Vault, 86K stone, against the 130K stone requirement), and the Victorian
   Vault's 210M steel sits under the 381M stone of the 30th tenement the Victorian gate
   requires.

The same sourcing rule applies to every building on its own: nothing may cost a resource
with no source in the building's own age (coal before the Renaissance, crypto before
Cyberpunk). Those buildings were dead content. The last age's wonder, which no gate
requires, is checked the same way.

`smoke.StaticGates` implements all of this and `TestGateCovenant` (smoke/static_test.go)
fails `go test ./...` when a balance change breaks it, so it runs in CI on every PR.
`TestGateCovenantCatchesBrokenGates` feeds it the old broken numbers and checks each one is
caught, and `TestGateCovenantCatchesColdStartTraps` does the same for the gold-priced trading
post (and checks that a market required by the Iron Age gate would have carried over and
fixed it). `TestGateCovenantCatchesBrokenLadder` breaks the storage ladder (a Victorian gate
cut to 10 tenements, a Victorian Vault at twice its price) and checks both are caught.

**Gate size.** With Law 3 the time an age takes follows mostly from its gate (required
buildings plus wonder, in price units) and the stock of older buildings. Gates that were
far smaller than their neighbors' finished in a fraction of the target and were enlarged
(Classical, Colonial, Industrial); the steel-heavy Space Age gate was trimmed, since steel
and titanium came only from the market there, and the Space Age's Orbital Refinery now
makes titanium.

**Levers, in order of preference.** When a gate breaks the covenant:

1. If it names an older age's building, retarget it to the same lineage's building in A
   (the lowest tier there, which is what the older building upgrades into, so upgraded
   copies count).
2. If the storage buildings of that age are out of band with the age's own prices, raise
   their per-copy storage. "In band" means the median first-copy price is 3-8% of the
   age's max storage, which is where Classical through Space sit.
3. Otherwise lower the count to the largest that fits, rounded down to a multiple of 5
   (exact below 10).
4. For a sourcing failure, drop the unobtainable resource from the price of the one
   producer that bootstraps it (or open an exchange in the age the resource unlocks),
   rather than inventing a new building.

Counts are the usual lever because the normalized cost curves (1.15 per copy, 1.13 for
housing and storage) made late copies explode: copy #80 costs 62,000× copy #1, so the
old late-game counts of 50-500 were never reachable at any storage.

### The Milestone Covenant (Law 1 applied to milestones)

Milestones reset with every run, and nothing proved they fit one: seven could never be
completed (the 50th Stone Pit costs 283K wood where the Stone Age stores at most 80K; a
billion people need more housing than all 22 ages hold) while the old check only compared
counts with MaxCount. The Milestone Covenant proves every milestone, chain and title from
config with the Gate Covenant's model: the most storage buildable, undiscounted prices and
the cold start.

**Due age.** A milestone is due by the end of the last age a normal run plays (the age
before `game.PrestigeMinAge`, the Atomic Age today), or by the end of its own `MinAge` when
that is later. A milestone naming a later age is deep-run content and must be doable in
that age: Tech Master names the Information Age because the tech tree reaches 50 there.

Each requirement must fit by the due age:

1. **Building counts:** a building is built only in its own age (the age lock), so the last
   required copy must cost at most the most storage buildable in that age, in every
   resource it costs, within MaxCount. `MinBuildingSum` (consecutive tiers of one lineage,
   in any mix) fits the sum of those ceilings.
2. **Population:** at most 40% (`MilestonePopShare`) of the housing ceiling, which is every
   housing building so far at its storage limit, never upgraded, plus the techs' housing.
   Nobody keeps every old tier maxed; the newest tier alone is about half the ceiling.
3. **Structures built in the run:** at most 50% (`MilestoneBuildShare`) of the build
   ceiling, which is every building of every age so far at its storage limit, each copy
   built once (7,561 in the whole game and 4,361 in a run today). Selling and rebuilding,
   which the counter also counts, and rebuilding after a catastrophe are churn and left out.
4. **Resources:** the amount fits the most storage buildable with the gate's 1.25x
   headroom, in an age from `MinAge` on that supplies it.
5. **Techs:** a count, or named techs, researchable by then, counting prerequisites and the
   knowledge storage each price needs.
6. **Wonders:** one per age, each part fitting one full store (the gate's wonder rule).
7. **Knowledge workers:** at most 40% of the worker slots in knowledge buildings, capped by
   housing.
8. **Soldiers trained:** what `FlowIncome` makes over the ages' pacing targets.

Play time (`MinTick`) always passes. A chain fails when any of its milestones does, and the
top title must not need more milestones than can be completed. Where the model simplifies,
it errs low: build_cost discounts, the refund an upgrade gets, and upgrades that add copies
of a building after its age can only add to what is possible.

`smoke.StaticMilestones` implements it and `TestMilestonesAreFeasible`
(smoke/static_milestones_test.go) fails `go test ./...` on any problem, naming the
milestone, the requirement and the math. `TestMilestoneFeasibilityCatchesBrokenMilestones`
feeds it the numbers master shipped before (the seven impossible milestones, three out of
reach in a run, three that named an age too early) and checks each is caught, and
`TestMilestoneFeasibilityFollowsPrestige` checks the due age moves with
`game.PrestigeMinAge`. The `static` smoke scenario lists every milestone's earliest age and
its tightest requirement.

**Levers**, when a milestone breaks it:

1. Lower the count, to at most about 80% of the proven ceiling.
2. If the count is right but only fits after the run, move `MinAge` to the age where it
   first fits, making it deep-run content (Tech Master, Wonder Empire, Tech Ascendant,
   Global City).
3. If it counts an old tier the game asks you to upgrade, count the lineage's consecutive
   tiers with `MinBuildingSum` (Trade Empire), so upgrading keeps the progress.

---

## MaxCount Policy

| Building Category | MaxCount | Reason |
|-------------------|----------|--------|
| Storage buildings | Yes, hard cap | Unlimited storage breaks resource pressure |
| Wonders | Yes, 1 | Unique by design |
| Production buildings | **No** | Geometric cost scaling is the natural cap |
| Housing buildings | **No** | Geometric cost scaling + population needs are the natural cap |
| Military buildings | **No** | Scale naturally |
| Research buildings | **No** | Scale naturally |

Players who leave gathering_camp running for 3 months and accumulate a sextillion wood are
playing correctly. That's the idle game working as designed.

---

## Implementation Phases

### Phase 1: Data Model
- [ ] Add `WorkerDomain string` and `WorkerCapacity int` to `BuildingDef` in config/buildings.go
- [ ] Add age-tiered worker class entries to config/villagers.go (see workers.md)
  - Each class: Domain, UnlockAge, FoodCost, OutputMultiplier, Name
- [ ] Update all ~80 building definitions with their domain + capacity values

### Phase 2: Engine
- [ ] Update VillagerManager to handle multi-tier workers per domain
- [ ] Update ResourceManager production calculation:
  `rate = building_base_rate × (0.20 + 0.80 × assigned/capacity)`
- [ ] Update housing pop values (hut: +3 → +10, all tiers rescaled)
- [ ] Remove MaxCount from all production/housing buildings in config/buildings.go

### Phase 3: Balance Numbers
- [x] Derive production from prices and the age targets (the Payback Rule, Law 3, 2026-09-27)
- [x] Verify Storage Covenant (Law 1) for all 22 age transitions (Gate Covenant, 2026-09-26)
- [ ] Tune worker food costs against the new rates (food producers keep hand-set rates)

### Phase 4: UI
- [ ] Update population panel to show current-tier workers prominently, legacy collapsed
- [ ] Update resource rate breakdown to show worker contribution separately from building base
- [ ] Worker assignment UI uses domain name (not class name) to avoid churn on age advance

---

## Appendix: Renaissance pace and 8-hour headroom (2026-09-28)

Two follow-ups to the check-in appendix below. Before: master (nightly run 36432409007,
weekly deep run 36456050959). After: this change (nightly-suite run 36466888675 with the
static, progression and idle scenarios; weekly deep run 36466891574).

### The Renaissance

The storage raise left the greedy bot finishing the Renaissance in 0.57x of its target.
Traced, the age is about 2.5 hours of economy and an hour of construction at the end (the
Sistine Chapel's build time, capped at a sixth of the target). Stone and steel come only
from the market, paid for with gold, and the market's throughput per trade is bounded by
what a store holds, which is why the storage raise sped it up.

What each lever did (local, 8 seeds, median; baseline Renaissance 3.7 h, Colonial 5.4 h):

| Change | Renaissance | Colonial |
|---|---|---|
| Colonial gate 10 exchanges, universities, art studios (was 8) | 4.0 h | 4.8 h |
| Sistine Chapel at 60 price units (was 40) | 3.9 h | 5.1 h |
| Payback 1.4x | 3.9 h | 5.7 h |
| Gate asks 45M gold | 3.8 h | 5.5 h |
| Gate asks 7.5M steel | 3.9 h | 5.0 h |
| Gate asks 30M knowledge | 4.8 h | 4.7 h |
| **Gate asks 30M knowledge, payback 1.3x** | **5.8 h** | **5.2 h** |

Every way of making the gate bigger moved time from the Renaissance into the Colonial
Age hour for hour: the bot keeps investing while it waits, and what it builds speeds the
next age up. Only a slower payback slows both. Knowledge is the lever with teeth because
it is the Renaissance's slow resource (universities make it; the market sells it at parity
with a 20% fee). The fix is both: `config.PaybackAdjust` stretches the Renaissance's payback
1.3x (1.5 h to 1.9 h; its exchanges make 3.76K gold a tick, were 4.89K), and the Colonial
gate asks for 30M knowledge (was 940K; 37% of the Renaissance's most storage, under the
Gate Covenant's 80%). The Storage Covenant still holds (the Renaissance's typical gold
income falls, so its vault holds more hours), and so does the Gate Covenant. The Steel
Era's Brace, 12% of the era's largest knowledge requirement, goes from 360K to 3.6M
knowledge, still under a thousandth of the Industrial Age's storage.

Greedy pacing, median time per age (deep tier, 5 seeds):

| Age | Target | Before | After |
|---|---|---|---|
| Primitive | 15 min | 17 min (1.2x) | 17 min (1.2x) |
| Stone | 45 min | 57 min (1.3x) | 57 min (1.3x) |
| Bronze | 1.5 h | 1.7 h (1.1x) | 1.7 h (1.1x) |
| Iron | 2.5 h | 3.1 h (1.3x) | 3.1 h (1.3x) |
| Classical | 3.5 h | 3.9 h (1.1x) | 3.9 h (1.1x) |
| Medieval | 4.5 h | 3.1 h (0.69x) | 3.1 h (0.69x) |
| **Renaissance** | 6 h | **3.4 h (0.57x)** | **5.8 h (0.97x)** |
| Colonial | 7 h | 5.4 h (0.78x) | 5.1 h (0.73x) |
| Industrial | 8 h | 6.4 h (0.80x) | 6.5 h (0.81x) |
| Victorian | 9 h | 6.1 h (0.67x) | 6.3 h (0.70x) |
| Electric | 10 h | 8.5 h (0.85x) | 8.6 h (0.86x) |
| Atomic | 12 h | 9.1 h (0.76x) | 9.3 h (0.77x) |
| Modern | 12 h | 12.2 h (1.0x) | 12.9 h (1.1x) |
| Information | 14 h | 20.9 h (1.5x) | 21.1 h (1.5x) |
| Digital | 16 h | 18.4 h (1.2x) | 18.3 h (1.1x) |
| Cyberpunk | 18 h | 20.1 h (1.1x) | 20.7 h (1.1x) |
| Fusion | 20 h | 16.7 h (0.84x) | 17.0 h (0.85x) |
| Space | 22 h | 1.3 d (1.4x) | 1.1 d (1.2x) |
| Interstellar | 24 h | 1.6 d (1.6x) | 1.6 d (1.6x) |
| Galactic | 24 h | 1.2 d (1.2x) | 1.3 d (1.3x) |

The first prestige moves from 2.1 to 2.3 days (nightly, 8 seeds). The tightest age is now
the Medieval at 0.69x.

### 8-hour check-ins

At 8-hour check-ins the first prestige took 7.2 days against 8 (10% headroom). Read visit
by visit (`-trace`, which now also lists required buildings still to start and an unbuilt
wonder), about one visit in five found the age finished except the wonder: its bank was
short, and the stone or iron to fill it sat in the stores, bought by a plan trade while the
player was away. The plan's wonder item waited for a full bank, overflow only takes what a
cap cuts off, and a trade into a full store stops. So the player lost eight hours to a
`wonder collect`. That is the game, not the bot: no plan item could have done it.

- **A planned wonder pays its bank from stock** (`game/plan.go`). Its price is what the bank
  still lacks; once what is held after the reservations above covers all of it, the plan
  banks it and starts the wonder. While it waits it reserves nothing (a wonder is 40 price
  units; holding that back stalls the plan for hours, and made 1-hour and 3-hour check-ins
  slower when tried). A part bigger than a full store still fills by deposits and overflow.
- **The idle bot** now plans its wonder last before the advance, so it takes what the age's
  other items leave; keeps the market-only inputs of its storage and required buildings
  topped up with a `plan trade` (the Atomic Age's vaults cost iron nothing makes there, and
  the vaults, then the bunkers behind them, waited a visit for it); and counts a planned
  trade as income once a trade building stands.

First prestige, median of three seeds (min-max):

| Check-in | Target | Before | After |
|---|---|---|---|
| 1 h | 3.5 d | 2.8 d (2.7-3.1) | 2.9 d (2.9-3.2) |
| 3 h | 5 d | 3.9 d (3.7-4.0) | 3.9 d (3.8-3.9) |
| 8 h | 8 d | 7.2 d (7.1-7.5) | 6.5 d (6.5-7.1) |

8-hour headroom goes from 10% to 23%. Measured locally before the Renaissance change, the
wonder fix took the 8-hour median from 7.4 to 6.8 days and the market top-ups to 6.5; the
Renaissance change costs the 1-hour player about a tenth of a day. At 8 hours the Atomic
Age now takes 11.7 h (was 1.2 d); the Industrial and Electric Ages take longer (19.2 h and
23.3 h, were 14.7 h and 21.3 h), and those are where the next hours are. At 1 hour the
Stone Age takes 5.6 h (was 2.8 h), most likely because the plan now advances out of the
Primitive Age mid-absence (the wonder no longer waits for a visit) and the Stone Age opening
it leaves waits for the next visit to recruit. The 1-hour total barely moved, so it is left
for later.

## Appendix: Check-in play: build plan, overflow, storage (2026-09-27)

The idle appendix below found that visits to the first prestige hardly depended on the
check-in interval, because storage capped what a visit could achieve. This change gives
the check-in player three tools and measures them.

**Build plan** (`plan`, `game/plan.go`). An ordered list of builds, techs, trades and an
advance that the engine walks after production every tick and during offline catch-up.
Items are paid when they start. The one real design choice was what a blocked item does to
the items after it:

- strict order (stop at the first unaffordable item) lets one big item stall the whole
  plan for hours;
- free skipping lets cheap items further down eat the resources the top item is saving
  for, forever;
- **skip with reservation** (chosen): a waiting item doesn't block, but it holds back its
  next price, and a later item only starts from what is left. The order is the priority,
  and resources the top items don't need still get spent. An item that money won't unblock
  holds nothing: a price over the cap, a resource the current income won't bring within a
  day, a wonder bank that isn't full, the next age's building before the advance. Without
  that exemption one item waiting on market-only stone held back the gold of everything
  below it.

Techs start in plan order (a research queue). Trade items hold back what they will sell
and sell once the market's supply pressure is under 2%, so they trade about once a minute
at within 0.6% of the market rate instead of losing up to 30% by trading every tick. An
advance item advances at its place in the plan, so the items below it can't spend what the
requirements count. Copies the plan finishes are staffed from idle workers.

**Offline** now runs in one-minute steps (`OfflineStepTicks`): production at 50% up to the
caps with overflow, construction and research advance, then the plan starts what the step
paid for. With nothing planned or under construction it pays exactly the old lump sum. A
day away with a plan takes about 10 ms. Construction and research used to stand still
offline.

**Wonder overflow** (on by default, `wonder overflow off`) banks what a cap cuts off into
the current wonder, up to its need. It only takes production that would be lost.

**Storage.** The time clause of Law 1 (1.5 hours of typical income) raised the storage per
copy of the Stone, Bronze, Iron, Classical, Medieval, Renaissance, Colonial, Industrial,
Victorian and Information Ages (from +2.5% for the Keep to +440% for the Renaissance Vault;
the numbers are in the CHANGELOG). The other eleven ages already kept it.

### Results

First prestige, median of three seeds (min-max). After: the nightly's `idle` scenario on a
GitHub runner (run 36347171804). Before: PR #125's idle bot on master.

| check-in every | before | after | target | visits after |
|---|---|---|---|---|
| 1 h | 7.3 d (7.0-8.2) | 2.8 d (2.7-3.1) | 3.5 d | 67 |
| 3 h | 17.4 d (17.2-18.1) | 3.9 d (3.7-4.0) | 5 d | 31 |
| 8 h | 44.5 d (44.1-45.5) | 7.2 d (7.1-7.5) | 8 d | 21 |

Time per age (median) at 3-hour check-ins, before → after: Primitive 12 h → 2.0 h, Stone
1.2 d → 6.0 h, Bronze 1.2 d → 4.8 h, Iron 1.4 d → 5.3 h, Classical 1.2 d → 8.8 h, Medieval
14.5 h → 6.3 h, Renaissance 22.8 h → 5.9 h, Colonial 20 h → 5.9 h, Industrial 2.4 d → 11.1 h,
Victorian 2.0 d → 8.0 h, Electric 2.7 d → 10.0 h, Atomic 2.7 d → 16.8 h. At 8 hours every
age takes one or two visits (Stone 8.8 h, Classical 16.2 h, Atomic 1.1 d); visits to the
prestige now fall with the interval (67, 31, 21) instead of staying near 150.

What each tool is worth, taking one away at a time (local runs, same seeds; `-no-plan`,
`-no-overflow`, and a build with the old storage numbers):

| | 1 h | 3 h | 8 h |
|---|---|---|---|
| all three | 2.8 d | 4.0 d | 7.4 d |
| no plan | 6.0 d | past 10 d | past 16 d |
| no overflow | 3.2 d | 4.3 d | 9.4 d |
| old storage | 3.1 d | 4.5 d | 8.3 d |

The plan does most of the work, overflow matters most at long intervals (more of the day's
income meets a full store), and storage adds about a tenth everywhere.

**Nothing overshot.** The idle player stays slower than the greedy one at every interval
(2.1 days for the greedy bot). Greedy pacing stays inside 0.5x-2x in every age, but the
storage raise sped the middle ages up: the Renaissance went from 0.72x to 0.57x of its
target and the first prestige from 2.4 to 2.1 days. The Renaissance Vault sits right on the
1.5-hour line (gold income jumps there), so if a later change pushes the Renaissance under
0.5x, the levers are that age's gate or its gold rates rather than its storage.

## Appendix: Idle play and the Iron Age gold trap (2026-09-27)

### The gold trap

The trading post cost 32K stone, 15K iron and 8.8K gold. It was the Iron Age's only gold
producer and its only trade building, so a player who skipped the optional Bronze Age
market could not get gold in the Iron Age at all (a smoke seed spent 2.5 days there). It
now costs stone and iron only; the Payback Rule sets its rate from the smaller price
(48.8 gold/tick, was 65.2) and gold's Iron Age price level falls from 8,000 to 7,500.
Rule 5 of the Gate Covenant now checks sourcing from a cold start (see above), and with
the trading post fixed it finds no other trap of the kind. Greedy pacing moved by at most
0.1x (Iron 2.7 h to 2.6 h, Classical 3.8 h to 4.1 h, Colonial 4.3 h to 5.0 h, seeds 1-8).

### Idle play

The target player checks in a few times a day. The smoke suite's `idle` style models one,
with the game running between visits. Time to the first prestige (Modern Age), seeds 1-3,
median (min-max), against about 2.4 days for the greedy bot:

| check-in every | before (one decision per visit) | after (a visit spends everything) | visits |
|---|---|---|---|
| 1 h | 27.0 d (25.9-27.2) | 7.3 d (7.0-8.2) | ~175 |
| 3 h | 76.3 d (71.2-76.3) | 17.4 d (17.2-18.1) | ~140 |
| 8 h | not reached in 83 d (stuck in the Classical Age) | 44.5 d (44.1-45.5) | ~134 |

Time per age at 3-hour check-ins, after: Primitive 12 h (48x the target), Stone 1.2 d,
Bronze 1.2 d, Iron 1.4 d, Classical 1.2 d, Medieval 14.5 h (3.2x), Renaissance 22.8 h,
Colonial 20 h (2.9x), Industrial 2.4 d (7.2x), Victorian 2.0 d, Electric 2.7 d, Atomic
2.7 d (5.4x). At 1-hour check-ins the Medieval through Colonial Ages are inside the band
(1.4x-1.5x) and the rest 2x-4x. No run soft-locked at any interval.

**What was the harness.** The idle bot made one decision per visit: one producer, one
storage copy and one trade every three hours. A visit is now rounds of decisions until
nothing more is worth doing, with storage bought for what comes in before the next visit
and whatever would be lost at a cap banked into the wonder or traded (`Bot.CheckIn`).

**What was the game.** `build` refused a second copy of a storage building while one was
under construction, so a player got one storage copy per visit (`build <key> N` queued any
number). Fixed: only unique buildings refuse.

**What still is the game.** The number of visits to the first prestige hardly depends on
the interval (about 130-180), so a visit's progress is capped, and the cap is storage. At
3-hour check-ins (seed 1, 149 visits) three in four of the needed resources that
something produces are at their cap when the player arrives, and a store that filled did
so in a median of 20-50 minutes (5-15 minutes from the Renaissance to the Victorian Age). Production is sized so producers repay in the payback
time (Law 3); storage is sized so the last required copy costs half of it (the Gate
Covenant). Nothing sizes storage to hours of production, so a player away for three hours
loses most of what their economy makes, and the gap to the targets does not close in the
later ages. The Industrial Age is worst: its buildings need stone and coal that nothing in
the age produces, so everything comes through the market, bounded by what the other stores
held.

Also: research runs one tech at a time with no queue, so a check-in player starts one tech
per visit even when a tech takes minutes. Techs rarely gate an advance, but it is the same
shape of problem.

**Recommendation.** Keep the targets (they describe an attentive player). To let a
check-in player keep up, give them a way to put the hours between visits to work that
doesn't need storage: a planned-build queue that pays for each copy when it starts rather
than when it is queued (the usual idle-game answer), or a wonder bank that takes overflow
automatically. Raising storage to hold a check-in's worth of production would also work
but removes the resource pressure storage exists for. A research queue is a smaller,
separate fix. None of this is in this change.

## Appendix: Pacing rebalance (2026-09-27)

Time in each age for the smoke bot, median (min-max) over seeds 1-5 at 1x, against the
Law 2 targets. Before: master at 62eefac, run with no age timeout. After: this change.

| Age | Target | Before | After |
|---|---|---|---|
| Primitive | 15 min | 5.8 h (5.7-6.0) | 18 min (17-18) |
| Stone | 45 min | 23.5 h (22.6-24.0) | 1.1 h (1.1-1.1) |
| Bronze | 1.5 h | 1.0 d (22.4 h to 1.1 d) | 2.3 h (2.2-2.3) |
| Iron | 2.5 h | 5.0 d (4.9-5.4) | 2.5 h (2.4-2.9) |
| Classical | 3.5 h | 16.1 d (16.0-18.0) | 4.1 h (4.0-4.5) |
| Medieval | 4.5 h | 7.9 d (7.5-8.7) | 2.9 h (2.6-3.2) |
| Renaissance | 6 h | 46.9 d (45.4-59.7) | 4.3 h (3.3-5.7) |
| Colonial | 7 h | never (stalled at 83 d) | 5.1 h (4.1-5.5) |
| Industrial | 8 h | - | 8.1 h (7.6-8.5) |
| Victorian | 9 h | - | 7.7 h (7.2-9.3) |
| Electric | 10 h | - | 8.9 h (7.9-9.4) |
| Atomic | 12 h | - | 9.5 h (9.4-9.8) |
| Modern | 12 h | - | 13.4 h (12.7-14.0) |
| Information | 14 h | - | 18.7 h (18.6-19.8) |
| Digital | 16 h | - | 18.7 h (16.3-19.1) |
| Cyberpunk | 18 h | - | 1.0 d (4.3 h to 1.1 d) |
| Fusion | 20 h | - | 17.7 h (16.6-18.0) |
| Space | 22 h | - | 1.9 d (1.8-2.0) |
| Interstellar | 24 h | - | 1.8 d (1.7-1.9) |
| Galactic | 24 h | - | 1.3 d (1.3-1.4) |

Every age through Atomic is inside 0.5x-2x of its target; the first prestige (Modern Age)
comes at about 2.4 days on seeds 1-10. Space (2.1x) is the one age still outside the band.

### Rules and constants (config/pacing.go)

| Constant | Value | Meaning |
|---|---|---|
| `PaybackDivisor`, `PaybackEpochExponent` | 16, 1.25 | payback = target × epochProgress^1.25 / 16 |
| `BuildTimeDivisor` | 6 | build time ≤ target / 6 (wonders too) |
| `StorageBuildTimeDivisor` | 48 | storage build time ≤ target / 48 |
| `ResearchTimeDivisor` | 8 | research time ≤ tech age target / 8 |
| `ExchangeFee` | 0.2 | market keeps 20% at parity |
| `WonderPriceUnits` | 40 | wonder price in price units of its age |

### Hand-set numbers

| What | Before | After | Why |
|---|---|---|---|
| Stone Age gate (raw) | 8K food, 5.2K wood, 1.4K knowledge, 20 huts | 500, 500, 75, 10 huts | 15-minute Primitive Age |
| Bronze Age gate (raw) | 15K food, 8K wood, 4K stone, 5K knowledge, 40 longhouses | 2K, 4K, 2K, 750, 15 longhouses | 45-minute Stone Age |
| Classical gate | 8 agoras, 5 trading posts | 12, 10 | Iron Age ran at half its target |
| Renaissance gate (raw) | 2K steel, 25K faith | 500, 4.6K | no Medieval steel building; faith is flow |
| Colonial gate | 5 exchanges, 3 universities, 5 art studios | 8, 8, 8 | Renaissance ran at under half its target |
| Industrial gate | 5 plantations, 8 ports | 8, 10 | Colonial ran at half its target |
| Interstellar gate | 20 orbital habitats | 15 | market-only steel made the last copies most of the gate |
| Gathering camp, forager post | 0.5, 1.0 food | 1.0, 1.5 | food is the Primitive bottleneck |
| Story circle … library | 0.05, 0.1, 0.2, 0.4, 0.8 knowledge | 0.2, 0.6, 2.0, 1.6, 3.2 | knowledge isn't a construction resource before the Medieval Age |
| Steel Forging | 0.1 steel/tick | 0.25 | the only Medieval steel source |
| Sacred Grove, Great Monolith food | 5K, 10K | 500, 1.5K | flow part of a wonder |
| Sistine Chapel faith | 6M | 20K | 6M faith was most of the 47-day Renaissance |
| Stellar Cradle | 940T uranium | no uranium | nothing in the Fusion Age produces it |

The Iron Age gate keeps its 80K food and 20K knowledge: the Stone Era harbinger's prices are
derived from it.

## Appendix: Pacing follow-ups (2026-09-27)

Time in each age for the smoke bot (Harbinger ignored), median (min-max) over seeds 1-5,
playing to a Quantum Age prestige. The Space Age was the one age out of band.

| Age | Target | Before | After |
|---|---|---|---|
| Fusion | 20 h | 17.7 h (16.8-18.1) | 16.8 h (16.3-16.9) |
| Space | 22 h | 2.1 d (1.8-2.2) | 1.2 d (1.1-1.2) |
| Interstellar | 24 h | 1.8 d (1.7-1.9) | 1.5 d (1.5-1.6) |
| Galactic | 24 h | 1.2 d (1.2-1.3) | 1.3 d (1.3-1.3) |

Every age from the Primitive to the Galactic is now inside 0.5x-2x; the ages before Fusion
moved by at most 0.2x (the bot's trading change below touches every age).

**Why the Space Age was slow.** Titanium unlocks there, nearly every Space building costs
it, and nothing made it. That alone is only the market's 20% fee, but the smoke bot, whose
slowest target was titanium, traded the plasma and electricity it was saving for the next
producer into titanium as fast as they came in, so it hardly invested in the Space Age at
all. The bot no longer sells what its most wanted purchase is saved for (unless that
resource sits at its cap), and the Orbital Refinery now makes titanium (it made dark matter,
which unlocks an age later) without costing it (its 88T titanium went to plasma), so the age
has a producer that can start the supply. The bot change did most of it: 2.3x to 1.3x with
the refinery either way.

### Hand-set numbers

| What | Before | After | Why |
|---|---|---|---|
| Orbital Refinery | 4 dark matter/tick; 88T titanium, 44T plasma, 110T electricity (typed) | titanium (Law 3 rate); 110T plasma, 110T electricity | a Space Age titanium producer that doesn't need titanium first |
| Sistine Chapel (typed) | 9.9M stone, 7M gold | 8.5M stone, 8.5M gold (24M each after sizing) | 27M stone was over the 25.5M Renaissance storage |
| World Simulation (typed) | 50T steel, 6T electricity | 30T steel, 18T electricity (34T, 20T after sizing) | 54T steel was over the 46.8T Digital storage |
| Iron Forged milestone | 40 coal | 40 iron | coal is locked until the Renaissance |
| Harbinger Appease | 15% of the passage storage | 1/4 of `FlowIncome` over the thread's ages | faith is a flow resource; see epochs.md |

## Appendix: Gate Covenant fixes (2026-09-26)

Every number changed to make every advance pass the Gate Covenant. Requirement counts are
the values the game uses (after `normalizeAgeRequirements`); prices are normalized.

### Age requirements

| Advance to | Before | After | Why |
|---|---|---|---|
| Bronze | 50 longhouse | 40 longhouse | copy #50 cost 43.9K wood vs 40K max Stone storage |
| Classical | 15 barracks, 5 market | 15 hunting lodge, 5 trading post | barracks and market are Bronze Age buildings (age lock) |
| Medieval | 20 library, 30 barracks | 15 library, 15 military academy | library copy #20 at 1.93×; barracks is Bronze Age (copy #30 cost 1.6× Bronze storage anyway) |
| Renaissance | 15 market | 10 guildhall | market is Bronze Age; guildhall copy #15 over 2× |
| Industrial | 5 market garden | removed | market garden is Renaissance; the Colonial food building (plantation, 5) was already required |
| Victorian | 1.1M oil | removed | nothing produces oil before the Victorian oil derrick |
| Electric | 20 steam turbine, 15 steel mill | 10 steam turbine, 10 bessemer plant | turbine copy #20 over max storage; steel mill is Industrial |
| Atomic | 20 electric arc furnace, 20 steam works | 15 electric arc furnace, 15 power station | furnace at 1.35×; steam works is Victorian |
| Modern | 30 nuclear reactor, 30 bunker complex | 15, 15 | copy #30 cost 1.4T steel / 2T stone vs 598B storage |
| Information | 50 think tank, 60 oil refinery | 20, 15 | copy #50/#60 cost 30-190× storage |
| Digital | 30 server farm, 80 media center, 30 innovation hub | 10, 15, 15 | up to 3,000× over |
| Cyberpunk | 80 AI research lab, 80 data center, 50 neural grid | 15, 15, 15 | up to 4,000× over |
| Fusion | 50 augmentation foundry, 80 arcology pod, 50 black market | 15, 25, 15 | up to 300× over |
| Space | 80 fusion reactor, 60 fusion reactor array, 50 plasma command | 10, 10, 10 | up to 6,000× over |
| Interstellar | 80 launch complex, 60 orbital habitat, 50 solar collector array | 10, 20, 10 | up to 4,600× over |
| Galactic | 80 warp drive plant, 60 generation ship, 50 orbital refinery | 15, 30, 15 antimatter forge | orbital refinery is Space Age; the rest up to 11,000× over |
| Quantum | 80 stellar exchange, 100 antimatter forge, 120 Dyson sphere habitat | 15, 15 stellar metallurgy, 30 | antimatter forge is Interstellar; up to 190,000× over |
| Transcendent | 500 reality academy, 300 reality forge, 200 probability war room | 20, 15, 15 | copy #500 cost 10^30× storage |

### Storage (cosmic era out of band: median first copy was 13-57% of max storage)

| Building | Before (per copy / max) | After (per copy / max) |
|---|---|---|
| Stellar Vault (Interstellar) | +500T / 12.5Q | +2Q / 50Q |
| Galactic Vault (Galactic) | +2Q / 50Q | +20Q / 500Q |
| Quantum Vault (Quantum) | +10Q / 250Q | +200Q / 5,000Q |

### Prices (sourcing)

| Building (age) | Removed from price | Why |
|---|---|---|
| Smithy (Bronze) | 850 iron | the only Bronze iron source; the age grants 30 iron |
| Mill (Renaissance) | 850K steel | every steel producer cost steel; only a 0.1/tick tech fed it |
| Ironworks, Smelter (Iron) | 6.4K, 4.2K coal | coal unlocks in the Renaissance |
| Forge (Classical) | 21K coal | same |
| Nuclear Extraction Plant (Electric) | 150M uranium | uranium unlocks in the Atomic Age |
| Crypto Exchange, Cyber Shrine, Logistics Hub (Digital) | 590B, 600B, 640B crypto | crypto unlocks in Cyberpunk |
| Monument of Ages (Modern) | 1.4M titanium | titanium unlocks in the Space Age |

### Exchange and mechanics

| Change | Before | After |
|---|---|---|
| gold ↔ data exchange | from Information Age | from Modern Age (where data unlocks and ten buildings, think tank included, cost it) |
| Storage upgrades (stash → storage pit, ...) | offered on every advance | never offered; storage never transforms |
| Upgrades into a capped building | ignored MaxCount | stop at MaxCount |
| Market exchange needs | a market or a port | any trade-lineage building (upgrading markets no longer shuts the exchange) |

