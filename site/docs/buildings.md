# Buildings

AgeForge uses a **13-lineage system** with 284 total buildings: 241 production buildings, 21 storage buildings, and 22 wonders. Buildings belong to lineages that span the full 22-age arc. When you advance an age, buildings that have a next-tier equivalent are **not** automatically transformed — instead they receive a pending upgrade marker and you upgrade them manually at your chosen pace using the `upgrade` command.

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

**`max`** tells the engine to build as many copies as you can afford using the full cumulative cost curve — it stops the moment the next unit would exceed your available resources. It is **not** a flat division of your resources by base cost; each unit is priced correctly at its scaled cost before the decision to continue is made.

## Building Upgrades

When you advance to a new age, buildings that have a next-tier equivalent are **not** automatically transformed. Instead, each such building gets a **pending upgrade** marker and continues producing at its old rate. You choose when to upgrade — building by building, at your own pace.

### The upgrade hint

In the Economy tab building list, upgradeable buildings display a gold hint line:

```
↑ Upgrade available → Farm   type: upgrade gathering_camp
```

This tells you exactly what command to run. Until you upgrade, the old building keeps producing normally — there is no production penalty for leaving it pending. The incentive to upgrade is the higher production rate of the new tier.

### Upgrade commands

```
upgrade <building>         — upgrade ALL copies of that building (default)
upgrade <building> <n>     — upgrade exactly n copies
upgrade <building> all     — same as no count argument
```

There is no `upgrade all` global command — upgrades are per building so you control the order and pacing.

**Examples:**
```
upgrade gathering_camp
upgrade forager_post 3
upgrade forager_post all
```

### Upgrade cost formula

Each copy you upgrade costs only the **delta** between the new building's cost and 50% of the old building's sell value:

```
old_copy_sell  = floor(oldBaseCost × oldCostScale ^ (N−1−i)) × 0.5
new_copy_cost  = floor(newBaseCost × newCostScale ^ i)
delta_per_copy = max(0, new_copy_cost − old_copy_sell)
```

Where `i` is the index of the copy being upgraded (0 = first) and `N` is the total number of old copies. Put simply: **you trade in your old building at 50% value toward the new one.** The net cost is always non-negative (never a free upgrade) but always cheaper than demolishing and rebuilding from scratch — typically 0.5–14% savings.

At high building counts (10+), the old building's sell value may completely cover the new copy's cost for the last few copies, making those final upgrades free. This is intentional — it rewards civilizations that invested heavily in a lineage before advancing.

### Workers and upgrades

- **Full upgrade** (all copies upgraded): workers transfer automatically to the new building. No reassignment needed.
- **Partial upgrade** (only some copies upgraded): workers remain on the old building. The new copies start with workers assigned from your idle pool if available.

### Legacy buildings (pending upgrade)

Once a building has a pending upgrade available:

- It **continues producing** at its current rate — no penalty
- You **cannot build more** copies of the old tier (it is greyed out in the build menu)
- You **can still** `assign`, `unassign`, and `sell` copies of the old tier normally
- The building counts toward your civilization stats as usual

Do not let pending upgrades accumulate for too long during high-demand periods — upgrading your food lineage first before a resource squeeze is almost always the right call.

### Strategic advice

- **Upgrade high-count buildings early.** The cost formula makes the last few copies cheaper to upgrade than the first, so a civilization with 15 Gathering Camps gets proportionally cheaper upgrades than one with 3. Reward your past investment.
- **Upgrade before starvation events.** A Farm produces significantly more than a Gathering Camp. If an epoch catastrophe is incoming, having upgraded food buildings gives you a wider safety margin.
- **You control the order.** You might upgrade your food lineage immediately on age advance and leave military or knowledge buildings pending until you've banked enough resources. There is no time pressure — pending buildings still produce.
- **Don't sell pending buildings for cash.** The 50% sell refund is already baked into the upgrade delta — you get that value back when you upgrade. Selling instead throws away the upgrade discount.

---

### Selling Buildings

```
sell <key>           — demolish 1 copy, recover 50% of its scaled cost
sell <key> <count>   — demolish that many copies (most expensive first)
```

You can sell non-wonder buildings from the **Stone Age onward** (sell is disabled in the Primitive Age). Selling removes copies from the most expensive end of the cost curve first — the refund for each copy is 50% of what that copy originally cost at its scale step.

