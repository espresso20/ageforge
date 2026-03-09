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

| Age | Storage Building | Capacity per building | Max Count |
|-----|-----------------|----------------------|-----------|
| Primitive | Stash | +300 all | 50 → 15,000 max |
| Stone | Storage Pit | +600 all | TBD |
| Bronze | Warehouse | +2,000 all | TBD |
| Iron | Granary (food only) | +5,000 food | TBD |
| ... | ... | ... | ... |

> **TODO:** Audit all storage buildings to verify Law 1 is satisfied for every age transition.
> The Primitive→Stone transition is the first known violation (hut costs exceed stash cap).

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
