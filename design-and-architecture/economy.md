# AgeForge — Economy Design

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

### Law 1 — The Storage Covenant
> At any stage of the game, a player's maximum possible storage for a resource **must be
> ≥ 2× the cost** of the most expensive building they are expected to build at that stage.

Violation of this law makes buildings literally unbuildable (costs exceed what can ever be
accumulated). This was the root cause of hut #68+ being impossible with 50 stashes.

For age gates this is made exact by the **Gate Covenant** (see Storage Design): the last
required copy of every required building costs at most half the storage buildable by then,
with no discounts, and a unit test enforces it.

### Law 2 — The Pacing Curve
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
  advance needs them), and 1/48 for storage buildings, which queue one copy at a time and
  are bought many times an age.
- **Research time** is at most 1/8 of the tech's age target, so the handful of techs an age
  offers fit in it one at a time.
- The per-building wait falls out of these: with a dozen producers each repaying in the
  payback time, the next copy is minutes away early and an hour or two late.

**Measuring it.** The smoke harness plays a greedy bot on fixed seeds and reports the time
spent in each age; `smoke/targets.go` holds the same table (a test keeps the two equal) and
`-pacing enforce` fails a run with a first-cycle age outside 0.5x-2x of it. CI runs enforced:
the per-PR fast tier grades the Primitive and Stone Ages, the nightly every age to the first
prestige. The bot has to play like a reasonable person for this to
measure the game rather than the bot (see the bot's strategy comment in `smoke/bot.go`).

### Law 3 — The Payback Rule
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
  1/4 in the Renaissance, 1/3 in the Victorian and 2/3 in the Space Age.

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

### Law 4 — The Coupling Law
> Every production building has a **worker capacity** (how many workers of a given type it
> can employ) and a **worker output** (production per assigned worker per tick). Buildings
> produce at **20% base rate without workers**. Full production requires workers.

This means:
- Housing → Population → Workers → Production is a real chain, not two parallel systems.
- Players must invest in housing to unlock production potential.
- Workers and buildings are complementary, not competing.
- The "idle" base rate (20%) means the game keeps running if you step away and forget to
  assign, but assigned workers provide a 5× multiplier — a meaningful optimization.

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
- **Storage buildings**: 1.15–1.20 (slow scale, you want many of these)
- **Production buildings**: 1.25–1.35 (moderate, you want several but not infinite)
- **Housing buildings**: 1.10–1.15 (keep cheap — housing is a prerequisite, not a sink)
- **Specialist buildings** (research, military): 1.35–1.50 (expensive per unit, few needed)
- **Late-game buildings** (space, quantum): 1.45–1.60 (steep is fine, everything takes days)

---

## Storage Design

Storage is the **resource pressure** mechanism. It keeps the game interesting — players can't
just walk away for a year and come back to infinite resources.

### Rules
- Every resource has a storage cap, set by the sum of storage buildings.
- Storage buildings have **MaxCount** (the only building type that does).
- At each age tier, the available storage buildings must satisfy **Law 1** — their maximum
  combined capacity must be ≥ 2× the most expensive building cost in that tier.
- Storage should scale roughly with production: a player at Bronze Age should have roughly
  5–10× more storage than at Primitive Age.

### Storage Buildings by Age

Every age has one storage building, each capped at **25 copies** (the stash at 50). Storage
never transforms or upgrades: the stashes you built in the Primitive Age still count in the
Quantum Age, and because of the age lock they can never be rebuilt later. The live numbers
are in `config/buildings.go`; the smoke report's "Tightest requirement per advance" table
shows how much headroom each advance has.

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
5. **Sourced:** every resource the gate asks for, directly, in a required building's price
   or in the wonder's, has a way in **in A** that does not need that resource first: a
   building of A that doesn't cost it, hand gathering (food, wood, stone, through the
   Medieval Age), a market exchange (a listed pair, or parity when the resource is a
   construction resource of A), or techs whose flat output alone covers the whole amount
   within 48 hours at 1x. Older ages' producers only count through the market's parity:
   the age lock stops you building more of them, and their output is sized to an older
   age's prices (the Stellar Cradle cost 940T uranium in the Fusion Age, where only old
   Atomic Age mines made any). A producer that costs its own output (the Bronze Age smithy
   and iron, the Renaissance mill and steel) doesn't count until something else supplies
   the first batch.