**Worker handling:** if the sold copies reduce total worker capacity below the number of workers currently assigned, the excess workers are automatically unassigned and returned to the idle pool. They are not dismissed — their population slot is preserved.

**Restrictions:**
- Wonders cannot be sold.
- Buildings currently under construction (in the build queue) cannot be sold until construction completes.
- Sell is not available in the Primitive Age.

### Cost Scaling

Each building has a `BaseCost` and a `CostScale`. The cost of building the next copy depends on how many are already **built** plus how many are currently **in the build queue**:

```
cost of next unit = floor(BaseCost × CostScale ^ (built + queued))
```

When buying multiple at once (`build <key> <count>` or `build <key> max`), each unit in the batch is priced individually at its correct exponent — the total is the sum, not a flat multiple:

```
total cost = sum of floor(BaseCost × CostScale ^ (built + queued + i))
             for i = 0 to count-1
```

This means bulk-buying is always more expensive per unit than buying one at a time, and queueing buildings increases the price of the next purchase even before construction completes.

Costs were rebalanced to flatter curves. Production, military, research, and trade buildings use `CostScale = 1.15`. Storage and housing buildings use the gentler `CostScale = 1.13`. Wonders are flat (`CostScale = 1.0`) — every copy costs the same, but they're capped to one each. Costs still grow exponentially for the scaling lineages, so plan your resource production before bulk-buying.

### Build-cost reductions

A handful of **milestone rewards** (Master Builder, Grand Architect, and others) and two **research techs** (Civil Engineering −5%, Nanofabrication −8%) grant build-cost reductions of −3% to −8% each. These now apply to your cumulative scaled cost: the cost above is multiplied by `(1 + Σ build_cost)`, floored so the cost never drops below 10% of base. Stacked, the currently available reductions reach roughly **−32%**. The discount is reflected in the cost the build menu shows you and in what you're actually charged — there's no hidden mismatch between display and charge. You don't need to do anything to opt in; earning the milestone or completing the tech is enough.

**Example:** Gathering Camp, BaseCost=16 wood, CostScale=1.15, none built or queued:
- 1st: 16 wood
- 2nd: 18 wood
- 5th: 27 wood
- Building 5 at once: 16+18+21+24+27 = **106 wood total** (not 16×5=80)

---

## Lineage System

Buildings no longer unlock individually per age. Instead, each lineage starts at a specific age and gains one new tier per age from there. When your civilization enters a new age:

1. All buildings in active lineages that have a next tier receive a **pending upgrade** marker.
2. The old tier becomes **legacy** — still producing at its old rate, cannot be built again.
3. The new tier becomes available to build (for fresh copies) and to upgrade into (from old copies).

