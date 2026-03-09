# Buildings

AgeForge uses a **13-lineage system** with 284 total buildings: 241 production buildings, 21 storage buildings, and 22 wonders. Buildings belong to lineages that span the full 22-age arc. When you advance an age, buildings in your current lineage tier automatically transform into the next tier — you keep your count and gain the upgraded building.

For the full wonder list see [Wonders](wonders.md).

---

## Why Buildings Matter

Buildings are the engine of your civilization. Every resource you'll ever need comes from buildings — workers assigned to production buildings are what turn raw capacity into actual output. More buildings mean:

- **More production** — each additional building of the same tier adds its base rate to your total
- **More worker slots** — worker capacity scales with building count, raising your production ceiling
- **More storage** — storage buildings expand caps for every resource simultaneously
- **Age unlock requirements** — many ages require a minimum building count or specific buildings to be present

The single most impactful action in AgeForge is: build more buildings, fill them with workers.

---

## Building Commands

```
build <key>              — queue one building (must have enough resources)
build <key> <count>      — queue and build that many copies
build <key> max          — build as many as you can afford right now
build cancel             — cancel the current build queue item
build                    — list all available buildings (with costs and count built)
```

**`max`** is a sentinel that tells the engine "keep building until resources run out or the MaxCount limit is hit." It's the fastest way to fill up a lineage tier at the start of an age when resources are plentiful.

### Cost Scaling

The cost of the N-th instance of a building is:

```
cost = BaseCost × CostScale ^ (N-1)
```

Most production buildings use `CostScale = 1.15`. Storage buildings use `CostScale = 1.20`. This means costs grow exponentially as you build more copies of the same building. Plan your resource production before bulk-buying.

**Example:** If a Gathering Camp costs 30 food at count 0, the second costs `30 × 1.15 = 34.5`, the fifth costs `30 × 1.15^4 ≈ 52.4`, and so on.

### Upgrade Command

Some buildings have explicit upgrade paths (distinct from automatic lineage transforms on age advance). These are manual one-time upgrades:

```
upgrade                  — list all available building upgrades
upgrade <building_key>   — upgrade all buildings of that type to the next tier
upgrade all              — upgrade everything affordable across all chains
```

Upgrades cost a fraction of the target building's base cost (typically **25%**). Workers remain assigned through the upgrade — no reassignment needed.

#### Upgrade Chains

| From | To | Min Age | Cost |
|------|----|---------|------|
| `hut` | `house` | Bronze | 25% of House base |
| `house` | `manor` | Medieval | 25% of Manor base |
| `manor` | `tenement` | Industrial | 25% of Tenement base |
| `tenement` | `tower_block` | Modern | 25% of Tower Block base |
| `tower_block` | `arcology_pod` | Cyberpunk | 25% of Arcology base |
| `arcology_pod` | `orbital_habitat` | Space | 25% of Orbital Habitat base |
| `stash` | `storage_pit` | Stone | 25% of Storage Pit base |
| `storage_pit` | `warehouse` | Bronze | 25% of Warehouse base |
| `warehouse` | `classical_vault` | Classical | 25% of Classical Vault base |
| `classical_vault` | `industrial_depot` | Industrial | 25% of Industrial Depot base |
| `industrial_depot` | `modern_depot` | Modern | 25% of Modern Depot base |
| `modern_depot` | `info_vault` | Information | 25% of Info Vault base |
| `info_vault` | `digital_archive` | Digital | 25% of Digital Archive base |
| `digital_archive` | `cyber_vault` | Cyberpunk | 25% of Cyber Vault base |
| `cyber_vault` | `fusion_vault` | Fusion | 25% of Fusion Vault base |
| `fusion_vault` | `orbital_depot` | Space | 25% of Orbital Depot base |
| `orbital_depot` | `stellar_vault` | Interstellar | 25% of Stellar Vault base |
| `stellar_vault` | `galactic_vault` | Galactic | 25% of Galactic Vault base |
| `galactic_vault` | `quantum_vault` | Quantum | 25% of Quantum Vault base |
| `story_circle` | `scriptorium` | Stone | 25% of Scriptorium base |
| `scriptorium` | `library` | Bronze | 25% of Library base |
| `library` | `university` | Medieval | 25% of University base |
| `gathering_camp` | `farm` | Bronze | 25% of Farm base |
| `woodcutter_camp` | `lumber_mill` | Bronze | 25% of Lumber Mill base |
| `stone_pit` | `quarry` | Bronze | 25% of Quarry base |

---

## Lineage System

Buildings no longer unlock individually per age. Instead, each lineage starts at a specific age and gains one new tier per age from there. When your civilization enters a new age:

1. All buildings in active lineages that have a next tier **transform** automatically.
2. The old tier becomes **legacy** — still producing, cannot be built again.
3. The new tier becomes the buildable version at its upgraded stats.

You do not lose progress. A Gathering Camp in the Food lineage may become a Farm, then a Plantation, then an Agri-Complex — your existing count carries forward with full worker assignments intact.

