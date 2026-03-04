# AgeForge — Resources

## Overview

AgeForge has **21 resources** organized into tiers. Resources are produced by buildings,
consumed by costs and workers (food drain), and capped by storage buildings.

Most resources are **flow resources** — produced and consumed continuously, capped at
storage max, wiped on prestige (with multiplier bonuses carrying over).

Two resources have **special accumulation rules**: Culture and Faith.

---

## Resource Index

| # | Resource | Type | Unlock Age | Notes |
|---|----------|------|-----------|-------|
| 1 | food | Flow | Primitive | Worker drain resource; negative = workers starve |
| 2 | wood | Flow | Primitive | Core construction material |
| 3 | stone | Flow | Primitive | Core construction material |
| 4 | knowledge | Flow | Primitive | Fuels research/tech; produced by Knowledge buildings |
| 5 | faith | Special (draining) | Primitive | Powers morale, cohesion, diplomacy — see below |
| 6 | gold | Flow | Bronze | Currency; trade, some building costs |
| 7 | iron | Flow | Iron | Smelted construction; mid-tier buildings |
| 8 | culture | Special (accumulating) | Classical | Accumulates forever; gates bonuses — see below |
| 9 | steel | Flow | Renaissance | Refined iron; advanced construction |
| 10 | coal | Flow | Industrial | Fuel for industrial buildings |
| 11 | oil | Flow | Modern | Fuel for modern buildings/energy |
| 12 | electricity | Flow | Electric | Powers electric+ buildings |
| 13 | uranium | Flow | Atomic | Nuclear fuel |
| 14 | titanium | Flow | Atomic | Advanced construction material |
| 15 | data | Flow | Information | Digital economy; fuels Hacker buildings |
| 16 | crypto | Flow | Digital | Digital currency; late trade routes |
| 17 | plasma | Flow | Fusion | Fusion energy; late-game fuel |
| 18 | dark_matter | Flow | Space | Exotic material; warp/dyson construction |
| 19 | antimatter | Flow | Galactic | Endgame construction; warp drives |
| 20 | quantum_flux | Flow | Quantum | Final-tier resource; gates quantum wonders |
| 21 | nanobots | Flow | Cyberpunk | Augmentation construction material |

---

## Standard Flow Resources

These follow the same rules:
- Produced by buildings (worker-modified)
- Drained by worker food costs, building queue costs, tech research costs
- Capped by storage buildings (see lineages.md Storage table)
- Reset to 0 on prestige (prestige bonuses increase production rate multiplier, not stored amount)
- Storage cap must always satisfy **Law 1 (Storage Covenant)** from economy.md

---

## Faith — Special Draining Resource

### What Faith Is

Faith is a **managed drain resource**. It is produced by Faith-lineage buildings (Shrine →
Transcendence Hall) via Faith workers (Shaman → Transcendent). It drains passively each
tick from a base "cultural entropy" rate plus a drain-per-citizen rate.

Players must actively maintain faith production or it decays toward zero. Unlike food
(where starvation is punishing), falling to 0 faith has tiered consequences that increase
over time — not an instant catastrophe.

### Production and Drain

```
faith_rate = Σ(faith_buildings × worker_output) - base_drain - (population × drain_per_capita)
```

`base_drain` is a small constant per age tier (increases each age).
`drain_per_capita` is tiny but scales with total population — larger civilizations need
more faith investment proportionally.

### Faith Storage

Faith uses standard storage (same as other resources). Storage buildings contribute to
faith cap. Faith does NOT accumulate like culture — it is spent, drained, and must be
replenished.

### Faith Uses

| Use | Mechanic | Notes |
|-----|----------|-------|
| **Morale bonus** | Soldiers gain +morale multiplier based on faith level | Full faith = max morale; 0 faith = morale penalty. Affects expedition success rates |
| **Cohesion** | Reduces speed/severity of negative random events | High faith = bad events less frequent and shorter; 0 faith = events hit harder |
| **Diplomatic influence** | Faith-aligned factions give better trade rates and opinion | Certain factions (religious, tribal) care about faith level in your civ |
| **Prestige multiplier** | At prestige time, faith level above threshold grants +% prestige bonus | Faith > 80% of cap = +10% prestige bonus per tier above threshold |
| **Wonder requirements** | Some wonders require a minimum faith threshold to unlock/build | e.g. "Temple of the Cosmos" requires 50,000 faith as a build prerequisite |

### Faith Threshold Effects

| Faith Level (% of cap) | Effect |
|------------------------|--------|
| 0% | −25% soldier morale; events hit +50% harder; no diplomatic faith bonus |
| 1–25% | −10% morale; minor cohesion penalty |
| 26–50% | Neutral (no bonus, no penalty) |
| 51–75% | +5% morale; events 10% less frequent |
| 76–99% | +10% morale; +15% cohesion; +5% faction opinion with faith-aligned civs |
| 100% (at cap) | +15% morale; +25% cohesion; +10% faction opinion; prestige multiplier active |

### Faith and Prestige

At prestige: faith is wiped to 0 (same as other resources). However:
- If faith was at 100% cap at time of prestige → +10% prestige bonus to all post-prestige multipliers
- Faith buildings and workers do NOT carry through prestige (same as everything else)