You do not lose progress or production. A Gathering Camp in the Food lineage may be upgraded to a Farm, then a Plantation, then an Agri-Complex — use `upgrade <building>` at each age transition to convert your existing copies. See the [Building Upgrades](#building-upgrades) section for cost details and strategy.

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
| 7 | Military | military | soldiers | 21 | War Camp | Probability War Room |
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

Beyond these effect types, the **Faith (worship) lineage** — shrines, temples, and their later-age tiers — and the **Culture/Arts lineage** (culture and entertainment buildings) now also **restore civilization morale each tick simply by existing** — no workers required. That makes them the primary lever for pushing morale into its production-bonus band, where it multiplies all worker output. See [Morale](morale.md).

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

**Nano Foundry** (Modern Age) is a standalone nanobot producer outside the lineage chain — it gives nanobots a production path as soon as they unlock, two ages before the lineage's own Bio Fabrication Lab. Built by `engineering` workers; +80 nanobots/tick.

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

When an age advance gives a lineage tier a pending upgrade, those buildings become legacy. Legacy buildings:

- Continue producing at their normal rate (no production loss on advance)
- Cannot be built again (greyed out in the build menu)
- Appear with a ☒ indicator in the Economy tab
- Show a gold `↑ Upgrade available` hint with the target building name
- Count toward building totals in your civilization stats
- Can still be assigned workers, unassigned, and sold normally

Your 20 Gathering Camps stay as Gathering Camps — producing at their full rate — until you run `upgrade gathering_camp`. Production only improves once the upgrade is complete, which is the incentive to do it. See [Building Upgrades](#building-upgrades) for the full guide.

---

## Storage Buildings (21 tiers)

Storage buildings form their own lineage and expand your resource cap for **every** resource simultaneously. Unlike production buildings, storage buildings require no workers — just build them and the caps go up.

Every storage building is **capped at 25 copies** (Stash at 50). The cap exists by design: each tier's storage-per-copy is at least its dominant build cost, so a full stack always provides more cap than it costs and a storage building never becomes an unaffordable wall — but the flattened cost curve would eventually outrun even that, so the stack is bounded just below where it would. Advance to the next tier instead of over-stacking.

| Building | Age | Effect | Max |
|----------|-----|--------|-----|
| Stash | Primitive | +300 all storage | 50 |
| Storage Pit | Stone | +600 all storage | 25 |
| Warehouse | Bronze | +4,000 all storage | 25 |
| Granary | Iron | +15,000 all storage | 25 |
| Classical Vault | Classical | +100,000 all storage | 25 |
| Keep | Medieval | +400,000 all storage | 25 |
| Renaissance Vault | Renaissance | +500,000 all storage | 25 |
| Colonial Warehouse | Colonial | +10M all storage | 25 |
| Industrial Depot | Industrial | +50M all storage | 25 |
| Victorian Vault | Victorian | +350M all storage | 25 |
| Electric Warehouse | Electric | +3.5B all storage | 25 |
| Atomic Vault | Atomic | +20B all storage | 25 |
| Modern Depot | Modern | +90B all storage | 25 |
| Info Vault | Information | +260B all storage | 25 |
| Digital Archive | Digital | +1.5T all storage | 25 |
| Cyber Vault | Cyberpunk | +8T all storage | 25 |
| Fusion Vault | Fusion | +30T all storage | 25 |
| Orbital Depot | Space | +200T all storage | 25 |
| Stellar Vault | Interstellar | +500T all storage | 25 |
| Galactic Vault | Galactic | +2Q all storage | 25 |
| Quantum Vault | Quantum | +10Q all storage | 25 |

> **Tip:** Stash is hard-capped at 50. Transition to Storage Pit the moment you enter the Stone Age. Storage bottlenecks stop all late-game resource accumulation dead — build storage first when entering every new age.

Storage buildings also participate in the upgrade system on age advance — use `upgrade <key>` to convert them to the next tier at the delta cost (see [Building Upgrades](#building-upgrades)).

---

## Cultural Monuments (4)

Cultural Monuments are one-off structures (Category: `monument`, capped at one copy each) that exist to turn surplus **culture** into a permanent payoff. Each costs a large lump of culture (plus other age-appropriate materials) and grants a permanent `production_all` bonus on construction. Unlike wonders, they need no resource banking — build them with the normal `build <key>` command.

| Monument | Age | Culture Cost | Permanent Bonus |
|----------|-----|--------------|-----------------|
| Cultural Obelisk | Classical | 2,500 | +1% all production |
| Grand Amphitheatre | Medieval | 25,000 | +2% all production |
| Eternal Library | Industrial | 500,000 | +3% all production |
| Monument of Ages | Modern | 25,000,000 | +5% all production |

Monuments are one of the two culture sinks (the other is the `festival` command); prestige gates remain the primary long-term sink. See [Culture](resources.md#culture).

---

## Wonders (22)

Wonders are unique one-of-a-kind buildings — you can only build each wonder once. They require resource banking before construction can begin:

```
wonder collect <resource> <amount|all>   — bank resources toward the active wonder
build <wonder_key>                       — begin construction once fully funded
```

Wonders provide powerful civilization-wide bonuses. Each wonder completed grants a permanent +0.5× speed boost to your game tick rate.

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

In addition to culture, these buildings now also help **restore morale each tick** (see [Morale](morale.md)), so they pull double duty — culture for epoch outcomes and morale for a production bonus.

Culture bonuses gate certain epoch event outcomes and diploma actions — a civilization rich in culture weathers epoch transitions more gracefully.

---

## Economy Tab

Press `status` to open the Economy tab. Buildings are displayed by lineage on the right panel. Legacy buildings show ☒. The left panel shows your current resource rates and worker assignments.

Use `rates` to print the current production/consumption rates for all resources directly in the command output.
