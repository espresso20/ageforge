# Multiplier / Bonus System — Investigation & Proposed Redesign

**Status:** Investigation / design proposal (no implementation yet)
**Author:** espresso + Claude
**Trello:** Refactor — "INVESTIGATION: unified multiplier registry/resolver" (`hJPR8YJX`)

---

## 1. Motivation

Every system in AgeForge produces or consumes multipliers — research, prestige,
epoch events, wonders, milestones, catastrophes, morale, building effects. The
machinery that tracks them is alpha-era layover: it grew one source at a time,
and there is **no single contract**. The symptom surfaced while adding morale to
the Active Multipliers panel — morale had to be wired in as *yet another* special
case alongside `research.GetBonus()`, `prestige.GetBonuses()`, the
`permanentBonuses` map, `SpeedMultiplier`, and a per-wonder loop.

The system is **functionally mostly correct**, but it is fragmented in a way that
is already producing player-visible bugs and makes every new source a
cross-cutting edit. This doc maps the current reality and proposes a unified
in-process **Modifier Resolver**.

> Note on framing: this is a single Go process, not "microservices." The right
> abstraction is an in-process registry/resolver — a concrete API that sources
> *contribute* to and consumers *query*. No service boundaries.

---

## 2. Current state — three aggregation paths

The same totals are derived **three independent times**, and the three have drifted:

1. **`recalculateRates()`** (`game/engine.go`) — the authoritative application.
   Combines building production × morale, then `production_all`, then per-resource
   `_rate`, then `gather_rate`, plus research/event flat production, diplomacy,
   storage. This is what actually hits the economy.
2. **`GetState()` → `permanentBonuses` copy** — what is persisted and handed to
   the UI. Critically, this is **only the permanent slice** — it omits active
   (timed) event bonuses and the dynamically-computed wonder/prestige/morale
   contributions.
3. **`overlay_stats.go` Active Multipliers (~L184–393)** — the UI **re-derives**
   attribution a third time, iterating milestone/research/prestige/legacy/wonder
   config defs to rebuild a `bonusAttrib` map, then displays totals from
   `state.PermanentBonuses`.

When three places compute the same thing, they disagree. They do.

### 2.1 Source catalog

| Source | Storage | Accessor | Target namespace | Persisted? |
|--------|---------|----------|------------------|-----------|
| Research | `ResearchManager.bonuses` | `GetBonus(t)` / `GetBonuses()` | any | reconstructed from techs |
| Prestige (passive) | `level` | `GetBonuses()` | `production_all` +2%/lvl, `tick_speed` +1%/lvl | level saved |
| Prestige (upgrades) | `upgrades` map | `GetBonuses()` | upgrade `EffectKey` (e.g. `gather_rate`, `build_cost`) | saved |
| Milestones | `permanentBonuses` (on complete) | direct write | any (`permanent_bonus` effects) | saved |
| Wonders | building `Effects` | `getWonderBonuses()` (per tick) | `bonus` targets | recomputed |
| Wonder speed | `speedMultiplier` field | `GetSpeedMultiplier()` | tick-interval divisor (not a bonus target) | saved |
| Legacy (Succumb) | `legacyBonuses` set + static config | `reapplyLegacyBonuses()` | `<resource>_rate` | set saved |
| Epoch events (good) | direct `permanentBonuses` write **or** `InjectEvent` | mixed | `production_all`, `tick_speed`, … | split |
| Milestone chain boosts | `EventManager.active` (injected) | `GetActiveEffects()` | `tick_speed` | not persisted |
| Catastrophe/event debuffs | `EventManager.active` | `GetActiveEffects()` | `production_all`, … | not persisted |
| Morale | `morale` value | `moraleMultiplier()` | multiplicative on all production | saved |
| Buildings | building `Effects` | `GetEffects()` / `GetStorageBonuses()` | `capacity`, `storage`, `morale` | from counts |

### 2.2 Target vocabulary (and its drift)

In use: `production_all`, `tick_speed`, `research_speed`, `population`,
`military_power`, `expedition_reward`, `gather_rate`, `<resource>_rate`, `all`
(storage), `<resource>` (storage), plus the non-bonus `speedMultiplier` divisor.

Drift / mismatches:
- `production_all` (multiplier) vs `production` (flat output) — same prefix,
  different meaning, easy to confuse; the UI and the engine loop on different
  ones.
