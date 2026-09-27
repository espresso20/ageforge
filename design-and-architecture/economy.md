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

### Law 2 — The Build Time Curve
> The time required to save up for and build the next building should follow a predictable,
> intentional curve across the full game. It should never feel instant and never feel impossible.

Target build times (approximate, at normal play pace):

| Age Tier | Target build time per building |
|----------|-------------------------------|
| Primitive | 1–5 min |
| Stone | 5–15 min |
| Bronze / Iron | 15–45 min |
| Classical → Renaissance | 1–4 hrs |
| Colonial → Industrial | 4–12 hrs |
| Victorian → Atomic | 12–48 hrs |
| Modern → Digital | 2–7 days |
| Cyberpunk → Fusion | 1–2 weeks |
| Space → Galactic | 2–4 weeks |
| Quantum | Prestige milestone |

**Prestige loop target: several weeks of real calendar time.**
Prestige is a major milestone, not a reset you do every few hours. The first prestige loop
should feel like an accomplishment. Subsequent loops are faster due to prestige bonuses,
eventually allowing players to reach and exceed their previous age ceiling.

### Law 3 — The Production Curve
> Total resource production rate must stay on a smooth exponential curve. At no point should
> a player's production rate plateau or drop. Each age tier should provide a meaningful
> production multiplier over the previous tier.

Target production multiplier per age advance: approximately **5–10×** total output.

This means:
- A player entering Bronze Age should produce roughly 5–10× more resources per tick than
  they did at peak Stone Age.
- Building costs at each age should be calibrated to this multiplier — not guessed.

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

Building costs should be **derived from expected production rate**, not guessed.

```
building_cost = production_rate_at_age × target_build_time_in_ticks
```

Where:
- `production_rate_at_age` = expected total resource production for the relevant resource
  at the point the player first encounters this building
- `target_build_time_in_ticks` = build time target for this age tier (from Law 2)
- `tick_interval` = ~1,500ms (1.5 seconds at base speed)

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
4. **Sourced:** every resource the gate asks for, directly or in a required building's price,
   has a way in by the end of A that does not need that resource first: a building that
   doesn't cost it, hand gathering (food, wood, stone, through the Medieval Age), a market
   exchange, or techs whose flat output alone covers the whole amount within 48 hours at 1x.
   A producer that costs its own output (the Bronze Age smithy and iron, the Renaissance mill
   and steel) doesn't count until something else supplies the first batch.

The same sourcing rule applies to every building on its own: nothing may cost a resource
with no source in the building's own age (coal before the Renaissance, crypto before
Cyberpunk). Those buildings were dead content.

`smoke.StaticGates` implements all of this and `TestGateCovenant` (smoke/static_test.go)
fails `go test ./...` when a balance change breaks it, so it runs in CI on every PR.

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
- [ ] Derive all building base costs using: `cost = production_rate_at_age × target_build_ticks`
- [ ] Verify Storage Covenant (Law 1) for all 22 age transitions
- [ ] Tune worker food costs and output multipliers against the build time curve (Law 2)

### Phase 4 — UI
- [ ] Update population panel to show current-tier workers prominently, legacy collapsed
- [ ] Update resource rate breakdown to show worker contribution separately from building base
- [ ] Worker assignment UI uses domain name (not class name) to avoid churn on age advance

---

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