---

## Culture — Special Accumulating Resource

### What Culture Is

Culture is a **permanent accumulation resource**. It never drains naturally. Once culture
is earned, it is yours forever — representing the lasting cultural achievements of your civilization.

Culture is produced automatically by Culture/Arts-lineage buildings (Amphitheater → Reality Art
Engine). **No workers are assigned to these buildings** — they produce culture passively
just by existing.

Culture does NOT tick down. Culture does NOT get spent on purchases. Culture is purely a
**milestone and unlock tracker** — it accumulates until the cap, unlocking bonuses at thresholds.

### Production

```
culture_per_tick = Σ(culture_buildings × culture_rate)
```

Each culture building has a fixed culture/tick rate (see lineages.md Lineage 10). There is no
worker assignment and no multiplier from workers. Buildings in higher ages produce more culture
per tick.

### Culture Cap

Culture has a **maximum cap** set by the total culture buildings owned:

```
culture_cap = Σ(culture_buildings × cap_contribution_per_building)
```

Once culture reaches its cap, production pauses (the culture display shows "at cap").
To grow culture further, build more culture buildings to expand the cap.

This creates an explicit investment loop: build culture buildings → cap rises → culture fills
to new threshold → unlock bonus → repeat.

### Culture Thresholds and Bonuses

Bonuses are **permanent** once unlocked — they don't go away if culture somehow drops
(which it normally can't, except prestige partial reset described below).

| Culture Milestone | Bonus Unlocked |
|-------------------|---------------|
| 500 | +5% food production (cultural tradition of harvest festivals) |
| 2,500 | Unlock first Culture Wonder requirement satisfied |
| 10,000 | +10% knowledge production (scholars inspired by culture) |
| 50,000 | +5% all resource production (cultural identity bonus) |
| 250,000 | Diplomatic: all factions start +10 opinion |
| 1,000,000 | Unlock Cultural Victory condition (prestige path) |
| 5,000,000 | +15% research speed |
| 25,000,000 | Golden Age trigger: temporary (30 min) 2× production for all resources |
| 100,000,000 | +20% prestige multiplier on next prestige |
| 500,000,000 | Galactic cultural hegemony: +30% all faction trade rates |
| 1,000,000,000 | Legacy bonus: carries +5% production multiplier through prestige |

> **Note:** Thresholds are design targets. Tune against the build time curve (Law 2) so
> milestones feel like genuine achievements, not instant unlocks.

### Culture and Prestige

Culture is **partially wiped** on prestige — not fully reset, not kept:

```
culture_after_prestige = culture_before_prestige × 0.20
```

The player keeps 20% of their culture, representing their civilization's enduring legacy.
This means:
- A player who grinded to 1B culture keeps 200M after prestige
- Previously unlocked culture bonuses **remain unlocked** (they're permanent once hit)
- They still need to fill back up to the cap to continue producing culture

This makes culture a satisfying long-term investment — it survives prestige partially,
and the bonuses are permanent, so there's no regret in investing in culture buildings.

### Culture Display

In the UI, culture is displayed differently from other resources:
- Shows as a progress bar toward the next threshold (not a raw number)
- The "cap" is shown clearly so players know when to build more culture buildings
- Tooltip shows: current / cap / next milestone / current bonuses active

---

## Resource Interactions (Key)

```
food → workers → production (all resources)
wood + stone → buildings → more production
knowledge → research → tech bonuses → multipliers on all production
faith → morale + cohesion + diplomacy + prestige multiplier
culture → threshold bonuses + wonder requirements + prestige legacy
gold → trade routes → resource conversion
iron → steel → titanium → advanced construction
coal → electricity → plasma → quantum_flux (energy chain)
data + crypto → late-game digital economy
dark_matter + antimatter → space/galactic construction
```

---

## Resource Unlock Chain by Age

| Age | New Resource(s) Unlocked |
|-----|------------------------|
| Primitive | food, wood, stone, knowledge, faith |
| Stone | (no new resources — scale up existing) |
| Bronze | gold |
| Iron | iron |
| Classical | culture |
| Medieval | (no new resources) |
| Renaissance | steel |
| Colonial | (no new resources) |
| Industrial | coal |
| Victorian | (no new resources) |
| Electric | electricity |
| Atomic | uranium, titanium |
| Modern | oil |
| Information | data |
| Digital | crypto |
| Cyberpunk | nanobots |
| Fusion | plasma |
| Space | dark_matter |
| Interstellar | (scale up dark_matter) |
| Galactic | antimatter |
| Quantum | quantum_flux |

---

## Implementation Notes

- `faith` needs: production rate (via workers), passive drain (base + per_capita), threshold
  effect table, prestige behavior
- `culture` needs: production rate (fixed per building, no workers), cap calculation,
  threshold milestone trigger, partial-wipe on prestige (×0.20), UI progress-bar display
- `ResourceDef` may need a `ResourceBehavior` enum: `flow`, `accumulating`, `draining`
  to differentiate behavior in ResourceManager
- Culture milestones should fire via the EventBus (same as regular milestones) so toast
  notifications appear and MilestoneManager can track them