- Legacy bonuses use `<resource>_rate` but render in a *separate* "Legacy"
  section, not the per-target breakdown.
- Morale and `speedMultiplier` live in their own namespaces entirely.
- `build_cost` is defined as a prestige target but exists in **no** vocabulary
  consumer.

---

## 3. Bug catalog (the "how broken is it" payload)

| # | Severity | Bug | Evidence |
|---|----------|-----|----------|
| 1 | **HIGH** | Active (timed) event bonuses **apply but are invisible** in Active Multipliers. They hit production in `recalculateRates` but the UI reads only `state.PermanentBonuses`. Player sees "+20%" in the log; the panel shows 0%. | engine.go `recalculateRates` (production_all incl. active events) vs overlay_stats.go reading `state.PermanentBonuses[target]` |
| 2 | **HIGH** | `build_cost` bonus target is **defined but never consumed** — `GetCost()` applies no bonuses. Any prestige upgrade targeting `build_cost` does nothing. | config prestige `build_cost` effects; `buildings.go GetCost()` has no bonus lookup |
| 3 | MED | `bonusAttrib.epoch` field is declared and checked in the UI but **never populated** — epoch permanent bonuses are misattributed (shown as 0 from "epoch"). | overlay_stats.go declares/checks `a.epoch`, nothing increments it |
| 4 | MED | Morale multiplier is applied **outside** all bonus tracking, with a fragile order-of-operations dependency (`× morale` then `× (1+production_all)`). Not in any breakdown. | engine.go morale applied first, production_all second |
| 5 | MED | Active-event `tick_speed` boosts (milestone chains) apply but have **no UI attribution** path. | tick speed recalculated from active events; no UI breakdown |
| 6 | LOW | Split-brain epoch application: some epoch events write `permanentBonuses` directly, others use `InjectEvent` timed effects — same source, two paths. | applyGoodEpochEvent: direct write vs InjectEvent |

Items 1 and 2 are genuine, shippable bugs. The rest are attribution/consistency
debt that guarantees more bugs as sources are added.

---

## 4. Proposed design — a Modifier Resolver

A single registry that all sources contribute to and all consumers query.

### 4.1 Core types

```go
// Op defines how a modifier combines with its peers on the same target.
type Op int
const (
    OpAdd Op = iota // additive percentage points: summed, then applied as ×(1+Σ)
    OpMul           // independent multiplier: multiplied in directly (morale, etc.)
)

type Modifier struct {
    Source string  // stable id for attribution: "research:masonry",
                   // "wonder:colossus", "prestige:passive", "morale",
                   // "event:peaceful_century", "milestone_chain:settlement"
    Target string  // canonical target (see vocabulary below)
    Op     Op
    Value  float64 // OpAdd: +0.10 = +10%.  OpMul: 1.18 = ×1.18
    // (timed effects already live in EventManager; the resolver is rebuilt each
    //  recalc, so expiry is handled by the source not emitting an expired mod.)
}
```

### 4.2 Resolver

```go
type Resolver struct{ mods map[string][]Modifier } // target -> contributions

// Total combines a target's modifiers: (1 + Σ OpAdd) × Π OpMul.
func (r *Resolver) Total(target string) float64

// Breakdown returns the contributing modifiers for UI attribution.
func (r *Resolver) Breakdown(target string) []Modifier
```

`Total` is the one formula. `Breakdown` is the same data, so the UI **cannot
disagree** with the economy.

### 4.3 Pull model (rebuild per recalc)

`recalculateRates()` already runs each tick and already pulls fresh from
`getWonderBonuses()` etc. Keep that: each source exposes `Modifiers() []Modifier`,
the engine builds a fresh `Resolver` per recalc by collecting from all sources.
No stale-state risk, no event-driven invalidation to get wrong. (Caching with
dirty-invalidation is a later optimization if profiling demands it; today's
per-tick re-derivation already costs this.)

```go
func (ge *GameEngine) buildResolver() *Resolver {
    r := NewResolver()
    r.AddAll(ge.Research.Modifiers())
    r.AddAll(ge.Prestige.Modifiers())
    r.AddAll(ge.wonderModifiers())
    r.AddAll(ge.permanentModifiers())   // milestones + legacy + epoch-permanent
    r.AddAll(ge.Events.ActiveModifiers()) // timed events — now first-class
    r.Add(Modifier{Source:"morale", Target:"production.all", Op:OpMul,
                   Value: ge.moraleMultiplier()})
    return r
}
```