6. **Flow within the target:** every flow resource the gate asks for (requirement, required
   copies and wonder together) is made within A's pacing target at `FlowIncome`, or the rest
   costs at most **10 price units** of A at the market's listed pairs (gold → food, gold →
   culture). Faith has no market, so a faith requirement must fit A's own output: the
   Renaissance's old 44K faith at about 1.5 faith/tick and the Sistine Chapel's old 6M faith
   both fail it.

The same sourcing rule applies to every building on its own: nothing may cost a resource
with no source in the building's own age (coal before the Renaissance, crypto before
Cyberpunk). Those buildings were dead content. The last age's wonder, which no gate
requires, is checked the same way.

`smoke.StaticGates` implements all of this and `TestGateCovenant` (smoke/static_test.go)
fails `go test ./...` when a balance change breaks it, so it runs in CI on every PR.
`TestGateCovenantCatchesBrokenGates` feeds it the old broken numbers and checks each one is
caught.

**Gate size.** With Law 3 the time an age takes follows mostly from its gate (required
buildings plus wonder, in price units) and the stock of older buildings. Gates that were
far smaller than their neighbours' finished in a fraction of the target and were enlarged
(Classical, Colonial, Industrial); the steel-heavy Space Age gate was trimmed, since steel
and titanium came only from the market there, and the Space Age's Orbital Refinery now
makes titanium.

**Levers, in order of preference.** When a gate breaks the covenant:

1. If it names an older age's building, retarget it to the same lineage's building in A
   (the lowest tier there, which is what the older building upgrades into, so upgraded
   copies count).
2. If the storage buildings of that age are out of band with the age's own prices, raise
   their per-copy storage. "In band" means the median first-copy price is 3–8% of the
   age's max storage, which is where Classical through Space sit.
3. Otherwise lower the count to the largest that fits, rounded down to a multiple of 5
   (exact below 10).
4. For a sourcing failure, drop the unobtainable resource from the price of the one
   producer that bootstraps it (or open an exchange in the age the resource unlocks),
   rather than inventing a new building.

Counts are the usual lever because the normalized cost curves (1.15 per copy, 1.13 for
housing and storage) made late copies explode: copy #80 costs 62,000× copy #1, so the
old late-game counts of 50–500 were never reachable at any storage.

---

## MaxCount Policy

| Building Category | MaxCount | Reason |
|-------------------|----------|--------|
| Storage buildings | Yes — hard cap | Unlimited storage breaks resource pressure |
| Wonders | Yes — 1 | Unique by design |
| Production buildings | **No** | Geometric cost scaling is the natural cap |
| Housing buildings | **No** | Geometric cost scaling + population needs are the natural cap |
| Military buildings | **No** | Scale naturally |
| Research buildings | **No** | Scale naturally |

Players who leave gathering_camp running for 3 months and accumulate a sextillion wood are
playing correctly. That's the idle game working as designed.

---

## Implementation Phases

### Phase 1 — Data Model
- [ ] Add `WorkerDomain string` and `WorkerCapacity int` to `BuildingDef` in config/buildings.go
- [ ] Add age-tiered worker class entries to config/villagers.go (see workers.md)
  - Each class: Domain, UnlockAge, FoodCost, OutputMultiplier, Name
- [ ] Update all ~80 building definitions with their domain + capacity values

### Phase 2 — Engine
- [ ] Update VillagerManager to handle multi-tier workers per domain
- [ ] Update ResourceManager production calculation:
  `rate = building_base_rate × (0.20 + 0.80 × assigned/capacity)`
- [ ] Update housing pop values (hut: +3 → +10, all tiers rescaled)
- [ ] Remove MaxCount from all production/housing buildings in config/buildings.go

### Phase 3 — Balance Numbers
- [x] Derive production from prices and the age targets (the Payback Rule, Law 3, 2026-09-27)
- [x] Verify Storage Covenant (Law 1) for all 22 age transitions (Gate Covenant, 2026-09-26)
- [ ] Tune worker food costs against the new rates (food producers keep hand-set rates)

### Phase 4 — UI
- [ ] Update population panel to show current-tier workers prominently, legacy collapsed
- [ ] Update resource rate breakdown to show worker contribution separately from building base
- [ ] Worker assignment UI uses domain name (not class name) to avoid churn on age advance

---

## Appendix — Pacing rebalance (2026-09-27)

Time in each age for the smoke bot, median (min–max) over seeds 1–5 at 1x, against the
Law 2 targets. Before: master at 62eefac, run with no age timeout. After: this change.