---

## The 13 Production Lineages

| # | Lineage | Domain | Primary Output | Tiers | First Building | Final Building |
|---|---------|--------|----------------|-------|----------------|----------------|
| 1 | Housing | None | Population cap | 21 | Hut (+10 pop) | Reality Fold (+1M pop) |
| 2 | Food | food | food | 21 | Gathering Camp | Quantum Cultivator |
| 3 | Organic Extraction | lumber | wood → coal → oil → nanobots → quantum flux | 21 | Wood Camp | Reality Harvester |
| 4 | Geological Extraction | masonry | stone → iron ore → uranium → titanium ore → dark matter crystals → antimatter | 21 | Stone Camp | Reality Excavator |
| 5 | Knowledge | knowledge | knowledge | 21 | Story Circle | Reality Academy |
| 6 | Faith | faith | faith | 21 | Shrine | Transcendence Hall |
| 7 | Military | military | soldiers | 21 | Hunting Lodge | Probability War Room |
| 8 | Trade | trade | gold | 19 | Market | Probability Market |
| 9 | Engineering | engineering | iron → steel → electricity → plasma → dark matter → quantum flux | 19 | Smithy | Reality Forge |
| 10 | Culture/Arts | None | culture | 17 | Amphitheater | Reality Art Engine |
| 11 | Metallurgy | metallurgy | iron → steel → titanium → dark matter → antimatter → quantum flux | 18 | Smelter | Quantum Metal Works |
| 12 | Energy | energy | coal → electricity → plasma → dark matter → quantum flux | 13 | Coal Plant | Zero Point Generator |
| 13 | Hacker/Digital | hacker | data | 8 | Server Farm | Reality Processor |

> Lineages 8–13 start in later ages (Bronze, Iron, Victorian, Information) so they have fewer tiers. The Housing and Culture lineages have no worker domain — they produce passively without worker assignment.

---

## Building Effects Reference

Every building's power is defined by one or more `Effect` entries. Understanding effect types helps you prioritize what to build.

| Effect Type | What It Does |
|-------------|-------------|
| `production` | Adds `Value` of `Target` resource per tick, worker-scaled. Most production buildings use this. |
| `storage` | Increases the global resource cap for `Target` resource (or `"all"` to raise every cap). |
| `capacity` | Increases a cap value — e.g. `Target: "population"` raises max worker population. |
| `bonus` | Applies a fractional multiplier bonus to a rate. Value is a fraction (`0.10` = +10%). |
| `unlock` | Unlocks a building key or resource key when built. Used by wonders and some techs. |
| `instant_resource` | Immediately adds `Value` of `Target` resource. Used by milestone rewards only. |
| `permanent_bonus` | Persistent multiplier on `Target` rate. Used by milestone rewards only. |

### Worker Scaling Formula

Production buildings with a worker domain use:

```
actual_rate = base_rate × count × (0.20 + 0.80 × workers_assigned / total_capacity)
```

At **0 workers** a building still produces at **20% of base rate** (the idle floor). At **full staffing** it produces at **100%**. This means an empty building isn't wasted — it's just throttled.

Buildings without a worker domain (Housing, Culture/Arts) produce at exactly `base_rate × count` regardless of workers.

---

## Lineage Output Progression

Several lineages change what resource they produce as you advance through epochs. The resource key in the `OutputResource` field remaps automatically on age transition — you don't need to rebuild or reconfigure anything.

### Organic Extraction (Lineage 3)

| Era | Output Resource |
|-----|----------------|
| Stone / Iron Era (tiers 1–5) | wood |
| Steel Era (tier 6+) | coal |
| Electric Era | oil |
| Digital Era+ | nanobots |
| Neon / Cosmic Era | quantum flux |

### Geological Extraction (Lineage 4)

| Era | Output Resource |
|-----|----------------|
| Stone / Bronze Era | stone |
| Iron Era | iron ore |
| Steel / Industrial Era | uranium |
| Modern Era | titanium ore |
| Neon Era | dark matter crystals |
| Cosmic Era | antimatter |

### Engineering (Lineage 9)

| Era | Output Resource |
|-----|----------------|
| Bronze / Iron Era | iron |
| Medieval / Steel Era | steel |
| Victorian / Electric Era | electricity |
| Fusion Era | plasma |
| Galactic Era | dark matter |
| Quantum Era | quantum flux |

### Metallurgy (Lineage 11)

Consumes raw ore from Geological Extraction and refines it into usable metals:

| Era | Ore In | Metal Out |
|-----|--------|----------|
| Iron / Medieval Era | iron ore | iron |
| Industrial / Electric Era | iron ore | steel |
| Modern / Space Era | titanium ore | titanium |
| Interstellar / Neon Era | dark matter crystals | dark matter |
| Cosmic / Galactic Era | dark matter crystals | antimatter |
| Quantum Era | — | quantum flux |

---

## Legacy Buildings