### 4.4 Canonical target vocabulary

Normalize to dotted, namespaced targets and map the legacy strings onto them:

```
production.all          rate.<resource>         tick_speed
research_speed          population              military_power
expedition_reward       gather_rate             cost.all / cost.<resource>
storage.all / storage.<resource>
```

`speedMultiplier` (wonder gate) stays its own concept — it is a tick-interval
divisor, not a percentage bonus — but is documented as such so nobody expects it
in the additive pool.

---

## 5. How this fixes the bugs

- **#1 active events invisible** → events emit `Modifier`s; `Breakdown` shows them. UI and economy read the same resolver.
- **#2 `build_cost` dead** → `GetCost()` multiplies by `1 + resolver.Total("cost.all")` (and per-resource). The target finally has a consumer; the prestige upgrades start working.
- **#3 epoch attribution** → epoch modifiers carry `Source:"epoch:..."`; `Breakdown` attributes them correctly. The hand-maintained `bonusAttrib.epoch` field is deleted.
- **#4 morale ordering** → morale is an `OpMul` modifier on `production.all`; combination order is defined once in `Total`, not implicit in line order.
- **#5 tick_speed attribution** → same `Breakdown` mechanism covers `tick_speed`.
- **#6 split-brain epoch** → permanent vs timed epoch effects both surface as modifiers (permanent ones persist, timed ones expire by not being re-emitted); one display path.

The UI's entire re-derivation (`overlay_stats.go` ~L184–393) collapses to
"render `resolver.Breakdown(target)` for each active target." ~200 lines of
hand-aggregation deleted.

---

## 6. Migration plan (incremental, each phase shippable & testable)

1. **Introduce types + resolver** (`game/modifiers.go`) with unit tests for the
   `Total` combination math (Add summing, Mul product, mixed). No behavior change.
2. **Sources emit `Modifiers()`** — add the method to research/prestige/wonders/
   events/permanent, each returning what it already computes, mapped to canonical
   targets. Golden test: resolver `Total` == the current scattered totals for a
   battery of game states (proves no economic drift before switching).
3. **Switch `recalculateRates` to consume the resolver** for the multiplier
   targets (production.all, rate.<r>, gather_rate, tick_speed, storage). Keep flat
   `production` effects as-is (they aren't multipliers). Re-run the golden test.
4. **Switch the UI to `Breakdown`**, delete the re-aggregation and the dead
   `epoch` field. Active events + morale now appear correctly.
5. **Fix the latent bugs** that the resolver makes trivial: wire `cost.all` into
   `GetCost()` (#2), confirm active events show (#1).
6. **Wiki/docs**: update the (now-accurate) Active Multipliers description.

Phases 1–2 are pure addition (zero risk). The economic switch (3) is gated by the
golden equality test, so "did we change the numbers" is answered mechanically.

---

## 7. Open questions / risks

- **Mul targets beyond morale?** Today only morale (and arguably speedMultiplier)
  is genuinely independent-multiplicative. Confirm nothing else should be `OpMul`
  before standardizing everything else as `OpAdd`.
- **Per-resource vs all interaction.** `production.all` and `rate.food` stack
  multiplicatively today (`× (1+all) × (1+food_rate)`). The resolver must preserve
  that — they are *different targets*, each resolved independently, applied in
  sequence by the consumer. Document the consumer-side application order once.
- **Save format.** `permanentBonuses` stays the persisted store for permanent
  modifiers (no save migration needed); the resolver is a runtime view rebuilt
  from it + live sources.
- **Cost of per-tick rebuild.** Measure; add dirty-caching only if it shows up.
- **Scope creep.** This is a *tracking/aggregation* refactor, not a balance
  change — the golden test must prove the numbers are identical. Any intended
  balance change rides separately.

---

## 8. Recommendation

Build it, phased as above. The two HIGH bugs (#1, #2) justify the work on their
own; the architecture win — one contract, one source of truth, new sources plug
in without touching consumers — is what stops the next morale-shaped paper cut.
Estimated as a multi-PR effort (one per phase); phases 1–2 are low-risk
groundwork that can land independently.