| Age | Target | Before | After |
|---|---|---|---|
| Primitive | 15 min | 5.8 h (5.7–6.0) | 18 min (17–18) |
| Stone | 45 min | 23.5 h (22.6–24.0) | 1.1 h (1.1–1.1) |
| Bronze | 1.5 h | 1.0 d (22.4 h–1.1 d) | 2.3 h (2.2–2.3) |
| Iron | 2.5 h | 5.0 d (4.9–5.4) | 2.5 h (2.4–2.9) |
| Classical | 3.5 h | 16.1 d (16.0–18.0) | 4.1 h (4.0–4.5) |
| Medieval | 4.5 h | 7.9 d (7.5–8.7) | 2.9 h (2.6–3.2) |
| Renaissance | 6 h | 46.9 d (45.4–59.7) | 4.3 h (3.3–5.7) |
| Colonial | 7 h | never (stalled at 83 d) | 5.1 h (4.1–5.5) |
| Industrial | 8 h | – | 8.1 h (7.6–8.5) |
| Victorian | 9 h | – | 7.7 h (7.2–9.3) |
| Electric | 10 h | – | 8.9 h (7.9–9.4) |
| Atomic | 12 h | – | 9.5 h (9.4–9.8) |
| Modern | 12 h | – | 13.4 h (12.7–14.0) |
| Information | 14 h | – | 18.7 h (18.6–19.8) |
| Digital | 16 h | – | 18.7 h (16.3–19.1) |
| Cyberpunk | 18 h | – | 1.0 d (4.3 h–1.1 d) |
| Fusion | 20 h | – | 17.7 h (16.6–18.0) |
| Space | 22 h | – | 1.9 d (1.8–2.0) |
| Interstellar | 24 h | – | 1.8 d (1.7–1.9) |
| Galactic | 24 h | – | 1.3 d (1.3–1.4) |

Every age through Atomic is inside 0.5x–2x of its target; the first prestige (Modern Age)
comes at about 2.4 days on seeds 1–10. Space (2.1x) is the one age still outside the band.

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

## Appendix — Pacing follow-ups (2026-09-27)

Time in each age for the smoke bot (Harbinger ignored), median (min–max) over seeds 1–5,
playing to a Quantum Age prestige. The Space Age was the one age out of band.

| Age | Target | Before | After |
|---|---|---|---|
| Fusion | 20 h | 17.7 h (16.8–18.1) | 16.8 h (16.3–16.9) |
| Space | 22 h | 2.1 d (1.8–2.2) | 1.2 d (1.1–1.2) |
| Interstellar | 24 h | 1.8 d (1.7–1.9) | 1.5 d (1.5–1.6) |
| Galactic | 24 h | 1.2 d (1.2–1.3) | 1.3 d (1.3–1.3) |

Every age from the Primitive to the Galactic is now inside 0.5x–2x; the ages before Fusion
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

## Appendix — Gate Covenant fixes (2026-09-26)

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
| Information | 50 think tank, 60 oil refinery | 20, 15 | copy #50/#60 cost 30–190× storage |
| Digital | 30 server farm, 80 media center, 30 innovation hub | 10, 15, 15 | up to 3,000× over |
| Cyberpunk | 80 AI research lab, 80 data center, 50 neural grid | 15, 15, 15 | up to 4,000× over |
| Fusion | 50 augmentation foundry, 80 arcology pod, 50 black market | 15, 25, 15 | up to 300× over |
| Space | 80 fusion reactor, 60 fusion reactor array, 50 plasma command | 10, 10, 10 | up to 6,000× over |
| Interstellar | 80 launch complex, 60 orbital habitat, 50 solar collector array | 10, 20, 10 | up to 4,600× over |
| Galactic | 80 warp drive plant, 60 generation ship, 50 orbital refinery | 15, 30, 15 antimatter forge | orbital refinery is Space Age; the rest up to 11,000× over |
| Quantum | 80 stellar exchange, 100 antimatter forge, 120 Dyson sphere habitat | 15, 15 stellar metallurgy, 30 | antimatter forge is Interstellar; up to 190,000× over |
| Transcendent | 500 reality academy, 300 reality forge, 200 probability war room | 20, 15, 15 | copy #500 cost 10^30× storage |

### Storage (cosmic era out of band: median first copy was 13–57% of max storage)

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