When a lineage tier upgrades, the previous-tier buildings become legacy. Legacy buildings:

- Continue producing at their normal rate
- Cannot be built again (greyed out in the build menu)
- Appear with a ☒ indicator in the Economy tab
- Count toward building totals in your civilization stats

You never lose production from an age advance. Your 20 Farms become 20 Plantations (or whatever the next tier is) automatically, with workers reassigned to match.

---

## Storage Buildings (21 tiers)

Storage buildings form their own lineage and expand your resource cap for **every** resource simultaneously. Unlike production buildings, storage buildings require no workers — just build them and the caps go up.

| Building | Age | Effect |
|----------|-----|--------|
| Stash | Primitive | +300 all storage (hard-capped at 50) |
| Storage Pit | Stone | +500 all storage |
| Warehouse | Bronze | +3,000 all storage |
| Granary | Iron | +12,000 all storage |
| Classical Vault | Classical | +25,000 all storage |
| Keep | Medieval | +60,000 all storage |
| Renaissance Vault | Renaissance | +500,000 all storage |
| Colonial Warehouse | Colonial | +10M all storage |
| Industrial Depot | Industrial | +50M all storage |
| Victorian Vault | Victorian | +350M all storage |
| Electric Warehouse | Electric | +3.5B all storage |
| Atomic Vault | Atomic | +15B all storage |
| Modern Depot | Modern | +45B all storage |
| Info Vault | Information | +250B all storage |
| Digital Archive | Digital | +1.5T all storage |
| Cyber Vault | Cyberpunk | +5T all storage |
| Fusion Vault | Fusion | +30T all storage |
| Orbital Depot | Space | +200T all storage |
| Stellar Vault | Interstellar | +500T all storage |
| Galactic Vault | Galactic | +2Q all storage |
| Quantum Vault | Quantum | +5Q all storage |

> **Tip:** Stash is hard-capped at 50. Transition to Storage Pit the moment you enter the Stone Age. Storage bottlenecks stop all late-game resource accumulation dead — build storage first when entering every new age.

You can manually upgrade storage buildings between tiers using the `upgrade` command at a 25% cost discount (see upgrade chains above).

---

## Wonders (22)

Wonders are unique one-of-a-kind buildings — you can only build each wonder once. They require resource banking before construction can begin:

```
wonder collect <resource> <amount|all>   — bank resources toward the active wonder
build <wonder_key>                       — begin construction once fully funded
```

Wonders provide powerful civilization-wide bonuses and contribute to your **WonderBank** — the fill level of your wonder collection feeds into epoch event defense rolls, reducing the probability of bad epoch outcomes.

Wonders are excluded from building destruction during catastrophe Endure events, and excluded from Succumb ruins generation. They are permanent fixtures of your civilization.

See [Wonders](wonders.md) for the full list with costs and effects.

---

## Ruins

Ruins are ancient remnants of a previous civilization — created when you choose **Succumb** during a catastrophe event. Up to 8 buildings are converted into ruins and carried forward into your new run.

**How ruins differ from normal buildings:**

| Property | Normal Building | Ruin |
|----------|----------------|------|
| Production | `base × count × (0.20 + 0.80 × fill)` | `base × count × 0.50` |
| Worker assignment | Required for full output | Not needed (produces automatically) |
| Can build more? | Yes | No (ruins only; not in build menu) |
| Destroyed by Endure? | Yes (20% randomly) | No (wonders only excluded; ruins can be destroyed) |
| Appears in Economy tab | Yes | Yes, marked as ruins |

Ruins give your new run a free head start. They produce at 50% base rate with no workers attached, which means food, wood, and stone (or whatever the ruin type produced) trickle in from tick 1 without any setup.

The more Succumbs you've accumulated across runs, the stronger your ruins collection can become — each Succumb lets you bank 8 new ruins on top of whatever survived from prior runs.

---

## Worker Capacity

Every production building has a `WorkerCapacity` value. The total available worker slots for a domain equals:

```
total_capacity = building_count × WorkerCapacity
```

More buildings mean more worker slots, which mean a higher production ceiling. Building the same lineage tier repeatedly is how you scale output in a given age before the next transformation.

Workers are recruited with `recruit [count|max]` and assigned with `assign <building> [count|all]`. Until workers are assigned to a building, that building produces at only 20% of its base rate.

---

## Culture Buildings (Lineage 10)

Culture buildings have no worker domain. They automatically produce culture each tick based on their base rate, regardless of worker assignment. They also contribute to your culture storage cap. Build them to accumulate culture passively and unlock culture threshold bonuses.

Culture bonuses gate certain epoch event outcomes and diploma actions — a civilization rich in culture weathers epoch transitions more gracefully.

---

## Economy Tab

Press `status` to open the Economy tab. Buildings are displayed by lineage on the right panel. Legacy buildings show ☒. The left panel shows your current resource rates and worker assignments.

Use `rates` to print the current production/consumption rates for all resources directly in the command output.
